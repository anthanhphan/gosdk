# Aurelion - Production-Ready HTTP Server Framework

<div align="center">

**A high-performance, production-ready HTTP server framework for Go, built on Fiber**

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Test Coverage](https://img.shields.io/badge/coverage-90%25-brightgreen.svg)](https://github.com/anthanhphan/gosdk)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[Features](#features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Examples](#examples)

</div>

---

## 🌟 **Features**

- 🚀 **High Performance** - Built on Fiber, one of the fastest Go web frameworks
- 🔒 **Security First** - Helmet, CORS, CSRF, rate limiting out of the box
- 📊 **Observability** - Request ID, Trace ID, structured logging built-in
- 🧪 **Testable** - Clean interface-based design for easy testing
- 🛠️ **Developer Friendly** - Fluent API, comprehensive error handling
- ⚡ **Production Ready** - Graceful shutdown, panic recovery, 90%+ test coverage
- 🎯 **Type Safe** - Strong typing for HTTP methods and routes
- 🔄 **Flexible** - Minimal opinions, maximum flexibility

---

## 📦 **Installation**

```bash
go get github.com/anthanhphan/gosdk/transport/aurelion
```

---

## 🚀 **Quick Start**

### Basic Server

```go
package main

import (
    "github.com/anthanhphan/gosdk/transport/aurelion"
)

func main() {
    // Create configuration
    config := &aurelion.Config{
        ServiceName: "My API",
        Port:        8080,
    }

    // Create server
    server, err := aurelion.NewHttpServer(config)
    if err != nil {
        panic(err)
    }

    // Add routes
    server.AddRoutes(
        aurelion.NewRoute("/hello").
            GET().
            Handler(func(ctx aurelion.Context) error {
                return aurelion.OK(ctx, "Hello, World!", nil)
            }),
    )

    // Start server
    if err := server.Start(); err != nil {
        panic(err)
    }
}
```

### With Authentication

```go
package main

import (
    "github.com/anthanhphan/gosdk/transport/aurelion"
)

func main() {
    config := &aurelion.Config{
        ServiceName:         "My API",
        Port:                8080,
        UseProperHTTPStatus: true, // Use proper HTTP status codes
    }

    // Create server with auth
    server, _ := aurelion.NewHttpServer(
        config,
        aurelion.WithAuthentication(authMiddleware()),
        aurelion.WithAuthorization(authzChecker()),
    )

    // Add routes
    server.AddRoutes(
        aurelion.NewRoute("/public").GET().Handler(publicHandler),
        aurelion.NewRoute("/protected").GET().Protected().Handler(protectedHandler),
    )

    server.Start()
}
```

---

## 📖 **Core Concepts**

### Routes

Define HTTP endpoints using a fluent builder API:

```go
// Simple routes
server.AddRoutes(
    aurelion.NewRoute("/users").GET().Handler(getUsers),
    aurelion.NewRoute("/users/:id").PUT().Handler(updateUser),
)

// Protected routes (require authentication)
server.AddRoutes(
    aurelion.NewRoute("/admin").
        GET().
        Protected().
        Handler(adminHandler),
)

// Routes with permissions
server.AddRoutes(
    aurelion.NewRoute("/admin/users").
        GET().
        Permissions("read:users", "admin").
        Handler(listAdminUsers),
)
```

### Route Groups

Group related routes with common prefix and middleware:

```go
server.AddGroupRoutes(
    aurelion.NewGroupRoute("/api/v1").Routes(
        aurelion.NewRoute("/users").GET().Handler(getUsersV1),
        aurelion.NewRoute("/posts").GET().Handler(getPostsV1),
    ),

    // Protected group
    aurelion.NewGroupRoute("/api/admin").
        Protected().
        Routes(
            aurelion.NewRoute("/dashboard").GET().Handler(dashboard),
        ),
)
```

### Request Handling

```go
func createUser(ctx aurelion.Context) error {
    // Define request structure
    type CreateUserRequest struct {
        Name  string `json:"name" validate:"required,min=3,max=50"`
        Email string `json:"email" validate:"required,email"`
        Age   int    `json:"age" validate:"min=18,max=100"`
    }

    // Validate and parse in one step
    var req CreateUserRequest
    if !aurelion.ValidateWithResponse(ctx, &req) {
        return nil // Validation errors already sent
    }

    // Get route parameters
    userID := ctx.Params("id")

    // Get query parameters
    includeDetails := ctx.Query("details", "false")

    // Access headers
    authToken := ctx.Get("Authorization")

    // Store values in context
    ctx.Locals("user_id", "123")
    ctx.Locals("lang", "en")

    // Your business logic here
    user := createUserInDB(req)

    // Return success response
    return aurelion.OK(ctx, "User created successfully", user)
}
```

### Error Handling

```go
// Standard HTTP errors
return aurelion.BadRequest(ctx, "Invalid input")
return aurelion.Unauthorized(ctx, "Authentication required")
return aurelion.Forbidden(ctx, "Access denied")
return aurelion.NotFound(ctx, "User not found")
return aurelion.InternalServerError(ctx, "Database error")

// Business errors with custom codes
return aurelion.Error(ctx, aurelion.NewError(1001, "User already exists"))
return aurelion.Error(ctx, aurelion.NewErrorf(1002, "Invalid email: %s", email))
```

### Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=18,max=100"`
    Website  string `json:"website" validate:"url"`
    Phone    string `json:"phone" validate:"numeric"`
}

