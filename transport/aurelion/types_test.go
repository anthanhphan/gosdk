package aurelion

import "testing"

func TestMethod_String(t *testing.T) {
	tests := []struct {
		name   string
		method Method
		want   string
	}{
		{"GET method should return correct string", GET, "GET"},
		{"POST method should return correct string", POST, "POST"},
		{"PUT method should return correct string", PUT, "PUT"},
		{"PATCH method should return correct string", PATCH, "PATCH"},
		{"DELETE method should return correct string", DELETE, "DELETE"},
		{"HEAD method should return correct string", HEAD, "HEAD"},
		{"OPTIONS method should return correct string", OPTIONS, "OPTIONS"},
		{"invalid method should return UNKNOWN", Method(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.String(); got != tt.want {
				t.Errorf("Method.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Clone(t *testing.T) {
	original := &Route{
		Path:                "/test",
		Method:              GET,
		Handler:             func(ctx Context) error { return nil },
		Middlewares:         []Middleware{func(ctx Context) error { return nil }},
		RequiredPermissions: []string{"read", "write"},
		IsProtected:         true,
	}

	clone := original.Clone()

	// Check that clone is not the same instance
	if clone == original {
		t.Error("Clone should create a new instance, not reuse the same pointer")
	}

	// Check that all fields are copied correctly
	if clone.Path != original.Path {
		t.Errorf("Clone Path = %v, want %v", clone.Path, original.Path)
	}

	if clone.Method != original.Method {
		t.Errorf("Clone Method = %v, want %v", clone.Method, original.Method)
	}

	if clone.IsProtected != original.IsProtected {
		t.Errorf("Clone IsProtected = %v, want %v", clone.IsProtected, original.IsProtected)
	}

	// Check that slices are deep copied
	if len(clone.Middlewares) != len(original.Middlewares) {
		t.Errorf("Clone Middlewares length = %v, want %v", len(clone.Middlewares), len(original.Middlewares))
	}

	if len(clone.RequiredPermissions) != len(original.RequiredPermissions) {
		t.Errorf("Clone RequiredPermissions length = %v, want %v",
			len(clone.RequiredPermissions), len(original.RequiredPermissions))
	}

	// Modify clone slices to ensure they're independent
	if len(clone.RequiredPermissions) > 0 {
		clone.RequiredPermissions[0] = "modified"
		if original.RequiredPermissions[0] == "modified" {
			t.Error("Clone should have independent slice, but original was modified")
		}
	}
}

func TestRoute_Clone_WithCORS(t *testing.T) {
	original := &Route{
		Path:   "/api/users",
		Method: GET,
		Handler: func(ctx Context) error {
			return nil
		},
		CORS: &CORSConfig{
			AllowOrigins: []string{"https://example.com", "https://test.com"},
			AllowMethods: []string{"GET", "POST"},
			AllowHeaders: []string{"Content-Type"},
			MaxAge:       3600,
		},
	}

	clone := original.Clone()

	if clone == nil {
		t.Fatal("Clone should not be nil")
	}

	// Check CORS config is deep copied
	if clone.CORS == nil {
		t.Error("Clone should have CORS config")
	}
	if clone.CORS == original.CORS {
		t.Error("Clone should have independent CORS config, not reuse the same pointer")
	}

	// Check CORS fields are copied
	if len(clone.CORS.AllowOrigins) != len(original.CORS.AllowOrigins) {
		t.Errorf("Clone CORS AllowOrigins length = %v, want %v",
			len(clone.CORS.AllowOrigins), len(original.CORS.AllowOrigins))
	}

	if clone.CORS.MaxAge != original.CORS.MaxAge {
		t.Errorf("Clone CORS MaxAge = %v, want %v", clone.CORS.MaxAge, original.CORS.MaxAge)
	}

	// Modify clone CORS to ensure it's independent
	if len(clone.CORS.AllowOrigins) > 0 {
		clone.CORS.AllowOrigins[0] = "modified"
		if original.CORS.AllowOrigins[0] == "modified" {
			t.Error("Clone should have independent CORS slice, but original was modified")
		}
	}
}

func TestRoute_Clone_Nil(t *testing.T) {
	var route *Route = nil
	clone := route.Clone()

	if clone != nil {
		t.Error("Cloning nil route should return nil")
	}
}

func TestRoute_String(t *testing.T) {
	tests := []struct {
		name  string
		route *Route
		want  string
	}{
		{
			name: "GET route should format correctly",
			route: &Route{
				Method: GET,
				Path:   "/users",
			},
			want: "GET /users",
		},
		{
			name: "POST route should format correctly",
			route: &Route{
				Method: POST,
				Path:   "/users/:id",
			},
			want: "POST /users/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.route.String(); got != tt.want {
				t.Errorf("Route.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
