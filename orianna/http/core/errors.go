package core

import (
	"errors"

	oerrors "github.com/anthanhphan/gosdk/orianna/shared/errors"
)

// HTTP-specific sentinel errors.
var (
	ErrRouteNotFound  = errors.New("route not found")
	ErrEmptyRoutePath = errors.New("route path cannot be empty")
	ErrDuplicateRoute = errors.New("duplicate route")
	ErrInvalidMethod  = errors.New("invalid HTTP method")
	ErrEmptyPrefix    = errors.New("prefix cannot be empty")
	ErrNoRoutes       = errors.New("group must have at least one route")
	ErrNilValidator   = errors.New("validator cannot be nil")
	ErrTimeout        = errors.New("request timeout")
	ErrRateLimited    = errors.New("rate limit exceeded")
)

// Re-export shared sentinel errors for convenience.
// Users can access these via core.ErrCircuitOpen instead of importing shared/errors.
var (
	ErrInvalidConfig    = oerrors.ErrInvalidConfig
	ErrHandlerNil       = oerrors.ErrHandlerNil
	ErrServerNotStarted = oerrors.ErrServerNotStarted
	ErrServerShutdown   = oerrors.ErrServerShutdown
	ErrNilChecker       = oerrors.ErrNilChecker
	ErrCircuitOpen      = oerrors.ErrCircuitOpen
)

// IsRouteError checks if an error is a route-related error.
func IsRouteError(err error) bool {
	return errors.Is(err, ErrRouteNotFound) ||
		errors.Is(err, ErrDuplicateRoute) ||
		errors.Is(err, ErrEmptyRoutePath)
}

// IsValidationRelatedError checks if an error is validation-related.
func IsValidationRelatedError(err error) bool {
	return errors.Is(err, ErrNilValidator)
}

// Re-export shared error checkers for convenience.
var (
	IsConfigError = oerrors.IsConfigError
	IsServerError = oerrors.IsServerError
)
