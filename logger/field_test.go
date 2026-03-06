// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"errors"
	"math"
	"testing"
)

func TestInt64(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		val   int64
		check func(t *testing.T, field Field)
	}{
		{
			name: "positive int64 should create field correctly",
			key:  "count",
			val:  12345,
			check: func(t *testing.T, field Field) {
				if field.Key != "count" {
					t.Errorf("Int64() Key = %v, want %v", field.Key, "count")
				}
				if field.Type != FieldTypeInt64 {
					t.Errorf("Int64() Type = %v, want FieldTypeInt64", field.Type)
				}
				if field.Integer != 12345 {
					t.Errorf("Int64() Integer = %v, want %v", field.Integer, int64(12345))
				}
			},
		},
		{
			name: "negative int64 should create field correctly",
			key:  "balance",
			val:  -100,
			check: func(t *testing.T, field Field) {
				if field.Key != "balance" {
					t.Errorf("Int64() Key = %v, want %v", field.Key, "balance")
				}
				if field.Integer != -100 {
					t.Errorf("Int64() Integer = %v, want %v", field.Integer, int64(-100))
				}
			},
		},
		{
			name: "zero int64 should create field correctly",
			key:  "zero",
			val:  0,
			check: func(t *testing.T, field Field) {
				if field.Key != "zero" {
					t.Errorf("Int64() Key = %v, want %v", field.Key, "zero")
				}
				if field.Integer != 0 {
					t.Errorf("Int64() Integer = %v, want %v", field.Integer, int64(0))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int64(tt.key, tt.val)
			tt.check(t, result)
		})
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		val   float64
		check func(t *testing.T, field Field)
	}{
		{
			name: "positive float64 should create field correctly",
			key:  "price",
			val:  99.99,
			check: func(t *testing.T, field Field) {
				if field.Key != "price" {
					t.Errorf("Float64() Key = %v, want %v", field.Key, "price")
				}
				if field.Type != FieldTypeFloat64 {
					t.Errorf("Float64() Type = %v, want FieldTypeFloat64", field.Type)
				}
				got := math.Float64frombits(uint64(field.Integer))
				if got != 99.99 {
					t.Errorf("Float64() decoded = %v, want %v", got, 99.99)
				}
			},
		},
		{
			name: "negative float64 should create field correctly",
			key:  "temperature",
			val:  -273.15,
			check: func(t *testing.T, field Field) {
				if field.Key != "temperature" {
					t.Errorf("Float64() Key = %v, want %v", field.Key, "temperature")
				}
				got := math.Float64frombits(uint64(field.Integer))
				if got != -273.15 {
					t.Errorf("Float64() decoded = %v, want %v", got, -273.15)
				}
			},
		},
		{
			name: "zero float64 should create field correctly",
			key:  "zero",
			val:  0.0,
			check: func(t *testing.T, field Field) {
				if field.Key != "zero" {
					t.Errorf("Float64() Key = %v, want %v", field.Key, "zero")
				}
				got := math.Float64frombits(uint64(field.Integer))
				if got != 0.0 {
					t.Errorf("Float64() decoded = %v, want %v", got, 0.0)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float64(tt.key, tt.val)
			tt.check(t, result)
		})
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		val   bool
		check func(t *testing.T, field Field)
	}{
		{
			name: "true bool should create field correctly",
			key:  "enabled",
			val:  true,
			check: func(t *testing.T, field Field) {
				if field.Key != "enabled" {
					t.Errorf("Bool() Key = %v, want %v", field.Key, "enabled")
				}
				if field.Type != FieldTypeBool {
					t.Errorf("Bool() Type = %v, want FieldTypeBool", field.Type)
				}
				if field.Integer != 1 {
					t.Errorf("Bool() Integer = %v, want 1 (true)", field.Integer)
				}
			},
		},
		{
			name: "false bool should create field correctly",
			key:  "disabled",
			val:  false,
			check: func(t *testing.T, field Field) {
				if field.Key != "disabled" {
					t.Errorf("Bool() Key = %v, want %v", field.Key, "disabled")
				}
				if field.Integer != 0 {
					t.Errorf("Bool() Integer = %v, want 0 (false)", field.Integer)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bool(tt.key, tt.val)
			tt.check(t, result)
		})
	}
}

func TestErrorField(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check func(t *testing.T, field Field)
	}{
		{
			name: "non-nil error should create string field with error message",
			err:  errors.New("test error"),
			check: func(t *testing.T, field Field) {
				if field.Key != "error" {
					t.Errorf("ErrorField() Key = %v, want %v", field.Key, "error")
				}
				if field.Type != FieldTypeString {
					t.Errorf("ErrorField() Type = %v, want FieldTypeString", field.Type)
				}
				if field.Str != "test error" {
					t.Errorf("ErrorField() Str = %v, want %v", field.Str, "test error")
				}
			},
		},
		{
			name: "nil error should create field with nil iface",
			err:  nil,
			check: func(t *testing.T, field Field) {
				if field.Key != "error" {
					t.Errorf("ErrorField() Key = %v, want %v", field.Key, "error")
				}
				if field.Type != FieldTypeAny {
					t.Errorf("ErrorField() Type = %v, want FieldTypeAny", field.Type)
				}
				if field.Iface != nil {
					t.Errorf("ErrorField() Iface = %v, want nil", field.Iface)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorField(tt.err)
			tt.check(t, result)
		})
	}
}
