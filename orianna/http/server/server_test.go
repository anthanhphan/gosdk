// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/orianna/shared/health"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/routing"
)

func TestServer_RegisterRoutes(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	route := routing.Route{
		Path:    "/test",
		Methods: []core.Method{core.GET},
		Handler: func(_ core.Context) error { return nil },
	}

	err = server.RegisterRoutes(route)
	if err != nil {
		t.Errorf("RegisterRoutes() error = %v", err)
	}
}

func TestServer_RegisterGroup(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	group := routing.RouteGroup{
		Prefix: "/api",
		Routes: []routing.Route{
			{
				Path:    "/users",
				Methods: []core.Method{core.GET},
				Handler: func(_ core.Context) error { return nil },
			},
		},
	}

	err = server.RegisterGroup(group)
	if err != nil {
		t.Errorf("RegisterGroup() error = %v", err)
	}
}

func TestServer_GetHealthManager(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	healthMgr := health.NewManager()
	server, err := NewServer(conf, WithHealthManager(healthMgr))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.GetHealthManager() == nil {
		t.Error("GetHealthManager() should return health manager")
	}
}

func TestServer_GetShutdownManager(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	shutdownMgr := &mockShutdownManager{}
	server, err := NewServer(conf, WithShutdownManager(shutdownMgr))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.GetShutdownManager() == nil {
		t.Error("GetShutdownManager() should return shutdown manager")
	}
}

type mockShutdownManager struct{}

func (m *mockShutdownManager) Shutdown(ctx context.Context) error {
	return nil
}

func TestServerOptions_WithGlobalMiddleware(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	mw := func(ctx core.Context) error { return ctx.Next() }
	server, err := NewServer(conf, WithGlobalMiddleware(mw))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if len(server.globalMiddlewares) != 1 {
		t.Errorf("globalMiddlewares count = %d, want 1", len(server.globalMiddlewares))
	}
}

func TestServerOptions_WithAuthentication(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	authMw := func(ctx core.Context) error { return ctx.Next() }
	server, err := NewServer(conf, WithAuthentication(authMw))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.authMiddleware == nil {
		t.Error("authMiddleware should be set")
	}
}

func TestServerOptions_WithAuthorization(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	authzChecker := func(_ core.Context, _ []string) error { return nil }
	server, err := NewServer(conf, WithAuthorization(authzChecker))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.authzChecker == nil {
		t.Error("authzChecker should be set")
	}
}

func TestServerOptions_WithRateLimiter(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	rateLimiter := func(ctx core.Context) error { return ctx.Next() }
	server, err := NewServer(conf, WithRateLimiter(rateLimiter))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.rateLimiter == nil {
		t.Error("rateLimiter should be set")
	}
}

func TestServerOptions_WithPanicRecover(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	panicRecover := func(ctx core.Context) error { return ctx.Next() }
	server, err := NewServer(conf, WithPanicRecover(panicRecover))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.panicRecover == nil {
		t.Error("panicRecover should be set")
	}
}

func TestServerOptions_WithHooks(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	hooks := core.NewHooks()
	server, err := NewServer(conf, WithHooks(hooks))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.hooks == nil {
		t.Error("hooks should be set")
	}
}

func TestServerOptions_WithMiddlewareConfig(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	mwConfig := &configuration.MiddlewareConfig{
		DisableHelmet: true,
	}
	server, err := NewServer(conf, WithMiddlewareConfig(mwConfig))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if !server.middlewareConfig.DisableHelmet {
		t.Error("middlewareConfig.DisableHelmet should be true")
	}
}

func TestServerShortcuts_GET(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.GET("/test", handler)
	if err != nil {
		t.Errorf("GET() error = %v", err)
	}
}

func TestServerShortcuts_POST(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.POST("/test", handler)
	if err != nil {
		t.Errorf("POST() error = %v", err)
	}
}

func TestServerShortcuts_PUT(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.PUT("/test/:id", handler)
	if err != nil {
		t.Errorf("PUT() error = %v", err)
	}
}

func TestServerShortcuts_DELETE(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.DELETE("/test/:id", handler)
	if err != nil {
		t.Errorf("DELETE() error = %v", err)
	}
}

func TestServerShortcuts_PATCH(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.PATCH("/test/:id", handler)
	if err != nil {
		t.Errorf("PATCH() error = %v", err)
	}
}

func TestServerShortcuts_HEAD(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.HEAD("/test", handler)
	if err != nil {
		t.Errorf("HEAD() error = %v", err)
	}
}

func TestServerShortcuts_OPTIONS(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.OPTIONS("/test", handler)
	if err != nil {
		t.Errorf("OPTIONS() error = %v", err)
	}
}

func TestServerShortcuts_Protected(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.Protected().GET("/admin", handler)
	if err != nil {
		t.Errorf("Protected().GET() error = %v", err)
	}
}

