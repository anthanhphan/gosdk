package aurelion

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Context defines the interface for request context operations
// It wraps fiber.Ctx and provides a clean, testable API
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

	// Middleware flow
	Next() error

	// Access underlying context for advanced use cases
	Context() interface{}
}

// Cookie represents an HTTP cookie
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

// contextWrapper wraps fiber.Ctx to implement our Context interface
type contextWrapper struct {
	fiberCtx *fiber.Ctx
}

// newContext creates a new context wrapper from fiber context.
// Returns nil only if fiberCtx is nil, which should never happen in normal operation.
func newContext(fiberCtx *fiber.Ctx) Context {
	if fiberCtx == nil {
		return nil
	}
	return &contextWrapper{fiberCtx: fiberCtx}
}

// Method returns the HTTP method
func (c *contextWrapper) Method() string {
	return c.fiberCtx.Method()
}

// Path returns the path part of the request URL
func (c *contextWrapper) Path() string {
	return c.fiberCtx.Path()
}

// OriginalURL returns the original request URL
func (c *contextWrapper) OriginalURL() string {
	return c.fiberCtx.OriginalURL()
}

// BaseURL returns the base URL (scheme + host)
func (c *contextWrapper) BaseURL() string {
	return c.fiberCtx.BaseURL()
}

// Protocol returns the protocol (http or https)
func (c *contextWrapper) Protocol() string {
	return c.fiberCtx.Protocol()
}

// Hostname returns the hostname from the request
func (c *contextWrapper) Hostname() string {
	return c.fiberCtx.Hostname()
}

// IP returns the remote IP address
func (c *contextWrapper) IP() string {
	return c.fiberCtx.IP()
}

// Secure returns true if the connection is secure (HTTPS)
func (c *contextWrapper) Secure() bool {
	return c.fiberCtx.Secure()
}

// Get returns the HTTP request header field
func (c *contextWrapper) Get(key string, defaultValue ...string) string {
	return c.fiberCtx.Get(key, defaultValue...)
}

// Set sets the HTTP response header field
func (c *contextWrapper) Set(key, value string) {
	c.fiberCtx.Set(key, value)
}

// Append appends the specified value to the HTTP response header field
func (c *contextWrapper) Append(field string, values ...string) {
	c.fiberCtx.Append(field, values...)
}

// Params returns the route parameter by key
func (c *contextWrapper) Params(key string, defaultValue ...string) string {
	return c.fiberCtx.Params(key, defaultValue...)
}

// AllParams returns all route parameters as a map
func (c *contextWrapper) AllParams() map[string]string {
	return c.fiberCtx.AllParams()
}

// ParamsParser binds the route parameters to a struct
func (c *contextWrapper) ParamsParser(out interface{}) error {
	return c.fiberCtx.ParamsParser(out)
}

// Query returns the query parameter by key
func (c *contextWrapper) Query(key string, defaultValue ...string) string {
	return c.fiberCtx.Query(key, defaultValue...)
}

// AllQueries returns all query parameters as a map
func (c *contextWrapper) AllQueries() map[string]string {
	return c.fiberCtx.Queries()
}

// QueryParser binds the query parameters to a struct
func (c *contextWrapper) QueryParser(out interface{}) error {
	return c.fiberCtx.QueryParser(out)
}

// Body returns the raw request body
func (c *contextWrapper) Body() []byte {
	return c.fiberCtx.Body()
}

// BodyParser binds the request body to a struct
func (c *contextWrapper) BodyParser(out interface{}) error {
	return c.fiberCtx.BodyParser(out)
}

// Cookies returns the cookie value by key
func (c *contextWrapper) Cookies(key string, defaultValue ...string) string {
	return c.fiberCtx.Cookies(key, defaultValue...)
}

// Cookie sets a cookie
func (c *contextWrapper) Cookie(cookie *Cookie) {
	fiberCookie := &fiber.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		MaxAge:   cookie.MaxAge,
		Expires:  cookie.Expires,
		Secure:   cookie.Secure,
		HTTPOnly: cookie.HTTPOnly,
		SameSite: cookie.SameSite,
	}
	c.fiberCtx.Cookie(fiberCookie)
}

// ClearCookie removes cookies by names
func (c *contextWrapper) ClearCookie(key ...string) {
	c.fiberCtx.ClearCookie(key...)
}

// Status sets the HTTP status code
func (c *contextWrapper) Status(status int) Context {
	c.fiberCtx.Status(status)
	return c
}

// JSON sends a JSON response
func (c *contextWrapper) JSON(data interface{}) error {
	return c.fiberCtx.JSON(data)
}

// XML sends an XML response
func (c *contextWrapper) XML(data interface{}) error {
	return c.fiberCtx.XML(data)
}

// SendString sends a string response
func (c *contextWrapper) SendString(s string) error {
	return c.fiberCtx.SendString(s)
}

// SendBytes sends a byte array response
func (c *contextWrapper) SendBytes(b []byte) error {
	return c.fiberCtx.Send(b)
}

// Redirect redirects to the specified URL
func (c *contextWrapper) Redirect(location string, status ...int) error {
	return c.fiberCtx.Redirect(location, status...)
}

// Accepts checks if the specified content types are acceptable
func (c *contextWrapper) Accepts(offers ...string) string {
	return c.fiberCtx.Accepts(offers...)
}

// AcceptsCharsets checks if the specified charsets are acceptable
func (c *contextWrapper) AcceptsCharsets(offers ...string) string {
	return c.fiberCtx.AcceptsCharsets(offers...)
}

// AcceptsEncodings checks if the specified encodings are acceptable
func (c *contextWrapper) AcceptsEncodings(offers ...string) string {
	return c.fiberCtx.AcceptsEncodings(offers...)
}

// AcceptsLanguages checks if the specified languages are acceptable
func (c *contextWrapper) AcceptsLanguages(offers ...string) string {
	return c.fiberCtx.AcceptsLanguages(offers...)
}

// Fresh returns true when the response is still "fresh"
func (c *contextWrapper) Fresh() bool {
	return c.fiberCtx.Fresh()
}

// Stale returns true when the response is "stale"
func (c *contextWrapper) Stale() bool {
	return c.fiberCtx.Stale()
}

// XHR returns true if the request's X-Requested-With header field is XMLHttpRequest
func (c *contextWrapper) XHR() bool {
	return c.fiberCtx.XHR()
}

// Locals stores and retrieves values scoped to the request
func (c *contextWrapper) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		c.fiberCtx.Locals(key, value[0])
		return value[0]
	}
	return c.fiberCtx.Locals(key)
}

// Next executes the next handler in the chain
func (c *contextWrapper) Next() error {
	return c.fiberCtx.Next()
}

// Context returns the underlying fiber context for advanced use cases
func (c *contextWrapper) Context() interface{} {
	return c.fiberCtx
}
