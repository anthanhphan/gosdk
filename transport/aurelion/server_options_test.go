package aurelion

import (
	"errors"
	"testing"
)

func TestWithGlobalMiddleware(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	middleware1 := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	middleware2 := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithGlobalMiddleware(middleware1, middleware2))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if len(server.globalMiddlewares) != 2 {
		t.Errorf("Expected 2 global middlewares, got %d", len(server.globalMiddlewares))
	}
}

func TestWithPanicRecover(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	panicRecover := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithPanicRecover(panicRecover))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.panicRecover == nil {
		t.Error("Panic recover middleware should be set")
	}
}

func TestWithAuthentication(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authMiddleware := Middleware(func(ctx Context) error {
		return ctx.Next()
	})

	server, err := NewHttpServer(config, WithAuthentication(authMiddleware))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.authMiddleware == nil {
		t.Error("Authentication middleware should be set")
	}
}

func TestWithAuthorization(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	authzChecker := AuthorizationFunc(func(ctx Context, perms []string) error {
		return nil
	})

	server, err := NewHttpServer(config, WithAuthorization(authzChecker))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.authzChecker == nil {
		t.Error("Authorization checker should be set")
	}
}

func TestWithAllOptions(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	middleware := Middleware(func(ctx Context) error { return ctx.Next() })
	panicRecover := Middleware(func(ctx Context) error { return ctx.Next() })
	authMiddleware := Middleware(func(ctx Context) error { return ctx.Next() })
	authzChecker := AuthorizationFunc(func(ctx Context, perms []string) error { return nil })

	server, err := NewHttpServer(config,
		WithGlobalMiddleware(middleware),
		WithPanicRecover(panicRecover),
		WithAuthentication(authMiddleware),
		WithAuthorization(authzChecker),
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.panicRecover == nil {
		t.Error("Panic recover should be set")
	}
	if server.authMiddleware == nil {
		t.Error("Auth middleware should be set")
	}
	if server.authzChecker == nil {
		t.Error("Authz checker should be set")
	}
}

func TestWithGlobalMiddleware_NoMiddleware(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config, WithGlobalMiddleware())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(server.globalMiddlewares) != 0 {
		t.Error("Should have no global middlewares")
	}
}

func TestServerOption_Error(t *testing.T) {
	// Test server option that returns error
	errorOption := func(s *HttpServer) error {
		return errors.New("option error")
	}

	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
	}

	server, err := NewHttpServer(config, errorOption)

	if err == nil {
		t.Error("Expected error from server option")
	}
	if server != nil {
		t.Error("Server should be nil on error")
	}
}
