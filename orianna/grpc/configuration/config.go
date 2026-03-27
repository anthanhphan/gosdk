// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package configuration

import (
	"errors"
	"fmt"
	"time"
)

// Validate checks the gRPC configuration for common mistakes and returns an error
// if anything is misconfigured. Call this at startup for fail-fast behavior.
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return errors.New("service_name is required")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port must be 0-65535, got %d", c.Port)
	}
	if err := c.validateTLS(); err != nil {
		return err
	}
	if err := c.validateTimeouts(); err != nil {
		return err
	}
	if c.MaxRecvMsgSize < 0 {
		return errors.New("max_recv_msg_size must be non-negative")
	}
	if c.MaxSendMsgSize < 0 {
		return errors.New("max_send_msg_size must be non-negative")
	}
	return nil
}

// validateTLS checks TLS/mTLS configuration completeness and exclusivity.
func (c *Config) validateTLS() error {
	if c.TLS != nil && c.MTLS != nil {
		return errors.New("TLS and MTLS cannot both be set; use TLS for server-only TLS or MTLS for mutual TLS")
	}
	if c.TLS != nil && (c.TLS.CertFile == "" || c.TLS.KeyFile == "") {
		return errors.New("TLS requires both cert_file and key_file")
	}
	if c.MTLS != nil {
		if c.MTLS.CertFile == "" || c.MTLS.KeyFile == "" {
			return errors.New("MTLS requires both cert_file and key_file")
		}
		if c.MTLS.CAFile == "" {
			return errors.New("MTLS requires ca_file for client certificate verification")
		}
	}
	if len(c.Permissions) > 0 && c.MTLS == nil {
		return errors.New("permissions require MTLS to be configured (certificate-based authorization needs client certificates)")
	}
	return nil
}

// validateTimeouts checks that all timeout durations are positive when set.
func (c *Config) validateTimeouts() error {
	checks := []struct {
		name  string
		value *time.Duration
	}{
		{"connection_timeout", c.ConnectionTimeout},
		{"keepalive_time", c.KeepaliveTime},
		{"keepalive_timeout", c.KeepaliveTimeout},
		{"graceful_shutdown_timeout", c.GracefulShutdownTimeout},
		{"keepalive_min_time", c.KeepaliveMinTime},
	}
	for _, tc := range checks {
		if tc.value != nil && *tc.value <= 0 {
			return fmt.Errorf("%s must be positive, got %v", tc.name, *tc.value)
		}
	}
	return nil
}

// TLSConfig represents server-side TLS configuration for the gRPC server.
// Server-side TLS encrypts traffic but does NOT require client certificates.
// Use this when clients authenticate via tokens (JWT, OAuth2) or are public-facing.
type TLSConfig struct {
	// CertFile is the path to the server TLS certificate file (required).
	CertFile string `yaml:"cert_file" json:"cert_file"`

	// KeyFile is the path to the server TLS private key file (required).
	KeyFile string `yaml:"key_file" json:"key_file"`
}

// MTLSConfig represents mutual TLS configuration for the gRPC server.
// mTLS always requires client certificate verification -- the client certificate
// IS the authentication mechanism (stateless, no tokens needed).
type MTLSConfig struct {
	// CertFile is the path to the server TLS certificate file (required).
	CertFile string `yaml:"cert_file" json:"cert_file"`

	// KeyFile is the path to the server TLS private key file (required).
	KeyFile string `yaml:"key_file" json:"key_file"`

	// CAFile is the path to the CA certificate file for client cert verification (required).
	// The server uses this CA to verify client certificates.
	CAFile string `yaml:"ca_file" json:"ca_file"`
}

// ClientPermission maps a certificate identity to allowed gRPC methods.
// The identity is matched against the client certificate's Common Name (CN).
// This provides stateless, certificate-based authorization.
type ClientPermission struct {
	// ClientIdentity is the expected Common Name (CN) from the client certificate.
	// Example: "user-service", "payment-gateway"
	ClientIdentity string `yaml:"client_identity" json:"client_identity"`

	// AllowedMethods is a list of gRPC method patterns this client can call.
	// Supports:
	//   - Exact: "/package.Service/Method"
	//   - Service wildcard: "/package.Service/*"
	//   - Global wildcard: "/*"
	AllowedMethods []string `yaml:"allowed_methods" json:"allowed_methods"`
}

