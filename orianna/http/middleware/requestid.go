// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/shared/ctxkeys"
	"github.com/anthanhphan/gosdk/orianna/shared/requestid"
)

// RequestIDMiddleware creates a middleware that generates and propagates request IDs.
// If a request ID is already present in the X-Request-ID header and is valid, it will be reused.
// Otherwise, a new UUID v7 is generated.
// Invalid request IDs (too long, special characters) are discarded for security.
//
// The request ID is:
// - Set in the response header X-Request-ID
// - Stored in context locals for access via ctx.RequestID()
func RequestIDMiddleware() core.Middleware {
	return func(ctx core.Context) error {
		reqID := ctx.Get(core.HeaderRequestID)
		if !requestid.IsValid(reqID) {
			reqID = requestid.Generate()
		}

		// Set response header
		ctx.Set(core.HeaderRequestID, reqID)
		// Store in locals for access via ctx.RequestID()
		ctx.Locals(ctxkeys.RequestID.Key(), reqID)

		return ctx.Next()
	}
}
