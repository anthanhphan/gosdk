// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

// wrappedServerStream wraps grpc.ServerStream to propagate a modified context.
// Used by interceptors that inject values into the context (request ID, tracing, auth).
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedServerStream) Context() context.Context {
	return s.ctx
}
