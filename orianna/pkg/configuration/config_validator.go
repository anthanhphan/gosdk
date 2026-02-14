// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package configuration

import (
	"errors"
	"fmt"
)

// ConfigValidator validates server configuration.
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator.
//
// Output:
//   - *ConfigValidator: A new validator instance
//
// Example:
//
//	validator := configuration.NewConfigValidator()
//	err := validator.Validate(config)
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate validates the server configuration.
//
// Input:
//   - config: The configuration to validate
//
// Output:
//   - error: Validation error if any
//
// Example:
//
//	validator := configuration.NewConfigValidator()
//	if err := validator.Validate(config); err != nil {
//	    log.Fatal(err)
//	}
func (cv *ConfigValidator) Validate(config *Config) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if config.ServiceName == "" {
		return errors.New("service_name is required")
	}
	if config.Port < 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535, got %d", config.Port)
	}

	if err := validateTimeouts(config); err != nil {
		return err
	}
	if err := validateLimits(config); err != nil {
		return err
	}
	return validateCORSCSRF(config)
}

func validateTimeouts(config *Config) error {
	if config.ReadTimeout != nil && *config.ReadTimeout < 0 {
		return errors.New("read_timeout cannot be negative")
	}
	if config.WriteTimeout != nil && *config.WriteTimeout < 0 {
		return errors.New("write_timeout cannot be negative")
	}
	if config.IdleTimeout != nil && *config.IdleTimeout < 0 {
		return errors.New("idle_timeout cannot be negative")
	}
	if config.GracefulShutdownTimeout != nil && *config.GracefulShutdownTimeout < 0 {
		return errors.New("graceful_shutdown_timeout cannot be negative")
	}
	return nil
}

func validateLimits(config *Config) error {
	if config.MaxBodySize < 0 {
		return errors.New("max_body_size cannot be negative")
	}
	if config.MaxConcurrentConnections < 0 {
		return errors.New("max_concurrent_connections cannot be negative")
	}
	return nil
}

func validateCORSCSRF(config *Config) error {
	if config.EnableCORS {
		if config.CORS == nil {
			return errors.New("cors config is required when enable_cors is true")
		}
		if len(config.CORS.AllowOrigins) == 0 {
			return errors.New("cors allow_origins is required")
		}
		if len(config.CORS.AllowMethods) == 0 {
			return errors.New("cors allow_methods is required")
		}
	}
	if config.EnableCSRF && config.CSRF == nil {
		return errors.New("csrf config is required when enable_csrf is true")
	}
	return nil
}
