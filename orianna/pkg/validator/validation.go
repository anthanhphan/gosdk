// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Validation Constants

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
	ruleOneOf    = "oneof"
	ruleLen      = "len"
	ruleGt       = "gt"
	ruleGte      = "gte"
	ruleLt       = "lt"
	ruleLte      = "lte"
)

const (
	msgRequired      = "is required"
	msgMinCharacters = "must be at least %d characters"
	msgMinItems      = "must contain at least %d items"
	msgMinValue      = "must be at least %d"
	msgMaxCharacters = "must be at most %d characters"
	msgMaxItems      = "must contain at most %d items"
	msgMaxValue      = "must be at most %d"
	msgInvalidEmail  = "must be a valid email address"
	msgInvalidURL    = "must be a valid URL"
	msgNumericOnly   = "must contain only numeric characters"
	msgAlphaOnly     = "must contain only alphabetic characters"
	msgNilPointer    = "cannot validate nil pointer"
	msgInvalidType   = "validate requires a struct or pointer to struct, got %T"
	msgOneOf         = "must be one of: %s"
	msgLen           = "must have exact length of %d"
	msgGt            = "must be greater than %s"
	msgGte           = "must be greater than or equal to %s"
	msgLt            = "must be less than %s"
	msgLte           = "must be less than or equal to %s"
)

// Exported validation regex patterns for reuse
var (
	// EmailRegex is the regular expression for validating email addresses
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// URLRegex is the regular expression for validating URLs
	URLRegex = regexp.MustCompile(`^https?://[^\s]+$`)

	// NumericRegex is the regular expression for validating numeric strings
	NumericRegex = regexp.MustCompile(`^[0-9]+$`)

	// AlphaRegex is the regular expression for validating alphabetic strings
	AlphaRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

	// AlphanumericRegex is the regular expression for validating alphanumeric strings
	AlphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// UUIDRegex is the regular expression for validating UUIDs
	UUIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

	// HexColorRegex is the regular expression for validating hex color codes
	HexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
)

// Custom Validation

// CustomValidationRule defines a custom validation rule function
type CustomValidationRule func(fieldName string, value reflect.Value, param string) *ValidationError

var (
	// customRules stores registered custom validation rules
	customRules   = make(map[string]CustomValidationRule)
	customRulesMu sync.RWMutex
)

// RegisterValidationRule registers a custom validation rule.
// Custom rules can be used by including their name in validate tags.
// This function is thread-safe.
//
// Input:
//   - name: The rule name to use in validate tags
//   - rule: The validation function to execute
//
// Output:
//   - None
//
// Example:
//
//	orianna.RegisterValidationRule("divisibleby", func(fieldName string, value reflect.Value, param string) *ValidationError {
//	    divisor, _ := strconv.Atoi(param)
//	    if value.Kind() == reflect.Int && value.Int() % int64(divisor) != 0 {
//	        return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be divisible by %d", divisor)}
//	    }
//	    return nil
//	})
//
//	// Usage: `validate:"divisibleby=5"`
func RegisterValidationRule(name string, rule CustomValidationRule) {
	if name != "" && rule != nil {
		customRulesMu.Lock()
		customRules[name] = rule
		customRulesMu.Unlock()
	}
}

// GetCustomValidationRule retrieves a registered custom validation rule.
// This function is thread-safe.
func GetCustomValidationRule(name string) (CustomValidationRule, bool) {
	customRulesMu.RLock()
	defer customRulesMu.RUnlock()
	rule, ok := customRules[name]
	return rule, ok
}

// Validation Types

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var b strings.Builder
	for i, err := range e {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Field)
		b.WriteString(": ")
		b.WriteString(err.Message)
	}
	return b.String()
}

// ToArray converts validation errors to array format for frontend
func (e ValidationErrors) ToArray() []map[string]string {
	if len(e) == 0 {
		return nil
	}
	result := make([]map[string]string, len(e))
	for i, err := range e {
		result[i] = map[string]string{"field": strings.ToLower(err.Field), "message": err.Message}
	}
	return result
}

// Struct Tag Cache

// cachedField holds pre-parsed validation metadata for a single struct field.
type cachedField struct {
	name       string
	rules      []string // split validation rules
	hasNested  bool     // true if field is a struct or *struct
	fieldIndex int      // index in struct type
}

// structFieldCache caches parsed field metadata per struct type.
// Key: reflect.Type, Value: []cachedField
var structFieldCache sync.Map

// getOrParseFields returns cached field metadata for the given struct type.
func getOrParseFields(typ reflect.Type) []cachedField {
	if cached, ok := structFieldCache.Load(typ); ok {
		return cached.([]cachedField)
	}

	fields := make([]cachedField, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(tagValidate)

		// Determine if field can have nested validation
		hasNested := false
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			hasNested = true
		}

		if tag == "" && !hasNested {
			continue
		}

		var rules []string
		if tag != "" {
			rules = strings.Split(tag, ruleSep)
		}

		fields = append(fields, cachedField{
			name:       field.Name,
			rules:      rules,
			hasNested:  hasNested,
			fieldIndex: i,
		})
	}

	structFieldCache.Store(typ, fields)
	return fields
}

