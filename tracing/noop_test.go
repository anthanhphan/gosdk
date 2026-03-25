package tracing

import (
	"context"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TestNoopClient_Coverage(t *testing.T) {
	client := NewNoopClient()

	if client.Tracer() == nil {
		t.Error("expected tracer")
	}

	ctx := context.Background()
	ctx, span := client.StartSpan(ctx, "noop-span")

	span.SetName("changed")
	span.SetStatus(codes.Error, "err")
	span.SetAttributes(attribute.String("k", "v"))
	span.RecordError(fmt.Errorf("err"))
	span.AddEvent("event")
	span.SpanContext()
	span.End()

	if err := client.Shutdown(ctx); err != nil {
		t.Errorf("noop shutdown shouldn't error: %v", err)
	}
}

func TestPropagation_Coverage(t *testing.T) {
	ctx := context.Background()

	traceID := TraceIDFromContext(ctx)
	if traceID != "" {
		t.Errorf("expected empty trace ID, got %s", traceID)
	}

	spanID := SpanIDFromContext(ctx)
	if spanID != "" {
		t.Errorf("expected empty span ID, got %s", spanID)
	}

	// For coverage, just ensure it doesn't panic on empty context
	sampled := IsSampled(ctx)
	if sampled {
		t.Error("expected false for un-sampled empty context")
	}
}
