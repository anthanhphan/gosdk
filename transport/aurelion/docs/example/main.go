package main

import (
	"net/http"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/transport/aurelion"
)

func main() {
	// Initialize logger
	undo := logger.InitDefaultLogger()
	defer undo()

	// Create server configuration
	config := &aurelion.Config{
		ServiceName: "Example API",
		Port:        8080,
	}

	// Create server
	server, err := aurelion.NewHttpServer(config)
	if err != nil {
		logger.NewLoggerWithFields().Fatalw("failed to create server", "error", err)
	}

	// Add routes
	server.AddRoutes(
		// Health check endpoint
		aurelion.NewRoute("/health").
			GET().
			Handler(func(ctx aurelion.Context) error {
				return aurelion.HealthCheck(ctx)
			}),

		// Simple GET endpoint
		aurelion.NewRoute("/users").
			GET().
			Handler(func(ctx aurelion.Context) error {
				users := []map[string]interface{}{
					{"id": 1, "name": "John Doe", "email": "john@example.com"},
					{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
				}
				return aurelion.OK(ctx, "Users retrieved", users)
			}),

		// GET with route parameter
		aurelion.NewRoute("/users/:id").
			GET().
			Handler(func(ctx aurelion.Context) error {
				userID := ctx.Params("id")
				if userID == "999" {
					return aurelion.Error(ctx, aurelion.NewError(1001, "User not found"))
				}
				user := map[string]interface{}{
					"id":    userID,
					"name":  "John Doe",
					"email": "john@example.com",
				}
				return aurelion.OK(ctx, "User details", user)
			}),

		// POST with body parsing
		aurelion.NewRoute("/users").
			POST().
			Handler(func(ctx aurelion.Context) error {
				var req struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}

				if err := ctx.BodyParser(&req); err != nil {
					return aurelion.BadRequest(ctx, "Invalid request body")
				}

				if req.Name == "" || req.Email == "" {
					return aurelion.BadRequest(ctx, "Name and email are required")
				}

				user := map[string]interface{}{
					"id":    123,
					"name":  req.Name,
					"email": req.Email,
				}
				return ctx.Status(http.StatusCreated).JSON(aurelion.Map{
					"success":   true,
					"code":      http.StatusCreated,
					"message":   "User created",
					"data":      user,
					"timestamp": time.Now().UnixMilli(),
				})
			}),

		// GET with query parameters
		aurelion.NewRoute("/search").
			GET().
			Handler(func(ctx aurelion.Context) error {
				query := ctx.Query("q", "")
				page := ctx.Query("page", "1")

				if query == "" {
					return aurelion.BadRequest(ctx, "Query parameter 'q' is required")
				}

				results := map[string]interface{}{
					"query": query,
					"page":  page,
					"items": []string{"result1", "result2", "result3"},
				}
				return aurelion.OK(ctx, "Search results", results)
			}),
	)

	// Add group routes
	server.AddGroupRoutes(
		aurelion.NewGroupRoute("/api/v1").
			Routes(
				aurelion.NewRoute("/status").
					GET().
					Handler(func(ctx aurelion.Context) error {
						return aurelion.OK(ctx, "API v1 status", aurelion.Map{"status": "ok"})
					}),
			),
	)

	// Start server
	logger.Info("Starting server on :8080")
	if err := server.Start(); err != nil {
		logger.NewLoggerWithFields().Fatalw("server error", "error", err)
	}
}
