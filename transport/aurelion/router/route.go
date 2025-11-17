package router

import (
	"context"
	"fmt"
	"reflect"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/gofiber/fiber/v2"
)

const (
	MaxRoutePathLength       = 1024
	MaxRouteHandlersPerRoute = 50
)

// Route represents a single HTTP route configuration.
type Route struct {
	Path                string
	Method              Method
	Handler             Handler
	Middlewares         []Middleware
	RequiredPermissions []string
	IsProtected         bool
	CORS                *CORSConfig
}

// GroupRoute represents a group of routes with a common prefix.
type GroupRoute struct {
	Prefix      string
	Routes      []Route
	Middlewares []Middleware
	IsProtected bool
}

// RouteBuilder provides a fluent interface for building routes.
type RouteBuilder struct {
	route *Route
}

// NewRoute creates a new route builder with the given path.
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
func (rb *RouteBuilder) Path(path string) *RouteBuilder {
	rb.route.Path = path
	return rb
}

// Method sets the HTTP method.
func (rb *RouteBuilder) Method(method interface{}) *RouteBuilder {
	if method == nil {
		return rb
	}
	// If it's already a router.Method, use it directly
	if routerMethod, ok := method.(Method); ok {
		rb.route.Method = routerMethod
		return rb
	}
	// If it's a core.Method, convert it
	if coreMethod, ok := method.(core.Method); ok {
		rb.route.Method = Method(coreMethod)
		return rb
	}
	// If it's a string, convert it
	if strMethod, ok := method.(string); ok {
		rb.route.Method = Method(strMethod)
		return rb
	}
	return rb
}

// GET sets the HTTP method to GET.
func (rb *RouteBuilder) GET() *RouteBuilder {
	return rb.Method(MethodGet)
}

// POST sets the HTTP method to POST.
func (rb *RouteBuilder) POST() *RouteBuilder {
	return rb.Method(MethodPost)
}

// PUT sets the HTTP method to PUT.
func (rb *RouteBuilder) PUT() *RouteBuilder {
	return rb.Method(MethodPut)
}

// PATCH sets the HTTP method to PATCH.
func (rb *RouteBuilder) PATCH() *RouteBuilder {
	return rb.Method(MethodPatch)
}

// DELETE sets the HTTP method to DELETE.
func (rb *RouteBuilder) DELETE() *RouteBuilder {
	return rb.Method(MethodDelete)
}

// Handler sets the route handler function.
func (rb *RouteBuilder) Handler(handler interface{}) *RouteBuilder {
	if handler == nil {
		return rb
	}
	// If it's already a router.Handler, use it directly
	if routerHandler, ok := handler.(Handler); ok {
		rb.route.Handler = routerHandler
		return rb
	}
	// If it's a core.Handler, convert it
	if coreHandler, ok := handler.(core.Handler); ok {
		rb.route.Handler = func(ctx ContextInterface) error {
			// Convert router.ContextInterface to core.Context
			coreCtx := AdaptRouterContextToCore(ctx)
			return coreHandler(coreCtx)
		}
		return rb
	}
	// If it's a function with the right signature, try to convert it using reflection
	// This handles cases where the function type doesn't match exactly (e.g., function literals)
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Func {
		handlerType := handlerValue.Type()
		// Check if it's func(ContextInterface) error (router.Handler signature)
		// Get the type of Handler by checking a sample function
		var sampleHandler Handler = func(ctx ContextInterface) error { return nil }
		routerHandlerType := reflect.TypeOf(sampleHandler)
		if handlerType.AssignableTo(routerHandlerType) {
			// Convert function literal to Handler type
			rb.route.Handler = Handler(handlerValue.Interface().(func(ContextInterface) error))
			return rb
		}
		// Check if it's func(core.Context) error
		if handlerType.NumIn() == 1 && handlerType.NumOut() == 1 {
			// Check if return type is error
			if handlerType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				// Check if the input parameter accepts core.Context
				firstParamType := handlerType.In(0)
				coreContextType := reflect.TypeOf((*core.Context)(nil)).Elem()
				// For interface types, check if parameter type is the interface
				if firstParamType.Kind() == reflect.Interface && firstParamType == coreContextType {
					// Try to convert - create a wrapper function
					rb.route.Handler = func(ctx ContextInterface) error {
						coreCtx := AdaptRouterContextToCore(ctx)
						// Call the handler using reflection
						results := handlerValue.Call([]reflect.Value{reflect.ValueOf(coreCtx)})
						if len(results) > 0 && !results[0].IsNil() {
							return results[0].Interface().(error)
						}
						return nil
					}
					return rb
				}
			}
		}
	}
	// If we can't convert it, return without setting (will be caught by validation)
	return rb
}

