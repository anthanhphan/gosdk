// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/anthanhphan/gosdk/metrics"
)

func main() {
	fmt.Println("=== metrics Example ===")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Basic Counter
	exampleCounter(ctx)

	// Example 2: Gauge Operations
	exampleGauge(ctx)

	// Example 3: Histogram
	exampleHistogram(ctx)

	// Example 4: Duration Measurement
	exampleDuration(ctx)

	// Example 5: Client Options
	exampleOptions(ctx)

	// Example 6: NoopClient
	exampleNoop(ctx)

	// Example 7: HTTP Handler
	exampleHTTPHandler()
}

// exampleCounter demonstrates counter metric operations.
func exampleCounter(ctx context.Context) {
	fmt.Println("1. Counter Metrics:")

	client := metrics.NewClient("myapp",
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)

	// Inc increments a counter by 1
	client.Inc(ctx, "requests_total", "method", "GET", "status", "200")
	client.Inc(ctx, "requests_total", "method", "POST", "status", "201")
	client.Inc(ctx, "requests_total", "method", "GET", "status", "200")

	// Add increments a counter by a specific value
	client.Add(ctx, "bytes_sent_total", 1024, "endpoint", "/upload")
	client.Add(ctx, "bytes_sent_total", 2048, "endpoint", "/upload")

	fmt.Println("   - Incremented requests_total counter with method/status tags")
	fmt.Println("   - Added 3072 bytes to bytes_sent_total counter")
	fmt.Println()
}

// exampleGauge demonstrates gauge metric operations.
func exampleGauge(ctx context.Context) {
	fmt.Println("2. Gauge Metrics:")

	client := metrics.NewClient("myapp",
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)

	// SetGauge sets a gauge to a specific value
	client.SetGauge(ctx, "active_connections", 42, "service", "api")
	fmt.Println("   - Set active_connections = 42")

	// GaugeInc increments a gauge by 1
	client.GaugeInc(ctx, "active_requests", "handler", "GetUser")
	client.GaugeInc(ctx, "active_requests", "handler", "GetUser")
	fmt.Println("   - Incremented active_requests to 2")

	// GaugeDec decrements a gauge by 1
	client.GaugeDec(ctx, "active_requests", "handler", "GetUser")
	fmt.Println("   - Decremented active_requests to 1")

	// Real-world pattern: track in-flight requests
	client.GaugeInc(ctx, "inflight", "endpoint", "/users")
	// ... handle request ...
	client.GaugeDec(ctx, "inflight", "endpoint", "/users")
	fmt.Println("   - Tracked in-flight request lifecycle")
	fmt.Println()
}

// exampleHistogram demonstrates histogram metric operations.
func exampleHistogram(ctx context.Context) {
	fmt.Println("3. Histogram Metrics:")

	client := metrics.NewClient("myapp",
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)

	// Record some request size observations
	sizes := []float64{256, 512, 1024, 2048, 4096}
	for _, size := range sizes {
		client.Histogram(ctx, "request_size_bytes", size, "endpoint", "/upload")
	}
	fmt.Printf("   - Recorded %d request size observations\n", len(sizes))

	// Record batch processing sizes
	client.Histogram(ctx, "batch_size", 100, "job", "import")
	client.Histogram(ctx, "batch_size", 250, "job", "import")
	client.Histogram(ctx, "batch_size", 50, "job", "export")
	fmt.Println("   - Recorded batch size observations with job tags")
	fmt.Println()
}

// exampleDuration demonstrates duration measurement.
func exampleDuration(ctx context.Context) {
	fmt.Println("4. Duration Measurement:")

	client := metrics.NewClient("myapp",
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)

	// Simulate measuring request duration
	start := time.Now()
	time.Sleep(time.Duration(rand.IntN(50)+10) * time.Millisecond) // simulate work
	client.Duration(ctx, "request_duration_seconds", start, "endpoint", "/users", "method", "GET")
	fmt.Printf("   - Recorded request duration: %v\n", time.Since(start).Round(time.Millisecond))

	// Measure database query time
	dbStart := time.Now()
	time.Sleep(time.Duration(rand.IntN(20)+5) * time.Millisecond) // simulate query
	client.Duration(ctx, "db_query_duration_seconds", dbStart, "query", "SELECT", "table", "users")
	fmt.Printf("   - Recorded DB query duration: %v\n", time.Since(dbStart).Round(time.Millisecond))
	fmt.Println()
}

