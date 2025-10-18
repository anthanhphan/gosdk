// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
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
