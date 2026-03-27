// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"context"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cache"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

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
		level := configuration.DefaultCompressionLevel
		if s.config.CompressionLevel != nil {
			level = *s.config.CompressionLevel
		}
		s.app.Use(compress.New(compress.Config{
			Level: compress.Level(level),
		}))
	}

	// Add ETag middleware
	if middlewareConfig == nil || !middlewareConfig.DisableETag {
		s.app.Use(etag.New())
	}

	// Add Cache middleware
	if middlewareConfig == nil || !middlewareConfig.DisableCache {
		expiration := configuration.DefaultCacheExpiration
		if s.config.CacheExpiration != nil {
			expiration = *s.config.CacheExpiration
		}
		s.app.Use(cache.New(cache.Config{
			Expiration:  expiration,
			CacheHeader: "X-Cache",
		}))
	}

	// Add request timeout middleware
	if s.config.RequestTimeout != nil && *s.config.RequestTimeout > 0 {
		timeout := *s.config.RequestTimeout
		s.app.Use(func(c fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(c.Context(), timeout)
			defer cancel()
			c.SetContext(ctx)
			err := c.Next()
			if ctx.Err() == context.DeadlineExceeded {
				return c.Status(408).JSON(core.NewErrorResponse("REQUEST_TIMEOUT", 408, "request timeout"))
			}
			return err
		})
	}
}

func (s *ServerAdapter) setupObservabilityMiddlewares(
	middlewareConfig *configuration.MiddlewareConfig,
	panicRecover core.Middleware,
	_ *logger.Logger,
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

	// NOTE: Logging middleware is NOT registered here.
	// It must be registered AFTER tracing middleware via SetupLoggingMiddleware
	// so that trace_id (from OTel span context) is available when logging.
}

// SetupLoggingMiddleware registers the request/response logging middleware.
// This MUST be called AFTER tracing middleware is registered, so that the
// OTel trace_id is present in the span context when the logging middleware reads it.
func (s *ServerAdapter) SetupLoggingMiddleware(
	middlewareConfig *configuration.MiddlewareConfig,
	log *logger.Logger,
) {
	if middlewareConfig == nil || !middlewareConfig.DisableLogging {
		if log != nil {
			s.app.Use(requestResponseLoggingMiddleware(log, s.config.VerboseLogging, s.config.VerboseLoggingSkipPaths))
		}
	}
}

func (s *ServerAdapter) setupCoreMiddlewares(globalMiddlewares []core.Middleware) {
	// Add custom global middlewares
	for _, middleware := range globalMiddlewares {
		s.app.Use(convertToFiberMiddlewareWithConfig(middleware, s.config))
	}
}
