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
