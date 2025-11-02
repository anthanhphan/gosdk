package aurelion

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

// HttpServer represents an HTTP server instance
type HttpServer struct {
	app               *fiber.App
	config            *Config
	globalMiddlewares []Middleware
	panicRecover      Middleware
	authMiddleware    Middleware
	authzChecker      AuthorizationFunc
	rateLimiter       Middleware
	routes            []*Route
	groupRoutes       []*GroupRoute
	logger            *zap.SugaredLogger
}

// NewHttpServer creates a new HTTP server with the given configuration and options
//
// Input:
//   - config: The server configuration
//   - options: Optional server configuration functions
//
// Output:
//   - *HttpServer: The created HTTP server
//   - error: Any error that occurred during creation
//
// Example:
//
//	config := &aurelion.Config{
//	    ServiceName: "My API",
//	    Port: 8080,
//	}
//	server, err := aurelion.NewHttpServer(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewHttpServer(config *Config, options ...ServerOption) (*HttpServer, error) {
	// Check for nil config
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Merge with default config
	config = config.Merge()

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create fiber config
	fiberConfig := fiber.Config{
		AppName:               config.ServiceName,
		BodyLimit:             config.MaxBodySize,
		Concurrency:           config.MaxConcurrentConnections,
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
	}

	// Set timeouts if configured
	if config.ReadTimeout != nil {
		fiberConfig.ReadTimeout = *config.ReadTimeout
	}
	if config.WriteTimeout != nil {
		fiberConfig.WriteTimeout = *config.WriteTimeout
	}
	if config.IdleTimeout != nil {
		fiberConfig.IdleTimeout = *config.IdleTimeout
	}

	// Create fiber app
	app := fiber.New(fiberConfig)

	// Create server
	server := &HttpServer{
		app:               app,
		config:            config,
		globalMiddlewares: make([]Middleware, 0),
		routes:            make([]*Route, 0),
		groupRoutes:       make([]*GroupRoute, 0),
		logger:            logger.NewLoggerWithFields(zap.String("package", "aurelion")),
	}

	// Apply options
	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Setup global middlewares
	server.setupGlobalMiddlewares()

	// Add default health check route
	server.AddRoutes(
		NewRoute(DefaultHealthCheckPath).
			GET().
			Handler(func(ctx Context) error {
				return HealthCheck(ctx)
			}),
	)

	return server, nil
}

// setupGlobalMiddlewares sets up global middlewares
func (s *HttpServer) setupGlobalMiddlewares() {
	// Add Helmet middleware (security headers)
	s.app.Use(helmet.New())

	// Add rate limiter middleware (always enabled)
	// Default: 500 requests per minute per IP address
	// Use WithRateLimiter option to customize (e.g., per user, per API key, global)
	if s.rateLimiter != nil {
		s.app.Use(middlewareToFiber(s.rateLimiter))
	} else {
		// Use default rate limiter: 500 requests per minute per IP
		s.app.Use(limiter.New(limiter.Config{
			Max:        DefaultRateLimitMax,
			Expiration: DefaultRateLimitExpiration,
		}))
	}

	// Add compression middleware (always enabled, default best speed)
	// Fiber automatically detects and uses Brotli if available and client supports it
	s.app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Add panic recovery middleware (built-in or custom)
	if s.panicRecover != nil {
		s.app.Use(middlewareToFiber(s.panicRecover))
	} else {
		// Use fiber's built-in recover middleware as default
		s.app.Use(recover.New())
	}

	// Add request ID middleware (always enabled)
	s.app.Use(requestIDMiddleware())

	// Add request/response logging middleware (configurable)
	s.app.Use(requestResponseLoggingMiddleware(s.logger, s.config.VerboseLogging))

	// Add custom global middlewares
	for _, middleware := range s.globalMiddlewares {
		s.app.Use(middlewareToFiber(middleware))
	}

	// Add CORS middleware if enabled
	if s.config.EnableCORS {
		corsConfig := cors.Config{
			AllowOrigins:     strings.Join(s.config.CORS.AllowOrigins, ","),
			AllowMethods:     strings.Join(s.config.CORS.AllowMethods, ","),
			AllowHeaders:     strings.Join(s.config.CORS.AllowHeaders, ","),
			AllowCredentials: s.config.CORS.AllowCredentials,
			ExposeHeaders:    strings.Join(s.config.CORS.ExposeHeaders, ","),
			MaxAge:           s.config.CORS.MaxAge,
		}
		s.app.Use(cors.New(corsConfig))
	}
}

