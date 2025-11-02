package aurelion

import (
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
//	server.AddGlobalMiddlewares(DefaultHeaderToLocalsMiddleware())
func DefaultHeaderToLocalsMiddleware() Middleware {
	return HeaderToLocalsMiddleware("", nil)
}
