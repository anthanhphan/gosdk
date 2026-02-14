// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"errors"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

func dummyHandler(_ core.Context) error { return nil }

func TestDuplicateRoute_SameMethodAndPath(t *testing.T) {
	rr := NewRouteRegistry()

	route := Route{Path: "/users", Method: core.GET, Handler: dummyHandler}

	if err := rr.RegisterRoute(route); err != nil {
		t.Fatalf("first RegisterRoute() error = %v", err)
	}

	err := rr.RegisterRoute(route)
	if err == nil {
		t.Fatal("second RegisterRoute() should return duplicate error")
	}
	if !errors.Is(err, core.ErrDuplicateRoute) {
		t.Errorf("error should wrap ErrDuplicateRoute, got: %v", err)
	}
}

func TestDuplicateRoute_DifferentMethodsSamePath(t *testing.T) {
	rr := NewRouteRegistry()

	get := Route{Path: "/users", Method: core.GET, Handler: dummyHandler}
	post := Route{Path: "/users", Method: core.POST, Handler: dummyHandler}

	if err := rr.RegisterRoute(get); err != nil {
		t.Fatalf("GET RegisterRoute() error = %v", err)
	}
	if err := rr.RegisterRoute(post); err != nil {
		t.Fatalf("POST RegisterRoute() should succeed, got error = %v", err)
	}
}

func TestDuplicateRoute_DifferentPathsSameMethod(t *testing.T) {
	rr := NewRouteRegistry()

	r1 := Route{Path: "/users", Method: core.GET, Handler: dummyHandler}
	r2 := Route{Path: "/items", Method: core.GET, Handler: dummyHandler}

	if err := rr.RegisterRoute(r1); err != nil {
		t.Fatalf("first RegisterRoute() error = %v", err)
	}
	if err := rr.RegisterRoute(r2); err != nil {
		t.Fatalf("second RegisterRoute() should succeed, got error = %v", err)
	}
}

func TestDuplicateRoute_MultiMethods(t *testing.T) {
	rr := NewRouteRegistry()

	r1 := Route{Path: "/users", Methods: []core.Method{core.GET, core.POST}, Handler: dummyHandler}
	if err := rr.RegisterRoute(r1); err != nil {
		t.Fatalf("RegisterRoute() error = %v", err)
	}

	// Register same path with overlapping method
	r2 := Route{Path: "/users", Method: core.GET, Handler: dummyHandler}
	err := rr.RegisterRoute(r2)
	if err == nil {
		t.Fatal("RegisterRoute() should return duplicate error for overlapping method")
	}
	if !errors.Is(err, core.ErrDuplicateRoute) {
		t.Errorf("error should wrap ErrDuplicateRoute, got: %v", err)
	}
}

func TestDuplicateRoute_RegisterRoutes_BatchDuplicate(t *testing.T) {
	rr := NewRouteRegistry()

	routes := []Route{
		{Path: "/a", Method: core.GET, Handler: dummyHandler},
		{Path: "/b", Method: core.POST, Handler: dummyHandler},
		{Path: "/a", Method: core.GET, Handler: dummyHandler}, // duplicate
	}

	err := rr.RegisterRoutes(routes...)
	if err == nil {
		t.Fatal("RegisterRoutes() should return duplicate error")
	}
	if !errors.Is(err, core.ErrDuplicateRoute) {
		t.Errorf("error should wrap ErrDuplicateRoute, got: %v", err)
	}
}

func TestDuplicateRoute_GroupDuplicate(t *testing.T) {
	rr := NewRouteRegistry()

	group := RouteGroup{
		Prefix: "/api",
		Routes: []Route{
			{Path: "/users", Method: core.GET, Handler: dummyHandler},
			{Path: "/users", Method: core.GET, Handler: dummyHandler}, // duplicate within group
		},
	}

	err := rr.RegisterGroup(group)
	if err == nil {
		t.Fatal("RegisterGroup() should return duplicate error")
	}
	if !errors.Is(err, core.ErrDuplicateRoute) {
		t.Errorf("error should wrap ErrDuplicateRoute, got: %v", err)
	}
}

func TestDuplicateRoute_GroupAndTopLevel(t *testing.T) {
	rr := NewRouteRegistry()

	// Register a top-level route
	route := Route{Path: "/api/users", Method: core.GET, Handler: dummyHandler}
	if err := rr.RegisterRoute(route); err != nil {
		t.Fatalf("RegisterRoute() error = %v", err)
	}

	// Register a group route that resolves to the same path
	group := RouteGroup{
		Prefix: "/api",
		Routes: []Route{
			{Path: "/users", Method: core.GET, Handler: dummyHandler},
		},
	}

	err := rr.RegisterGroup(group)
	if err == nil {
		t.Fatal("RegisterGroup() should return duplicate error for conflicting path")
	}
	if !errors.Is(err, core.ErrDuplicateRoute) {
		t.Errorf("error should wrap ErrDuplicateRoute, got: %v", err)
	}
}

func TestDuplicateRoute_NestedGroupDuplicate(t *testing.T) {
	rr := NewRouteRegistry()

	group := RouteGroup{
		Prefix: "/api",
		Routes: []Route{
			{Path: "/users", Method: core.GET, Handler: dummyHandler},
		},
		Groups: []RouteGroup{
			{
				Prefix: "/v2",
				Routes: []Route{
					{Path: "/users", Method: core.GET, Handler: dummyHandler}, // /api/v2/users - no conflict
				},
			},
		},
	}

	// This should succeed --different full paths
	if err := rr.RegisterGroup(group); err != nil {
		t.Fatalf("RegisterGroup() should succeed, got error = %v", err)
	}
}
