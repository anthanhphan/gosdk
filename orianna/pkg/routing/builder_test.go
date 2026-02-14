// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

func TestRouteBuilder_Path(t *testing.T) {
	builder := NewRoute("/test")
	builder.Path("/updated")

	route := builder.Build()
	if route.Path != "/updated" {
		t.Errorf("Path() = %s, want /updated", route.Path)
	}
}

func TestRouteBuilder_Handler(t *testing.T) {
	builder := NewRoute("/test")
	handler := func(_ core.Context) error { return nil }
	builder.Handler(handler)

	route := builder.Build()
	if route.Handler == nil {
		t.Error("Handler() should set handler")
	}
}

func TestRouteBuilder_Middleware(t *testing.T) {
	builder := NewRoute("/test")
	mw1 := func(_ core.Context) error { return nil }
	mw2 := func(_ core.Context) error { return nil }

	builder.Middleware(mw1, mw2)

	route := builder.Build()
	if len(route.Middlewares) != 2 {
		t.Errorf("Middleware() count = %d, want 2", len(route.Middlewares))
	}
}

func TestRouteBuilder_Permissions(t *testing.T) {
	builder := NewRoute("/test")
	builder.Permissions("read:users", "write:users")

	route := builder.Build()
	if len(route.RequiredPermissions) != 2 {
		t.Errorf("Permissions() count = %d, want 2", len(route.RequiredPermissions))
	}

	if !route.IsProtected {
		t.Error("Permissions() should set IsProtected to true")
	}

	if route.RequiredPermissions[0] != "read:users" {
		t.Errorf("Permission[0] = %s, want read:users", route.RequiredPermissions[0])
	}
}

func TestRouteBuilder_CORS(t *testing.T) {
	builder := NewRoute("/test")
	corsConfig := &configuration.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
	}

	builder.CORS(corsConfig)

	route := builder.Build()
	if route.CORS == nil {
		t.Error("CORS() should set CORS config")
	}

	if len(route.CORS.AllowOrigins) != 1 {
		t.Errorf("CORS AllowOrigins count = %d, want 1", len(route.CORS.AllowOrigins))
	}
}

func TestRouteBuilder_CORS_Nil(t *testing.T) {
	builder := NewRoute("/test")
	builder.CORS(nil)

	route := builder.Build()
	if route.CORS != nil {
		t.Error("CORS(nil) should not set CORS config")
	}
}

func TestRouteBuilder_POST(t *testing.T) {
	builder := NewRoute("/test")
	builder.POST()

	route := builder.Build()
	if route.Method != core.POST {
		t.Errorf("POST() = %v, want POST", route.Method)
	}
}

func TestRouteBuilder_PUT(t *testing.T) {
	builder := NewRoute("/test")
	builder.PUT()

	route := builder.Build()
	if route.Method != core.PUT {
		t.Errorf("PUT() = %v, want PUT", route.Method)
	}
}

func TestRouteBuilder_PATCH(t *testing.T) {
	builder := NewRoute("/test")
	builder.PATCH()

	route := builder.Build()
	if route.Method != core.PATCH {
		t.Errorf("PATCH() = %v, want PATCH", route.Method)
	}
}

func TestRouteBuilder_DELETE(t *testing.T) {
	builder := NewRoute("/test")
	builder.DELETE()

	route := builder.Build()
	if route.Method != core.DELETE {
		t.Errorf("DELETE() = %v, want DELETE", route.Method)
	}
}

func TestRouteBuilder_FluentAPI(t *testing.T) {
	handler := func(_ core.Context) error { return nil }
	mw := func(_ core.Context) error { return nil }

	route := NewRoute("/api/users").
		POST().
		Handler(handler).
		Middleware(mw).
		Protected().
		Permissions("admin").
		Build()

	if route.Path != "/api/users" {
		t.Errorf("Path = %s, want /api/users", route.Path)
	}

	if route.Method != core.POST {
		t.Errorf("Method = %v, want POST", route.Method)
	}

	if !route.IsProtected {
		t.Error("IsProtected should be true")
	}

	if len(route.Middlewares) != 1 {
		t.Errorf("Middlewares count = %d, want 1", len(route.Middlewares))
	}

	if len(route.RequiredPermissions) != 1 {
		t.Errorf("RequiredPermissions count = %d, want 1", len(route.RequiredPermissions))
	}
}

func TestGroupRouteBuilder_Middleware(t *testing.T) {
	builder := NewGroupRoute("/api")
	mw1 := func(_ core.Context) error { return nil }
	mw2 := func(_ core.Context) error { return nil }

	builder.Middleware(mw1, mw2)

	group := builder.Build()
	if len(group.Middlewares) != 2 {
		t.Errorf("Middleware() count = %d, want 2", len(group.Middlewares))
	}
}

func TestGroupRouteBuilder_Protected(t *testing.T) {
	builder := NewGroupRoute("/api")
	builder.Protected()

	group := builder.Build()
	if !group.IsProtected {
		t.Error("Protected() should set IsProtected to true")
	}
}

