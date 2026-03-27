// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/shared/observability"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"
)

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Client is an HTTP client with built-in observability (tracing, metrics),
// retry logic, and circuit breaker support.
type Client struct {
	httpClient *http.Client
	config     *Config
	tracing    tracing.Client
	metrics    metrics.Client
	retry      *resilience.RetryConfig
	circuit    *resilience.CircuitBreaker
	logger     *logger.Logger

	baseURL     string
	headers     map[string]string
	serviceName string

	// Pre-built lookup maps for O(1) retry decisions (built at construction time)
	retryStatusSet map[int]struct{}
	retryMethodSet map[string]struct{}

	// Pre-computed metric names (built at construction time to avoid per-request alloc)
	metricRequestsTotal   string
	metricRequestDuration string
	metricInFlight        string
	statusCodeCache       map[int]string
}

// =============================================================================
// Constructor
// =============================================================================

// NewClient creates a new HTTP client with observability.
// The client uses a production-grade http.Transport with connection pooling,
// keepalive, and TLS handshake timeouts by default.
//
// Example:
//
//	c, err := client.NewClient(
//	    client.WithBaseURL("https://api.example.com"),
//	    client.WithTracing(tracingClient),
//	    client.WithRetry(client.DefaultRetryConfig()),
//	    client.WithCircuitBreaker(client.DefaultCircuitBreakerConfig()),
//	)
func NewClient(opts ...Option) (*Client, error) {
	cfg := DefaultConfig()

	c := &Client{
		config:  cfg,
		logger:  logger.NewLoggerWithFields(logger.String("package", "http-client")),
		baseURL: cfg.BaseURL,
		headers: cfg.Headers,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	// Build production-grade transport with connection pooling
	maxIdleConns := c.config.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 100
	}
	maxIdleConnsPerHost := c.config.MaxIdleConnsPerHost
	if maxIdleConnsPerHost == 0 {
		maxIdleConnsPerHost = 10
	}

	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     c.config.TLSConfig,
	}

	c.httpClient = &http.Client{
		Timeout:   c.config.Timeout,
		Transport: transport,
	}

	// Pre-build retry lookup maps for O(1) checks instead of O(n) linear scans
	if c.retry != nil {
		c.retryStatusSet = make(map[int]struct{}, len(c.retry.RetryableStatusCodes))
		for _, code := range c.retry.RetryableStatusCodes {
			c.retryStatusSet[code] = struct{}{}
		}
		c.retryMethodSet = make(map[string]struct{}, len(c.retry.RetryableMethods))
		for _, m := range c.retry.RetryableMethods {
			c.retryMethodSet[m] = struct{}{}
		}
	}

	// Pre-compute metric names at construction time (zero alloc per request).
	// Uses shared observability suffix constants for consistency with server metrics.
	metricPrefix := "http"
	if c.serviceName != "" {
		metricPrefix = c.serviceName + "_http"
	}
	c.metricRequestsTotal = metricPrefix + observability.SuffixRequestsTotal
	c.metricRequestDuration = metricPrefix + observability.SuffixRequestDurationSeconds
	c.metricInFlight = metricPrefix + observability.SuffixInFlightRequests

	// Pre-build status code cache for O(1) string conversion on hot path
	c.statusCodeCache = observability.CodeStringCache([]int{
		200, 201, 204, 301, 302, 304, 400, 401, 403, 404, 405, 409, 422, 429, 500, 502, 503, 504,
	})

	return c, nil
}

// Close releases idle transport connections.
// Call this when the client is no longer needed.
func (c *Client) Close() {
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}
