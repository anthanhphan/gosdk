// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileSecurely(t *testing.T) {
	// Setup temp directory
	tempDir, err := os.MkdirTemp("", "io_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create test file
	testFile := "test.txt"
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create subdirectory file
	subDir := "subdir"
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	subFile := filepath.Join(subDir, "subfile.txt")
	subContent := "Subdirectory content"
	if err := os.WriteFile(subFile, []byte(subContent), 0644); err != nil {
		t.Fatalf("Failed to create subdirectory file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{"valid file should return content", testFile, testContent, false},
		{"file in subdirectory should return content", subFile, subContent, false},
		{"non-existent file should return error", "nonexistent.txt", "", true},
		{"directory traversal with .. should return error", "../test.txt", "", true},
		{"absolute path should return error", "/etc/passwd", "", true},
		{"path with .. in middle should work after cleaning", "subdir/../test.txt", testContent, false},
		{"empty path should return error", "", "", true},
		{"path starting with dot should work", "./test.txt", testContent, false},
		{"path with multiple slashes should return error", "//test.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ReadFileSecurely(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if string(result) != tt.want {
					t.Errorf("ReadFileSecurely() = %v, want %v", string(result), tt.want)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.path == "nonexistent.txt" && !strings.Contains(err.Error(), "no such file or directory") {
					t.Errorf("Error message = %v, want to contain 'no such file or directory'", err.Error())
				}
				if (tt.path == "../test.txt" || tt.path == "/etc/passwd" || tt.path == "//test.txt") && !strings.Contains(err.Error(), "directory traversal not allowed") {
					t.Errorf("Error message = %v, want to contain 'directory traversal not allowed'", err.Error())
				}
			}
		})
	}

	// Test working directory error by removing current directory
	os.RemoveAll(tempDir)
	result, err := ReadFileSecurely("test.txt")
	if err == nil {
		t.Error("Expected error but got none")
	}
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Error message = %v, want to contain 'no such file or directory'", err.Error())
	}
	if result != nil {
		t.Error("Expected nil result")
	}
}
