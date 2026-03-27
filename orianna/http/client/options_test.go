package client

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"
)

func TestClient_Options(t *testing.T) {
	tlsConf := &tls.Config{InsecureSkipVerify: true}
	tr := tracing.NewNoopClient()
	mc := metrics.NewNoopClient()
	logr := logger.NewLogger(&logger.Config{LogLevel: logger.LevelInfo}, []io.Writer{os.Stdout})

	c, err := NewClient(
		WithTLSConfig(tlsConf),
		WithTracing(tr),
		WithMetrics(mc),
		WithLogger(logr),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if c.config.TLSConfig != tlsConf {
		t.Error("expected tls config to be set")
	}
	if c.tracing == nil {
		t.Error("expected tracer to be set")
	}
	if c.metrics == nil {
		t.Error("expected metrics to be set")
	}
	if c.logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestClient_Verbs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	c, err := NewClient(WithBaseURL(ts.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := c.Put(context.Background(), "/test", map[string]string{"key": "val"})
	if err != nil || res.StatusCode != 200 {
		t.Error("Put failed")
	}

	res, err = c.Patch(context.Background(), "/test", map[string]string{"key": "val"})
	if err != nil || res.StatusCode != 200 {
		t.Error("Patch failed")
	}

	res, err = c.Delete(context.Background(), "/test")
	if err != nil || res.StatusCode != 200 {
		t.Error("Delete failed")
	}
}

func TestClient_Decode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":"test"}`))
	}))
	defer ts.Close()

	c, err := NewClient(WithBaseURL(ts.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	type result struct {
		Value string `json:"value"`
	}

	res, err := c.Get(context.Background(), "/test")
	if err != nil {
		t.Fatal(err)
	}

	r, err := Decode[result](res)
	if err != nil || r.Value != "test" {
		t.Error("Decode failed")
	}

	r2, err := GetJSON[result](c, context.Background(), "/test")
	if err != nil || r2.Value != "test" {
		t.Error("GetJSON failed")
	}

	r3, err := PostJSON[result](c, context.Background(), "/test", nil)
	if err != nil || r3.Value != "test" {
		t.Error("PostJSON failed")
	}

	r4, err := PutJSON[result](c, context.Background(), "/test", nil)
	if err != nil || r4.Value != "test" {
		t.Error("PutJSON failed")
	}

	r5, err := PatchJSON[result](c, context.Background(), "/test", nil)
	if err != nil || r5.Value != "test" {
		t.Error("PatchJSON failed")
	}

	r6, err := DeleteJSON[result](c, context.Background(), "/test")
	if err != nil || r6.Value != "test" {
		t.Error("DeleteJSON failed")
	}
}

func TestClient_Standalone(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if _, err := Get(context.Background(), ts.URL, "/test"); err != nil {
		t.Error(err)
	}
	if _, err := Post(context.Background(), ts.URL, "/test", nil); err != nil {
		t.Error(err)
	}
	if _, err := Put(context.Background(), ts.URL, "/test", nil); err != nil {
		t.Error(err)
	}
	if _, err := Patch(context.Background(), ts.URL, "/test", nil); err != nil {
		t.Error(err)
	}
	if _, err := Delete(context.Background(), ts.URL, "/test"); err != nil {
		t.Error(err)
	}
}

func TestClient_RetryConfig(t *testing.T) {
	rc := resilience.DefaultRetryConfig()
	if rc.MaxAttempts != 3 {
		t.Error("expected 3 attempts")
	}

	rc2 := resilience.DefaultRetryConfig()
	resilience.WithMaxAttempts(5)(rc2)
	resilience.WithBackoff(time.Second, 10*time.Second)(rc2)
	if rc2.MaxAttempts != 5 || rc2.InitialBackoff != time.Second {
		t.Error("expected overrides to be applied")
	}
}

func TestClient_WithHeaderOption(t *testing.T) {
	req := &Request{}
	WithHeader("X-Test", "123")(req)
	if req.Headers["X-Test"] != "123" {
		t.Error("expected header to be set")
	}
}

func TestClient_TracingAndMetrics(t *testing.T) {
	tr := tracing.NewNoopClient()
	mc := metrics.NewNoopClient()

	c, err := NewClient(
		WithTracing(tr),
		WithMetrics(mc),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, _ = c.Get(context.Background(), ts.URL)

	// Error path
	_, _ = c.Get(context.Background(), "http://invalid.local")
}
