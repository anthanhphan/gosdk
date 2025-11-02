package aurelion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRouteBuilder_Methods(t *testing.T) {
	tests := []struct {
		name    string
		builder func(string) *RouteBuilder
		want    Method
	}{
		{"GET method should set correctly", NewRoute, GET},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builder("/test")
			methods := []struct {
				name string
				call func() *RouteBuilder
				want Method
			}{
				{"GET", builder.GET, GET},
				{"POST", builder.POST, POST},
				{"PUT", builder.PUT, PUT},
				{"PATCH", builder.PATCH, PATCH},
				{"DELETE", builder.DELETE, DELETE},
				{"HEAD", func() *RouteBuilder { return builder.Method(HEAD) }, HEAD},
				{"OPTIONS", func() *RouteBuilder { return builder.Method(OPTIONS) }, OPTIONS},
			}

			for _, m := range methods {
				t.Run(m.name, func(t *testing.T) {
					route := m.call().Handler(func(ctx Context) error { return nil }).Build()
					if route.Method != m.want {
						t.Errorf("Method = %v, want %v", route.Method, m.want)
					}
				})
			}
		})
	}
}

func TestRouteBuilder_Path(t *testing.T) {
	route := NewRoute("/test").Path("/updated").Handler(func(ctx Context) error { return nil }).Build()

	if route.Path != "/updated" {
		t.Errorf("Path() = %v, want %v", route.Path, "/updated")
	}
}

func TestRouteBuilder_Method(t *testing.T) {
	route := NewRoute("/test").Method(POST).Handler(func(ctx Context) error { return nil }).Build()

	if route.Method != POST {
		t.Errorf("Method() = %v, want %v", route.Method, POST)
	}
}

func TestRouteBuilder_Handler(t *testing.T) {
	handler := func(ctx Context) error {
		return nil
	}

	route := NewRoute("/test").GET().Handler(handler).Build()

	if route.Handler == nil {
		t.Error("Handler should not be nil")
	}
}

func TestRouteBuilder_Middleware(t *testing.T) {
	middleware := func(ctx Context) error {
		return ctx.Next()
	}

	route := NewRoute("/test").GET().
		Middleware(middleware).
		Handler(func(ctx Context) error { return nil }).
		Build()

	if len(route.Middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(route.Middlewares))
	}
}

func TestRouteBuilder_MultipleMiddleware(t *testing.T) {
	m1 := func(ctx Context) error { return ctx.Next() }
	m2 := func(ctx Context) error { return ctx.Next() }

	route := NewRoute("/test").GET().
		Middleware(m1, m2).
		Handler(func(ctx Context) error { return nil }).
		Build()

	if len(route.Middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(route.Middlewares))
	}
}

func TestRouteBuilder_Protected(t *testing.T) {
	route := NewRoute("/test").GET().Protected().Handler(func(ctx Context) error { return nil }).Build()

	if !route.IsProtected {
		t.Error("Route should be protected")
	}
}

func TestRouteBuilder_Permissions(t *testing.T) {
	route := NewRoute("/test").GET().
		Permissions("read", "write").
		Handler(func(ctx Context) error { return nil }).
		Build()

	if !route.IsProtected {
		t.Error("Route should be protected when permissions are set")
	}

	if len(route.RequiredPermissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(route.RequiredPermissions))
	}

	expected := []string{"read", "write"}
	for i, perm := range route.RequiredPermissions {
		if perm != expected[i] {
			t.Errorf("Permission[%d] = %v, want %v", i, perm, expected[i])
		}
	}
}

func TestGroupRouteBuilder_Middleware(t *testing.T) {
	middleware := func(ctx Context) error {
		return ctx.Next()
	}

	group := NewGroupRoute("/api").
		Middleware(middleware).
		Routes(
			NewRoute("/users").GET().Handler(func(ctx Context) error { return nil }),
		).
		Build()

	if len(group.Middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(group.Middlewares))
	}
}

func TestGroupRouteBuilder_Protected(t *testing.T) {
	group := NewGroupRoute("/api").
		Protected().
		Routes(
			NewRoute("/users").GET().Handler(func(ctx Context) error { return nil }),
		).
		Build()

	if !group.IsProtected {
		t.Error("Group should be protected")
	}
}

