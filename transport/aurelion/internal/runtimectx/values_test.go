package runtimectx

import (
	"context"
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/internal/keys"
)

func TestGetContextKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		check func(t *testing.T, result interface{})
	}{
		{
			name: "request_id key should return RequestIDKey type",
			key:  keys.ContextKeyRequestID,
			check: func(t *testing.T, result interface{}) {
				if _, ok := result.(RequestIDKey); !ok {
					t.Errorf("Expected RequestIDKey type, got %T", result)
				}
			},
		},
		{
			name: "user_id key should return UserIDKey type",
			key:  keys.ContextKeyUserID,
			check: func(t *testing.T, result interface{}) {
				if _, ok := result.(UserIDKey); !ok {
					t.Errorf("Expected UserIDKey type, got %T", result)
				}
			},
		},
		{
			name: "lang key should return LangKey type",
			key:  keys.ContextKeyLanguage,
			check: func(t *testing.T, result interface{}) {
				if _, ok := result.(LangKey); !ok {
					t.Errorf("Expected LangKey type, got %T", result)
				}
			},
		},
		{
			name: "trace_id key should return TraceIDKey type",
			key:  keys.ContextKeyTraceID,
			check: func(t *testing.T, result interface{}) {
				if _, ok := result.(TraceIDKey); !ok {
					t.Errorf("Expected TraceIDKey type, got %T", result)
				}
			},
		},
		{
			name: "custom key should return CustomKey type",
			key:  "custom_key",
			check: func(t *testing.T, result interface{}) {
				if customKey, ok := result.(CustomKey); !ok {
					t.Errorf("Expected CustomKey type, got %T", result)
				} else if string(customKey) != "custom_key" {
					t.Errorf("Expected custom_key, got %s", customKey)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContextKey(tt.key)
			tt.check(t, result)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		key   string
		want  interface{}
		check func(t *testing.T, result interface{})
	}{
		{
			name: "existing key should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey(KeyRequestID), "req-123")
			},
			key: KeyRequestID,
			check: func(t *testing.T, result interface{}) {
				if result != "req-123" {
					t.Errorf("Get() = %v, want %v", result, "req-123")
				}
			},
		},
		{
			name: "non-existing key should return nil",
			setup: func() context.Context {
				return context.Background()
			},
			key: KeyRequestID,
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("Get() = %v, want nil", result)
				}
			},
		},
		{
			name: "nil context should return nil",
			setup: func() context.Context {
				return nil
			},
			key: KeyRequestID,
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("Get() = %v, want nil", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := Get(ctx, tt.key)
			tt.check(t, result)
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() context.Context
		key          string
		defaultValue string
		want         string
	}{
		{
			name: "existing string value should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey(KeyLanguage), "en")
			},
			key:          KeyLanguage,
			defaultValue: "vi",
			want:         "en",
		},
		{
			name: "non-existing key should return default",
			setup: func() context.Context {
				return context.Background()
			},
			key:          KeyLanguage,
			defaultValue: "vi",
			want:         "vi",
		},
		{
			name: "empty string value should return default",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey(KeyLanguage), "")
			},
			key:          KeyLanguage,
			defaultValue: "vi",
			want:         "vi",
		},
		{
			name: "non-string value should return default",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey(KeyLanguage), 123)
			},
			key:          KeyLanguage,
			defaultValue: "vi",
			want:         "vi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetString(ctx, tt.key, tt.defaultValue)
			if result != tt.want {
				t.Errorf("GetString() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() context.Context
		key          string
		defaultValue int
		want         int
	}{
		{
			name: "existing int value should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey("priority"), 10)
			},
			key:          "priority",
			defaultValue: 5,
			want:         10,
		},
		{
			name: "non-existing key should return default",
			setup: func() context.Context {
				return context.Background()
			},
			key:          "priority",
			defaultValue: 5,
			want:         5,
		},
		{
			name: "non-int value should return default",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey("priority"), "high")
			},
			key:          "priority",
			defaultValue: 5,
			want:         5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetInt(ctx, tt.key, tt.defaultValue)
			if result != tt.want {
				t.Errorf("GetInt() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() context.Context
		key          string
		defaultValue bool
		want         bool
	}{
		{
			name: "existing bool value true should return true",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey("enabled"), true)
			},
			key:          "enabled",
			defaultValue: false,
			want:         true,
		},
		{
			name: "existing bool value false should return false",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey("enabled"), false)
			},
			key:          "enabled",
			defaultValue: true,
			want:         false,
		},
		{
			name: "non-existing key should return default",
			setup: func() context.Context {
				return context.Background()
			},
			key:          "enabled",
			defaultValue: true,
			want:         true,
		},
		{
			name: "non-bool value should return default",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey("enabled"), "yes")
			},
			key:          "enabled",
			defaultValue: false,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetBool(ctx, tt.key, tt.defaultValue)
			if result != tt.want {
				t.Errorf("GetBool() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name  string
		ctx   context.Context
		key   string
		value interface{}
		check func(t *testing.T, result context.Context)
	}{
		{
			name:  "set string value should work",
			ctx:   context.Background(),
			key:   KeyLanguage,
			value: "en",
			check: func(t *testing.T, result context.Context) {
				got := result.Value(GetContextKey(KeyLanguage))
				if got != "en" {
					t.Errorf("Set() stored %v, want %v", got, "en")
				}
			},
		},
		{
			name:  "nil context should create new background context",
			ctx:   nil,
			key:   KeyLanguage,
			value: "en",
			check: func(t *testing.T, result context.Context) {
				if result == nil {
					t.Error("Set() should create new context when given nil")
				}
				got := result.Value(GetContextKey(KeyLanguage))
				if got != "en" {
					t.Errorf("Set() stored %v, want %v", got, "en")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Set(tt.ctx, tt.key, tt.value)
			tt.check(t, result)
		})
	}
}

func TestHas(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		key   string
		want  bool
	}{
		{
			name: "existing key should return true",
			setup: func() context.Context {
				return context.WithValue(context.Background(), GetContextKey(KeyRequestID), "req-123")
			},
			key:  KeyRequestID,
			want: true,
		},
		{
			name: "non-existing key should return false",
			setup: func() context.Context {
				return context.Background()
			},
			key:  KeyRequestID,
			want: false,
		},
		{
			name: "nil context should return false",
			setup: func() context.Context {
				return nil
			},
			key:  KeyRequestID,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := Has(ctx, tt.key)
			if result != tt.want {
				t.Errorf("Has() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetLanguageFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
	}{
		{
			name: "existing language should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), LangKey{}, "en")
			},
			want: "en",
		},
		{
			name: "non-existing language should return empty string",
			setup: func() context.Context {
				return context.Background()
			},
			want: "",
		},
		{
			name: "nil context should return empty string",
			setup: func() context.Context {
				return nil
			},
			want: "",
		},
		{
			name: "non-string value should return empty string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), LangKey{}, 123)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetLanguageFromContext(ctx)
			if result != tt.want {
				t.Errorf("GetLanguageFromContext() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		check func(t *testing.T, result interface{})
	}{
		{
			name: "existing user ID should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey{}, "user-123")
			},
			check: func(t *testing.T, result interface{}) {
				if result != "user-123" {
					t.Errorf("GetUserIDFromContext() = %v, want %v", result, "user-123")
				}
			},
		},
		{
			name: "non-existing user ID should return nil",
			setup: func() context.Context {
				return context.Background()
			},
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("GetUserIDFromContext() = %v, want nil", result)
				}
			},
		},
		{
			name: "nil context should return nil",
			setup: func() context.Context {
				return nil
			},
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("GetUserIDFromContext() = %v, want nil", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetUserIDFromContext(ctx)
			tt.check(t, result)
		})
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
	}{
		{
			name: "existing request ID should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), RequestIDKey{}, "req-123")
			},
			want: "req-123",
		},
		{
			name: "non-existing request ID should return empty string",
			setup: func() context.Context {
				return context.Background()
			},
			want: "",
		},
		{
			name: "nil context should return empty string",
			setup: func() context.Context {
				return nil
			},
			want: "",
		},
		{
			name: "non-string value should return empty string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), RequestIDKey{}, 123)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetRequestIDFromContext(ctx)
			if result != tt.want {
				t.Errorf("GetRequestIDFromContext() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetTraceIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
	}{
		{
			name: "existing trace ID should return value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TraceIDKey{}, "trace-456")
			},
			want: "trace-456",
		},
		{
			name: "non-existing trace ID should return empty string",
			setup: func() context.Context {
				return context.Background()
			},
			want: "",
		},
		{
			name: "nil context should return empty string",
			setup: func() context.Context {
				return nil
			},
			want: "",
		},
		{
			name: "non-string value should return empty string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TraceIDKey{}, 789)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetTraceIDFromContext(ctx)
			if result != tt.want {
				t.Errorf("GetTraceIDFromContext() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetAllContextValues(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		check func(t *testing.T, result map[string]string)
	}{
		{
			name: "nil context should return empty map",
			setup: func() context.Context {
				return nil
			},
			check: func(t *testing.T, result map[string]string) {
				if len(result) != 0 {
					t.Errorf("GetAllContextValues() = %v, want empty map", result)
				}
			},
		},
		{
			name: "context with values should return all values",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, RequestIDKey{}, "req-123")
				ctx = context.WithValue(ctx, TraceIDKey{}, "trace-456")
				ctx = context.WithValue(ctx, LangKey{}, "en")
				ctx = context.WithValue(ctx, UserIDKey{}, "user-789")
				return ctx
			},
			check: func(t *testing.T, result map[string]string) {
				if result[KeyRequestID] != "req-123" {
					t.Errorf("RequestID = %v, want req-123", result[KeyRequestID])
				}
				if result[KeyTraceID] != "trace-456" {
					t.Errorf("TraceID = %v, want trace-456", result[KeyTraceID])
				}
				if result[KeyLanguage] != "en" {
					t.Errorf("Language = %v, want en", result[KeyLanguage])
				}
				if result[KeyUserID] != "user-789" {
					t.Errorf("UserID = %v, want user-789", result[KeyUserID])
				}
			},
		},
		{
			name: "user ID as int should be converted to string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey{}, 123)
			},
			check: func(t *testing.T, result map[string]string) {
				if result[KeyUserID] != "123" {
					t.Errorf("UserID = %v, want 123", result[KeyUserID])
				}
			},
		},
		{
			name: "user ID as int64 should be converted to string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey{}, int64(456))
			},
			check: func(t *testing.T, result map[string]string) {
				if result[KeyUserID] != "456" {
					t.Errorf("UserID = %v, want 456", result[KeyUserID])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetAllContextValues(ctx)
			tt.check(t, result)
		})
	}
}
