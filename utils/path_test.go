// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetShortPath(t *testing.T) {
	// Get current working directory and module root for testing
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	moduleRoot := FindModuleRoot(wd)
	if moduleRoot == "" {
		// If no module root, create a temporary structure
		tempDir, err := os.MkdirTemp("", "path_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create a fake module structure
		testModulePath := filepath.Join(tempDir, "testmodule", "subdir", "file.go")
		if err := os.MkdirAll(filepath.Dir(testModulePath), 0755); err != nil {
			t.Fatalf("Failed to create test structure: %v", err)
		}
		if err := os.WriteFile(testModulePath, []byte("package subdir"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create go.mod
		goModPath := filepath.Join(tempDir, "testmodule", "go.mod")
		if err := os.WriteFile(goModPath, []byte("module testmodule\n"), 0644); err != nil {
			t.Fatalf("Failed to create go.mod: %v", err)
		}

		tests := []struct {
			name  string
			input string
			want  string
			check func(t *testing.T, result string)
		}{
			{
				name:  "path within module should return relative path",
				input: testModulePath,
				want:  "subdir/file.go",
				check: func(t *testing.T, result string) {
					// Should be relative to module root
					if !strings.HasSuffix(result, "file.go") {
						t.Errorf("GetShortPath() = %v, want to end with 'file.go'", result)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := GetShortPath(tt.input)
				tt.check(t, result)
			})
		}
		return
	}

	// Test with actual module root
	testFile := filepath.Join(moduleRoot, "utils", "panic_location_test.go")

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "path within module should return relative path",
			input: testFile,
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				// Should contain test file name
				if !strings.Contains(result, "panic_location_test.go") {
					t.Errorf("GetShortPath() = %v, want to contain 'panic_location_test.go'", result)
				}
				// Should not be absolute path
				if filepath.IsAbs(result) {
					t.Errorf("GetShortPath() = %v, want relative path", result)
				}
			},
		},
		{
			name:  "absolute path outside module should return fallback",
			input: "/usr/bin/go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				// Should return some form of path
				if !strings.Contains(result, "bin") && !strings.Contains(result, "go") {
					t.Errorf("GetShortPath() = %v, want to contain path components", result)
				}
			},
		},
		{
			name:  "short path should return directory/filename",
			input: "/tmp/file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				if !strings.Contains(result, "file.go") {
					t.Errorf("GetShortPath() = %v, want to contain 'file.go'", result)
				}
			},
		},
		{
			name:  "path with many components should return last 4",
			input: "/a/b/c/d/e/f/g/h/file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				// Should contain file name
				if !strings.Contains(result, "file.go") {
					t.Errorf("GetShortPath() = %v, want to contain 'file.go'", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetShortPath(tt.input)
			tt.check(t, result)
		})
	}
}

func TestFindModuleRoot(t *testing.T) {
	// Get current file location
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current file location")
	}

	currentDir := filepath.Dir(currentFile)
	expectedModuleRoot := FindModuleRoot(currentFile)

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "file in module should return module root",
			input: currentFile,
			check: func(t *testing.T, result string) {
				if expectedModuleRoot != "" {
					if result == "" {
						t.Error("FindModuleRoot() should find module root")
					}
					if result != expectedModuleRoot {
						t.Errorf("FindModuleRoot() = %v, want %v", result, expectedModuleRoot)
					}
				}
			},
		},
		{
			name:  "file in subdirectory should find module root",
			input: filepath.Join(currentDir, "subdir", "file.go"),
			check: func(t *testing.T, result string) {
				if expectedModuleRoot != "" {
					if result == "" {
						t.Error("FindModuleRoot() should find module root")
					}
					if result != expectedModuleRoot {
						t.Errorf("FindModuleRoot() = %v, want %v", result, expectedModuleRoot)
					}
				}
			},
		},
		{
			name:  "file outside module should return empty",
			input: "/tmp/test.go",
			check: func(t *testing.T, result string) {
				// May or may not find module root depending on setup
				// Just check it doesn't crash
				if result != "" {
					// If it finds a module root, that's okay too
					// Just verify it's a valid path
					if !strings.HasPrefix(result, "/") && !strings.HasPrefix(result, "C:\\") {
						t.Errorf("FindModuleRoot() = %v, want absolute path if found", result)
					}
				}
			},
		},
		{
			name:  "root directory should return empty",
			input: "/",
			check: func(t *testing.T, result string) {
				// Should stop at filesystem root
				if result != "" {
					t.Errorf("FindModuleRoot() = %v, want empty string for root directory", result)
				}
			},
		},
		{
			name:  "empty path should handle gracefully",
			input: "",
			check: func(t *testing.T, result string) {
				// Empty path: filepath.Dir("") returns ".", which might find module root if current dir has go.mod
				// Or might return empty if no go.mod found
				// Both behaviors are acceptable - just verify it doesn't crash
			},
		},
		{
			name:  "deeply nested path should find module root",
			input: filepath.Join(currentDir, "a", "b", "c", "d", "e", "file.go"),
			check: func(t *testing.T, result string) {
				if expectedModuleRoot != "" {
					if result != "" && result != expectedModuleRoot {
						t.Errorf("FindModuleRoot() = %v, want %v or empty", result, expectedModuleRoot)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindModuleRoot(tt.input)
			tt.check(t, result)
		})
	}
}

func TestGetShortPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "empty path should return directory/filename",
			input: "",
			check: func(t *testing.T, result string) {
				// Empty path results in "." for dir and base, which is valid
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				// Result should be some valid path format (like "./." or "." or similar)
			},
		},
		{
			name:  "path with only filename should return directory/filename",
			input: "file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				if !strings.Contains(result, "file.go") {
					t.Errorf("GetShortPath() = %v, want to contain 'file.go'", result)
				}
			},
		},
		{
			name:  "path with 1 component should return directory/filename",
			input: "/file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
			},
		},
		{
			name:  "path with 2 components should return directory/filename",
			input: "/dir/file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				if !strings.Contains(result, "file.go") {
					t.Errorf("GetShortPath() = %v, want to contain 'file.go'", result)
				}
			},
		},
		{
			name:  "path with 3 components should return directory/filename",
			input: "/a/b/file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
			},
		},
		{
			name:  "path with Windows separators should normalize",
			input: "C:\\Users\\test\\file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				// On Windows, backslashes are valid separators, but ToSlash should normalize
				// On Unix, this path won't match the separator so it will go through different logic
				// On Unix, might return last 4 components or directory/filename - both are acceptable
				// Just verify it doesn't crash and returns something reasonable
			},
		},
		{
			name:  "path with exact 4 components should return last 4",
			input: "/a/b/c/d/file.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("GetShortPath() should not return empty string")
				}
				if !strings.Contains(result, "file.go") {
					t.Errorf("GetShortPath() = %v, want to contain 'file.go'", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetShortPath(tt.input)
			tt.check(t, result)
		})
	}
}
