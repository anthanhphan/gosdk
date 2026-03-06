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

// stackBufPool reuses buffers for stack trace operations to reduce allocations.
var stackBufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 4096)
		return &buf
	},
}

// fieldSlicePool pools []Field slices to avoid allocation in parseKeysAndValues.
var fieldSlicePool = sync.Pool{
	New: func() any {
		s := make([]Field, 0, 8)
		return &s
	},
}

// callerCache caches runtime.Caller file paths → short paths to avoid
// repeated filepath.Rel / GetShortPath calls on every log entry.
var callerCache sync.Map // map[string]string

// Logger is the core logger implementation.
type Logger struct {
	config          *Config
	processedFields []Field // pre-processed default fields stored as ordered slice
	outputs         []WriteSyncer
	closers         []io.Closer
	mu              sync.RWMutex
	callerSkip      int
	encoder         Encoder
}

const (
	baseCallerSkip       = 3
	asyncCallerSkipDelta = 1
	globalCallerSkip     = 2
)

var defaultLevelOrder = map[Level]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
}

// NewLogger creates a new logger instance with the provided configuration and output writers.
//
// Input:
//   - config: Logger configuration containing log level, encoding, output paths, etc.
//   - outputs: Slice of io.Writer destinations for log output (stdout, stderr, files, etc.)
//   - fields: Optional Field parameters to add default context to all log messages
//
// Output:
//   - *Logger: A new logger instance ready for use
//
// Example:
//
//	config := &Config{
//	    LogLevel:    LevelInfo,
//	    LogEncoding: EncodingJSON,
//	}
//	logger := NewLogger(config, []io.Writer{os.Stdout}, String("app", "my-app"))
//	logger.Info("Application started")
func NewLogger(config *Config, outputs []io.Writer, fields ...Field) *Logger {
	if len(outputs) == 0 {
		outputs = []io.Writer{os.Stdout}
	}

	processed := make([]Field, 0, len(fields))
	for _, field := range fields {
		processed = append(processed, processField(field, config.MaskKey))
	}

	// Wrap outputs as BufferedWriteSyncers for non-blocking I/O
	wsOutputs := make([]WriteSyncer, 0, len(outputs))
	for _, w := range outputs {
		ws := AddSync(w)
		bws := NewBufferedWriteSyncer(Lock(ws), 0, 0)
		wsOutputs = append(wsOutputs, bws)
	}

	// Initialize encoder
	var encoder Encoder
	if config.LogEncoding == EncodingJSON {
		encoder = newJSONEncoder(config)
	} else {
		encoder = newConsoleEncoder(config)
	}

	return &Logger{
		config:          config,
		processedFields: processed,
		outputs:         wsOutputs,
		encoder:         encoder,
	}
}

func (l *Logger) setClosers(closers []io.Closer) {
	if len(closers) == 0 || l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closers = append(l.closers, closers...)
}

// With creates a new logger instance with additional fields that will be included in all log messages.
// The new logger shares the same configuration and outputs as the parent logger.
//
// Input:
//   - fields: Field parameters to add as persistent context to all log messages
//
// Output:
//   - *Logger: A new logger instance with the combined fields
//
// Example:
//
//	baseLogger := NewLogger(config, outputs)
//	serviceLogger := baseLogger.With(String("service", "user-service"))
//	serviceLogger.Info("User created") // Will include "service": "user-service" in all logs
func (l *Logger) With(fields ...Field) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newProcessed := make([]Field, len(l.processedFields), len(l.processedFields)+len(fields))
	copy(newProcessed, l.processedFields)
	for _, field := range fields {
		newProcessed = append(newProcessed, processField(field, l.config.MaskKey))
	}

	return &Logger{
		config:          l.config,
		processedFields: newProcessed,
		outputs:         l.outputs,
		closers:         l.closers,
		callerSkip:      l.callerSkip,
		encoder:         l.encoder,
	}
}

