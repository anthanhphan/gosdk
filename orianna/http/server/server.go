// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

//go:generate mockgen -source=server.go -destination=mocks/mock_server.go -package=mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/engine"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"github.com/anthanhphan/gosdk/tracing"

	"github.com/anthanhphan/gosdk/orianna/http/middleware"
	"github.com/anthanhphan/gosdk/orianna/http/platform/fiber"
	"github.com/anthanhphan/gosdk/orianna/http/routing"
)

// Server Implementation

// HealthCheckManager defines the interface for health check implementations.
type HealthCheckManager interface {
	Check(context.Context) *health.HealthReport
}

// ShutdownManager defines the interface for shutdown management.
type ShutdownManager interface {
	Shutdown(context.Context) error
}

// Server orchestrates domain objects and infrastructure adapters
type Server struct {
	config            *configuration.Config
	serverAdapter     engine.ServerEngine
	routeRegistry     *routing.RouteRegistry
	hooks             *core.Hooks
	healthManager     HealthCheckManager
	shutdownManager   ShutdownManager
	logger            *logger.Logger
	globalMiddlewares []core.Middleware
	panicRecover      core.Middleware
	authMiddleware    core.Middleware
	authzChecker      func(core.Context, []string) error
	rateLimiter       core.Middleware
	middlewareConfig  *configuration.MiddlewareConfig
	metricsClient     metrics.Client
	tracingClient     tracing.Client
}

// NewServer creates a new server instance with the given configuration and options.
func NewServer(
	conf *configuration.Config,
	options ...ServerOption,
) (*Server, error) {
	// Create logger
	log := logger.NewLoggerWithFields(logger.String("package", "transport"))
	// Validate config
	if conf == nil {
		return nil, fmt.Errorf("invalid config: config cannot be nil")
	}
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Merge with defaults
	conf = mergeConfig(conf)

	// Create server (adapter set after options)
	server := &Server{
		config:        conf,
		routeRegistry: routing.NewRouteRegistry(),
		hooks:         core.NewHooks(),

		logger:            log,
		globalMiddlewares: nil,
		middlewareConfig:  configuration.DefaultMiddlewareConfig(),
	}

	// Apply options (may set a custom server engine via WithServerEngine)
	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create default Fiber adapter if no custom engine was provided
	if server.serverAdapter == nil {
		serverAdapter, err := fiber.NewServerAdapter(conf)
		if err != nil {
			return nil, fmt.Errorf("failed to create server adapter: %w", err)
		}
		server.serverAdapter = serverAdapter
	}

	// Setup route registry with auth
	server.routeRegistry.SetAuthMiddleware(server.authMiddleware)
	server.routeRegistry.SetAuthzChecker(server.authzChecker)

	// Auto-disable legacy traceID middleware when OTel tracing is active
	// to avoid duplicate/conflicting trace IDs
	if server.tracingClient != nil {
		server.middlewareConfig.DisableTraceID = true
	}

	// Setup global middlewares on adapter
	server.serverAdapter.SetupGlobalMiddlewares(
		server.middlewareConfig,
		server.globalMiddlewares,
		server.panicRecover,
		server.rateLimiter,
		server.logger,
	)

	// Setup metrics if enabled
	if server.metricsClient != nil {
		server.Use(middleware.MetricsMiddleware(server.metricsClient, server.config.ServiceName))

		// Register /metrics endpoint for Prometheus scraping
		server.serverAdapter.RegisterMetricsHandler(server.metricsClient)
	}

	// Setup slow request detection if threshold is configured
	if server.config.SlowRequestThreshold > 0 {
		server.Use(middleware.SlowRequestDetector(server.config.SlowRequestThreshold))
	}

	// Setup tracing if enabled and not disabled via middleware config
	if server.tracingClient != nil && !server.middlewareConfig.DisableTracing {
		server.Use(middleware.TracingMiddleware(server.tracingClient))
	}

	// Setup logging middleware AFTER tracing so trace_id is available in logs
	server.serverAdapter.SetupLoggingMiddleware(
		server.middlewareConfig,
		server.logger,
	)

	// Setup hooks middleware AFTER logging so hooks see full request context
	// (trace_id, request_id, etc. are all available).
	// Register directly on adapter (not via Use()) to avoid polluting
	// the user-visible globalMiddlewares slice.
	if server.hooks != nil {
		server.serverAdapter.Use(hooksMiddleware(server.hooks))
	}

	// Routes will be registered when user calls RegisterRoutes
	// No need to register empty routes here

	return server, nil
}

