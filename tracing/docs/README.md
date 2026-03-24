# Tracing Package

A unified interface for distributed tracing with OpenTelemetry as the default backend. Supports exporting traces to **Tempo**, **Jaeger**, or any OTLP-compatible collector.

## Features

- **OpenTelemetry Backend** — Production-ready distributed tracing with OTel SDK
- **OTLP Export** — Exports traces via OTLP gRPC to Tempo, Jaeger, Grafana Alloy, etc.
- **Simplified API** — Intuitive methods for creating and managing spans
- **Context Propagation** — W3C TraceContext + Baggage propagation
- **Custom Options** — Configurable endpoint, sampling, service metadata
- **NoopClient** — Zero-cost no-op implementation for testing
- **Thread-Safe** — All operations are safe for concurrent use
- **Orianna Integration** — gRPC interceptor and REST middleware included

## Installation

```bash
go get github.com/anthanhphan/gosdk/tracing
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/anthanhphan/gosdk/tracing"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

func main() {
    // Create tracing client connected to an OTLP collector (Tempo, Jaeger, etc.)
    client, err := tracing.NewClient("order-service",
        tracing.WithEndpoint("otel-collector:4317"),
        tracing.WithInsecure(),
        tracing.WithEnvironment("production"),
        tracing.WithServiceVersion("v1.2.0"),
        tracing.WithSamplingRate(0.1), // Sample 10% of traces
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Shutdown(context.Background())

    ctx := context.Background()

    // Start a root span for an operation
    ctx, span := client.StartSpan(ctx, "process-order",
        tracing.WithSpanKind(tracing.SpanKindServer),
        tracing.WithAttributes(
            attribute.String("order.id", "12345"),
            attribute.String("customer.id", "cust-999"),
        ),
    )
    defer span.End()

    // Trace a child operation (child span is created automatically via context)
    if err := processPayment(ctx, client, "12345", 150.00); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return
    }

    span.SetStatus(codes.Ok, "order processed")
}

// processPayment creates a child span under the parent span in ctx
func processPayment(ctx context.Context, client tracing.Client, orderID string, amount float64) error {
    ctx, span := client.StartSpan(ctx, "payment.process",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("payment.order_id", orderID),
            attribute.Float64("payment.amount", amount),
            attribute.String("payment.currency", "VND"),
        ),
    )
    defer span.End()

    // Simulate payment gateway call
    span.AddEvent("payment.submitted", attribute.String("gateway", "stripe"))

    // Query database
    if err := queryDB(ctx, client); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.AddEvent("payment.confirmed")
    span.SetStatus(codes.Ok, "")
    return nil
}

// queryDB creates a client span for a database query
func queryDB(ctx context.Context, client tracing.Client) error {
    _, span := client.StartSpan(ctx, "db.query",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("db.system", "postgres"),
            attribute.String("db.name", "payments"),
            attribute.String("db.statement", "SELECT * FROM payments WHERE order_id = $1"),
        ),
    )
    defer span.End()

    // Execute query...
    span.SetStatus(codes.Ok, "")
    return nil
}
```

The above produces a trace with 3 nested spans:

```
process-order (Server)
  └── payment.process (Client)
        └── db.query (Client)
```

## Usage

### Creating Spans

Use `StartSpan` to create a span. The returned context carries the span, so any child
spans created from it will be nested under the parent:

```go
ctx, span := client.StartSpan(ctx, "operation-name",
    tracing.WithSpanKind(tracing.SpanKindServer),
    tracing.WithAttributes(
        attribute.String("key", "value"),
        attribute.Int("count", 42),
    ),
)
defer span.End() // Always defer End() to ensure the span is completed
```

### Span Kinds

| Kind | When to use |
|------|-------------|
| `SpanKindServer` | Handling incoming requests (HTTP, gRPC) |
| `SpanKindClient` | Making outgoing requests (HTTP, gRPC, DB) |
| `SpanKindProducer` | Sending messages to a queue (Kafka, RabbitMQ) |
| `SpanKindConsumer` | Receiving messages from a queue |
| `SpanKindInternal` | Internal operations (default) |

