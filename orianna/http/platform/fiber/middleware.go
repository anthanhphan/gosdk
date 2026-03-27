// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/middleware"
	"github.com/anthanhphan/gosdk/orianna/shared/ctxkeys"
	"github.com/anthanhphan/gosdk/orianna/shared/requestid"
	"github.com/anthanhphan/gosdk/tracing"
	"github.com/gofiber/fiber/v3"
)

// MaxLogPayloadBytes is the maximum size of request/response bodies logged.
// Prevents memory bloat, log storage costs, and accidental PII leakage.
const MaxLogPayloadBytes = 4096

// requestIDMiddleware creates middleware that generates and propagates request IDs.
// Uses shared/requestid.IsValid to reject crafted/malicious request IDs.
func requestIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		reqID := c.Get(core.HeaderRequestID)
		if !requestid.IsValid(reqID) {
			reqID = requestid.Generate()
		}

		c.Set(core.HeaderRequestID, reqID)
		c.Locals(ctxkeys.RequestID.Key(), reqID)
		return c.Next()
	}
}

// traceIDMiddleware creates middleware that generates or reuses trace ID from header
func traceIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if existingTraceID := c.Locals(ctxkeys.TraceID.Key()); existingTraceID != nil {
			if idStr, ok := existingTraceID.(string); ok && idStr != "" {
				c.Response().Header.Set(core.HeaderTraceID, idStr)
				return c.Next()
			}
		}

		traceID := c.Get(core.HeaderTraceID)
		if traceID == "" {
			traceID = c.Get(core.HeaderXB3TraceID)
			if traceID == "" {
				traceID = c.Get(core.HeaderTraceparent)
				if traceID != "" && len(traceID) >= core.TraceparentTraceIDEnd && traceID[2] == core.TraceparentVersionSeparator {
					traceID = traceID[core.TraceparentTraceIDStart:core.TraceparentTraceIDEnd]
				}
			}
		}

		if traceID == "" {
			traceID = requestid.Generate()
		}

		c.Locals(ctxkeys.TraceID.Key(), traceID)
		c.Set(core.HeaderTraceID, traceID)
		return c.Next()
	}
}

// truncatePayload truncates a string to MaxLogPayloadBytes.
func truncatePayload(s string) string {
	if len(s) > MaxLogPayloadBytes {
		return s[:MaxLogPayloadBytes] + "...(truncated)"
	}
	return s
}

// logFieldsPool pools []any slices for structured log fields.
// Capacity 16 covers all request + response fields without reallocation.
var logFieldsPool = sync.Pool{
	New: func() any {
		s := make([]any, 0, 16)
		return &s
	},
}

func acquireLogFields() []any {
	p := logFieldsPool.Get().(*[]any)
	return (*p)[:0]
}

func releaseLogFields(fields []any) {
	// Clear references to allow GC
	for i := range fields {
		fields[i] = nil
	}
	fields = fields[:0]
	logFieldsPool.Put(&fields)
}

// requestResponseLoggingMiddleware creates middleware that logs request and response information.
// Supports prefix-based skip paths (e.g., "/health" skips /health, /health/ready, etc.).
// Uses Warnw for error responses (>= 400) and Infow for success responses.
func requestResponseLoggingMiddleware(log *logger.Logger, verbose bool, skipPaths []string) fiber.Handler {
	// Separate exact paths and prefixes for efficient matching
	exactSkip := make(map[string]struct{})
	var prefixSkip []string
	for _, path := range skipPaths {
		if strings.HasSuffix(path, "*") {
			// "/health/*" -> prefix "/health/"
			prefixSkip = append(prefixSkip, path[:len(path)-1])
		} else {
			exactSkip[path] = struct{}{}
			// Also treat as prefix: "/health" matches "/health" and "/health/ready"
			prefixSkip = append(prefixSkip, path)
		}
	}

	// Sort prefixes for binary search
	sort.Strings(prefixSkip)

	return func(c fiber.Ctx) error {
		reqPath := c.Path()

		// Fast path: exact match
		if _, skip := exactSkip[reqPath]; skip {
			return c.Next()
		}

		// Binary search: O(log n) prefix match
		i := sort.SearchStrings(prefixSkip, reqPath)
		if i < len(prefixSkip) && strings.HasPrefix(reqPath, prefixSkip[i]) {
			return c.Next()
		}
		if i > 0 && strings.HasPrefix(reqPath, prefixSkip[i-1]) {
			return c.Next()
		}

		start := time.Now()
		requestID := getRequestIDFromContext(c)
		traceID := getTraceIDFromContext(c)

		// Use pooled log field slices to avoid 2x make per request
		reqFields := buildRequestLogFields(acquireLogFields(), c, verbose, requestID, traceID)
		log.Infow("incoming request", reqFields...)
		releaseLogFields(reqFields)

		err := c.Next()

		duration := time.Since(start)
		respFields := buildResponseLogFields(acquireLogFields(), c, verbose, duration, requestID, traceID, err)

		// Log severity by status code: Warnw for >= 400, Infow for < 400
		statusCode := c.Response().StatusCode()
		if err != nil || statusCode >= 400 {
			log.Warnw("request completed", respFields...)
		} else {
			log.Infow("request completed", respFields...)
		}
		releaseLogFields(respFields)

		return err
	}
}