// Middleware adds one or more middleware functions to the route.
func (rb *RouteBuilder) Middleware(middleware ...interface{}) *RouteBuilder {
	for _, m := range middleware {
		if m == nil {
			continue
		}
		// If it's already a router.Middleware, use it directly
		if routerMw, ok := m.(Middleware); ok {
			rb.route.Middlewares = append(rb.route.Middlewares, routerMw)
			continue
		}
		// If it's a core.Middleware, convert it
		if coreMw, ok := m.(core.Middleware); ok {
			routerMw := func(ctx ContextInterface) error {
				// Convert router.ContextInterface to core.Context
				coreCtx := AdaptRouterContextToCore(ctx)
				return coreMw(coreCtx)
			}
			rb.route.Middlewares = append(rb.route.Middlewares, routerMw)
			continue
		}
		// If it's a function with the right signature, try to convert it using reflection
		// This handles cases where the function type doesn't match exactly (e.g., function literals)
		mwValue := reflect.ValueOf(m)
		if mwValue.Kind() == reflect.Func {
			mwType := mwValue.Type()
			// Check if it's func(ContextInterface) error (router.Middleware signature)
			var sampleMiddleware Middleware = func(ctx ContextInterface) error { return nil }
			routerMiddlewareType := reflect.TypeOf(sampleMiddleware)
			if mwType.AssignableTo(routerMiddlewareType) {
				// Convert function literal to Middleware type
				rb.route.Middlewares = append(rb.route.Middlewares, Middleware(mwValue.Interface().(func(ContextInterface) error)))
				continue
			}
			// Check if it's func(core.Context) error
			if mwType.NumIn() == 1 && mwType.NumOut() == 1 {
				// Check if return type is error
				if mwType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
					// Check if the input parameter accepts core.Context
					firstParamType := mwType.In(0)
					coreContextType := reflect.TypeOf((*core.Context)(nil)).Elem()
					// For interface types, check if parameter type is the interface
					if firstParamType.Kind() == reflect.Interface && firstParamType == coreContextType {
						// Try to convert - create a wrapper function
						routerMw := func(ctx ContextInterface) error {
							coreCtx := AdaptRouterContextToCore(ctx)
							// Call the middleware using reflection
							results := mwValue.Call([]reflect.Value{reflect.ValueOf(coreCtx)})
							if len(results) > 0 && !results[0].IsNil() {
								return results[0].Interface().(error)
							}
							return nil
						}
						rb.route.Middlewares = append(rb.route.Middlewares, routerMw)
					}
				}
			}
		}
	}
	return rb
}

// Protected marks the route as requiring authentication.
func (rb *RouteBuilder) Protected() *RouteBuilder {
	rb.route.IsProtected = true
	return rb
}

// Permissions sets the required permissions for this route.
func (rb *RouteBuilder) Permissions(permissions ...string) *RouteBuilder {
	rb.route.RequiredPermissions = append(rb.route.RequiredPermissions, permissions...)
	rb.route.IsProtected = true
	return rb
}

// CORS sets per-route CORS configuration.
func (rb *RouteBuilder) CORS(corsConfig interface{}) *RouteBuilder {
	if corsConfig == nil {
		return rb
	}
	// If it's already a router.CORSConfig, use it directly
	if routerCORS, ok := corsConfig.(*CORSConfig); ok {
		rb.route.CORS = routerCORS
		return rb
	}
	// Try to convert from other CORS config types
	if configCORS, ok := corsConfig.(interface {
		GetAllowOrigins() []string
		GetAllowMethods() []string
		GetAllowHeaders() []string
		GetAllowCredentials() bool
		GetExposeHeaders() []string
		GetMaxAge() int
	}); ok {
		rb.route.CORS = &CORSConfig{
			AllowOrigins:     configCORS.GetAllowOrigins(),
			AllowMethods:     configCORS.GetAllowMethods(),
			AllowHeaders:     configCORS.GetAllowHeaders(),
			AllowCredentials: configCORS.GetAllowCredentials(),
			ExposeHeaders:    configCORS.GetExposeHeaders(),
			MaxAge:           configCORS.GetMaxAge(),
		}
		return rb
	}
	// Try to convert using reflection-like approach with struct fields
	// This is a fallback for config.CORSConfig and middleware.CORSConfig
	if corsMap, ok := corsConfig.(map[string]interface{}); ok {
		routerCORS := &CORSConfig{}
		if v, ok := corsMap["AllowOrigins"].([]string); ok {
			routerCORS.AllowOrigins = v
		}
		if v, ok := corsMap["AllowMethods"].([]string); ok {
			routerCORS.AllowMethods = v
		}
		if v, ok := corsMap["AllowHeaders"].([]string); ok {
			routerCORS.AllowHeaders = v
		}
		if v, ok := corsMap["AllowCredentials"].(bool); ok {
			routerCORS.AllowCredentials = v
		}
		if v, ok := corsMap["ExposeHeaders"].([]string); ok {
			routerCORS.ExposeHeaders = v
		}
		if v, ok := corsMap["MaxAge"].(int); ok {
			routerCORS.MaxAge = v
		}
		rb.route.CORS = routerCORS
		return rb
	}
	return rb
}

