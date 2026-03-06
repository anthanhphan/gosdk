// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

//go:generate mockgen -source=encoder.go -destination=mocks/mock_encoder.go -package=mocks

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	jsoniter "github.com/json-iterator/go"
)

var jsonLib = jsoniter.ConfigFastest

// Pre-computed JSON key prefixes as byte slices.
var (
	jsonKeyTimeBytes       = []byte(`"` + LogEncoderTimeKey + `":"`)
	jsonKeyLevelBytes      = []byte(`"` + LogEncoderLevelKey + `":"`)
	jsonKeyMessageBytes    = []byte(`"` + LogEncoderMessageKey + `":`)
	jsonKeyCallerBytes     = []byte(`"` + LogEncoderCallerKey + `":"`)
	jsonKeyStacktraceBytes = []byte(`"` + LogEncoderStacktraceKey + `":`)
)



// levelToLower returns the lowercase level string (switch avoids map lookup overhead).
func levelToLower(level Level) string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return strings.ToLower(string(level))
	}
}

// levelToUpper returns the uppercase level string.
func levelToUpper(level Level) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return strings.ToUpper(string(level))
	}
}

// bufPool pools byte slices for encoding.
var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return &b
	},
}

func getBuf() *[]byte {
	bp := bufPool.Get().(*[]byte)
	*bp = (*bp)[:0]
	return bp
}

func putBuf(bp *[]byte) {
	if cap(*bp) > 8192 {
		return
	}
	bufPool.Put(bp)
}

// resolveTimezone returns the time.Location for the given timezone string.
func resolveTimezone(timezone string) *time.Location {
	if timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

// Encoder encodes log entries.
type Encoder interface {
	// Encode encodes an entry and returns the result as a string.
	Encode(entry *Entry) string

	// EncodeTo encodes an entry and writes the bytes directly to the WriteSyncer.
	// Returns the number of bytes written. This is the preferred method for
	// production use as it avoids the []byte → string copy.
	EncodeTo(entry *Entry, ws WriteSyncer) (int, error)
}

// JSONEncoder encodes entries as JSON.
type JSONEncoder struct {
	config   *Config
	timezone *time.Location
}

func newJSONEncoder(config *Config) *JSONEncoder {
	return &JSONEncoder{
		config:   config,
		timezone: resolveTimezone(config.Timezone),
	}
}

// encodeToBuffer encodes an entry into a pooled byte buffer.
// Caller must call putBuf(bp) after consuming the buffer.
func (e *JSONEncoder) encodeToBuffer(entry *Entry) *[]byte {
	bp := getBuf()
	buf := *bp

	buf = append(buf, '{')

	// ts
	entryTime := entry.Time.In(e.timezone)
	buf = append(buf, jsonKeyTimeBytes...)
	buf = entryTime.AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, '"')

	// caller
	if entry.CallerDefined {
		buf = append(buf, ',')
		buf = append(buf, jsonKeyCallerBytes...)
		buf = append(buf, entry.CallerFile...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(entry.CallerLine), 10)
		buf = append(buf, '"')
	}

	// level
	buf = append(buf, ',')
	buf = append(buf, jsonKeyLevelBytes...)
	buf = append(buf, levelToLower(entry.Level)...)
	buf = append(buf, '"')

	// msg
	buf = append(buf, ',')
	buf = append(buf, jsonKeyMessageBytes...)
	buf = appendJSONString(buf, entry.Message)

	// stacktrace
	if entry.Stacktrace != "" {
		buf = append(buf, ',')
		buf = append(buf, jsonKeyStacktraceBytes...)
		buf = appendJSONString(buf, entry.Stacktrace)
	}

	// Single-pass: find priority field indices, write them first, then rest
	var priorityIdx [2]int // indices for trace_id, request_id (-1 = not found)
	priorityIdx[0] = -1
	priorityIdx[1] = -1
	for i := range entry.Fields {
		switch entry.Fields[i].Key {
		case "trace_id":
			priorityIdx[0] = i
		case "request_id":
			priorityIdx[1] = i
		}
	}

	// Write priority fields first
	for _, idx := range priorityIdx {
		if idx >= 0 {
			buf = append(buf, ',')
			buf = appendJSONString(buf, entry.Fields[idx].Key)
			buf = append(buf, ':')
			buf = appendTypedFieldValue(buf, &entry.Fields[idx])
		}
	}

	// Remaining fields (skip priority)
	for i := range entry.Fields {
		if i == priorityIdx[0] || i == priorityIdx[1] {
			continue
		}
		buf = append(buf, ',')
		buf = appendJSONString(buf, entry.Fields[i].Key)
		buf = append(buf, ':')
		buf = appendTypedFieldValue(buf, &entry.Fields[i])
	}

	buf = append(buf, '}', '\n')
	*bp = buf
	return bp
}

