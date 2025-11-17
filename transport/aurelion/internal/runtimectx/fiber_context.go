package runtimectx

import (
	goctx "context"
	"errors"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/anthanhphan/gosdk/transport/aurelion/internal/keys"
	"github.com/gofiber/fiber/v2"
)

// Use centralized key constant.
const trackedLocalsKey = keys.TrackedLocalsKey

// FiberContext wraps fiber.Ctx to implement core.Context.
type FiberContext struct {
	fiberCtx    *fiber.Ctx
	trackedKeys map[string]struct{}
}

// NewFiberContext creates a new context wrapper from fiber context.
func NewFiberContext(fiberCtx *fiber.Ctx) core.Context {
	if fiberCtx == nil {
		return nil
	}
	trackedKeys := getOrCreateTrackedLocalsMap(fiberCtx)
	if trackedKeys == nil {
		trackedKeys = make(map[string]struct{})
	}
	return &FiberContext{
		fiberCtx:    fiberCtx,
		trackedKeys: trackedKeys,
	}
}

// FiberFromContext safely extracts the underlying fiber.Ctx if available.
func FiberFromContext(ctx core.Context) (*fiber.Ctx, bool) {
	if ctx == nil {
		return nil, false
	}
	if fc, ok := ctx.(*FiberContext); ok {
		return fc.fiberCtx, true
	}
	return nil, false
}

// HandlerToFiber converts a core.Handler to fiber.Handler.
func HandlerToFiber(handler core.Handler) fiber.Handler {
	if handler == nil {
		return func(c *fiber.Ctx) error { return nil }
	}
	return func(c *fiber.Ctx) error {
		ctx := NewFiberContext(c)
		if ctx == nil {
			return errors.New("failed to create context wrapper")
		}
		return handler(ctx)
	}
}

