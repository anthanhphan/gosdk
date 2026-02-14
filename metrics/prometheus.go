// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ============================================================================
// Prometheus Client
// ============================================================================

// prometheusClient is the Prometheus-backed implementation of the Client interface.
// It manages counters, gauges, and histograms with thread-safe access.
type prometheusClient struct {
	registerer  prometheus.Registerer
	gatherer    prometheus.Gatherer
	namespace   string
	subsystem   string
	constLabels prometheus.Labels
	buckets     []float64

	counterMu   sync.RWMutex
	counters    map[string]*prometheus.CounterVec
	histogramMu sync.RWMutex
	histograms  map[string]*prometheus.HistogramVec
	gaugeMu     sync.RWMutex
	gauges      map[string]*prometheus.GaugeVec
}

// NewClient creates a new Prometheus metrics client with its own isolated registry.
// The namespace is used as a prefix for all metric names to avoid naming collisions.
//
// Input:
//   - namespace: Prefix for all metric names (e.g., "myapp" results in "myapp_requests_total")
//   - opts: Optional configuration options (buckets, const labels, subsystem, etc.)
//
// Output:
//   - Client: A new metrics client ready for use
//
// Example:
//
//	client := metrics.NewClient("myapp")
//	client.Inc(ctx, "requests_total", "method", "GET")
//	client.Duration(ctx, "request_duration", start, "endpoint", "/users")
//
//	// With options:
//	client := metrics.NewClient("myapp",
//	    metrics.WithSubsystem("http"),
//	    metrics.WithBuckets([]float64{0.01, 0.1, 0.5, 1, 5}),
//	)
func NewClient(namespace string, opts ...Option) Client {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	registry := prometheus.NewRegistry()

	// Register default collectors based on options
	if options.enableGoCollector {
		registry.MustRegister(collectors.NewGoCollector())
	}
	if options.enableProcessCollector {
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return &prometheusClient{
		registerer:  registry,
		gatherer:    registry,
		namespace:   namespace,
		subsystem:   options.subsystem,
		constLabels: options.constLabels,
		buckets:     options.buckets,
		counters:    make(map[string]*prometheus.CounterVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
		gauges:      make(map[string]*prometheus.GaugeVec),
	}
}

// NewClientWithRegisterer creates a new Prometheus metrics client with a custom registerer.
// Use this when you need to register metrics with a specific prometheus.Registerer instance.
//
// IMPORTANT: The registerer should also implement prometheus.Gatherer (e.g., *prometheus.Registry)
// to ensure Handler() returns the correct metrics. If it doesn't, a new registry will be created.
//
// Input:
//   - namespace: Prefix for all metric names
//   - registerer: Custom Prometheus registerer (e.g., for isolated registries)
//   - opts: Optional configuration options
//
// Output:
//   - Client: A new metrics client with custom registerer
//
// Example:
//
//	registry := prometheus.NewRegistry()
//	client := metrics.NewClientWithRegisterer("myapp", registry)
func NewClientWithRegisterer(namespace string, registerer prometheus.Registerer, opts ...Option) Client {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Determine gatherer: if registerer is also a Gatherer, use it.
	// Otherwise, create a wrapping registry so Handler() works correctly.
	var gatherer prometheus.Gatherer
	if g, ok := registerer.(prometheus.Gatherer); ok {
		gatherer = g
	} else {
		// Fallback: create a new registry to serve as gatherer
		registry := prometheus.NewRegistry()
		if options.enableGoCollector {
			registry.MustRegister(collectors.NewGoCollector())
		}
		if options.enableProcessCollector {
			registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		}
		gatherer = registry
	}

	return &prometheusClient{
		registerer:  registerer,
		gatherer:    gatherer,
		namespace:   namespace,
		subsystem:   options.subsystem,
		constLabels: options.constLabels,
		buckets:     options.buckets,
		counters:    make(map[string]*prometheus.CounterVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
		gauges:      make(map[string]*prometheus.GaugeVec),
	}
}

// NewClientWithRegistry creates a new Prometheus metrics client with a custom registry.
// This is useful for testing when you want to isolate metrics from the global registry.
//
// Input:
//   - namespace: Prefix for all metric names
//   - registry: Custom Prometheus registry for isolated metric collection
//
// Output:
//   - Client: A new metrics client with isolated registry
//
// Example:
//
//	registry := prometheus.NewRegistry()
//	client := metrics.NewClientWithRegistry("myapp", registry)
//	// Use in tests to avoid polluting global metrics
func NewClientWithRegistry(namespace string, registry *prometheus.Registry, opts ...Option) Client {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Register default collectors based on options
	if options.enableGoCollector {
		registry.MustRegister(collectors.NewGoCollector())
	}
	if options.enableProcessCollector {
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return &prometheusClient{
		registerer:  registry,
		gatherer:    registry,
		namespace:   namespace,
		subsystem:   options.subsystem,
		constLabels: options.constLabels,
		buckets:     options.buckets,
		counters:    make(map[string]*prometheus.CounterVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
		gauges:      make(map[string]*prometheus.GaugeVec),
	}
}

// ============================================================================
// Counter Operations
// ============================================================================

// Inc increments a counter metric by 1.
// Counters are monotonically increasing values, typically used for request counts.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the counter metric
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	client.Inc(ctx, "requests_total", "method", "GET", "status", "200")
func (c *prometheusClient) Inc(ctx context.Context, name string, tags ...string) {
	c.Add(ctx, name, 1, tags...)
}

// Add adds the given value to a counter metric.
// Use this when you need to increment by more than 1.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the counter metric
//   - value: Amount to add to the counter (must be positive)
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	client.Add(ctx, "bytes_sent", 1024, "endpoint", "/upload")
func (c *prometheusClient) Add(_ context.Context, name string, value int64, tags ...string) {
	labelNames := extractLabelNames(tags)

	c.counterMu.RLock()
	counter, exists := c.counters[name]
	c.counterMu.RUnlock()

	if !exists {
		c.counterMu.Lock()
		// Double-check after acquiring write lock
		if counter, exists = c.counters[name]; !exists {
			counter = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace:   c.namespace,
					Subsystem:   c.subsystem,
					Name:        name,
					Help:        name,
					ConstLabels: c.constLabels,
				},
				labelNames,
			)
			c.registerer.MustRegister(counter)
			c.counters[name] = counter
		}
		c.counterMu.Unlock()
	}

	labelValues := extractLabelValues(tags)
	counter.WithLabelValues(labelValues...).Add(float64(value))
}

// ============================================================================
// Gauge Operations
// ============================================================================

// SetGauge sets a gauge metric to a specific value.
// Gauges represent values that can go up or down, like active connections or memory usage.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the gauge metric
//   - value: The value to set the gauge to
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	client.SetGauge(ctx, "active_connections", 42, "service", "api")
//	client.SetGauge(ctx, "memory_usage_bytes", 1073741824, "pod", "web-1")
func (c *prometheusClient) SetGauge(_ context.Context, name string, value float64, tags ...string) {
	gauge := c.getOrCreateGauge(name, tags)
	labelValues := extractLabelValues(tags)
	gauge.WithLabelValues(labelValues...).Set(value)
}

// GaugeInc increments a gauge metric by 1.
// Use this for tracking values that increase and decrease, like active requests.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the gauge metric
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	// At request start
//	client.GaugeInc(ctx, "active_requests", "handler", "GetUser")
func (c *prometheusClient) GaugeInc(_ context.Context, name string, tags ...string) {
	gauge := c.getOrCreateGauge(name, tags)
	labelValues := extractLabelValues(tags)
	gauge.WithLabelValues(labelValues...).Inc()
}

// GaugeDec decrements a gauge metric by 1.
// Use this for tracking values that increase and decrease, like active requests.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the gauge metric
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	// At request end
//	client.GaugeDec(ctx, "active_requests", "handler", "GetUser")
func (c *prometheusClient) GaugeDec(_ context.Context, name string, tags ...string) {
	gauge := c.getOrCreateGauge(name, tags)
	labelValues := extractLabelValues(tags)
	gauge.WithLabelValues(labelValues...).Dec()
}

// getOrCreateGauge retrieves an existing gauge or creates a new one if it doesn't exist.
// This method is thread-safe and uses double-checked locking for performance.
func (c *prometheusClient) getOrCreateGauge(name string, tags []string) *prometheus.GaugeVec {
	labelNames := extractLabelNames(tags)

	c.gaugeMu.RLock()
	gauge, exists := c.gauges[name]
	c.gaugeMu.RUnlock()

	if !exists {
		c.gaugeMu.Lock()
		if gauge, exists = c.gauges[name]; !exists {
			gauge = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   c.namespace,
					Subsystem:   c.subsystem,
					Name:        name,
					Help:        name,
					ConstLabels: c.constLabels,
				},
				labelNames,
			)
			c.registerer.MustRegister(gauge)
			c.gauges[name] = gauge
		}
		c.gaugeMu.Unlock()
	}

	return gauge
}

// ============================================================================
// Histogram Operations
// ============================================================================

// Histogram records a value observation in a histogram metric.
// Histograms are useful for measuring distributions like request sizes or response times.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the histogram metric
//   - value: The observed value to record
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	client.Histogram(ctx, "request_size_bytes", 1024, "endpoint", "/upload")
//	client.Histogram(ctx, "batch_size", 100, "job", "import")
func (c *prometheusClient) Histogram(_ context.Context, name string, value float64, tags ...string) {
	histogram := c.getOrCreateHistogram(name, tags)
	labelValues := extractLabelValues(tags)
	histogram.WithLabelValues(labelValues...).Observe(value)
}

// Duration records the duration since start time in seconds as a histogram observation.
// This is a convenience method for timing operations using time.Now() as the start.
// Uses DefaultDurationBuckets for histogram bucket boundaries.
//
// Input:
//   - ctx: Context for the operation (reserved for future use)
//   - name: Name of the histogram metric
//   - start: Start time captured before the operation (typically via time.Now())
//   - tags: Alternating key-value pairs for metric labels
//
// Example:
//
//	start := time.Now()
//	// ... perform operation ...
//	client.Duration(ctx, "request_duration_seconds", start, "endpoint", "/users")
func (c *prometheusClient) Duration(_ context.Context, name string, start time.Time, tags ...string) {
	elapsed := float64(time.Since(start)) / float64(time.Second)
	histogram := c.getOrCreateHistogram(name, tags)
	labelValues := extractLabelValues(tags)
	histogram.WithLabelValues(labelValues...).Observe(elapsed)
}

// getOrCreateHistogram retrieves an existing histogram or creates a new one if it doesn't exist.
// This method is thread-safe and uses double-checked locking for performance.
func (c *prometheusClient) getOrCreateHistogram(name string, tags []string) *prometheus.HistogramVec {
	labelNames := extractLabelNames(tags)

	c.histogramMu.RLock()
	histogram, exists := c.histograms[name]
	c.histogramMu.RUnlock()

	if !exists {
		c.histogramMu.Lock()
		if histogram, exists = c.histograms[name]; !exists {
			buckets := c.buckets
			if len(buckets) == 0 {
				buckets = DefaultDurationBuckets()
			}
			histogram = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace:   c.namespace,
					Subsystem:   c.subsystem,
					Name:        name,
					Help:        name,
					Buckets:     buckets,
					ConstLabels: c.constLabels,
				},
				labelNames,
			)
			c.registerer.MustRegister(histogram)
			c.histograms[name] = histogram
		}
		c.histogramMu.Unlock()
	}

	return histogram
}

