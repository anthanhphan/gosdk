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

// Health check endpoint
const (
	// DefaultHealthCheckPath is the default path for health check endpoint
	DefaultHealthCheckPath = "/health"
)
