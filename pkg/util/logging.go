package util

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// loggerKey is used to store the logger in the context
type loggerKey struct{}

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor that helps to make a logger
// accessible from GRPC context.
func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = SetContext(ctx, logger)
		return handler(ctx, req)
	}
}

// SetContext sets the logger in the context
func SetContext(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFromContext returns the logger from the context
func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}

	// --8<-- [start:logger-from-context]
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		logger = logger.With("trace_id", spanContext.TraceID().String())
	}
	if spanContext.HasSpanID() {
		logger = logger.With("span_id", spanContext.SpanID().String())
	}
	// --8<-- [end:logger-from-context]

	return logger
}

// LogrFromSlog returns a logr.Logger from a slog.Logger
// This is useful for setting up the opentelemetry logger, which uses logr.
func LogrFromSlog(logger *slog.Logger) logr.Logger {
	return logr.FromSlogHandler(logger.Handler())
}

// NewLoggerForResource creates a new logger with some fields from the resource
// --8<-- [start:new-logger-for-resource]
func NewLoggerForResource(r *resource.Resource, level slog.Leveler) *slog.Logger {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	if hostName := GetResourceHostName(r); hostName != "" {
		logger = logger.With("hostname", hostName)
	}

	if serviceName := GetResourceServiceName(r); serviceName != "" {
		logger = logger.With("service_name", serviceName)
	}

	return logger
}

// --8<-- [end:new-logger-for-resource]
