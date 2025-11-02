package aurelion

import (
	"context"
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
	Context() context.Context

	// GetAllLocals returns all Locals keys and values as a map
	GetAllLocals() map[string]interface{}

	// Utility methods
	IsMethod(method string) bool
	RequestID() string
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
	// trackedKeys stores all keys that have been set through Locals() method
	// to enable GetAllLocals() without using unsafe operations
	trackedKeys map[string]struct{}
}

// newContext creates a new context wrapper from fiber context.
// Returns nil only if fiberCtx is nil, which should never happen in normal operation.
func newContext(fiberCtx *fiber.Ctx) Context {
	if fiberCtx == nil {
		return nil
	}
	return &contextWrapper{
		fiberCtx:    fiberCtx,
		trackedKeys: make(map[string]struct{}),
	}
}

// Method returns the HTTP method of the request.
//
// Output:
//   - string: The HTTP method (e.g., "GET", "POST", "PUT")
//
// Example:
//
//	method := ctx.Method()
//	if method == "POST" {
//	    // Handle POST request
//	}
func (c *contextWrapper) Method() string {
	return c.fiberCtx.Method()
}

// Path returns the route path from the request URL.
//
// Output:
//   - string: The route path (e.g., "/users/123")
//
// Example:
//
//	path := ctx.Path()
//	logger.Info("request path", "path", path)
func (c *contextWrapper) Path() string {
	return c.fiberCtx.Path()
}

// OriginalURL returns the original request URL including query string.
//
// Output:
//   - string: The full original URL
//
// Example:
//
//	url := ctx.OriginalURL()
//	logger.Info("original URL", "url", url)
func (c *contextWrapper) OriginalURL() string {
	return c.fiberCtx.OriginalURL()
}

// BaseURL returns the base URL (scheme + host) without path.
//
// Output:
//   - string: The base URL (e.g., "http://localhost:8080")
//
// Example:
//
//	baseURL := ctx.BaseURL()
//	fullURL := baseURL + ctx.Path()
func (c *contextWrapper) BaseURL() string {
	return c.fiberCtx.BaseURL()
}

// Protocol returns the protocol of the request.
//
// Output:
//   - string: The protocol ("http" or "https")
//
// Example:
//
//	protocol := ctx.Protocol()
//	if protocol == "https" {
//	    // Handle secure connection
//	}
func (c *contextWrapper) Protocol() string {
	return c.fiberCtx.Protocol()
}

// Hostname returns the hostname from the request.
//
// Output:
//   - string: The hostname (e.g., "example.com")
//
// Example:
//
//	hostname := ctx.Hostname()
//	logger.Info("request hostname", "hostname", hostname)
func (c *contextWrapper) Hostname() string {
	return c.fiberCtx.Hostname()
}

// IP returns the remote IP address of the client.
//
// Output:
//   - string: The client IP address
//
// Example:
//
//	ip := ctx.IP()
//	logger.Info("client IP", "ip", ip)
func (c *contextWrapper) IP() string {
	return c.fiberCtx.IP()
}

// Secure returns true if the connection is secure (HTTPS).
//
// Output:
//   - bool: True if using HTTPS, false otherwise
//
// Example:
//
//	if ctx.Secure() {
//	    // Handle secure connection
//	}
func (c *contextWrapper) Secure() bool {
	return c.fiberCtx.Secure()
}

// Get returns the HTTP request header field value.
//
// Input:
//   - key: The header field name
//   - defaultValue: Optional default value if header is not present
//
// Output:
//   - string: The header value, or default value if not found
//
// Example:
//
//	authHeader := ctx.Get("Authorization")
//	lang := ctx.Get("Accept-Language", "en") // Default to "en" if not present
func (c *contextWrapper) Get(key string, defaultValue ...string) string {
	return c.fiberCtx.Get(key, defaultValue...)
}

// Set sets the HTTP response header field.
//
// Input:
//   - key: The header field name
//   - value: The header field value
//
// Example:
//
//	ctx.Set("Content-Type", "application/json")
//	ctx.Set("X-Custom-Header", "value")
func (c *contextWrapper) Set(key, value string) {
	c.fiberCtx.Set(key, value)
}

// Append appends the specified value to the HTTP response header field.
//
// Input:
//   - field: The header field name
//   - values: One or more values to append
//
// Example:
//
//	ctx.Append("Set-Cookie", "cookie1=value1")
//	ctx.Append("Vary", "Accept-Encoding", "Accept-Language")
func (c *contextWrapper) Append(field string, values ...string) {
	c.fiberCtx.Append(field, values...)
}

