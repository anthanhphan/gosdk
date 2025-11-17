# Logger Package

A high-performance, self-built structured logging package for Go applications. Provides fast, zero-allocation logging with JSON/Console output formats, configurable log levels, timezone support, and easy structured data integration. Built from scratch without external dependencies (except standard library).

## Features

- **Self-built implementation** - No external dependencies, built from scratch
- **Structured logging** - JSON and Console output formats
- **Configurable log levels** - Debug, Info, Warn, Error, Fatal
- **Asynchronous logging** - Non-blocking logging with background worker
- **Timezone support** - Configurable timezone for timestamps
- **Caller information** - Automatic file and line number tracking
- **Stack traces** - Automatic stack trace capture for errors
- **Default fields** - Persistent fields added to all log messages
- **Component loggers** - Create loggers with persistent context
- **Secure file handling** - Directory traversal protection

## Installation

```bash
go get github.com/anthanhphan/gosdk/logger
```

## Quick Start

```go
package main

import (
	"github.com/anthanhphan/gosdk/logger"
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
		logger.String("service", "user-service"),
		logger.String("version", "1.0.0"),
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
	Timezone          string    // Timezone for timestamps (IANA timezone name)
}
```

### Log Levels

- **`LevelDebug`** - Debug messages (most verbose)
- **`LevelInfo`** - Informational messages
- **`LevelWarn`** - Warning messages
- **`LevelError`** - Error messages
- **`LevelFatal`** - Fatal messages (logs and exits with code 1)

### Output Formats

- **`EncodingJSON`** - Structured JSON output with ordered keys (ts and caller first)
- **`EncodingConsole`** - Human-readable console output with color support in development mode

### Output Destinations

- **Empty array `[]string{}`** - Log to console (stdout)
- **`"stdout"`** - Explicitly log to stdout
- **`"stderr"`** - Log to stderr
- **File paths `[]string{"log/app.log"}`** - Log to files (directory is created automatically)
- **Multiple paths** - Log to multiple destinations simultaneously

### Timezone Configuration

The `Timezone` field accepts IANA timezone names (e.g., "Asia/Ho_Chi_Minh", "America/New_York", "UTC"). If empty or invalid, defaults to UTC.

### Development Settings

When `IsDevelopment: true`:

- Console output includes ANSI color codes for log levels
- More readable output format
- Includes caller information by default
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

### Asynchronous Logging

For high-performance applications, use asynchronous logging to avoid blocking:

```go
// Initialize async logger (non-blocking, uses background worker)
undo := logger.InitAsyncLogger(&logger.Config{
	LogLevel:    logger.LevelInfo,
	LogEncoding: logger.EncodingJSON,
	Timezone:    "Asia/Ho_Chi_Minh",
	LogOutputPaths: []string{"log/app.log"},
})
defer undo() // Automatically flushes remaining entries

// All logging operations are non-blocking
logger.Info("Application started")
logger.Infow("User created", "user_id", 12345)

// Manually flush if needed before shutdown
logger.Flush()
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
logger.Fatal("Fatal error") // Logs and exits with code 1
```

#### Formatted Logging (Printf-style)

```go
// Formatted messages
logger.Debugf("User %s logged in at %s", username, time.Now())
logger.Infof("Processing %d items", count)
logger.Warnf("Connection attempt %d of %d failed", attempt, maxAttempts)
logger.Errorf("Failed to connect to %s: %v", host, err)
logger.Fatalf("Fatal error: %s", err.Error()) // Logs and exits with code 1
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
	LogOutputPaths:    []string{"log/app.log"},
	DisableCaller:     false,
	DisableStacktrace: true,
	IsDevelopment:     false,
	Timezone:          "Asia/Ho_Chi_Minh",
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
	logger.String("app_name", "my-app"),
	logger.String("app_version", "1.0.0"),
	logger.String("environment", "production"),
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
	logger.String("component", "auth"),
	logger.String("version", "2.1.0"),
)
authLog.Infow("User logged in", "user_id", "12345")

// Database component logger
dbLog := logger.NewLoggerWithFields(
	logger.String("component", "database"),
	logger.String("host", "localhost"),
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
	logger.String("request_id", "req-123"),
	logger.String("user_id", "user-456"),
)

if err != nil {
	log.Errorw("Failed to process request",
		"error", err.Error(),
		"operation", "user_creation",
	)
}
```

### Fatal Logging

Fatal methods log at error level and then exit the program with `os.Exit(1)`:

```go
// Fatal logging - logs and exits
if criticalError {
	logger.Fatal("Critical error occurred")
}

// Fatal formatted logging
if err != nil {
	logger.Fatalf("Failed to start server: %v", err)
}
```

## API Reference

### Initialization Functions

