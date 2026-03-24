// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/anthanhphan/gosdk/tracing"
)

func main() {
	fmt.Println("=== tracing Example ===")
	fmt.Println()

	ctx := context.Background()

	// Example 1: NoopClient (safe for running without a collector)
	exampleNoopClient(ctx)

	// Example 2: Creating Spans with Nesting
	exampleNestedSpans(ctx)

	// Example 3: Span Attributes
	exampleSpanAttributes(ctx)

	// Example 4: Error Recording
	exampleErrorRecording(ctx)

	// Example 5: Events
	exampleEvents(ctx)

	// Example 6: Trace Context Propagation
	examplePropagation(ctx)

	// Example 7: Configuration Options
	exampleOptions()

	// Example 8: Full Service Pattern
	exampleFullServicePattern(ctx)
}

// exampleNoopClient demonstrates the NoopClient for testing and environments
// where tracing is not available.
func exampleNoopClient(ctx context.Context) {
	fmt.Println("1. NoopClient (for testing / no collector):")

	client := tracing.NewNoopClient()

	// All operations are silently discarded -- no panics, no allocations
	ctx, span := client.StartSpan(ctx, "my-operation",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(attribute.String("key", "value")),
	)
	span.SetAttributes(attribute.Int("http.status_code", 200))
	span.AddEvent("cache.hit", attribute.String("key", "user:123"))
	span.SetStatus(codes.Ok, "")
	span.End()

	// Shutdown is also a no-op
	if err := client.Shutdown(ctx); err != nil {
		fmt.Printf("   - Error: %v\n", err)
	}

	// Use as a default when tracing is optional
	enableTracing := false
	var tc tracing.Client = tracing.NewNoopClient()
	if enableTracing {
		tc, _ = tracing.NewClient("my-service",
			tracing.WithEndpoint("localhost:4317"),
			tracing.WithInsecure(),
		)
	}
	_ = tc

	fmt.Println("   - All tracing operations silently discarded")
	fmt.Println("   - Safe to use as a default fallback")
	fmt.Println("   - Shutdown is a no-op")
	fmt.Println()
}

// exampleNestedSpans demonstrates creating parent-child span relationships.
func exampleNestedSpans(ctx context.Context) {
	fmt.Println("2. Nested Spans (parent -> child hierarchy):")

	client := tracing.NewNoopClient()

	// Root span -- typically created by middleware
	ctx, rootSpan := client.StartSpan(ctx, "GET /users/:id",
		tracing.WithSpanKind(tracing.SpanKindServer),
	)
	defer rootSpan.End()
	fmt.Println("   - Created root span: GET /users/:id (Server)")

	// Child span -- created in handler for business logic
	ctx, bizSpan := client.StartSpan(ctx, "user.validate",
		tracing.WithSpanKind(tracing.SpanKindInternal),
	)
	bizSpan.SetStatus(codes.Ok, "")
	bizSpan.End()
	fmt.Println("   - Created child span: user.validate (Internal)")

	// Grandchild span -- created for external call
	_, dbSpan := client.StartSpan(ctx, "db.query",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.system", "postgres"),
			attribute.String("db.statement", "SELECT * FROM users WHERE id = $1"),
		),
	)
	dbSpan.SetStatus(codes.Ok, "")
	dbSpan.End()
	fmt.Println("   - Created grandchild span: db.query (Client)")

	fmt.Println()
	fmt.Println("   Resulting trace hierarchy:")
	fmt.Println("   GET /users/:id (Server)")
	fmt.Println("     +-- user.validate (Internal)")
	fmt.Println("           +-- db.query (Client)")
	fmt.Println()
}

