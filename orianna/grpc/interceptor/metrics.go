// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/shared/observability"
)

// grpcCodeCache pre-computes string representations of all 17 gRPC status codes
// to avoid strconv.Itoa allocation on every request.
var grpcCodeCache = observability.CodeStringCache(func() []int {
	codes := make([]int, 17)
	for i := range codes {
		codes[i] = i
	}
	return codes
}())

// grpcCodeString returns a cached string for standard gRPC codes, falling back to strconv.Itoa.
func grpcCodeString(code int) string {
	return observability.CodeString(grpcCodeCache, code)
}

// MetricsInterceptor creates a unary interceptor that records gRPC metrics.
// Uses info.FullMethod to record route patterns, preventing unbounded cardinality.
// Includes client_identity label (from mTLS cert) for per-client observability.
//
// Metrics recorded:
//   - {subsystem}_requests_total: counter with labels method, code, client_identity
//   - {subsystem}_request_duration_seconds: histogram with labels method, code, client_identity
//   - {subsystem}_in_flight_requests: gauge tracking concurrent requests
func MetricsInterceptor(client metrics.Client, subsystem string) grpc.UnaryServerInterceptor {
	requestsTotalName := subsystem + observability.SuffixRequestsTotal
	requestDurationName := subsystem + observability.SuffixRequestDurationSeconds
	inFlightName := subsystem + observability.SuffixInFlightRequests

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Track in-flight requests
		client.GaugeInc(ctx, inFlightName)
		defer client.GaugeDec(ctx, inFlightName)

		start := time.Now()

		resp, err := handler(ctx, req)

		st, _ := status.FromError(err)
		codeStr := grpcCodeString(int(st.Code()))
		clientID := ClientIdentityFromContext(ctx)

		// Record request count
		client.Inc(ctx, requestsTotalName,
			"method", info.FullMethod,
			"code", codeStr,
			"client_identity", clientID,
		)

		// Record latency
		client.Duration(ctx, requestDurationName, start,
			"method", info.FullMethod,
			"code", codeStr,
			"client_identity", clientID,
		)

		return resp, err
	}
}

// StreamMetricsInterceptor creates a stream interceptor that records gRPC stream metrics.
// Includes client_identity label for per-client observability.
//
// Metrics recorded:
//   - {subsystem}_streams_total: counter with labels method, code, client_identity
//   - {subsystem}_stream_duration_seconds: histogram with labels method, code, client_identity
//   - {subsystem}_in_flight_requests: gauge tracking concurrent streams
func StreamMetricsInterceptor(client metrics.Client, subsystem string) grpc.StreamServerInterceptor {
	streamsTotalName := subsystem + observability.SuffixStreamsTotal
	streamDurationName := subsystem + observability.SuffixStreamDurationSeconds
	inFlightName := subsystem + observability.SuffixInFlightRequests

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Track in-flight streams
		client.GaugeInc(ss.Context(), inFlightName)
		defer client.GaugeDec(ss.Context(), inFlightName)

		start := time.Now()

		err := handler(srv, ss)

		st, _ := status.FromError(err)
		codeStr := grpcCodeString(int(st.Code()))
		clientID := ClientIdentityFromContext(ss.Context())

		client.Inc(ss.Context(), streamsTotalName,
			"method", info.FullMethod,
			"code", codeStr,
			"client_identity", clientID,
		)

		client.Duration(ss.Context(), streamDurationName, start,
			"method", info.FullMethod,
			"code", codeStr,
			"client_identity", clientID,
		)

		return err
	}
}
