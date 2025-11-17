package core

import (
	"context"
	"net/http"
	"time"
)

// Map is a convenient alias for map[string]interface{} used for flexible data structures.
// It provides a shorthand for JSON responses and dynamic data handling.
//
// Example:
//
//	ctx.JSON(core.Map{
//	    "success": true,
//	    "data": myData,
//	    "count": 42,
//	})
type Map map[string]interface{}

// Handler defines the function signature for route handlers.
//
// Handlers receive a Context and return an error. The error return enables:
//   - Centralized error handling through middleware
//   - Clean propagation of business logic errors
//   - Early returns for validation failures
//
// Input:
//   - ctx: The request/response context providing access to request data and response methods
//
// Output:
//   - error: Any error that occurred during handling (nil if successful)
//
// Example:
//
//	func getUserHandler(ctx core.Context) error {
//	    userID := ctx.Params("id")
//	    user, err := fetchUser(userID)
//	    if err != nil {
//	        return err // Will be handled by error middleware
//	    }
//	    return ctx.Status(200).JSON(user)
//	}
type Handler func(Context) error

// Middleware defines the function signature for middleware functions.
//
// Middleware enables cross-cutting concerns like logging, authentication,
// and validation. Middleware can:
//   - Pre-process requests before handlers
//   - Post-process responses after handlers
//   - Short-circuit execution by not calling ctx.Next()
//
// Input:
//   - ctx: The request/response context
//
// Output:
//   - error: Any error that occurred (stops middleware chain if non-nil)
//
// Example:
//
//	func authMiddleware(ctx core.Context) error {
//	    token := ctx.Get("Authorization")
//	    if token == "" {
//	        return errors.New("authentication required")
//	    }
//	    // Continue to next middleware or handler
//	    return ctx.Next()
//	}
type Middleware func(Context) error

// Method represents HTTP methods as type-safe constants.
type Method string

// HTTP method constants.
const (
	MethodGet     Method = http.MethodGet
	MethodPost    Method = http.MethodPost
	MethodPut     Method = http.MethodPut
	MethodPatch   Method = http.MethodPatch
	MethodDelete  Method = http.MethodDelete
	MethodHead    Method = http.MethodHead
	MethodOptions Method = http.MethodOptions
)

// Context defines the interface for request context operations.
//
// This interface abstracts HTTP request/response handling, providing a clean API
// that is independent of the underlying web framework. It enables:
//   - Framework-agnostic handler and middleware code
//   - Easy testing through mock implementations
//   - Type-safe access to request data
//   - Fluent response building
//
// The interface is implemented by internal/runtimectx.FiberContext which wraps
// Fiber's context while tracking state for proper context.Context integration.
//
// Example:
//
//	func handler(ctx core.Context) error {
//	    // Read request
//	    userID := ctx.Params("id")
//	    var req UpdateRequest
//	    if err := ctx.BodyParser(&req); err != nil {
//	        return err
//	    }
//
//	    // Store in context
//	    ctx.Locals("action", "update_user")
//
//	    // Send response
//	    return ctx.Status(200).JSON(core.Map{
//	        "success": true,
//	    })
//	}
type Context interface {
	// Request information
	Method() string
	Path() string
	OriginalURL() string
	BaseURL() string
	Protocol() string
	Hostname() string
	IP() string
	Secure() bool

	// Headers
	Get(key string, defaultValue ...string) string
	Set(key, value string)
	Append(field string, values ...string)

	// Route parameters
	Params(key string, defaultValue ...string) string
	AllParams() map[string]string
	ParamsParser(out interface{}) error

	// Query parameters
	Query(key string, defaultValue ...string) string
	AllQueries() map[string]string
	QueryParser(out interface{}) error

	// Request body
	Body() []byte
	BodyParser(out interface{}) error

	// Cookies
	Cookies(key string, defaultValue ...string) string
	Cookie(cookie *Cookie)
	ClearCookie(key ...string)

	// Response
	Status(status int) Context
	JSON(data interface{}) error
	XML(data interface{}) error
	SendString(s string) error
	SendBytes(b []byte) error
	Redirect(location string, status ...int) error

	// Content negotiation
	Accepts(offers ...string) string
	AcceptsCharsets(offers ...string) string
	AcceptsEncodings(offers ...string) string
	AcceptsLanguages(offers ...string) string

	// Request state
	Fresh() bool
	Stale() bool
	XHR() bool

	// Context storage (locals)
	Locals(key string, value ...interface{}) interface{}
	GetAllLocals() map[string]interface{}

	// Middleware flow
	Next() error

	// Access underlying context for advanced use cases
	Context() context.Context

	// Utility methods
	IsMethod(method string) bool
	RequestID() string
}

// Cookie represents an HTTP cookie with standard attributes.
//
// This struct provides a framework-agnostic way to set cookies in responses.
// It includes all standard cookie security attributes.
//
// Example:
//
//	cookie := &core.Cookie{
//	    Name:     "session_token",
//	    Value:    "abc123",
//	    Path:     "/",
//	    MaxAge:   3600,
//	    Secure:   true,
//	    HTTPOnly: true,
//	    SameSite: "Strict",
//	}
//	ctx.Cookie(cookie)
type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
	SameSite string
}
