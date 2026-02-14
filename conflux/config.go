// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package conflux

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/utils"
	"gopkg.in/yaml.v2"
)

// Environment constants are imported from utils for consistency
const (
	EnvLocal      = utils.EnvLocal
	EnvQC         = utils.EnvQC
	EnvStaging    = utils.EnvStaging
	EnvProduction = utils.EnvProduction
)

// ParseConfig parses a configuration file and unmarshals it into the provided model.
//
// Input:
//   - path: The file path to the configuration file
//   - model: A pointer to the struct that will hold the parsed configuration
//
// Output:
//   - *T: The parsed configuration model
//   - error: Any error that occurred during parsing
//
// Example:
//
//	type AppConfig struct {
//	    DatabaseURL string `json:"database_url" yaml:"database_url"`
//	    Port        int    `json:"port" yaml:"port"`
//	}
//
//	var config AppConfig
//	parsedConfig, err := ParseConfig("./config.json", &config)
//	if err != nil {
//	    log.Fatal("Failed to parse config:", err)
//	}
//	fmt.Printf("Database URL: %s\n", parsedConfig.DatabaseURL)
func ParseConfig[T any](path string, model *T) (*T, error) {
	log := logger.NewLoggerWithFields(logger.String("prefix", "conflux::ParseConfig"))

	if path == "" {
		log.Error("config path is required")
		return nil, fmt.Errorf("config path is required")
	}

	// Validate file extension
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if !isValidExtension(ext) {
		log.Errorf("unsupported file extension %s", ext)
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Read file safely to prevent directory traversal
	data, err := utils.ReadFileSecurely(path)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	// Parse based on file extension
	if err := unmarshalConfig(data, ext, model); err != nil {
		log.Errorf("failed to unmarshal %s: %v", ext, err)
		return nil, fmt.Errorf("failed to unmarshal %s: %w", ext, err)
	}

	return model, nil
}

// GetConfigPathFromEnv generates a configuration file path based on environment and file extension.
//
// Input:
//   - env: The environment name (qc, staging, production, local)
//   - ext: Optional file extension (defaults to json)
//
// Output:
//   - string: The generated configuration file path
//
// Example:
//
//	// Get JSON config for production
//	path := GetConfigPathFromEnv("production")
//	// Result: "./config/config.production.json"
//
//	// Get YAML config for staging
//	path := GetConfigPathFromEnv("staging", "yaml")
//	// Result: "./config/config.staging.yaml"
func GetConfigPathFromEnv(env string, ext ...string) string {
	fileExt := DefaultExtension
	if len(ext) > 0 && ext[0] != "" {
		fileExt = ext[0]
	}

	configPaths := map[string]string{
		EnvQC:         QcConfigPath,
		EnvStaging:    StagingConfigPath,
		EnvProduction: ProductionConfigPath,
		EnvLocal:      LocalConfigPath,
	}

	basePath, exists := configPaths[env]
	if !exists {
		basePath = DefaultConfigPath
	}

	return fmt.Sprintf("%s.%s", basePath, fileExt)
}

// isValidExtension checks if the provided file extension is supported.
func isValidExtension(ext string) bool {
	validExts := map[string]bool{
		ExtensionJSON: true,
		ExtensionYAML: true,
		ExtensionYML:  true,
	}
	return validExts[ext]
}

// unmarshalConfig unmarshals the data into the model based on the file extension.
func unmarshalConfig[T any](data []byte, ext string, model *T) error {
	switch ext {
	case ExtensionJSON:
		return jcodec.Unmarshal(data, model)
	case ExtensionYAML, ExtensionYML:
		return yaml.Unmarshal(data, model)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}
