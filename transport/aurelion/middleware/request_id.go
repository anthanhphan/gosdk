package middleware

import (
	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDMiddleware generates and propagates request IDs.
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if existing := c.Locals(ContextKeyRequestID); existing != nil {
			if idStr, ok := existing.(string); ok && idStr != "" {
				c.Set(RequestIDHeader, idStr)
				rctx.TrackFiberLocal(c, ContextKeyRequestID)
				return c.Next()
			}
		}

		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			uuidV7, err := uuid.NewV7()
			if err != nil {
				requestID = uuid.NewString()
			} else {
				requestID = uuidV7.String()
			}
		}

		c.Set(RequestIDHeader, requestID)
		c.Locals(ContextKeyRequestID, requestID)
		rctx.TrackFiberLocal(c, ContextKeyRequestID)

		return c.Next()
	}
}
