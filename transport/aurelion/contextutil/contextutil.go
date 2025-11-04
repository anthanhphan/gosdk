package contextutil

import (
	"context"
	"fmt"
)

// Context key types for storing values in request context
// Using typed keys prevents key collisions
type (
	// LangKey is the typed key for language in context
	LangKey struct{}
	// UserIDKey is the typed key for user ID in context
	UserIDKey struct{}
	// RequestIDKey is the typed key for request ID in context
	RequestIDKey struct{}
	// TraceIDKey is the typed key for trace ID in context
	TraceIDKey struct{}
	// CustomKey is the typed key for custom string keys in context
	CustomKey string
)

// Context key constants for storing values in request context
// These are for backward compatibility and documentation purposes
const (
	// KeyLanguage is the context key for language
	KeyLanguage = "lang"
	// KeyUserID is the context key for user ID
	KeyUserID = "user_id"
	// KeyRequestID is the context key for request ID
	KeyRequestID = "request_id"
	// KeyTraceID is the context key for trace ID
	KeyTraceID = "trace_id"
)

// GetContextKey returns the appropriate typed key for a given string key.
// This maps string keys to typed keys for better type safety.
// If the key is not in the predefined set, it returns a CustomKey.
//
// Input:
//   - key: The string key from Locals
//
// Output:
//   - interface{}: The typed key for use with context.WithValue
//
// Example:
//
//	typedKey := contextutil.GetContextKey("user_id")
//	ctx = context.WithValue(ctx, typedKey, value)
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
//
// Input:
//   - ctx: The context.Context
//   - key: The string key to look up
//
// Output:
//   - interface{}: The value, or nil if not found
//
// Example:
//
//	value := contextutil.Get(ctx, "custom_key")
//	if user, ok := value.(*User); ok {
//	    // Use user
//	}
func Get(ctx context.Context, key string) interface{} {
	if ctx == nil {
		return nil
	}

	typedKey := GetContextKey(key)
	return ctx.Value(typedKey)
}

// GetFromContext retrieves a string value from context.
// Deprecated: Use GetString(ctx, key, "") instead for better API.
func GetFromContext(ctx context.Context, key string) string {
	return GetString(ctx, key, "")
}

// GetString retrieves a string value with default.
//
// Input:
//   - ctx: The context.Context
//   - key: The string key
//   - defaultValue: Default if not found
//
// Output:
//   - string: The value or default
//
// Example:
//
//	lang := contextutil.GetString(ctx, "lang", "en")
func GetString(ctx context.Context, key string, defaultValue string) string {
	value := Get(ctx, key)
	if strVal, ok := value.(string); ok && strVal != "" {
		return strVal
	}
	return defaultValue
}

// GetInt retrieves an integer value from context.
//
// Input:
//   - ctx: The context.Context
//   - key: The string key
//   - defaultValue: Default if not found or invalid type
//
// Output:
//   - int: The value or default
//
// Example:
//
//	userID := contextutil.GetInt(ctx, "user_id", 0)
func GetInt(ctx context.Context, key string, defaultValue int) int {
	value := Get(ctx, key)
	if intVal, ok := value.(int); ok {
		return intVal
	}
	return defaultValue
}

// GetBool retrieves a boolean value from context.
//
// Input:
//   - ctx: The context.Context
//   - key: The string key
//   - defaultValue: Default if not found or invalid type
//
// Output:
//   - bool: The value or default
//
// Example:
//
//	isAdmin := contextutil.GetBool(ctx, "is_admin", false)
func GetBool(ctx context.Context, key string, defaultValue bool) bool {
	value := Get(ctx, key)
	if boolVal, ok := value.(bool); ok {
		return boolVal
	}
	return defaultValue
}

// Set sets a value of any type in context.
//
// Input:
//   - ctx: The context.Context
//   - key: The string key
//   - value: The value to set
//
// Output:
//   - context.Context: New context with value set
//
// Example:
//
//	ctx = contextutil.Set(ctx, "user", &User{ID: "123"})
func Set(ctx context.Context, key string, value interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	typedKey := GetContextKey(key)
	return context.WithValue(ctx, typedKey, value)
}

