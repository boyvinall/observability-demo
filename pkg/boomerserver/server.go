// Package boomerserver implements the boomer server. It makes things go boom.
package boomerserver

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	pb "github.com/boyvinall/observability-demo/pkg/boomer"
	"github.com/boyvinall/observability-demo/pkg/natscarrier"
	"github.com/boyvinall/observability-demo/pkg/util"
)

const (
	attributeKeyName = "boomer.name"
)

// Server implements the boomer server
type Server struct {
	pb.UnimplementedBoomerServer
	tracer trace.Tracer
	foo    metric.Int64Counter
	c      Connection
}

// Connection is an interface for publishing and requesting messages.
// It is satisfied by [nats.Conn], among others.
type Connection interface {
	Publish(subject string, msg []byte) error
	Request(subject string, req []byte, timeout time.Duration) (resp *nats.Msg, err error)
	RequestMsg(msg *nats.Msg, timeout time.Duration) (resp *nats.Msg, err error)
}

// New creates a new boomer server.
// The server is registered with the provided [grpc.ServiceRegistrar].
func New(r grpc.ServiceRegistrar, c Connection) (pb.BoomerServer, error) {
	s := &Server{
		tracer: otel.Tracer("boomer-server"),
		c:      c,
	}
	pb.RegisterBoomerServer(r, s)

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

// Boom implements the [pb.BoomerServer] GRPC interface
func (s *Server) Boom(ctx context.Context, req *pb.BoomRequest) (*pb.BoomResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(attributeKeyName, req.GetName()))

	logger := util.LoggerFromContext(ctx)
	logger.Info("boom", "boomer_name", req.GetName())

	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	tc := otel.GetTextMapPropagator()
	reqMsg := nats.NewMsg("req")
	reqMsg.Data = b
	tc.Inject(ctx, natscarrier.Header(reqMsg.Header))

	msg, err := s.c.RequestMsg(reqMsg, 10*time.Second)
	if err != nil {
		return nil, err
	}
	var resp pb.BoomResponse
	err = proto.Unmarshal(msg.Data, &resp)
	if err != nil {
		return nil, err
	}

	s.foo.Add(ctx, 1)

	logger = util.LoggerFromContext(ctx)
	logger.Info("boom-child", "name", req.GetName())

	span.AddEvent("tick", trace.WithAttributes(attribute.Int("pid", 1234), attribute.String("origin", "reddit")))
	span.AddEvent("tick", trace.WithAttributes(attribute.Int("pid", 5678), attribute.String("precedes", "gen-x")))
	return &resp, nil
}
