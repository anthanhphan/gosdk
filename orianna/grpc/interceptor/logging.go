// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/tracing"
)

// grpcLogFieldsPool pools []any slices for structured log fields.
// Capacity 16 covers all audit + metadata + payload fields without reallocation.
var grpcLogFieldsPool = sync.Pool{
	New: func() any {
		s := make([]any, 0, 16)
		return &s
	},
}

func acquireGRPCLogFields() []any {
	p := grpcLogFieldsPool.Get().(*[]any)
	return (*p)[:0]
}

func releaseGRPCLogFields(fields []any) {
	for i := range fields {
		fields[i] = nil
	}
	fields = fields[:0]
	grpcLogFieldsPool.Put(&fields)
}

// interceptorLog is a package-level logger for gRPC interceptors.
var interceptorLog = logger.NewLoggerWithFields(logger.String("package", "grpc-interceptor"))

// SensitiveMetadataKeys is the set of metadata/header keys that must never be logged.
// These are filtered from verbose logging to prevent PII/PCI/credential leakage.
// Aligned with HTTP SensitiveHTTPHeaders for cross-protocol consistency.
var SensitiveMetadataKeys = map[string]struct{}{
	"authorization":       {},
	"cookie":              {},
	"set-cookie":          {},
	"x-api-key":           {},
	"x-api-secret":        {},
	"proxy-authorization": {},
	"x-csrf-token":        {},
	"x-xsrf-token":        {},
	"x-refresh-token":     {},
	"access-token":        {},
	"refresh-token":       {},
	"secret-key":          {},
	"private-key":         {},
}

// skipMatcher checks whether a gRPC method should be skipped.
// Supports exact match AND wildcard patterns (/*,  /pkg.Service/*).
// Constructed once and shared between unary and stream interceptors.
type skipMatcher struct {
	exact    map[string]struct{} // exact method matches (fast path)
	patterns []string            // wildcard patterns (slow path, checked only when exact misses)
}

// newSkipMatcher builds a skip matcher from a list of method patterns.
// Exact methods go into a map for O(1) lookup. Patterns with wildcards are stored
// separately and checked via matchMethodPattern.
func newSkipMatcher(methods []string) *skipMatcher {
	sm := &skipMatcher{
		exact: make(map[string]struct{}, len(methods)),
	}
	for _, m := range methods {
		if m == "/*" || len(m) > 1 && m[len(m)-1] == '*' {
			sm.patterns = append(sm.patterns, m)
		} else {
			sm.exact[m] = struct{}{}
		}
	}
	return sm
}

// shouldSkip returns true if the given method matches any skip rule.
func (sm *skipMatcher) shouldSkip(fullMethod string) bool {
	if _, ok := sm.exact[fullMethod]; ok {
		return true
	}
	for _, p := range sm.patterns {
		if matchMethodPattern(p, fullMethod) {
			return true
		}
	}
	return false
}

// VerboseLoggingInterceptor creates a unary interceptor that logs both request and response
// with full audit trail: method, peer, client identity (from mTLS), trace_id.
// Methods matching skipMethods patterns will not be logged.
// Accepts optional *logger.Logger; defaults to package-level interceptorLog.
func VerboseLoggingInterceptor(skipMethods []string, log ...*logger.Logger) grpc.UnaryServerInterceptor {
	skip := newSkipMatcher(skipMethods)
	l := interceptorLog
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if skip.shouldSkip(info.FullMethod) {
			return handler(ctx, req)
		}

		peerAddr := extractPeerAddr(ctx)
		clientID := ClientIdentityFromContext(ctx)
		traceID := tracing.TraceIDFromContext(ctx)
		reqID := extractRequestIDFromMD(ctx)

		// Log incoming request with audit fields
		reqFields := buildGRPCRequestLogFields(acquireGRPCLogFields(), info.FullMethod, peerAddr, clientID, traceID, ctx, req)
		l.Infow("incoming request", reqFields...)
		releaseGRPCLogFields(reqFields)

		// Execute handler
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log response with audit fields
		respFields := buildGRPCResponseLogFields(acquireGRPCLogFields(), info.FullMethod, clientID, traceID, reqID, duration, resp, err)
		if err != nil {
			l.Warnw("request completed", respFields...)
		} else {
			l.Infow("request completed", respFields...)
		}
		releaseGRPCLogFields(respFields)

		return resp, err
	}
}

