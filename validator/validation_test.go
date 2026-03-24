package validator

import (
	"reflect"
	"strings"
	"testing"
)

// ============================================================================
// ValidationError / ValidationErrors
// ============================================================================

func TestValidationError_Error(t *testing.T) {
	e := &ValidationError{Field: "Name", Message: "is required"}
	if got := e.Error(); got != "Name: is required" {
		t.Errorf("got %q", got)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	// Empty
	var empty ValidationErrors
	if got := empty.Error(); got != "" {
		t.Errorf("empty got %q", got)
	}

	// Single
	single := ValidationErrors{{Field: "A", Message: "bad"}}
	if got := single.Error(); got != "A: bad" {
		t.Errorf("single got %q", got)
	}

	// Multiple
	multi := ValidationErrors{
		{Field: "A", Message: "bad"},
		{Field: "B", Message: "wrong"},
	}
	got := multi.Error()
	if !strings.Contains(got, "A: bad") || !strings.Contains(got, "B: wrong") || !strings.Contains(got, "; ") {
		t.Errorf("multi got %q", got)
	}
}

func TestValidationErrors_ToArray(t *testing.T) {
	// Nil
	var nilErrs ValidationErrors
	if nilErrs.ToArray() != nil {
		t.Error("nil should return nil")
	}

	// Empty
	empty := ValidationErrors{}
	if empty.ToArray() != nil {
		t.Error("empty should return nil")
	}

	// With errors
	errs := ValidationErrors{{Field: "Name", Message: "required"}}
	arr := errs.ToArray()
	if len(arr) != 1 {
		t.Fatalf("expected 1, got %d", len(arr))
	}
	if arr[0]["field"] != "name" || arr[0]["message"] != "required" {
		t.Errorf("got %v", arr[0])
	}
}

// ============================================================================
// Validate — entry point
// ============================================================================

func TestValidate_ValidStruct(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	if err := Validate(S{Name: "ok"}); err != nil {
		t.Errorf("should pass: %v", err)
	}
}

func TestValidate_InvalidStruct(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	if err := Validate(S{}); err == nil {
		t.Error("should fail")
	}
}

func TestValidate_Pointer(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	s := &S{Name: "ok"}
	if err := Validate(s); err != nil {
		t.Errorf("pointer should pass: %v", err)
	}
}

func TestValidate_NilPointer(t *testing.T) {
	err := Validate((*struct{})(nil))
	if err == nil || !strings.Contains(err.Error(), "nil pointer") {
		t.Errorf("nil pointer should fail, got: %v", err)
	}
}

func TestValidate_NonStruct(t *testing.T) {
	err := Validate("not a struct")
	if err == nil || !strings.Contains(err.Error(), "struct") {
		t.Errorf("non-struct should fail, got: %v", err)
	}
}

func TestValidate_NonStructPointer(t *testing.T) {
	s := "string"
	err := Validate(&s)
	if err == nil {
		t.Error("pointer to non-struct should fail")
	}
}

func TestValidate_NoTags(t *testing.T) {
	type S struct {
		Name string
	}
	if err := Validate(S{Name: "ok"}); err != nil {
		t.Errorf("no tags should pass: %v", err)
	}
}

func TestValidate_UnexportedFields(t *testing.T) {
	type S struct {
		exported   string
		Name       string `validate:"required"`
		unexported int
	}
	if err := Validate(S{exported: "a", Name: "ok", unexported: 1}); err != nil {
		t.Errorf("unexported fields should be skipped: %v", err)
	}
}

// ============================================================================
// Custom Rules
// ============================================================================

func TestRegisterValidationRule(t *testing.T) {
	// Register
	RegisterValidationRule("even", func(fieldName string, value reflect.Value, param string) *ValidationError {
		if value.Kind() == reflect.Int && value.Int()%2 != 0 {
			return &ValidationError{Field: fieldName, Message: "must be even"}
		}
		return nil
	})

	type S struct {
		Num int `validate:"even"`
	}

	if err := Validate(S{Num: 4}); err != nil {
		t.Errorf("even should pass: %v", err)
	}
	if err := Validate(S{Num: 3}); err == nil {
		t.Error("odd should fail")
	}
}

func TestRegisterValidationRule_Empty(t *testing.T) {
	// Empty name or nil rule should be ignored
	RegisterValidationRule("", nil)
	RegisterValidationRule("test", nil)
	RegisterValidationRule("", func(_ string, _ reflect.Value, _ string) *ValidationError { return nil })
}

func TestCustomRuleWithParam(t *testing.T) {
	RegisterValidationRule("divisibleby", func(fieldName string, value reflect.Value, param string) *ValidationError {
		if value.Kind() != reflect.Int || param == "" {
			return nil
		}
		divisor := 0
		for _, c := range param {
			divisor = divisor*10 + int(c-'0')
		}
		if value.Int()%int64(divisor) != 0 {
			return &ValidationError{Field: fieldName, Message: "not divisible"}
		}
		return nil
	})

	type S struct {
		Num int `validate:"divisibleby=5"`
	}
	if err := Validate(S{Num: 10}); err != nil {
		t.Errorf("divisible should pass: %v", err)
	}
	if err := Validate(S{Num: 7}); err == nil {
		t.Error("not divisible should fail")
	}
}

func TestUnknownRule(t *testing.T) {
	type S struct {
		Name string `validate:"nonexistent_rule_xyz"`
	}
	err := Validate(S{Name: "test"})
	if err == nil {
		t.Error("unknown rule should fail")
	}
	if !strings.Contains(err.Error(), "unknown validation rule") {
		t.Errorf("should mention unknown rule, got: %v", err)
	}
}

// ============================================================================
// Field Cache
// ============================================================================

func TestFieldCache_HitOnSecondCall(t *testing.T) {
	type CacheTest struct {
		A string `validate:"required"`
	}
	// First call parses
	_ = Validate(CacheTest{A: "ok"})
	// Second call should use cache
	if err := Validate(CacheTest{A: "ok"}); err != nil {
		t.Errorf("cached call should pass: %v", err)
	}
}

func TestFieldCache_NestedStruct(t *testing.T) {
	type Inner struct {
		X string `validate:"required"`
	}
	type Outer struct {
		Inner Inner
	}
	if err := Validate(Outer{Inner: Inner{X: "ok"}}); err != nil {
		t.Errorf("nested should pass: %v", err)
	}
	if err := Validate(Outer{}); err == nil {
		t.Error("nested empty should fail")
	}
}

func TestFieldCache_NestedStructNoTag(t *testing.T) {
	type Inner struct {
		X string `validate:"required"`
	}
	type Outer struct {
		Name  string `validate:"required"`
		Inner Inner  // no tag, but nested struct
	}
	err := Validate(Outer{Name: "ok"})
	if err == nil {
		t.Error("should fail for Inner.X")
	}
	if !strings.Contains(err.Error(), "Inner.X") {
		t.Errorf("should prefix with Inner, got: %v", err)
	}
}

func TestFieldCache_PointerToNestedStruct(t *testing.T) {
	type Inner struct {
		X string `validate:"required"`
	}
	type Outer struct {
		Inner *Inner
	}
	// Nil pointer nested — should skip
	if err := Validate(Outer{Inner: nil}); err != nil {
		t.Errorf("nil nested pointer should pass: %v", err)
	}

	// Non-nil pointer with invalid inner
	if err := Validate(Outer{Inner: &Inner{}}); err == nil {
		t.Error("pointer to invalid inner should fail")
	}

	// Valid
	if err := Validate(Outer{Inner: &Inner{X: "ok"}}); err != nil {
		t.Errorf("valid inner should pass: %v", err)
	}
}

// ============================================================================
// Dive with tag containing "dive" keyword
// ============================================================================

func TestDive_OnlyDive(t *testing.T) {
	type S struct {
		Tags []string `validate:"dive,required"`
	}
	if err := Validate(S{Tags: []string{"a"}}); err != nil {
		t.Errorf("should pass: %v", err)
	}
	err := Validate(S{Tags: []string{"a", ""}})
	if err == nil {
		t.Error("empty element should fail")
	}
}

func TestDive_PreDiveRules(t *testing.T) {
	type S struct {
		Tags []string `validate:"notempty,dive,required"`
	}
	// Empty slice fails pre-dive notempty
	if err := Validate(S{Tags: []string{}}); err == nil {
		t.Error("empty slice should fail notempty")
	}
}

func TestDive_NilSlice(t *testing.T) {
	type S struct {
		Tags []string `validate:"dive,required"`
	}
	if err := Validate(S{Tags: nil}); err != nil {
		t.Errorf("nil slice dive should pass: %v", err)
	}
}

func TestDive_Map(t *testing.T) {
	type S struct {
		Headers map[string]string `validate:"dive,required"`
	}
	if err := Validate(S{Headers: map[string]string{"k": "v"}}); err != nil {
		t.Errorf("map dive should pass: %v", err)
	}
	if err := Validate(S{Headers: map[string]string{"k": ""}}); err == nil {
		t.Error("map empty value should fail")
	}
}

func TestDive_NilMap(t *testing.T) {
	type S struct {
		M map[string]string `validate:"dive,required"`
	}
	if err := Validate(S{M: nil}); err != nil {
		t.Errorf("nil map dive should pass: %v", err)
	}
}
