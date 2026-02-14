// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"

	"github.com/anthanhphan/gosdk/orianna/pkg/validator"
)

// TypedHandler creates a handler with automatic request binding and response handling.
//
// Input:
//   - fn: Handler function that receives context and typed request, returns typed response
//
// Output:
//   - Handler: A handler with automatic marshaling
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name" validate:"required"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	type CreateUserResponse struct {
//	    ID   string `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	var createUser = core.TypedHandler(func(ctx core.Context, req CreateUserRequest) (CreateUserResponse, error) {
//	    user := service.CreateUser(req.Name, req.Email)
//	    return CreateUserResponse{ID: user.ID, Name: user.Name}, nil
//	})
func TypedHandler[Req, Resp any](fn func(Context, Req) (Resp, error)) Handler {
	return func(ctx Context) error {
		req, ok := MustBind[Req](ctx)
		if !ok {
			return nil
		}

		resp, err := fn(ctx, req)
		if err != nil {
			return handleTypedError(ctx, err)
		}

		return SendSuccess(ctx, NewSuccessResponse(StatusOK, "", resp))
	}
}

// SimpleHandler creates a handler with full control over the response.
//
// Input:
//   - fn: Handler function
//
// Output:
//   - Handler: The same handler function
//
// Example:
//
//	server.GET("/health", core.SimpleHandler(func(ctx core.Context) error {
//	    return core.SendSuccess(ctx, core.NewSuccessResponse(200, "OK", nil))
//	}))
func SimpleHandler(fn func(Context) error) Handler {
	return fn
}

func handleTypedError(ctx Context, err error) error {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		return SendError(ctx, errResp)
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return sendValidationErrorResponse(ctx, validationErrors)
	}

	var validationError *validator.ValidationError
	if errors.As(err, &validationError) {
		return sendValidationErrorResponse(ctx, validator.ValidationErrors{*validationError})
	}

	resp := NewErrorResponse("INTERNAL_ERROR", StatusInternalServerError, "An internal error occurred")
	return SendError(ctx, resp)
}
