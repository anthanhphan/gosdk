// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/utils"
)

// AsyncLogger wraps a Logger to provide asynchronous logging.
// Log entries are queued and written in a background goroutine.
type AsyncLogger struct {
	logger  *Logger
	queue   chan *Entry
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	stopped chan struct{}
}

// NewAsyncLogger creates a new async logger that wraps the given logger.
// Log entries are queued and written in a background goroutine for non-blocking logging.
//
// Input:
//   - logger: The base logger instance to wrap
//   - queueSize: Size of the log entry queue (if <= 0, defaults to 100)
//
// Output:
//   - *AsyncLogger: A new async logger instance
//
// Example:
//
//	baseLogger := NewLogger(config, outputs)
//	asyncLogger := NewAsyncLogger(baseLogger, 200)
//	asyncLogger.Info("Non-blocking log message")
func NewAsyncLogger(logger *Logger, queueSize int) *AsyncLogger {
	if queueSize <= 0 {
		queueSize = 100 // Default queue size
	}

	ctx, cancel := context.WithCancel(context.Background())
	al := &AsyncLogger{
		logger:  logger,
		queue:   make(chan *Entry, queueSize),
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}

	al.wg.Add(1)
	go al.worker()

	return al
}

func (al *AsyncLogger) worker() {
	defer al.wg.Done()

	for {
		select {
		case entry := <-al.queue:
			if entry == nil {
				return
			}
			al.logger.writeEntry(entry)
		case <-al.ctx.Done():
			al.drainQueue()
			close(al.stopped)
			return
		}
	}
}

func (al *AsyncLogger) drainQueue() {
	for {
		select {
		case entry := <-al.queue:
			if entry == nil {
				return
			}
			al.logger.writeEntry(entry)
		default:
			return
		}
	}
}

func (al *AsyncLogger) log(level Level, msg string, fields ...Field) {
	if !al.logger.shouldLog(level) {
		return
	}

	al.logger.mu.RLock()
	defaultFieldsCount := len(al.logger.fields)
	al.logger.mu.RUnlock()

	fieldsCapacity := defaultFieldsCount + len(fields)
	entry := &Entry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  make(map[string]interface{}, fieldsCapacity),
	}

	al.logger.mu.RLock()
	for k, v := range al.logger.fields {
		entry.Fields[k] = v
	}
	al.logger.mu.RUnlock()

	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	if !al.logger.config.DisableCaller {
		entry.Caller = al.getCaller()
	}

	if !al.logger.config.DisableStacktrace && level == LevelError {
		entry.Stacktrace = al.logger.getStacktrace()
	}

	select {
	case al.queue <- entry:
	case <-al.ctx.Done():
		al.logger.writeEntry(entry)
	default:
		al.logger.writeEntry(entry)
	}
}

// Debug logs a message at debug level asynchronously.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Debug("Debug message")
//	asyncLogger.Debug("Processing user", "user_id", 12345, "action", "create")
func (al *AsyncLogger) Debug(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelDebug, msg, fields...)
}

// Debugf logs a formatted message at debug level asynchronously using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Debugf("Processing user %s with id %d", "john", 12345)
func (al *AsyncLogger) Debugf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelDebug, msg)
}

// Debugw logs a message with structured key-value pairs at debug level asynchronously.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Debugw("Request received", "method", "GET", "path", "/api/users", "ip", "192.168.1.1")
func (al *AsyncLogger) Debugw(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelDebug, msg, fields...)
}

// Info logs a message at info level asynchronously.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Info("Application started")
//	asyncLogger.Info("User created", "user_id", 12345, "email", "user@example.com")
func (al *AsyncLogger) Info(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelInfo, msg, fields...)
}

// Infof logs a formatted message at info level asynchronously using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Infof("User %s logged in with id %d", "john", 12345)
func (al *AsyncLogger) Infof(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelInfo, msg)
}

// Infow logs a message with structured key-value pairs at info level asynchronously.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Infow("User created", "user_id", 12345, "email", "user@example.com")
func (al *AsyncLogger) Infow(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelInfo, msg, fields...)
}

// Warn logs a message at warning level asynchronously.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Warn("Slow query detected")
//	asyncLogger.Warn("Connection slow", "duration_ms", 1500, "host", "database.example.com")
func (al *AsyncLogger) Warn(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelWarn, msg, fields...)
}

