// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import oerrors "github.com/anthanhphan/gosdk/orianna/shared/errors"

// Re-export shared sentinel errors for backward compatibility.
var (
	ErrInvalidConfig    = oerrors.ErrInvalidConfig
	ErrHandlerNil       = oerrors.ErrHandlerNil
	ErrServerNotStarted = oerrors.ErrServerNotStarted
	ErrServerShutdown   = oerrors.ErrServerShutdown
	ErrNilChecker       = oerrors.ErrNilChecker
)

// IsConfigError checks if an error is a configuration error.
var IsConfigError = oerrors.IsConfigError

// IsServerError checks if an error is a server-related error.
var IsServerError = oerrors.IsServerError