// Has checks if a key exists in context.
//
// Input:
//   - ctx: The context.Context
//   - key: The string key
//
// Output:
//   - bool: True if key exists
//
// Example:
//
//	if contextutil.Has(ctx, "user_id") {
//	    // User is authenticated
//	}
func Has(ctx context.Context, key string) bool {
	if ctx == nil {
		return false
	}
	typedKey := GetContextKey(key)
	return ctx.Value(typedKey) != nil
}

// GetLanguageFromContext retrieves the language from the context.
//
// Input:
//   - ctx: The request context.Context
//
// Output:
//   - string: The language code, or empty string if not found
//
// Example:
//
//	lang := contextutil.GetLanguageFromContext(ctx)
//	if lang == "" {
//	    lang = "en" // default
//	}
func GetLanguageFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if lang, ok := ctx.Value(LangKey{}).(string); ok {
		return lang
	}
	return ""
}

// GetUserIDFromContext retrieves the user ID from the context.
// Returns the raw value (string or int64) for user to cast as needed.
//
// Input:
//   - ctx: The request context.Context
//
// Output:
//   - interface{}: The user ID value, or nil if not found
//
// Example:
//
//	// Cast to string
//	userID, ok := contextutil.GetUserIDFromContext(ctx).(string)
//	if !ok || userID == "" {
//	    return errors.New("User not authenticated")
//	}
//
//	// Cast to int64
//	userIDInt, ok := contextutil.GetUserIDFromContext(ctx).(int64)
//	if !ok || userIDInt == 0 {
//	    return errors.New("User not authenticated")
//	}
func GetUserIDFromContext(ctx context.Context) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(UserIDKey{})
}

// GetRequestIDFromContext retrieves the request ID from the context.
//
// Input:
//   - ctx: The request context.Context
//
// Output:
//   - string: The request ID, or empty string if not found
//
// Example:
//
//	requestID := contextutil.GetRequestIDFromContext(ctx)
//	logger.Info("Processing request", "request_id", requestID)
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}

// GetTraceIDFromContext retrieves the trace ID from the context.
//
// Input:
//   - ctx: The request context.Context
//
// Output:
//   - string: The trace ID, or empty string if not found
//
// Example:
//
//	traceID := contextutil.GetTraceIDFromContext(ctx)
//	if traceID == "" {
//	    traceID = generateTraceID()
//	    ctx = contextutil.SetTraceIDInContext(ctx, traceID)
//	}
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}

// Deprecated functions - use generic Set() instead

// SetCustomValueInContext sets a custom value in context.
// Deprecated: Use Set(ctx, key, value) instead.
func SetCustomValueInContext(ctx context.Context, key, value string) context.Context {
	return Set(ctx, key, value)
}

// SetLanguageInContext sets language in context.
// Deprecated: Use Set(ctx, KeyLanguage, lang) instead.
func SetLanguageInContext(ctx context.Context, lang string) context.Context {
	return Set(ctx, KeyLanguage, lang)
}

// SetUserIDInContext sets user ID in context.
// Deprecated: Use Set(ctx, KeyUserID, userID) instead.
func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	return Set(ctx, KeyUserID, userID)
}

// SetRequestIDInContext sets request ID in context.
// Deprecated: Use Set(ctx, KeyRequestID, requestID) instead.
func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return Set(ctx, KeyRequestID, requestID)
}

// SetTraceIDInContext sets trace ID in context.
// Deprecated: Use Set(ctx, KeyTraceID, traceID) instead.
func SetTraceIDInContext(ctx context.Context, traceID string) context.Context {
	return Set(ctx, KeyTraceID, traceID)
}

// GetAllContextValues returns all known context values.
// Deprecated: Use Get() for specific keys instead.
func GetAllContextValues(ctx context.Context) map[string]string {
	if ctx == nil {
		return make(map[string]string)
	}

	result := make(map[string]string)

	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		result[KeyRequestID] = requestID
	}
	if userIDValue := GetUserIDFromContext(ctx); userIDValue != nil {
		var userIDStr string
		switch v := userIDValue.(type) {
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
