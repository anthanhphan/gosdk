// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package configuration

import "time"

// Default server configuration values
const (
	DefaultPort                     = 8080
	DefaultMaxBodySize              = 4 * 1024 * 1024 // 4MB
	DefaultMaxConcurrentConnections = 256 * 1024
	DefaultReadTimeout              = 30 * time.Second
	DefaultWriteTimeout             = 30 * time.Second
	DefaultIdleTimeout              = 120 * time.Second
	DefaultGracefulShutdownTimeout  = 30 * time.Second
	DefaultHealthCheckTimeout       = 5 * time.Second
	DefaultRequestTimeout           = 30 * time.Second
)

// Rate limiter defaults
const (
	DefaultRateLimitMax        = 500
	DefaultRateLimitWindow     = 1 * time.Minute
	DefaultRateLimitExpiration = 1 * time.Minute
)

// Health check defaults
const (
	DefaultHealthCheckerName         = "http"
	HealthStatusThresholdClientError = 400
	HealthStatusThresholdServerError = 500
)

// Logging defaults
const (
	DefaultLogCapacityBase            = 8
	DefaultLogCapacityVerbose         = 14
	DefaultLogCapacityResponseBase    = 8
	DefaultLogCapacityResponseVerbose = 10
)

// Compression defaults
const (
	DefaultCompressionLevel = 1 // compress.LevelBestSpeed
)

// CSRF defaults
const (
	DefaultCSRFCookieName = "csrf_token"
	DefaultCSRFCookiePath = "/"
	DefaultCSRFKeyLookup  = "header:X-CSRF-Token"
	DefaultCSRFSameSite   = "Lax"
	DefaultCSRFExpiration = 24 * time.Hour
)
