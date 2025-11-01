// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"errors"

	"github.com/anthanhphan/gosdk/logger"
	"go.uber.org/zap"
)

func main() {
	// Example 1: Quick start with default configuration
	undo := logger.InitDefaultLogger()
	defer undo()

	// Example 2: Basic convenience logging functions
	logger.Info("Application started")
	logger.Debug("Debug information")
	logger.Warn("Warning message")
	logger.Error("Error occurred")

	// Example 3: Formatted logging (Printf-style)
	username := "john"
	userID := 12345
	logger.Infof("User %s logged in with id %d", username, userID)
	logger.Debugf("Processing request for user %s", username)

	// Example 4: Structured logging with key-value pairs
	logger.Infow("User created",
		"user_id", 12345,
		"username", "john",
		"email", "john@example.com",
	)

	logger.Debugw("Request received",
		"method", "GET",
		"path", "/api/users",
		"ip", "192.168.1.1",
	)

	// Example 5: Error handling with structured logging
	err := errors.New("connection timeout")
	logger.Errorw("Request failed",
		"error", err.Error(),
		"operation", "user_creation",
		"retry", 3,
	)

	// Example 6: Component-specific logger with persistent fields
	serviceLog := logger.NewLoggerWithFields(
		zap.String("service", "user-service"),
		zap.String("version", "1.0.0"),
	)
	serviceLog.Infow("Service initialized", "port", 8080)

	// Example 7: Multiple component loggers
	authLog := logger.NewLoggerWithFields(
		zap.String("component", "auth"),
	)
	authLog.Infow("User authenticated", "user_id", userID)

	dbLog := logger.NewLoggerWithFields(
		zap.String("component", "database"),
	)
	dbLog.Infow("Connection established", "pool_size", 10)

	logger.Info("Application shutdown complete")
}
