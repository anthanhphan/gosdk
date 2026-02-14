// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Method Tests

func TestMethod_String(t *testing.T) {
	tests := []struct {
		method   Method
		expected string
	}{
		{GET, "GET"},
		{POST, "POST"},
		{PUT, "PUT"},
		{PATCH, "PATCH"},
		{DELETE, "DELETE"},
		{HEAD, "HEAD"},
		{OPTIONS, "OPTIONS"},
		{Method(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.String())
		})
	}
}

// BindBody Additional Tests (complements binding_test.go)

type TestRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=0,max=150"`
}

func TestBindBody_WithValidation_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	})

	result, err := BindBody[TestRequest](mockCtx, true)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

func TestBindBody_WithValidation_Failure(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "Jo", // Too short
		"email": "invalid-email",
		"age":   30,
	})

	_, err := BindBody[TestRequest](mockCtx, true)
	require.Error(t, err)

	var validationErrs validator.ValidationErrors
	assert.ErrorAs(t, err, &validationErrs)
}

func TestBindBody_ParseError(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyParseError("invalid JSON")

	_, err := BindBody[TestRequest](mockCtx, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse request")
}

// ValidateAndRespond Tests

func TestValidateAndRespond_Success(t *testing.T) {
	mockCtx := NewMockContext()
	req := TestRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	ok, err := ValidateAndRespond(mockCtx, req)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestValidateAndRespond_Failure(t *testing.T) {
	mockCtx := NewMockContext()
	req := TestRequest{
		Name:  "Jo", // Too short
		Email: "invalid",
		Age:   30,
	}

	ok, err := ValidateAndRespond(mockCtx, req)
	assert.False(t, ok)
	assert.NoError(t, err) // No error returned, but response is sent
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

// MustValidate Tests

func TestMustValidate_Success(t *testing.T) {
	req := TestRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	assert.NotPanics(t, func() {
		MustValidate(req)
	})
}

func TestMustValidate_Panic(t *testing.T) {
	req := TestRequest{
		Name:  "Jo", // Too short
		Email: "invalid",
		Age:   30,
	}

	assert.Panics(t, func() {
		MustValidate(req)
	})
}

// GetParamInt Tests

func TestGetParamInt_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("id", "123")

	result, err := GetParamInt(mockCtx, "id")
	require.NoError(t, err)
	assert.Equal(t, 123, result)
}

func TestGetParamInt_NotFound(t *testing.T) {
	mockCtx := NewMockContext()

	_, err := GetParamInt(mockCtx, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetParamInt_InvalidFormat(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("id", "abc")

	_, err := GetParamInt(mockCtx, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid integer")
}

// GetParamInt64 Tests

func TestGetParamInt64_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("id", "9223372036854775807")

	result, err := GetParamInt64(mockCtx, "id")
	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), result)
}

func TestGetParamInt64_NotFound(t *testing.T) {
	mockCtx := NewMockContext()

	_, err := GetParamInt64(mockCtx, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetParamInt64_InvalidFormat(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("id", "not-a-number")

	_, err := GetParamInt64(mockCtx, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid integer")
}

// GetParamUUID Tests

func TestGetParamUUID_Success(t *testing.T) {
	mockCtx := NewMockContext()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	mockCtx.AddParam("id", uuid)

	result, err := GetParamUUID(mockCtx, "id")
	require.NoError(t, err)
	assert.Equal(t, uuid, result)
}

func TestGetParamUUID_NotFound(t *testing.T) {
	mockCtx := NewMockContext()

	_, err := GetParamUUID(mockCtx, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetParamUUID_InvalidFormat(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("id", "not-a-uuid")

	_, err := GetParamUUID(mockCtx, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid UUID")
}

// GetQueryInt Tests

func TestGetQueryInt_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("page", "10")

	result := GetQueryInt(mockCtx, "page", 1)
	assert.Equal(t, 10, result)
}

func TestGetQueryInt_NotFound_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()

	result := GetQueryInt(mockCtx, "page", 1)
	assert.Equal(t, 1, result)
}

func TestGetQueryInt_InvalidFormat_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("page", "abc")

	result := GetQueryInt(mockCtx, "page", 1)
	assert.Equal(t, 1, result)
}

// GetQueryInt64 Tests

func TestGetQueryInt64_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("offset", "1000000")

	result := GetQueryInt64(mockCtx, "offset", 0)
	assert.Equal(t, int64(1000000), result)
}

func TestGetQueryInt64_NotFound_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()

	result := GetQueryInt64(mockCtx, "offset", 100)
	assert.Equal(t, int64(100), result)
}

func TestGetQueryInt64_InvalidFormat_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("offset", "not-a-number")

	result := GetQueryInt64(mockCtx, "offset", 100)
	assert.Equal(t, int64(100), result)
}

// GetQueryBool Tests

func TestGetQueryBool_TruthyValues(t *testing.T) {
	truthy := []string{"true", "1", "yes", "on"}
	mockCtx := NewMockContext()

	for _, val := range truthy {
		t.Run(val, func(t *testing.T) {
			mockCtx.AddQuery("flag", val)
			result := GetQueryBool(mockCtx, "flag", false)
			assert.True(t, result)
		})
	}
}

func TestGetQueryBool_FalsyValues(t *testing.T) {
	falsy := []string{"false", "0", "no", "off"}
	mockCtx := NewMockContext()

	for _, val := range falsy {
		t.Run(val, func(t *testing.T) {
			mockCtx.AddQuery("flag", val)
			result := GetQueryBool(mockCtx, "flag", true)
			assert.False(t, result)
		})
	}
}

func TestGetQueryBool_NotFound_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()

	result := GetQueryBool(mockCtx, "flag", true)
	assert.True(t, result)
}

func TestGetQueryBool_InvalidValue_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("flag", "maybe")

	result := GetQueryBool(mockCtx, "flag", true)
	assert.True(t, result)
}

// GetQueryString Tests

func TestGetQueryString_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("sort", "name")

	result := GetQueryString(mockCtx, "sort", "created_at")
	assert.Equal(t, "name", result)
}

func TestGetQueryString_NotFound_ReturnsDefault(t *testing.T) {
	mockCtx := NewMockContext()

	result := GetQueryString(mockCtx, "sort", "created_at")
	assert.Equal(t, "created_at", result)
}
