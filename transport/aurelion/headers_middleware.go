package aurelion

import (
	"strconv"
	"strings"
)

// HeaderToLocalsMiddleware creates middleware that automatically parses
// all request headers and stores them in Locals for easy access.
// Headers are stored with lowercase keys for consistency.
// This middleware is dynamic - any new headers added in the future will
// automatically be parsed without needing to modify the package code.
//
// Input:
//   - prefix: Optional prefix for header keys in Locals (e.g., "header_").
//     If empty, headers are stored with their original names in lowercase.
//   - filter: Optional function to filter which headers to include.
//     If nil, all headers are included.
//
// Output:
//   - Middleware: The middleware function that parses headers to Locals
//
// Example:
//
//	// Parse all headers to Locals without prefix
//	middleware := HeaderToLocalsMiddleware("", nil)
//
//	// Parse only specific headers with prefix
//	middleware := HeaderToLocalsMiddleware("header_", func(key string) bool {
//	    return key == "uid" || key == "Accept-Language"
//	})
func HeaderToLocalsMiddleware(prefix string, filter func(string) bool) Middleware {
	return func(ctx Context) error {
		// Get fiber context to access headers
		fiberCtx := ctx.(*contextWrapper).fiberCtx

		// Iterate through all request headers
		// Use VisitHeaders to iterate through all headers dynamically
		fiberCtx.Request().Header.VisitAll(func(key, value []byte) {
			// Convert header key to lowercase string for consistency
			lowerKey := strings.ToLower(string(key))

			// Apply filter if provided
			if filter != nil && !filter(lowerKey) {
				return
			}

			// Determine the key name in Locals
			localsKey := lowerKey
			if prefix != "" {
				localsKey = prefix + lowerKey
			}

			// Store header value in Locals
			ctx.Locals(localsKey, string(value))
		})

		return ctx.Next()
	}
}

// DefaultHeaderToLocalsMiddleware creates middleware that automatically parses
// all request headers and stores them in Locals with lowercase keys.
// This is a convenience function that uses HeaderToLocalsMiddleware with no prefix
// and no filter, meaning all headers will be parsed.
//
// Output:
//   - Middleware: The middleware function that parses all headers to Locals
//
// Example:
//
//	server, _ := aurelion.NewHttpServer(
//	    config,
//	    aurelion.WithGlobalMiddleware(aurelion.DefaultHeaderToLocalsMiddleware()),
//	)
func DefaultHeaderToLocalsMiddleware() Middleware {
	return HeaderToLocalsMiddleware("", nil)
}

// GetHeader retrieves a header value from context Locals.
// Headers are stored in lowercase by HeaderToLocalsMiddleware.
//
// Input:
//   - ctx: The request context
//   - headerName: The header name (case-insensitive, will be converted to lowercase)
//   - defaultValue: Optional default value if header not found
//
// Output:
//   - string: The header value, or default value if not found
//
// Example:
//
//	// Get Accept-Language header
//	lang := aurelion.GetHeader(ctx, "Accept-Language", "en")
//
//	// Get custom uid header
//	uid := aurelion.GetHeader(ctx, "uid")
//
//	// Get test-header
//	testValue := aurelion.GetHeader(ctx, "test-header", "default")
func GetHeader(ctx Context, headerName string, defaultValue ...string) string {
	// Convert header name to lowercase for consistency
	lowerKey := strings.ToLower(headerName)

	// Try to get from Locals
	value := ctx.Locals(lowerKey)
	if value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}

	// Return default value if provided
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	// Return empty string if not found
	return ""
}

// GetHeaderInt retrieves a header value as integer from context Locals.
//
// Input:
//   - ctx: The request context
//   - headerName: The header name (case-insensitive)
//   - defaultValue: Default value if header not found or invalid
//
// Output:
//   - int: The header value as integer, or default value if not found/invalid
//
// Example:
//
//	limit := aurelion.GetHeaderInt(ctx, "X-Rate-Limit", 100)
//	timeout := aurelion.GetHeaderInt(ctx, "X-Timeout", 30)
func GetHeaderInt(ctx Context, headerName string, defaultValue int) int {
	strValue := GetHeader(ctx, headerName)
	if strValue == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetHeaderBool retrieves a header value as boolean from context Locals.
// Accepts: "true", "1", "yes", "on" as true (case-insensitive)
//
// Input:
//   - ctx: The request context
//   - headerName: The header name (case-insensitive)
//   - defaultValue: Default value if header not found or invalid
//
// Output:
//   - bool: The header value as boolean, or default value if not found/invalid
//
// Example:
//
//	debug := aurelion.GetHeaderBool(ctx, "X-Debug", false)
//	verbose := aurelion.GetHeaderBool(ctx, "X-Verbose", false)
func GetHeaderBool(ctx Context, headerName string, defaultValue bool) bool {
	strValue := strings.ToLower(GetHeader(ctx, headerName))
	if strValue == "" {
		return defaultValue
	}

	// Check for truthy values
	switch strValue {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetAllHeaders retrieves all headers from context Locals.
// Returns a map with lowercase header names as keys.
//
// Input:
//   - ctx: The request context
//
// Output:
//   - map[string]string: Map of all header names and values
//
// Example:
//
//	headers := aurelion.GetAllHeaders(ctx)
//	for name, value := range headers {
//	    log.Printf("Header: %s = %s", name, value)
//	}
func GetAllHeaders(ctx Context, prefix string) map[string]string {
	allLocals := ctx.GetAllLocals()
	headers := make(map[string]string)

	for key, value := range allLocals {
		// If prefix is specified, only include keys with that prefix
		if prefix != "" {
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			// Remove prefix from key
			key = strings.TrimPrefix(key, prefix)
		}

		// Only include string values
		if strValue, ok := value.(string); ok {
			headers[key] = strValue
		}
	}

	return headers
}
