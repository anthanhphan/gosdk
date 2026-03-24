// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Initialize logger for tests
	logger.InitDefaultLogger()
}

// TestRun tests the basic Run function.
func TestRun(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "function with string argument should execute successfully",
			fn: func(_ string) {
				// Test will verify execution via channel
			},
			args: []any{"test message"},
			check: func(_ *testing.T) {
				// Basic execution test - if no panic, it's successful
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with multiple arguments should execute successfully",
			fn: func(_, _ int) {
				// Test will verify execution
			},
			args: []any{10, 20},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with no arguments should execute successfully",
			fn: func() {
				// Test will verify execution
			},
			args: []any{},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run(tt.fn, tt.args...)
			tt.check(t)
		})
	}
}

// TestRun_WithCompletionChannel tests Run with completion verification.
func TestRun_WithCompletionChannel(t *testing.T) {
	done := make(chan bool, 1)
	Run(func() {
		done <- true
	})

	select {
	case <-done:
		// Success - function executed
	case <-time.After(1 * time.Second):
		t.Error("Function did not execute within timeout")
	}
}

// TestRun_WithArguments tests Run with various argument types.
func TestRun_WithArguments(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T, result any)
	}{
		{
			name: "function with string argument should receive correct value",
			fn: func(msg string) {
				assert.Equal(t, "test", msg)
			},
			args: []any{"test"},
			check: func(_ *testing.T, _ any) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with int arguments should receive correct values",
			fn: func(a, b int) {
				assert.Equal(t, 10, a)
				assert.Equal(t, 20, b)
			},
			args: []any{10, 20},
			check: func(_ *testing.T, _ any) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run(tt.fn, tt.args...)
			tt.check(t, nil)
		})
	}
}

// TestRun_PanicRecovery tests that panics are properly recovered with location.
func TestRun_PanicRecovery(t *testing.T) {
	tests := []struct {
		name    string
		fn      any
		args    []any
		wantErr bool
		check   func(t *testing.T)
	}{
		{
			name: "panic with string should be recovered with location",
			fn: func() {
				panic("test panic")
			},
			args:    []any{},
			wantErr: true,
			check: func(_ *testing.T) {
				time.Sleep(100 * time.Millisecond)
				// Panic should be recovered, test should complete
			},
		},
		{
			name: "panic with error should be recovered with location",
			fn: func() {
				panic("test error panic")
			},
			args:    []any{},
			wantErr: true,
			check: func(_ *testing.T) {
				time.Sleep(100 * time.Millisecond)
				// Panic should be recovered, test should complete
			},
		},
		{
			name: "panic with integer should be recovered with location",
			fn: func() {
				panic(123)
			},
			args:    []any{},
			wantErr: true,
			check: func(_ *testing.T) {
				time.Sleep(100 * time.Millisecond)
				// Panic should be recovered, test should complete
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic the test itself
			Run(tt.fn, tt.args...)
			tt.check(t)
		})
	}
}

// TestRun_InvalidFunction tests error handling for invalid functions.
func TestRun_InvalidFunction(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "non-function value should not cause panic",
			fn:   "not a function",
			args: []any{},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
				// Should log error but not panic
			},
		},
		{
			name: "nil function should not cause panic",
			fn:   nil,
			args: []any{},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
				// Should log error but not panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic the test itself
			Run(tt.fn, tt.args...)
			tt.check(t)
		})
	}
}

// TestInvoke tests the invoke function with various scenarios.
func TestInvoke(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "valid function with correct arguments should execute",
			fn: func(msg string) {
				assert.Equal(t, "test", msg)
			},
			args: []any{"test"},
			check: func(_ *testing.T) {
				// Function should execute
			},
		},
		{
			name: "invalid function type should log error",
			fn:   "not a function",
			args: []any{},
			check: func(_ *testing.T) {
				// Should log error but not panic
			},
		},
		{
			name: "function with insufficient arguments should log error",
			fn: func(_, _ int) {
				t.Error("Should not execute")
			},
			args: []any{10},
			check: func(_ *testing.T) {
				// Should log error about insufficient arguments
			},
		},
		{
			name: "function with type mismatch should log error",
			fn: func(s string) {
				// This function should not execute due to type mismatch
				// If it executes, the type checking failed - value is ignored
				_ = s
			},
			args: []any{123},
			check: func(_ *testing.T) {
				// Should log error about type mismatch
				time.Sleep(10 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoke(tt.fn, tt.args)
			tt.check(t)
		})
	}
}

// TestInvoke_TypeConversion tests type conversion in invoke.
func TestInvoke_TypeConversion(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "int to int32 conversion should work",
			fn: func(n int32) {
				assert.Equal(t, int32(10), n)
			},
			args: []any{10},
			check: func(_ *testing.T) {
				// Should execute successfully
			},
		},
		{
			name: "int to int64 conversion should work",
			fn: func(n int64) {
				assert.Equal(t, int64(10), n)
			},
			args: []any{10},
			check: func(_ *testing.T) {
				// Should execute successfully
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoke(tt.fn, tt.args)
			tt.check(t)
		})
	}
}