// Config represents the gRPC server configuration.
type Config struct {
	// ServiceName identifies the service for logging and health checks.
	ServiceName string `yaml:"service_name" json:"service_name"`

	// Version is the service version (optional).
	// Displayed in server startup logs.
	Version string `yaml:"version" json:"version"`

	// Port is the TCP port the gRPC server listens on.
	// Default: 50051
	Port int `yaml:"port" json:"port"`

	// MaxRecvMsgSize is the maximum message size the server can receive in bytes.
	// Default: 4MB
	MaxRecvMsgSize int `yaml:"max_recv_msg_size" json:"max_recv_msg_size"`

	// MaxSendMsgSize is the maximum message size the server can send in bytes.
	// Default: 4MB
	MaxSendMsgSize int `yaml:"max_send_msg_size" json:"max_send_msg_size"`

	// MaxConcurrentStreams is the maximum number of concurrent streams per connection.
	// Default: 256
	MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams" json:"max_concurrent_streams"`

	// ConnectionTimeout is the maximum time for a new connection to complete.
	// Default: 120 seconds.
	ConnectionTimeout *time.Duration `yaml:"connection_timeout" json:"connection_timeout"`

	// KeepaliveTime is the duration after which the server pings the client if no activity.
	// Default: 2 hours.
	KeepaliveTime *time.Duration `yaml:"keepalive_time" json:"keepalive_time"`

	// KeepaliveTimeout is the duration the server waits for a keepalive response.
	// Default: 20 seconds.
	KeepaliveTimeout *time.Duration `yaml:"keepalive_timeout" json:"keepalive_timeout"`

	// GracefulShutdownTimeout is the maximum time to wait for in-flight requests to complete.
	// Default: 30 seconds.
	GracefulShutdownTimeout *time.Duration `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`

	// TLS contains server-side TLS configuration (encryption only, no client cert required).
	// When nil, the server uses insecure connections (unless MTLS is set).
	// Cannot be used together with MTLS.
	//
	// Use TLS when:
	//   - Clients authenticate via tokens (JWT, OAuth2)
	//   - The server is behind a TLS-terminating load balancer
	//   - You need encryption without client certificate management
	TLS *TLSConfig `yaml:"tls" json:"tls"`

	// MTLS contains mutual TLS configuration (encryption + client certificate verification).
	// When set, mTLS is enforced: the server verifies client certificates at the transport level.
	// Cannot be used together with TLS.
	//
	// Use MTLS when:
	//   - Internal service-to-service communication
	//   - Certificate-based identity is the authentication mechanism
	MTLS *MTLSConfig `yaml:"mtls" json:"mtls"`

	// Permissions defines certificate-based method-level authorization per client.
	// Each entry maps a certificate identity (CN) to its allowed gRPC methods.
	// Only effective when MTLS is configured.
	Permissions []ClientPermission `yaml:"permissions" json:"permissions"`

	// EnableReflection enables gRPC server reflection for tools like grpcurl.
	// Default: false.
	EnableReflection bool `yaml:"enable_reflection" json:"enable_reflection"`

	// VerboseLogging enables detailed request/response logging.
	// Default: false.
	VerboseLogging bool `yaml:"verbose_logging" json:"verbose_logging"`

	// VerboseLoggingSkipMethods is a list of method patterns to exclude from verbose logging.
	// Supports exact match and wildcards (e.g., "/grpc.health.v1.Health/*").
	VerboseLoggingSkipMethods []string `yaml:"verbose_logging_skip_methods" json:"verbose_logging_skip_methods"`

	// LogPayloads enables logging of request/response payloads in verbose mode.
	// Payloads are truncated to MaxPayloadLogSize bytes.
	// WARNING: May contain sensitive data. Use only for debugging.
	// Default: false.
	LogPayloads bool `yaml:"log_payloads" json:"log_payloads"`

	// MaxPayloadLogSize is the maximum number of bytes to log from request/response payloads.
	// Only effective when LogPayloads is true.
	// Default: 1024 bytes.
	MaxPayloadLogSize int `yaml:"max_payload_log_size" json:"max_payload_log_size"`

	// SlowRequestThreshold is the duration threshold for slow RPC detection.
	// RPCs exceeding this threshold will be logged with a warning.
	// Set 0 to disable slow RPC detection (default).
	// Example: SlowRequestThreshold: 3 * time.Second
	SlowRequestThreshold time.Duration `yaml:"slow_request_threshold" json:"slow_request_threshold"`

	// KeepaliveMinTime is the minimum time between client pings.
	// Protects against client ping floods.
	// Default: 10 seconds.
	KeepaliveMinTime *time.Duration `yaml:"keepalive_min_time" json:"keepalive_min_time"`

	// KeepalivePermitWithoutStream allows pings even when there are no active streams.
	// Default: false (deny pings without streams).
	KeepalivePermitWithoutStream *bool `yaml:"keepalive_permit_without_stream" json:"keepalive_permit_without_stream"`

	// EnableCompression enables gzip payload compression for gRPC messages.
	// Reduces bandwidth at low CPU cost.
	// Default: false.
	EnableCompression bool `yaml:"enable_compression" json:"enable_compression"`

	// ReadBufferSize is the read buffer size per connection in bytes.
	// Larger buffers improve throughput for large messages.
	// Default: 32KB (gRPC default). Set 0 to use default.
	ReadBufferSize int `yaml:"read_buffer_size" json:"read_buffer_size"`

	// WriteBufferSize is the write buffer size per connection in bytes.
	// Default: 32KB (gRPC default). Set 0 to use default.
	WriteBufferSize int `yaml:"write_buffer_size" json:"write_buffer_size"`

	// InitialWindowSize is the initial flow-control window size per stream in bytes.
	// Larger windows allow more in-flight data before flow-control kicks in.
	// Default: 64KB (gRPC default). Set 0 to use default.
	InitialWindowSize int32 `yaml:"initial_window_size" json:"initial_window_size"`

	// InitialConnWindowSize is the initial flow-control window size per connection in bytes.
	// Default: 64KB (gRPC default). Set 0 to use default.
	InitialConnWindowSize int32 `yaml:"initial_conn_window_size" json:"initial_conn_window_size"`
}

