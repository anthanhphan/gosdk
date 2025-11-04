package aurelion

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Route represents a single HTTP route configuration
type Route struct {
	Path                string
	Method              Method
	Handler             Handler
	Middlewares         []Middleware
	RequiredPermissions []string
	IsProtected         bool
	CORS                *CORSConfig // Optional per-route CORS configuration
}

// GroupRoute represents a group of routes with a common prefix
type GroupRoute struct {
	Prefix      string
	Routes      []Route
	Middlewares []Middleware
	IsProtected bool
}

// RouteBuilder provides a fluent interface for building routes
type RouteBuilder struct {
	route *Route
}

// NewRoute creates a new route builder with the given path.
//
// Input:
//   - path: The route path (e.g., "/users")
//
// Output:
//   - *RouteBuilder: The route builder for fluent configuration
//
// Example:
//
//	route := aurelion.NewRoute("/users").
//		GET().
//		Handler(getUsersHandler)
func NewRoute(path string) *RouteBuilder {
	return &RouteBuilder{
		route: &Route{
			Path:                path,
			Middlewares:         make([]Middleware, 0),
			RequiredPermissions: make([]string, 0),
		},
	}
}

// Path sets the route path.
//
// Input:
//   - path: The route path (e.g., "/users")
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/old").Path("/new")
func (rb *RouteBuilder) Path(path string) *RouteBuilder {
	rb.route.Path = path
	return rb
}

// Method sets the HTTP method.
//
// Input:
//   - method: The HTTP method constant (GET, POST, PUT, etc.)
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users").Method(aurelion.GET)
func (rb *RouteBuilder) Method(method Method) *RouteBuilder {
	rb.route.Method = method
	return rb
}

// GET sets the HTTP method to GET.
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users").GET().Handler(getUsersHandler)
func (rb *RouteBuilder) GET() *RouteBuilder {
	return rb.Method(GET)
}

// POST sets the HTTP method to POST.
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users").POST().Handler(createUserHandler)
func (rb *RouteBuilder) POST() *RouteBuilder {
	return rb.Method(POST)
}

// PUT sets the HTTP method to PUT.
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users/:id").PUT().Handler(updateUserHandler)
func (rb *RouteBuilder) PUT() *RouteBuilder {
	return rb.Method(PUT)
}

// PATCH sets the HTTP method to PATCH.
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users/:id").PATCH().Handler(patchUserHandler)
func (rb *RouteBuilder) PATCH() *RouteBuilder {
	return rb.Method(PATCH)
}

// DELETE sets the HTTP method to DELETE.
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users/:id").DELETE().Handler(deleteUserHandler)
func (rb *RouteBuilder) DELETE() *RouteBuilder {
	return rb.Method(DELETE)
}

// Handler sets the route handler function.
//
// Input:
//   - handler: The function that handles requests to this route
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users").
//		GET().
//		Handler(func(ctx aurelion.Context) error {
//			return aurelion.OK(ctx, "Users list", users)
//		})
func (rb *RouteBuilder) Handler(handler Handler) *RouteBuilder {
	rb.route.Handler = handler
	return rb
}

// Middleware adds one or more middleware functions to the route.
//
// Input:
//   - middleware: One or more middleware functions to apply to this route
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/users").
//		GET().
//		Middleware(loggingMiddleware).
//		Handler(getUsersHandler)
func (rb *RouteBuilder) Middleware(middleware ...Middleware) *RouteBuilder {
	rb.route.Middlewares = append(rb.route.Middlewares, middleware...)
	return rb
}

// Protected marks the route as requiring authentication.
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/admin").
//		GET().
//		Protected().
//		Handler(adminHandler)
func (rb *RouteBuilder) Protected() *RouteBuilder {
	rb.route.IsProtected = true
	return rb
}

