// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Test marshalProtoMessage function with jcodec serialization
func TestMarshalProtoMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      any
		expected string
	}{
		{
			name:     "proto wrapper - StringValue",
			msg:      wrapperspb.String("test value"),
			expected: `{"value":"test value"}`,
		},
		{
			name:     "proto wrapper - Int64Value",
			msg:      wrapperspb.Int64(12345),
			expected: `{"value":12345}`,
		},
		{
			name:     "proto wrapper - BoolValue true",
			msg:      wrapperspb.Bool(true),
			expected: `{"value":true}`,
		},
		{
			name:     "proto wrapper - BoolValue false",
			msg:      wrapperspb.Bool(false),
			expected: `{}`, // jcodec omits zero-value fields
		},
		{
			name:     "non-proto message - string",
			msg:      "plain string",
			expected: "plain string", // jcodec CompactString returns raw string
		},
		{
			name:     "non-proto message - struct",
			msg:      struct{ Name string }{Name: "test"},
			expected: `{"Name":"test"}`,
		},
		{
			name:     "nil message",
			msg:      nil,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := marshalProtoMessage(tt.msg)
			if result != tt.expected {
				t.Errorf("marshalProtoMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test that marshalProtoMessage correctly handles proto.Message via jcodec
func TestMarshalProtoMessage_ProtoInterface(t *testing.T) {
	msg := wrapperspb.String("test")
	result := marshalProtoMessage(msg)

	// jcodec marshals proto wrapper types as full JSON objects
	expected := `{"value":"test"}`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Test marshalProtoMessage with empty proto message
func TestMarshalProtoMessage_EmptyProto(t *testing.T) {
	msg := &wrapperspb.StringValue{} // Empty StringValue
	result := marshalProtoMessage(msg)

	// jcodec marshals empty structs as empty JSON objects
	expected := `{}`
	if result != expected {
		t.Errorf("Expected %q for empty proto, got %q", expected, result)
	}
}

func TestMarshalProtoMessage_LargePayload(t *testing.T) {
	// Create a message larger than MaxLogPayloadBytes
	large := make([]byte, MaxLogPayloadBytes+100)
	for i := range large {
		large[i] = 'a'
	}
	result := marshalProtoMessage(string(large))
	if len(result) <= MaxLogPayloadBytes {
		t.Errorf("expected truncated result to be larger than %d, got %d", MaxLogPayloadBytes, len(result))
	}
	if !strings.Contains(result, "...(truncated)") {
		t.Error("expected truncated suffix")
	}
}

func TestNewSkipMatcher(t *testing.T) {
	sm := newSkipMatcher([]string{"/pkg.Svc/Exact", "/*", "/pkg.Svc/*"})
	if _, ok := sm.exact["/pkg.Svc/Exact"]; !ok {
		t.Error("expected exact match entry")
	}
	if len(sm.patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(sm.patterns))
	}
}

func TestShouldSkip(t *testing.T) {
	sm := newSkipMatcher([]string{"/grpc.health.v1.Health/Check", "/pkg.Svc/*"})

	// Exact match
	if !sm.shouldSkip("/grpc.health.v1.Health/Check") {
		t.Error("expected exact skip match")
	}
	// Wildcard match
	if !sm.shouldSkip("/pkg.Svc/Method") {
		t.Error("expected wildcard skip match")
	}
	// No match
	if sm.shouldSkip("/other.Svc/Method") {
		t.Error("should not skip unmatched method")
	}
}

func TestVerboseLoggingInterceptor_Success(t *testing.T) {
	interceptor := VerboseLoggingInterceptor(nil)
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}
	resp, err := interceptor(context.Background(), "request-body", info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestVerboseLoggingInterceptor_Error(t *testing.T) {
	interceptor := VerboseLoggingInterceptor(nil)
	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, expectedErr
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}
	_, err := interceptor(context.Background(), nil, info, handler)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestVerboseLoggingInterceptor_Skipped(t *testing.T) {
	interceptor := VerboseLoggingInterceptor([]string{"/grpc.health.v1.Health/Check"})
	handlerCalled := false
	handler := func(ctx context.Context, req any) (any, error) {
		handlerCalled = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}
	_, _ = interceptor(context.Background(), nil, info, handler)
	if !handlerCalled {
		t.Fatal("handler should have been called")
	}
}

func TestVerboseLoggingInterceptor_WithCustomLogger(t *testing.T) {
	l := logger.NewLoggerWithFields()
	interceptor := VerboseLoggingInterceptor(nil, l)
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}
	_, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStreamVerboseLoggingInterceptor_Success(t *testing.T) {
	interceptor := StreamVerboseLoggingInterceptor(nil)
	handler := func(srv any, ss grpc.ServerStream) error {
		return nil
	}

	mockStream := &testLoggingStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/pkg.Svc/StreamMethod"}
	err := interceptor(nil, mockStream, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStreamVerboseLoggingInterceptor_Error(t *testing.T) {
	interceptor := StreamVerboseLoggingInterceptor(nil)
	expectedErr := errors.New("stream error")
	handler := func(srv any, ss grpc.ServerStream) error {
		return expectedErr
	}

	mockStream := &testLoggingStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/pkg.Svc/StreamMethod"}
	err := interceptor(nil, mockStream, info, handler)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected stream error, got %v", err)
	}
}

func TestStreamVerboseLoggingInterceptor_Skipped(t *testing.T) {
	interceptor := StreamVerboseLoggingInterceptor([]string{"/skip/*"})
	handlerCalled := false
	handler := func(srv any, ss grpc.ServerStream) error {
		handlerCalled = true
		return nil
	}

	mockStream := &testLoggingStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/skip/Method"}
	_ = interceptor(nil, mockStream, info, handler)
	if !handlerCalled {
		t.Fatal("handler should have been called")
	}
}

func TestStreamVerboseLoggingInterceptor_WithCustomLogger(t *testing.T) {
	l := logger.NewLoggerWithFields()
	interceptor := StreamVerboseLoggingInterceptor(nil, l)
	handler := func(srv any, ss grpc.ServerStream) error {
		return nil
	}
	mockStream := &testLoggingStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/pkg.Svc/StreamMethod"}
	err := interceptor(nil, mockStream, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractPeerAddr(t *testing.T) {
	// Without peer
	addr := extractPeerAddr(context.Background())
	if addr != "" {
		t.Fatalf("expected empty, got %q", addr)
	}
}

func TestBuildGRPCRequestLogFields_WithMetadata(t *testing.T) {
	md := metadata.Pairs("x-request-id", "req-id", "authorization", "Bearer token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	fields := buildGRPCRequestLogFields(nil, "/svc/m", "127.0.0.1", "client", "trace-id", ctx, "req-body")
	if len(fields) == 0 {
		t.Fatal("expected non-empty fields")
	}
}

func TestBuildGRPCResponseLogFields(t *testing.T) {
	// Success
	fields := buildGRPCResponseLogFields(nil, "/svc/m", "client", "trace", "req-id", time.Second, "resp-body", nil)
	if len(fields) == 0 {
		t.Fatal("expected non-empty fields")
	}

	// Error
	fields = buildGRPCResponseLogFields(nil, "/svc/m", "client", "trace", "req-id", time.Second, nil, errors.New("fail"))
	if len(fields) == 0 {
		t.Fatal("expected non-empty fields with error")
	}
}

type testLoggingStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testLoggingStream) Context() context.Context {
	return s.ctx
}
