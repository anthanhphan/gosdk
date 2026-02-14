package configuration

import (
	"testing"
	"time"
)

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	if config == nil {
		t.Fatal("DefaultMiddlewareConfig() should not return nil")
	}

	if config.DisableHelmet != false {
		t.Error("DisableHelmet should be false by default")
	}

	if config.DisableRateLimit != false {
		t.Error("DisableRateLimit should be false by default")
	}

	if config.DisableCompression != false {
		t.Error("DisableCompression should be false by default")
	}

	if config.DisableRecovery != false {
		t.Error("DisableRecovery should be false by default")
	}

	if config.DisableRequestID != false {
		t.Error("DisableRequestID should be false by default")
	}

	if config.DisableTraceID != false {
		t.Error("DisableTraceID should be false by default")
	}

	if config.DisableLogging != false {
		t.Error("DisableLogging should be false by default")
	}

}

func TestMiddlewareConfigStructure(t *testing.T) {
	config := &MiddlewareConfig{
		DisableHelmet:      true,
		DisableRateLimit:   true,
		DisableCompression: true,
		DisableRecovery:    true,
		DisableRequestID:   true,
		DisableTraceID:     true,
		DisableLogging:     true,
	}

	if !config.DisableHelmet {
		t.Error("DisableHelmet should be true")
	}
	if !config.DisableRateLimit {
		t.Error("DisableRateLimit should be true")
	}
	if !config.DisableCompression {
		t.Error("DisableCompression should be true")
	}
	if !config.DisableRecovery {
		t.Error("DisableRecovery should be true")
	}
	if !config.DisableRequestID {
		t.Error("DisableRequestID should be true")
	}
	if !config.DisableTraceID {
		t.Error("DisableTraceID should be true")
	}
	if !config.DisableLogging {
		t.Error("DisableLogging should be true")
	}
}

func TestConfigStructure(t *testing.T) {
	readTimeout := 10 * time.Second
	writeTimeout := 15 * time.Second
	idleTimeout := 30 * time.Second
	gracefulTimeout := 20 * time.Second

	config := &Config{
		ServiceName:              "test-service",
		Port:                     8080,
		ReadTimeout:              &readTimeout,
		WriteTimeout:             &writeTimeout,
		IdleTimeout:              &idleTimeout,
		GracefulShutdownTimeout:  &gracefulTimeout,
		MaxBodySize:              1024 * 1024,
		MaxConcurrentConnections: 1000,
		EnableCORS:               true,
		EnableCSRF:               true,
		VerboseLogging:           true,
		VerboseLoggingSkipPaths:  []string{"/health", "/metrics"},
		UseProperHTTPStatus:      true,
	}

	if config.ServiceName != "test-service" {
		t.Errorf("ServiceName = %s, want test-service", config.ServiceName)
	}

	if config.Port != 8080 {
		t.Errorf("Port = %d, want 8080", config.Port)
	}

	if *config.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want 10s", *config.ReadTimeout)
	}

	if *config.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout = %v, want 15s", *config.WriteTimeout)
	}

	if *config.IdleTimeout != 30*time.Second {
		t.Errorf("IdleTimeout = %v, want 30s", *config.IdleTimeout)
	}

	if *config.GracefulShutdownTimeout != 20*time.Second {
		t.Errorf("GracefulShutdownTimeout = %v, want 20s", *config.GracefulShutdownTimeout)
	}

	if config.MaxBodySize != 1024*1024 {
		t.Errorf("MaxBodySize = %d, want %d", config.MaxBodySize, 1024*1024)
	}

	if config.MaxConcurrentConnections != 1000 {
		t.Errorf("MaxConcurrentConnections = %d, want 1000", config.MaxConcurrentConnections)
	}

	if !config.EnableCORS {
		t.Error("EnableCORS should be true")
	}

	if !config.EnableCSRF {
		t.Error("EnableCSRF should be true")
	}

	if !config.VerboseLogging {
		t.Error("VerboseLogging should be true")
	}

	if !config.UseProperHTTPStatus {
		t.Error("UseProperHTTPStatus should be true")
	}

	if len(config.VerboseLoggingSkipPaths) != 2 {
		t.Errorf("LoggingSkipPaths length = %d, want 2", len(config.VerboseLoggingSkipPaths))
	}
}