// WithOptions creates a new logger instance with additional options applied.
// Options can modify logger behavior such as caller skip frames.
//
// Input:
//   - opts: Option functions to modify the logger behavior
//
// Output:
//   - *Logger: A new logger instance with the options applied
//
// Example:
//
//	logger := NewLogger(config, outputs)
//	loggerWithSkip := logger.WithOptions(AddCallerSkip(1))
//	loggerWithSkip.Info("Message") // Caller info will skip one additional frame
func (l *Logger) WithOptions(opts ...Option) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newProcessed := make([]Field, len(l.processedFields))
	copy(newProcessed, l.processedFields)

	newLogger := &Logger{
		config:          l.config,
		processedFields: newProcessed,
		outputs:         l.outputs,
		closers:         l.closers,
		callerSkip:      l.callerSkip,
		encoder:         l.encoder,
	}

	for _, opt := range opts {
		opt(newLogger)
	}

	return newLogger
}

// Debug logs a message at debug level.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	logger.Debug("Debug message")
//	logger.Debug("Processing user", "user_id", 12345, "action", "create")
func (l *Logger) Debug(args ...any) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelDebug, 1, msg, fields...)
}

// Debugf logs a formatted message at debug level using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Debugf("Processing user %s with id %d", "john", 12345)
func (l *Logger) Debugf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelDebug, 1, msg)
}

// Debugw logs a message with structured key-value pairs at debug level.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Debugw("Request received", "method", "GET", "path", "/api/users", "ip", "192.168.1.1")
func (l *Logger) Debugw(msg string, keysAndValues ...any) {
	fsp, n := l.parseKeysAndValues(keysAndValues...)
	if fsp != nil {
		l.log(LevelDebug, 1, msg, (*fsp)[:n]...)
		fieldSlicePool.Put(fsp)
	} else {
		l.log(LevelDebug, 1, msg)
	}
}

// Info logs a message at info level.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	logger.Info("Application started")
//	logger.Info("User created", "user_id", 12345, "email", "user@example.com")
func (l *Logger) Info(args ...any) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelInfo, 1, msg, fields...)
}

// Infof logs a formatted message at info level using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Infof("User %s logged in with id %d", "john", 12345)
func (l *Logger) Infof(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelInfo, 1, msg)
}

// Infow logs a message with structured key-value pairs at info level.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Infow("User created", "user_id", 12345, "email", "user@example.com")
func (l *Logger) Infow(msg string, keysAndValues ...any) {
	fsp, n := l.parseKeysAndValues(keysAndValues...)
	if fsp != nil {
		l.log(LevelInfo, 1, msg, (*fsp)[:n]...)
		fieldSlicePool.Put(fsp)
	} else {
		l.log(LevelInfo, 1, msg)
	}
}
// Warn logs a message at warning level.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	logger.Warn("Slow query detected")
//	logger.Warn("Connection slow", "duration_ms", 1500, "host", "database.example.com")
func (l *Logger) Warn(args ...any) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelWarn, 1, msg, fields...)
}

// Warnf logs a formatted message at warning level using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Warnf("Connection attempt %d of %d failed", attempt, maxAttempts)
func (l *Logger) Warnf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelWarn, 1, msg)
}

// Warnw logs a message with structured key-value pairs at warning level.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Warnw("Slow query detected", "query", "SELECT * FROM users", "duration_ms", 1500)
func (l *Logger) Warnw(msg string, keysAndValues ...any) {
	fsp, n := l.parseKeysAndValues(keysAndValues...)
	if fsp != nil {
		l.log(LevelWarn, 1, msg, (*fsp)[:n]...)
		fieldSlicePool.Put(fsp)
	} else {
		l.log(LevelWarn, 1, msg)
	}
}
// Error logs a message at error level.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	logger.Error("Operation failed")
//	logger.Error("Database error", "error", err.Error(), "operation", "fetch_user")
func (l *Logger) Error(args ...any) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelError, 1, msg, fields...)
}

