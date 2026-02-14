package configuration

import (
	"testing"
	"time"
)

func TestConfigValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		check   func(t *testing.T, err error)
	}{
		{
			name: "valid config should not return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
			},
			wantErr: false,
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			},
		},
		{
			name:    "nil config should return error",
			config:  nil,
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for nil config")
				}
			},
		},
		{
			name: "empty service name should return error",
			config: &Config{
				ServiceName: "",
				Port:        8080,
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for empty service name")
				}
			},
		},
		{
			name: "invalid port should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        -1,
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for invalid port")
				}
			},
		},
		{
			name: "port out of range should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        70000,
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for port out of range")
				}
			},
		},
		{
			name: "negative timeout should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				ReadTimeout: func() *time.Duration {
					d := -1 * time.Second
					return &d
				}(),
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for negative timeout")
				}
			},
		},
		{
			name: "CORS enabled without config should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				EnableCORS:  true,
				CORS:        nil,
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for CORS enabled without config")
				}
			},
		},
		{
			name: "CSRF enabled without config should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				EnableCSRF:  true,
				CSRF:        nil,
			},
			wantErr: true,
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Validate() should return error for CSRF enabled without config")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewConfigValidator()
			err := validator.Validate(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.check(t, err)
		})
	}
}
