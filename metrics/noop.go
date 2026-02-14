// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package metrics

import (
	"context"
	"net/http"
	"time"
)

// ============================================================================
// NoopClient (No-Operation Client)
// ============================================================================

// noopClient is a no-operation implementation of the Client interface.
// All methods are no-ops that return immediately. This is useful for:
//   - Testing environments where metrics collection is not needed
//   - Disabling metrics without changing application code
//   - Providing a safe default when metrics are optional
//
// Example:
//
//	client := metrics.NewNoopClient()
//	client.Inc(ctx, "requests_total") // does nothing
type noopClient struct{}

// NewNoopClient creates a new no-operation metrics client.
// All metric operations are silently discarded.
//
// Example:
//
//	// Use in tests
//	client := metrics.NewNoopClient()
//
//	// Use as default when metrics are optional
//	var metricsClient metrics.Client = metrics.NewNoopClient()
//	if enableMetrics {
//	    metricsClient = metrics.NewClient("myapp")
//	}
func NewNoopClient() Client {
	return &noopClient{}
}

func (*noopClient) Inc(_ context.Context, _ string, _ ...string)                   {}
func (*noopClient) Add(_ context.Context, _ string, _ int64, _ ...string)          {}
func (*noopClient) SetGauge(_ context.Context, _ string, _ float64, _ ...string)   {}
func (*noopClient) GaugeInc(_ context.Context, _ string, _ ...string)              {}
func (*noopClient) GaugeDec(_ context.Context, _ string, _ ...string)              {}
func (*noopClient) Histogram(_ context.Context, _ string, _ float64, _ ...string)  {}
func (*noopClient) Duration(_ context.Context, _ string, _ time.Time, _ ...string) {}
func (*noopClient) Close() error                                                   { return nil }

// Handler returns a handler that responds with 200 OK and an empty body.
func (*noopClient) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
