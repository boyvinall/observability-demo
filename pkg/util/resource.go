package util

import (
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// NewDefaultResource creates an OTEL resource with a few useful attributes
func NewDefaultResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	hostName := os.Getenv("HOSTNAME")
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.HostName(hostName),
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)

	if err != nil {
		return nil, err
	}

	return r, nil
}

// GetResourceServiceName returns the service name from the resource
func GetResourceServiceName(r *resource.Resource) string {
	if v, ok := r.Set().Value(semconv.ServiceNameKey); ok {
		return v.AsString()
	}
	return ""
}

// GetResourceHostName returns the hostname from the resource
func GetResourceHostName(r *resource.Resource) string {
	if v, ok := r.Set().Value(semconv.HostNameKey); ok {
		return v.AsString()
	}
	return ""
}
