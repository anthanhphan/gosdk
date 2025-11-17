package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/runtimectx"
	"github.com/gofiber/fiber/v2"
)

func TestHeaderToLocals(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		filter     func(string) bool
		headers    map[string]string
		checkKey   string
		checkValue string
	}{
		{
			name:       "with prefix should store headers",
			prefix:     "hdr_",
			filter:     nil,
			headers:    map[string]string{"X-Custom": "value"},
			checkKey:   "hdr_x-custom",
			checkValue: "value",
		},
		{
			name:   "with filter should skip filtered headers",
			prefix: "hdr_",
			filter: func(key string) bool {
				return key != "x-skip"
			},
			headers:    map[string]string{"X-Custom": "value", "X-Skip": "skip-this"},
			checkKey:   "hdr_x-custom",
			checkValue: "value",
		},
		{
			name:       "multiple headers should be stored",
			prefix:     "h_",
			filter:     nil,
			headers:    map[string]string{"X-Header-1": "value1", "X-Header-2": "value2"},
			checkKey:   "h_x-header-1",
			checkValue: "value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			mw := HeaderToLocals(tt.prefix, tt.filter)

			// Convert middleware.MiddlewareFunc to core.Middleware
			coreMw := func(ctx core.Context) error {
				return mw(adaptMiddlewareContext(ctx))
			}
			app.Use(rctx.MiddlewareToFiber(coreMw))
			app.Get("/", func(c *fiber.Ctx) error {
				ctx := rctx.NewFiberContext(c)
				if got := GetHeader(ctx, tt.checkKey); got != tt.checkValue {
					t.Errorf("GetHeader() = %s, want %s", got, tt.checkValue)
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			if _, err := app.Test(req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultHeaderToLocals(t *testing.T) {
	app := fiber.New()
	mw := DefaultHeaderToLocals()

	// Convert middleware.MiddlewareFunc to core.Middleware
	coreMw := func(ctx core.Context) error {
		return mw(adaptMiddlewareContext(ctx))
	}
	app.Use(rctx.MiddlewareToFiber(coreMw))
	app.Get("/", func(c *fiber.Ctx) error {
		ctx := rctx.NewFiberContext(c)
		if got := GetHeader(ctx, "x-custom-header"); got != "test-value" {
			t.Errorf("GetHeader() = %s, want test-value", got)
		}
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Custom-Header", "test-value")

	if _, err := app.Test(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetHeader(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		localValue   interface{}
		want         string
	}{
		{
			name:         "existing key should return value",
			key:          "test-key",
			defaultValue: nil,
			localValue:   "test-value",
			want:         "test-value",
		},
		{
			name:         "missing key should return default value",
			key:          "missing-key",
			defaultValue: []string{"default-value"},
			localValue:   nil,
			want:         "default-value",
		},
		{
			name:         "missing key without default should return empty",
			key:          "missing-key",
			defaultValue: nil,
			localValue:   nil,
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				ctx := rctx.NewFiberContext(c)
				if tt.localValue != nil {
					ctx.Locals(tt.key, tt.localValue)
				}
				got := GetHeader(ctx, tt.key, tt.defaultValue...)
				if got != tt.want {
					t.Errorf("GetHeader() = %s, want %s", got, tt.want)
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if _, err := app.Test(req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetHeaderInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		localValue   interface{}
		want         int
	}{
		{
			name:         "valid int should return value",
			key:          "int-key",
			defaultValue: 0,
			localValue:   "123",
			want:         123,
		},
		{
			name:         "invalid int should return default",
			key:          "invalid-key",
			defaultValue: 999,
			localValue:   "not-a-number",
			want:         999,
		},
		{
			name:         "missing key should return default",
			key:          "missing-key",
			defaultValue: 42,
			localValue:   nil,
			want:         42,
		},
		{
			name:         "empty string should return default",
			key:          "empty-key",
			defaultValue: 100,
			localValue:   "",
			want:         100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				ctx := rctx.NewFiberContext(c)
				if tt.localValue != nil {
					ctx.Locals(tt.key, tt.localValue)
				}
				got := GetHeaderInt(ctx, tt.key, tt.defaultValue)
				if got != tt.want {
					t.Errorf("GetHeaderInt() = %d, want %d", got, tt.want)
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if _, err := app.Test(req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetHeaderBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		localValue   interface{}
		want         bool
	}{
		{
			name:         "true string should return true",
			key:          "bool-key",
			defaultValue: false,
			localValue:   "true",
			want:         true,
		},
		{
			name:         "false string should return false",
			key:          "bool-key",
			defaultValue: true,
			localValue:   "false",
			want:         false,
		},
		{
			name:         "1 should return true",
			key:          "bool-key",
			defaultValue: false,
			localValue:   "1",
			want:         true,
		},
		{
			name:         "0 should return false",
			key:          "bool-key",
			defaultValue: true,
			localValue:   "0",
			want:         false,
		},
		{
			name:         "yes should return true",
			key:          "bool-key",
			defaultValue: false,
			localValue:   "yes",
			want:         true,
		},
		{
			name:         "no should return false",
			key:          "bool-key",
			defaultValue: true,
			localValue:   "no",
			want:         false,
		},
		{
			name:         "on should return true",
			key:          "bool-key",
			defaultValue: false,
			localValue:   "on",
			want:         true,
		},
		{
			name:         "off should return false",
			key:          "bool-key",
			defaultValue: true,
			localValue:   "off",
			want:         false,
		},
		{
			name:         "invalid bool should return default",
			key:          "invalid-key",
			defaultValue: true,
			localValue:   "not-a-bool",
			want:         true,
		},
		{
			name:         "missing key should return default",
			key:          "missing-key",
			defaultValue: true,
			localValue:   nil,
			want:         true,
		},
		{
			name:         "empty string should return default",
			key:          "empty-key",
			defaultValue: false,
			localValue:   "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				ctx := rctx.NewFiberContext(c)
				if tt.localValue != nil {
					ctx.Locals(tt.key, tt.localValue)
				}
				got := GetHeaderBool(ctx, tt.key, tt.defaultValue)
				if got != tt.want {
					t.Errorf("GetHeaderBool() = %v, want %v", got, tt.want)
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if _, err := app.Test(req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetAllHeaders(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		locals map[string]interface{}
		want   map[string]string
	}{
		{
			name:   "should return all headers with prefix",
			prefix: "hdr_",
			locals: map[string]interface{}{
				"hdr_x-custom-1": "value1",
				"hdr_x-custom-2": "value2",
				"other_key":      "other-value",
			},
			want: map[string]string{
				"x-custom-1": "value1",
				"x-custom-2": "value2",
			},
		},
		{
			name:   "empty locals should return empty map",
			prefix: "hdr_",
			locals: map[string]interface{}{},
			want:   map[string]string{},
		},
		{
			name:   "no matching prefix should return empty map",
			prefix: "hdr_",
			locals: map[string]interface{}{
				"other_key1": "value1",
				"other_key2": "value2",
			},
			want: map[string]string{},
		},
		{
			name:   "no prefix should return all string values",
			prefix: "",
			locals: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": 123,
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				ctx := rctx.NewFiberContext(c)
				for key, value := range tt.locals {
					ctx.Locals(key, value)
				}
				result := GetAllHeaders(ctx, tt.prefix)
				if len(result) != len(tt.want) {
					t.Errorf("GetAllHeaders() length = %d, want %d", len(result), len(tt.want))
				}
				for key, value := range tt.want {
					if result[key] != value {
						t.Errorf("result[%s] = %s, want %s", key, result[key], value)
					}
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if _, err := app.Test(req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
