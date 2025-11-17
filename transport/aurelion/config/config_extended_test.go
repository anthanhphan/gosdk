package config

import (
	"context"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
)

func TestConfig_Validate_Comprehensive(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config should return error",
			config:  nil,
			wantErr: true,
		},
		{
			name: "negative read timeout should return error",
			config: func() *Config {
				neg := -1 * time.Second
				return &Config{
					ServiceName: "Test",
					Port:        8080,
					ReadTimeout: &neg,
				}
			}(),
			wantErr: true,
		},
		{
			name: "negative write timeout should return error",
			config: func() *Config {
				neg := -1 * time.Second
				return &Config{
					ServiceName:  "Test",
					Port:         8080,
					WriteTimeout: &neg,
				}
			}(),
			wantErr: true,
		},
		{
			name: "negative idle timeout should return error",
			config: func() *Config {
				neg := -1 * time.Second
				return &Config{
					ServiceName: "Test",
					Port:        8080,
					IdleTimeout: &neg,
				}
			}(),
			wantErr: true,
		},
		{
			name: "negative shutdown timeout should return error",
			config: func() *Config {
				neg := -1 * time.Second
				return &Config{
					ServiceName:             "Test",
					Port:                    8080,
					GracefulShutdownTimeout: &neg,
				}
			}(),
			wantErr: true,
		},
		{
			name: "CORS enabled with invalid config should return error",
			config: &Config{
				ServiceName: "Test",
				Port:        8080,
				EnableCORS:  true,
				CORS: &CORSConfig{
					AllowOrigins: []string{},
					AllowMethods: []string{"GET"},
				},
			},
			wantErr: true,
		},
		{
			name: "CSRF enabled with invalid config should return error",
			config: &Config{
				ServiceName: "Test",
				Port:        8080,
				EnableCSRF:  true,
				CSRF: &CSRFConfig{
					KeyLookup: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSRFConfig_Validate_AllSources(t *testing.T) {
	tests := []struct {
		name    string
		config  *CSRFConfig
		wantErr bool
	}{
		{
			name: "header source should be valid",
			config: &CSRFConfig{
				KeyLookup: "header:X-CSRF-Token",
			},
			wantErr: false,
		},
		{
			name: "query source should be valid",
			config: &CSRFConfig{
				KeyLookup: "query:csrf",
			},
			wantErr: false,
		},
		{
			name: "param source should be valid",
			config: &CSRFConfig{
				KeyLookup: "param:csrf",
			},
			wantErr: false,
		},
		{
			name: "form source should be valid",
			config: &CSRFConfig{
				KeyLookup: "form:csrf",
			},
			wantErr: false,
		},
		{
			name: "cookie source should be valid",
			config: &CSRFConfig{
				KeyLookup: "cookie:csrf",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCORSConfig_Validate_AllMethods(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range validMethods {
		t.Run("method "+method+" should be valid", func(t *testing.T) {
			config := &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{method},
			}
			if err := config.Validate(); err != nil {
				t.Errorf("Valid method %s should not return error: %v", method, err)
			}
		})
	}
}

func TestConfig_Merge_AllFields(t *testing.T) {
	tests := []struct {
		name  string
		input *Config
		check func(t *testing.T, merged *Config)
	}{
		{
			name: "zero port should be set to default",
			input: &Config{
				ServiceName: "Test",
				Port:        0,
			},
			check: func(t *testing.T, merged *Config) {
				if merged.Port != DefaultPort {
					t.Errorf("Port = %d, want %d", merged.Port, DefaultPort)
				}
			},
		},
		{
			name: "zero max body size should be set to default",
			input: &Config{
				ServiceName: "Test",
				Port:        8080,
				MaxBodySize: 0,
			},
			check: func(t *testing.T, merged *Config) {
				if merged.MaxBodySize != DefaultMaxBodySize {
					t.Errorf("MaxBodySize = %d, want %d", merged.MaxBodySize, DefaultMaxBodySize)
				}
			},
		},
		{
			name: "zero max concurrent connections should be set to default",
			input: &Config{
				ServiceName:              "Test",
				Port:                     8080,
				MaxConcurrentConnections: 0,
			},
			check: func(t *testing.T, merged *Config) {
				if merged.MaxConcurrentConnections != DefaultMaxConcurrentConnections {
					t.Errorf("MaxConcurrentConnections = %d, want %d", merged.MaxConcurrentConnections, DefaultMaxConcurrentConnections)
				}
			},
		},
		{
			name: "empty service name should be set to default",
			input: &Config{
				ServiceName: "",
				Port:        8080,
			},
			check: func(t *testing.T, merged *Config) {
				if merged.ServiceName != "HTTP Server" {
					t.Errorf("ServiceName = %s, want HTTP Server", merged.ServiceName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := tt.input.Merge()
			tt.check(t, merged)
		})
	}
}

// mockContext is a minimal implementation for testing
type mockContext struct {
	locals map[string]interface{}
}

func (m *mockContext) Method() string                                    { return "" }
func (m *mockContext) Path() string                                      { return "" }
func (m *mockContext) OriginalURL() string                               { return "" }
func (m *mockContext) BaseURL() string                                   { return "" }
func (m *mockContext) Protocol() string                                  { return "" }
func (m *mockContext) Hostname() string                                  { return "" }
func (m *mockContext) IP() string                                        { return "" }
func (m *mockContext) Secure() bool                                      { return false }
func (m *mockContext) Get(key string, defaultValue ...string) string     { return "" }
func (m *mockContext) Set(key, value string)                             {}
func (m *mockContext) Append(field string, values ...string)             {}
func (m *mockContext) Params(key string, defaultValue ...string) string  { return "" }
func (m *mockContext) AllParams() map[string]string                      { return nil }
func (m *mockContext) ParamsParser(out interface{}) error                { return nil }
func (m *mockContext) Query(key string, defaultValue ...string) string   { return "" }
func (m *mockContext) AllQueries() map[string]string                     { return nil }
func (m *mockContext) QueryParser(out interface{}) error                 { return nil }
func (m *mockContext) Body() []byte                                      { return nil }
func (m *mockContext) BodyParser(out interface{}) error                  { return nil }
func (m *mockContext) Cookies(key string, defaultValue ...string) string { return "" }
func (m *mockContext) Cookie(cookie *core.Cookie)                        {}
func (m *mockContext) ClearCookie(key ...string)                         {}
func (m *mockContext) Status(status int) core.Context                    { return m }
func (m *mockContext) JSON(data interface{}) error                       { return nil }
func (m *mockContext) XML(data interface{}) error                        { return nil }
func (m *mockContext) SendString(s string) error                         { return nil }
func (m *mockContext) SendBytes(b []byte) error                          { return nil }
func (m *mockContext) Redirect(location string, status ...int) error     { return nil }
func (m *mockContext) Accepts(offers ...string) string                   { return "" }
func (m *mockContext) AcceptsCharsets(offers ...string) string           { return "" }
func (m *mockContext) AcceptsEncodings(offers ...string) string          { return "" }
func (m *mockContext) AcceptsLanguages(offers ...string) string          { return "" }
func (m *mockContext) Fresh() bool                                       { return false }
func (m *mockContext) Stale() bool                                       { return false }
func (m *mockContext) XHR() bool                                         { return false }

func (m *mockContext) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}

func (m *mockContext) GetAllLocals() map[string]interface{} { return m.locals }
func (m *mockContext) Next() error                          { return nil }
func (m *mockContext) Context() context.Context             { return context.Background() }
func (m *mockContext) IsMethod(method string) bool          { return false }
func (m *mockContext) RequestID() string                    { return "" }

func TestStoreInContext_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		ctx    core.Context
		config *Config
	}{
		{
			name:   "nil context should not panic",
			ctx:    nil,
			config: &Config{},
		},
		{
			name:   "nil config should not panic",
			ctx:    &mockContext{locals: make(map[string]interface{})},
			config: nil,
		},
		{
			name:   "both nil should not panic",
			ctx:    nil,
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			StoreInContext(tt.ctx, tt.config)
		})
	}
}

func TestFromContext_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		setup func() core.Context
		check func(t *testing.T, cfg *Config)
	}{
		{
			name: "nil context should return nil",
			setup: func() core.Context {
				return nil
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg != nil {
					t.Error("FromContext(nil) should return nil")
				}
			},
		},
		{
			name: "context without config should return nil",
			setup: func() core.Context {
				return &mockContext{locals: make(map[string]interface{})}
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg != nil {
					t.Error("FromContext() should return nil when not found")
				}
			},
		},
		{
			name: "context with wrong type should return nil",
			setup: func() core.Context {
				ctx := &mockContext{locals: make(map[string]interface{})}
				ctx.locals[ContextKey] = "wrong type"
				return ctx
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg != nil {
					t.Error("FromContext() should return nil for wrong type")
				}
			},
		},
		{
			name: "context with valid config should return config",
			setup: func() core.Context {
				ctx := &mockContext{locals: make(map[string]interface{})}
				ctx.locals[ContextKey] = &Config{ServiceName: "Test", Port: 8080}
				return ctx
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg == nil {
					t.Error("FromContext() should return config")
					return
				}
				if cfg.ServiceName != "Test" {
					t.Errorf("ServiceName = %s, want Test", cfg.ServiceName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			cfg := FromContext(ctx)
			tt.check(t, cfg)
		})
	}
}

func TestCSRFConfig_Validate_AllEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *CSRFConfig
		wantErr bool
	}{
		{
			name:    "empty config should pass",
			config:  &CSRFConfig{},
			wantErr: false,
		},
		{
			name: "positive expiration should pass",
			config: func() *CSRFConfig {
				posExpiration := 1 * time.Hour
				return &CSRFConfig{
					Expiration: &posExpiration,
				}
			}(),
			wantErr: false,
		},
		{
			name: "all valid cookie same site values",
			config: &CSRFConfig{
				CookieSameSite: "Strict",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCORSConfig_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *CORSConfig
		wantErr bool
	}{
		{
			name: "zero max age should pass",
			config: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET"},
				MaxAge:       0,
			},
			wantErr: false,
		},
		{
			name: "positive max age should pass",
			config: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET"},
				MaxAge:       3600,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateCORS_Disabled(t *testing.T) {
	// CORS disabled should skip validation
	cfg := &Config{
		ServiceName: "Test",
		Port:        8080,
		EnableCORS:  false,
		CORS:        nil, // Can be nil when disabled
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("CORS disabled should not require CORS config: %v", err)
	}
}

func TestConfig_ValidateCSRF_Disabled(t *testing.T) {
	// CSRF disabled should skip validation
	cfg := &Config{
		ServiceName: "Test",
		Port:        8080,
		EnableCSRF:  false,
		CSRF:        nil, // Can be nil when disabled
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("CSRF disabled should not require CSRF config: %v", err)
	}
}

func TestConfig_ValidatePositiveTimeouts(t *testing.T) {
	// Test with positive timeouts
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 120 * time.Second
	shutdownTimeout := 30 * time.Second

	cfg := &Config{
		ServiceName:             "Test",
		Port:                    8080,
		ReadTimeout:             &readTimeout,
		WriteTimeout:            &writeTimeout,
		IdleTimeout:             &idleTimeout,
		GracefulShutdownTimeout: &shutdownTimeout,
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Valid timeouts should not return error: %v", err)
	}
}

func TestConfig_ValidateLimits_Zero(t *testing.T) {
	// Zero limits are valid (will be set to defaults by Merge)
	cfg := &Config{
		ServiceName:              "Test",
		Port:                     8080,
		MaxBodySize:              0,
		MaxConcurrentConnections: 0,
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Zero limits should not return error: %v", err)
	}
}

func TestConfig_EdgeCases_AllBranches(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "port at minimum should pass",
			config: &Config{
				ServiceName: "Test",
				Port:        MinPort,
			},
			wantErr: false,
		},
		{
			name: "port at maximum should pass",
			config: &Config{
				ServiceName: "Test",
				Port:        MaxPort,
			},
			wantErr: false,
		},
		{
			name: "CORS enabled with valid and invalid method mix should return error",
			config: &Config{
				ServiceName: "Test",
				Port:        8080,
				EnableCORS:  true,
				CORS: &CORSConfig{
					AllowOrigins: []string{"*"},
					AllowMethods: []string{"GET", "BADMETHOD"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSRFConfig_AllLookupSources_WithValues(t *testing.T) {
	sources := []string{"header", "query", "param", "form", "cookie"}

	for _, source := range sources {
		t.Run(source+" source should be valid", func(t *testing.T) {
			config := &CSRFConfig{
				KeyLookup: source + ":token",
			}
			if err := config.Validate(); err != nil {
				t.Errorf("Valid source %s should not return error: %v", source, err)
			}
		})
	}
}

func TestConfig_ValidateBasicFields_EmptyServiceName(t *testing.T) {
	cfg := &Config{
		ServiceName: "",
		Port:        8080,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("Empty service name should return error")
	}
}

func TestConfig_ValidateBasicFields_ValidPort(t *testing.T) {
	cfg := &Config{
		ServiceName: "Test",
		Port:        3000,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Valid port should not return error: %v", err)
	}
}

func TestCORSConfig_AllowMethods_Empty(t *testing.T) {
	cfg := &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("Empty AllowMethods should return error")
	}
}

func TestConfig_Merge_PreservesNonZeroValues(t *testing.T) {
	cfg := &Config{
		ServiceName:              "Custom",
		Port:                     3000,
		MaxBodySize:              1024,
		MaxConcurrentConnections: 1000,
	}

	merged := cfg.Merge()

	if merged.ServiceName != "Custom" {
		t.Error("Merge should preserve custom ServiceName")
	}
	if merged.Port != 3000 {
		t.Error("Merge should preserve custom Port")
	}
	if merged.MaxBodySize != 1024 {
		t.Error("Merge should preserve custom MaxBodySize")
	}
	if merged.MaxConcurrentConnections != 1000 {
		t.Error("Merge should preserve custom MaxConcurrentConnections")
	}
}

func TestCORSConfig_AllowOrigins_Empty(t *testing.T) {
	cfg := &CORSConfig{
		AllowOrigins: []string{},
		AllowMethods: []string{"GET"},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("Empty AllowOrigins should return error")
	}
}

func TestDefault_ReturnsNewInstance(t *testing.T) {
	cfg1 := Default()
	cfg2 := Default()

	// Should be different instances
	cfg1.ServiceName = "Modified"
	if cfg2.ServiceName == "Modified" {
		t.Error("Default() should return new instance each time")
	}
}
