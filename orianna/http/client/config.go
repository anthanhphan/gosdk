// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"crypto/tls"
	"time"
)

// DefaultMaxResponseBodySize is the maximum response body size (10MB).
// Prevents OOM from malicious or buggy upstream servers.
const DefaultMaxResponseBodySize = 10 * 1024 * 1024

// Config holds the configuration for the HTTP client.
type Config struct {
	// BaseURL is the base URL for all requests (e.g., "https://api.example.com")
	BaseURL string

	// Timeout is the default timeout for each request.
	// Includes connection, response, and retry timeouts.
	Timeout time.Duration

	// TLSConfig holds TLS configuration for HTTPS connections.
	TLSConfig *tls.Config

	// Headers are default headers applied to all requests.
	Headers map[string]string

	// MaxIdleConns controls the maximum number of idle (keep-alive) connections
	// across all hosts. Default: 100.
	MaxIdleConns int

	// MaxIdleConnsPerHost controls the maximum idle connections per host.
	// Default: 10.
	MaxIdleConnsPerHost int

	// MaxResponseBodySize is the maximum allowed response body size in bytes.
	// Prevents OOM from malicious or buggy upstream servers.
	// Default: 10MB (10485760 bytes). Set 0 to use default.
	MaxResponseBodySize int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Timeout: 30 * time.Second,
		Headers: make(map[string]string),
	}
}