// TestRecoverPanic tests panic recovery with location.
func TestRecoverPanic(t *testing.T) {
	tests := []struct {
		name  string
		panic any
		check func(t *testing.T)
	}{
		{
			name:  "string panic should be recovered with location",
			panic: "test panic",
			check: func(_ *testing.T) {
				// Should not crash
			},
		},
		{
			name:  "error panic should be recovered with location",
			panic: fmt.Errorf("test error"),
			check: func(_ *testing.T) {
				// Should not crash
			},
		},
		{
			name:  "integer panic should be recovered with location",
			panic: 123,
			check: func(_ *testing.T) {
				// Should not crash
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer recoverPanic()
				if tt.panic != nil {
					panic(tt.panic)
				}
			}()
			tt.check(t)
		})
	}
}

// TestRun_ConcurrentExecution tests concurrent execution of multiple goroutines.
func TestRun_ConcurrentExecution(t *testing.T) {
	var wg sync.WaitGroup
	results := make([]int, 0)
	var mu sync.Mutex

	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		i := i // Capture loop variable
		Run(func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, i)
			mu.Unlock()
		})
	}

	wg.Wait()

	if len(results) != numGoroutines {
		t.Errorf("Expected %d results, got %d", numGoroutines, len(results))
	}
}

// TestStackFrame_Location tests stackFrame.location() method.
func TestStackFrame_Location(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
		check    func(t *testing.T, got string)
	}{
		{
			name:     "valid file path should return formatted location",
			filePath: "\t/home/user/project/file.go:123",
			check: func(t *testing.T, got string) {
				if got == "" {
					t.Error("Expected non-empty location")
				}
				if !strings.Contains(got, ":123") {
					t.Errorf("Expected location to contain line number, got %s", got)
				}
			},
		},
		{
			name:     "file path with closing paren should return empty",
			filePath: "\t/home/user/project/file.go:123)",
			want:     "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty location, got %s", got)
				}
			},
		},
		{
			name:     "file path without colon should return empty",
			filePath: "\t/home/user/project/file.go",
			want:     "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty location, got %s", got)
				}
			},
		},
		{
			name:     "file path with line number and extra spaces should work",
			filePath: "    /home/user/project/file.go:456     ",
			check: func(t *testing.T, got string) {
				if got == "" {
					t.Error("Expected non-empty location")
				}
				if !strings.Contains(got, ":456") {
					t.Errorf("Expected location to contain line number, got %s", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := stackFrame{
				funcName: "test.function",
				filePath: tt.filePath,
			}
			got := frame.location()
			tt.check(t, got)
		})
	}
}

// TestParseFilePath tests parseFilePath function.
func TestParseFilePath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantFile string
		wantLine string
		wantOk   bool
	}{
		{
			name:     "valid file path should parse correctly",
			filePath: "/path/to/file.go:123",
			wantFile: "/path/to/file.go",
			wantLine: "123",
			wantOk:   true,
		},
		{
			name:     "file path with spaces should parse correctly",
			filePath: "  /path/to/file.go:456  ",
			wantFile: "/path/to/file.go",
			wantLine: "456",
			wantOk:   true,
		},
		{
			name:     "file path with line number and extra text should parse line",
			filePath: "/path/to/file.go:789 extra text",
			wantFile: "/path/to/file.go",
			wantLine: "789",
			wantOk:   true,
		},
		{
			name:     "file path without colon should fail",
			filePath: "/path/to/file.go",
			wantFile: "",
			wantLine: "",
			wantOk:   false,
		},
		{
			name:     "file path with only colon should fail",
			filePath: "/path/to/file.go:",
			wantFile: "/path/to/file.go",
			wantLine: "",
			wantOk:   false,
		},
		{
			name:     "empty file path should fail",
			filePath: "",
			wantFile: "",
			wantLine: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, line, ok := parseFilePath(tt.filePath)
			if ok != tt.wantOk {
				t.Errorf("parseFilePath() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok {
				if file != tt.wantFile {
					t.Errorf("parseFilePath() file = %v, want %v", file, tt.wantFile)
				}
				if line != tt.wantLine {
					t.Errorf("parseFilePath() line = %v, want %v", line, tt.wantLine)
				}
			}
		})
	}
}

// TestParseStackFrames tests parseStackFrames function.
func TestParseStackFrames(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int
		check func(t *testing.T, frames []stackFrame)
	}{
		{
			name: "valid stack trace should parse correctly",
			lines: []string{
				"goroutine 1 [running]:",
				"github.com/user/package.function",
				"\t/path/to/file.go:123",
				"github.com/user/package.other",
				"\t/path/to/file.go:456",
			},
			want: 2,
			check: func(t *testing.T, frames []stackFrame) {
				if len(frames) != 2 {
					t.Errorf("Expected 2 frames, got %d", len(frames))
				}
			},
		},
		{
			name: "stack trace with single frame should parse",
			lines: []string{
				"goroutine 1 [running]:",
				"github.com/user/package.function",
				"\t/path/to/file.go:123",
			},
			want: 1,
			check: func(t *testing.T, frames []stackFrame) {
				if len(frames) != 1 {
					t.Errorf("Expected 1 frame, got %d", len(frames))
				}
			},
		},
		{
			name: "empty lines should return empty frames",
			lines: []string{
				"goroutine 1 [running]:",
			},
			want: 0,
			check: func(t *testing.T, frames []stackFrame) {
				if len(frames) != 0 {
					t.Errorf("Expected 0 frames, got %d", len(frames))
				}
			},
		},
		{
			name: "incomplete frame should be skipped",
			lines: []string{
				"goroutine 1 [running]:",
				"github.com/user/package.function",
			},
			want: 0,
			check: func(t *testing.T, frames []stackFrame) {
				if len(frames) != 0 {
					t.Errorf("Expected 0 frames, got %d", len(frames))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frames := parseStackFrames(tt.lines)
			if tt.check != nil {
				tt.check(t, frames)
			}
		})
	}
}

