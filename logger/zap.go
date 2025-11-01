// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"log"

	"sync"

	"go.uber.org/zap"
)

var (
	zapLoggerInstance *zap.Logger
	once              sync.Once
)

// InitLogger initializes the logger with custom configuration and optional default log fields.
//
// Input:
//   - config: Logger configuration containing log level, encoding, output paths, etc.
//   - defaultLogFields: Optional zap.Field parameters to add default context to all log messages
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
//	    zap.String("app_name", "my-app"),
//	    zap.String("app_version", "1.0.0"),
//	    zap.String("environment", "production"),
//	)
//	defer undo()
func InitLogger(config *Config, defaultLogFields ...zap.Field) func() {
	var undo func()

	once.Do(func() {
		if err := config.Validate(); err != nil {
			log.Fatalf("invalid logger config: %v", err)
		}

		zapConfig := buildZapConfig(config)
		logger, err := zapConfig.Build()
		if err != nil {
			log.Fatalf("failed to build zap logger: %v", err)
		}

		// Add default log fields if provided
		if len(defaultLogFields) > 0 {
			logger = logger.With(defaultLogFields...)
		}

		zapLoggerInstance = logger
		undo = zap.ReplaceGlobals(logger)
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

// NewLoggerWithFields creates a sugared logger with additional structured fields from the global logger.
//
// Input:
//   - fields: Optional zap.Field parameters to add structured context to log messages
//
// Output:
//   - *zap.SugaredLogger: A sugared logger instance with the specified fields
//
// Example:
//
//	logger := NewLoggerWithFields(
//	    zap.String("service", "user-service"),
//	    zap.String("operation", "create_user"),
//	)
//	logger.Info("User created successfully", "user_id", 123)
func NewLoggerWithFields(fields ...zap.Field) *zap.SugaredLogger {
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	return zapLoggerInstance.With(fields...).Sugar()
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Debug(args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Debugf(template, args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Debugw(msg, keysAndValues...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Info(args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Infof(template, args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Infow(msg, keysAndValues...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Warn(args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Warnf(template, args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Warnw(msg, keysAndValues...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Error(args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Errorf(template, args...)
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
	if zapLoggerInstance == nil {
		InitDefaultLogger()
	}
	zapLoggerInstance.WithOptions(zap.AddCallerSkip(1)).Sugar().Errorw(msg, keysAndValues...)
}
