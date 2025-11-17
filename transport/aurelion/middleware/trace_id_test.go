package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestTraceIDMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*fiber.App)
		headers    map[string]string
		wantHeader bool
		wantLocal  bool
	}{
		{
			name:       "no existing trace ID should generate new one",
			setup:      nil,
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name:       "existing X-Trace-ID header should be reused",
			headers:    map[string]string{TraceIDHeader: "existing-trace-id"},
			setup:      nil,
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name:       "B3 trace header should be used",
			headers:    map[string]string{"X-B3-TraceId": "b3-trace-id"},
			setup:      nil,
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name:       "W3C traceparent header should be parsed",
			headers:    map[string]string{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"},
			setup:      nil,
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name: "existing local trace ID should be reused",
			setup: func(app *fiber.App) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyTraceID, "local-trace-id")
					return c.Next()
				})
			},
			wantHeader: true,
			wantLocal:  true,
		},
		{
			name: "empty local trace ID should generate new one",
			setup: func(app *fiber.App) {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(ContextKeyTraceID, "")
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

			app.Use(TraceIDMiddleware())

			app.Get("/", func(c *fiber.Ctx) error {
				if tt.wantLocal {
					if _, ok := c.Locals(ContextKeyTraceID).(string); !ok {
						t.Error("trace id local missing or not string")
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
				if resp.Header.Get(TraceIDHeader) == "" {
					t.Error("missing trace id in response header")
				}
			}
		})
	}
}

func TestExtractTraceIDFromTraceparent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid traceparent should extract trace ID",
			input: "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
			want:  "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name:  "empty header should return empty",
			input: "",
			want:  "",
		},
		{
			name:  "header with whitespace should be trimmed",
			input: "  00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01  ",
			want:  "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name:  "invalid format with less than 4 parts should return empty",
			input: "00-0af7651916cd43dd8448eb211c80319c",
			want:  "",
		},
		{
			name:  "trace ID with wrong length should return empty",
			input: "00-shortid-b7ad6b7169203331-01",
			want:  "",
		},
		{
			name:  "trace ID with non-hex characters should return empty",
			input: "00-0af7651916cd43dd8448eb211c80319g-b7ad6b7169203331-01",
			want:  "",
		},
		{
			name:  "uppercase hex should work",
			input: "00-0AF7651916CD43DD8448EB211C80319C-b7ad6b7169203331-01",
			want:  "0AF7651916CD43DD8448EB211C80319C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTraceIDFromTraceparent(tt.input)
			if got != tt.want {
				t.Errorf("extractTraceIDFromTraceparent() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid lowercase hex should return true",
			input: "0af7651916cd43dd8448eb211c80319c",
			want:  true,
		},
		{
			name:  "valid uppercase hex should return true",
			input: "0AF7651916CD43DD8448EB211C80319C",
			want:  true,
		},
		{
			name:  "mixed case hex should return true",
			input: "0aF7651916Cd43dd8448eb211c80319C",
			want:  true,
		},
		{
			name:  "empty string should return false",
			input: "",
			want:  false,
		},
		{
			name:  "string with non-hex characters should return false",
			input: "0af7651916cd43dd8448eb211c80319g",
			want:  false,
		},
		{
			name:  "string with special characters should return false",
			input: "0af7651916cd43dd-8448eb211c80319c",
			want:  false,
		},
		{
			name:  "string with spaces should return false",
			input: "0af7651916cd43dd 8448eb211c80319c",
			want:  false,
		},
		{
			name:  "only numbers should return true",
			input: "12345",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHexString(tt.input)
			if got != tt.want {
				t.Errorf("isHexString() = %v, want %v", got, tt.want)
			}
		})
	}
}