// Errorf logs a formatted message at error level using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Errorf("Failed to connect to %s on port %d", "database", 5432)
func (l *Logger) Errorf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelError, 1, msg)
}

// Errorw logs a message with structured key-value pairs at error level.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic any)
//
// Output:
//   - None
//
// Example:
//
//	logger.Errorw("Database connection failed", "error", err.Error(), "host", "localhost", "port", 5432)
func (l *Logger) Errorw(msg string, keysAndValues ...any) {
	fsp, n := l.parseKeysAndValues(keysAndValues...)
	if fsp != nil {
		l.log(LevelError, 1, msg, (*fsp)[:n]...)
		fieldSlicePool.Put(fsp)
	} else {
		l.log(LevelError, 1, msg)
	}
}
// Fatal logs a message at error level and then exits the program with os.Exit(1).
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None (exits program)
//
// Example:
//
//	logger.Fatal("Critical error occurred")
//	logger.Fatal("Database connection failed", "error", err.Error())
func (l *Logger) Fatal(args ...any) {
	msg, fields := l.formatArgs(args...)
	l.log(LevelError, 1, msg, fields...)
	os.Exit(1)
}

// Fatalf logs a formatted message at error level and then exits the program with os.Exit(1).
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic any)
//
// Output:
//   - None (exits program)
//
// Example:
//
//	logger.Fatalf("Failed to start server on port %d: %v", 8080, err)
func (l *Logger) Fatalf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(LevelError, 1, msg)
	os.Exit(1)
}

// Fatalw logs a message with structured key-value pairs at error level and then exits the program with os.Exit(1).
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic any)
//
// Output:
//   - None (exits program)
//
// Example:
//
//	logger.Fatalw("Critical error", "error", err.Error(), "component", "database")
func (l *Logger) Fatalw(msg string, keysAndValues ...any) {
	fsp, n := l.parseKeysAndValues(keysAndValues...)
	if fsp != nil {
		l.log(LevelError, 1, msg, (*fsp)[:n]...)
		fieldSlicePool.Put(fsp)
	} else {
		l.log(LevelError, 1, msg)
	}
	os.Exit(1)
}

func (l *Logger) formatArgs(args ...any) (string, []Field) {
	if len(args) == 0 {
		return "", nil
	}

	if len(args) == 1 {
		return fmt.Sprint(args[0]), nil
	}

	if msg, ok := args[0].(string); ok {
		fsp, n := l.parseKeysAndValues(args[1:]...)
		if fsp != nil {
			// Copy fields out since we need to return them and put the pool back
			fields := make([]Field, n)
			copy(fields, (*fsp)[:n])
			fieldSlicePool.Put(fsp)
			return msg, fields
		}
		return msg, nil
	}

	return fmt.Sprint(args...), nil
}

func (*Logger) parseKeysAndValues(keysAndValues ...any) (*[]Field, int) {
	if len(keysAndValues) == 0 {
		return nil, 0
	}

	fsp := fieldSlicePool.Get().(*[]Field)
	fields := (*fsp)[:0]

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			// Fast path: key is already a string (most common case)
			var key string
			if s, ok := keysAndValues[i].(string); ok {
				key = s
			} else {
				key = fmt.Sprint(keysAndValues[i])
			}
			// Use Any() to detect typed fields and avoid boxing
			fields = append(fields, Any(key, keysAndValues[i+1]))
		} else {
			fields = append(fields, Any("extra", keysAndValues[i]))
		}
	}

	*fsp = fields
	return fsp, len(fields)
}

func (l *Logger) log(level Level, skipOffset int, msg string, fields ...Field) {
	entry := l.createEntry(level, skipOffset, msg, fields)
	if entry == nil {
		return
	}
	l.writeEntry(entry)
}