// Params returns the route parameter value by key.
//
// Input:
//   - key: The route parameter key (e.g., "id" for route "/users/:id")
//   - defaultValue: Optional default value if parameter is not present
//
// Output:
//   - string: The parameter value, or default value if not found
//
// Example:
//
//	userID := ctx.Params("id")
//	page := ctx.Params("page", "1") // Default to "1" if not present
func (c *contextWrapper) Params(key string, defaultValue ...string) string {
	return c.fiberCtx.Params(key, defaultValue...)
}

// AllParams returns all route parameters as a map.
//
// Output:
//   - map[string]string: A map of all route parameter keys and values
//
// Example:
//
//	params := ctx.AllParams()
//	for key, value := range params {
//	    logger.Info("route param", "key", key, "value", value)
//	}
func (c *contextWrapper) AllParams() map[string]string {
	return c.fiberCtx.AllParams()
}

// ParamsParser binds the route parameters to a struct.
//
// Input:
//   - out: Pointer to struct to bind parameters to
//
// Output:
//   - error: Any error that occurred during parsing
//
// Example:
//
//	type UserParams struct {
//	    ID string `params:"id"`
//	}
//	var params UserParams
//	if err := ctx.ParamsParser(&params); err != nil {
//	    return aurelion.BadRequest(ctx, "Invalid parameters")
//	}
func (c *contextWrapper) ParamsParser(out interface{}) error {
	return c.fiberCtx.ParamsParser(out)
}

// Query returns the query parameter value by key.
//
// Input:
//   - key: The query parameter key
//   - defaultValue: Optional default value if parameter is not present
//
// Output:
//   - string: The query parameter value, or default value if not found
//
// Example:
//
//	page := ctx.Query("page")
//	limit := ctx.Query("limit", "10") // Default to "10" if not present
func (c *contextWrapper) Query(key string, defaultValue ...string) string {
	return c.fiberCtx.Query(key, defaultValue...)
}

// AllQueries returns all query parameters as a map.
//
// Output:
//   - map[string]string: A map of all query parameter keys and values
//
// Example:
//
//	queries := ctx.AllQueries()
//	for key, value := range queries {
//	    logger.Info("query param", "key", key, "value", value)
//	}
func (c *contextWrapper) AllQueries() map[string]string {
	return c.fiberCtx.Queries()
}

// QueryParser binds the query parameters to a struct.
//
// Input:
//   - out: Pointer to struct to bind query parameters to
//
// Output:
//   - error: Any error that occurred during parsing
//
// Example:
//
//	type PaginationParams struct {
//	    Page  int `query:"page"`
//	    Limit int `query:"limit"`
//	}
//	var params PaginationParams
//	if err := ctx.QueryParser(&params); err != nil {
//	    return aurelion.BadRequest(ctx, "Invalid query parameters")
//	}
func (c *contextWrapper) QueryParser(out interface{}) error {
	return c.fiberCtx.QueryParser(out)
}

// Body returns the raw request body as bytes.
//
// Output:
//   - []byte: The raw request body
//
// Example:
//
//	body := ctx.Body()
//	logger.Debug("request body", "body", string(body))
func (c *contextWrapper) Body() []byte {
	return c.fiberCtx.Body()
}

// BodyParser binds the request body to a struct.
// Supports JSON, XML, form data, and other formats based on Content-Type header.
//
// Input:
//   - out: Pointer to struct to bind body to
//
// Output:
//   - error: Any error that occurred during parsing
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//	var req CreateUserRequest
//	if err := ctx.BodyParser(&req); err != nil {
//	    return aurelion.BadRequest(ctx, "Invalid request body")
//	}
func (c *contextWrapper) BodyParser(out interface{}) error {
	return c.fiberCtx.BodyParser(out)
}

// Cookies returns the cookie value by key.
//
// Input:
//   - key: The cookie name
//   - defaultValue: Optional default value if cookie is not present
//
// Output:
//   - string: The cookie value, or default value if not found
//
// Example:
//
//	sessionID := ctx.Cookies("session_id")
//	theme := ctx.Cookies("theme", "light") // Default to "light" if not present
func (c *contextWrapper) Cookies(key string, defaultValue ...string) string {
	return c.fiberCtx.Cookies(key, defaultValue...)
}

// Cookie sets a cookie in the response.
//
// Input:
//   - cookie: The cookie configuration
//
// Example:
//
//	ctx.Cookie(&aurelion.Cookie{
//	    Name:     "session_id",
//	    Value:    sessionID,
//	    Path:     "/",
//	    MaxAge:   3600,
//	    Secure:   true,
//	    HTTPOnly: true,
//	})
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

// ClearCookie removes cookies by names.
//
// Input:
//   - key: One or more cookie names to remove
//
// Example:
//
//	ctx.ClearCookie("session_id")
//	ctx.ClearCookie("session_id", "auth_token")
func (c *contextWrapper) ClearCookie(key ...string) {
	c.fiberCtx.ClearCookie(key...)
}

