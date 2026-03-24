package validator

import (
	"strings"
	"testing"
)

// ============================================================================
// Validator Instance — WithFieldNameTag
// ============================================================================

func TestWithFieldNameTag_JSON(t *testing.T) {
	type User struct {
		FirstName string `json:"first_name" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(User{})
	if err == nil {
		t.Fatal("should fail")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "first_name") {
		t.Errorf("should use json tag, got: %s", errStr)
	}
	if !strings.Contains(errStr, "last_name") {
		t.Errorf("should use json tag, got: %s", errStr)
	}
}

func TestWithFieldNameTag_YAML(t *testing.T) {
	type Config struct {
		ServerPort int    `yaml:"server_port" validate:"required"`
		ServerName string `yaml:"server_name" validate:"required"`
	}

	v := New(WithFieldNameTag("yaml"))
	err := v.ValidateStruct(Config{})
	if err == nil {
		t.Fatal("should fail")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "server_port") {
		t.Errorf("should use yaml tag, got: %s", errStr)
	}
}

func TestWithFieldNameTag_Omitempty(t *testing.T) {
	type Req struct {
		Name string `json:"name,omitempty" validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(Req{})
	if err == nil {
		t.Fatal("should fail")
	}

	errStr := err.Error()
	if strings.Contains(errStr, "omitempty") {
		t.Error("should strip omitempty from json tag")
	}
	if !strings.Contains(errStr, "name:") {
		t.Errorf("should use 'name', got: %s", errStr)
	}
}

func TestWithFieldNameTag_Dash(t *testing.T) {
	type Item struct {
		ID int `json:"-" validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(Item{})
	if err == nil {
		t.Fatal("should fail")
	}

	// json:"-" should fall back to Go field name
	if !strings.Contains(err.Error(), "ID") {
		t.Errorf("dash should fallback to Go name, got: %s", err.Error())
	}
}

func TestWithFieldNameTag_EmptyTag(t *testing.T) {
	type Item struct {
		Value string `json:"" validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(Item{})
	if err == nil {
		t.Fatal("should fail")
	}

	// Empty json tag should fall back to Go field name
	if !strings.Contains(err.Error(), "Value") {
		t.Errorf("empty tag should fallback, got: %s", err.Error())
	}
}

func TestWithFieldNameTag_NoJsonTag(t *testing.T) {
	type Item struct {
		Data string `validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(Item{})
	if err == nil {
		t.Fatal("should fail")
	}

	// No json tag should use Go field name
	if !strings.Contains(err.Error(), "Data") {
		t.Errorf("missing tag should use Go name, got: %s", err.Error())
	}
}

func TestWithFieldNameTag_Nested(t *testing.T) {
	type Address struct {
		City string `json:"city" validate:"required"`
	}
	type User struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(User{Name: "John"})
	if err == nil {
		t.Fatal("should fail for missing city")
	}

	if !strings.Contains(err.Error(), "address.city") {
		t.Errorf("should show address.city, got: %s", err.Error())
	}
}

func TestWithFieldNameTag_Cache(t *testing.T) {
	type TagCache struct {
		Name string `json:"name" validate:"required"`
	}

	v := New(WithFieldNameTag("json"))
	// First call
	_ = v.ValidateStruct(TagCache{})
	// Second call should hit cache
	err := v.ValidateStruct(TagCache{})
	if err == nil {
		t.Fatal("should still fail on second call")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("cached call should use json tag, got: %s", err.Error())
	}
}

// ============================================================================
// Validator Instance — WithStopOnFirstError
// ============================================================================

func TestWithStopOnFirstError(t *testing.T) {
	type Form struct {
		Name  string `validate:"required"`
		Email string `validate:"required"`
		Age   int    `validate:"required"`
	}

	// Without stop — collects all errors
	v1 := New()
	err := v1.ValidateStruct(Form{})
	if err == nil {
		t.Fatal("should fail")
	}
	allErrs := err.(ValidationErrors)

	// With stop — only first error
	v2 := New(WithStopOnFirstError(true))
	err = v2.ValidateStruct(Form{})
	if err == nil {
		t.Fatal("should fail")
	}
	firstErr := err.(ValidationErrors)

	if len(firstErr) != 1 {
		t.Errorf("stopOnFirst: expected 1 error, got %d", len(firstErr))
	}
	if len(allErrs) <= 1 {
		t.Errorf("no stop: expected multiple errors, got %d", len(allErrs))
	}
}

func TestWithStopOnFirstError_Dive(t *testing.T) {
	type S struct {
		Tags []string `validate:"dive,required"`
	}

	v := New(WithStopOnFirstError(true))
	err := v.ValidateStruct(S{Tags: []string{"", ""}})
	if err == nil {
		t.Fatal("should fail")
	}
	// Dive collects all element errors as a batch, then stopOnFirst
	// prevents further field processing
	errs := err.(ValidationErrors)
	if len(errs) == 0 {
		t.Error("should have errors")
	}
}

func TestWithStopOnFirstError_PreDive(t *testing.T) {
	type S struct {
		Tags []string `validate:"notempty,dive,required"`
	}

	v := New(WithStopOnFirstError(true))
	err := v.ValidateStruct(S{Tags: []string{}})
	if err == nil {
		t.Fatal("should fail")
	}
	errs := err.(ValidationErrors)
	if len(errs) != 1 {
		t.Errorf("stopOnFirst pre-dive: expected 1 error, got %d", len(errs))
	}
}

func TestWithStopOnFirstError_Nested(t *testing.T) {
	type Inner struct {
		A string `validate:"required"`
		B string `validate:"required"`
	}
	type Outer struct {
		Name  string `validate:"required"`
		Inner Inner
	}

	v := New(WithStopOnFirstError(true))
	err := v.ValidateStruct(Outer{})
	if err == nil {
		t.Fatal("should fail")
	}
	errs := err.(ValidationErrors)
	if len(errs) != 1 {
		t.Errorf("stopOnFirst nested: expected 1 error, got %d: %v", len(errs), errs)
	}
}

// ============================================================================
// ValidateStruct — edge cases
// ============================================================================

func TestValidateStruct_NilPointer(t *testing.T) {
	v := New()
	err := v.ValidateStruct((*struct{})(nil))
	if err == nil {
		t.Error("nil pointer should fail")
	}
}

func TestValidateStruct_NonStruct(t *testing.T) {
	v := New()
	err := v.ValidateStruct("not a struct")
	if err == nil {
		t.Error("non-struct should fail")
	}
}

func TestValidateStruct_Valid(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	v := New()
	if err := v.ValidateStruct(S{Name: "ok"}); err != nil {
		t.Errorf("valid should pass: %v", err)
	}
}

func TestValidateStruct_Pointer(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	v := New()
	s := &S{Name: "ok"}
	if err := v.ValidateStruct(s); err != nil {
		t.Errorf("pointer should pass: %v", err)
	}
}

func TestValidateStruct_DefaultFieldCache(t *testing.T) {
	type DefaultCacheS struct {
		X string `validate:"required"`
	}
	v := New() // no tag
	// First call
	_ = v.ValidateStruct(DefaultCacheS{})
	// Second call should hit defaultFieldCache
	err := v.ValidateStruct(DefaultCacheS{})
	if err == nil {
		t.Error("should still fail")
	}
}

// ============================================================================
// Dive with nested structs
// ============================================================================

func TestDive_NestedStructsWithRules(t *testing.T) {
	type Item struct {
		Name string `validate:"required"`
	}
	type S struct {
		Items []Item `validate:"dive,required"`
	}

	// Valid nested
	if err := Validate(S{Items: []Item{{Name: "ok"}}}); err != nil {
		t.Errorf("dive nested should pass: %v", err)
	}

	// Invalid nested — struct validation happens in applyResolvedRules
	err := Validate(S{Items: []Item{{Name: ""}}})
	if err == nil {
		t.Fatal("dive nested should fail")
	}
	if !strings.Contains(err.Error(), "Items[0].Name") {
		t.Errorf("should show Items[0].Name, got: %v", err)
	}
}

func TestDive_WithFieldNameTag(t *testing.T) {
	type S struct {
		Servers []string `json:"servers" validate:"dive,required"`
	}
	v := New(WithFieldNameTag("json"))
	err := v.ValidateStruct(S{Servers: []string{"ok", ""}})
	if err == nil {
		t.Fatal("should fail")
	}
	if !strings.Contains(err.Error(), "servers[1]") {
		t.Errorf("should show servers[1], got: %v", err)
	}
}

// ============================================================================
// Default Validate() backward compat
// ============================================================================

func TestDefaultValidate_BackwardCompat(t *testing.T) {
	type Simple struct {
		Name string `validate:"required"`
	}
	if err := Validate(Simple{Name: "ok"}); err != nil {
		t.Errorf("global should pass: %v", err)
	}
	if err := Validate(Simple{}); err == nil {
		t.Error("global should fail")
	}
}
