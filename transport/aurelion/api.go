package aurelion

import (
	stdctx "context"

	"github.com/anthanhphan/gosdk/transport/aurelion/internal/context"
	"github.com/anthanhphan/gosdk/transport/aurelion/middleware"
	"github.com/anthanhphan/gosdk/transport/aurelion/response"
	"github.com/anthanhphan/gosdk/transport/aurelion/router"
	"github.com/anthanhphan/gosdk/transport/aurelion/validation"
)

// Re-export types from sub-packages for convenience.
type (
	APIResponse       = response.APIResponse
	ErrorData         = response.ErrorData
	BusinessError     = response.BusinessError
	ValidationErrors  = validation.ValidationErrors
	Route             = router.Route
	GroupRoute        = router.GroupRoute
	RouteBuilder      = router.RouteBuilder
	GroupRouteBuilder = router.GroupRouteBuilder
)

// Method constants re-exports.
const (
	GET     = MethodGet
	POST    = MethodPost
	PUT     = MethodPut
	PATCH   = MethodPatch
	DELETE  = MethodDelete
	HEAD    = MethodHead
	OPTIONS = MethodOptions
)

// Note: Config helpers, server constructors, and options are available directly
// as functions in the aurelion package (DefaultConfig, NewHttpServer, etc.)

// Route builders.
// These wrapper functions accept aurelion types (Handler, Middleware, Method, CORSConfig)
// and convert them internally to router types to avoid import cycles.
var (
	NewRoute      = NewRouteWrapper
	NewGroupRoute = NewGroupRouteWrapper
)

// Response helpers.
var (
	OK                  = OKPublic
	Error               = ErrorPublic
	BadRequest          = BadRequestPublic
	Unauthorized        = UnauthorizedPublic
	Forbidden           = ForbiddenPublic
	NotFound            = NotFoundPublic
	InternalServerError = InternalServerErrorPublic
	HealthCheck         = HealthCheckPublic
	ErrorWithDetails    = ErrorWithDetailsPublic
	NewError            = response.NewError
	NewErrorf           = response.NewErrorf
)

// Middleware helpers.
var (
	RequestIDMiddleware             = middleware.RequestIDMiddleware
	TraceIDMiddleware               = middleware.TraceIDMiddleware
	RequestResponseLogger           = middleware.RequestResponseLogger
	HeaderToLocalsMiddleware        = HeaderToLocalsPublic
	DefaultHeaderToLocalsMiddleware = DefaultHeaderToLocalsPublic
	GetHeader                       = GetHeaderPublic
	GetHeaderInt                    = GetHeaderIntPublic
	GetHeaderBool                   = GetHeaderBoolPublic
	GetAllHeaders                   = GetAllHeadersPublic
)

// Validation helpers.
var (
	Validate         = validation.Validate
	ValidateAndParse = ValidateAndParsePublic
)

// Context adapters.
var (
	MiddlewareToFiber = MiddlewareToFiberPublic
	HandlerToFiber    = HandlerToFiberPublic
	NewContext        = NewFiberContextPublic
)

// Error constants re-exports.
var (
	// Errors from main package are available directly.
	// Re-export response error types and context keys.
	MsgValidationFailed          = response.MsgValidationFailed
	ErrorTypeValidation          = response.ErrorTypeValidation
	ErrorTypeBusiness            = response.ErrorTypeBusiness
	ErrorTypePermission          = response.ErrorTypePermission
	ErrorTypeRateLimit           = response.ErrorTypeRateLimit
	ErrorTypeExternal            = response.ErrorTypeExternal
	ErrorTypeInternalServerError = response.ErrorTypeInternalServerError
	ContextKeyLanguage           = context.KeyLanguage
	ContextKeyUserID             = context.KeyUserID
	ContextKeyRequestID          = context.KeyRequestID
	ContextKeyTraceID            = context.KeyTraceID
)

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx Context) string {
	if ctx == nil {
		return ""
	}
	if id := ctx.Locals(context.ContextKeyRequestID); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// GetTraceID retrieves the trace ID from the context.
func GetTraceID(ctx Context) string {
	if ctx == nil {
		return ""
	}
	if id := ctx.Locals(context.ContextKeyTraceID); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// Context helpers (std context).
func ContextSet(ctx stdctx.Context, key string, value interface{}) stdctx.Context {
	return context.Set(ctx, key, value)
}

func ContextGet(ctx stdctx.Context, key string) interface{} {
	return context.Get(ctx, key)
}

func ContextHas(ctx stdctx.Context, key string) bool {
	return context.Has(ctx, key)
}

func ContextGetString(ctx stdctx.Context, key, defaultValue string) string {
	return context.GetString(ctx, key, defaultValue)
}

func ContextGetInt(ctx stdctx.Context, key string, defaultValue int) int {
	return context.GetInt(ctx, key, defaultValue)
}

func ContextGetBool(ctx stdctx.Context, key string, defaultValue bool) bool {
	return context.GetBool(ctx, key, defaultValue)
}

func ContextValues(ctx stdctx.Context) map[string]string {
	return context.GetAllContextValues(ctx)
}

func GetLanguageFromContext(ctx stdctx.Context) string {
	return context.GetLanguageFromContext(ctx)
}

func GetUserIDFromContext(ctx stdctx.Context) interface{} {
	return context.GetUserIDFromContext(ctx)
}

func GetRequestIDFromContext(ctx stdctx.Context) string {
	return context.GetRequestIDFromContext(ctx)
}

func GetTraceIDFromContext(ctx stdctx.Context) string {
	return context.GetTraceIDFromContext(ctx)
}
