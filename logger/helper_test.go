// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"strings"
	"testing"
)

func TestBuildLoggerConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		defaultFields []Field
		check         func(t *testing.T, logger *Logger)
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
			defaultFields: []Field{},
			check: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Fatal("Logger should be created")
				}
				if logger.config.LogLevel != LevelDebug {
					t.Errorf("LogLevel = %v, want %v", logger.config.LogLevel, LevelDebug)
				}
				if logger.config.LogEncoding != EncodingConsole {
					t.Errorf("LogEncoding = %v, want %v", logger.config.LogEncoding, EncodingConsole)
				}
				if logger.config.DisableCaller {
					t.Error("DisableCaller should be false")
				}
				if logger.config.DisableStacktrace {
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
				LogOutputPaths:    []string{"log/app.log"},
			},
			defaultFields: []Field{},
			check: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Fatal("Logger should be created")
				}
				if logger.config.LogLevel != LevelInfo {
					t.Errorf("LogLevel = %v, want %v", logger.config.LogLevel, LevelInfo)
				}
				if logger.config.LogEncoding != EncodingJSON {
					t.Errorf("LogEncoding = %v, want %v", logger.config.LogEncoding, EncodingJSON)
				}
				if logger.config.DisableCaller {
					t.Error("DisableCaller should be false")
				}
				if !logger.config.DisableStacktrace {
					t.Error("DisableStacktrace should be true")
				}
			},
		},
		{
			name: "config with default fields should include them",
			config: &Config{
				LogLevel:          LevelWarn,
				LogEncoding:       EncodingJSON,
				DisableCaller:     true,
				DisableStacktrace: false,
				IsDevelopment:     false,
				LogOutputPaths:    []string{},
			},
			defaultFields: []Field{
				String("app", "test"),
				Int("version", 1),
			},
			check: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Fatal("Logger should be created")
				}
				if logger.config.LogLevel != LevelWarn {
					t.Errorf("LogLevel = %v, want %v", logger.config.LogLevel, LevelWarn)
				}
				if !logger.config.DisableCaller {
					t.Error("DisableCaller should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := buildLoggerConfig(tt.config, tt.defaultFields...)
			tt.check(t, logger)
		})
	}
}

func TestGetShortPathForCaller(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string
		check    func(t *testing.T, result string)
	}{
		{
			name:     "short path should return as is",
			fullPath: "helper.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("getShortPathForCaller() should not return empty string")
				}
			},
		},
		{
			name:     "long path should return short path",
			fullPath: "/very/long/path/to/project/internal/logger/helper.go",
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("getShortPathForCaller() should not return empty string")
				}
				if strings.Contains(result, "/very/long/path") {
					t.Errorf("getShortPathForCaller() should return short path, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getShortPathForCaller(tt.fullPath)
			tt.check(t, result)
		})
	}
}