// Encode encodes an entry and returns the result as a string.
func (e *JSONEncoder) Encode(entry *Entry) string {
	bp := e.encodeToBuffer(entry)
	result := string(*bp)
	putBuf(bp)
	return result
}

// EncodeTo encodes an entry and writes bytes directly to the WriteSyncer.
// Avoids the []byte → string copy that Encode() requires.
func (e *JSONEncoder) EncodeTo(entry *Entry, ws WriteSyncer) (int, error) {
	bp := e.encodeToBuffer(entry)
	n, err := ws.Write(*bp)
	putBuf(bp)
	return n, err
}

// ConsoleEncoder encodes entries as human-readable console output.
type ConsoleEncoder struct {
	config   *Config
	timezone *time.Location
}

func newConsoleEncoder(config *Config) *ConsoleEncoder {
	return &ConsoleEncoder{
		config:   config,
		timezone: resolveTimezone(config.Timezone),
	}
}

// Encode encodes an entry as human-readable console output.
func (e *ConsoleEncoder) Encode(entry *Entry) string {
	bp := getBuf()
	buf := *bp

	entryTime := entry.Time.In(e.timezone)
	buf = entryTime.AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, '\t')

	levelStr := levelToUpper(entry.Level)
	if e.config.IsDevelopment {
		levelStr = colorizeLevel(levelStr, entry.Level)
	}
	buf = append(buf, levelStr...)

	if entry.CallerDefined {
		buf = append(buf, '\t')
		buf = append(buf, entry.CallerFile...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(entry.CallerLine), 10)
	}

	buf = append(buf, '\t')
	buf = append(buf, entry.Message...)

	if len(entry.Fields) > 0 {
		buf = append(buf, '\t')
		for i := range entry.Fields {
			if i > 0 {
				buf = append(buf, ' ')
			}
			buf = append(buf, entry.Fields[i].Key...)
			buf = append(buf, '=')
			buf = appendConsoleFieldValue(buf, &entry.Fields[i])
		}
	}

	buf = append(buf, '\n')

	if entry.Stacktrace != "" {
		buf = append(buf, entry.Stacktrace...)
		buf = append(buf, '\n')
	}

	*bp = buf
	result := string(*bp)
	putBuf(bp)
	return result
}

// EncodeTo writes console-encoded entry directly to WriteSyncer using pooled buffer.
func (e *ConsoleEncoder) EncodeTo(entry *Entry, ws WriteSyncer) (int, error) {
	bp := getBuf()
	buf := *bp

	entryTime := entry.Time.In(e.timezone)
	buf = entryTime.AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, '\t')

	levelStr := levelToUpper(entry.Level)
	if e.config.IsDevelopment {
		levelStr = colorizeLevel(levelStr, entry.Level)
	}
	buf = append(buf, levelStr...)

	if entry.CallerDefined {
		buf = append(buf, '\t')
		buf = append(buf, entry.CallerFile...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(entry.CallerLine), 10)
	}

	buf = append(buf, '\t')
	buf = append(buf, entry.Message...)

	if len(entry.Fields) > 0 {
		buf = append(buf, '\t')
		for i := range entry.Fields {
			if i > 0 {
				buf = append(buf, ' ')
			}
			buf = append(buf, entry.Fields[i].Key...)
			buf = append(buf, '=')
			buf = appendConsoleFieldValue(buf, &entry.Fields[i])
		}
	}

	buf = append(buf, '\n')

	if entry.Stacktrace != "" {
		buf = append(buf, entry.Stacktrace...)
		buf = append(buf, '\n')
	}

	*bp = buf
	n, err := ws.Write(*bp)
	putBuf(bp)
	return n, err
}

