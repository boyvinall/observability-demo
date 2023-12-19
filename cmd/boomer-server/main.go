// Package main is the entry point for the server binary
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/boyvinall/go-observability-app/pkg/server"
)

func serveMetrics() {
	log.Printf("serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe("0.0.0.0:2223", nil) //nolint:gosec // Ignoring G114: Use of net/http serve function that has no support for setting timeouts.
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}

func run(address string) error {
	ctx := context.Background()
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("MyBoomerServer"),
			semconv.ServiceVersion("0.0.0"),
		),
	)

	promexp, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(promexp),
		metric.WithResource(r),
	)
	otel.SetMeterProvider(mp)

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("tempo:4317"),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{"x-scope-orgid": "1"}),
	)
	if err != nil {
		log.Fatal(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exp),
		trace.WithResource(r),
	)
	otel.SetTracerProvider(tp)

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()

	// create the GRPC server first so that services can register themselves to it
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	_, err = server.New(grpcServer)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		return err
	}

	slog.Info("Listening", "address", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return err
	}

	reflection.Register(grpcServer)
	return grpcServer.Serve(lis)
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
