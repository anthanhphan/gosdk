package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
)

const (
	MinPort = 1
	MaxPort = 65535

	DefaultPort                     = 8080
	DefaultMaxBodySize              = 4 * 1024 * 1024
	DefaultMaxConcurrentConnections = 262144
	DefaultShutdownTimeout          = 30 * time.Second

	ContextKey = "aurelion_config"
)

// Config represents the HTTP server configuration.
type Config struct {
	ServiceName              string         `json:"service_name" yaml:"service_name" validate:"required"`
	Port                     int            `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	ReadTimeout              *time.Duration `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`
	WriteTimeout             *time.Duration `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`
	IdleTimeout              *time.Duration `json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty"`
	GracefulShutdownTimeout  *time.Duration `json:"graceful_shutdown_timeout,omitempty" yaml:"graceful_shutdown_timeout,omitempty"`
	MaxBodySize              int            `json:"max_body_size,omitempty" yaml:"max_body_size,omitempty"`
	MaxConcurrentConnections int            `json:"max_concurrent_connections,omitempty" yaml:"max_concurrent_connections,omitempty"`
	EnableCORS               bool           `json:"enable_cors,omitempty" yaml:"enable_cors,omitempty"`
	EnableCSRF               bool           `json:"enable_csrf,omitempty" yaml:"enable_csrf,omitempty"`
	CSRF                     *CSRFConfig    `json:"csrf,omitempty" yaml:"csrf,omitempty"`
	CORS                     *CORSConfig    `json:"cors,omitempty" yaml:"cors,omitempty"`
	VerboseLogging           bool           `json:"verbose_logging,omitempty" yaml:"verbose_logging,omitempty"`
	UseProperHTTPStatus      bool           `json:"use_proper_http_status,omitempty" yaml:"use_proper_http_status,omitempty"`
}

// Default creates a new Config instance with sensible default values.
//
// Input:
//   - None
//
// Output:
//   - *Config: A new configuration with default settings
//
// Example:
//
//	cfg := config.Default()
//	cfg.Port = 3000
//	server, _ := aurelion.NewHttpServer(cfg)
func Default() *Config {
	return &Config{
		ServiceName:              "HTTP Server",
		Port:                     DefaultPort,
		MaxBodySize:              DefaultMaxBodySize,
		MaxConcurrentConnections: DefaultMaxConcurrentConnections,
	}
}

// Merge fills in zero-value fields with defaults, returning the updated config.
//
// Input:
//   - None (receiver method on *Config)
//
// Output:
//   - *Config: The config with defaults merged in
//
// Example:
//
//	cfg := &config.Config{ServiceName: "My API"}
//	cfg = cfg.Merge() // Port, MaxBodySize, etc. will be set to defaults
func (c *Config) Merge() *Config {
	defaults := Default()

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

// Validate checks if the configuration is valid according to defined rules.
//
// Input:
//   - None (receiver method on *Config)
//
// Output:
//   - error: Validation error if config is invalid, nil otherwise
//
// Example:
//
//	cfg := &config.Config{ServiceName: "My API", Port: 8080}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func (c *Config) Validate() error {
	if c == nil {
		return core.ErrConfigNil
	}

	if err := c.validateBasicFields(); err != nil {
		return err
	}

	if err := c.validateTimeouts(); err != nil {
		return err
	}

	if err := c.validateLimits(); err != nil {
		return err
	}

	if err := c.validateCORS(); err != nil {
		return err
	}

	if err := c.validateCSRF(); err != nil {
		return fmt.Errorf("csrf config: %w", err)
	}

	return nil
}

func (c *Config) validateBasicFields() error {
	if c.ServiceName == "" {
		return errors.New("service_name is required")
	}

	if c.Port < MinPort || c.Port > MaxPort {
		return fmt.Errorf("port must be between %d and %d, got %d", MinPort, MaxPort, c.Port)
	}

	return nil
}

func (c *Config) validateTimeouts() error {
	if c.ReadTimeout != nil && *c.ReadTimeout < 0 {
		return errors.New("read_timeout cannot be negative")
	}

	if c.WriteTimeout != nil && *c.WriteTimeout < 0 {
		return errors.New("write_timeout cannot be negative")
	}

	if c.IdleTimeout != nil && *c.IdleTimeout < 0 {
		return errors.New("idle_timeout cannot be negative")
	}

	if c.GracefulShutdownTimeout != nil && *c.GracefulShutdownTimeout < 0 {
		return errors.New("graceful_shutdown_timeout cannot be negative")
	}

	return nil
}

func (c *Config) validateLimits() error {
	if c.MaxBodySize < 0 {
		return errors.New("max_body_size cannot be negative")
	}

	if c.MaxConcurrentConnections < 0 {
		return errors.New("max_concurrent_connections cannot be negative")
	}

	return nil
}

func (c *Config) validateCORS() error {
	if !c.EnableCORS {
		return nil
	}

	if c.CORS == nil {
		return errors.New("cors config is required when enable_cors is true")
	}

	if err := c.CORS.Validate(); err != nil {
		return fmt.Errorf("cors config: %w", err)
	}

	return nil
}

