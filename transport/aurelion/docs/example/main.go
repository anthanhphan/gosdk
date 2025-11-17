package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/transport/aurelion"
	"github.com/anthanhphan/gosdk/utils"
)

func main() {
	// Initialize logger
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelDebug,
		LogEncoding:       logger.EncodingJSON,
		DisableCaller:     false,
		DisableStacktrace: true,
	})
	defer undo()

	// Create server configuration
	shutdownTimeout := 30 * time.Second
	config := &aurelion.Config{
		ServiceName:             "Demo API",
		Port:                    8000,
		GracefulShutdownTimeout: &shutdownTimeout,
		EnableCORS:              false,
		VerboseLogging:          true, // Log request/response bodies for demo
		UseProperHTTPStatus:     true, // Use proper HTTP status codes (400, 401, etc.)
	}

	// Create server with middleware options
	server, err := aurelion.NewHttpServer(
		config,
		aurelion.WithPanicRecover(panicRecoverMiddleware()),
		aurelion.WithAuthentication(authMiddleware()),
		aurelion.WithAuthorization(authzChecker()),
		aurelion.WithGlobalMiddleware(aurelion.DefaultHeaderToLocalsMiddleware()),
	)
	if err != nil {
		logger.NewLoggerWithFields().Fatalw("failed to create server", "error", err)
	}

	// Register routes
	registerRoutes(server)

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		logger.NewLoggerWithFields().Fatalw("server error", "error", err)
	}
}

