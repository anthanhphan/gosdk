# Orianna Package

A production-ready, self-built HTTP framework for Go applications. Provides a clean, framework-agnostic API with fluent route builders, middleware composition, typed request binding with validation, structured error responses, health checks, lifecycle hooks, and built-in security features. Built on top of Fiber for high performance while maintaining a testable, platform-independent architecture.

## Features

- **Framework-agnostic API** - Clean `Context` interface that's testable and not tied to any HTTP framework
- **Fluent route builders** - Type-safe route and group builders with method chaining
- **Route shortcuts** - Direct `server.GET()`, `server.POST()` etc. for quick route registration
- **Protected routes** - Built-in authentication and authorization support with permissions
- **Typed request binding** - Generic `Bind[T]()` / `MustBind[T]()` with automatic validation
- **TypedHandler** - Automatic request/response marshaling with `TypedHandler[Req, Resp]()`
- **Middleware composition** - `Chain`, `Optional`, `SkipForPaths`, `Before`, `After`, `Recover`, `Timeout`
- **Structured responses** - Standardized `SuccessResponse` and `ErrorResponse` with request IDs
- **Health checks** - Worker-pool-based health manager with custom and HTTP checkers
- **Lifecycle hooks** - OnRequest, OnResponse, OnError, OnPanic, OnShutdown, OnServerStart
- **Built-in validation** - Struct tag-based validation with custom rules support
- **CORS/CSRF protection** - Configurable cross-origin and CSRF middleware
- **Metrics integration** - Prometheus-compatible request metrics
- **Duplicate route detection** - Catches conflicting route registrations at startup
- **Graceful shutdown** - Configurable shutdown timeout with hook support

## Installation

```bash
go get github.com/anthanhphan/gosdk/orianna
```

## Quick Start

```go
package main

import (
	"github.com/anthanhphan/gosdk/orianna"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
)

func main() {
	srv, err := orianna.NewServer(&configuration.Config{
		ServiceName: "my-api",
		Port:        8080,
	})
	if err != nil {
		panic(err)
	}

	// Register routes using shortcuts
	srv.GET("/", func(ctx orianna.Context) error {
		return ctx.OK(orianna.Map{"message": "Hello, World!"})
	})

	srv.GET("/users/:id", func(ctx orianna.Context) error {
		id, err := orianna.GetParamInt(ctx, "id")
		if err != nil {
			return ctx.BadRequestMsg("Invalid user ID")
		}
		return ctx.OK(orianna.Map{"id": id, "name": "Alice"})
	})

	// Start server (blocks until shutdown)
	if err := srv.Run(); err != nil {
		panic(err)
	}
}
```

## Configuration

### Config Structure

```go
type Config struct {
	ServiceName              string         // Service name (required, used in logs/metrics)
	Version                  string         // Service version (optional)
	Port                     int            // HTTP port (required, 0-65535)
	ReadTimeout              *time.Duration // Max duration for reading request (default: 30s)
	WriteTimeout             *time.Duration // Max duration for writing response (default: 30s)
	IdleTimeout              *time.Duration // Max idle time for keep-alive (default: 120s)
	GracefulShutdownTimeout  *time.Duration // Max wait for graceful shutdown (default: 30s)
	MaxBodySize              int            // Max request body in bytes (default: 4MB)
	MaxConcurrentConnections int            // Max concurrent connections (default: 256K)
	EnableCORS               bool           // Enable CORS middleware
	EnableCSRF               bool           // Enable CSRF protection
	CORS                     *CORSConfig    // CORS configuration
	CSRF                     *CSRFConfig    // CSRF configuration
	VerboseLogging           bool           // Enable detailed request/response logging
	VerboseLoggingSkipPaths  []string       // Paths to exclude from logging
	UseProperHTTPStatus      bool           // Use proper HTTP status codes for errors
}
```

### Server Options

Server behavior is customized through functional options passed to `NewServer`:

