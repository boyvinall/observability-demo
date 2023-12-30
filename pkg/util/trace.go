package util

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewTracerProviderForResource creates an OTEL TracerProvider with a default resource
func NewTracerProviderForResource(ctx context.Context, r *resource.Resource, opts ...otlptracegrpc.Option) (*trace.TracerProvider, error) {
	traceExp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(traceExp),
		trace.WithResource(r),
	)

	return tp, nil
}
