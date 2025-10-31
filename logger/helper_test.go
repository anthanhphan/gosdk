// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestBuildZapConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, zapConfig zap.Config)
	}{
		{
			name: "development config should set correct values",
			config: &Config{
				LogLevel:          LevelDebug,
				LogEncoding:       EncodingConsole,
				DisableCaller:     false,
				DisableStacktrace: false,
				IsDevelopment:     true,
				LogOutputPaths:    []string{},
			},
			check: func(t *testing.T, zapConfig zap.Config) {
				if zapConfig.Level.Level() != zapcore.DebugLevel {
					t.Errorf("Level = %v, want %v", zapConfig.Level.Level(), zapcore.DebugLevel)
				}
				if zapConfig.Encoding != "console" {
					t.Errorf("Encoding = %v, want console", zapConfig.Encoding)
				}
				if zapConfig.DisableCaller {
					t.Error("DisableCaller should be false")
				}
				if zapConfig.DisableStacktrace {
					t.Error("DisableStacktrace should be false")
				}
			},
		},
		{
			name: "production config should set correct values",
			config: &Config{
				LogLevel:          LevelInfo,
				LogEncoding:       EncodingJSON,
				DisableCaller:     false,
				DisableStacktrace: true,
				IsDevelopment:     false,
				LogOutputPaths:    []string{"app.log"},
			},
			check: func(t *testing.T, zapConfig zap.Config) {
				if zapConfig.Level.Level() != zapcore.InfoLevel {
					t.Errorf("Level = %v, want %v", zapConfig.Level.Level(), zapcore.InfoLevel)
				}
				if zapConfig.Encoding != "json" {
					t.Errorf("Encoding = %v, want json", zapConfig.Encoding)
				}
				if zapConfig.DisableCaller {
					t.Error("DisableCaller should be false")
				}
				if !zapConfig.DisableStacktrace {
					t.Error("DisableStacktrace should be true")
				}
				if len(zapConfig.OutputPaths) != 1 || zapConfig.OutputPaths[0] != "app.log" {
					t.Errorf("OutputPaths = %v, want [app.log]", zapConfig.OutputPaths)
				}
			},
		},
		{
			name: "config with empty output paths should use default",
			config: &Config{
				LogLevel:          LevelWarn,
				LogEncoding:       EncodingJSON,
				DisableCaller:     true,
				DisableStacktrace: false,
				IsDevelopment:     false,
				LogOutputPaths:    []string{},
			},
			check: func(t *testing.T, zapConfig zap.Config) {
				if zapConfig.Level.Level() != zapcore.WarnLevel {
					t.Errorf("Level = %v, want %v", zapConfig.Level.Level(), zapcore.WarnLevel)
				}
				if !zapConfig.DisableCaller {
					t.Error("DisableCaller should be true")
				}
				// OutputPaths will be set by zap's default config
				if len(zapConfig.OutputPaths) == 0 {
					t.Error("OutputPaths should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapConfig := buildZapConfig(tt.config)
			tt.check(t, zapConfig)
		})
	}
}

func TestBuildEncoder(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, encoderConfig zapcore.EncoderConfig)
	}{
		{
			name: "console encoding config should set correct keys",
			config: &Config{
				LogEncoding:       EncodingConsole,
				DisableCaller:     false,
				DisableStacktrace: false,
			},
			check: func(t *testing.T, encoderConfig zapcore.EncoderConfig) {
				if encoderConfig.MessageKey != LogEncoderMessageKey {
					t.Errorf("MessageKey = %v, want %v", encoderConfig.MessageKey, LogEncoderMessageKey)
				}
				if encoderConfig.TimeKey != LogEncoderTimeKey {
					t.Errorf("TimeKey = %v, want %v", encoderConfig.TimeKey, LogEncoderTimeKey)
				}
				if encoderConfig.LevelKey != LogEncoderLevelKey {
					t.Errorf("LevelKey = %v, want %v", encoderConfig.LevelKey, LogEncoderLevelKey)
				}
				if encoderConfig.CallerKey != LogEncoderCallerKey {
					t.Errorf("CallerKey = %v, want %v", encoderConfig.CallerKey, LogEncoderCallerKey)
				}
				if encoderConfig.StacktraceKey != LogEncoderStacktraceKey {
					t.Errorf("StacktraceKey = %v, want %v", encoderConfig.StacktraceKey, LogEncoderStacktraceKey)
				}
			},
		},
		{
			name: "json encoding config with disabled caller and stacktrace should have empty keys",
			config: &Config{
				LogEncoding:       EncodingJSON,
				DisableCaller:     true,
				DisableStacktrace: true,
			},
			check: func(t *testing.T, encoderConfig zapcore.EncoderConfig) {
				if encoderConfig.CallerKey != "" {
					t.Errorf("CallerKey = %v, want empty", encoderConfig.CallerKey)
				}
				if encoderConfig.StacktraceKey != "" {
					t.Errorf("StacktraceKey = %v, want empty", encoderConfig.StacktraceKey)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoderConfig := buildEncoder(tt.config)
			tt.check(t, encoderConfig)
		})
	}
}

