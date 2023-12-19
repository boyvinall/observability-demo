// Package main is the entry point for the CLI binary
package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/boyvinall/go-observability-app/pkg/boomer"
)

func main() {
	ctx := context.Background()
	address := "localhost:8080"
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}
	defer conn.Close()

	client := boomer.NewBoomerClient(conn)

	name := "old dude"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	resp, err := client.Boom(ctx, &boomer.BoomRequest{
		Name: name,
	})
	if err != nil {
		slog.Error("failed to boom", "error", err)
		return
	}
	slog.Info("response", "message", resp.Message)
}
