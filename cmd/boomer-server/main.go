// Package main is the entry point for the server binary
package main

import (
	"log/slog"
	"os"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "boomer",
		Usage: "make an explosive entrance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "otlp",
				Usage: "OTLP endpoint",
				Value: "tempo:4317",
			},
			&cli.StringFlag{
				Name:  "nats",
				Usage: "NATS endpoint",
				Value: "nats://nats:4222",
			},
			&cli.StringFlag{
				Name:  "listen-metrics",
				Usage: "listen address for prometheus metrics endpoint",
				Value: "0.0.0.0:2223",
			},
		},
		Commands: []*cli.Command{
			//--------------------------------------------------
			//  GRPC server
			//--------------------------------------------------
			{
				Name:  "server",
				Usage: "run the GRPC server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "listen-grpc",
						Usage: "listen address for GRPC server",
						Value: "0.0.0.0:8080",
					},
				},
				Action: func(c *cli.Context) error {
					return runServer(serverConfig{
						grpc: c.String("listen-grpc"),
						prom: c.String("listen-metrics"),
						nats: c.String("nats"),
						otlp: c.String("otlp"),
					})
				},
			},
			//--------------------------------------------------
			//  Worker
			//--------------------------------------------------
			{
				Name:  "worker",
				Usage: "run the NATS worker",
				Flags: []cli.Flag{},
				Action: func(c *cli.Context) error {
					return runWorker(workerConfig{
						prom: c.String("listen-metrics"),
						nats: c.String("nats"),
						otlp: c.String("otlp"),
					})
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("unable to run app", "error", err)
		os.Exit(1)
	}
}
