// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"time"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// Middleware Composition

// Chain combines multiple middleware functions into a single middleware.
// The middleware are executed in the order they are provided.
//
// IMPORTANT: Middleware used inside Chain should NOT call ctx.Next() themselves.
// Chain calls each middleware sequentially — if a middleware also calls
// ctx.Next(), the Fiber handler index will advance twice, skipping handlers.
// Middleware that needs to run code before AND after downstream handlers
// should use the Before/After combinators instead of Chain.
//
// If no middlewares are provided, Chain passes control to the next handler.
func Chain(middlewares ...core.Middleware) core.Middleware {
	return func(ctx core.Context) error {
		if len(middlewares) == 0 {
			return ctx.Next()
		}

		// Execute all middleware sequentially.
		// Each middleware's ctx.Next() call naturally advances Fiber's internal
		// handler index, so we just call them in sequence. If a middleware
		// does NOT call ctx.Next(), we stop (short-circuit behavior).
		for _, m := range middlewares {
			if err := m(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// Common Middleware Stacks

// Optional applies middleware only if the condition function returns true.
func Optional(condition func(core.Context) bool, middleware core.Middleware) core.Middleware {
	return func(ctx core.Context) error {
		if condition(ctx) {
			return middleware(ctx)
		}
		return ctx.Next()
	}
}

// OnlyForMethods applies middleware only for specific HTTP methods.
func OnlyForMethods(middleware core.Middleware, methods ...string) core.Middleware {
	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method] = true
	}

	return func(ctx core.Context) error {
		if methodMap[ctx.Method()] {
			return middleware(ctx)
		}
		return ctx.Next()
	}
}

// SkipForPaths skips middleware for specific paths (exact match).
func SkipForPaths(middleware core.Middleware, paths ...string) core.Middleware {
	pathMap := make(map[string]bool)
	for _, path := range paths {
		pathMap[path] = true
	}

	return func(ctx core.Context) error {
		if pathMap[ctx.Path()] {
			return ctx.Next()
		}
		return middleware(ctx)
	}
}

// SkipForPathPrefixes skips middleware for paths matching any of the given prefixes.
// For example, SkipForPathPrefixes(mw, "/health") will skip for /health, /health/ready, etc.
func SkipForPathPrefixes(middleware core.Middleware, prefixes ...string) core.Middleware {
	return func(ctx core.Context) error {
		p := ctx.Path()
		for _, prefix := range prefixes {
			if len(p) >= len(prefix) && p[:len(prefix)] == prefix {
				return ctx.Next()
			}
		}
		return middleware(ctx)
	}
}

// Before executes a function before the middleware
func Before(middleware core.Middleware, beforeFunc func(core.Context)) core.Middleware {
	return func(ctx core.Context) error {
		beforeFunc(ctx)
		return middleware(ctx)
	}
}

// After executes a function after the middleware
func After(middleware core.Middleware, afterFunc func(core.Context, error)) core.Middleware {
	return func(ctx core.Context) error {
		err := middleware(ctx)
		afterFunc(ctx, err)
		return err
	}
}

// Recover wraps middleware with panic recovery.
// The recovered error is returned to upstream middleware (e.g., logging, metrics)
// so the panic is visible in the middleware chain.
func Recover(mw core.Middleware) core.Middleware {
	return func(ctx core.Context) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				errResp := core.NewErrorResponse("INTERNAL_ERROR", core.StatusInternalServerError, "Internal server error")
				_ = core.SendError(ctx, errResp)
				returnErr = errResp
			}
		}()
		return mw(ctx)
	}
}

// Timeout wraps middleware with a timeout.
// The middleware runs synchronously on the same goroutine (safe for Fiber's
// non-thread-safe context). Handlers should check ctx.Context().Done() to
// cooperatively abort long-running work when the deadline fires.
func Timeout(middleware core.Middleware, timeout time.Duration) core.Middleware {
	return func(ctx core.Context) error {
		timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()

		// Wrap the context so the handler sees the deadline via ctx.Context()
		wrappedCtx := &timeoutContextWrapper{inner: ctx, timeoutCtx: timeoutCtx}

		err := middleware(wrappedCtx)

		// If the deadline fired during execution, return a timeout error
		if timeoutCtx.Err() == context.DeadlineExceeded {
			errResp := core.NewErrorResponse("TIMEOUT", core.StatusGatewayTimeout, "Request timeout")
			return core.SendError(ctx, errResp)
		}

		return err
	}
}

// timeoutContextWrapper wraps a core.Context and overrides Context() to return
// a deadline-aware context, so handlers can check ctx.Context().Done() for
// cooperative cancellation. We cannot embed core.Context directly because
// it has a method named Context() which would conflict with the embedded field name.
// Status() is overridden to return the wrapper for proper method chaining.
type timeoutContextWrapper struct {
	inner      core.Context
	timeoutCtx context.Context
}

// Compile-time interface check
var _ core.Context = (*timeoutContextWrapper)(nil)

func (w *timeoutContextWrapper) Context() context.Context     { return w.timeoutCtx }
func (w *timeoutContextWrapper) Status(code int) core.Context { w.inner.Status(code); return w }