func TestGroupRouteBuilder_Route(t *testing.T) {
	builder := NewGroupRoute("/api")
	route := NewRoute("/users").GET().Build()

	builder.Route(route)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Errorf("Route() count = %d, want 1", len(group.Routes))
	}
}

func TestGroupRouteBuilder_Route_Nil(t *testing.T) {
	builder := NewGroupRoute("/api")
	builder.Route(nil)

	group := builder.Build()
	if len(group.Routes) != 0 {
		t.Errorf("Route(nil) should not add route, got %d routes", len(group.Routes))
	}
}

func TestGroupRouteBuilder_Routes(t *testing.T) {
	builder := NewGroupRoute("/api")
	route1 := NewRoute("/users").GET().Build()
	route2 := NewRoute("/posts").POST().Build()

	builder.Routes(route1, route2)

	group := builder.Build()
	if len(group.Routes) != 2 {
		t.Errorf("Routes() count = %d, want 2", len(group.Routes))
	}
}

func TestGroupRouteBuilder_Group_Nil(t *testing.T) {
	builder := NewGroupRoute("/api")
	builder.Group(nil)

	group := builder.Build()
	if len(group.Groups) != 0 {
		t.Errorf("Group(nil) should not add group, got %d groups", len(group.Groups))
	}
}

func TestGroupRouteBuilder_POST(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }

	builder.POST("/users", handler)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("POST() should add route, got %d routes", len(group.Routes))
	}

	if group.Routes[0].Path != "/users" {
		t.Errorf("Route path = %s, want /users", group.Routes[0].Path)
	}

	if group.Routes[0].Method != core.POST {
		t.Errorf("Route method = %v, want POST", group.Routes[0].Method)
	}
}

func TestGroupRouteBuilder_PUT(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }

	builder.PUT("/users/:id", handler)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("PUT() should add route, got %d routes", len(group.Routes))
	}

	if group.Routes[0].Method != core.PUT {
		t.Errorf("Route method = %v, want PUT", group.Routes[0].Method)
	}
}

func TestGroupRouteBuilder_PATCH(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }

	builder.PATCH("/users/:id", handler)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("PATCH() should add route, got %d routes", len(group.Routes))
	}

	if group.Routes[0].Method != core.PATCH {
		t.Errorf("Route method = %v, want PATCH", group.Routes[0].Method)
	}
}

func TestGroupRouteBuilder_DELETE(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }

	builder.DELETE("/users/:id", handler)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("DELETE() should add route, got %d routes", len(group.Routes))
	}

	if group.Routes[0].Method != core.DELETE {
		t.Errorf("Route method = %v, want DELETE", group.Routes[0].Method)
	}
}

func TestGroupRouteBuilder_HEAD(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }

	builder.HEAD("/ping", handler)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("HEAD() should add route, got %d routes", len(group.Routes))
	}

	if group.Routes[0].Method != core.HEAD {
		t.Errorf("Route method = %v, want HEAD", group.Routes[0].Method)
	}
}

func TestGroupRouteBuilder_OPTIONS(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }

	builder.OPTIONS("/cors", handler)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("OPTIONS() should add route, got %d routes", len(group.Routes))
	}

	if group.Routes[0].Method != core.OPTIONS {
		t.Errorf("Route method = %v, want OPTIONS", group.Routes[0].Method)
	}
}

func TestGroupRouteBuilder_WithMiddleware(t *testing.T) {
	builder := NewGroupRoute("/api")
	handler := func(_ core.Context) error { return nil }
	mw := func(_ core.Context) error { return nil }

	builder.GET("/users", handler, mw)

	group := builder.Build()
	if len(group.Routes) != 1 {
		t.Fatalf("GET with middleware should add route, got %d routes", len(group.Routes))
	}

	if len(group.Routes[0].Middlewares) != 1 {
		t.Errorf("Route middlewares count = %d, want 1", len(group.Routes[0].Middlewares))
	}
}

func TestGroupRouteBuilder_FluentAPI(t *testing.T) {
	handler := func(_ core.Context) error { return nil }
	mw := func(_ core.Context) error { return nil }

	subGroup := NewGroupRoute("/v2").Build()

	group := NewGroupRoute("/api").
		Protected().
		Middleware(mw).
		GET("/users", handler).
		POST("/users", handler).
		Group(subGroup).
		Build()

	if group.Prefix != "/api" {
		t.Errorf("Prefix = %s, want /api", group.Prefix)
	}

	if !group.IsProtected {
		t.Error("IsProtected should be true")
	}

	if len(group.Middlewares) != 1 {
		t.Errorf("Middlewares count = %d, want 1", len(group.Middlewares))
	}

	if len(group.Routes) != 2 {
		t.Errorf("Routes count = %d, want 2", len(group.Routes))
	}

	if len(group.Groups) != 1 {
		t.Errorf("Groups count = %d, want 1", len(group.Groups))
	}
}
