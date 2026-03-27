// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package resilience provides shared resilience patterns (retry, circuit breaker)
// used across both HTTP and gRPC client implementations.
package resilience

import (
	"time"
)

// RetryConfig holds retry configuration.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int

	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// Multiplier is the backoff multiplier for each retry.
	Multiplier float64

	// RetryableStatusCodes are status codes that should trigger a retry.
	// For HTTP: status codes (429, 503, etc.). For gRPC: gRPC code integers.
	RetryableStatusCodes []int

	// RetryableMethods are methods that are safe to retry.
	// For HTTP: method strings ("GET", "HEAD"). For gRPC: full method names.
	// If empty, all methods are retryable.
	RetryableMethods []string
}

// DefaultRetryConfig returns a default retry configuration suitable for HTTP.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:          3,
		InitialBackoff:       100 * time.Millisecond,
		MaxBackoff:           5 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
		RetryableMethods:     []string{"GET", "HEAD", "OPTIONS", "DELETE"},
	}
}

// RetryOption configures retry behavior.
type RetryOption func(*RetryConfig)

// WithMaxAttempts sets the maximum retry attempts.
func WithMaxAttempts(max int) RetryOption {
	return func(c *RetryConfig) {
		c.MaxAttempts = max
	}
}

// WithBackoff sets the initial and max backoff durations.
func WithBackoff(initial, max time.Duration) RetryOption {
	return func(c *RetryConfig) {
		c.InitialBackoff = initial
		c.MaxBackoff = max
	}
}
