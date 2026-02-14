// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"reflect"
	"testing"
)

func TestValidate_PointerToStruct(t *testing.T) {
	type User struct {
		Name string `validate:"required"`
	}

	user := &User{Name: "John"}
	err := Validate(user)
	if err != nil {
		t.Errorf("Validate() should not error for valid pointer-to-struct, got %v", err)
	}

	empty := &User{}
	err = Validate(empty)
	if err == nil {
		t.Error("Validate() should error for pointer-to-struct with invalid data")
	}
}

func TestValidate_NestedPointerToStruct(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
	}

	type User struct {
		Name    string `validate:"required"`
		Address *Address
	}

	// Non-nil nested pointer with required field missing
	user := User{
		Name:    "John",
		Address: &Address{Street: ""},
	}

	err := Validate(user)
	if err == nil {
		t.Error("Validate() should error for non-nil pointer-to-struct with missing required field")
	}

	validationErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatal("Validate() should return ValidationErrors")
	}

	found := false
	for _, e := range validationErrs {
		if e.Field == "Address.Street" {
			found = true
		}
	}
	if !found {
		t.Error("Expected error for Address.Street")
	}
}

func TestValidateMax_String(t *testing.T) {
	type Test struct {
		Name string `validate:"max=5"`
	}

	valid := Test{Name: "John"}
	if err := Validate(valid); err != nil {
		t.Errorf("max=5 should pass for 4 chars, got %v", err)
	}

	invalid := Test{Name: "Jonathan"}
	if err := Validate(invalid); err == nil {
		t.Error("max=5 should fail for 8 chars")
	}
}

func TestValidateMax_Slice(t *testing.T) {
	type Test struct {
		Items []string `validate:"max=3"`
	}

	valid := Test{Items: []string{"a", "b"}}
	if err := Validate(valid); err != nil {
		t.Errorf("max=3 should pass for 2 items, got %v", err)
	}

	invalid := Test{Items: []string{"a", "b", "c", "d"}}
	if err := Validate(invalid); err == nil {
		t.Error("max=3 should fail for 4 items")
	}
}

func TestValidateMax_Uint(t *testing.T) {
	type Test struct {
		Count uint `validate:"max=10"`
	}

	valid := Test{Count: 5}
	if err := Validate(valid); err != nil {
		t.Errorf("max=10 should pass for 5, got %v", err)
	}

	invalid := Test{Count: 15}
	if err := Validate(invalid); err == nil {
		t.Error("max=10 should fail for 15")
	}
}

func TestValidateMax_Float(t *testing.T) {
	type Test struct {
		Value float64 `validate:"max=100"`
	}

	valid := Test{Value: 50.5}
	if err := Validate(valid); err != nil {
		t.Errorf("max=100 should pass for 50.5, got %v", err)
	}

	invalid := Test{Value: 150.5}
	if err := Validate(invalid); err == nil {
		t.Error("max=100 should fail for 150.5")
	}
}

func TestValidateMax_Int(t *testing.T) {
	type Test struct {
		Value int `validate:"max=10"`
	}

	valid := Test{Value: 5}
	if err := Validate(valid); err != nil {
		t.Errorf("max=10 should pass for 5, got %v", err)
	}

	invalid := Test{Value: 15}
	if err := Validate(invalid); err == nil {
		t.Error("max=10 should fail for 15")
	}
}

func TestIsZeroValue_Bool(t *testing.T) {
	type Test struct {
		Active bool `validate:"required"`
	}

	active := Test{Active: true}
	if err := Validate(active); err != nil {
		t.Errorf("required bool=true should pass, got %v", err)
	}

	inactive := Test{Active: false}
	if err := Validate(inactive); err == nil {
		t.Error("required bool=false should fail")
	}
}

func TestIsZeroValue_Uint(t *testing.T) {
	type Test struct {
		Count uint `validate:"required"`
	}

	valid := Test{Count: 1}
	if err := Validate(valid); err != nil {
		t.Errorf("required uint=1 should pass, got %v", err)
	}

	zero := Test{Count: 0}
	if err := Validate(zero); err == nil {
		t.Error("required uint=0 should fail")
	}
}

func TestIsZeroValue_Float(t *testing.T) {
	type Test struct {
		Rating float32 `validate:"required"`
	}

	valid := Test{Rating: 4.5}
	if err := Validate(valid); err != nil {
		t.Errorf("required float=4.5 should pass, got %v", err)
	}

	zero := Test{Rating: 0}
	if err := Validate(zero); err == nil {
		t.Error("required float=0 should fail")
	}
}

func TestIsZeroValue_Pointer(t *testing.T) {
	type Test struct {
		Data *string `validate:"required"`
	}

	s := "hello"
	valid := Test{Data: &s}
	if err := Validate(valid); err != nil {
		t.Errorf("required pointer=non-nil should pass, got %v", err)
	}

	nilPtr := Test{Data: nil}
	if err := Validate(nilPtr); err == nil {
		t.Error("required pointer=nil should fail")
	}
}

func TestIsZeroValue_Slice(t *testing.T) {
	type Test struct {
		Items []string `validate:"required"`
	}

	valid := Test{Items: []string{"a"}}
	if err := Validate(valid); err != nil {
		t.Errorf("required slice=non-nil should pass, got %v", err)
	}

	nilSlice := Test{Items: nil}
	if err := Validate(nilSlice); err == nil {
		t.Error("required slice=nil should fail")
	}
}

