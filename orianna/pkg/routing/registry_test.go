// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

func TestRouteRegistry_RegisterRoute(t *testing.T) {
	tests := []struct {
		name    string
		route   Route
		wantErr bool
		check   func(t *testing.T, registry *RouteRegistry, err error)
	}{
		{
			name: "valid route should register successfully",
			route: Route{
				Path:    "/test",
				Method:  core.GET,
				Handler: func(_ core.Context) error { return nil },
			},
			wantErr: false,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				if err != nil {
					t.Errorf("RegisterRoute() = %v, want nil", err)
				}
				if len(registry.GetRoutes()) != 1 {
					t.Errorf("RegisterRoute() routes count = %v, want 1", len(registry.GetRoutes()))
				}
			},
		},
		{
			name: "nil handler should return error",
			route: Route{
				Path:    "/test",
				Method:  core.GET,
				Handler: nil,
			},
			wantErr: true,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				if err == nil {
					t.Error("RegisterRoute() should return error for nil handler")
				}
			},
		},
		{
			name: "route with invalid path should return error",
			route: Route{
				Path:    "invalid-path",
				Method:  core.GET,
				Handler: func(_ core.Context) error { return nil },
			},
			wantErr: true,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				if err == nil {
					t.Error("RegisterRoute() should return error for invalid path")
				}
			},
		},
		{
			name: "protected route should have auth middleware applied",
			route: Route{
				Path:        "/protected",
				Method:      core.GET,
				Handler:     func(_ core.Context) error { return nil },
				IsProtected: true,
			},
			wantErr: false,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				routes := registry.GetRoutes()
				if len(routes) != 1 {
					t.Fatalf("RegisterRoute() routes count = %v, want 1", len(routes))
				}
				// Check that protection middleware was applied
				if len(routes[0].Middlewares) == 0 {
					t.Error("RegisterRoute() should apply protection middleware for protected route")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRouteRegistry()
			// Set auth middleware for protected route test
			authMiddleware := func(ctx core.Context) error { return ctx.Next() }
			registry.SetAuthMiddleware(authMiddleware)

			err := registry.RegisterRoute(tt.route)

			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.check(t, registry, err)
		})
	}
}

func TestRouteRegistry_RegisterGroup(t *testing.T) {
	tests := []struct {
		name    string
		group   RouteGroup
		wantErr bool
		check   func(t *testing.T, registry *RouteRegistry, err error)
	}{
		{
			name: "valid group should register successfully",
			group: RouteGroup{
				Prefix: "/api",
				Routes: []Route{
					{
						Path:    "/test",
						Method:  core.GET,
						Handler: func(_ core.Context) error { return nil },
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				if err != nil {
					t.Errorf("RegisterGroup() = %v, want nil", err)
				}
				if len(registry.GetGroups()) != 1 {
					t.Errorf("RegisterGroup() groups count = %v, want 1", len(registry.GetGroups()))
				}
			},
		},
		{
			name: "empty prefix should return error",
			group: RouteGroup{
				Prefix: "",
				Routes: []Route{
					{
						Path:    "/test",
						Method:  core.GET,
						Handler: func(_ core.Context) error { return nil },
					},
				},
			},
			wantErr: true,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				if err == nil {
					t.Error("RegisterGroup() should return error for empty prefix")
				}
			},
		},
		{
			name: "group with no routes should return error",
			group: RouteGroup{
				Prefix: "/api",
				Routes: []Route{},
			},
			wantErr: true,
			check: func(t *testing.T, registry *RouteRegistry, err error) {
				if err == nil {
					t.Error("RegisterGroup() should return error for group with no routes")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRouteRegistry()
			err := registry.RegisterGroup(tt.group)

			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.check(t, registry, err)
		})
	}
}
