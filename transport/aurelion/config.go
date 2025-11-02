package aurelion

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Config represents the HTTP server configuration
type Config struct {
	// ServiceName is the name of the service (required)
	ServiceName string `json:"service_name" yaml:"service_name" validate:"required"`

	// Port is the port to listen on (required, default: 8080)
	Port int `json:"port" yaml:"port" validate:"required,min=1,max=65535"`

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout *time.Duration `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`

	// WriteTimeout is the maximum duration before timing out writes
	WriteTimeout *time.Duration `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`

	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout *time.Duration `json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty"`

	// GracefulShutdownTimeout is the timeout for graceful shutdown
	GracefulShutdownTimeout *time.Duration `json:"graceful_shutdown_timeout,omitempty" yaml:"graceful_shutdown_timeout,omitempty"`

	// MaxBodySize is the maximum allowed size for request body
	MaxBodySize int `json:"max_body_size,omitempty" yaml:"max_body_size,omitempty"`

	// MaxConcurrentConnections is the maximum number of concurrent connections
	MaxConcurrentConnections int `json:"max_concurrent_connections,omitempty" yaml:"max_concurrent_connections,omitempty"`

	// EnableCORS enables CORS support
	EnableCORS bool `json:"enable_cors,omitempty" yaml:"enable_cors,omitempty"`

	// CORS configuration
	CORS *CORSConfig `json:"cors,omitempty" yaml:"cors,omitempty"`

	// VerboseLogging enables verbose request/response logging.
	// When enabled, logs include request/response bodies, query parameters, and route params.
	// This can impact performance in high-throughput scenarios.
	VerboseLogging bool `json:"verbose_logging,omitempty" yaml:"verbose_logging,omitempty"`
}

// Validate validates server configuration
//
// Input:
//   - none (receiver method)
//
// Output:
//   - error: Any validation error
//
// Example:
//
//	if err := config.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func (config *Config) Validate() error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if err := config.validateBasicFields(); err != nil {
		return err
	}

	if err := config.validateTimeouts(); err != nil {
		return err
	}

	if err := config.validateLimits(); err != nil {
		return err
	}

	if err := config.validateCORS(); err != nil {
		return err
	}

	return nil
}

// validateBasicFields validates service name and port
func (config *Config) validateBasicFields() error {
	if config.ServiceName == "" {
		return errors.New("service_name is required")
	}

	if config.Port < MinPort || config.Port > MaxPort {
		return fmt.Errorf("port must be between %d and %d, got %d", MinPort, MaxPort, config.Port)
	}

	return nil
}

// validateTimeouts validates timeout configurations
func (config *Config) validateTimeouts() error {
	if config.ReadTimeout != nil && *config.ReadTimeout < 0 {
		return errors.New("read_timeout cannot be negative")
	}

	if config.WriteTimeout != nil && *config.WriteTimeout < 0 {
		return errors.New("write_timeout cannot be negative")
	}

	if config.IdleTimeout != nil && *config.IdleTimeout < 0 {
		return errors.New("idle_timeout cannot be negative")
	}

	if config.GracefulShutdownTimeout != nil && *config.GracefulShutdownTimeout < 0 {
		return errors.New("graceful_shutdown_timeout cannot be negative")
	}

	return nil
}

// validateLimits validates size and connection limits
func (config *Config) validateLimits() error {
	if config.MaxBodySize < 0 {
		return errors.New("max_body_size cannot be negative")
	}

	if config.MaxConcurrentConnections < 0 {
		return errors.New("max_concurrent_connections cannot be negative")
	}

	return nil
}

// validateCORS validates CORS configuration
func (config *Config) validateCORS() error {
	if !config.EnableCORS {
		return nil
	}

	if config.CORS == nil {
		return errors.New("cors config is required when enable_cors is true")
	}

	if err := config.CORS.Validate(); err != nil {
		return fmt.Errorf("cors config: %w", err)
	}

	return nil
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	// AllowOrigins defines allowed origins
	AllowOrigins []string `json:"allow_origins" yaml:"allow_origins"`

	// AllowMethods defines allowed HTTP methods
	AllowMethods []string `json:"allow_methods" yaml:"allow_methods"`

	// AllowHeaders defines allowed headers
	AllowHeaders []string `json:"allow_headers" yaml:"allow_headers"`

	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool `json:"allow_credentials" yaml:"allow_credentials"`

	// ExposeHeaders defines headers that are safe to expose
	ExposeHeaders []string `json:"expose_headers,omitempty" yaml:"expose_headers,omitempty"`

	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge int `json:"max_age,omitempty" yaml:"max_age,omitempty"`
}

// Validate validates the CORS configuration
//
// Input:
//   - none (receiver method)
//
// Output:
//   - error: Any validation error
//
// Example:
//
//	corsConfig := &aurelion.CORSConfig{
//	    AllowOrigins: []string{"https://example.com"},
//	    AllowMethods: []string{"GET", "POST"},
//	    AllowHeaders: []string{"Content-Type"},
//	}
//	if err := corsConfig.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func (c *CORSConfig) Validate() error {
	if len(c.AllowOrigins) == 0 {
		return errors.New("allow_origins is required")
	}

	if len(c.AllowMethods) == 0 {
		return errors.New("allow_methods is required")
	}

	// Validate HTTP methods
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true,
		"DELETE": true, "HEAD": true, "OPTIONS": true,
	}

	for _, method := range c.AllowMethods {
		upperMethod := strings.ToUpper(method)
		if !validMethods[upperMethod] {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
	}

	if c.MaxAge < 0 {
		return errors.New("max_age cannot be negative")
	}

	return nil
}

// DefaultConfig returns a config with default values
//
// Input:
//   - none
//
// Output:
//   - *Config: A new Config instance with default values
//
// Example:
//
//	config := aurelion.DefaultConfig()
//	config.ServiceName = "My API"
//	config.Port = 3000
func DefaultConfig() *Config {
	return &Config{
		ServiceName:              "HTTP Server",
		Port:                     DefaultPort,
		MaxBodySize:              DefaultMaxBodySize,
		MaxConcurrentConnections: DefaultMaxConcurrentConnections,
	}
}

// Merge merges the provided config with default values.
// Fields with zero values are replaced with defaults.
//
// Input:
//   - none (receiver method)
//
// Output:
//   - *Config: The config instance with merged values
//
// Example:
//
//	config := &aurelion.Config{
//	    ServiceName: "My API",
//	    // Port defaults to 8080
//	}
//	config.Merge()
func (c *Config) Merge() *Config {
	defaults := DefaultConfig()

	if c.ServiceName == "" {
		c.ServiceName = defaults.ServiceName
	}

	if c.Port == 0 {
		c.Port = defaults.Port
	}

	if c.MaxBodySize == 0 {
		c.MaxBodySize = defaults.MaxBodySize
	}

	if c.MaxConcurrentConnections == 0 {
		c.MaxConcurrentConnections = defaults.MaxConcurrentConnections
	}

	return c
}
