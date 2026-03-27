// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package routing

import (
	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
)

// Route represents a single HTTP route configuration.
type Route struct {
	Path                string
	Methods             []core.Method
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