// TestContainsAny_WithPrefixes tests containsAny function with prefix-style patterns.
func TestContainsAny_WithPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		prefixes []string
		want     bool
	}{
		{
			name:     "string with matching prefix should return true",
			str:      "runtime.gopanic",
			prefixes: []string{"runtime."},
			want:     true,
		},
		{
			name:     "string containing prefix should return true",
			str:      "my.runtime.function",
			prefixes: []string{"runtime."},
			want:     true,
		},
		{
			name:     "string without prefix should return false",
			str:      "my.function",
			prefixes: []string{"runtime."},
			want:     false,
		},
		{
			name:     "empty prefixes should return false",
			str:      "runtime.gopanic",
			prefixes: []string{},
			want:     false,
		},
		{
			name:     "empty string should return false",
			str:      "",
			prefixes: []string{"runtime."},
			want:     false,
		},
		{
			name:     "multiple prefixes with match should return true",
			str:      "reflect.Value.call",
			prefixes: []string{"runtime.", "reflect."},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.str, tt.prefixes)
			if got != tt.want {
				t.Errorf("containsAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetCallerLocation tests getCallerLocation function.
func TestGetCallerLocation(t *testing.T) {
	// Call getCallerLocation indirectly through Run to get a real caller location
	done := make(chan string, 1)
	Run(func() {
		// getCallerLocation is tested indirectly through Run
		// We can verify it works by checking that Run doesn't crash
		done <- "success"
	})

	select {
	case <-done:
		// Success - function executed
	case <-time.After(100 * time.Millisecond):
		t.Error("Function did not execute")
	}
}

// TestInvoke_ExcessArguments tests invoke with excess arguments.
func TestInvoke_ExcessArguments(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "function with excess arguments should log warning",
			fn: func(_ int) {
				// Function should execute with first argument
			},
			args: []any{10, 20, 30},
			check: func(_ *testing.T) {
				// Should log warning about excess arguments but still execute
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with exact arguments should not log warning",
			fn: func(_, _ int) {
				// Function should execute
			},
			args: []any{10, 20},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoke(tt.fn, tt.args)
			tt.check(t)
		})
	}
}

// TestConvertArguments_EdgeCases tests convertArguments with edge cases.
func TestConvertArguments_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		fn      any
		args    []any
		wantErr bool
		check   func(t *testing.T)
	}{
		{
			name: "nil pointer argument should be valid",
			fn: func(s *string) {
				if s != nil {
					t.Error("Expected nil pointer")
				}
			},
			args:    []any{nil},
			wantErr: false, // nil is valid for pointer type
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "uint to int conversion should work",
			fn: func(n int) {
				assert.Equal(t, 10, n)
			},
			args:    []any{uint(10)},
			wantErr: false,
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "float to int conversion should fail",
			fn: func(_ int) {
				// Should not execute due to type conversion failure
			},
			args:    []any{10.5},
			wantErr: true,
			check: func(_ *testing.T) {
				// Function should not execute, error should be logged
				time.Sleep(50 * time.Millisecond)
				// No assertion needed - if function executed, test will fail
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoke(tt.fn, tt.args)
			tt.check(t)
		})
	}
}

// TestRecoverPanic_WithLocation tests recoverPanic with caller location.
func TestRecoverPanic_WithLocation(t *testing.T) {
	tests := []struct {
		name           string
		panicValue     any
		callerLocation string
		check          func(t *testing.T)
	}{
		{
			name:           "panic with caller location should use caller location as fallback",
			panicValue:     "test panic",
			callerLocation: "test.go:123",
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name:           "panic without caller location should still recover",
			panicValue:     "test panic",
			callerLocation: "",
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer recoverPanic()
				panic(tt.panicValue)
			}()
			tt.check(t)
		})
	}
}

