// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TypedHandler Tests

type HandlerTestRequest struct {
	Name string `json:"name" validate:"required,min=3"`
}

type HandlerTestResponse struct {
	Message string `json:"message"`
}

func TestTypedHandler_Success(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name": "John Doe",
	})

	handler := TypedHandler(func(ctx Context, req HandlerTestRequest) (HandlerTestResponse, error) {
		return HandlerTestResponse{Message: "Hello " + req.Name}, nil
	})

	err := handler(mockCtx)
	require.NoError(t, err)
	assert.Equal(t, StatusOK, mockCtx.ResponseStatusCode())
}

func TestTypedHandler_BindFailure(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name": "Jo", // Too short
	})

	handler := TypedHandler(func(ctx Context, req HandlerTestRequest) (HandlerTestResponse, error) {
		return HandlerTestResponse{Message: "Hello " + req.Name}, nil
	})

	err := handler(mockCtx)
	// MustBind returns nil error when binding fails (response already sent)
	assert.Nil(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestTypedHandler_HandlerError(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name": "John Doe",
	})

	handler := TypedHandler(func(ctx Context, req HandlerTestRequest) (HandlerTestResponse, error) {
		return HandlerTestResponse{}, errors.New("handler error")
	})

	err := handler(mockCtx)
	require.NoError(t, err) // Error is handled internally
	assert.Equal(t, StatusInternalServerError, mockCtx.ResponseStatusCode())
}

func TestTypedHandler_ErrorResponseError(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name": "John Doe",
	})

	errResp := NewErrorResponse("CUSTOM_ERROR", StatusBadRequest, "Custom error message")
	handler := TypedHandler(func(ctx Context, req HandlerTestRequest) (HandlerTestResponse, error) {
		return HandlerTestResponse{}, errResp
	})

	err := handler(mockCtx)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestTypedHandler_ValidationError(t *testing.T) {
	mockCtx := NewMockContext()
	mockCtx.SetBodyJSON(map[string]any{
		"name": "John Doe",
	})

	validationErr := validator.ValidationError{
		Field:   "email",
		Message: "email is required",
	}

	handler := TypedHandler(func(ctx Context, req HandlerTestRequest) (HandlerTestResponse, error) {
		return HandlerTestResponse{}, validator.ValidationErrors{validationErr}
	})

	err := handler(mockCtx)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

// SimpleHandler Tests

func TestSimpleHandler_Success(t *testing.T) {
	mockCtx := NewMockContext()

	handler := SimpleHandler(func(ctx Context) error {
		return ctx.OK(Map{"message": "success"})
	})

	err := handler(mockCtx)
	require.NoError(t, err)
	assert.Equal(t, StatusOK, mockCtx.ResponseStatusCode())
}

func TestSimpleHandler_Error(t *testing.T) {
	mockCtx := NewMockContext()

	handler := SimpleHandler(func(ctx Context) error {
		return errors.New("simple error")
	})

	err := handler(mockCtx)
	require.Error(t, err)
	assert.Equal(t, "simple error", err.Error())
}

// handleTypedError Tests

func TestHandleTypedError_ErrorResponse(t *testing.T) {
	mockCtx := NewMockContext()
	errResp := NewErrorResponse("TEST_ERROR", StatusBadRequest, "Test error")

	err := handleTypedError(mockCtx, errResp)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestHandleTypedError_WrappedErrorResponse(t *testing.T) {
	mockCtx := NewMockContext()
	errResp := NewErrorResponse("TEST_ERROR", StatusBadRequest, "Test error")
	wrappedErr := WrapError(errResp, "wrapped")

	err := handleTypedError(mockCtx, wrappedErr)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestHandleTypedError_ValidationErrors(t *testing.T) {
	mockCtx := NewMockContext()
	validationErr := validator.ValidationError{
		Field:   "name",
		Message: "name is required",
	}
	errs := validator.ValidationErrors{validationErr}

	err := handleTypedError(mockCtx, errs)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestHandleTypedError_SingleValidationError(t *testing.T) {
	mockCtx := NewMockContext()
	validationErr := &validator.ValidationError{
		Field:   "name",
		Message: "name is required",
	}

	err := handleTypedError(mockCtx, validationErr)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}

func TestHandleTypedError_GenericError(t *testing.T) {
	mockCtx := NewMockContext()
	genericErr := errors.New("generic error")

	err := handleTypedError(mockCtx, genericErr)
	require.NoError(t, err)
	assert.Equal(t, StatusInternalServerError, mockCtx.ResponseStatusCode())
}

// sendValidationErrorResponse Tests

func TestSendValidationErrorResponse(t *testing.T) {
	mockCtx := NewMockContext()
	validationErr := validator.ValidationError{
		Field:   "name",
		Message: "name is required",
	}
	errs := validator.ValidationErrors{validationErr}

	err := sendValidationErrorResponse(mockCtx, errs)
	require.NoError(t, err)
	assert.Equal(t, StatusBadRequest, mockCtx.ResponseStatusCode())
}
