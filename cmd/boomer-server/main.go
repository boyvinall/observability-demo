// Package main is the entry point for the server binary
package main

import (
	"log/slog"
	"net"
	"os"

	cli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/boyvinall/go-observability-app/pkg/server"
)

func run(address string) error {
	// create the GRPC server first so that services can register themselves to it
	grpcServer := grpc.NewServer(
		// grpc.Creds(credentials.NewServerTLSFromCert(&c.cert)),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		// grpc.ChainStreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	_, err := server.New(grpcServer)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		return err
	}

	slog.Info("Listening", "address", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return err
	}

	reflection.Register(grpcServer)
	return grpcServer.Serve(lis)
}

func main() {
	app := &cli.App{
		Name:  "boom",
		Usage: "make an explosive entrance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "address",
				Usage: "address to listen on",
				Value: "0.0.0.0:8080",
			},
		},
		Action: func(c *cli.Context) error {
			return run(c.String("address"))
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("unable to run app", "error", err)
		os.Exit(1)
	}
}
