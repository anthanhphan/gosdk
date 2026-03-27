// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/middleware"
	"github.com/anthanhphan/gosdk/orianna/http/routing"
	"github.com/anthanhphan/gosdk/orianna/http/server"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
)

// Orianna Complete Example - Full-featured API server
//
// This example demonstrates major Orianna features including:
// - Server configuration (timeouts, limits, CORS)
// - Routing (HTTP methods, parameters, groups, protected routes)
// - Security (authentication, authorization, request ID validation)
// - Request/response handling with validation
// - Middleware composition
// - Observability (logging, metrics, slow request detection)
// - Audit trail (InternalMessage & Cause logging in error responses)

const (
	validToken = "Bearer secret-token-12345"
	adminToken = "Bearer admin-token-67890"
)

func main() {
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelDebug,
		LogEncoding:       logger.EncodingJSON,
		DisableStacktrace: true,
	})
	defer undo()

	srv := createServer()

	if err := registerAllRoutes(srv); err != nil {
		logger.Fatalw("Failed to register routes", "error", err)
	}

	if err := srv.Run(); err != nil {
		logger.Fatalw("Server error", "error", err)
	}
}

// Server Configuration

func createServer() *server.Server {
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 30 * time.Second
	shutdownTimeout := 5 * time.Second
	config := &configuration.Config{
		ServiceName:              "complete-api",
		Version:                  "1.0.0",
		Port:                     8081,
		VerboseLogging:           true,
		VerboseLoggingSkipPaths:  []string{"/health", "/metrics"},
		UseProperHTTPStatus:      true,
		MaxBodySize:              4 * 1024 * 1024, // 4MB
		MaxConcurrentConnections: 256 * 1024,
		ReadTimeout:              &readTimeout,
		WriteTimeout:             &writeTimeout,
		IdleTimeout:              &idleTimeout,
		GracefulShutdownTimeout:  &shutdownTimeout,
		SlowRequestThreshold:     2 * time.Second, // Non-zero → auto-registers SlowRequestDetector
		EnableCORS:               true,
		CORS: &configuration.CORSConfig{
			AllowOrigins: []string{"http://localhost:3000", "https://app.example.com"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
			MaxAge:       3600,
		},
	}

	srv, err := server.NewServer(config,
		server.WithHooks(createHooks()),
		server.WithMiddlewareConfig(&configuration.MiddlewareConfig{
			DisableLogging: false,
		}),
		server.WithAuthentication(authMiddleware),
		server.WithAuthorization(authzChecker),
		server.WithMetrics(metrics.NewClient("complete-api")),

		// Health checks
		server.WithHealthChecker(
			health.NewCustomChecker("database", databaseHealthCheck),
		),
		server.WithHealthChecker(
			health.NewCustomChecker("cache", cacheHealthCheck),
		),
	)
	if err != nil {
		logger.Fatalw("Failed to create server", "error", err)
	}

	return srv
}

func createHooks() *core.Hooks {
	hooks := core.NewHooks()

	hooks.AddOnRequest(func(ctx core.Context) {
		logger.Debugw("REQUEST",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"ip", ctx.IP(),
			"user_agent", ctx.Get("User-Agent"),
		)
	})

	hooks.AddOnResponse(func(ctx core.Context, status int, latency time.Duration) {
		logger.Infow("RESPONSE",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"status", status,
			"latency_ms", latency.Milliseconds(),
		)
	})

	hooks.AddOnError(func(ctx core.Context, err error) {
		logger.Errorw("ERROR",
			"path", ctx.Path(),
			"error", err.Error(),
		)
	})

	hooks.AddOnShutdown(func() {
		logger.Infow("Server shutting down...")
	})

	return hooks
}

func databaseHealthCheck(_ context.Context) health.HealthCheck {
	return health.HealthCheck{
		Name:    "database",
		Status:  health.StatusHealthy,
		Message: "Connected to PostgreSQL",
	}
}

func cacheHealthCheck(_ context.Context) health.HealthCheck {
	return health.HealthCheck{
		Name:    "cache",
		Status:  health.StatusHealthy,
		Message: "Redis connection OK",
	}
}

// Route Registration

func registerAllRoutes(srv *server.Server) error {
	if err := srv.GET("/", homeHandler); err != nil {
		return err
	}
	if err := srv.GET("/health", healthHandler); err != nil {
		return err
	}
	if err := srv.GET("/users", listUsersHandler); err != nil {
		return err
	}
	if err := srv.GET("/users/:id", getUserHandler); err != nil {
		return err
	}
	if err := srv.POST("/users", createUserHandler); err != nil {
		return err
	}
	if err := srv.PUT("/users/:id", updateUserHandler); err != nil {
		return err
	}
	if err := srv.PATCH("/users/:id/status", patchUserStatusHandler); err != nil {
		return err
	}
	if err := srv.DELETE("/users/:id", deleteUserHandler); err != nil {
		return err
	}
	if err := srv.Protected().GET("/profile", profileHandler); err != nil {
		return err
	}
	if err := srv.Protected().PUT("/profile", updateProfileHandler); err != nil {
		return err
	}
	if err := srv.Protected().
		WithPermissions("admin:read").
		GET("/admin/stats", adminStatsHandler); err != nil {
		return err
	}

	if err := srv.Protected().
		WithPermissions("admin:write").
		Middleware(auditMiddleware).
		POST("/admin/settings", adminSettingsHandler); err != nil {
		return err
	}

	apiV1 := routing.NewGroupRoute("/api/v1").
		GET("/status", apiStatusHandler).
		GET("/version", apiVersionHandler).
		Build()
	if err := srv.RegisterGroup(*apiV1); err != nil {
		return err
	}
	if err := srv.GET("/demo/timing", demoHandler,
		timingMiddleware); err != nil {
		return err
	}
	if err := srv.GET("/demo/chain", demoHandler,
		middleware.Chain(timingMiddleware, headerMiddleware)); err != nil {
		return err
	}
	if err := srv.POST("/demo/audit", demoHandler,
		middleware.OnlyForMethods(auditMiddleware, "POST", "PUT")); err != nil {
		return err
	}
	if err := srv.GET("/demo/before-after", demoHandler,
		middleware.Before(func(ctx core.Context) error { return ctx.Next() }, func(ctx core.Context) { _ = beforeMiddleware(ctx) }),
		middleware.After(func(ctx core.Context) error { return ctx.Next() }, func(ctx core.Context, err error) { _ = afterMiddleware(ctx) })); err != nil {
		return err
	}
	if err := srv.GET("/demo/timeout", slowHandler,
		middleware.Timeout(func(ctx core.Context) error { return ctx.Next() }, 2*time.Second)); err != nil {
		return err
	}

	// Slow request demo — will trigger SlowRequestDetector warning (threshold: 2s)
	// Try: curl http://localhost:8081/demo/slow
	if err := srv.GET("/demo/slow", func(ctx core.Context) error {
		time.Sleep(3 * time.Second) // Exceeds 2s threshold
		return ctx.OK(map[string]interface{}{
			"message":  "This request was intentionally slow",
			"duration": "3s",
			"note":     "Check server logs for 'slow request detected' warning",
		})
	}); err != nil {
		return err
	}

	// Per-route slow detection with custom threshold (e.g., 500ms for critical paths)
	// Try: curl http://localhost:8081/demo/slow-custom
	if err := srv.GET("/demo/slow-custom", func(ctx core.Context) error {
		time.Sleep(600 * time.Millisecond) // Exceeds 500ms custom threshold
		return ctx.OK(map[string]interface{}{
			"message": "This route has a stricter 500ms threshold",
		})
	}, middleware.SlowRequestDetector(500*time.Millisecond)); err != nil {
		return err
	}

	errorGroup := routing.NewGroupRoute("/errors").
		GET("/business", errorBusinessHandler).
		GET("/validation", errorValidationHandler).
		GET("/wrapped", errorWrappedHandler).
		GET("/panic", errorPanicHandler).
		GET("/audit-demo", errorAuditDemoHandler). // NEW: demonstrates audit logging
		Build()
	if err := srv.RegisterGroup(*errorGroup); err != nil {
		return err
	}

	return nil
}

// Request/Response Types

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=1,max=150"`
	Role  string `json:"role" validate:"oneof=user admin moderator"`
}

type CreateUserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=3"`
}

type PatchStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
}

type UpdateProfileRequest struct {
	Bio    string `json:"bio" validate:"max=500"`
	Avatar string `json:"avatar" validate:"url"`
}

// Handlers: Public

func homeHandler(ctx core.Context) error {
	return ctx.OK(map[string]interface{}{
		"service":     "complete-api",
		"version":     "1.0.0",
		"description": "Full-featured Orianna example",
		"docs":        "https://github.com/anthanhphan/gosdk",
	})
}

func healthHandler(ctx core.Context) error {
	return ctx.OK(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    "operational",
	})
}

// Handlers: User CRUD
func listUsersHandler(ctx core.Context) error {
	page := core.GetQueryInt(ctx, "page", 1)
	limit := core.GetQueryInt(ctx, "limit", 10)
	active := core.GetQueryBool(ctx, "active", true)
	sortBy := core.GetQueryString(ctx, "sort", "created_at")

	users := []map[string]interface{}{
		{"id": 1, "name": "Alice", "email": "alice@example.com", "active": true},
		{"id": 2, "name": "Bob", "email": "bob@example.com", "active": true},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com", "active": false},
	}

	return ctx.OK(map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": len(users),
		},
		"filters": map[string]interface{}{
			"active": active,
			"sort":   sortBy,
		},
	})
}