### Recording Errors

Always record errors on the span AND set the span status:

```go
result, err := doSomething(ctx)
if err != nil {
    span.RecordError(err)                          // Adds error as a span event
    span.SetStatus(codes.Error, err.Error())       // Marks span as failed
    return err
}
span.SetStatus(codes.Ok, "")
```

### Adding Events

Events are timestamped annotations within a span, useful for marking key moments:

```go
span.AddEvent("cache.hit", attribute.String("key", "user:123"))
span.AddEvent("retry.attempt", attribute.Int("attempt", 2))
span.AddEvent("validation.passed")
```

### Setting Attributes After Creation

Add dynamic attributes to a span after it's been started:

```go
ctx, span := client.StartSpan(ctx, "http.request")
defer span.End()

// ... process request ...

span.SetAttributes(
    attribute.Int("http.status_code", 200),
    attribute.Int64("http.response_content_length", 1234),
)
```

### Extracting Trace ID for Log Correlation

Use `TraceIDFromContext` to correlate logs with traces:

```go
traceID := tracing.TraceIDFromContext(ctx)
spanID  := tracing.SpanIDFromContext(ctx)

logger.Infow("processing request",
    "trace_id", traceID,
    "span_id", spanID,
)
```

## Configuration Options

### Option Reference

| Option | Description | Default |
|--------|-------------|---------|
| `WithServiceVersion(ver)` | Sets `service.version` resource attribute | (none) |
| `WithEnvironment(env)` | Sets `deployment.environment` attribute | (none) |
| `WithEndpoint(url)` | OTLP collector endpoint | `localhost:4317` |
| `WithInsecure()` | Disable TLS for OTLP connection | TLS enabled |
| `WithSampler(sampler)` | Custom OTel sampler | AlwaysSample |
| `WithSamplingRate(rate)` | TraceIDRatioBased sampling (0.0 - 1.0) | 1.0 |
| `WithHeaders(map)` | Additional gRPC headers for OTLP exporter | (none) |

### Sampling Strategies

```go
// Always sample (default) — for development/staging
client, _ := tracing.NewClient("my-service")

// Sample 10% — for high-traffic production
client, _ := tracing.NewClient("my-service",
    tracing.WithSamplingRate(0.1),
)

// Custom sampler — respect parent decision, fallback to 50%
client, _ := tracing.NewClient("my-service",
    tracing.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5))),
)
```

### Authenticated OTLP Collector

```go
client, _ := tracing.NewClient("my-service",
    tracing.WithEndpoint("otel.grafana.net:443"),
    tracing.WithHeaders(map[string]string{
        "Authorization": "Basic <api-key>",
    }),
)
```

## Orianna Integration

### Full REST Server Example

