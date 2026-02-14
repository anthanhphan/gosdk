// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
)

// Orianna Complete Example - Full-featured API server
//
// This example demonstrates major Orianna features including:
// - Server configuration (timeouts, limits, CORS)
// - Routing (HTTP methods, parameters, groups, protected routes)
// - Security (authentication, authorization)
// - Request/response handling with validation
// - Middleware composition
// - Observability (logging, metrics)

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

func createServer() *orianna.Server {
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
		EnableCORS:               true,
		CORS: &configuration.CORSConfig{
			AllowOrigins: []string{"http://localhost:3000", "https://app.example.com"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
			MaxAge:       3600,
		},
	}

	srv, err := orianna.NewServer(config,
		orianna.WithHooks(createHooks()),
		orianna.WithMiddlewareConfig(&configuration.MiddlewareConfig{
			DisableLogging: false,
		}),

		orianna.WithAuthentication(authMiddleware),
		orianna.WithAuthorization(authzChecker),
		orianna.WithPanicRecover(orianna.Recover(func(ctx orianna.Context) error {
			return ctx.Next()
		})),
		orianna.WithMetrics(metrics.NewClient("complete-api")),

		// Health checks
		orianna.WithHealthChecker(
			orianna.NewCustomChecker("database", databaseHealthCheck),
		),
		orianna.WithHealthChecker(
			orianna.NewCustomChecker("cache", cacheHealthCheck),
		),
	)
	if err != nil {
		logger.Fatalw("Failed to create server", "error", err)
	}

	return srv
}
func createHooks() *orianna.RequestHooks {
	hooks := orianna.NewRequestHooks()

	hooks.AddOnRequest(func(ctx orianna.Context) {
		logger.Debugw("REQUEST",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"ip", ctx.IP(),
			"user_agent", ctx.Get("User-Agent"),
		)
	})

	hooks.AddOnResponse(func(ctx orianna.Context, status int, latency time.Duration) {
		logger.Infow("RESPONSE",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"status", status,
			"latency_ms", latency.Milliseconds(),
		)
	})

	hooks.AddOnError(func(ctx orianna.Context, err error) {
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

func databaseHealthCheck(_ context.Context) orianna.HealthCheck {
	return orianna.HealthCheck{
		Name:    "database",
		Status:  orianna.HealthStatusHealthy,
		Message: "Connected to PostgreSQL",
	}
}

func cacheHealthCheck(_ context.Context) orianna.HealthCheck {
	return orianna.HealthCheck{
		Name:    "cache",
		Status:  orianna.HealthStatusHealthy,
		Message: "Redis connection OK",
	}
}

// Route Registration

func registerAllRoutes(srv *orianna.Server) error {
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
	apiV1 := orianna.NewGroupRoute("/api/v1").
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
		orianna.Chain(timingMiddleware, headerMiddleware)); err != nil {
		return err
	}
	if err := srv.POST("/demo/audit", demoHandler,
		orianna.OnlyForMethods(auditMiddleware, "POST", "PUT")); err != nil {
		return err
	}
	if err := srv.GET("/demo/before-after", demoHandler,
		orianna.Before(func(ctx orianna.Context) error { return ctx.Next() }, func(ctx orianna.Context) { _ = beforeMiddleware(ctx) }),
		orianna.After(func(ctx orianna.Context) error { return ctx.Next() }, func(ctx orianna.Context, err error) { _ = afterMiddleware(ctx) })); err != nil {
		return err
	}
	if err := srv.GET("/demo/timeout", slowHandler,
		orianna.Timeout(func(ctx orianna.Context) error { return ctx.Next() }, 2*time.Second)); err != nil {
		return err
	}
	errorGroup := orianna.NewGroupRoute("/errors").
		GET("/business", errorBusinessHandler).
		GET("/validation", errorValidationHandler).
		GET("/wrapped", errorWrappedHandler).
		GET("/panic", errorPanicHandler).
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

func homeHandler(ctx orianna.Context) error {
	return ctx.OK(orianna.Map{
		"service":     "complete-api",
		"version":     "1.0.0",
		"description": "Full-featured Orianna example",
		"docs":        "https://github.com/anthanhphan/gosdk",
	})
}

func healthHandler(ctx orianna.Context) error {
	return ctx.OK(orianna.Map{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    "operational",
	})
}

// Handlers: User CRUD
func listUsersHandler(ctx orianna.Context) error {
	page := orianna.GetQueryInt(ctx, "page", 1)
	limit := orianna.GetQueryInt(ctx, "limit", 10)
	active := orianna.GetQueryBool(ctx, "active", true)
	sortBy := orianna.GetQueryString(ctx, "sort", "created_at")

	users := []orianna.Map{
		{"id": 1, "name": "Alice", "email": "alice@example.com", "active": true},
		{"id": 2, "name": "Bob", "email": "bob@example.com", "active": true},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com", "active": false},
	}

	return ctx.OK(orianna.Map{
		"users": users,
		"pagination": orianna.Map{
			"page":  page,
			"limit": limit,
			"total": len(users),
		},
		"filters": orianna.Map{
			"active": active,
			"sort":   sortBy,
		},
	})
}
func getUserHandler(ctx orianna.Context) error {
	id, err := orianna.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID: must be an integer")
	}

	return ctx.OK(orianna.Map{
		"id":    id,
		"name":  "Alice",
		"email": "alice@example.com",
		"role":  "admin",
	})
}

var createUserHandler = orianna.TypedHandler(
	func(_ orianna.Context, req CreateUserRequest) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:        123,
			Name:      req.Name,
			Email:     req.Email,
			Role:      req.Role,
			CreatedAt: time.Now(),
		}, nil
	},
)

