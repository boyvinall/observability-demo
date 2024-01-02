// Package worker implements a message-based worker for the boomer service.
package worker

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	pb "github.com/boyvinall/observability-demo/pkg/boomer"
	"github.com/boyvinall/observability-demo/pkg/natscarrier"
	"github.com/boyvinall/observability-demo/pkg/util"
)

// Connection is an interface for subscribing to messages
type Connection interface {
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
}

// Worker processes and responds to requests from a message queue
type Worker struct {
	tracer trace.Tracer
	sub    *nats.Subscription
}

// New creates a new boomer worker
func New(c Connection) (*Worker, error) {
	w := &Worker{
		tracer: otel.Tracer("boomer-worker"),
	}

	var err error
	w.sub, err = c.Subscribe("req", w.Handler)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Handler processes and responds to the [nats.Msg].
func (w *Worker) Handler(msg *nats.Msg) {
	tc := otel.GetTextMapPropagator()
	ctx := tc.Extract(context.Background(), natscarrier.Header(msg.Header))

	l := util.LoggerFromContext(ctx)
	l.Info("received request",
		"subject", msg.Subject,
		"reply", msg.Reply,
	)

	_, span := w.tracer.Start(ctx, "handler")
	defer span.End()

	var req pb.BoomRequest
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		_ = msg.Nak()
		return
	}

	resp := &pb.BoomResponse{Message: "Boom!"}
	b, err := proto.Marshal(resp)
	if err != nil {
		_ = msg.Nak()
		return
	}

	err = msg.Respond(b)
	if err != nil {
		_ = msg.Nak()
		return
	}
}