// Option 1: Manual validation
var req CreateUserRequest
if err := aurelion.ValidateAndParse(ctx, &req); err != nil {
    return aurelion.BadRequest(ctx, err.Error())
}

// Option 2: With custom error handling (array format for frontend)
var req CreateUserRequest
if err := aurelion.ValidateAndParse(ctx, &req); err != nil {
    if validationErr, ok := err.(aurelion.ValidationErrors); ok {
        return ctx.Status(400).JSON(aurelion.Map{
            "success": false,
            "code":    400,
            "message": "Validation failed",
            "errors":  validationErr.ToArray(), // Array format - easy to parse
        })
    }
    return aurelion.BadRequest(ctx, err.Error())
}

// Response format:
// {
//   "errors": [
//     {"field": "email", "message": "must be a valid email address"},
//     {"field": "age", "message": "must be at least 18"}
//   ]
// }
```

Supported validation rules:

- `required` - Field must not be empty
- `min=N` - Minimum length/value
- `max=N` - Maximum length/value
- `email` - Valid email format
- `url` - Valid URL format
- `numeric` - Only numeric characters
- `alpha` - Only alphabetic characters

---

## ⚙️ **Configuration**

### Basic Configuration

```go
config := &aurelion.Config{
    ServiceName: "My API",
    Port:        8080,
}
```

### Advanced Configuration

```go
import "time"

readTimeout := 10 * time.Second
writeTimeout := 10 * time.Second
shutdownTimeout := 30 * time.Second

