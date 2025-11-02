package aurelion

import (
	"errors"
	"strings"
	"time"

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
