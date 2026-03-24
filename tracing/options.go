// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package tracing

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ============================================================================
// Client Options (Functional Options Pattern)
// ============================================================================

// Option configures a tracing client.
// Use the With* functions to create options.
type Option func(*clientOptions)

// clientOptions holds all configurable options for the tracing client.
type clientOptions struct {
	// serviceName is the name of the service (used in trace resource).
	serviceName string

	// serviceVersion is the version of the service.
	serviceVersion string

	// environment is the deployment environment (e.g., "production", "staging").
	environment string

	// endpoint is the OTLP collector endpoint (default: "localhost:4317").
	endpoint string

	// insecure disables TLS for the OTLP gRPC connection.
	insecure bool

	// sampler is a custom sampler. If nil, AlwaysSample is used.
	sampler sdktrace.Sampler

	// headers are additional gRPC metadata headers for the OTLP exporter.
	headers map[string]string
}

// defaultClientOptions returns the default client options.
func defaultClientOptions() *clientOptions {
	return &clientOptions{
		endpoint: "localhost:4317",
		insecure: false,
	}
}

// WithServiceVersion sets the service version in the trace resource.
// This appears as the `service.version` resource attribute.
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithServiceVersion("v1.2.0"),
//	)
func WithServiceVersion(version string) Option {
	return func(o *clientOptions) {
		o.serviceVersion = version
	}
}

// WithEnvironment sets the deployment environment as a resource attribute.
// Common values: "production", "staging", "development".
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithEnvironment("production"),
//	)
func WithEnvironment(env string) Option {
	return func(o *clientOptions) {
		o.environment = env
	}
}

// WithEndpoint sets the OTLP collector endpoint.
// Default: "localhost:4317" (gRPC).
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithEndpoint("otel-collector.monitoring:4317"),
//	)
func WithEndpoint(endpoint string) Option {
	return func(o *clientOptions) {
		if endpoint != "" {
			o.endpoint = endpoint
		}
	}
}

// WithInsecure disables TLS for the OTLP gRPC connection.
// Use this for local development or when the collector is behind a trusted network.
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithEndpoint("localhost:4317"),
//	    tracing.WithInsecure(),
//	)
func WithInsecure() Option {
	return func(o *clientOptions) {
		o.insecure = true
	}
}

// WithSampler sets a custom sampling strategy.
// If not set, AlwaysSample is used.
//
// Common samplers:
//   - sdktrace.AlwaysSample() -- sample everything (default)
//   - sdktrace.NeverSample() -- sample nothing
//   - sdktrace.TraceIDRatioBased(0.1) -- sample 10% of traces
//   - sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1)) -- respect parent decision, fallback to 10%
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5))),
//	)
func WithSampler(sampler sdktrace.Sampler) Option {
	return func(o *clientOptions) {
		o.sampler = sampler
	}
}

// WithSamplingRate is a convenience option that sets TraceIDRatioBased sampling
// wrapped in ParentBased. The rate should be between 0.0 and 1.0.
//
//   - 0.0 = never sample (unless parent says yes)
//   - 0.5 = sample 50% of root traces
//   - 1.0 = always sample
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithSamplingRate(0.1), // Sample 10% of traces
//	)
func WithSamplingRate(rate float64) Option {
	return func(o *clientOptions) {
		o.sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(rate))
	}
}

// WithHeaders sets additional gRPC metadata headers sent with each OTLP export request.
// Useful for authentication tokens or tenant identifiers.
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithHeaders(map[string]string{
//	        "Authorization": "Bearer <token>",
//	    }),
//	)
func WithHeaders(headers map[string]string) Option {
	return func(o *clientOptions) {
		o.headers = headers
	}
}