- **`WithGlobalMiddleware(middlewares ...Middleware)`** - Add middleware applied to all routes
- **`WithPanicRecover(middleware Middleware)`** - Set panic recovery middleware
- **`WithAuthentication(middleware Middleware)`** - Set authentication middleware
- **`WithAuthorization(checker func(Context, []string) error)`** - Set authorization checker
- **`WithRateLimiter(middleware Middleware)`** - Set rate limiter middleware
- **`WithHooks(hooks *RequestHooks)`** - Set lifecycle hooks
- **`WithMiddlewareConfig(config *MiddlewareConfig)`** - Configure built-in middleware
- **`WithHealthManager(manager HealthCheckManager)`** - Set health check manager
- **`WithShutdownManager(manager ShutdownManager)`** - Set shutdown manager
- **`WithHealthChecker(checker HealthChecker)`** - Add a health checker
- **`WithMetrics(client metrics.Client)`** - Enable metrics collection

### MiddlewareConfig

Controls which built-in middleware are enabled:

```go
type MiddlewareConfig struct {
	DisableHelmet      bool // Security headers (X-Frame-Options, etc.)
	DisableRateLimit   bool // Rate limiting per IP
	DisableCompression bool // Response compression (gzip/deflate)
	DisableRecovery    bool // Panic recovery
	DisableRequestID   bool // Request ID generation
	DisableTraceID     bool // Distributed tracing ID
	DisableLogging     bool // Request/response logging
}
```

### CORS Configuration

```go
config := &configuration.Config{
	EnableCORS: true,
	CORS: &configuration.CORSConfig{
		AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	},
}
```

## Usage Examples

### Route Registration (Shortcuts)

The simplest way to register routes:

```go
srv.GET("/users", listUsersHandler)
srv.POST("/users", createUserHandler)
srv.PUT("/users/:id", updateUserHandler)
srv.PATCH("/users/:id", patchUserHandler)
srv.DELETE("/users/:id", deleteUserHandler)
srv.HEAD("/users/:id", checkUserHandler)
srv.OPTIONS("/api/users", corsHandler)
```

### Route Registration (Builder Pattern)

For more control, use the fluent route builder:

```go
route := orianna.NewRoute("/users/:id").
	GET().
	Handler(getUserHandler).
	Middleware(loggingMiddleware, cachingMiddleware).
	Build()

srv.RegisterRoutes(*route)

// Multi-method route
route := orianna.NewRoute("/users/:id").
	Methods(orianna.GET, orianna.HEAD).
	Handler(getUserHandler).
	Build()

srv.RegisterRoutes(*route)
```

### Route Groups

Group related routes with shared prefix and middleware:

```go
apiV1 := orianna.NewGroupRoute("/api/v1").
	Middleware(apiKeyMiddleware).
	GET("/status", statusHandler).
	GET("/version", versionHandler).
	POST("/users", createUserHandler).
	PUT("/users/:id", updateUserHandler).
	DELETE("/users/:id", deleteUserHandler).
	Build()

srv.RegisterGroup(*apiV1)

// Nested groups
api := orianna.NewGroupRoute("/api").
	Group(orianna.NewGroupRoute("/v1").
		GET("/users", listUsersV1).
		Build()).
	Group(orianna.NewGroupRoute("/v2").
		GET("/users", listUsersV2).
		Build()).
	Build()

srv.RegisterGroup(*api)
```

### Protected Routes

Routes requiring authentication and authorization:

```go
// Authentication only
srv.Protected().GET("/profile", profileHandler)
srv.Protected().PUT("/profile", updateProfileHandler)

// Authentication + authorization permissions
srv.Protected().
	WithPermissions("admin:read").
	GET("/admin/stats", adminStatsHandler)

srv.Protected().
	WithPermissions("admin:write").
	Middleware(auditMiddleware).
	POST("/admin/settings", adminSettingsHandler)
```

### Request Binding

#### Generic Bind

```go
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=1,max=150"`
	Role  string `json:"role" validate:"oneof=user admin moderator"`
}

func createUserHandler(ctx orianna.Context) error {
	// Bind with automatic validation (default: body source, validation enabled)
	req, err := orianna.Bind[CreateUserRequest](ctx)
	if err != nil {
		return err
	}
	// Use req...
}
```

#### MustBind (Auto Error Response)

```go
func updateUserHandler(ctx orianna.Context) error {
	req, ok := orianna.MustBind[UpdateUserRequest](ctx)
	if !ok {
		return nil // Error response already sent
	}
	// Use req...
}
```

#### Shorthand Binding

```go
// Bind from specific source
user, err := orianna.BindBody[CreateUserRequest](ctx, true)    // body + validate
query, err := orianna.BindQuery[SearchParams](ctx, false)       // query, no validate
params, err := orianna.BindParams[RouteParams](ctx, true)       // route params + validate

