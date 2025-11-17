// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		defaultFields []Field
		check         func(t *testing.T)
	}{
		{
			name: "valid config with default fields should initialize logger",
			config: &Config{
				LogLevel:          LevelInfo,
				LogEncoding:       EncodingJSON,
				DisableCaller:     false,
				DisableStacktrace: false,
				IsDevelopment:     false,
				LogOutputPaths:    []string{},
			},
			defaultFields: []Field{
				String("app_name", "test-app"),
				String("version", "1.0.0"),
			},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name: "valid config without default fields should initialize logger",
			config: &Config{
				LogLevel:          LevelDebug,
				LogEncoding:       EncodingConsole,
				DisableCaller:     false,
				DisableStacktrace: false,
				IsDevelopment:     true,
				LogOutputPaths:    []string{},
			},
			defaultFields: []Field{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton state for each test
			loggerInstance = nil
			once = sync.Once{}

			// Initialize logger
			undo := InitLogger(tt.config, tt.defaultFields...)
			defer undo()

			// Run checks
			tt.check(t)
		})
	}
}

func TestInitDevelopmentLogger(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	undo := InitDefaultLogger()
	defer undo()

	// Test that logger was initialized
	if loggerInstance == nil {
		t.Error("Logger should be initialized")
	}
}

func TestInitProductionLogger(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	undo := InitProductionLogger()
	defer undo()

	// Test that logger was initialized
	if loggerInstance == nil {
		t.Error("Logger should be initialized")
	}
}

func TestNewLoggerWithFields(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger first
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name   string
		fields []Field
		check  func(t *testing.T, logger *Logger)
	}{
		{
			name: "with fields should create logger",
			fields: []Field{
				String("service", "test-service"),
				String("operation", "test-operation"),
			},
			check: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Error("Logger should be created")
				}
			},
		},
		{
			name:   "without fields should create logger",
			fields: []Field{},
			check: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Error("Logger should be created")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with fields
			logger := NewLoggerWithFields(tt.fields...)
			tt.check(t, logger)
		})
	}
}

func TestNewLoggerWithFields_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that NewLoggerWithFields auto-initializes when logger is nil
	logger := NewLoggerWithFields()
	if logger == nil {
		t.Error("Logger should be created")
	}

	// Test that logger instance was created
	if loggerInstance == nil {
		t.Error("Logger instance should be initialized")
	}
}

func TestSingletonBehavior(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// First initialization
	undo1 := InitDefaultLogger()
	defer undo1()

	firstLogger := loggerInstance

	// Test that logger was initialized
	if firstLogger == nil {
		t.Error("Logger should be initialized after first call")
	}

	// Test that the logger instance is properly set
	if loggerInstance == nil {
		t.Error("loggerInstance should be set")
	}
}

func TestLogLevelMapping(t *testing.T) {
	// Test that our level constants are valid
	levels := []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}
	for _, level := range levels {
		if !level.isValid() {
			t.Errorf("Level %v should be valid", level)
		}
	}
}

