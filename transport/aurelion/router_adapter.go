package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/router"
	"github.com/gofiber/fiber/v2"
)

// buildRouteHandlers converts router.Route handlers to fiber.Handler slice.
func buildRouteHandlers(route *router.Route) []fiber.Handler {
	if route == nil {
		return nil
	}
	handlers := make([]fiber.Handler, 0, len(route.Middlewares)+1)

	// Convert middlewares
	for _, m := range route.Middlewares {
		if m != nil {
			handlers = append(handlers, convertRouterMiddleware(m))
		}
	}

	// Convert handler
	if route.Handler != nil {
		handlers = append(handlers, convertRouterHandler(route.Handler))
	}

	return handlers
}

// convertRouterHandler converts router.Handler to fiber.Handler.
func convertRouterHandler(handler router.Handler) fiber.Handler {
	if handler == nil {
		return nil
	}
	return HandlerToFiberPublic(convertRouterHandlerToAurelion(handler))
}

// convertRouterMiddleware converts router.Middleware to fiber.Handler.
func convertRouterMiddleware(middleware router.Middleware) fiber.Handler {
	if middleware == nil {
		return nil
	}
	return MiddlewareToFiberPublic(convertRouterMiddlewareToAurelion(middleware))
}

// routerContextAdapter adapts aurelion.Context to router.ContextInterface.
type routerContextAdapter struct {
	ctx Context
}

// Status implements router.ContextInterface.
func (a *routerContextAdapter) Status(status int) router.ContextInterface {
	a.ctx.Status(status)
	return a
}

// JSON implements router.ContextInterface.
func (a *routerContextAdapter) JSON(data interface{}) error {
	return a.ctx.JSON(data)
}

// Locals implements router.ContextInterface.
func (a *routerContextAdapter) Locals(key string, value ...interface{}) interface{} {
	return a.ctx.Locals(key, value...)
}

// GetAllLocals implements router.ContextInterface.
func (a *routerContextAdapter) GetAllLocals() map[string]interface{} {
	return a.ctx.GetAllLocals()
}

// Next implements router.ContextInterface.
func (a *routerContextAdapter) Next() error {
	return a.ctx.Next()
}

// Method implements router.ContextInterface.
func (a *routerContextAdapter) Method() string {
	return a.ctx.Method()
}

// Path implements router.ContextInterface.
func (a *routerContextAdapter) Path() string {
	return a.ctx.Path()
}

// Params implements router.ContextInterface.
func (a *routerContextAdapter) Params(key string, defaultValue ...string) string {
	return a.ctx.Params(key, defaultValue...)
}

// AllParams implements router.ContextInterface.
func (a *routerContextAdapter) AllParams() map[string]string {
	return a.ctx.AllParams()
}

// Query implements router.ContextInterface.
func (a *routerContextAdapter) Query(key string, defaultValue ...string) string {
	return a.ctx.Query(key, defaultValue...)
}

// AllQueries implements router.ContextInterface.
func (a *routerContextAdapter) AllQueries() map[string]string {
	return a.ctx.AllQueries()
}

// Body implements router.ContextInterface.
func (a *routerContextAdapter) Body() []byte {
	return a.ctx.Body()
}

// BodyParser implements router.ContextInterface.
func (a *routerContextAdapter) BodyParser(out interface{}) error {
	return a.ctx.BodyParser(out)
}

// ParamsParser implements router.ContextInterface.
func (a *routerContextAdapter) ParamsParser(out interface{}) error {
	return a.ctx.ParamsParser(out)
}

// QueryParser implements router.ContextInterface.
func (a *routerContextAdapter) QueryParser(out interface{}) error {
	return a.ctx.QueryParser(out)
}

// adaptContextForRouter adapts aurelion.Context to router.ContextInterface.
func adaptContextForRouter(ctx Context) router.ContextInterface {
	if ctx == nil {
		return nil
	}
	return &routerContextAdapter{ctx: ctx}
}

// extractContextFromRouterAdapter extracts the underlying aurelion.Context from router.ContextInterface.
// This is used by wrapper functions to convert router contexts back to aurelion contexts.
func extractContextFromRouterAdapter(ctx router.ContextInterface) Context {
	if ctx == nil {
		return nil
	}
	// Try to extract from adapter
	if adapter, ok := ctx.(*routerContextAdapter); ok {
		return adapter.ctx
	}
	// If not an adapter, we need to create one - but this means the context
	// is not from our system. For now, return nil (shouldn't happen in normal flow)
	return nil
}

// convertRouterHandlerToAurelion converts router.Handler to aurelion.Handler.
func convertRouterHandlerToAurelion(handler router.Handler) Handler {
	if handler == nil {
		return nil
	}
	return func(ctx Context) error {
		return handler(adaptContextForRouter(ctx))
	}
}

// convertRouterMiddlewareToAurelion converts router.Middleware to aurelion.Middleware.
func convertRouterMiddlewareToAurelion(middleware router.Middleware) Middleware {
	if middleware == nil {
		return nil
	}
	return func(ctx Context) error {
		return middleware(adaptContextForRouter(ctx))
	}
}

// buildGroupMiddlewares converts router.GroupRoute middlewares to fiber.Handler slice.
func buildGroupMiddlewares(group *router.GroupRoute) []fiber.Handler {
	if group == nil {
		return nil
	}
	fiberMiddlewares := make([]fiber.Handler, len(group.Middlewares))
	for i, m := range group.Middlewares {
		if m != nil {
			fiberMiddlewares[i] = convertRouterMiddleware(m)
		}
	}
	return fiberMiddlewares
}

// convertRouterMethod converts router.Method to aurelion.Method.
func convertRouterMethod(method router.Method) Method {
	return Method(method)
}

// convertRouterCORS converts router.CORSConfig to aurelion.CORSConfig.
func convertRouterCORS(cors *router.CORSConfig) *CORSConfig {
	if cors == nil {
		return nil
	}
	return &CORSConfig{
		AllowOrigins:     cors.AllowOrigins,
		AllowMethods:     cors.AllowMethods,
		AllowHeaders:     cors.AllowHeaders,
		AllowCredentials: cors.AllowCredentials,
		ExposeHeaders:    cors.ExposeHeaders,
		MaxAge:           cors.MaxAge,
	}
}
