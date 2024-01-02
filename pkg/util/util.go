// Package util has some helpers for setting up common components.
package util

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
)

// Config is passed to [SetupDefaultEnvironment] to configure the environment
type Config struct {
	ServiceName    string       // ServiceName is applied to the otel resource
	ServiceVersion string       // ServiceVersion is applied to the otel resource
	OTLPEndpoint   string       // OTLPEndpoint is the endpoint for the OTLP exporter
	LogLevel       slog.Leveler // LogLevel is the log level
}

// SetupDefaultEnvironment creates and registers components for logging, metrics, and tracing
func SetupDefaultEnvironment(ctx context.Context, c Config) error {

	// resource

	r, err := NewDefaultResource(c.ServiceName, c.ServiceVersion)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// logger

	logger := NewLoggerForResource(r, c.LogLevel)
	slog.SetDefault(logger)
	otel.SetLogger(LogrFromSlog(logger))

	// metrics

	mp, err := NewMeterProviderForResource(r)
	if err != nil {
		return fmt.Errorf("failed to create meter provider: %w", err)
	}
	otel.SetMeterProvider(mp)

	// traces

	tp, err := NewTracerProviderForResource(ctx, r,
		otlptracegrpc.WithEndpoint(c.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{"x-scope-orgid": "1"}),
	)
	if err != nil {
		return fmt.Errorf("failed to create tracer provider: %w", err)
	}
	otel.SetTracerProvider(tp)

	// TraceContext is used to propagate trace context across process boundaries

	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	return nil
}
