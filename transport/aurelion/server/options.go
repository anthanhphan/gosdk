package server

import (
	"fmt"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/runtimectx"
	"github.com/gofiber/fiber/v2"
)

// ServerOption defines a function type for configuring the server.
type ServerOption func(*HttpServer) error

// AuthenticationFunc defines the authentication middleware function.
type AuthenticationFunc core.Middleware

// AuthorizationFunc defines the authorization checker function.
type AuthorizationFunc func(core.Context, []string) error

// WithGlobalMiddleware adds global middleware that applies to all routes.
func WithGlobalMiddleware(middlewares ...interface{}) ServerOption {
	return func(s *HttpServer) error {
		for _, m := range middlewares {
			handler, err := normalizeFiberHandler(m)
			if err != nil {
				return fmt.Errorf("with global middleware: %w", err)
			}
			if handler != nil {
				s.globalMiddlewares = append(s.globalMiddlewares, handler)
			}
		}
		return nil
	}
}

// WithPanicRecover sets the panic recovery middleware.
func WithPanicRecover(m interface{}) ServerOption {
	return func(s *HttpServer) error {
		handler, err := normalizeFiberHandler(m)
		if err != nil {
			return fmt.Errorf("with panic recover: %w", err)
		}
		s.panicRecover = handler
		return nil
	}
}

// WithAuthentication sets the authentication middleware.
func WithAuthentication(m core.Middleware) ServerOption {
	return func(s *HttpServer) error {
		s.authMiddleware = m
		return nil
	}
}

// WithAuthorization sets the authorization checker function.
func WithAuthorization(checker AuthorizationFunc) ServerOption {
	return func(s *HttpServer) error {
		s.authzChecker = checker
		return nil
	}
}

// WithRateLimiter sets a custom rate limiter middleware.
func WithRateLimiter(m interface{}) ServerOption {
	return func(s *HttpServer) error {
		handler, err := normalizeFiberHandler(m)
		if err != nil {
			return fmt.Errorf("with rate limiter: %w", err)
		}
		s.rateLimiter = handler
		return nil
	}
}

func normalizeFiberHandler(m interface{}) (fiber.Handler, error) {
	if m == nil {
		return nil, nil
	}

	switch v := m.(type) {
	case fiber.Handler:
		return v, nil
	case core.Middleware:
		return rctx.MiddlewareToFiber(v), nil
	default:
		return nil, fmt.Errorf("unsupported middleware type %T", m)
	}
}
