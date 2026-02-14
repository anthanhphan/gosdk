// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/engine"
	"github.com/anthanhphan/gosdk/orianna/pkg/routing"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// ctxAdapterKey is the fiber Locals key for reusing ContextAdapter within a request.
const ctxAdapterKey = "__orianna_ctx"

// ServerAdapter implements ServerEngine using Fiber
type ServerAdapter struct {
	app    *fiber.App
	config *configuration.Config
	router engine.RouterEngine
}

// NewServerAdapter creates a new server adapter
func NewServerAdapter(conf *configuration.Config) (*ServerAdapter, error) {
	// Create fiber config with jcodec for optimal JSON performance
	fiberConfig := fiber.Config{
		AppName:               conf.ServiceName,
		BodyLimit:             conf.MaxBodySize,
		Concurrency:           conf.MaxConcurrentConnections,
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		JSONEncoder:           jcodec.Marshal,
		JSONDecoder:           jcodec.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Acquire context adapter
			ctx := AcquireContextAdapter(c, conf)
			defer ReleaseContextAdapter(ctx)

			// Try to handle ErrorResponse
			if core.HandleError(ctx, err) {
				return nil
			}

			// Handle Fiber errors (e.g., 404 Not Found, 405 Method Not Allowed)
			code := fiber.StatusInternalServerError
			msg := "Internal Server Error"
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
				msg = fiberErr.Message
			}

			// For standard errors, create ErrorResponse
			errResp := core.NewErrorResponse("ERROR", code, msg)
			errResp.RequestID = ctx.RequestID()

			return ctx.Status(code).JSON(errResp)
		},
	}

	// Set timeouts if configured
	if conf.ReadTimeout != nil {
		fiberConfig.ReadTimeout = *conf.ReadTimeout
	}
	if conf.WriteTimeout != nil {
		fiberConfig.WriteTimeout = *conf.WriteTimeout
	}
	if conf.IdleTimeout != nil {
		fiberConfig.IdleTimeout = *conf.IdleTimeout
	}

	// Create fiber app
	app := fiber.New(fiberConfig)

	// Create router adapter
	router := newRouterAdapterWithConfig(app, conf)

	return &ServerAdapter{
		app:    app,
		config: conf,
		router: router,
	}, nil
}

// Start starts the HTTP server and blocks until shutdown
func (s *ServerAdapter) Start() error {
	// Create channel for errors
	errChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if err := s.app.Listen(fmt.Sprintf(":%d", s.config.Port)); err != nil {
			errChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-quit:
		// Create shutdown context with timeout
		timeout := configuration.DefaultGracefulShutdownTimeout
		if s.config.GracefulShutdownTimeout != nil {
			timeout = *s.config.GracefulShutdownTimeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return s.Shutdown(ctx)
	}
}

// Shutdown gracefully shuts down the server
func (s *ServerAdapter) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

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

	// Build handler chain
	handlers := s.buildHandlerChain(route.Middlewares, route.Handler, &route)

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

	// Register for multiple methods if specified
	if len(route.Methods) > 0 {
		for _, method := range route.Methods {
			if err := register(method); err != nil {
				return err
			}
		}
		return nil
	}

	// Fallback to single method
	return register(route.Method)
}

// Use adds global middleware to the server
func (s *ServerAdapter) Use(middleware ...core.Middleware) {
	for _, mw := range middleware {
		s.app.Use(convertToFiberMiddlewareWithConfig(mw, s.config))
	}
}

// buildHandlerChain builds a handler chain from middlewares and handler
func (s *ServerAdapter) buildHandlerChain(middlewares []core.Middleware, handler core.Handler, _ *routing.Route) []core.Handler {
	chain := make([]core.Handler, 0, len(middlewares)+1)
	// Convert middlewares to handlers (they're compatible)
	for _, mw := range middlewares {
		chain = append(chain, core.Handler(mw))
	}

	// Use handler directly
	chain = append(chain, handler)
	return chain
}

