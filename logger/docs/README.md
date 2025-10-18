# Logger Package

A high-performance, structured logging package for Go applications built on [Uber's Zap](https://github.com/uber-go/zap). Provides fast, zero-allocation logging with JSON/Console output formats, configurable log levels, and easy structured data integration.

## Installation

```bash
go get github.com/anthanhphan/gosdk/logger
```

## Quick Start

```go
package main

import (
	"github.com/anthanhphan/gosdk/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelDebug,
		LogEncoding:       logger.EncodingJSON,
		DisableCaller:     false,
		DisableStacktrace: true,
		IsDevelopment:     true,
	})
	defer undo()

	// Basic logging
	log := logger.NewLoggerWithFields()
	log.Info("Application started")
	log.Debug("Debug information")
	log.Warn("Warning message")
	log.Error("Error occurred")

	// Structured logging with fields
	serviceLog := logger.NewLoggerWithFields(
		zap.String("service", "user-service"),
		zap.String("version", "1.0.0"),
	)
	serviceLog.Info("User created, user_id: %d, email: %s", 12345, "user@example.com")
}
```

## Configuration

### Config Structure

```go
type Config struct {
	LogLevel          Level     // Log level: debug, info, warn, error
	LogEncoding       Encoding  // Output format: json, console
	LogOutputPaths    []string  // Output destinations (empty = console)
	DisableCaller     bool      // Hide caller information
	DisableStacktrace bool      // Hide stack traces
	IsDevelopment     bool      // Development mode settings
}
```

### Log Levels

- **`LevelDebug`** - Debug messages (most verbose)
- **`LevelInfo`** - Informational messages
- **`LevelWarn`** - Warning messages
- **`LevelError`** - Error messages (least verbose)

### Output Formats

- **`EncodingJSON`** - Structured JSON output
- **`EncodingConsole`** - Human-readable console output

### Output Destinations

- **Empty array `[]string{}`** - Log to console
- **File paths `[]string{"/var/log/app.log"}`** - Log to files
- **Multiple paths** - Log to multiple destinations

### Development Settings

When `IsDevelopment: true`:

- More readable output format
- Includes caller information
- Optimized for development debugging

## Usage Examples

### Development Setup

```go
// Quick development setup
undo := logger.InitDevelopmentLogger()
defer undo()

log := logger.NewLoggerWithFields()
log.Debug("Debug message")
log.Info("App started")
```

### Production Setup

```go
// Production-ready setup
undo := logger.InitProductionLogger()
defer undo()

log := logger.NewLoggerWithFields(
	zap.String("app", "my-service"),
	zap.String("version", "1.0.0"),
)
log.Info("Service started", "port", 8080)
```

### File Output

```go
config := &logger.Config{
	LogLevel:       logger.LevelInfo,
	LogEncoding:    logger.EncodingJSON,
	LogOutputPaths: []string{"/var/log/app.log", "/var/log/error.log"},
}
undo := logger.InitLogger(config)
defer undo()
```

### Component Logging

```go
// Authentication component
authLog := logger.NewLoggerWithFields(
	zap.String("component", "auth"),
	zap.String("version", "2.1.0"),
)
authLog.Info("User logged in", "user_id", "12345")

// Database component
dbLog := logger.NewLoggerWithFields(
	zap.String("component", "database"),
	zap.String("host", "localhost"),
)
dbLog.Info("Connection established", "pool_size", 10)
```

### Error Handling

```go
log := logger.NewLoggerWithFields(
	zap.String("request_id", "req-123"),
	zap.String("user_id", "user-456"),
)

if err != nil {
	log.Error("Failed to process request",
		"error", err.Error(),
		"operation", "user_creation",
	)
}
```

### Custom Setup

```go
config := &logger.Config{
	LogLevel:          logger.LevelWarn,
	LogEncoding:       logger.EncodingConsole,
	LogOutputPaths:    []string{},
	DisableCaller:     false,
	DisableStacktrace: true,
	IsDevelopment:     false,
}
undo := logger.InitLogger(config)
defer undo()
```

## Output Formats

### JSON Output

```json
{
  "level": "info",
  "ts": "2025-10-12T20:15:29.807+0700",
  "caller": "main.go:25",
  "msg": "User created, user_id: 12345, email: user@example.com",
  "service": "user-service",
  "version": "1.0.0"
}
```

### Console Output

```
2025-10-12T20:15:29.807+0700	INFO	main.go:25	User created, user_id: 12345, email: user@example.com	{"service": "user-service", "version": "1.0.0"}
```
