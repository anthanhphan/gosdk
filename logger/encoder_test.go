// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"strings"
	"testing"
	"time"
)

func TestNewConsoleEncoder(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, encoder *ConsoleEncoder)
	}{
		{
			name: "valid config should create console encoder",
			config: &Config{
				LogLevel:    LevelInfo,
				LogEncoding: EncodingConsole,
			},
			check: func(t *testing.T, encoder *ConsoleEncoder) {
				if encoder == nil {
					t.Fatal("NewConsoleEncoder() should not return nil")
				}
				if encoder.config == nil {
					t.Error("NewConsoleEncoder() config should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewConsoleEncoder(tt.config)
			tt.check(t, encoder)
		})
	}
}

func TestConsoleEncoder_getTimezone(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, loc *time.Location)
	}{
		{
			name: "empty timezone should return UTC",
			config: &Config{
				Timezone: "",
			},
			check: func(t *testing.T, loc *time.Location) {
				if loc.String() != "UTC" {
					t.Errorf("getTimezone() = %v, want UTC", loc.String())
				}
			},
		},
		{
			name: "valid timezone should return correct location",
			config: &Config{
				Timezone: "Asia/Ho_Chi_Minh",
			},
			check: func(t *testing.T, loc *time.Location) {
				if loc == nil {
					t.Fatal("getTimezone() should not return nil")
				}
				if loc.String() != "Asia/Ho_Chi_Minh" {
					t.Errorf("getTimezone() = %v, want Asia/Ho_Chi_Minh", loc.String())
				}
			},
		},
		{
			name: "invalid timezone should fallback to UTC",
			config: &Config{
				Timezone: "Invalid/Timezone",
			},
			check: func(t *testing.T, loc *time.Location) {
				if loc.String() != "UTC" {
					t.Errorf("getTimezone() = %v, want UTC", loc.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewConsoleEncoder(tt.config)
			loc := encoder.getTimezone()
			tt.check(t, loc)
		})
	}
}

func TestJSONEncoder_getTimezone(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		check  func(t *testing.T, loc *time.Location)
	}{
		{
			name: "empty timezone should return UTC",
			config: &Config{
				Timezone: "",
			},
			check: func(t *testing.T, loc *time.Location) {
				if loc.String() != "UTC" {
					t.Errorf("getTimezone() = %v, want UTC", loc.String())
				}
			},
		},
		{
			name: "valid timezone should return correct location",
			config: &Config{
				Timezone: "America/New_York",
			},
			check: func(t *testing.T, loc *time.Location) {
				if loc == nil {
					t.Fatal("getTimezone() should not return nil")
				}
				if loc.String() != "America/New_York" {
					t.Errorf("getTimezone() = %v, want America/New_York", loc.String())
				}
			},
		},
		{
			name: "invalid timezone should fallback to UTC",
			config: &Config{
				Timezone: "Invalid/Timezone",
			},
			check: func(t *testing.T, loc *time.Location) {
				if loc.String() != "UTC" {
					t.Errorf("getTimezone() = %v, want UTC", loc.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewJSONEncoder(tt.config)
			loc := encoder.getTimezone()
			tt.check(t, loc)
		})
	}
}

func TestConsoleEncoder_Encode(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		entry  *Entry
		check  func(t *testing.T, output string)
	}{
		{
			name: "entry without caller should encode correctly",
			config: &Config{
				LogLevel:      LevelInfo,
				LogEncoding:   EncodingConsole,
				IsDevelopment: false,
				DisableCaller: true,
			},
			entry: &Entry{
				Time:    time.Now(),
				Level:   LevelInfo,
				Message: "test message",
				Fields:  map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "test message") {
					t.Errorf("Encode() output should contain message, got %v", output)
				}
				if !strings.Contains(output, "INFO") {
					t.Errorf("Encode() output should contain level, got %v", output)
				}
			},
		},
		{
			name: "entry with caller should encode correctly",
			config: &Config{
				LogLevel:      LevelInfo,
				LogEncoding:   EncodingConsole,
				IsDevelopment: false,
				DisableCaller: false,
			},
			entry: &Entry{
				Time:    time.Now(),
				Level:   LevelInfo,
				Message: "test message",
				Caller: &CallerInfo{
					File: "test.go",
					Line: 42,
				},
				Fields: map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "test message") {
					t.Errorf("Encode() output should contain message, got %v", output)
				}
				if !strings.Contains(output, "test.go:42") {
					t.Errorf("Encode() output should contain caller, got %v", output)
				}
			},
		},
		{
			name: "entry with fields should encode correctly",
			config: &Config{
				LogLevel:      LevelInfo,
				LogEncoding:   EncodingConsole,
				IsDevelopment: false,
				DisableCaller: true,
			},
			entry: &Entry{
				Time:    time.Now(),
				Level:   LevelInfo,
				Message: "test message",
				Fields: map[string]interface{}{
					"key1": "value1",
					"key2": 123,
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "test message") {
					t.Errorf("Encode() output should contain message, got %v", output)
				}
				if !strings.Contains(output, "key1=value1") {
					t.Errorf("Encode() output should contain field, got %v", output)
				}
			},
		},
		{
			name: "entry with stacktrace should encode correctly",
			config: &Config{
				LogLevel:      LevelError,
				LogEncoding:   EncodingConsole,
				IsDevelopment: false,
				DisableCaller: true,
			},
			entry: &Entry{
				Time:       time.Now(),
				Level:      LevelError,
				Message:    "test error",
				Stacktrace: "stack trace here",
				Fields:     map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "test error") {
					t.Errorf("Encode() output should contain message, got %v", output)
				}
				if !strings.Contains(output, "stack trace here") {
					t.Errorf("Encode() output should contain stacktrace, got %v", output)
				}
			},
		},
		{
			name: "development mode should colorize levels",
			config: &Config{
				LogLevel:      LevelInfo,
				LogEncoding:   EncodingConsole,
				IsDevelopment: true,
				DisableCaller: true,
			},
			entry: &Entry{
				Time:    time.Now(),
				Level:   LevelInfo,
				Message: "test message",
				Fields:  map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				// Check for ANSI color codes
				if !strings.Contains(output, "\033[32m") {
					t.Errorf("Encode() output should contain color code for INFO, got %v", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewConsoleEncoder(tt.config)
			output := encoder.Encode(tt.entry)
			tt.check(t, output)
		})
	}
}

