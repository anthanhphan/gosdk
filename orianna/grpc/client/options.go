// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/credentials"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"
)

// Config holds the configuration for the gRPC client.
type Config struct {
	// Address is the gRPC server address (e.g., "localhost:50051")
	Address string

	// ServiceName is the name of the service for metrics/tracing.
	ServiceName string

	// UseTLS enables TLS for the connection.
	UseTLS bool

	// TLSConfig holds TLS configuration.
	TLSConfig *TLSConfig

	// Timeout is the default timeout for each request.
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// EnableCompression enables gzip compression for gRPC messages.
	// Reduces bandwidth usage at low CPU cost.
	EnableCompression bool
}

// TLSConfig holds TLS certificate paths for mutual TLS authentication.
type TLSConfig struct {
	// CertFile is the path to the client TLS certificate.
	CertFile string
	// KeyFile is the path to the client TLS key.
	KeyFile string
	// CAFile is the path to the CA certificate for server verification.
	CAFile string
	// ServerNameOverride is used to verify server hostname.
	ServerNameOverride string
}

// DefaultConfig returns a default gRPC configuration.
func DefaultConfig() *Config {
	return &Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

// =============================================================================
// Unified Option type
// =============================================================================

// Option configures the gRPC client.
// All options are applied during NewClient construction.
//
// Example:
//
//	c, err := client.NewClient(
//	    client.WithAddress("localhost:50051"),
//	    client.WithServiceName("order-service"),
//	    client.WithTracing(tracingClient),
//	    client.WithMetrics(metricsClient),
//	)
type Option func(*Client) error

// WithAddress sets the gRPC server address.
func WithAddress(addr string) Option {
	return func(c *Client) error {
		c.config.Address = addr
		return nil
	}
}

// WithServiceName sets the service name for metrics.
func WithServiceName(name string) Option {
	return func(c *Client) error {
		c.config.ServiceName = name
		c.serviceName = name
		return nil
	}
}

// WithTLS enables TLS with the given configuration.
func WithTLS(cfg *TLSConfig) Option {
	return func(c *Client) error {
		c.config.UseTLS = true
		c.config.TLSConfig = cfg
		return nil
	}
}

// WithTracing adds distributed tracing support.
func WithTracing(tracingClient tracing.Client) Option {
	return func(c *Client) error {
		c.tracing = tracingClient
		return nil
	}
}

// WithMetrics adds metrics support.
func WithMetrics(metricsClient metrics.Client) Option {
	return func(c *Client) error {
		c.metrics = metricsClient
		return nil
	}
}

// WithRetry adds retry support.
func WithRetry(retryCfg *resilience.RetryConfig) Option {
	return func(c *Client) error {
		c.retry = retryCfg
		return nil
	}
}

// WithCircuitBreaker adds circuit breaker support.
func WithCircuitBreaker(cbCfg *resilience.CircuitBreakerConfig) Option {
	return func(c *Client) error {
		c.circuit = resilience.NewCircuitBreaker(cbCfg)
		return nil
	}
}

// WithLogger sets a custom logger for the gRPC client.
func WithLogger(log *logger.Logger) Option {
	return func(c *Client) error {
		c.logger = log
		return nil
	}
}

// =============================================================================
// TLS Helpers
// =============================================================================

// loadTLSCredentials loads TLS credentials from certificate files.
func loadTLSCredentials(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	// Load server name override
	if cfg.ServerNameOverride != "" {
		tlsCfg.ServerName = cfg.ServerNameOverride
	}

	// Load CA certificate for server verification
	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsCfg.RootCAs = certPool
	}

	// Load client certificate for mTLS
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return credentials.NewTLS(tlsCfg), nil
}
