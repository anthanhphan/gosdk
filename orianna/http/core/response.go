// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/logger"
)

// responseLog is a package-level logger for response-related logging.
var responseLog = logger.NewLoggerWithFields(logger.String("package", "http-core"))

// nowUTC returns the current time in UTC.
// Centralizes time.Now().UTC() calls to a single location.
func nowUTC() time.Time { return time.Now().UTC() }

// ============================================================================
// Response Pools — reduce allocations on the hot path
// ============================================================================

var successPool = sync.Pool{
	New: func() any { return &SuccessResponse{} },
}

var errorPool = sync.Pool{
	New: func() any { return &ErrorResponse{} },
}

// AcquireSuccessResponse gets a SuccessResponse from the pool.
// Call ReleaseSuccessResponse when done to return it to the pool.
//
// INTERNAL: This is an optimization for framework-internal hot paths.
// Application code should use NewSuccessResponse instead to avoid
// use-after-release bugs.
func AcquireSuccessResponse(httpStatus int, message string, data any) *SuccessResponse {
	resp := successPool.Get().(*SuccessResponse)
	resp.HTTPStatus = httpStatus
	resp.Code = "SUCCESS"
	resp.Message = message
	resp.Timestamp = nowUTC()
	resp.RequestID = ""
	resp.Data = data
	return resp
}

// ReleaseSuccessResponse returns a SuccessResponse to the pool.
// The response must not be used after calling this.
func ReleaseSuccessResponse(resp *SuccessResponse) {
	if resp == nil {
		return
	}
	resp.Data = nil
	resp.Message = ""
	resp.RequestID = ""
	successPool.Put(resp)
}

// AcquireErrorResponse gets an ErrorResponse from the pool.
// Call ReleaseErrorResponse when done to return it to the pool.
//
// INTERNAL: This is an optimization for framework-internal hot paths.
// Application code should use NewErrorResponse instead to avoid
// use-after-release bugs.
func AcquireErrorResponse(code string, httpStatus int, message string) *ErrorResponse {
	resp := errorPool.Get().(*ErrorResponse)
	resp.HTTPStatus = httpStatus
	resp.Code = code
	resp.Message = message
	resp.Timestamp = nowUTC()
	resp.RequestID = ""
	resp.InternalMessage = ""
	resp.Cause = nil
	resp.Details = nil
	return resp
}

// ReleaseErrorResponse returns an ErrorResponse to the pool.
// The response must not be used after calling this.
func ReleaseErrorResponse(resp *ErrorResponse) {
	if resp == nil {
		return
	}
	resp.Cause = nil
	resp.Details = nil
	resp.InternalMessage = ""
	resp.RequestID = ""
	resp.Message = ""
	resp.Code = ""
	errorPool.Put(resp)
}

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
			Timestamp:  nowUTC(),
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
			Timestamp:  nowUTC(),
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
	b.Grow(len(e.Code) + len(e.Message) + len(e.InternalMessage) + 16)
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

// WithDetails adds a detail field to the ErrorResponse and returns it for chaining.
func (e *ErrorResponse) WithDetails(key string, value any) *ErrorResponse {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithInternalMsg sets the internal message for server-side logging (never sent to client).
func (e *ErrorResponse) WithInternalMsg(format string, args ...any) *ErrorResponse {
	e.InternalMessage = fmt.Sprintf(format, args...)
	return e
}

// WithCause sets the underlying error cause for server-side logging (never sent to client).
func (e *ErrorResponse) WithCause(err error) *ErrorResponse {
	e.Cause = err
	return e
}

// WithRequestID sets the request ID on the ErrorResponse.
func (e *ErrorResponse) WithRequestID(requestID string) *ErrorResponse {
	e.RequestID = requestID
	return e
}

// WithHTTPStatus sets a different HTTP status code on the ErrorResponse.
//
// Example:
//
//	err := core.NewErrorResponse("CONFLICT", 409, "Already exists")
//	return core.SendError(ctx, err.WithHTTPStatus(422))
func (e *ErrorResponse) WithHTTPStatus(status int) *ErrorResponse {
	e.HTTPStatus = status
	return e
}

// SendSuccess sends a SuccessResponse as JSON or XML based on Accept header.
// Sets request ID and timestamp if not already set.
// Fast path: skips content-negotiation when Accept is empty or application/json (99% of API calls).
func SendSuccess(ctx Context, resp *SuccessResponse) error {
	if resp.RequestID == "" {
		resp.RequestID = ctx.RequestID()
	}
	if resp.Timestamp.IsZero() {
		resp.Timestamp = nowUTC()
	}

	// JSON fast-path — skip content negotiation for the common case
	accept := ctx.Get("Accept")
	if accept == "" || accept == "application/json" {
		return ctx.Status(resp.HTTPStatus).JSON(resp)
	}
	if ctx.Accepts("application/json", "application/xml") == "application/xml" {
		return ctx.Status(resp.HTTPStatus).XML(resp)
	}
	return ctx.Status(resp.HTTPStatus).JSON(resp)
}

// SendError sends an ErrorResponse as JSON or XML based on Accept header.
// Sets request ID and timestamp if not already set.
// Logs internal details (InternalMessage, Cause) server-side for audit.
func SendError(ctx Context, err *ErrorResponse) error {
	if err.RequestID == "" {
		err.RequestID = ctx.RequestID()
	}
	if err.Timestamp.IsZero() {
		err.Timestamp = nowUTC()
	}

	// Audit: log internal details server-side (never sent to client).
	if err.InternalMessage != "" || err.Cause != nil {
		responseLog.Warnw("error response",
			"request_id", err.RequestID,
			"error_code", err.Code,
			"http_status", err.HTTPStatus,
			"message", err.Message,
			"internal_message", err.InternalMessage,
			"cause", errorString(err.Cause),
		)
	}

	status := StatusOK
	if ctx.UseProperHTTPStatus() {
		status = err.HTTPStatus
	}

	// JSON fast-path — skip content negotiation for the common case
	accept := ctx.Get("Accept")
	if accept == "" || accept == "application/json" {
		return ctx.Status(status).JSON(err)
	}
	if ctx.Accepts("application/json", "application/xml") == "application/xml" {
		return ctx.Status(status).XML(err)
	}
	return ctx.Status(status).JSON(err)
}

// errorString safely converts an error to string, returning empty for nil.
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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
// Preserves the original error chain for errors.Is/errors.As compatibility.
// Creates a shallow copy to avoid mutating the original ErrorResponse.
// The original error is always preserved as the Cause for proper Unwrap() chains.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		clone := *errResp
		clone.InternalMessage = message
		// Always preserve the original error as cause.
		// If errResp already has a Cause, keep it (deeper root).
		// Otherwise, wrap the incoming error itself.
		if clone.Cause == nil {
			clone.Cause = err
		}
		// Deep copy Details to prevent shared mutation between original and clone
		if errResp.Details != nil {
			clone.Details = make(map[string]any, len(errResp.Details))
			for k, v := range errResp.Details {
				clone.Details[k] = v
			}
		}
		return &clone
	}
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			HTTPStatus: StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "An unexpected error occurred",
			Timestamp:  nowUTC(),
		},
		InternalMessage: message,
		Cause:           err,
	}
}

// WrapErrorf wraps an error with formatted message.
func WrapErrorf(err error, format string, args ...any) error {
	return WrapError(err, fmt.Sprintf(format, args...))
}