func TestIsZeroValue_Map(t *testing.T) {
	type Test struct {
		Meta map[string]string `validate:"required"`
	}

	valid := Test{Meta: map[string]string{"k": "v"}}
	if err := Validate(valid); err != nil {
		t.Errorf("required map=non-nil should pass, got %v", err)
	}

	nilMap := Test{Meta: nil}
	if err := Validate(nilMap); err == nil {
		t.Error("required map=nil should fail")
	}
}

func TestValidateComparison_Uint(t *testing.T) {
	type Test struct {
		Count uint `validate:"gt=5"`
	}

	valid := Test{Count: 10}
	if err := Validate(valid); err != nil {
		t.Errorf("gt=5 for uint 10 should pass, got %v", err)
	}

	invalid := Test{Count: 3}
	if err := Validate(invalid); err == nil {
		t.Error("gt=5 for uint 3 should fail")
	}
}

func TestValidateComparison_Float(t *testing.T) {
	type Test struct {
		Rate float64 `validate:"lt=100.0"`
	}

	valid := Test{Rate: 50.5}
	if err := Validate(valid); err != nil {
		t.Errorf("lt=100 for float 50.5 should pass, got %v", err)
	}

	invalid := Test{Rate: 150.0}
	if err := Validate(invalid); err == nil {
		t.Error("lt=100 for float 150.0 should fail")
	}
}

func TestValidateRule_CustomRule(t *testing.T) {
	ruleName := "test_even"
	RegisterValidationRule(ruleName, func(fieldName string, value reflect.Value, param string) *ValidationError {
		if value.Kind() == reflect.Int && value.Int()%2 != 0 {
			return &ValidationError{Field: fieldName, Message: "must be even"}
		}
		return nil
	})

	type Test struct {
		Number int `validate:"test_even"`
	}

	valid := Test{Number: 4}
	if err := Validate(valid); err != nil {
		t.Errorf("Custom rule should pass for even number, got %v", err)
	}

	invalid := Test{Number: 3}
	if err := Validate(invalid); err == nil {
		t.Error("Custom rule should fail for odd number")
	}
}

func TestValidateRule_UnknownRule(t *testing.T) {
	type Test struct {
		Value string `validate:"totally_unknown_rule"`
	}

	test := Test{Value: "hello"}
	// I1 fix: Unknown rules now produce a validation error to catch typos
	err := Validate(test)
	if err == nil {
		t.Fatal("Unknown validation rule should produce an error")
	}
	validationErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("Expected ValidationErrors, got %T", err)
	}
	if len(validationErrs) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(validationErrs))
	}
	if validationErrs[0].Field != "Value" {
		t.Errorf("Expected field 'Value', got %q", validationErrs[0].Field)
	}
}

func TestValidateMin_EmptyParam(t *testing.T) {
	type Test struct {
		Name string `validate:"min="`
	}

	test := Test{Name: "hello"}
	// Empty param should be handled gracefully (min with 0)
	err := Validate(test)
	if err != nil {
		t.Errorf("min= with empty param should not error, got %v", err)
	}
}

func TestValidateMax_EmptyParam(t *testing.T) {
	type Test struct {
		Name string `validate:"max="`
	}

	test := Test{Name: "hello"}
	// Empty param should be handled gracefully
	err := Validate(test)
	if err != nil {
		t.Errorf("max= with empty param should not error, got %v", err)
	}
}

func TestValidateAlphanumeric_Valid(t *testing.T) {
	type Test struct {
		Code string `validate:"alphanumeric"`
	}

	valid := Test{Code: "ABC123"}
	if err := Validate(valid); err != nil {
		t.Errorf("alphanumeric should pass for ABC123, got %v", err)
	}

	invalid := Test{Code: "ABC-123!"}
	if err := Validate(invalid); err == nil {
		t.Error("alphanumeric should fail for ABC-123!")
	}
}

func TestValidateUUID_Valid(t *testing.T) {
	type Test struct {
		ID string `validate:"uuid"`
	}

	valid := Test{ID: "550e8400-e29b-41d4-a716-446655440000"}
	if err := Validate(valid); err != nil {
		t.Errorf("uuid should pass for valid UUID, got %v", err)
	}

	invalid := Test{ID: "not-a-uuid"}
	if err := Validate(invalid); err == nil {
		t.Error("uuid should fail for invalid UUID")
	}
}

func TestValidate_UnexportedFields(t *testing.T) {
	type Test struct {
		Name     string `validate:"required"`
		internal string `validate:"required"`
	}

	test := Test{Name: "John"}
	_ = test.internal // field exists to test that unexported fields are skipped
	err := Validate(test)
	// Should only validate exported fields
	if err != nil {
		t.Errorf("Should not validate unexported fields, got %v", err)
	}
}

func TestValidateLen_Map(t *testing.T) {
	type Test struct {
		Meta map[string]string `validate:"len=2"`
	}

	valid := Test{Meta: map[string]string{"a": "1", "b": "2"}}
	if err := Validate(valid); err != nil {
		t.Errorf("len=2 should pass for map with 2 items, got %v", err)
	}

	invalid := Test{Meta: map[string]string{"a": "1"}}
	if err := Validate(invalid); err == nil {
		t.Error("len=2 should fail for map with 1 item")
	}
}
