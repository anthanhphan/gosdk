package aurelion

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetRequestID retrieves the request ID from the context
//
// Input:
//   - ctx: The request context
//
// Output:
//   - string: The request ID
//
// Example:
//
//	requestID := aurelion.GetRequestID(ctx)
//	logger.Info("Processing request", "request_id", requestID)
func GetRequestID(ctx Context) string {
	// Try to get from locals first
	if id := ctx.Locals(contextKeyRequestID); id != nil {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}

	// Try to get from header
	if fiberCtx, ok := ctx.Context().(*fiber.Ctx); ok {
		return fiberCtx.Get(RequestIDHeader, "")
	}

	return ""
}

// requestIDMiddleware is an internal middleware that generates and propagates request IDs.
// It checks if a request ID exists in the header, generates a new optimized ID if not,
// and stores it in both the response header and local context for access throughout the request lifecycle.
func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists in header
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			uuid, _ := uuid.NewV7()
			requestID = uuid.String()
		}

		// Set request ID in response header for client tracking
		c.Set(RequestIDHeader, requestID)

		// Store in locals for easy access by handlers and middleware
		c.Locals(contextKeyRequestID, requestID)

		return c.Next()
	}
}
