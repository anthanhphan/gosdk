# Conflux Package

A simple, type-safe configuration parser for Go applications. Supports JSON and YAML formats.

## Installation

```bash
go get github.com/anthanhphan/gosdk/conflux
```

## Quick Start

```go
package main

import (
	"github.com/anthanhphan/gosdk/conflux"
	"github.com/anthanhphan/gosdk/logger"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port" json:"port"`
		Name string `yaml:"name" json:"name"`
	} `yaml:"server" json:"server"`
	Logger logger.Config `yaml:"logger" json:"logger"`
}

func main() {
	// MustLoad panics on error — ideal for program init
	config := conflux.MustLoad[Config]("./config/config.local.yaml")

	log.Infof("Server starting on port %d", config.Server.Port)
}
```

## API

### `Load[T](path string) (*T, error)`

Parses a configuration file and returns the typed config. Returns an error if the file cannot be read or parsed.

```go
config, err := conflux.Load[Config]("./config/app.yaml")
if err != nil {
    log.Fatal(err)
}
```

### `MustLoad[T](path string) *T`

Same as `Load`, but panics on error. Intended for `main()` or `init()`.

```go
config := conflux.MustLoad[Config]("./config/app.yaml")
```

### Supported File Extensions

- **JSON** (`.json`)
- **YAML** (`.yaml`, `.yml`)

## Configuration File Examples

#### YAML

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

#### JSON

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