func TestServerShortcuts_ProtectedWithPermissions(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	err := server.Protected().WithPermissions("admin").GET("/admin", handler)
	if err != nil {
		t.Errorf("Protected().WithPermissions().GET() error = %v", err)
	}
}

func TestServerShortcuts_ProtectedWithMiddleware(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	mw := func(ctx core.Context) error { return ctx.Next() }
	err := server.Protected().Middleware(mw).GET("/admin", handler)
	if err != nil {
		t.Errorf("Protected().Middleware().GET() error = %v", err)
	}
}

func TestServerShortcuts_ProtectedAllMethods(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	protected := server.Protected()

	methods := []func(string, core.Handler, ...core.Middleware) error{
		protected.GET,
		protected.POST,
		protected.PUT,
		protected.DELETE,
		protected.PATCH,
		protected.HEAD,
		protected.OPTIONS,
	}

	for _, method := range methods {
		if err := method("/test", handler); err != nil {
			t.Errorf("Protected method error = %v", err)
		}
	}
}

func TestMergeConfig(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		// Leave defaults empty
	}

	merged := mergeConfig(conf)

	if merged.Port != configuration.DefaultPort {
		t.Errorf("Port = %d, want %d", merged.Port, configuration.DefaultPort)
	}

	if merged.MaxBodySize != configuration.DefaultMaxBodySize {
		t.Errorf("MaxBodySize = %d, want %d", merged.MaxBodySize, configuration.DefaultMaxBodySize)
	}

	if merged.MaxConcurrentConnections != configuration.DefaultMaxConcurrentConnections {
		t.Errorf("MaxConcurrentConnections = %d, want %d", merged.MaxConcurrentConnections, configuration.DefaultMaxConcurrentConnections)
	}
}

func TestMergeConfig_PreservesExisting(t *testing.T) {
	conf := &configuration.Config{
		ServiceName:              "test",
		Port:                     9000,
		MaxBodySize:              1024,
		MaxConcurrentConnections: 500,
	}

	merged := mergeConfig(conf)

	if merged.Port != 9000 {
		t.Errorf("Port = %d, want 9000", merged.Port)
	}

	if merged.MaxBodySize != 1024 {
		t.Errorf("MaxBodySize = %d, want 1024", merged.MaxBodySize)
	}

	if merged.MaxConcurrentConnections != 500 {
		t.Errorf("MaxConcurrentConnections = %d, want 500", merged.MaxConcurrentConnections)
	}
}

func TestNewServer_InvalidConfig(t *testing.T) {
	conf := &configuration.Config{
		// Missing required ServiceName
		Port: 8080,
	}

	_, err := NewServer(conf)
	if err == nil {
		t.Error("NewServer() should return error for invalid config")
	}
}

func TestHealthManager_Register(t *testing.T) {
	mgr := health.NewManager()
	checker := health.NewCustomChecker("test", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{
			Status:  health.StatusHealthy,
			Message: "OK",
		}
	})

	mgr.Register(checker)

	report := mgr.Check(context.Background())
	if len(report.Checks) != 1 {
		t.Errorf("Check count = %d, want 1", len(report.Checks))
	}
}

func TestHealthManager_Check_NoCheckers(t *testing.T) {
	mgr := health.NewManager()
	report := mgr.Check(context.Background())

	if report.Status != health.StatusHealthy {
		t.Errorf("Status = %v, want %v", report.Status, health.StatusHealthy)
	}

	if len(report.Checks) != 0 {
		t.Errorf("Check count = %d, want 0", len(report.Checks))
	}
}

func TestHealthManager_Check_Unhealthy(t *testing.T) {
	mgr := health.NewManager()
	checker := health.NewCustomChecker("test", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{
			Status:  health.StatusUnhealthy,
			Message: "Failed",
		}
	})

	mgr.Register(checker)
	report := mgr.Check(context.Background())

	if report.Status != health.StatusUnhealthy {
		t.Errorf("Status = %v, want %v", report.Status, health.StatusUnhealthy)
	}
}

func TestHealthManager_Check_Degraded(t *testing.T) {
	mgr := health.NewManager()
	checker := health.NewCustomChecker("test", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{
			Status:  health.StatusDegraded,
			Message: "Slow",
		}
	})

	mgr.Register(checker)
	report := mgr.Check(context.Background())

	if report.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", report.Status, health.StatusDegraded)
	}
}

func TestHealthManager_Check_Mixed(t *testing.T) {
	mgr := health.NewManager()

	healthy := health.NewCustomChecker("healthy", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{Status: health.StatusHealthy, Message: "OK"}
	})

	unhealthy := health.NewCustomChecker("unhealthy", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{Status: health.StatusUnhealthy, Message: "Failed"}
	})

	mgr.Register(healthy)
	mgr.Register(unhealthy)
	report := mgr.Check(context.Background())

	// Unhealthy takes precedence
	if report.Status != health.StatusUnhealthy {
		t.Errorf("Status = %v, want %v", report.Status, health.StatusUnhealthy)
	}

	if len(report.Checks) != 2 {
		t.Errorf("Check count = %d, want 2", len(report.Checks))
	}
}

