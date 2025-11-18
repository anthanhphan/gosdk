// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
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
	return goccyjson.Marshal(v)
}

// Unmarshal converts JSON bytes to a Go value using goccy/go-json's optimized engine.
func (e *goccyEngine) Unmarshal(data []byte, v interface{}) error {
	return goccyjson.Unmarshal(data, v)
}