// Permissions sets the required permissions for this route.
// Automatically enables protection when permissions are set.
//
// Input:
//   - permissions: One or more permission strings required to access this route
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/admin/users").
//		GET().
//		Permissions("read:users", "admin").
//		Handler(listUsersHandler)
func (rb *RouteBuilder) Permissions(permissions ...string) *RouteBuilder {
	rb.route.RequiredPermissions = append(rb.route.RequiredPermissions, permissions...)
	rb.route.IsProtected = true // Auto-enable protection when permissions are set
	return rb
}

// CORS sets per-route CORS configuration.
//
// Input:
//   - corsConfig: The CORS configuration for this route
//
// Output:
//   - *RouteBuilder: The route builder for chaining
//
// Example:
//
//	route := aurelion.NewRoute("/api/users").
//		GET().
//		CORS(&aurelion.CORSConfig{
//			AllowOrigins: []string{"https://example.com"},
//			AllowMethods: []string{"GET", "OPTIONS"},
//			AllowHeaders: []string{"Content-Type"},
//		}).
//		Handler(getUsersHandler)
func (rb *RouteBuilder) CORS(corsConfig *CORSConfig) *RouteBuilder {
	rb.route.CORS = corsConfig
	return rb
}

// Build returns the constructed route.
//
// Output:
//   - *Route: The fully configured route
//
// Example:
//
//	route := aurelion.NewRoute("/users").
//		GET().
//		Handler(getUsersHandler).
//		Build()
func (rb *RouteBuilder) Build() *Route {
	return rb.route
}

// GroupRouteBuilder provides a fluent interface for building group routes
type GroupRouteBuilder struct {
	group *GroupRoute
}

// NewGroupRoute creates a new group route builder with the given prefix.
//
// Input:
//   - prefix: The common path prefix for all routes in this group (e.g., "/api/v1")
//
// Output:
//   - *GroupRouteBuilder: The group route builder for fluent configuration
//
// Example:
//
//	group := aurelion.NewGroupRoute("/api/v1").
//		Routes(
//			aurelion.NewRoute("/users").GET().Handler(getUsersHandler),
//		)
func NewGroupRoute(prefix string) *GroupRouteBuilder {
	return &GroupRouteBuilder{
		group: &GroupRoute{
			Prefix:      prefix,
			Routes:      make([]Route, 0),
			Middlewares: make([]Middleware, 0),
		},
	}
}

// Middleware adds one or more middleware functions to the group.
//
// Input:
//   - middleware: One or more middleware functions to apply to all routes in this group
//
// Output:
//   - *GroupRouteBuilder: The group route builder for chaining
//
// Example:
//
//	group := aurelion.NewGroupRoute("/api/v1").
//		Middleware(loggingMiddleware, authMiddleware)
func (grb *GroupRouteBuilder) Middleware(middleware ...Middleware) *GroupRouteBuilder {
	grb.group.Middlewares = append(grb.group.Middlewares, middleware...)
	return grb
}

// Protected marks all routes in the group as requiring authentication.
//
// Output:
//   - *GroupRouteBuilder: The group route builder for chaining
//
// Example:
//
//	group := aurelion.NewGroupRoute("/admin").
//		Protected().
//		Routes(...)
func (grb *GroupRouteBuilder) Protected() *GroupRouteBuilder {
	grb.group.IsProtected = true
	return grb
}

// Route adds a single route to the group.
//
// Input:
//   - route: The route to add (accepts Route, *Route, or *RouteBuilder)
//
// Output:
//   - *GroupRouteBuilder: The group route builder for chaining
//
// Example:
//
//	group := aurelion.NewGroupRoute("/api/v1").
//		Route(aurelion.NewRoute("/users").GET().Handler(getUsersHandler))
func (grb *GroupRouteBuilder) Route(route interface{}) *GroupRouteBuilder {
	r := convertToRoute(route)
	if r == nil {
		return grb
	}

	grb.group.Routes = append(grb.group.Routes, *r)
	return grb
}