// Warnf logs a formatted message at warning level asynchronously using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Warnf("Connection attempt %d of %d failed", attempt, maxAttempts)
func (al *AsyncLogger) Warnf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelWarn, msg)
}

// Warnw logs a message with structured key-value pairs at warning level asynchronously.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Warnw("Slow query detected", "query", "SELECT * FROM users", "duration_ms", 1500)
func (al *AsyncLogger) Warnw(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelWarn, msg, fields...)
}

// Error logs a message at error level asynchronously.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Error("Operation failed")
//	asyncLogger.Error("Database error", "error", err.Error(), "operation", "fetch_user")
func (al *AsyncLogger) Error(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelError, msg, fields...)
}

// Errorf logs a formatted message at error level asynchronously using Printf-style formatting.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Errorf("Failed to connect to %s on port %d", "database", 5432)
func (al *AsyncLogger) Errorf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelError, msg)
}

// Errorw logs a message with structured key-value pairs at error level asynchronously.
//
// Input:
//   - msg: Log message
//   - keysAndValues: Alternating keys and values for structured logging (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	asyncLogger.Errorw("Database connection failed", "error", err.Error(), "host", "localhost", "port", 5432)
func (al *AsyncLogger) Errorw(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelError, msg, fields...)
}

// Fatal logs a message at error level and then exits the program with os.Exit(1).
// This is synchronous to ensure the message is logged before exit.
//
// Input:
//   - args: Variadic arguments to log (can be message string or key-value pairs)
//
// Output:
//   - None (exits program)
//
// Example:
//
//	asyncLogger.Fatal("Critical error occurred")
//	asyncLogger.Fatal("Database connection failed", "error", err.Error())
func (al *AsyncLogger) Fatal(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.logger.log(LevelError, msg, fields...)
	al.Flush()
	os.Exit(1)
}

// Fatalf logs a formatted message at error level and then exits the program with os.Exit(1).
// This is synchronous to ensure the message is logged before exit.
//
// Input:
//   - template: Format string (Printf-style)
//   - args: Arguments for the format string (variadic interface{})
//
// Output:
//   - None (exits program)
//
// Example:
//
//	asyncLogger.Fatalf("Failed to start server on port %d: %v", 8080, err)
func (al *AsyncLogger) Fatalf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.logger.log(LevelError, msg)
	al.Flush()
	os.Exit(1)
}

func (al *AsyncLogger) getCaller() *CallerInfo {
	skip := 4 + al.logger.callerSkip
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

// With creates a new async logger instance with additional fields that will be included in all log messages.
//
// Input:
//   - fields: Field parameters to add as persistent context to all log messages
//
// Output:
//   - *AsyncLogger: A new async logger instance with the combined fields
//
// Example:
//
//	baseAsyncLogger := NewAsyncLogger(logger, 100)
//	serviceAsyncLogger := baseAsyncLogger.With(String("service", "user-service"))
//	serviceAsyncLogger.Info("User created") // Will include "service": "user-service" in all logs
func (al *AsyncLogger) With(fields ...Field) *AsyncLogger {
	return &AsyncLogger{
		logger:  al.logger.With(fields...),
		queue:   al.queue,
		ctx:     al.ctx,
		cancel:  al.cancel,
		stopped: al.stopped,
	}
}

// Flush waits for all queued log entries to be written.
// This should be called before program exit to ensure all logs are persisted.
//
// Input:
//   - None
//
// Output:
//   - None
//
// Example:
//
//	defer asyncLogger.Flush() // Ensure all logs are written before exit
func (al *AsyncLogger) Flush() {
	select {
	case <-al.ctx.Done():
	default:
		al.cancel()
	}
	al.wg.Wait()
	select {
	case <-al.stopped:
	default:
	}
}

// Close stops the async logger and flushes all remaining entries.
// This is equivalent to calling Flush() and should be called before program exit.
//
// Input:
//   - None
//
// Output:
//   - error: Always returns nil (included for io.Closer interface compatibility)
//
// Example:
//
//	defer asyncLogger.Close() // Ensure all logs are written before exit
func (al *AsyncLogger) Close() error {
	al.Flush()
	return nil
}
