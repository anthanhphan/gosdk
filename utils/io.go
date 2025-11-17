// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileSecurely reads a file securely by preventing directory traversal attacks.
//
// Input:
//   - path: The file path to read (must be within the current working directory)
//
// Output:
//   - []byte: The file contents
//   - error: Any error that occurred during reading
//
// Example:
//
//	data, err := ReadFileSecurely("config/app.json")
//	if err != nil {
//	    log.Fatal("Failed to read file:", err)
//	}
//	fmt.Println(string(data))
func ReadFileSecurely(path string) ([]byte, error) {
	// Get the current working directory as the root
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Create a root filesystem to restrict access
	rootFS := os.DirFS(root)

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Ensure the path doesn't start with ".." or "/" to prevent directory traversal
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path: %s (directory traversal not allowed)", path)
	}

	// Open the file using the restricted filesystem
	file, err := rootFS.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// OpenFileSecurely opens a file for writing securely by preventing directory traversal attacks.
// Uses os.Root (Go 1.24+) to restrict file access to the current working directory.
//
// Input:
//   - path: The file path to open (must be within the current working directory)
//   - flag: File opening flags (e.g., os.O_CREATE|os.O_WRONLY|os.O_APPEND)
//   - perm: File permissions (e.g., 0600)
//
// Output:
//   - *os.File: The opened file
//   - error: Any error that occurred during opening
//
// Example:
//
//	file, err := OpenFileSecurely("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
//	if err != nil {
//	    log.Fatal("Failed to open file:", err)
//	}
//	defer file.Close()
func OpenFileSecurely(path string, flag int, perm os.FileMode) (*os.File, error) {
	// Get the current working directory as the root
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Ensure the path doesn't start with ".." or "/" to prevent directory traversal
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path: %s (directory traversal not allowed)", path)
	}

	// Use os.Root to restrict file access to the working directory
	// This prevents directory traversal attacks and avoids G304 detection
	rootFS, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("failed to open root directory: %w", err)
	}

	// Open the file using Root.OpenFile which restricts access to the root directory
	return rootFS.OpenFile(cleanPath, flag, perm)
}
