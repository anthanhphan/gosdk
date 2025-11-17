// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"log"
	"sync"
)

var (
	loggerInstance      *Logger
	asyncLoggerInstance *AsyncLogger
	once                sync.Once
	asyncOnce           sync.Once
)

// InitLogger initializes the logger with custom configuration and optional default log fields.
//
// Input:
//   - config: Logger configuration containing log level, encoding, output paths, etc.
//   - defaultLogFields: Optional Field parameters to add default context to all log messages
//
// Output:
//   - func(): Cleanup function to restore global logger state
//
// Example:
//
//	config := &Config{
//	    LogLevel: LevelInfo,
//	    LogEncoding: EncodingJSON,
//	    IsDevelopment: false,
//	}
//	undo := InitLogger(config,
//	    String("app_name", "my-app"),
//	    String("app_version", "1.0.0"),
//	    String("environment", "production"),
//	)
//	defer undo()
func InitLogger(config *Config, defaultLogFields ...Field) func() {
	var undo func()

	once.Do(func() {
		if err := config.Validate(); err != nil {
			log.Fatalf("invalid logger config: %v", err)
		}

		loggerInstance = buildLoggerConfig(config, defaultLogFields...)
		undo = func() {
			loggerInstance = nil
			once = sync.Once{}
		}
	})

	return undo
}

// InitDefaultLogger initializes the logger with default configuration.
//
// Input:
//   - None
//
// Output:
//   - func(): Cleanup function to restore global logger state
//
// Example:
//
//	undo := InitDefaultLogger()
//	defer undo()
func InitDefaultLogger() func() {
	return InitLogger(&DefaultConfig)
}

// InitProductionLogger initializes the logger with production configuration.
//
// Input:
//   - None
//
// Output:
//   - func(): Cleanup function to restore global logger state
//
// Example:
//
//	undo := InitProductionLogger()
//	defer undo()
func InitProductionLogger() func() {
	return InitLogger(&ProductionConfig)
}

// InitAsyncLogger initializes an asynchronous logger with custom configuration and optional default log fields.
// Log entries are queued and written in a background goroutine, providing non-blocking logging.
//
// Input:
//   - config: Logger configuration containing log level, encoding, output paths, etc.
//   - defaultLogFields: Optional Field parameters to add default context to all log messages
//
// Output:
//   - func(): Cleanup function to flush remaining entries and restore global logger state
//
// Example:
//
//	config := &Config{
//	    LogLevel: LevelInfo,
//	    LogEncoding: EncodingJSON,
//	    IsDevelopment: false,
//	}
//	undo := InitAsyncLogger(config,
//	    String("app_name", "my-app"),
//	    String("app_version", "1.0.0"),
//	)
//	defer undo()
func InitAsyncLogger(config *Config, defaultLogFields ...Field) func() {
	var undo func()

	asyncOnce.Do(func() {
		if err := config.Validate(); err != nil {
			log.Fatalf("invalid logger config: %v", err)
		}

		baseLogger := buildLoggerConfig(config, defaultLogFields...)
		// Use default queue size (100) - NewAsyncLogger handles 0 as default
		asyncLoggerInstance = NewAsyncLogger(baseLogger, 0)
		undo = func() {
			if asyncLoggerInstance != nil {
				asyncLoggerInstance.Flush()
			}
			asyncLoggerInstance = nil
			asyncOnce = sync.Once{}
		}
	})

	return undo
}

// NewLoggerWithFields creates a logger with additional structured fields from the global logger.
//
// Input:
//   - fields: Optional Field parameters to add structured context to log messages
//
// Output:
//   - *Logger: A logger instance with the specified fields
//
// Example:
//
//	logger := NewLoggerWithFields(
//	    String("service", "user-service"),
//	    String("operation", "create_user"),
//	)
//	logger.Info("User created successfully", "user_id", 123)
func NewLoggerWithFields(fields ...Field) *Logger {
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	return loggerInstance.With(fields...)
}