// Validation Functions

// Validate validates a struct using validation rules from struct field tags.
// It supports built-in rules (required, min, max, email, url, etc.) and custom rules.
// Struct tag parsing is cached per type for performance.
//
// Input:
//   - v: A struct or pointer to struct to validate
//
// Output:
//   - error: Returns ValidationErrors if validation fails, nil otherwise
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `validate:"required,min=3,max=50"`
//	    Email string `validate:"required,email"`
//	    Age   int    `validate:"min=18,max=120"`
//	}
//
//	req := CreateUserRequest{Name: "Jo", Email: "invalid"}
//	if err := orianna.Validate(req); err != nil {
//	    // err contains validation failures
//	}
func Validate(v any) error {
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
	cachedFields := getOrParseFields(typ)

	for _, cf := range cachedFields {
		value := val.Field(cf.fieldIndex)

		// Apply validation rules
		for _, rule := range cf.rules {
			if err := validateRule(cf.name, value, strings.TrimSpace(rule)); err != nil {
				errors = append(errors, *err)
			}
		}

		// Recursive validation for nested struct fields
		if cf.hasNested {
			errors = validateNestedFields(cf, value, errors)
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateNestedFields handles recursive validation of struct or *struct fields.
func validateNestedFields(cf cachedField, value reflect.Value, errors ValidationErrors) ValidationErrors {
	var checkVal reflect.Value
	if value.Kind() == reflect.Ptr && !value.IsNil() {
		checkVal = value.Elem()
	} else if value.Kind() == reflect.Struct {
		checkVal = value
	}

	if !checkVal.IsValid() || checkVal.Kind() != reflect.Struct {
		return errors
	}

	nestedErr := Validate(value.Interface())
	if nestedErr == nil {
		return errors
	}
	nestedErrors, ok := nestedErr.(ValidationErrors)
	if !ok {
		return errors
	}
	for _, ne := range nestedErrors {
		ne.Field = cf.name + "." + ne.Field
		errors = append(errors, ne)
	}
	return errors
}

// ruleHandler is a function that validates a single rule on a field value.
type ruleHandler func(fieldName string, value reflect.Value, param string) *ValidationError

// builtinRules maps rule names to their handler functions.
var builtinRules = map[string]ruleHandler{
	ruleRequired: func(fieldName string, value reflect.Value, _ string) *ValidationError {
		return validateRequired(fieldName, value)
	},
	ruleMin:        withIntParam(validateMin),
	ruleMax:        withIntParam(validateMax),
	ruleLen:        withIntParam(validateLen),
	ruleEmail:      noParam(validateEmail),
	ruleURL:        noParam(validateURL),
	ruleNumeric:    noParam(validateNumeric),
	ruleAlpha:      noParam(validateAlpha),
	"alphanumeric": noParam(validateAlphanumeric),
	"uuid":         noParam(validateUUID),
	ruleOneOf: func(fieldName string, value reflect.Value, param string) *ValidationError {
		if param == "" {
			return nil
		}
		return validateOneOf(fieldName, value, param)
	},
	ruleGt:  comparisonHandler(ruleGt),
	ruleGte: comparisonHandler(ruleGte),
	ruleLt:  comparisonHandler(ruleLt),
	ruleLte: comparisonHandler(ruleLte),
}

// withIntParam wraps a validator that takes an int parameter, handling the parsing.
func withIntParam(fn func(string, reflect.Value, int) *ValidationError) ruleHandler {
	return func(fieldName string, value reflect.Value, param string) *ValidationError {
		if param == "" {
			return nil
		}
		n, err := strconv.Atoi(param)
		if err != nil {
			return nil
		}
		return fn(fieldName, value, n)
	}
}

// noParam wraps a validator that takes no parameter.
func noParam(fn func(string, reflect.Value) *ValidationError) ruleHandler {
	return func(fieldName string, value reflect.Value, _ string) *ValidationError {
		return fn(fieldName, value)
	}
}

// comparisonHandler returns a ruleHandler for comparison rules (gt, gte, lt, lte).
func comparisonHandler(op string) ruleHandler {
	return func(fieldName string, value reflect.Value, param string) *ValidationError {
		if param == "" {
			return nil
		}
		return validateComparison(fieldName, value, param, op)
	}
}

func validateRule(fieldName string, value reflect.Value, rule string) *ValidationError {
	parts := strings.SplitN(rule, paramSep, 2)
	ruleName := parts[0]
	var ruleParam string
	if len(parts) > 1 {
		ruleParam = parts[1]
	}

	// Check for custom validation rules first
	if customRule, ok := GetCustomValidationRule(ruleName); ok {
		return customRule(fieldName, value, ruleParam)
	}

	// Look up built-in rule
	if handler, ok := builtinRules[ruleName]; ok {
		return handler(fieldName, value, ruleParam)
	}

	// Unknown rule name -- likely a typo in the validate tag
	return &ValidationError{
		Field:   fieldName,
		Message: fmt.Sprintf("unknown validation rule: %s", ruleName),
	}
}

// Validation Rules

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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if minVal >= 0 && value.Uint() < uint64(minVal) { // #nosec G115 -- guarded by minVal >= 0
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMinValue, minVal)}
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() < float64(minVal) {
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if maxVal < 0 || value.Uint() > uint64(maxVal) { // #nosec G115 -- guarded by maxVal >= 0
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxValue, maxVal)}
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() > float64(maxVal) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxValue, maxVal)}
		}
	}
	return nil
}

