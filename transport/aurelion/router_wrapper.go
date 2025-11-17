package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/router"
)

// RouteBuilderWrapper wraps router.RouteBuilder to accept aurelion types.
type RouteBuilderWrapper struct {
	*router.RouteBuilder
}

// Path wraps router.Path.
func (rb *RouteBuilderWrapper) Path(path string) *RouteBuilderWrapper {
	rb.RouteBuilder.Path(path)
	return rb
}

// Protected wraps router.Protected.
func (rb *RouteBuilderWrapper) Protected() *RouteBuilderWrapper {
	rb.RouteBuilder.Protected()
	return rb
}

// Permissions wraps router.Permissions.
func (rb *RouteBuilderWrapper) Permissions(permissions ...string) *RouteBuilderWrapper {
	rb.RouteBuilder.Permissions(permissions...)
	return rb
}

// Build wraps router.Build.
func (rb *RouteBuilderWrapper) Build() *router.Route {
	return rb.RouteBuilder.Build()
}

// Handler wraps router.Handler to accept aurelion.Handler.
func (rb *RouteBuilderWrapper) Handler(handler Handler) *RouteBuilderWrapper {
	if handler == nil {
		return rb
	}
	// Convert aurelion.Handler to router.Handler
	routerHandler := func(ctx router.ContextInterface) error {
		// Extract underlying context from adapter
		// The context passed here should already be a routerContextAdapter
		// We need to get the underlying aurelion.Context
		// Since routerContextAdapter is in router_adapter.go, we access it via the adapter
		underlyingCtx := extractContextFromRouterAdapter(ctx)
		return handler(underlyingCtx)
	}
	rb.RouteBuilder.Handler(routerHandler)
	return rb
}

// Middleware wraps router.Middleware to accept aurelion.Middleware.
func (rb *RouteBuilderWrapper) Middleware(middleware ...Middleware) *RouteBuilderWrapper {
	if len(middleware) == 0 {
		return rb
	}
	// Convert aurelion.Middleware to interface{} for the router builder
	routerMiddlewares := make([]interface{}, len(middleware))
	for i, m := range middleware {
		if m != nil {
			routerMiddlewares[i] = m
		}
	}
	rb.RouteBuilder.Middleware(routerMiddlewares...)
	return rb
}

// Method wraps router.Method to accept aurelion.Method.
func (rb *RouteBuilderWrapper) Method(method Method) *RouteBuilderWrapper {
	rb.RouteBuilder.Method(router.Method(method))
	return rb
}

// GET wraps router.GET.
func (rb *RouteBuilderWrapper) GET() *RouteBuilderWrapper {
	return rb.Method(MethodGet)
}

// POST wraps router.POST.
func (rb *RouteBuilderWrapper) POST() *RouteBuilderWrapper {
	return rb.Method(MethodPost)
}

// PUT wraps router.PUT.
func (rb *RouteBuilderWrapper) PUT() *RouteBuilderWrapper {
	return rb.Method(MethodPut)
}

// PATCH wraps router.PATCH.
func (rb *RouteBuilderWrapper) PATCH() *RouteBuilderWrapper {
	return rb.Method(MethodPatch)
}

// DELETE wraps router.DELETE.
func (rb *RouteBuilderWrapper) DELETE() *RouteBuilderWrapper {
	return rb.Method(MethodDelete)
}

// CORS wraps router.CORS to accept aurelion.CORSConfig.
func (rb *RouteBuilderWrapper) CORS(corsConfig *CORSConfig) *RouteBuilderWrapper {
	if corsConfig == nil {
		return rb
	}
	routerCORS := &router.CORSConfig{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		MaxAge:           corsConfig.MaxAge,
	}
	rb.RouteBuilder.CORS(routerCORS)
	return rb
}

// GroupRouteBuilderWrapper wraps router.GroupRouteBuilder to accept aurelion types.
type GroupRouteBuilderWrapper struct {
	*router.GroupRouteBuilder
}

// Protected wraps router.Protected.
func (grb *GroupRouteBuilderWrapper) Protected() *GroupRouteBuilderWrapper {
	grb.GroupRouteBuilder.Protected()
	return grb
}

// Route wraps router.Route to accept aurelion route types.
func (grb *GroupRouteBuilderWrapper) Route(route interface{}) *GroupRouteBuilderWrapper {
	// Convert various route types
	switch v := route.(type) {
	case *RouteBuilderWrapper:
		if v != nil {
			grb.GroupRouteBuilder.Route(v.Build())
		}
	case *router.RouteBuilder:
		if v != nil {
			grb.GroupRouteBuilder.Route(v.Build())
		}
	case *router.Route:
		grb.GroupRouteBuilder.Route(v)
	case router.Route:
		grb.GroupRouteBuilder.Route(&v)
	default:
		grb.GroupRouteBuilder.Route(route)
	}
	return grb
}

// Routes wraps router.Routes.
func (grb *GroupRouteBuilderWrapper) Routes(routes ...interface{}) *GroupRouteBuilderWrapper {
	grb.GroupRouteBuilder.Routes(routes...)
	return grb
}

// Build wraps router.Build.
func (grb *GroupRouteBuilderWrapper) Build() *router.GroupRoute {
	return grb.GroupRouteBuilder.Build()
}

// Middleware wraps router.Middleware to accept aurelion.Middleware.
func (grb *GroupRouteBuilderWrapper) Middleware(middleware ...Middleware) *GroupRouteBuilderWrapper {
	if len(middleware) == 0 {
		return grb
	}
	// Convert aurelion.Middleware to interface{} for the router builder
	routerMiddlewares := make([]interface{}, len(middleware))
	for i, m := range middleware {
		if m != nil {
			routerMiddlewares[i] = m
		}
	}
	grb.GroupRouteBuilder.Middleware(routerMiddlewares...)
	return grb
}

// NewRouteWrapper creates a new route builder wrapper that accepts aurelion types.
func NewRouteWrapper(path string) *RouteBuilderWrapper {
	return &RouteBuilderWrapper{RouteBuilder: router.NewRoute(path)}
}

// NewGroupRouteWrapper creates a new group route builder wrapper that accepts aurelion types.
func NewGroupRouteWrapper(prefix string) *GroupRouteBuilderWrapper {
	return &GroupRouteBuilderWrapper{GroupRouteBuilder: router.NewGroupRoute(prefix)}
}
