package contextutil

import (
	"context"
	"testing"
)

// This file intentionally tests nil context behavior for edge cases.
// We use a helper variable to avoid staticcheck SA1012 warnings while still testing nil behavior.
var nilContext context.Context // intentionally nil for testing

func init() {
	// Ensure nilContext is actually nil
	nilContext = nil
}

func TestGetLanguageFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name:  "nil context should return empty string",
			setup: func() context.Context { return nil }, //nolint:staticcheck // intentional nil test
			want:  "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetLanguageFromContext(nil) = %v, want empty string", got)
				}
			},
		},
		{
			name: "language in context should be returned",
			setup: func() context.Context {
				return SetLanguageInContext(context.Background(), "en")
			},
			want: "en",
			check: func(t *testing.T, got string) {
				if got != "en" {
					t.Errorf("GetLanguageFromContext() = %v, want en", got)
				}
			},
		},
		{
			name: "non-string language value should return empty string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), LangKey{}, 123)
			},
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetLanguageFromContext() = %v, want empty string", got)
				}
			},
		},
		{
			name:  "no language in context should return empty string",
			setup: func() context.Context { return context.Background() },
			want:  "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetLanguageFromContext() = %v, want empty string", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetLanguageFromContext(ctx)
			tt.check(t, got)
		})
	}
}

func TestSetLanguageInContext(t *testing.T) {
	tests := []struct {
		name  string
		lang  string
		check func(t *testing.T, ctx context.Context)
	}{
		{
			name: "set language should store in context",
			lang: "vi",
			check: func(t *testing.T, ctx context.Context) {
				if GetLanguageFromContext(ctx) != "vi" {
					t.Error("Language should be set in context")
				}
			},
		},
		{
			name: "set empty language should work",
			lang: "",
			check: func(t *testing.T, ctx context.Context) {
				if GetLanguageFromContext(ctx) != "" {
					t.Error("Empty language should be set correctly")
				}
			},
		},
		{
			name: "nil context should create new background context",
			lang: "en",
			check: func(t *testing.T, ctx context.Context) {
				if ctx == nil {
					t.Error("Context should not be nil")
				}
				if GetLanguageFromContext(ctx) != "en" {
					t.Error("Language should be set in new context")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "nil context should create new background context" {
				ctx = SetLanguageInContext(nilContext, tt.lang)
			} else {
				ctx = SetLanguageInContext(ctx, tt.lang)
			}
			tt.check(t, ctx)
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		check func(t *testing.T, got interface{})
	}{
		{
			name:  "nil context should return nil",
			setup: func() context.Context { return nil }, //nolint:staticcheck // intentional nil test
			check: func(t *testing.T, got interface{}) {
				if got != nil {
					t.Errorf("GetUserIDFromContext(nil) = %v, want nil", got)
				}
			},
		},
		{
			name: "string user ID in context should be returned",
			setup: func() context.Context {
				return SetUserIDInContext(context.Background(), "user123")
			},
			check: func(t *testing.T, got interface{}) {
				userID, ok := got.(string)
				if !ok {
					t.Errorf("GetUserIDFromContext() type = %T, want string", got)
				}
				if userID != "user123" {
					t.Errorf("GetUserIDFromContext() = %v, want user123", userID)
				}
			},
		},
		{
			name: "int64 user ID value should be returned as int64",
			setup: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey{}, int64(12345))
			},
			check: func(t *testing.T, got interface{}) {
				userID, ok := got.(int64)
				if !ok {
					t.Errorf("GetUserIDFromContext() type = %T, want int64", got)
				}
				if userID != 12345 {
					t.Errorf("GetUserIDFromContext() = %v, want 12345", userID)
				}
			},
		},
		{
			name: "int user ID value should be returned as int",
			setup: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey{}, 12345)
			},
			check: func(t *testing.T, got interface{}) {
				userID, ok := got.(int)
				if !ok {
					t.Errorf("GetUserIDFromContext() type = %T, want int", got)
				}
				if userID != 12345 {
					t.Errorf("GetUserIDFromContext() = %v, want 12345", userID)
				}
			},
		},
		{
			name:  "no user ID in context should return nil",
			setup: func() context.Context { return context.Background() },
			check: func(t *testing.T, got interface{}) {
				if got != nil {
					t.Errorf("GetUserIDFromContext() = %v, want nil", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetUserIDFromContext(ctx)
			tt.check(t, got)
		})
	}
}

func TestSetUserIDInContext(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		check  func(t *testing.T, ctx context.Context)
	}{
		{
			name:   "set user ID should store in context",
			userID: "user456",
			check: func(t *testing.T, ctx context.Context) {
				got, ok := GetUserIDFromContext(ctx).(string)
				if !ok || got != "user456" {
					t.Errorf("GetUserIDFromContext() = %v, want user456", got)
				}
			},
		},
		{
			name:   "set empty user ID should work",
			userID: "",
			check: func(t *testing.T, ctx context.Context) {
				got, ok := GetUserIDFromContext(ctx).(string)
				if !ok || got != "" {
					t.Errorf("GetUserIDFromContext() = %v, want empty string", got)
				}
			},
		},
		{
			name:   "nil context should create new background context",
			userID: "user789",
			check: func(t *testing.T, ctx context.Context) {
				if ctx == nil {
					t.Error("Context should not be nil")
				}
				got, ok := GetUserIDFromContext(ctx).(string)
				if !ok || got != "user789" {
					t.Errorf("GetUserIDFromContext() = %v, want user789", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "nil context should create new background context" {
				ctx = SetUserIDInContext(nilContext, tt.userID)
			} else {
				ctx = SetUserIDInContext(ctx, tt.userID)
			}
			tt.check(t, ctx)
		})
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name:  "nil context should return empty string",
			setup: func() context.Context { return nil }, //nolint:staticcheck // intentional nil test
			want:  "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetRequestIDFromContext(nil) = %v, want empty string", got)
				}
			},
		},
		{
			name: "request ID in context should be returned",
			setup: func() context.Context {
				return SetRequestIDInContext(context.Background(), "req-123")
			},
			want: "req-123",
			check: func(t *testing.T, got string) {
				if got != "req-123" {
					t.Errorf("GetRequestIDFromContext() = %v, want req-123", got)
				}
			},
		},
		{
			name:  "no request ID in context should return empty string",
			setup: func() context.Context { return context.Background() },
			want:  "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetRequestIDFromContext() = %v, want empty string", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetRequestIDFromContext(ctx)
			tt.check(t, got)
		})
	}
}