config := &aurelion.Config{
    ServiceName:             "My API",
    Port:                    8080,
    ReadTimeout:             &readTimeout,
    WriteTimeout:            &writeTimeout,
    GracefulShutdownTimeout: &shutdownTimeout,
    MaxBodySize:             4 * 1024 * 1024, // 4MB
    MaxConcurrentConnections: 100000,
    VerboseLogging:          false,
    UseProperHTTPStatus:     true, // Recommended for new projects

    // CORS
    EnableCORS: true,
    CORS: &aurelion.CORSConfig{
        AllowOrigins:     []string{"https://example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           3600,
    },

    // CSRF Protection
    EnableCSRF: true,
    CSRF: &aurelion.CSRFConfig{
        KeyLookup:      "header:X-Csrf-Token",
        CookieSecure:   true,
        CookieHTTPOnly: true,
        CookieSameSite: "Strict",
    },
}
```

### HTTP Status Code Behavior

**Backward Compatible Mode (Default):**

```go
config := &aurelion.Config{
    UseProperHTTPStatus: false, // default
}
// All error responses return HTTP 200 with error code in JSON body
// Example: HTTP 200 {"success": false, "code": 400, "message": "Bad request"}
```

**Proper HTTP Status Mode (Recommended for new projects):**

```go
config := &aurelion.Config{
    UseProperHTTPStatus: true,
}
// Error responses use proper HTTP status codes
// Example: HTTP 400 {"success": false, "code": 400, "message": "Bad request"}
```

---

## 🔒 **Security Features**

### Built-in Security

- **Helmet Middleware** - Security headers (XSS, clickjacking protection)
- **Rate Limiting** - 500 req/min per IP by default (configurable)
- **Panic Recovery** - Automatic panic recovery with logging
- **CORS Support** - Configurable cross-origin resource sharing
- **CSRF Protection** - Optional CSRF token validation
- **Request ID** - Automatic request tracking
- **Trace ID** - Distributed tracing support

### Authentication & Authorization

```go
// Authentication middleware
func authMiddleware() aurelion.Middleware {
    return func(ctx aurelion.Context) error {
        token := ctx.Get("Authorization")
        if token == "" {
            return aurelion.Unauthorized(ctx, "Token required")
        }

        // Validate token
        user, err := validateToken(token)
        if err != nil {
            return aurelion.Unauthorized(ctx, "Invalid token")
        }

        // Store user info
        ctx.Locals("user_id", user.ID)
        ctx.Locals("permissions", user.Permissions)

        return ctx.Next()
    }
}

// Authorization checker
func authzChecker() func(aurelion.Context, []string) error {
    return func(ctx aurelion.Context, required []string) error {
        userPerms := ctx.Locals("permissions").([]string)

        for _, perm := range required {
            if !contains(userPerms, perm) {
                return fmt.Errorf("missing permission: %s", perm)
            }
        }

        return nil
    }
}

// Use in server
server, _ := aurelion.NewHttpServer(
    config,
    aurelion.WithAuthentication(authMiddleware()),
    aurelion.WithAuthorization(authzChecker()),
)
```

---

## 📊 **Observability**

### Request ID

Every request automatically gets a unique request ID:

```go
func handler(ctx aurelion.Context) error {
    requestID := aurelion.GetRequestID(ctx)
    log.Printf("Processing request: %s", requestID)
    return nil
}
```

### Trace ID

Support for distributed tracing:

```go
func handler(ctx aurelion.Context) error {
    traceID := aurelion.GetTraceID(ctx)

    // Pass to downstream services
    httpClient.Do(req.WithHeader("X-Trace-ID", traceID))

    return nil
}
```

### Context Integration

Seamless integration with standard `context.Context`:

```go
import "github.com/anthanhphan/gosdk/transport/aurelion/contextutil"

func handler(ctx aurelion.Context) error {
    // Set values in Locals
    ctx.Locals("user_id", "123")
    ctx.Locals("lang", "en")

    // Access as standard context
    stdCtx := ctx.Context()

    // Use with contextutil
    userID := contextutil.GetUserIDFromContext(stdCtx)
    lang := contextutil.GetLanguageFromContext(stdCtx)

    // Pass to other functions expecting context.Context
    result, err := callExternalService(stdCtx, data)

    return aurelion.OK(ctx, "Success", result)
}
```

---

## 📚 **API Reference**

### Server Creation

- `NewHttpServer(config *Config, options ...ServerOption)` - Create new HTTP server

### Server Options

- `WithGlobalMiddleware(...Middleware)` - Add global middleware
- `WithAuthentication(Middleware)` - Set authentication middleware
- `WithAuthorization(AuthorizationFunc)` - Set authorization checker
- `WithPanicRecover(Middleware)` - Set custom panic recovery
- `WithRateLimiter(Middleware)` - Set custom rate limiter

### Routes

- `NewRoute(path string)` - Create new route builder
- `NewGroupRoute(prefix string)` - Create new group route builder

### Response Helpers

- `OK(ctx, message, data)` - 200 OK response
- `Error(ctx, err)` - Business error response
- `BadRequest(ctx, message)` - 400 Bad Request
- `Unauthorized(ctx, message)` - 401 Unauthorized
- `Forbidden(ctx, message)` - 403 Forbidden
- `NotFound(ctx, message)` - 404 Not Found
- `InternalServerError(ctx, message)` - 500 Internal Server Error

### Validation

- `Validate(v interface{})` - Validate struct with tags
- `ValidateAndParse(ctx, v)` - Parse and validate request body

### Utilities

- `GetRequestID(ctx)` - Get request ID
- `GetTraceID(ctx)` - Get trace ID

---

## 📝 **Examples**

See [`example/main.go`](example/main.go) for a comprehensive example demonstrating:

- Route creation and grouping
- Authentication and authorization
- Custom middleware
- Error handling
- Validation
- Context usage

---

## 🚀 **Migration Guide**

### From Other Frameworks

**From Gin:**

```go
// Gin
router.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

// Aurelion
server.AddRoutes(
    aurelion.NewRoute("/users/:id").GET().Handler(func(ctx aurelion.Context) error {
        id := ctx.Params("id")
        return aurelion.OK(ctx, "User found", aurelion.Map{"id": id})
    }),
)
```

---

## 🤝 **Contributing**

Contributions are welcome! Please follow:

1. Go coding standards
2. Include tests for new features
3. Update documentation
4. Follow existing code style

---

## 📄 **License**

[MIT License](LICENSE)

---

## 🙏 **Acknowledgments**

Built with ❤️ using:

- [Fiber](https://github.com/gofiber/fiber) - Express-inspired web framework
- [Zap](https://github.com/uber-go/zap) - Structured logging

---

## 📞 **Support**

- 📖 [Full Documentation](./IMPROVEMENTS.md)
- 🐛 [Issue Tracker](https://github.com/anthanhphan/gosdk/issues)
- 💬 [Discussions](https://github.com/anthanhphan/gosdk/discussions)

---

<div align="center">

**⭐ Star us on GitHub — it helps!**

Made with ❤️ by the gosdk team

</div>
