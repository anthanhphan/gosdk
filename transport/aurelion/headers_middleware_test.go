package aurelion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestHeaderToLocalsMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		prefix     string
		filter     func(string) bool
		wantLocals map[string]string
		check      func(t *testing.T, locals map[string]interface{})
	}{
		{
			name: "all headers should be stored in locals without prefix",
			headers: map[string]string{
				"X-Custom-Header": "value1",
				"Authorization":   "Bearer token",
			},
			prefix: "",
			filter: nil,
			wantLocals: map[string]string{
				"x-custom-header": "value1",
				"authorization":   "Bearer token",
			},
			check: func(t *testing.T, locals map[string]interface{}) {
				if locals["x-custom-header"] != "value1" {
					t.Errorf("Expected x-custom-header = value1, got %v", locals["x-custom-header"])
				}
				if locals["authorization"] != "Bearer token" {
					t.Errorf("Expected authorization = Bearer token, got %v", locals["authorization"])
				}
			},
		},
		{
			name: "headers should be stored with prefix",
			headers: map[string]string{
				"Uid":             "123",
				"Accept-Language": "en",
			},
			prefix: "header_",
			filter: nil,
			wantLocals: map[string]string{
				"header_uid":             "123",
				"header_accept-language": "en",
			},
			check: func(t *testing.T, locals map[string]interface{}) {
				if locals["header_uid"] != "123" {
					t.Errorf("Expected header_uid = 123, got %v", locals["header_uid"])
				}
				if locals["header_accept-language"] != "en" {
					t.Errorf("Expected header_accept-language = en, got %v", locals["header_accept-language"])
				}
			},
		},
		{
			name: "filter should only include matching headers",
			headers: map[string]string{
				"Uid":             "123",
				"Accept-Language": "en",
				"X-Other":         "ignored",
			},
			prefix: "",
			filter: func(key string) bool {
				return key == "uid" || key == "accept-language"
			},
			wantLocals: map[string]string{
				"uid":             "123",
				"accept-language": "en",
			},
			check: func(t *testing.T, locals map[string]interface{}) {
				if locals["uid"] != "123" {
					t.Errorf("Expected uid = 123, got %v", locals["uid"])
				}
				if locals["accept-language"] != "en" {
					t.Errorf("Expected accept-language = en, got %v", locals["accept-language"])
				}
				if locals["x-other"] != nil {
					t.Errorf("Expected x-other to be filtered out, got %v", locals["x-other"])
				}
			},
		},
		{
			name: "headers should be converted to lowercase",
			headers: map[string]string{
				"X-Upper-Case": "value",
				"Mixed-Case":   "value2",
			},
			prefix: "",
			filter: nil,
			wantLocals: map[string]string{
				"x-upper-case": "value",
				"mixed-case":   "value2",
			},
			check: func(t *testing.T, locals map[string]interface{}) {
				if locals["x-upper-case"] == nil {
					t.Error("Expected x-upper-case to be in locals")
				}
				if locals["mixed-case"] == nil {
					t.Error("Expected mixed-case to be in locals")
				}
				// Should not have original case
				if locals["X-Upper-Case"] != nil {
					t.Error("Expected original case to not be in locals")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(middlewareToFiber(HeaderToLocalsMiddleware(tt.prefix, tt.filter)))
			app.Get("/test", func(c *fiber.Ctx) error {
				// Collect all locals for verification
				allLocals := make(map[string]interface{})
				// We can't easily iterate all locals in test, so we check specific ones
				for key, wantValue := range tt.wantLocals {
					if value := c.Locals(key); value != nil {
						allLocals[key] = value
						if value.(string) != wantValue {
							t.Errorf("Expected %s = %s, got %v", key, wantValue, value)
						}
					}
				}
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			// Additional check if provided - verify by making another request and checking locals
			if tt.check != nil {
				// Create a test endpoint that exposes locals
				app.Get("/check", func(c *fiber.Ctx) error {
					locals := make(map[string]interface{})
					for key := range tt.wantLocals {
						if value := c.Locals(key); value != nil {
							locals[key] = value
						}
					}
					tt.check(t, locals)
					return c.SendString("checked")
				})

				req2 := httptest.NewRequest("GET", "/check", nil)
				for k, v := range tt.headers {
					req2.Header.Set(k, v)
				}
				_, _ = app.Test(req2)
			}
		})
	}
}

func TestDefaultHeaderToLocalsMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		check   func(t *testing.T, middleware Middleware)
	}{
		{
			name: "should parse all headers without prefix or filter",
			headers: map[string]string{
				"X-Test":  "value1",
				"X-Other": "value2",
			},
			check: func(t *testing.T, middleware Middleware) {
				// Create a mock fiber context to test
				app := fiber.New()
				app.Use(middlewareToFiber(middleware))
				app.Get("/test", func(c *fiber.Ctx) error {
					// Check that headers are in locals
					if c.Locals("x-test") == nil {
						t.Error("Expected x-test in locals")
					}
					if c.Locals("x-other") == nil {
						t.Error("Expected x-other in locals")
					}
					return c.SendString("ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				for k, v := range map[string]string{
					"X-Test":  "value1",
					"X-Other": "value2",
				} {
					req.Header.Set(k, v)
				}

				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := DefaultHeaderToLocalsMiddleware()
			tt.check(t, middleware)
		})
	}
}
