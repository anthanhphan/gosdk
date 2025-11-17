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

	// Start background worker
	al.wg.Add(1)
	go al.worker()

	return al
}

// worker processes log entries from the queue in the background.
func (al *AsyncLogger) worker() {
	defer al.wg.Done()

	for {
		select {
		case entry := <-al.queue:
			if entry == nil {
				// Nil entry signals shutdown
				return
			}
			al.logger.writeEntry(entry)
		case <-al.ctx.Done():
			// Process remaining entries before shutdown
			al.drainQueue()
			close(al.stopped)
			return
		}
	}
}

// drainQueue processes all remaining entries in the queue.
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

// log queues a log entry for asynchronous processing.
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

	// Add caller info if enabled
	if !al.logger.config.DisableCaller {
		// Need to adjust caller skip for async logger (one more level)
		entry.Caller = al.getCaller()
	}

	// Add stacktrace if enabled and level is error
	if !al.logger.config.DisableStacktrace && level == LevelError {
		entry.Stacktrace = al.logger.getStacktrace()
	}

	// Try to queue the entry (non-blocking)
	select {
	case al.queue <- entry:
		// Successfully queued
	case <-al.ctx.Done():
		// Logger is shutting down, write synchronously
		al.logger.writeEntry(entry)
	default:
		// Queue is full, write synchronously to avoid blocking
		al.logger.writeEntry(entry)
	}
}

// Debug logs a message at debug level asynchronously.
func (al *AsyncLogger) Debug(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelDebug, msg, fields...)
}

// Debugf logs a formatted message at debug level asynchronously.
func (al *AsyncLogger) Debugf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelDebug, msg)
}

// Debugw logs a message with structured key-value pairs at debug level asynchronously.
func (al *AsyncLogger) Debugw(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelDebug, msg, fields...)
}

// Info logs a message at info level asynchronously.
func (al *AsyncLogger) Info(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelInfo, msg, fields...)
}

// Infof logs a formatted message at info level asynchronously.
func (al *AsyncLogger) Infof(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelInfo, msg)
}

// Infow logs a message with structured key-value pairs at info level asynchronously.
func (al *AsyncLogger) Infow(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelInfo, msg, fields...)
}

// Warn logs a message at warning level asynchronously.
func (al *AsyncLogger) Warn(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelWarn, msg, fields...)
}

// Warnf logs a formatted message at warning level asynchronously.
func (al *AsyncLogger) Warnf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelWarn, msg)
}

// Warnw logs a message with structured key-value pairs at warning level asynchronously.
func (al *AsyncLogger) Warnw(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelWarn, msg, fields...)
}

// Error logs a message at error level asynchronously.
func (al *AsyncLogger) Error(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.log(LevelError, msg, fields...)
}

// Errorf logs a formatted message at error level asynchronously.
func (al *AsyncLogger) Errorf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.log(LevelError, msg)
}

// Errorw logs a message with structured key-value pairs at error level asynchronously.
func (al *AsyncLogger) Errorw(msg string, keysAndValues ...interface{}) {
	fields := al.logger.parseKeysAndValues(keysAndValues...)
	al.log(LevelError, msg, fields...)
}

// Fatal logs a message at error level and then exits the program with os.Exit(1).
// This is synchronous to ensure the message is logged before exit.
func (al *AsyncLogger) Fatal(args ...interface{}) {
	msg, fields := al.logger.formatArgs(args...)
	al.logger.log(LevelError, msg, fields...)
	al.Flush()
	os.Exit(1)
}

// Fatalf logs a formatted message at error level and then exits the program with os.Exit(1).
// This is synchronous to ensure the message is logged before exit.
func (al *AsyncLogger) Fatalf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	al.logger.log(LevelError, msg)
	al.Flush()
	os.Exit(1)
}

// getCaller retrieves caller information for async logger.
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

// With creates a new async logger with additional fields.
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
func (al *AsyncLogger) Flush() {
	// Signal to drain queue (only if not already cancelled)
	select {
	case <-al.ctx.Done():
		// Already cancelled, just wait
	default:
		al.cancel()
	}
	// Wait for worker to finish
	al.wg.Wait()
	// Wait for stopped signal (non-blocking if already closed)
	select {
	case <-al.stopped:
	default:
	}
}

// Close stops the async logger and flushes all remaining entries.
func (al *AsyncLogger) Close() error {
	al.Flush()
	return nil
}
