// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/anthanhphan/gosdk/utils"
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

	// Timezone specifies the timezone for timestamp formatting. If empty, uses UTC.
	// Must be a valid IANA timezone name (e.g., "America/New_York", "Asia/Tokyo", "UTC").
	Timezone string `yaml:"timezone" json:"timezone"`
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

// buildLoggerConfig creates a logger instance from the config.
func buildLoggerConfig(config *Config, defaultFields ...Field) *Logger {
	// Get output writers
	outputs := getOutputWriters(config.LogOutputPaths)

	// Create logger
	return NewLogger(config, outputs, defaultFields...)
}

// getOutputWriters returns a slice of io.Writer based on output paths.
func getOutputWriters(paths []string) []io.Writer {
	if len(paths) == 0 {
		return []io.Writer{os.Stdout}
	}

	writers := make([]io.Writer, 0, len(paths))
	for _, path := range paths {
		switch path {
		case "stdout", "":
			writers = append(writers, os.Stdout)
		case "stderr":
			writers = append(writers, os.Stderr)
		default:
			// Open file securely using utils.OpenFileSecurely to prevent directory traversal
			// Use 0600 permissions (read/write for owner only) for security
			file, err := utils.OpenFileSecurely(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				// Fallback to stdout if file cannot be opened
				// Note: This error is silently handled to avoid breaking logger initialization
				// In production, you may want to log this error or handle it differently
				writers = append(writers, os.Stdout)
			} else {
				writers = append(writers, file)
			}
		}
	}

	if len(writers) == 0 {
		return []io.Writer{os.Stdout}
	}

	return writers
}

// getShortPathForCaller returns a short path for caller information.
func getShortPathForCaller(fullPath string) string {
	return utils.GetShortPath(fullPath)
}
