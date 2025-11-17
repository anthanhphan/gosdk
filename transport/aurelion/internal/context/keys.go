package context

// HTTP Headers used across the framework for correlation and tracing.
const (
	RequestIDHeader   = "X-Request-ID"
	TraceIDHeader     = "X-Trace-ID"
	TraceParentHeader = "traceparent"
	B3TraceIDHeader   = "X-B3-TraceId"
)

// Context Keys used for storing values in fiber.Ctx Locals.
const (
	ContextKeyRequestID = "request_id"
	ContextKeyTraceID   = "trace_id"
	ContextKeyLanguage  = "lang"
	ContextKeyUserID    = "user_id"
)

// Internal Keys used for framework implementation details.
const (
	TrackedLocalsKey = "__aurelion_tracked_locals__"
)
