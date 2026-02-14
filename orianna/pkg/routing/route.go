// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// Route represents a single HTTP route configuration
type Route struct {
	Path                string
	Method              core.Method
	Methods             []core.Method // Support for multiple methods (replaces Method if set)
	Handler             core.Handler
	Middlewares         []core.Middleware
	RequiredPermissions []string
	IsProtected         bool
	CORS                *configuration.CORSConfig // Optional per-route CORS configuration
}

// RouteGroup represents a group of routes with a common prefix
type RouteGroup struct {
	Prefix      string
	Routes      []Route
	Groups      []RouteGroup // Nested groups
	Middlewares []core.Middleware
	IsProtected bool
}
