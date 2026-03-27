// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip" // Register gzip compressor
	"google.golang.org/grpc/keepalive"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/shared/observability"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"
)

// Compile-time interface compliance checks.
var _ io.Closer = (*Client)(nil)

// Client is a gRPC client with built-in observability,
// retry logic, and circuit breaker support.
type Client struct {
	conn        *grpc.ClientConn
	config      *Config
	tracing     tracing.Client
	metrics     metrics.Client
	retry       *resilience.RetryConfig
	circuit     *resilience.CircuitBreaker
	logger      *logger.Logger
	serviceName string

	// Pre-built lookup map for O(1) retry decisions (built at construction time)
	retryStatusSet map[int]struct{}

	// Pre-computed metric names (built at construction time to avoid per-request alloc)
	metricRequestsTotal   string
	metricRequestDuration string
	metricInFlight        string
}

// =============================================================================
// Constructor
// =============================================================================

// NewClient creates a new gRPC client with observability.
//
// Example:
//
//	c, err := client.NewClient(
//	    client.WithAddress("localhost:50051"),
//	    client.WithServiceName("order-service"),
//	    client.WithTracing(tracingClient),
//	)
func NewClient(opts ...Option) (*Client, error) {
	cfg := DefaultConfig()

	c := &Client{
		config: cfg,
		logger: logger.NewLoggerWithFields(logger.String("package", "grpc-client")),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply gRPC option: %w", err)
		}
	}

	if c.config.Address == "" {
		return nil, fmt.Errorf("gRPC address is required")
	}

	c.serviceName = c.config.ServiceName

	// Pre-build retry lookup map for O(1) checks instead of O(n) linear scans
	if c.retry != nil {
		c.retryStatusSet = make(map[int]struct{}, len(c.retry.RetryableStatusCodes))
		for _, code := range c.retry.RetryableStatusCodes {
			c.retryStatusSet[code] = struct{}{}
		}
	}

	// Pre-compute metric names at construction time (zero alloc per request)
	c.metricRequestsTotal = c.serviceName + "_grpc" + observability.SuffixRequestsTotal
	c.metricRequestDuration = c.serviceName + "_grpc" + observability.SuffixRequestDurationSeconds
	c.metricInFlight = c.serviceName + "_grpc" + observability.SuffixInFlightRequests

	// Build dial options
	var dialOpts []grpc.DialOption

	// Transport credentials
	if c.config.UseTLS && c.config.TLSConfig != nil {
		creds, err := loadTLSCredentials(c.config.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Keepalive — prevents idle connections from being killed by load balancers
	// and detects dead connections early.
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: false,
	}))

	// Default call options — message size limits
	defaultCallOpts := []grpc.CallOption{
		grpc.MaxCallRecvMsgSize(4 * 1024 * 1024),
		grpc.MaxCallSendMsgSize(4 * 1024 * 1024),
	}
	if c.config.EnableCompression {
		defaultCallOpts = append(defaultCallOpts, grpc.UseCompressor("gzip"))
	}
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(defaultCallOpts...))

	// Unary interceptor for observability
	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(
		newUnaryClientInterceptor(c.serviceName),
	))

	// Stream interceptor for observability
	dialOpts = append(dialOpts, grpc.WithStreamInterceptor(
		newStreamClientInterceptor(c.serviceName),
	))

	conn, err := grpc.NewClient(c.config.Address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	c.conn = conn
	return c, nil
}

// =============================================================================
// Client Methods
// =============================================================================

// Connection returns the underlying gRPC connection.
func (c *Client) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
