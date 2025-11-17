package middleware

import (
	"strconv"
	"strings"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	runtimectx "github.com/anthanhphan/gosdk/transport/aurelion/internal/runtimectx"
	"github.com/gofiber/fiber/v2"
)

// adaptMiddlewareContext adapts core.Context to middleware.ContextInterface
func adaptMiddlewareContext(ctx core.Context) ContextInterface {
	if ctx == nil {
		return nil
	}
	// If it's already a ContextInterface, return it
	if ci, ok := ctx.(ContextInterface); ok {
		return ci
	}
	// Create an adapter
	return &middlewareContextAdapter{ctx: ctx}
}

// middlewareContextAdapter adapts core.Context to middleware.ContextInterface
type middlewareContextAdapter struct {
	ctx core.Context
}

func (a *middlewareContextAdapter) Locals(key string, value ...interface{}) interface{} {
	return a.ctx.Locals(key, value...)
}

func (a *middlewareContextAdapter) GetAllLocals() map[string]interface{} {
	return a.ctx.GetAllLocals()
}

func (a *middlewareContextAdapter) Next() error {
	return a.ctx.Next()
}

// HeaderToLocals returns a middleware that copies request headers into context locals.
// Note: This returns a function that will be wrapped by the main package to match aurelion.Middleware.
func HeaderToLocals(prefix string, filter func(string) bool) MiddlewareFunc {
	return func(ctx ContextInterface) error {
		// Get fiber context - try multiple extraction methods
		var fiberCtx *fiber.Ctx

		// Try to extract from middlewareContextAdapter (most common case in tests)
		if adapter, ok := ctx.(*middlewareContextAdapter); ok && adapter != nil && adapter.ctx != nil {
			// Try to extract directly from the underlying core.Context
			if fc, ok := runtimectx.FiberFromContext(adapter.ctx); ok {
				fiberCtx = fc
				goto found
			}
		}

		// Try to extract if ctx itself implements core.Context (unlikely but possible)
		// This handles the case where ContextInterface also implements core.Context
		if coreCtx, ok := ctx.(core.Context); ok {
			if fc, ok := runtimectx.FiberFromContext(coreCtx); ok {
				fiberCtx = fc
				goto found
			}
		}

		// If we can't get fiber context, skip header copying
		return ctx.Next()

	found:
		// Extract headers from fiber context and store in context locals
		// Store via context interface which handles tracking automatically
		fiberCtx.Request().Header.VisitAll(func(key, value []byte) {
			lowerKey := strings.ToLower(string(key))
			if filter != nil && !filter(lowerKey) {
				return
			}

			localsKey := lowerKey
			if prefix != "" {
				localsKey = prefix + lowerKey
			}

			// Store via context interface - this ensures proper tracking
			// The context interface will delegate to the underlying FiberContext
			// which will store in fiberCtx.Locals() and track the key automatically
			ctx.Locals(localsKey, string(value))
			// Also store directly in fiber context to ensure it's accessible
			// This is the source of truth for fiber context locals
			fiberCtx.Locals(localsKey, string(value))
			// Track the key explicitly to ensure it's in the tracked map
			runtimectx.TrackFiberLocal(fiberCtx, localsKey)
		})

		return ctx.Next()
	}
}

// DefaultHeaderToLocals returns middleware that copies all headers without prefix.
func DefaultHeaderToLocals() MiddlewareFunc {
	return HeaderToLocals("", nil)
}

// GetHeader retrieves a header value from context locals.
// It accepts both ContextInterface and core.Context.
func GetHeader(ctx interface{}, headerName string, defaultValue ...string) string {
	lowerKey := strings.ToLower(headerName)

	// First, try to extract fiber context and read directly from it
	// This is the most reliable way since values are stored directly in fiber context locals
	// We read directly from fiberCtx.Locals() to bypass any tracking requirements
	if coreCtx, ok := ctx.(core.Context); ok {
		if fiberCtx, ok := runtimectx.FiberFromContext(coreCtx); ok {
			// Read directly from fiber context locals - this bypasses tracking
			if value := fiberCtx.Locals(lowerKey); value != nil {
				if strValue, ok := value.(string); ok {
					return strValue
				}
			}
		}
		// Fallback to core.Context.Locals() which handles tracking
		if value := coreCtx.Locals(lowerKey); value != nil {
			if strValue, ok := value.(string); ok {
				return strValue
			}
		}
	}

	// Try to get value from ContextInterface
	if ci, ok := ctx.(ContextInterface); ok {
		if value := ci.Locals(lowerKey); value != nil {
			if strValue, ok := value.(string); ok {
				return strValue
			}
		}
	}

	// Try to adapt other context types to ContextInterface
	if adapter := adaptContextForGetHeader(ctx); adapter != nil {
		if value := adapter.Locals(lowerKey); value != nil {
			if strValue, ok := value.(string); ok {
				return strValue
			}
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// adaptContextForGetHeader adapts various context types to ContextInterface for GetHeader.
// It handles core.Context by wrapping it in a middlewareContextAdapter.
func adaptContextForGetHeader(ctx interface{}) ContextInterface {
	if ctx == nil {
		return nil
	}
	// If it's already a ContextInterface, return it
	if ci, ok := ctx.(ContextInterface); ok {
		return ci
	}
	// Try to adapt by checking if it has the methods we need
	// This handles core.Context and other compatible types
	if hasLocalsMethod(ctx) {
		return &getHeaderContextAdapter{ctx: ctx}
	}
	return nil
}

// getHeaderContextAdapter adapts any context with Locals/GetAllLocals/Next methods to ContextInterface.
type getHeaderContextAdapter struct {
	ctx interface{}
}

func (a *getHeaderContextAdapter) Locals(key string, value ...interface{}) interface{} {
	// Use reflection or type assertion to call Locals
	// For now, try to call it directly if it's a known type
	if ctx, ok := a.ctx.(interface {
		Locals(key string, value ...interface{}) interface{}
	}); ok {
		return ctx.Locals(key, value...)
	}
	return nil
}

func (a *getHeaderContextAdapter) GetAllLocals() map[string]interface{} {
	if ctx, ok := a.ctx.(interface {
		GetAllLocals() map[string]interface{}
	}); ok {
		return ctx.GetAllLocals()
	}
	return make(map[string]interface{})
}

func (a *getHeaderContextAdapter) Next() error {
	if ctx, ok := a.ctx.(interface {
		Next() error
	}); ok {
		return ctx.Next()
	}
	return nil
}

// hasLocalsMethod checks if the context has the Locals method.
func hasLocalsMethod(ctx interface{}) bool {
	_, ok := ctx.(interface {
		Locals(key string, value ...interface{}) interface{}
		GetAllLocals() map[string]interface{}
		Next() error
	})
	return ok
}

// GetHeaderInt retrieves a header value as integer from context locals.
func GetHeaderInt(ctx ContextInterface, headerName string, defaultValue int) int {
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

// GetHeaderBool retrieves a header value as boolean from context locals.
func GetHeaderBool(ctx ContextInterface, headerName string, defaultValue bool) bool {
	strValue := strings.ToLower(GetHeader(ctx, headerName))
	if strValue == "" {
		return defaultValue
	}

	switch strValue {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetAllHeaders retrieves all headers from context locals.
func GetAllHeaders(ctx ContextInterface, prefix string) map[string]string {
	allLocals := ctx.GetAllLocals()
	headers := make(map[string]string)

	for key, value := range allLocals {
		if prefix != "" {
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			key = strings.TrimPrefix(key, prefix)
		}

		if strValue, ok := value.(string); ok {
			headers[key] = strValue
		}
	}

	return headers
}
