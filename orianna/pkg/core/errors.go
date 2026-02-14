// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "errors"

// Sentinel errors for internal use
var (
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrRouteNotFound    = errors.New("route not found")
	ErrHandlerNil       = errors.New("handler cannot be nil")
	ErrEmptyRoutePath   = errors.New("route path cannot be empty")
	ErrDuplicateRoute   = errors.New("duplicate route")
	ErrServerNotStarted = errors.New("server not started")
	ErrServerShutdown   = errors.New("server shutdown")
	ErrInvalidMethod    = errors.New("invalid HTTP method")
	ErrEmptyPrefix      = errors.New("prefix cannot be empty")
	ErrNoRoutes         = errors.New("group must have at least one route")
	ErrNilValidator     = errors.New("validator cannot be nil")
	ErrNilChecker       = errors.New("health checker cannot be nil")
	ErrTimeout          = errors.New("request timeout")
	ErrRateLimited      = errors.New("rate limit exceeded")
)

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

// IsRouteError checks if an error is a route-related error.
func IsRouteError(err error) bool {
	return errors.Is(err, ErrRouteNotFound) ||
		errors.Is(err, ErrDuplicateRoute) ||
		errors.Is(err, ErrEmptyRoutePath)
}

// IsServerError checks if an error is a server-related error.
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerNotStarted) ||
		errors.Is(err, ErrServerShutdown)
}

// IsValidationRelatedError checks if an error is validation-related.
func IsValidationRelatedError(err error) bool {
	return errors.Is(err, ErrNilValidator) ||
		errors.Is(err, ErrNilChecker)
}
