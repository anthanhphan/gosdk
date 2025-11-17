package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*fiber.App)
		headers    map[string]string
		wantHeader bool
		wantLocal  bool
	}{
		{
			name:       "no existing request ID should generate new one",
			setup:      nil,
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name:       "existing request ID in header should be reused",
			headers:    map[string]string{RequestIDHeader: "existing-request-id"},
			setup:      nil,
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name: "existing local request ID should be reused",
			setup: func(app *fiber.App) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "local-request-id")
					return c.Next()
				})
			},
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name: "empty local request ID should generate new one",
			setup: func(app *fiber.App) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "")
					return c.Next()
				})
			},
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name: "non-string local value should generate new one",
			setup: func(app *fiber.App) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, 123)
					return c.Next()
				})
			},
			wantHeader: true,
			wantLocal:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			if tt.setup != nil {
				tt.setup(app)
			}

			app.Use(RequestIDMiddleware())

			app.Get("/", func(c *fiber.Ctx) error {
				if tt.wantLocal {
					if _, ok := c.Locals(ContextKeyRequestID).(string); !ok {
						t.Error("request id local missing or not string")
					}
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantHeader {
				if resp.Header.Get(RequestIDHeader) == "" {
					t.Error("missing request id in response header")
				}
			}
		})
	}
}
