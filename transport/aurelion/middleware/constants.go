package middleware

import "github.com/anthanhphan/gosdk/transport/aurelion/internal/context"

// Re-export commonly used constants for middleware convenience.
const (
	// HTTP Headers for correlation and tracing.
	RequestIDHeader = context.RequestIDHeader
	TraceIDHeader   = context.TraceIDHeader

	// Context keys for storing values in request locals.
	ContextKeyRequestID = context.ContextKeyRequestID
	ContextKeyTraceID   = context.ContextKeyTraceID
	ContextKeyLanguage  = context.ContextKeyLanguage
	ContextKeyUserID    = context.ContextKeyUserID
)
