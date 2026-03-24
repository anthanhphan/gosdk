// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package tracing provides a unified interface for distributed tracing
// with OpenTelemetry as the default backend. It supports exporting traces
// to Tempo, Jaeger, or any OTLP-compatible collector.
//
// The package follows the same design patterns as the metrics package:
//   - Client interface for abstraction
//   - OpenTelemetry implementation for production
//   - NoopClient for testing and when tracing is disabled
//   - Functional options for configuration
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithEndpoint("localhost:4317"),
//	    tracing.WithInsecure(),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Shutdown(context.Background())
//
//	ctx, span := client.StartSpan(ctx, "my-operation")
//	defer span.End()
package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen --source=client.go -destination=./mocks/client.go -package=mocks

// ============================================================================
// Client Interface
// ============================================================================

// Client is the main interface for distributed tracing.
// It provides methods for creating spans and managing the tracer lifecycle.
type Client interface {
	// StartSpan starts a new span with the given name and returns the updated
	// context containing the span, along with a Span handle for adding
	// attributes, events, and ending the span.
	//
	// The returned context should be passed to downstream operations
	// to propagate trace context.
	//
	// Example:
	//
	//	ctx, span := client.StartSpan(ctx, "db.query",
	//	    tracing.WithSpanKind(tracing.SpanKindClient),
	//	    tracing.WithAttributes(attribute.String("db.system", "postgres")),
	//	)
	//	defer span.End()
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// Shutdown flushes any pending spans and releases resources.
	// Should be called before application exit.
	Shutdown(ctx context.Context) error

	// Tracer returns the underlying OTel tracer for advanced use cases
	// that require direct access to the OpenTelemetry API.
	Tracer() trace.Tracer
}

// ============================================================================
// Span Interface
// ============================================================================

// Span represents an active trace span. It wraps the OpenTelemetry span
// interface with a simplified API.
type Span interface {
	// End completes the span. Must be called when the operation is done.
	// After End is called, the span should not be modified.
	End()

	// SetAttributes sets key-value attributes on the span.
	// Use this to record contextual information about the operation.
	//
	// Example:
	//
	//	span.SetAttributes(
	//	    attribute.String("http.method", "GET"),
	//	    attribute.Int("http.status_code", 200),
	//	)
	SetAttributes(attrs ...attribute.KeyValue)

	// SetStatus sets the status of the span.
	// Use codes.Error for failed operations, codes.Ok for explicit success.
	//
	// Example:
	//
	//	span.SetStatus(codes.Error, "database connection failed")
	SetStatus(code codes.Code, description string)

	// RecordError records an error on the span as an exception event.
	// This does not change the span status; use SetStatus for that.
	//
	// Example:
	//
	//	if err != nil {
	//	    span.RecordError(err)
	//	    span.SetStatus(codes.Error, err.Error())
	//	}
	RecordError(err error)

	// SetName updates the name of the span after creation.
	// This is useful when the final operation name is not known at span start
	// (e.g., HTTP route pattern resolved after middleware creates the span).
	//
	// Example:
	//
	//	span.SetName("POST /api/orders")
	SetName(name string)

	// AddEvent adds a timestamped event to the span.
	// Events represent something that occurred during the span's lifetime.
	//
	// Example:
	//
	//	span.AddEvent("cache.miss", attribute.String("key", "user:123"))
	AddEvent(name string, attrs ...attribute.KeyValue)

	// SpanContext returns the SpanContext of this span.
	// Useful for extracting trace ID, span ID, and trace flags.
	SpanContext() trace.SpanContext
}

// ============================================================================
// Span Kind
// ============================================================================

// SpanKind represents the relationship of a span to its parent and children.
type SpanKind int

const (
	// SpanKindInternal is the default span kind for internal operations.
	SpanKindInternal SpanKind = iota
	// SpanKindServer indicates the span covers server-side handling of an RPC or HTTP request.
	SpanKindServer
	// SpanKindClient indicates the span covers the client-side of an RPC or HTTP request.
	SpanKindClient
	// SpanKindProducer indicates the span covers the sending of a message to a queue.
	SpanKindProducer
	// SpanKindConsumer indicates the span covers the receiving of a message from a queue.
	SpanKindConsumer
)

// toOTelSpanKind converts our SpanKind to the OTel SpanKind.
func (k SpanKind) toOTelSpanKind() trace.SpanKind {
	switch k {
	case SpanKindServer:
		return trace.SpanKindServer
	case SpanKindClient:
		return trace.SpanKindClient
	case SpanKindProducer:
		return trace.SpanKindProducer
	case SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

// ============================================================================
// Span Options
// ============================================================================

// SpanOption configures how a span is created.
type SpanOption func(*spanOptions)

// spanOptions holds configuration for span creation.
type spanOptions struct {
	kind       SpanKind
	attributes []attribute.KeyValue
}

// WithSpanKind sets the kind of span being created.
//
// Example:
//
//	ctx, span := client.StartSpan(ctx, "db.query",
//	    tracing.WithSpanKind(tracing.SpanKindClient),
//	)
func WithSpanKind(kind SpanKind) SpanOption {
	return func(o *spanOptions) {
		o.kind = kind
	}
}

// WithAttributes sets initial attributes on the span.
//
// Example:
//
//	ctx, span := client.StartSpan(ctx, "http.request",
//	    tracing.WithAttributes(
//	        attribute.String("http.method", "GET"),
//	        attribute.String("http.url", "/users"),
//	    ),
//	)
func WithAttributes(attrs ...attribute.KeyValue) SpanOption {
	return func(o *spanOptions) {
		o.attributes = append(o.attributes, attrs...)
	}
}
