package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/boyvinall/observability-demo/pkg/boomerserver"
	"github.com/boyvinall/observability-demo/pkg/util"
)

type serverConfig struct {
	grpc string
	prom string
	nats string
	otlp string
}

func runServer(config serverConfig) error {
	ctx := context.Background()
	g := errgroup.Group{}

	//--------------------------------------------------
	//
	//  Setup OTEL components
	//  Do this first because it registers a bunch of globals
	//
	//--------------------------------------------------

	err := util.SetupDefaultEnvironment(ctx, util.Config{
		ServiceName:    "MyBoomerServer",
		ServiceVersion: "0.0.0",
		OTLPEndpoint:   config.otlp,
		LogLevel:       slog.LevelDebug,
	})
	if err != nil {
		return fmt.Errorf("failed to setup default environment: %w", err)
	}
	g.Go(util.ServeMetrics(config.prom)) // Start the prometheus HTTP server

	//--------------------------------------------------
	//
	//  set up the app
	//
	//--------------------------------------------------

	// create the GRPC server first so that services can register themselves to it

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			util.UnaryServerInterceptor(nil),
		),
	)
	reflection.Register(grpcServer)

	// messaging

	c, err := nats.Connect(config.nats)
	if err != nil {
		return err
	}

	// create the server

	_, err = boomerserver.New(grpcServer, c)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	slog.Info("Listening", "address", config.grpc)
	lis, err := net.Listen("tcp", config.grpc)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// start the GRPC server

	g.Go(func() error {
		return grpcServer.Serve(lis)
	})

	//--------------------------------------------------
	//
	//  wait for the app to exit
	//
	//--------------------------------------------------

	return g.Wait()
}
