// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

//go:build amd64 || 386

package jcodec

import (
	"fmt"
	"io"

	"github.com/bytedance/sonic"
)

// sonicEngine implements engine using Sonic for AMD64/x86_64 architectures.
// Sonic leverages JIT compilation and SIMD instructions (AVX2, SSE4.2)
// to achieve 3-10x better performance than the standard library on x86 architectures.
type sonicEngine struct{}

// newSonicEngine creates a new Sonic-based engine optimized for AMD64/x86_64.
func newSonicEngine() engine {
	return &sonicEngine{}
}

// Marshal converts a Go value to JSON bytes using Sonic's high-performance engine.
func (e *sonicEngine) Marshal(v interface{}) ([]byte, error) {
	data, err := sonic.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("sonic engine: %w", err)
	}
	return data, nil
}

// Unmarshal converts JSON bytes to a Go value using Sonic's high-performance engine.
func (e *sonicEngine) Unmarshal(data []byte, v interface{}) error {
	if err := sonic.Unmarshal(data, v); err != nil {
		return fmt.Errorf("sonic engine: %w", err)
	}
	return nil
}

// MarshalIndent converts a Go value to pretty-printed JSON bytes using Sonic's high-performance engine.
func (e *sonicEngine) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	data, err := sonic.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, fmt.Errorf("sonic engine: %w", err)
	}
	return data, nil
}

// Valid reports whether data is valid JSON using Sonic's validation.
func (e *sonicEngine) Valid(data []byte) bool {
	return sonic.Valid(data)
}

// NewEncoder returns a new encoder that writes to w using Sonic's high-performance engine.
func (e *sonicEngine) NewEncoder(w io.Writer) Encoder {
	return sonic.ConfigDefault.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r using Sonic's high-performance engine.
func (e *sonicEngine) NewDecoder(r io.Reader) Decoder {
	return sonic.ConfigDefault.NewDecoder(r)
}
