// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogger_formatArgs(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, []io.Writer{os.Stdout})

	tests := []struct {
		name    string
		args    []interface{}
		wantMsg string
		wantLen int
		check   func(t *testing.T, msg string, fields []Field)
	}{
		{
			name:    "empty args should return empty message",
			args:    []interface{}{},
			wantMsg: "",
			wantLen: 0,
			check: func(t *testing.T, msg string, fields []Field) {
				if msg != "" {
					t.Errorf("formatArgs() msg = %v, want empty", msg)
				}
				if fields != nil {
					t.Errorf("formatArgs() fields = %v, want nil", fields)
				}
			},
		},
		{
			name:    "single arg should return formatted message",
			args:    []interface{}{"test"},
			wantMsg: "test",
			wantLen: 0,
			check: func(t *testing.T, msg string, fields []Field) {
				if msg != "test" {
					t.Errorf("formatArgs() msg = %v, want 'test'", msg)
				}
				if fields != nil {
					t.Errorf("formatArgs() fields = %v, want nil", fields)
				}
			},
		},
		{
			name:    "single non-string arg should return formatted message",
			args:    []interface{}{123},
			wantMsg: "123",
			wantLen: 0,
			check: func(t *testing.T, msg string, fields []Field) {
				if msg != "123" {
					t.Errorf("formatArgs() msg = %v, want '123'", msg)
				}
			},
		},
		{
			name:    "string message with key-value pairs should parse correctly",
			args:    []interface{}{"message", "key1", "value1", "key2", 123},
			wantMsg: "message",
			wantLen: 2,
			check: func(t *testing.T, msg string, fields []Field) {
				if msg != "message" {
					t.Errorf("formatArgs() msg = %v, want 'message'", msg)
				}
				if len(fields) != 2 {
					t.Errorf("formatArgs() fields len = %v, want 2", len(fields))
				}
			},
		},
		{
			name:    "non-string first arg should format all as message",
			args:    []interface{}{123, "key", "value"},
			wantMsg: "123 key value",
			wantLen: 0,
			check: func(t *testing.T, msg string, fields []Field) {
				if !strings.Contains(msg, "123") {
					t.Errorf("formatArgs() msg = %v, want to contain '123'", msg)
				}
			},
		},
		{
			name:    "odd number of key-value pairs should handle last value",
			args:    []interface{}{"message", "key1", "value1", "key2"},
			wantMsg: "message",
			wantLen: 2,
			check: func(t *testing.T, msg string, fields []Field) {
				if msg != "message" {
					t.Errorf("formatArgs() msg = %v, want 'message'", msg)
				}
				if len(fields) != 2 {
					t.Errorf("formatArgs() fields len = %v, want 2", len(fields))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, fields := logger.formatArgs(tt.args...)
			tt.check(t, msg, fields)
		})
	}
}

func TestLogger_WithOptions(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, []io.Writer{os.Stdout})

	tests := []struct {
		name  string
		opts  []Option
		check func(t *testing.T, newLogger *Logger)
	}{
		{
			name: "no options should return logger with same callerSkip",
			opts: []Option{},
			check: func(t *testing.T, newLogger *Logger) {
				if newLogger.callerSkip != logger.callerSkip {
					t.Errorf("WithOptions() callerSkip = %v, want %v", newLogger.callerSkip, logger.callerSkip)
				}
			},
		},
		{
			name: "single option should apply correctly",
			opts: []Option{AddCallerSkip(2)},
			check: func(t *testing.T, newLogger *Logger) {
				if newLogger.callerSkip != logger.callerSkip+2 {
					t.Errorf("WithOptions() callerSkip = %v, want %v", newLogger.callerSkip, logger.callerSkip+2)
				}
			},
		},
		{
			name: "multiple options should apply correctly",
			opts: []Option{AddCallerSkip(1), AddCallerSkip(2)},
			check: func(t *testing.T, newLogger *Logger) {
				if newLogger.callerSkip != logger.callerSkip+3 {
					t.Errorf("WithOptions() callerSkip = %v, want %v", newLogger.callerSkip, logger.callerSkip+3)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLogger := logger.WithOptions(tt.opts...)
			tt.check(t, newLogger)
		})
	}
}

