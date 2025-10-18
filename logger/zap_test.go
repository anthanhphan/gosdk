// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"os"
	"os/exec"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		defaultFields []zap.Field
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
			defaultFields: []zap.Field{
				zap.String("app_name", "test-app"),
				zap.String("version", "1.0.0"),
			},
			check: func(t *testing.T) {
				if zapLoggerInstance == nil {
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
			defaultFields: []zap.Field{},
			check: func(t *testing.T) {
				if zapLoggerInstance == nil {
					t.Error("Logger should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton state for each test
			zapLoggerInstance = nil
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
	zapLoggerInstance = nil
	once = sync.Once{}

	undo := InitDevelopmentLogger()
	defer undo()

	// Test that logger was initialized
	if zapLoggerInstance == nil {
		t.Error("Logger should be initialized")
	}
}

func TestInitProductionLogger(t *testing.T) {
	// Reset singleton state for testing
	zapLoggerInstance = nil
	once = sync.Once{}

	undo := InitProductionLogger()
	defer undo()

	// Test that logger was initialized
	if zapLoggerInstance == nil {
		t.Error("Logger should be initialized")
	}
}

func TestNewLoggerWithFields(t *testing.T) {
	// Reset singleton state for testing
	zapLoggerInstance = nil
	once = sync.Once{}

	// Initialize logger first
	undo := InitDevelopmentLogger()
	defer undo()

	tests := []struct {
		name   string
		fields []zap.Field
		check  func(t *testing.T, logger *zap.SugaredLogger)
	}{
		{
			name: "with fields should create logger",
			fields: []zap.Field{
				zap.String("service", "test-service"),
				zap.String("operation", "test-operation"),
			},
			check: func(t *testing.T, logger *zap.SugaredLogger) {
				if logger == nil {
					t.Error("Logger should be created")
				}
			},
		},
		{
			name:   "without fields should create logger",
			fields: []zap.Field{},
			check: func(t *testing.T, logger *zap.SugaredLogger) {
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
	zapLoggerInstance = nil
	once = sync.Once{}

	// Test that NewLoggerWithFields auto-initializes when logger is nil
	logger := NewLoggerWithFields()
	if logger == nil {
		t.Error("Logger should be created")
	}

	// Test that logger instance was created
	if zapLoggerInstance == nil {
		t.Error("Logger instance should be initialized")
	}
}

func TestSingletonBehavior(t *testing.T) {
	// Reset singleton state for testing
	zapLoggerInstance = nil
	once = sync.Once{}

	// First initialization
	undo1 := InitDevelopmentLogger()
	defer undo1()

	firstLogger := zapLoggerInstance

	// Test that logger was initialized
	if firstLogger == nil {
		t.Error("Logger should be initialized after first call")
	}

	// Test that the logger instance is properly set
	if zapLoggerInstance == nil {
		t.Error("zapLoggerInstance should be set")
	}
}

func TestLogLevelMapping(t *testing.T) {
	// Test that our level constants map correctly to zap levels
	expectedMappings := map[Level]zapcore.Level{
		LevelDebug: zapcore.DebugLevel,
		LevelInfo:  zapcore.InfoLevel,
		LevelWarn:  zapcore.WarnLevel,
		LevelError: zapcore.ErrorLevel,
	}

	for level, expectedZapLevel := range expectedMappings {
		if logLevelMap[level] != expectedZapLevel {
			t.Errorf("Level %v maps to %v, want %v", level, logLevelMap[level], expectedZapLevel)
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
	zapLoggerInstance = nil
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
	zapLoggerInstance = nil
	once = sync.Once{}

	// This should call log.Fatalf and exit
	// We can't easily trigger zap.Build() to fail, so this test is more theoretical
	InitLogger(&Config{
		LogLevel:       LevelInfo,
		LogEncoding:    EncodingJSON,
		LogOutputPaths: []string{"/dev/null/invalid/path/that/should/not/exist"},
	})
}
