package middleware

import (
	"strings"

	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TraceIDMiddleware generates or reuses trace IDs from headers.
func TraceIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if existing := c.Locals(ContextKeyTraceID); existing != nil {
			if idStr, ok := existing.(string); ok && idStr != "" {
				c.Set(TraceIDHeader, idStr)
				rctx.TrackFiberLocal(c, ContextKeyTraceID)
				return c.Next()
			}
		}

		traceID := c.Get(TraceIDHeader)
		if traceID == "" {
			// Try B3 propagation format
			traceID = c.Get("X-B3-TraceId")
			if traceID == "" {
				// Try W3C trace context format
				traceID = extractTraceIDFromTraceparent(c.Get("traceparent"))
			}
		}

		if traceID == "" {
			uuidV7, err := uuid.NewV7()
			if err != nil {
				traceID = uuid.NewString()
			} else {
				traceID = uuidV7.String()
			}
		}

		c.Set(TraceIDHeader, traceID)
		c.Locals(ContextKeyTraceID, traceID)
		rctx.TrackFiberLocal(c, ContextKeyTraceID)

		return c.Next()
	}
}

func extractTraceIDFromTraceparent(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	parts := strings.Split(header, "-")
	if len(parts) < 4 {
		return ""
	}

	traceIDPart := parts[1]
	if len(traceIDPart) != 32 || !isHexString(traceIDPart) {
		return ""
	}

	return traceIDPart
}

func isHexString(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}

	return true
}