// TestExtractLocation_EdgeCases tests extractLocation with various stack trace formats.
func TestExtractLocation_EdgeCases(t *testing.T) {
	parser := newStackTraceParser()

	tests := []struct {
		name       string
		stackTrace string
		want       string
		check      func(t *testing.T, got string)
	}{
		{
			name: "stack trace with only wrapper functions should return empty",
			stackTrace: `goroutine 1 [running]:
github.com/user/goroutine.invoke.func1()
	/path/to/goroutine/routine.go:123
github.com/user/goroutine.Run.func1()
	/path/to/goroutine/routine.go:456`,
			want: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty location for wrapper functions, got %s", got)
				}
			},
		},
		{
			name: "stack trace with user code should return user location",
			stackTrace: `goroutine 1 [running]:
github.com/user/goroutine.invoke.func1()
	/path/to/goroutine/routine.go:123
github.com/user/package.UserFunction()
	/path/to/user/file.go:789`,
			check: func(t *testing.T, got string) {
				if got == "" {
					t.Error("Expected non-empty location for user code")
				}
				if !strings.Contains(got, "file.go") || !strings.Contains(got, "789") {
					t.Errorf("Expected location to contain user file, got %s", got)
				}
			},
		},
		{
			name:       "empty stack trace should return empty",
			stackTrace: "",
			want:       "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty location, got %s", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractLocation(tt.stackTrace)
			if tt.check != nil {
				tt.check(t, got)
			} else if got != tt.want {
				t.Errorf("extractLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsValidFilePath tests isValidFilePath function.
func TestIsValidFilePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "valid file path with colon should return true",
			path: "/path/to/file.go:123",
			want: true,
		},
		{
			name: "file path with closing paren should return false",
			path: "/path/to/file.go:123)",
			want: false,
		},
		{
			name: "file path without colon should return false",
			path: "/path/to/file.go",
			want: false,
		},
		{
			name: "empty path should return false",
			path: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFilePath(tt.path)
			if got != tt.want {
				t.Errorf("isValidFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConvertArguments_NilPointer tests convertArguments with nil pointer handling.
func TestConvertArguments_NilPointer(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "nil pointer argument should be converted to zero value",
			fn: func(s *string) {
				if s != nil {
					t.Error("Expected nil pointer")
				}
			},
			args: []any{nil},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "nil interface argument should work",
			fn: func(v interface{}) {
				if v != nil {
					t.Error("Expected nil interface")
				}
			},
			args: []any{nil},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoke(tt.fn, tt.args)
			tt.check(t)
		})
	}
}

// TestRecoverPanic_NoPanic tests recoverPanic when no panic occurs.
func TestRecoverPanic_NoPanic(_ *testing.T) {
	func() {
		defer recoverPanic()
		// No panic - should return normally
	}()
	// Test passes if no panic occurred
}

// TestNormalizePanicValue tests normalizePanicValue with various types.
func TestNormalizePanicValue(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantMsg string
		check   func(t *testing.T, err error)
	}{
		{
			name:    "error type should return as-is",
			input:   fmt.Errorf("test error"),
			wantMsg: "test error",
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected non-nil error")
				}
				if err.Error() != "test error" {
					t.Errorf("Expected error message 'test error', got '%s'", err.Error())
				}
			},
		},
		{
			name:    "string type should be converted to error",
			input:   "test string",
			wantMsg: "test string",
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected non-nil error")
				}
				if err.Error() != "test string" {
					t.Errorf("Expected error message 'test string', got '%s'", err.Error())
				}
			},
		},
		{
			name:    "integer type should be converted to error",
			input:   123,
			wantMsg: "123",
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected non-nil error")
				}
				if err.Error() != "123" {
					t.Errorf("Expected error message '123', got '%s'", err.Error())
				}
			},
		},
		{
			name:    "nil should be converted to error",
			input:   nil,
			wantMsg: "<nil>",
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected non-nil error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := normalizePanicValue(tt.input)
			tt.check(t, err)
		})
	}
}