func getUserHandler(ctx core.Context) error {
	id, err := core.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID: must be an integer")
	}

	return ctx.OK(map[string]interface{}{
		"id":    id,
		"name":  "Alice",
		"email": "alice@example.com",
		"role":  "admin",
	})
}

var createUserHandler = core.TypedHandler(core.StatusCreated,
	func(_ core.Context, req CreateUserRequest) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:        123,
			Name:      req.Name,
			Email:     req.Email,
			Role:      req.Role,
			CreatedAt: time.Now(),
		}, nil
	},
)

func updateUserHandler(ctx core.Context) error {
	id, err := core.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID")
	}

	req, ok := core.MustBind[UpdateUserRequest](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(map[string]interface{}{
		"id":      id,
		"name":    req.Name,
		"updated": true,
	})
}

func patchUserStatusHandler(ctx core.Context) error {
	id, err := core.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID")
	}

	req, ok := core.MustBind[PatchStatusRequest](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(map[string]interface{}{
		"id":     id,
		"status": req.Status,
	})
}

func deleteUserHandler(ctx core.Context) error {
	id, err := core.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID")
	}

	log.Printf("Deleted user %d", id)
	return ctx.NoContent()
}

// Handlers: Protected

func profileHandler(ctx core.Context) error {
	userID := ctx.Locals("user_id")
	return ctx.OK(map[string]interface{}{
		"user_id": userID,
		"name":    "Alice",
		"email":   "alice@example.com",
		"bio":     "Software engineer",
	})
}