func (c *Config) validateCSRF() error {
	if !c.EnableCSRF {
		return nil
	}

	if c.CSRF == nil {
		return errors.New("csrf config is required when enable_csrf is true")
	}

	if err := c.CSRF.Validate(); err != nil {
		return fmt.Errorf("csrf config: %w", err)
	}

	return nil
}

// CSRFConfig represents CSRF protection configuration.
type CSRFConfig struct {
	KeyLookup         string         `json:"key_lookup,omitempty" yaml:"key_lookup,omitempty"`
	CookieName        string         `json:"cookie_name,omitempty" yaml:"cookie_name,omitempty"`
	CookiePath        string         `json:"cookie_path,omitempty" yaml:"cookie_path,omitempty"`
	CookieDomain      string         `json:"cookie_domain,omitempty" yaml:"cookie_domain,omitempty"`
	CookieSameSite    string         `json:"cookie_same_site,omitempty" yaml:"cookie_same_site,omitempty"`
	CookieSecure      bool           `json:"cookie_secure,omitempty" yaml:"cookie_secure,omitempty"`
	CookieHTTPOnly    bool           `json:"cookie_http_only,omitempty" yaml:"cookie_http_only,omitempty"`
	CookieSessionOnly bool           `json:"cookie_session_only,omitempty" yaml:"cookie_session_only,omitempty"`
	SingleUseToken    bool           `json:"single_use_token,omitempty" yaml:"single_use_token,omitempty"`
	Expiration        *time.Duration `json:"expiration,omitempty" yaml:"expiration,omitempty"`
}

// Validate checks if the CSRF configuration is valid.
//
// Input:
//   - None (receiver method on *CSRFConfig)
//
// Output:
//   - error: Validation error if CSRF config is invalid, nil otherwise
//
// Example:
//
//	csrfCfg := &config.CSRFConfig{KeyLookup: "header:X-Csrf-Token"}
//	if err := csrfCfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func (c *CSRFConfig) Validate() error {
	if c.KeyLookup != "" {
		parts := strings.Split(c.KeyLookup, ":")
		if len(parts) != 2 {
			return fmt.Errorf("key_lookup must be in the form of <source>:<key>, got: %s", c.KeyLookup)
		}
		validSources := map[string]bool{
			"header": true,
			"query":  true,
			"param":  true,
			"form":   true,
			"cookie": true,
		}
		if !validSources[parts[0]] {
			return fmt.Errorf("invalid key_lookup source: %s (must be header, query, param, form, or cookie)", parts[0])
		}
		if parts[1] == "" {
			return errors.New("key_lookup key cannot be empty")
		}
	}

	if c.Expiration != nil && *c.Expiration < 0 {
		return errors.New("expiration cannot be negative")
	}

	if c.CookieSameSite != "" {
		validSameSite := map[string]bool{
			"Strict": true,
			"Lax":    true,
			"None":   true,
		}
		if !validSameSite[c.CookieSameSite] {
			return fmt.Errorf("invalid cookie_same_site: %s (must be Strict, Lax, or None)", c.CookieSameSite)
		}
	}

	return nil
}

// CORSConfig represents CORS configuration.
type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins" yaml:"allow_origins"`
	AllowMethods     []string `json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers" yaml:"allow_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	ExposeHeaders    []string `json:"expose_headers,omitempty" yaml:"expose_headers,omitempty"`
	MaxAge           int      `json:"max_age,omitempty" yaml:"max_age,omitempty"`
}

// Validate checks if the CORS configuration is valid.
//
// Input:
//   - None (receiver method on *CORSConfig)
//
// Output:
//   - error: Validation error if CORS config is invalid, nil otherwise
//
// Example:
//
//	corsCfg := &config.CORSConfig{
//	    AllowOrigins: []string{"https://example.com"},
//	    AllowMethods: []string{"GET", "POST"},
//	}
//	if err := corsCfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func (c *CORSConfig) Validate() error {
	if len(c.AllowOrigins) == 0 {
		return errors.New("allow_origins is required")
	}

	if len(c.AllowMethods) == 0 {
		return errors.New("allow_methods is required")
	}

	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
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

// StoreInContext stores the server config in the request context for handler access.
//
// Input:
//   - ctx: The request context
//   - cfg: The config to store
//
// Output:
//   - None
//
// Example:
//
//	config.StoreInContext(ctx, serverConfig)
func StoreInContext(ctx core.Context, cfg *Config) {
	if ctx == nil || cfg == nil {
		return
	}
	ctx.Locals(ContextKey, cfg)
}

// FromContext retrieves the server configuration stored in the request context.
//
// Input:
//   - ctx: The request context
//
// Output:
//   - *Config: The stored configuration, or nil if not found
//
// Example:
//
//	cfg := config.FromContext(ctx)
//	if cfg != nil && cfg.UseProperHTTPStatus {
//	    // Use proper HTTP status codes
//	}
func FromContext(ctx core.Context) *Config {
	if ctx == nil {
		return nil
	}
	if value := ctx.Locals(ContextKey); value != nil {
		if cfg, ok := value.(*Config); ok {
			return cfg
		}
	}
	return nil
}
