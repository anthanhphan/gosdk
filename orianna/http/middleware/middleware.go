// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/shared/ctxkeys"
	"github.com/anthanhphan/gosdk/utils"
)

// Middleware Composition

// Chain combines multiple middleware functions into a single middleware.
// The middleware are executed in the order they are provided.
//
// IMPORTANT: Middleware used inside Chain should NOT call ctx.Next() themselves.
// Chain calls each middleware sequentially -- if a middleware also calls
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
	methodSet := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		methodSet[method] = struct{}{}
	}

	return func(ctx core.Context) error {
		if _, ok := methodSet[ctx.Method()]; ok {
			return middleware(ctx)
		}
		return ctx.Next()
	}
}

// SkipForPaths skips middleware for specific paths (exact match).
func SkipForPaths(middleware core.Middleware, paths ...string) core.Middleware {
	pathSet := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		pathSet[path] = struct{}{}
	}

	return func(ctx core.Context) error {
		if _, ok := pathSet[ctx.Path()]; ok {
			return ctx.Next()
		}
		return middleware(ctx)
	}
}

// SkipForPathPrefixes skips middleware for paths matching any of the given prefixes.
// For example, SkipForPathPrefixes(mw, "/health") will skip for /health, /health/ready, etc.
// Prefixes are sorted at construction time and matched via binary search (O(log n) per request).
func SkipForPathPrefixes(middleware core.Middleware, prefixes ...string) core.Middleware {
	sorted := make([]string, len(prefixes))
	copy(sorted, prefixes)
	sort.Strings(sorted)

	return func(ctx core.Context) error {
		p := ctx.Path()
		// Binary search: find the first prefix >= p
		i := sort.SearchStrings(sorted, p)

		// Check sorted[i]: exact match or p has this prefix
		if i < len(sorted) && strings.HasPrefix(p, sorted[i]) {
			return ctx.Next()
		}
		// Check sorted[i-1]: p may start with a shorter prefix
		if i > 0 && strings.HasPrefix(p, sorted[i-1]) {
			return ctx.Next()
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
// Captures the goroutine stack trace and logs it server-side for debugging.
// The recovered error is returned to upstream middleware (e.g., logging, metrics)
// so the panic is visible in the middleware chain.
// Includes request_id and trace_id for incident correlation.
// Accepts an optional *logger.Logger; defaults to package-level logger.
func Recover(mw core.Middleware, log ...*logger.Logger) core.Middleware {
	l := defaultLog
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	}
	return func(ctx core.Context) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				location, _ := utils.GetPanicLocation()
				requestID := ctx.RequestID()
				traceID, _ := ctx.Locals(ctxkeys.TraceID.Key()).(string)
				l.Errorw("panic recovered",
					"error", fmt.Sprint(r),
					"location", location,
					"path", ctx.Path(),
					"method", ctx.Method(),
					"request_id", requestID,
					"trace_id", traceID,
				)
				// Use NewErrorResponse (not pool-based) because errResp escapes
				// via returnErr and may be inspected by upstream middleware.
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
		origCtx := ctx.Context()
		timeoutCtx, cancel := context.WithTimeout(origCtx, timeout)
		defer cancel()

		// Inject deadline-aware context so handlers see it via ctx.Context()
		ctx.SetContext(timeoutCtx)

		err := middleware(ctx)

		// Restore original context
		ctx.SetContext(origCtx)

		// If the deadline fired during execution, return a timeout error
		if timeoutCtx.Err() == context.DeadlineExceeded {
			errResp := core.NewErrorResponse("TIMEOUT", core.StatusGatewayTimeout, "Request timeout")
			_ = core.SendError(ctx, errResp)
			return errResp
		}

		return err
	}
}
