// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/anthanhphan/gosdk/logger"
)

// Package-level logger for interceptors (avoid per-call allocation).
var interceptorLogger = logger.NewLoggerWithFields(logger.String("package", "grpc-client"))

func newUnaryClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, args any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, args, reply, cc, opts...)
		duration := time.Since(start)

		if err != nil {
			interceptorLogger.Warnw("gRPC unary call",
				"service", serviceName,
				"method", method,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
			)
		} else {
			interceptorLogger.Debugw("gRPC unary call",
				"service", serviceName,
				"method", method,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}

func newStreamClientInterceptor(serviceName string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()
		stream, err := streamer(ctx, desc, cc, method, opts...)
		duration := time.Since(start)

		if err != nil {
			interceptorLogger.Warnw("gRPC stream created",
				"service", serviceName,
				"method", method,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
			)
		} else {
			interceptorLogger.Debugw("gRPC stream created",
				"service", serviceName,
				"method", method,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return stream, err
	}
}
