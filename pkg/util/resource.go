package util

import (
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func NewDefaultResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	hostName := os.Getenv("HOSTNAME")
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.HostName(hostName),
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.0.0"),
		),
	)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func GetResourceServiceName(r *resource.Resource) string {
	if v, ok := r.Set().Value(semconv.ServiceNameKey); ok {
		return v.AsString()
	}
	return ""
}

func GetResourceHostName(r *resource.Resource) string {
	if v, ok := r.Set().Value(semconv.HostNameKey); ok {
		return v.AsString()
	}
	return ""
}
