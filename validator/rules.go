// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// Rule Names
// ============================================================================

const (
	// Core rules
	ruleRequired = "required"
	ruleMin      = "min"
	ruleMax      = "max"
	ruleLen      = "len"
	ruleOneOf    = "oneof"

	// Comparison rules
	ruleGt  = "gt"
	ruleGte = "gte"
	ruleLt  = "lt"
	ruleLte = "lte"

	// String rules
	ruleContains   = "contains"
	ruleStartsWith = "startswith"
	ruleEndsWith   = "endswith"
	ruleLowercase  = "lowercase"
	ruleUppercase  = "uppercase"
	ruleExcludes   = "excludes"

	// Format rules
	ruleEmail        = "email"
	ruleURL          = "url"
	ruleNumeric      = "numeric"
	ruleAlpha        = "alpha"
	ruleAlphanumeric = "alphanumeric"
	ruleUUID         = "uuid"
	ruleHexColor     = "hexcolor"
	ruleDatetime     = "datetime"

	// Network rules
	ruleIP   = "ip"
	ruleIPv4 = "ipv4"
	ruleIPv6 = "ipv6"

	// Collection rules
	ruleNotEmpty = "notempty"
	ruleUnique   = "unique"
	ruleDive     = "dive"
)

// ============================================================================
// Error Messages
// ============================================================================

const (
	msgRequired      = "is required"
	msgMinCharacters = "must be at least %d characters"
	msgMinItems      = "must contain at least %d items"
	msgMinValue      = "must be at least %d"
	msgMaxCharacters = "must be at most %d characters"
	msgMaxItems      = "must contain at most %d items"
	msgMaxValue      = "must be at most %d"
	msgOneOf         = "must be one of: %s"
	msgLen           = "must have exact length of %d"
	msgGt            = "must be greater than %s"
	msgGte           = "must be greater than or equal to %s"
	msgLt            = "must be less than %s"
	msgLte           = "must be less than or equal to %s"
	msgContains      = "must contain '%s'"
	msgStartsWith    = "must start with '%s'"
	msgEndsWith      = "must end with '%s'"
	msgLowercase     = "must be lowercase"
	msgUppercase     = "must be uppercase"
	msgExcludes      = "must not contain '%s'"
	msgInvalidEmail  = "must be a valid email address"
	msgInvalidURL    = "must be a valid URL"
	msgNumericOnly   = "must contain only numeric characters"
	msgAlphaOnly     = "must contain only alphabetic characters"
	msgAlphanumeric  = "must contain only alphanumeric characters"
	msgInvalidUUID   = "must be a valid UUID"
	msgInvalidHex    = "must be a valid hex color (e.g., #FFF or #FFFFFF)"
	msgInvalidDate   = "must match datetime format '%s'"
	msgInvalidIP     = "must be a valid IP address"
	msgInvalidIPv4   = "must be a valid IPv4 address"
	msgInvalidIPv6   = "must be a valid IPv6 address"
	msgNotEmpty      = "must not be empty"
	msgUnique        = "must contain unique values"
)

// ============================================================================
// Rule Registry — rule name → factory that returns a pre-resolved handler
//
// Factory pattern: each factory receives the raw param string and returns
// a ruleHandler with ALL parsing done upfront. No parsing at validation time.
// ============================================================================

// ruleFactory creates a ruleHandler with the param pre-parsed.
type ruleFactory func(param string) ruleHandler

// builtinFactories maps rule names to their factory functions.
// Each factory pre-parses the param and returns a closure that captures
// the parsed value — zero parsing at validation time.
var builtinFactories = map[string]ruleFactory{
	// Core
	ruleRequired: func(_ string) ruleHandler {
		return func(fieldName string, value reflect.Value, _ string) *ValidationError {
			return validateRequired(fieldName, value)
		}
	},
	ruleMin: intFactory(validateMin),
	ruleMax: intFactory(validateMax),
	ruleLen: intFactory(validateLen),
	ruleOneOf: func(param string) ruleHandler {
		allowed := strings.Fields(param) // pre-split once
		return func(fieldName string, value reflect.Value, _ string) *ValidationError {
			return validateOneOf(fieldName, value, param, allowed)
		}
	},

	// Comparison — pre-parse float threshold
	ruleGt:  floatComparisonFactory(ruleGt),
	ruleGte: floatComparisonFactory(ruleGte),
	ruleLt:  floatComparisonFactory(ruleLt),
	ruleLte: floatComparisonFactory(ruleLte),

	// String
	ruleContains:   paramPassthrough(validateContains),
	ruleStartsWith: paramPassthrough(validateStartsWith),
	ruleEndsWith:   paramPassthrough(validateEndsWith),
	ruleLowercase:  noParamFactory(validateLowercase),
	ruleUppercase:  noParamFactory(validateUppercase),
	ruleExcludes:   paramPassthrough(validateExcludes),

	// Format
	ruleEmail:        noParamFactory(validateEmail),
	ruleURL:          noParamFactory(validateURL),
	ruleNumeric:      noParamFactory(validateNumeric),
	ruleAlpha:        noParamFactory(validateAlpha),
	ruleAlphanumeric: noParamFactory(validateAlphanumeric),
	ruleUUID:         noParamFactory(validateUUID),
	ruleHexColor:     noParamFactory(validateHexColor),
	ruleDatetime: func(param string) ruleHandler {
		layout := param
		if layout == "" {
			layout = time.RFC3339
		}
		return func(fieldName string, value reflect.Value, _ string) *ValidationError {
			return validateDatetime(fieldName, value, layout)
		}
	},

	// Network
	ruleIP:   noParamFactory(validateIP),
	ruleIPv4: noParamFactory(validateIPv4),
	ruleIPv6: noParamFactory(validateIPv6),

	// Collection
	ruleNotEmpty: noParamFactory(validateNotEmpty),
	ruleUnique:   noParamFactory(validateUnique),
}

