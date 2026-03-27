// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/tracing"
)

// SlowRequestDetector creates a middleware that logs a warning when a request
// exceeds the given duration threshold. This enables rapid identification of
// performance degradation in production without requiring external tooling.
//
// The log includes request_id, trace_id, method, route path, actual duration,
// status, and threshold for immediate incident correlation.
func SlowRequestDetector(threshold time.Duration, log ...*logger.Logger) core.Middleware {
	l := defaultLog
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	}
	thresholdMs := threshold.Milliseconds()
	return func(ctx core.Context) error {
		start := time.Now()

		err := ctx.Next()

		duration := time.Since(start)
		if duration >= threshold {
			traceID := tracing.TraceIDFromContext(ctx.Context())
			l.Warnw("slow request detected",
				"request_id", ctx.RequestID(),
				"trace_id", traceID,
				"method", ctx.Method(),
				"path", ctx.RoutePath(),
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", thresholdMs,
				"status", ctx.ResponseStatusCode(),
			)
		}

		return err
	}
}
