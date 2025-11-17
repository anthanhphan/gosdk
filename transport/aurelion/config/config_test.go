package config

import (
	"net/http/httptest"
	"testing"
	"time"

	rctx "github.com/anthanhphan/gosdk/transport/aurelion/internal/runtimectx"
	"github.com/gofiber/fiber/v2"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				ServiceName: "test",
				Port:        8080,
			},
			wantErr: false,
		},
		{
			name:    "invalid port",
			cfg:     &Config{ServiceName: "svc", Port: 70000},
			wantErr: true,
		},
		{
			name: "negative body size",
			cfg: &Config{
				ServiceName: "svc",
				Port:        8080,
				MaxBodySize: -1,
			},
			wantErr: true,
		},
		{
			name: "negative max concurrent connections",
			cfg: &Config{
				ServiceName:              "svc",
				Port:                     8080,
				MaxConcurrentConnections: -1,
			},
			wantErr: true,
		},
		{
			name: "enable CORS without config",
			cfg: &Config{
				ServiceName: "svc",
				Port:        8080,
				EnableCORS:  true,
				CORS:        nil,
			},
			wantErr: true,
		},
		{
			name: "enable CSRF without config",
			cfg: &Config{
				ServiceName: "svc",
				Port:        8080,
				EnableCSRF:  true,
				CSRF:        nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.cfg == nil {
			continue
		}
		cfg := tt.cfg.Merge()
		err := cfg.Validate()
		if (err != nil) != tt.wantErr {
			t.Fatalf("%s: Validate() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestCORSConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *CORSConfig
		wantErr bool
	}{
		{
			name: "valid CORS config",
			cfg: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET", "POST"},
			},
			wantErr: false,
		},
		{
			name: "missing allow origins",
			cfg: &CORSConfig{
				AllowMethods: []string{"GET"},
			},
			wantErr: true,
		},
		{
			name: "missing allow methods",
			cfg: &CORSConfig{
				AllowOrigins: []string{"*"},
			},
			wantErr: true,
		},
		{
			name: "invalid HTTP method",
			cfg: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"INVALID"},
			},
			wantErr: true,
		},
		{
			name: "negative max age",
			cfg: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET"},
				MaxAge:       -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSRFConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *CSRFConfig
		wantErr bool
	}{
		{
			name: "valid CSRF config",
			cfg: &CSRFConfig{
				KeyLookup:  "header:X-CSRF-Token",
				CookieName: "csrf_token",
			},
			wantErr: false,
		},
		{
			name: "invalid key_lookup format",
			cfg: &CSRFConfig{
				KeyLookup: "invalid-format",
			},
			wantErr: true,
		},
		{
			name: "invalid key_lookup source",
			cfg: &CSRFConfig{
				KeyLookup: "invalid:X-Token",
			},
			wantErr: true,
		},
		{
			name: "empty key_lookup key",
			cfg: &CSRFConfig{
				KeyLookup: "header:",
			},
			wantErr: true,
		},
		{
			name: "negative expiration",
			cfg: &CSRFConfig{
				Expiration: func() *time.Duration { d := -1 * time.Second; return &d }(),
			},
			wantErr: true,
		},
		{
			name: "invalid cookie_same_site",
			cfg: &CSRFConfig{
				CookieSameSite: "Invalid",
			},
			wantErr: true,
		},
		{
			name: "valid cookie_same_site Strict",
			cfg: &CSRFConfig{
				CookieSameSite: "Strict",
			},
			wantErr: false,
		},
		{
			name: "valid cookie_same_site Lax",
			cfg: &CSRFConfig{
				CookieSameSite: "Lax",
			},
			wantErr: false,
		},
		{
			name: "valid cookie_same_site None",
			cfg: &CSRFConfig{
				CookieSameSite: "None",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoreInContextAndFromContext(t *testing.T) {
	app := fiber.New()
	cfg := &Config{ServiceName: "test", Port: 8080}

	app.Get("/", func(c *fiber.Ctx) error {
		ctx := rctx.NewFiberContext(c)
		StoreInContext(ctx, cfg)
		if got := FromContext(ctx); got == nil || got.ServiceName != "test" {
			t.Fatalf("FromContext() = %v", got)
		}
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}
