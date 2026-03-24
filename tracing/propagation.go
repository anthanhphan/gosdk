// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// ============================================================================
// Trace Context Extraction
// ============================================================================

// TraceIDFromContext extracts the trace ID from an active span in the context.
// Returns an empty string if no active span exists or the span context is invalid.
//
// Example:
//
//	traceID := tracing.TraceIDFromContext(ctx)
//	logger.Infow("processing request", "trace_id", traceID)
func TraceIDFromContext(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

// SpanIDFromContext extracts the span ID from an active span in the context.
// Returns an empty string if no active span exists or the span context is invalid.
//
// Example:
//
//	spanID := tracing.SpanIDFromContext(ctx)
func SpanIDFromContext(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasSpanID() {
		return ""
	}
	return sc.SpanID().String()
}

// IsSampled returns true if the span in the context is being sampled/recorded.
func IsSampled(ctx context.Context) bool {
	return trace.SpanContextFromContext(ctx).IsSampled()
}

// ============================================================================
// Header Propagation
// ============================================================================

// HeaderCarrier is a map-based carrier for HTTP headers.
// It implements propagation.TextMapCarrier for injecting/extracting
// trace context from HTTP headers.
type HeaderCarrier map[string]string

// Get returns the value associated with the passed key.
func (c HeaderCarrier) Get(key string) string {
	return c[key]
}

// Set stores the key-value pair.
func (c HeaderCarrier) Set(key, value string) {
	c[key] = value
}

// Keys returns the keys for which this carrier has a value.
func (c HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// ExtractContext extracts trace context from carrier headers into a Go context.
// This is useful for extracting traceparent/tracestate headers from incoming
// HTTP or gRPC requests.
//
// Example:
//
//	headers := tracing.HeaderCarrier{
//	    "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
//	}
//	ctx = tracing.ExtractContext(ctx, headers)
func ExtractContext(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// InjectContext injects trace context from a Go context into carrier headers.
// This is useful for propagating trace context to outgoing HTTP or gRPC requests.
//
// Example:
//
//	headers := make(tracing.HeaderCarrier)
//	tracing.InjectContext(ctx, headers)
//	// headers now contains "traceparent" and optionally "tracestate"
func InjectContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}
