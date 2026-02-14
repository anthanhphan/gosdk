// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Sentinel Error Tests

func TestSentinelErrors_Defined(t *testing.T) {
	// Verify all sentinel errors are defined
	assert.NotNil(t, ErrInvalidConfig)
	assert.NotNil(t, ErrRouteNotFound)
	assert.NotNil(t, ErrHandlerNil)
	assert.NotNil(t, ErrEmptyRoutePath)
	assert.NotNil(t, ErrDuplicateRoute)
	assert.NotNil(t, ErrServerNotStarted)
	assert.NotNil(t, ErrServerShutdown)
	assert.NotNil(t, ErrInvalidMethod)
	assert.NotNil(t, ErrEmptyPrefix)
	assert.NotNil(t, ErrNoRoutes)
	assert.NotNil(t, ErrNilValidator)
	assert.NotNil(t, ErrNilChecker)
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrRateLimited)
}

// IsConfigError Tests

func TestIsConfigError_True(t *testing.T) {
	err := ErrInvalidConfig
	assert.True(t, IsConfigError(err))
}

func TestIsConfigError_Wrapped(t *testing.T) {
	err := WrapError(ErrInvalidConfig, "wrapped config error")
	assert.True(t, IsConfigError(err))
}

func TestIsConfigError_False(t *testing.T) {
	err := ErrRouteNotFound
	assert.False(t, IsConfigError(err))
}

func TestIsConfigError_Nil(t *testing.T) {
	assert.False(t, IsConfigError(nil))
}

// IsRouteError Tests

func TestIsRouteError_RouteNotFound(t *testing.T) {
	err := ErrRouteNotFound
	assert.True(t, IsRouteError(err))
}

func TestIsRouteError_DuplicateRoute(t *testing.T) {
	err := ErrDuplicateRoute
	assert.True(t, IsRouteError(err))
}

func TestIsRouteError_EmptyRoutePath(t *testing.T) {
	err := ErrEmptyRoutePath
	assert.True(t, IsRouteError(err))
}

func TestIsRouteError_Wrapped(t *testing.T) {
	err := WrapError(ErrDuplicateRoute, "wrapped route error")
	assert.True(t, IsRouteError(err))
}

func TestIsRouteError_False(t *testing.T) {
	err := ErrInvalidConfig
	assert.False(t, IsRouteError(err))
}

func TestIsRouteError_Nil(t *testing.T) {
	assert.False(t, IsRouteError(nil))
}

// IsServerError Tests

func TestIsServerError_ServerNotStarted(t *testing.T) {
	err := ErrServerNotStarted
	assert.True(t, IsServerError(err))
}

func TestIsServerError_ServerShutdown(t *testing.T) {
	err := ErrServerShutdown
	assert.True(t, IsServerError(err))
}

func TestIsServerError_Wrapped(t *testing.T) {
	err := WrapError(ErrServerShutdown, "wrapped server error")
	assert.True(t, IsServerError(err))
}

func TestIsServerError_False(t *testing.T) {
	err := ErrInvalidConfig
	assert.False(t, IsServerError(err))
}

func TestIsServerError_Nil(t *testing.T) {
	assert.False(t, IsServerError(nil))
}

// IsValidationRelatedError Tests

func TestIsValidationRelatedError_NilValidator(t *testing.T) {
	err := ErrNilValidator
	assert.True(t, IsValidationRelatedError(err))
}

func TestIsValidationRelatedError_NilChecker(t *testing.T) {
	err := ErrNilChecker
	assert.True(t, IsValidationRelatedError(err))
}

func TestIsValidationRelatedError_Wrapped(t *testing.T) {
	err := WrapError(ErrNilValidator, "wrapped validation error")
	assert.True(t, IsValidationRelatedError(err))
}

func TestIsValidationRelatedError_False(t *testing.T) {
	err := ErrInvalidConfig
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
