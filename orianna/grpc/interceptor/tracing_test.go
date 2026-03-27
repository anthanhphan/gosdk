// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/tracing"
)

// ============================================================================
// TracingInterceptor Tests
// ============================================================================

func TestTracingInterceptor_Success(t *testing.T) {
	client := tracing.NewNoopClient()
	interceptor := TracingInterceptor(client)

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	resp, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "response" {
		t.Errorf("expected 'response', got %v", resp)
	}
}

func TestTracingInterceptor_Error(t *testing.T) {
	client := tracing.NewNoopClient()
	interceptor := TracingInterceptor(client)

	expectedErr := status.Error(codes.NotFound, "not found")
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, expectedErr
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	_, err := interceptor(context.Background(), "request", info, handler)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestTracingInterceptor_WithMetadata(t *testing.T) {
	client := tracing.NewNoopClient()
	interceptor := TracingInterceptor(client)

	// Create context with incoming metadata containing traceparent
	md := metadata.New(map[string]string{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	resp, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected 'ok', got %v", resp)
	}
}

// ============================================================================
// StreamTracingInterceptor Tests
// ============================================================================

func TestStreamTracingInterceptor_Success(t *testing.T) {
	client := tracing.NewNoopClient()
	interceptor := StreamTracingInterceptor(client)

	handler := func(srv any, stream grpc.ServerStream) error {
		return nil
	}

	info := &grpc.StreamServerInfo{
		FullMethod:     "/test.Service/StreamMethod",
		IsServerStream: true,
	}

	err := interceptor(nil, &mockServerStream{ctx: context.Background()}, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStreamTracingInterceptor_Error(t *testing.T) {
	client := tracing.NewNoopClient()
	interceptor := StreamTracingInterceptor(client)

	expectedErr := errors.New("stream error")
	handler := func(srv any, stream grpc.ServerStream) error {
		return expectedErr
	}

	info := &grpc.StreamServerInfo{
		FullMethod:     "/test.Service/StreamMethod",
		IsServerStream: true,
	}

	err := interceptor(nil, &mockServerStream{ctx: context.Background()}, info, handler)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================================
// grpcMetadataCarrier Tests
// ============================================================================

func TestGrpcMetadataCarrier(t *testing.T) {
	md := metadata.New(map[string]string{
		"traceparent": "00-trace-span-01",
		"key":         "value",
	})
	carrier := &core.GRPCMetadataCarrier{MD: md}

	t.Run("Get existing key", func(t *testing.T) {
		val := carrier.Get("traceparent")
		if val != "00-trace-span-01" {
			t.Errorf("expected '00-trace-span-01', got %q", val)
		}
	})

	t.Run("Get missing key", func(t *testing.T) {
		val := carrier.Get("nonexistent")
		if val != "" {
			t.Errorf("expected empty string, got %q", val)
		}
	})

	t.Run("Set key", func(t *testing.T) {
		carrier.Set("newkey", "newvalue")
		val := carrier.Get("newkey")
		if val != "newvalue" {
			t.Errorf("expected 'newvalue', got %q", val)
		}
	})

	t.Run("Keys returns all keys", func(t *testing.T) {
		keys := carrier.Keys()
		if len(keys) < 2 {
			t.Errorf("expected at least 2 keys, got %d", len(keys))
		}
	})
}

// ============================================================================
// extractTraceContext Tests
// ============================================================================

func TestExtractTraceContext_NoMetadata(t *testing.T) {
	ctx := context.Background()
	result := extractTraceContext(ctx)
	if result == nil {
		t.Fatal("expected non-nil context")
	}
}

func TestExtractTraceContext_WithMetadata(t *testing.T) {
	md := metadata.New(map[string]string{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	result := extractTraceContext(ctx)
	if result == nil {
		t.Fatal("expected non-nil context")
	}
}
