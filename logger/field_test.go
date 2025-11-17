// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"errors"
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
				if field.Value != int64(12345) {
					t.Errorf("Int64() Value = %v, want %v", field.Value, int64(12345))
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
				if field.Value != int64(-100) {
					t.Errorf("Int64() Value = %v, want %v", field.Value, int64(-100))
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
				if field.Value != int64(0) {
					t.Errorf("Int64() Value = %v, want %v", field.Value, int64(0))
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
				if field.Value != 99.99 {
					t.Errorf("Float64() Value = %v, want %v", field.Value, 99.99)
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
				if field.Value != -273.15 {
					t.Errorf("Float64() Value = %v, want %v", field.Value, -273.15)
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
				if field.Value != 0.0 {
					t.Errorf("Float64() Value = %v, want %v", field.Value, 0.0)
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
				if field.Value != true {
					t.Errorf("Bool() Value = %v, want %v", field.Value, true)
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
				if field.Value != false {
					t.Errorf("Bool() Value = %v, want %v", field.Value, false)
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
			name: "non-nil error should create field with error message",
			err:  errors.New("test error"),
			check: func(t *testing.T, field Field) {
				if field.Key != "error" {
					t.Errorf("ErrorField() Key = %v, want %v", field.Key, "error")
				}
				if field.Value != "test error" {
					t.Errorf("ErrorField() Value = %v, want %v", field.Value, "test error")
				}
			},
		},
		{
			name: "nil error should create field with nil value",
			err:  nil,
			check: func(t *testing.T, field Field) {
				if field.Key != "error" {
					t.Errorf("ErrorField() Key = %v, want %v", field.Key, "error")
				}
				if field.Value != nil {
					t.Errorf("ErrorField() Value = %v, want %v", field.Value, nil)
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
