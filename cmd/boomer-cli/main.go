package main

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/boyvinall/go-observability-app/pkg/boomer"
)

func main() {
	ctx := context.Background()
	address := "localhost:8080"
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure())
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}
	defer conn.Close()

	client := boomer.NewBoomerClient(conn)
	resp, err := client.Boom(ctx, &boomer.BoomRequest{})
	if err != nil {
		slog.Error("failed to boom", "error", err)
		return
	}
	slog.Info("response", "message", resp.Message)
}
