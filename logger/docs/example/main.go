// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"github.com/anthanhphan/gosdk/logger"
	"go.uber.org/zap"
)

func main() {
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelDebug,
		LogEncoding:       logger.EncodingJSON,
		DisableCaller:     false,
		DisableStacktrace: true,
		IsDevelopment:     true,
	})
	defer undo()

	log := logger.NewLoggerWithFields()
	log.Info("Application started")
	log.Debug("Debug information")
	log.Warn("Warning message")
	log.Error("Error occurred")

	serviceLog := logger.NewLoggerWithFields(
		zap.String("service", "user-service"),
		zap.String("version", "1.0.0"),
	)
	serviceLog.Infof("User created, user_id: %d, email: %s", 12345, "user@example.com")
}