// Delegated methods — all 1:1 forwarded to inner context
func (w *timeoutContextWrapper) Next() error                       { return w.inner.Next() }
func (w *timeoutContextWrapper) Method() string                    { return w.inner.Method() }
func (w *timeoutContextWrapper) Path() string                      { return w.inner.Path() }
func (w *timeoutContextWrapper) RoutePath() string                 { return w.inner.RoutePath() }
func (w *timeoutContextWrapper) OriginalURL() string               { return w.inner.OriginalURL() }
func (w *timeoutContextWrapper) BaseURL() string                   { return w.inner.BaseURL() }
func (w *timeoutContextWrapper) Protocol() string                  { return w.inner.Protocol() }
func (w *timeoutContextWrapper) Hostname() string                  { return w.inner.Hostname() }
func (w *timeoutContextWrapper) IP() string                        { return w.inner.IP() }
func (w *timeoutContextWrapper) Secure() bool                      { return w.inner.Secure() }
func (w *timeoutContextWrapper) Get(k string, dv ...string) string { return w.inner.Get(k, dv...) }
func (w *timeoutContextWrapper) Set(k, v string)                   { w.inner.Set(k, v) }
func (w *timeoutContextWrapper) Append(k string, v ...string)      { w.inner.Append(k, v...) }
func (w *timeoutContextWrapper) Params(k string, dv ...string) string {
	return w.inner.Params(k, dv...)
}
func (w *timeoutContextWrapper) AllParams() map[string]string        { return w.inner.AllParams() }
func (w *timeoutContextWrapper) ParamsParser(out any) error          { return w.inner.ParamsParser(out) }
func (w *timeoutContextWrapper) Query(k string, dv ...string) string { return w.inner.Query(k, dv...) }
func (w *timeoutContextWrapper) AllQueries() map[string]string       { return w.inner.AllQueries() }
func (w *timeoutContextWrapper) QueryParser(out any) error           { return w.inner.QueryParser(out) }
func (w *timeoutContextWrapper) Body() []byte                        { return w.inner.Body() }
func (w *timeoutContextWrapper) BodyParser(out any) error            { return w.inner.BodyParser(out) }
func (w *timeoutContextWrapper) Cookies(k string, dv ...string) string {
	return w.inner.Cookies(k, dv...)
}
func (w *timeoutContextWrapper) Cookie(c *core.Cookie)      { w.inner.Cookie(c) }
func (w *timeoutContextWrapper) ClearCookie(keys ...string) { w.inner.ClearCookie(keys...) }
func (w *timeoutContextWrapper) UseProperHTTPStatus() bool  { return w.inner.UseProperHTTPStatus() }
func (w *timeoutContextWrapper) ResponseStatusCode() int    { return w.inner.ResponseStatusCode() }
func (w *timeoutContextWrapper) JSON(data any) error        { return w.inner.JSON(data) }
func (w *timeoutContextWrapper) XML(data any) error         { return w.inner.XML(data) }
func (w *timeoutContextWrapper) SendString(s string) error  { return w.inner.SendString(s) }
func (w *timeoutContextWrapper) SendBytes(b []byte) error   { return w.inner.SendBytes(b) }
func (w *timeoutContextWrapper) Redirect(loc string, s ...int) error {
	return w.inner.Redirect(loc, s...)
}
func (w *timeoutContextWrapper) Accepts(types ...string) string { return w.inner.Accepts(types...) }
func (w *timeoutContextWrapper) AcceptsCharsets(cs ...string) string {
	return w.inner.AcceptsCharsets(cs...)
}
func (w *timeoutContextWrapper) AcceptsEncodings(en ...string) string {
	return w.inner.AcceptsEncodings(en...)
}
func (w *timeoutContextWrapper) AcceptsLanguages(lg ...string) string {
	return w.inner.AcceptsLanguages(lg...)
}
func (w *timeoutContextWrapper) Fresh() bool { return w.inner.Fresh() }
func (w *timeoutContextWrapper) Stale() bool { return w.inner.Stale() }
func (w *timeoutContextWrapper) XHR() bool   { return w.inner.XHR() }
func (w *timeoutContextWrapper) Locals(key string, value ...any) any {
	return w.inner.Locals(key, value...)
}
func (w *timeoutContextWrapper) GetAllLocals() map[string]any   { return w.inner.GetAllLocals() }
func (w *timeoutContextWrapper) IsMethod(m string) bool         { return w.inner.IsMethod(m) }
func (w *timeoutContextWrapper) RequestID() string              { return w.inner.RequestID() }
func (w *timeoutContextWrapper) OK(data any) error              { return w.inner.OK(data) }
func (w *timeoutContextWrapper) Created(data any) error         { return w.inner.Created(data) }
func (w *timeoutContextWrapper) NoContent() error               { return w.inner.NoContent() }
func (w *timeoutContextWrapper) BadRequestMsg(msg string) error { return w.inner.BadRequestMsg(msg) }
func (w *timeoutContextWrapper) UnauthorizedMsg(msg string) error {
	return w.inner.UnauthorizedMsg(msg)
}
func (w *timeoutContextWrapper) ForbiddenMsg(msg string) error { return w.inner.ForbiddenMsg(msg) }
func (w *timeoutContextWrapper) NotFoundMsg(msg string) error  { return w.inner.NotFoundMsg(msg) }
func (w *timeoutContextWrapper) InternalErrorMsg(msg string) error {
	return w.inner.InternalErrorMsg(msg)
}