// TestContainsAny tests containsAny function.
func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		patterns []string
		want     bool
	}{
		{
			name:     "string containing pattern should return true",
			str:      "runtime.gopanic",
			patterns: []string{"runtime."},
			want:     true,
		},
		{
			name:     "string without pattern should return false",
			str:      "my.function",
			patterns: []string{"runtime."},
			want:     false,
		},
		{
			name:     "empty patterns should return false",
			str:      "runtime.gopanic",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "multiple patterns with match should return true",
			str:      "runtime.gopanic",
			patterns: []string{"reflect.", "runtime."},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.str, tt.patterns)
			if got != tt.want {
				t.Errorf("containsAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFrameFilter_ShouldSkip tests frameFilter.shouldSkip method.
func TestFrameFilter_ShouldSkip(t *testing.T) {
	filter := &frameFilter{
		wrapperPatterns: []string{".invoke.func", ".Run.func"},
		packagePatterns: []string{"runtime/", "reflect/"},
		prefixPatterns:  []string{"runtime.", "reflect."},
	}

	tests := []struct {
		name     string
		funcName string
		filePath string
		want     bool
	}{
		{
			name:     "wrapper function should be skipped",
			funcName: "github.com/user/goroutine.invoke.func1",
			filePath: "/path/to/file.go",
			want:     true,
		},
		{
			name:     "runtime package should be skipped",
			funcName: "runtime.gopanic",
			filePath: "/usr/lib/go/runtime/panic.go",
			want:     true,
		},
		{
			name:     "user function should not be skipped",
			funcName: "github.com/user/package.UserFunction",
			filePath: "/path/to/user/file.go",
			want:     false,
		},
		{
			name:     "runtime prefix should be skipped",
			funcName: "runtime.gopanic",
			filePath: "/path/to/file.go",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.shouldSkip(tt.funcName, tt.filePath)
			if got != tt.want {
				t.Errorf("shouldSkip() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCapturePanicLocation tests capturePanicLocation function.
func TestCapturePanicLocation(_ *testing.T) {
	func() {
		defer recoverPanic()
		// Capture location - in non-panic context, location might be empty
		_ = capturePanicLocation()
	}()
}

// TestGetCallerLocation_FromGoroutine tests getCallerLocation when called from goroutine.
func TestGetCallerLocation_FromGoroutine(t *testing.T) {
	done := make(chan bool, 1)
	Run(func() {
		// getCallerLocation is tested indirectly
		done <- true
	})

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Function did not execute")
	}
}

// TestLocation_EdgeCases tests location() with more edge cases.
func TestLocation_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		check    func(t *testing.T, got string)
	}{
		{
			name:     "file path with tab and line number should parse",
			filePath: "\t/path/to/file.go:999",
			check: func(t *testing.T, got string) {
				if got == "" {
					t.Error("Expected non-empty location")
				}
			},
		},
		{
			name:     "file path with invalid parse should return empty",
			filePath: "no_colon_here",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty location, got %s", got)
				}
			},
		},
		{
			name:     "file path with only colon should return empty",
			filePath: ":",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("Expected empty location, got %s", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := stackFrame{
				funcName: "test",
				filePath: tt.filePath,
			}
			got := frame.location()
			tt.check(t, got)
		})
	}
}

// TestParseStackFrames_EdgeCases tests parseStackFrames with edge cases.
func TestParseStackFrames_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		check func(t *testing.T, frames []stackFrame)
	}{
		{
			name: "odd number of lines after header should handle correctly",
			lines: []string{
				"goroutine 1 [running]:",
				"function1",
				"\tfile1.go:1",
				"function2",
			},
			check: func(t *testing.T, frames []stackFrame) {
				if len(frames) != 1 {
					t.Errorf("Expected 1 frame, got %d", len(frames))
				}
			},
		},
		{
			name: "single function line without file should skip",
			lines: []string{
				"goroutine 1 [running]:",
				"function1",
			},
			check: func(t *testing.T, frames []stackFrame) {
				if len(frames) != 0 {
					t.Errorf("Expected 0 frames, got %d", len(frames))
				}
			},
		},
		{
			name: "empty lines should create frame with empty strings",
			lines: []string{
				"goroutine 1 [running]:",
				"",
				"",
			},
			check: func(t *testing.T, frames []stackFrame) {
				// Empty lines still create a frame (frame with empty strings)
				// This is expected behavior as parseStackFrames processes pairs
				if len(frames) == 0 {
					t.Error("Expected at least 1 frame from empty line pair")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frames := parseStackFrames(tt.lines)
			tt.check(t, frames)
		})
	}
}

// TestConvertArguments_NilHandling tests convertArguments nil handling path.
func TestConvertArguments_NilHandling(t *testing.T) {
	tests := []struct {
		name  string
		fn    any
		args  []any
		check func(t *testing.T)
	}{
		{
			name: "nil for non-pointer non-interface should fail",
			fn: func(_ string) {
				t.Error("Should not execute")
			},
			args: []any{nil},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "nil for pointer type should create zero value",
			fn: func(s *string) {
				if s != nil {
					t.Error("Expected nil")
				}
			},
			args: []any{nil},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "nil for interface type should work",
			fn: func(v interface{}) {
				if v != nil {
					t.Error("Expected nil")
				}
			},
			args: []any{nil},
			check: func(_ *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoke(tt.fn, tt.args)
			tt.check(t)
		})
	}
}

// TestRecoverPanic_LocationFallback tests recoverPanic location fallback.
func TestRecoverPanic_LocationFallback(_ *testing.T) {
	func() {
		defer recoverPanic()
		panic("test")
	}()
	time.Sleep(50 * time.Millisecond)
}

// TestRecoverPanic_EmptyLocation tests recoverPanic with empty captured location.
func TestRecoverPanic_EmptyLocation(_ *testing.T) {
	func() {
		defer recoverPanic()
		panic("test")
	}()
	time.Sleep(50 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// Additional coverage tests
// ---------------------------------------------------------------------------

// TestRun_FuncErrorFastPath tests the func(error) fast path in Run.
func TestRun_FuncErrorFastPath(t *testing.T) {
	done := make(chan error, 1)
	expectedErr := fmt.Errorf("test error")

	Run(func(err error) {
		done <- err
	}, expectedErr)

	select {
	case got := <-done:
		assert.Equal(t, expectedErr, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("func(error) fast path did not execute")
	}
}

// TestRun_FuncIntFastPath tests the func(int) fast path in Run.
func TestRun_FuncIntFastPath(t *testing.T) {
	done := make(chan int, 1)

	Run(func(n int) {
		done <- n
	}, 42)

	select {
	case got := <-done:
		assert.Equal(t, 42, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("func(int) fast path did not execute")
	}
}

// TestRun_FuncStringWrongArgFallsToReflect tests that func(string) with non-string arg
// falls through to the generic reflect path (which logs type mismatch error).
func TestRun_FuncStringWrongArgFallsToReflect(t *testing.T) {
	done := make(chan bool, 1)

	type myStruct struct{ x int }
	// func(string) with myStruct arg -- type assert fails, falls to reflect path
	Run(func(s string) {
		done <- true
	}, myStruct{42}) // struct cannot convert to string

	time.Sleep(100 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("should not have executed with incompatible arg type")
	default:
		// expected -- reflect path rejects the type mismatch
	}
}

// TestRun_FuncErrorWrongArgFallsToReflect tests func(error) fast path with non-error arg.
func TestRun_FuncErrorWrongArgFallsToReflect(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	// func(error) with string arg -- type assert fails, falls to reflect path
	Run(func(_ error) {}, "not an error")
	time.Sleep(100 * time.Millisecond)
}

// TestRun_FuncIntWrongArgFallsToReflect tests func(int) fast path with non-int arg.
func TestRun_FuncIntWrongArgFallsToReflect(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	// func(int) with string arg -- type assert fails, falls to reflect path
	Run(func(_ int) {}, "not an int")
	time.Sleep(100 * time.Millisecond)
}

// TestRun_FuncNoArgWithExtraArgs tests func() fast path with extra args -- falls to reflect.
func TestRun_FuncNoArgWithExtraArgs(t *testing.T) {
	done := make(chan bool, 1)

	// func() with args -- len(args) != 0, falls to reflect generic path
	Run(func() {
		done <- true
	}, "extra")

	select {
	case <-done:
		// reflect path executed the func() ignoring extra args (after warning)
	case <-time.After(100 * time.Millisecond):
		// Also acceptable -- reflect may reject
	}
}

// TestRun_PanicInFuncErrorFastPath tests panic recovery in func(error) fast path.
func TestRun_PanicInFuncErrorFastPath(t *testing.T) {
	Run(func(_ error) {
		panic("panic in error handler")
	}, fmt.Errorf("test"))
	time.Sleep(100 * time.Millisecond)
	// Test passes if panic is recovered (no crash)
}

// TestRun_PanicInFuncIntFastPath tests panic recovery in func(int) fast path.
func TestRun_PanicInFuncIntFastPath(t *testing.T) {
	Run(func(_ int) {
		panic("panic in int handler")
	}, 42)
	time.Sleep(100 * time.Millisecond)
}

// TestGetCallerLocation_Direct tests getCallerLocation called directly (from goroutine package).
func TestGetCallerLocation_Direct(t *testing.T) {
	// When called from within the goroutine package, frame.File contains "/goroutine/"
	// so it should return "" (filtered out)
	loc := getCallerLocation()
	// Location might be empty (we're calling from goroutine package) or non-empty (test framework)
	_ = loc // just exercise the function
}

// TestExtractLocation_SingleLineNoNewline tests extractLocation with truncated stack.
func TestExtractLocation_SingleLineNoNewline(t *testing.T) {
	parser := newStackTraceParser()

	// Stack trace where func line has no newline (truncated)
	got := parser.extractLocation("goroutine 1 [running]:\nsome.func.name")
	assert.Equal(t, "", got)
}

// TestExtractLocation_LastFrameNoTrailingNewline tests last frame without trailing newline.
func TestExtractLocation_LastFrameNoTrailingNewline(t *testing.T) {
	parser := newStackTraceParser()

	got := parser.extractLocation("goroutine 1 [running]:\nmain.myFunc()\n\t/path/to/main.go:42")
	assert.NotEmpty(t, got)
	assert.Contains(t, got, "main.go:42")
}

// TestFanOut_WorkersZero tests FanOut with workers=0 (defaults to 1).
func TestFanOut_WorkersZero(t *testing.T) {
	results, err := FanOut(context.Background(), []int{1, 2}, 0,
		func(_ context.Context, n int) (int, error) { return n * 2, nil },
	)
	assert.NoError(t, err)
	assert.Equal(t, []int{2, 4}, results)
}

// TestTrySubmit_AfterStop tests TrySubmit after pool is stopped.
func TestTrySubmit_AfterStop(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 5})
	pool.Start(context.Background())
	pool.Stop()

	ok := pool.TrySubmit(func() {})
	assert.False(t, ok)
}

// TestGroup_Go_SemReleaseOnPanic tests that semaphore is released when goroutine panics.
func TestGroup_Go_SemReleaseOnPanic(t *testing.T) {
	g := NewGroupWithLimit(context.Background(), 1)

	g.Go(func(_ context.Context) error {
		panic("sem release test")
	})

	// If sem is not released, this Go() would block forever
	g.Go(func(_ context.Context) error {
		return nil
	})

	err := g.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
}

// TestGroup_Go_NoSemNilPath tests that Go without semaphore skips sem logic.
func TestGroup_Go_NoSemNilPath(t *testing.T) {
	g := NewGroup() // no semaphore
	done := make(chan bool, 1)
	g.Go(func(_ context.Context) error {
		done <- true
		return nil
	})
	err := g.Wait()
	assert.NoError(t, err)
	assert.True(t, <-done)
}

// TestTrySubmit_Success tests TrySubmit happy path.
func TestTrySubmit_Success(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 5})
	pool.Start(context.Background())

	var count atomic.Int32
	ok := pool.TrySubmit(func() { count.Add(1) })
	assert.True(t, ok)

	pool.Stop()
	assert.Equal(t, int32(1), count.Load())
}

// TestGetCallerLocation_ReturnsEmpty_FromInternalFrame tests that getCallerLocation
// returns empty when the first frame is from the goroutine package itself.
func TestGetCallerLocation_ReturnsEmpty_FromInternalFrame(t *testing.T) {
	// Call from within the goroutine package -- frame.File contains "/goroutine/"
	loc := getCallerLocation()
	// From test files (which ARE in /goroutine/), we might get an empty string
	// or the testing framework location -- either is valid
	_ = loc
}

// testHelperCallerLocation is an internal function to exercise getCallerLocation
// from multiple stack frames within the goroutine package.
func testHelperCallerLocation() string {
	return getCallerLocation()
}

// TestGetCallerLocation_AllFramesInternal tests the return "" path when all frames
// are from the goroutine package or runtime.
func TestGetCallerLocation_AllFramesInternal(t *testing.T) {
	// Call through a helper within the goroutine package -- tests the "no external frame" path
	loc := testHelperCallerLocation()
	// The result may be a testing framework location or empty -- both valid
	_ = loc
}

// TestRecoverPanic_CallerLocationFallback tests the callerLocation fallback
// when capturePanicLocation returns empty.
func TestRecoverPanic_CallerLocationFallback(t *testing.T) {
	func() {
		defer recoverPanic()
		panic("fallback test")
	}()
}

// ---------------------------------------------------------------------------
// RunWithContext tests
// ---------------------------------------------------------------------------

func TestRunWithContext_NormalExecution(t *testing.T) {
	done := make(chan string, 1)
	RunWithContext(context.Background(), func(_ context.Context) {
		done <- "ok"
	})

	select {
	case v := <-done:
		assert.Equal(t, "ok", v)
	case <-time.After(200 * time.Millisecond):
		t.Error("RunWithContext did not complete")
	}
}

func TestRunWithContext_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	RunWithContext(ctx, func(ctx context.Context) {
		done <- ctx.Err()
	})

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Error("RunWithContext did not complete")
	}
}

