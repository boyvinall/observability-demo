package server

import (
	"context"

	"github.com/boyvinall/go-observability-app/pkg/boomer"
	"google.golang.org/grpc"
)

type server struct {
	boomer.UnimplementedBoomerServer
}

func New(grpcServer *grpc.Server) (boomer.BoomerServer, error) {
	s := &server{}
	boomer.RegisterBoomerServer(grpcServer, s)
	return s, nil
}

func (s *server) Boom(ctx context.Context, req *boomer.BoomRequest) (*boomer.BoomResponse, error) {
	return &boomer.BoomResponse{Message: "Boom!"}, nil
}
