// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"context"
	"errors"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

func TestRegisterRoutes_Multiple(t *testing.T) {
	rr := NewRouteRegistry()

	routes := []Route{
		{Path: "/a", Method: core.GET, Handler: func(_ core.Context) error { return nil }},
		{Path: "/b", Method: core.POST, Handler: func(_ core.Context) error { return nil }},
	}

	err := rr.RegisterRoutes(routes...)
	if err != nil {
		t.Errorf("RegisterRoutes() error = %v", err)
	}

	if len(rr.GetRoutes()) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(rr.GetRoutes()))
	}
}

func TestRegisterRoutes_FailsOnInvalid(t *testing.T) {
	rr := NewRouteRegistry()

	routes := []Route{
		{Path: "/valid", Method: core.GET, Handler: func(_ core.Context) error { return nil }},
		{Path: "/invalid", Method: core.GET, Handler: nil}, // nil handler
	}

	err := rr.RegisterRoutes(routes...)
	if err == nil {
		t.Error("RegisterRoutes() should fail on invalid route")
	}

	// First route should have been registered before failure
	if len(rr.GetRoutes()) != 1 {
		t.Errorf("Expected 1 route (before failure), got %d", len(rr.GetRoutes()))
	}
}

func TestSetAuthzChecker(t *testing.T) {
	rr := NewRouteRegistry()

	checker := func(_ core.Context, _ []string) error { return nil }
	rr.SetAuthzChecker(checker)

	if rr.authzChecker == nil {
		t.Error("SetAuthzChecker() should set the checker")
	}
}

func TestApplyGroupProtection_ProtectedGroup(t *testing.T) {
	rr := NewRouteRegistry()

	authMw := func(ctx core.Context) error { return ctx.Next() }
	rr.SetAuthMiddleware(authMw)

	group := RouteGroup{
		Prefix:      "/api",
		IsProtected: true,
		Routes: []Route{
			{Path: "/users", Method: core.GET, Handler: func(_ core.Context) error { return nil }},
			{Path: "/items", Method: core.GET, Handler: func(_ core.Context) error { return nil }},
		},
	}

	err := rr.RegisterGroup(group)
	if err != nil {
		t.Errorf("RegisterGroup() error = %v", err)
	}

	groups := rr.GetGroups()
	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	// Routes in a protected group should have IsProtected set
	for _, route := range groups[0].Routes {
		if !route.IsProtected {
			t.Errorf("Route %s should be protected", route.Path)
		}
		// Should have auth middleware applied
		if len(route.Middlewares) == 0 {
			t.Errorf("Route %s should have middleware from protection", route.Path)
		}
	}
}

func TestApplyGroupProtection_NestedGroups(t *testing.T) {
	rr := NewRouteRegistry()

	authMw := func(ctx core.Context) error { return ctx.Next() }
	rr.SetAuthMiddleware(authMw)

	group := RouteGroup{
		Prefix:      "/api",
		IsProtected: true,
		Routes: []Route{
			{Path: "/users", Method: core.GET, Handler: func(_ core.Context) error { return nil }},
		},
		Groups: []RouteGroup{
			{
				Prefix: "/admin",
				Routes: []Route{
					{Path: "/dashboard", Method: core.GET, Handler: func(_ core.Context) error { return nil }},
				},
			},
		},
	}

	err := rr.RegisterGroup(group)
	if err != nil {
		t.Errorf("RegisterGroup() error = %v", err)
	}

	groups := rr.GetGroups()
	// Nested group should also be protected
	if !groups[0].Groups[0].IsProtected {
		t.Error("Nested group should inherit IsProtected from parent")
	}
}

func TestCreateAuthorizationMiddleware(t *testing.T) {
	rr := NewRouteRegistry()

	checkerCalled := false
	checker := func(_ core.Context, perms []string) error {
		checkerCalled = true
		if len(perms) != 1 || perms[0] != "admin" {
			t.Errorf("Expected permissions [admin], got %v", perms)
		}
		return nil
	}
	rr.SetAuthzChecker(checker)

	authzMw := rr.createAuthorizationMiddleware([]string{"admin"})

	// Create a simple mock context
	mockCtx := &simpleContext{}
	err := authzMw(mockCtx)
	if err != nil {
		t.Errorf("authz middleware should not return error, got %v", err)
	}
	if !checkerCalled {
		t.Error("authz checker should have been called")
	}
}

func TestCreateAuthorizationMiddleware_Denied(t *testing.T) {
	rr := NewRouteRegistry()

	checker := func(_ core.Context, perms []string) error {
		return errors.New("access denied")
	}
	rr.SetAuthzChecker(checker)

	authzMw := rr.createAuthorizationMiddleware([]string{"admin"})
	mockCtx := &simpleContext{}
	err := authzMw(mockCtx)
	if err == nil {
		t.Error("authz middleware should return error for denied access")
	}
}

func TestApplyProtectionMiddleware_WithAuthzChecker(t *testing.T) {
	rr := NewRouteRegistry()

	authMw := func(ctx core.Context) error { return ctx.Next() }
	authzChecker := func(_ core.Context, _ []string) error { return nil }

	rr.SetAuthMiddleware(authMw)
	rr.SetAuthzChecker(authzChecker)

	route := Route{
		Path:                "/admin",
		Method:              core.GET,
		Handler:             func(_ core.Context) error { return nil },
		IsProtected:         true,
		RequiredPermissions: []string{"admin"},
	}

	err := rr.RegisterRoute(route)
	if err != nil {
		t.Fatalf("RegisterRoute() error = %v", err)
	}

	routes := rr.GetRoutes()
	// Should have auth + authz middleware
	if len(routes[0].Middlewares) < 2 {
		t.Errorf("Expected at least 2 middlewares (auth+authz), got %d", len(routes[0].Middlewares))
	}
}

