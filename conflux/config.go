// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package conflux

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/utils"
	"github.com/anthanhphan/gosdk/validator"
	"gopkg.in/yaml.v3"
)

// Load parses a configuration file at the given path and returns the parsed config.
// Supported formats: JSON (.json), YAML (.yaml, .yml).
// After parsing, struct validation tags are automatically validated.
//
// Example:
//
//	config, err := conflux.Load[AppConfig]("./config/config.local.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load[T any](path string) (*T, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is required")
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if !validExts[ext] {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	data, err := utils.ReadFileSecurely(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg T
	if err := unmarshal(data, ext, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", ext, err)
	}

	if err := validator.Validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// MustLoad is like Load but panics on error. Use in main() or init().
//
// Example:
//
//	config := conflux.MustLoad[AppConfig]("./config/config.local.yaml")
func MustLoad[T any](path string) *T {
	cfg, err := Load[T](path)
	if err != nil {
		panic(fmt.Sprintf("conflux: %v", err))
	}
	return cfg
}

// validExts is the set of supported config file extensions.
var validExts = map[string]bool{
	ExtensionJSON: true,
	ExtensionYAML: true,
	ExtensionYML:  true,
}

// unmarshal decodes data into dst based on file extension.
func unmarshal(data []byte, ext string, dst any) error {
	switch ext {
	case ExtensionJSON:
		return jcodec.Unmarshal(data, dst)
	case ExtensionYAML, ExtensionYML:
		return yaml.Unmarshal(data, dst)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}
