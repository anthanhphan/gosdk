// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"strings"

	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// buildCORSConfig converts CORSConfig to fiber cors.Config
func buildCORSConfig(corsConfig *configuration.CORSConfig) cors.Config {
	return cors.Config{
		AllowOrigins:     strings.Join(corsConfig.AllowOrigins, ","),
		AllowMethods:     strings.Join(corsConfig.AllowMethods, ","),
		AllowHeaders:     strings.Join(corsConfig.AllowHeaders, ","),
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    strings.Join(corsConfig.ExposeHeaders, ","),
		MaxAge:           corsConfig.MaxAge,
	}
}

// buildCSRFConfig converts CSRFConfig to fiber csrf.Config
func buildCSRFConfig(csrfConfig *configuration.CSRFConfig) csrf.Config {
	conf := csrf.Config{
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
		conf.CookieSameSite = "Strict"
	case "Lax":
		conf.CookieSameSite = "Lax"
	case "None":
		conf.CookieSameSite = "None"
	default:
		conf.CookieSameSite = "Lax"
	}

	// Set expiration if provided
	if csrfConfig.Expiration != nil {
		conf.Expiration = *csrfConfig.Expiration
	}

	return conf
}
