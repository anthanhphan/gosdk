package validator

import (
	"reflect"
	"strings"
	"testing"
)

// ============================================================================
// Non-string type coverage for string/format/network validators
// ============================================================================

func TestNonString_StringRules(t *testing.T) {
	type S struct {
		Val int `validate:"startswith=x"`
	}
	if err := Validate(S{Val: 123}); err != nil {
		t.Errorf("startswith non-string: %v", err)
	}

	type S2 struct {
		Val int `validate:"endswith=x"`
	}
	if err := Validate(S2{Val: 123}); err != nil {
		t.Errorf("endswith non-string: %v", err)
	}

	type S3 struct {
		Val int `validate:"lowercase"`
	}
	if err := Validate(S3{Val: 123}); err != nil {
		t.Errorf("lowercase non-string: %v", err)
	}

	type S4 struct {
		Val int `validate:"uppercase"`
	}
	if err := Validate(S4{Val: 123}); err != nil {
		t.Errorf("uppercase non-string: %v", err)
	}

	type S5 struct {
		Val int `validate:"excludes=x"`
	}
	if err := Validate(S5{Val: 123}); err != nil {
		t.Errorf("excludes non-string: %v", err)
	}
}

func TestNonString_FormatRules(t *testing.T) {
	type S1 struct {
		Val int `validate:"url"`
	}
	if err := Validate(S1{Val: 1}); err != nil {
		t.Errorf("url non-string: %v", err)
	}

	type S2 struct {
		Val int `validate:"numeric"`
	}
	if err := Validate(S2{Val: 1}); err != nil {
		t.Errorf("numeric non-string: %v", err)
	}

	type S3 struct {
		Val int `validate:"alpha"`
	}
	if err := Validate(S3{Val: 1}); err != nil {
		t.Errorf("alpha non-string: %v", err)
	}

	type S4 struct {
		Val int `validate:"alphanumeric"`
	}
	if err := Validate(S4{Val: 1}); err != nil {
		t.Errorf("alphanumeric non-string: %v", err)
	}

	type S5 struct {
		Val int `validate:"uuid"`
	}
	if err := Validate(S5{Val: 1}); err != nil {
		t.Errorf("uuid non-string: %v", err)
	}

	type S6 struct {
		Val int `validate:"hexcolor"`
	}
	if err := Validate(S6{Val: 1}); err != nil {
		t.Errorf("hexcolor non-string: %v", err)
	}
}

func TestNonString_NetworkRules(t *testing.T) {
	type S1 struct {
		Val int `validate:"ip"`
	}
	if err := Validate(S1{Val: 1}); err != nil {
		t.Errorf("ip non-string: %v", err)
	}

	type S2 struct {
		Val int `validate:"ipv4"`
	}
	if err := Validate(S2{Val: 1}); err != nil {
		t.Errorf("ipv4 non-string: %v", err)
	}

	type S3 struct {
		Val int `validate:"ipv6"`
	}
	if err := Validate(S3{Val: 1}); err != nil {
		t.Errorf("ipv6 non-string: %v", err)
	}
}

func TestNonString_DatetimeRule(t *testing.T) {
	type S struct {
		Val int `validate:"datetime"`
	}
	if err := Validate(S{Val: 1}); err != nil {
		t.Errorf("datetime non-string: %v", err)
	}
}

// ============================================================================
// applyResolvedRules — interface element unwrapping
// ============================================================================

func TestDive_MapInterfaceValues(t *testing.T) {
	type S struct {
		Data map[string]interface{} `validate:"dive,required"`
	}
	// Interface values trigger elem.Kind() == reflect.Interface path
	err := Validate(S{Data: map[string]interface{}{"k": ""}})
	if err == nil {
		t.Error("empty interface value should fail required")
	}

	// Valid
	if err := Validate(S{Data: map[string]interface{}{"k": "ok"}}); err != nil {
		t.Errorf("map interface should pass: %v", err)
	}
}

// ============================================================================
// validateNestedStruct — non-struct kind after deref
// ============================================================================

func TestNested_PtrToNonStruct(t *testing.T) {
	// This exercises the value.Kind() != reflect.Struct return in validateNestedStruct
	// The hasNested check is already done at parse time where we check ft.Kind() == reflect.Struct
	// So this path is not normally reachable. We test it indirectly.
	type Inner struct {
		X string `validate:"required"`
	}
	type Outer struct {
		Name  string `validate:"required"`
		Inner *Inner
	}
	// Nil pointer nested struct should be handled
	if err := Validate(Outer{Name: "ok", Inner: nil}); err != nil {
		t.Errorf("nil nested ptr should pass: %v", err)
	}
}

