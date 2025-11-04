package aurelion

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validation tag names
const (
	tagValidate = "validate"
	ruleSep     = ","
	paramSep    = "="
)

// Validation rule names
const (
	ruleRequired = "required"
	ruleMin      = "min"
	ruleMax      = "max"
	ruleEmail    = "email"
	ruleURL      = "url"
	ruleNumeric  = "numeric"
	ruleAlpha    = "alpha"
)

// Validation error messages
const (
	msgRequired         = "is required"
	msgMinCharacters    = "must be at least %d characters"
	msgMinItems         = "must contain at least %d items"
	msgMinValue         = "must be at least %d"
	msgMaxCharacters    = "must be at most %d characters"
	msgMaxItems         = "must contain at most %d items"
	msgMaxValue         = "must be at most %d"
	msgInvalidEmail     = "must be a valid email address"
	msgInvalidURL       = "must be a valid URL"
	msgNumericOnly      = "must contain only numeric characters"
	msgAlphaOnly        = "must contain only alphabetic characters"
	msgNilPointer       = "cannot validate nil pointer"
	msgInvalidType      = "validate requires a struct or pointer to struct, got %T"
	msgInvalidBodyParse = "invalid request body: %w"
)

// Compiled regex patterns for performance (compile once, use many times)
var (
	emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	urlRegex     = regexp.MustCompile(`^https?://[^\s]+$`)
	numericRegex = regexp.MustCompile(`^[0-9]+$`)
	alphaRegex   = regexp.MustCompile(`^[a-zA-Z]+$`)
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ToMap converts validation errors to a map for JSON response.
// Deprecated: Use ToArray() instead for better frontend parsing.
func (e ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string)
	for _, err := range e {
		result[strings.ToLower(err.Field)] = err.Message
	}
	return result
}

// ToArray converts validation errors to an array format for easy frontend parsing.
// This is the recommended format as it preserves order and is easier to iterate.
//
// Output:
//   - []map[string]string: Array of error objects with "field" and "message" keys
//
// Example:
//
//	if validationErr, ok := err.(aurelion.ValidationErrors); ok {
//	    return ctx.Status(400).JSON(aurelion.Map{
//	        "errors": validationErr.ToArray(),
//	    })
//	}
//	// Response: {"errors": [{"field": "email", "message": "..."}, {"field": "age", "message": "..."}]}
func (e ValidationErrors) ToArray() []map[string]string {
	result := make([]map[string]string, len(e))
	for i, err := range e {
		result[i] = map[string]string{
			"field":   strings.ToLower(err.Field),
			"message": err.Message,
		}
	}
	return result
}

