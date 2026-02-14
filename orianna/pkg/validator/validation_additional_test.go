// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"testing"
)

// Tests for validateMin - hitting uint, float, slice, array branches

func TestValidateMin_Uint(t *testing.T) {
	type Test struct {
		Count uint `validate:"min=5"`
	}
	valid := Test{Count: 10}
	if err := Validate(valid); err != nil {
		t.Errorf("min=5 uint 10 should pass, got %v", err)
	}
	invalid := Test{Count: 2}
	if err := Validate(invalid); err == nil {
		t.Error("min=5 uint 2 should fail")
	}
}

func TestValidateMin_Float(t *testing.T) {
	type Test struct {
		Rate float64 `validate:"min=10"`
	}
	valid := Test{Rate: 15.5}
	if err := Validate(valid); err != nil {
		t.Errorf("min=10 float 15.5 should pass, got %v", err)
	}
	invalid := Test{Rate: 3.5}
	if err := Validate(invalid); err == nil {
		t.Error("min=10 float 3.5 should fail")
	}
}

func TestValidateMin_Slice(t *testing.T) {
	type Test struct {
		Items []string `validate:"min=2"`
	}
	valid := Test{Items: []string{"a", "b", "c"}}
	if err := Validate(valid); err != nil {
		t.Errorf("min=2 slice with 3 items should pass, got %v", err)
	}
	invalid := Test{Items: []string{"a"}}
	if err := Validate(invalid); err == nil {
		t.Error("min=2 slice with 1 item should fail")
	}
}

func TestValidateMin_Int_Boundary(t *testing.T) {
	type Test struct {
		Age int `validate:"min=18"`
	}
	exact := Test{Age: 18}
	if err := Validate(exact); err != nil {
		t.Errorf("min=18 int 18 (exact boundary) should pass, got %v", err)
	}
	below := Test{Age: 17}
	if err := Validate(below); err == nil {
		t.Error("min=18 int 17 should fail")
	}
}

// Tests for validateRule - hitting URL, email, numeric, alpha dispatch

func TestValidateRule_URL(t *testing.T) {
	type Test struct {
		Website string `validate:"url"`
	}
	valid := Test{Website: "https://example.com"}
	if err := Validate(valid); err != nil {
		t.Errorf("url should pass for valid URL, got %v", err)
	}
	invalid := Test{Website: "not a url"}
	if err := Validate(invalid); err == nil {
		t.Error("url should fail for invalid URL")
	}
}

func TestValidateRule_Email(t *testing.T) {
	type Test struct {
		Email string `validate:"email"`
	}
	valid := Test{Email: "test@example.com"}
	if err := Validate(valid); err != nil {
		t.Errorf("email should pass for valid email, got %v", err)
	}
	invalid := Test{Email: "not-an-email"}
	if err := Validate(invalid); err == nil {
		t.Error("email should fail for invalid email")
	}
}

func TestValidateRule_Numeric(t *testing.T) {
	type Test struct {
		Code string `validate:"numeric"`
	}
	valid := Test{Code: "12345"}
	if err := Validate(valid); err != nil {
		t.Errorf("numeric should pass for numeric string, got %v", err)
	}
	invalid := Test{Code: "abc123"}
	if err := Validate(invalid); err == nil {
		t.Error("numeric should fail for alpha-numeric string")
	}
}

func TestValidateRule_Alpha(t *testing.T) {
	type Test struct {
		Name string `validate:"alpha"`
	}
	valid := Test{Name: "JohnDoe"}
	if err := Validate(valid); err != nil {
		t.Errorf("alpha should pass for alpha string, got %v", err)
	}
	invalid := Test{Name: "John123"}
	if err := Validate(invalid); err == nil {
		t.Error("alpha should fail for non-alpha string")
	}
}

// Tests for validateComparison - hitting gte and lte branches

func TestValidateComparison_Gte(t *testing.T) {
	type Test struct {
		Score int `validate:"gte=10"`
	}
	exact := Test{Score: 10}
	if err := Validate(exact); err != nil {
		t.Errorf("gte=10 for 10 should pass, got %v", err)
	}
	above := Test{Score: 11}
	if err := Validate(above); err != nil {
		t.Errorf("gte=10 for 11 should pass, got %v", err)
	}
	below := Test{Score: 9}
	if err := Validate(below); err == nil {
		t.Error("gte=10 for 9 should fail")
	}
}