// MiddlewareToFiber converts a core.Middleware to fiber.Handler.
func MiddlewareToFiber(middleware core.Middleware) fiber.Handler {
	if middleware == nil {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	return func(c *fiber.Ctx) error {
		ctx := NewFiberContext(c)
		if ctx == nil {
			return errors.New("failed to create context wrapper")
		}
		return middleware(ctx)
	}
}

// TrackFiberLocal marks a fiber local key so it can later be merged into context.Context.
func TrackFiberLocal(c *fiber.Ctx, key string) {
	if c == nil || key == "" {
		return
	}
	tracked := getOrCreateTrackedLocalsMap(c)
	if tracked == nil {
		return
	}
	tracked[key] = struct{}{}
}

func getOrCreateTrackedLocalsMap(c *fiber.Ctx) map[string]struct{} {
	if c == nil {
		return nil
	}
	if tracked, ok := c.Locals(trackedLocalsKey).(map[string]struct{}); ok && tracked != nil {
		return tracked
	}
	tracked := make(map[string]struct{})
	c.Locals(trackedLocalsKey, tracked)
	return tracked
}

func (c *FiberContext) ensureTrackedMap() map[string]struct{} {
	if c.trackedKeys == nil {
		c.trackedKeys = getOrCreateTrackedLocalsMap(c.fiberCtx)
		if c.trackedKeys == nil {
			c.trackedKeys = make(map[string]struct{})
		}
	}
	return c.trackedKeys
}

// Method returns the HTTP method of the request.
func (c *FiberContext) Method() string { return c.fiberCtx.Method() }

// Path returns the route path from the request URL.
func (c *FiberContext) Path() string { return c.fiberCtx.Path() }

// OriginalURL returns the original request URL including query string.
func (c *FiberContext) OriginalURL() string { return c.fiberCtx.OriginalURL() }

// BaseURL returns the base URL (scheme + host) without path.
func (c *FiberContext) BaseURL() string { return c.fiberCtx.BaseURL() }

// Protocol returns the protocol of the request.
func (c *FiberContext) Protocol() string { return c.fiberCtx.Protocol() }

// Hostname returns the hostname from the request.
func (c *FiberContext) Hostname() string { return c.fiberCtx.Hostname() }

// IP returns the remote IP address of the client.
func (c *FiberContext) IP() string { return c.fiberCtx.IP() }

// Secure returns true if the connection is secure (HTTPS).
func (c *FiberContext) Secure() bool { return c.fiberCtx.Secure() }

// Get returns a request header value.
func (c *FiberContext) Get(key string, defaultValue ...string) string {
	return c.fiberCtx.Get(key, defaultValue...)
}

// Set sets a response header value.
func (c *FiberContext) Set(key, value string) { c.fiberCtx.Set(key, value) }

// Append appends a response header value.
func (c *FiberContext) Append(field string, values ...string) { c.fiberCtx.Append(field, values...) }

// Params returns the route parameter value by key.
func (c *FiberContext) Params(key string, defaultValue ...string) string {
	return c.fiberCtx.Params(key, defaultValue...)
}

// AllParams returns all route parameters.
func (c *FiberContext) AllParams() map[string]string { return c.fiberCtx.AllParams() }

// ParamsParser binds route params to a struct.
func (c *FiberContext) ParamsParser(out interface{}) error { return c.fiberCtx.ParamsParser(out) }

// Query returns the query parameter value by key.
func (c *FiberContext) Query(key string, defaultValue ...string) string {
	return c.fiberCtx.Query(key, defaultValue...)
}

// AllQueries returns all query parameters.
func (c *FiberContext) AllQueries() map[string]string { return c.fiberCtx.Queries() }

// QueryParser binds query parameters to a struct.
func (c *FiberContext) QueryParser(out interface{}) error { return c.fiberCtx.QueryParser(out) }

// Body returns the raw request body.
func (c *FiberContext) Body() []byte { return c.fiberCtx.Body() }

// BodyParser binds the request body to a struct.
func (c *FiberContext) BodyParser(out interface{}) error { return c.fiberCtx.BodyParser(out) }

// Cookies returns a cookie value by key.
func (c *FiberContext) Cookies(key string, defaultValue ...string) string {
	return c.fiberCtx.Cookies(key, defaultValue...)
}

// Cookie sets a cookie in the response.
func (c *FiberContext) Cookie(cookie *core.Cookie) {
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
func (c *FiberContext) ClearCookie(key ...string) { c.fiberCtx.ClearCookie(key...) }

// Status sets the HTTP status code for the response.
func (c *FiberContext) Status(status int) core.Context {
	c.fiberCtx.Status(status)
	return c
}

// JSON sends a JSON response.
func (c *FiberContext) JSON(data interface{}) error { return c.fiberCtx.JSON(data) }

// XML sends an XML response.
func (c *FiberContext) XML(data interface{}) error { return c.fiberCtx.XML(data) }

// SendString sends a string response.
func (c *FiberContext) SendString(s string) error { return c.fiberCtx.SendString(s) }

// SendBytes sends a byte slice response.
func (c *FiberContext) SendBytes(b []byte) error { return c.fiberCtx.Send(b) }

// Redirect redirects the client.
func (c *FiberContext) Redirect(location string, status ...int) error {
	return c.fiberCtx.Redirect(location, status...)
}

// Accepts checks acceptable content types.
func (c *FiberContext) Accepts(offers ...string) string { return c.fiberCtx.Accepts(offers...) }

// AcceptsCharsets checks acceptable charsets.
func (c *FiberContext) AcceptsCharsets(offers ...string) string {
	return c.fiberCtx.AcceptsCharsets(offers...)
}

// AcceptsEncodings checks acceptable encodings.
func (c *FiberContext) AcceptsEncodings(offers ...string) string {
	return c.fiberCtx.AcceptsEncodings(offers...)
}

// AcceptsLanguages checks acceptable languages.
func (c *FiberContext) AcceptsLanguages(offers ...string) string {
	return c.fiberCtx.AcceptsLanguages(offers...)
}

// Fresh reports whether the response is still fresh.
func (c *FiberContext) Fresh() bool { return c.fiberCtx.Fresh() }

// Stale reports whether the response is stale.
func (c *FiberContext) Stale() bool { return c.fiberCtx.Stale() }

// XHR reports whether the request was made via XMLHttpRequest.
func (c *FiberContext) XHR() bool { return c.fiberCtx.XHR() }

// Locals stores and retrieves values scoped to the current request.
func (c *FiberContext) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		c.fiberCtx.Locals(key, value[0])
		tracker := c.ensureTrackedMap()
		tracker[key] = struct{}{}
		TrackFiberLocal(c.fiberCtx, key)
		return value[0]
	}
	if v := c.fiberCtx.Locals(key); v != nil {
		tracker := c.ensureTrackedMap()
		tracker[key] = struct{}{}
		TrackFiberLocal(c.fiberCtx, key)
		return v
	}
	return nil
}

// GetAllLocals returns all Locals keys and values as a map.
func (c *FiberContext) GetAllLocals() map[string]interface{} {
	result := make(map[string]interface{})
	for key := range c.ensureTrackedMap() {
		if key == trackedLocalsKey {
			continue
		}
		if value := c.fiberCtx.Locals(key); value != nil {
			result[key] = value
		}
	}
	return result
}

// Next executes the next handler in the middleware chain.
func (c *FiberContext) Next() error { return c.fiberCtx.Next() }

// Context returns the underlying context.Context with Locals merged in.
func (c *FiberContext) Context() goctx.Context {
	reqCtx := c.fiberCtx.Context()
	var resultCtx goctx.Context = reqCtx
	for key, value := range c.GetAllLocals() {
		switch typed := value.(type) {
		case string, int, int64, bool:
			resultCtx = goctx.WithValue(resultCtx, GetContextKey(key), typed)
		}
	}
	return resultCtx
}

// IsMethod checks if the request method matches the provided method.
func (c *FiberContext) IsMethod(method string) bool { return c.fiberCtx.Method() == method }

// RequestID retrieves the request ID from Locals when available.
func (c *FiberContext) RequestID() string {
	if id := c.Locals(keys.ContextKeyRequestID); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// ResetTracking clears tracked locals; useful for tests.
func (c *FiberContext) ResetTracking() {
	c.trackedKeys = make(map[string]struct{})
	TrackFiberLocal(c.fiberCtx, trackedLocalsKey)
}
