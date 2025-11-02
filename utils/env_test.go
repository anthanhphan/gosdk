package utils

import (
	"os"
	"testing"
)

func TestGetEnvironment(t *testing.T) {
	// Save original ENV value
	originalEnv := os.Getenv("ENV")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	tests := []struct {
		name         string
		envValue     string
		want         string
		shouldSetEnv bool
	}{
		{
			name:         "ENV not set should return local",
			envValue:     "",
			want:         EnvLocal,
			shouldSetEnv: false,
		},
		{
			name:         "ENV set to local should return local",
			envValue:     EnvLocal,
			want:         EnvLocal,
			shouldSetEnv: true,
		},
		{
			name:         "ENV set to qc should return qc",
			envValue:     EnvQC,
			want:         EnvQC,
			shouldSetEnv: true,
		},
		{
			name:         "ENV set to staging should return staging",
			envValue:     EnvStaging,
			want:         EnvStaging,
			shouldSetEnv: true,
		},
		{
			name:         "ENV set to production should return production",
			envValue:     EnvProduction,
			want:         EnvProduction,
			shouldSetEnv: true,
		},
		{
			name:         "ENV set to empty string should return local",
			envValue:     "",
			want:         EnvLocal,
			shouldSetEnv: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.shouldSetEnv {
				os.Setenv("ENV", tt.envValue)
			} else {
				os.Unsetenv("ENV")
			}

			// Execute
			got := GetEnvironment()

			// Verify
			if got != tt.want {
				t.Errorf("GetEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{
			name:    "local environment should not return error",
			env:     EnvLocal,
			wantErr: false,
		},
		{
			name:    "qc environment should not return error",
			env:     EnvQC,
			wantErr: false,
		},
		{
			name:    "staging environment should not return error",
			env:     EnvStaging,
			wantErr: false,
		},
		{
			name:    "production environment should not return error",
			env:     EnvProduction,
			wantErr: false,
		},
		{
			name:    "invalid environment should return error",
			env:     "invalid",
			wantErr: true,
		},
		{
			name:    "empty environment should return error",
			env:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironment(tt.env)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !contains(err.Error(), tt.env) {
				t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.env)
			}
		})
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
