# Routine Package

A safe goroutine management package for Go applications that provides panic recovery with accurate location tracking and dynamic function invocation with automatic type conversion. Built with best practices for production-ready applications.

## Installation

```bash
go get github.com/anthanhphan/gosdk/goroutine
```

## Quick Start

```go
package main

import (
	"time"

	routine "github.com/anthanhphan/gosdk/goroutine"
	"github.com/anthanhphan/gosdk/logger"
)

func main() {
	// Initialize logger
	undo := logger.InitDefaultLogger()
	defer undo()

	// Run a simple function in a goroutine
	routine.Run(func(msg string) {
		logger.Info("Message:", msg)
	}, "Hello, world!")

	// Panic recovery demonstration
	routine.Run(func() {
		panic("This panic will be recovered and logged with location")
	})

	time.Sleep(100 * time.Millisecond)
}
```

## Features

- **Panic Recovery**: Automatic panic recovery and logging with accurate panic location tracking
- **Dynamic Invocation**: Use reflection to invoke functions with any signature
- **Type Safety**: Automatic type conversion and validation for compatible types
- **Error Logging**: Structured logging for all errors and panics with location information
- **Production Ready**: Follows Go best practices and conventions

## API Reference

### Run

Starts a new goroutine and invokes the provided function with the given arguments. Any panic that occurs within the goroutine is automatically recovered and logged with the exact panic location.

```go
func Run(fn any, args ...any)
```

**Parameters:**

- `fn`: The function to be invoked in the new goroutine. Must be a valid function type.
- `args`: Variadic list of arguments to pass to the invoked function.

**Output:**

- None (function executes asynchronously in a goroutine)

**Example:**

```go
// Function with string argument
routine.Run(func(msg string) {
	logger.Info("Message:", msg)
}, "Hello, world!")

// Function with multiple arguments
routine.Run(func(a, b int) {
	logger.Infof("Sum: %d", a+b)
}, 10, 20)

// Function with no arguments
routine.Run(func() {
	logger.Info("Running in background")
})

// Panic recovery - automatically logged with location
routine.Run(func() {
	panic("something went wrong")
})
// Output: {"level":"error","msg":"panic recovered in goroutine","error":"something went wrong","panic_at":"file.go:123"}
```

## Usage Examples

### Basic Background Task

```go
routine.Run(func() {
	// Long-running background task
	for {
		// Do work
		time.Sleep(1 * time.Second)
	}
})
```

### Task with Arguments

```go
routine.Run(func(userID int, message string) {
	logger.Infow("Sending notification",
		"user_id", userID,
		"message", message,
	)
}, 12345, "Welcome to our service!")
```

### Concurrent Processing

```go
items := []string{"item1", "item2", "item3"}

for _, item := range items {
	item := item // Capture loop variable
	routine.Run(func(i string) {
		logger.Infow("Processing item", "item", i)
		// Process item
	}, item)
}
```

### Panic Recovery with Location Tracking

```go
// Panic will be automatically recovered and logged with accurate location
routine.Run(func() {
	// Your code that might panic
	panic("something went wrong")
})

// The panic is recovered by the package with location information:
// {"level":"error","msg":"panic recovered in goroutine","error":"something went wrong","panic_at":"file.go:123"}
// Your application continues running safely
```

## Advanced Features

### Type Conversion

The package automatically handles type conversion for compatible types:

```go
// int to int32
routine.Run(func(n int32) {
	logger.Infof("Number: %d", n)
}, 10) // int(10) is converted to int32

// int to int64
routine.Run(func(n int64) {
	logger.Infof("Number: %d", n)
}, 10) // int(10) is converted to int64

// uint to int
routine.Run(func(n int) {
	logger.Infof("Number: %d", n)
}, uint(10)) // uint(10) is converted to int
```

### Invalid Function Handling

If an invalid function is provided, an error is logged but the application continues:

```go
// This will log an error but not panic
routine.Run("not a function", "arg1", "arg2")
```

### Argument Validation

The package validates argument count and types:

```go
// Insufficient arguments - error logged
routine.Run(func(a, b int) {
	// ...
}, 10) // Missing second argument
// Error: "insufficient arguments: expected 2, got 1"

// Type mismatch - error logged
routine.Run(func(s string) {
	// ...
}, 123) // int cannot be converted to string
// Error: "type mismatch at index 0: expected string, got int"

// Excess arguments - warning logged, function still executes
routine.Run(func(a int) {
	// ...
}, 10, 20, 30) // Excess arguments logged as warning
```

### Nil Pointer Arguments

Nil values are properly handled for pointer and interface types:

```go
// Nil pointer argument
routine.Run(func(s *string) {
	if s == nil {
		logger.Info("Received nil pointer")
	}
}, nil)

// Nil interface argument
routine.Run(func(v interface{}) {
	if v == nil {
		logger.Info("Received nil interface")
	}
}, nil)
```

## Integration with Logger

The routine package uses the logger package for structured logging. Make sure to initialize the logger before using the routine package:

```go
import (
	routine "github.com/anthanhphan/gosdk/goroutine"
	"github.com/anthanhphan/gosdk/logger"
)

func main() {
	// Initialize logger
	undo := logger.InitDefaultLogger()
	defer undo()

	// Now you can use routine package
	routine.Run(func() {
		logger.Info("Running in goroutine")
	})
}
```

## Panic Location Tracking

The package automatically captures and logs the exact location where a panic occurs:

```go
routine.Run(func() {
	panic("critical error")
})

// Log output includes panic location:
// {
//   "level": "error",
//   "ts": "2025-11-02T...",
//   "caller": "goroutine/routine.go:401",
//   "msg": "panic recovered in goroutine",
//   "prefix": "routine::recoverPanic",
//   "error": "critical error",
//   "panic_at": "user/app.go:42"
// }
```

The `panic_at` field shows the exact file and line number where the panic occurred in your code, making debugging much easier.

## Best Practices

1. **Initialize logger**: Make sure to initialize the logger before using the routine package
2. **Use structured logging**: The package uses structured logging with panic location tracking
3. **Avoid passing nil functions**: Always validate function parameters before calling `Run`
4. **Capture loop variables**: When using loops with goroutines, capture loop variables properly
5. **Handle panics gracefully**: Panics are automatically recovered, but design your code to avoid panics when possible
6. **Type compatibility**: Ensure argument types are compatible or convertible to function parameter types
7. **Error handling**: Check argument validation errors in logs for debugging

## Performance Considerations

- The package uses reflection for dynamic function invocation, which has a small performance overhead
- For high-performance scenarios, consider using direct goroutine calls if you don't need panic recovery
- Stack trace parsing uses a pool of buffers to minimize allocations
- Panic location extraction filters out wrapper and runtime frames efficiently
- Type conversion is performed at runtime, so ensure types are compatible for best performance

## Examples

See the [example directory](./example/main.go) for complete working examples demonstrating all features of the package.

## License

Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>
