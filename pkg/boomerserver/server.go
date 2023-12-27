// Package boomerserver implements the boomer server. It makes things go boom.
package boomerserver

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	pb "github.com/boyvinall/go-observability-app/pkg/boomer"
	"github.com/boyvinall/go-observability-app/pkg/logutil"
)

const (
	attributeKeyName = "boomer.name"
)

type server struct {
	pb.UnimplementedBoomerServer
	tracer trace.Tracer
	foo    metric.Int64Counter
}

// New creates a new boomer server
//
// The server is registered with the provided GRPC server
func New(grpcServer *grpc.Server) (pb.BoomerServer, error) {
	s := &server{
		tracer: otel.Tracer("boomer-server"),
	}
	pb.RegisterBoomerServer(grpcServer, s)

	m := otel.GetMeterProvider().Meter("app_or_package_name")

	var err error
	s.foo, err = m.Int64Counter("foo",
		metric.WithDescription("fooo"),
		metric.WithUnit("ms"))
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *server) Boom(ctx context.Context, req *pb.BoomRequest) (*pb.BoomResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(attributeKeyName, req.GetName()))

	logger := logutil.LoggerFromContext(ctx)
	logger.Info("boom", "name", req.GetName())

	ctx, childSpan := s.tracer.Start(ctx, "child")
	defer childSpan.End()

	s.foo.Add(ctx, 1)

	logger = logutil.LoggerFromContext(ctx)
	logger.Info("boom-child", "name", req.GetName())

	childSpan.AddEvent("tick", trace.WithAttributes(attribute.Int("pid", 1234), attribute.String("origin", "reddit")))
	childSpan.AddEvent("tick", trace.WithAttributes(attribute.Int("pid", 5678), attribute.String("precedes", "gen-x")))
	return &pb.BoomResponse{Message: "Boom!"}, nil
}