// exampleSpanAttributes demonstrates adding attributes to spans.
func exampleSpanAttributes(ctx context.Context) {
	fmt.Println("3. Span Attributes:")

	client := tracing.NewNoopClient()

	// Set attributes at span creation
	ctx, span := client.StartSpan(ctx, "http.request",
		tracing.WithAttributes(
			attribute.String("http.method", "POST"),
			attribute.String("http.route", "/api/orders"),
			attribute.String("http.scheme", "https"),
		),
	)
	defer span.End()
	fmt.Println("   - Set attributes at creation: http.method, http.route, http.scheme")

	// Add attributes after creation (e.g., after processing)
	span.SetAttributes(
		attribute.Int("http.status_code", 201),
		attribute.Int64("http.response_content_length", 256),
		attribute.String("http.user_agent", "Go-SDK/1.0"),
	)
	fmt.Println("   - Added attributes after processing: http.status_code, http.response_content_length")

	// Common attribute patterns
	_, dbSpan := client.StartSpan(ctx, "db.query",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.system", "postgres"),
			attribute.String("db.name", "orders_db"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.statement", "SELECT * FROM orders WHERE status = $1"),
		),
	)
	dbSpan.End()
	fmt.Println("   - Database span attributes: db.system, db.name, db.operation, db.statement")

	_, cacheSpan := client.StartSpan(ctx, "cache.get",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("cache.system", "redis"),
			attribute.String("cache.key", "user:123"),
			attribute.Bool("cache.hit", true),
		),
	)
	cacheSpan.End()
	fmt.Println("   - Cache span attributes: cache.system, cache.key, cache.hit")
	fmt.Println()
}

// exampleErrorRecording demonstrates how to record errors on spans.
func exampleErrorRecording(ctx context.Context) {
	fmt.Println("4. Error Recording:")

	client := tracing.NewNoopClient()

	// Pattern 1: RecordError + SetStatus
	ctx, span := client.StartSpan(ctx, "process.order")
	err := fmt.Errorf("insufficient inventory for item SKU-001")
	if err != nil {
		span.RecordError(err)                    // Adds error as event with stack trace
		span.SetStatus(codes.Error, err.Error()) // Marks span as failed
	}
	span.End()
	fmt.Println("   - Recorded error: insufficient inventory for item SKU-001")
	fmt.Println("   - RecordError adds the error as a span event")
	fmt.Println("   - SetStatus marks the span as failed in the trace UI")

	// Pattern 2: Success path
	_, successSpan := client.StartSpan(ctx, "validate.input")
	// ... validation succeeded ...
	successSpan.SetStatus(codes.Ok, "")
	successSpan.End()
	fmt.Println("   - Success: SetStatus(codes.Ok, \"\")")

	// Pattern 3: Multiple errors in one span
	_, multiSpan := client.StartSpan(ctx, "batch.process")
	multiSpan.RecordError(fmt.Errorf("item 1 failed: invalid format"))
	multiSpan.RecordError(fmt.Errorf("item 3 failed: missing required field"))
	multiSpan.SetStatus(codes.Error, "2 of 5 items failed")
	multiSpan.End()
	fmt.Println("   - Multiple errors recorded on a single span")
	fmt.Println()
}

// exampleEvents demonstrates adding events (annotations) to spans.
func exampleEvents(ctx context.Context) {
	fmt.Println("5. Span Events:")

	client := tracing.NewNoopClient()

	ctx, span := client.StartSpan(ctx, "user.signup")
	defer span.End()

	// Events mark specific moments within a span's lifetime
	span.AddEvent("validation.started")
	fmt.Println("   - Event: validation.started")

	span.AddEvent("validation.passed",
		attribute.Int("rules_checked", 5),
	)
	fmt.Println("   - Event: validation.passed (rules_checked=5)")

	span.AddEvent("email.sent",
		attribute.String("email.to", "user@example.com"),
		attribute.String("email.template", "welcome"),
	)
	fmt.Println("   - Event: email.sent (to=user@example.com, template=welcome)")

	span.AddEvent("user.created",
		attribute.String("user.id", "usr-12345"),
		attribute.String("user.role", "member"),
	)
	fmt.Println("   - Event: user.created (id=usr-12345, role=member)")

	// Real-world: track retry attempts
	_, retrySpan := client.StartSpan(ctx, "external.api.call")
	for i := 1; i <= 3; i++ {
		retrySpan.AddEvent("retry.attempt",
			attribute.Int("attempt", i),
			attribute.String("reason", "connection_timeout"),
		)
	}
	retrySpan.End()
	fmt.Println("   - Retry tracking: 3 attempt events on a single span")
	fmt.Println()
}

