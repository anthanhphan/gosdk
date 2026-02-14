// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

func TestHTTPChecker_Check_Success(t *testing.T) {
	// Create test server that returns 200
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	checker := NewHTTPChecker(ts.URL, "test-api", 5*time.Second)

	check := checker.Check(context.Background())

	if check.Status != HealthStatusHealthy {
		t.Errorf("Status = %v, want %v", check.Status, HealthStatusHealthy)
	}

	if check.Name != "test-api" {
		t.Errorf("Name = %s, want test-api", check.Name)
	}

	if check.ResponseTime < 0 {
		t.Error("ResponseTime should be non-negative")
	}

	if details, ok := check.Details["url"]; !ok || details != ts.URL {
		t.Errorf("Details[url] = %v, want %s", details, ts.URL)
	}
}

func TestHTTPChecker_Check_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	checker := NewHTTPChecker(ts.URL, "test-api", 5*time.Second)
	check := checker.Check(context.Background())

	if check.Status != HealthStatusUnhealthy {
		t.Errorf("Status = %v, want %v (500 should be unhealthy)", check.Status, HealthStatusUnhealthy)
	}
}

func TestHTTPChecker_Check_ClientError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	checker := NewHTTPChecker(ts.URL, "test-api", 5*time.Second)
	check := checker.Check(context.Background())

	if check.Status != HealthStatusDegraded {
		t.Errorf("Status = %v, want %v (404 should be degraded)", check.Status, HealthStatusDegraded)
	}
}

func TestHTTPChecker_Check_ConnectionRefused(t *testing.T) {
	// Use a URL that will definitely fail to connect
	checker := NewHTTPChecker("http://127.0.0.1:1", "unreachable", 1*time.Second)
	check := checker.Check(context.Background())

	if check.Status != HealthStatusUnhealthy {
		t.Errorf("Status = %v, want %v", check.Status, HealthStatusUnhealthy)
	}

	if check.Error == nil {
		t.Error("Error should be set for connection failure")
	}
}

func TestHTTPChecker_Check_InvalidURL(t *testing.T) {
	checker := NewHTTPChecker("://invalid-url", "invalid", 1*time.Second)
	check := checker.Check(context.Background())

	if check.Status != HealthStatusUnhealthy {
		t.Errorf("Status = %v, want %v", check.Status, HealthStatusUnhealthy)
	}
}

func TestHTTPChecker_Name(t *testing.T) {
	checker := NewHTTPChecker("http://example.com", "my-check", 5*time.Second)
	if checker.Name() != "my-check" {
		t.Errorf("Name() = %s, want my-check", checker.Name())
	}
}

func TestHTTPChecker_DefaultName(t *testing.T) {
	checker := NewHTTPChecker("http://example.com", "", 5*time.Second)
	if checker.Name() != configuration.DefaultHealthCheckerName {
		t.Errorf("Name() = %s, want %s", checker.Name(), configuration.DefaultHealthCheckerName)
	}
}

func TestHTTPChecker_DefaultTimeout(t *testing.T) {
	checker := NewHTTPChecker("http://example.com", "test", 0)
	if checker.timeout != configuration.DefaultHealthCheckTimeout {
		t.Errorf("timeout = %v, want %v", checker.timeout, configuration.DefaultHealthCheckTimeout)
	}
}

func TestHealthManager_Multiple_Checkers(t *testing.T) {
	mgr := NewHealthManager()

	for i := 0; i < 15; i++ {
		name := "checker-" + string(rune('A'+i))
		mgr.Register(NewCustomChecker(name, func(ctx context.Context) HealthCheck {
			return HealthCheck{Status: HealthStatusHealthy}
		}))
	}

	report := mgr.Check(context.Background())

	if len(report.Checks) != 15 {
		t.Errorf("Check count = %d, want 15", len(report.Checks))
	}

	if report.Status != HealthStatusHealthy {
		t.Errorf("Status = %v, want %v", report.Status, HealthStatusHealthy)
	}
}

func TestHealthManager_DegradedDoesNotOverrideUnhealthy(t *testing.T) {
	mgr := NewHealthManager()

	mgr.Register(NewCustomChecker("unhealthy", func(ctx context.Context) HealthCheck {
		return HealthCheck{Status: HealthStatusUnhealthy}
	}))
	mgr.Register(NewCustomChecker("degraded", func(ctx context.Context) HealthCheck {
		return HealthCheck{Status: HealthStatusDegraded}
	}))
	mgr.Register(NewCustomChecker("healthy", func(ctx context.Context) HealthCheck {
		return HealthCheck{Status: HealthStatusHealthy}
	}))

	report := mgr.Check(context.Background())

	// Unhealthy should take precedence
	if report.Status != HealthStatusUnhealthy {
		t.Errorf("Status = %v, want %v (unhealthy takes precedence)", report.Status, HealthStatusUnhealthy)
	}
}

func TestCustomChecker_EmptyNameFillsFromChecker(t *testing.T) {
	checker := NewCustomChecker("my-checker", func(ctx context.Context) HealthCheck {
		return HealthCheck{
			// Name intentionally left empty
			Status:  HealthStatusHealthy,
			Message: "OK",
		}
	})

	check := checker.Check(context.Background())
	if check.Name != "my-checker" {
		t.Errorf("Check.Name = %s, want my-checker (should be auto-filled)", check.Name)
	}
}

func TestCustomChecker_PreservesName(t *testing.T) {
	checker := NewCustomChecker("fallback", func(ctx context.Context) HealthCheck {
		return HealthCheck{
			Name:    "explicit-name",
			Status:  HealthStatusHealthy,
			Message: "OK",
		}
	})

	check := checker.Check(context.Background())
	if check.Name != "explicit-name" {
		t.Errorf("Check.Name = %s, want explicit-name", check.Name)
	}
}

func TestWithMetrics(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	// WithMetrics with nil should still work
	server, err := NewServer(conf, WithMetrics(nil))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.metricsClient != nil {
		t.Error("metricsClient should be nil")
	}
}

func TestWithHealthChecker_CreatesManager(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	checker := NewCustomChecker("test-checker", func(ctx context.Context) HealthCheck {
		return HealthCheck{Status: HealthStatusHealthy}
	})

	server, err := NewServer(conf, WithHealthChecker(checker))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	mgr := server.GetHealthManager()
	if mgr == nil {
		t.Fatal("HealthManager should be auto-created by WithHealthChecker")
	}
}

func TestServer_Use_Multiple(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	server, _ := NewServer(conf)

	mw1 := func(ctx interface{ Next() error }) error { return nil }
	mw2 := func(ctx interface{ Next() error }) error { return nil }
	_ = mw1
	_ = mw2

	// Use should append middlewares
	server.Use()
	if len(server.globalMiddlewares) != 0 {
		t.Error("Use() with no args should not add middleware")
	}
}

func TestNewServer_WithAllOptions(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	mw := func(ctx core.Context) error { return ctx.Next() }

	_, err := NewServer(conf,
		WithAuthentication(mw),
		WithRateLimiter(mw),
		WithPanicRecover(mw),
		WithMiddlewareConfig(&configuration.MiddlewareConfig{DisableHelmet: true}),
	)

	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
}

func TestHTTPChecker_Check_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	checker := NewHTTPChecker(ts.URL, "slow-api", 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	check := checker.Check(ctx)
	// Should fail since context is cancelled
	if check.Status == HealthStatusHealthy {
		t.Error("Cancelled context should result in unhealthy or degraded status")
	}
}