func updateProfileHandler(ctx core.Context) error {
	userID := ctx.Locals("user_id")

	req, ok := core.MustBind[UpdateProfileRequest](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(map[string]interface{}{
		"user_id": userID,
		"bio":     req.Bio,
		"avatar":  req.Avatar,
		"updated": true,
	})
}

// Handlers: Admin

func adminStatsHandler(ctx core.Context) error {
	return ctx.OK(map[string]interface{}{
		"total_users":    1250,
		"active_users":   1000,
		"requests_today": 45000,
		"avg_latency_ms": 23,
	})
}

func adminSettingsHandler(ctx core.Context) error {
	req, ok := core.MustBind[map[string]interface{}](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(map[string]interface{}{
		"settings": req,
		"applied":  true,
	})
}

// Handlers: API Group

func apiStatusHandler(ctx core.Context) error {
	return ctx.OK(map[string]interface{}{
		"status":  "operational",
		"version": "v1",
	})
}

func apiVersionHandler(ctx core.Context) error {
	return ctx.OK(map[string]interface{}{
		"api_version":       "1.0.0",
		"go_version":        "1.23",
		"framework":         "orianna",
		"framework_version": "1.0.0",
	})
}

// Handlers: Demo

func demoHandler(ctx core.Context) error {
	return ctx.OK(map[string]interface{}{
		"message": "Demo endpoint",
		"path":    ctx.Path(),
		"method":  ctx.Method(),
	})
}

func slowHandler(ctx core.Context) error {
	time.Sleep(1 * time.Second)
	return ctx.OK(map[string]interface{}{"status": "completed"})
}

// Handlers: Error Demos

func errorBusinessHandler(_ core.Context) error {
	return core.NewErrorResponse(
		"INSUFFICIENT_BALANCE",
		core.StatusBadRequest,
		"Account balance too low",
	).
		WithDetails("required", 100).
		WithDetails("current", 42).
		WithInternalMsg("User tried to withdraw %d but only has %d", 100, 42)
}

func errorValidationHandler(ctx core.Context) error {
	invalid := CreateUserRequest{
		Name:  "Jo",           // Too short (min=3)
		Email: "not-an-email", // Invalid email
		Age:   -1,             // Invalid age
		Role:  "invalid",      // Not in allowed list
	}

	if ok, err := core.ValidateAndRespond(ctx, invalid); !ok {
		return err
	}
	return nil
}

func errorWrappedHandler(_ core.Context) error {
	baseErr := fmt.Errorf("connection refused")
	return core.WrapErrorf(baseErr, "failed to connect to user-service")
}

func errorPanicHandler(_ core.Context) error {
	panic("simulated panic for demo purposes")
}

// errorAuditDemoHandler demonstrates how InternalMessage and Cause are
// automatically logged server-side for audit purposes, while never being
// exposed to the client in the JSON response.
// Try: curl http://localhost:8081/errors/audit-demo
// Then check server logs for "error response" with internal_message and cause.
func errorAuditDemoHandler(_ core.Context) error {
	// Simulate a real-world scenario: database query failed
	dbErr := fmt.Errorf("pq: connection refused on 10.0.1.42:5432")

	return core.NewErrorResponse(
		"DATA_FETCH_FAILED",
		core.StatusInternalServerError,
		"Unable to retrieve data. Please try again later.", // Client-safe message
	).
		WithInternalMsg("Failed to query user_accounts table in primary DB"). // Logged server-side only
		WithCause(dbErr).                                                     // Root cause — logged, never sent to client
		WithDetails("retry_after_seconds", 30)
}

// Authentication & Authorization

func authMiddleware(ctx core.Context) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return core.NewErrorResponse(
			"UNAUTHORIZED",
			core.StatusUnauthorized,
			"Missing authorization token",
		)
	}

	switch token {
	case validToken:
		ctx.Locals("user_id", "user-123")
		ctx.Locals("role", "user")
	case adminToken:
		ctx.Locals("user_id", "admin-456")
		ctx.Locals("role", "admin")
	default:
		return core.NewErrorResponse(
			"UNAUTHORIZED",
			core.StatusUnauthorized,
			"Invalid authorization token",
		)
	}

	return ctx.Next()
}

func authzChecker(ctx core.Context, permissions []string) error {
	role := ctx.Locals("role")
	if role == nil {
		return fmt.Errorf("role not found in context")
	}

	if role == "admin" {
		return nil
	}

	for _, perm := range permissions {
		if perm == "admin:write" || perm == "admin:read" {
			return core.NewErrorResponse(
				"FORBIDDEN",
				core.StatusForbidden,
				fmt.Sprintf("Insufficient permissions: requires %v", permissions),
			)
		}
	}

	return nil
}

// Middleware

func timingMiddleware(ctx core.Context) error {
	start := time.Now()
	err := ctx.Next()
	duration := time.Since(start)
	logger.Infow("TIMING", "path", ctx.Path(), "duration_ms", duration.Milliseconds())
	return err
}

func headerMiddleware(ctx core.Context) error {
	ctx.Set("X-Custom-Header", "orianna-demo")
	ctx.Set("X-Request-Time", time.Now().UTC().Format(time.RFC3339))
	ctx.Set("X-Server-Version", "1.0.0")
	return ctx.Next()
}

func auditMiddleware(ctx core.Context) error {
	userID := ctx.Locals("user_id")
	logger.Infow("AUDIT",
		"action", fmt.Sprintf("%s %s", ctx.Method(), ctx.Path()),
		"user_id", userID,
		"ip", ctx.IP(),
	)
	return ctx.Next()
}

func beforeMiddleware(ctx core.Context) error {
	logger.Debugw("BEFORE middleware executed")
	ctx.Locals("before_time", time.Now())
	return nil
}

func afterMiddleware(ctx core.Context) error {
	if startTime, ok := ctx.Locals("before_time").(time.Time); ok {
		logger.Debugw("AFTER middleware executed",
			"elapsed", time.Since(startTime).Milliseconds())
	}
	return nil
}
