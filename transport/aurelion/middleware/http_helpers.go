package middleware

import (
	"reflect"
	"strings"

	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// Config represents server configuration (duplicated to avoid import cycle).
type Config struct {
	ServiceName string
	Port        int
	// Add other fields as needed, or use interface{}
}

// CORSConfig represents CORS configuration (duplicated to avoid import cycle).
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// CSRFConfig represents CSRF configuration (duplicated to avoid import cycle).
type CSRFConfig struct {
	KeyLookup         string
	CookieName        string
	CookiePath        string
	CookieDomain      string
	CookieSameSite    string
	CookieSecure      bool
	CookieHTTPOnly    bool
	CookieSessionOnly bool
	SingleUseToken    bool
	Expiration        interface{} // Use interface{} to avoid importing time.Duration
}

// ConfigInjector returns a fiber middleware that stores server config in request locals.
func ConfigInjector(cfg interface{}) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if cfg != nil {
			c.Locals("aurelion_config", cfg)
			rctx.TrackFiberLocal(c, "aurelion_config")
		}
		return c.Next()
	}
}

// BuildCORSConfig converts CORSConfig to fiber cors.Config.
func BuildCORSConfig(corsConfig interface{}) cors.Config {
	if corsConfig == nil {
		return cors.Config{}
	}
	// If it's already a middleware.CORSConfig, use it directly
	if mwCORS, ok := corsConfig.(*CORSConfig); ok {
		return cors.Config{
			AllowOrigins:     strings.Join(mwCORS.AllowOrigins, ","),
			AllowMethods:     strings.Join(mwCORS.AllowMethods, ","),
			AllowHeaders:     strings.Join(mwCORS.AllowHeaders, ","),
			AllowCredentials: mwCORS.AllowCredentials,
			ExposeHeaders:    strings.Join(mwCORS.ExposeHeaders, ","),
			MaxAge:           mwCORS.MaxAge,
		}
	}
	// Try to convert from other CORS config types using reflection-like approach
	// This handles config.CORSConfig and router.CORSConfig which have the same structure
	mwCORS := convertToMiddlewareCORS(corsConfig)
	if mwCORS == nil {
		return cors.Config{}
	}
	return cors.Config{
		AllowOrigins:     strings.Join(mwCORS.AllowOrigins, ","),
		AllowMethods:     strings.Join(mwCORS.AllowMethods, ","),
		AllowHeaders:     strings.Join(mwCORS.AllowHeaders, ","),
		AllowCredentials: mwCORS.AllowCredentials,
		ExposeHeaders:    strings.Join(mwCORS.ExposeHeaders, ","),
		MaxAge:           mwCORS.MaxAge,
	}
}

// convertToMiddlewareCORS converts any CORS config type to middleware.CORSConfig.
func convertToMiddlewareCORS(corsConfig interface{}) *CORSConfig {
	if corsConfig == nil {
		return nil
	}
	mwCORS := &CORSConfig{}

	// Use reflection to copy fields from any CORS config struct
	rv := reflect.ValueOf(corsConfig)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	// Copy fields by name using helper function
	copyCORSField(rv, "AllowOrigins", &mwCORS.AllowOrigins)
	copyCORSField(rv, "AllowMethods", &mwCORS.AllowMethods)
	copyCORSField(rv, "AllowHeaders", &mwCORS.AllowHeaders)
	copyCORSBoolField(rv, "AllowCredentials", &mwCORS.AllowCredentials)
	copyCORSField(rv, "ExposeHeaders", &mwCORS.ExposeHeaders)
	copyCORSAgeField(rv, "MaxAge", &mwCORS.MaxAge)

	return mwCORS
}

// copyCORSField copies a string slice field from reflection value to target
func copyCORSField(rv reflect.Value, fieldName string, target *[]string) {
	if field := rv.FieldByName(fieldName); field.IsValid() && field.Kind() == reflect.Slice {
		if slice, ok := field.Interface().([]string); ok {
			*target = slice
		}
	}
}

