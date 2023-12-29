package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"

	"github.com/boyvinall/go-observability-app/pkg/util"
	"github.com/boyvinall/go-observability-app/pkg/worker"
)

type workerConfig struct {
	prom string
	nats string
	otlp string
}

func runWorker(config workerConfig) error {
	ctx := context.Background()
	g := errgroup.Group{}

	// Setup OTEL components
	// Do this first because it registers a bunch of globals

	err := util.SetupDefaultEnvironment(ctx, util.Config{
		ServiceName:    "MyBoomerWorker",
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

	// messaging
	c, err := nats.Connect(config.nats)
	if err != nil {
		return err
	}

	// create the worker
	_, err = worker.New(c)
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	//--------------------------------------------------
	//
	//  wait for the app to exit
	//
	//--------------------------------------------------

	return g.Wait()
}
