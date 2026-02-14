// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"errors"
	"strconv"
	"time"

	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// MetricsMiddleware creates a middleware that records HTTP metrics using the provided client.
// Uses ctx.RoutePath() instead of ctx.Path() to record route patterns (e.g., "/users/:id")
// rather than actual paths (e.g., "/users/123"), preventing unbounded Prometheus cardinality.
func MetricsMiddleware(client metrics.Client, subsystem string) core.Middleware {
	// Optimization: Pre-calculate metric names to avoid concatenation on every request
	requestsTotalName := subsystem + "_requests_total"
	requestDurationName := subsystem + "_request_duration_seconds"

	return func(ctx core.Context) error {
		start := time.Now()

		err := ctx.Next()

		statusCode := ctx.ResponseStatusCode()
		status := strconv.Itoa(statusCode)
		errorCode := "none"

		if err != nil {
			var errResp *core.ErrorResponse
			if errors.As(err, &errResp) {
				errorCode = errResp.Code
			} else {
				errorCode = "INTERNAL_ERROR"
			}
		}

		// Record request count
		client.Inc(ctx.Context(), requestsTotalName,
			"method", ctx.Method(),
			"path", ctx.RoutePath(),
			"status", status,
			"error_code", errorCode,
		)

		// Record latency
		client.Duration(ctx.Context(), requestDurationName, start,
			"method", ctx.Method(),
			"path", ctx.RoutePath(),
			"status", status,
			"error_code", errorCode,
		)

		return err
	}
}
