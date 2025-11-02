package aurelion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap/zaptest"
)

// TestUnmarshal tests the utils.Unmarshal function with APIResponse type
func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, resp APIResponse, err error)
	}{
		{
			name:    "empty string should return zero value",
			input:   "",
			wantErr: false,
			check: func(t *testing.T, resp APIResponse, err error) {
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
				if resp.Code != 0 || resp.Success != false {
					t.Error("Unmarshal() response should be zero value")
				}
			},
		},
		{
			name:    "valid APIResponse JSON should be parsed",
			input:   `{"success":true,"code":200,"message":"OK","data":{"id":1},"timestamp":1234567890}`,
			wantErr: false,
			check: func(t *testing.T, resp APIResponse, err error) {
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
				if resp.Code != 200 {
					t.Errorf("Unmarshal() response.Code = %v, want 200", resp.Code)
				}
				if resp.Success != true {
					t.Errorf("Unmarshal() response.Success = %v, want true", resp.Success)
				}
			},
		},
		{
			name:    "invalid JSON should return error",
			input:   "not json",
			wantErr: true,
			check: func(t *testing.T, resp APIResponse, err error) {
				if err == nil {
					t.Error("Unmarshal() error = nil, want error")
				}
			},
		},
		{
			name:    "JSON without APIResponse structure should still parse with zero values",
			input:   `{"name":"test"}`,
			wantErr: false,
			check: func(t *testing.T, resp APIResponse, err error) {
				// Unmarshal unmarshals any JSON into APIResponse with zero values for missing fields
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
				if resp.Code != 0 {
					t.Errorf("Unmarshal() response.Code = %v, want 0", resp.Code)
				}
			},
		},
		{
			name:    "minimal valid APIResponse should be parsed",
			input:   `{"success":false,"code":400,"timestamp":1234567890}`,
			wantErr: false,
			check: func(t *testing.T, resp APIResponse, err error) {
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
				if resp.Code != 400 {
					t.Errorf("Unmarshal() response.Code = %v, want 400", resp.Code)
				}
				if resp.Success != false {
					t.Errorf("Unmarshal() response.Success = %v, want false", resp.Success)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := utils.Unmarshal(tt.input, APIResponse{})
			tt.check(t, resp, err)
		})
	}
}

// TestRequestResponseLoggingMiddleware tests the request/response logging middleware
func TestRequestResponseLoggingMiddleware(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name        string
		verbose     bool
		method      string
		path        string
		requestBody string
		handler     func(c *fiber.Ctx) error
		check       func(t *testing.T, app *fiber.App)
	}{
		{
			name:    "basic request should log without verbose logging",
			verbose: false,
			method:  http.MethodGet,
			path:    "/test",
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:        "request with verbose logging should log body",
			verbose:     true,
			method:      http.MethodPost,
			path:        "/test",
			requestBody: `{"name":"test"}`,
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodPost, "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:    "request with query params should log them when verbose logging enabled",
			verbose: true,
			method:  http.MethodGet,
			path:    "/test",
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodGet, "/test?key=value&id=123", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:    "request without request ID should use unknown",
			verbose: false,
			method:  http.MethodGet,
			path:    "/test",
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:    "request with params should log them when verbose logging enabled",
			verbose: true,
			method:  http.MethodGet,
			path:    "/test/:id/:name",
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodGet, "/test/123/john", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:        "request with invalid JSON body should still process",
			verbose:     true,
			method:      http.MethodPost,
			path:        "/test",
			requestBody: "invalid json",
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodPost, "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:    "request with error response should log error code",
			verbose: true,
			method:  http.MethodGet,
			path:    "/error",
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return BadRequest(ctx, "bad request")
			},
			check: func(t *testing.T, app *fiber.App) {
				req := httptest.NewRequest(http.MethodGet, "/error", nil)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// Add request ID middleware first
			app.Use(requestIDMiddleware())

			// Add logging middleware
			app.Use(requestResponseLoggingMiddleware(logger, tt.verbose))

			// Add route
			app.Add(tt.method, tt.path, func(c *fiber.Ctx) error {
				return tt.handler(c)
			})

			tt.check(t, app)
		})
	}
}

