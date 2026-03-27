// Copyright (c) 2026 anthanhphan <can.thanhphan.work@gmail.com>

package middleware

import (
	"time"

	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/shared/httputil"
	"github.com/anthanhphan/gosdk/orianna/shared/observability"
)

// statusCodeCache pre-computes string representations of common HTTP status codes
// to avoid strconv.Itoa allocation on the hot path.
var statusCodeCache = observability.CodeStringCache([]int{
	200, 201, 204, 301, 302, 304, 400, 401, 403, 404, 405, 409, 422, 429, 500, 502, 503, 504,
})

// statusString returns a cached string for common codes, falling back to strconv.Itoa.
// Used by both MetricsMiddleware (this file) and TracingMiddleware (tracing.go).
func statusString(code int) string {
	return observability.CodeString(statusCodeCache, code)
}

// MetricsMiddleware creates a middleware that records HTTP metrics using the provided client.
// Uses ctx.RoutePath() instead of ctx.Path() to record route patterns (e.g., "/users/:id")
// rather than actual paths (e.g., "/users/123"), preventing unbounded Prometheus cardinality.
//
// Metrics recorded:
//   - {subsystem}_requests_total: counter with labels method, path, status, error_class
//   - {subsystem}_request_duration_seconds: histogram with labels method, path, status, error_class
//   - {subsystem}_in_flight_requests: gauge tracking concurrent requests
//   - {subsystem}_response_body_bytes: histogram of response payload sizes
func MetricsMiddleware(client metrics.Client, subsystem string) core.Middleware {
	// Optimization: Pre-calculate metric names to avoid concatenation on every request
	requestsTotalName := subsystem + observability.SuffixRequestsTotal
	requestDurationName := subsystem + observability.SuffixRequestDurationSeconds
	inFlightName := subsystem + observability.SuffixInFlightRequests

	return func(ctx core.Context) error {
		// Track in-flight requests
		client.GaugeInc(ctx.Context(), inFlightName)
		defer client.GaugeDec(ctx.Context(), inFlightName)

		start := time.Now()

		err := ctx.Next()

		// Cache values to avoid double method calls
		method := ctx.Method()
		routePath := ctx.RoutePath()
		statusCode := ctx.ResponseStatusCode()
		status := statusString(statusCode)
		errClass := httputil.ErrorClassFromStatus(statusCode)

		// Record request count
		client.Inc(ctx.Context(), requestsTotalName,
			"method", method,
			"path", routePath,
			"status", status,
			"error_class", errClass,
		)

		// Record latency
		client.Duration(ctx.Context(), requestDurationName, start,
			"method", method,
			"path", routePath,
			"status", status,
			"error_class", errClass,
		)

		return err
	}
}
