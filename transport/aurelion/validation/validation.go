package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	tagValidate = "validate"
	ruleSep     = ","
	paramSep    = "="
)

const (
	ruleRequired = "required"
	ruleMin      = "min"
	ruleMax      = "max"
	ruleEmail    = "email"
	ruleURL      = "url"
	ruleNumeric  = "numeric"
	ruleAlpha    = "alpha"
)

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

var (
	emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	urlRegex     = regexp.MustCompile(`^https?://[^\s]+$`)
	numericRegex = regexp.MustCompile(`^[0-9]+$`)
	alphaRegex   = regexp.MustCompile(`^[a-zA-Z]+$`)
)

// ValidationError represents a validation error with field and message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
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

// ToArray converts validation errors to an array format for easy frontend parsing.
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
func Validate(v interface{}) error {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("%s", msgNilPointer)
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf(msgInvalidType, v)
	}

	var errors ValidationErrors

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get(tagValidate)
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ruleSep)
		for _, rule := range rules {
			if err := validateRule(field.Name, value, strings.TrimSpace(rule)); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func validateRule(fieldName string, value reflect.Value, rule string) *ValidationError {
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
			return nil
		}
		return validateMin(fieldName, value, minVal)
	case ruleMax:
		if ruleParam == "" {
			return nil
		}
		maxVal, err := strconv.Atoi(ruleParam)
		if err != nil {
			return nil
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

func validateRequired(fieldName string, value reflect.Value) *ValidationError {
	if isZeroValue(value) {
		return &ValidationError{Field: fieldName, Message: msgRequired}
	}
	return nil
}

func validateMin(fieldName string, value reflect.Value, minVal int) *ValidationError {
	switch value.Kind() {
	case reflect.String:
		if len(value.String()) < minVal {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMinCharacters, minVal)}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() < minVal {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMinItems, minVal)}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() < int64(minVal) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMinValue, minVal)}
		}
	}
	return nil
}

func validateMax(fieldName string, value reflect.Value, maxVal int) *ValidationError {
	switch value.Kind() {
	case reflect.String:
		if len(value.String()) > maxVal {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxCharacters, maxVal)}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() > maxVal {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxItems, maxVal)}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() > int64(maxVal) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxValue, maxVal)}
		}
	}
	return nil
}

func validateEmail(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	email := value.String()
	if email == "" {
		return nil
	}
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: fieldName, Message: msgInvalidEmail}
	}
	return nil
}

func validateURL(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	url := value.String()
	if url == "" {
		return nil
	}
	if !urlRegex.MatchString(url) {
		return &ValidationError{Field: fieldName, Message: msgInvalidURL}
	}
	return nil
}

func validateNumeric(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	str := value.String()
	if str == "" {
		return nil
	}
	if !numericRegex.MatchString(str) {
		return &ValidationError{Field: fieldName, Message: msgNumericOnly}
	}
	return nil
}

func validateAlpha(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	str := value.String()
	if str == "" {
		return nil
	}
	if !alphaRegex.MatchString(str) {
		return &ValidationError{Field: fieldName, Message: msgAlphaOnly}
	}
	return nil
}

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

// ValidateAndParse combines BodyParser and Validate helper functions.
func ValidateAndParse(ctx ContextInterface, v interface{}) error {
	if err := ctx.BodyParser(v); err != nil {
		return fmt.Errorf(msgInvalidBodyParse, err)
	}

	if err := Validate(v); err != nil {
		return err
	}

	return nil
}