// AddRoutes adds multiple routes to the server
//
// Input:
//   - routes: The routes to add (RouteBuilder or Route)
//
// Output:
//   - *HttpServer: The server instance (for chaining)
//
// Example:
//
//	server.AddRoutes(
//	    aurelion.NewRoute("/users").GET().Handler(getUsersHandler),
//	    aurelion.NewRoute("/users/:id").GET().Handler(getUserHandler),
//	)
func (s *HttpServer) AddRoutes(routes ...interface{}) *HttpServer {
	for _, r := range routes {
		var route *Route

		switch v := r.(type) {
		case *RouteBuilder:
			route = v.Build()
		case *Route:
			route = v
		case Route:
			route = &v
		default:
			s.logger.Warnw("invalid route type, skipping", "type", fmt.Sprintf("%T", r))
			continue
		}

		// Validate route
		if err := validateRoute(route); err != nil {
			s.logger.Errorw("invalid route", "error", err, "path", route.Path)
			continue
		}

		// Apply protection middleware if needed
		s.applyProtectionMiddleware(route)

		// Register route with optional CORS
		s.registerRouteWithOptionalCORS(route)

		// Store route for reference
		s.routes = append(s.routes, route)
	}

	return s
}

// registerRouteWithOptionalCORS registers a route with optional per-route CORS configuration
func (s *HttpServer) registerRouteWithOptionalCORS(route *Route) {
	if route == nil {
		return
	}

	// If route has CORS config, create a group with CORS middleware
	if route.CORS != nil {
		corsConfig := cors.Config{
			AllowOrigins:     strings.Join(route.CORS.AllowOrigins, ","),
			AllowMethods:     strings.Join(route.CORS.AllowMethods, ","),
			AllowHeaders:     strings.Join(route.CORS.AllowHeaders, ","),
			AllowCredentials: route.CORS.AllowCredentials,
			ExposeHeaders:    strings.Join(route.CORS.ExposeHeaders, ","),
			MaxAge:           route.CORS.MaxAge,
		}
		g := s.app.Group("", cors.New(corsConfig))
		handlers := route.buildHandlers()
		registerRoute(g, route.Method, route.Path, handlers)
		return
	}

	// No CORS config, register directly
	route.register(s.app)
}

// AddGroupRoutes adds multiple group routes to the server
//
// Input:
//   - groups: The group routes to add (GroupRouteBuilder or GroupRoute)
//
// Output:
//   - *HttpServer: The server instance (for chaining)
//
// Example:
//
//	server.AddGroupRoutes(
//	    aurelion.NewGroupRoute("/api/v1").Routes(
//	        aurelion.NewRoute("/status").GET().Handler(statusHandler),
//	        aurelion.NewRoute("/health").GET().Handler(healthHandler),
//	    ),
//	)
func (s *HttpServer) AddGroupRoutes(groups ...interface{}) *HttpServer {
	for _, g := range groups {
		var group *GroupRoute

		switch v := g.(type) {
		case *GroupRouteBuilder:
			group = v.Build()
		case *GroupRoute:
			group = v
		case GroupRoute:
			group = &v
		default:
			s.logger.Warnw("invalid group type, skipping", "type", fmt.Sprintf("%T", g))
			continue
		}

		// Validate group
		if err := validateGroupRoute(group); err != nil {
			s.logger.Errorw("invalid group route", "error", err, "prefix", group.Prefix)
			continue
		}

		// Apply protection middleware to routes based on group and individual route settings
		s.applyGroupProtection(group)

		// Register group
		group.register(s.app)

		// Store group for reference
		s.groupRoutes = append(s.groupRoutes, group)
	}

	return s
}

