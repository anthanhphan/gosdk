// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"log"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/server"
)

// ===========================================================================
// Orianna Quickstart - Minimal example to get started quickly
// ===========================================================================
//
// This example demonstrates:
// - Basic server setup and configuration
// - Simple GET/POST routes
// - Query parameter handling
// - Request body binding with validation
// - JSON responses and error handling
//
// Run:  go run main.go
// Test: curl http://localhost:8080/
//
// ===========================================================================

func main() {
	// -------------------------------------------------------------------------
	// Step 1: Create server with basic configuration
	// -------------------------------------------------------------------------
	srv, err := server.NewServer(&configuration.Config{
		ServiceName:         "quickstart-api",
		Port:                8080,
		VerboseLogging:      true, // Enable detailed request logging
		UseProperHTTPStatus: true, // Use HTTP status codes in responses
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// -------------------------------------------------------------------------
	// Step 2: Register routes
	// -------------------------------------------------------------------------

	// Simple GET endpoint returning JSON
	if err := srv.GET("/", func(ctx core.Context) error {
		return ctx.OK(map[string]interface{}{
			"message": "Hello, Orianna!",
			"version": "1.0.0",
		})
	}); err != nil {
		log.Fatal(err)
	}

	// GET with query parameters
	// Try: curl "http://localhost:8080/users?page=2&limit=20&active=true"
	if err := srv.GET("/users", func(ctx core.Context) error {
		// Parse query params with defaults using package-level helpers
		page := core.GetQueryInt(ctx, "page", 1)
		limit := core.GetQueryInt(ctx, "limit", 10)
		active := core.GetQueryBool(ctx, "active", true)

		return ctx.OK(map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "Alice", "email": "alice@example.com"},
				{"id": 2, "name": "Bob", "email": "bob@example.com"},
			},
			"pagination": map[string]interface{}{
				"page":   page,
				"limit":  limit,
				"active": active,
			},
		})
	}); err != nil {
		log.Fatal(err)
	}

	// GET with URL parameters
	// Try: curl http://localhost:8080/users/123
	if err := srv.GET("/users/:id", func(ctx core.Context) error {
		// Parse URL param as integer using package-level helper
		id, err := core.GetParamInt(ctx, "id")
		if err != nil {
			return ctx.BadRequestMsg("Invalid user ID: must be a number")
		}

		return ctx.OK(map[string]interface{}{
			"id":    id,
			"name":  "Alice",
			"email": "alice@example.com",
		})
	}); err != nil {
		log.Fatal(err)
	}

	// POST with automatic request binding and validation
	// Try: curl -X POST http://localhost:8080/users \
	//   -H "Content-Type: application/json" \
	//   -d '{"name":"Charlie","email":"charlie@example.com","age":25}'
	if err := srv.POST("/users", func(ctx core.Context) error {
		// Define request structure with validation tags
		type CreateUserRequest struct {
			Name  string `json:"name" validate:"required,min=3,max=50"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"required,min=1,max=150"`
		}

		// Bind and validate request body using package-level helper
		req, err := core.Bind[CreateUserRequest](ctx, core.BindOptions{
			Source:   core.BindSourceBody,
			Validate: true,
		})
		if err != nil {
			// Handle validation errors
			return ctx.BadRequestMsg("Invalid request body: " + err.Error())
		}

		// Success response with created status
		return ctx.Created(map[string]interface{}{
			"id":    123,
			"name":  req.Name,
			"email": req.Email,
			"age":   req.Age,
		})
	}); err != nil {
		log.Fatal(err)
	}

	// PUT endpoint for updates
	// Try: curl -X PUT http://localhost:8080/users/123 \
	//   -H "Content-Type: application/json" \
	//   -d '{"name":"Charlie Updated"}'
	if err := srv.PUT("/users/:id", func(ctx core.Context) error {
		id, err := core.GetParamInt(ctx, "id")
		if err != nil {
			return ctx.BadRequestMsg("Invalid user ID")
		}

		type UpdateUserRequest struct {
			Name string `json:"name" validate:"required,min=3"`
		}

		// Use MustBind - sends error response automatically on failure
		req, ok := core.MustBind[UpdateUserRequest](ctx)
		if !ok {
			return nil // Error already sent to client
		}

		return ctx.OK(map[string]interface{}{
			"id":      id,
			"name":    req.Name,
			"updated": true,
		})
	}); err != nil {
		log.Fatal(err)
	}

	// DELETE endpoint
	// Try: curl -X DELETE http://localhost:8080/users/123
	if err := srv.DELETE("/users/:id", func(ctx core.Context) error {
		id, err := core.GetParamInt(ctx, "id")
		if err != nil {
			return ctx.BadRequestMsg("Invalid user ID")
		}

		// NoContent returns 204 with no body
		log.Printf("Deleted user %d", id)
		return ctx.NoContent()
	}); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------------------
	// Step 3: Start server
	// -------------------------------------------------------------------------
	log.Printf("Server starting on http://localhost:8080")
	log.Println("Try these endpoints:")
	log.Println("  GET    http://localhost:8080/")
	log.Println("  GET    http://localhost:8080/users?page=1")
	log.Println("  GET    http://localhost:8080/users/123")
	log.Println("  POST   http://localhost:8080/users")
	log.Println("  PUT    http://localhost:8080/users/123")
	log.Println("  DELETE http://localhost:8080/users/123")

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
