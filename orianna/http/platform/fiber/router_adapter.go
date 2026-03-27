// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/engine"
	"github.com/gofiber/fiber/v3"
)

// routerAdapter implements Router using Fiber
type routerAdapter struct {
	router fiber.Router
	config *configuration.Config
}

// newRouterAdapterWithConfig creates a new router adapter with config
func newRouterAdapterWithConfig(router fiber.Router, conf *configuration.Config) engine.RouterEngine {
	return &routerAdapter{router: router, config: conf}
}

// GET registers a GET route
func (r *routerAdapter) GET(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Get(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// POST registers a POST route
func (r *routerAdapter) POST(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Post(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// PUT registers a PUT route
func (r *routerAdapter) PUT(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Put(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// PATCH registers a PATCH route
func (r *routerAdapter) PATCH(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Patch(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// DELETE registers a DELETE route
func (r *routerAdapter) DELETE(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Delete(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// HEAD registers a HEAD route
func (r *routerAdapter) HEAD(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Head(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// OPTIONS registers an OPTIONS route
func (r *routerAdapter) OPTIONS(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlersToAny(handlers)
	r.router.Options(path, fiberHandlers[0], fiberHandlers[1:]...)
}

// Use adds middleware to the router
func (r *routerAdapter) Use(middleware ...core.Middleware) {
	for _, mw := range middleware {
		r.router.Use(r.convertMiddleware(mw))
	}
}

// Group creates a route group with a prefix
func (r *routerAdapter) Group(prefix string, middleware ...core.Middleware) engine.RouterEngine {
	fiberMiddlewares := r.convertMiddlewaresToAny(middleware)
	group := r.router.Group(prefix, fiberMiddlewares...)
	return newRouterAdapterWithConfig(group, r.config)
}

// convertHandlersToAny converts domain handlers to []any for fiber v3 router methods
func (r *routerAdapter) convertHandlersToAny(handlers []core.Handler) []any {
	result := make([]any, len(handlers))
	for i, handler := range handlers {
		result[i] = r.convertHandler(handler)
	}
	return result
}

// convertMiddlewaresToAny converts domain middlewares to []any for fiber v3 router methods
func (r *routerAdapter) convertMiddlewaresToAny(middlewares []core.Middleware) []any {
	result := make([]any, len(middlewares))
	for i, middleware := range middlewares {
		result[i] = r.convertMiddleware(middleware)
	}
	return result
}

// convertHandler converts a domain handler to fiber handler.
// Reuses an existing ContextAdapter from the request if available.
func (r *routerAdapter) convertHandler(handler core.Handler) fiber.Handler {
	if handler == nil {
		return func(_ fiber.Ctx) error {
			return nil
		}
	}
	conf := r.config
	return func(c fiber.Ctx) error {
		return withContextAdapter(c, conf, func(ctx *ContextAdapter) error {
			return handler(ctx)
		})
	}
}

// convertMiddleware converts a domain middleware to fiber middleware.
// Reuses an existing ContextAdapter from the request if available.
func (r *routerAdapter) convertMiddleware(middleware core.Middleware) fiber.Handler {
	if middleware == nil {
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}
	conf := r.config
	return func(c fiber.Ctx) error {
		return withContextAdapter(c, conf, func(ctx *ContextAdapter) error {
			return middleware(ctx)
		})
	}
}