// ============================================================================
// Factory Helpers
// ============================================================================

// intFactory creates a factory that pre-parses an int param.
// At validation time, no strconv.Atoi call is needed.
func intFactory(fn func(string, reflect.Value, int) *ValidationError) ruleFactory {
	return func(param string) ruleHandler {
		if param == "" {
			return noop
		}
		n, err := strconv.Atoi(param)
		if err != nil {
			return noop
		}
		return func(fieldName string, value reflect.Value, _ string) *ValidationError {
			return fn(fieldName, value, n)
		}
	}
}

// floatComparisonFactory creates a factory that pre-parses a float threshold.
func floatComparisonFactory(op string) ruleFactory {
	msg := comparisonMessages[op]
	return func(param string) ruleHandler {
		if param == "" {
			return noop
		}
		threshold, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return noop
		}
		return func(fieldName string, value reflect.Value, _ string) *ValidationError {
			return validateComparisonPreParsed(fieldName, value, threshold, op, msg, param)
		}
	}
}

// noParamFactory wraps a no-param validator into a factory.
func noParamFactory(fn func(string, reflect.Value) *ValidationError) ruleFactory {
	// Create handler once, reuse for all instances
	h := ruleHandler(func(fieldName string, value reflect.Value, _ string) *ValidationError {
		return fn(fieldName, value)
	})
	return func(_ string) ruleHandler { return h }
}

// paramPassthrough wraps a string-param validator into a factory.
func paramPassthrough(fn func(string, reflect.Value, string) *ValidationError) ruleFactory {
	return func(param string) ruleHandler {
		return func(fieldName string, value reflect.Value, _ string) *ValidationError {
			return fn(fieldName, value, param)
		}
	}
}

// noop is a handler that does nothing.
func noop(_ string, _ reflect.Value, _ string) *ValidationError { return nil }

// ============================================================================
// Core Rules — required, min, max, len, oneof
// ============================================================================

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
		if minVal >= 0 && value.Uint() < uint64(minVal) {
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
		if maxVal < 0 || value.Uint() > uint64(maxVal) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxValue, maxVal)}
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() > float64(maxVal) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgMaxValue, maxVal)}
		}
	}
	return nil
}

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

// validateOneOf validates that a value is one of the pre-split allowed values.
func validateOneOf(fieldName string, value reflect.Value, rawParam string, allowed []string) *ValidationError {
	if matchesOneOf(value, allowed) {
		return nil
	}
	return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgOneOf, rawParam)}
}

// matchesOneOf checks if value matches any allowed value.
func matchesOneOf(value reflect.Value, allowed []string) bool {
	switch value.Kind() {
	case reflect.String:
		s := value.String()
		if s == "" {
			return true
		}
		for _, a := range allowed {
			if s == a {
				return true
			}
		}
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return numericInAllowed(value, allowed)
	default:
		return true
	}
}

// numericInAllowed checks if a numeric value matches any of the allowed string values.
func numericInAllowed(value reflect.Value, allowed []string) bool {
	v := toFloat64(value)
	for _, a := range allowed {
		if n, err := strconv.ParseFloat(a, 64); err == nil && v == n {
			return true
		}
	}
	return false
}

// toFloat64 converts an int/uint/float reflect.Value to float64.
func toFloat64(value reflect.Value) float64 {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return value.Float()
	}
	return 0
}

// ============================================================================
// Comparison Rules — gt, gte, lt, lte (pre-parsed threshold)
// ============================================================================

