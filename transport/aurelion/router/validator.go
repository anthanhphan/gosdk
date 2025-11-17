package router

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateRoute validates a single route.
func ValidateRoute(route *Route) error {
	if route == nil {
		return errors.New("route cannot be nil")
	}

	if route.Path != "" {
		if len(route.Path) > MaxRoutePathLength {
			return fmt.Errorf("route path exceeds maximum length of %d characters", MaxRoutePathLength)
		}
		if !strings.HasPrefix(route.Path, "/") {
			return fmt.Errorf("route path must start with '/': %s", route.Path)
		}
	}

	if route.Handler == nil {
		return fmt.Errorf("route handler cannot be nil for path: %s", route.Path)
	}

	totalHandlers := len(route.Middlewares) + 1
	if totalHandlers > MaxRouteHandlersPerRoute {
		return fmt.Errorf("route has too many handlers (max %d): %s", MaxRouteHandlersPerRoute, route.Path)
	}

	return nil
}

// ValidateGroupRoute validates a group route.
func ValidateGroupRoute(group *GroupRoute) error {
	if group == nil {
		return errors.New("group route cannot be nil")
	}

	if group.Prefix == "" {
		return errors.New("group route prefix cannot be empty")
	}

	if !strings.HasPrefix(group.Prefix, "/") {
		return fmt.Errorf("group route prefix must start with '/': %s", group.Prefix)
	}

	if len(group.Prefix) > MaxRoutePathLength {
		return fmt.Errorf("group route prefix exceeds maximum length of %d characters", MaxRoutePathLength)
	}

	if len(group.Routes) == 0 {
		return fmt.Errorf("group route has no routes: %s", group.Prefix)
	}

	if len(group.Middlewares) > MaxRouteHandlersPerRoute {
		return fmt.Errorf("group has too many middlewares (max %d): %s", MaxRouteHandlersPerRoute, group.Prefix)
	}

	for i, route := range group.Routes {
		if route.Path != "" && !strings.HasPrefix(route.Path, "/") {
			return fmt.Errorf("invalid route at index %d in group %s: route path must start with '/': %s", i, group.Prefix, route.Path)
		}
		if err := ValidateRoute(&route); err != nil {
			return fmt.Errorf("invalid route at index %d in group %s: %w", i, group.Prefix, err)
		}
	}

	return nil
}