// getRequestIDFromContext extracts request ID from fiber context
func getRequestIDFromContext(c fiber.Ctx) string {
	requestID, ok := c.Locals(ctxkeys.RequestID.Key()).(string)
	if !ok {
		return core.DefaultUnknownRequestID
	}
	return requestID
}

// getTraceIDFromContext extracts trace ID from fiber context.
// Prefers OTel trace_id when available, falls back to legacy locals.
func getTraceIDFromContext(c fiber.Ctx) string {
	// Prefer OTel trace_id (set by TracingMiddleware)
	// In Fiber v3, Ctx implements context.Context directly
	if otelTraceID := tracing.TraceIDFromContext(c); otelTraceID != "" {
		return otelTraceID
	}
	// Fall back to legacy trace_id local
	traceID, ok := c.Locals(ctxkeys.TraceID.Key()).(string)
	if !ok {
		return ""
	}
	return traceID
}

// buildRequestLogFields appends request log fields to the provided slice (avoids allocation).
func buildRequestLogFields(fields []any, c fiber.Ctx, verbose bool, requestID, traceID string) []any {
	fields = append(fields,
		"request_id", requestID,
		"trace_id", traceID,
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"user-agent", c.Get("User-Agent"),
	)

	if verbose {
		if queryString := string(c.Request().URI().QueryString()); queryString != "" {
			fields = append(fields, "query", queryString)
		}

		route := c.Route()
		if route != nil && len(route.Params) > 0 {
			// Only allocate params map when params exist
			params := make(map[string]string, len(route.Params))
			for _, param := range route.Params {
				params[param] = c.Params(param)
			}
			fields = append(fields, "params", params)
		}

		// Log request headers with sensitive values redacted
		headers := make(map[string]string)
		for key, value := range c.Request().Header.All() {
			k := string(key)
			v := string(value)
			headers[k] = middleware.SanitizeHeaderValue(k, v)
		}
		if len(headers) > 0 {
			fields = append(fields, "headers", headers)
		}

		// Check raw byte length first to avoid string conversion on empty body
		if rawBody := c.Body(); len(rawBody) > 0 {
			fields = append(fields, "body", truncatePayload(string(rawBody)))
		}
	}

	return fields
}

// buildResponseLogFields appends response log fields to the provided slice (avoids allocation).
func buildResponseLogFields(fields []any, c fiber.Ctx, verbose bool, duration time.Duration, requestID, traceID string, err error) []any {
	statusCode := c.Response().StatusCode()
	var errorResponse any

	// If error occurred, try to determine the correct status code
	if err != nil {
		var e *fiber.Error
		if errors.As(err, &e) {
			statusCode = e.Code
			errorResponse = e
		} else {
			// Check for core.ErrorResponse
			var ce *core.ErrorResponse
			if errors.As(err, &ce) {
				statusCode = ce.HTTPStatus
				errorResponse = ce
			}
		}

		// If status is still success but we have an error, default to 500
		if statusCode >= 200 && statusCode < 300 {
			statusCode = fiber.StatusInternalServerError
		}
	}

	fields = append(fields,
		"request_id", requestID,
		"trace_id", traceID,
		"method", c.Method(),
		"path", c.Path(),
		"http_code", statusCode,
		"duration_ms", duration.Milliseconds(),
	)

	if verbose {
		// Check raw byte length first to avoid string conversion on empty body
		rawBody := c.Response().Body()
		var responseBody string
		if len(rawBody) > 0 {
			responseBody = string(rawBody)
		}

		// If empty and we have an error, marshal the error object using jcodec
		if responseBody == "" && err != nil {
			target := errorResponse
			if target == nil {
				target = err // Fallback to error interface
			}

			// Try to marshal using jcodec
			if jsonStr, je := jcodec.CompactString(target); je == nil {
				responseBody = jsonStr
			} else if target != nil {
				// Fallback to error string if marshal fails
				responseBody = err.Error()
			}
		}

		if responseBody != "" {
			fields = append(fields, "response", truncatePayload(responseBody))
		}
	}

	return fields
}
