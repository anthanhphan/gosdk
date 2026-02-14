// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BindSource Tests

func TestDefaultBindOptions(t *testing.T) {
	opts := DefaultBindOptions()
	assert.Equal(t, BindSourceBody, opts.Source)
	assert.True(t, opts.Validate)
}

// Bind Tests

type BindTestRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
}

func TestBind_FromBody_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})

	result, err := Bind[BindTestRequest](mockCtx)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
}

func TestBind_FromBody_ValidationFailure(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "Jo", // Too short
		"email": "invalid-email",
	})

	_, err := Bind[BindTestRequest](mockCtx)
	require.Error(t, err)

	var validationErrs validator.ValidationErrors
	assert.ErrorAs(t, err, &validationErrs)
}

func TestBind_FromQuery_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("name", "John Doe")
	mockCtx.AddQuery("email", "john@example.com")

	opts := BindOptions{Source: BindSourceQuery, Validate: true}
	result, err := Bind[BindTestRequest](mockCtx, opts)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

func TestBind_FromParams_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("name", "John Doe")
	mockCtx.AddParam("email", "john@example.com")

	opts := BindOptions{Source: BindSourceParams, Validate: true}
	result, err := Bind[BindTestRequest](mockCtx, opts)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

func TestBind_FromHeaders_NotSupported(t *testing.T) {
	mockCtx := NewMockContext()

	opts := BindOptions{Source: BindSourceHeaders, Validate: false}
	_, err := Bind[BindTestRequest](mockCtx, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "headers not yet supported")
}

func TestBind_ParseError(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyParseError("invalid JSON")

	_, err := Bind[BindTestRequest](mockCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse request")
}

func TestBind_WithoutValidation(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "Jo", // Would fail validation but validation is disabled
		"email": "invalid",
	})

	opts := BindOptions{Source: BindSourceBody, Validate: false}
	result, err := Bind[BindTestRequest](mockCtx, opts)
	require.NoError(t, err)
	assert.Equal(t, "Jo", result.Name)
}

// MustBind Tests

func TestMustBind_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})

	result, ok := MustBind[BindTestRequest](mockCtx)
	assert.True(t, ok)
	assert.Equal(t, "John Doe", result.Name)
}

func TestMustBind_Failure(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "Jo", // Too short
		"email": "invalid",
	})

	_, ok := MustBind[BindTestRequest](mockCtx)
	assert.False(t, ok)
	// Verify error response was sent
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestMustBind_ParseError(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyParseError("invalid JSON")

	_, ok := MustBind[BindTestRequest](mockCtx)
	assert.False(t, ok)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

// BindBody Tests

func TestBindBody_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})

	result, err := BindBody[BindTestRequest](mockCtx, true)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

func TestBindBody_WithoutValidation(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name":  "Jo",
		"email": "invalid",
	})

	result, err := BindBody[BindTestRequest](mockCtx, false)
	require.NoError(t, err)
	assert.Equal(t, "Jo", result.Name)
}

// BindQuery Tests

func TestBindQuery_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddQuery("name", "John Doe")
	mockCtx.AddQuery("email", "john@example.com")

	result, err := BindQuery[BindTestRequest](mockCtx, true)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

// BindParams Tests

func TestBindParams_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.AddParam("name", "John Doe")
	mockCtx.AddParam("email", "john@example.com")

	result, err := BindParams[BindTestRequest](mockCtx, true)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

// handleBindError Tests

func TestHandleBindError_ValidationError(t *testing.T) {
	mockCtx := NewMockContext()

	// Create a validation error
	validationErr := validator.ValidationError{
		Field:   "name",
		Message: "name is required",
	}

	handleBindError(mockCtx, validator.ValidationErrors{validationErr})

	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestHandleBindError_GenericError(t *testing.T) {
	mockCtx := NewMockContext()

	handleBindError(mockCtx, WrapError(ErrInvalidConfig, "generic error"))

	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}
