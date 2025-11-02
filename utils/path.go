// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetShortPath converts absolute file path to short format relative to module root.
//
// Input:
//   - fullPath: The absolute file path to convert
//
// Output:
//   - string: The relative path from module root, or fallback path if module root not found
//
// The function automatically finds the Go module root (where go.mod is) and returns relative path.
// It uses multiple strategies:
//  1. Try to find Go module root (directory containing go.mod) and return relative path
//  2. Try relative to current working directory
//  3. Return last 4 path components as fallback
//  4. Last resort: return directory/filename
//
// Example:
//
//	path := utils.GetShortPath("/path/to/module/example/handler/controller/handler.go")
//	// Returns: "example/handler/controller/handler.go" (relative to module root)
func GetShortPath(fullPath string) string {
	// Strategy 1: Find Go module root (directory containing go.mod) and return relative path
	if moduleRoot := FindModuleRoot(fullPath); moduleRoot != "" {
		if rel, err := filepath.Rel(moduleRoot, fullPath); err == nil {
			// Normalize path separators to forward slashes for consistency
			return filepath.ToSlash(rel)
		}
	}

	// Strategy 2: Try relative to current working directory
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, fullPath); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}

	// Strategy 3: Return last 4 path components as fallback
	parts := strings.Split(fullPath, string(filepath.Separator))
	// Filter out empty strings (from leading/trailing separators)
	nonEmptyParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}
	if len(nonEmptyParts) >= 4 {
		return strings.Join(nonEmptyParts[len(nonEmptyParts)-4:], "/")
	}

	// Last resort: return directory/filename
	dir := filepath.Dir(fullPath)
	fileName := filepath.Base(fullPath)
	dirName := filepath.Base(dir)
	// Normalize path separators to forward slashes for consistency
	return filepath.ToSlash(filepath.Join(dirName, fileName))
}

// FindModuleRoot walks up the directory tree from the given file path to find the directory containing go.mod file.
//
// Input:
//   - filePath: The absolute file path to start searching from
//
// Output:
//   - string: The module root directory path, or empty string if not found
//
// Example:
//
//	root := utils.FindModuleRoot("/path/to/module/subdir/file.go")
//	if root == "" {
//	    log.Fatal("Module root not found")
//	}
//	fmt.Println(root) // Prints: "/path/to/module"
func FindModuleRoot(filePath string) string {
	dir := filepath.Dir(filePath)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}
	return ""
}