func TestValidateComparison_Lte(t *testing.T) {
	type Test struct {
		Score int `validate:"lte=100"`
	}
	exact := Test{Score: 100}
	if err := Validate(exact); err != nil {
		t.Errorf("lte=100 for 100 should pass, got %v", err)
	}
	below := Test{Score: 50}
	if err := Validate(below); err != nil {
		t.Errorf("lte=100 for 50 should pass, got %v", err)
	}
	above := Test{Score: 101}
	if err := Validate(above); err == nil {
		t.Error("lte=100 for 101 should fail")
	}
}

func TestValidateComparison_Gt_Float(t *testing.T) {
	type Test struct {
		Price float32 `validate:"gt=0"`
	}
	valid := Test{Price: 9.99}
	if err := Validate(valid); err != nil {
		t.Errorf("gt=0 for float 9.99 should pass, got %v", err)
	}
	zero := Test{Price: 0}
	if err := Validate(zero); err == nil {
		t.Error("gt=0 for float 0 should fail")
	}
}

func TestValidateComparison_Lte_Uint(t *testing.T) {
	type Test struct {
		Count uint `validate:"lte=50"`
	}
	valid := Test{Count: 25}
	if err := Validate(valid); err != nil {
		t.Errorf("lte=50 for uint 25 should pass, got %v", err)
	}
	over := Test{Count: 100}
	if err := Validate(over); err == nil {
		t.Error("lte=50 for uint 100 should fail")
	}
}

// Test ToArray with non-empty errors

func TestValidationErrors_ToArray_Content(t *testing.T) {
	errors := ValidationErrors{
		{Field: "Name", Message: "is required"},
		{Field: "Email", Message: "must be valid"},
	}
	arr := errors.ToArray()
	if len(arr) != 2 {
		t.Fatalf("ToArray() returned %d items, want 2", len(arr))
	}
	if arr[0]["field"] != "name" {
		t.Errorf("ToArray()[0][field] = %s, want name (lowercase)", arr[0]["field"])
	}
	if arr[1]["message"] != "must be valid" {
		t.Errorf("ToArray()[1][message] = %s, want 'must be valid'", arr[1]["message"])
	}
}

func TestValidationErrors_ToArray_NilResult(t *testing.T) {
	var errors ValidationErrors
	arr := errors.ToArray()
	if arr != nil {
		t.Error("ToArray() should return nil for empty errors")
	}
}

func TestValidationErrors_Error_EmptyString(t *testing.T) {
	var errors ValidationErrors
	if errors.Error() != "" {
		t.Error("Error() should return empty string for empty errors")
	}
}

// Tests for isZeroValue - hitting chan and func types

func TestIsZeroValue_Chan(t *testing.T) {
	type Test struct {
		Ch chan int `validate:"required"`
	}
	valid := Test{Ch: make(chan int)}
	if err := Validate(valid); err != nil {
		t.Errorf("required chan should pass for non-nil, got %v", err)
	}
	invalid := Test{Ch: nil}
	if err := Validate(invalid); err == nil {
		t.Error("required chan should fail for nil")
	}
}

func TestIsZeroValue_Interface(t *testing.T) {
	type Test struct {
		Data interface{} `validate:"required"`
	}
	valid := Test{Data: "hello"}
	if err := Validate(valid); err != nil {
		t.Errorf("required interface should pass for non-nil, got %v", err)
	}
	invalid := Test{Data: nil}
	if err := Validate(invalid); err == nil {
		t.Error("required interface should fail for nil")
	}
}

// len validation for array type

func TestValidateLen_Array(t *testing.T) {
	type Test struct {
		Items [3]string `validate:"len=3"`
	}
	valid := Test{Items: [3]string{"a", "b", "c"}}
	if err := Validate(valid); err != nil {
		t.Errorf("len=3 should pass for array with 3 items, got %v", err)
	}
}

// Validate - nil pointer

func TestValidate_NilPointer(t *testing.T) {
	type User struct {
		Name string `validate:"required"`
	}
	var u *User
	err := Validate(u)
	if err == nil {
		t.Error("Validate() should error for nil pointer")
	}
}