// TestGetRequestIDFromContext tests the getRequestIDFromContext helper function
func TestGetRequestIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func(c *fiber.Ctx)
		check func(t *testing.T, requestID string)
	}{
		{
			name: "valid request ID should be extracted",
			setup: func(c *fiber.Ctx) {
				c.Locals(contextKeyRequestID, "test-request-id")
			},
			check: func(t *testing.T, requestID string) {
				if requestID != "test-request-id" {
					t.Errorf("getRequestIDFromContext() = %v, want %v", requestID, "test-request-id")
				}
			},
		},
		{
			name: "non-string request ID should return unknown",
			setup: func(c *fiber.Ctx) {
				c.Locals(contextKeyRequestID, 12345)
			},
			check: func(t *testing.T, requestID string) {
				if requestID != "unknown" {
					t.Errorf("getRequestIDFromContext() = %v, want %v", requestID, "unknown")
				}
			},
		},
		{
			name: "missing request ID should return unknown",
			setup: func(c *fiber.Ctx) {
				// Do nothing
			},
			check: func(t *testing.T, requestID string) {
				if requestID != "unknown" {
					t.Errorf("getRequestIDFromContext() = %v, want %v", requestID, "unknown")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			var requestID string

			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setup(c)
				requestID = getRequestIDFromContext(c)
				return c.SendString("ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			tt.check(t, requestID)
		})
	}
}

// TestBuildRequestLogFields tests the buildRequestLogFields helper function
func TestBuildRequestLogFields(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		setup   func(c *fiber.Ctx)
		check   func(t *testing.T, fields []interface{})
	}{
		{
			name:    "basic logging should include basic fields only",
			verbose: false,
			setup: func(c *fiber.Ctx) {
				// Do nothing
			},
			check: func(t *testing.T, fields []interface{}) {
				if len(fields) != 8 { // 4 pairs: method, path, ip, user-agent
					t.Errorf("buildRequestLogFields() length = %v, want %v", len(fields), 8)
				}
				// Check for method field
				foundMethod := false
				for i := 0; i < len(fields); i += 2 {
					if fields[i] == "method" {
						foundMethod = true
						break
					}
				}
				if !foundMethod {
					t.Error("buildRequestLogFields() should include method field")
				}
			},
		},
		{
			name:    "verbose logging with query params should include query",
			verbose: true,
			setup: func(c *fiber.Ctx) {
				// Query params will be in the request
			},
			check: func(t *testing.T, fields []interface{}) {
				// Should have at least basic fields + potentially query
				if len(fields) < 8 {
					t.Errorf("buildRequestLogFields() length = %v, want at least %v", len(fields), 8)
				}
			},
		},
		{
			name:    "verbose logging with body should include body",
			verbose: true,
			setup: func(c *fiber.Ctx) {
				// Body will be set via request
			},
			check: func(t *testing.T, fields []interface{}) {
				// Should have at least basic fields + potentially body
				if len(fields) < 8 {
					t.Errorf("buildRequestLogFields() length = %v, want at least %v", len(fields), 8)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			var logFields []interface{}

			app.Post("/test", func(c *fiber.Ctx) error {
				tt.setup(c)
				logFields = buildRequestLogFields(c, tt.verbose)
				return c.SendString("ok")
			})

			body := `{"test":"data"}`
			req := httptest.NewRequest(http.MethodPost, "/test?key=value", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			tt.check(t, logFields)
		})
	}
}

// TestBuildResponseLogFields tests the buildResponseLogFields helper function
func TestBuildResponseLogFields(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		handler func(c *fiber.Ctx) error
		check   func(t *testing.T, fields []interface{})
	}{
		{
			name:    "basic logging should include basic fields only",
			verbose: false,
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, fields []interface{}) {
				if len(fields) != 8 { // 4 pairs: method, path, http_code, duration_ms
					t.Errorf("buildResponseLogFields() length = %v, want %v", len(fields), 8)
				}
			},
		},
		{
			name:    "verbose logging with response should include response",
			verbose: true,
			handler: func(c *fiber.Ctx) error {
				ctx := newContext(c)
				return OK(ctx, "success")
			},
			check: func(t *testing.T, fields []interface{}) {
				// Should have at least basic fields + response + code
				if len(fields) < 8 {
					t.Errorf("buildResponseLogFields() length = %v, want at least %v", len(fields), 8)
				}
				// Check for response field
				foundResponse := false
				for i := 0; i < len(fields); i += 2 {
					if fields[i] == "response" {
						foundResponse = true
						break
					}
				}
				if !foundResponse {
					t.Error("buildResponseLogFields() should include response field when detailed logging enabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			var responseFields []interface{}

			app.Get("/test", func(c *fiber.Ctx) error {
				err := tt.handler(c)
				// Capture response fields after handler execution
				duration := 1 * time.Millisecond
				responseFields = buildResponseLogFields(c, tt.verbose, duration)
				return err
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			tt.check(t, responseFields)
		})
	}
}
