// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/shared/requestid"
	"github.com/anthanhphan/gosdk/tracing"
	"github.com/anthanhphan/gosdk/utils"
)

// Chain combines multiple unary interceptors into a single interceptor.
// The interceptors are executed in the order they are provided.
// Short-circuits for 0 or 1 interceptor at build time to avoid per-request overhead.
// The closure chain is pre-built at construction time — subsequent calls
// invoke the pre-built chain with zero per-request allocations.
func Chain(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	switch len(interceptors) {
	case 0:
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	case 1:
		return interceptors[0]
	default:
		return chainUnaryInterceptors(interceptors)
	}
}

// chainUnaryInterceptors builds a composite interceptor from a snapshot of the input slice.
// Uses indexed recursive dispatch to avoid N closure allocations per request.
func chainUnaryInterceptors(interceptors []grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	// Take a snapshot to prevent mutation after construction
	chain := make([]grpc.UnaryServerInterceptor, len(interceptors))
	copy(chain, interceptors)
	n := len(chain)

	// Recursive dispatch: interceptors[i] calls interceptors[i+1] via handler
	var dispatch func(i int, ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error)
	dispatch = func(i int, ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if i == n {
			return handler(ctx, req)
		}
		return chain[i](ctx, req, info, func(ctx context.Context, req any) (any, error) {
			return dispatch(i+1, ctx, req, info, handler)
		})
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return dispatch(0, ctx, req, info, handler)
	}
}

// StreamChain combines multiple stream interceptors into a single interceptor.
// Short-circuits for 0 or 1 interceptor at build time to avoid per-request overhead.
// The closure chain is pre-built at construction time.
func StreamChain(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	switch len(interceptors) {
	case 0:
		return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, ss)
		}
	case 1:
		return interceptors[0]
	default:
		return chainStreamInterceptors(interceptors)
	}
}

// chainStreamInterceptors builds a composite stream interceptor from a snapshot of the input slice.
// Uses indexed recursive dispatch to avoid N closure allocations per request.
func chainStreamInterceptors(interceptors []grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	chain := make([]grpc.StreamServerInterceptor, len(interceptors))
	copy(chain, interceptors)
	n := len(chain)

	var dispatch func(i int, srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
	dispatch = func(i int, srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if i == n {
			return handler(srv, ss)
		}
		return chain[i](srv, ss, info, func(srv any, ss grpc.ServerStream) error {
			return dispatch(i+1, srv, ss, info, handler)
		})
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return dispatch(0, srv, ss, info, handler)
	}
}

// Recover wraps a unary interceptor with panic recovery.
// Captures the goroutine stack trace and logs it server-side for debugging.
// The error returned to clients is sanitized (no stack trace).
// Includes request_id and trace_id for incident correlation.
func Recover() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, retErr error) {
		defer func() {
			if r := recover(); r != nil {
				location, _ := utils.GetPanicLocation()
				// Extract request_id and trace_id for incident correlation
				reqID := extractRequestIDFromMD(ctx)
				traceID := tracing.TraceIDFromContext(ctx)
				logger.Errorw("panic recovered",
					"error", fmt.Sprint(r),
					"location", location,
					"method", info.FullMethod,
					"request_id", reqID,
					"trace_id", traceID,
				)
				retErr = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamRecover wraps a stream interceptor with panic recovery.
// Captures the goroutine stack trace and logs it server-side for debugging.
// The error returned to clients is sanitized (no stack trace).
// Includes request_id and trace_id for incident correlation.
func StreamRecover() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
		defer func() {
			if r := recover(); r != nil {
				location, _ := utils.GetPanicLocation()
				reqID := ""
				traceID := ""
				if ss != nil {
					ctx := ss.Context()
					reqID = extractRequestIDFromMD(ctx)
					traceID = tracing.TraceIDFromContext(ctx)
				}
				logger.Errorw("panic recovered",
					"error", fmt.Sprint(r),
					"location", location,
					"method", info.FullMethod,
					"request_id", reqID,
					"trace_id", traceID,
				)
				retErr = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(srv, ss)
	}
}

// Timeout wraps a unary handler with a timeout.
func Timeout(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := handler(timeoutCtx, req)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, status.Error(codes.DeadlineExceeded, "request timeout")
		}
		return resp, err
	}
}

// Optional applies an interceptor only if the condition function returns true.
func Optional(condition func(ctx context.Context, info *grpc.UnaryServerInfo) bool, interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if condition(ctx, info) {
			return interceptor(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}

// SkipForMethods skips the interceptor for specific full method names.
func SkipForMethods(interceptor grpc.UnaryServerInterceptor, methods ...string) grpc.UnaryServerInterceptor {
	methodSet := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		methodSet[m] = struct{}{}
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := methodSet[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		return interceptor(ctx, req, info, handler)
	}
}

// RequestIDInterceptor creates a unary interceptor that ensures every request
// has a unique request ID. If the client sends x-request-id metadata, it is
// preserved. Otherwise, a UUID v7 is generated.
// The request ID is injected back into incoming metadata and set as an
// outgoing header for client visibility.
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = ensureRequestID(ctx)
		return handler(ctx, req)
	}
}

// StreamRequestIDInterceptor creates a stream interceptor that ensures every
// stream has a unique request ID, mirroring RequestIDInterceptor for streams.
func StreamRequestIDInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ensureRequestID(ss.Context())
		wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrappedStream)
	}
}

// ensureRequestID checks for an existing x-request-id in incoming metadata.
// If absent or invalid, generates a UUID v7 and injects it into the context's metadata.
// Uses shared/requestid.IsValid to prevent log injection attacks.
func ensureRequestID(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	reqID := ""
	if vals := md.Get("x-request-id"); len(vals) > 0 {
		reqID = vals[0]
	}
	if !requestid.IsValid(reqID) {
		reqID = requestid.Generate()
	}

	// Inject into incoming metadata so downstream reads see it
	md = md.Copy()
	md.Set("x-request-id", reqID)
	ctx = metadata.NewIncomingContext(ctx, md)

	// Set outgoing header so the client receives the request ID
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", reqID))

	return ctx
}

// extractRequestIDFromMD extracts the request ID from incoming gRPC metadata.
// Returns empty string if not found.
func extractRequestIDFromMD(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if vals := md.Get("x-request-id"); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
