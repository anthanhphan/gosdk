// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Validator is a configurable validation instance.
type Validator struct {
	fieldNameTag     string
	stopOnFirstError bool
}

// ValidatorOption is a functional option for configuring a Validator.
type ValidatorOption func(*Validator)

// New creates a new Validator instance with the given options.
func New(opts ...ValidatorOption) *Validator {
	v := &Validator{}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// WithFieldNameTag uses the specified struct tag for field names in errors.
//
//	v := validator.New(validator.WithFieldNameTag("json"))
func WithFieldNameTag(tag string) ValidatorOption {
	return func(v *Validator) { v.fieldNameTag = tag }
}

// WithStopOnFirstError stops validation after the first error.
func WithStopOnFirstError(stop bool) ValidatorOption {
	return func(v *Validator) { v.stopOnFirstError = stop }
}

// ============================================================================
// Instance Field Cache
// ============================================================================

type instanceCachedField struct {
	cachedField
	displayName string
}

type instanceFieldCacheKey struct {
	typ reflect.Type
	tag string
}

// instanceFieldCache caches fields with resolved display names.
// Key: instanceFieldCacheKey (type-safe), Value: []instanceCachedField
var instanceFieldCache sync.Map

// defaultFieldCache caches fields for the default (no-tag) validator.
// Key: reflect.Type, Value: []instanceCachedField
var defaultFieldCache sync.Map

// resolveFields returns cached fields with displayName resolved from the given tag.
func (v *Validator) resolveFields(typ reflect.Type) []instanceCachedField {
	if v.fieldNameTag == "" {
		// Fast path: no tag override — use dedicated cache to avoid allocation
		if cached, ok := defaultFieldCache.Load(typ); ok {
			return cached.([]instanceCachedField)
		}
		fields := getOrParseFields(typ)
		result := make([]instanceCachedField, len(fields))
		for i := range fields {
			result[i] = instanceCachedField{cachedField: fields[i], displayName: fields[i].name}
		}
		defaultFieldCache.Store(typ, result)
		return result
	}

	cacheKey := instanceFieldCacheKey{typ: typ, tag: v.fieldNameTag}
	if cached, ok := instanceFieldCache.Load(cacheKey); ok {
		return cached.([]instanceCachedField)
	}

	baseFields := getOrParseFields(typ)
	result := make([]instanceCachedField, len(baseFields))
	for i := range baseFields {
		displayName := baseFields[i].name
		structField := typ.Field(baseFields[i].fieldIndex)
		if tagValue := structField.Tag.Get(v.fieldNameTag); tagValue != "" {
			if idx := strings.Index(tagValue, ","); idx != -1 {
				tagValue = tagValue[:idx]
			}
			if tagValue != "" && tagValue != "-" {
				displayName = tagValue
			}
		}
		result[i] = instanceCachedField{cachedField: baseFields[i], displayName: displayName}
	}

	instanceFieldCache.Store(cacheKey, result)
	return result
}

// ============================================================================
// Validation
// ============================================================================

// ValidateStruct validates a struct using the Validator's configuration.
// After the first call for each struct type, all subsequent calls have
// zero string parsing and zero map lookups.
func (v *Validator) ValidateStruct(s any) error {
	val := reflect.ValueOf(s)
	kind := val.Kind()
	if kind == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("%s", msgNilPointer)
		}
		val = val.Elem()
		kind = val.Kind()
	}
	if kind != reflect.Struct {
		return fmt.Errorf(msgInvalidType, s)
	}

	var errors ValidationErrors
	fields := v.resolveFields(val.Type())

	for i := range fields {
		cf := &fields[i]
		value := val.Field(cf.fieldIndex)

		if fieldErrs := v.validateField(cf, value); len(fieldErrs) > 0 {
			errors = append(errors, fieldErrs...)
			if v.stopOnFirstError {
				return errors
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateField validates a single field and returns any errors.
func (v *Validator) validateField(cf *instanceCachedField, value reflect.Value) ValidationErrors {
	var errors ValidationErrors

	if cf.hasDive {
		// Pre-dive rules on the field itself
		for j := range cf.rules {
			if err := cf.rules[j].handler(cf.displayName, value, ""); err != nil {
				errors = append(errors, *err)
			}
		}
		// Dive: validate each element
		if len(cf.diveRules) > 0 {
			errors = append(errors, validateDiveResolved(cf.displayName, value, cf.diveRules)...)
		}
		return errors
	}

	// Apply pre-resolved rules
	for j := range cf.rules {
		if err := cf.rules[j].handler(cf.displayName, value, ""); err != nil {
			errors = append(errors, *err)
		}
	}

	// Recursive validation for nested struct fields
	if cf.hasNested {
		errors = append(errors, v.validateNestedStruct(cf, value)...)
	}

	return errors
}

// validateNestedStruct handles recursive validation.
func (v *Validator) validateNestedStruct(cf *instanceCachedField, value reflect.Value) ValidationErrors {
	if cf.isPtr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil
	}

	nestedErr := v.ValidateStruct(value.Interface())
	if nestedErr == nil {
		return nil
	}
	nestedErrors, ok := nestedErr.(ValidationErrors)
	if !ok {
		return nil
	}

	prefix := cf.displayName + "."
	result := make(ValidationErrors, len(nestedErrors))
	for i, ne := range nestedErrors {
		ne.Field = prefix + ne.Field
		result[i] = ne
	}
	return result
}

// validateDiveResolved validates each element using pre-resolved rules.
func validateDiveResolved(fieldName string, value reflect.Value, rules []resolvedRule) ValidationErrors {
	var errors ValidationErrors
	kind := value.Kind()

	switch kind {
	case reflect.Slice, reflect.Array:
		if value.IsNil() {
			return nil
		}
		n := value.Len()
		for i := 0; i < n; i++ {
			elemName := fmt.Sprintf("%s[%d]", fieldName, i)
			errors = applyResolvedRules(elemName, value.Index(i), rules, errors)
		}

	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		for _, key := range value.MapKeys() {
			elemName := fmt.Sprintf("%s[%v]", fieldName, key.Interface())
			errors = applyResolvedRules(elemName, value.MapIndex(key), rules, errors)
		}
	}

	return errors
}

// applyResolvedRules applies pre-resolved rules to a single element.
func applyResolvedRules(elemName string, elem reflect.Value, rules []resolvedRule, errors ValidationErrors) ValidationErrors {
	if elem.Kind() == reflect.Interface {
		elem = elem.Elem()
	}

	for i := range rules {
		if err := rules[i].handler(elemName, elem, ""); err != nil {
			errors = append(errors, *err)
		}
	}

	// Recursive struct validation
	actual := elem
	if actual.Kind() == reflect.Ptr && !actual.IsNil() {
		actual = actual.Elem()
	}
	if actual.Kind() == reflect.Struct {
		if nestedErr := Validate(actual.Interface()); nestedErr != nil {
			if nestedErrors, ok := nestedErr.(ValidationErrors); ok {
				prefix := elemName + "."
				for _, ne := range nestedErrors {
					ne.Field = prefix + ne.Field
					errors = append(errors, ne)
				}
			}
		}
	}

	return errors
}
