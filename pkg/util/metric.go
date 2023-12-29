package util

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewMeterProviderForResource(r *resource.Resource) (*metric.MeterProvider, error) {
	// Create the prometheus exporter
	// For an example with more config options, see https://docs.daocloud.io/en/insight/06UserGuide/01quickstart/otel/meter/#create-an-initialization-function-using-the-opentelemetry-sdk
	metricExp, err := otelprom.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prometheus exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metricExp),
		metric.WithResource(r),
	)

	return mp, nil
}

func ServeMetrics(address string) func() error {
	return func() error {
		slog.Info("serving metrics", "address", address)

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{EnableOpenMetrics: true}))
		metricServer := &http.Server{
			Addr:              address,
			ReadHeaderTimeout: 3 * time.Second, // fix for gosec G114
			Handler:           mux,
		}
		return metricServer.ListenAndServe()
	}
}
