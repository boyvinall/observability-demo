// Package main is the entry point for the server binary
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	cli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/boyvinall/go-observability-app/pkg/boomerserver"
	"github.com/boyvinall/go-observability-app/pkg/logutil"
)

//nolint:funlen
func run(address string) error {
	hostName := os.Getenv("HOSTNAME")
	serviceName := "MyBoomerServer"

	ctx := context.Background()
	g := errgroup.Group{}

	//--------------------------------------------------
	//
	//  set up the logger
	//
	//   - do this before creating any other objects
	//     so that other constructors can use it
	//
	//--------------------------------------------------

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).
		With("hostname", hostName).
		With("service_name", serviceName)
	slog.SetDefault(logger)
	otel.SetLogger(logutil.LogrFromSlog(logger))

	//--------------------------------------------------
	//
	//  create the resource
	//
	//  - this is used by metrics and traces
	//
	//--------------------------------------------------

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.HostName(hostName),
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	//--------------------------------------------------
	//
	//  metrics
	//
	//--------------------------------------------------

	promexp, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to initialize prometheus exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(promexp),
		metric.WithResource(r),
	)
	otel.SetMeterProvider(mp)

	// Start the prometheus HTTP server
	g.Go(func() error {
		slog.Info("serving metrics", "address", "localhost:2223/metrics")

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		metricServer := &http.Server{
			Addr:              "0.0.0.0:2223",
			ReadHeaderTimeout: 3 * time.Second, // fix for gosec G114
			Handler:           mux,
		}
		return metricServer.ListenAndServe()
	})

	//--------------------------------------------------
	//
	//  traces
	//
	//--------------------------------------------------

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("tempo:4317"),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{"x-scope-orgid": "1"}),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exp),
		trace.WithResource(r),
	)
	otel.SetTracerProvider(tp)

	//--------------------------------------------------
	//
	//  set up the app
	//
	//--------------------------------------------------

	// create the GRPC server first so that services can register themselves to it
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			logutil.UnaryServerInterceptor(logger),
			// Add any other interceptor you want
		),
	)
	reflection.Register(grpcServer)

	// create the server
	_, err = boomerserver.New(grpcServer)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	logger.Info("Listening", "address", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// start the GRPC server
	g.Go(func() error {
		return grpcServer.Serve(lis)
	})

	//--------------------------------------------------
	//
	//  wait for the app to exit
	//
	//--------------------------------------------------

	return g.Wait()
}

func main() {
	app := &cli.App{
		Name:  "boom",
		Usage: "make an explosive entrance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "address",
				Usage: "address to listen on",
				Value: "0.0.0.0:8080",
			},
		},
		Action: func(c *cli.Context) error {
			return run(c.String("address"))
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("unable to run app", "error", err)
		os.Exit(1)
	}
}
