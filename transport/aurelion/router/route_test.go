package router

import (
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/gofiber/fiber/v2"
)

func TestRouteBuilder_Build(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *RouteBuilder
		check func(t *testing.T, route *Route)
	}{
		{
			name: "basic GET route should build correctly",
			setup: func() *RouteBuilder {
				return NewRoute("/users").GET().Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if route.Path != "/users" {
					t.Errorf("Path = %s, want /users", route.Path)
				}
				if route.Method != MethodGet {
					t.Errorf("Method = %s, want GET", route.Method)
				}
				if route.Handler == nil {
					t.Error("Handler should not be nil")
				}
			},
		},
		{
			name: "POST route should build correctly",
			setup: func() *RouteBuilder {
				return NewRoute("/users").POST().Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if route.Method != MethodPost {
					t.Errorf("Method = %s, want POST", route.Method)
				}
			},
		},
		{
			name: "PUT route should build correctly",
			setup: func() *RouteBuilder {
				return NewRoute("/users/:id").PUT().Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if route.Method != MethodPut {
					t.Errorf("Method = %s, want PUT", route.Method)
				}
			},
		},
		{
			name: "PATCH route should build correctly",
			setup: func() *RouteBuilder {
				return NewRoute("/users/:id").PATCH().Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if route.Method != MethodPatch {
					t.Errorf("Method = %s, want PATCH", route.Method)
				}
			},
		},
		{
			name: "DELETE route should build correctly",
			setup: func() *RouteBuilder {
				return NewRoute("/users/:id").DELETE().Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if route.Method != MethodDelete {
					t.Errorf("Method = %s, want DELETE", route.Method)
				}
			},
		},
		{
			name: "protected route should set IsProtected",
			setup: func() *RouteBuilder {
				return NewRoute("/admin").GET().Protected().Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if !route.IsProtected {
					t.Error("IsProtected should be true")
				}
			},
		},
		{
			name: "route with permissions should set IsProtected and permissions",
			setup: func() *RouteBuilder {
				return NewRoute("/admin/users").GET().
					Permissions("admin", "read:users").
					Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if !route.IsProtected {
					t.Error("IsProtected should be true when permissions are set")
				}
				if len(route.RequiredPermissions) != 2 {
					t.Errorf("RequiredPermissions length = %d, want 2", len(route.RequiredPermissions))
				}
			},
		},
		{
			name: "route with middleware should store middlewares",
			setup: func() *RouteBuilder {
				mw1 := func(ctx ContextInterface) error { return ctx.Next() }
				mw2 := func(ctx ContextInterface) error { return ctx.Next() }
				return NewRoute("/test").GET().
					Middleware(mw1, mw2).
					Handler(func(ctx ContextInterface) error { return nil })
			},
			check: func(t *testing.T, route *Route) {
				if len(route.Middlewares) != 2 {
					t.Errorf("Middlewares length = %d, want 2", len(route.Middlewares))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			route := builder.Build()
			tt.check(t, route)
		})
	}
}

func TestRouteBuilder_Path(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "set path should override initial path",
			path:     "/new-path",
			wantPath: "/new-path",
		},
		{
			name:     "path with params should be set correctly",
			path:     "/users/:id/posts/:postId",
			wantPath: "/users/:id/posts/:postId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := NewRoute("/initial").Path(tt.path).GET().Handler(func(ctx ContextInterface) error { return nil }).Build()
			if route.Path != tt.wantPath {
				t.Errorf("Path = %s, want %s", route.Path, tt.wantPath)
			}
		})
	}
}

func TestRouteBuilder_Method(t *testing.T) {
	tests := []struct {
		name       string
		method     interface{}
		wantMethod Method
	}{
		{
			name:       "set GET method",
			method:     core.MethodGet,
			wantMethod: MethodGet,
		},
		{
			name:       "set POST method",
			method:     core.MethodPost,
			wantMethod: MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := NewRoute("/test").Method(tt.method).Handler(func(ctx ContextInterface) error { return nil }).Build()
			// Convert core.Method to router.Method for comparison
			wantMethod := Method(tt.wantMethod)
			if route.Method != wantMethod {
				t.Errorf("Method = %s, want %s", route.Method, wantMethod)
			}
		})
	}
}

