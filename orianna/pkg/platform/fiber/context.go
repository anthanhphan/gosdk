// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"context"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/gofiber/fiber/v2"
)

var contextPool = sync.Pool{
	New: func() any {
		return &ContextAdapter{
			trackedKeys: make(map[string]struct{}),
		}
	},
}

// ContextAdapter wraps fiber.Ctx to implement Context interface
type ContextAdapter struct {
	fiberCtx            *fiber.Ctx
	trackedKeys         map[string]struct{}
	useProperHTTPStatus bool
	cachedCtx           context.Context // lazily built, invalidated on Locals write
	ctxDirty            bool            // true when Locals changed since last Context() call
}

// AcquireContextAdapter acquires a context adapter from the pool
func AcquireContextAdapter(fiberCtx *fiber.Ctx, conf *configuration.Config) *ContextAdapter {
	if fiberCtx == nil {
		return nil
	}
	ctx := contextPool.Get().(*ContextAdapter)
	ctx.Reset(fiberCtx, conf)
	return ctx
}

// ReleaseContextAdapter releases the context adapter back to the pool
func ReleaseContextAdapter(c *ContextAdapter) {
	if c != nil {
		c.reset()
		contextPool.Put(c)
	}
}

// Reset resets the context adapter with new fiber context and config
func (c *ContextAdapter) Reset(fiberCtx *fiber.Ctx, conf *configuration.Config) {
	c.fiberCtx = fiberCtx
	if conf != nil {
		c.useProperHTTPStatus = conf.UseProperHTTPStatus
	} else {
		c.useProperHTTPStatus = false
	}
}

// reset clears the context adapter state
func (c *ContextAdapter) reset() {
	c.fiberCtx = nil
	// Clear tracked keys for reuse (Go 1.21+)
	clear(c.trackedKeys)
	c.useProperHTTPStatus = false
	c.cachedCtx = nil
	c.ctxDirty = false
}

// Method returns the HTTP method of the request
func (c *ContextAdapter) Method() string {
	return c.fiberCtx.Method()
}

// Path returns the route path from the request URL
func (c *ContextAdapter) Path() string {
	return c.fiberCtx.Path()
}

// RoutePath returns the matched route pattern (e.g., "/users/:id" instead of "/users/123")
// This is used for metrics to prevent unbounded cardinality
func (c *ContextAdapter) RoutePath() string {
	route := c.fiberCtx.Route()
	if route != nil {
		return route.Path
	}
	// Fallback to actual path if route is not available
	return c.fiberCtx.Path()
}

// OriginalURL returns the original request URL including query string
func (c *ContextAdapter) OriginalURL() string {
	return c.fiberCtx.OriginalURL()
}

// BaseURL returns the base URL (scheme + host) without path
func (c *ContextAdapter) BaseURL() string {
	return c.fiberCtx.BaseURL()
}

// Protocol returns the protocol of the request
func (c *ContextAdapter) Protocol() string {
	return c.fiberCtx.Protocol()
}

// Hostname returns the hostname from the request
func (c *ContextAdapter) Hostname() string {
	return c.fiberCtx.Hostname()
}

// IP returns the remote IP address of the client
func (c *ContextAdapter) IP() string {
	return c.fiberCtx.IP()
}

// Secure returns true if the connection is secure (HTTPS)
func (c *ContextAdapter) Secure() bool {
	return c.fiberCtx.Secure()
}

// Get returns the HTTP request header field value
func (c *ContextAdapter) Get(key string, defaultValue ...string) string {
	return c.fiberCtx.Get(key, defaultValue...)
}

// Set sets the HTTP response header field
func (c *ContextAdapter) Set(key, value string) {
	c.fiberCtx.Set(key, value)
}

// Append appends the specified value to the HTTP response header field
func (c *ContextAdapter) Append(field string, values ...string) {
	c.fiberCtx.Append(field, values...)
}

// Params returns the route parameter value by key
func (c *ContextAdapter) Params(key string, defaultValue ...string) string {
	return c.fiberCtx.Params(key, defaultValue...)
}

// AllParams returns all route parameters as a map
func (c *ContextAdapter) AllParams() map[string]string {
	return c.fiberCtx.AllParams()
}

// ParamsParser binds the route parameters to a struct
func (c *ContextAdapter) ParamsParser(out any) error {
	return c.fiberCtx.ParamsParser(out)
}

// Query returns the query parameter value by key
func (c *ContextAdapter) Query(key string, defaultValue ...string) string {
	return c.fiberCtx.Query(key, defaultValue...)
}

