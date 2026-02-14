// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package configuration contains configuration types and defaults for the orianna HTTP server.
package configuration

import "time"

// MiddlewareConfig holds configuration for default middlewares.
// All fields default to false (middleware enabled).
type MiddlewareConfig struct {
	// DisableHelmet disables security headers middleware (helmet).
	// Helmet adds security headers like X-Frame-Options, X-Content-Type-Options, etc.
	// Default: false (enabled)
	// Set true only if you handle security headers manually
	DisableHelmet bool

	// DisableRateLimit disables rate limiting middleware.
	// Rate limiting prevents abuse by limiting requests per IP/user.
	// Default: false (enabled)
	// Set true for internal services or when using external rate limiting
	DisableRateLimit bool

	// DisableCompression disables response compression middleware.
	// Compression reduces bandwidth usage (gzip/deflate).
	// Default: false (enabled)
	// Set true if using CDN compression or serving pre-compressed files
	DisableCompression bool

	// DisableRecovery disables panic recovery middleware.
	// Recovery prevents panics from crashing the server.
	// Default: false (enabled)
	// WARNING: Never disable in production!
	DisableRecovery bool

	// DisableRequestID disables request ID generation middleware.
	// Request IDs help track requests across services and logs.
	// Default: false (enabled)
	// Set true only if using custom request ID generation
	DisableRequestID bool

	// DisableTraceID disables distributed tracing ID middleware.
	// Trace IDs help correlate requests in distributed systems.
	// Default: false (enabled)
	// Set true if using external tracing system (e.g., OpenTelemetry)
	DisableTraceID bool

	// DisableLogging disables request/response logging middleware.
	// Logs HTTP method, path, status, duration, and errors.
	// Default: false (enabled)
	// Set true only for very high-traffic services where logging impacts performance
	DisableLogging bool
}

// DefaultMiddlewareConfig returns the default middleware configuration.
//
// Output:
//   - *MiddlewareConfig: Configuration with all middlewares enabled
//
// Example:
//
//	config := configuration.DefaultMiddlewareConfig()
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		DisableHelmet:      false,
		DisableRateLimit:   false,
		DisableCompression: false,
		DisableRecovery:    false,
		DisableRequestID:   false,
		DisableTraceID:     false,
		DisableLogging:     false,
	}
}

// Config represents the HTTP server configuration.
type Config struct {
	// ServiceName is the name of your service (required).
	// Used in logs, metrics, and server startup messages.
	// Example: "user-api", "payment-service"
	ServiceName string `yaml:"service_name" json:"service_name"`

	// Version is the service version (optional).
	// Displayed in server startup logs and can be used for monitoring.
	// Example: "v1.0.0", "2024.01.15"
	Version string `yaml:"version" json:"version"`

	// Port is the HTTP server port (required, 0-65535).
	// The port where the server listens for incoming requests.
	// Example: 8080, 3000
	Port int `yaml:"port" json:"port"`

	// ReadTimeout is the maximum duration for reading the entire request, including body.
	// Prevents slow clients from holding connections indefinitely.
	// Default: 15 seconds. Set nil to use default.
	// Example: 30*time.Second for slow clients
	ReadTimeout *time.Duration `yaml:"read_timeout" json:"read_timeout"`

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// Prevents slow clients from affecting server performance.
	// Default: 15 seconds. Set nil to use default.
	// Example: 60*time.Second for large file downloads
	WriteTimeout *time.Duration `yaml:"write_timeout" json:"write_timeout"`

	// IdleTimeout is the maximum time to wait for the next request when keep-alives are enabled.
	// Closes idle connections to free up resources.
	// Default: 60 seconds. Set nil to use default.
	// Example: 120*time.Second for long-polling applications
	IdleTimeout *time.Duration `yaml:"idle_timeout" json:"idle_timeout"`

	// GracefulShutdownTimeout is the maximum duration to wait for the server to shutdown gracefully.
	// Allows in-flight requests to complete before forcing shutdown.
	// Default: 30 seconds. Set nil to use default.
	// Example: 60*time.Second for long-running operations
	GracefulShutdownTimeout *time.Duration `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`

	// MaxBodySize is the maximum allowed request body size in bytes.
	// Prevents memory exhaustion from overly large requests.
	// Default: 4MB (4194304 bytes). Set 0 to use default.
	// Example: 10*1024*1024 for 10MB file uploads
	MaxBodySize int `yaml:"max_body_size" json:"max_body_size"`

	// MaxConcurrentConnections is the maximum number of concurrent connections.
	// Limits resource usage and prevents connection exhaustion.
	// Default: 0 (unlimited). Set positive value to limit.
	// Example: 1000 for high-traffic services
	MaxConcurrentConnections int `yaml:"max_concurrent_connections" json:"max_concurrent_connections"`

	// EnableCORS enables Cross-Origin Resource Sharing middleware.
	// Required for browser-based clients from different domains.
	// Requires CORS config to be set if enabled.
	// Default: false
	EnableCORS bool `yaml:"enable_cors" json:"enable_cors"`

	// EnableCSRF enables Cross-Site Request Forgery protection.
	// Protects against CSRF attacks with token validation.
	// Requires CSRF config to be set if enabled.
	// Default: false
	EnableCSRF bool `yaml:"enable_csrf" json:"enable_csrf"`

	// CSRF contains CSRF protection configuration.
	// Required when EnableCSRF is true.
	// Defines token lookup, cookie settings, and expiration.
	CSRF *CSRFConfig `yaml:"csrf" json:"csrf"`

	// CORS contains CORS configuration.
	// Required when EnableCORS is true.
	// Defines allowed origins, methods, headers, and credentials.
	CORS *CORSConfig `yaml:"cors" json:"cors"`

	// VerboseLogging enables detailed request/response logging.
	// Useful for debugging but may impact performance in production.
	// Default: false
	VerboseLogging bool `yaml:"verbose_logging" json:"verbose_logging"`

	// VerboseLoggingSkipPaths is a list of paths to exclude from request logging.
	// Reduces log noise for health checks and metrics endpoints.
	// Example: []string{"/health", "/metrics", "/ready"}
	VerboseLoggingSkipPaths []string `yaml:"verbose_logging_skip_paths" json:"verbose_logging_skip_paths"`

	// UseProperHTTPStatus determines whether to use proper HTTP status codes for errors.
	// If true: error responses use appropriate HTTP status (400, 404, 500, etc.)
	// If false: all responses use 200 OK with error details in body (legacy API style)
	// Default: false (for backward compatibility)
	UseProperHTTPStatus bool `yaml:"use_proper_http_status" json:"use_proper_http_status"`
}

