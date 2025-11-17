package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRequestResponseLogger(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		setupApp func(app *fiber.App, logger *zap.SugaredLogger, verbose bool)
		request  func() *http.Request
		wantCode int
	}{
		{
			name:    "non-verbose mode should log basic info",
			verbose: false,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "test-request-id")
					return c.Next()
				})
				app.Use(RequestResponseLogger(logger, verbose))
				app.Get("/test", func(c *fiber.Ctx) error {
					return c.SendString("ok")
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			wantCode: 200,
		},
		{
			name:    "verbose mode should log detailed info",
			verbose: true,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "test-request-id-verbose")
					return c.Next()
				})
				app.Use(RequestResponseLogger(logger, verbose))
				app.Get("/users/:id", func(c *fiber.Ctx) error {
					return c.JSON(map[string]interface{}{
						"code":    200,
						"message": "success",
					})
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/users/123?filter=active", nil)
			},
			wantCode: 200,
		},
		{
			name:    "verbose mode with POST body should log body",
			verbose: true,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "test-post-request")
					return c.Next()
				})
				app.Use(RequestResponseLogger(logger, verbose))
				app.Post("/users", func(c *fiber.Ctx) error {
					return c.JSON(map[string]interface{}{
						"code":    201,
						"message": "created",
					})
				})
			},
			request: func() *http.Request {
				body := bytes.NewBufferString(`{"name":"John","email":"john@example.com"}`)
				req := httptest.NewRequest("POST", "/users", body)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: 200,
		},
		{
			name:    "verbose mode with invalid JSON body should still work",
			verbose: true,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "test-invalid-json")
					return c.Next()
				})
				app.Use(RequestResponseLogger(logger, verbose))
				app.Post("/data", func(c *fiber.Ctx) error {
					return c.SendString("processed")
				})
			},
			request: func() *http.Request {
				body := bytes.NewBufferString("not valid json")
				return httptest.NewRequest("POST", "/data", body)
			},
			wantCode: 200,
		},
		{
			name:    "verbose mode without request ID should still work",
			verbose: true,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(RequestResponseLogger(logger, verbose))
				app.Get("/health", func(c *fiber.Ctx) error {
					return c.SendString("healthy")
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/health", nil)
			},
			wantCode: 200,
		},
		{
			name:    "verbose mode with empty response body",
			verbose: true,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "test-empty-response")
					return c.Next()
				})
				app.Use(RequestResponseLogger(logger, verbose))
				app.Get("/empty", func(c *fiber.Ctx) error {
					return c.SendStatus(204)
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/empty", nil)
			},
			wantCode: 204,
		},
		{
			name:    "verbose mode with non-JSON response",
			verbose: true,
			setupApp: func(app *fiber.App, logger *zap.SugaredLogger, verbose bool) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyRequestID, "test-non-json")
					return c.Next()
				})
				app.Use(RequestResponseLogger(logger, verbose))
				app.Get("/text", func(c *fiber.Ctx) error {
					return c.SendString("plain text response")
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/text", nil)
			},
			wantCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test logger
			var buf bytes.Buffer
			encoderConfig := zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(&buf),
				zapcore.InfoLevel,
			)
			logger := zap.New(core).Sugar()

			// Create Fiber app
			app := fiber.New()
			tt.setupApp(app, logger, tt.verbose)

			// Test request
			req := tt.request()
			resp, err := app.Test(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("Status = %d, want %d", resp.StatusCode, tt.wantCode)
			}

			// Verify logs were written
			if buf.Len() == 0 {
				t.Error("Expected logs to be written")
			}
		})
	}
}