func validateEmail(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String || value.String() == "" {
		return nil
	}
	if !EmailRegex.MatchString(value.String()) {
		return &ValidationError{Field: fieldName, Message: msgInvalidEmail}
	}
	return nil
}

func validateURL(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String || value.String() == "" {
		return nil
	}
	if !URLRegex.MatchString(value.String()) {
		return &ValidationError{Field: fieldName, Message: msgInvalidURL}
	}
	return nil
}

func validateNumeric(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String || value.String() == "" {
		return nil
	}
	if !NumericRegex.MatchString(value.String()) {
		return &ValidationError{Field: fieldName, Message: msgNumericOnly}
	}
	return nil
}

func validateAlpha(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String || value.String() == "" {
		return nil
	}
	if !AlphaRegex.MatchString(value.String()) {
		return &ValidationError{Field: fieldName, Message: msgAlphaOnly}
	}
	return nil
}

// validateAlphanumeric validates that a string contains only alphanumeric characters
func validateAlphanumeric(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String || value.String() == "" {
		return nil
	}
	if !AlphanumericRegex.MatchString(value.String()) {
		return &ValidationError{Field: fieldName, Message: "must contain only alphanumeric characters"}
	}
	return nil
}

// validateUUID validates that a string is a valid UUID
func validateUUID(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String || value.String() == "" {
		return nil
	}
	if !UUIDRegex.MatchString(value.String()) {
		return &ValidationError{Field: fieldName, Message: "must be a valid UUID"}
	}
	return nil
}

// validateOneOf validates that a value is one of the allowed values.
// Supports string, int, uint, and float types.
// Allowed values are space-separated in the param string.
//
// Usage: `validate:"oneof=active inactive pending"` (string)
// Usage: `validate:"oneof=1 2 3"` (int)
func validateOneOf(fieldName string, value reflect.Value, param string) *ValidationError {
	allowed := strings.Fields(param)

	switch value.Kind() {
	case reflect.String:
		if validateOneOfString(value.String(), allowed) {
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if validateOneOfInt(value.Int(), allowed) {
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if validateOneOfUint(value.Uint(), allowed) {
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if validateOneOfFloat(value.Float(), allowed) {
			return nil
		}
	default:
		return nil
	}
	return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgOneOf, param)}
}

// validateOneOfString checks if a string value is in the allowed list
func validateOneOfString(str string, allowed []string) bool {
	if str == "" {
		return true
	}
	for _, a := range allowed {
		if str == a {
			return true
		}
	}
	return false
}

// validateOneOfInt checks if an int value is in the allowed list
func validateOneOfInt(val int64, allowed []string) bool {
	for _, a := range allowed {
		if n, err := strconv.ParseInt(a, 10, 64); err == nil && val == n {
			return true
		}
	}
	return false
}

// validateOneOfUint checks if a uint value is in the allowed list
func validateOneOfUint(val uint64, allowed []string) bool {
	for _, a := range allowed {
		if n, err := strconv.ParseUint(a, 10, 64); err == nil && val == n {
			return true
		}
	}
	return false
}

// validateOneOfFloat checks if a float value is in the allowed list
func validateOneOfFloat(val float64, allowed []string) bool {
	for _, a := range allowed {
		if n, err := strconv.ParseFloat(a, 64); err == nil && val == n {
			return true
		}
	}
	return false
}

// validateLen validates that a value has an exact length.
// Supports strings (character length), slices, arrays, and maps (element count).
//
// Usage: `validate:"len=10"`
func validateLen(fieldName string, value reflect.Value, length int) *ValidationError {
	switch value.Kind() {
	case reflect.String:
		if len(value.String()) != length {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgLen, length)}
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if value.Len() != length {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgLen, length)}
		}
	}
	return nil
}

// validateComparison validates numeric comparison rules: gt, gte, lt, lte.
func validateComparison(fieldName string, value reflect.Value, param string, op string) *ValidationError {
	var v, threshold float64
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil
		}
		v, threshold = float64(value.Int()), t
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil
		}
		v, threshold = float64(value.Uint()), t
	case reflect.Float32, reflect.Float64:
		t, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil
		}
		v, threshold = value.Float(), t
	default:
		return nil
	}

	var pass bool
	switch op {
	case ruleGt:
		pass = v > threshold
	case ruleGte:
		pass = v >= threshold
	case ruleLt:
		pass = v < threshold
	case ruleLte:
		pass = v <= threshold
	}

	if !pass {
		msgMap := map[string]string{
			ruleGt: msgGt, ruleGte: msgGte, ruleLt: msgLt, ruleLte: msgLte,
		}
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMap[op], param)}
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