// copyCORSBoolField copies a bool field from reflection value to target
func copyCORSBoolField(rv reflect.Value, fieldName string, target *bool) {
	if field := rv.FieldByName(fieldName); field.IsValid() && field.Kind() == reflect.Bool {
		*target = field.Bool()
	}
}

// copyCORSAgeField copies an int field from reflection value to target
func copyCORSAgeField(rv reflect.Value, fieldName string, target *int) {
	if field := rv.FieldByName(fieldName); field.IsValid() && field.Kind() == reflect.Int {
		*target = int(field.Int())
	}
}

// convertToMiddlewareCSRF converts any CSRF config type to middleware.CSRFConfig.
func convertToMiddlewareCSRF(csrfConfig interface{}) *CSRFConfig {
	if csrfConfig == nil {
		return nil
	}
	mwCSRF := &CSRFConfig{}

	// Use reflection to copy fields from any CSRF config struct
	rv := reflect.ValueOf(csrfConfig)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	// Copy fields by name using helper functions
	copyCSRFStringField(rv, "KeyLookup", &mwCSRF.KeyLookup)
	copyCSRFStringField(rv, "CookieName", &mwCSRF.CookieName)
	copyCSRFStringField(rv, "CookiePath", &mwCSRF.CookiePath)
	copyCSRFStringField(rv, "CookieDomain", &mwCSRF.CookieDomain)
	copyCSRFStringField(rv, "CookieSameSite", &mwCSRF.CookieSameSite)
	copyCSRFBoolField(rv, "CookieSecure", &mwCSRF.CookieSecure)
	copyCSRFBoolField(rv, "CookieHTTPOnly", &mwCSRF.CookieHTTPOnly)
	copyCSRFBoolField(rv, "CookieSessionOnly", &mwCSRF.CookieSessionOnly)
	copyCSRFBoolField(rv, "SingleUseToken", &mwCSRF.SingleUseToken)
	copyCSRFExpirationField(rv, "Expiration", &mwCSRF.Expiration)

	return mwCSRF
}

// copyCSRFStringField copies a string field from reflection value to target
func copyCSRFStringField(rv reflect.Value, fieldName string, target *string) {
	if field := rv.FieldByName(fieldName); field.IsValid() && field.Kind() == reflect.String {
		*target = field.String()
	}
}

// copyCSRFBoolField copies a bool field from reflection value to target
func copyCSRFBoolField(rv reflect.Value, fieldName string, target *bool) {
	if field := rv.FieldByName(fieldName); field.IsValid() && field.Kind() == reflect.Bool {
		*target = field.Bool()
	}
}

// copyCSRFExpirationField copies an expiration field from reflection value to target
func copyCSRFExpirationField(rv reflect.Value, fieldName string, target *interface{}) {
	if field := rv.FieldByName(fieldName); field.IsValid() {
		*target = field.Interface()
	}
}

// BuildCSRFConfig converts CSRFConfig to fiber csrf.Config.
func BuildCSRFConfig(csrfConfig interface{}) csrf.Config {
	if csrfConfig == nil {
		return csrf.Config{}
	}
	// If it's already a middleware.CSRFConfig, use it directly
	if mwCSRF, ok := csrfConfig.(*CSRFConfig); ok {
		return buildCSRFFromConfig(mwCSRF)
	}
	// Try to convert from other CSRF config types
	mwCSRF := convertToMiddlewareCSRF(csrfConfig)
	if mwCSRF == nil {
		return csrf.Config{}
	}
	return buildCSRFFromConfig(mwCSRF)
}

// buildCSRFFromConfig builds fiber CSRF config from middleware.CSRFConfig.
func buildCSRFFromConfig(csrfConfig *CSRFConfig) csrf.Config {
	if csrfConfig == nil {
		return csrf.Config{}
	}

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

	switch csrfConfig.CookieSameSite {
	case "Strict":
		config.CookieSameSite = "Strict"
	case "Lax":
		config.CookieSameSite = "Lax"
	case "None":
		config.CookieSameSite = "None"
	default:
		config.CookieSameSite = "Lax"
	}

	// Handle expiration - would need type assertion if using interface{}
	// For now, skip expiration handling to avoid import cycle

	return config
}
