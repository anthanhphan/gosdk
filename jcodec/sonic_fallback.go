// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

//go:build !amd64 && !386

package jcodec

// This file provides a fallback for non-AMD64 architectures.
// On non-AMD64 architectures, newSonicEngine will not be available due to build tags,
// so we provide a stub that falls back to goccy engine.
// This should never be called in normal operation since newEngineForArch
// routes non-AMD64 architectures to goccy, but it's here for safety.

// newSonicEngine is a fallback stub for non-AMD64 architectures.
// It returns a goccy engine as a safe fallback.
// This should never be reached in normal operation.
func newSonicEngine() engine {
	// Fallback to goccy engine for non-AMD64 architectures
	// This ensures the code compiles on all platforms
	return newGoccyEngine()
}
