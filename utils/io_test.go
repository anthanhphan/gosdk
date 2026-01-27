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
	defer func() { _ = os.RemoveAll(tempDir) }()

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
	_ = os.RemoveAll(tempDir)
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

func TestOpenFileSecurely(t *testing.T) {
	// Setup temp directory
	tempDir, err := os.MkdirTemp("", "io_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

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
	testFile := "test.log"
	testContent := "test content"

	tests := []struct {
		name    string
		path    string
		flag    int
		perm    os.FileMode
		wantErr bool
		check   func(t *testing.T, file *os.File, err error)
	}{
		{
			name:    "valid file should open successfully",
			path:    testFile,
			flag:    os.O_CREATE | os.O_WRONLY | os.O_TRUNC,
			perm:    0600,
			wantErr: false,
			check: func(t *testing.T, file *os.File, err error) {
				if err != nil {
					t.Errorf("OpenFileSecurely() unexpected error: %v", err)
					return
				}
				if file == nil {
					t.Error("OpenFileSecurely() should not return nil file")
					return
				}
				_ = file.Close()
			},
		},
		{
			name:    "file in subdirectory should open successfully",
			path:    filepath.Join("subdir", "subfile.log"),
			flag:    os.O_CREATE | os.O_WRONLY | os.O_TRUNC,
			perm:    0600,
			wantErr: false,
			check: func(t *testing.T, file *os.File, err error) {
				if err != nil {
					t.Errorf("OpenFileSecurely() unexpected error: %v", err)
					return
				}
				if file == nil {
					t.Error("OpenFileSecurely() should not return nil file")
					return
				}
				_ = file.Close()
			},
		},
		{
			name:    "directory traversal with .. should return error",
			path:    "../test.log",
			flag:    os.O_CREATE | os.O_WRONLY | os.O_TRUNC,
			perm:    0600,
			wantErr: true,
			check: func(t *testing.T, _ *os.File, err error) {
				if err == nil {
					t.Error("OpenFileSecurely() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), "directory traversal not allowed") {
					t.Errorf("OpenFileSecurely() error = %v, want to contain 'directory traversal not allowed'", err)
				}
			},
		},
		{
			name:    "absolute path should return error",
			path:    "/etc/passwd",
			flag:    os.O_CREATE | os.O_WRONLY | os.O_TRUNC,
			perm:    0600,
			wantErr: true,
			check: func(t *testing.T, _ *os.File, err error) {
				if err == nil {
					t.Error("OpenFileSecurely() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), "directory traversal not allowed") {
					t.Errorf("OpenFileSecurely() error = %v, want to contain 'directory traversal not allowed'", err)
				}
			},
		},
		{
			name:    "path with .. in middle should work after cleaning",
			path:    filepath.Join("subdir", "..", testFile),
			flag:    os.O_CREATE | os.O_WRONLY | os.O_TRUNC,
			perm:    0600,
			wantErr: false,
			check: func(t *testing.T, file *os.File, err error) {
				if err != nil {
					t.Errorf("OpenFileSecurely() unexpected error: %v", err)
					return
				}
				if file == nil {
					t.Error("OpenFileSecurely() should not return nil file")
					return
				}
				_ = file.Close()
			},
		},
		{
			name:    "append mode should append to existing file",
			path:    testFile,
			flag:    os.O_CREATE | os.O_WRONLY | os.O_APPEND,
			perm:    0600,
			wantErr: false,
			check: func(t *testing.T, file *os.File, err error) {
				if err != nil {
					t.Errorf("OpenFileSecurely() unexpected error: %v", err)
					return
				}
				if file == nil {
					t.Error("OpenFileSecurely() should not return nil file")
					return
				}
				// Write content
				if _, err := file.WriteString(testContent); err != nil {
					t.Errorf("Failed to write to file: %v", err)
				}
				_ = file.Close()

				// Verify content was written
				data, err := os.ReadFile(testFile)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				if string(data) != testContent {
					t.Errorf("File content = %v, want %v", string(data), testContent)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subdirectory if needed for subdirectory test
			if strings.Contains(tt.path, "subdir") && !tt.wantErr {
				if err := os.MkdirAll(filepath.Dir(tt.path), 0755); err != nil {
					t.Fatalf("Failed to create subdirectory: %v", err)
				}
			}
			file, err := OpenFileSecurely(tt.path, tt.flag, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenFileSecurely() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.check(t, file, err)
		})
	}

	// Test with invalid root directory
	_ = os.RemoveAll(tempDir)
	file, err := OpenFileSecurely("test.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err == nil {
		t.Error("OpenFileSecurely() expected error for invalid root directory")
		if file != nil {
			_ = file.Close()
		}
	}
}