// Custom bind options
req, err := orianna.Bind[Request](ctx, orianna.BindOptions{
	Source:   orianna.BindSourceQuery,
	Validate: true,
})
```

### TypedHandler (Automatic Marshaling)

Eliminates boilerplate for request parsing and response sending:

```go
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type CreateUserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var createUser = orianna.TypedHandler(
	func(ctx orianna.Context, req CreateUserRequest) (CreateUserResponse, error) {
		// Request is already parsed and validated
		// Return response struct - automatically wrapped in SuccessResponse
		return CreateUserResponse{
			ID:    123,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	},
)

srv.POST("/users", createUser)
```

### Response Helpers

#### Shorthand Responses

```go
func handler(ctx orianna.Context) error {
	// Success responses
	return ctx.OK(data)                           // 200 OK
	return ctx.Created(data)                      // 201 Created
	return ctx.NoContent()                        // 204 No Content

	// Error responses
	return ctx.BadRequestMsg("Invalid input")     // 400 Bad Request
	return ctx.UnauthorizedMsg("Not logged in")   // 401 Unauthorized
	return ctx.ForbiddenMsg("Access denied")      // 403 Forbidden
	return ctx.NotFoundMsg("User not found")      // 404 Not Found
	return ctx.InternalErrorMsg("Server error")   // 500 Internal Server Error
}
```

#### Structured Responses

```go
// Success response
resp := orianna.NewSuccessResponse(orianna.StatusOK, "User created", user)
return orianna.SendSuccess(ctx, resp)

// Error response with details
errResp := orianna.NewErrorResponse("INSUFFICIENT_BALANCE", orianna.StatusBadRequest, "Balance too low").
	WithDetails("required", 100).
	WithDetails("current", 42).
	WithInternalMsg("User %s tried to withdraw %d", userID, amount).
	WithCause(originalErr)
return orianna.SendError(ctx, errResp)
```

#### Error Handling Utilities

```go
// Check error type
if orianna.IsErrorCode(err, "NOT_FOUND") {
	// Handle specific error
}

// Wrap errors with context
return orianna.WrapError(err, "failed to create user")
return orianna.WrapErrorf(err, "failed to process user %d", userID)

// Auto-handle ErrorResponse errors
if orianna.HandleError(ctx, err) {
	return nil // ErrorResponse was sent
}

// Validate and respond
if ok, err := orianna.ValidateAndRespond(ctx, request); !ok {
	return err // Validation error response already sent
}
```

### Query & Parameter Helpers

```go
// Route parameters
id, err := orianna.GetParamInt(ctx, "id")
id64, err := orianna.GetParamInt64(ctx, "id")
uuid, err := orianna.GetParamUUID(ctx, "id")

// Query parameters (with defaults)
page := orianna.GetQueryInt(ctx, "page", 1)
limit := orianna.GetQueryInt64(ctx, "limit", 10)
active := orianna.GetQueryBool(ctx, "active", true)    // accepts: true/1/yes/on, false/0/no/off
sortBy := orianna.GetQueryString(ctx, "sort", "created_at")
```

### Middleware

#### Custom Middleware

```go
func timingMiddleware(ctx orianna.Context) error {
	start := time.Now()
	err := ctx.Next()
	duration := time.Since(start)
	logger.Infow("Request timing", "path", ctx.Path(), "duration_ms", duration.Milliseconds())
	return err
}

// Register per-route
srv.GET("/users", listUsersHandler, timingMiddleware)

// Register globally
srv.Use(timingMiddleware)
```

#### Middleware Composition

```go
// Chain multiple middleware into one
combined := orianna.Chain(authMiddleware, loggingMiddleware, rateLimitMiddleware)
srv.GET("/api/data", handler, combined)

// Conditional middleware
srv.Use(orianna.Optional(func(ctx orianna.Context) bool {
	return ctx.Get("X-Debug") != ""
}, debugMiddleware))

// Method-specific
srv.Use(orianna.OnlyForMethods(auditMiddleware, "POST", "PUT", "DELETE"))

// Skip specific paths
srv.Use(orianna.SkipForPaths(loggingMiddleware, "/health", "/metrics"))
srv.Use(orianna.SkipForPathPrefixes(authMiddleware, "/public"))

// Before/After hooks
withBefore := orianna.Before(myMiddleware, func(ctx orianna.Context) {
	ctx.Locals("start_time", time.Now())
})

withAfter := orianna.After(myMiddleware, func(ctx orianna.Context, err error) {
	logger.Infow("Request completed", "error", err)
})

// Panic recovery wrapper
safe := orianna.Recover(unsafeMiddleware)

// Timeout wrapper
withTimeout := orianna.Timeout(slowMiddleware, 5*time.Second)
```

### Lifecycle Hooks

```go
hooks := orianna.NewRequestHooks()

hooks.AddOnRequest(func(ctx orianna.Context) {
	logger.Debugw("Request received", "method", ctx.Method(), "path", ctx.Path())
})

hooks.AddOnResponse(func(ctx orianna.Context, status int, latency time.Duration) {
	logger.Infow("Response sent", "status", status, "latency_ms", latency.Milliseconds())
})

hooks.AddOnError(func(ctx orianna.Context, err error) {
	logger.Errorw("Request error", "path", ctx.Path(), "error", err.Error())
})

hooks.AddOnPanic(func(ctx orianna.Context, recovered any, stack []byte) {
	logger.Errorw("Panic recovered", "panic", recovered)
})

hooks.AddOnShutdown(func() {
	logger.Infow("Server shutting down, cleaning up resources...")
})

hooks.AddOnServerStart(func(server any) error {
	logger.Infow("Server starting up")
	return nil
})

srv, _ := orianna.NewServer(config, orianna.WithHooks(hooks))
```

### Health Checks

```go
srv, _ := orianna.NewServer(config,
	// Custom health checker
	orianna.WithHealthChecker(
		orianna.NewCustomChecker("database", func(ctx context.Context) orianna.HealthCheck {
			return orianna.HealthCheck{
				Name:    "database",
				Status:  orianna.HealthStatusHealthy,
				Message: "Connected to PostgreSQL",
			}
		}),
	),
	// HTTP endpoint health checker
	orianna.WithHealthChecker(
		orianna.NewHTTPChecker("http://redis:6379/ping", "cache", 5*time.Second),
	),
)

// Access health manager directly
report := srv.GetHealthManager().Check(ctx)
// report.Status: "healthy" | "unhealthy" | "degraded"
// report.Checks: map of individual check results
```

#### Health Status Values

- **`HealthStatusHealthy`** - Component is operating normally
- **`HealthStatusUnhealthy`** - Component is down or non-functional
- **`HealthStatusDegraded`** - Component is working but with reduced performance

### Authentication & Authorization

```go
// Authentication middleware
func authMiddleware(ctx orianna.Context) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return orianna.NewErrorResponse("UNAUTHORIZED", orianna.StatusUnauthorized, "Missing token")
	}

	// Validate token and set user context
	userID, err := validateToken(token)
	if err != nil {
		return orianna.NewErrorResponse("UNAUTHORIZED", orianna.StatusUnauthorized, "Invalid token")
	}

	ctx.Locals("user_id", userID)
	return ctx.Next()
}

