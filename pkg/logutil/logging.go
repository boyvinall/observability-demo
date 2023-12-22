// Package logutil has some helpers for logging
//
// It's currently useful for setting up a logger with trace context but
// might be expanded later.
package logutil

import (
	"context"
	"log/slog"

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
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFromContext returns the logger from the context
func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		logger = logger.With("trace_id", spanContext.TraceID().String())
	}
	if spanContext.HasSpanID() {
		logger = logger.With("span_id", spanContext.SpanID().String())
	}
	return logger
}
