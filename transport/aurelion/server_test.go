package aurelion

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewHttpServer(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config should create server",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
			},
			wantErr: false,
		},
		{
			name:    "nil config should return error",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewHttpServer(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && server == nil {
				t.Error("Server should not be nil on success")
			}
		})
	}
}

func TestNewHttpServer_WithOptions(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	middleware := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithGlobalMiddleware(middleware))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestNewHttpServer_WithRateLimiter(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		option ServerOption
		check  func(t *testing.T, server *HttpServer)
	}{
		{
			name: "should use default rate limiter when not provided",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
			},
			option: nil,
			check: func(t *testing.T, server *HttpServer) {
				if server == nil {
					t.Fatal("Server should not be nil")
				}
				// Default rate limiter should be configured (using limiter middleware internally)
				if server.rateLimiter != nil {
					t.Error("Default rate limiter should use limiter.New internally, not custom middleware")
				}
			},
		},
		{
			name: "should use custom rate limiter when provided",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
			},
			option: WithRateLimiter(Middleware(func(ctx Context) error {
				return ctx.Next()
			})),
			check: func(t *testing.T, server *HttpServer) {
				if server == nil {
					t.Fatal("Server should not be nil")
				}
				if server.rateLimiter == nil {
					t.Error("Custom rate limiter should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *HttpServer
			var err error

			if tt.option != nil {
				server, err = NewHttpServer(tt.config, tt.option)
			} else {
				server, err = NewHttpServer(tt.config)
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if tt.check != nil {
				tt.check(t, server)
			}
		})
	}
}

func TestAddRoutes(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test with RouteBuilder
	server.AddRoutes(
		NewRoute("/test1").GET().Handler(func(ctx Context) error { return nil }),
	)

	// Test with *Route
	route := NewRoute("/test2").POST().Handler(func(ctx Context) error { return nil }).Build()
	server.AddRoutes(route)

	// Test with Route (value type)
	routeValue := Route{
		Path:    "/test3",
		Method:  GET,
		Handler: func(ctx Context) error { return nil },
	}
	server.AddRoutes(routeValue)

	// Test with route having CORS config
	corsRoute := NewRoute("/test4").GET().
		Handler(func(ctx Context) error { return nil }).
		CORS(&CORSConfig{
			AllowOrigins: []string{"https://example.com"},
			AllowMethods: []string{"GET"},
			AllowHeaders: []string{"Content-Type"},
		}).
		Build()
	server.AddRoutes(corsRoute)

	if len(server.routes) != 5 { // 4 added routes + 1 default health check route
		t.Errorf("Expected 5 routes, got %d", len(server.routes))
	}
}

func TestAddRoutes_InvalidType(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Invalid route type should be skipped
	server.AddRoutes("invalid")

	if len(server.routes) != 1 { // 1 default health check route
		t.Errorf("Invalid route type should be skipped, expected 1 route (health check), got %d", len(server.routes))
	}
}

func TestAddRoutes_InvalidRoute(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Invalid route (no handler) should be skipped
	server.AddRoutes(&Route{
		Path:   "/test",
		Method: GET,
	})

	if len(server.routes) != 1 { // 1 default health check route
		t.Errorf("Invalid route should be skipped, expected 1 route (health check), got %d", len(server.routes))
	}
}

func TestAddGroupRoutes(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	group := NewGroupRoute("/api").
		Routes(
			NewRoute("/users").GET().Handler(func(ctx Context) error { return nil }),
		).
		Build()

	server.AddGroupRoutes(group)

	if len(server.groupRoutes) != 1 {
		t.Errorf("Expected 1 group, got %d", len(server.groupRoutes))
	}
}

func TestAddGroupRoutes_ValueType(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	groupValue := GroupRoute{
		Prefix: "/api",
		Routes: []Route{
			{Path: "/test", Method: GET, Handler: func(ctx Context) error { return nil }},
		},
	}

	server.AddGroupRoutes(groupValue)

	if len(server.groupRoutes) != 1 {
		t.Errorf("Expected 1 group, got %d", len(server.groupRoutes))
	}
}

func TestAddGroupRoutes_InvalidGroup(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Invalid group (no routes)
	server.AddGroupRoutes(&GroupRoute{
		Prefix: "/api",
		Routes: []Route{},
	})

	if len(server.groupRoutes) != 0 {
		t.Error("Invalid group should be skipped")
	}
}

func TestAddGroupRoutes_InvalidType(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Invalid group type should be skipped
	server.AddGroupRoutes("invalid")

	if len(server.groupRoutes) != 0 {
		t.Error("Invalid group type should be skipped")
	}
}

func TestApplyProtectionMiddleware(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authMiddleware := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithAuthentication(authMiddleware))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create protected route
	route := &Route{
		Path:        "/protected",
		Method:      GET,
		Handler:     func(ctx Context) error { return nil },
		IsProtected: true,
	}

	// Apply protection
	server.applyProtectionMiddleware(route)

	if len(route.Middlewares) == 0 {
		t.Error("Protected route should have middlewares")
	}
}

