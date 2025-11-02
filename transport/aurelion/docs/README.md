# Aurelion Package

A production-ready HTTP server framework for Go applications built on [Fiber](https://github.com/gofiber/fiber). Provides a fluent API for building REST APIs with built-in middleware, request/response handling, automatic request/trace ID generation, CORS, CSRF protection, and structured logging integration.

## Installation

```bash
go get github.com/anthanhphan/gosdk/transport/aurelion
```

## Quick Start

```go
package main

import (
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/transport/aurelion"
)

func main() {
	// Initialize logger
	undo := logger.InitDefaultLogger()
	defer undo()

	// Create server configuration
	config := &aurelion.Config{
		ServiceName: "My API",
		Port:        8080,
	}

	// Create server
	server, err := aurelion.NewHttpServer(config)
	if err != nil {
		logger.NewLoggerWithFields().Fatalw("failed to create server", "error", err)
	}

	// Add routes
	server.AddRoutes(
		aurelion.NewRoute("/users").
			GET().
			Handler(func(ctx aurelion.Context) error {
				return aurelion.OK(ctx, "Users list", []map[string]interface{}{
					{"id": 1, "name": "John"},
					{"id": 2, "name": "Jane"},
				})
			}),
	)

	// Start server
	if err := server.Start(); err != nil {
		logger.NewLoggerWithFields().Fatalw("server error", "error", err)
	}
}
```

## Configuration

### Config Structure

```go
type Config struct {
	ServiceName             string        // Service name (required)
	Port                    int           // Port to listen on (required, default: 8080)
	ReadTimeout             *time.Duration // Max duration for reading request
	WriteTimeout            *time.Duration // Max duration for writing response
	IdleTimeout             *time.Duration // Max idle time between requests
	GracefulShutdownTimeout *time.Duration // Shutdown timeout (default: 30s)
	MaxBodySize             int           // Max request body size (default: 4MB)
	MaxConcurrentConnections int          // Max concurrent connections (default: 262144)
	EnableCORS              bool          // Enable CORS support
	EnableCSRF              bool          // Enable CSRF protection
	CORS                    *CORSConfig   // CORS configuration
	CSRF                    *CSRFConfig   // CSRF configuration
	VerboseLogging          bool          // Enable verbose request/response logging
}
```

### Default Values

- **Port**: `8080`
- **MaxBodySize**: `4MB`
- **MaxConcurrentConnections**: `262144`
- **GracefulShutdownTimeout**: `30s`
- **Rate Limiting**: `500 requests/minute per IP`

## Usage Examples

### Basic Server Setup

```go
package main

import (
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/transport/aurelion"
)

func main() {
	// Initialize logger
	undo := logger.InitDefaultLogger()
	defer undo()

	// Create server with default config
	config := aurelion.DefaultConfig()
	config.ServiceName = "My API"
	config.Port = 3000

	server, err := aurelion.NewHttpServer(config)
	if err != nil {
		logger.Fatalw("failed to create server", "error", err)
	}

	// Add a simple route
	server.AddRoutes(
		aurelion.NewRoute("/health").
			GET().
			Handler(func(ctx aurelion.Context) error {
				return aurelion.HealthCheck(ctx)
			}),
	)

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		logger.Fatalw("server error", "error", err)
	}
}
```

### Creating Routes

#### Simple GET Route

```go
server.AddRoutes(
	aurelion.NewRoute("/users").
		GET().
		Handler(func(ctx aurelion.Context) error {
			users := []map[string]interface{}{
				{"id": 1, "name": "John"},
				{"id": 2, "name": "Jane"},
			}
			return aurelion.OK(ctx, "Users retrieved", users)
		}),
)
```

#### POST Route with Body Parsing

```go
import (
	"net/http"
	"time"
)

server.AddRoutes(
	aurelion.NewRoute("/users").
		POST().
		Handler(func(ctx aurelion.Context) error {
			var req struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}

			if err := ctx.BodyParser(&req); err != nil {
				return aurelion.BadRequest(ctx, "Invalid request body")
			}

			// Create user logic here
			user := map[string]interface{}{
				"id":    123,
				"name":  req.Name,
				"email": req.Email,
			}

			return ctx.Status(http.StatusCreated).JSON(aurelion.Map{
				"success":   true,
				"code":      http.StatusCreated,
				"message":   "User created",
				"data":      user,
				"timestamp": time.Now().UnixMilli(),
			})
		}),
)
```

#### Route with Parameters

```go
server.AddRoutes(
	aurelion.NewRoute("/users/:id").
		GET().
		Handler(func(ctx aurelion.Context) error {
			userID := ctx.Params("id")

			if userID == "999" {
				return aurelion.Error(ctx, aurelion.NewError(1001, "User not found"))
			}

			user := map[string]interface{}{
				"id":   userID,
				"name": "John Doe",
			}
			return aurelion.OK(ctx, "User details", user)
		}),
)
```

#### Route with Query Parameters

```go
server.AddRoutes(
	aurelion.NewRoute("/search").
		GET().
		Handler(func(ctx aurelion.Context) error {
			query := ctx.Query("q", "")
			page := ctx.Query("page", "1")

			if query == "" {
				return aurelion.BadRequest(ctx, "Query parameter 'q' is required")
			}

			results := map[string]interface{}{
				"query": query,
				"page":  page,
				"items": []string{"result1", "result2"},
			}
			return aurelion.OK(ctx, "Search results", results)
		}),
)
```

#### Protected Route (Authentication Required)

```go
server.AddRoutes(
	aurelion.NewRoute("/admin").
		GET().
		Protected(). // Requires authentication
		Handler(func(ctx aurelion.Context) error {
			// This route requires authentication middleware
			userID := ctx.Locals("user_id")
			return aurelion.OK(ctx, "Admin panel", aurelion.Map{"user_id": userID})
		}),
)
```

#### Route with Permissions (Authorization Required)

```go
server.AddRoutes(
	aurelion.NewRoute("/admin/users").
		GET().
		Protected().
		Permissions("read:users", "admin"). // Requires specific permissions
		Handler(func(ctx aurelion.Context) error {
			return aurelion.OK(ctx, "Users list", []string{"user1", "user2"})
		}),
)
```

#### Route with Custom Middleware

```go
customMiddleware := aurelion.Middleware(func(ctx aurelion.Context) error {
	logger.Info("Custom middleware executed", "path", ctx.Path())
	return ctx.Next()
})

server.AddRoutes(
	aurelion.NewRoute("/users").
		GET().
		Middleware(customMiddleware).
		Handler(func(ctx aurelion.Context) error {
			return aurelion.OK(ctx, "Users list", []map[string]interface{}{})
		}),
)
```

#### Route with CORS Configuration

```go
server.AddRoutes(
	aurelion.NewRoute("/api/users").
		GET().
		CORS(&aurelion.CORSConfig{
			AllowOrigins:     []string{"https://example.com"},
			AllowMethods:     []string{"GET", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
		}).
		Handler(func(ctx aurelion.Context) error {
			return aurelion.OK(ctx, "Users list", []map[string]interface{}{})
		}),
)
```

### Group Routes

#### Basic Group Route

```go
server.AddGroupRoutes(
	aurelion.NewGroupRoute("/api/v1").
		Routes(
			aurelion.NewRoute("/users").
				GET().
				Handler(func(ctx aurelion.Context) error {
					return aurelion.OK(ctx, "Users", []string{})
				}),
			aurelion.NewRoute("/posts").
				GET().
				Handler(func(ctx aurelion.Context) error {
					return aurelion.OK(ctx, "Posts", []string{})
				}),
		),
)
```

#### Group Route with Middleware

```go
loggingMiddleware := aurelion.Middleware(func(ctx aurelion.Context) error {
	logger.Info("Group middleware", "path", ctx.Path())
	return ctx.Next()
})

server.AddGroupRoutes(
	aurelion.NewGroupRoute("/api/v1").
		Middleware(loggingMiddleware).
		Routes(
			aurelion.NewRoute("/users").GET().Handler(userHandler),
			aurelion.NewRoute("/posts").GET().Handler(postHandler),
		),
)
```

#### Protected Group Route

```go
server.AddGroupRoutes(
	aurelion.NewGroupRoute("/admin").
		Protected(). // All routes in group require authentication
		Routes(
			aurelion.NewRoute("/users").GET().Handler(adminUsersHandler),
			aurelion.NewRoute("/settings").GET().Handler(adminSettingsHandler),
		),
)
```

### Server Options

#### With Authentication Middleware

```go
authMiddleware := aurelion.Middleware(func(ctx aurelion.Context) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return aurelion.Unauthorized(ctx, "Authentication required")
	}

	// Validate token and extract user ID
	userID := "123" // Your token validation logic
	ctx.Locals("user_id", userID)

	return ctx.Next()
})

server, err := aurelion.NewHttpServer(
	config,
	aurelion.WithAuthentication(authMiddleware),
)
```

#### With Authorization Checker

```go
authChecker := func(ctx aurelion.Context, requiredPermissions []string) error {
	userPermissions := ctx.Locals("permissions").([]string)

	for _, required := range requiredPermissions {
		if !contains(userPermissions, required) {
			return errors.New("missing permission: " + required)
		}
	}
	return nil
}

server, err := aurelion.NewHttpServer(
	config,
	aurelion.WithAuthentication(authMiddleware),
	aurelion.WithAuthorization(authChecker),
)
```

#### With Rate Limiter

```go
import "github.com/gofiber/fiber/v2/middleware/limiter"

customLimiter := limiter.New(limiter.Config{
	Max:        1000,
	Expiration: 1 * time.Minute,
	KeyGenerator: func(c *fiber.Ctx) string {
		// Rate limit by user ID
		if userID := c.Locals("user_id"); userID != nil {
			return fmt.Sprintf("user:%v", userID)
		}
		return c.IP()
	},
})

server, err := aurelion.NewHttpServer(
	config,
	aurelion.WithRateLimiter(customLimiter),
)
```

#### With Global Middleware

```go
loggingMiddleware := aurelion.Middleware(func(ctx aurelion.Context) error {
	start := time.Now()
	err := ctx.Next()
	duration := time.Since(start)

	logger.Infow("request completed",
		"method", ctx.Method(),
		"path", ctx.Path(),
		"duration_ms", duration.Milliseconds(),
	)

	return err
})

server, err := aurelion.NewHttpServer(
	config,
	aurelion.WithGlobalMiddleware(loggingMiddleware),
)
```

#### With Panic Recovery

```go
panicRecover := aurelion.Middleware(func(ctx aurelion.Context) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("panic recovered", "panic", r)
			aurelion.InternalServerError(ctx, "Internal server error")
		}
	}()
	return ctx.Next()
})

server, err := aurelion.NewHttpServer(
	config,
	aurelion.WithPanicRecover(panicRecover),
)
```

### Response Helpers

#### Success Responses

```go
import (
	"net/http"
	"time"
)

// 200 OK with message and data
return aurelion.OK(ctx, "Operation successful", aurelion.Map{"result": "data"})

// 201 Created
return ctx.Status(http.StatusCreated).JSON(aurelion.Map{
	"success":   true,
	"code":      http.StatusCreated,
	"message":   "Resource created",
	"data":      createdResource,
	"timestamp": time.Now().UnixMilli(),
})

// 202 Accepted
return ctx.Status(http.StatusAccepted).JSON(aurelion.Map{
	"success":   true,
	"code":      http.StatusAccepted,
	"message":   "Request accepted for processing",
	"timestamp": time.Now().UnixMilli(),
})

// 204 No Content
return ctx.Status(http.StatusNoContent).SendString("")
```

#### Error Responses

```go
// 400 Bad Request
return aurelion.BadRequest(ctx, "Invalid input")

// 401 Unauthorized
return aurelion.Unauthorized(ctx, "Authentication required")

// 403 Forbidden
return aurelion.Forbidden(ctx, "Access denied")

// 404 Not Found
return aurelion.NotFound(ctx, "Resource not found")

// 409 Conflict
return ctx.Status(http.StatusOK).JSON(aurelion.Map{
	"success":   false,
	"code":      http.StatusConflict,
	"message":   "Resource already exists",
	"timestamp": time.Now().UnixMilli(),
})
// Note: Use aurelion.OK() and aurelion.Error() helpers for common responses.
// For custom status codes, use ctx.Status() and ctx.JSON() directly.

// 500 Internal Server Error
return aurelion.InternalServerError(ctx, "Internal server error")

// Custom business error
err := aurelion.NewError(1001, "User not found")
return aurelion.Error(ctx, err)

// Formatted error
err := aurelion.NewErrorf(1002, "User %d not found", userID)
return aurelion.Error(ctx, err)
```

### Request/Trace ID

Request ID and Trace ID are automatically generated and included in all requests. You can retrieve them:

```go
// Get request ID (auto-generated UUID v7 or from X-Request-ID header)
requestID := aurelion.GetRequestID(ctx)

// Get trace ID (auto-generated UUID v7 or from X-Trace-ID/X-B3-TraceId/traceparent header)
traceID := aurelion.GetTraceID(ctx)

// Convenience method
requestID := ctx.RequestID()

// Use in logging
logger.Infow("processing request",
	"request_id", requestID,
	"trace_id", traceID,
)
```

### Context and Locals

```go
// Store values in request context (Locals)
ctx.Locals("user_id", "123")
ctx.Locals("lang", "vi")

// Retrieve values
userID := ctx.Locals("user_id")
lang := ctx.Locals("lang")

// Get all Locals
allLocals := ctx.GetAllLocals()
for key, value := range allLocals {
	logger.Info("local value", "key", key, "value", value)
}

// Convert to standard context.Context
stdCtx := ctx.Context()
// Use with contextutil package or standard context APIs
```

### CORS Configuration

```go
config := &aurelion.Config{
	ServiceName: "My API",
	Port:        8080,
	EnableCORS:  true,
	CORS: &aurelion.CORSConfig{
		AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Request-ID", "X-Trace-ID"},
		MaxAge:           3600,
	},
}
```

### CSRF Protection

```go
shutdownTimeout := 30 * time.Second
csrfExpiration := 24 * time.Hour

config := &aurelion.Config{
	ServiceName:             "My API",
	Port:                    8080,
	GracefulShutdownTimeout: &shutdownTimeout,
	EnableCSRF:              true,
	CSRF: &aurelion.CSRFConfig{
		KeyLookup:         "header:X-Csrf-Token",
		CookieName:        "csrf_token",
		CookiePath:        "/",
		CookieSecure:      true,
		CookieHTTPOnly:    true,
		CookieSameSite:    "Strict",
		CookieSessionOnly: true,
		Expiration:        &csrfExpiration,
	},
}
```

### Header to Locals Middleware

Automatically parse request headers and store them in Locals:

```go
server, err := aurelion.NewHttpServer(
	config,
	aurelion.WithGlobalMiddleware(aurelion.DefaultHeaderToLocalsMiddleware()),
)

// All headers are now available in Locals (lowercase keys)
// Example: Accept-Language header -> "accept-language" in Locals
lang := ctx.Locals("accept-language")

// Or with prefix and filter
customMiddleware := aurelion.HeaderToLocalsMiddleware("header_", func(key string) bool {
	return key == "uid" || key == "accept-language"
})
```

## API Reference

### Server Functions

- **`NewHttpServer(config *Config, options ...ServerOption) (*HttpServer, error)`** - Create a new HTTP server
- **`DefaultConfig() *Config`** - Get default configuration

### Server Methods

- **`Start() error`** - Start the server (blocks until shutdown)
- **`Shutdown(ctx context.Context) error`** - Gracefully shutdown the server
- **`AddRoutes(routes ...interface{}) *HttpServer`** - Add routes to the server
- **`AddGroupRoutes(groups ...interface{}) *HttpServer`** - Add group routes to the server

### Route Builder

- **`NewRoute(path string) *RouteBuilder`** - Create a new route builder
- **`NewGroupRoute(prefix string) *GroupRouteBuilder`** - Create a new group route builder

### Route Builder Methods

- **`GET()`, `POST()`, `PUT()`, `PATCH()`, `DELETE()`, `HEAD()`, `OPTIONS()`** - Set HTTP method
- **`Method(method Method)`** - Set HTTP method explicitly
- **`Path(path string)`** - Set route path
- **`Handler(handler Handler)`** - Set route handler
- **`Middleware(middleware ...Middleware)`** - Add middleware to route
- **`Protected()`** - Mark route as requiring authentication
- **`Permissions(permissions ...string)`** - Set required permissions for route
- **`CORS(corsConfig *CORSConfig)`** - Set per-route CORS configuration
- **`Build() *Route`** - Build the route

### Group Route Builder Methods

- **`Middleware(middleware ...Middleware)`** - Add middleware to group
- **`Protected()`** - Mark all routes in group as requiring authentication
- **`Route(route interface{})`** - Add single route to group
- **`Routes(routes ...interface{})`** - Add multiple routes to group
- **`Build() *GroupRoute`** - Build the group route

### Response Functions

- **`OK(ctx Context, message string, data ...interface{}) error`** - Send 200 OK response
- **`BadRequest(ctx Context, message string) error`** - Send 400 Bad Request response
- **`Unauthorized(ctx Context, message string) error`** - Send 401 Unauthorized response
- **`Forbidden(ctx Context, message string) error`** - Send 403 Forbidden response
- **`NotFound(ctx Context, message string) error`** - Send 404 Not Found response
- **`InternalServerError(ctx Context, message string) error`** - Send 500 Internal Server Error response
- **`Error(ctx Context, err error) error`** - Send error response (handles BusinessError)
- **`HealthCheck(ctx Context) error`** - Send health check response

### Error Functions

- **`NewError(code int, message string) *BusinessError`** - Create business error
- **`NewErrorf(code int, format string, args ...interface{}) *BusinessError`** - Create formatted business error

### Request ID Functions

- **`GetRequestID(ctx Context) string`** - Get request ID from context
- **`GetTraceID(ctx Context) string`** - Get trace ID from context

### Middleware Functions

- **`HeaderToLocalsMiddleware(prefix string, filter func(string) bool) Middleware`** - Parse headers to Locals
- **`DefaultHeaderToLocalsMiddleware() Middleware`** - Parse all headers to Locals (no prefix, no filter)

### Server Options

- **`WithGlobalMiddleware(middlewares ...Middleware) ServerOption`** - Add global middleware
- **`WithPanicRecover(middleware Middleware) ServerOption`** - Set panic recovery middleware
- **`WithAuthentication(middleware Middleware) ServerOption`** - Set authentication middleware
- **`WithAuthorization(checker AuthorizationFunc) ServerOption`** - Set authorization checker
- **`WithRateLimiter(middleware Middleware) ServerOption`** - Set custom rate limiter middleware

### Context Methods

The `Context` interface provides methods for accessing request data and sending responses. See the package documentation for the full API.

## Best Practices

### 1. Initialize Logger First

Always initialize the logger before creating the server:

```go
undo := logger.InitDefaultLogger()
defer undo()

server, err := aurelion.NewHttpServer(config)
```

### 2. Use Structured Logging

Use structured logging with request/trace IDs for better observability:

```go
requestID := aurelion.GetRequestID(ctx)
traceID := aurelion.GetTraceID(ctx)

logger.Infow("processing request",
	"request_id", requestID,
	"trace_id", traceID,
	"method", ctx.Method(),
	"path", ctx.Path(),
)
```

### 3. Validate Input Early

Validate and return errors early in handlers:

```go
handler := func(ctx aurelion.Context) error {
	var req CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return aurelion.BadRequest(ctx, "Invalid request body")
	}

	if req.Name == "" {
		return aurelion.BadRequest(ctx, "Name is required")
	}

	// Process request...
}
```

### 4. Use Business Errors

Use business errors for application-level errors:

```go
user, err := getUserByID(userID)
if err != nil {
	if errors.Is(err, ErrUserNotFound) {
		return aurelion.Error(ctx, aurelion.NewError(1001, "User not found"))
	}
	return aurelion.InternalServerError(ctx, "Failed to get user")
}
```

### 5. Use Group Routes for API Versioning

Organize routes using group routes:

```go
server.AddGroupRoutes(
	aurelion.NewGroupRoute("/api/v1").Routes(
		aurelion.NewRoute("/users").GET().Handler(getUsersHandler),
		aurelion.NewRoute("/posts").GET().Handler(getPostsHandler),
	),
	aurelion.NewGroupRoute("/api/v2").Routes(
		aurelion.NewRoute("/users").GET().Handler(getUsersV2Handler),
	),
)
```

### 6. Graceful Shutdown

Always configure graceful shutdown timeout:

```go
shutdownTimeout := 30 * time.Second
config := &aurelion.Config{
	ServiceName:             "My API",
	Port:                    8080,
	GracefulShutdownTimeout: &shutdownTimeout,
}
```

### 7. Use Protected Routes

Mark routes that require authentication:

```go
server.AddRoutes(
	aurelion.NewRoute("/profile").
		GET().
		Protected(). // Requires authentication
		Handler(getProfileHandler),
)
```

### 8. Use Permissions for Fine-Grained Authorization

Use permissions for route-level authorization:

```go
server.AddRoutes(
	aurelion.NewRoute("/admin/users").
		GET().
		Protected().
		Permissions("read:users", "admin").
		Handler(getUsersHandler),
)
```

## Features

### Built-in Middleware

- **Helmet** - Security headers (always enabled)
- **Rate Limiting** - Default 500 req/min per IP (configurable)
- **Compression** - Automatic response compression (always enabled)
- **Panic Recovery** - Graceful panic recovery (always enabled)
- **Request ID** - Automatic request ID generation (always enabled)
- **Trace ID** - Automatic trace ID generation (always enabled)
- **Request/Response Logging** - Structured logging with configurable verbosity
- **CORS** - Configurable CORS support
- **CSRF Protection** - Configurable CSRF protection

### Request ID and Trace ID

- Automatically generated UUID v7 (with v4 fallback)
- Can be provided via headers (`X-Request-ID`, `X-Trace-ID`, `X-B3-TraceId`, `traceparent`)
- Included in response headers
- Available via `GetRequestID()` and `GetTraceID()` functions
- Stored in Locals and automatically merged into `context.Context`

### Context Integration

- All Locals values are automatically merged into standard `context.Context`
- Compatible with `contextutil` package for type-safe context value access
- Supports custom keys without hard-coding

### Response Format

All responses follow a standard format:

```json
{
	"success": true,
	"code": 200,
	"message": "Operation successful",
	"data": {...},
	"timestamp": 1696234567890
}
```

Errors follow the same format with `success: false` and appropriate error code.
