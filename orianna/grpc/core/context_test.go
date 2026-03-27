// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNewServerContext(t *testing.T) {
	ctx := context.Background()
	sc := NewServerContext(ctx, "/pkg.Service/Method")

	assert.Equal(t, "/pkg.Service/Method", sc.FullMethod())
	assert.Equal(t, "pkg.Service", sc.ServiceName())
	assert.Equal(t, "Method", sc.MethodName())
}

func TestServerContext_EmptyMethod(t *testing.T) {
	sc := NewServerContext(context.Background(), "")
	assert.Empty(t, sc.ServiceName())
	assert.Empty(t, sc.MethodName())
}

func TestServerContext_Locals(t *testing.T) {
	sc := NewServerContext(context.Background(), "/svc/m")

	assert.Nil(t, sc.GetLocal("key"))

	sc.SetLocal("key", "value")
	assert.Equal(t, "value", sc.GetLocal("key"))
}

func TestServerContext_Metadata(t *testing.T) {
	md := metadata.Pairs("x-request-id", "abc123")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	sc := NewServerContext(ctx, "/svc/m")

	assert.Equal(t, "abc123", sc.GetMetadata("x-request-id"))
	assert.Empty(t, sc.GetMetadata("nonexistent"))
}

func TestServerContext_RequestID(t *testing.T) {
	t.Run("with request id", func(t *testing.T) {
		md := metadata.Pairs("x-request-id", "req-123")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		sc := NewServerContext(ctx, "/svc/m")
		assert.Equal(t, "req-123", sc.RequestID())
	})

	t.Run("without request id", func(t *testing.T) {
		sc := NewServerContext(context.Background(), "/svc/m")
		assert.Equal(t, "unknown", sc.RequestID())
	})
}

func TestServerContext_Peer(t *testing.T) {
	sc := NewServerContext(context.Background(), "/svc/m")
	assert.Empty(t, sc.Peer())
}

func TestServerContext_IsSecure(t *testing.T) {
	sc := NewServerContext(context.Background(), "/svc/m")
	assert.False(t, sc.IsSecure())
}

func TestServerContext_MetadataValues(t *testing.T) {
	md := metadata.Pairs("key", "val1", "key", "val2")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	sc := NewServerContext(ctx, "/svc/m")

	vals := sc.GetMetadataValues("key")
	assert.Len(t, vals, 2)
}

func TestServerContext_IncomingMetadata(t *testing.T) {
	t.Run("with metadata", func(t *testing.T) {
		md := metadata.Pairs("k", "v")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		sc := NewServerContext(ctx, "/svc/m")
		assert.NotEmpty(t, sc.IncomingMetadata())
	})

	t.Run("without metadata", func(t *testing.T) {
		sc := NewServerContext(context.Background(), "/svc/m")
		assert.Empty(t, sc.IncomingMetadata())
	})
}

func TestSplitFullMethod(t *testing.T) {
	tests := []struct {
		input   string
		service string
		method  string
	}{
		{"/pkg.Svc/Method", "pkg.Svc", "Method"},
		{"pkg.Svc/Method", "pkg.Svc", "Method"},
		{"/Service/", "Service", ""},
		{"NoSlash", "NoSlash", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		s, m := splitFullMethod(tt.input)
		assert.Equal(t, tt.service, s)
		assert.Equal(t, tt.method, m)
	}
}

func TestServerContext_Context(t *testing.T) {
	ctx := context.Background()
	sc := NewServerContext(ctx, "/svc/m")
	assert.Equal(t, ctx, sc.Context())
}

func TestServerContext_SetOutgoingHeader(t *testing.T) {
	// SetOutgoingHeader calls grpc.SetHeader which will fail outside a gRPC stream,
	// but we just need to exercise the code path for coverage.
	sc := NewServerContext(context.Background(), "/svc/m")
	// Should not panic
	sc.SetOutgoingHeader("x-custom", "v1", "v2")
}

func TestServerContext_SetOutgoingTrailer(t *testing.T) {
	sc := NewServerContext(context.Background(), "/svc/m")
	// Should not panic
	sc.SetOutgoingTrailer("x-trailer", "v1")
}

func TestServerContext_TraceID(t *testing.T) {
	sc := NewServerContext(context.Background(), "/svc/m")
	// Without trace context, TraceID returns ""
	assert.Empty(t, sc.TraceID())
}