// Status sets the HTTP status code for the response.
//
// Input:
//   - status: The HTTP status code (e.g., 200, 404, 500)
//
// Output:
//   - Context: The context instance (for chaining)
//
// Example:
//
//	ctx.Status(http.StatusCreated)
//	return ctx.JSON(data)
func (c *contextWrapper) Status(status int) Context {
	c.fiberCtx.Status(status)
	return c
}

// JSON sends a JSON response with automatic Content-Type header.
//
// Input:
//   - data: The data to serialize to JSON
//
// Output:
//   - error: Any error that occurred during JSON encoding
//
// Example:
//
//	user := map[string]interface{}{
//	    "id":   123,
//	    "name": "John",
//	}
//	return ctx.JSON(user)
func (c *contextWrapper) JSON(data interface{}) error {
	return c.fiberCtx.JSON(data)
}

// XML sends an XML response with automatic Content-Type header.
//
// Input:
//   - data: The data to serialize to XML
//
// Output:
//   - error: Any error that occurred during XML encoding
//
// Example:
//
//	type User struct {
//	    XMLName xml.Name `xml:"user"`
//	    ID      int      `xml:"id"`
//	    Name    string   `xml:"name"`
//	}
//	user := User{ID: 123, Name: "John"}
//	return ctx.XML(user)
func (c *contextWrapper) XML(data interface{}) error {
	return c.fiberCtx.XML(data)
}

// SendString sends a plain text string response.
//
// Input:
//   - s: The string to send
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return ctx.SendString("Hello, World!")
func (c *contextWrapper) SendString(s string) error {
	return c.fiberCtx.SendString(s)
}

// SendBytes sends a byte array response.
//
// Input:
//   - b: The byte array to send
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	imageData := []byte{...}
//	return ctx.SendBytes(imageData)
func (c *contextWrapper) SendBytes(b []byte) error {
	return c.fiberCtx.Send(b)
}

// Redirect redirects the client to the specified URL.
//
// Input:
//   - location: The URL to redirect to
//   - status: Optional HTTP status code (defaults to 302)
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return ctx.Redirect("https://example.com")
//	return ctx.Redirect("/login", http.StatusFound)
func (c *contextWrapper) Redirect(location string, status ...int) error {
	return c.fiberCtx.Redirect(location, status...)
}

// Accepts checks if the specified content types are acceptable by the client.
//
// Input:
//   - offers: One or more content types to check (e.g., "application/json", "text/html")
//
// Output:
//   - string: The first acceptable content type, or empty string if none are acceptable
//
// Example:
//
//	contentType := ctx.Accepts("application/json", "text/html")
//	if contentType == "application/json" {
//	    return ctx.JSON(data)
//	}
func (c *contextWrapper) Accepts(offers ...string) string {
	return c.fiberCtx.Accepts(offers...)
}

// AcceptsCharsets checks if the specified character sets are acceptable by the client.
//
// Input:
//   - offers: One or more character sets to check (e.g., "utf-8", "iso-8859-1")
//
// Output:
//   - string: The first acceptable charset, or empty string if none are acceptable
//
// Example:
//
//	charset := ctx.AcceptsCharsets("utf-8", "iso-8859-1")
//	logger.Info("accepted charset", "charset", charset)
func (c *contextWrapper) AcceptsCharsets(offers ...string) string {
	return c.fiberCtx.AcceptsCharsets(offers...)
}

// AcceptsEncodings checks if the specified encodings are acceptable by the client.
//
// Input:
//   - offers: One or more encodings to check (e.g., "gzip", "deflate", "br")
//
// Output:
//   - string: The first acceptable encoding, or empty string if none are acceptable
//
// Example:
//
//	encoding := ctx.AcceptsEncodings("gzip", "deflate")
//	logger.Info("accepted encoding", "encoding", encoding)
func (c *contextWrapper) AcceptsEncodings(offers ...string) string {
	return c.fiberCtx.AcceptsEncodings(offers...)
}

// AcceptsLanguages checks if the specified languages are acceptable by the client.
//
// Input:
//   - offers: One or more language codes to check (e.g., "en", "fr", "vi")
//
// Output:
//   - string: The first acceptable language, or empty string if none are acceptable
//
// Example:
//
//	lang := ctx.AcceptsLanguages("en", "vi")
//	if lang == "vi" {
//	    // Return Vietnamese response
//	}
func (c *contextWrapper) AcceptsLanguages(offers ...string) string {
	return c.fiberCtx.AcceptsLanguages(offers...)
}

// Fresh returns true when the response is still "fresh" (not stale).
// This is determined by comparing the request's If-None-Match and If-Modified-Since headers
// with the response's ETag and Last-Modified headers.
//
// Output:
//   - bool: True if the response is fresh, false otherwise
//
// Example:
//
//	if ctx.Fresh() {
//	    return ctx.Status(http.StatusNotModified).SendString("")
//	}
func (c *contextWrapper) Fresh() bool {
	return c.fiberCtx.Fresh()
}