// Debug logs a message at debug level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
//
// Input:
//   - args: Arguments to log (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	logger.Debug("Processing request")
//	logger.Debug("User", "john", "logged in")
func Debug(args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Debug(args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Debug(args...)
}

// Debugf logs a formatted message at debug level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Debugf("User %s logged in with id %d", "john", 12345)
func Debugf(template string, args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Debugf(template, args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Debugf(template, args...)
}

// Debugw logs a message with structured key-value pairs at debug level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Debugw("Request received", "method", "GET", "path", "/api/users")
func Debugw(msg string, keysAndValues ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Debugw(msg, keysAndValues...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Debugw(msg, keysAndValues...)
}

// Info logs a message at info level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
//
// Input:
//   - args: Arguments to log (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	logger.Info("Application started")
//	logger.Info("Service ready on port", 8080)
func Info(args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Info(args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Info(args...)
}

// Infof logs a formatted message at info level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Infof("User %s created with id %d", "jane", 456)
func Infof(template string, args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Infof(template, args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Infof(template, args...)
}

// Infow logs a message with structured key-value pairs at info level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Infow("User created", "user_id", 12345, "email", "user@example.com")
func Infow(msg string, keysAndValues ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Infow(msg, keysAndValues...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Infow(msg, keysAndValues...)
}

// Warn logs a message at warning level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
//
// Input:
//   - args: Arguments to log (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	logger.Warn("Connection timeout")
//	logger.Warn("Retry attempt", 3, "of", 5)
func Warn(args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Warn(args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Warn(args...)
}

// Warnf logs a formatted message at warning level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Warnf("Connection to %s failed after %d retries", "localhost", 3)
func Warnf(template string, args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Warnf(template, args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Warnf(template, args...)
}

// Warnw logs a message with structured key-value pairs at warning level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Warnw("Slow query detected", "query", "SELECT * FROM users", "duration_ms", 1500)
func Warnw(msg string, keysAndValues ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Warnw(msg, keysAndValues...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Warnw(msg, keysAndValues...)
}

// Error logs a message at error level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
//
// Input:
//   - args: Arguments to log (variadic interface{})
//
// Output:
//   - None
//
// Example:
//
//	logger.Error("Database connection failed")
//	logger.Error("Failed to process request:", err)
func Error(args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Error(args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Error(args...)
}

// Errorf logs a formatted message at error level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Errorf("Failed to connect to %s on port %d", "database", 5432)
func Errorf(template string, args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Errorf(template, args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Errorf(template, args...)
}

// Errorw logs a message with structured key-value pairs at error level using the global logger.
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Errorw("Request failed", "error", err.Error(), "retry", 3)
func Errorw(msg string, keysAndValues ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Errorw(msg, keysAndValues...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Errorw(msg, keysAndValues...)
}

// Fatal logs a message at error level using the global logger and then exits the program with os.Exit(1).
// Automatically initializes with default configuration if logger is not initialized.
//
// Input:
//   - args: Arguments to log (variadic interface{})
//
// Output:
//   - None (exits program)
//
// Example:
//
//	logger.Fatal("Critical error occurred")
//	logger.Fatal("Database connection failed:", err)
func Fatal(args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Fatal(args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Fatal(args...)
}

// Fatalf logs a formatted message at error level using the global logger and then exits the program with os.Exit(1).
// Automatically initializes with default configuration if logger is not initialized.
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
//	logger.Fatalf("Failed to start server on port %d: %v", 8080, err)
func Fatalf(template string, args ...interface{}) {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Fatalf(template, args...)
		return
	}
	if loggerInstance == nil {
		InitDefaultLogger()
	}
	loggerInstance.WithOptions(AddCallerSkip(1)).Fatalf(template, args...)
}

// Flush waits for all queued log entries to be written (for async logger).
// If async logger is not initialized, this is a no-op.
//
// Input:
//   - None
//
// Output:
//   - None
//
// Example:
//
//	logger.Flush() // Ensure all logs are written before exit
func Flush() {
	if asyncLoggerInstance != nil {
		asyncLoggerInstance.Flush()
	}
}