func TestSetRequestIDInContext(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		check     func(t *testing.T, ctx context.Context)
	}{
		{
			name:      "set request ID should store in context",
			requestID: "req-456",
			check: func(t *testing.T, ctx context.Context) {
				if GetRequestIDFromContext(ctx) != "req-456" {
					t.Error("Request ID should be set in context")
				}
			},
		},
		{
			name:      "set empty request ID should work",
			requestID: "",
			check: func(t *testing.T, ctx context.Context) {
				if GetRequestIDFromContext(ctx) != "" {
					t.Error("Empty request ID should be set correctly")
				}
			},
		},
		{
			name:      "nil context should create new background context",
			requestID: "req-789",
			check: func(t *testing.T, ctx context.Context) {
				if ctx == nil {
					t.Error("Context should not be nil")
				}
				if GetRequestIDFromContext(ctx) != "req-789" {
					t.Error("Request ID should be set in new context")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "nil context should create new background context" {
				ctx = SetRequestIDInContext(nilContext, tt.requestID)
			} else {
				ctx = SetRequestIDInContext(ctx, tt.requestID)
			}
			tt.check(t, ctx)
		})
	}
}

func TestGetTraceIDFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name:  "nil context should return empty string",
			setup: func() context.Context { return nil }, //nolint:staticcheck // intentional nil test
			want:  "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetTraceIDFromContext(nil) = %v, want empty string", got)
				}
			},
		},
		{
			name: "trace ID in context should be returned",
			setup: func() context.Context {
				return SetTraceIDInContext(context.Background(), "trace-12345")
			},
			want: "trace-12345",
			check: func(t *testing.T, got string) {
				if got != "trace-12345" {
					t.Errorf("GetTraceIDFromContext() = %v, want trace-12345", got)
				}
			},
		},
		{
			name: "non-string trace ID value should return empty string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TraceIDKey{}, 99999)
			},
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetTraceIDFromContext() = %v, want empty string", got)
				}
			},
		},
		{
			name:  "no trace ID in context should return empty string",
			setup: func() context.Context { return context.Background() },
			want:  "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("GetTraceIDFromContext() = %v, want empty string", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetTraceIDFromContext(ctx)
			tt.check(t, got)
		})
	}
}

