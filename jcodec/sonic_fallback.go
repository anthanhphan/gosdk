// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

//go:build !amd64 && !386

package jcodec

// This file provides a fallback for non-AMD64 architectures.
// On non-AMD64 architectures, newSonicEngine will not be available due to build tags,
// so we provide a stub that falls back to goccy engine.
// This should never be called in normal operation since newEngineForArch
// routes non-AMD64 architectures to goccy, but it's here for safety.

// newSonicEngine is a fallback stub for non-AMD64 architectures.
// This should never be called due to build tag routing in newEngineForArch.
// If this is reached, it indicates a serious logic error.
func newSonicEngine() engine {
	// This panic indicates a serious bug in the architecture routing logic
	panic("jcodec: sonic engine requested on non-AMD64 architecture - this should never happen due to build tags")
}
