// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

//go:generate mockgen -source=server.go -destination=mocks/mock_server.go -package=mocks

import (
	"context"
	"fmt"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/engine"

	"github.com/anthanhphan/gosdk/orianna/pkg/middleware"
	"github.com/anthanhphan/gosdk/orianna/pkg/platform/fiber"
	"github.com/anthanhphan/gosdk/orianna/pkg/routing"
)

// Server Implementation

// HealthCheckManager defines the interface for health check implementations.
type HealthCheckManager interface {
	Check(context.Context) *HealthReport
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
	configValidator   *configuration.ConfigValidator
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
}

// NewServer creates a new server instance with the given configuration and options.
func NewServer(
	conf *configuration.Config,
	options ...ServerOption,
) (*Server, error) {
	// Create logger
	log := logger.NewLoggerWithFields(logger.String("package", "transport"))
	// Validate config
	validator := configuration.NewConfigValidator()
	if err := validator.Validate(conf); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Merge with defaults
	conf = mergeConfig(conf)

	// Create server adapter (default to Fiber)
	serverAdapter, err := fiber.NewServerAdapter(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create server adapter: %w", err)
	}

	// Create server
	server := &Server{
		config:          conf,
		serverAdapter:   serverAdapter,
		routeRegistry:   routing.NewRouteRegistry(),
		hooks:           core.NewHooks(),
		configValidator: validator,

		logger:            log,
		globalMiddlewares: make([]core.Middleware, 0),
		middlewareConfig:  configuration.DefaultMiddlewareConfig(),
	}

	// Apply options
	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Setup route registry with auth
	server.routeRegistry.SetAuthMiddleware(server.authMiddleware)
	server.routeRegistry.SetAuthzChecker(server.authzChecker)

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

	// Routes will be registered when user calls RegisterRoutes
	// No need to register empty routes here

	return server, nil
}

// Start starts the HTTP server and begins listening for incoming requests.
func (s *Server) Start() error {
	// Invoke OnServerStart hook
	if err := s.hooks.ExecuteOnServerStart(s); err != nil {
		return fmt.Errorf("OnServerStart hook failed: %w", err)
	}

	// Log server start with version if available
	if s.config.Version != "" {
		s.logger.Infow("Server started successfully",
			"service", s.config.ServiceName,
			"version", s.config.Version,
			"port", s.config.Port,
		)
	} else {
		s.logger.Infow("Server started successfully",
			"service", s.config.ServiceName,
			"port", s.config.Port,
		)
	}

	return s.serverAdapter.Start()
}

// Shutdown gracefully shuts down the server.
// It always executes shutdown hooks and stops the adapter, even if the shutdown manager fails.
func (s *Server) Shutdown(ctx context.Context) error {
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
