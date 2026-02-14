# Metrics Package

A unified interface for application metrics collection with Prometheus as the default backend. Provides a simplified API for tracking counters, gauges, and histograms with support for labeled dimensions (tags).

## Features

- **Prometheus Backend** - Production-ready Prometheus metrics collection
- **Simplified API** - Intuitive methods for counters, gauges, and histograms
- **Tagged Metrics** - Flexible label dimensions via alternating key-value pairs
- **Duration Tracking** - Built-in convenience for timing operations
- **Custom Options** - Configurable buckets, const labels, subsystem grouping
- **Isolated Registries** - Each client uses its own registry to avoid collisions
- **NoopClient** - Zero-cost no-op implementation for testing
- **Thread-Safe** - All operations are safe for concurrent use
- **HTTP Handler** - Built-in handler for Prometheus scraping endpoint

## Installation

```bash
go get github.com/anthanhphan/gosdk/metrics
```

## Quick Start

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/anthanhphan/gosdk/metrics"
)

func main() {
    client := metrics.NewClient("myapp")
    ctx := context.Background()

    // Count requests
    client.Inc(ctx, "requests_total", "method", "GET", "status", "200")

    // Track active connections
    client.SetGauge(ctx, "active_connections", 42, "service", "api")

    // Measure request duration
    start := time.Now()
    // ... handle request ...
    client.Duration(ctx, "request_duration_seconds", start, "endpoint", "/users")

    // Expose metrics for Prometheus scraping
    http.Handle("/metrics", client.Handler())
    http.ListenAndServe(":8080", nil)
}
```

## Usage

### Counter Metrics

Counters are monotonically increasing values, used for request counts, bytes sent, etc.

```go
// Increment by 1
client.Inc(ctx, "requests_total", "method", "GET", "status", "200")

// Increment by a specific value
client.Add(ctx, "bytes_sent_total", 1024, "endpoint", "/upload")
```

### Gauge Metrics

Gauges represent values that can go up or down, like active connections or memory usage.

```go
// Set to a specific value
client.SetGauge(ctx, "active_connections", 42, "service", "api")

// Increment / decrement by 1
client.GaugeInc(ctx, "active_requests", "handler", "GetUser")
client.GaugeDec(ctx, "active_requests", "handler", "GetUser")
```

### Histogram Metrics

Histograms measure distributions of values like request sizes or response times.

```go
// Record a value observation
client.Histogram(ctx, "request_size_bytes", 1024, "endpoint", "/upload")
```

### Duration Measurement

Convenience method for timing operations using `time.Now()` as the start.

```go
start := time.Now()
// ... perform operation ...
client.Duration(ctx, "request_duration_seconds", start, "endpoint", "/users")
```

### Tags

Tags are passed as alternating key-value strings to add labeled dimensions to metrics:

```go
// Tags: method=GET, status=200
client.Inc(ctx, "requests_total", "method", "GET", "status", "200")

// Tags: endpoint=/users, method=GET
client.Duration(ctx, "request_duration_seconds", start, "endpoint", "/users", "method", "GET")
```

If an odd number of tag values is provided, `"unknown"` is appended as the last value.

## Configuration Options

### WithSubsystem

Groups related metrics under a subsystem name inserted between namespace and metric name:

```go
// Metrics will be named: myapp_http_requests_total
client := metrics.NewClient("myapp",
    metrics.WithSubsystem("http"),
)
```

### WithConstLabels

Sets constant labels applied to every metric. Use for service-level identifiers:

```go
client := metrics.NewClient("myapp",
    metrics.WithConstLabels(map[string]string{
        "env":    "production",
        "region": "us-east-1",
    }),
)
```

### WithBuckets

Sets custom histogram bucket boundaries. If not set, `DefaultDurationBuckets` are used:

```go
client := metrics.NewClient("myapp",
    metrics.WithBuckets([]float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0}),
)
```

Default buckets: `[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]`

### WithoutGoCollector / WithoutProcessCollector

Disables the Go runtime or process metrics collectors. Useful in testing or to reduce metric cardinality:

```go
client := metrics.NewClient("myapp",
    metrics.WithoutGoCollector(),
    metrics.WithoutProcessCollector(),
)
```

## API Reference

### Client Constructors

- **`NewClient(namespace string, opts ...Option) Client`** - Creates a new Prometheus client with its own isolated registry
- **`NewClientWithRegisterer(namespace string, registerer prometheus.Registerer, opts ...Option) Client`** - Creates a client with a custom Prometheus registerer
- **`NewClientWithRegistry(namespace string, registry *prometheus.Registry, opts ...Option) Client`** - Creates a client with a custom Prometheus registry (useful for testing)
- **`NewNoopClient() Client`** - Creates a no-op client where all operations are silently discarded

### Client Interface

```go
type Client interface {
    Inc(ctx context.Context, name string, tags ...string)
    Add(ctx context.Context, name string, value int64, tags ...string)
    SetGauge(ctx context.Context, name string, value float64, tags ...string)
    GaugeInc(ctx context.Context, name string, tags ...string)
    GaugeDec(ctx context.Context, name string, tags ...string)
    Histogram(ctx context.Context, name string, value float64, tags ...string)
    Duration(ctx context.Context, name string, start time.Time, tags ...string)
    Handler() http.Handler
    Close() error
}
```

| Method | Description |
|--------|-------------|
| `Inc` | Increments a counter by 1 |
| `Add` | Adds a specific value to a counter |
| `SetGauge` | Sets a gauge to a specific value |
| `GaugeInc` | Increments a gauge by 1 |
| `GaugeDec` | Decrements a gauge by 1 |
| `Histogram` | Records a value observation in a histogram |
| `Duration` | Records elapsed duration since start time as a histogram observation |
| `Handler` | Returns an HTTP handler for Prometheus metric scraping |
| `Close` | Performs cleanup (no-op for Prometheus backend) |

### Option Functions

| Option | Description |
|--------|-------------|
| `WithBuckets(buckets []float64)` | Sets custom histogram bucket boundaries |
| `WithConstLabels(labels map[string]string)` | Sets constant labels for all metrics |
| `WithSubsystem(subsystem string)` | Sets subsystem name between namespace and metric name |
| `WithoutGoCollector()` | Disables the Go runtime metrics collector |
| `WithoutProcessCollector()` | Disables the process metrics collector |

### Utility Functions

- **`DefaultDurationBuckets() []float64`** - Returns a copy of the default histogram buckets

## NoopClient

The `NoopClient` silently discards all metric operations. Use it for:

- Testing environments where metrics collection is not needed
- Disabling metrics without changing application code
- Providing a safe default when metrics are optional

```go
// Use in tests
client := metrics.NewNoopClient()

// Use as a conditional default
var metricsClient metrics.Client = metrics.NewNoopClient()
if enableMetrics {
    metricsClient = metrics.NewClient("myapp")
}
```

## HTTP Handler

Expose metrics for Prometheus scraping by mounting the handler:

```go
client := metrics.NewClient("myapp")
http.Handle("/metrics", client.Handler())
http.ListenAndServe(":8080", nil)
```

Each client exposes only its own metrics via an isolated gatherer, preventing cross-contamination when using multiple clients.

## Concurrency

The metrics package is fully thread-safe:

- Metric creation uses double-checked locking with `sync.RWMutex` for optimal read performance
- All counter, gauge, and histogram operations are concurrent-safe
- Multiple goroutines can safely share a single `Client` instance

## License

Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>
