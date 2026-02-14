// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package metrics

import "github.com/prometheus/client_golang/prometheus"

// ============================================================================
// Client Options (Functional Options Pattern)
// ============================================================================

// Option configures a metrics client.
// Use the With* functions to create options.
type Option func(*clientOptions)

// clientOptions holds all configurable options for the metrics client.
type clientOptions struct {
	// buckets overrides the default histogram buckets
	buckets []float64

	// constLabels are labels that are applied to every metric
	constLabels prometheus.Labels

	// subsystem is an optional subsystem name added between namespace and metric name
	subsystem string

	// enableGoCollector enables the Go runtime collector (default: true)
	enableGoCollector bool

	// enableProcessCollector enables the process collector (default: true)
	enableProcessCollector bool
}

// defaultClientOptions returns the default client options.
func defaultClientOptions() *clientOptions {
	return &clientOptions{
		buckets:                DefaultDurationBuckets(),
		enableGoCollector:      true,
		enableProcessCollector: true,
	}
}

// WithBuckets sets custom histogram buckets for duration and histogram metrics.
// If not set, DefaultDurationBuckets will be used.
//
// Example:
//
//	client := metrics.NewClient("myapp",
//	    metrics.WithBuckets([]float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0}),
//	)
func WithBuckets(buckets []float64) Option {
	return func(o *clientOptions) {
		if len(buckets) > 0 {
			o.buckets = buckets
		}
	}
}

// WithConstLabels sets constant labels that are applied to every metric.
// These labels cannot be changed after creation. Use for service-level identifiers
// like environment, region, or pod name.
//
// Example:
//
//	client := metrics.NewClient("myapp",
//	    metrics.WithConstLabels(map[string]string{
//	        "env":    "production",
//	        "region": "us-east-1",
//	    }),
//	)
func WithConstLabels(labels map[string]string) Option {
	return func(o *clientOptions) {
		o.constLabels = labels
	}
}

// WithSubsystem sets an optional subsystem name that is inserted between
// the namespace and metric name. Useful for grouping related metrics.
//
// Example:
//
//	// Metrics will be named: myapp_http_requests_total
//	client := metrics.NewClient("myapp",
//	    metrics.WithSubsystem("http"),
//	)
func WithSubsystem(subsystem string) Option {
	return func(o *clientOptions) {
		o.subsystem = subsystem
	}
}

// WithoutGoCollector disables the Go runtime metrics collector.
// Useful in testing or when you want to reduce metric cardinality.
func WithoutGoCollector() Option {
	return func(o *clientOptions) {
		o.enableGoCollector = false
	}
}

// WithoutProcessCollector disables the process metrics collector.
// Useful in testing or when you want to reduce metric cardinality.
func WithoutProcessCollector() Option {
	return func(o *clientOptions) {
		o.enableProcessCollector = false
	}
}