// applyGroupProtection applies protection middleware to routes in a group.
// It handles both group-level and individual route-level protection settings.
func (s *HttpServer) applyGroupProtection(group *GroupRoute) {
	for i := range group.Routes {
		// If group is protected, protect all routes in the group
		if group.IsProtected {
			group.Routes[i].IsProtected = true
		}

		// Apply protection middleware if route is protected
		if group.Routes[i].IsProtected {
			s.applyProtectionMiddleware(&group.Routes[i])
		}
	}
}

// applyProtectionMiddleware applies authentication and authorization middleware to a route.
// It prepends auth/authz middlewares to the route's existing middleware chain.
// This ensures security checks run before any route-specific logic.
func (s *HttpServer) applyProtectionMiddleware(route *Route) {
	if route == nil || !route.IsProtected {
		return
	}

	// Calculate capacity: auth + authz (if needed) + existing middlewares
	capacity := len(route.Middlewares)
	if s.authMiddleware != nil {
		capacity++
	}
	if len(route.RequiredPermissions) > 0 && s.authzChecker != nil {
		capacity++
	}

	// Create a new middlewares slice with protection middlewares prepended
	protectionMiddlewares := make([]Middleware, 0, capacity)

	// Add authentication middleware first
	if s.authMiddleware != nil {
		protectionMiddlewares = append(protectionMiddlewares, s.authMiddleware)
	}

	// Add authorization middleware if there are required permissions
	if len(route.RequiredPermissions) > 0 && s.authzChecker != nil {
		authzMiddleware := s.createAuthorizationMiddleware(route.RequiredPermissions)
		protectionMiddlewares = append(protectionMiddlewares, authzMiddleware)
	}

	// Add existing route middlewares
	protectionMiddlewares = append(protectionMiddlewares, route.Middlewares...)

	// Update route middlewares
	route.Middlewares = protectionMiddlewares
}

// createAuthorizationMiddleware creates an authorization middleware for the given permissions.
// It wraps the authzChecker function and returns a proper error response on failure.
func (s *HttpServer) createAuthorizationMiddleware(permissions []string) Middleware {
	return func(ctx Context) error {
		if ctx == nil {
			return fmt.Errorf("context cannot be nil")
		}
		if s.authzChecker == nil {
			return fmt.Errorf("authorization checker not configured")
		}

		if err := s.authzChecker(ctx, permissions); err != nil {
			return Forbidden(ctx, fmt.Sprintf("Insufficient permissions: %v", err))
		}
		return ctx.Next()
	}
}

// Start starts the HTTP server and blocks until shutdown
//
// Input:
//   - none
//
// Output:
//   - error: Any error that occurred during server operation
//
// Example:
//
//	if err := server.Start(); err != nil {
//	    log.Fatal(err)
//	}
func (s *HttpServer) Start() error {
	s.logger.Infow("starting HTTP server",
		"service", s.config.ServiceName,
		"port", s.config.Port,
		"environment", utils.GetEnvironment(),
		"routes", len(s.routes),
		"groups", len(s.groupRoutes),
	)

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
		s.logger.Info("shutting down server...")

		// Create shutdown context with timeout
		timeout := DefaultShutdownTimeout
		if s.config.GracefulShutdownTimeout != nil {
			timeout = *s.config.GracefulShutdownTimeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return s.Shutdown(ctx)
	}
}

// Shutdown gracefully shuts down the server
//
// Input:
//   - ctx: Context for controlling the shutdown timeout
//
// Output:
//   - error: Any error that occurred during shutdown
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), aurelion.DefaultShutdownTimeout)
//	defer cancel()
//	if err := server.Shutdown(ctx); err != nil {
//	    log.Error("shutdown error", err)
//	}
func (s *HttpServer) Shutdown(ctx context.Context) error {
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		s.logger.Errorw("server shutdown failed", "error", err)
		return err
	}

	s.logger.Info("server stopped")
	return nil
}
