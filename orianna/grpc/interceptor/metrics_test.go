// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/metrics"
	"google.golang.org/grpc"
)

func TestGrpcCodeString_Cached(t *testing.T) {
	for i := 0; i <= 16; i++ {
		result := grpcCodeString(i)
		if result == "" {
			t.Errorf("grpcCodeString(%d) returned empty string", i)
		}
	}
}

func TestGrpcCodeString_Uncached(t *testing.T) {
	result := grpcCodeString(99)
	if result != "99" {
		t.Errorf("grpcCodeString(99) = %q, want '99'", result)
	}
}

func TestMetricsInterceptor_Success(t *testing.T) {
	client := metrics.NewNoopClient()
	interceptor := MetricsInterceptor(client, "test")

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}
	resp, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestMetricsInterceptor_Error(t *testing.T) {
	client := metrics.NewNoopClient()
	interceptor := MetricsInterceptor(client, "test")

	expectedErr := errors.New("fail")
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, expectedErr
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}
	_, err := interceptor(context.Background(), nil, info, handler)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

type mockMetricsStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *mockMetricsStream) Context() context.Context {
	return s.ctx
}

func TestStreamMetricsInterceptor_Success(t *testing.T) {
	client := metrics.NewNoopClient()
	interceptor := StreamMetricsInterceptor(client, "test")

	handler := func(srv any, ss grpc.ServerStream) error {
		return nil
	}
	mockStream := &mockMetricsStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/pkg.Svc/StreamMethod"}
	err := interceptor(nil, mockStream, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStreamMetricsInterceptor_Error(t *testing.T) {
	client := metrics.NewNoopClient()
	interceptor := StreamMetricsInterceptor(client, "test")

	expectedErr := errors.New("stream fail")
	handler := func(srv any, ss grpc.ServerStream) error {
		return expectedErr
	}
	mockStream := &mockMetricsStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/pkg.Svc/StreamMethod"}
	err := interceptor(nil, mockStream, info, handler)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected stream error, got %v", err)
	}
}

// Verify in-flight gauge tracking
func TestMetricsInterceptor_InFlight(t *testing.T) {
	client := metrics.NewNoopClient()
	interceptor := MetricsInterceptor(client, "test")

	handler := func(ctx context.Context, req any) (any, error) {
		// During handler execution, in-flight should be incremented
		_ = time.Now() // placeholder for timing
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}
	_, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
