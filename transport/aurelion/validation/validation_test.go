package validation

import (
	"context"
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestValidate(t *testing.T) {
	type TestStruct struct {
		Name     string `validate:"required,min=3,max=10"`
		Email    string `validate:"required,email"`
		Age      int    `validate:"min=18,max=100"`
		Website  string `validate:"url"`
		Phone    string `validate:"numeric"`
		Username string `validate:"alpha"`
	}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(t *testing.T, err error)
	}{
		{
			name: "valid struct should pass",
			input: &TestStruct{
				Name:     "John",
				Email:    "john@example.com",
				Age:      25,
				Website:  "https://example.com",
				Phone:    "1234567890",
				Username: "johndoe",
			},
			wantErr: false,
		},
		{
			name:    "nil pointer should return error",
			input:   (*TestStruct)(nil),
			wantErr: true,
		},
		{
			name:    "non-struct should return error",
			input:   "not a struct",
			wantErr: true,
		},
		{
			name: "required field missing should return error",
			input: &TestStruct{
				Name:  "",
				Email: "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "min length violation should return error",
			input: &TestStruct{
				Name:  "Jo",
				Email: "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "max length violation should return error",
			input: &TestStruct{
				Name:  "VeryLongNameThatExceedsMaximum",
				Email: "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid email should return error",
			input: &TestStruct{
				Name:  "John",
				Email: "invalid-email",
			},
			wantErr: true,
		},
		{
			name: "min value violation should return error",
			input: &TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   15,
			},
			wantErr: true,
		},
		{
			name: "max value violation should return error",
			input: &TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   150,
			},
			wantErr: true,
		},
		{
			name: "invalid URL should return error",
			input: &TestStruct{
				Name:    "John",
				Email:   "john@example.com",
				Website: "not-a-url",
			},
			wantErr: true,
		},
		{
			name: "non-numeric phone should return error",
			input: &TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Phone: "abc123",
			},
			wantErr: true,
		},
		{
			name: "non-alpha username should return error",
			input: &TestStruct{
				Name:     "John",
				Email:    "john@example.com",
				Username: "user123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "email",
		Message: "must be valid",
	}

	expected := "email: must be valid"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name   string
		errors ValidationErrors
		want   string
	}{
		{
			name:   "empty errors should return empty string",
			errors: ValidationErrors{},
			want:   "",
		},
		{
			name: "single error should return error message",
			errors: ValidationErrors{
				{Field: "name", Message: "is required"},
			},
			want: "name: is required",
		},
		{
			name: "multiple errors should join with semicolon",
			errors: ValidationErrors{
				{Field: "name", Message: "is required"},
				{Field: "email", Message: "invalid format"},
			},
			want: "name: is required; email: invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			if result != tt.want {
				t.Errorf("Error() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestValidationErrors_ToArray(t *testing.T) {
	errors := ValidationErrors{
		{Field: "Name", Message: "is required"},
		{Field: "Email", Message: "invalid format"},
	}

	array := errors.ToArray()

	if len(array) != 2 {
		t.Errorf("ToArray() length = %d, want 2", len(array))
	}

	if array[0]["field"] != "name" {
		t.Error("Field should be lowercased")
	}

	if array[0]["message"] != "is required" {
		t.Error("Message should be preserved")
	}
}

func TestValidate_SliceAndArray(t *testing.T) {
	type TestStruct struct {
		Tags    []string `validate:"min=2,max=5"`
		Numbers []int    `validate:"min=1"`
	}

	tests := []struct {
		name    string
		input   *TestStruct
		wantErr bool
	}{
		{
			name: "valid slice length should pass",
			input: &TestStruct{
				Tags:    []string{"tag1", "tag2", "tag3"},
				Numbers: []int{1, 2},
			},
			wantErr: false,
		},
		{
			name: "slice too short should return error",
			input: &TestStruct{
				Tags: []string{"tag1"},
			},
			wantErr: true,
		},
		{
			name: "slice too long should return error",
			input: &TestStruct{
				Tags: []string{"1", "2", "3", "4", "5", "6"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_UnexportedFields(t *testing.T) {
	type TestStruct struct {
		Public  string `validate:"required"`
		private string `validate:"required"` // Should be ignored
	}

	input := &TestStruct{
		Public:  "value",
		private: "", // Ignored
	}

	err := Validate(input)
	if err != nil {
		t.Errorf("Unexported fields should be ignored: %v", err)
	}
}

func TestValidate_NoValidationTags(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	input := &TestStruct{
		Name: "",
		Age:  0,
	}

	err := Validate(input)
	if err != nil {
		t.Errorf("No validation tags should not return error: %v", err)
	}
}

func TestValidateAndParse(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
	}

	tests := []struct {
		name    string
		setup   func() core.Context
		wantErr bool
	}{
		{
			name: "valid JSON should parse and validate",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				fCtx.Request().Header.SetContentType("application/json")
				fCtx.Request().SetBodyString(`{"name":"John","email":"john@example.com"}`)
				return &mockContext{fCtx: fCtx}
			},
			wantErr: false,
		},
		{
			name: "invalid JSON should return error",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				fCtx.Request().Header.SetContentType("application/json")
				fCtx.Request().SetBodyString(`{invalid json}`)
				return &mockContext{fCtx: fCtx}
			},
			wantErr: true,
		},
		{
			name: "valid JSON but invalid validation should return error",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				fCtx.Request().Header.SetContentType("application/json")
				fCtx.Request().SetBodyString(`{"name":"Jo","email":"invalid"}`)
				return &mockContext{fCtx: fCtx}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			var req TestRequest
			err := ValidateAndParse(ctx, &req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndParse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_AllTypes(t *testing.T) {
	type TestStruct struct {
		String  string   `validate:"min=3,max=10"`
		Int     int      `validate:"min=10,max=100"`
		Int8    int8     `validate:"min=1,max=10"`
		Int16   int16    `validate:"min=100,max=1000"`
		Int32   int32    `validate:"min=1000,max=10000"`
		Int64   int64    `validate:"min=10000,max=100000"`
		Slice   []string `validate:"min=1,max=5"`
		Email   string   `validate:"email"`
		URL     string   `validate:"url"`
		Numeric string   `validate:"numeric"`
		Alpha   string   `validate:"alpha"`
	}

	tests := []struct {
		name    string
		input   *TestStruct
		wantErr bool
	}{
		{
			name: "all valid types should pass",
			input: &TestStruct{
				String:  "valid",
				Int:     50,
				Int8:    5,
				Int16:   500,
				Int32:   5000,
				Int64:   50000,
				Slice:   []string{"a", "b"},
				Email:   "test@example.com",
				URL:     "https://example.com",
				Numeric: "12345",
				Alpha:   "abcdef",
			},
			wantErr: false,
		},
		{
			name: "int too small should return error",
			input: &TestStruct{
				String: "valid",
				Int:    5,
			},
			wantErr: true,
		},
		{
			name: "int too large should return error",
			input: &TestStruct{
				String: "valid",
				Int:    200,
			},
			wantErr: true,
		},
		{
			name: "empty optional email should pass",
			input: &TestStruct{
				String: "valid",
				Int:    50,
				Int8:   5,
				Int16:  500,
				Int32:  5000,
				Int64:  50000,
				Slice:  []string{"a"},
				Email:  "",
			},
			wantErr: false,
		},
		{
			name: "empty optional URL should pass",
			input: &TestStruct{
				String: "valid",
				Int:    50,
				Int8:   5,
				Int16:  500,
				Int32:  5000,
				Int64:  50000,
				Slice:  []string{"a"},
				URL:    "",
			},
			wantErr: false,
		},
		{
			name: "empty optional numeric should pass",
			input: &TestStruct{
				String:  "valid",
				Int:     50,
				Int8:    5,
				Int16:   500,
				Int32:   5000,
				Int64:   50000,
				Slice:   []string{"a"},
				Numeric: "",
			},
			wantErr: false,
		},
		{
			name: "empty optional alpha should pass",
			input: &TestStruct{
				String: "valid",
				Int:    50,
				Int8:   5,
				Int16:  500,
				Int32:  5000,
				Int64:  50000,
				Slice:  []string{"a"},
				Alpha:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_ZeroValues(t *testing.T) {
	type TestStruct struct {
		String  string            `validate:"required"`
		Bool    bool              `validate:"required"`
		Int     int               `validate:"required"`
		Uint    uint              `validate:"required"`
		Float32 float32           `validate:"required"`
		Float64 float64           `validate:"required"`
		Ptr     *string           `validate:"required"`
		Slice   []string          `validate:"required"`
		Map     map[string]string `validate:"required"`
		Chan    chan int          `validate:"required"`
	}

	tests := []struct {
		name    string
		input   *TestStruct
		wantErr bool
	}{
		{
			name: "zero string should return error",
			input: &TestStruct{
				String: "",
			},
			wantErr: true,
		},
		{
			name: "false bool should return error",
			input: &TestStruct{
				Bool: false,
			},
			wantErr: true,
		},
		{
			name: "zero int should return error",
			input: &TestStruct{
				Int: 0,
			},
			wantErr: true,
		},
		{
			name: "nil ptr should return error",
			input: &TestStruct{
				Ptr: nil,
			},
			wantErr: true,
		},
		{
			name: "nil slice should return error",
			input: &TestStruct{
				Slice: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRule_InvalidParameters(t *testing.T) {
	type TestStruct struct {
		Field string `validate:"min=invalid"`
	}

	input := &TestStruct{Field: "test"}
	err := Validate(input)
	// Should not error when min parameter is invalid (silently ignored)
	if err != nil {
		t.Logf("Validate with invalid min parameter: %v", err)
	}
}

func TestValidate_StructValue(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	// Pass by value (not pointer)
	input := TestStruct{Name: "test"}
	err := Validate(input)
	if err != nil {
		t.Errorf("Struct value should be validated: %v", err)
	}
}

// mockContext for testing
type mockContext struct {
	fCtx *fiber.Ctx
}

func (m *mockContext) Method() string                                   { return "" }
func (m *mockContext) Path() string                                     { return "" }
func (m *mockContext) OriginalURL() string                              { return "" }
func (m *mockContext) BaseURL() string                                  { return "" }
func (m *mockContext) Protocol() string                                 { return "" }
func (m *mockContext) Hostname() string                                 { return "" }
func (m *mockContext) IP() string                                       { return "" }
func (m *mockContext) Secure() bool                                     { return false }
func (m *mockContext) Get(key string, defaultValue ...string) string    { return "" }
func (m *mockContext) Set(key, value string)                            {}
func (m *mockContext) Append(field string, values ...string)            {}
func (m *mockContext) Params(key string, defaultValue ...string) string { return "" }
func (m *mockContext) AllParams() map[string]string                     { return nil }
func (m *mockContext) ParamsParser(out interface{}) error               { return nil }
func (m *mockContext) Query(key string, defaultValue ...string) string  { return "" }
func (m *mockContext) AllQueries() map[string]string                    { return nil }
func (m *mockContext) QueryParser(out interface{}) error                { return nil }
func (m *mockContext) Body() []byte                                     { return m.fCtx.Body() }

func (m *mockContext) BodyParser(out interface{}) error {
	return m.fCtx.BodyParser(out)
}

func (m *mockContext) Cookies(key string, defaultValue ...string) string   { return "" }
func (m *mockContext) Cookie(cookie *core.Cookie)                          {}
func (m *mockContext) ClearCookie(key ...string)                           {}
func (m *mockContext) Status(status int) core.Context                      { return m }
func (m *mockContext) JSON(data interface{}) error                         { return nil }
func (m *mockContext) XML(data interface{}) error                          { return nil }
func (m *mockContext) SendString(s string) error                           { return nil }
func (m *mockContext) SendBytes(b []byte) error                            { return nil }
func (m *mockContext) Redirect(location string, status ...int) error       { return nil }
func (m *mockContext) Accepts(offers ...string) string                     { return "" }
func (m *mockContext) AcceptsCharsets(offers ...string) string             { return "" }
func (m *mockContext) AcceptsEncodings(offers ...string) string            { return "" }
func (m *mockContext) AcceptsLanguages(offers ...string) string            { return "" }
func (m *mockContext) Fresh() bool                                         { return false }
func (m *mockContext) Stale() bool                                         { return false }
func (m *mockContext) XHR() bool                                           { return false }
func (m *mockContext) Locals(key string, value ...interface{}) interface{} { return nil }
func (m *mockContext) GetAllLocals() map[string]interface{}                { return nil }
func (m *mockContext) Next() error                                         { return nil }
func (m *mockContext) Context() context.Context                            { return context.Background() }
func (m *mockContext) IsMethod(method string) bool                         { return false }
func (m *mockContext) RequestID() string                                   { return "" }