// registerRoutes demonstrates all key features
func registerRoutes(server *aurelion.HttpServer) {
	// 1. Public GET routes
	server.AddRoutes(
		// Basic GET with query parameter
		aurelion.NewRoute("/search").
			GET().
			Handler(func(ctx aurelion.Context) error {
				query := ctx.Query("q", "")
				if query == "" {
					return aurelion.BadRequest(ctx, "Query parameter 'q' is required")
				}
				return aurelion.OK(ctx, "Search results", aurelion.Map{"query": query, "results": []string{"result1", "result2"}})
			}),

		// GET with route parameter
		aurelion.NewRoute("/users/:id").
			GET().
			Handler(func(ctx aurelion.Context) error {
				userID := ctx.Params("id")
				// Simulate: user not found
				if userID == "999" {
					return aurelion.Error(ctx, aurelion.NewErrorf(1001, "User with ID %s not found", userID))
				}
				return aurelion.OK(ctx, "User details", aurelion.Map{
					"id":   userID,
					"name": "John Doe",
				})
			}),

		// GET with request ID (auto-generated)
		aurelion.NewRoute("/info").
			GET().
			Protected().
			Handler(func(ctx aurelion.Context) error {
				requestID := aurelion.GetRequestID(ctx)

				testFunc := func(ctx context.Context, requestID string) error {
					if requestID == "" {
						return errors.New("request ID is empty")
					}

					userID := aurelion.GetUserIDFromContext(ctx)
					if userID == "" {
						return errors.New("user ID is empty")
					}

					lang := aurelion.GetLanguageFromContext(ctx)
					if lang == "" {
						return errors.New("language is empty")
					}

					logger.Infow("request ID", "request_id", aurelion.GetRequestIDFromContext(ctx), "user_id", userID, "lang", lang, "trace_id", aurelion.GetTraceIDFromContext(ctx))
					return nil
				}
				return testFunc(ctx.Context(), requestID)
			}),
	)

	// 2. POST routes with body parsing
	server.AddRoutes(
		// POST with manual validation (old way)
		aurelion.NewRoute("/users").
			POST().
			Handler(func(ctx aurelion.Context) error {
				type CreateUserRequest struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}

				var req CreateUserRequest
				if err := ctx.BodyParser(&req); err != nil {
					return aurelion.BadRequest(ctx, "Invalid request body")
				}

				// Business validation
				if req.Email == "existing@example.com" {
					return aurelion.Error(ctx, aurelion.NewErrorf(1002, "User with email %s already exists", req.Email))
				}

				return ctx.Status(201).JSON(aurelion.Map{
					"message": "User created",
					"user":    aurelion.Map{"id": "123", "name": req.Name, "email": req.Email},
				})
			}),

		// POST with struct validation (recommended way)
		aurelion.NewRoute("/users/validated").
			POST().
			Handler(func(ctx aurelion.Context) error {
				type CreateUserRequest struct {
					Name  string `json:"name" validate:"required,min=3,max=50"`
					Email string `json:"email" validate:"required,email"`
					Age   int    `json:"age" validate:"min=18,max=100"`
				}

				var req CreateUserRequest

				// Parse and validate
				if err := aurelion.ValidateAndParse(ctx, &req); err != nil {
					// Use ErrorWithDetails for validation errors with constants
					if validationErr, ok := err.(aurelion.ValidationErrors); ok {
						return aurelion.ErrorWithDetails(ctx, 400, aurelion.MsgValidationFailed, &aurelion.ErrorData{
							Type:       aurelion.ErrorTypeValidation,
							Validation: validationErr.ToArray(),
						})
					}
					return aurelion.BadRequest(ctx, err.Error())
				}

				// Business validation
				if req.Email == "existing@example.com" {
					return aurelion.Error(ctx, aurelion.NewErrorf(1002, "User with email %s already exists", req.Email))
				}

				return ctx.Status(201).JSON(aurelion.Map{
					"message": "User created with validation",
					"user":    aurelion.Map{"id": "123", "name": req.Name, "email": req.Email, "age": req.Age},
				})
			}),
	)

	// 3. Protected routes (requires authentication)
	server.AddRoutes(
		aurelion.NewRoute("/protected").
			GET().
			Protected().
			Handler(func(ctx aurelion.Context) error {
				userID := ctx.Locals("user_id")
				return aurelion.OK(ctx, "Protected endpoint", aurelion.Map{"user_id": userID})
			}),
	)

	// 4. Protected routes with authorization (requires specific permissions)
	server.AddRoutes(
		aurelion.NewRoute("/admin/users").
			GET().
			Protected().
			Permissions("read:users").
			Handler(func(ctx aurelion.Context) error {
				return aurelion.OK(ctx, "Users list", aurelion.Map{
					"users": []aurelion.Map{{"id": 1, "name": "User 1"}},
				})
			}),
	)

	// 5. Routes with custom middleware
	server.AddRoutes(
		// Route with custom logging middleware
		aurelion.NewRoute("/log-me").
			GET().
			Middleware(customLoggingMiddleware()).
			Handler(func(ctx aurelion.Context) error {
				return aurelion.OK(ctx, "Logged request", aurelion.Map{"message": "This route has custom logging"})
			}),

		// Route with rate limiting middleware (custom)
		aurelion.NewRoute("/limited").
			GET().
			Middleware(rateLimitMiddleware()).
			Handler(func(ctx aurelion.Context) error {
				return aurelion.OK(ctx, "Rate limited", aurelion.Map{"message": "This route has custom rate limiting"})
			}),
	)

	// 6. Route groups
	server.AddGroupRoutes(
		aurelion.NewGroupRoute("/api/v1").
			Routes(
				aurelion.NewRoute("/status").
					GET().
					Handler(func(ctx aurelion.Context) error {
						return aurelion.OK(ctx, "API status", aurelion.Map{"version": "1.0.0"})
					}),
			),

		// Protected group (all routes require authentication)
		aurelion.NewGroupRoute("/api/admin").
			Protected().
			Routes(
				aurelion.NewRoute("/dashboard").
					GET().
					Handler(func(ctx aurelion.Context) error {
						return aurelion.OK(ctx, "Dashboard", aurelion.Map{"stats": aurelion.Map{"users": 100}})
					}),
			),

		// Group with custom middleware (applies to all routes in the group)
		aurelion.NewGroupRoute("/api/metrics").
			Middleware(metricsMiddleware()).
			Routes(
				aurelion.NewRoute("/visitors").
					GET().
					Handler(func(ctx aurelion.Context) error {
						return aurelion.OK(ctx, "Visitor metrics", aurelion.Map{"visitors": 1000})
					}),

				aurelion.NewRoute("/requests").
					GET().
					Handler(func(ctx aurelion.Context) error {
						return aurelion.OK(ctx, "Request metrics", aurelion.Map{"requests": 5000})
					}),
			),
	)

	// 7. Error examples
	server.AddRoutes(
		aurelion.NewRoute("/errors/bad-request").
			GET().
			Handler(func(ctx aurelion.Context) error {
				return aurelion.BadRequest(ctx, "Invalid input")
			}),

		aurelion.NewRoute("/errors/not-found").
			GET().
			Handler(func(ctx aurelion.Context) error {
				return aurelion.NotFound(ctx, "Resource not found")
			}),

		aurelion.NewRoute("/panic").
			GET().
			Handler(func(ctx aurelion.Context) error {
				panic("Demonstrates panic recovery")
			}),
	)
}

