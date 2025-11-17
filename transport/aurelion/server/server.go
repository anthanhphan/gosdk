package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/transport/aurelion/config"
	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/anthanhphan/gosdk/transport/aurelion/middleware"
	"github.com/anthanhphan/gosdk/transport/aurelion/response"
	"github.com/anthanhphan/gosdk/transport/aurelion/router"
	"github.com/anthanhphan/gosdk/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

const (
	DefaultHealthCheckPath = "/health"
	DefaultRateLimitMax    = 500
	DefaultRateWindow      = time.Minute
)

// HttpServer represents an HTTP server instance.
type HttpServer struct {
	app               *fiber.App
	config            *config.Config
	globalMiddlewares []fiber.Handler
	panicRecover      fiber.Handler
	authMiddleware    core.Middleware
	authzChecker      AuthorizationFunc
	rateLimiter       fiber.Handler
	routes            []*router.Route
	groupRoutes       []*router.GroupRoute
	logger            *zap.SugaredLogger
}

// NewHttpServer creates a new HTTP server with the given configuration and options.
func NewHttpServer(cfg *config.Config, options ...ServerOption) (*HttpServer, error) {
	if cfg == nil {
		return nil, core.ErrConfigNil
	}

	cfg = cfg.Merge()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	fiberConfig := fiber.Config{
		AppName:               cfg.ServiceName,
		BodyLimit:             cfg.MaxBodySize,
		Concurrency:           cfg.MaxConcurrentConnections,
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
	}

	if cfg.ReadTimeout != nil {
		fiberConfig.ReadTimeout = *cfg.ReadTimeout
	}
	if cfg.WriteTimeout != nil {
		fiberConfig.WriteTimeout = *cfg.WriteTimeout
	}
	if cfg.IdleTimeout != nil {
		fiberConfig.IdleTimeout = *cfg.IdleTimeout
	}

	app := fiber.New(fiberConfig)

	server := &HttpServer{
		app:               app,
		config:            cfg,
		globalMiddlewares: make([]fiber.Handler, 0),
		routes:            make([]*router.Route, 0),
		groupRoutes:       make([]*router.GroupRoute, 0),
		logger:            logger.NewLoggerWithFields(zap.String("package", "aurelion/server")),
	}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	server.setupGlobalMiddlewares()

	server.AddRoutes(router.NewRoute(DefaultHealthCheckPath).
		GET().
		Handler(func(ctx core.Context) error {
			config.StoreInContext(ctx, cfg)
			return response.HealthCheck(ctx)
		}),
	)

	return server, nil
}

func (s *HttpServer) setupGlobalMiddlewares() {
	s.app.Use(helmet.New())

	if s.rateLimiter != nil {
		s.app.Use(s.rateLimiter)
	} else {
		s.app.Use(limiter.New(limiter.Config{
			Max:        DefaultRateLimitMax,
			Expiration: DefaultRateWindow,
		}))
	}

	s.app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))

	if s.panicRecover != nil {
		s.app.Use(s.panicRecover)
	} else {
		s.app.Use(recover.New())
	}

	s.app.Use(middleware.ConfigInjector(s.config))
	s.app.Use(middleware.RequestIDMiddleware())
	s.app.Use(middleware.TraceIDMiddleware())
	s.app.Use(middleware.RequestResponseLogger(s.logger, s.config.VerboseLogging))

	for _, handler := range s.globalMiddlewares {
		s.app.Use(handler)
	}

	if s.config.EnableCORS {
		s.app.Use(cors.New(middleware.BuildCORSConfig(s.config.CORS)))
	}

	if s.config.EnableCSRF {
		s.app.Use(csrf.New(middleware.BuildCSRFConfig(s.config.CSRF)))
	}
}

// AddRoutes adds multiple routes to the server.
func (s *HttpServer) AddRoutes(routes ...interface{}) *HttpServer {
	for _, entry := range routes {
		route := router.ToRoute(entry)
		if route == nil {
			s.logger.Warnw("invalid route type, skipping", "type", fmt.Sprintf("%T", entry))
			continue
		}

		routeCopy := route.Clone()
		if err := router.ValidateRoute(routeCopy); err != nil {
			s.logger.Errorw("invalid route", "error", err, "path", routeCopy.Path)
			continue
		}

		s.applyProtection(routeCopy)
		s.registerRoute(routeCopy)
		s.routes = append(s.routes, routeCopy)
	}

	return s
}