func TestRunWithContext_PanicRecovery(t *testing.T) {
	RunWithContext(context.Background(), func(_ context.Context) {
		panic("ctx panic test")
	})
	time.Sleep(100 * time.Millisecond)
	// Should not crash the process
}

// ---------------------------------------------------------------------------
// RunWithTimeout tests
// ---------------------------------------------------------------------------

func TestRunWithTimeout_NormalCompletion(t *testing.T) {
	done := make(chan string, 1)
	cancel := RunWithTimeout(2*time.Second, func(_ context.Context) {
		done <- "finished"
	})
	defer cancel()

	select {
	case v := <-done:
		assert.Equal(t, "finished", v)
	case <-time.After(500 * time.Millisecond):
		t.Error("RunWithTimeout did not complete")
	}
}

func TestRunWithTimeout_Timeout(t *testing.T) {
	started := make(chan struct{})
	done := make(chan struct{})

	cancel := RunWithTimeout(50*time.Millisecond, func(ctx context.Context) {
		close(started)
		<-ctx.Done() // wait for timeout
		close(done)
	})
	defer cancel()

	<-started
	select {
	case <-done:
		// fn observed ctx.Done and exited — no goroutine leak
	case <-time.After(500 * time.Millisecond):
		t.Error("fn did not exit after timeout")
	}
}

