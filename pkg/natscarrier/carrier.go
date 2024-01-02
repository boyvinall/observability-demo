// Package natscarrier provides an OTEL TextMapCarrier implementation for NATS messages.
//
// The carrier can be used to inject and extract trace/span IDs from a [nats.Msg] via
// [go.opentelemetry.io/otel/propagation.TextMapPropagator].
package natscarrier

import (
	"strings"

	"github.com/nats-io/nats.go"
)

// Header implements the [go.opentelemetry.io/otel/propagation.TextMapCarrier] interface
// using a [nats.Header] held in memory as a storage
type Header nats.Header

// Get returns the value associated with the passed key
func (h Header) Get(key string) string {
	v, found := h[key]
	if !found {
		return ""
	}
	if len(v) == 0 {
		return ""
	}
	return h[key][0]
}

// Set stores the key-value pair
func (h Header) Set(key string, value string) {
	h[key] = []string{value}
}

// Keys lists the keys stored in this carrier
func (h Header) Keys() []string {
	keys := make([]string, len(h))
	i := 0
	for k := range h {
		keys[i] = k
		i++
	}
	return keys
}

// String implements the [fmt.Stringer] interface, see [Header]
func (h Header) String() string {
	s := make([]string, 0, len(h))
	for k, v := range h {
		s = append(s, k+"="+strings.Join(v, " "))
	}
	return strings.Join(s, " ")
}
