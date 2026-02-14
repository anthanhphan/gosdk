// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"fmt"
	"testing"
)

// ErrorResponse Constructor Tests

func TestNewErrorResponse(t *testing.T) {
	err := NewErrorResponse("NOT_FOUND", StatusNotFound, "Resource not found")

	if err.Code != "NOT_FOUND" {
		t.Errorf("Code = %s, want NOT_FOUND", err.Code)
	}
	if err.HTTPStatus != StatusNotFound {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, StatusNotFound)
	}
	if err.Message != "Resource not found" {
		t.Errorf("Message = %s, want 'Resource not found'", err.Message)
	}
	if err.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := NewSuccessResponse(StatusOK, "Success", data)

	if resp.Code != "SUCCESS" {
		t.Errorf("Code = %s, want SUCCESS", resp.Code)
	}
	if resp.HTTPStatus != StatusOK {
		t.Errorf("HTTPStatus = %d, want %d", resp.HTTPStatus, StatusOK)
	}
	if resp.Message != "Success" {
		t.Errorf("Message = %s, want 'Success'", resp.Message)
	}
	if resp.Data == nil {
		t.Error("Data should not be nil")
	}
}

// ErrorResponse Error Interface Tests

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ErrorResponse
		contains []string
	}{
		{
			name:     "basic error",
			err:      NewErrorResponse("BAD_REQUEST", StatusBadRequest, "Bad request"),
			contains: []string{"[BAD_REQUEST]", "Bad request"},
		},
		{
			name: "with internal message",
			err: NewErrorResponse("INTERNAL_ERROR", StatusInternalServerError, "Error").
				WithInternalMsg("detailed info"),
			contains: []string{"[INTERNAL_ERROR]", "Error", "internal: detailed info"},
		},
		{
			name: "with cause",
			err: NewErrorResponse("DB_ERROR", StatusInternalServerError, "Database error").
				WithCause(errors.New("connection refused")),
			contains: []string{"[DB_ERROR]", "connection refused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			for _, want := range tt.contains {
				if !containsString(msg, want) {
					t.Errorf("Error() = %q, should contain %q", msg, want)
				}
			}
		})
	}
}

// Fluent Builder Tests

func TestErrorResponse_FluentBuilders(t *testing.T) {
	cause := errors.New("root cause")
	err := NewErrorResponse("TEST", 400, "test").
		WithDetails("field", "name").
		WithDetails("reason", "too short").
		WithInternalMsg("user %s failed validation", "john").
		WithCause(cause).
		WithRequestID("req-123")

	if err.Details["field"] != "name" {
		t.Errorf("Details[field] = %v, want 'name'", err.Details["field"])
	}
	if err.Details["reason"] != "too short" {
		t.Errorf("Details[reason] = %v, want 'too short'", err.Details["reason"])
	}
	if err.InternalMessage != "user john failed validation" {
		t.Errorf("InternalMessage = %s, want 'user john failed validation'", err.InternalMessage)
	}
	if !errors.Is(err.Cause, cause) {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
	if err.RequestID != "req-123" {
		t.Errorf("RequestID = %s, want 'req-123'", err.RequestID)
	}
}

// Error Matching Tests

func TestErrorResponse_Is(t *testing.T) {
	err1 := NewErrorResponse("NOT_FOUND", 404, "Not found")
	err2 := NewErrorResponse("NOT_FOUND", 404, "Different message")
	err3 := NewErrorResponse("BAD_REQUEST", 400, "Not found")

	if !errors.Is(err1, err2) {
		t.Error("errors.Is should match errors with same code")
	}
	if errors.Is(err1, err3) {
		t.Error("errors.Is should not match errors with different codes")
	}
}

func TestErrorResponse_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := NewErrorResponse("TEST", 500, "test").WithCause(cause)

	unwrapped := errors.Unwrap(err)
	if !errors.Is(unwrapped, cause) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestErrorResponse_Unwrap_nil(t *testing.T) {
	err := NewErrorResponse("TEST", 500, "test")
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

// IsErrorCode Tests

func TestIsErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{
			name: "matching code",
			err:  NewErrorResponse("NOT_FOUND", 404, "not found"),
			code: "NOT_FOUND",
			want: true,
		},
		{
			name: "non-matching code",
			err:  NewErrorResponse("NOT_FOUND", 404, "not found"),
			code: "BAD_REQUEST",
			want: false,
		},
		{
			name: "wrapped ErrorResponse",
			err:  fmt.Errorf("context: %w", NewErrorResponse("TIMEOUT", 504, "timeout")),
			code: "TIMEOUT",
			want: true,
		},
		{
			name: "plain error",
			err:  errors.New("plain error"),
			code: "NOT_FOUND",
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			code: "NOT_FOUND",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsErrorCode(tt.err, tt.code)
			if got != tt.want {
				t.Errorf("IsErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// WrapError Tests

func TestWrapError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := WrapError(nil, "context")
		if result != nil {
			t.Errorf("WrapError(nil) = %v, want nil", result)
		}
	})

	t.Run("wraps ErrorResponse with internal message", func(t *testing.T) {
		original := NewErrorResponse("BAD_REQUEST", 400, "Bad request")
		result := WrapError(original, "user service")

		var errResp *ErrorResponse
		if !errors.As(result, &errResp) {
			t.Fatal("WrapError should return ErrorResponse")
		}
		if errResp.Code != "BAD_REQUEST" {
			t.Errorf("Code = %s, want BAD_REQUEST", errResp.Code)
		}
		if errResp.InternalMessage != "user service" {
			t.Errorf("InternalMessage = %s, want 'user service'", errResp.InternalMessage)
		}
	})

	t.Run("wraps plain error into ErrorResponse", func(t *testing.T) {
		plainErr := errors.New("db connection failed")
		result := WrapError(plainErr, "user repository")

		var errResp *ErrorResponse
		if !errors.As(result, &errResp) {
			t.Fatal("WrapError should return ErrorResponse for plain errors")
		}
		if errResp.Code != "INTERNAL_ERROR" {
			t.Errorf("Code = %s, want INTERNAL_ERROR", errResp.Code)
		}
		if errResp.HTTPStatus != StatusInternalServerError {
			t.Errorf("HTTPStatus = %d, want %d", errResp.HTTPStatus, StatusInternalServerError)
		}
		if errResp.InternalMessage != "user repository" {
			t.Errorf("InternalMessage = %s, want 'user repository'", errResp.InternalMessage)
		}
		if !errors.Is(errResp.Cause, plainErr) {
			t.Errorf("Cause = %v, want %v", errResp.Cause, plainErr)
		}
	})
}

func TestWrapErrorf(t *testing.T) {
	err := errors.New("timeout")
	result := WrapErrorf(err, "failed to call %s after %d ms", "user-service", 500)

	var errResp *ErrorResponse
	if !errors.As(result, &errResp) {
		t.Fatal("WrapErrorf should return ErrorResponse")
	}
	if errResp.InternalMessage != "failed to call user-service after 500 ms" {
		t.Errorf("InternalMessage = %s, want formatted message", errResp.InternalMessage)
	}
}

// Helpers

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
