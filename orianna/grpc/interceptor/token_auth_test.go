// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockTokenValidator implements TokenValidator for testing.
type mockTokenValidator struct {
	claims map[string]any
	err    error
}

func (m *mockTokenValidator) Validate(_ context.Context, _ string) (map[string]any, error) {
	return m.claims, m.err
}

func TestTokenAuthInterceptor_Success(t *testing.T) {
	validator := &mockTokenValidator{
		claims: map[string]any{"sub": "user-123", "role": "admin"},
	}
	interceptor := TokenAuthInterceptor(validator)

	md := metadata.New(map[string]string{
		"authorization": "Bearer valid-token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req any) (any, error) {
		claims := TokenClaimsFromContext(ctx)
		if claims == nil {
			t.Error("expected claims in context")
		}
		if claims["sub"] != "user-123" {
			t.Errorf("expected sub=user-123, got %v", claims["sub"])
		}
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	resp, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected 'ok', got %v", resp)
	}
}

func TestTokenAuthInterceptor_MissingMetadata(t *testing.T) {
	validator := &mockTokenValidator{}
	interceptor := TokenAuthInterceptor(validator)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	_, err := interceptor(context.Background(), "request", info,
		func(ctx context.Context, req any) (any, error) { return nil, nil })
	if err == nil {
		t.Fatal("expected error")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestTokenAuthInterceptor_MissingAuthHeader(t *testing.T) {
	validator := &mockTokenValidator{}
	interceptor := TokenAuthInterceptor(validator)

	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	_, err := interceptor(ctx, "request", info,
		func(ctx context.Context, req any) (any, error) { return nil, nil })
	if err == nil {
		t.Fatal("expected error")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestTokenAuthInterceptor_InvalidFormat(t *testing.T) {
	validator := &mockTokenValidator{}
	interceptor := TokenAuthInterceptor(validator)

	md := metadata.New(map[string]string{
		"authorization": "Basic dXNlcjpwYXNz",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	_, err := interceptor(ctx, "request", info,
		func(ctx context.Context, req any) (any, error) { return nil, nil })
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTokenAuthInterceptor_ValidationFailure(t *testing.T) {
	validator := &mockTokenValidator{
		err: status.Error(codes.Unauthenticated, "token expired"),
	}
	interceptor := TokenAuthInterceptor(validator)

	md := metadata.New(map[string]string{
		"authorization": "Bearer expired-token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	_, err := interceptor(ctx, "request", info,
		func(ctx context.Context, req any) (any, error) { return nil, nil })
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTokenClaimsFromContext_NoClaims(t *testing.T) {
	claims := TokenClaimsFromContext(context.Background())
	if claims != nil {
		t.Error("expected nil claims")
	}
}

func TestStreamTokenAuthInterceptor_Success(t *testing.T) {
	validator := &mockTokenValidator{
		claims: map[string]any{"sub": "user-456"},
	}
	interceptor := StreamTokenAuthInterceptor(validator)

	md := metadata.New(map[string]string{
		"authorization": "Bearer valid-token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(srv any, stream grpc.ServerStream) error {
		claims := TokenClaimsFromContext(stream.Context())
		if claims == nil {
			t.Error("expected claims in stream context")
		}
		return nil
	}

	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/StreamMethod"}
	err := interceptor(nil, &mockServerStream{ctx: ctx}, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExtractBearerToken_EmptyToken(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer ",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := extractBearerToken(ctx)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
