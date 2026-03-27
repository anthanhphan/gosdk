// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"net/http"
	"time"

	"github.com/anthanhphan/gosdk/orianna/shared/httputil"
	"github.com/anthanhphan/gosdk/orianna/shared/observability"
	"github.com/anthanhphan/gosdk/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// startTracing starts a tracing span for the request.
func (c *Client) startTracing(httpReq *http.Request, req *Request, reqURL string, attempt int) (*http.Request, tracing.Span) {
	spanName := req.Method + " " + req.Path
	var span tracing.Span

	if c.tracing != nil {
		var spanCtx context.Context
		spanCtx, span = c.tracing.StartSpan(httpReq.Context(), spanName,
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", reqURL),
				attribute.Int("http.attempt", attempt),
			),
		)
		httpReq = httpReq.WithContext(spanCtx)
	}

	return httpReq, span
}

// recordMetrics records request metrics.
func (c *Client) recordMetrics(ctx context.Context, req *Request, attempt int, httpResp *http.Response, start time.Time) {
	statusCode := 0
	if httpResp != nil {
		statusCode = httpResp.StatusCode
	}

	// Use cached status code string to avoid strconv.Itoa allocation
	status := observability.CodeString(c.statusCodeCache, statusCode)
	errClass := httputil.ErrorClassFromStatus(statusCode)
	attemptStr := observability.AttemptString(attempt)

	c.metrics.Inc(ctx, c.metricRequestsTotal,
		"method", req.Method,
		"path", req.Path,
		"status", status,
		"error_class", errClass,
		"attempt", attemptStr,
	)

	c.metrics.Duration(ctx, c.metricRequestDuration, start,
		"method", req.Method,
		"path", req.Path,
		"status", status,
		"attempt", attemptStr,
	)
}

// handleSpanError handles span errors.
func (c *Client) handleSpanError(span tracing.Span, err error) {
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
	}
}

// endSpan ends the tracing span with appropriate status.
func (c *Client) endSpan(span tracing.Span, httpResp *http.Response) {
	if span != nil {
		statusCode := httpResp.StatusCode
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
		)
		if statusCode >= 400 {
			// Use cached string concat instead of fmt.Sprintf
			span.SetStatus(codes.Error, "HTTP "+observability.CodeString(c.statusCodeCache, statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}
}