// ============================================================================
// HTTP Handler
// ============================================================================

// Handler returns an HTTP handler for exposing metrics in Prometheus format.
// Mount this handler at a path like "/metrics" to enable Prometheus scraping.
//
// This handler uses the instance's own gatherer, ensuring only metrics registered
// with this client are exposed. This is critical when using custom registries.
//
// Output:
//   - http.Handler: Handler that serves metrics in Prometheus exposition format
//
// Example:
//
//	client := metrics.NewClient("myapp")
//	http.Handle("/metrics", client.Handler())
//	http.ListenAndServe(":8080", nil)
func (c *prometheusClient) Handler() http.Handler {
	return promhttp.HandlerFor(c.gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Close performs cleanup for the Prometheus client.
// For Prometheus, this is a no-op since metrics are scraped by the server.
// This method exists to satisfy the Client interface for backends that need cleanup.
func (*prometheusClient) Close() error {
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// extractLabelNames extracts the label names (keys) from alternating key-value tag pairs.
// If the number of tags is odd, it appends "unknown" to make it even.
func extractLabelNames(tags []string) []string {
	if len(tags)%2 != 0 {
		tags = append(tags, "unknown")
	}
	names := make([]string, 0, len(tags)/2)
	for i := 0; i < len(tags); i += 2 {
		names = append(names, tags[i])
	}
	return names
}

// extractLabelValues extracts the label values from alternating key-value tag pairs.
// If the number of tags is odd, it appends "unknown" to make it even.
func extractLabelValues(tags []string) []string {
	if len(tags)%2 != 0 {
		tags = append(tags, "unknown")
	}
	values := make([]string, 0, len(tags)/2)
	for i := 1; i < len(tags); i += 2 {
		values = append(values, tags[i])
	}
	return values
}