func updateUserHandler(ctx orianna.Context) error {
	id, err := orianna.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID")
	}

	req, ok := orianna.MustBind[UpdateUserRequest](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(orianna.Map{
		"id":      id,
		"name":    req.Name,
		"updated": true,
	})
}
func patchUserStatusHandler(ctx orianna.Context) error {
	id, err := orianna.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID")
	}

	req, ok := orianna.MustBind[PatchStatusRequest](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(orianna.Map{
		"id":     id,
		"status": req.Status,
	})
}
func deleteUserHandler(ctx orianna.Context) error {
	id, err := orianna.GetParamInt(ctx, "id")
	if err != nil {
		return ctx.BadRequestMsg("Invalid user ID")
	}

	log.Printf("Deleted user %d", id)
	return ctx.NoContent()
}

// Handlers: Protected

func profileHandler(ctx orianna.Context) error {
	userID := ctx.Locals("user_id")
	return ctx.OK(orianna.Map{
		"user_id": userID,
		"name":    "Alice",
		"email":   "alice@example.com",
		"bio":     "Software engineer",
	})
}

func updateProfileHandler(ctx orianna.Context) error {
	userID := ctx.Locals("user_id")

	req, ok := orianna.MustBind[UpdateProfileRequest](ctx)
	if !ok {
		return nil
	}

	return ctx.OK(orianna.Map{
		"user_id": userID,
		"bio":     req.Bio,
		"avatar":  req.Avatar,
		"updated": true,
	})
}

// Handlers: Admin

func adminStatsHandler(ctx orianna.Context) error {
	return ctx.OK(orianna.Map{
		"total_users":    1250,
		"active_users":   1000,
		"requests_today": 45000,
		"avg_latency_ms": 23,
	})
}

func adminSettingsHandler(ctx orianna.Context) error {
	req, ok := orianna.MustBind[orianna.Map](ctx, orianna.BindOptions{
		Source:   orianna.BindSourceBody,
		Validate: false,
	})
	if !ok {
		return nil
	}

	return ctx.OK(orianna.Map{
		"settings": req,
		"applied":  true,
	})
}

// Handlers: API Group

func apiStatusHandler(ctx orianna.Context) error {
	return ctx.OK(orianna.Map{
		"status":  "operational",
		"version": "v1",
	})
}