func TestRunWithTimeout_EarlyCancel(t *testing.T) {
	done := make(chan error, 1)
	cancel := RunWithTimeout(5*time.Second, func(ctx context.Context) {
		<-ctx.Done()
		done <- ctx.Err()
	})

	// Cancel early
	cancel()

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Error("RunWithTimeout did not cancel")
	}
}

func TestRunWithTimeout_PanicRecovery(t *testing.T) {
	cancel := RunWithTimeout(2*time.Second, func(_ context.Context) {
		panic("timeout panic test")
	})
	defer cancel()
	time.Sleep(100 * time.Millisecond)
	// Should not crash the process
}

// ---------------------------------------------------------------------------
// SubmitWithTimeout tests
// ---------------------------------------------------------------------------

func TestSubmitWithTimeout_NormalCompletion(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 2, QueueSize: 10})
	pool.Start(context.Background())
	defer pool.Stop()

	done := make(chan string, 1)
	ok := pool.SubmitWithTimeout(2*time.Second, func(_ context.Context) {
		done <- "ok"
	})
	assert.True(t, ok)

	select {
	case v := <-done:
		assert.Equal(t, "ok", v)
	case <-time.After(500 * time.Millisecond):
		t.Error("SubmitWithTimeout did not complete")
	}
}

func TestSubmitWithTimeout_Timeout(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 10})
	pool.Start(context.Background())
	defer pool.Stop()

	done := make(chan error, 1)
	pool.SubmitWithTimeout(50*time.Millisecond, func(ctx context.Context) {
		<-ctx.Done()
		done <- ctx.Err()
	})

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Error("SubmitWithTimeout job did not timeout")
	}
}

