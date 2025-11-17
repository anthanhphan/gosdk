// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/utils"
)

// Logger is the core logger implementation.
type Logger struct {
	config         *Config
	fields         map[string]interface{}
	outputs        []io.Writer
	mu             sync.RWMutex
	callerSkip     int
	levelOrder     map[Level]int
	jsonEncoder    *JSONEncoder
	consoleEncoder *ConsoleEncoder
	encoderMu      sync.RWMutex
}

// NewLogger creates a new logger instance.
func NewLogger(config *Config, outputs []io.Writer, fields ...Field) *Logger {
	if len(outputs) == 0 {
		outputs = []io.Writer{os.Stdout}
	}

	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}

	levelOrder := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	return &Logger{
		config:     config,
		fields:     fieldMap,
		outputs:    outputs,
		levelOrder: levelOrder,
	}
}

// With creates a new logger with additional fields.
func (l *Logger) With(fields ...Field) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for _, field := range fields {
		newFields[field.Key] = field.Value
	}

	return &Logger{
		config:         l.config,
		fields:         newFields,
		outputs:        l.outputs,
		callerSkip:     l.callerSkip,
		levelOrder:     l.levelOrder,
		jsonEncoder:    l.jsonEncoder,
		consoleEncoder: l.consoleEncoder,
	}
}

// WithOptions creates a new logger with additional options.
func (l *Logger) WithOptions(opts ...Option) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &Logger{
		config:         l.config,
		fields:         make(map[string]interface{}),
		outputs:        l.outputs,
		callerSkip:     l.callerSkip,
		levelOrder:     l.levelOrder,
		jsonEncoder:    l.jsonEncoder,
		consoleEncoder: l.consoleEncoder,
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	for _, opt := range opts {
		opt(newLogger)
	}

	return newLogger
}

// Debug logs a message at debug level.
func (l *Logger) Debug(args ...interface{}) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelDebug, msg, fields...)
}

// Debugf logs a formatted message at debug level.
func (l *Logger) Debugf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelDebug, msg)
}

// Debugw logs a message with structured key-value pairs at debug level.
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	fields := l.parseKeysAndValues(keysAndValues...)
	l.log(LevelDebug, msg, fields...)
}

// Info logs a message at info level.
func (l *Logger) Info(args ...interface{}) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelInfo, msg, fields...)
}

// Infof logs a formatted message at info level.
func (l *Logger) Infof(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelInfo, msg)
}

// Infow logs a message with structured key-value pairs at info level.
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	fields := l.parseKeysAndValues(keysAndValues...)
	l.log(LevelInfo, msg, fields...)
}

// Warn logs a message at warning level.
func (l *Logger) Warn(args ...interface{}) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelWarn, msg, fields...)
}

// Warnf logs a formatted message at warning level.
func (l *Logger) Warnf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelWarn, msg)
}

// Warnw logs a message with structured key-value pairs at warning level.
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	fields := l.parseKeysAndValues(keysAndValues...)
	l.log(LevelWarn, msg, fields...)
}

// Error logs a message at error level.
func (l *Logger) Error(args ...interface{}) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelError, msg, fields...)
}

// Errorf logs a formatted message at error level.
func (l *Logger) Errorf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelError, msg)
}

// Errorw logs a message with structured key-value pairs at error level.
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	fields := l.parseKeysAndValues(keysAndValues...)
	l.log(LevelError, msg, fields...)
}

// Fatal logs a message at error level and then exits the program with os.Exit(1).
func (l *Logger) Fatal(args ...interface{}) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelError, msg, fields...)
	os.Exit(1)
}

// Fatalf logs a formatted message at error level and then exits the program with os.Exit(1).
func (l *Logger) Fatalf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelError, msg)
	os.Exit(1)
}

// formatArgs formats variadic arguments into a message and fields.
func (l *Logger) formatArgs(args ...interface{}) (string, []Field) {
	if len(args) == 0 {
		return "", nil
	}

	// If first arg is a string, use it as message
	if len(args) == 1 {
		return fmt.Sprint(args[0]), nil
	}

	// If first arg is a string and there are more args, treat as message + key-value pairs
	if msg, ok := args[0].(string); ok {
		fields := l.parseKeysAndValues(args[1:]...)
		return msg, fields
	}

	// Otherwise, format all args as message
	return fmt.Sprint(args...), nil
}

// parseKeysAndValues parses alternating keys and values into fields.
func (l *Logger) parseKeysAndValues(keysAndValues ...interface{}) []Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	estimatedCapacity := (len(keysAndValues) + 1) / 2
	fields := make([]Field, 0, estimatedCapacity)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprint(keysAndValues[i])
			value := keysAndValues[i+1]
			fields = append(fields, Any(key, value))
		} else {
			fields = append(fields, Any("extra", keysAndValues[i]))
		}
	}

	return fields
}

// log creates a log entry and writes it.
func (l *Logger) log(level Level, msg string, fields ...Field) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.RLock()
	defaultFieldsCount := len(l.fields)
	l.mu.RUnlock()

	fieldsCapacity := defaultFieldsCount + len(fields)
	entry := &Entry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  make(map[string]interface{}, fieldsCapacity),
	}

	l.mu.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	l.mu.RUnlock()

	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	if !l.config.DisableCaller {
		entry.Caller = l.getCaller()
	}

	if !l.config.DisableStacktrace && level == LevelError {
		entry.Stacktrace = l.getStacktrace()
	}

	l.writeEntry(entry)
}

// shouldLog checks if the log level should be logged.
func (l *Logger) shouldLog(level Level) bool {
	levelVal, ok := l.levelOrder[level]
	if !ok {
		return false
	}
	configLevelVal, ok := l.levelOrder[l.config.LogLevel]
	if !ok {
		return true
	}
	return levelVal >= configLevelVal
}

// getCaller retrieves caller information.
func (l *Logger) getCaller() *CallerInfo {
	skip := 3 + l.callerSkip
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	shortPath := utils.GetShortPath(file)

	return &CallerInfo{
		File: shortPath,
		Line: line,
	}
}

// getStacktrace retrieves the stack trace.
func (l *Logger) getStacktrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// writeEntry writes the log entry to all outputs.
func (l *Logger) writeEntry(entry *Entry) {
	var encoder Encoder
	if l.config.LogEncoding == EncodingJSON {
		if l.jsonEncoder == nil {
			l.encoderMu.Lock()
			if l.jsonEncoder == nil {
				l.jsonEncoder = NewJSONEncoder(l.config)
			}
			l.encoderMu.Unlock()
		}
		encoder = l.jsonEncoder
	} else {
		if l.consoleEncoder == nil {
			l.encoderMu.Lock()
			if l.consoleEncoder == nil {
				l.consoleEncoder = NewConsoleEncoder(l.config)
			}
			l.encoderMu.Unlock()
		}
		encoder = l.consoleEncoder
	}

	output := encoder.Encode(entry)
	if output == "" {
		return
	}

	l.mu.RLock()
	outputs := l.outputs
	l.mu.RUnlock()

	for _, w := range outputs {
		fmt.Fprint(w, output)
		if file, ok := w.(*os.File); ok && file != os.Stdout && file != os.Stderr {
			_ = file.Sync()
		}
	}
}

// Option is a function that modifies a logger.
type Option func(*Logger)

// AddCallerSkip adds skip frames to the caller.
func AddCallerSkip(skip int) Option {
	return func(l *Logger) {
		l.callerSkip += skip
	}
}
