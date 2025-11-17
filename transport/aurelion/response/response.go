package response

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
)

// contextAdapter adapts core.Context to ContextInterface.
type contextAdapter struct {
	ctx core.Context
}

// Status implements ContextInterface.
func (a *contextAdapter) Status(status int) ContextInterface {
	a.ctx.Status(status)
	return a
}

// JSON implements ContextInterface.
func (a *contextAdapter) JSON(data interface{}) error {
	return a.ctx.JSON(data)
}

// Locals implements ContextInterface.
func (a *contextAdapter) Locals(key string, value ...interface{}) interface{} {
	return a.ctx.Locals(key, value...)
}

// Next implements ContextInterface.
func (a *contextAdapter) Next() error {
	return a.ctx.Next()
}

// GetAllLocals implements ContextInterface.
func (a *contextAdapter) GetAllLocals() map[string]interface{} {
	return a.ctx.GetAllLocals()
}

// adaptCoreContext adapts core.Context to ContextInterface.
func adaptCoreContext(ctx core.Context) ContextInterface {
	if ctx == nil {
		return nil
	}
	// Note: We can't check if ctx is already a ContextInterface because
	// core.Context and ContextInterface have incompatible Status method signatures.
	// Always create an adapter.
	return &contextAdapter{ctx: ctx}
}

// toContextInterface converts interface{} to ContextInterface, handling both ContextInterface and core.Context.
func toContextInterface(ctx interface{}) ContextInterface {
	if ctx == nil {
		return nil
	}
	// If it's already a ContextInterface, return it
	if ci, ok := ctx.(ContextInterface); ok {
		return ci
	}
	// If it's a core.Context, adapt it
	// Note: We can't use direct type assertion because core.Context and ContextInterface
	// have incompatible Status method signatures. We use a helper function to check.
	if isCoreContext(ctx) {
		return adaptCoreContext(ctx.(core.Context))
	}
	return nil
}

// isCoreContext checks if the value implements core.Context
// This avoids the impossible type assertion issue by using a helper that checks the underlying type
func isCoreContext(ctx interface{}) bool {
	if ctx == nil {
		return false
	}
	// Try direct type assertion (this will work if it's actually a core.Context)
	// We use a helper function to avoid the impossible type assertion warning
	// since core.Context and ContextInterface have incompatible Status method signatures
	_, ok := ctx.(core.Context)
	return ok
}

