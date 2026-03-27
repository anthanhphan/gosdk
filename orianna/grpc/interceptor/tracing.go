// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/tracing"
)

// TracingInterceptor creates a unary interceptor that starts a span for each gRPC call.
// It extracts trace context from incoming metadata, creates a server span with
// standard RPC attributes, and records the gRPC status code.
//
// Input:
//   - client: The tracing client to use for creating spans
//
// Example:
//
//	srv := server.NewServer(config,
//	    orianna.WithGrpcGlobalUnaryInterceptor(
//	        interceptor.TracingInterceptor(tracingClient),
//	    ),
//	)
func TracingInterceptor(client tracing.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Extract trace context from incoming gRPC metadata
		ctx = extractTraceContext(ctx)

		// Start server span
		ctx, span := client.StartSpan(ctx, info.FullMethod,
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
				attribute.String("rpc.grpc.client_identity", ClientIdentityFromContext(ctx)),
			),
		)
		defer span.End()

		// Execute handler
		resp, err := handler(ctx, req)

		// Record status
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(attribute.String("rpc.grpc.status_code", grpcCodeString(int(st.Code()))))
			span.RecordError(err)
			span.SetStatus(codes.Error, st.Message())
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", "0"))
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// StreamTracingInterceptor creates a stream interceptor that starts a span
// for each gRPC stream. It records stream lifecycle events and status.
//
// Input:
//   - client: The tracing client to use for creating spans
//
// Example:
//
//	srv := server.NewServer(config,
//	    orianna.WithGrpcGlobalStreamInterceptor(
//	        interceptor.StreamTracingInterceptor(tracingClient),
//	    ),
//	)
func StreamTracingInterceptor(client tracing.Client) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Extract trace context from incoming gRPC metadata
		ctx = extractTraceContext(ctx)

		// Start server span
		ctx, span := client.StartSpan(ctx, info.FullMethod,
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
				attribute.Bool("rpc.grpc.is_server_stream", info.IsServerStream),
				attribute.Bool("rpc.grpc.is_client_stream", info.IsClientStream),
				attribute.String("rpc.grpc.client_identity", ClientIdentityFromContext(ctx)),
			),
		)
		defer span.End()

		// Wrap server stream with traced context
		wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		err := handler(srv, wrappedStream)

		// Record status
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(attribute.String("rpc.grpc.status_code", grpcCodeString(int(st.Code()))))
			span.RecordError(err)
			span.SetStatus(codes.Error, st.Message())
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", "0"))
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// extractTraceContext extracts trace context from incoming gRPC metadata.
func extractTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	return tracing.ExtractContext(ctx, &core.GRPCMetadataCarrier{MD: md})
}