// Build returns the constructed route.
func (rb *RouteBuilder) Build() *Route {
	return rb.route
}

// GroupRouteBuilder provides a fluent interface for building group routes.
type GroupRouteBuilder struct {
	group *GroupRoute
}

// NewGroupRoute creates a new group route builder with the given prefix.
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
func (grb *GroupRouteBuilder) Middleware(middleware ...interface{}) *GroupRouteBuilder {
	for _, m := range middleware {
		if m == nil {
			continue
		}
		// If it's already a router.Middleware, use it directly
		if routerMw, ok := m.(Middleware); ok {
			grb.group.Middlewares = append(grb.group.Middlewares, routerMw)
			continue
		}
		// If it's a core.Middleware, convert it
		if coreMw, ok := m.(core.Middleware); ok {
			routerMw := func(ctx ContextInterface) error {
				// Convert router.ContextInterface to core.Context
				coreCtx := AdaptRouterContextToCore(ctx)
				return coreMw(coreCtx)
			}
			grb.group.Middlewares = append(grb.group.Middlewares, routerMw)
			continue
		}
		// If it's a function with the right signature, try to convert it using reflection
		// This handles cases where the function type doesn't match exactly (e.g., function literals)
		mwValue := reflect.ValueOf(m)
		if mwValue.Kind() == reflect.Func {
			mwType := mwValue.Type()
			// Check if it's func(ContextInterface) error (router.Middleware signature)
			var sampleMiddleware Middleware = func(ctx ContextInterface) error { return nil }
			routerMiddlewareType := reflect.TypeOf(sampleMiddleware)
			if mwType.AssignableTo(routerMiddlewareType) {
				// Convert function literal to Middleware type
				grb.group.Middlewares = append(grb.group.Middlewares, Middleware(mwValue.Interface().(func(ContextInterface) error)))
				continue
			}
			// Check if it's func(core.Context) error
			if mwType.NumIn() == 1 && mwType.NumOut() == 1 {
				// Check if return type is error
				if mwType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
					// Check if the input parameter accepts core.Context
					firstParamType := mwType.In(0)
					coreContextType := reflect.TypeOf((*core.Context)(nil)).Elem()
					// For interface types, check if parameter type is the interface
					if firstParamType.Kind() == reflect.Interface && firstParamType == coreContextType {
						// Try to convert - create a wrapper function
						routerMw := func(ctx ContextInterface) error {
							coreCtx := AdaptRouterContextToCore(ctx)
							// Call the middleware using reflection
							results := mwValue.Call([]reflect.Value{reflect.ValueOf(coreCtx)})
							if len(results) > 0 && !results[0].IsNil() {
								return results[0].Interface().(error)
							}
							return nil
						}
						grb.group.Middlewares = append(grb.group.Middlewares, routerMw)
					}
				}
			}
		}
	}
	return grb
}

// Protected marks all routes in the group as requiring authentication.
func (grb *GroupRouteBuilder) Protected() *GroupRouteBuilder {
	grb.group.IsProtected = true
	return grb
}

// Route adds a single route to the group.
func (grb *GroupRouteBuilder) Route(route interface{}) *GroupRouteBuilder {
	r := ToRoute(route)
	if r == nil {
		return grb
	}
	grb.group.Routes = append(grb.group.Routes, *r)
	return grb
}