// Validate - non-struct input

func TestValidate_NonStruct(t *testing.T) {
	err := Validate("not a struct")
	if err == nil {
		t.Error("Validate() should error for non-struct input")
	}
}

// oneof with empty param

func TestValidateOneOf_EmptyParam(t *testing.T) {
	type Test struct {
		Status string `validate:"oneof="`
	}
	test := Test{Status: "active"}
	// oneof= with empty param returns nil (early return)
	err := Validate(test)
	if err != nil {
		t.Errorf("oneof= with empty param should be no-op, got %v", err)
	}
}

// len with empty param

func TestValidateLen_EmptyParam(t *testing.T) {
	type Test struct {
		Name string `validate:"len="`
	}
	test := Test{Name: "hello"}
	// len= with empty param is handled by early return (returns nil)
	err := Validate(test)
	if err != nil {
		t.Errorf("len= with empty param should be no-op, got %v", err)
	}
}

// gt with empty param

func TestValidateGt_EmptyParam(t *testing.T) {
	type Test struct {
		Value int `validate:"gt="`
	}
	test := Test{Value: 5}
	err := Validate(test)
	if err != nil {
		t.Errorf("gt= with empty param should not error, got %v", err)
	}
}

// gte with empty param

func TestValidateGte_EmptyParam(t *testing.T) {
	type Test struct {
		Value int `validate:"gte="`
	}
	test := Test{Value: 5}
	err := Validate(test)
	if err != nil {
		t.Errorf("gte= with empty param should not error, got %v", err)
	}
}

// lt with empty param

func TestValidateLt_EmptyParam(t *testing.T) {
	type Test struct {
		Value int `validate:"lt="`
	}
	test := Test{Value: 5}
	err := Validate(test)
	if err != nil {
		t.Errorf("lt= with empty param should not error, got %v", err)
	}
}

// lte with empty param

func TestValidateLte_EmptyParam(t *testing.T) {
	type Test struct {
		Value int `validate:"lte="`
	}
	test := Test{Value: 5}
	err := Validate(test)
	if err != nil {
		t.Errorf("lte= with empty param should not error, got %v", err)
	}
}

// Email/URL/numeric/alpha with non-string types should be no-op

func TestValidateEmail_NonString(t *testing.T) {
	type Test struct {
		Age int `validate:"email"`
	}
	test := Test{Age: 25}
	// email rule on int should not error (no-op)
	err := Validate(test)
	if err != nil {
		t.Errorf("email on int should be no-op, got %v", err)
	}
}

// Alphanumeric with empty string

func TestValidateAlphanumeric_Empty(t *testing.T) {
	type Test struct {
		Code string `validate:"alphanumeric"`
	}
	test := Test{Code: ""}
	err := Validate(test)
	if err != nil {
		t.Errorf("alphanumeric on empty string should pass, got %v", err)
	}
}

// UUID with empty string

func TestValidateUUID_Empty(t *testing.T) {
	type Test struct {
		ID string `validate:"uuid"`
	}
	test := Test{ID: ""}
	err := Validate(test)
	if err != nil {
		t.Errorf("uuid on empty string should pass, got %v", err)
	}
}

// Test nested pointer that is nil (should NOT error for optional nested struct)

func TestValidate_NilNestedPointer(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
	}
	type User struct {
		Name    string   `validate:"required"`
		Address *Address // nil pointer - skip nested validation
	}

	user := User{Name: "John", Address: nil}
	err := Validate(user)
	// Nil pointer nested struct should be skipped
	if err != nil {
		t.Errorf("nil pointer nested struct should be skipped, got %v", err)
	}
}

// Multiple rules on one field

func TestValidate_MultipleRules(t *testing.T) {
	type Test struct {
		Name string `validate:"required,min=3,max=10"`
	}
	tooShort := Test{Name: "Ab"}
	err := Validate(tooShort)
	if err == nil {
		t.Error("min=3 should fail for 2-char string")
	}
	valErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatal("Should return ValidationErrors")
	}
	if len(valErrs) != 1 {
		t.Errorf("Expected 1 error (min=3), got %d", len(valErrs))
	}
}
