package aurelion

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config should not return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
			},
			wantErr: false,
		},
		{
			name:    "nil config should return error",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "empty service name should return error",
			config: &Config{
				ServiceName: "",
				Port:        8080,
			},
			wantErr: true,
			errMsg:  "service_name is required",
		},
		{
			name: "port below minimum should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        MinPort - 1,
			},
			wantErr: true,
			errMsg:  "port must be between",
		},
		{
			name: "port above maximum should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        MaxPort + 1,
			},
			wantErr: true,
			errMsg:  "port must be between",
		},
		{
			name: "negative read timeout should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				ReadTimeout: func() *time.Duration { d := time.Duration(-1); return &d }(),
			},
			wantErr: true,
			errMsg:  "read_timeout cannot be negative",
		},
		{
			name: "negative write timeout should return error",
			config: &Config{
				ServiceName:  "Test Service",
				Port:         8080,
				WriteTimeout: func() *time.Duration { d := time.Duration(-1); return &d }(),
			},
			wantErr: true,
			errMsg:  "write_timeout cannot be negative",
		},
		{
			name: "negative idle timeout should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				IdleTimeout: func() *time.Duration { d := time.Duration(-1); return &d }(),
			},
			wantErr: true,
			errMsg:  "idle_timeout cannot be negative",
		},
		{
			name: "negative graceful shutdown timeout should return error",
			config: &Config{
				ServiceName:             "Test Service",
				Port:                    8080,
				GracefulShutdownTimeout: func() *time.Duration { d := time.Duration(-1); return &d }(),
			},
			wantErr: true,
			errMsg:  "graceful_shutdown_timeout cannot be negative",
		},
		{
			name: "negative max body size should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				MaxBodySize: -1,
			},
			wantErr: true,
			errMsg:  "max_body_size cannot be negative",
		},
		{
			name: "negative max concurrent connections should return error",
			config: &Config{
				ServiceName:              "Test Service",
				Port:                     8080,
				MaxConcurrentConnections: -1,
			},
			wantErr: true,
			errMsg:  "max_concurrent_connections cannot be negative",
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
			errMsg:  "cors config is required when enable_cors is true",
		},
		{
			name: "CORS with empty origins should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				EnableCORS:  true,
				CORS: &CORSConfig{
					AllowOrigins: []string{},
					AllowMethods: []string{"GET"},
				},
			},
			wantErr: true,
			errMsg:  "allow_origins is required",
		},
		{
			name: "CORS with invalid method should return error",
			config: &Config{
				ServiceName: "Test Service",
				Port:        8080,
				EnableCORS:  true,
				CORS: &CORSConfig{
					AllowOrigins: []string{"*"},
					AllowMethods: []string{"INVALID"},
				},
			},
			wantErr: true,
			errMsg:  "invalid HTTP method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestCORSConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cors    *CORSConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid CORS config should not return error",
			cors: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET", "POST"},
			},
			wantErr: false,
		},
		{
			name: "empty origins should return error",
			cors: &CORSConfig{
				AllowOrigins: []string{},
				AllowMethods: []string{"GET"},
			},
			wantErr: true,
			errMsg:  "allow_origins is required",
		},
		{
			name: "empty methods should return error",
			cors: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{},
			},
			wantErr: true,
			errMsg:  "allow_methods is required",
		},
		{
			name: "invalid method should return error",
			cors: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"INVALID"},
			},
			wantErr: true,
			errMsg:  "invalid HTTP method",
		},
		{
			name: "case insensitive method validation",
			cors: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"get", "post", "PUT"},
			},
			wantErr: false,
		},
		{
			name: "negative max age should return error",
			cors: &CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET"},
				MaxAge:       -1,
			},
			wantErr: true,
			errMsg:  "max_age cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cors.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestConfig_Merge(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedName   string
		expectedPort   int
		expectedBodySz int
	}{
		{
			name: "empty config should use defaults",
			config: &Config{
				ServiceName: "",
				Port:        0,
			},
			expectedName:   "HTTP Server",
			expectedPort:   DefaultPort,
			expectedBodySz: DefaultMaxBodySize,
		},
		{
			name: "partial config should merge defaults",
			config: &Config{
				ServiceName: "Custom Service",
				Port:        0,
			},
			expectedName: "Custom Service",
			expectedPort: DefaultPort,
		},
		{
			name: "full config should not be overridden",
			config: &Config{
				ServiceName: "Custom Service",
				Port:        9000,
			},
			expectedName: "Custom Service",
			expectedPort: 9000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Merge()

			if result.ServiceName != tt.expectedName {
				t.Errorf("ServiceName = %v, want %v", result.ServiceName, tt.expectedName)
			}

			if result.Port != tt.expectedPort {
				t.Errorf("Port = %v, want %v", result.Port, tt.expectedPort)
			}

			if tt.expectedBodySz > 0 && result.MaxBodySize != tt.expectedBodySz {
				t.Errorf("MaxBodySize = %v, want %v", result.MaxBodySize, tt.expectedBodySz)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ServiceName != "HTTP Server" {
		t.Errorf("Expected ServiceName to be 'HTTP Server', got %s", config.ServiceName)
	}

	if config.Port != DefaultPort {
		t.Errorf("Expected Port to be %d, got %d", DefaultPort, config.Port)
	}

	if config.MaxBodySize != DefaultMaxBodySize {
		t.Errorf("Expected MaxBodySize to be %d, got %d", DefaultMaxBodySize, config.MaxBodySize)
	}

	if config.MaxConcurrentConnections != DefaultMaxConcurrentConnections {
		t.Errorf("Expected MaxConcurrentConnections to be %d, got %d",
			DefaultMaxConcurrentConnections, config.MaxConcurrentConnections)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