func TestGroupRouteBuilder_Route(t *testing.T) {
	group := NewGroupRoute("/api").
		Route(NewRoute("/users").GET().Handler(func(ctx Context) error { return nil })).
		Build()

	if len(group.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(group.Routes))
	}

	if group.Routes[0].Path != "/users" {
		t.Errorf("Route path = %v, want %v", group.Routes[0].Path, "/users")
	}
}

func TestGroupRouteBuilder_Routes(t *testing.T) {
	group := NewGroupRoute("/api").
		Routes(
			NewRoute("/users").GET().Handler(func(ctx Context) error { return nil }),
			NewRoute("/posts").GET().Handler(func(ctx Context) error { return nil }),
		).
		Build()

	if len(group.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(group.Routes))
	}
}

func TestGroupRouteBuilder_InvalidRouteType(t *testing.T) {
	group := NewGroupRoute("/api").
		Route("invalid").
		Build()

	if len(group.Routes) != 0 {
		t.Error("Invalid route type should be skipped")
	}
}

func TestConvertToRoute(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "Route value should convert",
			input: Route{Path: "/test", Method: GET},
			want:  true,
		},
		{
			name:  "*Route should convert",
			input: &Route{Path: "/test", Method: GET},
			want:  true,
		},
		{
			name:  "*RouteBuilder should convert",
			input: NewRoute("/test").GET().Handler(func(ctx Context) error { return nil }),
			want:  true,
		},
		{
			name:  "invalid type should return nil",
			input: "invalid",
			want:  false,
		},
		{
			name:  "nil should return nil",
			input: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToRoute(tt.input)
			if (result != nil) != tt.want {
				t.Errorf("convertToRoute() result = %v, want not nil = %v", result != nil, tt.want)
			}
		})
	}
}

func TestGroupRouteBuilder_RouteWithNilBuilder(t *testing.T) {
	group := NewGroupRoute("/api").
		Route((*RouteBuilder)(nil)).
		Build()

	if len(group.Routes) != 0 {
		t.Error("Nil route builder should be skipped")
	}
}

func TestGroupRouteBuilder_RouteWithNilRoute(t *testing.T) {
	group := NewGroupRoute("/api").
		Route((*Route)(nil)).
		Build()

	if len(group.Routes) != 0 {
		t.Error("Nil route should be skipped")
	}
}

func TestBuildHandlers_WithNilMiddleware(t *testing.T) {
	route := &Route{
		Path:    "/test",
		Method:  GET,
		Handler: func(ctx Context) error { return nil },
		Middlewares: []Middleware{
			nil,
			func(ctx Context) error { return ctx.Next() },
		},
	}

	handlers := route.buildHandlers()

	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(handlers))
	}
}

func TestHandlerToFiber_NilHandler(t *testing.T) {
	fiberHandler := handlerToFiber(nil)

	app := fiber.New()
	app.Get("/test", fiberHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMiddlewareToFiber_NilMiddleware(t *testing.T) {
	fiberHandler := middlewareToFiber(nil)

	app := fiber.New()
	var callOrder []string
	app.Get("/test", fiberHandler, func(c *fiber.Ctx) error {
		callOrder = append(callOrder, "next")
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if len(callOrder) != 1 || callOrder[0] != "next" {
		t.Error("Nil middleware should pass through to next handler")
	}
}

func TestRegisterRoute_AllMethods(t *testing.T) {
	app := fiber.New()

	methods := []Method{GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS}

	for _, method := range methods {
		registerRoute(app, method, "/"+strings.ToLower(method.String()), []fiber.Handler{
			func(c *fiber.Ctx) error { return c.SendString("ok") },
		})
	}

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test POST
	req = httptest.NewRequest(http.MethodPost, "/post", nil)
	resp, err = app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Test PUT
	req = httptest.NewRequest(http.MethodPut, "/put", nil)
	resp, err = app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Test PATCH
	req = httptest.NewRequest(http.MethodPatch, "/patch", nil)
	resp, err = app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Test DELETE
	req = httptest.NewRequest(http.MethodDelete, "/delete", nil)
	resp, err = app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Test HEAD
	req = httptest.NewRequest(http.MethodHead, "/head", nil)
	resp, err = app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Test OPTIONS
	req = httptest.NewRequest(http.MethodOptions, "/options", nil)
	resp, err = app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestRegisterRoute_NilOrEmptyHandlers(t *testing.T) {
	app := fiber.New()

	// Test with nil handlers
	registerRoute(app, GET, "/test", nil)
	registerRoute(app, GET, "/test2", []fiber.Handler{})

	// Should not panic
}
