// Package server implements the boomer server. It makes things go boom.
package server

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/boyvinall/go-observability-app/pkg/boomer"
	"github.com/boyvinall/go-observability-app/pkg/tracelog"
)

const (
	attributeKeyName = "boomer.name"
)

type server struct {
	boomer.UnimplementedBoomerServer
	tracer trace.Tracer
}

// New creates a new boomer server
//
// The server is registered with the provided GRPC server
func New(grpcServer *grpc.Server) (boomer.BoomerServer, error) {
	s := &server{
		tracer: otel.Tracer("boomer-server"),
	}
	boomer.RegisterBoomerServer(grpcServer, s)
	return s, nil
}

func (s *server) Boom(ctx context.Context, req *boomer.BoomRequest) (*boomer.BoomResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(attributeKeyName, req.GetName()))

	l := tracelog.LoggerFromContext(ctx)
	l.Info("boom", "name", req.GetName())

	ctx, childSpan := s.tracer.Start(ctx, "child")
	defer childSpan.End()

	l = tracelog.LoggerFromContext(ctx)
	l.Info("boom", "name", req.GetName())

	childSpan.AddEvent("tick", trace.WithAttributes(attribute.Int("pid", 1234), attribute.String("origin", "reddit")))
	childSpan.AddEvent("tick", trace.WithAttributes(attribute.Int("pid", 5678), attribute.String("precedes", "gen-x")))
	return &boomer.BoomResponse{Message: "Boom!"}, nil
}
