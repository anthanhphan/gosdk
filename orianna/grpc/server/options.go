// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"google.golang.org/grpc"

	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/orianna/grpc/interceptor"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"github.com/anthanhphan/gosdk/tracing"
)

// ServerOption defines a function type for configuring the gRPC server.
type ServerOption func(*Server) error

// WithGlobalUnaryInterceptor adds global unary interceptors that apply to all RPCs.
func WithGlobalUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) error {
		s.globalUnary = append(s.globalUnary, interceptors...)
		return nil
	}
}

// WithGlobalStreamInterceptor adds global stream interceptors.
func WithGlobalStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) error {
		s.globalStream = append(s.globalStream, interceptors...)
		return nil
	}
}

// WithPanicRecover sets the panic recovery interceptor.
func WithPanicRecover(unary grpc.UnaryServerInterceptor, stream grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) error {
		s.panicRecover = unary
		s.streamRecover = stream
		return nil
	}
}

// WithDisableRecovery disables the automatic panic recovery interceptor.
// WARNING: Never disable in production!
func WithDisableRecovery() ServerOption {
	return func(s *Server) error {
		s.disableRecovery = true
		return nil
	}
}

// WithRateLimiter sets a custom rate limiter interceptor.
func WithRateLimiter(interceptor grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) error {
		s.rateLimiter = interceptor
		return nil
	}
}

// WithHooks sets lifecycle hooks for the server.
func WithHooks(hooks *core.Hooks) ServerOption {
	return func(s *Server) error {
		s.hooks = hooks
		return nil
	}
}

// WithHealthManager sets the health check manager.
func WithHealthManager(manager *health.Manager) ServerOption {
	return func(s *Server) error {
		s.healthManager = manager
		return nil
	}
}

// WithMetrics adds metrics interceptors to the server.
func WithMetrics(client metrics.Client) ServerOption {
	return func(s *Server) error {
		s.metricsClient = client
		return nil
	}
}

// WithTracing adds tracing interceptors to the server.
// Automatically wires both unary and stream tracing interceptors.
func WithTracing(client tracing.Client) ServerOption {
	return func(s *Server) error {
		s.tracingClient = client
		return nil
	}
}

// WithHealthChecker adds a health checker to the server.
// If no health manager exists, one is created automatically.
func WithHealthChecker(checker health.Checker) ServerOption {
	return func(s *Server) error {
		if s.healthManager == nil {
			s.healthManager = health.NewManager()
		}
		s.healthManager.Register(checker)
		return nil
	}
}

// WithTokenAuth sets a token-based authentication validator for the server.
// The TokenAuthInterceptor is automatically wired into the interceptor chain
// after certificate auth (if configured) and before rate limiting.
// This is designed for internal service-to-service authentication using
// JWT, OAuth2, or any token-based mechanism.
//
// Example:
//
//	srv, err := server.NewServer(config,
//	    server.WithTokenAuth(myJWTValidator),
//	)
func WithTokenAuth(validator interceptor.TokenValidator) ServerOption {
	return func(s *Server) error {
		s.tokenValidator = validator
		return nil
	}
}
