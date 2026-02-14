package configuration

import (
	"testing"
	"time"
)

func TestConfigValidator_Validate_WriteTimeout_Negative(t *testing.T) {
	validator := NewConfigValidator()
	d := -1 * time.Second
	config := &Config{
		ServiceName:  "test",
		Port:         8080,
		WriteTimeout: &d,
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for negative write_timeout")
	}
}

func TestConfigValidator_Validate_IdleTimeout_Negative(t *testing.T) {
	validator := NewConfigValidator()
	d := -1 * time.Second
	config := &Config{
		ServiceName: "test",
		Port:        8080,
		IdleTimeout: &d,
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for negative idle_timeout")
	}
}

func TestConfigValidator_Validate_GracefulShutdownTimeout_Negative(t *testing.T) {
	validator := NewConfigValidator()
	d := -1 * time.Second
	config := &Config{
		ServiceName:             "test",
		Port:                    8080,
		GracefulShutdownTimeout: &d,
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for negative graceful_shutdown_timeout")
	}
}

func TestConfigValidator_Validate_MaxBodySize_Negative(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName: "test",
		Port:        8080,
		MaxBodySize: -1,
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for negative max_body_size")
	}
}

func TestConfigValidator_Validate_MaxConcurrentConnections_Negative(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName:              "test",
		Port:                     8080,
		MaxConcurrentConnections: -1,
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for negative max_concurrent_connections")
	}
}

func TestConfigValidator_Validate_CORS_EmptyOrigins(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName: "test",
		Port:        8080,
		EnableCORS:  true,
		CORS: &CORSConfig{
			AllowOrigins: []string{},
			AllowMethods: []string{"GET"},
		},
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for CORS with empty origins")
	}
}

func TestConfigValidator_Validate_CORS_EmptyMethods(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName: "test",
		Port:        8080,
		EnableCORS:  true,
		CORS: &CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{},
		},
	}
	err := validator.Validate(config)
	if err == nil {
		t.Error("Validate() should return error for CORS with empty methods")
	}
}

func TestConfigValidator_Validate_CORS_Valid(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName: "test",
		Port:        8080,
		EnableCORS:  true,
		CORS: &CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST"},
		},
	}
	err := validator.Validate(config)
	if err != nil {
		t.Errorf("Validate() should not return error for valid CORS config, got %v", err)
	}
}

func TestConfigValidator_Validate_CSRF_Valid(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName: "test",
		Port:        8080,
		EnableCSRF:  true,
		CSRF:        &CSRFConfig{KeyLookup: "header:X-CSRF-Token"},
	}
	err := validator.Validate(config)
	if err != nil {
		t.Errorf("Validate() should not return error for valid CSRF config, got %v", err)
	}
}

func TestConfigValidator_Validate_Port_Zero(t *testing.T) {
	validator := NewConfigValidator()
	config := &Config{
		ServiceName: "test",
		Port:        0,
	}
	// Port 0 is valid (auto-assign)
	err := validator.Validate(config)
	if err != nil {
		t.Errorf("Validate() should accept port 0, got %v", err)
	}
}

func TestConfigValidator_Validate_AllTimeouts_Valid(t *testing.T) {
	validator := NewConfigValidator()
	read := 5 * time.Second
	write := 10 * time.Second
	idle := 30 * time.Second
	grace := 15 * time.Second
	config := &Config{
		ServiceName:             "test",
		Port:                    8080,
		ReadTimeout:             &read,
		WriteTimeout:            &write,
		IdleTimeout:             &idle,
		GracefulShutdownTimeout: &grace,
	}
	err := validator.Validate(config)
	if err != nil {
		t.Errorf("Validate() should accept valid timeouts, got %v", err)
	}
}
