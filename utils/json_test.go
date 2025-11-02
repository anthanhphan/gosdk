// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"encoding/json"
	"testing"
)

// TestMarshalCompact_JSONString tests MarshalCompact with JSON string inputs
func TestMarshalCompact_JSONString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result string, err error)
	}{
		{
			name:    "empty string should return empty string",
			input:   "",
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result != `""` {
					t.Errorf("MarshalCompact() = %v, want %v", result, `""`)
				}
			},
		},
		{
			name:    "valid JSON string should be formatted and sorted",
			input:   `{"name":"john","age":30}`,
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result != `{"age":30,"name":"john"}` {
					t.Errorf("MarshalCompact() = %v, want %v", result, `{"age":30,"name":"john"}`)
				}
			},
		},
		{
			name:    "invalid JSON should return error",
			input:   "not json",
			wantErr: true,
			check: func(t *testing.T, result string, err error) {
				if err == nil {
					t.Error("MarshalCompact() error = nil, want error")
				}
			},
		},
		{
			name:    "JSON with nested objects should be formatted",
			input:   `{"user":{"name":"john","age":30},"active":true}`,
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result == "" {
					t.Error("MarshalCompact() result should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalCompact(tt.input)
			tt.check(t, result, err)
		})
	}
}

// TestUnmarshal tests the Unmarshal function with various inputs
func TestUnmarshal(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		input   string
		target  User
		wantErr bool
		check   func(t *testing.T, result User, err error)
	}{
		{
			name:    "empty string should return zero value",
			input:   "",
			target:  User{},
			wantErr: false,
			check: func(t *testing.T, result User, err error) {
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
				if result.Name != "" || result.Age != 0 {
					t.Error("Unmarshal() result should be zero value")
				}
			},
		},
		{
			name:    "valid JSON should parse into struct",
			input:   `{"name":"john","age":30}`,
			target:  User{},
			wantErr: false,
			check: func(t *testing.T, result User, err error) {
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
				if result.Name != "john" {
					t.Errorf("Unmarshal() user.Name = %v, want %v", result.Name, "john")
				}
				if result.Age != 30 {
					t.Errorf("Unmarshal() user.Age = %v, want %v", result.Age, 30)
				}
			},
		},
		{
			name:    "invalid JSON should return error",
			input:   "not json",
			target:  User{},
			wantErr: true,
			check: func(t *testing.T, result User, err error) {
				if err == nil {
					t.Error("Unmarshal() error = nil, want error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Unmarshal(tt.input, tt.target)
			tt.check(t, result, err)
		})
	}
}

// TestMarshalCompact_Struct tests MarshalCompact with struct inputs
func TestMarshalCompact_Struct(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(t *testing.T, result string, err error)
	}{
		{
			name:    "valid struct should marshal to compact JSON",
			input:   User{Name: "john", Age: 30},
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result == "" {
					t.Error("MarshalCompact() result should not be empty")
				}
				// Unmarshal and verify content (keys order may vary)
				var user User
				if err := json.Unmarshal([]byte(result), &user); err != nil {
					t.Errorf("MarshalCompact() result is not valid JSON: %v", err)
				}
				if user.Name != "john" || user.Age != 30 {
					t.Errorf("MarshalCompact() unmarshaled user = %+v, want {Name:john Age:30}", user)
				}
			},
		},
		{
			name:    "empty struct should marshal to empty JSON",
			input:   User{},
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result == "" {
					t.Error("MarshalCompact() result should not be empty")
				}
				// Unmarshal and verify content (keys order may vary)
				var user User
				if err := json.Unmarshal([]byte(result), &user); err != nil {
					t.Errorf("MarshalCompact() result is not valid JSON: %v", err)
				}
				if user.Name != "" || user.Age != 0 {
					t.Errorf("MarshalCompact() unmarshaled user = %+v, want zero value", user)
				}
			},
		},
		{
			name:    "map should marshal to compact JSON",
			input:   map[string]interface{}{"key": "value", "number": 42},
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result == "" {
					t.Error("MarshalCompact() result should not be empty")
				}
			},
		},
		{
			name:    "slice should marshal to compact JSON",
			input:   []string{"a", "b", "c"},
			wantErr: false,
			check: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("MarshalCompact() error = %v, want nil", err)
				}
				if result != `["a","b","c"]` {
					t.Errorf("MarshalCompact() = %v, want %v", result, `["a","b","c"]`)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalCompact(tt.input)
			tt.check(t, result, err)
		})
	}
}
