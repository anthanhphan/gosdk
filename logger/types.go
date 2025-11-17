// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

// Level represents the log level for filtering log messages.
type Level string

var validLevels = map[Level]struct{}{
	LevelDebug: {},
	LevelInfo:  {},
	LevelWarn:  {},
	LevelError: {},
}

var levelValuesCache = []string{"debug", "info", "warn", "error"}

func (l Level) isValid() bool {
	_, ok := validLevels[l]
	return ok
}

func levelValues() []string {
	return levelValuesCache
}

// Encoding represents the output format for log messages.
type Encoding string

var validEncodings = map[Encoding]struct{}{
	EncodingJSON:    {},
	EncodingConsole: {},
}

var encodingValuesCache = []string{"json", "console"}

func (e Encoding) isValid() bool {
	_, ok := validEncodings[e]
	return ok
}

func encodingValues() []string {
	return encodingValuesCache
}

// Log level constants for filtering log messages.
const (
	// LevelDebug represents debug level logs (most verbose).
	LevelDebug Level = "debug"
	// LevelInfo represents informational level logs.
	LevelInfo Level = "info"
	// LevelWarn represents warning level logs.
	LevelWarn Level = "warn"
	// LevelError represents error level logs (least verbose).
	LevelError Level = "error"
)

// Log encoding constants for output format.
const (
	// EncodingJSON represents structured JSON output format.
	EncodingJSON Encoding = "json"
	// EncodingConsole represents human-readable console output format.
	EncodingConsole Encoding = "console"
)

// Log encoder key constants for structured log fields.
const (
	// LogEncoderMessageKey is the key for log message content.
	LogEncoderMessageKey = "msg"
	// LogEncoderTimeKey is the key for timestamp field.
	LogEncoderTimeKey = "ts"
	// LogEncoderLevelKey is the key for log level field.
	LogEncoderLevelKey = "level"
	// LogEncoderCallerKey is the key for caller information (file:line).
	LogEncoderCallerKey = "caller"
	// LogEncoderNameKey is the key for logger name field.
	LogEncoderNameKey = "logger"
	// LogEncoderStacktraceKey is the key for stack trace field.
	LogEncoderStacktraceKey = "stacktrace"
)

// DevelopmentConfig provides a pre-configured logger setup optimized for development.
// It enables debug level logging, includes caller information and stack traces,
// and uses JSON encoding for structured output.
var DevelopmentConfig = Config{
	LogLevel:          LevelDebug,
	LogEncoding:       EncodingJSON,
	DisableCaller:     false,
	DisableStacktrace: false,
	IsDevelopment:     true,
}

// DefaultConfig is kept for backward compatibility.
// Deprecated: Use DevelopmentConfig instead.
var DefaultConfig = DevelopmentConfig

// ProductionConfig provides a pre-configured logger setup optimized for production.
// It uses info level logging, disables stack traces for performance,
// and uses JSON encoding for structured output.
var ProductionConfig = Config{
	LogLevel:          LevelInfo,
	LogEncoding:       EncodingJSON,
	DisableCaller:     false,
	DisableStacktrace: true,
	IsDevelopment:     false,
}
