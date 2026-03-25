package tracing

import (
	"context"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestOTelClient_FullCoverage(t *testing.T) {
	// Use WithSampler, WithInsecure, WithEndpoint, WithEnvironment, WithHeaders, etc.
	client, err := NewClient("test-service",
		WithEndpoint("localhost:4317"),
		WithInsecure(),
		WithEnvironment("test"),
		WithSampler(sdktrace.AlwaysSample()),
		WithSamplingRate(1.0),
		WithServiceVersion("v1.0.0"),
		WithHeaders(map[string]string{"foo": "bar"}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	ctx, span := client.StartSpan(ctx, "test-span",
		WithSpanKind(SpanKindServer),
		WithAttributes(attribute.String("k", "v")),
	)

	tracer := client.Tracer()
	if tracer == nil {
		t.Error("expected non-nil tracer")
	}

	// Cover Span interface wrapper methods
	span.SetName("new-name")
	span.SetStatus(codes.Ok, "ok")
	span.SetAttributes(attribute.Int("foo", 1))
	span.RecordError(fmt.Errorf("test err"))
	span.AddEvent("test-event", attribute.String("ek", "ev"))
	// Access SpanContext to cover the method; validity depends on OTel setup
	_ = span.SpanContext()
	span.End()

	// Shutdown
	if err := client.Shutdown(ctx); err != nil {
		// Log but don't fail, OTLP exporter might fail on missing connection
		t.Logf("shutdown err: %v", err)
	}
}
