// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

//go:build amd64 || 386

package jcodec

import (
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
	return sonic.Marshal(v)
}

// Unmarshal converts JSON bytes to a Go value using Sonic's high-performance engine.
func (e *sonicEngine) Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}
