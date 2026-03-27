// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"crypto/tls"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"
)

// Option configures the HTTP client.
// All options are applied during NewClient construction.
//
// Example:
//
//	c, err := client.NewClient(
//	    client.WithBaseURL("https://api.example.com"),
//	    client.WithTimeout(10 * time.Second),
//	    client.WithTracing(tracingClient),
//	    client.WithRetry(client.DefaultRetryConfig()),
//	)
type Option func(*Client) error

// =============================================================================
// Config Options
// =============================================================================

// WithBaseURL sets the base URL for the client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		c.config.BaseURL = baseURL
		c.baseURL = baseURL
		return nil
	}
}

// WithTimeout sets the default request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		c.config.Timeout = timeout
		return nil
	}
}

// WithTLSConfig sets the TLS configuration.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *Client) error {
		c.config.TLSConfig = tlsConfig
		return nil
	}
}

// WithDefaultHeader adds a default header to all requests.
func WithDefaultHeader(key, value string) Option {
	return func(c *Client) error {
		c.config.Headers[key] = value
		c.headers[key] = value
		return nil
	}
}

// WithServiceName sets the service name used as a prefix for metric names.
// This aligns HTTP client metrics with gRPC client conventions.
// When set, metrics are named: {serviceName}_http_requests_total, etc.
// When empty (default), metrics are named: http_requests_total, etc.
func WithServiceName(name string) Option {
	return func(c *Client) error {
		c.serviceName = name
		return nil
	}
}

// =============================================================================
// Observability Options
// =============================================================================

// WithTracing adds distributed tracing support to the client.
func WithTracing(tracingClient tracing.Client) Option {
	return func(c *Client) error {
		c.tracing = tracingClient
		return nil
	}
}

// WithMetrics adds metrics support to the client.
func WithMetrics(metricsClient metrics.Client) Option {
	return func(c *Client) error {
		c.metrics = metricsClient
		return nil
	}
}

// WithLogger sets a custom logger.
func WithLogger(log *logger.Logger) Option {
	return func(c *Client) error {
		c.logger = log
		return nil
	}
}

// =============================================================================
// Resilience Options
// =============================================================================

// WithRetry configures retry behavior with the given config.
func WithRetry(retryCfg *resilience.RetryConfig) Option {
	return func(c *Client) error {
		c.retry = retryCfg
		return nil
	}
}

// WithCircuitBreaker configures circuit breaker behavior.
func WithCircuitBreaker(cbCfg *resilience.CircuitBreakerConfig) Option {
	return func(c *Client) error {
		c.circuit = resilience.NewCircuitBreaker(cbCfg)
		return nil
	}
}
