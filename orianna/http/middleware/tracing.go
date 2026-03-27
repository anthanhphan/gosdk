// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/shared/ctxkeys"
	"github.com/anthanhphan/gosdk/tracing"
)

// TracingMiddleware creates a middleware that starts a span for each HTTP request.
// It extracts trace context from incoming headers, creates a server span with
// standard HTTP attributes, and stores the trace ID in locals for log correlation.
//
// Uses ctx.RoutePath() instead of ctx.Path() to record route patterns (e.g., "/users/:id")
// rather than actual paths (e.g., "/users/123"), preventing unbounded span name cardinality.
func TracingMiddleware(client tracing.Client) core.Middleware {
	return func(ctx core.Context) error {
		// Cache method to avoid repeated interface method calls
		method := ctx.Method()

		// Extract trace context from incoming headers.
		// Allocate carrier once — it is reused by being passed to InjectContext below.
		headers := tracing.HeaderCarrier{
			"traceparent": ctx.Get("traceparent"),
			"tracestate":  ctx.Get("tracestate"),
		}
		goCtx := tracing.ExtractContext(ctx.Context(), headers)

		// Start server span with a constant name to avoid string concatenation
		// before route resolution. The span name is updated after ctx.Next()
		// once the actual route pattern is matched.
		goCtx, span := client.StartSpan(goCtx, "HTTP",
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.url", ctx.OriginalURL()),
				attribute.String("http.scheme", ctx.Protocol()),
				attribute.String("net.host.name", ctx.Hostname()),
				attribute.String("http.user_agent", ctx.Get("User-Agent")),
				attribute.String("net.peer.ip", ctx.IP()),
			),
		)
		defer span.End()

		// Store trace context for downstream use
		ctx.SetContext(goCtx)

		// Store trace ID in locals for log correlation
		traceID := tracing.TraceIDFromContext(goCtx)
		if traceID != "" {
			ctx.Locals(ctxkeys.TraceID.Key(), traceID)
			ctx.Set("X-Trace-ID", traceID)
		}

		// Execute next handler
		err := ctx.Next()

		// Now the route is resolved — update span name and record all response
		// attributes in a single SetAttributes call to reduce marshaling overhead.
		routePath := ctx.RoutePath()
		statusCode := ctx.ResponseStatusCode()
		span.SetName(method + " " + routePath)
		span.SetAttributes(
			attribute.String("http.route", routePath),
			attribute.Int("http.status_code", statusCode),
		)

		// Inject trace context into response headers
		// Reuse the incoming HeaderCarrier to avoid a second map allocation
		for k := range headers {
			delete(headers, k)
		}
		tracing.InjectContext(goCtx, headers)
		for k, v := range headers {
			ctx.Set(k, v)
		}

		// Record error/status with enriched attributes for ErrorResponse
		if err != nil {
			span.RecordError(err)
			var errResp *core.ErrorResponse
			if errors.As(err, &errResp) {
				span.SetAttributes(
					attribute.String("error.code", errResp.Code),
					attribute.String("error.internal_message", errResp.InternalMessage),
				)
			}
			span.SetStatus(codes.Error, err.Error())
		} else if statusCode >= 400 {
			span.SetStatus(codes.Error, "HTTP "+statusString(statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}
