// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/gofiber/fiber/v3"
)

func newTestConf() *configuration.Config {
	return &configuration.Config{
		ServiceName: "test",
		Port:        8080,
	}
}

func TestContextAdapter_AcquireRelease(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		if ctx == nil {
			t.Fatal("AcquireContextAdapter() returned nil")
		}
		ReleaseContextAdapter(ctx)
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestContextAdapter_RequestInfo(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/users/:id", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		defer ReleaseContextAdapter(ctx)

		if ctx.Method() != "GET" {
			t.Errorf("Method() = %v, want GET", ctx.Method())
		}
		if ctx.Path() != "/users/123" {
			t.Errorf("Path() = %v, want /users/123", ctx.Path())
		}
		if ctx.Params("id") != "123" {
			t.Errorf("Params(id) = %v, want 123", ctx.Params("id"))
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestContextAdapter_Locals(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		defer ReleaseContextAdapter(ctx)

		ctx.Locals("key1", "value1")

		if got := ctx.Locals("key1"); got != "value1" {
			t.Errorf("Locals(key1) = %v, want value1", got)
		}
		if got := ctx.Locals("nonexistent"); got != nil {
			t.Errorf("Locals(nonexistent) = %v, want nil", got)
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

func TestContextAdapter_Headers(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		defer ReleaseContextAdapter(ctx)

		if got := ctx.Get("X-Custom-Header"); got != "test-value" {
			t.Errorf("Get(X-Custom-Header) = %v, want test-value", got)
		}

		ctx.Set("X-Response-Header", "response-value")
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-Header", "test-value")
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

func TestContextAdapter_QueryParams(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		defer ReleaseContextAdapter(ctx)

		if got := ctx.Query("page"); got != "1" {
			t.Errorf("Query(page) = %v, want 1", got)
		}
		if got := ctx.Query("missing"); got != "" {
			t.Errorf("Query(missing) = %v, want empty", got)
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?page=1", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

func TestContextAdapter_SetContext(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		defer ReleaseContextAdapter(ctx)

		origCtx := ctx.Context()
		if origCtx == nil {
			t.Fatal("Context() should not be nil")
		}

		// SetContext should update the cached context
		ctx.SetContext(origCtx)
		newCtx := ctx.Context()
		if newCtx == nil {
			t.Fatal("Context() after SetContext() should not be nil")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

func TestContextAdapter_StatusAndJSON(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := AcquireContextAdapter(c, conf)
		defer ReleaseContextAdapter(ctx)

		return ctx.Status(200).JSON(map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestWithContextAdapter_Reuse(t *testing.T) {
	app := fiber.New()
	conf := newTestConf()

	adapterCount := 0

	app.Use(func(c fiber.Ctx) error {
		return withContextAdapter(c, conf, func(ctx *ContextAdapter) error {
			adapterCount++
			ctx.Locals("middleware_ran", true)
			return c.Next()
		})
	})

	app.Get("/test", func(c fiber.Ctx) error {
		return withContextAdapter(c, conf, func(ctx *ContextAdapter) error {
			adapterCount++
			if got := ctx.Locals("middleware_ran"); got != true {
				t.Error("expected middleware_ran to be true (adapter should be reused)")
			}
			return ctx.Status(200).JSON(map[string]string{"ok": "true"})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if adapterCount != 2 {
		t.Errorf("adapter invocations = %d, want 2", adapterCount)
	}
}
