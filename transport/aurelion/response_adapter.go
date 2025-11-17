package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/response"
)

// responseContextAdapter adapts aurelion.Context to response.ContextInterface.
type responseContextAdapter struct {
	ctx Context
}

// Status implements response.ContextInterface.
func (a *responseContextAdapter) Status(status int) response.ContextInterface {
	a.ctx.Status(status)
	return a
}

// JSON implements response.ContextInterface.
func (a *responseContextAdapter) JSON(data interface{}) error {
	return a.ctx.JSON(data)
}

// Locals implements response.ContextInterface.
func (a *responseContextAdapter) Locals(key string, value ...interface{}) interface{} {
	return a.ctx.Locals(key, value...)
}

// Next implements response.ContextInterface.
func (a *responseContextAdapter) Next() error {
	return a.ctx.Next()
}

// GetAllLocals implements response.ContextInterface.
func (a *responseContextAdapter) GetAllLocals() map[string]interface{} {
	return a.ctx.GetAllLocals()
}

// adaptContextForResponse adapts aurelion.Context to response.ContextInterface.
func adaptContextForResponse(ctx Context) response.ContextInterface {
	if ctx == nil {
		return nil
	}
	return &responseContextAdapter{ctx: ctx}
}

// Response adapter functions to convert between response types and public aurelion types.

// OKPublic wraps response.OK to accept aurelion.Context.
func OKPublic(ctx Context, message string, data ...interface{}) error {
	return response.OK(adaptContextForResponse(ctx), message, data...)
}

// ErrorPublic wraps response.Error to accept aurelion.Context.
func ErrorPublic(ctx Context, err error) error {
	return response.Error(adaptContextForResponse(ctx), err)
}

// BadRequestPublic wraps response.BadRequest to accept aurelion.Context.
func BadRequestPublic(ctx Context, message string) error {
	return response.BadRequest(adaptContextForResponse(ctx), message)
}

// UnauthorizedPublic wraps response.Unauthorized to accept aurelion.Context.
func UnauthorizedPublic(ctx Context, message string) error {
	return response.Unauthorized(adaptContextForResponse(ctx), message)
}

// ForbiddenPublic wraps response.Forbidden to accept aurelion.Context.
func ForbiddenPublic(ctx Context, message string) error {
	return response.Forbidden(adaptContextForResponse(ctx), message)
}

// NotFoundPublic wraps response.NotFound to accept aurelion.Context.
func NotFoundPublic(ctx Context, message string) error {
	return response.NotFound(adaptContextForResponse(ctx), message)
}

// InternalServerErrorPublic wraps response.InternalServerError to accept aurelion.Context.
func InternalServerErrorPublic(ctx Context, message string) error {
	return response.InternalServerError(adaptContextForResponse(ctx), message)
}

// HealthCheckPublic wraps response.HealthCheck to accept aurelion.Context.
func HealthCheckPublic(ctx Context) error {
	return response.HealthCheck(adaptContextForResponse(ctx))
}

// ErrorWithDetailsPublic wraps response.ErrorWithDetails to accept aurelion.Context.
func ErrorWithDetailsPublic(ctx Context, code int, message string, errorData *response.ErrorData) error {
	return response.ErrorWithDetails(adaptContextForResponse(ctx), code, message, errorData)
}
