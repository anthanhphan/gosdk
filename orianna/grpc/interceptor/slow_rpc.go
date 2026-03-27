// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/tracing"
)

// SlowRPCDetector creates a unary interceptor that logs a warning when an RPC
// exceeds the given duration threshold. This enables rapid identification of
// performance degradation in production.
// Accepts optional *logger.Logger; defaults to package-level interceptorLog.
func SlowRPCDetector(threshold time.Duration, log ...*logger.Logger) grpc.UnaryServerInterceptor {
	l := interceptorLog
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	}
	thresholdMs := threshold.Milliseconds()
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if duration >= threshold {
			reqID := extractRequestIDFromMD(ctx)
			traceID := tracing.TraceIDFromContext(ctx)
			l.Warnw("slow RPC detected",
				"request_id", reqID,
				"trace_id", traceID,
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", thresholdMs,
			)
		}

		return resp, err
	}
}

// StreamSlowRPCDetector creates a stream interceptor that logs a warning when
// a stream's total duration exceeds the given threshold.
// Accepts optional *logger.Logger; defaults to package-level interceptorLog.
func StreamSlowRPCDetector(threshold time.Duration, log ...*logger.Logger) grpc.StreamServerInterceptor {
	l := interceptorLog
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	}
	thresholdMs := threshold.Milliseconds()
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		if duration >= threshold {
			reqID := ""
			traceID := ""
			if ss != nil {
				ctx := ss.Context()
				reqID = extractRequestIDFromMD(ctx)
				traceID = tracing.TraceIDFromContext(ctx)
			}
			l.Warnw("slow stream RPC detected",
				"request_id", reqID,
				"trace_id", traceID,
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", thresholdMs,
			)
		}

		return err
	}
}
