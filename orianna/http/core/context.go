// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

//go:generate mockgen -source=context.go -destination=mocks/mock_context.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/anthanhphan/gosdk/validator"
)

// Context Interface (composed of smaller interfaces for ISP)

// RequestInfo provides read-only access to request information
type RequestInfo interface {
	Method() string
	Path() string
	RoutePath() string
	OriginalURL() string
	BaseURL() string
	Protocol() string
	Hostname() string
	IP() string
	Secure() bool
}

// HeaderManager handles HTTP headers
type HeaderManager interface {
	Get(key string, defaultValue ...string) string
	Set(key, value string)
	Append(field string, values ...string)
	HeadersParser(out any) error
}

// ParamGetter provides route parameter access
type ParamGetter interface {
	Params(key string, defaultValue ...string) string
	AllParams() map[string]string
	ParamsParser(out any) error
}

// QueryGetter provides query string parameter access
type QueryGetter interface {
	Query(key string, defaultValue ...string) string
	AllQueries() map[string]string
	QueryParser(out any) error
}

// BodyReader handles request body parsing
type BodyReader interface {
	Body() []byte
	BodyParser(out any) error
}

// CookieManager handles cookies
type CookieManager interface {
	Cookies(key string, defaultValue ...string) string
	Cookie(cookie *Cookie)
	ClearCookie(key ...string)
}

// ResponseWriter handles response writing
type ResponseWriter interface {
	Status(status int) Context
	ResponseStatusCode() int
	JSON(data any) error
	XML(data any) error
	SendString(s string) error
	SendBytes(b []byte) error
	Redirect(location string, status ...int) error
	// SendStream sets the response body from an io.Reader for streaming large payloads
	// with O(1) memory usage. Optional size parameter sets Content-Length header.
	SendStream(stream io.Reader, size ...int) error
	// SendFile transfers a file from the filesystem as the response.
	// Content-Type is automatically detected from the file extension.
	SendFile(file string) error
}

// ContentNegotiator handles content negotiation
type ContentNegotiator interface {
	Accepts(offers ...string) string
	AcceptsCharsets(offers ...string) string
	AcceptsEncodings(offers ...string) string
	AcceptsLanguages(offers ...string) string
}

// RequestState provides request state information
type RequestState interface {
	Fresh() bool
	Stale() bool
	XHR() bool
}

// LocalsStorage handles request-scoped local storage
type LocalsStorage interface {
	Locals(key string, value ...any) any
	GetAllLocals() map[string]any
}

// ShorthandResponder provides shorthand response methods
type ShorthandResponder interface {
	OK(data any) error
	Created(data any) error
	NoContent() error
	BadRequestMsg(message string) error
	UnauthorizedMsg(message string) error
	ForbiddenMsg(message string) error
	NotFoundMsg(message string) error
	InternalErrorMsg(message string) error
}

// Context defines the interface for request context operations.
// It provides a clean, testable API that is framework-agnostic.
// Context is composed of smaller interfaces following the Interface Segregation Principle,
// allowing consumers to depend only on the methods they need.
type Context interface {
	// Embedded interfaces
	RequestInfo
	HeaderManager
	ParamGetter
	QueryGetter
	BodyReader
	CookieManager
	ResponseWriter
	ContentNegotiator
	RequestState
	LocalsStorage
	ShorthandResponder

	// Middleware flow
	Next() error

	// Access underlying context for advanced use cases
	Context() context.Context

	// SetContext replaces the underlying context (e.g., to inject a timeout/deadline).
	SetContext(ctx context.Context)

	// Utility methods
	IsMethod(method string) bool
	RequestID() string

	// Configuration access
	UseProperHTTPStatus() bool
}

// Validation Helpers

// ValidateAndRespond validates the given value and sends an error response if invalid.
// Returns true if valid, false if invalid (error response already sent).
//
// Example:
//
//	var req CreateUserRequest
//	if err := ctx.BodyParser(&req); err != nil {
//	    return ctx.BadRequest("Invalid body")
//	}
//
//	if ok, err := orianna.ValidateAndRespond(ctx, req); !ok {
//	    return err
//	}
//	// Continue with valid request...
func ValidateAndRespond(ctx Context, v any) (bool, error) {
	if err := validator.Validate(v); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return false, sendValidationErrorResponse(ctx, validationErrs)
		}
		errResp := NewErrorResponse("BAD_REQUEST", StatusBadRequest, "Validation failed")
		return false, SendError(ctx, errResp)
	}
	return true, nil
}

// MustValidate validates a value and panics if invalid.
//
// WARNING: This function panics on failure. Use only in tests or initialization
// code where a panic is acceptable. In production handlers, use validator.Validate instead.
//
// Example:
//
//	config := &MyConfig{Port: 8080}
//	orianna.MustValidate(config)
func MustValidate(v any) {
	if err := validator.Validate(v); err != nil {
		panic(fmt.Sprintf("validation failed: %v", err))
	}
}