- **`InitLogger(config *Config, defaultLogFields ...Field) func()`** - Initialize synchronous logger with custom configuration and optional default fields
- **`InitDefaultLogger() func()`** - Initialize with default configuration (debug level, development mode)
- **`InitProductionLogger() func()`** - Initialize with production configuration (info level, optimized settings)
- **`InitAsyncLogger(config *Config, defaultLogFields ...Field) func()`** - Initialize asynchronous logger with custom configuration (non-blocking, uses background worker)

### Convenience Logging Functions

All convenience functions auto-initialize the logger with default configuration if not already initialized. If async logger is initialized, these functions use async logging automatically.

#### Simple Logging

- **`Debug(args ...interface{})`** - Log debug message
- **`Info(args ...interface{})`** - Log info message
- **`Warn(args ...interface{})`** - Log warning message
- **`Error(args ...interface{})`** - Log error message
- **`Fatal(args ...interface{})`** - Log error message and exit with code 1

#### Formatted Logging (Printf-style)

- **`Debugf(template string, args ...interface{})`** - Log formatted debug message
- **`Infof(template string, args ...interface{})`** - Log formatted info message
- **`Warnf(template string, args ...interface{})`** - Log formatted warning message
- **`Errorf(template string, args ...interface{})`** - Log formatted error message
- **`Fatalf(template string, args ...interface{})`** - Log formatted error message and exit with code 1

#### Structured Logging (Key-Value pairs)

- **`Debugw(msg string, keysAndValues ...interface{})`** - Log debug with structured fields
- **`Infow(msg string, keysAndValues ...interface{})`** - Log info with structured fields
- **`Warnw(msg string, keysAndValues ...interface{})`** - Log warning with structured fields
- **`Errorw(msg string, keysAndValues ...interface{})`** - Log error with structured fields

### Utility Functions

- **`Flush()`** - Flush all queued log entries (for async logger)
- **`NewLoggerWithFields(fields ...Field) *Logger`** - Create a logger with persistent fields

### Field Constructors

- **`String(key, value string) Field`** - Create a string field
- **`Int(key string, value int) Field`** - Create an integer field
- **`Int64(key string, value int64) Field`** - Create an int64 field
- **`Float64(key string, value float64) Field`** - Create a float64 field
- **`Bool(key string, value bool) Field`** - Create a boolean field
- **`ErrorField(err error) Field`** - Create an error field
- **`Any(key string, value interface{}) Field`** - Create a field with any value type

## Output Formats

### JSON Output

JSON output has ordered keys with `ts` and `caller` as the first keys:

```json
{
  "ts": "2025-11-17T13:57:39+07:00",
  "caller": "logger/docs/example/main.go:25",
  "level": "info",
  "msg": "User created",
  "user_id": 12345,
  "email": "user@example.com"
}
```

### Console Output

Console output is human-readable with optional color support:

```
2025-11-17T13:57:39+07:00	INFO	logger/docs/example/main.go:25	User created	{"user_id": 12345, "email": "user@example.com"}
```

In development mode, log levels are colorized:

- Debug: Cyan
- Info: Green
- Warn: Yellow
- Error: Red

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
- **Fatal**: Critical errors that require immediate program termination

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
authLog := logger.NewLoggerWithFields(logger.String("component", "auth"))
dbLog := logger.NewLoggerWithFields(logger.String("component", "database"))
```

### 5. Use Async Logger for High-Throughput Applications

For applications with high logging volume, use async logger to avoid blocking:

```go
undo := logger.InitAsyncLogger(&logger.Config{
	LogLevel:    logger.LevelInfo,
	LogEncoding: logger.EncodingJSON,
	LogOutputPaths: []string{"log/app.log"},
})
defer undo()
```

### 6. Configure Timezone for Timestamps

Set timezone for more readable timestamps in your local timezone:

```go
config := &logger.Config{
	LogLevel:    logger.LevelInfo,
	LogEncoding: logger.EncodingJSON,
	Timezone:    "Asia/Ho_Chi_Minh", // IANA timezone name
}
```

### 7. File Output Best Practices

- Use relative paths like `"log/app.log"` (directory is created automatically)
- The `log/` folder should be added to `.gitignore`
- Multiple output paths are supported for redundancy

## Security

The logger package includes built-in security features:

- **Directory traversal protection** - File paths are validated to prevent directory traversal attacks
- **Secure file permissions** - Log files are created with `0600` permissions (read/write for owner only)
- **Path validation** - All file paths are validated and resolved to absolute paths within the working directory

## Performance

- **Zero-allocation logging** - Optimized for minimal memory allocations
- **Asynchronous logging** - Non-blocking logging with configurable queue size (default: 100)
- **Efficient encoding** - Fast JSON and Console encoding
- **Concurrent-safe** - All operations are safe for concurrent use

## Migration from Zap

If you're migrating from zap, note these differences:

- Use `logger.String()` instead of `zap.String()`
- Use `logger.Field` instead of `zap.Field`
- No `SugaredLogger` - all methods are directly on `Logger`
- Use `InitAsyncLogger()` for async logging instead of zap's async options
- Field constructors are in the `logger` package, not `zap`
