// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

//go:generate mockgen -source=encoder.go -destination=mocks/mock_encoder.go -package=mocks

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// Encoder encodes log entries to strings.
type Encoder interface {
	Encode(entry *Entry) string
}

// JSONEncoder encodes entries as JSON.
type JSONEncoder struct {
	config   *Config
	timezone *time.Location
}

// newJSONEncoder creates a new JSON encoder with the provided configuration.
func newJSONEncoder(config *Config) *JSONEncoder {
	encoder := &JSONEncoder{config: config}
	encoder.timezone = encoder.getTimezone()
	return encoder
}

func (e *JSONEncoder) getTimezone() *time.Location {
	if e.config.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(e.config.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

// Encode encodes an entry as JSON with ts and caller as first keys.
//
// Input:
//   - entry: The log entry to encode
//
// Output:
//   - string: JSON-encoded log entry as a string
//
// Example:
//
//	entry := &Entry{
//	    Time:    time.Now(),
//	    Level:   LevelInfo,
//	    Message: "User created",
//	    Fields:  map[string]interface{}{"user_id": 12345},
//	}
//	json := encoder.Encode(entry)
func (e *JSONEncoder) Encode(entry *Entry) string {
	entryTime := entry.Time.In(e.timezone)

	estimatedSize := 256 + len(entry.Message) + len(entry.Fields)*32
	var builder strings.Builder
	builder.Grow(estimatedSize)

	builder.WriteByte('{')

	builder.WriteString(`"` + LogEncoderTimeKey + `":"`)
	builder.WriteString(entryTime.Format(time.RFC3339Nano))
	builder.WriteByte('"')

	if entry.Caller != nil {
		builder.WriteByte(',')
		builder.WriteString(`"` + LogEncoderCallerKey + `":"`)
		builder.WriteString(entry.Caller.File)
		builder.WriteByte(':')
		builder.WriteString(strconv.Itoa(entry.Caller.Line))
		builder.WriteByte('"')
	}

	builder.WriteByte(',')
	builder.WriteString(`"` + LogEncoderLevelKey + `":"`)
	builder.WriteString(strings.ToLower(string(entry.Level)))
	builder.WriteByte('"')

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

// newConsoleEncoder creates a new console encoder with the provided configuration.
func newConsoleEncoder(config *Config) *ConsoleEncoder {
	encoder := &ConsoleEncoder{config: config}
	encoder.timezone = encoder.getTimezone()
	return encoder
}

func (e *ConsoleEncoder) getTimezone() *time.Location {
	if e.config.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(e.config.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

// Encode encodes an entry as human-readable console output.
//
// Input:
//   - entry: The log entry to encode
//
// Output:
//   - string: Console-formatted log entry as a string
//
// Example:
//
//	entry := &Entry{
//	    Time:    time.Now(),
//	    Level:   LevelInfo,
//	    Message: "User created",
//	    Fields:  map[string]interface{}{"user_id": 12345},
//	}
//	console := encoder.Encode(entry)
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
		builder.WriteString(strconv.Itoa(entry.Caller.Line))
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

func colorizeLevel(levelStr string, level Level) string {
	switch level {
	case LevelDebug:
		return "\033[36m" + levelStr + "\033[0m"
	case LevelInfo:
		return "\033[32m" + levelStr + "\033[0m"
	case LevelWarn:
		return "\033[33m" + levelStr + "\033[0m"
	case LevelError:
		return "\033[31m" + levelStr + "\033[0m"
	default:
		return levelStr
	}
}
