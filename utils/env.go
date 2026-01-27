// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"fmt"
	"os"
)

// Environment name constants for different deployment environments
const (
	// EnvLocal represents local development environment
	EnvLocal = "local"
	// EnvQC represents quality control/testing environment
	EnvQC = "qc"
	// EnvStaging represents staging environment
	EnvStaging = "staging"
	// EnvProduction represents production environment
	EnvProduction = "production"
)

// GetEnvironment returns the current environment name from environment variables.
// Defaults to EnvLocal if not set.
//
// Input:
//   - None
//
// Output:
//   - string: The environment name (local, qc, staging, production)
//
// Example:
//
//	env := utils.GetEnvironment()
//	if env == utils.EnvProduction {
//	    logger.InitProductionLogger()
//	} else {
//	    logger.InitDefaultLogger()
//	}
func GetEnvironment() string {
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return EnvLocal
}

// validEnvs is a map of valid environment names for O(1) lookup.
var validEnvs = map[string]struct{}{
	EnvLocal:      {},
	EnvQC:         {},
	EnvStaging:    {},
	EnvProduction: {},
}

// ValidateEnvironment returns an error if the environment is invalid.
//
// Input:
//   - env: The environment name to validate
//
// Output:
//   - error: Error if environment is invalid, nil otherwise
//
// Example:
//
//	if err := utils.ValidateEnvironment(env); err != nil {
//	    log.Fatal(err)
//	}
func ValidateEnvironment(env string) error {
	if _, ok := validEnvs[env]; ok {
		return nil
	}
	return fmt.Errorf("invalid environment: %s (must be one of: %s, %s, %s, %s)",
		env, EnvLocal, EnvQC, EnvStaging, EnvProduction)
}
