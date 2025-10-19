package conflux

import (
	"os"
	"testing"
)

// TestConfig represents a test configuration structure
type TestConfig struct {
	DatabaseURL string `json:"database_url" yaml:"database_url"`
	Port        int    `json:"port" yaml:"port"`
	Debug       bool   `json:"debug" yaml:"debug"`
}

func TestParseConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "conflux_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create config directory
	if err := os.Mkdir("config", 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Test data
	testConfig := TestConfig{
		DatabaseURL: "postgres://localhost:5432/testdb",
		Port:        8080,
		Debug:       true,
	}

	tests := []struct {
		name        string
		path        string
		content     string
		wantErr     bool
		errContains string
		setup       func() error
	}{
		{
			name:    "valid JSON config should parse successfully",
			path:    "config.json",
			content: `{"database_url":"postgres://localhost:5432/testdb","port":8080,"debug":true}`,
			wantErr: false,
			setup: func() error {
				return os.WriteFile("config.json", []byte(`{"database_url":"postgres://localhost:5432/testdb","port":8080,"debug":true}`), 0644)
			},
		},
		{
			name:    "valid YAML config should parse successfully",
			path:    "config.yaml",
			content: "database_url: postgres://localhost:5432/testdb\nport: 8080\ndebug: true",
			wantErr: false,
			setup: func() error {
				return os.WriteFile("config.yaml", []byte("database_url: postgres://localhost:5432/testdb\nport: 8080\ndebug: true"), 0644)
			},
		},
		{
			name:    "valid YML config should parse successfully",
			path:    "config.yml",
			content: "database_url: postgres://localhost:5432/testdb\nport: 8080\ndebug: true",
			wantErr: false,
			setup: func() error {
				return os.WriteFile("config.yml", []byte("database_url: postgres://localhost:5432/testdb\nport: 8080\ndebug: true"), 0644)
			},
		},
		{
			name:        "empty path should return error",
			path:        "",
			wantErr:     true,
			errContains: "config path is required",
		},
		{
			name:        "unsupported extension should return error",
			path:        "config.xml",
			wantErr:     true,
			errContains: "unsupported file extension: xml",
		},
		{
			name:        "non-existent file should return error",
			path:        "nonexistent.json",
			wantErr:     true,
			errContains: "no such file or directory",
		},
		{
			name:        "invalid JSON should return error",
			path:        "invalid.json",
			wantErr:     true,
			errContains: "failed to unmarshal json",
			setup: func() error {
				return os.WriteFile("invalid.json", []byte(`{"invalid": json}`), 0644)
			},
		},
		{
			name:        "invalid YAML should return error",
			path:        "invalid.yaml",
			wantErr:     true,
			errContains: "failed to unmarshal yaml",
			setup: func() error {
				return os.WriteFile("invalid.yaml", []byte("invalid: yaml: content"), 0644)
			},
		},
		{
			name:        "directory traversal should return error",
			path:        "../config.json",
			wantErr:     true,
			errContains: "invalid path",
		},
		{
			name:        "absolute path should return error",
			path:        "/etc/passwd",
			wantErr:     true,
			errContains: "unsupported file extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Execute
			var config TestConfig
			result, err := ParseConfig(tt.path, &config)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errContains)
				}
				if result != nil {
					t.Error("Expected nil result for error case")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result for success case")
				}
				if result != nil && *result != testConfig {
					t.Errorf("ParseConfig() = %v, want %v", *result, testConfig)
				}
			}
		})
	}
}

func TestGetConfigPathFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		ext      []string
		expected string
	}{
		{
			name:     "qc environment with default extension should return qc config path",
			env:      EnvQC,
			ext:      []string{},
			expected: "./config/config.qc.json",
		},
		{
			name:     "staging environment with yaml extension should return staging yaml config path",
			env:      EnvStaging,
			ext:      []string{"yaml"},
			expected: "./config/config.staging.yaml",
		},
		{
			name:     "production environment with yml extension should return production yml config path",
			env:      EnvProduction,
			ext:      []string{"yml"},
			expected: "./config/config.production.yml",
		},
		{
			name:     "local environment with json extension should return local json config path",
			env:      EnvLocal,
			ext:      []string{"json"},
			expected: "./config/config.local.json",
		},
		{
			name:     "unknown environment should return default config path",
			env:      "unknown",
			ext:      []string{},
			expected: "./config/config.local.json",
		},
		{
			name:     "empty environment should return default config path",
			env:      "",
			ext:      []string{},
			expected: "./config/config.local.json",
		},
		{
			name:     "environment with empty extension should use default extension",
			env:      EnvQC,
			ext:      []string{""},
			expected: "./config/config.qc.json",
		},
		{
			name:     "environment with multiple extensions should use first one",
			env:      EnvStaging,
			ext:      []string{"yaml", "json"},
			expected: "./config/config.staging.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigPathFromEnv(tt.env, tt.ext...)
			if result != tt.expected {
				t.Errorf("GetConfigPathFromEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidExtension(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected bool
	}{
		{"json extension should be valid", ExtensionJSON, true},
		{"yaml extension should be valid", ExtensionYAML, true},
		{"yml extension should be valid", ExtensionYML, true},
		{"xml extension should be invalid", "xml", false},
		{"txt extension should be invalid", "txt", false},
		{"empty extension should be invalid", "", false},
		{"uppercase JSON should be invalid", "JSON", false},
		{"mixed case yaml should be invalid", "YaMl", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidExtension(tt.ext)
			if result != tt.expected {
				t.Errorf("isValidExtension(%s) = %v, want %v", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestUnmarshalConfig(t *testing.T) {
	testData := []byte(`{"database_url":"postgres://localhost:5432/testdb","port":8080,"debug":true}`)
	yamlData := []byte("database_url: postgres://localhost:5432/testdb\nport: 8080\ndebug: true")

	tests := []struct {
		name        string
		data        []byte
		ext         string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid JSON data should unmarshal successfully",
			data:    testData,
			ext:     ExtensionJSON,
			wantErr: false,
		},
		{
			name:    "valid YAML data should unmarshal successfully",
			data:    yamlData,
			ext:     ExtensionYAML,
			wantErr: false,
		},
		{
			name:    "valid YML data should unmarshal successfully",
			data:    yamlData,
			ext:     ExtensionYML,
			wantErr: false,
		},
		{
			name:        "unsupported extension should return error",
			data:        testData,
			ext:         "xml",
			wantErr:     true,
			errContains: "unsupported file extension: xml",
		},
		{
			name:        "invalid JSON data should return error",
			data:        []byte(`{"invalid": json}`),
			ext:         ExtensionJSON,
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name:        "invalid YAML data should return error",
			data:        []byte("invalid: yaml: content"),
			ext:         ExtensionYAML,
			wantErr:     true,
			errContains: "mapping values are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config TestConfig
			err := unmarshalConfig(tt.data, tt.ext, &config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

// Helper function to check substring containment
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
