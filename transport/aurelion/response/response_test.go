package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/config"
	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/runtimectx"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestDetermineHTTPStatusWithProperCodes(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{ServiceName: "test", Port: 8080, UseProperHTTPStatus: true}
	app.Get("/", func(c *fiber.Ctx) error {
		ctx := rctx.NewFiberContext(c)
		config.StoreInContext(ctx, cfg)
		return BadRequest(ctx, "invalid")
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestDetermineHTTPStatusWithLegacyBehaviour(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{ServiceName: "test", Port: 8080, UseProperHTTPStatus: false}
	app.Get("/", func(c *fiber.Ctx) error {
		ctx := rctx.NewFiberContext(c)
		config.StoreInContext(ctx, cfg)
		return BadRequest(ctx, "invalid")
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestOK(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    interface{}
		check   func(t *testing.T, ctx core.Context)
	}{
		{
			name:    "OK with message and data should work",
			message: "Success",
			data:    map[string]string{"key": "value"},
			check: func(t *testing.T, ctx core.Context) {
				// Response sent successfully
			},
		},
		{
			name:    "OK with only message should work",
			message: "Success",
			data:    nil,
			check: func(t *testing.T, ctx core.Context) {
				// Response sent successfully
			},
		},
		{
			name:    "OK with nil context should return error",
			message: "Success",
			data:    nil,
			check: func(t *testing.T, ctx core.Context) {
				// Will error due to nil context
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx core.Context
			if tt.name != "OK with nil context should return error" {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fCtx)
				ctx = rctx.NewFiberContext(fCtx)
			}

			err := OK(ctx, tt.message, tt.data)
			if tt.name == "OK with nil context should return error" {
				if err == nil {
					t.Error("OK with nil context should return error")
				}
			}
		})
	}
}

func TestError_Responses(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(core.Context) error
		wantErr bool
	}{
		{
			name: "nil context should return error",
			fn: func(ctx core.Context) error {
				return BadRequest(nil, "test")
			},
			wantErr: true,
		},
		{
			name: "BadRequest should work",
			fn: func(ctx core.Context) error {
				return BadRequest(ctx, "bad request")
			},
			wantErr: false,
		},
		{
			name: "Unauthorized should work",
			fn: func(ctx core.Context) error {
				return Unauthorized(ctx, "unauthorized")
			},
			wantErr: false,
		},
		{
			name: "Forbidden should work",
			fn: func(ctx core.Context) error {
				return Forbidden(ctx, "forbidden")
			},
			wantErr: false,
		},
		{
			name: "NotFound should work",
			fn: func(ctx core.Context) error {
				return NotFound(ctx, "not found")
			},
			wantErr: false,
		},
		{
			name: "InternalServerError should work",
			fn: func(ctx core.Context) error {
				return InternalServerError(ctx, "server error")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx core.Context
			if tt.name != "nil context should return error" {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fCtx)
				ctx = rctx.NewFiberContext(fCtx)
			}

			err := tt.fn(ctx)
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

func TestBusinessError(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		msg   string
		check func(t *testing.T, err *BusinessError)
	}{
		{
			name: "NewError should create business error",
			code: 1001,
			msg:  "User not found",
			check: func(t *testing.T, err *BusinessError) {
				if err.Code != 1001 {
					t.Errorf("Code = %d, want 1001", err.Code)
				}
				if err.Message != "User not found" {
					t.Errorf("Message = %s, want User not found", err.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.msg)
			tt.check(t, err)
		})
	}
}

func TestBusinessError_Error(t *testing.T) {
	err := &BusinessError{
		Code:    1001,
		Message: "Test error",
	}

	expected := "[1001] Test error"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestBusinessError_Is(t *testing.T) {
	err1 := &BusinessError{Code: 1001, Message: "Error 1"}
	err2 := &BusinessError{Code: 1001, Message: "Error 1"}
	err3 := &BusinessError{Code: 1002, Message: "Error 2"}

	tests := []struct {
		name   string
		err    *BusinessError
		target error
		want   bool
	}{
		{
			name:   "same error should return true",
			err:    err1,
			target: err2,
			want:   true,
		},
		{
			name:   "different error should return false",
			err:    err1,
			target: err3,
			want:   false,
		},
		{
			name:   "nil target should return false",
			err:    err1,
			target: nil,
			want:   false,
		},
		{
			name:   "non-BusinessError target should return false",
			err:    err1,
			target: errors.New("other error"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			if result != tt.want {
				t.Errorf("Is() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNewErrorf(t *testing.T) {
	err := NewErrorf(1001, "User %s not found", "john")

	if err.Code != 1001 {
		t.Errorf("Code = %d, want 1001", err.Code)
	}

	expected := "User john not found"
	if err.Message != expected {
		t.Errorf("Message = %s, want %s", err.Message, expected)
	}
}

func TestError_WithBusinessError(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)
	bizErr := NewError(1001, "Custom error")

	err := Error(ctx, bizErr)
	if err != nil {
		t.Errorf("Error() should not return error: %v", err)
	}
}

func TestError_WithNilError(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)

	err := Error(ctx, nil)
	// Should return InternalServerError for nil error
	if err != nil {
		t.Errorf("Error(nil) should handle gracefully: %v", err)
	}
}

func TestError_WithRegularError(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)
	regularErr := errors.New("regular error")

	err := Error(ctx, regularErr)
	if err != nil {
		t.Errorf("Error() should not return error: %v", err)
	}
}

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)

	err := HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}
}

func TestHealthCheck_NilContext(t *testing.T) {
	err := HealthCheck(nil)
	if err == nil {
		t.Error("HealthCheck(nil) should return error")
	}
}

func TestErrorWithDetails(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)

	errorData := &ErrorData{
		Type: ErrorTypeValidation,
		Validation: []map[string]string{
			{"field": "email", "message": "invalid"},
		},
	}

	err := ErrorWithDetails(ctx, 400, "Validation failed", errorData)
	if err != nil {
		t.Errorf("ErrorWithDetails() error = %v", err)
	}
}

func TestErrorWithDetails_NilContext(t *testing.T) {
	errorData := &ErrorData{Type: ErrorTypeValidation}
	err := ErrorWithDetails(nil, 400, "test", errorData)
	if err == nil {
		t.Error("ErrorWithDetails(nil) should return error")
	}
}

func TestDetermineHTTPStatus_WithProperStatus(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)
	cfg := &config.Config{UseProperHTTPStatus: true}
	config.StoreInContext(ctx, cfg)

	// Should return proper status code
	status := determineHTTPStatus(ctx, 404)
	if status != 404 {
		t.Errorf("determineHTTPStatus() = %d, want 404", status)
	}
}

func TestDetermineHTTPStatus_LegacyMode(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)
	cfg := &config.Config{UseProperHTTPStatus: false}
	config.StoreInContext(ctx, cfg)

	// Should return 200 for non-5xx errors
	status := determineHTTPStatus(ctx, 404)
	if status != 200 {
		t.Errorf("determineHTTPStatus() = %d, want 200", status)
	}

	// Should return 500+ as-is
	status = determineHTTPStatus(ctx, 500)
	if status != 500 {
		t.Errorf("determineHTTPStatus() for 500 = %d, want 500", status)
	}
}

func TestDetermineHTTPStatus_NoConfig(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := rctx.NewFiberContext(fCtx)
	// No config stored

	// Should default to legacy behavior (200 for non-5xx)
	status := determineHTTPStatus(ctx, 404)
	if status != 200 {
		t.Errorf("determineHTTPStatus() without config = %d, want 200", status)
	}
}

func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     core.Context
		wantErr bool
	}{
		{
			name:    "nil context should return error",
			ctx:     nil,
			wantErr: true,
		},
		{
			name: "valid context should not return error",
			ctx: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return rctx.NewFiberContext(fCtx)
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildAPIResponse(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		code    int
		message string
		data    interface{}
		check   func(t *testing.T, resp APIResponse)
	}{
		{
			name:    "response with data should include data",
			success: true,
			code:    200,
			message: "Success",
			data:    map[string]string{"key": "value"},
			check: func(t *testing.T, resp APIResponse) {
				if !resp.Success {
					t.Error("Success should be true")
				}
				if resp.Code != 200 {
					t.Error("Code should be 200")
				}
				if resp.Data == nil {
					t.Error("Data should not be nil")
				}
			},
		},
		{
			name:    "response without data should have nil data",
			success: false,
			code:    400,
			message: "Error",
			data:    nil,
			check: func(t *testing.T, resp APIResponse) {
				if resp.Success {
					t.Error("Success should be false")
				}
				if resp.Data != nil {
					t.Error("Data should be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp APIResponse
			if tt.data != nil {
				resp = buildAPIResponse(tt.success, tt.code, tt.message, tt.data)
			} else {
				resp = buildAPIResponse(tt.success, tt.code, tt.message)
			}
			tt.check(t, resp)
		})
	}
}

func TestAllResponseHelpers(t *testing.T) {
	cfg := &config.Config{ServiceName: "test", Port: 8080, UseProperHTTPStatus: true}

	tests := []struct {
		name     string
		handler  func(core.Context) error
		wantCode int
	}{
		{
			name: "OK should return 200",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return OK(ctx, "success", nil)
			},
			wantCode: 200,
		},
		{
			name: "BadRequest should return 400",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return BadRequest(ctx, "bad request")
			},
			wantCode: 400,
		},
		{
			name: "Unauthorized should return 401",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return Unauthorized(ctx, "unauthorized")
			},
			wantCode: 401,
		},
		{
			name: "Forbidden should return 403",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return Forbidden(ctx, "forbidden")
			},
			wantCode: 403,
		},
		{
			name: "NotFound should return 404",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return NotFound(ctx, "not found")
			},
			wantCode: 404,
		},
		{
			name: "InternalServerError should return 500",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return InternalServerError(ctx, "server error")
			},
			wantCode: 500,
		},
		{
			name: "HealthCheck should return 200",
			handler: func(ctx core.Context) error {
				config.StoreInContext(ctx, cfg)
				return HealthCheck(ctx)
			},
			wantCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh app instance for each test case
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				ctx := rctx.NewFiberContext(c)
				return tt.handler(ctx)
			})

			resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
			if err != nil {
				t.Errorf("Test request error: %v", err)
			}

			if resp.StatusCode != tt.wantCode {
				t.Errorf("Status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestError_WithNilContext(t *testing.T) {
	err := Error(nil, NewError(1001, "test"))
	if err == nil {
		t.Error("Error(nil) should return error")
	}
}
