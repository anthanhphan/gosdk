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

// InitDevelopmentLogger initializes the logger with development configuration.
//
// Input:
//   - None
//
// Output:
//   - func(): Cleanup function to restore global logger state
//
// Example:
//
//	undo := InitDevelopmentLogger()
//	defer undo()
func InitDevelopmentLogger() func() {
	return InitLogger(&DevelopmentConfig)
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
		InitDevelopmentLogger()
	}
	return zapLoggerInstance.With(fields...).Sugar()
}