func (grb *GroupRouteBuilder) Routes(routes ...interface{}) *GroupRouteBuilder {
	for _, route := range routes {
		grb.Route(route)
	}
	return grb
}

// Build returns the constructed group route.
func (grb *GroupRouteBuilder) Build() *GroupRoute {
	return grb.group
}

// String returns a string representation of the route.
func (r *Route) String() string {
	return fmt.Sprintf("%s %s", string(r.Method), r.Path)
}

// Register registers the route with the fiber app.
// Note: This method is kept for compatibility but handlers conversion
// should be done by the server using buildRouteHandlers.
func (route *Route) Register(app *fiber.App) {
	if route == nil || app == nil {
		return
	}
	// Handlers will be converted by the server
	handlers := route.buildHandlers()
	registerRoute(app, route.Method, route.Path, handlers)
}

// Register registers the group and its routes with the fiber app.
// Note: This method is kept for compatibility but handlers conversion
// should be done by the server using buildRouteHandlers and buildGroupMiddlewares.
func (group *GroupRoute) Register(app *fiber.App) {
	if group == nil || app == nil {
		return
	}
	// Convert group middlewares to fiber handlers
	fiberMiddlewares := buildGroupMiddlewares(group)

	g := app.Group(group.Prefix, fiberMiddlewares...)

	for i := range group.Routes {
		handlers := group.Routes[i].buildHandlers()
		path := group.Routes[i].Path
		if path == "" {
			path = "/"
		}
		registerRoute(g, group.Routes[i].Method, path, handlers)
	}
}

// buildGroupMiddlewares converts group middlewares to fiber handlers.
func buildGroupMiddlewares(group *GroupRoute) []fiber.Handler {
	if group == nil {
		return nil
	}
	fiberMiddlewares := make([]fiber.Handler, 0, len(group.Middlewares))
	for _, m := range group.Middlewares {
		if m != nil {
			fiberMiddlewares = append(fiberMiddlewares, convertMiddlewareToFiber(m))
		}
	}
	return fiberMiddlewares
}

// Clone creates a deep copy of the route.
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

	if route.CORS != nil {
		clone.CORS = &CORSConfig{
			AllowOrigins:     append([]string{}, route.CORS.AllowOrigins...),
			AllowMethods:     append([]string{}, route.CORS.AllowMethods...),
			AllowHeaders:     append([]string{}, route.CORS.AllowHeaders...),
			AllowCredentials: route.CORS.AllowCredentials,
			ExposeHeaders:    append([]string{}, route.CORS.ExposeHeaders...),
			MaxAge:           route.CORS.MaxAge,
		}
	}
	return clone
}

func (route *Route) buildHandlers() []fiber.Handler {
	if route == nil {
		return nil
	}
	handlers := make([]fiber.Handler, 0, len(route.Middlewares)+1)

	// Convert middlewares
	for _, m := range route.Middlewares {
		if m != nil {
			handlers = append(handlers, convertMiddlewareToFiber(m))
		}
	}

	// Convert handler
	if route.Handler != nil {
		handlers = append(handlers, convertHandlerToFiber(route.Handler))
	}

	return handlers
}

// convertHandlerToFiber converts router.Handler to fiber.Handler.
func convertHandlerToFiber(handler Handler) fiber.Handler {
	if handler == nil {
		return nil
	}
	return func(c *fiber.Ctx) error {
		ctx := adaptFiberCtxToContextInterface(c)
		return handler(ctx)
	}
}

// convertMiddlewareToFiber converts router.Middleware to fiber.Handler.
func convertMiddlewareToFiber(middleware Middleware) fiber.Handler {
	if middleware == nil {
		return nil
	}
	return func(c *fiber.Ctx) error {
		ctx := adaptFiberCtxToContextInterface(c)
		return middleware(ctx)
	}
}

// fiberContextAdapter adapts fiber.Ctx to ContextInterface.
type fiberContextAdapter struct {
	c *fiber.Ctx
}

func (a *fiberContextAdapter) Status(status int) ContextInterface {
	a.c.Status(status)
	return a
}

func (a *fiberContextAdapter) JSON(data interface{}) error {
	return a.c.JSON(data)
}

func (a *fiberContextAdapter) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		a.c.Locals(key, value[0])
		return value[0]
	}
	return a.c.Locals(key)
}

func (a *fiberContextAdapter) GetAllLocals() map[string]interface{} {
	result := make(map[string]interface{})
	a.c.Context().VisitUserValues(func(key []byte, value interface{}) {
		result[string(key)] = value
	})
	return result
}

