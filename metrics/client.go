// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

// Package metrics provides a unified interface for application metrics collection
// with Prometheus as the default backend. It offers a simplified API for tracking
// counters, gauges, and histograms with support for labeled dimensions (tags).
//
// The package supports three main metric types:
//   - Counters: For monotonically increasing values (e.g., request counts)
//   - Gauges: For values that can go up or down (e.g., active connections)
//   - Histograms: For measuring distributions (e.g., request latencies)
//
// Tags are passed as alternating key-value strings to add dimensions to metrics.
//
// Example:
//
//	client := metrics.NewClient("myapp")
//	client.Inc(ctx, "requests_total", "method", "GET", "status", "200")
//	client.Duration(ctx, "request_duration", startTime, "endpoint", "/users")
//	client.SetGauge(ctx, "active_connections", 42, "service", "api")
package metrics

import (
	"context"
	"net/http"
	"time"
)

//go:generate mockgen --source=client.go -destination=./mocks/client.go -package=mocks

// ============================================================================
// Client Interface
// ============================================================================

// Client is the main interface for collecting metrics.
// It combines Counter, Gauge, and Histogram operations in a simplified interface.
// Tags are passed as alternating key-value strings (e.g., "method", "GET", "status", "200").
type Client interface {
	// Inc increments a counter by 1
	Inc(ctx context.Context, name string, tags ...string)

	// Add adds the given value to a counter
	Add(ctx context.Context, name string, value int64, tags ...string)

	// SetGauge sets a gauge to a specific value
	SetGauge(ctx context.Context, name string, value float64, tags ...string)

	// GaugeInc increments a gauge by 1
	GaugeInc(ctx context.Context, name string, tags ...string)

	// GaugeDec decrements a gauge by 1
	GaugeDec(ctx context.Context, name string, tags ...string)

	// Histogram records a value observation
	Histogram(ctx context.Context, name string, value float64, tags ...string)

	// Duration records the duration since start time
	Duration(ctx context.Context, name string, start time.Time, tags ...string)

	// Handler returns an HTTP handler for exposing metrics
	Handler() http.Handler

	// Close performs any cleanup needed by the metrics client.
	// For Prometheus, this is a no-op. For other backends, it may flush buffers.
	Close() error
}

// ============================================================================
// Default Histogram Buckets
// ============================================================================

// defaultDurationBuckets are the default buckets for duration metrics (in seconds)
var defaultDurationBuckets = []float64{
	0.001, // 1ms
	0.005, // 5ms
	0.01,  // 10ms
	0.025, // 25ms
	0.05,  // 50ms
	0.1,   // 100ms
	0.25,  // 250ms
	0.5,   // 500ms
	1.0,   // 1s
	2.5,   // 2.5s
	5.0,   // 5s
	10.0,  // 10s
}

// DefaultDurationBuckets returns the default buckets for duration metrics (in seconds).
// A new slice is returned each call so callers cannot mutate the defaults.
func DefaultDurationBuckets() []float64 {
	out := make([]float64, len(defaultDurationBuckets))
	copy(out, defaultDurationBuckets)
	return out
}
