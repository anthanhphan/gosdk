// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "net/http"

// HTTP Status Codes (aliases to net/http for convenience)
const (
	StatusOK                            = http.StatusOK
	StatusCreated                       = http.StatusCreated
	StatusAccepted                      = http.StatusAccepted
	StatusNoContent                     = http.StatusNoContent
	StatusPartialContent                = http.StatusPartialContent
	StatusMultiStatus                   = http.StatusMultiStatus
	StatusAlreadyReported               = http.StatusAlreadyReported
	StatusIMUsed                        = http.StatusIMUsed
	StatusMultipleChoices               = http.StatusMultipleChoices
	StatusMovedPermanently              = http.StatusMovedPermanently
	StatusFound                         = http.StatusFound
	StatusSeeOther                      = http.StatusSeeOther
	StatusNotModified                   = http.StatusNotModified
	StatusTemporaryRedirect             = http.StatusTemporaryRedirect
	StatusPermanentRedirect             = http.StatusPermanentRedirect
	StatusBadRequest                    = http.StatusBadRequest
	StatusUnauthorized                  = http.StatusUnauthorized
	StatusPaymentRequired               = http.StatusPaymentRequired
	StatusForbidden                     = http.StatusForbidden
	StatusNotFound                      = http.StatusNotFound
	StatusMethodNotAllowed              = http.StatusMethodNotAllowed
	StatusNotAcceptable                 = http.StatusNotAcceptable
	StatusProxyAuthRequired             = http.StatusProxyAuthRequired
	StatusRequestTimeout                = http.StatusRequestTimeout
	StatusConflict                      = http.StatusConflict
	StatusGone                          = http.StatusGone
	StatusLengthRequired                = http.StatusLengthRequired
	StatusPreconditionFailed            = http.StatusPreconditionFailed
	StatusRequestEntityTooLarge         = http.StatusRequestEntityTooLarge
	StatusRequestURITooLong             = http.StatusRequestURITooLong
	StatusUnsupportedMediaType          = http.StatusUnsupportedMediaType
	StatusRequestedRangeNotSatisfiable  = http.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed             = http.StatusExpectationFailed
	StatusTeapot                        = http.StatusTeapot
	StatusMisdirectedRequest            = http.StatusMisdirectedRequest
	StatusUnprocessableEntity           = http.StatusUnprocessableEntity
	StatusLocked                        = http.StatusLocked
	StatusFailedDependency              = http.StatusFailedDependency
	StatusTooEarly                      = http.StatusTooEarly
	StatusUpgradeRequired               = http.StatusUpgradeRequired
	StatusPreconditionRequired          = http.StatusPreconditionRequired
	StatusTooManyRequests               = http.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge   = http.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons    = http.StatusUnavailableForLegalReasons
	StatusInternalServerError           = http.StatusInternalServerError
	StatusNotImplemented                = http.StatusNotImplemented
	StatusBadGateway                    = http.StatusBadGateway
	StatusServiceUnavailable            = http.StatusServiceUnavailable
	StatusGatewayTimeout                = http.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       = http.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         = http.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           = http.StatusInsufficientStorage
	StatusLoopDetected                  = http.StatusLoopDetected
	StatusNotExtended                   = http.StatusNotExtended
	StatusNetworkAuthenticationRequired = http.StatusNetworkAuthenticationRequired
)

// contextKey is a private type for context keys to prevent collisions.
type contextKey struct{ name string }

func (k contextKey) String() string { return "orianna." + k.name }
func (k contextKey) Key() string    { return k.name }

// Typed context keys
var (
	ContextKeyRequestID     = contextKey{"request_id"}
	ContextKeyTraceID       = contextKey{"trace_id"}
	ContextKeyConfig        = contextKey{"server_config"}
	ContextKeyUserID        = contextKey{"user_id"}
	ContextKeyTenantID      = contextKey{"tenant_id"}
	ContextKeyCorrelationID = contextKey{"correlation_id"}
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
