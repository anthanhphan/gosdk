// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
	"fmt"
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

// Marshal converts a Go value to JSON bytes using goccy/go-json's optimized engine.
func (e *goccyEngine) Marshal(v interface{}) ([]byte, error) {
	data, err := goccyjson.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("goccy engine: %w", err)
	}
	return data, nil
}

// Unmarshal converts JSON bytes to a Go value using goccy/go-json's optimized engine.
func (e *goccyEngine) Unmarshal(data []byte, v interface{}) error {
	if err := goccyjson.Unmarshal(data, v); err != nil {
		return fmt.Errorf("goccy engine: %w", err)
	}
	return nil
}

// MarshalIndent converts a Go value to pretty-printed JSON bytes using goccy/go-json's optimized engine.
func (e *goccyEngine) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	data, err := goccyjson.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, fmt.Errorf("goccy engine: %w", err)
	}
	return data, nil
}

// Valid reports whether data is valid JSON using goccy/go-json's validation.
func (e *goccyEngine) Valid(data []byte) bool {
	return goccyjson.Valid(data)
}

// NewEncoder returns a new encoder that writes to w using goccy/go-json's optimized engine.
func (e *goccyEngine) NewEncoder(w io.Writer) Encoder {
	return goccyjson.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r using goccy/go-json's optimized engine.
func (e *goccyEngine) NewDecoder(r io.Reader) Decoder {
	return goccyjson.NewDecoder(r)
}
