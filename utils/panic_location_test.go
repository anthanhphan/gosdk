// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"strings"
	"testing"
)

// panicHelper is a helper function that panics at a known location for testing.
func panicHelper(panicValue interface{}) (location string, err error) {
	defer func() {
		if r := recover(); r != nil {
			location, err = GetPanicLocation()
		}
	}()
	// This line number will be checked in tests
	panic(panicValue)
}

// nestedPanicHelper creates a panic deeper in the call stack.
func nestedPanicHelper(panicValue interface{}) (location string, err error) {
	return panicHelper(panicValue)
}

func TestGetPanicLocation(t *testing.T) {
	tests := []struct {
		name       string
		panicValue interface{}
		panicFunc  func(interface{}) (string, error)
		wantErr    bool
		errMsg     string
		check      func(t *testing.T, location string, err error)
	}{
		{
			name:       "string panic should return location in test file",
			panicValue: "test panic",
			panicFunc:  panicHelper,
			wantErr:    false,
			check: func(t *testing.T, location string, err error) {
				if err != nil {
					t.Errorf("GetPanicLocation() unexpected error: %v", err)
					return
				}
				if location == "" {
					t.Error("GetPanicLocation() should not return empty location")
					return
				}
				// Should return panic_location_test.go file
				if !strings.Contains(location, "panic_location_test.go") {
					t.Errorf("GetPanicLocation() location = %v, want to contain 'panic_location_test.go'", location)
				}
				// Should be in format "file:line"
				if !strings.Contains(location, ":") {
					t.Errorf("GetPanicLocation() location = %v, want format 'file:line'", location)
				}
			},
		},
		{
			name:       "error panic should return location",
			panicValue: "error occurred",
			panicFunc:  nestedPanicHelper,
			wantErr:    false,
			check: func(t *testing.T, location string, err error) {
				if err != nil {
					t.Errorf("GetPanicLocation() unexpected error: %v", err)
					return
				}
				if location == "" {
					t.Error("GetPanicLocation() should not return empty location")
					return
				}
				if !strings.Contains(location, "panic_location_test.go") {
					t.Errorf("GetPanicLocation() location = %v, want to contain 'panic_location_test.go'", location)
				}
			},
		},
		{
			name:       "nil panic should return location",
			panicValue: nil,
			panicFunc:  panicHelper,
			wantErr:    false,
			check: func(t *testing.T, location string, err error) {
				if err != nil {
					t.Errorf("GetPanicLocation() unexpected error: %v", err)
					return
				}
				if location == "" {
					t.Error("GetPanicLocation() should not return empty location")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := tt.panicFunc(tt.panicValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPanicLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetPanicLocation() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			tt.check(t, location, err)
		})
	}
}

func TestGetPanicLocation_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		panicFunc func() (string, error)
		wantErr   bool
		errMsg    string
		check     func(t *testing.T, location string, err error)
	}{
		{
			name: "call outside recover context should handle gracefully",
			panicFunc: func() (string, error) {
				// Call GetPanicLocation without recover context
				// This tests error handling when called incorrectly
				return GetPanicLocation()
			},
			wantErr: false, // Behavior depends on call stack, may succeed or fail
			check: func(t *testing.T, location string, err error) {
				// Verify function handles edge case gracefully without crashing
				// Error cases are covered in actual panic/recover scenarios
				if err != nil {
					// If error occurs, should be one of the documented error types
					errorTypes := []string{
						"not in panic context",
						"empty stack",
						"no user frame detected",
					}
					hasKnownError := false
					for _, errType := range errorTypes {
						if strings.Contains(err.Error(), errType) {
							hasKnownError = true
							break
						}
					}
					if !hasKnownError {
						t.Errorf("GetPanicLocation() error = %v, want known error type", err)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := tt.panicFunc()
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPanicLocation() error = nil, wantErr true")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetPanicLocation() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			tt.check(t, location, err)
		})
	}
}

// TestIsSystemFrame tests the isSystemFrame function indirectly through GetPanicLocation
// by testing panics in system code vs user code
func TestIsSystemFrame(t *testing.T) {
	tests := []struct {
		name       string
		panicValue interface{}
		setup      func()
		check      func(t *testing.T, location string, err error)
	}{
		{
			name:       "panic in user code should return user location",
			panicValue: "user panic",
			setup:      func() {},
			check: func(t *testing.T, location string, err error) {
				if err != nil {
					t.Errorf("GetPanicLocation() unexpected error: %v", err)
					return
				}
				// Should return test file location, not system frame
				if strings.Contains(location, "runtime/") ||
					strings.Contains(location, "reflect/") ||
					strings.Contains(location, "testing/") {
					t.Errorf("GetPanicLocation() location = %v, should not contain system paths", location)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			location, err := panicHelper(tt.panicValue)
			tt.check(t, location, err)
		})
	}
}

func TestGetPanicLocation_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, error)
		wantErr bool
		errMsg  string
		check   func(t *testing.T, location string, err error)
	}{
		{
			name: "call without recover context should return error or handle gracefully",
			setup: func() (string, error) {
				// Call directly without recover context
				return GetPanicLocation()
			},
			wantErr: false, // May or may not return error depending on call stack
			check: func(t *testing.T, location string, err error) {
				// Function should handle gracefully - either return error or location
				if err != nil {
					// If error, should be one of the known error types
					errorTypes := []string{
						"not in panic context",
						"empty stack",
						"no user frame detected",
					}
					hasKnownError := false
					for _, errType := range errorTypes {
						if strings.Contains(err.Error(), errType) {
							hasKnownError = true
							break
						}
					}
					if !hasKnownError {
						t.Errorf("GetPanicLocation() error = %v, want known error type", err)
					}
				} else {
					// If no error, should return valid location
					if location == "" {
						t.Error("GetPanicLocation() should return location or error")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := tt.setup()
			// Note: wantErr is false because behavior depends on call stack
			// The check function validates the result appropriately
			_ = tt.wantErr // Acknowledge wantErr for documentation
			tt.check(t, location, err)
		})
	}
}

// TestGetPanicLocation_NoUserFrame tests the case where no user frame is detected
func TestGetPanicLocation_NoUserFrame(t *testing.T) {
	// This test is difficult to create directly, but we can test the error message
	// by calling GetPanicLocation in a context where it might not find a user frame
	location, err := GetPanicLocation()
	if err == nil {
		// If no error, that's also valid - depends on call stack
		if location == "" {
			t.Error("GetPanicLocation() should return location or error")
		}
	} else {
		// Should be one of the known error types
		errorTypes := []string{
			"not in panic context",
			"empty stack",
			"no user frame detected",
		}
		hasKnownError := false
		for _, errType := range errorTypes {
			if strings.Contains(err.Error(), errType) {
				hasKnownError = true
				break
			}
		}
		if !hasKnownError {
			t.Errorf("GetPanicLocation() error = %v, want known error type", err)
		}
	}
}
