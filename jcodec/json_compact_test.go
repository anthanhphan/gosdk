// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
	"testing"
)

func TestCompactString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "valid json string should be compacted",
			input:   `{  "id":  1,  "name":  "John"  }`,
			want:    `{"id":1,"name":"John"}`,
			wantErr: false,
		},
		{
			name:    "valid json string with newlines should be compacted",
			input:   "{\n  \"id\": 1,\n  \"name\": \"John\"\n}",
			want:    `{"id":1,"name":"John"}`,
			wantErr: false,
		},
		{
			name:    "invalid json string should return error",
			input:   `{ "id": 1, "name": }`,
			wantErr: true,
		},
		{
			name:    "struct input should be marshaled",
			input:   testUser{ID: 1, Name: "John", Email: "john@example.com", Active: true},
			want:    `{"id":1,"name":"John","email":"john@example.com","created_at":"0001-01-01T00:00:00Z","active":true}`,
			wantErr: false,
		},
		{
			name:    "primitive int input should be marshaled",
			input:   123,
			want:    "123",
			wantErr: false,
		},
		{
			name:    "primitive string input (not json) should be treated as json string and fail if invalid json",
			input:   "not json",
			wantErr: true,
		},
		{
			name:    "primitive quoted string input (valid json string) should be compacted",
			input:   `"hello"`,
			want:    `"hello"`,
			wantErr: false,
		},
		{
			name:    "nil input should be marshaled to null",
			input:   nil,
			want:    "null",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompactString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompactString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CompactString() = %v, want %v", got, tt.want)
			}
		})
	}
}
