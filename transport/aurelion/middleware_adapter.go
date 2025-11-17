package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/middleware"
)

// Middleware adapter functions to convert between middleware types and public aurelion types.

// HeaderToLocalsPublic wraps the internal HeaderToLocals to return aurelion.Middleware.
func HeaderToLocalsPublic(prefix string, filter func(string) bool) Middleware {
	internalMw := middleware.HeaderToLocals(prefix, filter)
	return func(ctx Context) error {
		return internalMw(ctx)
	}
}

// DefaultHeaderToLocalsPublic wraps DefaultHeaderToLocals to return aurelion.Middleware.
func DefaultHeaderToLocalsPublic() Middleware {
	internalMw := middleware.DefaultHeaderToLocals()
	return func(ctx Context) error {
		return internalMw(ctx)
	}
}

// GetHeaderPublic wraps GetHeader to work with aurelion.Context.
func GetHeaderPublic(ctx Context, headerName string, defaultValue ...string) string {
	return middleware.GetHeader(ctx, headerName, defaultValue...)
}

// GetHeaderIntPublic wraps GetHeaderInt to work with aurelion.Context.
func GetHeaderIntPublic(ctx Context, headerName string, defaultValue int) int {
	return middleware.GetHeaderInt(ctx, headerName, defaultValue)
}

// GetHeaderBoolPublic wraps GetHeaderBool to work with aurelion.Context.
func GetHeaderBoolPublic(ctx Context, headerName string, defaultValue bool) bool {
	return middleware.GetHeaderBool(ctx, headerName, defaultValue)
}

// GetAllHeadersPublic wraps GetAllHeaders to work with aurelion.Context.
func GetAllHeadersPublic(ctx Context, prefix string) map[string]string {
	return middleware.GetAllHeaders(ctx, prefix)
}
