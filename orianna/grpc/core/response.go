// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatusError represents a gRPC error with status code, message, and optional details.
// It wraps a grpc/status.Status and provides builder methods for adding context.
type StatusError struct {
	Code            codes.Code `json:"code"`
	Message         string     `json:"message"`
	InternalMessage string     `json:"-"`
	Cause           error      `json:"-"`
}

// NewStatusError creates a new StatusError with the given code and message.
func NewStatusError(code codes.Code, message string) *StatusError {
	return &StatusError{
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface.
func (e *StatusError) Error() string {
	var b strings.Builder
	b.Grow(len(e.Code.String()) + len(e.Message) + len(e.InternalMessage) + 16)
	b.WriteByte('[')
	b.WriteString(e.Code.String())
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

// Unwrap implements the errors unwrap interface.
func (e *StatusError) Unwrap() error { return e.Cause }

// Is checks if the target error is a StatusError with the same code.
func (e *StatusError) Is(target error) bool {
	t, ok := target.(*StatusError)
	return ok && e.Code == t.Code
}

// GRPCStatus converts the StatusError to a grpc status.Status.
func (e *StatusError) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Message)
}

// ToGRPCError converts the StatusError to a gRPC error suitable for returning from handlers.
func (e *StatusError) ToGRPCError() error {
	return e.GRPCStatus().Err()
}

// WithInternalMsg returns a copy of the StatusError with the internal message set.
// The receiver is not mutated, making this safe for concurrent use.
func (e *StatusError) WithInternalMsg(format string, args ...any) *StatusError {
	clone := *e
	clone.InternalMessage = fmt.Sprintf(format, args...)
	return &clone
}

// WithCause returns a copy of the StatusError with the underlying error cause set.
// The receiver is not mutated, making this safe for concurrent use.
func (e *StatusError) WithCause(err error) *StatusError {
	clone := *e
	clone.Cause = err
	return &clone
}

// IsCode checks if any error in the chain is a StatusError with the given code.
func IsCode(err error, code codes.Code) bool {
	var se *StatusError
	if errors.As(err, &se) {
		return se.Code == code
	}
	st, ok := status.FromError(err)
	if ok {
		return st.Code() == code
	}
	return false
}

// HandleError checks if error is a StatusError and returns the gRPC error.
// Returns the gRPC error and true if it was a known error, false otherwise.
func HandleError(err error) (error, bool) {
	var se *StatusError
	if errors.As(err, &se) {
		return se.ToGRPCError(), true
	}
	return nil, false
}

// WrapError wraps an error with context, converting it to a StatusError.
// Preserves the original error chain for errors.Is/errors.As compatibility.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	var se *StatusError
	if errors.As(err, &se) {
		wrapped := se.WithInternalMsg("%s", message)
		if wrapped.Cause == nil {
			wrapped.Cause = err
		}
		return wrapped
	}
	return &StatusError{
		Code:            codes.Internal,
		Message:         "An unexpected error occurred",
		InternalMessage: message,
		Cause:           err,
	}
}

// WrapErrorf wraps an error with formatted message.
func WrapErrorf(err error, format string, args ...any) error {
	return WrapError(err, fmt.Sprintf(format, args...))
}