// authMiddleware demonstrates authentication
func authMiddleware() aurelion.Middleware {
	return func(ctx aurelion.Context) error {
		token := ctx.Get("Authorization")
		if token == "" {
			return aurelion.Unauthorized(ctx, "Authentication required")
		}

		// Demo: simple token validation
		if token != "Bearer valid-token" {
			return aurelion.Unauthorized(ctx, "Invalid token")
		}

		// Set user info in context (using Locals for aurelion.Context compatibility)
		// The Context() method will automatically merge Locals values into context.Context
		ctx.Locals("user_id", "123")
		ctx.Locals("username", "demo_user")
		ctx.Locals("role", "admin")
		ctx.Locals("permissions", []string{"read:users", "write:users"})

		return ctx.Next()
	}
}

// authzChecker demonstrates authorization
func authzChecker() func(aurelion.Context, []string) error {
	return func(ctx aurelion.Context, requiredPermissions []string) error {
		userPermissions, ok := ctx.Locals("permissions").([]string)
		if !ok {
			return errors.New("no permissions found")
		}

		for _, required := range requiredPermissions {
			found := false
			for _, perm := range userPermissions {
				if perm == required {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("missing permission: %s", required)
			}
		}

		return nil
	}
}

// panicRecoverMiddleware demonstrates panic recovery
func panicRecoverMiddleware() aurelion.Middleware {
	return func(ctx aurelion.Context) error {
		defer func() {
			if r := recover(); r != nil {
				panicLocation, _ := utils.GetPanicLocation()
				logger.Errorw("panic recovered", "panic", r, "panic_file", panicLocation)
				_ = aurelion.InternalServerError(ctx, "Internal server error")
			}
		}()
		return ctx.Next()
	}
}

// customLoggingMiddleware demonstrates custom middleware for specific routes
func customLoggingMiddleware() aurelion.Middleware {
	return func(ctx aurelion.Context) error {
		start := time.Now()
		logger.Infow("custom middleware: request started", "path", ctx.Path(), "method", ctx.Method())

		err := ctx.Next()

		duration := time.Since(start)
		logger.Infow("custom middleware: request completed", "duration_ms", duration.Milliseconds())

		return err
	}
}

// rateLimitMiddleware demonstrates simple rate limiting middleware
func rateLimitMiddleware() aurelion.Middleware {
	// Simple in-memory rate limiter (not production-ready)
	rateLimiter := make(map[string]time.Time)
	return func(ctx aurelion.Context) error {
		clientIP := ctx.IP()

		// Check if client exceeded rate limit (1 request per 2 seconds)
		if lastRequest, exists := rateLimiter[clientIP]; exists {
			if time.Since(lastRequest) < 2*time.Second {
				return aurelion.Error(ctx, aurelion.NewErrorf(429, "Rate limit exceeded. Please wait %d seconds", 2-int(time.Since(lastRequest).Seconds())))
			}
		}

		// Update rate limiter
		rateLimiter[clientIP] = time.Now()

		return ctx.Next()
	}
}

// metricsMiddleware demonstrates metrics collection middleware
func metricsMiddleware() aurelion.Middleware {
	return func(ctx aurelion.Context) error {
		// Collect metrics before processing
		ctx.Locals("metrics_start", time.Now())

		err := ctx.Next()

		// Log metrics after processing
		if startTime, ok := ctx.Locals("metrics_start").(time.Time); ok {
			duration := time.Since(startTime)
			logger.Infow("metrics collected", "path", ctx.Path(), "status", ctx.Get("X-Response-Code"), "duration_ms", duration.Milliseconds())
		}

		return err
	}
}
