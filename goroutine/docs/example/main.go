// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"time"

	routine "github.com/anthanhphan/gosdk/goroutine"
	"github.com/anthanhphan/gosdk/logger"
)

func main() {
	// Example 1: Initialize logger with custom configuration (disable stacktrace)
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelInfo,
		LogEncoding:       logger.EncodingJSON,
		DisableCaller:     false,
		DisableStacktrace: true,
		IsDevelopment:     false,
	})
	defer undo()

	// Example 2: Run a simple function in a goroutine with string argument
	routine.Run(func(msg string) {
		logger.Info("Message: ", msg)
	}, "Hello, world!")

	// Example 3: Run a function with multiple arguments
	routine.Run(func(a, b int) {
		logger.Infof("Sum: %d + %d = %d", a, b, a+b)
	}, 10, 20)

	// Example 4: Run a function with no arguments
	routine.Run(func() {
		logger.Info("Running in background goroutine")
	})

	// Example 5: Panic recovery demonstration
	// The panic will be automatically recovered and logged with accurate location
	routine.Run(func() {
		logger.Info("About to panic...")
		panic("This panic will be recovered and logged with location")
	})

	// Example 6: Type conversion demonstration
	// int is automatically converted to int32
	routine.Run(func(n int32) {
		logger.Infof("Received int32: %d", n)
	}, 42)

	// Wait for goroutines to complete
	time.Sleep(200 * time.Millisecond)
	logger.Info("All examples completed")
}