// examplePropagation demonstrates trace context propagation across service boundaries.
func examplePropagation(ctx context.Context) {
	fmt.Println("6. Trace Context Propagation:")

	client := tracing.NewNoopClient()

	// Start a span
	ctx, span := client.StartSpan(ctx, "api.handler",
		tracing.WithSpanKind(tracing.SpanKindServer),
	)
	defer span.End()

	// Extract trace/span IDs from context
	traceID := tracing.TraceIDFromContext(ctx)
	spanID := tracing.SpanIDFromContext(ctx)
	isSampled := tracing.IsSampled(ctx)
	fmt.Printf("   - TraceID: %q\n", traceID)
	fmt.Printf("   - SpanID: %q\n", spanID)
	fmt.Printf("   - IsSampled: %v\n", isSampled)

	// Inject context into outgoing HTTP headers
	headers := make(tracing.HeaderCarrier)
	tracing.InjectContext(ctx, headers)
	fmt.Printf("   - Injected into headers: %v\n", headers)

	// Extract context from incoming HTTP headers (simulated)
	incomingHeaders := tracing.HeaderCarrier{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"tracestate":  "vendor=value",
	}
	extractedCtx := tracing.ExtractContext(context.Background(), incomingHeaders)
	extractedTraceID := tracing.TraceIDFromContext(extractedCtx)
	fmt.Printf("   - Extracted trace ID from headers: %q\n", extractedTraceID)

	// Cross-service propagation pattern:
	// Service A -> Service B
	fmt.Println()
	fmt.Println("   Cross-service propagation flow:")
	fmt.Println("   1. Service A creates a span")
	fmt.Println("   2. Service A injects context into outgoing headers (traceparent)")
	fmt.Println("   3. Service B extracts context from incoming headers")
	fmt.Println("   4. Service B creates child span -> automatically linked to Service A's trace")
	fmt.Println()
}

// exampleOptions demonstrates all available configuration options.
func exampleOptions() {
	fmt.Println("7. Configuration Options:")

	// Note: NewClient connects to a real OTLP collector, so we only
	// demonstrate the option patterns here without calling NewClient.

	fmt.Println("   // Basic -- connect to local collector")
	fmt.Println(`   client, _ := tracing.NewClient("my-service",`)
	fmt.Println(`       tracing.WithEndpoint("localhost:4317"),`)
	fmt.Println(`       tracing.WithInsecure(),`)
	fmt.Println("   )")
	fmt.Println()

	fmt.Println("   // Production -- with sampling, metadata, and TLS")
	fmt.Println(`   client, _ := tracing.NewClient("order-service",`)
	fmt.Println(`       tracing.WithEndpoint("otel-collector.monitoring:4317"),`)
	fmt.Println(`       tracing.WithServiceVersion("v2.1.0"),`)
	fmt.Println(`       tracing.WithEnvironment("production"),`)
	fmt.Println(`       tracing.WithSamplingRate(0.1), // 10% of traces`)
	fmt.Println("   )")
	fmt.Println()

	fmt.Println("   // Cloud / SaaS -- with auth headers")
	fmt.Println(`   client, _ := tracing.NewClient("my-service",`)
	fmt.Println(`       tracing.WithEndpoint("otel.grafana.net:443"),`)
	fmt.Println(`       tracing.WithHeaders(map[string]string{`)
	fmt.Println(`           "Authorization": "Basic <api-key>",`)
	fmt.Println("       }),")
	fmt.Println("   )")
	fmt.Println()
}

