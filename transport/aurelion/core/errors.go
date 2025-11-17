package core

import (
	"errors"
	"fmt"
)

// Sentinel errors for common validation and runtime failures.
var (
	ErrContextNil   = errors.New("context cannot be nil")
	ErrUnknownError = errors.New("unknown error")
	ErrConfigNil    = errors.New("config cannot be nil")
	ErrHandlerNil   = errors.New("handler cannot be nil")
	ErrInvalidRoute = errors.New("invalid route configuration")
)

// HTTPStatusCode represents standard HTTP status codes used in the framework.
const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusInternalServerError = 500
)

// ConfigValidationError represents configuration validation failures with detailed context.
type ConfigValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error implements the error interface for ConfigValidationError.
//
// Input:
//   - None (receiver method)
//
// Output:
//   - string: Human-readable error message
//
// Example:
//
//	err := &ConfigValidationError{Field: "port", Message: "must be between 1 and 65535", Value: 0}
//	fmt.Println(err.Error()) // Output: "config validation failed: port: must be between 1 and 65535 (got: 0)"
func (e *ConfigValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("config validation failed: %s: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("config validation failed: %s: %s", e.Field, e.Message)
}

// Unwrap returns the underlying error for error chain compatibility.
func (e *ConfigValidationError) Unwrap() error {
	return errors.New(e.Message)
}

// RouteValidationError represents route configuration validation failures.
type RouteValidationError struct {
	Path    string
	Method  string
	Message string
}

// Error implements the error interface for RouteValidationError.
//
// Input:
//   - None (receiver method)
//
// Output:
//   - string: Human-readable error message with route context
//
// Example:
//
//	err := &RouteValidationError{Path: "/users", Method: "GET", Message: "handler is required"}
//	fmt.Println(err.Error()) // Output: "route validation failed: GET /users: handler is required"
func (e *RouteValidationError) Error() string {
	if e.Method != "" && e.Path != "" {
		return fmt.Sprintf("route validation failed: %s %s: %s", e.Method, e.Path, e.Message)
	}
	if e.Path != "" {
		return fmt.Sprintf("route validation failed: %s: %s", e.Path, e.Message)
	}
	return fmt.Sprintf("route validation failed: %s", e.Message)
}

// Unwrap returns the underlying error for error chain compatibility.
func (e *RouteValidationError) Unwrap() error {
	return ErrInvalidRoute
}

// MiddlewareError wraps errors that occur during middleware execution.
type MiddlewareError struct {
	MiddlewareName string
	Cause          error
}

// Error implements the error interface for MiddlewareError.
//
// Input:
//   - None (receiver method)
//
// Output:
//   - string: Error message with middleware context
//
// Example:
//
//	err := &MiddlewareError{MiddlewareName: "authentication", Cause: errors.New("token expired")}
//	fmt.Println(err.Error()) // Output: "middleware 'authentication' failed: token expired"
func (e *MiddlewareError) Error() string {
	if e.MiddlewareName != "" {
		return fmt.Sprintf("middleware '%s' failed: %v", e.MiddlewareName, e.Cause)
	}
	return fmt.Sprintf("middleware failed: %v", e.Cause)
}

// Unwrap returns the underlying error for error chain compatibility.
func (e *MiddlewareError) Unwrap() error {
	return e.Cause
}

// ServerError represents server-level errors with additional context.
type ServerError struct {
	Operation string
	Cause     error
}

// Error implements the error interface for ServerError.
//
// Input:
//   - None (receiver method)
//
// Output:
//   - string: Error message with server operation context
//
// Example:
//
//	err := &ServerError{Operation: "start", Cause: errors.New("port already in use")}
//	fmt.Println(err.Error()) // Output: "server operation 'start' failed: port already in use"
func (e *ServerError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("server operation '%s' failed: %v", e.Operation, e.Cause)
	}
	return fmt.Sprintf("server operation failed: %v", e.Cause)
}

// Unwrap returns the underlying error for error chain compatibility.
func (e *ServerError) Unwrap() error {
	return e.Cause
}