func (l *Logger) createEntry(level Level, skipOffset int, msg string, fields []Field) *Entry {
	if !l.shouldLog(level) {
		return nil
	}

	entry := getEntry()
	entry.Time = time.Now()
	entry.Level = level
	entry.Message = msg

	// Append pre-processed default fields (already processed at With/NewLogger time)
	l.mu.RLock()
	entry.Fields = append(entry.Fields, l.processedFields...)
	l.mu.RUnlock()

	// Append per-call fields (processField handles struct tag omit/mask)
	for i := range fields {
		entry.Fields = append(entry.Fields, processField(fields[i], l.config.MaskKey))
	}

	if !l.config.DisableCaller {
		l.setCallerInfo(entry, skipOffset)
	}

	if !l.config.DisableStacktrace && level == LevelError {
		entry.Stacktrace = l.getStacktrace()
	}

	return entry
}

func (l *Logger) shouldLog(level Level) bool {
	levelVal, ok := defaultLevelOrder[level]
	if !ok {
		return false
	}
	configLevelVal, ok := defaultLevelOrder[l.config.LogLevel]
	if !ok {
		return true
	}
	return levelVal >= configLevelVal
}

func (l *Logger) setCallerInfo(entry *Entry, skipOffset int) {
	skip := baseCallerSkip + l.callerSkip + skipOffset
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return
	}

	// Cache the short path to avoid filepath.Rel on every log
	if cached, found := callerCache.Load(file); found {
		entry.CallerFile = cached.(string)
	} else {
		short := utils.GetShortPath(file)
		callerCache.Store(file, short)
		entry.CallerFile = short
	}
	entry.CallerLine = line
	entry.CallerDefined = true
}

func (*Logger) getStacktrace() string {
	bufPtr := stackBufPool.Get().(*[]byte)
	defer stackBufPool.Put(bufPtr)

	buf := *bufPtr
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func (l *Logger) writeEntry(entry *Entry) {
	l.mu.RLock()
	outputs := l.outputs
	l.mu.RUnlock()

	switch len(outputs) {
	case 0:
		// no outputs
	case 1:
		// Single output: zero-copy path
		_, _ = l.encoder.EncodeTo(entry, outputs[0])
	default:
		// Multiple outputs: encode once, write bytes to all
		encoded := l.encoder.Encode(entry)
		b := []byte(encoded)
		for _, ws := range outputs {
			_, _ = ws.Write(b)
		}
	}

	// Return entry to pool after all writes
	putEntry(entry)
}

// Sync flushes all buffered output to the underlying writers.
// This should be called when you need to ensure all logged data has been written,
// for example during graceful shutdown or after critical error logs.
func (l *Logger) Sync() {
	l.flushOutputs()
}

func (l *Logger) flushOutputs() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, ws := range l.outputs {
		_ = ws.Sync()
	}
}

// stopOutputs stops all BufferedWriteSyncers (final flush + stop goroutine).
// This should only be called during application shutdown.
func (l *Logger) stopOutputs() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, ws := range l.outputs {
		if bws, ok := ws.(*BufferedWriteSyncer); ok {
			_ = bws.Stop()
		} else {
			_ = ws.Sync()
		}
	}
}

func (l *Logger) closeOutputs() {
	// Stop buffered writers first
	l.stopOutputs()

	l.mu.Lock()
	closers := l.closers
	l.closers = nil
	l.mu.Unlock()

	for _, closer := range closers {
		_ = closer.Close()
	}
}

// Option is a function that modifies a logger.
type Option func(*Logger)

// AddCallerSkip creates an Option that adds skip frames to the caller information.
// This is useful when wrapping the logger to adjust the caller location.
//
// Input:
//   - skip: Number of additional frames to skip when determining caller location
//
// Output:
//   - Option: An option function that can be used with WithOptions
//
// Example:
//
//	logger := NewLogger(config, outputs)
//	loggerWithSkip := logger.WithOptions(AddCallerSkip(1))
func AddCallerSkip(skip int) Option {
	return func(l *Logger) {
		l.callerSkip += skip
	}
}