func TestColorizeLevel(t *testing.T) {
	tests := []struct {
		name     string
		levelStr string
		level    Level
		check    func(t *testing.T, output string)
	}{
		{
			name:     "debug level should be cyan",
			levelStr: "DEBUG",
			level:    LevelDebug,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "\033[36m") {
					t.Errorf("colorizeLevel() should contain cyan color code, got %v", output)
				}
			},
		},
		{
			name:     "info level should be green",
			levelStr: "INFO",
			level:    LevelInfo,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "\033[32m") {
					t.Errorf("colorizeLevel() should contain green color code, got %v", output)
				}
			},
		},
		{
			name:     "warn level should be yellow",
			levelStr: "WARN",
			level:    LevelWarn,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "\033[33m") {
					t.Errorf("colorizeLevel() should contain yellow color code, got %v", output)
				}
			},
		},
		{
			name:     "error level should be red",
			levelStr: "ERROR",
			level:    LevelError,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "\033[31m") {
					t.Errorf("colorizeLevel() should contain red color code, got %v", output)
				}
			},
		},
		{
			name:     "unknown level should return unchanged",
			levelStr: "UNKNOWN",
			level:    Level("unknown"),
			check: func(t *testing.T, output string) {
				// The output will be the full encoded entry, so we check it contains UNKNOWN without color codes
				if !strings.Contains(output, "UNKNOWN") {
					t.Errorf("colorizeLevel() output should contain UNKNOWN, got %v", output)
				}
				// Should not contain color codes for unknown level
				if strings.Contains(output, "\033[") {
					t.Errorf("colorizeLevel() should not contain color codes for unknown level, got %v", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{IsDevelopment: true}
			encoder := NewConsoleEncoder(config)
			// We need to test colorizeLevel indirectly through Encode
			entry := &Entry{
				Time:    time.Now(),
				Level:   tt.level,
				Message: "test",
				Fields:  map[string]interface{}{},
			}
			output := encoder.Encode(entry)
			tt.check(t, output)
		})
	}
}