func TestCSRFConfigStructure(t *testing.T) {
	expiration := 24 * time.Hour

	csrfConfig := &CSRFConfig{
		KeyLookup:         "header:X-CSRF-Token",
		CookieName:        "csrf_token",
		CookiePath:        "/",
		CookieDomain:      "example.com",
		CookieSameSite:    "Lax",
		CookieSecure:      true,
		CookieHTTPOnly:    true,
		CookieSessionOnly: false,
		SingleUseToken:    false,
		Expiration:        &expiration,
	}

	if csrfConfig.KeyLookup != "header:X-CSRF-Token" {
		t.Errorf("KeyLookup = %s, want header:X-CSRF-Token", csrfConfig.KeyLookup)
	}

	if csrfConfig.CookieName != "csrf_token" {
		t.Errorf("CookieName = %s, want csrf_token", csrfConfig.CookieName)
	}

	if !csrfConfig.CookieSecure {
		t.Error("CookieSecure should be true")
	}

	if csrfConfig.CookiePath != "/" {
		t.Errorf("CookiePath = %s, want /", csrfConfig.CookiePath)
	}

	if csrfConfig.CookieDomain != "example.com" {
		t.Errorf("CookieDomain = %s, want example.com", csrfConfig.CookieDomain)
	}

	if csrfConfig.CookieSameSite != "Lax" {
		t.Errorf("CookieSameSite = %s, want Lax", csrfConfig.CookieSameSite)
	}

	if !csrfConfig.CookieHTTPOnly {
		t.Error("CookieHTTPOnly should be true")
	}

	if csrfConfig.CookieSessionOnly {
		t.Error("CookieSessionOnly should be false")
	}

	if csrfConfig.SingleUseToken {
		t.Error("SingleUseToken should be false")
	}

	if *csrfConfig.Expiration != 24*time.Hour {
		t.Errorf("Expiration = %v, want 24h", *csrfConfig.Expiration)
	}
}

func TestCORSConfigStructure(t *testing.T) {
	corsConfig := &CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Request-ID"},
		MaxAge:           3600,
	}

	if len(corsConfig.AllowOrigins) != 2 {
		t.Errorf("AllowOrigins length = %d, want 2", len(corsConfig.AllowOrigins))
	}

	if len(corsConfig.AllowMethods) != 4 {
		t.Errorf("AllowMethods length = %d, want 4", len(corsConfig.AllowMethods))
	}

	if len(corsConfig.AllowHeaders) != 2 {
		t.Errorf("AllowHeaders length = %d, want 2", len(corsConfig.AllowHeaders))
	}

	if !corsConfig.AllowCredentials {
		t.Error("AllowCredentials should be true")
	}

	if len(corsConfig.ExposeHeaders) != 1 {
		t.Errorf("ExposeHeaders length = %d, want 1", len(corsConfig.ExposeHeaders))
	}

	if corsConfig.MaxAge != 3600 {
		t.Errorf("MaxAge = %d, want 3600", corsConfig.MaxAge)
	}
}

func TestConfigWithNilTimeouts(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		Port:        8080,
		// All timeout fields intentionally nil
	}

	if config.ServiceName != "test-service" {
		t.Errorf("ServiceName = %s, want test-service", config.ServiceName)
	}

	if config.Port != 8080 {
		t.Errorf("Port = %d, want 8080", config.Port)
	}

	if config.ReadTimeout != nil {
		t.Error("ReadTimeout should be nil")
	}

	if config.WriteTimeout != nil {
		t.Error("WriteTimeout should be nil")
	}

	if config.IdleTimeout != nil {
		t.Error("IdleTimeout should be nil")
	}

	if config.GracefulShutdownTimeout != nil {
		t.Error("GracefulShutdownTimeout should be nil")
	}
}