// StreamVerboseLoggingInterceptor creates a stream interceptor that logs stream lifecycle.
// Methods matching skipMethods patterns will not be logged.
// Accepts optional *logger.Logger; defaults to package-level interceptorLog.
func StreamVerboseLoggingInterceptor(skipMethods []string, log ...*logger.Logger) grpc.StreamServerInterceptor {
	skip := newSkipMatcher(skipMethods)
	l := interceptorLog
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if skip.shouldSkip(info.FullMethod) {
			return handler(srv, ss)
		}

		peerAddr := extractPeerAddr(ss.Context())
		clientID := ClientIdentityFromContext(ss.Context())
		traceID := tracing.TraceIDFromContext(ss.Context())
		reqID := extractRequestIDFromMD(ss.Context())

		// Log incoming stream with audit fields
		l.Infow("incoming stream",
			"method", info.FullMethod,
			"peer", peerAddr,
			"client_identity", clientID,
			"trace_id", traceID,
			"request_id", reqID,
			"is_server_stream", info.IsServerStream,
			"is_client_stream", info.IsClientStream,
		)

		// Execute handler
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		// Log stream completed
		st, _ := status.FromError(err)
		fields := acquireGRPCLogFields()
		fields = append(fields,
			"method", info.FullMethod,
			"client_identity", clientID,
			"trace_id", traceID,
			"request_id", reqID,
			"code", st.Code().String(),
			"code_num", int(st.Code()),
			"duration_ms", duration.Milliseconds(),
		)
		if err != nil {
			fields = append(fields, "error", st.Message())
			l.Warnw("stream completed", fields...)
		} else {
			l.Infow("stream completed", fields...)
		}
		releaseGRPCLogFields(fields)

		return err
	}
}

// extractPeerAddr extracts the peer address from the gRPC context.
func extractPeerAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return p.Addr.String()
	}
	return ""
}

// buildGRPCRequestLogFields builds log fields for an incoming gRPC request.
// Includes client_identity and trace_id for audit trail.
// Sensitive metadata keys are redacted using SensitiveMetadataKeys.
func buildGRPCRequestLogFields(fields []any, fullMethod, peerAddr, clientID, traceID string, ctx context.Context, req any) []any {
	fields = append(fields,
		"method", fullMethod,
		"peer", peerAddr,
		"client_identity", clientID,
		"trace_id", traceID,
	)

	// Include metadata with sensitive values redacted
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-request-id"); len(vals) > 0 {
			fields = append(fields, "request_id", vals[0])
		}
		redacted := make(map[string]string, len(md))
		for key, vals := range md {
			if _, sensitive := SensitiveMetadataKeys[key]; sensitive {
				redacted[key] = "[REDACTED]"
			} else if len(vals) > 0 {
				redacted[key] = vals[0]
			}
		}
		if len(redacted) > 0 {
			fields = append(fields, "metadata", redacted)
		}
	}

	// Log request body
	if req != nil {
		fields = append(fields, "request", marshalProtoMessage(req))
	}

	return fields
}

// buildGRPCResponseLogFields builds log fields for a completed gRPC response.
// Includes client_identity, trace_id, and request_id for audit trail and code_num for automated analysis.
func buildGRPCResponseLogFields(fields []any, fullMethod, clientID, traceID, requestID string, duration time.Duration, resp any, err error) []any {
	st, _ := status.FromError(err)

	fields = append(fields,
		"method", fullMethod,
		"client_identity", clientID,
		"trace_id", traceID,
		"request_id", requestID,
		"code", st.Code().String(),
		"code_num", int(st.Code()),
		"duration_ms", duration.Milliseconds(),
	)

	if err != nil {
		fields = append(fields, "error", st.Message())
	}

	// Log response body
	if resp != nil {
		fields = append(fields, "response", marshalProtoMessage(resp))
	}

	return fields
}

// MaxLogPayloadBytes is the maximum size of request/response bodies logged in verbose mode.
// Prevents memory bloat, log storage costs, and accidental PII leakage from large payloads.
const MaxLogPayloadBytes = 4096

// marshalProtoMessage marshals any message (including proto) to JSON string using jcodec.
// Falls back to fmt.Sprintf if jcodec serialization fails.
// Truncates output to MaxLogPayloadBytes to prevent log bloat.
func marshalProtoMessage(msg any) string {
	if s, err := jcodec.CompactString(msg); err == nil {
		if len(s) > MaxLogPayloadBytes {
			return s[:MaxLogPayloadBytes] + "...(truncated)"
		}
		return s
	}
	s := fmt.Sprintf("%+v", msg)
	if len(s) > MaxLogPayloadBytes {
		return s[:MaxLogPayloadBytes] + "...(truncated)"
	}
	return s
}
