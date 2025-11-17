package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/internal/context"
	"github.com/gofiber/fiber/v2"
)

// Adapter functions to convert between internal context and public types.
// These bridge the internal ContextInterface with the public Context interface.

// ensureFiberContextImplementsContext ensures compile-time check that FiberContext implements Context.
// Note: This will fail at compile time if FiberContext doesn't implement all Context methods.
// We use a wrapper to handle the Status() return type difference and Cookie type conversion.
type fiberContextWrapper struct {
	*context.FiberContext
}

func (w *fiberContextWrapper) Status(status int) Context {
	w.FiberContext.Status(status)
	return w
}

func (w *fiberContextWrapper) Cookie(cookie *Cookie) {
	// Convert public Cookie to internal Cookie
	internalCookie := &context.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		MaxAge:   cookie.MaxAge,
		Expires:  cookie.Expires,
		Secure:   cookie.Secure,
		HTTPOnly: cookie.HTTPOnly,
		SameSite: cookie.SameSite,
	}
	w.FiberContext.Cookie(internalCookie)
}

// Ensure wrapper implements Context
var _ Context = (*fiberContextWrapper)(nil)

// handlerToFiberInternal converts a public Handler to fiber.Handler using internal context.
func handlerToFiberInternal(handler Handler) fiber.Handler {
	if handler == nil {
		return func(c *fiber.Ctx) error { return nil }
	}
	return func(c *fiber.Ctx) error {
		internalCtx := context.NewFiberContext(c)
		if internalCtx == nil {
			return nil
		}
		// Wrap to ensure Status() returns Context
		ctx := &fiberContextWrapper{FiberContext: internalCtx.(*context.FiberContext)}
		return handler(ctx)
	}
}

// middlewareToFiberInternal converts a public Middleware to fiber.Handler using internal context.
func middlewareToFiberInternal(middleware Middleware) fiber.Handler {
	if middleware == nil {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	return func(c *fiber.Ctx) error {
		internalCtx := context.NewFiberContext(c)
		if internalCtx == nil {
			return nil
		}
		// Wrap to ensure Status() returns Context
		ctx := &fiberContextWrapper{FiberContext: internalCtx.(*context.FiberContext)}
		return middleware(ctx)
	}
}

// HandlerToFiberPublic converts a public Handler to fiber.Handler.
func HandlerToFiberPublic(handler Handler) fiber.Handler {
	return handlerToFiberInternal(handler)
}

// MiddlewareToFiberPublic converts a public Middleware to fiber.Handler.
func MiddlewareToFiberPublic(middleware Middleware) fiber.Handler {
	return middlewareToFiberInternal(middleware)
}

// NewFiberContextPublic creates a new context from fiber context, returning the public interface.
func NewFiberContextPublic(fiberCtx *fiber.Ctx) Context {
	internalCtx := context.NewFiberContext(fiberCtx)
	if internalCtx == nil {
		return nil
	}
	// Wrap to ensure Status() returns Context
	fc, ok := internalCtx.(*context.FiberContext)
	if !ok {
		return nil
	}
	return &fiberContextWrapper{FiberContext: fc}
}

// FiberFromContextPublic safely extracts fiber.Ctx from public Context.
func FiberFromContextPublic(ctx Context) (*fiber.Ctx, bool) {
	if ctx == nil {
		return nil, false
	}
	// Try to unwrap if it's a wrapper
	if wrapper, ok := ctx.(*fiberContextWrapper); ok {
		return context.FiberFromContext(wrapper.FiberContext)
	}
	// Try to extract fiber context using a helper method if available
	// Note: We can't use direct type assertion because aurelion.Context and context.ContextInterface
	// have incompatible method signatures. Instead, we use a helper that checks for the underlying type.
	if fiberCtx := extractFiberContextFromAurelion(ctx); fiberCtx != nil {
		return fiberCtx, true
	}
	return nil, false
}

// extractFiberContextFromAurelion attempts to extract fiber context from aurelion.Context
func extractFiberContextFromAurelion(ctx Context) *fiber.Ctx {
	// Try to unwrap the wrapper to get the underlying FiberContext
	if wrapper, ok := ctx.(*fiberContextWrapper); ok {
		if fiberCtx, ok := context.FiberFromContext(wrapper.FiberContext); ok {
			return fiberCtx
		}
	}
	// Try to get fiber context through the runtimectx package
	// First convert aurelion.Context to core.Context if possible
	// Since we can't directly convert, we'll try to extract from the wrapper
	return nil
}