// CSRFConfig represents CSRF protection configuration.
type CSRFConfig struct {
	// KeyLookup defines where to find the CSRF token in the request.
	// Format: "source:key" where source can be: header, form, query, cookie
	// Example: "header:X-CSRF-Token", "form:csrf_token"
	// Default: "header:X-CSRF-Token"
	KeyLookup string `yaml:"key_lookup" json:"key_lookup"`

	// CookieName is the name of the CSRF cookie.
	// The cookie stores the CSRF token for validation.
	// Default: "_csrf"
	// Example: "csrf_token", "xsrf-token"
	CookieName string `yaml:"cookie_name" json:"cookie_name"`

	// CookiePath restricts the cookie to a specific path.
	// Default: "/" (available site-wide)
	// Example: "/api" to limit to API routes only
	CookiePath string `yaml:"cookie_path" json:"cookie_path"`

	// CookieDomain restricts the cookie to a specific domain.
	// Empty means current domain only.
	// Example: ".example.com" for subdomain sharing
	CookieDomain string `yaml:"cookie_domain" json:"cookie_domain"`

	// CookieSameSite controls when the cookie is sent.
	// Values: "Strict", "Lax", "None"
	// Default: "Strict" (only same-site requests)
	// "Lax": sent on top-level navigation
	// "None": requires CookieSecure=true
	CookieSameSite string `yaml:"cookie_same_site" json:"cookie_same_site"`

	// CookieSecure makes the cookie HTTPS-only.
	// Prevents token theft over insecure connections.
	// Default: false. Set true in production.
	CookieSecure bool `yaml:"cookie_secure" json:"cookie_secure"`

	// CookieHTTPOnly prevents JavaScript access to the cookie.
	// Protects against XSS attacks.
	// Default: false
	// Recommended: true for security
	CookieHTTPOnly bool `yaml:"cookie_http_only" json:"cookie_http_only"`

	// CookieSessionOnly makes the cookie expire when browser closes.
	// If false, uses Expiration duration.
	// Default: false
	CookieSessionOnly bool `yaml:"cookie_session_only" json:"cookie_session_only"`

	// SingleUseToken invalidates token after first use.
	// Stronger protection but requires token refresh for multi-request forms.
	// Default: false
	// Set true for critical operations (e.g., payment, delete)
	SingleUseToken bool `yaml:"single_use_token" json:"single_use_token"`

	// Expiration is how long the CSRF token is valid.
	// Only used if CookieSessionOnly is false.
	// Default: 24 hours
	// Example: 1*time.Hour for short-lived tokens
	Expiration *time.Duration `yaml:"expiration" json:"expiration"`
}

// CORSConfig represents CORS configuration.
type CORSConfig struct {
	// AllowOrigins is a list of origins allowed to make cross-origin requests.
	// Use "*" to allow all origins (not recommended for production).
	// Required when EnableCORS is true.
	// Example: []string{"https://example.com", "https://app.example.com"}
	AllowOrigins []string `yaml:"allow_origins" json:"allow_origins"`

	// AllowMethods is a list of HTTP methods allowed for cross-origin requests.
	// Common methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
	// Required when EnableCORS is true.
	// Example: []string{"GET", "POST", "PUT", "DELETE"}
	AllowMethods []string `yaml:"allow_methods" json:"allow_methods"`

	// AllowHeaders is a list of request headers allowed in cross-origin requests.
	// Common headers: Content-Type, Authorization, X-Requested-With
	// Example: []string{"Content-Type", "Authorization", "X-API-Key"}
	AllowHeaders []string `yaml:"allow_headers" json:"allow_headers"`

	// AllowCredentials indicates whether credentials (cookies, auth headers) are allowed.
	// If true, AllowOrigins cannot be "*" for security reasons.
	// Default: false
	// Set true when using cookies or authentication
	AllowCredentials bool `yaml:"allow_credentials" json:"allow_credentials"`

	// ExposeHeaders lists response headers that browsers are allowed to access.
	// By default, browsers only expose simple headers (Cache-Control, Content-Type, etc.)
	// Example: []string{"X-Total-Count", "X-Page-Number"}
	ExposeHeaders []string `yaml:"expose_headers" json:"expose_headers"`

	// MaxAge is how long (in seconds) browsers can cache preflight request results.
	// Reduces preflight requests for better performance.
	// Default: 0 (no caching)
	// Example: 3600 (1 hour), 86400 (24 hours)
	MaxAge int `yaml:"max_age" json:"max_age"`
}
