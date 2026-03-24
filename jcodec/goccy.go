// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
	"io"

	goccyjson "github.com/goccy/go-json"
)

// goccyEngine implements engine using goccy/go-json for non-AMD64 architectures.
// goccy/go-json is a pure Go implementation that provides excellent performance
// on ARM64 and other architectures, typically 1.4-8x faster than the standard library.
type goccyEngine struct{}

// newGoccyEngine creates a new goccy/go-json-based engine optimized for ARM64 and other architectures.
func newGoccyEngine() engine {
	return &goccyEngine{}
}

// Marshal converts a Go value to JSON bytes using goccy/go-json.
func (*goccyEngine) Marshal(v any) ([]byte, error) {
	return goccyjson.Marshal(v)
}

// Unmarshal converts JSON bytes to a Go value using goccy/go-json.
func (*goccyEngine) Unmarshal(data []byte, v any) error {
	return goccyjson.Unmarshal(data, v)
}

// MarshalIndent converts a Go value to pretty-printed JSON bytes using goccy/go-json.
func (*goccyEngine) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return goccyjson.MarshalIndent(v, prefix, indent)
}

// Valid reports whether data is valid JSON using goccy/go-json.
func (*goccyEngine) Valid(data []byte) bool {
	return goccyjson.Valid(data)
}

// NewEncoder returns a new encoder that writes to w using goccy/go-json.
func (*goccyEngine) NewEncoder(w io.Writer) Encoder {
	return goccyjson.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r using goccy/go-json.
func (*goccyEngine) NewDecoder(r io.Reader) Decoder {
	return goccyjson.NewDecoder(r)
}