// Stale returns true when the response is "stale" (not fresh).
//
// Output:
//   - bool: True if the response is stale, false otherwise
//
// Example:
//
//	if ctx.Stale() {
//	    // Generate new response
//	}
func (c *contextWrapper) Stale() bool {
	return c.fiberCtx.Stale()
}

// XHR returns true if the request's X-Requested-With header field is "XMLHttpRequest".
// This is commonly used to detect AJAX requests.
//
// Output:
//   - bool: True if the request is an AJAX request, false otherwise
//
// Example:
//
//	if ctx.XHR() {
//	    // Handle AJAX request differently
//	}
func (c *contextWrapper) XHR() bool {
	return c.fiberCtx.XHR()
}

// Locals stores and retrieves values scoped to the current request.
// Values stored in Locals are available throughout the request lifecycle
// and are automatically merged into context.Context when ctx.Context() is called.
//
// Input:
//   - key: The key to store or retrieve
//   - value: Optional value to store (if provided, stores the value)
//
// Output:
//   - interface{}: The stored value if retrieving, or the value that was stored if setting
//
// Example:
//
//	// Store a value
//	ctx.Locals("user_id", "123")
//	ctx.Locals("lang", "vi")
//
//	// Retrieve a value
//	userID := ctx.Locals("user_id")
//	lang := ctx.Locals("lang")
func (c *contextWrapper) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		c.fiberCtx.Locals(key, value[0])
		// Track the key when setting a value
		c.trackedKeys[key] = struct{}{}
		return value[0]
	}
	// Track the key when getting a value (in case it was set outside our wrapper)
	// This helps discover keys that were set directly on fiberCtx
	if c.fiberCtx.Locals(key) != nil {
		c.trackedKeys[key] = struct{}{}
	}
	return c.fiberCtx.Locals(key)
}

// GetAllLocals returns all Locals keys and values as a map.
// This allows iteration through all custom values stored in Locals.
//
// Output:
//   - map[string]interface{}: A map of all Locals keys and values
//
// Example:
//
//	allLocals := ctx.GetAllLocals()
//	for key, value := range allLocals {
//	    logger.Infow("local value", "key", key, "value", value)
//	}
func (c *contextWrapper) GetAllLocals() map[string]interface{} {
	result := make(map[string]interface{})
	// Iterate through tracked keys and retrieve their values
	for key := range c.trackedKeys {
		if value := c.fiberCtx.Locals(key); value != nil {
			result[key] = value
		}
	}
	return result
}

// Next executes the next handler in the middleware chain.
// Call this in middleware functions to pass control to the next handler.
// If not called, the request handling stops at the current middleware.
//
// Output:
//   - error: Any error returned by subsequent handlers
//
// Example:
//
//	middleware := aurelion.Middleware(func(ctx aurelion.Context) error {
//	    logger.Info("before handler")
//	    err := ctx.Next()
//	    logger.Info("after handler")
//	    return err
//	})
func (c *contextWrapper) Next() error {
	return c.fiberCtx.Next()
}

// Context returns the underlying standard context.Context.
// It automatically merges all string values from Locals into the standard context.Context.
// This allows seamless integration with standard context.Context APIs and libraries.
// The merging is done dynamically, so all custom keys set in Locals are automatically included.
//
// Output:
//   - context.Context: The standard context.Context with all Locals values merged
//
// Example:
//
//	stdCtx := ctx.Context()
//	userID := contextutil.GetUserIDFromContext(stdCtx)
//	lang := contextutil.GetLanguageFromContext(stdCtx)
func (c *contextWrapper) Context() context.Context {
	reqCtx := c.fiberCtx.Context()
	var resultCtx context.Context = reqCtx

	// Get all Locals keys and values using tracked keys
	allLocals := c.GetAllLocals()

	// Merge all string values into context (only strings can be used in context.Context)
	for key, value := range allLocals {
		if strValue, ok := value.(string); ok {
			typedKey := getTypedContextKey(key)
			resultCtx = context.WithValue(resultCtx, typedKey, strValue)
		}
	}

	return resultCtx
}

// IsMethod checks if the request method matches the given method.
//
// Input:
//   - method: The HTTP method to check (e.g., "GET", "POST", "PUT")
//
// Output:
//   - bool: True if the request method matches, false otherwise
//
// Example:
//
//	if ctx.IsMethod("POST") {
//	    // Handle POST request
//	}
func (c *contextWrapper) IsMethod(method string) bool {
	return c.fiberCtx.Method() == method
}

// RequestID retrieves the request ID from the context.
// This is a convenience method that wraps GetRequestID.
//
// Output:
//   - string: The request ID
func (c *contextWrapper) RequestID() string {
	return GetRequestID(c)
}
