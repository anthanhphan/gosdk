// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package orianna

import (
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/middleware"
	"github.com/anthanhphan/gosdk/orianna/pkg/routing"
	"github.com/anthanhphan/gosdk/orianna/pkg/server"
	ovalidator "github.com/anthanhphan/gosdk/orianna/pkg/validator"
)

// Core types
type (
	Context          = core.Context
	Map              = core.Map
	Handler          = core.Handler
	Middleware       = core.Middleware
	Method           = core.Method
	BindSource       = core.BindSource
	BindOptions      = core.BindOptions
	BaseResponse     = core.BaseResponse
	SuccessResponse  = core.SuccessResponse
	ErrorResponse    = core.ErrorResponse
	RequestHooks     = core.Hooks
	ValidationErrors = ovalidator.ValidationErrors
)

// Server types
type (
	Server       = server.Server
	ServerOption = server.ServerOption
	Config       = configuration.Config
)

// Routing types
type (
	Route             = routing.Route
	RouteGroup        = routing.RouteGroup
	RouteBuilder      = routing.RouteBuilder
	GroupRouteBuilder = routing.GroupRouteBuilder
)

// Health types
type (
	HealthStatus  = server.HealthStatus
	HealthCheck   = server.HealthCheck
	HealthReport  = server.HealthReport
	HealthChecker = server.HealthChecker
	HealthManager = server.HealthManager
	CustomChecker = server.CustomChecker
	HTTPChecker   = server.HTTPChecker
)

// Bind source constants
const (
	BindSourceBody    = core.BindSourceBody
	BindSourceQuery   = core.BindSourceQuery
	BindSourceParams  = core.BindSourceParams
	BindSourceHeaders = core.BindSourceHeaders
)

// HTTP status constants
const (
	StatusOK                  = core.StatusOK
	StatusCreated             = core.StatusCreated
	StatusAccepted            = core.StatusAccepted
	StatusNoContent           = core.StatusNoContent
	StatusBadRequest          = core.StatusBadRequest
	StatusUnauthorized        = core.StatusUnauthorized
	StatusForbidden           = core.StatusForbidden
	StatusNotFound            = core.StatusNotFound
	StatusConflict            = core.StatusConflict
	StatusUnprocessableEntity = core.StatusUnprocessableEntity
	StatusTooManyRequests     = core.StatusTooManyRequests
	StatusInternalServerError = core.StatusInternalServerError
	StatusServiceUnavailable  = core.StatusServiceUnavailable
	StatusGatewayTimeout      = core.StatusGatewayTimeout
)

// Health status constants
const (
	HealthStatusHealthy   = server.HealthStatusHealthy
	HealthStatusUnhealthy = server.HealthStatusUnhealthy
	HealthStatusDegraded  = server.HealthStatusDegraded
)

// HTTP method constants
var (
	GET     = core.GET
	POST    = core.POST
	PUT     = core.PUT
	DELETE  = core.DELETE
	PATCH   = core.PATCH
	HEAD    = core.HEAD
	OPTIONS = core.OPTIONS
)

// Response helpers
var (
	NewErrorResponse   = core.NewErrorResponse
	NewSuccessResponse = core.NewSuccessResponse
	SendSuccess        = core.SendSuccess
	SendError          = core.SendError
	HandleError        = core.HandleError
	NewRequestHooks    = core.NewHooks
	ValidateAndRespond = core.ValidateAndRespond
	IsErrorCode        = core.IsErrorCode
)

// Error helpers
var (
	WrapError     = core.WrapError
	WrapErrorf    = core.WrapErrorf
	IsConfigError = core.IsConfigError
	IsRouteError  = core.IsRouteError
	IsServerError = core.IsServerError
)

// Query/param helpers
var (
	GetParamInt    = core.GetParamInt
	GetParamInt64  = core.GetParamInt64
	GetParamUUID   = core.GetParamUUID
	GetQueryInt    = core.GetQueryInt
	GetQueryInt64  = core.GetQueryInt64
	GetQueryBool   = core.GetQueryBool
	GetQueryString = core.GetQueryString
)

// Middleware helpers
var (
	Chain               = middleware.Chain
	Optional            = middleware.Optional
	OnlyForMethods      = middleware.OnlyForMethods
	SkipForPaths        = middleware.SkipForPaths
	SkipForPathPrefixes = middleware.SkipForPathPrefixes
	Before              = middleware.Before
	After               = middleware.After
	Recover             = middleware.Recover
	Timeout             = middleware.Timeout
)

// Server options
var (
	WithGlobalMiddleware = server.WithGlobalMiddleware
	WithPanicRecover     = server.WithPanicRecover
	WithAuthentication   = server.WithAuthentication
	WithAuthorization    = server.WithAuthorization
	WithRateLimiter      = server.WithRateLimiter
	WithHooks            = server.WithHooks
	WithMiddlewareConfig = server.WithMiddlewareConfig
	WithHealthManager    = server.WithHealthManager
	WithShutdownManager  = server.WithShutdownManager
	WithHealthChecker    = server.WithHealthChecker
	WithMetrics          = server.WithMetrics
)

// Constructors
var (
	NewServer        = server.NewServer
	NewRoute         = routing.NewRoute
	NewGroupRoute    = routing.NewGroupRoute
	NewHealthManager = server.NewHealthManager
	NewCustomChecker = server.NewCustomChecker
	NewHTTPChecker   = server.NewHTTPChecker
	SimpleHandler    = core.SimpleHandler
)

// Generic functions (require wrappers for type parameters)

// Bind parses request data into a typed struct.
func Bind[T any](ctx Context, opts ...BindOptions) (T, error) {
	return core.Bind[T](ctx, opts...)
}

// MustBind parses request and sends error response on failure.
func MustBind[T any](ctx Context, opts ...BindOptions) (T, bool) {
	return core.MustBind[T](ctx, opts...)
}

// BindBody is a shorthand for binding from request body.
func BindBody[T any](ctx Context, validate bool) (T, error) {
	return core.BindBody[T](ctx, validate)
}

// BindQuery is a shorthand for binding from query parameters.
func BindQuery[T any](ctx Context, validate bool) (T, error) {
	return core.BindQuery[T](ctx, validate)
}

// BindParams is a shorthand for binding from route parameters.
func BindParams[T any](ctx Context, validate bool) (T, error) {
	return core.BindParams[T](ctx, validate)
}

// TypedHandler creates a handler with automatic request/response marshaling.
func TypedHandler[Req, Resp any](fn func(Context, Req) (Resp, error)) Handler {
	return core.TypedHandler(fn)
}
