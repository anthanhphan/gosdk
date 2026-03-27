// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"fmt"

	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/engine"
	"github.com/anthanhphan/gosdk/orianna/http/routing"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
)

// RegisterRoutes registers multiple routes
func (s *ServerAdapter) RegisterRoutes(routes ...routing.Route) error {
	for _, route := range routes {
		if err := s.registerRoute(route); err != nil {
			return err
		}
	}
	return nil
}

// RegisterGroup registers a route group
func (s *ServerAdapter) RegisterGroup(group routing.RouteGroup) error {
	return s.registerGroupToRouter(s.router, group)
}

// RegisterMetricsHandler registers a /metrics endpoint for Prometheus scraping
func (s *ServerAdapter) RegisterMetricsHandler(client engine.MetricsClient) {
	if client == nil {
		return
	}

	// Get the http.Handler from the metrics client
	handler := client.Handler()

	// Register /metrics route adapting the net/http handler to fiber
	// We use a custom adapter since fiber uses fasthttp internally
	s.app.Get("/metrics", adaptor.HTTPHandler(handler))
}

// RegisterStaticFiles serves static files from the filesystem.
// Does nothing if config is nil.
func (s *ServerAdapter) RegisterStaticFiles(conf *configuration.StaticFileConfig) {
	if conf == nil || conf.Root == "" {
		return
	}
	prefix := conf.Prefix
	if prefix == "" {
		prefix = "/static"
	}
	s.app.Use(prefix, static.New(conf.Root, static.Config{
		Browse: conf.Browse,
		MaxAge: conf.MaxAge,
	}))
}

// registerGroupToRouter registers a group to a specific router
func (s *ServerAdapter) registerGroupToRouter(router engine.RouterEngine, group routing.RouteGroup) error {
	// Create group router
	groupRouter := router.Group(group.Prefix, group.Middlewares...)

	// Register all routes in the group
	for _, route := range group.Routes {
		if err := s.registerRouteToRouter(groupRouter, route); err != nil {
			return err
		}
	}

	// Register nested groups
	for _, subGroup := range group.Groups {
		if err := s.registerGroupToRouter(groupRouter, subGroup); err != nil {
			return err
		}
	}

	return nil
}

// registerRoute registers a single route
func (s *ServerAdapter) registerRoute(route routing.Route) error {
	return s.registerRouteToRouter(s.router, route)
}

// registerRouteToRouter registers a route to a specific router
func (s *ServerAdapter) registerRouteToRouter(router engine.RouterEngine, route routing.Route) error {
	path := route.Path
	if path == "" {
		path = "/"
	}

	// Build handler chain: route.Middlewares already includes protection middleware
	// (auth/authz) applied by the RouteRegistry. buildHandlerChain chains these
	// route-level middlewares with the final handler into an ordered handler slice.
	handlers := s.buildHandlerChain(route.Middlewares, route.Handler)

	// Helper to register for a single method
	register := func(method core.Method) error {
		switch method {
		case core.GET:
			router.GET(path, handlers...)
		case core.POST:
			router.POST(path, handlers...)
		case core.PUT:
			router.PUT(path, handlers...)
		case core.PATCH:
			router.PATCH(path, handlers...)
		case core.DELETE:
			router.DELETE(path, handlers...)
		case core.HEAD:
			router.HEAD(path, handlers...)
		case core.OPTIONS:
			router.OPTIONS(path, handlers...)
		default:
			return fmt.Errorf("unsupported HTTP method: %v", method)
		}
		return nil
	}

	for _, method := range route.Methods {
		if err := register(method); err != nil {
			return err
		}
	}
	return nil
}

// buildHandlerChain builds a handler chain from middlewares and handler
func (s *ServerAdapter) buildHandlerChain(middlewares []core.Middleware, handler core.Handler) []core.Handler {
	chain := make([]core.Handler, 0, len(middlewares)+1)
	// Convert middlewares to handlers (they're compatible)
	for _, mw := range middlewares {
		chain = append(chain, core.Handler(mw))
	}

	// Use handler directly
	chain = append(chain, handler)
	return chain
}

// convertToFiberMiddlewareWithConfig converts domain middleware to fiber handler
// with config provided at setup time (avoids per-request Locals lookup).
func convertToFiberMiddlewareWithConfig(middleware core.Middleware, conf *configuration.Config) fiber.Handler {
	if middleware == nil {
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}
	return func(c fiber.Ctx) error {
		return withContextAdapter(c, conf, func(ctx *ContextAdapter) error {
			return middleware(ctx)
		})
	}
}

// withContextAdapter handles ContextAdapter acquire/reuse/release for Fiber handlers.
// If a ContextAdapter already exists in the request (stored in Locals), it is reused.
// Otherwise, a new one is acquired from the pool and released after fn completes.
func withContextAdapter(c fiber.Ctx, conf *configuration.Config, fn func(*ContextAdapter) error) error {
	if existing, ok := c.Locals(ctxAdapterKey).(*ContextAdapter); ok && existing != nil {
		return fn(existing)
	}
	ctx := AcquireContextAdapter(c, conf)
	c.Locals(ctxAdapterKey, ctx)
	defer func() {
		c.Locals(ctxAdapterKey, nil)
		ReleaseContextAdapter(ctx)
	}()
	return fn(ctx)
}
