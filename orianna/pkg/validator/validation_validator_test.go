// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
		check   func(t *testing.T, err error)
	}{
		{
			name: "valid struct should not return error",
			input: struct {
				Name  string `validate:"required,min=3"`
				Email string `validate:"required,email"`
			}{
				Name:  "John",
				Email: "john@example.com",
			},
			wantErr: false,
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			},
		},
		{
			name: "missing required field should return error",
			input: struct {
				Name  string `validate:"required"`
				Email string `validate:"required"`
			}{
				Name: "John",
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for missing required field")
				}
				if validationErr, ok := err.(ValidationErrors); ok {
					if len(validationErr) == 0 {
						t.Error("ValidationErrors should not be empty")
					}
				}
			},
		},
		{
			name:    "nil pointer should return error",
			input:   (*struct{ Name string })(nil),
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for nil pointer")
				}
			},
		},
		{
			name:    "non-struct type should return error",
			input:   "not a struct",
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for non-struct type")
				}
			},
		},
		{
			name: "struct with no validation tags should not return error",
			input: struct {
				Name  string
				Email string
			}{
				Name:  "John",
				Email: "john@example.com",
			},
			wantErr: false,
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.check(t, err)
		})
	}
}
