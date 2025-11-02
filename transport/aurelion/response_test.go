package aurelion

import (
	"net/http"
	"testing"
)

// mockContext implements Context interface for testing
type mockContext struct {
	statusCode  int
	response    interface{}
	headerSet   map[string]string
	errorReturn error
	returnEmpty bool
}

func (m *mockContext) Method() string                                    { return "GET" }
func (m *mockContext) Path() string                                      { return "/test" }
func (m *mockContext) OriginalURL() string                               { return "/test" }
func (m *mockContext) BaseURL() string                                   { return "http://localhost:8080" }
func (m *mockContext) Protocol() string                                  { return "http" }
func (m *mockContext) Hostname() string                                  { return "localhost" }
func (m *mockContext) IP() string                                        { return "127.0.0.1" }
func (m *mockContext) Secure() bool                                      { return false }
func (m *mockContext) Get(key string, defaultValue ...string) string     { return "" }
func (m *mockContext) Set(key, value string)                             { m.headerSet[key] = value }
func (m *mockContext) Append(field string, values ...string)             {}
func (m *mockContext) Params(key string, defaultValue ...string) string  { return "" }
func (m *mockContext) AllParams() map[string]string                      { return map[string]string{} }
func (m *mockContext) ParamsParser(out interface{}) error                { return nil }
func (m *mockContext) Query(key string, defaultValue ...string) string   { return "" }
func (m *mockContext) AllQueries() map[string]string                     { return map[string]string{} }
func (m *mockContext) QueryParser(out interface{}) error                 { return nil }
func (m *mockContext) Body() []byte                                      { return nil }
func (m *mockContext) BodyParser(out interface{}) error                  { return nil }
func (m *mockContext) Cookies(key string, defaultValue ...string) string { return "" }
func (m *mockContext) Cookie(cookie *Cookie)                             {}
func (m *mockContext) ClearCookie(key ...string)                         {}
func (m *mockContext) Status(status int) Context                         { m.statusCode = status; return m }
func (m *mockContext) JSON(data interface{}) error                       { m.response = data; return m.errorReturn }
func (m *mockContext) XML(data interface{}) error                        { return m.errorReturn }
func (m *mockContext) SendString(s string) error                         { return m.errorReturn }
func (m *mockContext) SendBytes(b []byte) error                          { return m.errorReturn }
func (m *mockContext) Redirect(location string, status ...int) error     { return m.errorReturn }
func (m *mockContext) Accepts(offers ...string) string                   { return "" }
func (m *mockContext) AcceptsCharsets(offers ...string) string           { return "" }
func (m *mockContext) AcceptsEncodings(offers ...string) string          { return "" }
func (m *mockContext) AcceptsLanguages(offers ...string) string          { return "" }
func (m *mockContext) Fresh() bool                                       { return true }
func (m *mockContext) Stale() bool                                       { return false }
func (m *mockContext) XHR() bool                                         { return false }
func (m *mockContext) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		return value[0]
	}
	return nil
}
func (m *mockContext) Next() error { return nil }
func (m *mockContext) Context() interface{} {
	if m.returnEmpty {
		return "non-fiber-context"
	}
	return nil
}

func newMockContext() *mockContext {
	return &mockContext{
		headerSet: make(map[string]string),
	}
}

func TestOK(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    []interface{}
		wantErr bool
	}{
		{
			name:    "OK without data should succeed",
			message: "Success",
			data:    nil,
			wantErr: false,
		},
		{
			name:    "OK with data should succeed",
			message: "Success with data",
			data:    []interface{}{map[string]string{"key": "value"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockContext()
			err := OK(ctx, tt.message, tt.data...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
			}

			if !tt.wantErr && ctx.statusCode != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, ctx.statusCode)
			}
		})
	}
}

func TestOK_NilContext(t *testing.T) {
	err := OK(nil, "test")
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "BusinessError should return formatted error",
			err:     NewError(1001, "Test error"),
			wantErr: false,
		},
		{
			name:    "Generic error should return internal server error",
			err:     &mockError{msg: "Generic error"},
			wantErr: false,
		},
		{
			name:    "nil error should return internal server error",
			err:     nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockContext()
			err := Error(ctx, tt.err)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestBusinessError_Error(t *testing.T) {
	err := NewError(1001, "Test error message")
	expected := "[1001] Test error message"

	if err.Error() != expected {
		t.Errorf("BusinessError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestNewError(t *testing.T) {
	err := NewError(1001, "Test error")

	if err.Code != 1001 {
		t.Errorf("Expected Code to be 1001, got %d", err.Code)
	}

	if err.Message != "Test error" {
		t.Errorf("Expected Message to be 'Test error', got %s", err.Message)
	}
}

func TestNewErrorf(t *testing.T) {
	err := NewErrorf(1002, "User %s not found", "john")

	if err.Code != 1002 {
		t.Errorf("Expected Code to be 1002, got %d", err.Code)
	}

	expected := "User john not found"
	if err.Message != expected {
		t.Errorf("Expected Message to be '%s', got '%s'", expected, err.Message)
	}
}

func TestBadRequest(t *testing.T) {
	ctx := newMockContext()
	err := BadRequest(ctx, "Invalid request")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ctx.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, ctx.statusCode)
	}
}

func TestUnauthorized(t *testing.T) {
	ctx := newMockContext()
	err := Unauthorized(ctx, "Unauthorized")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ctx.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, ctx.statusCode)
	}
}

func TestForbidden(t *testing.T) {
	ctx := newMockContext()
	err := Forbidden(ctx, "Forbidden")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ctx.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, ctx.statusCode)
	}
}

func TestNotFound(t *testing.T) {
	ctx := newMockContext()
	err := NotFound(ctx, "Not found")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ctx.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, ctx.statusCode)
	}
}

func TestInternalServerError(t *testing.T) {
	ctx := newMockContext()
	err := InternalServerError(ctx, "Internal error")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ctx.statusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, ctx.statusCode)
	}
}

func TestBadRequest_NilContext(t *testing.T) {
	err := BadRequest(nil, "test")
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestUnauthorized_NilContext(t *testing.T) {
	err := Unauthorized(nil, "test")
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestForbidden_NilContext(t *testing.T) {
	err := Forbidden(nil, "test")
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestNotFound_NilContext(t *testing.T) {
	err := NotFound(nil, "test")
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestInternalServerError_NilContext(t *testing.T) {
	err := InternalServerError(nil, "test")
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestError_NilContext(t *testing.T) {
	err := Error(nil, NewError(1001, "test"))
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

// mockError is a simple error type for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestHealthCheck(t *testing.T) {
	ctx := newMockContext()
	err := HealthCheck(ctx)

	if err != nil {
		t.Errorf("HealthCheck should not return error, got: %v", err)
	}

	if ctx.statusCode != http.StatusOK {
		t.Errorf("HealthCheck should return status 200, got: %d", ctx.statusCode)
	}

	response, ok := ctx.response.(APIResponse)
	if !ok {
		t.Errorf("HealthCheck should return APIResponse, got: %T", ctx.response)
	}

	if response.Success != true {
		t.Errorf("HealthCheck should return success=true, got: %v", response.Success)
	}

	if response.Code != http.StatusOK {
		t.Errorf("HealthCheck should return code=200, got: %d", response.Code)
	}

	if response.Message != "Server is healthy" {
		t.Errorf("HealthCheck should return correct message, got: %s", response.Message)
	}
}

func TestHealthCheck_NilContext(t *testing.T) {
	err := HealthCheck(nil)
	if err == nil {
		t.Error("HealthCheck with nil context should return error")
	}
}
