// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ============================================================================
// Core Functionality Tests
// ============================================================================

func TestPrometheusClient(t *testing.T) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("test", registry)
	ctx := context.Background()

	t.Run("Counter_Inc", func(t *testing.T) {
		client.Inc(ctx, "requests_total", "method", "GET", "path", "/users")
		client.Inc(ctx, "requests_total", "method", "GET", "path", "/users")

		// Verify the counter was incremented
		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}
		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_requests_total" {
				found = true
				for _, m := range mf.GetMetric() {
					if m.GetCounter().GetValue() != 2 {
						t.Errorf("expected counter value 2, got %v", m.GetCounter().GetValue())
					}
				}
			}
		}
		if !found {
			t.Error("counter metric 'test_requests_total' not found")
		}
	})

	t.Run("Counter_Add", func(t *testing.T) {
		client.Add(ctx, "bytes_total", 1024, "endpoint", "/upload")

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}
		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_bytes_total" {
				found = true
				for _, m := range mf.GetMetric() {
					if m.GetCounter().GetValue() != 1024 {
						t.Errorf("expected counter value 1024, got %v", m.GetCounter().GetValue())
					}
				}
			}
		}
		if !found {
			t.Error("counter metric 'test_bytes_total' not found")
		}
	})

	t.Run("Gauge_Set", func(t *testing.T) {
		client.SetGauge(ctx, "connections", 42)

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}
		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_connections" {
				found = true
				for _, m := range mf.GetMetric() {
					if m.GetGauge().GetValue() != 42 {
						t.Errorf("expected gauge value 42, got %v", m.GetGauge().GetValue())
					}
				}
			}
		}
		if !found {
			t.Error("gauge metric 'test_connections' not found")
		}
	})

	t.Run("Gauge_IncDec", func(t *testing.T) {
		client.SetGauge(ctx, "active_requests", 10)
		client.GaugeInc(ctx, "active_requests")
		client.GaugeInc(ctx, "active_requests")
		client.GaugeDec(ctx, "active_requests")

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_active_requests" {
				for _, m := range mf.GetMetric() {
					if m.GetGauge().GetValue() != 11 {
						t.Errorf("expected gauge value 11, got %v", m.GetGauge().GetValue())
					}
				}
			}
		}
	})

	t.Run("Histogram", func(t *testing.T) {
		client.Histogram(ctx, "latency_seconds", 0.15, "endpoint", "/users")

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}
		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_latency_seconds" {
				found = true
				for _, m := range mf.GetMetric() {
					if m.GetHistogram().GetSampleCount() != 1 {
						t.Errorf("expected histogram count 1, got %v", m.GetHistogram().GetSampleCount())
					}
				}
			}
		}
		if !found {
			t.Error("histogram metric 'test_latency_seconds' not found")
		}
	})

	t.Run("Duration", func(t *testing.T) {
		start := time.Now()
		time.Sleep(time.Millisecond)
		client.Duration(ctx, "request_duration", start, "method", "GET")

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}
		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_request_duration" {
				found = true
				for _, m := range mf.GetMetric() {
					if m.GetHistogram().GetSampleSum() <= 0 {
						t.Error("expected positive duration, got 0")
					}
				}
			}
		}
		if !found {
			t.Error("histogram metric 'test_request_duration' not found")
		}
	})
}

// ============================================================================
// Handler Tests
// ============================================================================

func TestHandler_ReturnsInstanceMetrics(t *testing.T) {
	// Create an isolated registry (no Go/Process collectors for cleaner output)
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("myapp", registry,
		WithoutGoCollector(),
		WithoutProcessCollector(),
	)

	ctx := context.Background()
	client.Inc(ctx, "test_counter", "label", "value")

	handler := client.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Should contain our metric
	if !strings.Contains(body, "myapp_test_counter") {
		t.Errorf("expected handler to expose myapp_test_counter, got:\n%s", body)
	}

	// Should NOT contain go_ metrics (since we disabled Go collector)
	if strings.Contains(body, "go_goroutines") {
		t.Error("expected handler NOT to expose go_goroutines when Go collector is disabled")
	}
}

func TestHandler_WithGoCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("myapp", registry)

	handler := client.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "go_") {
		t.Error("expected Go metrics when collectors are enabled")
	}
}

// ============================================================================
// NoopClient Tests
// ============================================================================