// SetupGlobalMiddlewares sets up global middlewares on the server
func (s *ServerAdapter) SetupGlobalMiddlewares(
	middlewareConfig *configuration.MiddlewareConfig,
	globalMiddlewares []core.Middleware,
	panicRecover core.Middleware,
	rateLimiter core.Middleware,
	log *logger.Logger,
) {
	s.setupSecurityMiddlewares(middlewareConfig)
	s.setupTrafficMiddlewares(middlewareConfig, rateLimiter)
	s.setupObservabilityMiddlewares(middlewareConfig, panicRecover, log)
	s.setupCoreMiddlewares(globalMiddlewares)
}

func (s *ServerAdapter) setupSecurityMiddlewares(middlewareConfig *configuration.MiddlewareConfig) {
	// Add Helmet middleware (security headers)
	if middlewareConfig == nil || !middlewareConfig.DisableHelmet {
		s.app.Use(helmet.New())
	}

	// Add CORS middleware if enabled
	if s.config.EnableCORS && s.config.CORS != nil {
		s.app.Use(cors.New(buildCORSConfig(s.config.CORS)))
	}

	// Add CSRF protection middleware if enabled
	if s.config.EnableCSRF && s.config.CSRF != nil {
		s.app.Use(csrf.New(buildCSRFConfig(s.config.CSRF)))
	}
}

func (s *ServerAdapter) setupTrafficMiddlewares(middlewareConfig *configuration.MiddlewareConfig, rateLimiter core.Middleware) {
	// Add rate limiter middleware
	if middlewareConfig == nil || !middlewareConfig.DisableRateLimit {
		if rateLimiter != nil {
			s.app.Use(convertToFiberMiddlewareWithConfig(rateLimiter, s.config))
		} else {
			// Use default rate limiter configuration
			s.app.Use(limiter.New(limiter.Config{
				Max:        configuration.DefaultRateLimitMax,
				Expiration: configuration.DefaultRateLimitExpiration,
			}))
		}
	}

	// Add compression middleware
	if middlewareConfig == nil || !middlewareConfig.DisableCompression {
		s.app.Use(compress.New(compress.Config{
			Level: configuration.DefaultCompressionLevel,
		}))
	}
}

func (s *ServerAdapter) setupObservabilityMiddlewares(
	middlewareConfig *configuration.MiddlewareConfig,
	panicRecover core.Middleware,
	log *logger.Logger,
) {
	// Add panic recovery middleware
	if middlewareConfig == nil || !middlewareConfig.DisableRecovery {
		if panicRecover != nil {
			s.app.Use(convertToFiberMiddlewareWithConfig(panicRecover, s.config))
		} else {
			s.app.Use(recover.New())
		}
	}

	// Add request ID middleware
	if middlewareConfig == nil || !middlewareConfig.DisableRequestID {
		s.app.Use(requestIDMiddleware())
	}

	// Add trace ID middleware
	if middlewareConfig == nil || !middlewareConfig.DisableTraceID {
		s.app.Use(traceIDMiddleware())
	}

	// Add request/response logging middleware
	if middlewareConfig == nil || !middlewareConfig.DisableLogging {
		if log != nil {
			s.app.Use(requestResponseLoggingMiddleware(log, s.config.VerboseLogging, s.config.VerboseLoggingSkipPaths))
		}
	}
}

func (s *ServerAdapter) setupCoreMiddlewares(globalMiddlewares []core.Middleware) {
	// Add config middleware to store config in context
	s.app.Use(configMiddleware(s.config))

	// Add custom global middlewares
	for _, middleware := range globalMiddlewares {
		s.app.Use(convertToFiberMiddlewareWithConfig(middleware, s.config))
	}
}

// Router Adapter

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
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Get(path, fiberHandlers...)
}

// POST registers a POST route
func (r *routerAdapter) POST(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Post(path, fiberHandlers...)
}

// PUT registers a PUT route
func (r *routerAdapter) PUT(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Put(path, fiberHandlers...)
}

// PATCH registers a PATCH route
func (r *routerAdapter) PATCH(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Patch(path, fiberHandlers...)
}

// DELETE registers a DELETE route
func (r *routerAdapter) DELETE(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Delete(path, fiberHandlers...)
}

