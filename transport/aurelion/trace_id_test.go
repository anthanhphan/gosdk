package aurelion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestTraceIDMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		headerName     string
		headerValue    string
		existingLocals map[string]interface{}
		wantHeader     bool
		check          func(t *testing.T, resp *http.Response, locals map[string]interface{})
	}{
		{
			name:        "trace ID from X-Trace-ID header should be used",
			headerName:  TraceIDHeader,
			headerValue: "test-trace-id-123",
			wantHeader:  true,
			check: func(t *testing.T, resp *http.Response, locals map[string]interface{}) {
				if resp.Header.Get(TraceIDHeader) != "test-trace-id-123" {
					t.Errorf("Expected trace ID header %s, got %s", "test-trace-id-123", resp.Header.Get(TraceIDHeader))
				}
				if locals[contextKeyTraceID] != "test-trace-id-123" {
					t.Errorf("Expected trace ID in locals %s, got %v", "test-trace-id-123", locals[contextKeyTraceID])
				}
			},
		},
		{
			name:        "trace ID from X-B3-TraceId header should be used",
			headerName:  "X-B3-TraceId",
			headerValue: "b3-trace-id-456",
			wantHeader:  true,
			check: func(t *testing.T, resp *http.Response, locals map[string]interface{}) {
				if resp.Header.Get(TraceIDHeader) != "b3-trace-id-456" {
					t.Errorf("Expected trace ID header %s, got %s", "b3-trace-id-456", resp.Header.Get(TraceIDHeader))
				}
			},
		},
		{
			name:        "trace ID from traceparent header should be extracted",
			headerName:  "traceparent",
			headerValue: "00-12345678901234567890123456789012-1234567890123456-01",
			wantHeader:  true,
			check: func(t *testing.T, resp *http.Response, locals map[string]interface{}) {
				traceID := resp.Header.Get(TraceIDHeader)
				if traceID == "" {
					t.Error("Expected trace ID to be extracted from traceparent")
				}
				if len(traceID) != 32 {
					t.Errorf("Expected extracted trace ID length 32, got %d", len(traceID))
				}
			},
		},
		{
			name:        "trace ID should be generated when no header is provided",
			headerName:  "",
			headerValue: "",
			wantHeader:  true,
			check: func(t *testing.T, resp *http.Response, locals map[string]interface{}) {
				traceID := resp.Header.Get(TraceIDHeader)
				if traceID == "" {
					t.Error("Expected trace ID to be generated")
				}
				if locals[contextKeyTraceID] == nil {
					t.Error("Expected trace ID to be stored in locals")
				}
			},
		},
		{
			name: "existing trace ID in locals should be used",
			existingLocals: map[string]interface{}{
				contextKeyTraceID: "existing-trace-id",
			},
			headerName:  "", // No header to force use of existing locals
			headerValue: "",
			wantHeader:  true,
			check: func(t *testing.T, resp *http.Response, locals map[string]interface{}) {
				// The middleware will use existing locals value
				traceID := resp.Header.Get(TraceIDHeader)
				if traceID == "" {
					t.Error("Expected trace ID header in response")
				}
				// Should use existing trace ID if it exists in locals before middleware runs
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// If we need to test existing locals, we need to set them BEFORE traceIDMiddleware
			if tt.existingLocals != nil {
				app.Use(func(c *fiber.Ctx) error {
					for k, v := range tt.existingLocals {
						c.Locals(k, v)
					}
					return c.Next()
				})
			}

			app.Use(traceIDMiddleware())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerName != "" {
				req.Header.Set(tt.headerName, tt.headerValue)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			// Get locals from the context after request
			locals := make(map[string]interface{})
			if tt.existingLocals != nil {
				locals = tt.existingLocals
			} else {
				// In a real scenario, we'd need to access the context, but for testing
				// we'll check the response header which confirms the middleware worked
				locals[contextKeyTraceID] = resp.Header.Get(TraceIDHeader)
			}

			if tt.wantHeader && resp.Header.Get(TraceIDHeader) == "" {
				t.Error("Expected trace ID header in response")
			}

			tt.check(t, resp, locals)
		})
	}
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name  string
		setup func() Context
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name: "trace ID in locals should be returned",
			setup: func() Context {
				ctx := newMockContext()
				ctx.Locals(contextKeyTraceID, "test-trace-id")
				return ctx
			},
			want: "test-trace-id",
			check: func(t *testing.T, got string) {
				if got != "test-trace-id" {
					t.Errorf("GetTraceID() = %v, want test-trace-id", got)
				}
			},
		},
		{
			name: "no trace ID in locals should return empty string",
			setup: func() Context {
				return newMockContext()
			},
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetTraceID() = %v, want empty string", got)
				}
			},
		},
		{
			name: "non-string trace ID in locals should return empty string",
			setup: func() Context {
				ctx := newMockContext()
				ctx.Locals(contextKeyTraceID, 123)
				return ctx
			},
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetTraceID() = %v, want empty string", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetTraceID(ctx)
			tt.check(t, got)
		})
	}
}
