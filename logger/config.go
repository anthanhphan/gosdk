// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"errors"
	"strings"
)

// Config represents the configuration for the logger.
// It defines log level, encoding format, output destinations, and various logging options.
type Config struct {
	// LogLevel specifies the minimum log level to output (e.g., DEBUG, INFO, WARN, ERROR).
	LogLevel Level `yaml:"log_level" json:"log_level"`

	// LogEncoding defines the output format for log entries (e.g., JSON, CONSOLE).
	LogEncoding Encoding `yaml:"log_encoding" json:"log_encoding"`

	// DisableCaller controls whether to include caller information (file:line) in log entries.
	DisableCaller bool `yaml:"disable_caller" json:"disable_caller"`

	// DisableStacktrace controls whether to include stack traces for error-level logs.
	DisableStacktrace bool `yaml:"disable_stacktrace" json:"disable_stacktrace"`

	// IsDevelopment enables development mode with more verbose output and human-readable formatting.
	IsDevelopment bool `yaml:"is_development" json:"is_development"`

	// LogOutputPaths specifies file paths to write log output to. If empty, logs to console.
	LogOutputPaths []string `yaml:"log_output_paths" json:"log_output_paths"`
}

// Validate checks if the configuration is valid and all required fields are set.
//
// Input:
//   - None
//
// Output:
//   - error: Validation error if config is invalid, nil if valid
//
// Example:
//
//	config := &Config{
//	    LogLevel: LevelInfo,
//	    LogEncoding: EncodingJSON,
//	}
//	if err := config.Validate(); err != nil {
//	    log.Fatal("Invalid config:", err)
//	}
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required, nil is not allowed")
	}
	if c.LogLevel == "" {
		return errors.New("level is required")
	}
	if c.LogEncoding == "" {
		return errors.New("encoding is required")
	}
	if !c.LogLevel.isValid() {
		return errors.New("level is invalid, must be one of: " + strings.Join(levelValues(), ", "))
	}
	if !c.LogEncoding.isValid() {
		return errors.New("encoding is invalid, must be one of: " + strings.Join(encodingValues(), ", "))
	}

	return nil
}