// appendConsoleFieldValue appends a field value for console output to a byte buffer.
func appendConsoleFieldValue(buf []byte, f *Field) []byte {
	switch f.Type {
	case FieldTypeString:
		return append(buf, f.Str...)
	case FieldTypeInt64:
		return strconv.AppendInt(buf, f.Integer, 10)
	case FieldTypeBool:
		return strconv.AppendBool(buf, f.Integer == 1)
	case FieldTypeFloat64:
		return strconv.AppendFloat(buf, math.Float64frombits(uint64(f.Integer)), 'f', -1, 64)
	default:
		return append(buf, fmt.Sprintf("%v", f.Iface)...)
	}
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

// --- Typed field encoding (zero-alloc for string/int64/bool/float64) ---

// appendTypedFieldValue encodes a typed Field's value to JSON without interface boxing.
func appendTypedFieldValue(buf []byte, f *Field) []byte {
	switch f.Type {
	case FieldTypeString:
		return appendJSONString(buf, f.Str)
	case FieldTypeInt64:
		return strconv.AppendInt(buf, f.Integer, 10)
	case FieldTypeBool:
		return strconv.AppendBool(buf, f.Integer == 1)
	case FieldTypeFloat64:
		return appendJSONFloat(buf, math.Float64frombits(uint64(f.Integer)))
	default:
		return appendJSONValue(buf, f.Iface)
	}
}



// appendJSONValue appends a JSON-encoded value for FieldTypeAny (fallback).
func appendJSONValue(buf []byte, v any) []byte {
	switch val := v.(type) {
	case string:
		return appendJSONString(buf, val)
	case int:
		return strconv.AppendInt(buf, int64(val), 10)
	case int8:
		return strconv.AppendInt(buf, int64(val), 10)
	case int16:
		return strconv.AppendInt(buf, int64(val), 10)
	case int32:
		return strconv.AppendInt(buf, int64(val), 10)
	case int64:
		return strconv.AppendInt(buf, int64(val), 10)
	case uint:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint8:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint16:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint32:
		return strconv.AppendUint(buf, uint64(val), 10)
	case uint64:
		return strconv.AppendUint(buf, uint64(val), 10)
	case float32:
		return appendJSONFloat(buf, float64(val))
	case float64:
		return appendJSONFloat(buf, val)
	case bool:
		return strconv.AppendBool(buf, val)
	case nil:
		return append(buf, "null"...)
	case error:
		return appendJSONString(buf, val.Error())
	default:
		jsonBytes, err := jsonLib.Marshal(val)
		if err != nil {
			return appendJSONString(buf, fmt.Sprintf("%v", val))
		}
		return append(buf, jsonBytes...)
	}
}

// appendJSONFloat appends a JSON float.
func appendJSONFloat(buf []byte, f float64) []byte {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return appendJSONString(buf, strconv.FormatFloat(f, 'f', -1, 64))
	}
	return strconv.AppendFloat(buf, f, 'f', -1, 64)
}

// appendJSONString appends a JSON-escaped string to buf (zero-alloc, RFC 8259).
func appendJSONString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	start := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b >= utf8.RuneSelf {
			_, size := utf8.DecodeRuneInString(s[i:])
			i += size
			continue
		}
		if b >= 0x20 && b != '"' && b != '\\' {
			i++
			continue
		}
		if start < i {
			buf = append(buf, s[start:i]...)
		}
		switch b {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		case '\b':
			buf = append(buf, '\\', 'b')
		case '\f':
			buf = append(buf, '\\', 'f')
		default:
			buf = append(buf, '\\', 'u', '0', '0')
			buf = append(buf, hexChars[b>>4], hexChars[b&0x0f])
		}
		i++
		start = i
	}
	if start < len(s) {
		buf = append(buf, s[start:]...)
	}
	buf = append(buf, '"')
	return buf
}

const hexChars = "0123456789abcdef"
