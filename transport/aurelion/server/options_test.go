package server

import (
	"net/http/httptest"
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/config"
	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/anthanhphan/gosdk/transport/aurelion/router"
	"github.com/gofiber/fiber/v2"
)

func TestWithRateLimiterAcceptsFiberHandler(t *testing.T) {
	cfg := &config.Config{ServiceName: "test", Port: 8081}

	called := false
	limiter := func(c *fiber.Ctx) error {
		called = true
		return c.Next()
	}

	srv, err := NewHttpServer(cfg, WithRateLimiter(limiter))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.AddRoutes(router.NewRoute("/ping").GET().Handler(func(ctx core.Context) error {
		return ctx.Status(200).SendString("pong")
	}))

	resp, err := srv.App().Test(httptest.NewRequest("GET", "/ping", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	if !called {
		t.Fatal("custom rate limiter was not invoked")
	}
}

func TestWithAuthentication(t *testing.T) {
	cfg := &config.Config{ServiceName: "test", Port: 8082}

	authCalled := false
	auth := func(ctx core.Context) error {
		authCalled = true
		return ctx.Next()
	}

	srv, err := NewHttpServer(cfg, WithAuthentication(auth))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.AddRoutes(router.NewRoute("/secure").GET().Protected().Handler(func(ctx core.Context) error {
		if !authCalled {
			t.Fatal("authentication middleware was not invoked")
		}
		return ctx.Status(200).SendString("ok")
	}))

	resp, err := srv.App().Test(httptest.NewRequest("GET", "/secure", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestWithGlobalMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		middlewares []interface{}
		wantErr     bool
	}{
		{
			name: "fiber.Handler should be added",
			middlewares: []interface{}{
				func(c *fiber.Ctx) error { return c.Next() },
			},
			wantErr: false,
		},
		{
			name: "core.Middleware should be added",
			middlewares: []interface{}{
				core.Middleware(func(ctx core.Context) error { return ctx.Next() }),
			},
			wantErr: false,
		},
		{
			name: "multiple middlewares should be added",
			middlewares: []interface{}{
				func(c *fiber.Ctx) error { return c.Next() },
				core.Middleware(func(ctx core.Context) error { return ctx.Next() }),
			},
			wantErr: false,
		},
		{
			name: "nil middleware should be ignored",
			middlewares: []interface{}{
				nil,
			},
			wantErr: false,
		},
		{
			name: "invalid middleware type should return error",
			middlewares: []interface{}{
				"invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{ServiceName: "test", Port: 8083}
			_, err := NewHttpServer(cfg, WithGlobalMiddleware(tt.middlewares...))
			if (err != nil) != tt.wantErr {
				t.Errorf("WithGlobalMiddleware() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithPanicRecover(t *testing.T) {
	tests := []struct {
		name    string
		handler interface{}
		wantErr bool
	}{
		{
			name:    "fiber.Handler should be accepted",
			handler: func(c *fiber.Ctx) error { return c.Next() },
			wantErr: false,
		},
		{
			name:    "core.Middleware should be accepted",
			handler: core.Middleware(func(ctx core.Context) error { return ctx.Next() }),
			wantErr: false,
		},
		{
			name:    "nil handler should be accepted",
			handler: nil,
			wantErr: false,
		},
		{
			name:    "invalid handler type should return error",
			handler: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{ServiceName: "test", Port: 8084}
			_, err := NewHttpServer(cfg, WithPanicRecover(tt.handler))
			if (err != nil) != tt.wantErr {
				t.Errorf("WithPanicRecover() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithAuthorization(t *testing.T) {
	cfg := &config.Config{ServiceName: "test", Port: 8085}

	authzCalled := false
	authz := func(ctx core.Context, roles []string) error {
		authzCalled = true
		return nil
	}

	srv, err := NewHttpServer(cfg, WithAuthorization(authz))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.AddRoutes(router.NewRoute("/admin").GET().Permissions("admin").Handler(func(ctx core.Context) error {
		return ctx.Status(200).SendString("ok")
	}))

	resp, err := srv.App().Test(httptest.NewRequest("GET", "/admin", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	if !authzCalled {
		t.Fatal("authorization checker was not invoked")
	}
}

func TestWithRateLimiterNil(t *testing.T) {
	cfg := &config.Config{ServiceName: "test", Port: 8086}
	srv, err := NewHttpServer(cfg, WithRateLimiter(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("server should not be nil")
	}
}

func TestWithRateLimiterInvalidType(t *testing.T) {
	cfg := &config.Config{ServiceName: "test", Port: 8087}
	_, err := NewHttpServer(cfg, WithRateLimiter("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid rate limiter type")
	}
}

func TestNormalizeFiberHandler(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantNil bool
		wantErr bool
	}{
		{
			name:    "nil should return nil handler",
			input:   nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "fiber.Handler should be returned as-is",
			input:   func(c *fiber.Ctx) error { return c.Next() },
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "core.Middleware should be converted",
			input:   core.Middleware(func(ctx core.Context) error { return ctx.Next() }),
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "unsupported type should return error",
			input:   "invalid",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := normalizeFiberHandler(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeFiberHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (handler == nil) != tt.wantNil {
				t.Errorf("normalizeFiberHandler() handler = %v, wantNil %v", handler, tt.wantNil)
			}
		})
	}
}
