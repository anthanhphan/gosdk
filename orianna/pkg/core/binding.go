// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"fmt"

	"github.com/anthanhphan/gosdk/orianna/pkg/validator"
)

// Bind Source Types

// BindSource specifies where to bind data from
type BindSource int

const (
	// BindSourceBody binds from request body (JSON, XML, etc)
	BindSourceBody BindSource = iota
	// BindSourceQuery binds from query parameters
	BindSourceQuery
	// BindSourceParams binds from route parameters
	BindSourceParams
	// BindSourceHeaders binds from request headers
	BindSourceHeaders
)

// Bind Options

// BindOptions configures binding behavior
type BindOptions struct {
	// Source specifies where to bind data from (default: Body)
	Source BindSource
	// Validate enables automatic validation after binding
	Validate bool
}

// DefaultBindOptions returns default binding options
func DefaultBindOptions() BindOptions {
	return BindOptions{
		Source:   BindSourceBody,
		Validate: true,
	}
}

// Generic Binding Functions

// Bind parses request data into a typed struct with optional validation.
// This is the recommended way to parse requests with type safety.
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name" validate:"required,min=3"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	user, err := orianna.Bind[CreateUserRequest](ctx)
//	if err != nil {
//	    return orianna.SendValidationError(ctx, err)
//	}
func Bind[T any](ctx Context, opts ...BindOptions) (T, error) {
	var result T
	opt := DefaultBindOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}

	var parseErr error
	switch opt.Source {
	case BindSourceBody:
		parseErr = ctx.BodyParser(&result)
	case BindSourceQuery:
		parseErr = ctx.QueryParser(&result)
	case BindSourceParams:
		parseErr = ctx.ParamsParser(&result)
	case BindSourceHeaders:
		// Headers binding is not yet implemented
		parseErr = fmt.Errorf("bind source headers not yet supported")
	default:
		parseErr = ctx.BodyParser(&result)
	}

	if parseErr != nil {
		return result, WrapError(parseErr, "failed to parse request")
	}

	if opt.Validate {
		if validationErr := validator.Validate(result); validationErr != nil {
			return result, validationErr
		}
	}

	return result, nil
}

// MustBind parses the request and automatically sends error response on failure.
// Returns the parsed value and a boolean indicating success.
// If false is returned, an error response has already been sent.
//
// Example:
//
//	user, ok := orianna.MustBind[CreateUserRequest](ctx)
//	if !ok {
//	    return nil // Error response already sent
//	}
//	// Use user...
func MustBind[T any](ctx Context, opts ...BindOptions) (T, bool) {
	result, err := Bind[T](ctx, opts...)
	if err != nil {
		handleBindError(ctx, err)
		return result, false
	}
	return result, true
}

// BindBody is a shorthand for binding from request body
func BindBody[T any](ctx Context, validate bool) (T, error) {
	return Bind[T](ctx, BindOptions{Source: BindSourceBody, Validate: validate})
}

// BindQuery is a shorthand for binding from query parameters
func BindQuery[T any](ctx Context, validate bool) (T, error) {
	return Bind[T](ctx, BindOptions{Source: BindSourceQuery, Validate: validate})
}

// BindParams is a shorthand for binding from route parameters
func BindParams[T any](ctx Context, validate bool) (T, error) {
	return Bind[T](ctx, BindOptions{Source: BindSourceParams, Validate: validate})
}

// Error Handling

// handleBindError sends appropriate error response based on error type.
// Uses errors.As for proper support of wrapped errors.
func handleBindError(ctx Context, err error) {
	// Check if it's a validation error (supports wrapped errors)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		_ = sendValidationErrorResponse(ctx, validationErrors)
		return
	}

	// Check if single validation error (supports wrapped errors)
	var validationError *validator.ValidationError
	if errors.As(err, &validationError) {
		_ = sendValidationErrorResponse(ctx, validator.ValidationErrors{*validationError})
		return
	}

	// Default to bad request
	errResp := NewErrorResponse("BAD_REQUEST", StatusBadRequest, err.Error())
	_ = SendError(ctx, errResp)
}

// sendValidationErrorResponse sends a validation error response.
// This is a shared helper to avoid code duplication between handleBindError and handleTypedError.
func sendValidationErrorResponse(ctx Context, errs validator.ValidationErrors) error {
	resp := NewErrorResponse("VALIDATION_FAILED", StatusBadRequest, "Validation failed")
	resp.Details = Map{"errors": errs.ToArray()}
	return SendError(ctx, resp)
}