// exampleFullServicePattern demonstrates a complete end-to-end trace pattern
// that would occur in a real microservice.
func exampleFullServicePattern(ctx context.Context) {
	fmt.Println("8. Full Service Pattern (end-to-end):")

	client := tracing.NewNoopClient()

	// Simulate an incoming HTTP request with traceparent
	incomingHeaders := tracing.HeaderCarrier{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	}
	ctx = tracing.ExtractContext(ctx, incomingHeaders)
	fmt.Println("   -> Incoming request with traceparent header")

	// Root span (normally created by TracingMiddleware)
	ctx, rootSpan := client.StartSpan(ctx, "POST /api/orders",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", "POST"),
			attribute.String("http.route", "/api/orders"),
			attribute.String("net.peer.ip", "10.0.0.5"),
		),
	)
	fmt.Println("   -> Created root span: POST /api/orders")

	// Step 1: Validate input
	_, validateSpan := client.StartSpan(ctx, "order.validate")
	time.Sleep(1 * time.Millisecond)
	validateSpan.AddEvent("validation.passed", attribute.Int("rules", 8))
	validateSpan.SetStatus(codes.Ok, "")
	validateSpan.End()
	fmt.Println("   -> Validated input (child span)")

	// Step 2: Check inventory via external service
	ctx2, inventorySpan := client.StartSpan(ctx, "inventory.check",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", "/inventory.InventoryService/CheckStock"),
		),
	)
	time.Sleep(5 * time.Millisecond)

	// Inject context for outgoing gRPC call
	outgoingHeaders := make(tracing.HeaderCarrier)
	tracing.InjectContext(ctx2, outgoingHeaders)

	inventorySpan.AddEvent("stock.available", attribute.Int("quantity", 42))
	inventorySpan.SetStatus(codes.Ok, "")
	inventorySpan.End()
	fmt.Println("   -> Checked inventory via gRPC (client span, trace context propagated)")

	// Step 3: Persist to database
	_, dbSpan := client.StartSpan(ctx, "db.insert",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.system", "postgres"),
			attribute.String("db.name", "orders"),
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.statement", "INSERT INTO orders (customer_id, product_id, qty) VALUES ($1, $2, $3)"),
		),
	)
	time.Sleep(2 * time.Millisecond)
	dbSpan.SetStatus(codes.Ok, "")
	dbSpan.End()
	fmt.Println("   -> Inserted into database (client span)")

	// Step 4: Send event to message queue
	_, mqSpan := client.StartSpan(ctx, "kafka.publish",
		tracing.WithSpanKind(tracing.SpanKindProducer),
		tracing.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", "order-events"),
			attribute.String("messaging.operation", "publish"),
		),
	)
	time.Sleep(1 * time.Millisecond)
	mqSpan.AddEvent("message.sent",
		attribute.String("messaging.message_id", "msg-abc-123"),
	)
	mqSpan.SetStatus(codes.Ok, "")
	mqSpan.End()
	fmt.Println("   -> Published event to Kafka (producer span)")

	// Complete root span
	rootSpan.SetAttributes(attribute.Int("http.status_code", 201))
	rootSpan.SetStatus(codes.Ok, "")
	rootSpan.End()
	fmt.Println("   -> Completed with HTTP 201")

	// Inject context into response headers
	responseHeaders := make(tracing.HeaderCarrier)
	tracing.InjectContext(ctx, responseHeaders)

	fmt.Println()
	fmt.Println("   Resulting trace:")
	fmt.Println("   POST /api/orders (Server)")
	fmt.Println("     |-- order.validate (Internal)")
	fmt.Println("     |-- inventory.check (Client -> gRPC)")
	fmt.Println("     |-- db.insert (Client -> Postgres)")
	fmt.Println("     +-- kafka.publish (Producer -> Kafka)")
	fmt.Println()
	fmt.Println("   View this trace in Tempo/Jaeger UI using the trace ID")
	fmt.Println("   from the X-Trace-ID response header.")
}
