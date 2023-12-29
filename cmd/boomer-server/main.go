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

	"github.com/nats-io/nats.go"
	p "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/boyvinall/go-observability-app/pkg/boomerserver"
	"github.com/boyvinall/go-observability-app/pkg/logutil"
	"github.com/boyvinall/go-observability-app/pkg/worker"
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

	// Create the prometheus exporter
	// For an example with more config options, see https://docs.daocloud.io/en/insight/06UserGuide/01quickstart/otel/meter/#create-an-initialization-function-using-the-opentelemetry-sdk
	metricExp, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to initialize prometheus exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metricExp),
		metric.WithResource(r),
	)
	otel.SetMeterProvider(mp)

	// Start the prometheus HTTP server
	g.Go(func() error {
		slog.Info("serving metrics", "address", "localhost:2223/metrics")

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(p.DefaultGatherer, promhttp.HandlerOpts{EnableOpenMetrics: true}))
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

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("tempo:4317"),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{"x-scope-orgid": "1"}),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(traceExp),
		trace.WithResource(r),
	)
	otel.SetTracerProvider(tp)

	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

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

	// messaging
	c, err := nats.Connect("nats://nats:4222")
	if err != nil {
		return err
	}

	// create the worker
	_, err = worker.New(c)
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	// create the server
	_, err = boomerserver.New(grpcServer, c)
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
