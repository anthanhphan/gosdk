// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestSlowRPCDetector(t *testing.T) {
	t.Run("fast RPC does not trigger warning", func(t *testing.T) {
		detector := SlowRPCDetector(1 * time.Second)

		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/FastMethod"}
		handler := func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		}

		resp, err := detector(ctx, nil, info, handler)
		if err != nil {
			t.Fatalf("SlowRPCDetector() error = %v, want nil", err)
		}
		if resp != "ok" {
			t.Fatalf("SlowRPCDetector() resp = %v, want 'ok'", resp)
		}
	})

	t.Run("slow RPC triggers warning", func(t *testing.T) {
		detector := SlowRPCDetector(10 * time.Millisecond)

		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/SlowMethod"}
		handler := func(ctx context.Context, req any) (any, error) {
			time.Sleep(15 * time.Millisecond)
			return "slow-ok", nil
		}

		resp, err := detector(ctx, nil, info, handler)
		if err != nil {
			t.Fatalf("SlowRPCDetector() error = %v, want nil", err)
		}
		if resp != "slow-ok" {
			t.Fatalf("resp = %v, want 'slow-ok'", resp)
		}
	})

	t.Run("propagates handler error", func(t *testing.T) {
		detector := SlowRPCDetector(1 * time.Second)

		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/ErrorMethod"}
		expectedErr := context.DeadlineExceeded
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, expectedErr
		}

		_, err := detector(ctx, nil, info, handler)
		if err != expectedErr {
			t.Fatalf("error = %v, want %v", err, expectedErr)
		}
	})
}

func TestStreamSlowRPCDetector(t *testing.T) {
	t.Run("fast stream does not trigger", func(t *testing.T) {
		detector := StreamSlowRPCDetector(1 * time.Second)

		info := &grpc.StreamServerInfo{FullMethod: "/test.Service/FastStream"}
		handler := func(srv any, ss grpc.ServerStream) error {
			return nil
		}

		err := detector(nil, nil, info, handler)
		if err != nil {
			t.Fatalf("StreamSlowRPCDetector() error = %v, want nil", err)
		}
	})

	t.Run("slow stream triggers warning", func(t *testing.T) {
		detector := StreamSlowRPCDetector(10 * time.Millisecond)

		info := &grpc.StreamServerInfo{FullMethod: "/test.Service/SlowStream"}
		handler := func(srv any, ss grpc.ServerStream) error {
			time.Sleep(15 * time.Millisecond)
			return nil
		}

		err := detector(nil, nil, info, handler)
		if err != nil {
			t.Fatalf("StreamSlowRPCDetector() error = %v, want nil", err)
		}
	})
}
