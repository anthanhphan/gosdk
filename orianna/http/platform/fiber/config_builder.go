// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"strings"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
)

// buildCORSConfig converts CORSConfig to fiber cors.Config
func buildCORSConfig(corsConfig *configuration.CORSConfig) cors.Config {
	return cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		MaxAge:           corsConfig.MaxAge,
	}
}

// buildCSRFConfig converts CSRFConfig to fiber csrf.Config.
// Applies secure-by-default values: CookieHTTPOnly=true, CookieSecure=true, SameSite=Strict.
// Users must explicitly set *bool to false to opt out of security hardening.
func buildCSRFConfig(csrfConfig *configuration.CSRFConfig) csrf.Config {
	// Secure defaults: true unless explicitly set to false
	cookieSecure := true
	if csrfConfig.CookieSecure != nil {
		cookieSecure = *csrfConfig.CookieSecure
	}
	cookieHTTPOnly := true
	if csrfConfig.CookieHTTPOnly != nil {
		cookieHTTPOnly = *csrfConfig.CookieHTTPOnly
	}

	conf := csrf.Config{
		CookieName:        csrfConfig.CookieName,
		CookiePath:        csrfConfig.CookiePath,
		CookieDomain:      csrfConfig.CookieDomain,
		CookieSecure:      cookieSecure,
		CookieHTTPOnly:    cookieHTTPOnly,
		CookieSessionOnly: csrfConfig.CookieSessionOnly,
		SingleUseToken:    csrfConfig.SingleUseToken,
	}

	// Convert KeyLookup string (e.g. "header:X-CSRF-Token") to v3 Extractor.
	if csrfConfig.KeyLookup != "" {
		conf.Extractor = parseKeyLookupToExtractor(csrfConfig.KeyLookup)
	}

	// Set SameSite cookie attribute (default to "Strict" for maximum security)
	switch csrfConfig.CookieSameSite {
	case "Strict", "Lax", "None":
		conf.CookieSameSite = csrfConfig.CookieSameSite
	default:
		conf.CookieSameSite = "Strict"
	}

	// Set idle timeout if provided
	if csrfConfig.Expiration != nil {
		conf.IdleTimeout = *csrfConfig.Expiration
	}

	return conf
}

// parseKeyLookupToExtractor converts a v2-style KeyLookup string to a v3 Extractor.
// Supported formats: "header:<name>", "form:<name>", "query:<name>", "param:<name>".
func parseKeyLookupToExtractor(keyLookup string) extractors.Extractor {
	parts := strings.SplitN(keyLookup, ":", 2)
	if len(parts) != 2 {
		// Fallback: treat entire string as header name
		return extractors.FromHeader(keyLookup)
	}

	source, name := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	switch strings.ToLower(source) {
	case "header":
		return extractors.FromHeader(name)
	case "form":
		return extractors.FromForm(name)
	case "query":
		return extractors.FromQuery(name)
	case "param":
		return extractors.FromParam(name)
	default:
		return extractors.FromHeader(name)
	}
}