func TestNoopClient(t *testing.T) {
	client := NewNoopClient()
	ctx := context.Background()

	// All operations should be no-ops (no panics)
	t.Run("counter operations", func(t *testing.T) {
		client.Inc(ctx, "counter", "key", "value")
		client.Add(ctx, "counter", 10, "key", "value")
	})

	t.Run("gauge operations", func(t *testing.T) {
		client.SetGauge(ctx, "gauge", 42)
		client.GaugeInc(ctx, "gauge")
		client.GaugeDec(ctx, "gauge")
	})

	t.Run("histogram operations", func(t *testing.T) {
		client.Histogram(ctx, "histogram", 0.5)
		client.Duration(ctx, "duration", time.Now())
	})

	t.Run("handler returns 200", func(t *testing.T) {
		handler := client.Handler()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("close returns nil", func(t *testing.T) {
		if err := client.Close(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
}

// ============================================================================
// Options Tests
// ============================================================================

func TestWithSubsystem(t *testing.T) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("myapp", registry, WithSubsystem("http"))

	ctx := context.Background()
	client.Inc(ctx, "requests_total", "method", "GET")

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	found := false
	for _, mf := range metricFamilies {
		// With subsystem, metric name should be myapp_http_requests_total
		if mf.GetName() == "myapp_http_requests_total" {
			found = true
		}
	}
	if !found {
		names := make([]string, 0)
		for _, mf := range metricFamilies {
			names = append(names, mf.GetName())
		}
		t.Errorf("expected metric 'myapp_http_requests_total', got metrics: %v", names)
	}
}

func TestWithConstLabels(t *testing.T) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("myapp", registry,
		WithConstLabels(map[string]string{"env": "test"}),
	)

	ctx := context.Background()
	client.Inc(ctx, "requests_total", "method", "GET")

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	for _, mf := range metricFamilies {
		if mf.GetName() == "myapp_requests_total" {
			for _, m := range mf.GetMetric() {
				hasEnvLabel := false
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "env" && lp.GetValue() == "test" {
						hasEnvLabel = true
					}
				}
				if !hasEnvLabel {
					t.Error("expected const label 'env=test' on metric")
				}
			}
		}
	}
}

func TestWithBuckets(t *testing.T) {
	registry := prometheus.NewRegistry()
	customBuckets := []float64{0.1, 0.5, 1.0}
	client := NewClientWithRegistry("myapp", registry, WithBuckets(customBuckets))

	ctx := context.Background()
	client.Histogram(ctx, "request_duration", 0.3)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	for _, mf := range metricFamilies {
		if mf.GetName() == "myapp_request_duration" {
			for _, m := range mf.GetMetric() {
				bucketCount := len(m.GetHistogram().GetBucket())
				// 3 custom buckets + 1 implicit +Inf = 4 total (but prometheus Gather returns explicit ones)
				if bucketCount != len(customBuckets) {
					t.Errorf("expected %d buckets, got %d", len(customBuckets), bucketCount)
				}
			}
		}
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestOddTagsHandling(t *testing.T) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("test", registry)
	ctx := context.Background()

	// Odd number of tags should not panic
	client.Inc(ctx, "odd_tags_counter", "method")
	client.SetGauge(ctx, "odd_tags_gauge", 1, "key")
	client.Histogram(ctx, "odd_tags_histogram", 0.5, "endpoint")

	// Verify metrics were created (with "unknown" as the missing value)
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expectedMetrics := map[string]bool{
		"test_odd_tags_counter":   false,
		"test_odd_tags_gauge":     false,
		"test_odd_tags_histogram": false,
	}

	for _, mf := range metricFamilies {
		if _, ok := expectedMetrics[mf.GetName()]; ok {
			expectedMetrics[mf.GetName()] = true
		}
	}

	for name, found := range expectedMetrics {
		if !found {
			t.Errorf("expected metric %q to be registered", name)
		}
	}
}

func TestNewClientWithRegisterer(t *testing.T) {
	t.Run("with custom registry", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		client := NewClientWithRegisterer("myapp", registry,
			WithSubsystem("test"),
			WithConstLabels(map[string]string{"env": "test"}),
		)

		ctx := context.Background()
		client.Inc(ctx, "requests_total", "method", "GET")

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == "myapp_test_requests_total" {
				found = true
			}
		}
		if !found {
			t.Error("expected metric 'myapp_test_requests_total' not found")
		}
	})

	t.Run("with custom registerer that is not a gatherer", func(t *testing.T) {
		// Use a custom registerer implementation that doesn't implement Gatherer
		customReg := &mockRegisterer{}
		client := NewClientWithRegisterer("app", customReg)

		ctx := context.Background()
		// Should not panic even though registerer is not a gatherer
		client.Inc(ctx, "counter", "key", "value")
		client.SetGauge(ctx, "gauge", 42.0)

		// Handler should still work (using fallback registry)
		handler := client.Handler()
		if handler == nil {
			t.Error("expected non-nil handler even with non-Gatherer registerer")
		}
	})

	t.Run("with options", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		client := NewClientWithRegisterer("myapp", registry,
			WithBuckets([]float64{0.1, 1.0, 10.0}),
			WithoutGoCollector(),
			WithoutProcessCollector(),
		)

		ctx := context.Background()
		client.Histogram(ctx, "request_duration", 0.5)

		// Verify the client was created successfully
		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		hasGoMetrics := false
		for _, mf := range metricFamilies {
			if strings.Contains(mf.GetName(), "go_") {
				hasGoMetrics = true
			}
		}
		if hasGoMetrics {
			t.Error("expected no Go metrics when disabled via options")
		}
	})
}

// mockRegisterer is a simple Registerer that doesn't implement Gatherer
type mockRegisterer struct{}

func (m *mockRegisterer) Register(prometheus.Collector) error {
	return nil
}

func (m *mockRegisterer) MustRegister(...prometheus.Collector) {}

func (m *mockRegisterer) Unregister(prometheus.Collector) bool {
	return true
}

func TestClose(t *testing.T) {
	client := NewClient("test")
	if err := client.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkPrometheusInc(b *testing.B) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("bench", registry)
	ctx := context.Background()

	// Pre-create the metric
	client.Inc(ctx, "counter", "method", "GET")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.Inc(ctx, "counter", "method", "GET")
		}
	})
}

func BenchmarkPrometheusDuration(b *testing.B) {
	registry := prometheus.NewRegistry()
	client := NewClientWithRegistry("bench", registry)
	ctx := context.Background()

	// Pre-create the metric
	client.Duration(ctx, "duration", time.Now(), "method", "GET")

	b.ResetTimer()
	start := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.Duration(ctx, "duration", start, "method", "GET")
		}
	})
}
