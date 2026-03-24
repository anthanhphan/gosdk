// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// ============================================================================
// Constants
// ============================================================================

const (
	tagValidate = "validate"
	ruleSep     = ","
	paramSep    = "="
)

const (
	msgNilPointer  = "cannot validate nil pointer"
	msgInvalidType = "validate requires a struct or pointer to struct, got %T"
)

// ============================================================================
// Regex Patterns (exported for reuse)
// ============================================================================

var (
	EmailRegex        = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	URLRegex          = regexp.MustCompile(`^https?://[^\s]+$`)
	AlphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	UUIDRegex         = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	HexColorRegex     = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
)

// ============================================================================
// Custom Validation
// ============================================================================

// CustomValidationRule defines a custom validation rule function.
type CustomValidationRule func(fieldName string, value reflect.Value, param string) *ValidationError

var (
	customRules   = make(map[string]CustomValidationRule)
	customRulesMu sync.RWMutex
)

// RegisterValidationRule registers a custom validation rule.
// This function is thread-safe.
func RegisterValidationRule(name string, rule CustomValidationRule) {
	if name != "" && rule != nil {
		customRulesMu.Lock()
		customRules[name] = rule
		customRulesMu.Unlock()
	}
}

// getCustomRule retrieves a registered custom validation rule (thread-safe).
func getCustomRule(name string) (CustomValidationRule, bool) {
	customRulesMu.RLock()
	defer customRulesMu.RUnlock()
	rule, ok := customRules[name]
	return rule, ok
}

// ============================================================================
// Validation Types
// ============================================================================

// ValidationError represents a validation error with field and message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationErrors holds multiple validation errors.
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

// ToArray converts validation errors to array format for frontend.
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

// ============================================================================
// Pre-resolved Rule Cache
// ============================================================================

// ruleHandler is a function that validates a single rule on a field value.
type ruleHandler func(fieldName string, value reflect.Value, param string) *ValidationError

// resolvedRule is a pre-resolved validation rule: handler already has param baked in.
// Zero parsing at validation time.
type resolvedRule struct {
	handler ruleHandler
}

// cachedField holds all pre-computed validation metadata for a single struct field.
type cachedField struct {
	name       string         // Go struct field name
	fieldIndex int            // index in struct type
	fieldKind  reflect.Kind   // cached kind
	isPtr      bool           // field is a pointer
	hasNested  bool           // field is a struct or *struct
	hasDive    bool           // has "dive" rule
	rules      []resolvedRule // pre-resolved rules (before dive if dive exists)
	diveRules  []resolvedRule // pre-resolved rules after dive (nil if no dive)
}

// structFieldCache caches parsed field metadata per struct type.
var structFieldCache sync.Map

// resolveRuleHandler resolves a raw rule string into a resolvedRule.
// All param parsing happens here — at cache time, not at validation time.
func resolveRuleHandler(raw string) resolvedRule {
	parts := strings.SplitN(raw, paramSep, 2)
	name := parts[0]
	param := ""
	if len(parts) > 1 {
		param = parts[1]
	}

	// Check custom rules first
	if customFn, ok := getCustomRule(name); ok {
		p := param // capture
		return resolvedRule{
			handler: func(fieldName string, value reflect.Value, _ string) *ValidationError {
				return customFn(fieldName, value, p)
			},
		}
	}

	// Lookup built-in factory — creates handler with param pre-parsed
	if factory, ok := builtinFactories[name]; ok {
		return resolvedRule{handler: factory(param)}
	}

	// Unknown rule
	msg := "unknown validation rule: " + name
	return resolvedRule{
		handler: func(fieldName string, _ reflect.Value, _ string) *ValidationError {
			return &ValidationError{Field: fieldName, Message: msg}
		},
	}
}

// getOrParseFields returns cached field metadata for the given struct type.
func getOrParseFields(typ reflect.Type) []cachedField {
	if cached, ok := structFieldCache.Load(typ); ok {
		return cached.([]cachedField)
	}

	n := typ.NumField()
	fields := make([]cachedField, 0, n)
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(tagValidate)

		ft := field.Type
		kind := ft.Kind()
		isPtr := kind == reflect.Ptr
		if isPtr {
			ft = ft.Elem()
		}
		hasNested := ft.Kind() == reflect.Struct

		if tag == "" && !hasNested {
			continue
		}

		cf := cachedField{
			name:       field.Name,
			fieldIndex: i,
			fieldKind:  kind,
			isPtr:      isPtr,
			hasNested:  hasNested,
		}

		if tag != "" {
			rawRules := strings.Split(tag, ruleSep)

			// Find dive index
			diveAt := -1
			for j, r := range rawRules {
				if strings.TrimSpace(r) == ruleDive {
					diveAt = j
					break
				}
			}

			if diveAt >= 0 {
				cf.hasDive = true
				// Pre-resolve rules before dive
				for _, r := range rawRules[:diveAt] {
					cf.rules = append(cf.rules, resolveRuleHandler(strings.TrimSpace(r)))
				}
				// Pre-resolve dive rules
				for _, r := range rawRules[diveAt+1:] {
					cf.diveRules = append(cf.diveRules, resolveRuleHandler(strings.TrimSpace(r)))
				}
			} else {
				cf.rules = make([]resolvedRule, 0, len(rawRules))
				for _, r := range rawRules {
					cf.rules = append(cf.rules, resolveRuleHandler(strings.TrimSpace(r)))
				}
			}
		}

		fields = append(fields, cf)
	}

	structFieldCache.Store(typ, fields)
	return fields
}

// ============================================================================
// Validation Engine
// ============================================================================

// defaultValidator is the package-level default Validator instance.
var defaultValidator = New()

// Validate validates a struct using validation rules from struct field tags.
// After the first call for each struct type, all subsequent calls have
// zero string parsing and zero map lookups — only pre-resolved function calls.
func Validate(v any) error {
	return defaultValidator.ValidateStruct(v)
}
