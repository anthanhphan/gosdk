package contextutil

import (
	"context"
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

// GetFromContext retrieves a value from context by string key.
// It supports both typed keys (for predefined keys) and custom keys.
//
// Input:
//   - ctx: The context.Context
//   - key: The string key to look up
//
// Output:
//   - string: The value as string, or empty string if not found
//
// Example:
//
//	value := contextutil.GetFromContext(ctx, "custom_key")
func GetFromContext(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}

	// Try typed key
	typedKey := GetContextKey(key)
	if value, ok := ctx.Value(typedKey).(string); ok {
		return value
	}

	return ""
}

// SetCustomValueInContext sets a custom string value in the context and returns a new context.
// This is useful for setting custom keys that are not in the predefined set.
//
// Input:
//   - ctx: The request context.Context
//   - key: The custom string key
//   - value: The string value to set
//
// Output:
//   - context.Context: A new context with the custom value set
//
// Example:
//
//	ctx = contextutil.SetCustomValueInContext(ctx, "custom_key", "custom_value")
func SetCustomValueInContext(ctx context.Context, key, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, CustomKey(key), value)
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

// SetLanguageInContext sets the language in the context and returns a new context.
//
// Input:
//   - ctx: The request context.Context
//   - lang: The language code to set
//
// Output:
//   - context.Context: A new context with the language value set
//
// Example:
//
//	ctx = contextutil.SetLanguageInContext(ctx, "en")
func SetLanguageInContext(ctx context.Context, lang string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, LangKey{}, lang)
}

// GetUserIDFromContext retrieves the user ID from the context.
//
// Input:
//   - ctx: The request context.Context
//
// Output:
//   - string: The user ID, or empty string if not found
//
// Example:
//
//	userID := contextutil.GetUserIDFromContext(ctx)
//	if userID == "" {
//	    return errors.New("User not authenticated")
//	}
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value(UserIDKey{}).(string); ok {
		return userID
	}
	return ""
}

// SetUserIDInContext sets the user ID in the context and returns a new context.
//
// Input:
//   - ctx: The request context.Context
//   - userID: The user ID to set
//
// Output:
//   - context.Context: A new context with the user ID value set
//
// Example:
//
//	ctx = contextutil.SetUserIDInContext(ctx, "user123")
func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, UserIDKey{}, userID)
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

// SetRequestIDInContext sets the request ID in the context and returns a new context.
//
// Input:
//   - ctx: The request context.Context
//   - requestID: The request ID to set
//
// Output:
//   - context.Context: A new context with the request ID value set
//
// Example:
//
//	ctx = contextutil.SetRequestIDInContext(ctx, "req-12345")
func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, RequestIDKey{}, requestID)
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

// SetTraceIDInContext sets the trace ID in the context and returns a new context.
//
// Input:
//   - ctx: The request context.Context
//   - traceID: The trace ID to set
//
// Output:
//   - context.Context: A new context with the trace ID value set
//
// Example:
//
//	ctx = contextutil.SetTraceIDInContext(ctx, "trace-12345")
func SetTraceIDInContext(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, TraceIDKey{}, traceID)
}

// GetAllContextValues returns all string values from the context as a map.
// This only includes values that were set using contextutil typed keys.
// Note: context.Context doesn't support iteration, so this only returns
// values for known keys. For all Locals values, use aurelion.Context.Locals()
// or access them through the aurelion.Context interface.
//
// Input:
//   - ctx: The context.Context to extract values from
//
// Output:
//   - map[string]string: A map of all string values in the context
//
// Example:
//
//	values := contextutil.GetAllContextValues(ctx)
//	for key, value := range values {
//	    logger.Infow("context value", "key", key, "value", value)
//	}
func GetAllContextValues(ctx context.Context) map[string]string {
	if ctx == nil {
		return make(map[string]string)
	}

	result := make(map[string]string)

	// Get all predefined keys
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		result[KeyRequestID] = requestID
	}
	if userID := GetUserIDFromContext(ctx); userID != "" {
		result[KeyUserID] = userID
	}
	if lang := GetLanguageFromContext(ctx); lang != "" {
		result[KeyLanguage] = lang
	}
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		result[KeyTraceID] = traceID
	}

	// Note: We cannot iterate through all context values because context.Context
	// doesn't provide an iteration API. To get all custom values, access them
	// through aurelion.Context.Locals() or use reflection to access the underlying
	// context values.

	return result
}