func TestHealthManager_Check_Panic(t *testing.T) {
	mgr := health.NewManager()
	checker := health.NewCustomChecker("panic", func(ctx context.Context) health.HealthCheck {
		panic("test panic")
	})

	mgr.Register(checker)
	report := mgr.Check(context.Background())

	if report.Status != health.StatusUnhealthy {
		t.Errorf("Status = %v, want %v (panic should be caught)", report.Status, health.StatusUnhealthy)
	}

	if check, ok := report.Checks["panic"]; ok {
		if check.Status != health.StatusUnhealthy {
			t.Error("Panicked checker should be marked unhealthy")
		}
	}
}

func TestHealthManager_Check_Context_Cancelled(t *testing.T) {
	mgr := health.NewManagerWithPoolSize(1)

	slowChecker := health.NewCustomChecker("slow", func(ctx context.Context) health.HealthCheck {
		time.Sleep(100 * time.Millisecond)
		return health.HealthCheck{Status: health.StatusHealthy}
	})

	mgr.Register(slowChecker)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	report := mgr.Check(ctx)

	// With cancelled context, FanOut short-circuits without processing.
	// All items get zero-value results and are marked as panicked/unhealthy.
	if report.Status != health.StatusUnhealthy {
		t.Errorf("Status = %v, want %v for cancelled context", report.Status, health.StatusUnhealthy)
	}
}

func TestNewManagerWithPoolSize(t *testing.T) {
	mgr := health.NewManagerWithPoolSize(5)
	if mgr.WorkerPoolSize() != 5 {
		t.Errorf("WorkerPoolSize() = %d, want 5", mgr.WorkerPoolSize())
	}
}

func TestNewManagerWithPoolSize_Invalid(t *testing.T) {
	mgr := health.NewManagerWithPoolSize(0)
	if mgr.WorkerPoolSize() != health.DefaultWorkerPoolSize {
		t.Errorf("WorkerPoolSize() = %d, want %d", mgr.WorkerPoolSize(), health.DefaultWorkerPoolSize)
	}

	mgr2 := health.NewManagerWithPoolSize(-1)
	if mgr2.WorkerPoolSize() != health.DefaultWorkerPoolSize {
		t.Errorf("WorkerPoolSize() = %d, want %d", mgr2.WorkerPoolSize(), health.DefaultWorkerPoolSize)
	}
}

func TestCustomChecker_Check(t *testing.T) {
	checker := health.NewCustomChecker("test", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{
			Status:  health.StatusHealthy,
			Message: "Test OK",
		}
	})

	check := checker.Check(context.Background())
	if check.Name != "test" {
		t.Errorf("Name = %s, want test", check.Name)
	}

	if check.Status != health.StatusHealthy {
		t.Errorf("Status = %v, want %v", check.Status, health.StatusHealthy)
	}
}

func TestCustomChecker_Name(t *testing.T) {
	checker := health.NewCustomChecker("my-checker", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{Status: health.StatusHealthy}
	})

	if checker.Name() != "my-checker" {
		t.Errorf("Name() = %s, want my-checker", checker.Name())
	}
}

func TestWithHealthChecker(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}

	checker := health.NewCustomChecker("test", func(ctx context.Context) health.HealthCheck {
		return health.HealthCheck{Status: health.StatusHealthy}
	})

	server, err := NewServer(conf, WithHealthChecker(checker))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	mgr := server.GetHealthManager()
	if mgr == nil {
		t.Fatal("Health manager should be created")
	}

	report := mgr.Check(context.Background())
	if len(report.Checks) != 1 {
		t.Errorf("Check count = %d, want 1", len(report.Checks))
	}
}

func TestRouteShortcuts_WithMiddleware(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test",
		Port:        0,
	}
	server, _ := NewServer(conf)

	handler := func(_ core.Context) error { return nil }
	mw1 := func(ctx core.Context) error { return ctx.Next() }
	mw2 := func(ctx core.Context) error { return ctx.Next() }

	err := server.Protected().
		Middleware(mw1, mw2).
		GET("/admin", handler)

	if err != nil {
		t.Errorf("Protected().Middleware().GET() error = %v", err)
	}
}

func TestHealthCheck_ResponseTime(t *testing.T) {
	mgr := health.NewManager()

	checker := health.NewCustomChecker("slow", func(ctx context.Context) health.HealthCheck {
		time.Sleep(10 * time.Millisecond)
		return health.HealthCheck{Status: health.StatusHealthy}
	})

	mgr.Register(checker)
	report := mgr.Check(context.Background())

	if check, ok := report.Checks["slow"]; ok {
		if check.ResponseTime < 10 {
			t.Errorf("ResponseTime = %d ms, want >= 10 ms", check.ResponseTime)
		}
	} else {
		t.Error("Check 'slow' not found in report")
	}
}
