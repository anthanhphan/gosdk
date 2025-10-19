# Conflux Package

A simple, type-safe configuration management package for Go applications. Parses JSON and YAML configuration files with environment-based path resolution.

## Installation

```bash
go get github.com/anthanhphan/gosdk/conflux
```

## Quick Start

```go
package main

import (
	"os"

	"github.com/anthanhphan/gosdk/conflux"
	"github.com/anthanhphan/gosdk/logger"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port" json:"port"`
		Name string `yaml:"name" json:"name"`
	} `yaml:"server" json:"server"`
	Logger logger.Config `yaml:"logger" json:"logger"`
}

func main() {
	// Parse configuration based on environment
	config, err := conflux.ParseConfig(
		conflux.GetConfigPathFromEnv(os.Getenv("ENV"), conflux.ExtensionYAML),
		&Config{},
	)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	log.Infof("Server starting on port %d", config.Server.Port)
}
```

## Configuration

### Environment Paths

| Environment  | Path                         |
| ------------ | ---------------------------- |
| `local`      | `./config/config.local`      |
| `qc`         | `./config/config.qc`         |
| `staging`    | `./config/config.staging`    |
| `production` | `./config/config.production` |

### File Extensions

- **JSON** (`.json`) - Default format
- **YAML** (`.yaml`, `.yml`) - Human-readable format

## Usage Examples

### Basic Configuration

```go
type AppConfig struct {
    DatabaseURL string `json:"database_url" yaml:"database_url"`
    Port        int    `json:"port" yaml:"port"`
}

var config AppConfig
parsedConfig, err := conflux.ParseConfig("./config.json", &config)
if err != nil {
    log.Fatal("Failed to parse config:", err)
}
fmt.Printf("Database URL: %s\n", parsedConfig.DatabaseURL)
```

### Environment-Based Configuration

```go
// Get JSON config for production
path := conflux.GetConfigPathFromEnv("production")
// Result: "./config/config.production.json"

// Get YAML config for staging
path := conflux.GetConfigPathFromEnv("staging", "yaml")
// Result: "./config/config.staging.yaml"
```

### Configuration Files

#### JSON Configuration File

```json
{
  "server": {
    "port": 8080,
    "name": "example"
  },
  "logger": {
    "level": "debug",
    "encoding": "json"
  }
}
```

#### YAML Configuration File

```yaml
server:
  port: 8080
  name: example

logger:
  level: debug
  encoding: json
  disable_caller: false
  is_development: true
```
