package aurelion

import (
	"fmt"
	"net/http"
	"time"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// BusinessError represents a business logic error with code and message
type BusinessError struct {
	Code    int
	Message string
}

// Error implements the error interface
func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewError creates a new business error
//
// Input:
//   - code: Custom business error code
//   - message: Error message
//
// Output:
//   - *BusinessError: The business error
//
// Example:
//
//	err := aurelion.NewError(1001, "User not found")
//	return aurelion.Error(ctx, err)
func NewError(code int, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}

// NewErrorf creates a new business error with formatted message
//
// Input:
//   - code: Custom business error code
//   - format: Error message format string
//   - args: Format arguments
//
// Output:
//   - *BusinessError: The business error
//
// Example:
//
//	err := aurelion.NewErrorf(1002, "User %d not found", userID)
//	return aurelion.Error(ctx, err)
func NewErrorf(code int, format string, args ...interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// OK sends a successful response with HTTP 200
//
// Input:
//   - ctx: The request context
//   - message: Success message
//   - data: Optional response data
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.OK(ctx, "User created", user)
func OK(ctx Context, message string, data ...interface{}) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	response := APIResponse{
		Success:   true,
		Code:      200,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return ctx.Status(http.StatusOK).JSON(response)
}

// Error sends a business error response with HTTP 200 and custom error code
//
// Input:
//   - ctx: The request context
//   - err: The error (BusinessError or generic error)
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.Error(ctx, aurelion.NewError(1001, "User not found"))
func Error(ctx Context, err error) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if err == nil {
		return InternalServerError(ctx, "unknown error")
	}

	if bizErr, ok := err.(*BusinessError); ok {
		return ctx.Status(http.StatusOK).JSON(APIResponse{
			Success:   false,
			Code:      bizErr.Code,
			Message:   bizErr.Message,
			Timestamp: time.Now().UnixMilli(),
		})
	}
	return InternalServerError(ctx, err.Error())
}

// BadRequest sends a bad request error response with HTTP 200 and code 400
//
// Input:
//   - ctx: The request context
//   - message: Error message
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.BadRequest(ctx, "Invalid input")
func BadRequest(ctx Context, message string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Status(http.StatusOK).JSON(APIResponse{
		Success:   false,
		Code:      http.StatusBadRequest,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	})
}

// Unauthorized sends an unauthorized error response with HTTP 200 and code 401
//
// Input:
//   - ctx: The request context
//   - message: Error message
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.Unauthorized(ctx, "Authentication required")
func Unauthorized(ctx Context, message string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Status(http.StatusOK).JSON(APIResponse{
		Success:   false,
		Code:      http.StatusUnauthorized,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	})
}

// Forbidden sends a forbidden error response with HTTP 200 and code 403
//
// Input:
//   - ctx: The request context
//   - message: Error message
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.Forbidden(ctx, "Access denied")
func Forbidden(ctx Context, message string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Status(http.StatusOK).JSON(APIResponse{
		Success:   false,
		Code:      http.StatusForbidden,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	})
}

// NotFound sends a not found error response with HTTP 200 and code 404
//
// Input:
//   - ctx: The request context
//   - message: Error message
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.NotFound(ctx, "User not found")
func NotFound(ctx Context, message string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Status(http.StatusOK).JSON(APIResponse{
		Success:   false,
		Code:      http.StatusNotFound,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	})
}

// InternalServerError sends an internal server error response with HTTP 500
//
// Input:
//   - ctx: The request context
//   - message: Error message
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.InternalServerError(ctx, "Database error")
func InternalServerError(ctx Context, message string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success:   false,
		Code:      http.StatusInternalServerError,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	})
}

// HealthCheck sends a health check response indicating the server is healthy
//
// Input:
//   - ctx: The request context
//
// Output:
//   - error: Any error that occurred
//
// Example:
//
//	return aurelion.HealthCheck(ctx)
func HealthCheck(ctx Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Status(http.StatusOK).JSON(APIResponse{
		Success: true,
		Code:    http.StatusOK,
		Message: "Server is healthy",
		Data: Map{
			"status":    "healthy",
			"timestamp": time.Now().UnixMilli(),
		},
		Timestamp: time.Now().UnixMilli(),
	})
}
