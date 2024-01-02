package main

import (
	"log/slog"

	"github.com/cenkalti/backoff/v4"
	"github.com/nats-io/nats.go"
)

func setupNatsConnection(address string) (*nats.Conn, error) {
	var c *nats.Conn
	b := backoff.NewExponentialBackOff()

	err := backoff.Retry(func() error {
		var e error
		slog.Info("Connecting to NATS", "address", address)
		c, e = nats.Connect(address)
		return e
	}, b)

	return c, err
}