func TestSetTraceIDInContext(t *testing.T) {
	tests := []struct {
		name    string
		traceID string
		check   func(t *testing.T, ctx context.Context)
	}{
		{
			name:    "set trace ID should store in context",
			traceID: "trace-67890",
			check: func(t *testing.T, ctx context.Context) {
				if GetTraceIDFromContext(ctx) != "trace-67890" {
					t.Error("Trace ID should be set in context")
				}
			},
		},
		{
			name:    "set empty trace ID should work",
			traceID: "",
			check: func(t *testing.T, ctx context.Context) {
				if GetTraceIDFromContext(ctx) != "" {
					t.Error("Empty trace ID should be set correctly")
				}
			},
		},
		{
			name:    "nil context should create new background context",
			traceID: "trace-11111",
			check: func(t *testing.T, ctx context.Context) {
				if ctx == nil {
					t.Error("Context should not be nil")
				}
				if GetTraceIDFromContext(ctx) != "trace-11111" {
					t.Error("Trace ID should be set in new context")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "nil context should create new background context" {
				ctx = SetTraceIDInContext(nilContext, tt.traceID)
			} else {
				ctx = SetTraceIDInContext(ctx, tt.traceID)
			}
			tt.check(t, ctx)
		})
	}
}

func TestGetAllContextValues(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  map[string]string
		check func(t *testing.T, got map[string]string)
	}{
		{
			name: "nil context should return empty map",
			setup: func() context.Context {
				return nilContext
			},
			want: make(map[string]string),
			check: func(t *testing.T, got map[string]string) {
				if len(got) != 0 {
					t.Errorf("GetAllContextValues(nil) should return empty map, got %v", got)
				}
			},
		},
		{
			name: "context with all predefined values should return all",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = SetRequestIDInContext(ctx, "req-123")
				ctx = SetUserIDInContext(ctx, "user-456")
				ctx = SetLanguageInContext(ctx, "en")
				ctx = SetTraceIDInContext(ctx, "trace-789")
				return ctx
			},
			want: map[string]string{
				KeyRequestID: "req-123",
				KeyUserID:    "user-456",
				KeyLanguage:  "en",
				KeyTraceID:   "trace-789",
			},
			check: func(t *testing.T, got map[string]string) {
				if got[KeyRequestID] != "req-123" {
					t.Errorf("Expected request_id = req-123, got %s", got[KeyRequestID])
				}
				if got[KeyUserID] != "user-456" {
					t.Errorf("Expected user_id = user-456, got %s", got[KeyUserID])
				}
				if got[KeyLanguage] != "en" {
					t.Errorf("Expected lang = en, got %s", got[KeyLanguage])
				}
				if got[KeyTraceID] != "trace-789" {
					t.Errorf("Expected trace_id = trace-789, got %s", got[KeyTraceID])
				}
			},
		},
		{
			name: "context with partial values should return only set values",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = SetRequestIDInContext(ctx, "req-999")
				ctx = SetUserIDInContext(ctx, "user-888")
				return ctx
			},
			want: map[string]string{
				KeyRequestID: "req-999",
				KeyUserID:    "user-888",
			},
			check: func(t *testing.T, got map[string]string) {
				if got[KeyRequestID] != "req-999" {
					t.Errorf("Expected request_id = req-999, got %s", got[KeyRequestID])
				}
				if got[KeyUserID] != "user-888" {
					t.Errorf("Expected user_id = user-888, got %s", got[KeyUserID])
				}
				if got[KeyLanguage] != "" {
					t.Errorf("Expected empty lang, got %s", got[KeyLanguage])
				}
				if got[KeyTraceID] != "" {
					t.Errorf("Expected empty trace_id, got %s", got[KeyTraceID])
				}
			},
		},
		{
			name: "empty context should return empty map",
			setup: func() context.Context {
				return context.Background()
			},
			want: make(map[string]string),
			check: func(t *testing.T, got map[string]string) {
				if len(got) != 0 {
					t.Errorf("GetAllContextValues(empty) should return empty map, got %v", got)
				}
			},
		},
		{
			name: "context with custom values should not return them",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = SetCustomValueInContext(ctx, "custom_key", "custom_value")
				ctx = SetRequestIDInContext(ctx, "req-111")
				return ctx
			},
			want: map[string]string{
				KeyRequestID: "req-111",
			},
			check: func(t *testing.T, got map[string]string) {
				if got[KeyRequestID] != "req-111" {
					t.Errorf("Expected request_id = req-111, got %s", got[KeyRequestID])
				}
				// Custom values are not included because GetAllContextValues only checks predefined keys
				if got["custom_key"] != "" {
					t.Errorf("Custom keys should not be in GetAllContextValues result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetAllContextValues(ctx)
			tt.check(t, got)
		})
	}
}

func TestGetContextKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected interface{}
		check    func(t *testing.T, got interface{})
	}{
		{
			name: "request_id should return RequestIDKey",
			key:  KeyRequestID,
			check: func(t *testing.T, got interface{}) {
				if _, ok := got.(RequestIDKey); !ok {
					t.Errorf("Expected RequestIDKey, got %T", got)
				}
			},
		},
		{
			name: "user_id should return UserIDKey",
			key:  KeyUserID,
			check: func(t *testing.T, got interface{}) {
				if _, ok := got.(UserIDKey); !ok {
					t.Errorf("Expected UserIDKey, got %T", got)
				}
			},
		},
		{
			name: "lang should return LangKey",
			key:  KeyLanguage,
			check: func(t *testing.T, got interface{}) {
				if _, ok := got.(LangKey); !ok {
					t.Errorf("Expected LangKey, got %T", got)
				}
			},
		},
		{
			name: "trace_id should return TraceIDKey",
			key:  KeyTraceID,
			check: func(t *testing.T, got interface{}) {
				if _, ok := got.(TraceIDKey); !ok {
					t.Errorf("Expected TraceIDKey, got %T", got)
				}
			},
		},
		{
			name: "custom key should return CustomKey",
			key:  "custom_key",
			check: func(t *testing.T, got interface{}) {
				ck, ok := got.(CustomKey)
				if !ok {
					t.Errorf("Expected CustomKey, got %T", got)
				}
				if string(ck) != "custom_key" {
					t.Errorf("Expected CustomKey(\"custom_key\"), got %v", ck)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetContextKey(tt.key)
			tt.check(t, got)
		})
	}
}

func TestGetFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		key   string
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name: "nil context should return empty string",
			setup: func() context.Context {
				return nilContext
			},
			key:  "any_key",
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty string for nil context, got %s", got)
				}
			},
		},
		{
			name: "predefined key should return value",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = SetUserIDInContext(ctx, "user123")
				return ctx
			},
			key:  KeyUserID,
			want: "user123",
			check: func(t *testing.T, got string) {
				if got != "user123" {
					t.Errorf("Expected user123, got %s", got)
				}
			},
		},
		{
			name: "custom key should return value",
			setup: func() context.Context {
				ctx := context.Background()
				ctx = SetCustomValueInContext(ctx, "custom_key", "custom_value")
				return ctx
			},
			key:  "custom_key",
			want: "custom_value",
			check: func(t *testing.T, got string) {
				if got != "custom_value" {
					t.Errorf("Expected custom_value, got %s", got)
				}
			},
		},
		{
			name: "non-existent key should return empty string",
			setup: func() context.Context {
				return context.Background()
			},
			key:  "non_existent",
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty string, got %s", got)
				}
			},
		},
		{
			name: "empty context should return empty string",
			setup: func() context.Context {
				return context.Background()
			},
			key:  KeyRequestID,
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty string, got %s", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetFromContext(ctx, tt.key)
			tt.check(t, got)
		})
	}
}
