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
	// Initialize logger with default config
	undo := logger.InitDefaultLogger()
	defer undo()

	// Basic logging using convenience functions
	logger.Info("Application started")
	logger.Debug("Debug information")
	logger.Warn("Warning message")
	logger.Error("Error occurred")

	// Formatted logging
	logger.Infof("User %s logged in with id %d", "john", 12345)
	logger.Debugf("Processing request for user %s", "john")

	// Structured logging with key-value pairs
	logger.Infow("User created", "user_id", 12345, "email", "user@example.com")
	logger.Errorw("Request failed", "error", "connection timeout", "retry", 3)

	// Create logger with persistent fields
	serviceLog := logger.NewLoggerWithFields(
		zap.String("service", "user-service"),
		zap.String("version", "1.0.0"),
	)
	serviceLog.Infow("Service initialized", "port", 8080)
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
// Quick development setup with debug level logging
undo := logger.InitDefaultLogger()
defer undo()

logger.Debug("Debug message")
logger.Info("App started")
```

### Production Setup

```go
// Production-ready setup with info level logging and optimized settings
undo := logger.InitProductionLogger()
defer undo()

logger.Info("Service started")
logger.Infow("Service initialized", "port", 8080, "version", "1.0.0")
```

### Convenience Logging Functions

The package provides global logging functions for quick and easy logging without creating logger instances:

#### Simple Logging

```go
// Basic logging - automatically initializes with default config if needed
logger.Debug("Debug message")
logger.Info("Information message")
logger.Warn("Warning message")
logger.Error("Error message")
```

#### Formatted Logging (Printf-style)

```go
// Formatted messages with type-safe formatting
logger.Debugf("User %s logged in at %s", username, time.Now())
logger.Infof("Processing %d items", count)
logger.Warnf("Connection attempt %d of %d failed", attempt, maxAttempts)
logger.Errorf("Failed to connect to %s: %v", host, err)
```

#### Structured Logging (Key-Value pairs)

```go
// Structured logging with key-value pairs for better searchability
logger.Debugw("Request received",
	"method", "GET",
	"path", "/api/users",
	"ip", "192.168.1.1",
)

logger.Infow("User created",
	"user_id", 12345,
	"username", "john",
	"email", "john@example.com",
)

logger.Warnw("Slow query detected",
	"query", "SELECT * FROM users",
	"duration_ms", 1500,
)

logger.Errorw("Database connection failed",
	"error", err.Error(),
	"host", "localhost",
	"port", 5432,
)
```

### Custom Configuration

```go
// Initialize with custom config
config := &logger.Config{
	LogLevel:          logger.LevelInfo,
	LogEncoding:       logger.EncodingJSON,
	LogOutputPaths:    []string{"/var/log/app.log"},
	DisableCaller:     false,
	DisableStacktrace: true,
	IsDevelopment:     false,
}
undo := logger.InitLogger(config)
defer undo()

// Use convenience functions
logger.Info("Logger initialized")
logger.Infow("Application ready", "version", "1.0.0")
```

### Logger with Default Fields

```go
// Initialize with default fields that will be added to all log messages
undo := logger.InitLogger(&logger.Config{
	LogLevel:    logger.LevelInfo,
	LogEncoding: logger.EncodingJSON,
},
	zap.String("app_name", "my-app"),
	zap.String("app_version", "1.0.0"),
	zap.String("environment", "production"),
)
defer undo()

// All log messages will include the default fields
logger.Info("User logged in") // Will include app_name, app_version, environment
```

### Component-Specific Logging

Create loggers with persistent fields for different components:

```go
// Initialize global logger first
undo := logger.InitDefaultLogger()
defer undo()

// Authentication component logger
authLog := logger.NewLoggerWithFields(
	zap.String("component", "auth"),
	zap.String("version", "2.1.0"),
)
authLog.Infow("User logged in", "user_id", "12345")