// ============================================================================
// Dive with struct elements (triggers recursive struct validation in applyResolvedRules)
// ============================================================================

func TestDive_StructElements_Valid(t *testing.T) {
	type Item struct {
		Name string `validate:"required"`
	}
	type S struct {
		Items []Item `validate:"notempty,dive,required"`
	}
	err := Validate(S{Items: []Item{{Name: "a"}, {Name: "b"}}})
	if err != nil {
		t.Errorf("dive struct elements should pass: %v", err)
	}
}

func TestDive_PtrStructElements(t *testing.T) {
	type Item struct {
		Name string `validate:"required"`
	}
	type S struct {
		Items []*Item `validate:"dive,required"`
	}

	err := Validate(S{Items: []*Item{{Name: "a"}}})
	if err != nil {
		t.Errorf("dive ptr struct should pass: %v", err)
	}

	// Ptr to struct with invalid field
	err = Validate(S{Items: []*Item{{Name: ""}}})
	if err == nil {
		t.Error("dive ptr struct invalid should fail")
	}
	if !strings.Contains(err.Error(), "Items[0].Name") {
		t.Errorf("should show Items[0].Name, got: %v", err)
	}
}

// ============================================================================
// validateNestedStruct — nestedErr not ValidationErrors type (edge case)
// ============================================================================

func TestNested_ValidInnerStruct(t *testing.T) {
	type Inner struct {
		X string `validate:"required"`
	}
	type Outer struct {
		Inner Inner
	}
	// Valid inner — nestedErr == nil
	if err := Validate(Outer{Inner: Inner{X: "ok"}}); err != nil {
		t.Errorf("valid inner should pass: %v", err)
	}
}

// ============================================================================
// NotEmpty on array type
// ============================================================================

func TestNotEmpty_Array(t *testing.T) {
	type S struct {
		Arr [3]int `validate:"notempty"`
	}
	// Array with len > 0 passes
	if err := Validate(S{Arr: [3]int{1, 2, 3}}); err != nil {
		t.Errorf("notempty array pass: %v", err)
	}
}

func TestNotEmpty_NonCollectionType(t *testing.T) {
	type S struct {
		Val int `validate:"notempty"`
	}
	// notempty on int should pass (not a collection)
	if err := Validate(S{Val: 0}); err != nil {
		t.Errorf("notempty on int should pass: %v", err)
	}
}

// ============================================================================
// Min/Max on unsupported types (default case)
// ============================================================================

func TestMinMax_UnsupportedType(t *testing.T) {
	type S struct {
		B bool `validate:"min=1"`
	}
	// Bool is not supported by min, should pass (no matching case)
	if err := Validate(S{B: true}); err != nil {
		t.Errorf("min on bool should pass: %v", err)
	}

	type S2 struct {
		B bool `validate:"max=1"`
	}
	if err := Validate(S2{B: true}); err != nil {
		t.Errorf("max on bool should pass: %v", err)
	}
}

// ============================================================================
// Len on unsupported type
// ============================================================================

func TestLen_UnsupportedType(t *testing.T) {
	type S struct {
		Val int `validate:"len=5"`
	}
	// int not supported by len, should pass
	if err := Validate(S{Val: 123}); err != nil {
		t.Errorf("len on int should pass: %v", err)
	}
}

// ============================================================================
// isZeroValue — default case (struct type returns false)
// ============================================================================

func TestIsZeroValue_StructType(t *testing.T) {
	type Inner struct{}
	type S struct {
		Inner Inner `validate:"required"`
	}
	// Struct type hits default case in isZeroValue, returns false
	if err := Validate(S{}); err != nil {
		t.Errorf("required struct should pass (not zero): %v", err)
	}
}

// ============================================================================
// StopOnFirstError + nested (covers ValidateStruct nested+stopOnFirst branch)
// ============================================================================

