package worker

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	pb "github.com/boyvinall/go-observability-app/pkg/boomer"
	"github.com/boyvinall/go-observability-app/pkg/natscarrier"
	"github.com/boyvinall/go-observability-app/pkg/util"
)

type Connection interface {
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
}

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

// Handler handles requests
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
		msg.Nak()
		return
	}

	resp := &pb.BoomResponse{Message: "Boom!"}
	b, err := proto.Marshal(resp)
	if err != nil {
		msg.Nak()
		return
	}

	err = msg.Respond(b)
	if err != nil {
		msg.Nak()
		return
	}
}