// MergeConfigDefaults returns a copy of the config with defaults filled in for zero-value fields.
// Does not mutate the original config.
func MergeConfigDefaults(conf *Config) *Config {
	merged := *conf // shallow copy

	if merged.Port == 0 {
		merged.Port = DefaultPort
	}
	if merged.MaxRecvMsgSize == 0 {
		merged.MaxRecvMsgSize = DefaultMaxRecvMsgSize
	}
	if merged.MaxSendMsgSize == 0 {
		merged.MaxSendMsgSize = DefaultMaxSendMsgSize
	}
	if merged.MaxConcurrentStreams == 0 {
		merged.MaxConcurrentStreams = DefaultMaxConcurrentStreams
	}
	if merged.MaxPayloadLogSize == 0 {
		merged.MaxPayloadLogSize = DefaultMaxPayloadLogSize
	}

	setDurationDefault(&merged.ConnectionTimeout, DefaultConnectionTimeout)
	setDurationDefault(&merged.KeepaliveTime, DefaultKeepaliveTime)
	setDurationDefault(&merged.KeepaliveTimeout, DefaultKeepaliveTimeout)
	setDurationDefault(&merged.GracefulShutdownTimeout, DefaultGracefulShutdownTimeout)
	setDurationDefault(&merged.KeepaliveMinTime, DefaultKeepaliveMinTime)

	if merged.KeepalivePermitWithoutStream == nil {
		d := DefaultKeepalivePermitWithoutStream
		merged.KeepalivePermitWithoutStream = &d
	}
	return &merged
}

// setDurationDefault sets a *time.Duration field to the default if it is nil.
func setDurationDefault(field **time.Duration, defaultVal time.Duration) {
	if *field == nil {
		d := defaultVal
		*field = &d
	}
}