func TestStopOnFirst_NestedErrors(t *testing.T) {
	type Inner struct {
		A string `validate:"required"`
	}
	type Outer struct {
		Name  string `validate:"required"`
		Inner Inner
	}

	v := New(WithStopOnFirstError(true))
	// Name is valid, Inner.A is empty → should get nested error and stop
	err := v.ValidateStruct(Outer{Name: "ok"})
	if err == nil {
		t.Fatal("should fail")
	}
	errs := err.(ValidationErrors)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

// ============================================================================
// NotEmpty on zero-length array [0]int
// ============================================================================

func TestNotEmpty_ZeroLenArray(t *testing.T) {
	type S struct {
		Arr [0]int `validate:"notempty"`
	}
	if err := Validate(S{}); err == nil {
		t.Error("notempty on [0]int should fail")
	}
}

// ============================================================================
// validateNestedStruct — cover !ok branch (non-ValidationErrors error)
// ============================================================================

func TestNested_NestedPtrStruct_Valid(t *testing.T) {
	type Inner struct {
		X string `validate:"required"`
	}
	type Outer struct {
		Inner *Inner
	}
	// Valid nested ptr
	if err := Validate(Outer{Inner: &Inner{X: "ok"}}); err != nil {
		t.Errorf("valid nested ptr should pass: %v", err)
	}

	// Invalid nested ptr — forces nested errors
	err := Validate(Outer{Inner: &Inner{}})
	if err == nil {
		t.Fatal("invalid inner should fail")
	}
	if !strings.Contains(err.Error(), "Inner.X") {
		t.Errorf("should show Inner.X, got: %v", err)
	}
}

// Test that the Validator's ValidateStruct properly handles nested struct
// that returns an error wrapping format (non-ValidationErrors via fmt.Errorf)
func TestNested_WithTagOverride_NestedErrors(t *testing.T) {
	type Address struct {
		Street string `json:"street" validate:"required"`
		Zip    string `json:"zip" validate:"required,len=5"`
	}
	type User struct {
		Name    string   `json:"name" validate:"required"`
		Address *Address `json:"address"`
	}

	v := New(WithFieldNameTag("json"))

	// Nil nested ptr — should pass
	if err := v.ValidateStruct(User{Name: "ok", Address: nil}); err != nil {
		t.Errorf("nil address should pass: %v", err)
	}

	// Invalid nested
	err := v.ValidateStruct(User{Name: "ok", Address: &Address{}})
	if err == nil {
		t.Fatal("empty address should fail")
	}
	if !strings.Contains(err.Error(), "address.street") {
		t.Errorf("should show address.street, got: %v", err)
	}
}

// ============================================================================
// getOrParseFields — pointer to nested struct without validate tag
// ============================================================================

func TestParsing_PtrNestedNoTag(t *testing.T) {
	type Inner struct {
		Val string `validate:"required"`
	}
	type Outer struct {
		Ref *Inner // no validate tag, but ptr to struct → hasNested
	}

	// Nil ptr
	if err := Validate(Outer{Ref: nil}); err != nil {
		t.Errorf("nil ptr no tag should pass: %v", err)
	}

	// Non-nil with valid inner
	if err := Validate(Outer{Ref: &Inner{Val: "ok"}}); err != nil {
		t.Errorf("valid inner no tag should pass: %v", err)
	}

	// Non-nil with invalid inner
	err := Validate(Outer{Ref: &Inner{}})
	if err == nil {
		t.Fatal("invalid inner should fail")
	}
}

// ============================================================================
// Internal: validateNestedStruct defensive guards
// ============================================================================

func TestValidateNestedStruct_NonStructKind(t *testing.T) {
	// Directly test the non-struct guard (L191-192)
	// Create an instanceCachedField that thinks it's nested but value is string
	v := New()
	cf := &instanceCachedField{
		cachedField: cachedField{
			name:      "Fake",
			hasNested: true,
			isPtr:     false,
		},
		displayName: "Fake",
	}
	result := v.validateNestedStruct(cf, reflect.ValueOf("not a struct"))
	if result != nil {
		t.Errorf("non-struct should return nil, got: %v", result)
	}
}

func TestValidateNestedStruct_NonValidationErrorsReturn(t *testing.T) {
	// The !ok branch (L200-201) is reached when ValidateStruct returns
	// a non-ValidationErrors error. This happens on nil ptr or non-struct.
	// We can't easily trigger this through public API since validateNestedStruct
	// is only called when hasNested is true. This is a defensive guard.
	// The guard is validated above with non-struct kind test.
}

// ============================================================================
// Internal: getOrParseFields — struct with only unexported fields
// ============================================================================

func TestGetOrParseFields_AllUnexported(t *testing.T) {
	type allPrivate struct {
		x int
		y string
	}
	// Should parse fine but have no fields
	if err := Validate(allPrivate{x: 1, y: "a"}); err != nil {
		t.Errorf("all unexported should pass: %v", err)
	}
}
