// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ============================================================================
// NoopClient Tests
// ============================================================================

func TestNoopClient(t *testing.T) {
	client := NewNoopClient()
	ctx := context.Background()

	t.Run("StartSpan returns context and span", func(t *testing.T) {
		newCtx, span := client.StartSpan(ctx, "test-operation")
		if newCtx == nil {
			t.Fatal("expected non-nil context")
		}
		if span == nil {
			t.Fatal("expected non-nil span")
		}
		span.End() // should not panic
	})

	t.Run("Span operations are no-ops", func(t *testing.T) {
		_, span := client.StartSpan(ctx, "test-operation")

		// All operations should be no-ops (no panics)
		span.SetAttributes(attribute.String("key", "value"))
		span.SetStatus(codes.Error, "test error")
		span.RecordError(context.DeadlineExceeded)
		span.AddEvent("test-event", attribute.Int("count", 42))

		sc := span.SpanContext()
		if sc.HasTraceID() {
			t.Error("noop span should not have a valid trace ID")
		}
		if sc.HasSpanID() {
			t.Error("noop span should not have a valid span ID")
		}

		span.End()
	})

	t.Run("Shutdown returns nil", func(t *testing.T) {
		if err := client.Shutdown(ctx); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("Tracer returns non-nil", func(t *testing.T) {
		tracer := client.Tracer()
		if tracer == nil {
			t.Fatal("expected non-nil tracer")
		}
	})
}

// ============================================================================
// SpanKind Tests
// ============================================================================

func TestSpanKind_toOTelSpanKind(t *testing.T) {
	tests := []struct {
		name     string
		kind     SpanKind
		expected trace.SpanKind
	}{
		{"Internal", SpanKindInternal, trace.SpanKindInternal},
		{"Server", SpanKindServer, trace.SpanKindServer},
		{"Client", SpanKindClient, trace.SpanKindClient},
		{"Producer", SpanKindProducer, trace.SpanKindProducer},
		{"Consumer", SpanKindConsumer, trace.SpanKindConsumer},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.kind.toOTelSpanKind()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// ============================================================================
// SpanOption Tests
// ============================================================================

func TestWithSpanKind(t *testing.T) {
	opts := &spanOptions{}
	WithSpanKind(SpanKindClient)(opts)

	if opts.kind != SpanKindClient {
		t.Errorf("expected SpanKindClient, got %v", opts.kind)
	}
}

func TestWithAttributes(t *testing.T) {
	opts := &spanOptions{}
	attrs := []attribute.KeyValue{
		attribute.String("key1", "val1"),
		attribute.Int("key2", 42),
	}
	WithAttributes(attrs...)(opts)

	if len(opts.attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(opts.attributes))
	}
}

func TestWithAttributes_Append(t *testing.T) {
	opts := &spanOptions{}
	WithAttributes(attribute.String("key1", "val1"))(opts)
	WithAttributes(attribute.String("key2", "val2"))(opts)

	if len(opts.attributes) != 2 {
		t.Errorf("expected 2 attributes after two calls, got %d", len(opts.attributes))
	}
}

// ============================================================================
// Propagation Tests
// ============================================================================

func TestTraceIDFromContext_NoSpan(t *testing.T) {
	ctx := context.Background()
	traceID := TraceIDFromContext(ctx)
	if traceID != "" {
		t.Errorf("expected empty trace ID, got %q", traceID)
	}
}

func TestSpanIDFromContext_NoSpan(t *testing.T) {
	ctx := context.Background()
	spanID := SpanIDFromContext(ctx)
	if spanID != "" {
		t.Errorf("expected empty span ID, got %q", spanID)
	}
}

func TestIsSampled_NoSpan(t *testing.T) {
	ctx := context.Background()
	if IsSampled(ctx) {
		t.Error("expected not sampled for empty context")
	}
}

func TestHeaderCarrier(t *testing.T) {
	carrier := HeaderCarrier{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"tracestate":  "vendor=value",
	}

	t.Run("Get returns existing value", func(t *testing.T) {
		val := carrier.Get("traceparent")
		if val == "" {
			t.Error("expected non-empty traceparent")
		}
	})

	t.Run("Get returns empty for missing key", func(t *testing.T) {
		val := carrier.Get("nonexistent")
		if val != "" {
			t.Errorf("expected empty value, got %q", val)
		}
	})

	t.Run("Set adds/overwrites value", func(t *testing.T) {
		carrier.Set("newkey", "newvalue")
		if carrier.Get("newkey") != "newvalue" {
			t.Error("expected Set to store value")
		}
	})

	t.Run("Keys returns all keys", func(t *testing.T) {
		keys := carrier.Keys()
		if len(keys) < 2 {
			t.Errorf("expected at least 2 keys, got %d", len(keys))
		}
	})
}

func TestExtractAndInjectContext(t *testing.T) {
	ctx := context.Background()

	// Extract from empty carrier should not panic
	emptyCarrier := HeaderCarrier{}
	ctx = ExtractContext(ctx, emptyCarrier)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Inject into carrier should not panic
	outputCarrier := make(HeaderCarrier)
	InjectContext(ctx, outputCarrier)
	// No active span, so no traceparent should be injected
}

// ============================================================================
// Client Options Tests
// ============================================================================

func TestDefaultClientOptions(t *testing.T) {
	opts := defaultClientOptions()

	if opts.endpoint != "localhost:4317" {
		t.Errorf("expected default endpoint 'localhost:4317', got %q", opts.endpoint)
	}
	if opts.insecure {
		t.Error("expected insecure to be false by default")
	}
}

func TestWithServiceVersion(t *testing.T) {
	opts := defaultClientOptions()
	WithServiceVersion("v1.2.3")(opts)

	if opts.serviceVersion != "v1.2.3" {
		t.Errorf("expected version 'v1.2.3', got %q", opts.serviceVersion)
	}
}

func TestWithEnvironment(t *testing.T) {
	opts := defaultClientOptions()
	WithEnvironment("production")(opts)

	if opts.environment != "production" {
		t.Errorf("expected environment 'production', got %q", opts.environment)
	}
}

func TestWithEndpoint(t *testing.T) {
	opts := defaultClientOptions()
	WithEndpoint("collector:4317")(opts)

	if opts.endpoint != "collector:4317" {
		t.Errorf("expected endpoint 'collector:4317', got %q", opts.endpoint)
	}
}

func TestWithEndpoint_EmptyIgnored(t *testing.T) {
	opts := defaultClientOptions()
	WithEndpoint("")(opts)

	if opts.endpoint != "localhost:4317" {
		t.Errorf("expected default endpoint preserved, got %q", opts.endpoint)
	}
}

func TestWithInsecure(t *testing.T) {
	opts := defaultClientOptions()
	WithInsecure()(opts)

	if !opts.insecure {
		t.Error("expected insecure to be true")
	}
}

func TestWithSamplingRate(t *testing.T) {
	opts := defaultClientOptions()
	WithSamplingRate(0.5)(opts)

	if opts.sampler == nil {
		t.Error("expected non-nil sampler")
	}
}

func TestWithHeaders(t *testing.T) {
	opts := defaultClientOptions()
	headers := map[string]string{
		"Authorization": "Bearer token",
	}
	WithHeaders(headers)(opts)

	if opts.headers == nil {
		t.Fatal("expected non-nil headers")
	}
	if opts.headers["Authorization"] != "Bearer token" {
		t.Error("expected Authorization header")
	}
}