func apiVersionHandler(ctx orianna.Context) error {
	return ctx.OK(orianna.Map{
		"api_version":       "1.0.0",
		"go_version":        "1.23",
		"framework":         "orianna",
		"framework_version": "1.0.0",
	})
}

// Handlers: Demo

func demoHandler(ctx orianna.Context) error {
	return ctx.OK(orianna.Map{
		"message": "Demo endpoint",
		"path":    ctx.Path(),
		"method":  ctx.Method(),
	})
}

func slowHandler(ctx orianna.Context) error {
	time.Sleep(1 * time.Second)
	return ctx.OK(orianna.Map{"status": "completed"})
}

// Handlers: Error Demos

func errorBusinessHandler(_ orianna.Context) error {
	return orianna.NewErrorResponse(
		"INSUFFICIENT_BALANCE",
		orianna.StatusBadRequest,
		"Account balance too low",
	).
		WithDetails("required", 100).
		WithDetails("current", 42).
		WithInternalMsg("User tried to withdraw %d but only has %d", 100, 42)
}

func errorValidationHandler(ctx orianna.Context) error {
	invalid := CreateUserRequest{
		Name:  "Jo",           // Too short (min=3)
		Email: "not-an-email", // Invalid email
		Age:   -1,             // Invalid age
		Role:  "invalid",      // Not in allowed list
	}

	if ok, err := orianna.ValidateAndRespond(ctx, invalid); !ok {
		return err
	}
	return nil
}

func errorWrappedHandler(_ orianna.Context) error {
	baseErr := fmt.Errorf("connection refused")
	return orianna.WrapErrorf(baseErr, "failed to connect to user-service")
}

func errorPanicHandler(_ orianna.Context) error {
	panic("simulated panic for demo purposes")
}

// Authentication & Authorization

func authMiddleware(ctx orianna.Context) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return orianna.NewErrorResponse(
			"UNAUTHORIZED",
			orianna.StatusUnauthorized,
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
		return orianna.NewErrorResponse(
			"UNAUTHORIZED",
			orianna.StatusUnauthorized,
			"Invalid authorization token",
		)
	}

	return ctx.Next()
}

func authzChecker(ctx orianna.Context, permissions []string) error {
	role := ctx.Locals("role")
	if role == nil {
		return fmt.Errorf("role not found in context")
	}

	if role == "admin" {
		return nil
	}

	for _, perm := range permissions {
		if perm == "admin:write" || perm == "admin:read" {
			return orianna.NewErrorResponse(
				"FORBIDDEN",
				orianna.StatusForbidden,
				fmt.Sprintf("Insufficient permissions: requires %v", permissions),
			)
		}
	}

	return nil
}

// Middleware

func timingMiddleware(ctx orianna.Context) error {
	start := time.Now()
	err := ctx.Next()
	duration := time.Since(start)
	logger.Infow("TIMING", "path", ctx.Path(), "duration_ms", duration.Milliseconds())
	return err
}

func headerMiddleware(ctx orianna.Context) error {
	ctx.Set("X-Custom-Header", "orianna-demo")
	ctx.Set("X-Request-Time", time.Now().UTC().Format(time.RFC3339))
	ctx.Set("X-Server-Version", "1.0.0")
	return ctx.Next()
}

func auditMiddleware(ctx orianna.Context) error {
	userID := ctx.Locals("user_id")
	logger.Infow("AUDIT",
		"action", fmt.Sprintf("%s %s", ctx.Method(), ctx.Path()),
		"user_id", userID,
		"ip", ctx.IP(),
	)
	return ctx.Next()
}

func beforeMiddleware(ctx orianna.Context) error {
	logger.Debugw("BEFORE middleware executed")
	ctx.Locals("before_time", time.Now())
	return nil
}

func afterMiddleware(ctx orianna.Context) error {
	if startTime, ok := ctx.Locals("before_time").(time.Time); ok {
		logger.Debugw("AFTER middleware executed",
			"elapsed", time.Since(startTime).Milliseconds())
	}
	return nil
}
