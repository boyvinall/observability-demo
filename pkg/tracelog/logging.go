package tracelog

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type loggerKey struct{}

func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = context.WithValue(ctx, loggerKey{}, logger)
		return handler(ctx, req)
	}
}

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
