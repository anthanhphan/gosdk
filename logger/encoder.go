// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Encoder encodes log entries to strings.
type Encoder interface {
	Encode(entry *Entry) string
}

// JSONEncoder encodes entries as JSON.
type JSONEncoder struct {
	config   *Config
	timezone *time.Location
}

// NewJSONEncoder creates a new JSON encoder.
func NewJSONEncoder(config *Config) *JSONEncoder {
	encoder := &JSONEncoder{config: config}
	encoder.timezone = encoder.getTimezone()
	return encoder
}

// getTimezone returns the timezone from config or UTC as default.
func (e *JSONEncoder) getTimezone() *time.Location {
	if e.config.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(e.config.Timezone)
	if err != nil {
		// If timezone is invalid, fall back to UTC
		return time.UTC
	}
	return loc
}

// Encode encodes an entry as JSON with ts and caller as first keys.
func (e *JSONEncoder) Encode(entry *Entry) string {
	entryTime := entry.Time.In(e.timezone)

	estimatedSize := 256 + len(entry.Message) + len(entry.Fields)*32
	var builder strings.Builder
	builder.Grow(estimatedSize)

	builder.WriteByte('{')

	tsValue := entryTime.Format(time.RFC3339)
	tsJSON, _ := json.Marshal(tsValue)
	builder.WriteString(`"` + LogEncoderTimeKey + `":`)
	builder.Write(tsJSON)

	if entry.Caller != nil {
		builder.WriteByte(',')
		callerValue := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		callerJSON, _ := json.Marshal(callerValue)
		builder.WriteString(`"` + LogEncoderCallerKey + `":`)
		builder.Write(callerJSON)
	}

	builder.WriteByte(',')
	levelValue := strings.ToLower(string(entry.Level))
	levelJSON, _ := json.Marshal(levelValue)
	builder.WriteString(`"` + LogEncoderLevelKey + `":`)
	builder.Write(levelJSON)

	builder.WriteByte(',')
	msgJSON, _ := json.Marshal(entry.Message)
	builder.WriteString(`"` + LogEncoderMessageKey + `":`)
	builder.Write(msgJSON)

	if entry.Stacktrace != "" {
		builder.WriteByte(',')
		stackJSON, _ := json.Marshal(entry.Stacktrace)
		builder.WriteString(`"` + LogEncoderStacktraceKey + `":`)
		builder.Write(stackJSON)
	}

	for k, v := range entry.Fields {
		fieldJSON, err := json.Marshal(v)
		if err != nil {
			continue
		}
		builder.WriteByte(',')
		keyJSON, _ := json.Marshal(k)
		builder.Write(keyJSON)
		builder.WriteByte(':')
		builder.Write(fieldJSON)
	}

	builder.WriteByte('}')
	builder.WriteByte('\n')

	return builder.String()
}

// ConsoleEncoder encodes entries as human-readable console output.
type ConsoleEncoder struct {
	config   *Config
	timezone *time.Location
}

// NewConsoleEncoder creates a new console encoder.
func NewConsoleEncoder(config *Config) *ConsoleEncoder {
	encoder := &ConsoleEncoder{config: config}
	encoder.timezone = encoder.getTimezone()
	return encoder
}

// getTimezone returns the timezone from config or UTC as default.
func (e *ConsoleEncoder) getTimezone() *time.Location {
	if e.config.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(e.config.Timezone)
	if err != nil {
		// If timezone is invalid, fall back to UTC
		return time.UTC
	}
	return loc
}

// Encode encodes an entry as console output.
func (e *ConsoleEncoder) Encode(entry *Entry) string {
	entryTime := entry.Time.In(e.timezone)
	timeStr := entryTime.Format(time.RFC3339Nano)

	levelStr := strings.ToUpper(string(entry.Level))
	if e.config.IsDevelopment {
		levelStr = colorizeLevel(levelStr, entry.Level)
	}

	var builder strings.Builder
	estimatedSize := len(timeStr) + len(levelStr) + len(entry.Message) + 64
	if entry.Caller != nil {
		estimatedSize += len(entry.Caller.File) + 16
	}
	if len(entry.Fields) > 0 {
		estimatedSize += len(entry.Fields) * 32
	}
	builder.Grow(estimatedSize)

	builder.WriteString(timeStr)
	builder.WriteByte('\t')
	builder.WriteString(levelStr)

	if entry.Caller != nil {
		builder.WriteByte('\t')
		builder.WriteString(entry.Caller.File)
		builder.WriteByte(':')
		builder.WriteString(fmt.Sprintf("%d", entry.Caller.Line))
	}

	builder.WriteByte('\t')
	builder.WriteString(entry.Message)

	if len(entry.Fields) > 0 {
		builder.WriteByte('\t')
		first := true
		for k, v := range entry.Fields {
			if !first {
				builder.WriteByte(' ')
			}
			builder.WriteString(k)
			builder.WriteByte('=')
			builder.WriteString(fmt.Sprintf("%v", v))
			first = false
		}
	}

	builder.WriteByte('\n')

	if entry.Stacktrace != "" {
		builder.WriteString(entry.Stacktrace)
		builder.WriteByte('\n')
	}

	return builder.String()
}

// colorizeLevel adds ANSI color codes to log levels.
func colorizeLevel(levelStr string, level Level) string {
	switch level {
	case LevelDebug:
		return "\033[36m" + levelStr + "\033[0m" // Cyan
	case LevelInfo:
		return "\033[32m" + levelStr + "\033[0m" // Green
	case LevelWarn:
		return "\033[33m" + levelStr + "\033[0m" // Yellow
	case LevelError:
		return "\033[31m" + levelStr + "\033[0m" // Red
	default:
		return levelStr
	}
}
