// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
)

func TestNewClient_DefaultConfig(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if c.config.Timeout != 30*time.Second {
		t.Errorf("default timeout = %v, want %v", c.config.Timeout, 30*time.Second)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("https://example.com"),
		WithTimeout(10*time.Second),
		WithDefaultHeader("X-Custom", "value"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %v, want https://example.com", c.baseURL)
	}
	if c.config.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", c.config.Timeout)
	}
	if c.headers["X-Custom"] != "value" {
		t.Errorf("header X-Custom = %v, want value", c.headers["X-Custom"])
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("path = %v, want /test", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := c.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
	if len(resp.Body) == 0 {
		t.Error("Body should not be empty")
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %v, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", ct)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	c, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	body := map[string]string{"name": "test"}
	resp, err := c.Post(context.Background(), "/items", body)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("StatusCode = %v, want 201", resp.StatusCode)
	}
}

func TestClient_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Errorf("query param page = %v, want 1", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("query param limit = %v, want 10", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = c.Get(context.Background(), "/items",
		WithQuery("page", "1"),
		WithQuery("limit", "10"),
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %v, want Bearer test-token", got)
		}
		if got := r.Header.Get("X-Default"); got != "default-value" {
			t.Errorf("X-Default = %v, want default-value", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := NewClient(
		WithBaseURL(server.URL),
		WithDefaultHeader("X-Default", "default-value"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = c.Get(context.Background(), "/test", WithAuth("Bearer test-token"))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_RetryOnServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := NewClient(
		WithBaseURL(server.URL),
		WithRetry(&resilience.RetryConfig{
			MaxAttempts:          3,
			InitialBackoff:       1 * time.Millisecond,
			MaxBackoff:           10 * time.Millisecond,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{503},
			RetryableMethods:     []string{"GET"},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := c.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("attempts = %v, want 3", attempts)
	}
}

func TestClient_RetryExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c, err := NewClient(
		WithBaseURL(server.URL),
		WithRetry(&resilience.RetryConfig{
			MaxAttempts:          2,
			InitialBackoff:       1 * time.Millisecond,
			MaxBackoff:           10 * time.Millisecond,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{503},
			RetryableMethods:     []string{"GET"},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = c.Get(context.Background(), "/test")
	if err == nil {
		t.Error("expected error when retries exhausted")
	}
}

func TestClient_CircuitBreakerIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c, err := NewClient(
		WithBaseURL(server.URL),
		WithCircuitBreaker(&resilience.CircuitBreakerConfig{
			FailureThreshold:    2,
			SuccessThreshold:    1,
			Timeout:             1 * time.Second,
			HalfOpenMaxRequests: 1,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Make requests until circuit opens
	for i := 0; i < 3; i++ {
		_, _ = c.Get(context.Background(), "/fail")
	}

	// Circuit should now be open
	_, err = c.Get(context.Background(), "/fail")
	if err == nil {
		t.Error("expected circuit breaker error")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = c.Get(ctx, "/slow")
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestClient_Backoff(t *testing.T) {
	c := &Client{
		retry: &resilience.RetryConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     1 * time.Second,
			Multiplier:     2.0,
		},
	}

	// With jitter (0-25%), backoff should be in [base, base * 1.25]
	tests := []struct {
		attempt int
		base    time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1 * time.Second}, // capped at max
	}

	for _, tt := range tests {
		got := c.calculateBackoff(tt.attempt)
		maxExpected := time.Duration(float64(tt.base) * 1.25)
		if got < tt.base || got > maxExpected {
			t.Errorf("calculateBackoff(%d) = %v, want between %v and %v", tt.attempt, got, tt.base, maxExpected)
		}
	}
}

func TestClient_IsRetryable(t *testing.T) {
	c := &Client{
		retry: &resilience.RetryConfig{
			RetryableStatusCodes: []int{503, 429},
			RetryableMethods:     []string{"GET", "HEAD"},
		},
		retryStatusSet: map[int]struct{}{503: {}, 429: {}},
		retryMethodSet: map[string]struct{}{"GET": {}, "HEAD": {}},
	}

	tests := []struct {
		name   string
		method string
		resp   *Response
		err    error
		want   bool
	}{
		{"GET 503 retryable", "GET", &Response{StatusCode: 503}, nil, true},
		{"GET 429 retryable", "GET", &Response{StatusCode: 429}, nil, true},
		{"GET 200 not retryable", "GET", &Response{StatusCode: 200}, nil, false},
		{"POST 503 not retryable (method)", "POST", &Response{StatusCode: 503}, nil, false},
		{"GET nil resp not retryable", "GET", nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.isRetryable(tt.method, tt.resp, tt.err)
			if got != tt.want {
				t.Errorf("isRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_IsRetryable_NetworkError(t *testing.T) {
	c := &Client{
		retry:          &resilience.RetryConfig{RetryableMethods: []string{"GET"}},
		retryStatusSet: map[int]struct{}{},
		retryMethodSet: map[string]struct{}{"GET": {}},
	}

	// Network error with retryable method
	got := c.isRetryable("GET", nil, fmt.Errorf("connection refused"))
	if !got {
		t.Error("isRetryable() should return true for network errors with retryable method")
	}

	// Network error with non-retryable method
	got = c.isRetryable("POST", nil, fmt.Errorf("connection refused"))
	if got {
		t.Error("isRetryable() should return false for network errors with non-retryable method")
	}
}

func TestClient_IsMethodRetryable_EmptySet(t *testing.T) {
	c := &Client{
		retry:          &resilience.RetryConfig{},
		retryMethodSet: map[string]struct{}{},
	}
	// Empty method set means all methods are retryable
	if !c.isMethodRetryable("POST") {
		t.Error("isMethodRetryable() with empty set should return true for any method")
	}
	if !c.isMethodRetryable("DELETE") {
		t.Error("isMethodRetryable() with empty set should return true for DELETE")
	}
}

func TestDecode_NilResponse(t *testing.T) {
	_, err := Decode[map[string]string](nil)
	if err == nil {
		t.Error("Decode(nil) should return error")
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	resp := &Response{Body: []byte("not json")}
	_, err := Decode[map[string]string](resp)
	if err == nil {
		t.Error("Decode(invalid json) should return error")
	}
}

func TestJSON_ErrorPaths(t *testing.T) {
	// Use a non-reachable server to trigger errors
	c, _ := NewClient(WithBaseURL("http://127.0.0.1:1"))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := GetJSON[map[string]string](c, ctx, "/err"); err == nil {
		t.Error("GetJSON should error on unreachable server")
	}
	if _, err := PostJSON[map[string]string](c, ctx, "/err", nil); err == nil {
		t.Error("PostJSON should error on unreachable server")
	}
	if _, err := PutJSON[map[string]string](c, ctx, "/err", nil); err == nil {
		t.Error("PutJSON should error on unreachable server")
	}
	if _, err := PatchJSON[map[string]string](c, ctx, "/err", nil); err == nil {
		t.Error("PatchJSON should error on unreachable server")
	}
	if _, err := DeleteJSON[map[string]string](c, ctx, "/err"); err == nil {
		t.Error("DeleteJSON should error on unreachable server")
	}
}

func TestClient_BuildRequest_MarshalError(t *testing.T) {
	c, _ := NewClient(WithBaseURL("http://localhost"))
	// func values cannot be marshaled to JSON
	_, err := c.Post(context.Background(), "/test", func() {})
	if err == nil {
		t.Error("Post with unmarshalable body should error")
	}
}

func TestClient_RetryWithNetworkError(t *testing.T) {
	c, _ := NewClient(
		WithBaseURL("http://127.0.0.1:1"),
		WithRetry(&resilience.RetryConfig{
			MaxAttempts:      2,
			InitialBackoff:   1 * time.Millisecond,
			MaxBackoff:       10 * time.Millisecond,
			Multiplier:       2.0,
			RetryableMethods: []string{},
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, err := c.Get(ctx, "/fail")
	if err == nil {
		t.Error("expected error on unreachable server with retry")
	}
}

// ---------- Close ----------

func TestClient_Close(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	// Close should not panic
	c.Close()
}

// ---------- WithServiceName ----------

func TestWithServiceName(t *testing.T) {
	c, err := NewClient(WithServiceName("user-service"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.serviceName != "user-service" {
		t.Errorf("serviceName = %v, want user-service", c.serviceName)
	}
	// Metric prefix should include service name
	if c.metricRequestsTotal != "user-service_http_requests_total" {
		t.Errorf("metricRequestsTotal = %v, want user-service_http_requests_total", c.metricRequestsTotal)
	}
}

// ---------- CRUD Methods ----------

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %v, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, _ := NewClient(WithBaseURL(server.URL))
	resp, err := c.Put(context.Background(), "/items/1", map[string]string{"name": "updated"})
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %v, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, _ := NewClient(WithBaseURL(server.URL))
	resp, err := c.Patch(context.Background(), "/items/1", map[string]string{"name": "patched"})
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %v, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c, _ := NewClient(WithBaseURL(server.URL))
	resp, err := c.Delete(context.Background(), "/items/1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("StatusCode = %v, want 204", resp.StatusCode)
	}
}

// ---------- Standalone Functions ----------

func TestStandalone_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %v, want GET", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Clean up the default clients map before and after test
	defaultClientsMu.Lock()
	delete(defaultClients, server.URL)
	defaultClientsMu.Unlock()
	defer func() {
		defaultClientsMu.Lock()
		delete(defaultClients, server.URL)
		defaultClientsMu.Unlock()
	}()

	resp, err := Get(context.Background(), server.URL, "/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestStandalone_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %v, want POST", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	defaultClientsMu.Lock()
	delete(defaultClients, server.URL)
	defaultClientsMu.Unlock()

	resp, err := Post(context.Background(), server.URL, "/items", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("StatusCode = %v, want 201", resp.StatusCode)
	}
}

func TestStandalone_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	defaultClientsMu.Lock()
	delete(defaultClients, server.URL)
	defaultClientsMu.Unlock()

	resp, err := Put(context.Background(), server.URL, "/items/1", map[string]string{"name": "updated"})
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestStandalone_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	defaultClientsMu.Lock()
	delete(defaultClients, server.URL)
	defaultClientsMu.Unlock()

	resp, err := Patch(context.Background(), server.URL, "/items/1", map[string]string{"name": "patched"})
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestStandalone_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	defaultClientsMu.Lock()
	delete(defaultClients, server.URL)
	defaultClientsMu.Unlock()

	resp, err := Delete(context.Background(), server.URL, "/items/1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("StatusCode = %v, want 204", resp.StatusCode)
	}
}

// ---------- Standalone Client Caching ----------

func TestStandalone_GetOrCreateClient_CacheReuse(t *testing.T) {
	defaultClientsMu.Lock()
	delete(defaultClients, "http://test-cache:8080")
	defaultClientsMu.Unlock()

	c1, err := getOrCreateClient("http://test-cache:8080")
	if err != nil {
		t.Fatalf("getOrCreateClient() error = %v", err)
	}

	c2, err := getOrCreateClient("http://test-cache:8080")
	if err != nil {
		t.Fatalf("getOrCreateClient() error = %v", err)
	}

	if c1 != c2 {
		t.Error("expected same client to be returned from cache")
	}

	// Cleanup
	defaultClientsMu.Lock()
	delete(defaultClients, "http://test-cache:8080")
	defaultClientsMu.Unlock()
}
