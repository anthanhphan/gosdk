// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"strings"

	"github.com/anthanhphan/gosdk/orianna/http/core"
)

// SensitiveHTTPHeaders is the set of HTTP headers that must never be logged.
// These are filtered from verbose logging to prevent PII/PCI/credential leakage.
var SensitiveHTTPHeaders = map[string]struct{}{
	"authorization":       {},
	"x-api-key":           {},
	"x-api-secret":        {},
	"proxy-authorization": {},
	"cookie":              {},
	"set-cookie":          {},
	"x-csrf-token":        {},
	"x-xsrf-token":        {},
	"x-refresh-token":     {},
	"access-token":        {},
	"refresh-token":       {},
	"secret-key":          {},
	"private-key":         {},
}

// SanitizeHeaders returns a new map with sensitive headers redacted.
// This should be used before logging HTTP headers.
func SanitizeHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}

	sanitized := make(map[string]string, len(headers))
	for k, v := range headers {
		keyLower := strings.ToLower(k)
		if _, sensitive := SensitiveHTTPHeaders[keyLower]; sensitive {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

// SanitizeHeaderValue redacts a single header value if the header name is sensitive.
func SanitizeHeaderValue(headerName, value string) string {
	keyLower := strings.ToLower(headerName)
	if _, sensitive := SensitiveHTTPHeaders[keyLower]; sensitive {
		return "[REDACTED]"
	}
	return value
}

// SanitizeHeadersFromContext extracts and sanitizes headers from the context.
// This is useful for logging middleware.
// NOTE: Since core.Context does not expose a raw header iterator, we check
// a comprehensive list of commonly used headers. If you need ALL headers,
// use Fiber's Request().Header.VisitAll() at the adapter layer instead.
func SanitizeHeadersFromContext(ctx core.Context) map[string]string {
	headers := make(map[string]string)

	// Comprehensive list of commonly used HTTP headers
	commonHeaders := []string{
		// Standard
		"Content-Type",
		"Content-Length",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"User-Agent",
		"Host",
		"Origin",
		"Referer",
		// Tracking/Correlation
		"X-Request-ID",
		"X-Trace-ID",
		"X-Correlation-ID",
		"X-B3-TraceId",
		"X-B3-SpanId",
		"traceparent",
		// Proxy/Forwarding
		"X-Forwarded-For",
		"X-Forwarded-Proto",
		"X-Forwarded-Host",
		"X-Real-IP",
		// Security (will be redacted)
		"Authorization",
		"Cookie",
		"X-API-Key",
		"X-CSRF-Token",
	}

	for _, h := range commonHeaders {
		if v := ctx.Get(h); v != "" {
			headers[h] = v
		}
	}

	return SanitizeHeaders(headers)
}
