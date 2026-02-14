// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// ServerOption defines a function type for configuring the server
type ServerOption func(*Server) error

// WithGlobalMiddleware adds global middleware that applies to all routes
func WithGlobalMiddleware(middlewares ...core.Middleware) ServerOption {
	return func(s *Server) error {
		s.globalMiddlewares = append(s.globalMiddlewares, middlewares...)
		return nil
	}
}

// WithPanicRecover sets the panic recovery middleware
func WithPanicRecover(middleware core.Middleware) ServerOption {
	return func(s *Server) error {
		s.panicRecover = middleware
		return nil
	}
}

// WithAuthentication sets the authentication middleware
func WithAuthentication(middleware core.Middleware) ServerOption {
	return func(s *Server) error {
		s.authMiddleware = middleware
		return nil
	}
}

// WithAuthorization sets the authorization checker function
func WithAuthorization(checker func(core.Context, []string) error) ServerOption {
	return func(s *Server) error {
		s.authzChecker = checker
		return nil
	}
}

// WithRateLimiter sets a custom rate limiter middleware
func WithRateLimiter(middleware core.Middleware) ServerOption {
	return func(s *Server) error {
		s.rateLimiter = middleware
		return nil
	}
}

// WithHooks sets lifecycle hooks for the server
func WithHooks(hooks *core.Hooks) ServerOption {
	return func(s *Server) error {
		s.hooks = hooks
		return nil
	}
}

// WithMiddlewareConfig sets the middleware configuration
func WithMiddlewareConfig(config *configuration.MiddlewareConfig) ServerOption {
	return func(s *Server) error {
		s.middlewareConfig = config
		return nil
	}
}

// WithHealthManager sets the health check manager
func WithHealthManager(manager HealthCheckManager) ServerOption {
	return func(s *Server) error {
		s.healthManager = manager
		return nil
	}
}

// WithShutdownManager sets the shutdown manager
func WithShutdownManager(manager ShutdownManager) ServerOption {
	return func(s *Server) error {
		s.shutdownManager = manager
		return nil
	}
}

// WithMetrics adds metrics middleware to the server
func WithMetrics(client metrics.Client) ServerOption {
	return func(s *Server) error {
		s.metricsClient = client
		return nil
	}
}

// WithHealthChecker adds a health checker to the server
func WithHealthChecker(checker HealthChecker) ServerOption {
	return func(s *Server) error {
		// Get or create health manager
		manager := s.GetHealthManager()
		if manager == nil {
			// Create default health manager
			healthManager := NewHealthManager()
			manager = healthManager
			if err := WithHealthManager(healthManager)(s); err != nil {
				return err
			}
		}

		// Register checker (need to access underlying manager)
		if mgr, ok := manager.(*HealthManager); ok {
			mgr.Register(checker)
		}
		return nil
	}
}