func (a *fiberContextAdapter) Next() error {
	return a.c.Next()
}

func (a *fiberContextAdapter) Method() string {
	return a.c.Method()
}

func (a *fiberContextAdapter) Path() string {
	return a.c.Path()
}

func (a *fiberContextAdapter) Params(key string, defaultValue ...string) string {
	return a.c.Params(key, defaultValue...)
}

func (a *fiberContextAdapter) AllParams() map[string]string {
	return a.c.AllParams()
}

func (a *fiberContextAdapter) ParamsParser(out interface{}) error {
	return a.c.ParamsParser(out)
}

func (a *fiberContextAdapter) Query(key string, defaultValue ...string) string {
	return a.c.Query(key, defaultValue...)
}

func (a *fiberContextAdapter) AllQueries() map[string]string {
	return a.c.Queries()
}

func (a *fiberContextAdapter) QueryParser(out interface{}) error {
	return a.c.QueryParser(out)
}

func (a *fiberContextAdapter) Body() []byte {
	return a.c.Body()
}

func (a *fiberContextAdapter) BodyParser(out interface{}) error {
	return a.c.BodyParser(out)
}

// adaptFiberCtxToContextInterface adapts fiber.Ctx to ContextInterface.
func adaptFiberCtxToContextInterface(c *fiber.Ctx) ContextInterface {
	if c == nil {
		return nil
	}
	return &fiberContextAdapter{c: c}
}

// extractFiberCtxFromContextInterface extracts the underlying fiber.Ctx from a ContextInterface.
// This is used by routerToCoreAdapter to access fiber-specific methods like SendString.
func extractFiberCtxFromContextInterface(ctx ContextInterface) *fiber.Ctx {
	if ctx == nil {
		return nil
	}
	// Check if it's a fiberContextAdapter
	if adapter, ok := ctx.(*fiberContextAdapter); ok {
		return adapter.c
	}
	return nil
}

// Handlers exposes the prepared handler chain for the route.
func (route *Route) Handlers() []fiber.Handler {
	return route.buildHandlers()
}

func registerRoute(router fiber.Router, method Method, path string, handlers []fiber.Handler) {
	if router == nil || len(handlers) == 0 {
		return
	}

	switch method {
	case MethodGet:
		router.Get(path, handlers...)
	case MethodPost:
		router.Post(path, handlers...)
	case MethodPut:
		router.Put(path, handlers...)
	case MethodPatch:
		router.Patch(path, handlers...)
	case MethodDelete:
		router.Delete(path, handlers...)
	case MethodHead:
		router.Head(path, handlers...)
	case MethodOptions:
		router.Options(path, handlers...)
	}
}

// ToRoute attempts to convert various route types to a *Route pointer.
func ToRoute(route interface{}) *Route {
	switch v := route.(type) {
	case *RouteBuilder:
		if v == nil {
			return nil
		}
		return v.Build()
	case *Route:
		return v
	case Route:
		return &v
	default:
		return nil
	}
}

// ToGroupRoute attempts to convert various group route types to a *GroupRoute pointer.
func ToGroupRoute(group interface{}) *GroupRoute {
	switch v := group.(type) {
	case *GroupRouteBuilder:
		if v == nil {
			return nil
		}
		return v.Build()
	case *GroupRoute:
		return v
	case GroupRoute:
		return &v
	default:
		return nil
	}
}

// AdaptRouterContextToCore adapts router.ContextInterface to core.Context.
// This is a simplified adapter that only implements the methods needed by core.Handler and core.Middleware.
func AdaptRouterContextToCore(ctx ContextInterface) core.Context {
	if ctx == nil {
		return nil
	}
	// Note: We can't use direct type assertion because router.ContextInterface and core.Context
	// have incompatible Status method signatures. Instead, we check if it's a core.Context
	// by attempting to access it through a helper that checks for the underlying type.
	// For now, always create an adapter since router.ContextInterface is a subset of core.Context
	return &routerToCoreAdapter{ctx: ctx}
}

// routerToCoreAdapter adapts router.ContextInterface to core.Context.
type routerToCoreAdapter struct {
	ctx ContextInterface
}

func (a *routerToCoreAdapter) Method() string { return a.ctx.Method() }
func (a *routerToCoreAdapter) Path() string   { return a.ctx.Path() }

