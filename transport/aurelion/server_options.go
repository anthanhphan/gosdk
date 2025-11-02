package aurelion

// ServerOption defines a function type for configuring the server
type ServerOption func(*HttpServer) error

// AuthenticationFunc defines the authentication middleware function
type AuthenticationFunc func(Context) error

// AuthorizationFunc defines the authorization checker function
type AuthorizationFunc func(Context, []string) error

// WithGlobalMiddleware adds global middleware that applies to all routes
//
// Input:
//   - middlewares: One or more middleware functions
//
// Output:
//   - ServerOption: The server option function
//
// Example:
//
//	loggingMiddleware := aurelion.Middleware(func(ctx aurelion.Context) error {
//	    start := time.Now()
//	    err := ctx.Next()
//	    logger.Info("request", "duration", time.Since(start), "path", ctx.Path())
//	    return err
//	})
//	server, err := aurelion.NewHttpServer(config, aurelion.WithGlobalMiddleware(loggingMiddleware))
func WithGlobalMiddleware(middlewares ...Middleware) ServerOption {
	return func(s *HttpServer) error {
		s.globalMiddlewares = append(s.globalMiddlewares, middlewares...)
		return nil
	}
}

// WithPanicRecover sets the panic recovery middleware
//
// Input:
//   - middleware: The panic recovery middleware
//
// Output:
//   - ServerOption: The server option function
//
// Example:
//
//	panicRecover := aurelion.Middleware(func(ctx aurelion.Context) error {
//	    defer func() {
//	        if r := recover(); r != nil {
//	            logger.Error("panic recovered", "panic", r)
//	            aurelion.InternalServerError(ctx, "Internal server error")
//	        }
//	    }()
//	    return ctx.Next()
//	})
//	server, err := aurelion.NewHttpServer(config, aurelion.WithPanicRecover(panicRecover))
func WithPanicRecover(middleware Middleware) ServerOption {
	return func(s *HttpServer) error {
		s.panicRecover = middleware
		return nil
	}
}

// WithAuthentication sets the authentication middleware
//
// Input:
//   - middleware: The authentication middleware
//
// Output:
//   - ServerOption: The server option function
//
// Example:
//
//	authMiddleware := aurelion.Middleware(func(ctx aurelion.Context) error {
//	    token := ctx.Get("Authorization")
//	    if token == "" {
//	        return aurelion.Unauthorized(ctx, "Authentication required")
//	    }
//	    // Validate token and set user info in context
//	    ctx.Locals("user_id", userID)
//	    return ctx.Next()
//	})
//	server, err := aurelion.NewHttpServer(config, aurelion.WithAuthentication(authMiddleware))
func WithAuthentication(middleware Middleware) ServerOption {
	return func(s *HttpServer) error {
		s.authMiddleware = middleware
		return nil
	}
}

// WithAuthorization sets the authorization checker function
//
// Input:
//   - checker: The authorization checker function
//
// Output:
//   - ServerOption: The server option function
//
// Example:
//
//	authChecker := func(ctx aurelion.Context, requiredPermissions []string) error {
//	    userPermissions := ctx.Locals("permissions").([]string)
//	    for _, required := range requiredPermissions {
//	        if !contains(userPermissions, required) {
//	            return errors.New("missing permission")
//	        }
//	    }
//	    return nil
//	}
//	server, err := aurelion.NewHttpServer(config, aurelion.WithAuthorization(authChecker))
func WithAuthorization(checker AuthorizationFunc) ServerOption {
	return func(s *HttpServer) error {
		s.authzChecker = checker
		return nil
	}
}

// WithRateLimiter sets a custom rate limiter middleware.
// If not provided, a default rate limiter of 500 requests per minute per IP is used.
//
// Input:
//   - middleware: The rate limiter middleware
//
// Output:
//   - ServerOption: The server option function
//
// Example:
//
//	// Per User Rate Limiting
//	import "github.com/gofiber/fiber/v2/middleware/limiter"
//
//	customLimiter := limiter.New(limiter.Config{
//	    Max:        1000,
//	    Expiration: 1 * time.Minute,
//	    KeyGenerator: func(c *fiber.Ctx) string {
//	        // Rate limit by user ID instead of IP
//	        if userID := c.Locals("user_id"); userID != nil {
//	            return fmt.Sprintf("user:%v", userID)
//	        }
//	        return c.IP()
//	    },
//	})
//	server, err := aurelion.NewHttpServer(config, aurelion.WithRateLimiter(customLimiter))
//
//	// Global Rate Limiting (all users share same limit)
//	globalLimiter := limiter.New(limiter.Config{
//	    Max:        10000,
//	    Expiration: 1 * time.Minute,
//	    KeyGenerator: func(c *fiber.Ctx) string {
//	        return "global" // All requests share one counter
//	    },
//	})
//	server, err := aurelion.NewHttpServer(config, aurelion.WithRateLimiter(globalLimiter))
//
//	// Per API Key Rate Limiting
//	apiKeyLimiter := limiter.New(limiter.Config{
//	    Max:        5000,
//	    Expiration: 1 * time.Hour,
//	    KeyGenerator: func(c *fiber.Ctx) string {
//	        return c.Get("X-API-Key", "anonymous")
//	    },
//	})
//	server, err := aurelion.NewHttpServer(config, aurelion.WithRateLimiter(apiKeyLimiter))
func WithRateLimiter(middleware Middleware) ServerOption {
	return func(s *HttpServer) error {
		s.rateLimiter = middleware
		return nil
	}
}
