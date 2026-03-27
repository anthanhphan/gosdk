// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"testing"
)

// Tests for query.go helpers

func TestGetQueryInt_DefaultValue(t *testing.T) {
	ctx := NewMockContext()
	result := GetQueryInt(ctx, "page", 1)
	if result != 1 {
		t.Fatalf("expected default 1, got %d", result)
	}
}

func TestGetQueryInt_ValidValue(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddQuery("page", "5")

	result := GetQueryInt(ctx, "page", 1)
	if result != 5 {
		t.Fatalf("expected 5, got %d", result)
	}
}

func TestGetQueryInt_InvalidValue(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddQuery("page", "not-a-number")

	result := GetQueryInt(ctx, "page", 1)
	if result != 1 {
		t.Fatalf("expected default 1 for invalid input, got %d", result)
	}
}

func TestGetQueryInt64_DefaultValue(t *testing.T) {
	ctx := NewMockContext()
	result := GetQueryInt64(ctx, "id", 100)
	if result != 100 {
		t.Fatalf("expected default 100, got %d", result)
	}
}

func TestGetQueryInt64_ValidValue(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddQuery("id", "9999999999")

	result := GetQueryInt64(ctx, "id", 0)
	if result != 9999999999 {
		t.Fatalf("expected 9999999999, got %d", result)
	}
}

func TestGetQueryInt64_InvalidValue(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddQuery("id", "abc")

	result := GetQueryInt64(ctx, "id", 42)
	if result != 42 {
		t.Fatalf("expected default 42 for invalid input, got %d", result)
	}
}

func TestGetQueryBool_TrueValues(t *testing.T) {
	truthy := []string{"true", "1", "yes", "on"}
	for _, v := range truthy {
		ctx := NewMockContext()
		ctx.AddQuery("flag", v)

		result := GetQueryBool(ctx, "flag", false)
		if !result {
			t.Fatalf("expected true for %q, got false", v)
		}
	}
}

func TestGetQueryBool_FalseValues(t *testing.T) {
	falsy := []string{"false", "0", "no", "off"}
	for _, v := range falsy {
		ctx := NewMockContext()
		ctx.AddQuery("flag", v)

		result := GetQueryBool(ctx, "flag", true)
		if result {
			t.Fatalf("expected false for %q, got true", v)
		}
	}
}

func TestGetQueryBool_DefaultValue(t *testing.T) {
	ctx := NewMockContext()
	result := GetQueryBool(ctx, "flag", true)
	if !result {
		t.Fatal("expected default true, got false")
	}
}

func TestGetQueryBool_UnknownValue(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddQuery("flag", "maybe")

	result := GetQueryBool(ctx, "flag", false)
	if result {
		t.Fatal("expected default false for unknown value, got true")
	}
}

func TestGetQueryString_DefaultValue(t *testing.T) {
	ctx := NewMockContext()
	result := GetQueryString(ctx, "sort", "created_at")
	if result != "created_at" {
		t.Fatalf("expected 'created_at', got %q", result)
	}
}

func TestGetQueryString_ValidValue(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddQuery("sort", "name")

	result := GetQueryString(ctx, "sort", "created_at")
	if result != "name" {
		t.Fatalf("expected 'name', got %q", result)
	}
}

// Tests for params.go helpers

func TestGetParamInt_ValidParam(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddParam("id", "42")

	result, err := GetParamInt(ctx, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

func TestGetParamInt_InvalidParam(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddParam("id", "abc")

	_, err := GetParamInt(ctx, "id")
	if err == nil {
		t.Fatal("expected error for invalid integer param")
	}
}

func TestGetParamInt_MissingParam(t *testing.T) {
	ctx := NewMockContext()

	_, err := GetParamInt(ctx, "id")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}

func TestGetParamInt64_ValidParam(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddParam("id", "9999999999")

	result, err := GetParamInt64(ctx, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 9999999999 {
		t.Fatalf("expected 9999999999, got %d", result)
	}
}

func TestGetParamInt64_MissingParam(t *testing.T) {
	ctx := NewMockContext()

	_, err := GetParamInt64(ctx, "id")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}

func TestGetParamUUID_ValidParam(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddParam("id", "550e8400-e29b-41d4-a716-446655440000")

	result, err := GetParamUUID(ctx, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected UUID: %s", result)
	}
}

func TestGetParamUUID_InvalidParam(t *testing.T) {
	ctx := NewMockContext()
	ctx.AddParam("id", "not-a-uuid")

	_, err := GetParamUUID(ctx, "id")
	if err == nil {
		t.Fatal("expected error for invalid UUID param")
	}
}

func TestGetParamUUID_MissingParam(t *testing.T) {
	ctx := NewMockContext()

	_, err := GetParamUUID(ctx, "id")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}
