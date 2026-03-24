// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// GetPanicLocation extracts the panic location from call stack where panic occurred.
//
// Output:
//   - string: The panic location in format "file:line" (relative to module root)
//   - error: Any error that occurred during location detection (empty stack, no panic context, etc.)
//
// The function uses runtime.Callers() to automatically extract the call stack and works by:
//  1. Collecting all frames from the call stack
//  2. Finding the recover function (defer func with ".func" in function name)
//  3. Finding the first user code frame after the recover function (this is where panic occurred)
//
// This function must be called from within a recover() handler to work correctly.
//
// Example:
//
//	func recoverPanic() {
//	    defer func() {
//	        if r := recover(); r != nil {
//	            location, err := utils.GetPanicLocation()
//	            if err != nil {
//	                log.Printf("Failed to get panic location: %v", err)
//	                return
//	            }
//	            log.Printf("Panic %v at %s", r, location)
//	        }
//	    }()
//	    panic("something went wrong")
//	}
func GetPanicLocation() (location string, err error) {
	const maxDepth = 64
	pcs := make([]uintptr, maxDepth)

	n := runtime.Callers(2, pcs)
	if n == 0 {
		return "", fmt.Errorf("empty stack (no panic context)")
	}

	frames := runtime.CallersFrames(pcs[:n])

	// Single-pass: find recover frame, then first user frame after it.
	recoverFound := false
	var recoverFile string
	var recoverLine int

	for {
		frame, more := frames.Next()

		if !recoverFound {
			// Looking for the recover frame (anonymous defer function)
			if !isSystemFrame(frame.File) && strings.Contains(frame.Function, ".func") {
				recoverFound = true
				recoverFile = frame.File
				recoverLine = frame.Line
			}
		} else {
			// Looking for first user frame after recover frame
			if !isSystemFrame(frame.File) {
				if frame.File != recoverFile || frame.Line != recoverLine {
					shortPath := GetShortPath(frame.File)
					return fmt.Sprintf("%s:%d", shortPath, frame.Line), nil
				}
			}
		}

		if !more {
			break
		}
	}

	if !recoverFound {
		return "", fmt.Errorf("not in panic context (recover frame not found)")
	}

	return "", fmt.Errorf("panic context found but no user frame detected")
}

// isSystemFrame filters files belonging to runtime, reflect, internal, framework.
// It filters out known system packages and framework files.
func isSystemFrame(file string) bool {
	ignorePrefixes := []string{
		"runtime/",
		"reflect/",
		"testing/",
		"net/http",
	}

	for _, p := range ignorePrefixes {
		if strings.Contains(file, p) {
			return true
		}
	}

	return false
}