// Authorization checker
func authzChecker(ctx orianna.Context, permissions []string) error {
	role := ctx.Locals("role").(string)
	for _, perm := range permissions {
		if !hasPermission(role, perm) {
			return orianna.NewErrorResponse("FORBIDDEN", orianna.StatusForbidden,
				fmt.Sprintf("Missing permission: %s", perm))
		}
	}
	return nil
}

srv, _ := orianna.NewServer(config,
	orianna.WithAuthentication(authMiddleware),
	orianna.WithAuthorization(authzChecker),
)

// Use protected routes
srv.Protected().GET("/profile", profileHandler)
srv.Protected().WithPermissions("admin:write").POST("/admin/settings", handler)
```

### Full Production Example

```go
package main

import (
	"context"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
)

func main() {
	undo := logger.InitProductionLogger()
	defer undo()

	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	shutdownTimeout := 30 * time.Second

	config := &configuration.Config{
		ServiceName:         "user-api",
		Version:             "1.0.0",
		Port:                8080,
		UseProperHTTPStatus: true,
		ReadTimeout:         &readTimeout,
		WriteTimeout:        &writeTimeout,
		GracefulShutdownTimeout: &shutdownTimeout,
		EnableCORS: true,
		CORS: &configuration.CORSConfig{
			AllowOrigins: []string{"https://app.example.com"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
		},
	}

	srv, err := orianna.NewServer(config,
		orianna.WithAuthentication(authMiddleware),
		orianna.WithAuthorization(authzChecker),
		orianna.WithMetrics(metrics.NewClient("user-api")),
		orianna.WithHealthChecker(
			orianna.NewCustomChecker("database", dbHealthCheck),
		),
	)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Public routes
	srv.GET("/health", healthHandler)
	srv.GET("/users", listUsersHandler)
	srv.GET("/users/:id", getUserHandler)
	srv.POST("/users", createUserHandler)

	// Protected routes
	srv.Protected().GET("/profile", profileHandler)
	srv.Protected().WithPermissions("admin:read").GET("/admin/stats", adminStatsHandler)

	// Route groups
	apiV1 := orianna.NewGroupRoute("/api/v1").
		GET("/status", statusHandler).
		GET("/version", versionHandler).
		Build()
	srv.RegisterGroup(*apiV1)

	if err := srv.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
```

## API Reference

### Server

- **`NewServer(conf *Config, options ...ServerOption) (*Server, error)`** - Create a new server
- **`(*Server).Start() error`** - Start the HTTP server
- **`(*Server).Run() error`** - Alias for Start
- **`(*Server).Shutdown(ctx context.Context) error`** - Graceful shutdown
- **`(*Server).RegisterRoutes(routes ...Route) error`** - Register routes
- **`(*Server).RegisterGroup(group RouteGroup) error`** - Register a route group
- **`(*Server).Use(middleware ...Middleware)`** - Add global middleware
- **`(*Server).Protected() *RouteShortcuts`** - Start protected route chain
- **`(*Server).GetHealthManager() HealthCheckManager`** - Get health manager
- **`(*Server).GetShutdownManager() ShutdownManager`** - Get shutdown manager

### Route Shortcuts

- **`(*Server).GET(path, handler, middleware...) error`**
- **`(*Server).POST(path, handler, middleware...) error`**
- **`(*Server).PUT(path, handler, middleware...) error`**
- **`(*Server).PATCH(path, handler, middleware...) error`**
- **`(*Server).DELETE(path, handler, middleware...) error`**
- **`(*Server).HEAD(path, handler, middleware...) error`**
- **`(*Server).OPTIONS(path, handler, middleware...) error`**

### Route Builder

- **`NewRoute(path string) *RouteBuilder`** - Create route builder
- **`(*RouteBuilder).GET() / .POST() / .PUT() / .PATCH() / .DELETE()`** - Set HTTP method
- **`(*RouteBuilder).Methods(methods ...Method)`** - Set multiple methods
- **`(*RouteBuilder).Handler(handler Handler)`** - Set handler
- **`(*RouteBuilder).Middleware(middleware ...Middleware)`** - Add middleware
- **`(*RouteBuilder).Protected()`** - Mark as requiring authentication
- **`(*RouteBuilder).Permissions(permissions ...string)`** - Set required permissions
- **`(*RouteBuilder).CORS(corsConfig *CORSConfig)`** - Set per-route CORS
- **`(*RouteBuilder).Build() *Route`** - Build the route

### Group Route Builder

- **`NewGroupRoute(prefix string) *GroupRouteBuilder`** - Create group builder
- **`(*GroupRouteBuilder).GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS(path, handler, middleware...)`** - Add route
- **`(*GroupRouteBuilder).Route(route *Route)`** - Add a built route
- **`(*GroupRouteBuilder).Routes(routes ...*Route)`** - Add multiple built routes
- **`(*GroupRouteBuilder).Group(group *RouteGroup)`** - Add nested group
- **`(*GroupRouteBuilder).Middleware(middleware ...Middleware)`** - Add group middleware
- **`(*GroupRouteBuilder).Protected()`** - Mark all routes as protected
- **`(*GroupRouteBuilder).Build() *RouteGroup`** - Build the group

### Request Binding

- **`Bind[T](ctx Context, opts ...BindOptions) (T, error)`** - Parse request with optional validation
- **`MustBind[T](ctx Context, opts ...BindOptions) (T, bool)`** - Parse request, auto-send error on failure
- **`BindBody[T](ctx Context, validate bool) (T, error)`** - Bind from request body
- **`BindQuery[T](ctx Context, validate bool) (T, error)`** - Bind from query parameters
- **`BindParams[T](ctx Context, validate bool) (T, error)`** - Bind from route parameters

### Handler Helpers

- **`TypedHandler[Req, Resp](fn func(Context, Req) (Resp, error)) Handler`** - Auto request/response marshaling
- **`SimpleHandler(fn func(Context) error) Handler`** - Pass-through handler wrapper

### Response Functions

- **`NewSuccessResponse(httpStatus int, message string, data any) *SuccessResponse`**
- **`NewErrorResponse(code string, httpStatus int, message string) *ErrorResponse`**
- **`SendSuccess(ctx Context, resp *SuccessResponse) error`**
- **`SendError(ctx Context, err *ErrorResponse) error`**
- **`HandleError(ctx Context, err error) bool`**
- **`ValidateAndRespond(ctx Context, v any) (bool, error)`**
- **`IsErrorCode(err error, code string) bool`**
- **`WrapError(err error, message string) error`**
- **`WrapErrorf(err error, format string, args ...any) error`**

### Query & Parameter Helpers

- **`GetParamInt(ctx Context, key string) (int, error)`**
- **`GetParamInt64(ctx Context, key string) (int64, error)`**
- **`GetParamUUID(ctx Context, key string) (string, error)`**
- **`GetQueryInt(ctx Context, key string, defaultValue int) int`**
- **`GetQueryInt64(ctx Context, key string, defaultValue int64) int64`**
- **`GetQueryBool(ctx Context, key string, defaultValue bool) bool`**
- **`GetQueryString(ctx Context, key string, defaultValue string) string`**

### Middleware Functions

- **`Chain(middlewares ...Middleware) Middleware`** - Combine multiple middleware
- **`Optional(condition func(Context) bool, middleware Middleware) Middleware`** - Conditional middleware
- **`OnlyForMethods(middleware Middleware, methods ...string) Middleware`** - Method filter
- **`SkipForPaths(middleware Middleware, paths ...string) Middleware`** - Exact path skip
- **`SkipForPathPrefixes(middleware Middleware, prefixes ...string) Middleware`** - Prefix path skip
- **`Before(middleware Middleware, beforeFunc func(Context)) Middleware`** - Pre-execution hook
- **`After(middleware Middleware, afterFunc func(Context, error)) Middleware`** - Post-execution hook
- **`Recover(middleware Middleware) Middleware`** - Panic recovery wrapper
- **`Timeout(middleware Middleware, timeout time.Duration) Middleware`** - Timeout wrapper

### Health Checks

- **`NewHealthManager() *HealthManager`** - Create health manager
- **`NewCustomChecker(name string, checkFn func(context.Context) HealthCheck) *CustomChecker`** - Custom checker
- **`NewHTTPChecker(url, name string, timeout time.Duration) *HTTPChecker`** - HTTP endpoint checker

### Error Classification

- **`IsConfigError(err error) bool`** - Check for configuration errors
- **`IsRouteError(err error) bool`** - Check for route-related errors
- **`IsServerError(err error) bool`** - Check for server-related errors

## Context Interface

The `Context` interface provides a framework-agnostic API composed of smaller interfaces following the Interface Segregation Principle:

| Interface           | Methods                                                       |
|---------------------|---------------------------------------------------------------|
| `RequestInfo`       | `Method()`, `Path()`, `RoutePath()`, `IP()`, `Hostname()` etc. |
| `HeaderManager`     | `Get()`, `Set()`, `Append()`                                   |
| `ParamGetter`       | `Params()`, `AllParams()`, `ParamsParser()`                    |
| `QueryGetter`       | `Query()`, `AllQueries()`, `QueryParser()`                     |
| `BodyReader`        | `Body()`, `BodyParser()`                                       |
| `CookieManager`     | `Cookies()`, `Cookie()`, `ClearCookie()`                       |
| `ResponseWriter`    | `Status()`, `JSON()`, `XML()`, `SendString()`, `Redirect()`   |
| `ContentNegotiator` | `Accepts()`, `AcceptsCharsets()`, `AcceptsEncodings()`         |
| `RequestState`      | `Fresh()`, `Stale()`, `XHR()`                                 |
| `LocalsStorage`     | `Locals()`, `GetAllLocals()`                                   |
| `ShorthandResponder`| `OK()`, `Created()`, `NoContent()`, `BadRequestMsg()` etc.    |

Additional methods: `Next()`, `Context()`, `IsMethod()`, `RequestID()`, `UseProperHTTPStatus()`

## Validation

Struct field validation uses the `validate` tag:

```go
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=1,max=150"`
	Role  string `json:"role" validate:"oneof=user admin moderator"`
}
```

### Supported Rules

| Rule         | Description                            | Example                       |
|--------------|----------------------------------------|-------------------------------|
| `required`   | Field must not be zero value           | `validate:"required"`         |
| `min`        | Minimum value/length                   | `validate:"min=3"`            |
| `max`        | Maximum value/length                   | `validate:"max=100"`          |
| `len`        | Exact length                           | `validate:"len=10"`           |
| `email`      | Valid email format                     | `validate:"email"`            |
| `url`        | Valid URL format                       | `validate:"url"`              |
| `numeric`    | Numeric characters only                | `validate:"numeric"`          |
| `alpha`      | Alphabetic characters only             | `validate:"alpha"`            |
| `oneof`      | Value must be one of listed values     | `validate:"oneof=a b c"`      |
| `gt`         | Greater than                           | `validate:"gt=0"`             |
| `gte`        | Greater than or equal to               | `validate:"gte=1"`            |
| `lt`         | Less than                              | `validate:"lt=100"`           |
| `lte`        | Less than or equal to                  | `validate:"lte=99"`           |

## Best Practices

### 1. Use Route Shortcuts for Simple Routes

```go
// Preferred for simple routes
srv.GET("/users", listUsersHandler)
srv.POST("/users", createUserHandler)

// Use builder for complex routes (per-route CORS, multiple methods, etc.)
route := orianna.NewRoute("/users").
	Methods(orianna.GET, orianna.HEAD).
	Handler(handler).
	CORS(corsConfig).
	Build()
```

### 2. Use TypedHandler for CRUD Operations

```go
// Eliminates boilerplate for request parsing and response sending
var createUser = orianna.TypedHandler(
	func(ctx orianna.Context, req CreateUserRequest) (CreateUserResponse, error) {
		user, err := service.Create(req)
		if err != nil {
			return CreateUserResponse{}, err
		}
		return CreateUserResponse{ID: user.ID}, nil
	},
)
```

### 3. Use MustBind for Manual Handlers

```go
func handler(ctx orianna.Context) error {
	req, ok := orianna.MustBind[Request](ctx)
	if !ok {
		return nil // Error already sent
	}
	// Use req...
}
```

### 4. Use Structured Error Responses

```go
// Good: Structured error with code and details
return orianna.NewErrorResponse("INSUFFICIENT_BALANCE", orianna.StatusBadRequest, "Balance too low").
	WithDetails("required", 100).
	WithDetails("current", 42)

// Less optimal: Generic error
return fmt.Errorf("balance too low")
```

### 5. Group Related Routes

```go
// Organize API with groups for shared prefix/middleware
apiV1 := orianna.NewGroupRoute("/api/v1").
	Middleware(versionMiddleware).
	GET("/users", listUsers).
	POST("/users", createUser).
	Build()
```

### 6. Configure Health Checks

```go
// Always configure health checks for production
srv, _ := orianna.NewServer(config,
	orianna.WithHealthChecker(orianna.NewCustomChecker("database", dbCheck)),
	orianna.WithHealthChecker(orianna.NewCustomChecker("cache", cacheCheck)),
	orianna.WithHealthChecker(orianna.NewHTTPChecker("http://auth:8080/health", "auth-service", 5*time.Second)),
)
```

### 7. Use Proper HTTP Status Codes

```go
// Recommended: Enable proper HTTP status codes
config := &configuration.Config{
	UseProperHTTPStatus: true, // 400, 404, 500 etc. instead of always 200
}
```

## Performance

- **High-performance Fiber backend** - Built on Fiber (fasthttp) for maximum throughput
- **Struct tag caching** - Validation metadata is parsed once and cached per type
- **Worker pool health checks** - Concurrent health checks with configurable pool size (default: 10)
- **Pre-calculated metric names** - Metric string concatenation happens once, not per request
- **Concurrent-safe** - All operations safe for concurrent use
- **Efficient middleware chain** - Minimal overhead per middleware layer