```go
package main

import (
    "context"
    "log"

    "github.com/anthanhphan/gosdk/metrics"
    "github.com/anthanhphan/gosdk/orianna"
    "github.com/anthanhphan/gosdk/tracing"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

func main() {
    // Initialize tracing
    tracingClient, err := tracing.NewClient("user-service",
        tracing.WithEndpoint("tempo:4317"),
        tracing.WithInsecure(),
        tracing.WithEnvironment("production"),
        tracing.WithServiceVersion("v2.1.0"),
        tracing.WithSamplingRate(0.1),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer tracingClient.Shutdown(context.Background())

    // Initialize metrics (optional, can be used alongside tracing)
    metricsClient := metrics.NewClient("user_service")

    // Create REST server with tracing + metrics
    config := &orianna.Config{
        ServiceName: "user-service",
        Port:        8080,
    }

    srv, err := orianna.NewHttpServer(config,
        orianna.WithTracing(tracingClient),   // <-- enables tracing middleware
        orianna.WithMetrics(metricsClient),   // <-- enables metrics middleware
    )
    if err != nil {
        log.Fatal(err)
    }

    // Register routes
    srv.RegisterRoutes(
        orianna.NewRoute(orianna.GET, "/users/:id", getUserHandler(tracingClient)),
    )

    srv.Start()
}

// getUserHandler shows manual span creation inside a handler
func getUserHandler(tc tracing.Client) orianna.Handler {
    return func(ctx orianna.Context) error {
        userID := ctx.Params("id")

        // The tracing middleware already created a root span
        // for "GET /users/:id". Create a child span for DB query.
        goCtx, span := tc.StartSpan(ctx.Context(), "db.find_user",
            tracing.WithSpanKind(tracing.SpanKindClient),
            tracing.WithAttributes(
                attribute.String("db.system", "postgres"),
                attribute.String("db.statement", "SELECT * FROM users WHERE id = $1"),
                attribute.String("db.params.id", userID),
            ),
        )
        defer span.End()

        // Simulate DB query...
        user := map[string]string{"id": userID, "name": "Phan An Thanh"}

        span.SetStatus(codes.Ok, "")

        // Log with trace correlation
        // traceID is automatically set in ctx.Locals("trace_id") by middleware
        _ = goCtx // goCtx would be passed to downstream services

        return ctx.OK(user)
    }
}
```

When a request comes in, the tracing middleware automatically:
1. Extracts `traceparent` header from the incoming request
2. Creates a server span: `GET /users/:id`
3. Sets attributes: `http.method`, `http.route`, `http.url`, `http.scheme`, `net.peer.ip`
4. Stores `trace_id` in `ctx.Locals("trace_id")` for log correlation
5. Sets `X-Trace-ID` response header
6. Records `http.status_code` after the handler completes
7. Injects `traceparent` into response headers for downstream tracing

### Full gRPC Server Example

```go
package main

import (
    "context"
    "log"

    "github.com/anthanhphan/gosdk/metrics"
    "github.com/anthanhphan/gosdk/orianna"
    "github.com/anthanhphan/gosdk/tracing"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

func main() {
    // Initialize tracing
    tracingClient, err := tracing.NewClient("payment-service",
        tracing.WithEndpoint("tempo:4317"),
        tracing.WithInsecure(),
        tracing.WithEnvironment("production"),
        tracing.WithSamplingRate(0.5),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer tracingClient.Shutdown(context.Background())

    // Create gRPC server with tracing interceptors
    config := &orianna.GrpcConfig{
        Port: 50051,
    }

    metricsClient := metrics.NewClient("payment_service")

    srv, err := orianna.NewGrpcServer(config,
        // Tracing interceptors
        orianna.WithGrpcGlobalUnaryInterceptor(
            orianna.GrpcTracingInterceptor(tracingClient),
        ),
        orianna.WithGrpcGlobalStreamInterceptor(
            orianna.GrpcStreamTracingInterceptor(tracingClient),
        ),
        // Metrics interceptors (alongside tracing)
        orianna.WithGrpcGlobalUnaryInterceptor(
            orianna.GrpcMetricsInterceptor(metricsClient, "payment"),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Register gRPC services...
    srv.Start()
}

// In your gRPC service handler, create child spans:
func (s *paymentService) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
    // The tracing interceptor already created a root span
    // for "/payment.PaymentService/ProcessPayment"

    // Create child span for business logic
    ctx, span := s.tracingClient.StartSpan(ctx, "payment.validate",
        tracing.WithSpanKind(tracing.SpanKindInternal),
        tracing.WithAttributes(
            attribute.String("payment.method", req.Method),
            attribute.Float64("payment.amount", req.Amount),
        ),
    )
    defer span.End()

    // Validate payment...
    if req.Amount <= 0 {
        span.SetStatus(codes.Error, "invalid amount")
        return nil, status.Error(grpccodes.InvalidArgument, "amount must be positive")
    }

    // Create another child span for external call
    ctx, dbSpan := s.tracingClient.StartSpan(ctx, "db.insert_payment",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("db.system", "postgres"),
            attribute.String("db.operation", "INSERT"),
        ),
    )
    // ... insert into DB ...
    dbSpan.SetStatus(codes.Ok, "")
    dbSpan.End()

    span.SetStatus(codes.Ok, "")
    return &pb.PaymentResponse{TransactionID: "txn-123"}, nil
}
```

