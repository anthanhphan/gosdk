// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

//go:build amd64 || 386

package jcodec

import (
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

// Marshal converts a Go value to JSON bytes using Sonic.
func (*sonicEngine) Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal converts JSON bytes to a Go value using Sonic.
func (*sonicEngine) Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

// MarshalIndent converts a Go value to pretty-printed JSON bytes using Sonic.
func (*sonicEngine) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return sonic.MarshalIndent(v, prefix, indent)
}

// Valid reports whether data is valid JSON using Sonic.
func (*sonicEngine) Valid(data []byte) bool {
	return sonic.Valid(data)
}

// NewEncoder returns a new encoder that writes to w using Sonic.
func (*sonicEngine) NewEncoder(w io.Writer) Encoder {
	return sonic.ConfigDefault.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r using Sonic.
func (*sonicEngine) NewDecoder(r io.Reader) Decoder {
	return sonic.ConfigDefault.NewDecoder(r)
}
