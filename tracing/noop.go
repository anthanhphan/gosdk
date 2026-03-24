// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// ============================================================================
// NoopClient (No-Operation Client)
// ============================================================================

// Compile-time interface compliance checks.
var (
	_ Client = (*noopClient)(nil)
	_ Span   = (*noopSpan)(nil)
)

// noopClient is a no-operation implementation of the Client interface.
// All methods are no-ops that return immediately. This is useful for:
//   - Testing environments where tracing is not needed
//   - Disabling tracing without changing application code
//   - Providing a safe default when tracing is optional
//
// Example:
//
//	client := tracing.NewNoopClient()
//	ctx, span := client.StartSpan(ctx, "operation") // does nothing
//	span.End()                                        // does nothing
type noopClient struct {
	tracer trace.Tracer
}

// NewNoopClient creates a new no-operation tracing client.
// All tracing operations are silently discarded.
//
// Example:
//
//	// Use in tests
//	client := tracing.NewNoopClient()
//
//	// Use as default when tracing is optional
//	var tracingClient tracing.Client = tracing.NewNoopClient()
//	if enableTracing {
//	    tracingClient, _ = tracing.NewClient("myapp")
//	}
func NewNoopClient() Client {
	return &noopClient{
		tracer: noop.NewTracerProvider().Tracer("noop"),
	}
}

func (c *noopClient) StartSpan(ctx context.Context, _ string, _ ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

func (c *noopClient) Shutdown(_ context.Context) error {
	return nil
}

func (c *noopClient) Tracer() trace.Tracer {
	return c.tracer
}

// ============================================================================
// NoopSpan
// ============================================================================

// noopSpan is a no-operation implementation of the Span interface.
type noopSpan struct{}

func (*noopSpan) End()                                       {}
func (*noopSpan) SetAttributes(_ ...attribute.KeyValue)      {}
func (*noopSpan) SetStatus(_ codes.Code, _ string)           {}
func (*noopSpan) RecordError(_ error)                        {}
func (*noopSpan) SetName(_ string)                           {}
func (*noopSpan) AddEvent(_ string, _ ...attribute.KeyValue) {}
func (*noopSpan) SpanContext() trace.SpanContext             { return trace.SpanContext{} }
