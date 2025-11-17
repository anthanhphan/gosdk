// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"io"
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config should return error",
			config:  nil,
			wantErr: true,
			errMsg:  "config is required, nil is not allowed",
		},
		{
			name: "empty log level should return error",
			config: &Config{
				LogLevel:    "",
				LogEncoding: EncodingJSON,
			},
			wantErr: true,
			errMsg:  "level is required",
		},
		{
			name: "empty log encoding should return error",
			config: &Config{
				LogLevel:    LevelInfo,
				LogEncoding: "",
			},
			wantErr: true,
			errMsg:  "encoding is required",
		},
		{
			name: "invalid log level should return error",
			config: &Config{
				LogLevel:    Level("invalid"),
				LogEncoding: EncodingJSON,
			},
			wantErr: true,
			errMsg:  "level is invalid, must be one of: debug, info, warn, error",
		},
		{
			name: "invalid log encoding should return error",
			config: &Config{
				LogLevel:    LevelInfo,
				LogEncoding: Encoding("invalid"),
			},
			wantErr: true,
			errMsg:  "encoding is invalid, must be one of: json, console",
		},
		{
			name: "valid config should not return error",
			config: &Config{
				LogLevel:          LevelInfo,
				LogEncoding:       EncodingJSON,
				DisableCaller:     false,
				DisableStacktrace: false,
				IsDevelopment:     false,
				LogOutputPaths:    []string{},
			},
			wantErr: false,
		},
		{
			name: "valid config with file output should not return error",
			config: &Config{
				LogLevel:          LevelDebug,
				LogEncoding:       EncodingConsole,
				DisableCaller:     true,
				DisableStacktrace: true,
				IsDevelopment:     true,
				LogOutputPaths:    []string{"app.log", "error.log"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, wantErrMsg %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestGetOutputWriters(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		check func(t *testing.T, writers []io.Writer)
	}{
		{
			name:  "empty paths should return stdout",
			paths: []string{},
			check: func(t *testing.T, writers []io.Writer) {
				if len(writers) != 1 {
					t.Errorf("getOutputWriters() len = %v, want 1", len(writers))
				}
				if writers[0] != os.Stdout {
					t.Errorf("getOutputWriters() should return stdout for empty paths")
				}
			},
		},
		{
			name:  "stdout path should return stdout",
			paths: []string{"stdout"},
			check: func(t *testing.T, writers []io.Writer) {
				if len(writers) != 1 {
					t.Errorf("getOutputWriters() len = %v, want 1", len(writers))
				}
				if writers[0] != os.Stdout {
					t.Errorf("getOutputWriters() should return stdout")
				}
			},
		},
		{
			name:  "empty string path should return stdout",
			paths: []string{""},
			check: func(t *testing.T, writers []io.Writer) {
				if len(writers) != 1 {
					t.Errorf("getOutputWriters() len = %v, want 1", len(writers))
				}
				if writers[0] != os.Stdout {
					t.Errorf("getOutputWriters() should return stdout for empty string")
				}
			},
		},
		{
			name:  "stderr path should return stderr",
			paths: []string{"stderr"},
			check: func(t *testing.T, writers []io.Writer) {
				if len(writers) != 1 {
					t.Errorf("getOutputWriters() len = %v, want 1", len(writers))
				}
				if writers[0] != os.Stderr {
					t.Errorf("getOutputWriters() should return stderr")
				}
			},
		},
		{
			name:  "multiple paths should return multiple writers",
			paths: []string{"stdout", "stderr"},
			check: func(t *testing.T, writers []io.Writer) {
				if len(writers) != 2 {
					t.Errorf("getOutputWriters() len = %v, want 2", len(writers))
				}
				if writers[0] != os.Stdout {
					t.Errorf("getOutputWriters() first writer should be stdout")
				}
				if writers[1] != os.Stderr {
					t.Errorf("getOutputWriters() second writer should be stderr")
				}
			},
		},
		{
			name:  "invalid file path should fallback to stdout",
			paths: []string{"/invalid/path/that/does/not/exist.log"},
			check: func(t *testing.T, writers []io.Writer) {
				if len(writers) != 1 {
					t.Errorf("getOutputWriters() len = %v, want 1", len(writers))
				}
				// Should fallback to stdout for invalid paths
				if writers[0] != os.Stdout {
					t.Errorf("getOutputWriters() should fallback to stdout for invalid paths")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writers := getOutputWriters(tt.paths)
			tt.check(t, writers)
		})
	}
}
