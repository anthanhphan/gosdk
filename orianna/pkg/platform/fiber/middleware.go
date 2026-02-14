// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"errors"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// logFieldPool reduces per-request allocations for log field slices.
var logFieldPool = sync.Pool{
	New: func() any {
		s := make([]any, 0, 16)
		return &s
	},
}

// Header constants for request tracking
const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = core.HeaderRequestID

	// TraceIDHeader is the header name for trace ID
	TraceIDHeader = core.HeaderTraceID
)

// requestIDMiddleware creates middleware that generates and propagates request IDs
func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			uuidV7, err := uuid.NewV7()
			if err != nil {
				requestID = uuid.NewString()
			} else {
				requestID = uuidV7.String()
			}
		}

		c.Set(RequestIDHeader, requestID)
		c.Locals(core.ContextKeyRequestID.Key(), requestID)
		return c.Next()
	}
}

// configMiddleware creates middleware that stores the entire server config in context locals
func configMiddleware(conf *configuration.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(core.ContextKeyConfig.Key(), conf)
		return c.Next()
	}
}

// traceIDMiddleware creates middleware that generates or reuses trace ID from header
func traceIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if existingTraceID := c.Locals(core.ContextKeyTraceID.Key()); existingTraceID != nil {
			if idStr, ok := existingTraceID.(string); ok && idStr != "" {
				c.Response().Header.Set(TraceIDHeader, idStr)
				return c.Next()
			}
		}

		traceID := c.Get(TraceIDHeader)
		if traceID == "" {
			traceID = c.Get(core.HeaderXB3TraceID)
			if traceID == "" {
				traceID = c.Get(core.HeaderTraceparent)
				if traceID != "" && len(traceID) > core.TraceparentPrefixLength && traceID[2] == core.TraceparentVersionSeparator {
					traceID = traceID[core.TraceparentTraceIDStart:core.TraceparentTraceIDEnd]
				}
			}
		}

		if traceID == "" {
			uuidV7, err := uuid.NewV7()
			if err != nil {
				traceID = uuid.NewString()
			} else {
				traceID = uuidV7.String()
			}
		}

		c.Locals(core.ContextKeyTraceID.Key(), traceID)
		c.Set(TraceIDHeader, traceID)
		return c.Next()
	}
}

// requestResponseLoggingMiddleware creates middleware that logs request and response information
func requestResponseLoggingMiddleware(log *logger.Logger, verbose bool, skipPaths []string) fiber.Handler {
	// Build skip path map for O(1) lookup
	skipPathMap := make(map[string]bool, len(skipPaths))
	for _, path := range skipPaths {
		skipPathMap[path] = true
	}

	return func(c *fiber.Ctx) error {
		// Skip logging for configured paths
		if skipPathMap[c.Path()] {
			return c.Next()
		}

		start := time.Now()
		requestID := getRequestIDFromContext(c)
		traceID := getTraceIDFromContext(c)

		// Use pooled slices for log fields
		reqFields := logFieldPool.Get().(*[]any)
		*reqFields = (*reqFields)[:0]
		*reqFields = buildRequestLogFields(*reqFields, c, verbose, requestID, traceID)
		log.Infow("incoming request", *reqFields...)
		logFieldPool.Put(reqFields)

		err := c.Next()

		duration := time.Since(start)
		respFields := logFieldPool.Get().(*[]any)
		*respFields = (*respFields)[:0]
		*respFields = buildResponseLogFields(*respFields, c, verbose, duration, requestID, traceID, err)
		log.Infow("request completed", *respFields...)
		logFieldPool.Put(respFields)

		return err
	}
}

// getRequestIDFromContext extracts request ID from fiber context
func getRequestIDFromContext(c *fiber.Ctx) string {
	requestID, ok := c.Locals(core.ContextKeyRequestID.Key()).(string)
	if !ok {
		return core.DefaultUnknownRequestID
	}
	return requestID
}

// getTraceIDFromContext extracts trace ID from fiber context
func getTraceIDFromContext(c *fiber.Ctx) string {
	traceID, ok := c.Locals(core.ContextKeyTraceID.Key()).(string)
	if !ok {
		return ""
	}
	return traceID
}

// buildRequestLogFields appends request log fields to the provided slice (avoids allocation).
func buildRequestLogFields(fields []any, c *fiber.Ctx, verbose bool, requestID, traceID string) []any {
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

		if params := c.AllParams(); len(params) > 0 {
			fields = append(fields, "params", params)
		}

		if body := string(c.Body()); body != "" {
			fields = append(fields, "body", body)
		}
	}

	return fields
}

// buildResponseLogFields appends response log fields to the provided slice (avoids allocation).
func buildResponseLogFields(fields []any, c *fiber.Ctx, verbose bool, duration time.Duration, requestID, traceID string, err error) []any {
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
				if ce.RequestID == "" {
					ce.RequestID = requestID
				}
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
		// Use a copy of the body to avoid issues
		responseBody := string(c.Response().Body())

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
			fields = append(fields, "response", responseBody)
		}
	}

	return fields
}
