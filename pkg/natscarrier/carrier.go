// Package natscarrier provides an OTEL TextMapCarrierimplementation for NATS messages
package natscarrier

import (
	"strings"

	"github.com/nats-io/nats.go"
)

// Header is a TextMapCarrier that uses a nats.Header held in memory as a storage
type Header nats.Header

// Get implements the TextMapCarrier interface
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

// Set implements the TextMapCarrier interface
func (h Header) Set(key string, value string) {
	h[key] = []string{value}
}

// Keys implements the TextMapCarrier interface
func (h Header) Keys() []string {
	keys := make([]string, len(h))
	i := 0
	for k := range h {
		keys[i] = k
		i++
	}
	return keys
}

// String implements the fmt.Stringer interface
func (h Header) String() string {
	s := make([]string, 0, len(h))
	for k, v := range h {
		s = append(s, k+"="+strings.Join(v, " "))
	}
	return strings.Join(s, " ")
}