// Validate validates a struct using basic validation rules from tags.
// This is a simple implementation suitable for basic validation.
// For production with complex validation needs, integrate with github.com/go-playground/validator/v10
//
// Supported tags:
//   - required: Field must not be zero value
//   - min=N: String/slice/array must have minimum length N
//   - max=N: String/slice/array must have maximum length N
//   - email: String must be valid email format
//   - url: String must be valid URL format
//   - numeric: String must contain only numeric characters
//   - alpha: String must contain only alphabetic characters
//
// Input:
//   - v: The struct to validate (must be a pointer or struct)
//
// Output:
//   - error: ValidationErrors if validation fails, nil otherwise
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `validate:"required,min=3,max=50"`
//	    Email string `validate:"required,email"`
//	    Age   int    `validate:"min=18,max=100"`
//	}
//
//	var req CreateUserRequest
//	if err := ctx.BodyParser(&req); err != nil {
//	    return aurelion.BadRequest(ctx, "Invalid request body")
//	}
//
//	if err := aurelion.Validate(&req); err != nil {
//	    return aurelion.BadRequest(ctx, err.Error())
//	}
func Validate(v interface{}) error {
	val := reflect.ValueOf(v)

	// Handle pointer types
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("%s", msgNilPointer)
		}
		val = val.Elem()
	}

	// Ensure we're validating a struct
	if val.Kind() != reflect.Struct {
		return fmt.Errorf(msgInvalidType, v)
	}

	var errors ValidationErrors

	// Iterate through struct fields
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// Skip unexported fields - they cannot be validated
		if !field.IsExported() {
			continue
		}

		// Get validation tag
		tag := field.Tag.Get(tagValidate)
		if tag == "" {
			continue
		}

		// Parse and validate each rule
		rules := strings.Split(tag, ruleSep)
		for _, rule := range rules {
			if err := validateRule(field.Name, value, strings.TrimSpace(rule)); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	// Return validation errors if any
	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateRule validates a single validation rule against a field value.
// Returns ValidationError if validation fails, nil if validation passes.
// This is an internal helper function that routes to specific validation functions.
//
// Supported rules:
//   - required: Field must not be zero value
//   - min=N: Minimum length/value constraint
//   - max=N: Maximum length/value constraint
//   - email: Valid email format
//   - url: Valid URL format
//   - numeric: Only numeric characters
//   - alpha: Only alphabetic characters
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field
//   - rule: The validation rule string (e.g., "required", "min=3", "email")
//
// Output:
//   - *ValidationError: Error if validation fails, nil if passes
func validateRule(fieldName string, value reflect.Value, rule string) *ValidationError {
	// Parse rule with parameter (e.g., "min=3" -> ruleName="min", ruleParam="3")
	parts := strings.SplitN(rule, paramSep, 2)
	ruleName := parts[0]
	var ruleParam string
	if len(parts) > 1 {
		ruleParam = parts[1]
	}

	switch ruleName {
	case ruleRequired:
		return validateRequired(fieldName, value)

	case ruleMin:
		if ruleParam == "" {
			return nil
		}
		minVal, err := strconv.Atoi(ruleParam)
		if err != nil {
			return nil // Invalid parameter, skip validation
		}
		return validateMin(fieldName, value, minVal)

	case ruleMax:
		if ruleParam == "" {
			return nil
		}
		maxVal, err := strconv.Atoi(ruleParam)
		if err != nil {
			return nil // Invalid parameter, skip validation
		}
		return validateMax(fieldName, value, maxVal)

	case ruleEmail:
		return validateEmail(fieldName, value)

	case ruleURL:
		return validateURL(fieldName, value)

	case ruleNumeric:
		return validateNumeric(fieldName, value)

	case ruleAlpha:
		return validateAlpha(fieldName, value)
	}

	return nil
}

// isZeroValue checks if a value is the zero value for its type.
// This is an internal helper function used by validation rules.
//
// Input:
//   - v: The reflection value to check
//
// Output:
//   - bool: True if the value is zero, false otherwise
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	}
	return false
}

// ValidateAndParse is a helper that combines BodyParser and Validate.
//
// Input:
//   - ctx: The request context
//   - v: Pointer to struct to parse and validate
//
// Output:
//   - error: Any parsing or validation error
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name" validate:"required,min=3"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	var req CreateUserRequest
//	if err := aurelion.ValidateAndParse(ctx, &req); err != nil {
//	    if validationErr, ok := err.(aurelion.ValidationErrors); ok {
//	        return aurelion.BadRequest(ctx, validationErr.Error())
//	    }
//	    return aurelion.BadRequest(ctx, "Invalid request body")
//	}
func ValidateAndParse(ctx Context, v interface{}) error {
	// Parse request body into struct
	if err := ctx.BodyParser(v); err != nil {
		return fmt.Errorf(msgInvalidBodyParse, err)
	}

	// Validate the parsed struct
	if err := Validate(v); err != nil {
		return err
	}

	return nil
}