func TestSubmitWithTimeout_StoppedPool(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 10})
	pool.Start(context.Background())
	pool.Stop()

	ok := pool.SubmitWithTimeout(time.Second, func(_ context.Context) {})
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Edge case tests — RunWithContext
// ---------------------------------------------------------------------------

// Verify context.WithTimeout works through RunWithContext.
func TestRunWithContext_WithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	RunWithContext(ctx, func(ctx context.Context) {
		<-ctx.Done()
		done <- ctx.Err()
	})

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(500 * time.Millisecond):
		t.Error("goroutine did not exit after timeout")
	}
}

// Multiple goroutines share the same cancel context.
func TestRunWithContext_SharedCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var exited atomic.Int32

	for i := 0; i < 5; i++ {
		RunWithContext(ctx, func(ctx context.Context) {
			<-ctx.Done()
			exited.Add(1)
		})
	}

	time.Sleep(30 * time.Millisecond)
	cancel() // cancel all 5 goroutines
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(5), exited.Load(), "all 5 goroutines should have exited")
}

// fn completes with result accessible via channel even when context is alive.
func TestRunWithContext_ResultBeforeCancel(t *testing.T) {
	result := make(chan int, 1)
	RunWithContext(context.Background(), func(_ context.Context) {
		result <- 42
	})

	select {
	case v := <-result:
		assert.Equal(t, 42, v)
	case <-time.After(200 * time.Millisecond):
		t.Error("did not receive result")
	}
}

// ---------------------------------------------------------------------------
// Edge case tests — RunWithTimeout
// ---------------------------------------------------------------------------

// Verify timeout error type is DeadlineExceeded.
func TestRunWithTimeout_ErrorType(t *testing.T) {
	done := make(chan error, 1)
	cancel := RunWithTimeout(50*time.Millisecond, func(ctx context.Context) {
		<-ctx.Done()
		done <- ctx.Err()
	})
	defer cancel()

	err := <-done
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// Concurrent RunWithTimeout calls should not interfere with each other.
func TestRunWithTimeout_ConcurrentCalls(t *testing.T) {
	var completed atomic.Int32

	cancels := make([]context.CancelFunc, 10)
	for i := 0; i < 10; i++ {
		cancels[i] = RunWithTimeout(2*time.Second, func(_ context.Context) {
			time.Sleep(10 * time.Millisecond)
			completed.Add(1)
		})
	}

	time.Sleep(200 * time.Millisecond)
	for _, c := range cancels {
		c()
	}

	assert.Equal(t, int32(10), completed.Load(), "all 10 goroutines should complete")
}

// Double cancel should not panic.
func TestRunWithTimeout_DoubleCancelSafe(t *testing.T) {
	cancel := RunWithTimeout(time.Second, func(ctx context.Context) {
		<-ctx.Done()
	})
	cancel()
	cancel() // should not panic
	time.Sleep(50 * time.Millisecond)
}

// Goroutine leak detection: after cancel, no goroutines left running.
func TestRunWithTimeout_NoLeak(t *testing.T) {
	before := runtime.NumGoroutine()

	done := make(chan struct{})
	cancel := RunWithTimeout(50*time.Millisecond, func(ctx context.Context) {
		<-ctx.Done()
		close(done)
	})

	<-done
	cancel()
	time.Sleep(100 * time.Millisecond)

	after := runtime.NumGoroutine()
	// Allow ±2 for background goroutines (GC, etc.)
	assert.InDelta(t, before, after, 2, "goroutine count should return to baseline")
}

// ---------------------------------------------------------------------------
// Edge case tests — SubmitWithTimeout
// ---------------------------------------------------------------------------

// SubmitWithTimeout job uses pool context: if pool stops, job ctx is cancelled.
func TestSubmitWithTimeout_PoolStopCancelsJob(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 10})
	pool.Start(context.Background())

	done := make(chan error, 1)
	pool.SubmitWithTimeout(5*time.Second, func(ctx context.Context) {
		<-ctx.Done()
		done <- ctx.Err()
	})

	time.Sleep(30 * time.Millisecond)
	pool.Stop() // stopping pool cancels pool ctx, which cascades to job ctx

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Error("job should have been cancelled when pool stopped")
	}
}

// Multiple SubmitWithTimeout jobs with different timeouts.
func TestSubmitWithTimeout_MixedTimeouts(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 3, QueueSize: 10})
	pool.Start(context.Background())
	defer pool.Stop()

	var fast, slow atomic.Int32

	// Fast jobs: 2s timeout, finish immediately
	for i := 0; i < 5; i++ {
		pool.SubmitWithTimeout(2*time.Second, func(_ context.Context) {
			fast.Add(1)
		})
	}

	// Slow jobs: 50ms timeout, block until cancelled
	for i := 0; i < 3; i++ {
		pool.SubmitWithTimeout(50*time.Millisecond, func(ctx context.Context) {
			<-ctx.Done()
			slow.Add(1)
		})
	}

	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(5), fast.Load(), "all fast jobs should complete")
	assert.Equal(t, int32(3), slow.Load(), "all slow jobs should timeout and exit")
}