// HEAD registers a HEAD route
func (r *routerAdapter) HEAD(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Head(path, fiberHandlers...)
}

// OPTIONS registers an OPTIONS route
func (r *routerAdapter) OPTIONS(path string, handlers ...core.Handler) {
	fiberHandlers := r.convertHandlers(handlers)
	r.router.Options(path, fiberHandlers...)
}

// Use adds middleware to the router
func (r *routerAdapter) Use(middleware ...core.Middleware) {
	for _, mw := range middleware {
		r.router.Use(r.convertMiddleware(mw))
	}
}

// Group creates a route group with a prefix
func (r *routerAdapter) Group(prefix string, middleware ...core.Middleware) engine.RouterEngine {
	fiberMiddlewares := r.convertMiddlewares(middleware)
	group := r.router.Group(prefix, fiberMiddlewares...)
	return newRouterAdapterWithConfig(group, r.config)
}

// convertHandlers converts domain handlers to fiber handlers
func (r *routerAdapter) convertHandlers(handlers []core.Handler) []fiber.Handler {
	fiberHandlers := make([]fiber.Handler, len(handlers))
	for i, handler := range handlers {
		fiberHandlers[i] = r.convertHandler(handler)
	}
	return fiberHandlers
}

// convertHandler converts a domain handler to fiber handler.
// Reuses an existing ContextAdapter from the request if available.
func (r *routerAdapter) convertHandler(handler core.Handler) fiber.Handler {
	if handler == nil {
		return func(_ *fiber.Ctx) error {
			return nil
		}
	}
	conf := r.config
	return func(c *fiber.Ctx) error {
		// Reuse existing adapter if already created in this request
		if existing, ok := c.Locals(ctxAdapterKey).(*ContextAdapter); ok && existing != nil {
			return handler(existing)
		}
		// First handler in chain: acquire and store for reuse
		ctx := AcquireContextAdapter(c, conf)
		c.Locals(ctxAdapterKey, ctx)
		defer func() {
			c.Locals(ctxAdapterKey, nil)
			ReleaseContextAdapter(ctx)
		}()
		return handler(ctx)
	}
}

// convertMiddlewares converts domain middlewares to fiber middlewares
func (r *routerAdapter) convertMiddlewares(middlewares []core.Middleware) []fiber.Handler {
	fiberMiddlewares := make([]fiber.Handler, len(middlewares))
	for i, middleware := range middlewares {
		fiberMiddlewares[i] = r.convertMiddleware(middleware)
	}
	return fiberMiddlewares
}

// convertMiddleware converts a domain middleware to fiber middleware.
// Reuses an existing ContextAdapter from the request if available.
func (r *routerAdapter) convertMiddleware(middleware core.Middleware) fiber.Handler {
	if middleware == nil {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	conf := r.config
	return func(c *fiber.Ctx) error {
		// Reuse existing adapter if already created in this request
		if existing, ok := c.Locals(ctxAdapterKey).(*ContextAdapter); ok && existing != nil {
			return middleware(existing)
		}
		ctx := AcquireContextAdapter(c, conf)
		c.Locals(ctxAdapterKey, ctx)
		defer func() {
			c.Locals(ctxAdapterKey, nil)
			ReleaseContextAdapter(ctx)
		}()
		return middleware(ctx)
	}
}

// convertToFiberMiddlewareWithConfig converts domain middleware to fiber handler
// with config provided at setup time (avoids per-request Locals lookup).
func convertToFiberMiddlewareWithConfig(middleware core.Middleware, conf *configuration.Config) fiber.Handler {
	if middleware == nil {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	return func(c *fiber.Ctx) error {
		// Reuse existing adapter if already created in this request
		if existing, ok := c.Locals(ctxAdapterKey).(*ContextAdapter); ok && existing != nil {
			return middleware(existing)
		}
		ctx := AcquireContextAdapter(c, conf)
		c.Locals(ctxAdapterKey, ctx)
		defer func() {
			c.Locals(ctxAdapterKey, nil)
			ReleaseContextAdapter(ctx)
		}()
		return middleware(ctx)
	}
}