func TestGetZapConfigByMode(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, zapConfig zap.Config)
	}{
		{
			name: "development mode should return valid config",
			config: &Config{
				IsDevelopment: true,
			},
			check: func(t *testing.T, zapConfig zap.Config) {
				if zapConfig.Level.Level() == zapcore.InvalidLevel {
					t.Error("Level should be valid")
				}
				if zapConfig.Encoding == "" {
					t.Error("Encoding should not be empty")
				}
			},
		},
		{
			name: "production mode should return valid config",
			config: &Config{
				IsDevelopment: false,
			},
			check: func(t *testing.T, zapConfig zap.Config) {
				if zapConfig.Level.Level() == zapcore.InvalidLevel {
					t.Error("Level should be valid")
				}
				if zapConfig.Encoding == "" {
					t.Error("Encoding should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapConfig := getZapConfigByMode(tt.config)
			tt.check(t, zapConfig)
		})
	}
}

func TestGetCallerKey(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "caller enabled should return caller key",
			config: &Config{
				DisableCaller: false,
			},
			expected: LogEncoderCallerKey,
		},
		{
			name: "caller disabled should return empty string",
			config: &Config{
				DisableCaller: true,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCallerKey(tt.config)
			if result != tt.expected {
				t.Errorf("getCallerKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetStacktraceKey(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "stacktrace enabled should return stacktrace key",
			config: &Config{
				DisableStacktrace: false,
			},
			expected: LogEncoderStacktraceKey,
		},
		{
			name: "stacktrace disabled should return empty string",
			config: &Config{
				DisableStacktrace: true,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStacktraceKey(tt.config)
			if result != tt.expected {
				t.Errorf("getStacktraceKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEncodeLevel(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, result zapcore.LevelEncoder)
	}{
		{
			name: "console encoding should return non-nil encoder",
			config: &Config{
				LogEncoding: EncodingConsole,
			},
			check: func(t *testing.T, result zapcore.LevelEncoder) {
				if result == nil {
					t.Error("Level encoder should not be nil")
				}
			},
		},
		{
			name: "json encoding should return non-nil encoder",
			config: &Config{
				LogEncoding: EncodingJSON,
			},
			check: func(t *testing.T, result zapcore.LevelEncoder) {
				if result == nil {
					t.Error("Level encoder should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEncodeLevel(tt.config)
			tt.check(t, result)
		})
	}
}

func TestGetCallerEncoder(t *testing.T) {
	tests := []struct {
		name   string
		caller zapcore.EntryCaller
		check  func(t *testing.T, result string)
	}{
		{
			name: "caller with valid file path should format correctly",
			caller: zapcore.EntryCaller{
				Defined: true,
				PC:      0,
				File:    "/Users/test/project/logger/helper.go",
				Line:    42,
			},
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("getCallerEncoder() should not return empty string")
				}
				if !strings.Contains(result, ":42") {
					t.Errorf("getCallerEncoder() should contain line number, got %v", result)
				}
				if !strings.Contains(result, "helper.go") {
					t.Errorf("getCallerEncoder() should contain filename, got %v", result)
				}
			},
		},
		{
			name: "caller with long path should return short path",
			caller: zapcore.EntryCaller{
				Defined: true,
				PC:      0,
				File:    "/very/long/path/to/project/internal/logger/helper.go",
				Line:    100,
			},
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("getCallerEncoder() should not return empty string")
				}
				if !strings.Contains(result, ":100") {
					t.Errorf("getCallerEncoder() should contain line number, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock encoder to capture the result
			var capturedResult string
			mockEnc := &mockPrimitiveArrayEncoder{
				appendStringFunc: func(s string) {
					capturedResult = s
				},
			}

			// Call getCallerEncoder
			getCallerEncoder(tt.caller, mockEnc)

			// Verify the result
			tt.check(t, capturedResult)
		})
	}
}