const (
	MsgHealthCheckHealthy = "Server is healthy"
	MsgValidationFailed   = "Validation failed"

	ErrorTypeValidation          = "validation_error"
	ErrorTypeBusiness            = "business_error"
	ErrorTypePermission          = "permission_error"
	ErrorTypeRateLimit           = "rate_limit_error"
	ErrorTypeExternal            = "external_api_error"
	ErrorTypeInternalServerError = "internal_server_error"

	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// APIResponse represents a standard API response structure.
type APIResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorData  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorData represents flexible error information that can be extended.
type ErrorData struct {
	Type       string                 `json:"type,omitempty"`
	Validation []map[string]string    `json:"validation,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// BusinessError represents a business logic error with code and message.
type BusinessError struct {
	Code    int
	Message string
}

// Error implements the error interface.
func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Is checks if the error matches the target error type.
func (e *BusinessError) Is(target error) bool {
	if target == nil {
		return false
	}
	if t, ok := target.(*BusinessError); ok {
		return e.Code == t.Code && e.Message == t.Message
	}
	return false
}

// NewError creates a new business error.
func NewError(code int, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}

// NewErrorf creates a new business error with formatted message.
func NewErrorf(code int, format string, args ...interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// OK sends a successful response with HTTP 200.
func OK(ctx interface{}, message string, data ...interface{}) error {
	ctxInterface := toContextInterface(ctx)
	if err := validateContext(ctxInterface); err != nil {
		return err
	}

	response := buildAPIResponse(true, http.StatusOK, message, data...)
	return ctxInterface.Status(http.StatusOK).JSON(response)
}

// Error sends a business error response with HTTP 200 and custom error code.
func Error(ctx interface{}, err error) error {
	ctxInterface := toContextInterface(ctx)
	if validateErr := validateContext(ctxInterface); validateErr != nil {
		return validateErr
	}
	if err == nil {
		return InternalServerError(ctxInterface, "unknown error")
	}

	if bizErr, ok := err.(*BusinessError); ok {
		return ctxInterface.Status(http.StatusOK).JSON(buildAPIResponse(false, bizErr.Code, bizErr.Message))
	}
	return InternalServerError(ctxInterface, err.Error())
}

// BadRequest sends a bad request error response.
func BadRequest(ctx interface{}, message string) error {
	return sendProperStatus(toContextInterface(ctx), http.StatusBadRequest, message)
}

// Unauthorized sends an unauthorized error response.
func Unauthorized(ctx interface{}, message string) error {
	return sendProperStatus(toContextInterface(ctx), http.StatusUnauthorized, message)
}

// Forbidden sends a forbidden error response.
func Forbidden(ctx interface{}, message string) error {
	return sendProperStatus(toContextInterface(ctx), http.StatusForbidden, message)
}

// NotFound sends a not found error response.
func NotFound(ctx interface{}, message string) error {
	return sendProperStatus(toContextInterface(ctx), http.StatusNotFound, message)
}

// InternalServerError sends an internal server error response with HTTP 500.
func InternalServerError(ctx interface{}, message string) error {
	ctxInterface := toContextInterface(ctx)
	if err := validateContext(ctxInterface); err != nil {
		return err
	}
	return ctxInterface.Status(http.StatusInternalServerError).JSON(buildAPIResponse(false, http.StatusInternalServerError, message))
}

// HealthCheck sends a health check response indicating the server is healthy.
func HealthCheck(ctx interface{}) error {
	ctxInterface := toContextInterface(ctx)
	if err := validateContext(ctxInterface); err != nil {
		return err
	}

	response := buildAPIResponse(true, http.StatusOK, MsgHealthCheckHealthy, Map{
		"status":    HealthStatusHealthy,
		"timestamp": time.Now().UnixMilli(),
	})
	return ctxInterface.Status(http.StatusOK).JSON(response)
}

// ErrorWithDetails sends an error response with detailed error information.
func ErrorWithDetails(ctx interface{}, code int, message string, errorData *ErrorData) error {
	ctxInterface := toContextInterface(ctx)
	if err := validateContext(ctxInterface); err != nil {
		return err
	}

	// Use a helper that can access config properly
	statusCode := determineHTTPStatusInternal(ctxInterface, code)
	response := APIResponse{
		Success:   false,
		Code:      code,
		Message:   message,
		Error:     errorData,
		Timestamp: time.Now().UnixMilli(),
	}

	return ctxInterface.Status(statusCode).JSON(response)
}

func sendProperStatus(ctx ContextInterface, properStatus int, message string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	statusCode := determineHTTPStatusInternal(ctx, properStatus)
	return ctx.Status(statusCode).JSON(buildAPIResponse(false, properStatus, message))
}

func buildAPIResponse(success bool, code int, message string, data ...interface{}) APIResponse {
	response := APIResponse{
		Success:   success,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return response
}

// determineHTTPStatus determines the HTTP status code based on the context configuration.
func determineHTTPStatus(ctx interface{}, properStatusCode int) int {
	return determineHTTPStatusInternal(toContextInterface(ctx), properStatusCode)
}

func determineHTTPStatusInternal(ctx ContextInterface, properStatusCode int) int {
	// Check if context is nil
	if ctx == nil {
		return getDefaultStatus(properStatusCode)
	}
	// Try to get config from context locals
	cfgValue := ctx.Locals("aurelion_config")
	if cfgValue == nil {
		return getDefaultStatus(properStatusCode)
	}
	// Check if config has UseProperHTTPStatus enabled
	if hasUseProperHTTPStatus(cfgValue) {
		return properStatusCode
	}
	// Default behavior
	return getDefaultStatus(properStatusCode)
}

// getDefaultStatus returns the default status code based on properStatusCode
func getDefaultStatus(properStatusCode int) int {
	if properStatusCode >= http.StatusInternalServerError {
		return properStatusCode
	}
	return http.StatusOK
}

// hasUseProperHTTPStatus checks if the config value has UseProperHTTPStatus enabled
func hasUseProperHTTPStatus(cfgValue interface{}) bool {
	// Check struct field using reflection
	if checkStructField(cfgValue) {
		return true
	}
	// Check map fields
	return checkMapFields(cfgValue)
}

// checkStructField checks if a struct has UseProperHTTPStatus field set to true
func checkStructField(cfgValue interface{}) bool {
	cfgValueType := reflect.TypeOf(cfgValue)
	if cfgValueType.Kind() == reflect.Ptr {
		cfgValueType = cfgValueType.Elem()
	}
	if cfgValueType.Kind() != reflect.Struct {
		return false
	}
	cfgValueValue := reflect.ValueOf(cfgValue)
	if cfgValueValue.Kind() == reflect.Ptr {
		cfgValueValue = cfgValueValue.Elem()
	}
	field := cfgValueValue.FieldByName("UseProperHTTPStatus")
	return field.IsValid() && field.Kind() == reflect.Bool && field.Bool()
}

// checkMapFields checks if a map has UseProperHTTPStatus field set to true
func checkMapFields(cfgValue interface{}) bool {
	cfgMap, ok := cfgValue.(map[string]interface{})
	if !ok {
		return false
	}
	if useProper, ok := cfgMap["use_proper_http_status"].(bool); ok && useProper {
		return true
	}
	if useProper, ok := cfgMap["UseProperHTTPStatus"].(bool); ok && useProper {
		return true
	}
	return false
}

func validateContext(ctx interface{}) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	ctxInterface := toContextInterface(ctx)
	if ctxInterface == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}
