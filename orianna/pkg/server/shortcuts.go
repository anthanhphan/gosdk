// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/routing"
)

// Server Route Shortcuts

// registerRoute is an internal helper to register routes with a specific HTTP method
func (s *Server) registerRoute(method core.Method, path string, handler core.Handler, middleware ...core.Middleware) error {
	route := routing.NewRoute(path).
		Method(method).
		Handler(handler).
		Middleware(middleware...).
		Build()
	return s.RegisterRoutes(*route)
}

// GET registers a GET route with optional middleware.
// This is a simplified alternative to using NewRoute().GET().Build().
//
// Input:
//   - path: The URL path for the route
//   - handler: The handler function for this route
//   - middleware: Optional middleware functions to apply to this route
//
// Output:
//   - error: Returns an error if route registration fails
//
// Example:
//
//	server.GET("/users", listUsersHandler)
//	server.GET("/admin", adminHandler, authMiddleware, adminMiddleware)
func (s *Server) GET(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.GET, path, handler, middleware...)
}

// POST registers a POST route with optional middleware.
// This is a simplified alternative to using NewRoute().POST().Build().
//
// Input:
//   - path: The URL path for the route
//   - handler: The handler function for this route
//   - middleware: Optional middleware functions to apply to this route
//
// Output:
//   - error: Returns an error if route registration fails
//
// Example:
//
//	server.POST("/users", createUserHandler)
//	server.POST("/login", loginHandler, rateLimitMiddleware)
func (s *Server) POST(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.POST, path, handler, middleware...)
}

// PUT registers a PUT route with optional middleware
// This is a simplified alternative to using NewRoute().PUT().Build()
//
// Example:
//
//	server.PUT("/users/:id", updateUserHandler)
func (s *Server) PUT(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.PUT, path, handler, middleware...)
}

// DELETE registers a DELETE route with optional middleware
// This is a simplified alternative to using NewRoute().DELETE().Build()
//
// Example:
//
//	server.DELETE("/users/:id", deleteUserHandler)
func (s *Server) DELETE(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.DELETE, path, handler, middleware...)
}

// PATCH registers a PATCH route with optional middleware
// This is a simplified alternative to using NewRoute().PATCH().Build()
//
// Example:
//
//	server.PATCH("/users/:id", patchUserHandler)
func (s *Server) PATCH(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.PATCH, path, handler, middleware...)
}

// HEAD registers a HEAD route with optional middleware
// This is a simplified alternative to using NewRoute().HEAD().Build()
//
// Example:
//
//	server.HEAD("/users/:id", checkUserExistsHandler)
func (s *Server) HEAD(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.HEAD, path, handler, middleware...)
}

// OPTIONS registers an OPTIONS route with optional middleware
// This is a simplified alternative to using NewRoute().OPTIONS().Build()
//
// Example:
//
//	server.OPTIONS("/api/users", corsPreflightHandler)
func (s *Server) OPTIONS(path string, handler core.Handler, middleware ...core.Middleware) error {
	return s.registerRoute(core.OPTIONS, path, handler, middleware...)
}

// Protected Route Shortcuts

// RouteShortcuts provides a fluent API for registering protected routes
type RouteShortcuts struct {
	server      *Server
	isProtected bool
	permissions []string
	middleware  []core.Middleware
}

// Protected returns a RouteShortcuts instance for registering protected routes.
// Protected routes require authentication and optionally specific permissions.
//
// Input:
//   - None
//
// Output:
//   - *RouteShortcuts: A builder for protected route registration
//
// Example:
//
//	server.Protected().GET("/profile", profileHandler)
//	server.Protected().WithPermissions("admin").GET("/admin", adminHandler)
func (s *Server) Protected() *RouteShortcuts {
	return &RouteShortcuts{
		server:      s,
		isProtected: true,
	}
}

// WithPermissions sets required permissions for the protected routes
//
// Example:
//
//	server.Protected().WithPermissions("read:users", "write:users").GET("/admin/users", handler)
func (rs *RouteShortcuts) WithPermissions(permissions ...string) *RouteShortcuts {
	rs.permissions = permissions
	return rs
}

// Middleware adds middleware to the protected routes
//
// Example:
//
//	server.Protected().Middleware(auditMiddleware).GET("/admin", handler)
func (rs *RouteShortcuts) Middleware(middleware ...core.Middleware) *RouteShortcuts {
	rs.middleware = append(rs.middleware, middleware...)
	return rs
}

// registerProtectedRoute is an internal helper to register protected routes
func (rs *RouteShortcuts) registerProtectedRoute(method core.Method, path string, handler core.Handler, middleware ...core.Middleware) error {
	// Create new slice to avoid modifying rs.middleware's backing array
	allMiddleware := make([]core.Middleware, 0, len(rs.middleware)+len(middleware))
	allMiddleware = append(allMiddleware, rs.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	builder := routing.NewRoute(path).
		Method(method).
		Handler(handler).
		Middleware(allMiddleware...)

	if rs.isProtected {
		builder.Protected()
	}
	if len(rs.permissions) > 0 {
		builder.Permissions(rs.permissions...)
	}

	route := builder.Build()
	return rs.server.RegisterRoutes(*route)
}

// GET registers a protected GET route
func (rs *RouteShortcuts) GET(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.GET, path, handler, middleware...)
}

// POST registers a protected POST route
func (rs *RouteShortcuts) POST(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.POST, path, handler, middleware...)
}

// PUT registers a protected PUT route
func (rs *RouteShortcuts) PUT(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.PUT, path, handler, middleware...)
}

// DELETE registers a protected DELETE route
func (rs *RouteShortcuts) DELETE(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.DELETE, path, handler, middleware...)
}

// PATCH registers a protected PATCH route
func (rs *RouteShortcuts) PATCH(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.PATCH, path, handler, middleware...)
}

// HEAD registers a protected HEAD route
func (rs *RouteShortcuts) HEAD(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.HEAD, path, handler, middleware...)
}

// OPTIONS registers a protected OPTIONS route
func (rs *RouteShortcuts) OPTIONS(path string, handler core.Handler, middleware ...core.Middleware) error {
	return rs.registerProtectedRoute(core.OPTIONS, path, handler, middleware...)
}

// Server Lifecycle Shortcuts

// Run starts the server and blocks until shutdown.
// This is equivalent to calling Start() but with a clearer name.
//
// Input:
//   - None
//
// Output:
//   - error: Returns an error if the server fails to start
//
// Example:
//
//	if err := server.Run(); err != nil {
//	    log.Fatal(err)
//	}
func (s *Server) Run() error {
	return s.Start()
}
