// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package requestid

import (
	"strings"
	"testing"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		// Valid cases
		{"uuid v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"uuid v7", "0192f45a-c67a-7def-8012-3456789abcdf", true},
		{"alphanumeric", "abc123", true},
		{"with dashes", "request-id-123", true},
		{"with underscores", "request_id_123", true},
		{"mixed", "Req-ID_v7-abc123", true},
		{"single char", "a", true},
		{"max length", strings.Repeat("a", MaxLength), true},

		// Invalid cases
		{"empty", "", false},
		{"too long", strings.Repeat("a", MaxLength+1), false},
		{"contains space", "request id", false},
		{"contains newline", "request\nid", false},
		{"contains tab", "request\tid", false},
		{"contains dot", "request.id", false},
		{"contains colon", "request:id", false},
		{"contains slash", "request/id", false},
		{"contains backslash", "request\\id", false},
		{"json injection", `{"key":"value"}`, false},
		{"log injection newline", "valid-id\n{\"injected\":true}", false},
		{"null byte", "valid\x00id", false},
		{"unicode", "réquest-ïd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValid(tt.id)
			if got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	id := Generate()
	if id == "" {
		t.Error("Generate() returned empty string")
	}
	if !IsValid(id) {
		t.Errorf("Generate() = %q, but IsValid returned false", id)
	}

	// Should generate unique IDs
	id2 := Generate()
	if id == id2 {
		t.Errorf("Generate() returned same ID twice: %s", id)
	}
}

func BenchmarkIsValid(b *testing.B) {
	id := "0192f45a-c67a-7def-8012-3456789abcdf"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValid(id)
	}
}