// convertToRoute converts various route types to a *Route pointer.
// Supports Route, *Route, and *RouteBuilder types.
// This is a wrapper around convertToRouteType for backward compatibility.
// Internal use only - called automatically when routes are added.
//
// Input:
//   - route: The route interface to convert (RouteBuilder, *Route, or Route)
//
// Output:
//   - *Route: The converted route, or nil if conversion fails
func convertToRoute(route interface{}) *Route {
	return convertToRouteType(route)
}

// Routes adds multiple routes to the group.
//
// Input:
//   - routes: Variable number of routes to add (accepts Route, *Route, or *RouteBuilder)
//
// Output:
//   - *GroupRouteBuilder: The group route builder for chaining
//
// Example:
//
//	group := aurelion.NewGroupRoute("/api/v1").
//		Routes(
//			aurelion.NewRoute("/users").GET().Handler(getUsersHandler),
//			aurelion.NewRoute("/posts").GET().Handler(getPostsHandler),
//		)
func (grb *GroupRouteBuilder) Routes(routes ...interface{}) *GroupRouteBuilder {
	for _, route := range routes {
		grb.Route(route)
	}
	return grb
}

// Build returns the constructed group route.
//
// Output:
//   - *GroupRoute: The fully configured group route
//
// Example:
//
//	group := aurelion.NewGroupRoute("/api/v1").
//		Routes(...).
//		Build()
func (grb *GroupRouteBuilder) Build() *GroupRoute {
	return grb.group
}

// String returns a string representation of the route.
//
// Output:
//   - string: A string in the format "METHOD /path"
//
// Example:
//
//	route := aurelion.NewRoute("/users").GET()
//	fmt.Println(route.Build().String()) // Output: "GET /users"
func (r *Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Path)
}

// register registers the route with the fiber app.
// Converts the route's handler chain to fiber handlers and registers them.
// Internal use only - called automatically when routes are added to the server.
//
// Input:
//   - app: The fiber application instance
func (route *Route) register(app *fiber.App) {
	if route == nil || app == nil {
		return
	}
	handlers := route.buildHandlers()
	registerRoute(app, route.Method, route.Path, handlers)
}

// register registers the group and its routes with the fiber app.
// Creates a fiber group with the prefix and applies group middlewares.
// Then registers all routes in the group. Internal use only.
//
// Input:
//   - app: The fiber application instance
func (group *GroupRoute) register(app *fiber.App) {
	if group == nil || app == nil {
		return
	}
	// Convert group middlewares to fiber handlers
	fiberMiddlewares := make([]fiber.Handler, len(group.Middlewares))
	for i, m := range group.Middlewares {
		fiberMiddlewares[i] = middlewareToFiber(m)
	}

	g := app.Group(group.Prefix, fiberMiddlewares...)

	for i := range group.Routes {
		handlers := group.Routes[i].buildHandlers()
		// Empty path in group means route at group prefix, use "/"
		path := group.Routes[i].Path
		if path == "" {
			path = "/"
		}
		registerRoute(g, group.Routes[i].Method, path, handlers)
	}
}

// buildHandlers builds the handler chain for the route.
// Converts route middlewares and handler to fiber handlers.
// Returns a slice of fiber handlers in the correct order: middlewares first, then handler.
// Internal use only - called automatically during route registration.
//
// Output:
//   - []fiber.Handler: The handler chain ready for fiber registration
func (route *Route) buildHandlers() []fiber.Handler {
	if route == nil {
		return nil
	}

	// Pre-allocate with exact size for better performance
	handlers := make([]fiber.Handler, 0, len(route.Middlewares)+1)

	// Convert middlewares to fiber handlers
	for _, m := range route.Middlewares {
		if m != nil {
			handlers = append(handlers, middlewareToFiber(m))
		}
	}

	// Convert handler to fiber handler
	if route.Handler != nil {
		handlers = append(handlers, handlerToFiber(route.Handler))
	}

	return handlers
}

