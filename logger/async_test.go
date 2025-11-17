// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestNewAsyncLogger(t *testing.T) {
	tests := []struct {
		name      string
		logger    *Logger
		queueSize int
		check     func(t *testing.T, al *AsyncLogger)
	}{
		{
			name: "valid logger with custom queue size should create async logger",
			logger: NewLogger(&Config{
				LogLevel:    LevelInfo,
				LogEncoding: EncodingJSON,
			}, nil),
			queueSize: 50,
			check: func(t *testing.T, al *AsyncLogger) {
				if al == nil {
					t.Fatal("NewAsyncLogger() should not return nil")
				}
				if al.logger == nil {
					t.Error("NewAsyncLogger() logger should not be nil")
				}
				if al.rt == nil || al.rt.queue == nil {
					t.Error("NewAsyncLogger() queue should not be nil")
				}
				if cap(al.rt.queue) != 50 {
					t.Errorf("NewAsyncLogger() queue capacity = %v, want %v", cap(al.rt.queue), 50)
				}
			},
		},
		{
			name: "zero queue size should use default",
			logger: NewLogger(&Config{
				LogLevel:    LevelInfo,
				LogEncoding: EncodingJSON,
			}, nil),
			queueSize: 0,
			check: func(t *testing.T, al *AsyncLogger) {
				if al == nil {
					t.Fatal("NewAsyncLogger() should not return nil")
				}
				if al.rt == nil || al.rt.queue == nil {
					t.Fatal("NewAsyncLogger() queue should not be nil")
				}
				if cap(al.rt.queue) != 100 {
					t.Errorf("NewAsyncLogger() queue capacity = %v, want %v (default)", cap(al.rt.queue), 100)
				}
			},
		},
		{
			name: "negative queue size should use default",
			logger: NewLogger(&Config{
				LogLevel:    LevelInfo,
				LogEncoding: EncodingJSON,
			}, nil),
			queueSize: -10,
			check: func(t *testing.T, al *AsyncLogger) {
				if al == nil {
					t.Fatal("NewAsyncLogger() should not return nil")
				}
				if al.rt == nil || al.rt.queue == nil {
					t.Fatal("NewAsyncLogger() queue should not be nil")
				}
				if cap(al.rt.queue) != 100 {
					t.Errorf("NewAsyncLogger() queue capacity = %v, want %v (default)", cap(al.rt.queue), 100)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := NewAsyncLogger(tt.logger, tt.queueSize)
			// Give worker time to start
			time.Sleep(10 * time.Millisecond)
			tt.check(t, al)
			// Cleanup
			al.Close()
		})
	}
}

func TestAsyncLogger_Debug(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Debug("test debug message")
	al.Debug("key", "value")

	// Give worker time to process
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Debugf(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Debugf("test %s message", "debug")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Debugw(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Debugw("test message", "key1", "value1", "key2", 123)
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Info(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Info("test info message")
	al.Info("key", "value")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Infof(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Infof("test %s message", "info")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Infow(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Infow("test message", "key1", "value1", "key2", 123)
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Warn(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelWarn,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Warn("test warn message")
	al.Warn("key", "value")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Warnf(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelWarn,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Warnf("test %s message", "warn")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Warnw(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelWarn,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Warnw("test message", "key1", "value1", "key2", 123)
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Error(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:          LevelError,
		LogEncoding:       EncodingJSON,
		DisableStacktrace: false,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Error("test error message")
	al.Error("key", "value")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Errorf(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:          LevelError,
		LogEncoding:       EncodingJSON,
		DisableStacktrace: false,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Errorf("test %s message", "error")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_Errorw(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:          LevelError,
		LogEncoding:       EncodingJSON,
		DisableStacktrace: false,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Errorw("test message", "key1", "value1", "key2", 123)
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_With(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil, String("base", "value"))
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	newAl := al.With(String("new", "field"))
	if newAl == nil {
		t.Fatal("With() should not return nil")
	}
	if newAl.logger == nil {
		t.Error("With() logger should not be nil")
	}

	// Test that new logger has both fields
	newAl.Info("test")
	time.Sleep(50 * time.Millisecond)
	newAl.Flush()
}

func TestAsyncLogger_Flush(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)

	// Add some entries
	al.Info("message 1")
	al.Info("message 2")
	al.Info("message 3")

	// Flush should wait for all entries to be processed
	al.Flush()

	// Flush again should not block (already flushed)
	al.Flush()
}

func TestAsyncLogger_Close(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)

	al.Info("message 1")
	al.Info("message 2")

	err := al.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestAsyncLogger_LevelFiltering(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelWarn,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	// Debug and Info should be filtered
	al.Debug("should not log")
	al.Info("should not log")
	al.Warn("should log")
	al.Error("should log")

	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_QueueFull(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)
	// Use very small queue to test queue full scenario
	al := NewAsyncLogger(logger, 2)
	defer al.Close()

	// Fill queue quickly
	al.Info("message 1")
	al.Info("message 2")
	al.Info("message 3") // This should trigger synchronous write

	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_WithCaller(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:      LevelInfo,
		LogEncoding:   EncodingJSON,
		DisableCaller: false,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Info("test with caller")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

func TestAsyncLogger_WithStacktrace(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:          LevelError,
		LogEncoding:       EncodingJSON,
		DisableStacktrace: false,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Error("test with stacktrace")
	time.Sleep(50 * time.Millisecond)
	al.Flush()
}

// TestFatalErrorCases tests the fatal error paths using subprocesses
func TestAsyncLogger_FatalErrorCases(t *testing.T) {
	testBinary := os.Args[0]

	tests := []struct {
		name             string
		testFunc         string
		expectedExitCode int
	}{
		{
			name:             "Fatal should exit with code 1",
			testFunc:         "TestAsyncFatal",
			expectedExitCode: 1,
		},
		{
			name:             "Fatalf should exit with code 1",
			testFunc:         "TestAsyncFatalf",
			expectedExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(testBinary, "-test.run", tt.testFunc)
			cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

			err := cmd.Run()

			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() != tt.expectedExitCode {
					t.Errorf("Expected exit code %d, got %d", tt.expectedExitCode, exitError.ExitCode())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestAsyncFatal tests the fatal error path
func TestAsyncFatal(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	logger := NewLogger(&Config{
		LogLevel:    LevelError,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Fatal("fatal error")
}

// TestAsyncFatalf tests the fatal error path
func TestAsyncFatalf(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	logger := NewLogger(&Config{
		LogLevel:    LevelError,
		LogEncoding: EncodingJSON,
	}, nil)
	al := NewAsyncLogger(logger, 10)
	defer al.Close()

	al.Fatalf("fatal error: %s", "test")
}