// AddGroupRoutes adds multiple group routes to the server.
func (s *HttpServer) AddGroupRoutes(groups ...interface{}) *HttpServer {
	for _, entry := range groups {
		group := router.ToGroupRoute(entry)
		if group == nil {
			s.logger.Warnw("invalid group type, skipping", "type", fmt.Sprintf("%T", entry))
			continue
		}

		groupCopy := *group
		if err := router.ValidateGroupRoute(&groupCopy); err != nil {
			s.logger.Errorw("invalid group route", "error", err, "prefix", groupCopy.Prefix)
			continue
		}

		s.applyGroupProtection(&groupCopy)
		groupCopy.Register(s.app)
		groupClone := groupCopy
		s.groupRoutes = append(s.groupRoutes, &groupClone)
	}

	return s
}

func (s *HttpServer) applyGroupProtection(group *router.GroupRoute) {
	for i := range group.Routes {
		if group.IsProtected {
			group.Routes[i].IsProtected = true
		}
		s.applyProtection(&group.Routes[i])
	}
}

func (s *HttpServer) applyProtection(route *router.Route) {
	if route == nil || !route.IsProtected {
		return
	}

	// Convert router.Middleware to core.Middleware for auth middlewares
	authMiddlewares := make([]router.Middleware, 0, 2)

	if s.authMiddleware != nil {
		// Convert core.Middleware to router.Middleware
		authMiddlewares = append(authMiddlewares, func(ctx router.ContextInterface) error {
			coreCtx := router.AdaptRouterContextToCore(ctx)
			return s.authMiddleware(coreCtx)
		})
	}

	if len(route.RequiredPermissions) > 0 && s.authzChecker != nil {
		authzMw := s.createAuthorizationMiddleware(route.RequiredPermissions)
		// Convert core.Middleware to router.Middleware
		authMiddlewares = append(authMiddlewares, func(ctx router.ContextInterface) error {
			coreCtx := router.AdaptRouterContextToCore(ctx)
			return authzMw(coreCtx)
		})
	}

	// Prepend auth middlewares to route middlewares
	route.Middlewares = append(authMiddlewares, route.Middlewares...)
}

func (s *HttpServer) createAuthorizationMiddleware(permissions []string) core.Middleware {
	return func(ctx core.Context) error {
		if ctx == nil {
			return core.ErrContextNil
		}
		if s.authzChecker == nil {
			return response.Forbidden(ctx, "authorization checker not configured")
		}

		if err := s.authzChecker(ctx, permissions); err != nil {
			return response.Forbidden(ctx, fmt.Sprintf("insufficient permissions: %v", err))
		}
		return ctx.Next()
	}
}

func (s *HttpServer) registerRoute(route *router.Route) {
	if route == nil {
		return
	}

	if route.CORS != nil {
		group := s.app.Group("", cors.New(middleware.BuildCORSConfig(route.CORS)))
		handlers := routeHandlers(route)
		registerWithRouter(group, route.Method, route.Path, handlers)
		return
	}

	route.Register(s.app)
}

func routeHandlers(route *router.Route) []fiber.Handler {
	return route.Clone().Handlers()
}

func registerWithRouter(rt fiber.Router, method interface{}, path string, handlers []fiber.Handler) {
	// Convert method to string
	var methodStr string
	switch m := method.(type) {
	case core.Method:
		methodStr = string(m)
	case string:
		methodStr = m
	default:
		// Try to convert router.Method by checking if it's a string type
		if str, ok := m.(string); ok {
			methodStr = str
		} else {
			return
		}
	}
	if rt == nil || len(handlers) == 0 {
		return
	}

	switch methodStr {
	case "GET":
		rt.Get(path, handlers...)
	case "POST":
		rt.Post(path, handlers...)
	case "PUT":
		rt.Put(path, handlers...)
	case "PATCH":
		rt.Patch(path, handlers...)
	case "DELETE":
		rt.Delete(path, handlers...)
	case "HEAD":
		rt.Head(path, handlers...)
	case "OPTIONS":
		rt.Options(path, handlers...)
	}
}

// Start starts the HTTP server and blocks until shutdown.
func (s *HttpServer) Start() error {
	s.logger.Infow("starting HTTP server",
		"service", s.config.ServiceName,
		"port", s.config.Port,
		"environment", utils.GetEnvironment(),
		"routes", len(s.routes),
		"groups", len(s.groupRoutes),
	)

	errChan := make(chan error, 1)

	go func() {
		if err := s.app.Listen(fmt.Sprintf(":%d", s.config.Port)); err != nil {
			errChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-quit:
		s.logger.Info("shutting down server...")
		return s.shutdown()
	}
}

func (s *HttpServer) shutdown() error {
	timeout := config.DefaultShutdownTimeout
	if s.config.GracefulShutdownTimeout != nil {
		timeout = *s.config.GracefulShutdownTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		s.logger.Errorw("server shutdown failed", "error", err)
		return err
	}

	s.logger.Info("server stopped")
	return nil
}

// App exposes the underlying fiber app for advanced customisation.
func (s *HttpServer) App() *fiber.App {
	return s.app
}