// registerRoute registers a route with the appropriate HTTP method (internal use).
// It maps our Method type to the corresponding fiber router method.
//
// Input:
//   - router: The fiber router to register the route on
//   - method: The HTTP method (GET, POST, etc.)
//   - path: The route path
//   - handlers: The handler chain (middlewares + handler)
func registerRoute(router fiber.Router, method Method, path string, handlers []fiber.Handler) {
	if router == nil || len(handlers) == 0 {
		return
	}

	switch method {
	case GET:
		router.Get(path, handlers...)
	case POST:
		router.Post(path, handlers...)
	case PUT:
		router.Put(path, handlers...)
	case PATCH:
		router.Patch(path, handlers...)
	case DELETE:
		router.Delete(path, handlers...)
	case HEAD:
		router.Head(path, handlers...)
	case OPTIONS:
		router.Options(path, handlers...)
	}
}

// Clone creates a deep copy of the route.
//
// Output:
//   - *Route: A deep copy of the route, or nil if the route is nil
//
// Example:
//
//	route1 := aurelion.NewRoute("/users").GET()
//	route2 := route1.Build().Clone()
func (route *Route) Clone() *Route {
	if route == nil {
		return nil
	}

	clone := &Route{
		Path:                route.Path,
		Method:              route.Method,
		Handler:             route.Handler,
		IsProtected:         route.IsProtected,
		Middlewares:         make([]Middleware, len(route.Middlewares)),
		RequiredPermissions: make([]string, len(route.RequiredPermissions)),
	}
	copy(clone.Middlewares, route.Middlewares)
	copy(clone.RequiredPermissions, route.RequiredPermissions)

	// Deep copy CORS config if present
	if route.CORS != nil {
		clone.CORS = &CORSConfig{
			AllowOrigins:     make([]string, len(route.CORS.AllowOrigins)),
			AllowMethods:     make([]string, len(route.CORS.AllowMethods)),
			AllowHeaders:     make([]string, len(route.CORS.AllowHeaders)),
			AllowCredentials: route.CORS.AllowCredentials,
			ExposeHeaders:    make([]string, len(route.CORS.ExposeHeaders)),
			MaxAge:           route.CORS.MaxAge,
		}
		copy(clone.CORS.AllowOrigins, route.CORS.AllowOrigins)
		copy(clone.CORS.AllowMethods, route.CORS.AllowMethods)
		copy(clone.CORS.AllowHeaders, route.CORS.AllowHeaders)
		copy(clone.CORS.ExposeHeaders, route.CORS.ExposeHeaders)
	}

	return clone
}

// handlerToFiber converts our Handler type to fiber.Handler.
// It wraps the fiber context in our custom Context interface.
// Internal use only - called automatically during route registration.
//
// Input:
//   - handler: The aurelion Handler function to convert
//
// Output:
//   - fiber.Handler: The fiber-compatible handler function
func handlerToFiber(handler Handler) fiber.Handler {
	if handler == nil {
		// Return a no-op handler if nil (defensive programming)
		return func(c *fiber.Ctx) error {
			return nil
		}
	}

	return func(c *fiber.Ctx) error {
		ctx := newContext(c)
		if ctx == nil {
			return fmt.Errorf("failed to create context wrapper")
		}
		return handler(ctx)
	}
}

// middlewareToFiber converts our Middleware type to fiber.Handler.
// It wraps the fiber context in our custom Context interface.
// Internal use only - called automatically during route registration.
//
// Input:
//   - middleware: The aurelion Middleware function to convert
//
// Output:
//   - fiber.Handler: The fiber-compatible middleware function
func middlewareToFiber(middleware Middleware) fiber.Handler {
	if middleware == nil {
		// Return a pass-through middleware if nil (defensive programming)
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		ctx := newContext(c)
		if ctx == nil {
			return fmt.Errorf("failed to create context wrapper")
		}
		return middleware(ctx)
	}
}
