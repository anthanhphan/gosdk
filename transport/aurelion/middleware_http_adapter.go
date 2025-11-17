package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// ConfigInjectorPublic wraps middleware.ConfigInjector to accept public Config type.
func ConfigInjectorPublic(cfg *Config) func(*fiber.Ctx) error {
	return middleware.ConfigInjector(cfg)
}

// BuildCORSConfigPublic converts public CORSConfig to fiber cors.Config.
func BuildCORSConfigPublic(corsConfig *CORSConfig) cors.Config {
	if corsConfig == nil {
		return cors.Config{}
	}
	internalCORS := &middleware.CORSConfig{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		MaxAge:           corsConfig.MaxAge,
	}
	return middleware.BuildCORSConfig(internalCORS)
}

// BuildCSRFConfigPublic converts public CSRFConfig to fiber csrf.Config.
func BuildCSRFConfigPublic(csrfConfig *CSRFConfig) csrf.Config {
	if csrfConfig == nil {
		return csrf.Config{}
	}
	internalCSRF := &middleware.CSRFConfig{
		KeyLookup:         csrfConfig.KeyLookup,
		CookieName:        csrfConfig.CookieName,
		CookiePath:        csrfConfig.CookiePath,
		CookieDomain:      csrfConfig.CookieDomain,
		CookieSameSite:    csrfConfig.CookieSameSite,
		CookieSecure:      csrfConfig.CookieSecure,
		CookieHTTPOnly:    csrfConfig.CookieHTTPOnly,
		CookieSessionOnly: csrfConfig.CookieSessionOnly,
		SingleUseToken:    csrfConfig.SingleUseToken,
		Expiration:        csrfConfig.Expiration, // interface{} can hold *time.Duration
	}
	result := middleware.BuildCSRFConfig(internalCSRF)
	// Handle expiration if provided
	if csrfConfig.Expiration != nil {
		result.Expiration = *csrfConfig.Expiration
	}
	return result
}
