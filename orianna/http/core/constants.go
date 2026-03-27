// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "net/http"

// HTTP Status Codes -- only codes used by the framework are aliased here.
// For other status codes, use net/http directly (e.g., http.StatusTeapot).
const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusAccepted            = http.StatusAccepted
	StatusNoContent           = http.StatusNoContent
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusUnprocessableEntity = http.StatusUnprocessableEntity
	StatusTooManyRequests     = http.StatusTooManyRequests
	StatusInternalServerError = http.StatusInternalServerError
	StatusServiceUnavailable  = http.StatusServiceUnavailable
	StatusGatewayTimeout      = http.StatusGatewayTimeout
)

// HTTP Headers
const (
	HeaderRequestID       = "X-Request-ID"
	HeaderTraceID         = "X-Trace-ID"
	HeaderCorrelationID   = "X-Correlation-ID"
	HeaderUserAgent       = "User-Agent"
	HeaderContentType     = "Content-Type"
	HeaderContentLength   = "Content-Length"
	HeaderAuthorization   = "Authorization"
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderAcceptLanguage  = "Accept-Language"
	HeaderCacheControl    = "Cache-Control"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedProto = "X-Forwarded-Proto"
	HeaderXForwardedHost  = "X-Forwarded-Host"
	HeaderXRealIP         = "X-Real-IP"
	HeaderXB3TraceID      = "X-B3-TraceId"
	HeaderTraceparent     = "traceparent"
)

// Response Messages
const (
	MessageOK                        = "OK"
	MessageCreated                   = "Created"
	MessageAccepted                  = "Accepted"
	MessageNoContent                 = "No Content"
	MessageBadRequest                = "Bad Request"
	MessageUnauthorized              = "Unauthorized"
	MessageForbidden                 = "Forbidden"
	MessageNotFound                  = "Not Found"
	MessageConflict                  = "Conflict"
	MessageUnprocessableEntity       = "Unprocessable Entity"
	MessageTooManyRequests           = "Too Many Requests"
	MessageInternalServerError       = "Internal Server Error"
	MessageServiceUnavailable        = "Service Unavailable"
	DefaultUnknownRequestID          = "unknown"
	HealthMessageHealthy             = "HTTP endpoint is healthy"
	HealthMessageUnhealthy           = "HTTP endpoint returned error status: %d"
	HealthMessageDegraded            = "HTTP endpoint returned client error: %d"
	HealthMessageCreateRequestFailed = "Failed to create request: %v"
)

// Traceparent format constants
const (
	TraceparentPrefixLength     = 3
	TraceparentVersionSeparator = '-'
	TraceparentTraceIDStart     = 3
	TraceparentTraceIDEnd       = 35
)