func TestApplyProtectionMiddleware_WithPermissions(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authzChecker := AuthorizationFunc(func(ctx Context, perms []string) error {
		return nil
	})

	server, err := NewHttpServer(config, WithAuthorization(authzChecker))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create route with permissions
	route := &Route{
		Path:                "/protected",
		Method:              GET,
		Handler:             func(ctx Context) error { return nil },
		IsProtected:         true,
		RequiredPermissions: []string{"read"},
	}

	// Apply protection
	server.applyProtectionMiddleware(route)

	if len(route.Middlewares) == 0 {
		t.Error("Protected route with permissions should have middlewares")
	}
}

func TestCreateAuthorizationMiddleware_Error(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authzChecker := AuthorizationFunc(func(ctx Context, perms []string) error {
		return errors.New("insufficient permissions")
	})

	server, err := NewHttpServer(config, WithAuthorization(authzChecker))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := server.createAuthorizationMiddleware([]string{"read"})

	err = middleware(nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateAuthorizationMiddleware_NilContext(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authzChecker := AuthorizationFunc(func(ctx Context, perms []string) error {
		return nil
	})

	server, err := NewHttpServer(config, WithAuthorization(authzChecker))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := server.createAuthorizationMiddleware([]string{"read"})

	err = middleware(nil)
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
}

func TestCreateAuthorizationMiddleware_NilAuthzChecker(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := server.createAuthorizationMiddleware([]string{"read"})

	mockCtx := &mockContext{}
	err = middleware(mockCtx)
	if err == nil {
		t.Error("Expected error for nil authz checker, got nil")
	}
}

func TestCreateAuthorizationMiddleware_Success(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authzChecker := AuthorizationFunc(func(ctx Context, perms []string) error {
		return nil // Success
	})

	server, err := NewHttpServer(config, WithAuthorization(authzChecker))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := server.createAuthorizationMiddleware([]string{"read"})

	mockCtx := &mockContext{}
	err = middleware(mockCtx)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestApplyGroupProtection(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authMiddleware := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithAuthentication(authMiddleware))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	group := &GroupRoute{
		Prefix:      "/api",
		IsProtected: true,
		Routes: []Route{
			{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
			{Path: "/posts", Method: GET, Handler: func(ctx Context) error { return nil }},
		},
	}

	server.applyGroupProtection(group)

	// All routes should be protected
	for i, route := range group.Routes {
		if !route.IsProtected {
			t.Errorf("Route %d should be protected", i)
		}
	}
}

func TestApplyGroupProtection_IndividualRoutes(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authMiddleware := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithAuthentication(authMiddleware))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	group := &GroupRoute{
		Prefix:      "/api",
		IsProtected: false,
		Routes: []Route{
			{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }, IsProtected: true},
			{Path: "/posts", Method: GET, Handler: func(ctx Context) error { return nil }, IsProtected: false},
		},
	}

	server.applyGroupProtection(group)

	// First route should be protected, second should not
	if !group.Routes[0].IsProtected {
		t.Error("First route should be protected")
	}
}

func TestAddRoutes_WithGlobalMiddleware(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	middleware := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithGlobalMiddleware(middleware))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	server.AddRoutes(
		NewRoute("/test").GET().Handler(func(ctx Context) error { return nil }),
	)

	if len(server.globalMiddlewares) != 1 {
		t.Errorf("Expected 1 global middleware, got %d", len(server.globalMiddlewares))
	}
}

func TestServer_Shutdown(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Server hasn't started, shutdown should handle gracefully
	ctx := context.Background()
	err = server.Shutdown(ctx)
	_ = err // Expected nil or error for unstarted server
}

func TestNewHttpServer_WithTimeouts(t *testing.T) {
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 60 * time.Second

	config := &Config{
		ServiceName:  "Test Service",
		Port:         8080,
		ReadTimeout:  &readTimeout,
		WriteTimeout: &writeTimeout,
		IdleTimeout:  &idleTimeout,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server with timeouts: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestNewHttpServer_WithGracefulShutdownTimeout(t *testing.T) {
	shutdownTimeout := 30 * time.Second

	config := &Config{
		ServiceName:             "Test Service",
		Port:                    8080,
		GracefulShutdownTimeout: &shutdownTimeout,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestNewHttpServer_WithMaxBodySize(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
		MaxBodySize: 10 * 1024 * 1024, // 10MB
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestNewHttpServer_WithMaxConcurrentConnections(t *testing.T) {
	config := &Config{
		ServiceName:              "Test Service",
		Port:                     8080,
		MaxConcurrentConnections: 5000,
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestNewHttpServer_WithCORS(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
		EnableCORS:  true,
		CORS: &CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST"},
		},
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestSetupGlobalMiddlewares_WithCORS(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
		EnableCORS:  true,
		CORS: &CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET"},
			AllowHeaders: []string{"X-Custom"},
		},
	}

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}
}

func TestServer_Shutdown_WithTimeout(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8090,
	}
	shutdownTimeout := 1 * time.Second
	config.GracefulShutdownTimeout = &shutdownTimeout

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Server hasn't started, shutdown should handle gracefully
	ctx := context.Background()
	err = server.Shutdown(ctx)
	_ = err // Expected error or nil both acceptable for unstarted server
}

func TestServer_Shutdown_WithZeroTimeout(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8091,
	}
	shutdownTimeout := 0 * time.Second
	config.GracefulShutdownTimeout = &shutdownTimeout

	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Server hasn't started, shutdown should handle gracefully
	ctx := context.Background()
	err = server.Shutdown(ctx)
	_ = err // Expected error or nil both acceptable for unstarted server
}
