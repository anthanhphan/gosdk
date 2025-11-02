package aurelion

import (
	"context"

	"github.com/anthanhphan/gosdk/transport/aurelion/contextutil"
)

// getTypedContextKey returns the appropriate typed key for a given string key.
// This is used internally when merging Locals values into context.Context.
// It maps string keys to typed keys from contextutil for type-safe context value storage.
// This ensures compatibility with contextutil functions like GetUserIDFromContext.
//
// Input:
//   - key: The string key from Locals
//
// Output:
//   - interface{}: The typed key for use with context.WithValue
func getTypedContextKey(key string) interface{} {
	// Use contextutil.GetContextKey to get the proper typed key
	// This ensures compatibility with contextutil functions
	return contextutil.GetContextKey(key)
}

// getValueFromContext retrieves a string value from context.Context using a typed key.
// This is a helper function for internal use when accessing context values.
// It uses contextutil.GetFromContext for consistency.
//
// Input:
//   - ctx: The context.Context to read from
//   - key: The string key (will be converted to typed key)
//
// Output:
//   - string: The value if found, empty string otherwise
func getValueFromContext(ctx context.Context, key string) string {
	return contextutil.GetFromContext(ctx, key)
}