func TestValidateGroup_NoPrefix(t *testing.T) {
	rr := NewRouteRegistry()
	group := RouteGroup{
		Prefix: "",
		Routes: []Route{
			{Path: "/test", Handler: func(_ core.Context) error { return nil }},
		},
	}
	err := rr.RegisterGroup(group)
	if err == nil {
		t.Error("RegisterGroup() should error for empty prefix")
	}
}

func TestValidateGroup_InvalidPrefix(t *testing.T) {
	rr := NewRouteRegistry()
	group := RouteGroup{
		Prefix: "no-slash",
		Routes: []Route{
			{Path: "/test", Handler: func(_ core.Context) error { return nil }},
		},
	}
	err := rr.RegisterGroup(group)
	if err == nil {
		t.Error("RegisterGroup() should error for prefix without leading /")
	}
}

func TestValidateGroup_NoRoutes(t *testing.T) {
	rr := NewRouteRegistry()
	group := RouteGroup{
		Prefix: "/api",
	}
	err := rr.RegisterGroup(group)
	if err == nil {
		t.Error("RegisterGroup() should error for group with no routes or subgroups")
	}
}

func TestValidateGroup_InvalidSubgroup(t *testing.T) {
	rr := NewRouteRegistry()
	group := RouteGroup{
		Prefix: "/api",
		Groups: []RouteGroup{
			{
				Prefix: "", // invalid
				Routes: []Route{
					{Path: "/test", Handler: func(_ core.Context) error { return nil }},
				},
			},
		},
	}
	err := rr.RegisterGroup(group)
	if err == nil {
		t.Error("RegisterGroup() should error for invalid subgroup")
	}
}

func TestValidateRoute_NilRoute(t *testing.T) {
	err := validateRoute(nil)
	if err == nil {
		t.Error("validateRoute() should error for nil route")
	}
}

func TestValidateRoute_NoLeadingSlash(t *testing.T) {
	route := &Route{
		Path:    "no-slash",
		Handler: func(_ core.Context) error { return nil },
	}
	err := validateRoute(route)
	if err == nil {
		t.Error("validateRoute() should error for path without leading /")
	}
}

// simpleContext is a minimal mock for testing authorization middleware
type simpleContext struct{}

func (c *simpleContext) Next() error                       { return nil }
func (c *simpleContext) Context() context.Context          { return context.Background() }
func (c *simpleContext) Method() string                    { return "GET" }
func (c *simpleContext) Path() string                      { return "/" }
func (c *simpleContext) RoutePath() string                 { return "/" }
func (c *simpleContext) OriginalURL() string               { return "/" }
func (c *simpleContext) BaseURL() string                   { return "" }
func (c *simpleContext) Protocol() string                  { return "http" }
func (c *simpleContext) Hostname() string                  { return "localhost" }
func (c *simpleContext) IP() string                        { return "127.0.0.1" }
func (c *simpleContext) Secure() bool                      { return false }
func (c *simpleContext) Get(string, ...string) string      { return "" }
func (c *simpleContext) Set(string, string)                {}
func (c *simpleContext) Append(string, ...string)          {}
func (c *simpleContext) Params(string, ...string) string   { return "" }
func (c *simpleContext) AllParams() map[string]string      { return nil }
func (c *simpleContext) ParamsParser(any) error            { return nil }
func (c *simpleContext) Query(string, ...string) string    { return "" }
func (c *simpleContext) AllQueries() map[string]string     { return nil }
func (c *simpleContext) QueryParser(any) error             { return nil }
func (c *simpleContext) Body() []byte                      { return nil }
func (c *simpleContext) BodyParser(any) error              { return nil }
func (c *simpleContext) Cookies(string, ...string) string  { return "" }
func (c *simpleContext) Cookie(*core.Cookie)               {}
func (c *simpleContext) ClearCookie(...string)             {}
func (c *simpleContext) Status(int) core.Context           { return c }
func (c *simpleContext) ResponseStatusCode() int           { return 200 }
func (c *simpleContext) JSON(any) error                    { return nil }
func (c *simpleContext) XML(any) error                     { return nil }
func (c *simpleContext) SendString(string) error           { return nil }
func (c *simpleContext) SendBytes([]byte) error            { return nil }
func (c *simpleContext) Redirect(string, ...int) error     { return nil }
func (c *simpleContext) Accepts(...string) string          { return "" }
func (c *simpleContext) AcceptsCharsets(...string) string  { return "" }
func (c *simpleContext) AcceptsEncodings(...string) string { return "" }
func (c *simpleContext) AcceptsLanguages(...string) string { return "" }
func (c *simpleContext) Fresh() bool                       { return false }
func (c *simpleContext) Stale() bool                       { return false }
func (c *simpleContext) XHR() bool                         { return false }
func (c *simpleContext) Locals(string, ...any) any         { return nil }
func (c *simpleContext) GetAllLocals() map[string]any      { return nil }
func (c *simpleContext) OK(any) error                      { return nil }
func (c *simpleContext) Created(any) error                 { return nil }
func (c *simpleContext) NoContent() error                  { return nil }
func (c *simpleContext) BadRequestMsg(string) error        { return nil }
func (c *simpleContext) UnauthorizedMsg(string) error      { return nil }
func (c *simpleContext) ForbiddenMsg(string) error         { return nil }
func (c *simpleContext) NotFoundMsg(string) error          { return nil }
func (c *simpleContext) InternalErrorMsg(string) error     { return nil }
func (c *simpleContext) IsMethod(string) bool              { return false }
func (c *simpleContext) RequestID() string                 { return "" }
func (c *simpleContext) UseProperHTTPStatus() bool         { return false }
