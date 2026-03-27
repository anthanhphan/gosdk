// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Sentinel Error Tests

func TestSentinelErrors_Defined(t *testing.T) {
	// Verify all HTTP-specific sentinel errors are defined
	assert.NotNil(t, ErrRouteNotFound)
	assert.NotNil(t, ErrEmptyRoutePath)
	assert.NotNil(t, ErrDuplicateRoute)
	assert.NotNil(t, ErrInvalidMethod)
	assert.NotNil(t, ErrEmptyPrefix)
	assert.NotNil(t, ErrNoRoutes)
	assert.NotNil(t, ErrNilValidator)
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrRateLimited)
}

// IsRouteError Tests

func TestIsRouteError_RouteNotFound(t *testing.T) {
	assert.True(t, IsRouteError(ErrRouteNotFound))
}

func TestIsRouteError_DuplicateRoute(t *testing.T) {
	assert.True(t, IsRouteError(ErrDuplicateRoute))
}

func TestIsRouteError_EmptyRoutePath(t *testing.T) {
	assert.True(t, IsRouteError(ErrEmptyRoutePath))
}

func TestIsRouteError_Wrapped(t *testing.T) {
	err := WrapError(ErrDuplicateRoute, "wrapped route error")
	assert.True(t, IsRouteError(err))
}

func TestIsRouteError_False(t *testing.T) {
	err := ErrNilValidator
	assert.False(t, IsRouteError(err))
}

func TestIsRouteError_Nil(t *testing.T) {
	assert.False(t, IsRouteError(nil))
}

// IsValidationRelatedError Tests

func TestIsValidationRelatedError_NilValidator(t *testing.T) {
	assert.True(t, IsValidationRelatedError(ErrNilValidator))
}

func TestIsValidationRelatedError_Wrapped(t *testing.T) {
	err := WrapError(ErrNilValidator, "wrapped validation error")
	assert.True(t, IsValidationRelatedError(err))
}

func TestIsValidationRelatedError_False(t *testing.T) {
	err := ErrRouteNotFound
	assert.False(t, IsValidationRelatedError(err))
}

func TestIsValidationRelatedError_Nil(t *testing.T) {
	assert.False(t, IsValidationRelatedError(nil))
}

// Error Wrapping Tests

func TestWrapError_Unwrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := WrapError(original, "wrapped")

	assert.True(t, errors.Is(wrapped, original))
}

func TestWrapErrorf_Unwrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := WrapErrorf(original, "wrapped with %s", "formatting")

	assert.True(t, errors.Is(wrapped, original))
	assert.Contains(t, wrapped.Error(), "formatting")
}
