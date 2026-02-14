// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		check func(t *testing.T, builder *RouteBuilder)
	}{
		{
			name: "should create route builder with path",
			path: "/test",
			check: func(t *testing.T, builder *RouteBuilder) {
				if builder == nil {
					t.Fatal("NewRoute() should not return nil")
					return
				}
				if builder.route.Path != "/test" {
					t.Errorf("NewRoute() path = %v, want '/test'", builder.route.Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRoute(tt.path)
			tt.check(t, builder)
		})
	}
}

func TestRouteBuilder_Method(t *testing.T) {
	tests := []struct {
		name   string
		method core.Method
		check  func(t *testing.T, builder *RouteBuilder)
	}{
		{
			name:   "should set GET method",
			method: core.GET,
			check: func(t *testing.T, builder *RouteBuilder) {
				if builder.route.Method != core.GET {
					t.Errorf("Method() = %v, want GET", builder.route.Method)
				}
			},
		},
		{
			name:   "should set POST method",
			method: core.POST,
			check: func(t *testing.T, builder *RouteBuilder) {
				if builder.route.Method != core.POST {
					t.Errorf("Method() = %v, want POST", builder.route.Method)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRoute("/test")
			builder.Method(tt.method)
			tt.check(t, builder)
		})
	}
}

func TestRouteBuilder_Methods(t *testing.T) {
	tests := []struct {
		name    string
		methods []core.Method
		check   func(t *testing.T, builder *RouteBuilder)
	}{
		{
			name:    "should set multiple methods",
			methods: []core.Method{core.GET, core.POST},
			check: func(t *testing.T, builder *RouteBuilder) {
				if len(builder.route.Methods) != 2 {
					t.Errorf("Methods length = %d, want 2", len(builder.route.Methods))
				}
				// Check content (order depends on how we append, but here it's simple append)
				if builder.route.Methods[0] != core.GET {
					t.Errorf("Methods[0] = %v, want GET", builder.route.Methods[0])
				}
				if builder.route.Methods[1] != core.POST {
					t.Errorf("Methods[1] = %v, want POST", builder.route.Methods[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRoute("/test")
			builder.Methods(tt.methods...)
			tt.check(t, builder)
		})
	}
}

func TestRouteBuilder_GET(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, builder *RouteBuilder)
	}{
		{
			name: "should set GET method",
			check: func(t *testing.T, builder *RouteBuilder) {
				if builder.route.Method != core.GET {
					t.Errorf("GET() = %v, want GET", builder.route.Method)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRoute("/test")
			builder.GET()
			tt.check(t, builder)
		})
	}
}

func TestRouteBuilder_Protected(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, builder *RouteBuilder)
	}{
		{
			name: "should mark route as protected",
			check: func(t *testing.T, builder *RouteBuilder) {
				if !builder.route.IsProtected {
					t.Error("Protected() should set IsProtected to true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRoute("/test")
			builder.Protected()
			tt.check(t, builder)
		})
	}
}

func TestRouteBuilder_Build(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*RouteBuilder)
		check func(t *testing.T, route *Route)
	}{
		{
			name: "should build route with all settings",
			setup: func(builder *RouteBuilder) {
				builder.Path("/test").
					GET().
					Handler(func(_ core.Context) error { return nil }).
					Protected()
			},
			check: func(t *testing.T, route *Route) {
				if route == nil {
					t.Fatal("Build() should not return nil")
					return
				}
				if route.Path != "/test" {
					t.Errorf("Build() path = %v, want '/test'", route.Path)
				}
				if route.Method != core.GET {
					t.Errorf("Build() method = %v, want GET", route.Method)
				}
				if !route.IsProtected {
					t.Error("Build() should preserve IsProtected")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRoute("/test")
			tt.setup(builder)
			route := builder.Build()
			tt.check(t, route)
		})
	}
}

func TestNewGroupRoute(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		check  func(t *testing.T, builder *GroupRouteBuilder)
	}{
		{
			name:   "should create group route builder with prefix",
			prefix: "/api",
			check: func(t *testing.T, builder *GroupRouteBuilder) {
				if builder == nil {
					t.Fatal("NewGroupRoute() should not return nil")
					return
				}
				if builder.group.Prefix != "/api" {
					t.Errorf("NewGroupRoute() prefix = %v, want '/api'", builder.group.Prefix)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewGroupRoute(tt.prefix)
			tt.check(t, builder)
		})
	}
}

func TestGroupRouteBuilder_Group(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*GroupRouteBuilder)
		check func(t *testing.T, group *RouteGroup)
	}{
		{
			name: "should add subgroup",
			setup: func(builder *GroupRouteBuilder) {
				subGroup := NewGroupRoute("/sub").Build()
				builder.Group(subGroup)
			},
			check: func(t *testing.T, group *RouteGroup) {
				if len(group.Groups) != 1 {
					t.Errorf("Groups length = %d, want 1", len(group.Groups))
				}
				if group.Groups[0].Prefix != "/sub" {
					t.Errorf("Subgroup prefix = %s, want '/sub'", group.Groups[0].Prefix)
				}
			},
		},
		{
			name: "should add multiple subgroups",
			setup: func(builder *GroupRouteBuilder) {
				sub1 := NewGroupRoute("/sub1").Build()
				sub2 := NewGroupRoute("/sub2").Build()
				builder.Groups(sub1, sub2)
			},
			check: func(t *testing.T, group *RouteGroup) {
				if len(group.Groups) != 2 {
					t.Errorf("Groups length = %d, want 2", len(group.Groups))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewGroupRoute("/api")
			tt.setup(builder)
			group := builder.Build()
			tt.check(t, group)
		})
	}
}