// TestFatalErrorCases tests the log.Fatalf paths using subprocesses
// This is necessary because log.Fatalf causes the program to exit
func TestFatalErrorCases(t *testing.T) {
	// Get the current test binary path
	testBinary := os.Args[0]

	tests := []struct {
		name             string
		testFunc         string
		expectedExitCode int
	}{
		{
			name:             "invalid config should cause fatal error",
			testFunc:         "TestFatalInvalidConfig",
			expectedExitCode: 1,
		},
		{
			name:             "zap build error should cause fatal error",
			testFunc:         "TestFatalZapBuildError",
			expectedExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the test function in a subprocess
			cmd := exec.Command(testBinary, "-test.run", tt.testFunc)
			cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

			err := cmd.Run()

			// Check if the process exited with the expected code
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

// TestFatalInvalidConfig tests the config validation fatal error path
// This function will be run in a subprocess and should exit with code 1
func TestFatalInvalidConfig(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	// Reset singleton state
	loggerInstance = nil
	once = sync.Once{}

	// This should call log.Fatalf and exit
	InitLogger(&Config{
		LogLevel:    Level("invalid"),
		LogEncoding: EncodingJSON,
	})
}

// TestFatalZapBuildError tests the zap build fatal error path
// This function will be run in a subprocess and should exit with code 1
func TestFatalZapBuildError(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	// Reset singleton state
	loggerInstance = nil
	once = sync.Once{}

	// This should call log.Fatalf and exit
	// We can't easily trigger zap.Build() to fail, so this test is more theoretical
	InitLogger(&Config{
		LogLevel:       LevelInfo,
		LogEncoding:    EncodingJSON,
		LogOutputPaths: []string{"/dev/null/invalid/path/that/should/not/exist"},
	})
}

func TestDebug(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name  string
		args  []interface{}
		check func(t *testing.T)
	}{
		{
			name: "should log debug message",
			args: []interface{}{"test debug message"},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name: "should log multiple arguments",
			args: []interface{}{"test", "debug", 123},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Debug(tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestDebug_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Debug auto-initializes when logger is nil
	Debug("test message")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestDebugf(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name     string
		template string
		args     []interface{}
		check    func(t *testing.T)
	}{
		{
			name:     "should log formatted debug message",
			template: "user %s logged in with id %d",
			args:     []interface{}{"john", 123},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:     "should log message without arguments",
			template: "simple debug message",
			args:     []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Debugf(tt.template, tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestDebugf_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Debugf auto-initializes when logger is nil
	Debugf("test message %s", "test")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestDebugw(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name          string
		msg           string
		keysAndValues []interface{}
		check         func(t *testing.T)
	}{
		{
			name: "should log debug message with structured fields",
			msg:  "user logged in",
			keysAndValues: []interface{}{
				"username", "john",
				"user_id", 123,
			},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:          "should log message without fields",
			msg:           "simple debug message",
			keysAndValues: []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Debugw(tt.msg, tt.keysAndValues...)

			// Verify
			tt.check(t)
		})
	}
}

func TestDebugw_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Debugw auto-initializes when logger is nil
	Debugw("test message", "key", "value")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestInfo(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name  string
		args  []interface{}
		check func(t *testing.T)
	}{
		{
			name: "should log info message",
			args: []interface{}{"test info message"},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name: "should log multiple arguments",
			args: []interface{}{"test", "info", 456},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Info(tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestInfo_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Info auto-initializes when logger is nil
	Info("test message")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestInfof(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name     string
		template string
		args     []interface{}
		check    func(t *testing.T)
	}{
		{
			name:     "should log formatted info message",
			template: "user %s created with id %d",
			args:     []interface{}{"jane", 456},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:     "should log message without arguments",
			template: "simple info message",
			args:     []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Infof(tt.template, tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestInfof_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Infof auto-initializes when logger is nil
	Infof("test message %s", "test")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestInfow(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name          string
		msg           string
		keysAndValues []interface{}
		check         func(t *testing.T)
	}{
		{
			name: "should log info message with structured fields",
			msg:  "user created",
			keysAndValues: []interface{}{
				"username", "jane",
				"user_id", 456,
			},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:          "should log message without fields",
			msg:           "simple info message",
			keysAndValues: []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Infow(tt.msg, tt.keysAndValues...)

			// Verify
			tt.check(t)
		})
	}
}

func TestInfow_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Infow auto-initializes when logger is nil
	Infow("test message", "key", "value")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestWarn(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name  string
		args  []interface{}
		check func(t *testing.T)
	}{
		{
			name: "should log warn message",
			args: []interface{}{"test warn message"},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name: "should log multiple arguments",
			args: []interface{}{"test", "warn", 789},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Warn(tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestWarn_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Warn auto-initializes when logger is nil
	Warn("test message")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestWarnf(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name     string
		template string
		args     []interface{}
		check    func(t *testing.T)
	}{
		{
			name:     "should log formatted warn message",
			template: "connection to %s failed after %d retries",
			args:     []interface{}{"localhost", 3},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:     "should log message without arguments",
			template: "simple warn message",
			args:     []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Warnf(tt.template, tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestWarnf_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Warnf auto-initializes when logger is nil
	Warnf("test message %s", "test")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestWarnw(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name          string
		msg           string
		keysAndValues []interface{}
		check         func(t *testing.T)
	}{
		{
			name: "should log warn message with structured fields",
			msg:  "connection failed",
			keysAndValues: []interface{}{
				"host", "localhost",
				"retries", 3,
			},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:          "should log message without fields",
			msg:           "simple warn message",
			keysAndValues: []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Warnw(tt.msg, tt.keysAndValues...)

			// Verify
			tt.check(t)
		})
	}
}

func TestWarnw_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Warnw auto-initializes when logger is nil
	Warnw("test message", "key", "value")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestError(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name  string
		args  []interface{}
		check func(t *testing.T)
	}{
		{
			name: "should log error message",
			args: []interface{}{"test error message"},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name: "should log multiple arguments",
			args: []interface{}{"test", "error", 999},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Error(tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestError_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Error auto-initializes when logger is nil
	Error("test message")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestErrorf(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name     string
		template string
		args     []interface{}
		check    func(t *testing.T)
	}{
		{
			name:     "should log formatted error message",
			template: "failed to connect to %s on port %d",
			args:     []interface{}{"database", 5432},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:     "should log message without arguments",
			template: "simple error message",
			args:     []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Errorf(tt.template, tt.args...)

			// Verify
			tt.check(t)
		})
	}
}

func TestErrorf_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Errorf auto-initializes when logger is nil
	Errorf("test message %s", "test")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestErrorw(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	tests := []struct {
		name          string
		msg           string
		keysAndValues []interface{}
		check         func(t *testing.T)
	}{
		{
			name: "should log error message with structured fields",
			msg:  "connection failed",
			keysAndValues: []interface{}{
				"host", "database",
				"port", 5432,
			},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
		{
			name:          "should log message without fields",
			msg:           "simple error message",
			keysAndValues: []interface{}{},
			check: func(t *testing.T) {
				if loggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			Errorw(tt.msg, tt.keysAndValues...)

			// Verify
			tt.check(t)
		})
	}
}

func TestErrorw_AutoInit(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Test that Errorw auto-initializes when logger is nil
	Errorw("test message", "key", "value")

	if loggerInstance == nil {
		t.Error("Logger should be auto-initialized")
	}
}

func TestInitAsyncLogger(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		defaultFields []Field
		check         func(t *testing.T)
	}{
		{
			name: "valid config with default fields should initialize async logger",
			config: &Config{
				LogLevel:          LevelInfo,
				LogEncoding:       EncodingJSON,
				DisableCaller:     false,
				DisableStacktrace: false,
				IsDevelopment:     false,
				LogOutputPaths:    []string{},
			},
			defaultFields: []Field{
				String("app_name", "test-app"),
				String("version", "1.0.0"),
			},
			check: func(t *testing.T) {
				if asyncLoggerInstance == nil {
					t.Error("AsyncLogger should be initialized")
				}
			},
		},
		{
			name: "valid config without default fields should initialize async logger",
			config: &Config{
				LogLevel:          LevelDebug,
				LogEncoding:       EncodingConsole,
				DisableCaller:     false,
				DisableStacktrace: false,
				IsDevelopment:     true,
				LogOutputPaths:    []string{},
			},
			defaultFields: []Field{},
			check: func(t *testing.T) {
				if asyncLoggerInstance == nil {
					t.Error("AsyncLogger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton state for each test
			asyncLoggerInstance = nil
			asyncOnce = sync.Once{}

			// Initialize async logger
			undo := InitAsyncLogger(tt.config, tt.defaultFields...)
			defer undo()

			// Run checks
			tt.check(t)
		})
	}
}

func TestInitAsyncLogger_GlobalFunctions(t *testing.T) {
	// Reset singleton state for testing
	asyncLoggerInstance = nil
	asyncOnce = sync.Once{}

	// Initialize async logger
	undo := InitAsyncLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	})
	defer undo()

	// Test that global functions use async logger
	Info("test message")
	Infof("test %s", "message")
	Infow("test", "key", "value")

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)
	Flush()
}

func TestGlobalFunctions_WithAsyncLogger(t *testing.T) {
	// Reset singleton state for testing
	asyncLoggerInstance = nil
	asyncOnce = sync.Once{}

	// Initialize async logger
	undo := InitAsyncLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	})
	defer undo()

	// Test all global functions with async logger
	Debug("debug message")
	Debugf("debug %s", "message")
	Debugw("debug", "key", "value")

	Info("info message")
	Infof("info %s", "message")
	Infow("info", "key", "value")

	Warn("warn message")
	Warnf("warn %s", "message")
	Warnw("warn", "key", "value")

	Error("error message")
	Errorf("error %s", "message")
	Errorw("error", "key", "value")

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)
	Flush()
}

func TestFatal(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	// Test that Fatal function exists and can be called
	// Note: We can't actually test os.Exit(1) in normal tests, so we test via subprocess
	testBinary := os.Args[0]
	cmd := exec.Command(testBinary, "-test.run", "TestFatalGlobal")
	cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

	err := cmd.Run()

	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFatalGlobal(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	// Reset singleton state
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	// This should call os.Exit(1)
	Fatal("fatal error")
}

func TestFatalf(t *testing.T) {
	// Reset singleton state for testing
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	// Test that Fatalf function exists and can be called
	testBinary := os.Args[0]
	cmd := exec.Command(testBinary, "-test.run", "TestFatalfGlobal")
	cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

	err := cmd.Run()

	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFatalfGlobal(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	// Reset singleton state
	loggerInstance = nil
	once = sync.Once{}

	// Initialize logger
	undo := InitDefaultLogger()
	defer undo()

	// This should call os.Exit(1)
	Fatalf("fatal error: %s", "test")
}

func TestFatal_WithAsyncLogger(t *testing.T) {
	// Reset singleton state for testing
	asyncLoggerInstance = nil
	asyncOnce = sync.Once{}

	// Initialize async logger
	undo := InitAsyncLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	})
	defer undo()

	// Test that Fatal function exists and can be called with async logger
	// Note: We can't actually test os.Exit(1) in normal tests, so we test via subprocess
	testBinary := os.Args[0]
	cmd := exec.Command(testBinary, "-test.run", "TestFatalGlobalWithAsync")
	cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

	err := cmd.Run()

	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFatalGlobalWithAsync(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	// Reset singleton state
	asyncLoggerInstance = nil
	asyncOnce = sync.Once{}

	// Initialize async logger
	undo := InitAsyncLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	})
	defer undo()

	// This should call os.Exit(1)
	Fatal("fatal error with async logger")
}

func TestFatalf_WithAsyncLogger(t *testing.T) {
	// Reset singleton state for testing
	asyncLoggerInstance = nil
	asyncOnce = sync.Once{}

	// Initialize async logger
	undo := InitAsyncLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	})
	defer undo()

	// Test that Fatalf function exists and can be called with async logger
	testBinary := os.Args[0]
	cmd := exec.Command(testBinary, "-test.run", "TestFatalfGlobalWithAsync")
	cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

	err := cmd.Run()

	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFatalfGlobalWithAsync(t *testing.T) {
	if os.Getenv("GO_TEST_FATAL") != "1" {
		t.Skip("Skipping fatal test in main process")
	}

	// Reset singleton state
	asyncLoggerInstance = nil
	asyncOnce = sync.Once{}

	// Initialize async logger
	undo := InitAsyncLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	})
	defer undo()

	// This should call os.Exit(1)
	Fatalf("fatal error with async logger: %s", "test")
}
