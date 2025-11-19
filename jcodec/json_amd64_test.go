// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

//go:build amd64 || 386

package jcodec

import "testing"

// TestNewEngineForArch_AMD64 tests the AMD64/386 architecture paths.
// This test only runs on AMD64 and 386 architectures.
func TestNewEngineForArch_AMD64(t *testing.T) {
	tests := []struct {
		name  string
		arch  string
		check func(t *testing.T, engine engine)
	}{
		{
			name: "amd64 architecture should return sonic engine",
			arch: "amd64",
			check: func(t *testing.T, e engine) {
				if e == nil {
					t.Error("Engine should not be nil")
				}
				// Test that it works
				data, err := e.Marshal(testUser{ID: 1, Name: "Test"})
				if err != nil {
					t.Errorf("Engine Marshal should work: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name: "386 architecture should return sonic engine",
			arch: "386",
			check: func(t *testing.T, e engine) {
				if e == nil {
					t.Error("Engine should not be nil")
				}
				// Test that it works
				data, err := e.Marshal(testUser{ID: 1, Name: "Test"})
				if err != nil {
					t.Errorf("Engine Marshal should work: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := newEngineForArch(tt.arch)
			tt.check(t, engine)
		})
	}
}
