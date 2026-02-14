// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// RouteRegistry manages route registration and validation.
// It is safe for concurrent use.
type RouteRegistry struct {
	mu              sync.RWMutex
	routes          []Route
	groups          []RouteGroup
	registeredPaths map[string]struct{} // tracks "METHOD /path" to detect duplicates
	authMiddleware  core.Middleware
	authzChecker    func(core.Context, []string) error
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes:          make([]Route, 0),
		groups:          make([]RouteGroup, 0),
		registeredPaths: make(map[string]struct{}),
	}
}

// RegisterRoute registers a single route
func (rr *RouteRegistry) RegisterRoute(route Route) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	return rr.registerRouteLocked(&route)
}

// RegisterRoutes registers multiple routes atomically.
// Either all routes are registered or none (on first validation error).
func (rr *RouteRegistry) RegisterRoutes(routes ...Route) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for i := range routes {
		if err := rr.registerRouteLocked(&routes[i]); err != nil {
			return fmt.Errorf("route at index %d: %w", i, err)
		}
	}
	return nil
}

// registerRouteLocked registers a single route. Caller must hold mu.Lock().
func (rr *RouteRegistry) registerRouteLocked(route *Route) error {
	if err := validateRoute(route); err != nil {
		return fmt.Errorf("invalid route: %w", err)
	}

	// Check for duplicate routes
	if err := rr.checkDuplicateRoute(route, ""); err != nil {
		return err
	}

	// Apply protection middleware if needed
	rr.applyProtectionMiddleware(route)

	rr.routes = append(rr.routes, *route)
	return nil
}

// RegisterGroup registers a route group
func (rr *RouteRegistry) RegisterGroup(group RouteGroup) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if err := rr.validateGroup(&group); err != nil {
		return fmt.Errorf("invalid group: %w", err)
	}

	// Check for duplicate routes within the group
	if err := rr.checkGroupDuplicates(&group, ""); err != nil {
		return err
	}

	// Apply protection to routes in group
	rr.applyGroupProtection(&group)

	rr.groups = append(rr.groups, group)
	return nil
}

// GetRoutes returns a copy of all registered routes
func (rr *RouteRegistry) GetRoutes() []Route {
	rr.mu.RLock()
	defer rr.mu.RUnlock()
	out := make([]Route, len(rr.routes))
	copy(out, rr.routes)
	return out
}

// GetGroups returns a copy of all registered groups
func (rr *RouteRegistry) GetGroups() []RouteGroup {
	rr.mu.RLock()
	defer rr.mu.RUnlock()
	out := make([]RouteGroup, len(rr.groups))
	copy(out, rr.groups)
	return out
}

// SetAuthMiddleware sets the authentication middleware
func (rr *RouteRegistry) SetAuthMiddleware(middleware core.Middleware) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.authMiddleware = middleware
}

// SetAuthzChecker sets the authorization checker
func (rr *RouteRegistry) SetAuthzChecker(checker func(core.Context, []string) error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.authzChecker = checker
}

// routeKey builds the lookup key "METHOD /path" for duplicate detection.
func routeKey(method core.Method, prefix, path string) string {
	fullPath := prefix + path
	return method.String() + " " + fullPath
}

// checkDuplicateRoute checks if any method+path combination is already registered.
// prefix is prepended to the route path (used for group routes).
func (rr *RouteRegistry) checkDuplicateRoute(route *Route, prefix string) error {
	methods := route.Methods
	if len(methods) == 0 {
		methods = []core.Method{route.Method}
	}

	for _, m := range methods {
		key := routeKey(m, prefix, route.Path)
		if _, exists := rr.registeredPaths[key]; exists {
			return fmt.Errorf("%w: %s", core.ErrDuplicateRoute, key)
		}
	}

	// Register all method+path combinations
	for _, m := range methods {
		rr.registeredPaths[routeKey(m, prefix, route.Path)] = struct{}{}
	}
	return nil
}

