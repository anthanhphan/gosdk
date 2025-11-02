package aurelion

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     Context
		wantErr bool
	}{
		{
			name:    "nil context should return error",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "valid context should not return error",
			ctx:     newMockContext(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err.Error() != ErrContextNil {
				t.Errorf("validateContext() error message = %v, want %v", err.Error(), ErrContextNil)
			}
		})
	}
}

func TestBuildAPIResponse(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		code    int
		message string
		data    []interface{}
		check   func(t *testing.T, resp APIResponse)
	}{
		{
			name:    "success response without data should work",
			success: true,
			code:    200,
			message: "Success",
			data:    nil,
			check: func(t *testing.T, resp APIResponse) {
				if !resp.Success {
					t.Error("Response should be successful")
				}
				if resp.Code != 200 {
					t.Errorf("Code = %v, want 200", resp.Code)
				}
				if resp.Message != "Success" {
					t.Errorf("Message = %v, want Success", resp.Message)
				}
				if resp.Data != nil {
					t.Error("Data should be nil")
				}
				if resp.Timestamp == 0 {
					t.Error("Timestamp should be set")
				}
			},
		},
		{
			name:    "success response with data should work",
			success: true,
			code:    201,
			message: "Created",
			data:    []interface{}{map[string]string{"id": "123"}},
			check: func(t *testing.T, resp APIResponse) {
				if resp.Data == nil {
					t.Error("Data should not be nil")
				}
			},
		},
		{
			name:    "error response should work",
			success: false,
			code:    400,
			message: "Bad Request",
			data:    nil,
			check: func(t *testing.T, resp APIResponse) {
				if resp.Success {
					t.Error("Response should not be successful")
				}
				if resp.Code != 400 {
					t.Errorf("Code = %v, want 400", resp.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := buildAPIResponse(tt.success, tt.code, tt.message, tt.data...)
			tt.check(t, resp)
		})
	}
}

func TestBuildCORSConfig(t *testing.T) {
	tests := []struct {
		name       string
		corsConfig *CORSConfig
		check      func(t *testing.T, config interface{})
	}{
		{
			name: "valid CORS config should be converted correctly",
			corsConfig: &CORSConfig{
				AllowOrigins:     []string{"https://example.com", "https://test.com"},
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Content-Type"},
				AllowCredentials: true,
				ExposeHeaders:    []string{"X-Custom"},
				MaxAge:           3600,
			},
			check: func(t *testing.T, config interface{}) {
				c, ok := config.(cors.Config)
				if !ok {
					t.Fatal("Config should be cors.Config type")
				}
				if c.AllowOrigins != "https://example.com,https://test.com" {
					t.Errorf("AllowOrigins = %v, want joined string", c.AllowOrigins)
				}
				if c.AllowMethods != "GET,POST" {
					t.Errorf("AllowMethods = %v, want GET,POST", c.AllowMethods)
				}
				if !c.AllowCredentials {
					t.Error("AllowCredentials should be true")
				}
				if c.MaxAge != 3600 {
					t.Errorf("MaxAge = %v, want 3600", c.MaxAge)
				}
			},
		},
		{
			name: "empty CORS config should work",
			corsConfig: &CORSConfig{
				AllowOrigins: []string{},
				AllowMethods: []string{},
			},
			check: func(t *testing.T, config interface{}) {
				c := config.(cors.Config)
				if c.AllowOrigins != "" {
					t.Error("AllowOrigins should be empty string")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := buildCORSConfig(tt.corsConfig)
			tt.check(t, config)
		})
	}
}

func TestConvertToRouteType(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, route *Route)
	}{
		{
			name:  "RouteBuilder should be converted",
			input: NewRoute("/test").GET().Handler(func(ctx Context) error { return nil }),
			check: func(t *testing.T, route *Route) {
				if route == nil {
					t.Fatal("Route should not be nil")
				}
				if route.Path != "/test" {
					t.Errorf("Path = %v, want /test", route.Path)
				}
			},
		},
		{
			name:  "*Route should be returned as is",
			input: &Route{Path: "/test", Method: GET},
			check: func(t *testing.T, route *Route) {
				if route == nil {
					t.Fatal("Route should not be nil")
				}
				if route.Path != "/test" {
					t.Errorf("Path = %v, want /test", route.Path)
				}
			},
		},
		{
			name:  "Route value should be converted to pointer",
			input: Route{Path: "/test", Method: POST},
			check: func(t *testing.T, route *Route) {
				if route == nil {
					t.Fatal("Route should not be nil")
				}
				if route.Method != POST {
					t.Errorf("Method = %v, want POST", route.Method)
				}
			},
		},
		{
			name:  "nil *RouteBuilder should return nil",
			input: (*RouteBuilder)(nil),
			check: func(t *testing.T, route *Route) {
				if route != nil {
					t.Error("Route should be nil")
				}
			},
		},
		{
			name:  "nil *Route should return nil",
			input: (*Route)(nil),
			check: func(t *testing.T, route *Route) {
				if route != nil {
					t.Error("Route should be nil")
				}
			},
		},
		{
			name:  "invalid type should return nil",
			input: "invalid",
			check: func(t *testing.T, route *Route) {
				if route != nil {
					t.Error("Route should be nil for invalid type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := convertToRouteType(tt.input)
			tt.check(t, route)
		})
	}
}

func TestConvertToGroupRouteType(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, group *GroupRoute)
	}{
		{
			name:  "GroupRouteBuilder should be converted",
			input: NewGroupRoute("/api").Routes(NewRoute("/test").GET().Handler(func(ctx Context) error { return nil })),
			check: func(t *testing.T, group *GroupRoute) {
				if group == nil {
					t.Fatal("Group should not be nil")
				}
				if group.Prefix != "/api" {
					t.Errorf("Prefix = %v, want /api", group.Prefix)
				}
			},
		},
		{
			name:  "*GroupRoute should be returned as is",
			input: &GroupRoute{Prefix: "/api"},
			check: func(t *testing.T, group *GroupRoute) {
				if group == nil {
					t.Fatal("Group should not be nil")
				}
				if group.Prefix != "/api" {
					t.Errorf("Prefix = %v, want /api", group.Prefix)
				}
			},
		},
		{
			name:  "GroupRoute value should be converted to pointer",
			input: GroupRoute{Prefix: "/v1"},
			check: func(t *testing.T, group *GroupRoute) {
				if group == nil {
					t.Fatal("Group should not be nil")
				}
				if group.Prefix != "/v1" {
					t.Errorf("Prefix = %v, want /v1", group.Prefix)
				}
			},
		},
		{
			name:  "nil *GroupRouteBuilder should return nil",
			input: (*GroupRouteBuilder)(nil),
			check: func(t *testing.T, group *GroupRoute) {
				if group != nil {
					t.Error("Group should be nil")
				}
			},
		},
		{
			name:  "nil *GroupRoute should return nil",
			input: (*GroupRoute)(nil),
			check: func(t *testing.T, group *GroupRoute) {
				if group != nil {
					t.Error("Group should be nil")
				}
			},
		},
		{
			name:  "invalid type should return nil",
			input: 123,
			check: func(t *testing.T, group *GroupRoute) {
				if group != nil {
					t.Error("Group should be nil for invalid type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := convertToGroupRouteType(tt.input)
			tt.check(t, group)
		})
	}
}

func TestBuildCSRFConfig(t *testing.T) {
	expiration := 5 * time.Minute

	tests := []struct {
		name       string
		csrfConfig *CSRFConfig
		check      func(t *testing.T, config csrf.Config)
	}{
		{
			name: "minimal config should set defaults",
			csrfConfig: &CSRFConfig{
				KeyLookup: "header:X-Csrf-Token",
			},
			check: func(t *testing.T, config csrf.Config) {
				if config.KeyLookup != "header:X-Csrf-Token" {
					t.Errorf("Expected KeyLookup = header:X-Csrf-Token, got %s", config.KeyLookup)
				}
				if config.CookieSameSite != "Lax" {
					t.Errorf("Expected default CookieSameSite = Lax, got %s", config.CookieSameSite)
				}
			},
		},
		{
			name: "full config should set all fields",
			csrfConfig: &CSRFConfig{
				KeyLookup:         "cookie:csrf_token",
				CookieName:        "csrf",
				CookiePath:        "/",
				CookieDomain:      "example.com",
				CookieSecure:      true,
				CookieHTTPOnly:    true,
				CookieSessionOnly: true,
				CookieSameSite:    "Strict",
				SingleUseToken:    true,
				Expiration:        &expiration,
			},
			check: func(t *testing.T, config csrf.Config) {
				if config.KeyLookup != "cookie:csrf_token" {
					t.Errorf("Expected KeyLookup = cookie:csrf_token, got %s", config.KeyLookup)
				}
				if config.CookieName != "csrf" {
					t.Errorf("Expected CookieName = csrf, got %s", config.CookieName)
				}
				if config.CookiePath != "/" {
					t.Errorf("Expected CookiePath = /, got %s", config.CookiePath)
				}
				if config.CookieDomain != "example.com" {
					t.Errorf("Expected CookieDomain = example.com, got %s", config.CookieDomain)
				}
				if !config.CookieSecure {
					t.Error("Expected CookieSecure = true")
				}
				if !config.CookieHTTPOnly {
					t.Error("Expected CookieHTTPOnly = true")
				}
				if !config.CookieSessionOnly {
					t.Error("Expected CookieSessionOnly = true")
				}
				if config.CookieSameSite != "Strict" {
					t.Errorf("Expected CookieSameSite = Strict, got %s", config.CookieSameSite)
				}
				if !config.SingleUseToken {
					t.Error("Expected SingleUseToken = true")
				}
				if config.Expiration != expiration {
					t.Errorf("Expected Expiration = %v, got %v", expiration, config.Expiration)
				}
			},
		},
		{
			name: "SameSite Lax should be set correctly",
			csrfConfig: &CSRFConfig{
				CookieSameSite: "Lax",
			},
			check: func(t *testing.T, config csrf.Config) {
				if config.CookieSameSite != "Lax" {
					t.Errorf("Expected CookieSameSite = Lax, got %s", config.CookieSameSite)
				}
			},
		},
		{
			name: "SameSite None should be set correctly",
			csrfConfig: &CSRFConfig{
				CookieSameSite: "None",
			},
			check: func(t *testing.T, config csrf.Config) {
				if config.CookieSameSite != "None" {
					t.Errorf("Expected CookieSameSite = None, got %s", config.CookieSameSite)
				}
			},
		},
		{
			name: "invalid SameSite should default to Lax",
			csrfConfig: &CSRFConfig{
				CookieSameSite: "Invalid",
			},
			check: func(t *testing.T, config csrf.Config) {
				if config.CookieSameSite != "Lax" {
					t.Errorf("Expected default CookieSameSite = Lax, got %s", config.CookieSameSite)
				}
			},
		},
		{
			name: "nil expiration should not set expiration",
			csrfConfig: &CSRFConfig{
				Expiration: nil,
			},
			check: func(t *testing.T, config csrf.Config) {
				// Expiration should be zero value (0)
				if config.Expiration != 0 {
					t.Errorf("Expected Expiration = 0, got %v", config.Expiration)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := buildCSRFConfig(tt.csrfConfig)
			tt.check(t, config)
		})
	}
}