// exampleOptions demonstrates client configuration options.
func exampleOptions(ctx context.Context) {
	fmt.Println("5. Client Options:")

	// WithSubsystem groups related metrics
	httpClient := metrics.NewClient("myapp",
		metrics.WithSubsystem("http"),
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)
	httpClient.Inc(ctx, "requests_total", "method", "GET")
	fmt.Println("   - WithSubsystem: metrics named myapp_http_requests_total")

	// WithConstLabels adds labels to every metric
	client := metrics.NewClient("myapp",
		metrics.WithConstLabels(map[string]string{
			"env":    "production",
			"region": "us-east-1",
		}),
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)
	client.Inc(ctx, "requests_total", "method", "GET")
	fmt.Println("   - WithConstLabels: env=production, region=us-east-1 on all metrics")

	// WithBuckets customizes histogram bucket boundaries
	latencyClient := metrics.NewClient("myapp",
		metrics.WithBuckets([]float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0}),
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)
	latencyClient.Histogram(ctx, "latency_seconds", 0.035, "op", "read")
	fmt.Println("   - WithBuckets: custom histogram buckets [10ms, 50ms, 100ms, 500ms, 1s, 5s]")

	// Show default buckets
	fmt.Printf("   - Default buckets: %v\n", metrics.DefaultDurationBuckets())
	fmt.Println()
}

// exampleNoop demonstrates the NoopClient for testing.
func exampleNoop(ctx context.Context) {
	fmt.Println("6. NoopClient (for testing):")

	client := metrics.NewNoopClient()

	// All operations are silently discarded
	client.Inc(ctx, "requests_total", "method", "GET")
	client.Add(ctx, "bytes_sent", 1024, "endpoint", "/upload")
	client.SetGauge(ctx, "active_connections", 42, "service", "api")
	client.GaugeInc(ctx, "active_requests", "handler", "GetUser")
	client.GaugeDec(ctx, "active_requests", "handler", "GetUser")
	client.Histogram(ctx, "request_size", 256, "endpoint", "/upload")
	client.Duration(ctx, "request_duration", time.Now(), "endpoint", "/users")

	fmt.Println("   - All metric operations silently discarded")

	// Use as a default when metrics are optional
	var metricsClient metrics.Client = metrics.NewNoopClient()
	enableMetrics := false
	if enableMetrics {
		metricsClient = metrics.NewClient("myapp")
	}
	metricsClient.Inc(ctx, "requests_total")
	fmt.Println("   - Safe to use as a default fallback")

	if err := client.Close(); err != nil {
		log.Fatal("Close error:", err)
	}
	fmt.Println("   - Close is a no-op")
	fmt.Println()
}

// exampleHTTPHandler demonstrates exposing metrics via HTTP.
func exampleHTTPHandler() {
	fmt.Println("7. HTTP Handler:")

	client := metrics.NewClient("myapp",
		metrics.WithSubsystem("http"),
		metrics.WithoutGoCollector(),
		metrics.WithoutProcessCollector(),
	)

	// Mount the metrics handler at /metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", client.Handler())

	fmt.Println("   // Example server setup:")
	fmt.Println("   // client := metrics.NewClient(\"myapp\")")
	fmt.Println("   // http.Handle(\"/metrics\", client.Handler())")
	fmt.Println("   // http.ListenAndServe(\":8080\", nil)")
	fmt.Println()
	fmt.Println("   - Metrics handler ready for Prometheus scraping")
	fmt.Println("   - Access at http://localhost:8080/metrics")
}
