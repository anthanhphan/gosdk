// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package conflux

// Supported file extensions for configuration files
const (
	ExtensionJSON = "json"
	ExtensionYAML = "yaml"
	ExtensionYML  = "yml"
)

// Configuration file paths for different environments
const (
	QcConfigPath         = "./config/config.qc"
	StagingConfigPath    = "./config/config.staging"
	ProductionConfigPath = "./config/config.production"
	LocalConfigPath      = "./config/config.local"
)

// Default configuration values
const (
	DefaultConfigPath = LocalConfigPath
	DefaultExtension  = ExtensionJSON
)
