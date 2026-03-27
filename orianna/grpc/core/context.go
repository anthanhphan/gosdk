// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/tracing"
)

// Context defines the interface for gRPC request context operations.
// It provides a clean, testable API that is framework-agnostic.
type Context interface {
	// Access underlying Go context
	Context() context.Context

	// Request information
	FullMethod() string
	ServiceName() string
	MethodName() string
	Peer() string
	IsSecure() bool

	// Metadata operations
	GetMetadata(key string) string
	GetMetadataValues(key string) []string
	IncomingMetadata() metadata.MD
	SetOutgoingHeader(key string, values ...string)
	SetOutgoingTrailer(key string, values ...string)

	// Local storage
	GetLocal(key string) any
	SetLocal(key string, value any)

	// Utility methods
	RequestID() string

	// Trace ID for distributed tracing correlation
	TraceID() string
}

// Compile-time interface compliance check.
var _ Context = (*ServerContext)(nil)

// ServerContext implements the Context interface for gRPC server-side operations.
type ServerContext struct {
	ctx         context.Context
	fullMethod  string
	serviceName string
	methodName  string
	locals      map[string]any
	mu          sync.RWMutex
	cachedMD    metadata.MD
	mdOnce      sync.Once
}

// NewServerContext creates a new ServerContext from a Go context and full method name.
func NewServerContext(ctx context.Context, fullMethod string) *ServerContext {
	service, method := splitFullMethod(fullMethod)
	return &ServerContext{
		ctx:         ctx,
		fullMethod:  fullMethod,
		serviceName: service,
		methodName:  method,
	}
}

// incomingMD returns the cached incoming metadata, parsing it on first access.
func (c *ServerContext) incomingMD() metadata.MD {
	c.mdOnce.Do(func() {
		c.cachedMD, _ = metadata.FromIncomingContext(c.ctx)
	})
	return c.cachedMD
}

// Context returns the underlying Go context.
func (c *ServerContext) Context() context.Context { return c.ctx }

// FullMethod returns the full gRPC method name (e.g., "/package.Service/Method").
func (c *ServerContext) FullMethod() string { return c.fullMethod }

// ServiceName returns the gRPC service name.
func (c *ServerContext) ServiceName() string { return c.serviceName }

// MethodName returns the gRPC method name.
func (c *ServerContext) MethodName() string { return c.methodName }

// Peer returns the address of the client.
func (c *ServerContext) Peer() string {
	p, ok := peer.FromContext(c.ctx)
	if !ok || p.Addr == nil {
		return ""
	}
	return p.Addr.String()
}

// IsSecure returns whether the connection uses TLS.
func (c *ServerContext) IsSecure() bool {
	p, ok := peer.FromContext(c.ctx)
	if !ok {
		return false
	}
	return p.AuthInfo != nil
}

// GetMetadata returns the first value of the given metadata key from incoming metadata.
func (c *ServerContext) GetMetadata(key string) string {
	values := c.incomingMD().Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// GetMetadataValues returns all values of the given metadata key from incoming metadata.
func (c *ServerContext) GetMetadataValues(key string) []string {
	return c.incomingMD().Get(key)
}

// IncomingMetadata returns the incoming metadata from the request.
func (c *ServerContext) IncomingMetadata() metadata.MD {
	md := c.incomingMD()
	if md == nil {
		return metadata.MD{}
	}
	return md
}

// SetOutgoingHeader sets a header in the server response metadata.
func (c *ServerContext) SetOutgoingHeader(key string, values ...string) {
	md := metadata.MD{}
	for _, v := range values {
		md.Append(key, v)
	}
	if err := grpc.SetHeader(c.ctx, md); err != nil {
		logger.Warnw("failed to set outgoing header", "key", key, "error", err)
	}
}

// SetOutgoingTrailer sets a trailer in the server response metadata.
func (c *ServerContext) SetOutgoingTrailer(key string, values ...string) {
	md := metadata.MD{}
	for _, v := range values {
		md.Append(key, v)
	}
	if err := grpc.SetTrailer(c.ctx, md); err != nil {
		logger.Warnw("failed to set outgoing trailer", "key", key, "error", err)
	}
}

// GetLocal returns a request-scoped local value.
func (c *ServerContext) GetLocal(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.locals == nil {
		return nil
	}
	return c.locals[key]
}

// SetLocal sets a request-scoped local value.
func (c *ServerContext) SetLocal(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.locals == nil {
		c.locals = make(map[string]any, 4)
	}
	c.locals[key] = value
}

// RequestID returns the request ID from metadata or "unknown".
func (c *ServerContext) RequestID() string {
	reqID := c.GetMetadata(MetadataKeyRequestID)
	if reqID == "" {
		return DefaultUnknownRequestID
	}
	return reqID
}

// TraceID returns the trace ID from the context for distributed tracing correlation.
// Returns empty string if no trace ID is present.
func (c *ServerContext) TraceID() string {
	return tracing.TraceIDFromContext(c.ctx)
}

// splitFullMethod splits a gRPC full method name (e.g., "/package.Service/Method")
// into service name and method name.
func splitFullMethod(fullMethod string) (string, string) {
	if fullMethod == "" {
		return "", ""
	}
	// Remove leading "/"
	if fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return fullMethod[:i], fullMethod[i+1:]
		}
	}
	return fullMethod, ""
}