// checkGroupDuplicates recursively checks all routes in a group for duplicates.
func (rr *RouteRegistry) checkGroupDuplicates(group *RouteGroup, parentPrefix string) error {
	fullPrefix := parentPrefix + group.Prefix

	for i := range group.Routes {
		if err := rr.checkDuplicateRoute(&group.Routes[i], fullPrefix); err != nil {
			return err
		}
	}

	for i := range group.Groups {
		if err := rr.checkGroupDuplicates(&group.Groups[i], fullPrefix); err != nil {
			return err
		}
	}
	return nil
}

// validateRoute validates a route configuration
func validateRoute(route *Route) error {
	if route == nil {
		return errors.New("route cannot be nil")
	}

	if route.Path != "" {
		if !strings.HasPrefix(route.Path, "/") {
			return fmt.Errorf("route path must start with '/': %s", route.Path)
		}
	}

	if route.Handler == nil {
		return fmt.Errorf("route handler cannot be nil for path: %s", route.Path)
	}

	return nil
}

// validateGroup validates a route group
func (rr *RouteRegistry) validateGroup(group *RouteGroup) error {
	if group == nil {
		return errors.New("group route cannot be nil")
	}

	if group.Prefix == "" {
		return errors.New("group route prefix cannot be empty")
	}

	if !strings.HasPrefix(group.Prefix, "/") {
		return fmt.Errorf("group route prefix must start with '/': %s", group.Prefix)
	}

	if len(group.Routes) == 0 && len(group.Groups) == 0 {
		return fmt.Errorf("group route has no routes or subgroups: %s", group.Prefix)
	}

	for _, subGroup := range group.Groups {
		if err := rr.validateGroup(&subGroup); err != nil {
			return err
		}
	}

	return nil
}

// applyProtectionMiddleware applies authentication and authorization middleware to a route
func (rr *RouteRegistry) applyProtectionMiddleware(route *Route) {
	if route == nil || !route.IsProtected {
		return
	}

	// Calculate how many protection middlewares we need
	protectionCount := 0
	if rr.authMiddleware != nil {
		protectionCount++
	}
	if len(route.RequiredPermissions) > 0 && rr.authzChecker != nil {
		protectionCount++
	}

	// If no protection middlewares needed, return early
	if protectionCount == 0 {
		return
	}

	// Pre-allocate new slice with exact capacity to avoid reallocation
	// Capacity = existing middlewares + protection middlewares
	newMiddlewares := make([]core.Middleware, 0, len(route.Middlewares)+protectionCount)

	// Add authentication middleware first
	if rr.authMiddleware != nil {
		newMiddlewares = append(newMiddlewares, rr.authMiddleware)
	}

	// Add authorization middleware if there are required permissions
	if len(route.RequiredPermissions) > 0 && rr.authzChecker != nil {
		authzMiddleware := rr.createAuthorizationMiddleware(route.RequiredPermissions)
		newMiddlewares = append(newMiddlewares, authzMiddleware)
	}

	// Append existing middlewares
	newMiddlewares = append(newMiddlewares, route.Middlewares...)

	// Replace route middlewares with new slice (single assignment, no intermediate allocations)
	route.Middlewares = newMiddlewares
}

// applyGroupProtection applies protection middleware to routes in a group and its subgroups
func (rr *RouteRegistry) applyGroupProtection(group *RouteGroup) {
	// Apply protection to direct routes
	for i := range group.Routes {
		// If group is protected, protect all routes in the group
		if group.IsProtected {
			group.Routes[i].IsProtected = true
		}

		// Apply protection middleware if route is protected
		if group.Routes[i].IsProtected {
			rr.applyProtectionMiddleware(&group.Routes[i])
		}
	}

	// Apply protection to nested groups
	for i := range group.Groups {
		if group.IsProtected {
			group.Groups[i].IsProtected = true
		}
		rr.applyGroupProtection(&group.Groups[i])
	}
}

// createAuthorizationMiddleware creates an authorization middleware
func (rr *RouteRegistry) createAuthorizationMiddleware(permissions []string) core.Middleware {
	checker := rr.authzChecker // capture the checker at creation time
	return func(ctx core.Context) error {
		if err := checker(ctx, permissions); err != nil {
			return fmt.Errorf("insufficient permissions: %w", err)
		}
		return ctx.Next()
	}
}