// Methods not in router.ContextInterface - return empty/default values
func (a *routerToCoreAdapter) OriginalURL() string { return "" }
func (a *routerToCoreAdapter) BaseURL() string     { return "" }
func (a *routerToCoreAdapter) Protocol() string    { return "" }
func (a *routerToCoreAdapter) Hostname() string    { return "" }
func (a *routerToCoreAdapter) IP() string          { return "" }
func (a *routerToCoreAdapter) Secure() bool        { return false }

func (a *routerToCoreAdapter) Get(key string, defaultValue ...string) string {
	// Not available in router.ContextInterface, return default
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}
func (a *routerToCoreAdapter) Set(key, value string)                 {}
func (a *routerToCoreAdapter) Append(field string, values ...string) {}

func (a *routerToCoreAdapter) Params(key string, defaultValue ...string) string {
	return a.ctx.Params(key, defaultValue...)
}
func (a *routerToCoreAdapter) AllParams() map[string]string { return a.ctx.AllParams() }
func (a *routerToCoreAdapter) ParamsParser(out interface{}) error {
	return a.ctx.ParamsParser(out)
}

func (a *routerToCoreAdapter) Query(key string, defaultValue ...string) string {
	return a.ctx.Query(key, defaultValue...)
}
func (a *routerToCoreAdapter) AllQueries() map[string]string { return a.ctx.AllQueries() }
func (a *routerToCoreAdapter) QueryParser(out interface{}) error {
	// QueryParser is available in router.ContextInterface
	return a.ctx.QueryParser(out)
}

func (a *routerToCoreAdapter) Body() []byte { return a.ctx.Body() }
func (a *routerToCoreAdapter) BodyParser(out interface{}) error {
	return a.ctx.BodyParser(out)
}

func (a *routerToCoreAdapter) Cookies(key string, defaultValue ...string) string {
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}
func (a *routerToCoreAdapter) Cookie(cookie *core.Cookie) {}
func (a *routerToCoreAdapter) ClearCookie(key ...string)  {}

func (a *routerToCoreAdapter) Status(status int) core.Context {
	a.ctx.Status(status)
	return a
}
func (a *routerToCoreAdapter) JSON(data interface{}) error { return a.ctx.JSON(data) }
func (a *routerToCoreAdapter) XML(data interface{}) error  { return fmt.Errorf("XML not supported") }
func (a *routerToCoreAdapter) SendString(s string) error {
	// Try to extract fiber.Ctx from the underlying context
	if fiberCtx := extractFiberCtxFromContextInterface(a.ctx); fiberCtx != nil {
		return fiberCtx.SendString(s)
	}
	return fmt.Errorf("SendString not supported: cannot access underlying fiber context")
}
func (a *routerToCoreAdapter) SendBytes(b []byte) error {
	// Try to extract fiber.Ctx from the underlying context
	if fiberCtx := extractFiberCtxFromContextInterface(a.ctx); fiberCtx != nil {
		return fiberCtx.Send(b)
	}
	return fmt.Errorf("SendBytes not supported: cannot access underlying fiber context")
}
func (a *routerToCoreAdapter) Redirect(location string, status ...int) error {
	return fmt.Errorf("Redirect not supported")
}

func (a *routerToCoreAdapter) Accepts(offers ...string) string          { return "" }
func (a *routerToCoreAdapter) AcceptsCharsets(offers ...string) string  { return "" }
func (a *routerToCoreAdapter) AcceptsEncodings(offers ...string) string { return "" }
func (a *routerToCoreAdapter) AcceptsLanguages(offers ...string) string { return "" }

func (a *routerToCoreAdapter) Fresh() bool { return false }
func (a *routerToCoreAdapter) Stale() bool { return true }
func (a *routerToCoreAdapter) XHR() bool   { return false }

func (a *routerToCoreAdapter) Locals(key string, value ...interface{}) interface{} {
	return a.ctx.Locals(key, value...)
}
func (a *routerToCoreAdapter) GetAllLocals() map[string]interface{} {
	return a.ctx.GetAllLocals()
}

func (a *routerToCoreAdapter) Next() error { return a.ctx.Next() }

func (a *routerToCoreAdapter) Context() context.Context {
	// Return a basic context - this is a limitation of the adapter
	return context.Background()
}

func (a *routerToCoreAdapter) IsMethod(method string) bool {
	return a.ctx.Method() == method
}
func (a *routerToCoreAdapter) RequestID() string { return "" }
