// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package engine

//go:generate mockgen -source=interfaces.go -destination=mocks/mock_engine.go -package=mocks

import (
	"context"
	"net/http"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/routing"
)

// ServerEngine defines the port interface for HTTP server implementations
// It allows for framework-agnostic application logic and easy swapping of HTTP framework implementations.
type ServerEngine interface {
	Start() error
	Shutdown(ctx context.Context) error
	RegisterRoutes(routes ...routing.Route) error
	RegisterGroup(group routing.RouteGroup) error
	SetupGlobalMiddlewares(config *configuration.MiddlewareConfig, global []core.Middleware, panicRecover core.Middleware, rateLimiter core.Middleware, log *logger.Logger)
	Use(middleware ...core.Middleware)
	// RegisterMetricsHandler registers a /metrics endpoint using the provided metrics client
	RegisterMetricsHandler(client MetricsClient)
}

// MetricsClient is a minimal interface for metrics clients that provide an HTTP handler
type MetricsClient interface {
	Handler() http.Handler
}

// RouterEngine defines the port interface for HTTP router implementations
type RouterEngine interface {
	GET(path string, handlers ...core.Handler)
	POST(path string, handlers ...core.Handler)
	PUT(path string, handlers ...core.Handler)
	PATCH(path string, handlers ...core.Handler)
	DELETE(path string, handlers ...core.Handler)
	HEAD(path string, handlers ...core.Handler)
	OPTIONS(path string, handlers ...core.Handler)
	Use(middleware ...core.Middleware)
	Group(prefix string, middleware ...core.Middleware) RouterEngine
}