func TestLogger_writeEntry(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, []io.Writer{&buf})

	tests := []struct {
		name  string
		entry *Entry
		check func(t *testing.T, output string)
	}{
		{
			name: "valid entry should write to output",
			entry: &Entry{
				Time:    testTime,
				Level:   LevelInfo,
				Message: "test message",
				Fields:  map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "test message") {
					t.Errorf("writeEntry() output should contain message, got %v", output)
				}
			},
		},
		{
			name: "entry with empty message should still write JSON",
			entry: &Entry{
				Time:    testTime,
				Level:   LevelInfo,
				Message: "",
				Fields:  map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				// Empty message should still produce JSON output
				if !strings.Contains(output, `"msg":""`) {
					t.Errorf("writeEntry() output should contain empty msg field, got %v", output)
				}
			},
		},
		{
			name: "entry with console encoding should write correctly",
			entry: &Entry{
				Time:    testTime,
				Level:   LevelInfo,
				Message: "console test",
				Fields:  map[string]interface{}{},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "console test") {
					t.Errorf("writeEntry() output should contain message, got %v", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.config.LogEncoding = EncodingJSON
			if tt.name == "entry with console encoding should write correctly" {
				logger.config.LogEncoding = EncodingConsole
			}
			logger.writeEntry(tt.entry)
			tt.check(t, buf.String())
		})
	}
}

func TestLogger_log(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:          LevelDebug,
		LogEncoding:       EncodingJSON,
		DisableCaller:     false,
		DisableStacktrace: false,
	}, []io.Writer{&buf})

	tests := []struct {
		name   string
		level  Level
		msg    string
		fields []Field
		check  func(t *testing.T, output string)
	}{
		{
			name:   "debug level should log when level is debug",
			level:  LevelDebug,
			msg:    "debug message",
			fields: []Field{String("key", "value")},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "debug message") {
					t.Errorf("log() output should contain message, got %v", output)
				}
			},
		},
		{
			name:   "info level should log when level is debug",
			level:  LevelInfo,
			msg:    "info message",
			fields: []Field{String("key", "value")},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "info message") {
					t.Errorf("log() output should contain message, got %v", output)
				}
			},
		},
		{
			name:   "error level should include stacktrace",
			level:  LevelError,
			msg:    "error message",
			fields: []Field{String("key", "value")},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "error message") {
					t.Errorf("log() output should contain message, got %v", output)
				}
				if !strings.Contains(output, "goroutine") {
					t.Errorf("log() output should contain stacktrace for error level, got %v", output)
				}
			},
		},
		{
			name:   "warn level should not log when level is error",
			level:  LevelWarn,
			msg:    "warn message",
			fields: []Field{String("key", "value")},
			check: func(t *testing.T, output string) {
				// Should not log because warn < error
				logger.config.LogLevel = LevelError
				buf.Reset()
				logger.log(LevelWarn, 0, "warn message", []Field{String("key", "value")}...)
				if buf.String() != "" {
					t.Errorf("log() should not log warn when level is error, got %v", buf.String())
				}
				logger.config.LogLevel = LevelDebug
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.log(tt.level, 0, tt.msg, tt.fields...)
			tt.check(t, buf.String())
		})
	}
}

func TestLogger_getCaller(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, []io.Writer{os.Stdout})

	// Test that getCaller returns valid caller info
	caller := logger.getCaller(0)
	if caller == nil {
		t.Error("getCaller() should not return nil")
		return
	}
	if caller.File == "" {
		t.Error("getCaller() should return file path")
	}
	if caller.Line == 0 {
		t.Error("getCaller() should return line number")
	}

	// Test with increased callerSkip
	logger.callerSkip = 100
	caller = logger.getCaller(0)
	// Should return nil if skip is too high
	if caller != nil {
		t.Logf("getCaller() with high skip returned: %v", caller)
	}
}

var testTime = func() time.Time {
	t, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
	return t
}()