// hooksMiddleware creates a middleware that fires request lifecycle hooks.
// Hooks are executed with panic recovery to prevent a faulty hook from
// crashing the server.
func hooksMiddleware(hooks *core.Hooks) core.Middleware {
	return func(ctx core.Context) error {
		hooks.ExecuteOnRequest(ctx)

		start := time.Now()
		err := ctx.Next()
		latency := time.Since(start)

		hooks.ExecuteOnResponse(ctx, ctx.ResponseStatusCode(), latency)

		if err != nil {
			hooks.ExecuteOnError(ctx, err)
		}
		return err
	}
}

// Start starts the HTTP server and begins listening for incoming requests.
func (s *Server) Start() error {
	// Invoke OnServerStart hook
	if err := s.hooks.ExecuteOnServerStart(s); err != nil {
		return fmt.Errorf("OnServerStart hook failed: %w", err)
	}

	// Log server start with version if available
	logFields := []any{"service", s.config.ServiceName, "port", s.config.Port}
	if s.config.Version != "" {
		logFields = append(logFields, "version", s.config.Version)
	}
	s.logger.Infow("Server started successfully", logFields...)

	return s.serverAdapter.Start()
}

// Shutdown gracefully shuts down the server.
// It always executes shutdown hooks and stops the adapter, even if the shutdown manager fails.
// If the caller's context has no deadline and GracefulShutdownTimeout is configured,
// a timeout is applied automatically to ensure in-flight requests can complete.
func (s *Server) Shutdown(ctx context.Context) error {
	// Apply configured timeout if caller didn't set a deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && s.config.GracefulShutdownTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *s.config.GracefulShutdownTimeout)
		defer cancel()
	}

	var firstErr error

	// Use shutdown manager if available
	if s.shutdownManager != nil {
		if err := s.shutdownManager.Shutdown(ctx); err != nil {
			firstErr = fmt.Errorf("shutdown manager failed: %w", err)
		}
	}

	// Always invoke OnShutdown hooks regardless of shutdown manager result
	s.hooks.ExecuteOnShutdown()

	// Always stop the server adapter to release the listening port
	if err := s.serverAdapter.Shutdown(ctx); err != nil {
		if firstErr == nil {
			firstErr = fmt.Errorf("server adapter shutdown failed: %w", err)
		}
	}

	return firstErr
}

// RegisterRoutes registers one or more routes with the server.
func (s *Server) RegisterRoutes(routes ...routing.Route) error {
	if err := s.routeRegistry.RegisterRoutes(routes...); err != nil {
		return err
	}
	// Register newly added routes to adapter
	return s.serverAdapter.RegisterRoutes(routes...)
}

// RegisterGroup registers a route group containing multiple routes with a common prefix.
func (s *Server) RegisterGroup(group routing.RouteGroup) error {
	if err := s.routeRegistry.RegisterGroup(group); err != nil {
		return err
	}
	// Register to adapter
	return s.serverAdapter.RegisterGroup(group)
}

// Use adds global middleware to the server.
func (s *Server) Use(middleware ...core.Middleware) {
	s.globalMiddlewares = append(s.globalMiddlewares, middleware...)
	s.serverAdapter.Use(middleware...)
}

// GetHealthManager returns the health check manager
func (s *Server) GetHealthManager() HealthCheckManager {
	return s.healthManager
}

// GetShutdownManager returns the shutdown manager
func (s *Server) GetShutdownManager() ShutdownManager {
	return s.shutdownManager
}

// mergeConfig merges config with defaults (creates a copy to avoid mutating the input)
func mergeConfig(conf *configuration.Config) *configuration.Config {
	merged := *conf // shallow copy
	if merged.MaxBodySize == 0 {
		merged.MaxBodySize = configuration.DefaultMaxBodySize
	}
	if merged.MaxConcurrentConnections == 0 {
		merged.MaxConcurrentConnections = configuration.DefaultMaxConcurrentConnections
	}
	if merged.Port == 0 {
		merged.Port = configuration.DefaultPort
	}
	// Merge timeout defaults when user hasn't set them
	if merged.ReadTimeout == nil {
		d := configuration.DefaultReadTimeout
		merged.ReadTimeout = &d
	}
	if merged.WriteTimeout == nil {
		d := configuration.DefaultWriteTimeout
		merged.WriteTimeout = &d
	}
	if merged.IdleTimeout == nil {
		d := configuration.DefaultIdleTimeout
		merged.IdleTimeout = &d
	}
	if merged.GracefulShutdownTimeout == nil {
		d := configuration.DefaultGracefulShutdownTimeout
		merged.GracefulShutdownTimeout = &d
	}
	return &merged
}
