// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// BaseResponse contains common fields for all API responses.
type BaseResponse struct {
	HTTPStatus int       `json:"http_status"`
	Code       string    `json:"code"`
	Message    string    `json:"message,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id,omitempty"`
}

// SuccessResponse represents a standard successful API response.
type SuccessResponse struct {
	BaseResponse
	Data any `json:"data,omitempty"`
}

// ErrorResponse represents a standard error API response.
type ErrorResponse struct {
	BaseResponse
	InternalMessage string         `json:"-"`
	Cause           error          `json:"-"`
	Details         map[string]any `json:"details,omitempty"`
}

// Map is a convenient alias for map[string]any.
type Map map[string]any

// NewSuccessResponse creates a new SuccessResponse.
//
// Input:
//   - httpStatus: HTTP status code
//   - message: Optional success message
//   - data: Response data
//
// Output:
//   - *SuccessResponse: The created response
//
// Example:
//
//	resp := core.NewSuccessResponse(200, "User created", user)
//	return core.SendSuccess(ctx, resp)
func NewSuccessResponse(httpStatus int, message string, data any) *SuccessResponse {
	return &SuccessResponse{
		BaseResponse: BaseResponse{
			HTTPStatus: httpStatus,
			Code:       "SUCCESS",
			Message:    message,
			Timestamp:  time.Now().UTC(),
		},
		Data: data,
	}
}

// NewErrorResponse creates a new ErrorResponse.
//
// Input:
//   - code: Error code
//   - httpStatus: HTTP status code
//   - message: Error message
//
// Output:
//   - *ErrorResponse: The created error response
//
// Example:
//
//	err := core.NewErrorResponse("NOT_FOUND", 404, "User not found")
//	return core.SendError(ctx, err)
func NewErrorResponse(code string, httpStatus int, message string) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			HTTPStatus: httpStatus,
			Code:       code,
			Message:    message,
			Timestamp:  time.Now().UTC(),
		},
	}
}

// IsErrorCode checks if any error in the chain is an ErrorResponse with the given code.
func IsErrorCode(err error, code string) bool {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		return errResp.Code == code
	}
	return false
}

func (e *ErrorResponse) Error() string {
	var b strings.Builder
	b.WriteByte('[')
	b.WriteString(e.Code)
	b.WriteString("] ")
	b.WriteString(e.Message)
	if e.InternalMessage != "" {
		b.WriteString(" (internal: ")
		b.WriteString(e.InternalMessage)
		b.WriteByte(')')
	}
	if e.Cause != nil {
		b.WriteString(": ")
		b.WriteString(e.Cause.Error())
	}
	return b.String()
}

func (e *ErrorResponse) Unwrap() error { return e.Cause }

func (e *ErrorResponse) Is(target error) bool {
	t, ok := target.(*ErrorResponse)
	return ok && e.Code == t.Code
}

// WithDetails adds a detail field to the error response.
func (e *ErrorResponse) WithDetails(key string, value any) *ErrorResponse {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithInternalMsg sets the internal message for logging.
func (e *ErrorResponse) WithInternalMsg(format string, args ...any) *ErrorResponse {
	e.InternalMessage = fmt.Sprintf(format, args...)
	return e
}

// WithCause sets the underlying error cause.
func (e *ErrorResponse) WithCause(err error) *ErrorResponse {
	e.Cause = err
	return e
}

// WithRequestID sets the request ID.
func (e *ErrorResponse) WithRequestID(requestID string) *ErrorResponse {
	e.RequestID = requestID
	return e
}

// SendSuccess sends a SuccessResponse as JSON.
//
// Input:
//   - ctx: Request context
//   - resp: Success response to send
//
// Output:
//   - error: Error if sending fails
//
// Example:
//
//	resp := core.NewSuccessResponse(200, "OK", data)
//	return core.SendSuccess(ctx, resp)
func SendSuccess(ctx Context, resp *SuccessResponse) error {
	if resp.RequestID == "" {
		resp.RequestID = ctx.RequestID()
	}
	if resp.Timestamp.IsZero() {
		resp.Timestamp = time.Now().UTC()
	}
	return ctx.Status(resp.HTTPStatus).JSON(resp)
}

// SendError sends an ErrorResponse as JSON.
//
// Input:
//   - ctx: Request context
//   - err: Error response to send
//
// Output:
//   - error: Error if sending fails
//
// Example:
//
//	err := core.NewErrorResponse("NOT_FOUND", 404, "Not found")
//	return core.SendError(ctx, err)
func SendError(ctx Context, err *ErrorResponse) error {
	if err.RequestID == "" {
		err.RequestID = ctx.RequestID()
	}
	if err.Timestamp.IsZero() {
		err.Timestamp = time.Now().UTC()
	}

	status := StatusOK
	if ctx.UseProperHTTPStatus() {
		status = err.HTTPStatus
	}
	return ctx.Status(status).JSON(err)
}

// HandleError checks if error is an ErrorResponse and sends it.
func HandleError(ctx Context, err error) bool {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		_ = SendError(ctx, errResp)
		return true
	}
	return false
}

// WrapError wraps an error with context.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		clone := *errResp
		return clone.WithInternalMsg("%s", message)
	}
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			HTTPStatus: StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "An unexpected error occurred",
			Timestamp:  time.Now().UTC(),
		},
		InternalMessage: message,
		Cause:           err,
	}
}

// WrapErrorf wraps an error with formatted message.
func WrapErrorf(err error, format string, args ...any) error {
	return WrapError(err, fmt.Sprintf(format, args...))
}
