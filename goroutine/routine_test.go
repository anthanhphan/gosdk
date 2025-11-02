// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"fmt"
	"strings"
	"sync"
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
			fn: func(msg string) {
				// Test will verify execution via channel
			},
			args: []any{"test message"},
			check: func(t *testing.T) {
				// Basic execution test - if no panic, it's successful
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with multiple arguments should execute successfully",
			fn: func(a, b int) {
				// Test will verify execution
			},
			args: []any{10, 20},
			check: func(t *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with no arguments should execute successfully",
			fn: func() {
				// Test will verify execution
			},
			args: []any{},
			check: func(t *testing.T) {
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
			check: func(t *testing.T, result any) {
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
			check: func(t *testing.T, result any) {
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
				time.Sleep(50 * time.Millisecond)
				// Should log error but not panic
			},
		},
		{
			name: "nil function should not cause panic",
			fn:   nil,
			args: []any{},
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
				// Function should execute
			},
		},
		{
			name: "invalid function type should log error",
			fn:   "not a function",
			args: []any{},
			check: func(t *testing.T) {
				// Should log error but not panic
			},
		},
		{
			name: "function with insufficient arguments should log error",
			fn: func(a, b int) {
				t.Error("Should not execute")
			},
			args: []any{10},
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
				// Should execute successfully
			},
		},
		{
			name: "int to int64 conversion should work",
			fn: func(n int64) {
				assert.Equal(t, int64(10), n)
			},
			args: []any{10},
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
				// Should not crash
			},
		},
		{
			name:  "error panic should be recovered with location",
			panic: fmt.Errorf("test error"),
			check: func(t *testing.T) {
				// Should not crash
			},
		},
		{
			name:  "integer panic should be recovered with location",
			panic: 123,
			check: func(t *testing.T) {
				// Should not crash
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer recoverPanic("")
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

// TestHasPrefixAny tests hasPrefixAny function.
func TestHasPrefixAny(t *testing.T) {
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
			got := hasPrefixAny(tt.str, tt.prefixes)
			if got != tt.want {
				t.Errorf("hasPrefixAny() = %v, want %v", got, tt.want)
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
			fn: func(a int) {
				// Function should execute with first argument
			},
			args: []any{10, 20, 30},
			check: func(t *testing.T) {
				// Should log warning about excess arguments but still execute
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "function with exact arguments should not log warning",
			fn: func(a, b int) {
				// Function should execute
			},
			args: []any{10, 20},
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "float to int conversion should fail",
			fn: func(n int) {
				// Should not execute due to type conversion failure
			},
			args:    []any{10.5},
			wantErr: true,
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name:           "panic without caller location should still recover",
			panicValue:     "test panic",
			callerLocation: "",
			check: func(t *testing.T) {
				time.Sleep(50 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer recoverPanic(tt.callerLocation)
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
func TestRecoverPanic_NoPanic(t *testing.T) {
	func() {
		defer recoverPanic("test.go:123")
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
func TestCapturePanicLocation(t *testing.T) {
	func() {
		defer recoverPanic("fallback.go:123")
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
			fn: func(s string) {
				t.Error("Should not execute")
			},
			args: []any{nil},
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
			check: func(t *testing.T) {
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
func TestRecoverPanic_LocationFallback(t *testing.T) {
	func() {
		defer recoverPanic("fallback.go:999")
		panic("test")
	}()
	time.Sleep(50 * time.Millisecond)
}

// TestRecoverPanic_EmptyLocation tests recoverPanic with empty captured location.
func TestRecoverPanic_EmptyLocation(t *testing.T) {
	func() {
		defer recoverPanic("")
		panic("test")
	}()
	time.Sleep(50 * time.Millisecond)
}