// AllQueries returns all query parameters as a map
func (c *ContextAdapter) AllQueries() map[string]string {
	return c.fiberCtx.Queries()
}

// QueryParser binds the query parameters to a struct
func (c *ContextAdapter) QueryParser(out any) error {
	return c.fiberCtx.QueryParser(out)
}

// Body returns the raw request body as bytes
func (c *ContextAdapter) Body() []byte {
	return c.fiberCtx.Body()
}

// BodyParser binds the request body to a struct
func (c *ContextAdapter) BodyParser(out any) error {
	return c.fiberCtx.BodyParser(out)
}

// Cookies returns the cookie value by key
func (c *ContextAdapter) Cookies(key string, defaultValue ...string) string {
	return c.fiberCtx.Cookies(key, defaultValue...)
}

// Cookie sets a cookie in the response
func (c *ContextAdapter) Cookie(cookie *core.Cookie) {
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
func (c *ContextAdapter) ClearCookie(key ...string) {
	c.fiberCtx.ClearCookie(key...)
}

// Status sets the HTTP status code for the response
func (c *ContextAdapter) Status(status int) core.Context {
	c.fiberCtx.Status(status)
	return c
}

// UseProperHTTPStatus returns whether to use proper HTTP status codes for errors
func (c *ContextAdapter) UseProperHTTPStatus() bool {
	return c.useProperHTTPStatus
}

// ResponseStatusCode returns the HTTP status code of the response
func (c *ContextAdapter) ResponseStatusCode() int {
	return c.fiberCtx.Response().StatusCode()
}

// JSON sends a JSON response with automatic Content-Type header
func (c *ContextAdapter) JSON(data any) error {
	return c.fiberCtx.JSON(data)
}

// XML sends an XML response with automatic Content-Type header
func (c *ContextAdapter) XML(data any) error {
	return c.fiberCtx.XML(data)
}

// SendString sends a plain text string response
func (c *ContextAdapter) SendString(s string) error {
	return c.fiberCtx.SendString(s)
}

// SendBytes sends a byte array response
func (c *ContextAdapter) SendBytes(b []byte) error {
	return c.fiberCtx.Send(b)
}

// Redirect redirects the client to the specified URL
func (c *ContextAdapter) Redirect(location string, status ...int) error {
	return c.fiberCtx.Redirect(location, status...)
}

// Accepts checks if the specified content types are acceptable by the client
func (c *ContextAdapter) Accepts(offers ...string) string {
	return c.fiberCtx.Accepts(offers...)
}

// AcceptsCharsets checks if the specified character sets are acceptable by the client
func (c *ContextAdapter) AcceptsCharsets(offers ...string) string {
	return c.fiberCtx.AcceptsCharsets(offers...)
}

// AcceptsEncodings checks if the specified encodings are acceptable by the client
func (c *ContextAdapter) AcceptsEncodings(offers ...string) string {
	return c.fiberCtx.AcceptsEncodings(offers...)
}

// AcceptsLanguages checks if the specified languages are acceptable by the client
func (c *ContextAdapter) AcceptsLanguages(offers ...string) string {
	return c.fiberCtx.AcceptsLanguages(offers...)
}

// Fresh returns true when the response is still "fresh" (not stale)
func (c *ContextAdapter) Fresh() bool {
	return c.fiberCtx.Fresh()
}

// Stale returns true when the response is "stale" (not fresh)
func (c *ContextAdapter) Stale() bool {
	return c.fiberCtx.Stale()
}

// XHR returns true if the request's X-Requested-With header field is "XMLHttpRequest"
func (c *ContextAdapter) XHR() bool {
	return c.fiberCtx.XHR()
}

// Locals stores and retrieves values scoped to the current request
func (c *ContextAdapter) Locals(key string, value ...any) any {
	if len(value) > 0 {
		c.fiberCtx.Locals(key, value[0])
		c.trackedKeys[key] = struct{}{}
		c.ctxDirty = true // invalidate cached context
		return value[0]
	}
	if c.fiberCtx.Locals(key) != nil {
		c.trackedKeys[key] = struct{}{}
	}
	return c.fiberCtx.Locals(key)
}

// GetAllLocals returns all Locals keys and values as a map
func (c *ContextAdapter) GetAllLocals() map[string]any {
	result := make(map[string]any)
	for key := range c.trackedKeys {
		if value := c.fiberCtx.Locals(key); value != nil {
			result[key] = value
		}
	}
	return result
}

// Next executes the next handler in the middleware chain
func (c *ContextAdapter) Next() error {
	return c.fiberCtx.Next()
}

// Context returns the underlying standard context.Context.
// The result is cached and only rebuilt when Locals are modified.
func (c *ContextAdapter) Context() context.Context {
	if c.cachedCtx != nil && !c.ctxDirty {
		return c.cachedCtx
	}

	reqCtx := c.fiberCtx.Context()
	var resultCtx context.Context = reqCtx

	allLocals := c.GetAllLocals()
	for key, value := range allLocals {
		if strValue, ok := value.(string); ok {
			typedKey := getTypedContextKey(key)
			resultCtx = context.WithValue(resultCtx, typedKey, strValue)
		}
	}

	c.cachedCtx = resultCtx
	c.ctxDirty = false
	return resultCtx
}

// IsMethod checks if the request method matches the given method
func (c *ContextAdapter) IsMethod(method string) bool {
	return c.fiberCtx.Method() == method
}

// RequestID retrieves the request ID from the context
func (c *ContextAdapter) RequestID() string {
	if id := c.Locals(core.ContextKeyRequestID.Key()); id != nil {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}
	return ""
}

// OK sends a 200 OK response with data wrapped in SuccessResponse
func (c *ContextAdapter) OK(data any) error {
	response := &core.SuccessResponse{
		BaseResponse: core.BaseResponse{
			HTTPStatus: core.StatusOK,
			Code:       "SUCCESS",
			Message:    core.MessageOK,
			Timestamp:  time.Now().UTC(),
			RequestID:  c.RequestID(),
		},
		Data: data,
	}
	return c.Status(core.StatusOK).JSON(response)
}

// Created sends a 201 Created response with data
func (c *ContextAdapter) Created(data any) error {
	response := &core.SuccessResponse{
		BaseResponse: core.BaseResponse{
			HTTPStatus: core.StatusCreated,
			Code:       "SUCCESS",
			Message:    core.MessageCreated,
			Timestamp:  time.Now().UTC(),
			RequestID:  c.RequestID(),
		},
		Data: data,
	}
	return c.Status(core.StatusCreated).JSON(response)
}

// NoContent sends a 204 No Content response
func (c *ContextAdapter) NoContent() error {
	return c.Status(core.StatusNoContent).SendString("")
}

// buildErrorResponse is a helper to build consistent error responses
// This reduces code duplication across error response methods
func (c *ContextAdapter) buildErrorResponse(httpStatus int, apiCode string, message string) error {
	status := core.StatusOK
	if c.useProperHTTPStatus {
		status = httpStatus
	}

	errResp := core.NewErrorResponse(apiCode, httpStatus, message)
	errResp.RequestID = c.RequestID()

	return c.Status(status).JSON(errResp)
}

// BadRequestMsg sends a 400 Bad Request response with a message
func (c *ContextAdapter) BadRequestMsg(message string) error {
	return c.buildErrorResponse(core.StatusBadRequest, "BAD_REQUEST", message)
}

// UnauthorizedMsg sends a 401 Unauthorized response with a message
func (c *ContextAdapter) UnauthorizedMsg(message string) error {
	return c.buildErrorResponse(core.StatusUnauthorized, "UNAUTHORIZED", message)
}

// ForbiddenMsg sends a 403 Forbidden response with a message
func (c *ContextAdapter) ForbiddenMsg(message string) error {
	return c.buildErrorResponse(core.StatusForbidden, "FORBIDDEN", message)
}

// NotFoundMsg sends a 404 Not Found response with a message
func (c *ContextAdapter) NotFoundMsg(message string) error {
	return c.buildErrorResponse(core.StatusNotFound, "NOT_FOUND", message)
}

// InternalErrorMsg sends a 500 Internal Server Error response with a message
func (c *ContextAdapter) InternalErrorMsg(message string) error {
	return c.buildErrorResponse(core.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// getTypedContextKey returns the appropriate typed key for a given string key.
// For known context keys, it returns the proper typed key; otherwise returns the string.
func getTypedContextKey(key string) any {
	switch key {
	case core.ContextKeyRequestID.Key():
		return core.ContextKeyRequestID
	case core.ContextKeyTraceID.Key():
		return core.ContextKeyTraceID
	case core.ContextKeyConfig.Key():
		return core.ContextKeyConfig
	case core.ContextKeyUserID.Key():
		return core.ContextKeyUserID
	case core.ContextKeyTenantID.Key():
		return core.ContextKeyTenantID
	case core.ContextKeyCorrelationID.Key():
		return core.ContextKeyCorrelationID
	default:
		return key
	}
}
