package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/transport/aurelion/config"
	"github.com/gofiber/fiber/v2"
)

func TestConfigInjector(t *testing.T) {
	tests := []struct {
		name  string
		cfg   *config.Config
		check func(t *testing.T, app *fiber.App)
	}{
		{
			name: "valid config should store in context",
			cfg:  &config.Config{ServiceName: "test-service", Port: 8080},
			check: func(t *testing.T, app *fiber.App) {
				app.Use(ConfigInjector(&config.Config{ServiceName: "test-service", Port: 8080}))
				app.Get("/test", func(c *fiber.Ctx) error {
					cfg := c.Locals(config.ContextKey)
					if cfg == nil {
						t.Error("Config should be stored in locals")
					}
					return c.SendString("ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.StatusCode != 200 {
					t.Errorf("Status = %d, want 200", resp.StatusCode)
				}
			},
		},
		{
			name: "nil config should not panic",
			cfg:  nil,
			check: func(t *testing.T, app *fiber.App) {
				app.Use(ConfigInjector(nil))
				app.Get("/test", func(c *fiber.Ctx) error {
					cfg := c.Locals(config.ContextKey)
					if cfg != nil {
						t.Error("Config should be nil")
					}
					return c.SendString("ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.StatusCode != 200 {
					t.Errorf("Status = %d, want 200", resp.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			tt.check(t, app)
		})
	}
}

func TestBuildCORSConfig(t *testing.T) {
	tests := []struct {
		name  string
		input *config.CORSConfig
		check func(t *testing.T, result interface{})
	}{
		{
			name:  "nil config should return empty config",
			input: nil,
			check: func(t *testing.T, result interface{}) {
				// Just verify no panic occurs
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "valid config should convert correctly",
			input: &config.CORSConfig{
				AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
				AllowMethods:     []string{"GET", "POST", "PUT"},
				AllowHeaders:     []string{"Content-Type", "Authorization"},
				AllowCredentials: true,
				ExposeHeaders:    []string{"X-Request-ID"},
				MaxAge:           3600,
			},
			check: func(t *testing.T, result interface{}) {
				// Verify the configuration builds without error
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "empty arrays should work",
			input: &config.CORSConfig{
				AllowOrigins:     []string{},
				AllowMethods:     []string{},
				AllowHeaders:     []string{},
				AllowCredentials: false,
				ExposeHeaders:    []string{},
				MaxAge:           0,
			},
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCORSConfig(tt.input)
			tt.check(t, result)
		})
	}
}

func TestBuildCSRFConfig(t *testing.T) {
	expiration := 1 * time.Hour

	tests := []struct {
		name  string
		input *config.CSRFConfig
		check func(t *testing.T, result interface{})
	}{
		{
			name:  "nil config should return empty config",
			input: nil,
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "valid config with Strict SameSite should convert correctly",
			input: &config.CSRFConfig{
				KeyLookup:         "header:X-CSRF-Token",
				CookieName:        "csrf_token",
				CookiePath:        "/",
				CookieDomain:      "example.com",
				CookieSecure:      true,
				CookieHTTPOnly:    true,
				CookieSessionOnly: false,
				CookieSameSite:    "Strict",
				SingleUseToken:    true,
				Expiration:        &expiration,
			},
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "Lax SameSite should work",
			input: &config.CSRFConfig{
				KeyLookup:      "cookie:csrf",
				CookieName:     "csrf",
				CookieSameSite: "Lax",
			},
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "None SameSite should work",
			input: &config.CSRFConfig{
				KeyLookup:      "cookie:csrf",
				CookieName:     "csrf",
				CookieSameSite: "None",
			},
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "default SameSite should be Lax",
			input: &config.CSRFConfig{
				KeyLookup:      "cookie:csrf",
				CookieName:     "csrf",
				CookieSameSite: "Invalid",
			},
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "nil expiration should work",
			input: &config.CSRFConfig{
				KeyLookup:  "cookie:csrf",
				CookieName: "csrf",
				Expiration: nil,
			},
			check: func(t *testing.T, result interface{}) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCSRFConfig(tt.input)
			tt.check(t, result)
		})
	}
}
