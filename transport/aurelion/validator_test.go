package aurelion

import (
	"testing"
)

func TestValidateRoute(t *testing.T) {
	tests := []struct {
		name    string
		route   *Route
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil route should return error",
			route:   nil,
			wantErr: true,
			errMsg:  "route cannot be nil",
		},
		{
			name: "valid route should not return error",
			route: &Route{
				Path:    "/test",
				Method:  GET,
				Handler: func(ctx Context) error { return nil },
			},
			wantErr: false,
		},
		{
			name: "route path without leading slash should return error",
			route: &Route{
				Path:    "test",
				Method:  GET,
				Handler: func(ctx Context) error { return nil },
			},
			wantErr: true,
			errMsg:  "must start with '/'",
		},
		{
			name: "route with empty path should be valid",
			route: &Route{
				Path:    "",
				Method:  GET,
				Handler: func(ctx Context) error { return nil },
			},
			wantErr: false,
		},
		{
			name: "invalid HTTP method should return error",
			route: &Route{
				Path:    "/test",
				Method:  Method(999),
				Handler: func(ctx Context) error { return nil },
			},
			wantErr: true,
			errMsg:  "invalid HTTP method",
		},
		{
			name: "nil handler should return error",
			route: &Route{
				Path:    "/test",
				Method:  GET,
				Handler: nil,
			},
			wantErr: true,
			errMsg:  "handler cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoute(tt.route)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateRoute_PathLength(t *testing.T) {
	longPath := "/" + string(make([]byte, MaxRoutePathLength+1))

	route := &Route{
		Path:    longPath,
		Method:  GET,
		Handler: func(ctx Context) error { return nil },
	}

	err := validateRoute(route)
	if err == nil {
		t.Error("Expected error for path exceeding max length")
	}
	if !contains(err.Error(), "exceeds maximum length") {
		t.Errorf("Error message = %v, want to contain 'exceeds maximum length'", err.Error())
	}
}

func TestValidateRoute_TooManyHandlers(t *testing.T) {
	route := &Route{
		Path:    "/test",
		Method:  GET,
		Handler: func(ctx Context) error { return nil },
	}

	// Add too many middlewares
	for i := 0; i < MaxRouteHandlersPerRoute; i++ {
		route.Middlewares = append(route.Middlewares, func(ctx Context) error { return nil })
	}

	err := validateRoute(route)
	if err == nil {
		t.Error("Expected error for too many handlers")
	}
	if !contains(err.Error(), "too many handlers") {
		t.Errorf("Error message = %v, want to contain 'too many handlers'", err.Error())
	}
}

func TestValidateGroupRoute(t *testing.T) {
	tests := []struct {
		name    string
		group   *GroupRoute
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil group should return error",
			group:   nil,
			wantErr: true,
			errMsg:  "group route cannot be nil",
		},
		{
			name: "valid group should not return error",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{
					{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
				},
			},
			wantErr: false,
		},
		{
			name: "empty prefix should return error",
			group: &GroupRoute{
				Prefix: "",
				Routes: []Route{
					{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
				},
			},
			wantErr: true,
			errMsg:  "prefix cannot be empty",
		},
		{
			name: "prefix without leading slash should return error",
			group: &GroupRoute{
				Prefix: "api",
				Routes: []Route{
					{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
				},
			},
			wantErr: true,
			errMsg:  "must start with '/'",
		},
		{
			name: "prefix exceeding max length should return error",
			group: &GroupRoute{
				Prefix: "/" + string(make([]byte, MaxRoutePathLength+1)),
				Routes: []Route{
					{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
				},
			},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name: "group with no routes should return error",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{},
			},
			wantErr: true,
			errMsg:  "has no routes",
		},
		{
			name: "group with too many middlewares should return error",
			group: &GroupRoute{
				Prefix: "/api",
				Routes: []Route{
					{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
				},
			},
			wantErr: true,
			errMsg:  "too many middlewares",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "group with too many middlewares should return error" {
				// Add too many middlewares
				for i := 0; i < MaxRouteHandlersPerRoute+1; i++ {
					tt.group.Middlewares = append(tt.group.Middlewares, func(ctx Context) error { return nil })
				}
			}

			err := validateGroupRoute(tt.group)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateGroupRoute_InvalidChildRoute(t *testing.T) {
	group := &GroupRoute{
		Prefix: "/api",
		Routes: []Route{
			{
				Path:    "invalid",
				Method:  GET,
				Handler: func(ctx Context) error { return nil },
			},
		},
	}

	err := validateGroupRoute(group)
	if err == nil {
		t.Error("Expected error for invalid child route")
	}
	if !contains(err.Error(), "must start with '/'") {
		t.Errorf("Error message = %v, want to contain 'must start with '/'", err.Error())
	}
}

func TestValidateGroupRoute_TooManyMiddlewares(t *testing.T) {
	group := &GroupRoute{
		Prefix: "/api",
		Routes: []Route{
			{Path: "/users", Method: GET, Handler: func(ctx Context) error { return nil }},
		},
	}

	// Add too many middlewares
	for i := 0; i < MaxRouteHandlersPerRoute+1; i++ {
		group.Middlewares = append(group.Middlewares, func(ctx Context) error { return nil })
	}

	err := validateGroupRoute(group)
	if err == nil {
		t.Error("Expected error for too many middlewares")
	}
	if !contains(err.Error(), "too many middlewares") {
		t.Errorf("Error message = %v, want to contain 'too many middlewares'", err.Error())
	}
}
