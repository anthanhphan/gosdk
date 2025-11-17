package runtimectx

import (
	"context"
	"fmt"

	"github.com/anthanhphan/gosdk/transport/aurelion/internal/keys"
)

// Typed keys used when storing values in context.Context to avoid collisions.
type (
	LangKey      struct{}
	UserIDKey    struct{}
	RequestIDKey struct{}
	TraceIDKey   struct{}
	CustomKey    string
)

// Well-known context keys - re-export from keys package for consistency.
const (
	KeyLanguage  = keys.ContextKeyLanguage
	KeyUserID    = keys.ContextKeyUserID
	KeyRequestID = keys.ContextKeyRequestID
	KeyTraceID   = keys.ContextKeyTraceID
)

// GetContextKey maps a string key to a typed key for use with context.WithValue.
func GetContextKey(key string) interface{} {
	switch key {
	case KeyRequestID:
		return RequestIDKey{}
	case KeyUserID:
		return UserIDKey{}
	case KeyLanguage:
		return LangKey{}
	case KeyTraceID:
		return TraceIDKey{}
	default:
		return CustomKey(key)
	}
}

// Get retrieves a value of any type from context by string key.
func Get(ctx context.Context, key string) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(GetContextKey(key))
}

// GetString retrieves a string value with default.
func GetString(ctx context.Context, key, defaultValue string) string {
	if value, ok := Get(ctx, key).(string); ok && value != "" {
		return value
	}
	return defaultValue
}

// GetInt retrieves an integer value from context.
func GetInt(ctx context.Context, key string, defaultValue int) int {
	if value, ok := Get(ctx, key).(int); ok {
		return value
	}
	return defaultValue
}

// GetBool retrieves a boolean value from context.
func GetBool(ctx context.Context, key string, defaultValue bool) bool {
	if value, ok := Get(ctx, key).(bool); ok {
		return value
	}
	return defaultValue
}

// Set stores a value of any type in context.
func Set(ctx context.Context, key string, value interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, GetContextKey(key), value)
}

// Has checks if a key exists in context.
func Has(ctx context.Context, key string) bool {
	if ctx == nil {
		return false
	}
	return ctx.Value(GetContextKey(key)) != nil
}

// GetLanguageFromContext retrieves the language code from context.
func GetLanguageFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if lang, ok := ctx.Value(LangKey{}).(string); ok {
		return lang
	}
	return ""
}

// GetUserIDFromContext retrieves the user identifier from context.
func GetUserIDFromContext(ctx context.Context) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(UserIDKey{})
}

// GetRequestIDFromContext retrieves the request ID stored in context.
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetTraceIDFromContext retrieves the trace ID stored in context.
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetAllContextValues collects well-known values from context into a map.
func GetAllContextValues(ctx context.Context) map[string]string {
	result := make(map[string]string)
	if ctx == nil {
		return result
	}
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		result[KeyRequestID] = requestID
	}
	if userID := GetUserIDFromContext(ctx); userID != nil {
		var userIDStr string
		switch v := userID.(type) {
		case string:
			userIDStr = v
		case int64:
			userIDStr = fmt.Sprintf("%d", v)
		case int:
			userIDStr = fmt.Sprintf("%d", v)
		default:
			userIDStr = fmt.Sprintf("%v", v)
		}
		if userIDStr != "" {
			result[KeyUserID] = userIDStr
		}
	}
	if lang := GetLanguageFromContext(ctx); lang != "" {
		result[KeyLanguage] = lang
	}
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		result[KeyTraceID] = traceID
	}
	return result
}