The gRPC interceptor automatically:
1. Extracts `traceparent` from incoming gRPC metadata
2. Creates a server span: `/payment.PaymentService/ProcessPayment`
3. Sets attributes: `rpc.system=grpc`, `rpc.method`, `rpc.grpc.client_identity`
4. Records `rpc.grpc.status_code` after handler completes
5. Records errors and sets span status on failures

### Cross-Service Trace Propagation

When Service A calls Service B, the trace context propagates automatically:

```go
// Service A (REST) → calls Service B (gRPC)
func callServiceB(ctx context.Context, tc tracing.Client) error {
    ctx, span := tc.StartSpan(ctx, "grpc.call.service_b",
        tracing.WithSpanKind(tracing.SpanKindClient),
    )
    defer span.End()

    // Inject trace context into outgoing gRPC metadata
    md := metadata.New(nil)
    carrier := tracing.HeaderCarrier{}
    tracing.InjectContext(ctx, carrier)
    for k, v := range carrier {
        md.Set(k, v)
    }
    ctx = metadata.NewOutgoingContext(ctx, md)

    // Make gRPC call — Service B's tracing interceptor will extract the context
    resp, err := grpcClient.SomeMethod(ctx, req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "")
    return nil
}
```

This produces a distributed trace across services:

```
[Service A] GET /orders/:id (Server)
  └── grpc.call.service_b (Client)
        └── [Service B] /payment.PaymentService/ProcessPayment (Server)
              └── db.insert_payment (Client)
```

## API Reference

### Client Interface

```go
type Client interface {
    StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
    Shutdown(ctx context.Context) error
    Tracer() trace.Tracer
}
```

### Span Interface

```go
type Span interface {
    End()
    SetAttributes(attrs ...attribute.KeyValue)
    SetStatus(code codes.Code, description string)
    RecordError(err error)
    AddEvent(name string, attrs ...attribute.KeyValue)
    SpanContext() trace.SpanContext
}
```

### Constructors

| Function | Description |
|----------|-------------|
| `NewClient(name, opts...)` | Creates OTel client with OTLP gRPC exporter |
| `NewNoopClient()` | Creates no-op client (all operations discarded) |

### Propagation Helpers

| Function | Description |
|----------|-------------|
| `TraceIDFromContext(ctx)` | Extracts trace ID string from context |
| `SpanIDFromContext(ctx)` | Extracts span ID string from context |
| `IsSampled(ctx)` | Returns true if current span is being sampled |
| `ExtractContext(ctx, carrier)` | Extracts trace context from carrier into Go context |
| `InjectContext(ctx, carrier)` | Injects trace context from Go context into carrier |

## NoopClient

The `NoopClient` silently discards all tracing operations. Use it for:

- Testing environments where tracing is not needed
- Disabling tracing without changing application code
- Providing a safe default when tracing is optional

```go
// Use in tests
client := tracing.NewNoopClient()

// Use as a conditional default
var tc tracing.Client = tracing.NewNoopClient()
if enableTracing {
    tc, _ = tracing.NewClient("my-service",
        tracing.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
    )
}
```

## Concurrency

The tracing package is fully thread-safe:

- `Client` methods are safe for concurrent use from multiple goroutines
- Span operations (`SetAttributes`, `AddEvent`, `RecordError`, `End`) are thread-safe
- Multiple goroutines can share a single `Client` instance

## License

Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>
