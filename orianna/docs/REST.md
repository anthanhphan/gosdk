# Orianna — HTTP / REST

Complete guide to building HTTP services with Orianna. All import paths, types, defaults, and API signatures are sourced directly from the codebase.

## Table of Contents

- [Configuration](#configuration)
  - [Server Config](#server-config)
  - [Config Defaults](#config-defaults)
  - [CORS Config](#cors-config)
  - [CSRF Config](#csrf-config)
  - [Middleware Config](#middleware-config)
  - [Static File Config](#static-file-config)
  - [Server Options](#server-options)
- [Server Lifecycle](#server-lifecycle)
- [Routing](#routing)
  - [Route Shortcuts](#route-shortcuts)
  - [Route Builder](#route-builder)
  - [Route Groups](#route-groups)
  - [Protected Routes](#protected-routes)
- [Request Binding & Validation](#request-binding--validation)
  - [Bind](#bind)
  - [MustBind](#mustbind)
  - [Shorthand Binding](#shorthand-binding)
  - [TypedHandler](#typedhandler)
  - [Validation Rules](#validation-rules)
- [Response Helpers](#response-helpers)
  - [Shorthand Responses](#shorthand-responses)
  - [Structured Responses](#structured-responses)
  - [Error Utilities](#error-utilities)
  - [Query & Parameter Helpers](#query--parameter-helpers)
- [Middleware](#middleware)
  - [Custom Middleware](#custom-middleware)
  - [Middleware Composition](#middleware-composition)
- [Authentication & Authorization](#authentication--authorization)
- [Health Checks](#health-checks)
- [Lifecycle Hooks](#lifecycle-hooks)
- [HTTP Client](#http-client)
- [Context Interface (ISP)](#context-interface-isp)

---

## Configuration

### Server Config

```go
import "github.com/anthanhphan/gosdk/orianna/http/configuration"

readTimeout := 10 * time.Second
writeTimeout := 10 * time.Second
shutdownTimeout := 30 * time.Second

config := &configuration.Config{
    // ── Required ──
    ServiceName: "user-api",
    Port:        8080,

    // ── Optional ──
    Version: "v1.0.0",  // displayed in startup logs

    // ── Timeouts (pointer-based: nil = use default) ──
    ReadTimeout:             &readTimeout,   // default: 30s
    WriteTimeout:            &writeTimeout,  // default: 30s
    IdleTimeout:             nil,            // default: 120s
    GracefulShutdownTimeout: &shutdownTimeout,
    RequestTimeout:          nil,            // default: 30s (per-request deadline)

    // ── Limits ──
    MaxBodySize:              4 * 1024 * 1024, // default: 4MB
    MaxConcurrentConnections: 256 * 1024,      // default: 256K

    // ── Features ──
    VerboseLogging:          true,
    VerboseLoggingSkipPaths: []string{"/health", "/metrics"},
    UseProperHTTPStatus:     true,              // 400/404/500 instead of always 200
    SlowRequestThreshold:    2 * time.Second,   // auto-registers slow request detector

    // ── CORS ──
    EnableCORS: true,
    CORS: &configuration.CORSConfig{
        AllowOrigins:     []string{"https://app.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
        ExposeHeaders:    []string{"X-Total-Count"},
        MaxAge:           3600,
    },

    // ── CSRF ──
    EnableCSRF: true,
    CSRF: &configuration.CSRFConfig{
        // CookieSecure and CookieHTTPOnly default to true (secure-by-default via *bool)
    },

    // ── Compression ──
    CompressionLevel: nil, // default: 1 (BestSpeed)

    // ── Cache ──
    CacheExpiration: nil, // default: 1 minute

    // ── Prefork ──
    EnablePrefork: false, // WARNING: in-process state NOT shared across processes

    // ── Static files ──
    Static: &configuration.StaticFileConfig{
        Prefix: "/static",
        Root:   "./public",
        Browse: false,
        MaxAge: 3600,
    },
}
```

### Config Defaults

| Field | Default | Notes |
|-------|---------|-------|
| `Port` | `8080` | |
| `ReadTimeout` | `30s` | Nil = default |
| `WriteTimeout` | `30s` | Nil = default |
| `IdleTimeout` | `120s` | Nil = default |
| `GracefulShutdownTimeout` | `30s` | Nil = default |
| `RequestTimeout` | `30s` | Nil = default |
| `MaxBodySize` | `4MB` | 0 = default |
| `MaxConcurrentConnections` | `256K` | 0 = default |
| `CompressionLevel` | `1` (BestSpeed) | |
| `CacheExpiration` | `1 minute` | |
| `CookieSecure` (CSRF) | `true` | `*bool`, nil = true |
| `CookieHTTPOnly` (CSRF) | `true` | `*bool`, nil = true |
| `CookieSameSite` (CSRF) | `"Strict"` | |
| `CookieName` (CSRF) | `"csrf_token"` | |
| `CookiePath` (CSRF) | `"/"` | |
| `Expiration` (CSRF) | `24h` | |

### CORS Config

```go
CORS: &configuration.CORSConfig{
    AllowOrigins:     []string{"https://app.example.com"}, // required
    AllowMethods:     []string{"GET", "POST"},             // required
    AllowHeaders:     []string{"Authorization"},
    AllowCredentials: true,   // cannot be true when origin is "*"
    ExposeHeaders:    []string{"X-Total-Count"},
    MaxAge:           3600,   // preflight cache (seconds)
}
```

> **Validation:** If `AllowCredentials: true` and `AllowOrigins` contains `"*"`, validation fails at startup (browsers reject this).

### CSRF Config

```go
CSRF: &configuration.CSRFConfig{
    KeyLookup:      "header:X-CSRF-Token", // source:key format
    CookieName:     "csrf_token",
    CookiePath:     "/",
    CookieDomain:   "",                     // current domain
    CookieSameSite: "Strict",               // Strict | Lax | None
    CookieSecure:   nil,                    // nil = true (secure-by-default)
    CookieHTTPOnly: nil,                    // nil = true (secure-by-default)
    SingleUseToken: false,                  // true for critical ops (payment, delete)
    Expiration:     nil,                    // nil = 24h
}
```

### Middleware Config

Controls which built-in middleware is active. All default to `false` (enabled).

```go
server.WithMiddlewareConfig(&configuration.MiddlewareConfig{
    DisableHelmet:      false, // Security headers (X-Frame-Options, X-Content-Type-Options)
    DisableRecovery:    false, // Panic recovery — WARNING: never disable in production
    DisableRequestID:   false, // UUIDv7 request ID generation
    DisableTraceID:     false, // Distributed tracing ID (auto-disabled when OTel tracing active)
    DisableLogging:     false, // Request/response logging
    DisableRateLimit:   false, // Rate limiting
    DisableCompression: false, // Response compression (gzip/deflate)
    DisableTracing:     false, // OpenTelemetry tracing middleware
    DisableETag:        false, // ETag generation for cache validation
    DisableCache:       false, // Response caching
})
```

### Static File Config

```go
Static: &configuration.StaticFileConfig{
    Prefix: "/static", // URL prefix
    Root:   "./public", // filesystem directory
    Browse: false,      // directory listing
    MaxAge: 3600,       // Cache-Control max-age (seconds)
}
```

### Server Options

```go
import "github.com/anthanhphan/gosdk/orianna/http/server"

srv, err := server.NewServer(config,
    server.WithAuthentication(authMiddleware),
    server.WithAuthorization(authzChecker),
    server.WithGlobalMiddleware(loggingMW, metricsMW),
    server.WithPanicRecover(recoveryMW),
    server.WithRateLimiter(rateLimiterMW),
    server.WithHooks(hooks),
    server.WithMetrics(metricsClient),
    server.WithTracing(tracingClient),
    server.WithHealthManager(healthMgr),
    server.WithHealthChecker(dbChecker),
    server.WithShutdownManager(shutdownMgr),
    server.WithMiddlewareConfig(mwConfig),
    server.WithServerEngine(customEngine), // Strategy pattern: swap Fiber for custom engine
)
```

| Option | Description |
|--------|-------------|
| `WithAuthentication(mw)` | Set auth middleware for `Protected()` routes |
| `WithAuthorization(fn)` | Set permission checker `func(Context, []string) error` |
| `WithGlobalMiddleware(mws...)` | Add middleware to all routes |
| `WithPanicRecover(mw)` | Custom panic recovery middleware |
| `WithRateLimiter(mw)` | Custom rate limiter middleware |
| `WithHooks(hooks)` | Set lifecycle hooks |
| `WithMetrics(client)` | Enable Prometheus metrics + `/metrics` endpoint |
| `WithTracing(client)` | Enable OpenTelemetry tracing (auto-disables legacy traceID) |
| `WithHealthManager(mgr)` | Set custom health check manager |
| `WithHealthChecker(checker)` | Add health checker (auto-creates manager if nil) |
| `WithShutdownManager(mgr)` | Set custom shutdown manager |
| `WithMiddlewareConfig(cfg)` | Control built-in middleware toggles |
| `WithServerEngine(engine)` | Swap the underlying server engine (Strategy pattern) |

---

## Server Lifecycle

```go
// Create
srv, err := server.NewServer(config, opts...)

// Start (blocks until OS signal SIGTERM/SIGINT)
err = srv.Start()

// Programmatic shutdown (applies GracefulShutdownTimeout if no deadline in ctx)
err = srv.Shutdown(ctx)
```

**Shutdown sequence:**
1. Shutdown manager runs (if set)
2. `OnShutdown` hooks fire
3. Server adapter stops accepting new connections
4. In-flight requests complete (up to `GracefulShutdownTimeout`)

---

## Routing

### Route Shortcuts

```go
srv.GET("/users", listUsersHandler)
srv.POST("/users", createUserHandler)
srv.PUT("/users/:id", updateUserHandler)
srv.PATCH("/users/:id", patchUserHandler)
srv.DELETE("/users/:id", deleteUserHandler)
srv.HEAD("/users/:id", checkUserHandler)
srv.OPTIONS("/api/users", corsHandler)

// With per-route middleware
srv.GET("/users", listUsersHandler, cacheMiddleware, loggingMiddleware)
```

### Route Builder

```go
import "github.com/anthanhphan/gosdk/orianna/http/routing"

// Single method
route := routing.NewRoute("/users/:id").
    GET().
    Handler(getUserHandler).
    Middleware(cacheMiddleware).
    Build()
srv.RegisterRoutes(*route)

// Multiple methods
route := routing.NewRoute("/users/:id").
    Methods(core.GET, core.HEAD).
    Handler(getUserHandler).
    Build()
srv.RegisterRoutes(*route)

// Protected route with permissions and CORS
route := routing.NewRoute("/admin/settings").
    POST().
    Handler(adminHandler).
    Protected().
    Permissions("admin:write").
    CORS(corsConfig).
    Build()
```

### Route Groups

```go
apiV1 := routing.NewGroupRoute("/api/v1").
    Middleware(apiKeyMiddleware).
    GET("/status", statusHandler).
    GET("/version", versionHandler).
    POST("/users", createUserHandler).
    PUT("/users/:id", updateUserHandler).
    DELETE("/users/:id", deleteUserHandler).
    Build()

srv.RegisterGroup(*apiV1)

// Nested groups
api := routing.NewGroupRoute("/api").
    Group(routing.NewGroupRoute("/v1").
        GET("/users", listUsersV1).Build()).
    Group(routing.NewGroupRoute("/v2").
        GET("/users", listUsersV2).Build()).
    Build()

srv.RegisterGroup(*api)
```

> **Atomic Registration:** Routes are validated first, then registered. If any route in a batch fails validation, none are registered.

### Protected Routes

```go
// Authentication only
srv.Protected().GET("/profile", profileHandler)
srv.Protected().PUT("/profile", updateProfileHandler)

// Authentication + permissions
srv.Protected().
    WithPermissions("admin:read").
    GET("/admin/stats", adminStatsHandler)

// With additional middleware
srv.Protected().
    WithPermissions("admin:write").
    Middleware(auditMiddleware).
    POST("/admin/settings", adminSettingsHandler)
```

---

## Request Binding & Validation

### Bind

Returns parsed struct + error. You handle the error response.

```go
import "github.com/anthanhphan/gosdk/orianna/http/core"

type CreateUserRequest struct {
    Name  string `json:"name"  validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age"   validate:"required,min=1,max=150"`
    Role  string `json:"role"  validate:"oneof=user admin moderator"`
}

func createUserHandler(ctx core.Context) error {
    req, err := core.Bind[CreateUserRequest](ctx, core.BindOptions{
        Source:   core.BindSourceBody,
        Validate: true,
    })
    if err != nil {
        return ctx.BadRequestMsg("Invalid request: " + err.Error())
    }
    // use req...
}
```

### MustBind

Parses + validates + auto-sends 400 on failure. Returns `(T, bool)`.

```go
func updateUserHandler(ctx core.Context) error {
    req, ok := core.MustBind[UpdateUserRequest](ctx)
    if !ok {
        return nil // 400 error response already sent to client
    }
    // use req...
}
```

### Shorthand Binding

```go
user, err := core.BindBody[CreateUserRequest](ctx, true)    // body + validate
query, err := core.BindQuery[SearchParams](ctx, false)       // query, skip validation
params, err := core.BindParams[RouteParams](ctx, true)       // URL params + validate
```

### TypedHandler

Zero-boilerplate handler that auto-parses, validates, and marshals:

```go
var createUser = core.TypedHandler(core.StatusCreated,
    func(ctx core.Context, req CreateUserRequest) (CreateUserResponse, error) {
        // req is already parsed and validated
        // returned value is auto-marshaled to JSON with the specified status code
        return CreateUserResponse{ID: 123, Name: req.Name}, nil
    },
)

srv.POST("/users", createUser)
```

### Validation Rules

| Rule | Description | Example |
|------|-------------|---------|
| `required` | Must not be zero value | `validate:"required"` |
| `min` / `max` | Min/max value or length | `validate:"min=3,max=50"` |
| `len` | Exact length | `validate:"len=10"` |
| `email` | Valid email format | `validate:"email"` |
| `url` | Valid URL format | `validate:"url"` |
| `numeric` / `alpha` | Character type | `validate:"numeric"` |
| `oneof` | Allowed values | `validate:"oneof=a b c"` |
| `gt` / `gte` / `lt` / `lte` | Comparisons | `validate:"gt=0,lte=100"` |

---

## Response Helpers

### Shorthand Responses

```go
// ── Success ──
return ctx.OK(data)                           // 200 + JSON
return ctx.Created(data)                      // 201 + JSON
return ctx.NoContent()                        // 204

// ── Client Errors ──
return ctx.BadRequestMsg("Invalid input")     // 400
return ctx.UnauthorizedMsg("Not logged in")   // 401
return ctx.ForbiddenMsg("Access denied")      // 403
return ctx.NotFoundMsg("User not found")      // 404

// ── Server Errors ──
return ctx.InternalErrorMsg("Server error")   // 500
```

### Structured Responses

**Success:**

```go
resp := core.NewSuccessResponse(200, "User created", userData)
return core.SendSuccess(ctx, resp)
```

Response body:
```json
{
    "http_status": 200,
    "code": "SUCCESS",
    "message": "User created",
    "timestamp": "2026-03-27T07:00:00Z",
    "request_id": "01952...",
    "data": { ... }
}
```

**Error with details:**

```go
return core.SendError(ctx, core.NewErrorResponse("INSUFFICIENT_BALANCE", 400, "Balance too low").
    WithDetails("required", 100).
    WithDetails("current", 42).
    WithInternalMsg("User %s tried to withdraw %d", userID, amount). // server-side log only
    WithCause(originalErr).                                           // server-side log only
    WithHTTPStatus(422),                                              // override status code
)
```

> **`UseProperHTTPStatus`**: When `false` (legacy mode), all responses return HTTP 200 with error details in the body. When `true`, the actual HTTP status code is used.

### Error Utilities

```go
// Check error code in error chain
if core.IsErrorCode(err, "NOT_FOUND") { ... }

// Wrap errors (preserves error chain, deep-copies Details)
return core.WrapError(err, "failed to create user")
return core.WrapErrorf(err, "failed to process user %d", userID)

// Auto-handle ErrorResponse (sends response + returns true, or returns false)
if core.HandleError(ctx, err) {
    return nil // error response was sent
}

// Validate struct and auto-send 400 on failure
if ok, err := core.ValidateAndRespond(ctx, request); !ok {
    return err
}
```

### Query & Parameter Helpers

```go
// ── URL parameters ──
id, err := core.GetParamInt(ctx, "id")       // int
id64, err := core.GetParamInt64(ctx, "id")   // int64
uuid, err := core.GetParamUUID(ctx, "id")    // string (validated UUID)

// ── Query parameters (with defaults) ──
page := core.GetQueryInt(ctx, "page", 1)              // default: 1
limit := core.GetQueryInt64(ctx, "limit", 10)          // default: 10
active := core.GetQueryBool(ctx, "active", true)       // accepts: true/1/yes/on, false/0/no/off
sortBy := core.GetQueryString(ctx, "sort", "created_at")
```

---

## Middleware

### Custom Middleware

```go
// Middleware signature: func(core.Context) error
func timingMiddleware(ctx core.Context) error {
    start := time.Now()
    err := ctx.Next() // call next handler
    logger.Infow("timing", "path", ctx.Path(), "ms", time.Since(start).Milliseconds())
    return err
}

srv.GET("/users", listUsersHandler, timingMiddleware)  // per-route
srv.Use(timingMiddleware)                               // global
```

### Middleware Composition

```go
import "github.com/anthanhphan/gosdk/orianna/http/middleware"

// Chain multiple middleware into one
combined := middleware.Chain(authMW, loggingMW, rateLimitMW)

// Conditional — apply only when condition is true
middleware.Optional(func(ctx core.Context) bool {
    return ctx.Get("X-Debug") != ""
}, debugMW)

// Method filter — apply only for specific HTTP methods
middleware.OnlyForMethods(auditMW, "POST", "PUT", "DELETE")

// Path exclusion — skip for specific paths or prefixes
middleware.SkipForPaths(loggingMW, "/health", "/metrics")
middleware.SkipForPathPrefixes(authMW, "/public")

// Before/After hooks — run code before or after a middleware
middleware.Before(myMW, func(ctx core.Context) { ctx.Locals("t", time.Now()) })
middleware.After(myMW, func(ctx core.Context, err error) { log.Println(err) })

// Safety wrappers
middleware.Recover(unsafeMW)                    // catch panics from third-party middleware
middleware.Timeout(slowMW, 5*time.Second)       // cancel if middleware exceeds timeout
```

---

## Authentication & Authorization

```go
// 1. Implement auth middleware
func authMiddleware(ctx core.Context) error {
    token := ctx.Get("Authorization")
    if token == "" {
        return core.NewErrorResponse("UNAUTHORIZED", core.StatusUnauthorized, "Missing token")
    }
    userID, err := validateToken(token)
    if err != nil {
        return core.NewErrorResponse("UNAUTHORIZED", core.StatusUnauthorized, "Invalid token")
    }
    ctx.Locals("user_id", userID)
    ctx.Locals("role", userRole)
    return ctx.Next()
}

// 2. Implement authorization checker
func authzChecker(ctx core.Context, permissions []string) error {
    role := ctx.Locals("role").(string)
    for _, perm := range permissions {
        if !hasPermission(role, perm) {
            return core.NewErrorResponse("FORBIDDEN", core.StatusForbidden,
                fmt.Sprintf("Missing permission: %s", perm))
        }
    }
    return nil
}

// 3. Wire into server
srv, _ := server.NewServer(config,
    server.WithAuthentication(authMiddleware),
    server.WithAuthorization(authzChecker),
)

// 4. Use protected routes
srv.Protected().GET("/profile", profileHandler)
srv.Protected().WithPermissions("admin:write").POST("/admin/settings", handler)
```

---

## Health Checks

```go
import "github.com/anthanhphan/gosdk/orianna/shared/health"

srv, _ := server.NewServer(config,
    server.WithHealthChecker(
        health.NewCustomChecker("database", func(ctx context.Context) health.HealthCheck {
            if err := db.PingContext(ctx); err != nil {
                return health.HealthCheck{
                    Status:  health.StatusUnhealthy,
                    Message: "PostgreSQL unreachable",
                    Error:   err,
                }
            }
            return health.HealthCheck{
                Status:  health.StatusHealthy,
                Message: "PostgreSQL connected",
            }
        }),
    ),
    server.WithHealthChecker(
        health.NewHTTPChecker("http://redis:6379/ping", "cache", 5*time.Second),
    ),
)

// Programmatic access
report := srv.GetHealthManager().Check(ctx)
// report.Status: "healthy" | "unhealthy" | "degraded"
```

**Health status values:**

| Status | Meaning |
|--------|---------|
| `health.StatusHealthy` | All checkers pass |
| `health.StatusDegraded` | Some checkers fail, service partially available |
| `health.StatusUnhealthy` | Critical checkers fail, service unavailable |

---

## Lifecycle Hooks

```go
import "github.com/anthanhphan/gosdk/orianna/http/core"

hooks := core.NewHooks()

hooks.AddOnRequest(func(ctx core.Context) {
    logger.Debugw("request", "method", ctx.Method(), "path", ctx.Path())
})

hooks.AddOnResponse(func(ctx core.Context, status int, latency time.Duration) {
    logger.Infow("response", "status", status, "ms", latency.Milliseconds())
})

hooks.AddOnError(func(ctx core.Context, err error) {
    logger.Errorw("error", "path", ctx.Path(), "error", err)
})

hooks.AddOnPanic(func(ctx core.Context, recovered any, stack []byte) {
    logger.Errorw("panic", "recovered", recovered)
})

hooks.AddOnShutdown(func() {
    db.Close()
    logger.Infow("shutting down")
})

hooks.AddOnServerStart(func(server any) error {
    logger.Infow("server started")
    return nil // return error to abort startup
})

srv, _ := server.NewServer(config, server.WithHooks(hooks))
```

---

## HTTP Client

Production-grade HTTP client with connection pooling, retry, circuit breaker, and observability.

```go
import "github.com/anthanhphan/gosdk/orianna/http/client"
import "github.com/anthanhphan/gosdk/orianna/shared/resilience"

httpClient, err := client.NewClient(
    client.WithBaseURL("https://api.example.com"),
    client.WithTimeout(30 * time.Second),
    client.WithServiceName("user-service"),    // metric prefix: user-service_http_*
    client.WithDefaultHeader("X-Client", "my-service"),
    client.WithTracing(tracingClient),
    client.WithMetrics(metricsClient),
    client.WithLogger(myLogger),
    client.WithRetry(&resilience.RetryConfig{
        MaxAttempts:          3,
        InitialBackoff:       100 * time.Millisecond,
        MaxBackoff:           5 * time.Second,
        Multiplier:           2.0,
        RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
        RetryableMethods:     []string{"GET", "HEAD"},
    }),
    client.WithCircuitBreaker(&resilience.CircuitBreakerConfig{
        FailureThreshold: 5,
        SuccessThreshold: 3,
        Timeout:          30 * time.Second,
    }),
)
defer httpClient.Close() // releases idle transport connections

// ── CRUD Methods ──
resp, err := httpClient.Get(ctx, "/users/1")
resp, err = httpClient.Post(ctx, "/users", body)
resp, err = httpClient.Put(ctx, "/users/1", body)
resp, err = httpClient.Delete(ctx, "/users/1")

// ── Request Options ──
resp, err = httpClient.Post(ctx, "/users", body, client.WithAuth("Bearer token"))
resp, err = httpClient.Get(ctx, "/posts", client.WithQuery("userId", "1"))

// ── Standalone Functions (no client setup needed) ──
resp, err = client.Get(ctx, "https://api.example.com", "/users/1")
```

| Option | Description |
|--------|-------------|
| `WithBaseURL(url)` | Base URL for all requests |
| `WithTimeout(d)` | Default request timeout (default: 30s) |
| `WithServiceName(name)` | Metric name prefix |
| `WithDefaultHeader(k, v)` | Header added to every request |
| `WithTracing(client)` | OpenTelemetry distributed tracing |
| `WithMetrics(client)` | Prometheus-compatible metrics |
| `WithLogger(log)` | Custom structured logger |
| `WithRetry(cfg)` | Retry with exponential backoff + jitter |
| `WithCircuitBreaker(cfg)` | Circuit breaker (fail-fast when downstream is down) |
| `WithTLSConfig(cfg)` | Custom TLS configuration |

---

## Context Interface (ISP)

`core.Context` is composed of 11 focused interfaces following the Interface Segregation Principle:

| Interface | Methods |
|-----------|---------|
| `RequestInfo` | `Method()`, `Path()`, `RoutePath()`, `OriginalURL()`, `BaseURL()`, `Protocol()`, `Hostname()`, `IP()`, `Secure()` |
| `HeaderManager` | `Get(key)`, `Set(key, value)`, `Append(field, values...)`, `HeadersParser(out)` |
| `ParamGetter` | `Params(key)`, `AllParams()`, `ParamsParser(out)` |
| `QueryGetter` | `Query(key)`, `AllQueries()`, `QueryParser(out)` |
| `BodyReader` | `Body()`, `BodyParser(out)` |
| `CookieManager` | `Cookies(key)`, `Cookie(cookie)`, `ClearCookie(keys...)` |
| `ResponseWriter` | `Status(code)`, `JSON(data)`, `XML(data)`, `SendString(s)`, `SendBytes(b)`, `SendStream(r, size...)`, `SendFile(path)`, `Redirect(url, status...)`, `ResponseStatusCode()` |
| `ContentNegotiator` | `Accepts(offers...)`, `AcceptsCharsets(...)`, `AcceptsEncodings(...)`, `AcceptsLanguages(...)` |
| `RequestState` | `Fresh()`, `Stale()`, `XHR()` |
| `LocalsStorage` | `Locals(key, value...)`, `GetAllLocals()` |
| `ShorthandResponder` | `OK(data)`, `Created(data)`, `NoContent()`, `BadRequestMsg(msg)`, `UnauthorizedMsg(msg)`, `ForbiddenMsg(msg)`, `NotFoundMsg(msg)`, `InternalErrorMsg(msg)` |

**Additional Context methods:** `Next()`, `Context()`, `SetContext(ctx)`, `IsMethod(method)`, `RequestID()`, `UseProperHTTPStatus()`
