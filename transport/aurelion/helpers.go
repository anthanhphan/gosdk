package aurelion

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// validateContext checks if the context is nil and returns an error if it is.
//
// Input:
//   - ctx: The context to validate
//
// Output:
//   - error: An error if the context is nil, nil otherwise
func validateContext(ctx Context) error {
	if ctx == nil {
		return errors.New(ErrContextNil)
	}
	return nil
}

// buildAPIResponse creates an APIResponse with the current timestamp.
//
// Input:
//   - success: Whether the response is successful
//   - code: The response code
//   - message: The response message
//   - data: Optional response data
//
// Output:
//   - APIResponse: The constructed API response
func buildAPIResponse(success bool, code int, message string, data ...interface{}) APIResponse {
	response := APIResponse{
		Success:   success,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return response
}

// buildCORSConfig converts CORSConfig to fiber cors.Config.
//
// Input:
//   - corsConfig: The CORS configuration
//
// Output:
//   - cors.Config: The fiber CORS configuration
func buildCORSConfig(corsConfig *CORSConfig) cors.Config {
	return cors.Config{
		AllowOrigins:     strings.Join(corsConfig.AllowOrigins, ","),
		AllowMethods:     strings.Join(corsConfig.AllowMethods, ","),
		AllowHeaders:     strings.Join(corsConfig.AllowHeaders, ","),
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    strings.Join(corsConfig.ExposeHeaders, ","),
		MaxAge:           corsConfig.MaxAge,
	}
}

// convertToRouteType attempts to convert an interface to a Route pointer.
// It handles RouteBuilder, *Route, and Route types.
//
// Input:
//   - r: The route interface to convert
//
// Output:
//   - *Route: The converted route, or nil if conversion fails
func convertToRouteType(r interface{}) *Route {
	switch v := r.(type) {
	case *RouteBuilder:
		if v == nil {
			return nil
		}
		return v.Build()
	case *Route:
		if v == nil {
			return nil
		}
		return v
	case Route:
		return &v
	default:
		return nil
	}
}

// convertToGroupRouteType attempts to convert an interface to a GroupRoute pointer.
// It handles GroupRouteBuilder, *GroupRoute, and GroupRoute types.
//
// Input:
//   - g: The group route interface to convert
//
// Output:
//   - *GroupRoute: The converted group route, or nil if conversion fails
func convertToGroupRouteType(g interface{}) *GroupRoute {
	switch v := g.(type) {
	case *GroupRouteBuilder:
		if v == nil {
			return nil
		}
		return v.Build()
	case *GroupRoute:
		if v == nil {
			return nil
		}
		return v
	case GroupRoute:
		return &v
	default:
		return nil
	}
}

// buildCSRFConfig converts CSRFConfig to fiber csrf.Config.
//
// Input:
//   - csrfConfig: The CSRF configuration
//
// Output:
//   - csrf.Config: The fiber CSRF configuration
func buildCSRFConfig(csrfConfig *CSRFConfig) csrf.Config {
	config := csrf.Config{
		KeyLookup:         csrfConfig.KeyLookup,
		CookieName:        csrfConfig.CookieName,
		CookiePath:        csrfConfig.CookiePath,
		CookieDomain:      csrfConfig.CookieDomain,
		CookieSecure:      csrfConfig.CookieSecure,
		CookieHTTPOnly:    csrfConfig.CookieHTTPOnly,
		CookieSessionOnly: csrfConfig.CookieSessionOnly,
		SingleUseToken:    csrfConfig.SingleUseToken,
	}

	// Set SameSite cookie attribute
	switch csrfConfig.CookieSameSite {
	case "Strict":
		config.CookieSameSite = "Strict"
	case "Lax":
		config.CookieSameSite = "Lax"
	case "None":
		config.CookieSameSite = "None"
	default:
		// Default to Lax if not specified
		config.CookieSameSite = "Lax"
	}

	// Set expiration if provided
	if csrfConfig.Expiration != nil {
		config.Expiration = *csrfConfig.Expiration
	}

	return config
}

// configMiddleware stores the server config in the request context.
// This allows response functions to access config settings like UseProperHTTPStatus.
//
// Input:
//   - config: The server configuration
//
// Output:
//   - fiber.Handler: The middleware handler
func configMiddleware(config *Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if config != nil {
			c.Locals(contextKeyConfig, config)
		}
		return c.Next()
	}
}

// getConfigFromContext retrieves the server config from context.
// Returns nil if config is not found in context.
//
// Input:
//   - ctx: The request context
//
// Output:
//   - *Config: The server configuration, or nil if not found
func getConfigFromContext(ctx Context) *Config {
	if ctx == nil {
		return nil
	}
	config, ok := ctx.Locals(contextKeyConfig).(*Config)
	if !ok {
		return nil
	}
	return config
}

// determineHTTPStatus determines the appropriate HTTP status code based on server configuration.
// If UseProperHTTPStatus is true, returns the proper HTTP status code.
// Otherwise, returns HTTP 200 for backward compatibility.
//
// Input:
//   - ctx: The request context
//   - properStatusCode: The proper HTTP status code (e.g., 400, 401, 404, 500)
//
// Output:
//   - int: The HTTP status code to use
func determineHTTPStatus(ctx Context, properStatusCode int) int {
	config := getConfigFromContext(ctx)
	if config != nil && config.UseProperHTTPStatus {
		return properStatusCode
	}
	// Default behavior for backward compatibility: always return 200
	// Only exception is InternalServerError which always uses 500
	if properStatusCode >= 500 {
		return properStatusCode
	}
	return 200
}
