package aurelion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestGetRequestID(t *testing.T) {
	app := fiber.New()

	app.Use(requestIDMiddleware())

	var receivedID string
	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		receivedID = GetRequestID(ctx)
		return c.SendString("test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if receivedID == "" {
		t.Error("Request ID should not be empty")
	}
}

func TestGetRequestID_WithHeader(t *testing.T) {
	app := fiber.New()

	app.Use(requestIDMiddleware())

	var receivedID string
	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		receivedID = GetRequestID(ctx)
		return c.SendString("test")
	})

	customID := "custom-request-id-12345"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, customID)

	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if receivedID != customID {
		t.Errorf("Request ID = %v, want %v", receivedID, customID)
	}
}

func TestGetRequestID_ResponseHeader(t *testing.T) {
	app := fiber.New()

	app.Use(requestIDMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	responseID := resp.Header.Get(RequestIDHeader)
	if responseID == "" {
		t.Error("Request ID should be in response header")
	}
}

func TestRequestIDMiddleware_Generated(t *testing.T) {
	app := fiber.New()

	app.Use(requestIDMiddleware())

	var receivedID string
	app.Get("/test", func(c *fiber.Ctx) error {
		receivedID = c.Locals(contextKeyRequestID).(string)
		return c.SendString("test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if receivedID == "" {
		t.Error("Request ID should be generated and stored in locals")
	}

	if resp.Header.Get(RequestIDHeader) != receivedID {
		t.Error("Request ID in response header should match locals")
	}
}

func TestRequestIDMiddleware_ReuseExisting(t *testing.T) {
	app := fiber.New()

	app.Use(requestIDMiddleware())

	var receivedID string
	app.Get("/test", func(c *fiber.Ctx) error {
		receivedID = c.Locals(contextKeyRequestID).(string)
		return c.SendString("test")
	})

	customID := "existing-request-id"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, customID)

	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if receivedID != customID {
		t.Errorf("Request ID = %v, want %v", receivedID, customID)
	}
}

func TestGetRequestID_WithMockContext(t *testing.T) {
	// Test with mock context that has no request ID in locals
	mockCtx := &mockContext{}
	rid := GetRequestID(mockCtx)

	if rid != "" {
		t.Errorf("Request ID = %v, want empty string", rid)
	}
}

func TestGetRequestID_WithNonStringLocals(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		// Store non-string value in locals
		c.Locals(contextKeyRequestID, 12345)

		ctx := newContext(c)
		rid := GetRequestID(ctx)

		if rid != "" {
			t.Error("Request ID should be empty when locals has non-string value")
		}

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

// Note: Unmarshal tests are in middleware_test.go
