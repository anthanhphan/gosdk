package aurelion

import "time"

// Server configuration constants
const (
	// MinPort is the minimum valid port number
	MinPort = 1

	// MaxPort is the maximum valid port number
	MaxPort = 65535

	// DefaultPort is the default server port
	DefaultPort = 8080

	// MaxRoutePathLength is the maximum allowed length for a route path
	MaxRoutePathLength = 1024

	// MaxRouteHandlersPerRoute is the maximum number of handlers (middlewares + handler) per route
	MaxRouteHandlersPerRoute = 50

	// DefaultMaxBodySize is the default maximum request body size (4MB)
	DefaultMaxBodySize = 4 * 1024 * 1024

	// DefaultMaxConcurrentConnections is the default maximum concurrent connections
	DefaultMaxConcurrentConnections = 262144

	// DefaultShutdownTimeout is the default graceful shutdown timeout
	DefaultShutdownTimeout = 30 * time.Second

	// DefaultRateLimitMax is the default maximum number of requests per IP per time window
	DefaultRateLimitMax = 500

	// DefaultRateLimitExpiration is the default time window for rate limiting
	DefaultRateLimitExpiration = 1 * time.Minute
)

// Request ID header
const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"

	// contextKeyRequestID is the context key for request ID (internal use)
	contextKeyRequestID = "request_id"
)

// Config key
const (
	// contextKeyConfig is the context key for server config (internal use)
	contextKeyConfig = "aurelion_config"
)

// Trace ID header
const (
	// TraceIDHeader is the header name for trace ID
	TraceIDHeader = "X-Trace-ID"

	// contextKeyTraceID is the context key for trace ID (internal use)
	contextKeyTraceID = "trace_id"
)

// Health check endpoint
const (
	// DefaultHealthCheckPath is the default path for health check endpoint
	DefaultHealthCheckPath = "/health"
)

// Error messages
const (
	// ErrContextNil is the error message when context is nil
	ErrContextNil = "context cannot be nil"
	// ErrUnknownError is the error message for unknown errors
	ErrUnknownError = "unknown error"
	// ErrConfigNil is the error message when config is nil
	ErrConfigNil = "config cannot be nil"
)

// Response messages
const (
	// MsgHealthCheckHealthy is the health check success message
	MsgHealthCheckHealthy = "Server is healthy"
	// MsgValidationFailed is the validation failure message
	MsgValidationFailed = "Validation failed"
)

// Error type constants
const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation = "validation_error"
	// ErrorTypeBusiness represents business logic errors
	ErrorTypeBusiness = "business_error"
	// ErrorTypePermission represents permission/authorization errors
	ErrorTypePermission = "permission_error"
	// ErrorTypeRateLimit represents rate limiting errors
	ErrorTypeRateLimit = "rate_limit_error"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal = "external_api_error"
	// ErrorTypeInternalServerError represents internal server errors
	ErrorTypeInternalServerError = "internal_server_error"
)

// Health check status
const (
	// HealthStatusHealthy indicates the server is healthy
	HealthStatusHealthy = "healthy"
	// HealthStatusDegraded indicates the server is degraded
	HealthStatusDegraded = "degraded"
	// HealthStatusUnhealthy indicates the server is unhealthy
	HealthStatusUnhealthy = "unhealthy"
)
