// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package errors provides shared sentinel errors used across
// both HTTP and gRPC server implementations.
package errors

import "errors"

// Shared sentinel errors for server lifecycle and configuration.
var (
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrHandlerNil       = errors.New("handler cannot be nil")
	ErrServerNotStarted = errors.New("server not started")
	ErrServerShutdown   = errors.New("server shutdown")
	ErrNilChecker       = errors.New("health checker cannot be nil")
	ErrCircuitOpen      = errors.New("service temporarily unavailable")
)

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

// IsServerError checks if an error is a server-related error.
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerNotStarted) ||
		errors.Is(err, ErrServerShutdown)
}