// validateRequired checks if a field has a non-zero value.
// Returns a ValidationError if the field is empty/zero, nil otherwise.
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field
//
// Output:
//   - *ValidationError: Error if field is zero, nil if has value
func validateRequired(fieldName string, value reflect.Value) *ValidationError {
	if isZeroValue(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: msgRequired,
		}
	}
	return nil
}

// validateMin validates minimum length/value constraint.
// Supports strings (length), slices/arrays (length), and integers (value).
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field
//   - minVal: The minimum value/length required
//
// Output:
//   - *ValidationError: Error if constraint violated, nil otherwise
func validateMin(fieldName string, value reflect.Value, minVal int) *ValidationError {
	switch value.Kind() {
	case reflect.String:
		if len(value.String()) < minVal {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf(msgMinCharacters, minVal),
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() < minVal {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf(msgMinItems, minVal),
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() < int64(minVal) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf(msgMinValue, minVal),
			}
		}
	}
	return nil
}

// validateMax validates maximum length/value constraint.
// Supports strings (length), slices/arrays (length), and integers (value).
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field
//   - maxVal: The maximum value/length allowed
//
// Output:
//   - *ValidationError: Error if constraint violated, nil otherwise
func validateMax(fieldName string, value reflect.Value, maxVal int) *ValidationError {
	switch value.Kind() {
	case reflect.String:
		if len(value.String()) > maxVal {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf(msgMaxCharacters, maxVal),
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() > maxVal {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf(msgMaxItems, maxVal),
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() > int64(maxVal) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf(msgMaxValue, maxVal),
			}
		}
	}
	return nil
}

// validateEmail validates email format using pre-compiled regex.
// Uses a pre-compiled regex pattern for better performance.
// Empty strings are considered valid (use "required" rule to enforce non-empty).
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field (must be string)
//
// Output:
//   - *ValidationError: Error if email format is invalid, nil otherwise
func validateEmail(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}

	email := value.String()
	if email == "" {
		return nil // Empty is valid unless required is also set
	}

	// Use pre-compiled regex for better performance
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   fieldName,
			Message: msgInvalidEmail,
		}
	}
	return nil
}

// validateURL validates URL format using pre-compiled regex.
// Validates that the string is a valid HTTP/HTTPS URL.
// Uses a pre-compiled regex pattern for better performance.
// Empty strings are considered valid (use "required" rule to enforce non-empty).
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field (must be string)
//
// Output:
//   - *ValidationError: Error if URL format is invalid, nil otherwise
func validateURL(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}

	url := value.String()
	if url == "" {
		return nil // Empty is valid unless required is also set
	}

	// Use pre-compiled regex for better performance
	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   fieldName,
			Message: msgInvalidURL,
		}
	}
	return nil
}

// validateNumeric validates that string contains only numeric characters.
// Uses a pre-compiled regex pattern for better performance.
// Empty strings are considered valid (use "required" rule to enforce non-empty).
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field (must be string)
//
// Output:
//   - *ValidationError: Error if string contains non-numeric characters, nil otherwise
func validateNumeric(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}

	str := value.String()
	if str == "" {
		return nil
	}

	// Use pre-compiled regex for better performance
	if !numericRegex.MatchString(str) {
		return &ValidationError{
			Field:   fieldName,
			Message: msgNumericOnly,
		}
	}
	return nil
}

// validateAlpha validates that string contains only alphabetic characters.
// Uses a pre-compiled regex pattern for better performance.
// Empty strings are considered valid (use "required" rule to enforce non-empty).
//
// Input:
//   - fieldName: The name of the field being validated
//   - value: The reflection value of the field (must be string)
//
// Output:
//   - *ValidationError: Error if string contains non-alphabetic characters, nil otherwise
func validateAlpha(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}

	str := value.String()
	if str == "" {
		return nil
	}

	// Use pre-compiled regex for better performance
	if !alphaRegex.MatchString(str) {
		return &ValidationError{
			Field:   fieldName,
			Message: msgAlphaOnly,
		}
	}
	return nil
}