// Database component logger
dbLog := logger.NewLoggerWithFields(
	zap.String("component", "database"),
	zap.String("host", "localhost"),
)
dbLog.Infow("Connection established", "pool_size", 10)

// Or use global convenience functions
logger.Infow("Global event", "component", "main", "event", "startup")
```

### Error Handling

```go
// Using global convenience functions
if err != nil {
	logger.Errorw("Failed to process request",
		"error", err.Error(),
		"operation", "user_creation",
		"request_id", "req-123",
	)
}

// Using component-specific logger
log := logger.NewLoggerWithFields(
	zap.String("request_id", "req-123"),
	zap.String("user_id", "user-456"),
)

if err != nil {
	log.Errorw("Failed to process request",
		"error", err.Error(),
		"operation", "user_creation",
	)
}
```

## API Reference

### Initialization Functions

- **`InitLogger(config *Config, defaultLogFields ...zap.Field) func()`** - Initialize logger with custom configuration and optional default fields
- **`InitDefaultLogger() func()`** - Initialize with default configuration (debug level, development mode)
- **`InitProductionLogger() func()`** - Initialize with production configuration (info level, optimized settings)

### Convenience Logging Functions

All convenience functions auto-initialize the logger with default configuration if not already initialized.

#### Simple Logging

- **`Debug(args ...interface{})`** - Log debug message
- **`Info(args ...interface{})`** - Log info message
- **`Warn(args ...interface{})`** - Log warning message
- **`Error(args ...interface{})`** - Log error message

#### Formatted Logging (Printf-style)

- **`Debugf(template string, args ...interface{})`** - Log formatted debug message
- **`Infof(template string, args ...interface{})`** - Log formatted info message
- **`Warnf(template string, args ...interface{})`** - Log formatted warning message
- **`Errorf(template string, args ...interface{})`** - Log formatted error message

#### Structured Logging (Key-Value pairs)

- **`Debugw(msg string, keysAndValues ...interface{})`** - Log debug with structured fields
- **`Infow(msg string, keysAndValues ...interface{})`** - Log info with structured fields
- **`Warnw(msg string, keysAndValues ...interface{})`** - Log warning with structured fields
- **`Errorw(msg string, keysAndValues ...interface{})`** - Log error with structured fields

### Component Logger

- **`NewLoggerWithFields(fields ...zap.Field) *zap.SugaredLogger`** - Create a sugared logger with persistent fields

## Output Formats

### JSON Output

```json
{
  "level": "info",
  "ts": "2025-11-01T14:24:43.047+0700",
  "caller": "logger/zap.go:145",
  "msg": "User created",
  "user_id": 12345,
  "email": "user@example.com"
}
```

### Console Output

```
2025-11-01T14:24:43.047+0700	INFO	zap.go:145	User created	{"user_id": 12345, "email": "user@example.com"}
```

## Best Practices

### 1. Use Structured Logging

Prefer structured logging with key-value pairs over formatted strings for better searchability and analysis:

```go
// Good: Structured logging
logger.Infow("User created", "user_id", 12345, "email", "user@example.com")

// Less optimal: Formatted string
logger.Infof("User created with id %d and email %s", 12345, "user@example.com")
```

### 2. Use Appropriate Log Levels

- **Debug**: Detailed information for debugging purposes
- **Info**: General informational messages about application progress
- **Warn**: Warning messages for potentially harmful situations
- **Error**: Error messages for error events

### 3. Initialize Once

Initialize the logger once at application startup:

```go
func main() {
	undo := logger.InitProductionLogger()
	defer undo()

	// Rest of your application
}
```

### 4. Use Component-Specific Loggers

Create loggers with persistent fields for different components:

```go
authLog := logger.NewLoggerWithFields(zap.String("component", "auth"))
dbLog := logger.NewLoggerWithFields(zap.String("component", "database"))
```

### 5. Auto-Initialization

Convenience functions auto-initialize with default configuration if logger is not initialized:

```go
// No need to explicitly initialize for simple use cases
logger.Info("Application started") // Auto-initializes if needed
```
