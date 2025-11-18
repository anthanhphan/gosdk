// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

// Package jcodec provides an intelligent, architecture-aware JSON marshaling solution
// that automatically selects the optimal JSON library based on the target CPU architecture.
// It seamlessly switches between Sonic (for AMD64/x86_64) and goccy/go-json (for ARM64/others)
// to deliver peak performance on any platform.
package jcodec

import (
	"runtime"
	"sync"
)

const (
	// Architecture constants for engine selection.
	archAMD64 = "amd64"
	arch386   = "386"
)

// engine represents a JSON marshaling engine implementation.
type engine interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

var (
	// defaultEngine is the global default engine instance, lazily initialized.
	defaultEngine engine
	once          sync.Once
)

// getDefaultEngine returns the global default engine instance.
// It is initialized on first use using auto-selection strategy.
func getDefaultEngine() engine {
	once.Do(func() {
		defaultEngine = newEngineForArch(runtime.GOARCH)
	})
	return defaultEngine
}

// newEngineForArch creates an engine optimized for a specific architecture.
func newEngineForArch(arch string) engine {
	switch arch {
	case archAMD64, arch386:
		return newSonicEngine()
	default:
		return newGoccyEngine()
	}
}

// Marshal converts a Go value to JSON bytes using the optimal engine for the current architecture.
// It automatically selects Sonic on AMD64/x86_64 and goccy/go-json on other architectures.
//
// Input:
//   - v: The value to marshal
//
// Output:
//   - []byte: The marshaled JSON bytes
//   - error: Any error that occurred during marshaling
//
// Example:
//
//	user := User{ID: 1, Name: "John"}
//	data, err := jcodec.Marshal(user)
//	if err != nil {
//	    return fmt.Errorf("marshal failed: %w", err)
//	}
func Marshal(v interface{}) ([]byte, error) {
	return getDefaultEngine().Marshal(v)
}

// Unmarshal converts JSON bytes to a Go value using the optimal engine for the current architecture.
// It automatically selects Sonic on AMD64/x86_64 and goccy/go-json on other architectures.
//
// Input:
//   - data: The JSON bytes to unmarshal
//   - v: A pointer to the value to unmarshal into
//
// Output:
//   - error: Any error that occurred during unmarshaling
//
// Example:
//
//	var user User
//	err := jcodec.Unmarshal(data, &user)
//	if err != nil {
//	    return fmt.Errorf("unmarshal failed: %w", err)
//	}
func Unmarshal(data []byte, v interface{}) error {
	return getDefaultEngine().Unmarshal(data, v)
}