func TestValidateRoute(t *testing.T) {
	tests := []struct {
		name    string
		route   *Route
		wantErr bool
	}{
		{
			name: "valid route should pass",
			route: &Route{
				Path:    "/test",
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
			},
			wantErr: false,
		},
		{
			name:    "nil route should return error",
			route:   nil,
			wantErr: true,
		},
		{
			name: "route without leading slash should return error",
			route: &Route{
				Path:    "test",
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
			},
			wantErr: true,
		},
		{
			name: "route without handler should return error",
			route: &Route{
				Path:   "/test",
				Method: MethodGet,
			},
			wantErr: true,
		},
		{
			name: "path longer than max should return error",
			route: &Route{
				Path:    "/" + string(make([]byte, MaxRoutePathLength+1)),
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
			},
			wantErr: true,
		},
		{
			name: "too many handlers should return error",
			route: &Route{
				Path:        "/test",
				Method:      MethodGet,
				Handler:     func(ctx ContextInterface) error { return nil },
				Middlewares: make([]Middleware, MaxRouteHandlersPerRoute),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoute(tt.route)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPerRouteCORS(t *testing.T) {
	route := NewRoute("/test").GET().CORS(&CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Content-Type"},
	}).Handler(func(ctx ContextInterface) error { return nil }).Build()

	if route.CORS == nil {
		t.Fatal("expected CORS config")
	}

	clone := route.Clone()
	if clone == route || clone.CORS == route.CORS {
		t.Fatal("expected deep copy of route and CORS config")
	}
}

func TestRoute_Clone(t *testing.T) {
	tests := []struct {
		name  string
		route *Route
		check func(t *testing.T, original *Route, clone *Route)
	}{
		{
			name:  "nil route should return nil",
			route: nil,
			check: func(t *testing.T, original *Route, clone *Route) {
				if clone != nil {
					t.Error("Clone of nil route should be nil")
				}
			},
		},
		{
			name: "cloned route should be deep copy",
			route: &Route{
				Path:                "/test",
				Method:              MethodGet,
				Handler:             func(ctx ContextInterface) error { return nil },
				Middlewares:         []Middleware{func(ctx ContextInterface) error { return nil }},
				RequiredPermissions: []string{"admin"},
				IsProtected:         true,
			},
			check: func(t *testing.T, original *Route, clone *Route) {
				if clone == original {
					t.Error("Clone should not be same pointer as original")
				}
				if clone.Path != original.Path {
					t.Error("Path should be copied")
				}
				if len(clone.Middlewares) != len(original.Middlewares) {
					t.Error("Middlewares length should match")
				}
				if len(clone.RequiredPermissions) != len(original.RequiredPermissions) {
					t.Error("RequiredPermissions length should match")
				}
			},
		},
		{
			name: "clone with CORS should deep copy CORS config",
			route: &Route{
				Path:    "/test",
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
				CORS: &CORSConfig{
					AllowOrigins: []string{"*"},
					AllowMethods: []string{"GET"},
				},
			},
			check: func(t *testing.T, original *Route, clone *Route) {
				if clone.CORS == original.CORS {
					t.Error("CORS config should be deep copied")
				}
				if len(clone.CORS.AllowOrigins) != len(original.CORS.AllowOrigins) {
					t.Error("AllowOrigins should be copied")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.route.Clone()
			tt.check(t, tt.route, clone)
		})
	}
}

func TestRoute_String(t *testing.T) {
	route := &Route{
		Path:   "/users",
		Method: MethodGet,
	}

	result := route.String()
	expected := "GET /users"
	if result != expected {
		t.Errorf("String() = %s, want %s", result, expected)
	}
}

func TestRoute_Register(t *testing.T) {
	tests := []struct {
		name  string
		route *Route
		check func(t *testing.T, app *fiber.App)
	}{
		{
			name:  "nil route should not panic",
			route: nil,
			check: func(t *testing.T, app *fiber.App) {
				// Should not panic
			},
		},
		{
			name: "valid route should register",
			route: &Route{
				Path:    "/test",
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
			},
			check: func(t *testing.T, app *fiber.App) {
				// Route registered, check via app.Stack()
				if len(app.Stack()) == 0 {
					t.Error("Expected route to be registered")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			if tt.route != nil {
				tt.route.Register(app)
			}
			tt.check(t, app)
		})
	}
}

func TestGroupRouteBuilder(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *GroupRouteBuilder
		check func(t *testing.T, group *GroupRoute)
	}{
		{
			name: "basic group should build correctly",
			setup: func() *GroupRouteBuilder {
				return NewGroupRoute("/api/v1").Routes(
					NewRoute("/users").GET().Handler(func(ctx ContextInterface) error { return nil }),
				)
			},
			check: func(t *testing.T, group *GroupRoute) {
				if group.Prefix != "/api/v1" {
					t.Errorf("Prefix = %s, want /api/v1", group.Prefix)
				}
				if len(group.Routes) != 1 {
					t.Errorf("Routes length = %d, want 1", len(group.Routes))
				}
			},
		},
		{
			name: "protected group should mark all routes as protected",
			setup: func() *GroupRouteBuilder {
				return NewGroupRoute("/admin").Protected().Routes(
					NewRoute("/users").GET().Handler(func(ctx ContextInterface) error { return nil }),
					NewRoute("/posts").GET().Handler(func(ctx ContextInterface) error { return nil }),
				)
			},
			check: func(t *testing.T, group *GroupRoute) {
				if !group.IsProtected {
					t.Error("IsProtected should be true")
				}
				if len(group.Routes) != 2 {
					t.Errorf("Routes length = %d, want 2", len(group.Routes))
				}
			},
		},
		{
			name: "group with middleware should store middlewares",
			setup: func() *GroupRouteBuilder {
				mw := func(ctx ContextInterface) error { return ctx.Next() }
				return NewGroupRoute("/api").Middleware(mw).Routes(
					NewRoute("/test").GET().Handler(func(ctx ContextInterface) error { return nil }),
				)
			},
			check: func(t *testing.T, group *GroupRoute) {
				if len(group.Middlewares) != 1 {
					t.Errorf("Middlewares length = %d, want 1", len(group.Middlewares))
				}
			},
		},
		{
			name: "single route addition should work",
			setup: func() *GroupRouteBuilder {
				return NewGroupRoute("/api").
					Route(NewRoute("/single").GET().Handler(func(ctx ContextInterface) error { return nil }))
			},
			check: func(t *testing.T, group *GroupRoute) {
				if len(group.Routes) != 1 {
					t.Errorf("Routes length = %d, want 1", len(group.Routes))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			group := builder.Build()
			tt.check(t, group)
		})
	}
}

func TestGroupRoute_Register(t *testing.T) {
	tests := []struct {
		name  string
		group *GroupRoute
		check func(t *testing.T, app *fiber.App)
	}{
		{
			name:  "nil group should not panic",
			group: nil,
			check: func(t *testing.T, app *fiber.App) {
				// Should not panic
			},
		},
		{
			name: "valid group should register all routes",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{
					{
						Path:    "/test",
						Method:  MethodGet,
						Handler: func(ctx ContextInterface) error { return nil },
					},
				},
			},
			check: func(t *testing.T, app *fiber.App) {
				if len(app.Stack()) == 0 {
					t.Error("Expected routes to be registered")
				}
			},
		},
		{
			name: "group with empty route path should use root",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{
					{
						Path:    "",
						Method:  MethodGet,
						Handler: func(ctx ContextInterface) error { return nil },
					},
				},
			},
			check: func(t *testing.T, app *fiber.App) {
				// Should register without error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			if tt.group != nil {
				tt.group.Register(app)
			}
			tt.check(t, app)
		})
	}
}

func TestValidateGroupRoute(t *testing.T) {
	tests := []struct {
		name    string
		group   *GroupRoute
		wantErr bool
	}{
		{
			name:    "nil group should return error",
			group:   nil,
			wantErr: true,
		},
		{
			name: "group without prefix should return error",
			group: &GroupRoute{
				Routes: []Route{
					{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
				},
			},
			wantErr: true,
		},
		{
			name: "group with invalid prefix should return error",
			group: &GroupRoute{
				Prefix: "api",
				Routes: []Route{
					{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
				},
			},
			wantErr: true,
		},
		{
			name: "group without routes should return error",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{},
			},
			wantErr: true,
		},
		{
			name: "valid group should pass",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{
					{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupRoute(tt.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGroupRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToRoute(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, result *Route)
	}{
		{
			name:  "RouteBuilder should convert to Route",
			input: NewRoute("/test").GET().Handler(func(ctx ContextInterface) error { return nil }),
			check: func(t *testing.T, result *Route) {
				if result == nil {
					t.Error("ToRoute should return non-nil Route")
					return
				}
				if result.Path != "/test" {
					t.Error("Path should be preserved")
				}
			},
		},
		{
			name:  "nil RouteBuilder should return nil",
			input: (*RouteBuilder)(nil),
			check: func(t *testing.T, result *Route) {
				if result != nil {
					t.Error("ToRoute of nil RouteBuilder should return nil")
				}
			},
		},
		{
			name:  "*Route should return itself",
			input: &Route{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
			check: func(t *testing.T, result *Route) {
				if result == nil {
					t.Error("ToRoute should return non-nil Route")
				}
			},
		},
		{
			name:  "Route value should convert to pointer",
			input: Route{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
			check: func(t *testing.T, result *Route) {
				if result == nil {
					t.Error("ToRoute should return non-nil Route")
				}
			},
		},
		{
			name:  "invalid type should return nil",
			input: "invalid",
			check: func(t *testing.T, result *Route) {
				if result != nil {
					t.Error("ToRoute of invalid type should return nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToRoute(tt.input)
			tt.check(t, result)
		})
	}
}

func TestToGroupRoute(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, result *GroupRoute)
	}{
		{
			name: "GroupRouteBuilder should convert to GroupRoute",
			input: NewGroupRoute("/api").Routes(
				NewRoute("/test").GET().Handler(func(ctx ContextInterface) error { return nil }),
			),
			check: func(t *testing.T, result *GroupRoute) {
				if result == nil {
					t.Error("ToGroupRoute should return non-nil GroupRoute")
					return
				}
				if result.Prefix != "/api" {
					t.Error("Prefix should be preserved")
				}
			},
		},
		{
			name:  "nil GroupRouteBuilder should return nil",
			input: (*GroupRouteBuilder)(nil),
			check: func(t *testing.T, result *GroupRoute) {
				if result != nil {
					t.Error("ToGroupRoute of nil GroupRouteBuilder should return nil")
				}
			},
		},
		{
			name:  "*GroupRoute should return itself",
			input: &GroupRoute{Prefix: "/api", Routes: []Route{{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }}}},
			check: func(t *testing.T, result *GroupRoute) {
				if result == nil {
					t.Error("ToGroupRoute should return non-nil GroupRoute")
				}
			},
		},
		{
			name:  "GroupRoute value should convert to pointer",
			input: GroupRoute{Prefix: "/api", Routes: []Route{{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }}}},
			check: func(t *testing.T, result *GroupRoute) {
				if result == nil {
					t.Error("ToGroupRoute should return non-nil GroupRoute")
				}
			},
		},
		{
			name:  "invalid type should return nil",
			input: "invalid",
			check: func(t *testing.T, result *GroupRoute) {
				if result != nil {
					t.Error("ToGroupRoute of invalid type should return nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToGroupRoute(tt.input)
			tt.check(t, result)
		})
	}
}

func TestRoute_Handlers(t *testing.T) {
	tests := []struct {
		name  string
		route *Route
		check func(t *testing.T, handlers []fiber.Handler)
	}{
		{
			name: "route with handler and middleware should return all handlers",
			route: &Route{
				Path:    "/test",
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
				Middlewares: []Middleware{
					func(ctx ContextInterface) error { return ctx.Next() },
				},
			},
			check: func(t *testing.T, handlers []fiber.Handler) {
				if len(handlers) != 2 {
					t.Errorf("Handlers length = %d, want 2", len(handlers))
				}
			},
		},
		{
			name:  "nil route should return nil handlers",
			route: nil,
			check: func(t *testing.T, handlers []fiber.Handler) {
				if handlers != nil {
					t.Error("Handlers of nil route should be nil")
				}
			},
		},
		{
			name: "route with nil middleware should skip it",
			route: &Route{
				Path:    "/test",
				Method:  MethodGet,
				Handler: func(ctx ContextInterface) error { return nil },
				Middlewares: []Middleware{
					nil,
					func(ctx ContextInterface) error { return ctx.Next() },
				},
			},
			check: func(t *testing.T, handlers []fiber.Handler) {
				// Should have handler + 1 valid middleware (nil skipped)
				if len(handlers) < 2 {
					t.Errorf("Handlers length = %d, should have at least 2", len(handlers))
				}
			},
		},
		{
			name: "route without handler should return only middlewares",
			route: &Route{
				Path:    "/test",
				Method:  MethodGet,
				Handler: nil,
				Middlewares: []Middleware{
					func(ctx ContextInterface) error { return ctx.Next() },
				},
			},
			check: func(t *testing.T, handlers []fiber.Handler) {
				if len(handlers) != 1 {
					t.Errorf("Handlers length = %d, want 1", len(handlers))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handlers []fiber.Handler
			if tt.route != nil {
				handlers = tt.route.Handlers()
			}
			tt.check(t, handlers)
		})
	}
}

func TestRegisterRoute_AllMethods(t *testing.T) {
	tests := []struct {
		name   string
		method interface{}
	}{
		{name: "POST method", method: MethodPost},
		{name: "PUT method", method: MethodPut},
		{name: "PATCH method", method: MethodPatch},
		{name: "DELETE method", method: MethodDelete},
		{name: "HEAD method", method: core.MethodHead},
		{name: "OPTIONS method", method: core.MethodOptions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			// Convert method to router.Method
			var method Method
			switch m := tt.method.(type) {
			case Method:
				method = m
			case core.Method:
				method = Method(m)
			default:
				t.Fatalf("unexpected method type: %T", m)
			}
			route := &Route{
				Path:    "/test",
				Method:  method,
				Handler: func(ctx ContextInterface) error { return nil },
			}
			route.Register(app)

			// Verify route was registered
			if len(app.Stack()) == 0 {
				t.Error("Expected route to be registered")
			}
		})
	}
}

func TestRoute_NilAppRegister(t *testing.T) {
	route := &Route{
		Path:    "/test",
		Method:  MethodGet,
		Handler: func(ctx ContextInterface) error { return nil },
	}
	// Should not panic
	route.Register(nil)
}

func TestGroupRoute_NilAppRegister(t *testing.T) {
	group := &GroupRoute{
		Prefix: "/api",
		Routes: []Route{
			{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
		},
	}
	// Should not panic
	group.Register(nil)
}

func TestGroupRouteBuilder_NilRoute(t *testing.T) {
	// Test Route() with nil builder
	builder := NewGroupRoute("/api")
	result := builder.Route(nil)
	if result != builder {
		t.Error("Route() should return builder for chaining")
	}

	group := builder.Build()
	if len(group.Routes) != 0 {
		t.Error("Nil route should not be added")
	}
}

func TestValidateGroupRoute_InvalidRoutes(t *testing.T) {
	tests := []struct {
		name    string
		group   *GroupRoute
		wantErr bool
	}{
		{
			name: "group with too many middlewares should return error",
			group: &GroupRoute{
				Prefix:      "/api",
				Middlewares: make([]Middleware, MaxRouteHandlersPerRoute+1),
				Routes: []Route{
					{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
				},
			},
			wantErr: true,
		},
		{
			name: "group with invalid route in middle should return error",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{
					{Path: "/valid", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
					{Path: "invalid", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
				},
			},
			wantErr: true,
		},
		{
			name: "group prefix longer than max should return error",
			group: &GroupRoute{
				Prefix: "/" + string(make([]byte, MaxRoutePathLength+1)),
				Routes: []Route{
					{Path: "/test", Method: MethodGet, Handler: func(ctx ContextInterface) error { return nil }},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupRoute(tt.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGroupRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
