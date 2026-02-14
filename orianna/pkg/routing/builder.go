// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// Route Builder

// RouteBuilder provides a fluent interface for building routes
type RouteBuilder struct {
	route *Route
}

// NewRoute creates a new route builder with the given path.
func NewRoute(path string) *RouteBuilder {
	return &RouteBuilder{
		route: &Route{
			Path:                path,
			Middlewares:         make([]core.Middleware, 0),
			RequiredPermissions: make([]string, 0),
		},
	}
}

// Path sets the route path
func (rb *RouteBuilder) Path(path string) *RouteBuilder {
	rb.route.Path = path
	return rb
}

// Method sets the HTTP method
func (rb *RouteBuilder) Method(method core.Method) *RouteBuilder {
	rb.route.Method = method
	return rb
}

// GET sets the HTTP method to GET
func (rb *RouteBuilder) GET() *RouteBuilder {
	return rb.Method(core.GET)
}

// POST sets the HTTP method to POST
func (rb *RouteBuilder) POST() *RouteBuilder {
	return rb.Method(core.POST)
}

// PUT sets the HTTP method to PUT
func (rb *RouteBuilder) PUT() *RouteBuilder {
	return rb.Method(core.PUT)
}

// PATCH sets the HTTP method to PATCH
func (rb *RouteBuilder) PATCH() *RouteBuilder {
	return rb.Method(core.PATCH)
}

// DELETE sets the HTTP method to DELETE
func (rb *RouteBuilder) DELETE() *RouteBuilder {
	return rb.Method(core.DELETE)
}

// Methods sets multiple HTTP methods for the route
func (rb *RouteBuilder) Methods(methods ...core.Method) *RouteBuilder {
	rb.route.Methods = append(rb.route.Methods, methods...)
	return rb
}

// Handler sets the route handler function
func (rb *RouteBuilder) Handler(handler core.Handler) *RouteBuilder {
	rb.route.Handler = handler
	return rb
}

// Middleware adds one or more middleware functions to the route
func (rb *RouteBuilder) Middleware(middleware ...core.Middleware) *RouteBuilder {
	rb.route.Middlewares = append(rb.route.Middlewares, middleware...)
	return rb
}

// Protected marks the route as requiring authentication
func (rb *RouteBuilder) Protected() *RouteBuilder {
	rb.route.IsProtected = true
	return rb
}

// Permissions sets the required permissions for this route
func (rb *RouteBuilder) Permissions(permissions ...string) *RouteBuilder {
	rb.route.RequiredPermissions = append(rb.route.RequiredPermissions, permissions...)
	rb.route.IsProtected = true
	return rb
}

// CORS sets per-route CORS configuration
func (rb *RouteBuilder) CORS(corsConfig *configuration.CORSConfig) *RouteBuilder {
	if corsConfig != nil {
		rb.route.CORS = corsConfig
	}
	return rb
}

// Build returns the constructed route
func (rb *RouteBuilder) Build() *Route {
	return rb.route
}

// Group Route Builder

// GroupRouteBuilder provides a fluent interface for building group routes
type GroupRouteBuilder struct {
	group *RouteGroup
}

// NewGroupRoute creates a new group route builder with the given prefix
func NewGroupRoute(prefix string) *GroupRouteBuilder {
	return &GroupRouteBuilder{
		group: &RouteGroup{
			Prefix:      prefix,
			Routes:      make([]Route, 0),
			Groups:      make([]RouteGroup, 0),
			Middlewares: make([]core.Middleware, 0),
		},
	}
}

// Middleware adds one or more middleware functions to the group
func (grb *GroupRouteBuilder) Middleware(middleware ...core.Middleware) *GroupRouteBuilder {
	grb.group.Middlewares = append(grb.group.Middlewares, middleware...)
	return grb
}

// Protected marks all routes in the group as requiring authentication
func (grb *GroupRouteBuilder) Protected() *GroupRouteBuilder {
	grb.group.IsProtected = true
	return grb
}

// Route adds a single route to the group
func (grb *GroupRouteBuilder) Route(route *Route) *GroupRouteBuilder {
	if route != nil {
		grb.group.Routes = append(grb.group.Routes, *route)
	}
	return grb
}

// Routes adds multiple routes to the group
func (grb *GroupRouteBuilder) Routes(routes ...*Route) *GroupRouteBuilder {
	for _, route := range routes {
		grb.Route(route)
	}
	return grb
}

// Group adds a subgroup to the group
func (grb *GroupRouteBuilder) Group(group *RouteGroup) *GroupRouteBuilder {
	if group != nil {
		grb.group.Groups = append(grb.group.Groups, *group)
	}
	return grb
}

// Groups adds multiple subgroups to the group
func (grb *GroupRouteBuilder) Groups(groups ...*RouteGroup) *GroupRouteBuilder {
	for _, group := range groups {
		grb.Group(group)
	}
	return grb
}

// Build returns the constructed group route
func (grb *GroupRouteBuilder) Build() *RouteGroup {
	return grb.group
}

// GET adds a GET route to the group
func (grb *GroupRouteBuilder) GET(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.GET, path, handler, middleware...)
}

// POST adds a POST route to the group
func (grb *GroupRouteBuilder) POST(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.POST, path, handler, middleware...)
}

// PUT adds a PUT route to the group
func (grb *GroupRouteBuilder) PUT(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.PUT, path, handler, middleware...)
}

// PATCH adds a PATCH route to the group
func (grb *GroupRouteBuilder) PATCH(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.PATCH, path, handler, middleware...)
}

// DELETE adds a DELETE route to the group
func (grb *GroupRouteBuilder) DELETE(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.DELETE, path, handler, middleware...)
}

// HEAD adds a HEAD route to the group
func (grb *GroupRouteBuilder) HEAD(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.HEAD, path, handler, middleware...)
}

// OPTIONS adds an OPTIONS route to the group
func (grb *GroupRouteBuilder) OPTIONS(path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	return grb.addRoute(core.OPTIONS, path, handler, middleware...)
}

// addRoute is a helper to create and add a route
func (grb *GroupRouteBuilder) addRoute(method core.Method, path string, handler core.Handler, middleware ...core.Middleware) *GroupRouteBuilder {
	route := Route{
		Path:        path,
		Method:      method,
		Handler:     handler,
		Middlewares: middleware,
	}
	grb.group.Routes = append(grb.group.Routes, route)
	return grb
}
