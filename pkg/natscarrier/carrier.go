package natscarrier

import (
	"strings"

	"github.com/nats-io/nats.go"
)

type Header nats.Header

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

func (h Header) Set(key string, value string) {
	h[key] = []string{value}
}

func (h Header) Keys() []string {
	keys := make([]string, len(h))
	i := 0
	for k := range h {
		keys[i] = k
		i++
	}
	return keys
}

func (h Header) String() string {
	s := make([]string, 0, len(h))
	for k, v := range h {
		s = append(s, k+"="+strings.Join(v, " "))
	}
	return strings.Join(s, " ")
}
