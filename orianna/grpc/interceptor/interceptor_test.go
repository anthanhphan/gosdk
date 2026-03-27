// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestChain_Empty(t *testing.T) {
	chain := Chain()
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := chain(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestChain_Order(t *testing.T) {
	var order []int

	interceptor1 := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		order = append(order, 1)
		return handler(ctx, req)
	}
	interceptor2 := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		order = append(order, 2)
		return handler(ctx, req)
	}

	chain := Chain(interceptor1, interceptor2)
	handler := func(ctx context.Context, req any) (any, error) {
		order = append(order, 3)
		return "ok", nil
	}

	_, err := chain(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("expected order [1,2,3], got %v", order)
	}
}

func TestChain_ShortCircuit(t *testing.T) {
	expectedErr := errors.New("interceptor error")

	interceptor1 := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return nil, expectedErr
	}

	chain := Chain(interceptor1)
	handler := func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	_, err := chain(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected interceptor error, got %v", err)
	}
}

func TestStreamChain_Empty(t *testing.T) {
	chain := StreamChain()
	handler := func(srv any, ss grpc.ServerStream) error {
		return nil
	}

	err := chain(nil, nil, &grpc.StreamServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStreamChain_Order(t *testing.T) {
	var order []int

	interceptor1 := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		order = append(order, 1)
		return handler(srv, ss)
	}
	interceptor2 := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		order = append(order, 2)
		return handler(srv, ss)
	}

	chain := StreamChain(interceptor1, interceptor2)
	handler := func(srv any, ss grpc.ServerStream) error {
		order = append(order, 3)
		return nil
	}

	err := chain(nil, nil, &grpc.StreamServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("expected order [1,2,3], got %v", order)
	}
}

func TestRecover_PanicRecovery(t *testing.T) {
	interceptor := Recover()

	handler := func(ctx context.Context, req any) (any, error) {
		panic("test panic")
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestRecover_NoPanic(t *testing.T) {
	interceptor := Recover()

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestStreamRecover_PanicRecovery(t *testing.T) {
	interceptor := StreamRecover()

	handler := func(srv any, ss grpc.ServerStream) error {
		panic("stream panic")
	}

	err := interceptor(nil, nil, &grpc.StreamServerInfo{}, handler)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestTimeout_WithinDeadline(t *testing.T) {
	interceptor := Timeout(5 * time.Second)

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestTimeout_Exceeded(t *testing.T) {
	interceptor := Timeout(1 * time.Millisecond)

	handler := func(ctx context.Context, req any) (any, error) {
		time.Sleep(50 * time.Millisecond)
		return "ok", nil
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", st.Code())
	}
}

func TestOptional_ConditionTrue(t *testing.T) {
	interceptorCalled := false
	inner := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		interceptorCalled = true
		return handler(ctx, req)
	}

	condition := func(ctx context.Context, info *grpc.UnaryServerInfo) bool {
		return true
	}

	opt := Optional(condition, inner)
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	_, _ = opt(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if !interceptorCalled {
		t.Fatal("interceptor should have been called")
	}
}

func TestOptional_ConditionFalse(t *testing.T) {
	interceptorCalled := false
	inner := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		interceptorCalled = true
		return handler(ctx, req)
	}

	condition := func(ctx context.Context, info *grpc.UnaryServerInfo) bool {
		return false
	}

	opt := Optional(condition, inner)
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	_, _ = opt(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if interceptorCalled {
		t.Fatal("interceptor should NOT have been called")
	}
}

func TestSkipForMethods_Skipped(t *testing.T) {
	interceptorCalled := false
	inner := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		interceptorCalled = true
		return handler(ctx, req)
	}

	skip := SkipForMethods(inner, "/pkg.Svc/Health")
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Health"}
	_, _ = skip(context.Background(), nil, info, handler)
	if interceptorCalled {
		t.Fatal("interceptor should have been skipped")
	}
}

func TestSkipForMethods_NotSkipped(t *testing.T) {
	interceptorCalled := false
	inner := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		interceptorCalled = true
		return handler(ctx, req)
	}

	skip := SkipForMethods(inner, "/pkg.Svc/Health")
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/GetUser"}
	_, _ = skip(context.Background(), nil, info, handler)
	if !interceptorCalled {
		t.Fatal("interceptor should have been called")
	}
}

func TestRequestIDInterceptor_GeneratesID(t *testing.T) {
	interceptor := RequestIDInterceptor()
	handler := func(ctx context.Context, req any) (any, error) {
		// Verify request ID is present in incoming metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			t.Fatal("expected incoming metadata")
		}
		vals := md.Get("x-request-id")
		if len(vals) == 0 || vals[0] == "" {
			t.Fatal("expected non-empty x-request-id")
		}
		return "ok", nil
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestIDInterceptor_PreservesExisting(t *testing.T) {
	interceptor := RequestIDInterceptor()

	existingID := "01234567-89ab-cdef-0123-456789abcdef"
	md := metadata.Pairs("x-request-id", existingID)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req any) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		vals := md.Get("x-request-id")
		if len(vals) == 0 || vals[0] != existingID {
			t.Fatalf("expected preserved ID %q, got %v", existingID, vals)
		}
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStreamRequestIDInterceptor(t *testing.T) {
	interceptor := StreamRequestIDInterceptor()

	handlerCalled := false
	handler := func(srv any, ss grpc.ServerStream) error {
		handlerCalled = true
		// Verify the stream has a modified context with request ID
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			t.Fatal("expected incoming metadata in stream context")
		}
		vals := md.Get("x-request-id")
		if len(vals) == 0 || vals[0] == "" {
			t.Fatal("expected non-empty x-request-id in stream context")
		}
		return nil
	}

	mockStream := &testServerStream{ctx: context.Background()}
	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler was not called")
	}
}

func TestWrappedServerStream_Context(t *testing.T) {
	type testCtxKey struct{}
	innerCtx := context.WithValue(context.Background(), testCtxKey{}, "test-value")
	ws := &wrappedServerStream{
		ServerStream: &testServerStream{ctx: context.Background()},
		ctx:          innerCtx,
	}
	if ws.Context() != innerCtx {
		t.Fatal("wrapped stream should return the overridden context")
	}
}

func TestStreamRecover_WithMockStream(t *testing.T) {
	interceptor := StreamRecover()

	md := metadata.Pairs("x-request-id", "test-req-id")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	mockStream := &testServerStream{ctx: ctx}

	handler := func(srv any, ss grpc.ServerStream) error {
		panic("stream panic with context")
	}

	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{FullMethod: "/svc/m"}, handler)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Fatalf("expected Internal status, got %v", err)
	}
}

func TestExtractRequestIDFromMD(t *testing.T) {
	// With request ID
	md := metadata.Pairs("x-request-id", "req-123")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	id := extractRequestIDFromMD(ctx)
	if id != "req-123" {
		t.Fatalf("expected req-123, got %q", id)
	}

	// Without metadata
	id = extractRequestIDFromMD(context.Background())
	if id != "" {
		t.Fatalf("expected empty, got %q", id)
	}

	// With metadata but no request ID
	md2 := metadata.Pairs("other", "val")
	ctx2 := metadata.NewIncomingContext(context.Background(), md2)
	id = extractRequestIDFromMD(ctx2)
	if id != "" {
		t.Fatalf("expected empty, got %q", id)
	}
}

// testServerStream is a mock ServerStream for interceptor tests
type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testServerStream) Context() context.Context {
	return s.ctx
}