var comparisonMessages = map[string]string{
	ruleGt: msgGt, ruleGte: msgGte, ruleLt: msgLt, ruleLte: msgLte,
}

// validateComparisonPreParsed validates with a pre-parsed float threshold.
// No strconv.ParseFloat at validation time.
func validateComparisonPreParsed(fieldName string, value reflect.Value, threshold float64, op, msg, param string) *ValidationError {
	var v float64
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		v = value.Float()
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
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msg, param)}
	}
	return nil
}

// ============================================================================
// String Rules — extract value.String() once per call
// ============================================================================

func validateContains(fieldName string, value reflect.Value, param string) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	if !strings.Contains(s, param) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgContains, param)}
	}
	return nil
}

func validateStartsWith(fieldName string, value reflect.Value, param string) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	if !strings.HasPrefix(s, param) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgStartsWith, param)}
	}
	return nil
}

func validateEndsWith(fieldName string, value reflect.Value, param string) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	if !strings.HasSuffix(s, param) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgEndsWith, param)}
	}
	return nil
}

func validateLowercase(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && s != strings.ToLower(s) {
		return &ValidationError{Field: fieldName, Message: msgLowercase}
	}
	return nil
}

func validateUppercase(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && s != strings.ToUpper(s) {
		return &ValidationError{Field: fieldName, Message: msgUppercase}
	}
	return nil
}

func validateExcludes(fieldName string, value reflect.Value, param string) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && strings.Contains(s, param) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgExcludes, param)}
	}
	return nil
}

// ============================================================================
// Format Rules — extract value.String() once per call
// ============================================================================

func validateEmail(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && !EmailRegex.MatchString(s) {
		return &ValidationError{Field: fieldName, Message: msgInvalidEmail}
	}
	return nil
}

func validateURL(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && !URLRegex.MatchString(s) {
		return &ValidationError{Field: fieldName, Message: msgInvalidURL}
	}
	return nil
}

func validateNumeric(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return &ValidationError{Field: fieldName, Message: msgNumericOnly}
		}
	}
	return nil
}

func validateAlpha(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return &ValidationError{Field: fieldName, Message: msgAlphaOnly}
		}
	}
	return nil
}

func validateAlphanumeric(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return &ValidationError{Field: fieldName, Message: msgAlphanumeric}
		}
	}
	return nil
}

func validateUUID(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && !UUIDRegex.MatchString(s) {
		return &ValidationError{Field: fieldName, Message: msgInvalidUUID}
	}
	return nil
}

func validateHexColor(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && !HexColorRegex.MatchString(s) {
		return &ValidationError{Field: fieldName, Message: msgInvalidHex}
	}
	return nil
}

func validateDatetime(fieldName string, value reflect.Value, layout string) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	if _, err := time.Parse(layout, s); err != nil {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf(msgInvalidDate, layout)}
	}
	return nil
}

// ============================================================================
// Network Rules
// ============================================================================

func validateIP(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s != "" && net.ParseIP(s) == nil {
		return &ValidationError{Field: fieldName, Message: msgInvalidIP}
	}
	return nil
}

func validateIPv4(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() == nil {
		return &ValidationError{Field: fieldName, Message: msgInvalidIPv4}
	}
	return nil
}

func validateIPv6(fieldName string, value reflect.Value) *ValidationError {
	if value.Kind() != reflect.String {
		return nil
	}
	s := value.String()
	if s == "" {
		return nil
	}
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() != nil {
		return &ValidationError{Field: fieldName, Message: msgInvalidIPv6}
	}
	return nil
}

// ============================================================================
// Collection Rules — notempty, unique
// ============================================================================

func validateNotEmpty(fieldName string, value reflect.Value) *ValidationError {
	switch value.Kind() {
	case reflect.Slice, reflect.Map:
		if value.IsNil() || value.Len() == 0 {
			return &ValidationError{Field: fieldName, Message: msgNotEmpty}
		}
	case reflect.Array:
		if value.Len() == 0 {
			return &ValidationError{Field: fieldName, Message: msgNotEmpty}
		}
	case reflect.String:
		if value.String() == "" {
			return &ValidationError{Field: fieldName, Message: msgNotEmpty}
		}
	}
	return nil
}

func validateUnique(fieldName string, value reflect.Value) *ValidationError {
	kind := value.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil
	}
	if value.IsNil() || value.Len() <= 1 {
		return nil
	}

	seen := make(map[any]struct{}, value.Len())
	for i := 0; i < value.Len(); i++ {
		key := value.Index(i).Interface()
		if _, exists := seen[key]; exists {
			return &ValidationError{Field: fieldName, Message: msgUnique}
		}
		seen[key] = struct{}{}
	}
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

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
