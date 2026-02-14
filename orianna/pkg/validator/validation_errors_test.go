// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name  string
		err   *ValidationError
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name: "error message should contain field and message",
			err: &ValidationError{
				Field:   "Email",
				Message: "is required",
			},
			check: func(t *testing.T, got string) {
				if got == "" {
					t.Error("Error() should not return empty string")
				}
				if got != "Email: is required" {
					t.Errorf("Error() = %v, want 'Email: is required'", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			tt.check(t, got)
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name  string
		errs  ValidationErrors
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name: "empty errors should return empty string",
			errs: ValidationErrors{},
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Error() = %v, want empty string", got)
				}
			},
		},
		{
			name: "multiple errors should be joined",
			errs: ValidationErrors{
				{Field: "Email", Message: "is required"},
				{Field: "Name", Message: "must be at least 3 characters"},
			},
			check: func(t *testing.T, got string) {
				if got == "" {
					t.Error("Error() should not return empty string")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.errs.Error()
			tt.check(t, got)
		})
	}
}

func TestValidationErrors_ToArray(t *testing.T) {
	tests := []struct {
		name  string
		errs  ValidationErrors
		check func(t *testing.T, result []map[string]string)
	}{
		{
			name: "should convert errors to array format",
			errs: ValidationErrors{
				{Field: "Email", Message: "is required"},
				{Field: "Name", Message: "must be at least 3 characters"},
			},
			check: func(t *testing.T, result []map[string]string) {
				if len(result) != 2 {
					t.Errorf("ToArray() length = %v, want 2", len(result))
				}
				if result[0]["field"] != "email" {
					t.Errorf("ToArray() first field = %v, want 'email'", result[0]["field"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errs.ToArray()
			tt.check(t, result)
		})
	}
}
