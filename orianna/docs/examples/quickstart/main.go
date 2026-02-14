// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"log"

	"github.com/anthanhphan/gosdk/orianna"
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
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
	server, err := orianna.NewServer(&configuration.Config{
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
	if err := server.GET("/", func(ctx orianna.Context) error {
		return ctx.OK(orianna.Map{
			"message": "Hello, Orianna!",
			"version": "1.0.0",
		})
	}); err != nil {
		log.Fatal(err)
	}

	// GET with query parameters
	// Try: curl "http://localhost:8080/users?page=2&limit=20&active=true"
	if err := server.GET("/users", func(ctx orianna.Context) error {
		// Parse query params with defaults
		page := orianna.GetQueryInt(ctx, "page", 1)
		limit := orianna.GetQueryInt(ctx, "limit", 10)
		active := orianna.GetQueryBool(ctx, "active", true)

		return ctx.OK(orianna.Map{
			"users": []orianna.Map{
				{"id": 1, "name": "Alice", "email": "alice@example.com"},
				{"id": 2, "name": "Bob", "email": "bob@example.com"},
			},
			"pagination": orianna.Map{
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
	if err := server.GET("/users/:id", func(ctx orianna.Context) error {
		// Parse URL param as integer
		id, err := orianna.GetParamInt(ctx, "id")
		if err != nil {
			return ctx.BadRequestMsg("Invalid user ID: must be a number")
		}

		return ctx.OK(orianna.Map{
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
	if err := server.POST("/users", func(ctx orianna.Context) error {
		// Define request structure with validation tags
		type CreateUserRequest struct {
			Name  string `json:"name" validate:"required,min=3,max=50"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"required,min=1,max=150"`
		}

		// Bind and validate request body
		req, err := orianna.BindBody[CreateUserRequest](ctx, true)
		if err != nil {
			// Handle validation errors
			if valErrs, ok := err.(orianna.ValidationErrors); ok {
				return ctx.JSON(orianna.NewErrorResponse(
					"VALIDATION_FAILED",
					orianna.StatusBadRequest,
					"Request validation failed",
				).WithDetails("errors", valErrs.ToArray()))
			}
			// Handle other binding errors
			return ctx.BadRequestMsg("Invalid request body")
		}

		// Success response with created status
		return ctx.Created(orianna.Map{
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
	if err := server.PUT("/users/:id", func(ctx orianna.Context) error {
		id, err := orianna.GetParamInt(ctx, "id")
		if err != nil {
			return ctx.BadRequestMsg("Invalid user ID")
		}

		type UpdateUserRequest struct {
			Name string `json:"name" validate:"required,min=3"`
		}

		// MustBind automatically sends error response if binding fails
		req, ok := orianna.MustBind[UpdateUserRequest](ctx)
		if !ok {
			return nil // Error already sent to client
		}

		return ctx.OK(orianna.Map{
			"id":      id,
			"name":    req.Name,
			"updated": true,
		})
	}); err != nil {
		log.Fatal(err)
	}

	// DELETE endpoint
	// Try: curl -X DELETE http://localhost:8080/users/123
	if err := server.DELETE("/users/:id", func(ctx orianna.Context) error {
		id, err := orianna.GetParamInt(ctx, "id")
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

	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
