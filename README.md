# GoSDK

<div align="center">

**A collection of high-performance Go packages. Built from scratch, ready for production.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

[**Get Started**](#quick-start) • [**Packages**](#packages) • [**Examples**](#try-it-out)

</div>

---

## What's Inside?

GoSDK provides essential packages that make Go development easier and faster:

| Package | What it does | Why use it |
|---------|-------------|-----------|
| **[orianna](./orianna)** | HTTP framework | Production-ready server with routing, middleware, validation |
| **[jcodec](./jcodec)** | Fast JSON encoding/decoding | 5x faster than standard library |
| **[logger](./logger)** | Structured logging | Zero-allocation, async support |
| **[goroutine](./goroutine)** | Safe concurrent code | Auto panic recovery |
| **[conflux](./conflux)** | Config management | JSON/YAML with type safety |
| **[utils](./utils)** | Common utilities | File I/O, environment helpers |

> **New to Go?** Each package has its own README with detailed guides and examples. Start with [orianna](./orianna/docs/README.md), [logger](./logger/docs/README.md), or [jcodec](./jcodec/docs/README.md)!

---

## Quick Start

### 1. Install a package

```bash
# Choose what you need
go get github.com/anthanhphan/gosdk/orianna
go get github.com/anthanhphan/gosdk/logger
go get github.com/anthanhphan/gosdk/jcodec
```

### 2. Use it in your code

```go
package main

import "github.com/anthanhphan/gosdk/logger"

func main() {
    // Initialize logger
    undo := logger.InitDefaultLogger()
    defer undo()

    // Start logging!
    logger.Info("Hello, GoSDK!")
    logger.Infow("User logged in", "username", "alice", "id", 42)
}
```

### 3. Run it

```bash
go run main.go
```

That's it!

---

## Try It Out

Each package has runnable examples. Clone this repo and try:

```bash
# Clone the repository
git clone https://github.com/anthanhphan/gosdk
cd gosdk

# Try orianna example
go run orianna/docs/examples/complete/main.go

# Try logger example
go run logger/docs/example/main.go

# Try jcodec example
go run jcodec/docs/example/main.go

# Try goroutine example
go run goroutine/docs/example/main.go
```

---

## Packages

### orianna - HTTP Framework

Production-ready HTTP framework with routing, middleware, typed request binding, and more.

```go
import (
    "github.com/anthanhphan/gosdk/orianna"
    "github.com/anthanhphan/gosdk/orianna/pkg/configuration"
)

srv, _ := orianna.NewServer(&configuration.Config{
    ServiceName: "my-api",
    Port:        8080,
})

srv.GET("/users", listUsersHandler)
srv.Protected().GET("/profile", profileHandler)
srv.Run()
```

[Full docs](./orianna/docs/README.md) • [Examples](./orianna/docs/examples/complete/main.go)

---

### jcodec - Fast JSON

Drop-in replacement for `encoding/json` that's **5x faster**.

```go
import "github.com/anthanhphan/gosdk/jcodec"

data, err := jcodec.Marshal(user)          // Like json.Marshal
jcodec.Unmarshal(data, &user)              // Like json.Unmarshal
pretty, _ := jcodec.MarshalIndent(user, "", "  ")  // Pretty print
```

[Full docs](./jcodec/docs/README.md) • [Examples](./jcodec/docs/example/main.go)

---

### logger - Structured Logging

Built from scratch, zero dependencies, blazing fast.

```go
import "github.com/anthanhphan/gosdk/logger"

// Quick start
undo := logger.InitDefaultLogger()
defer undo()

logger.Info("Simple message")
logger.Infow("Structured log", "key", "value")
```

[Full docs](./logger/docs/README.md) • [Examples](./logger/docs/example/main.go)

---

### goroutine - Safe Concurrency

Run goroutines without worrying about panics.

```go
import routine "github.com/anthanhphan/gosdk/goroutine"

// Panics are caught and logged automatically
routine.Run(func() {
    // Your code here
    panic("No problem!") // This won't crash your app
})
```

[Full docs](./goroutine/docs/README.md) • [Examples](./goroutine/docs/example/main.go)

---

### conflux - Easy Config

Load JSON/YAML configs with environment support.

```go
import "github.com/anthanhphan/gosdk/conflux"

type Config struct {
    Port int `json:"port"`
}

config, err := conflux.ParseConfig("config.json", &Config{})
```

[Full docs](./conflux/docs/README.md) • [Examples](./conflux/docs/example/main.go)

---

## Explore the Code

**For developers:**

```bash
# Format code
make fmt

# Run tests
make test_race

# Run all checks (before committing)
make all

# See all available commands
make help
```

**Project structure:**
```
gosdk/
├── orianna/         # HTTP framework package
│   ├── docs/        # Documentation
│   └── pkg/         # Sub-packages (core, server, routing, middleware, ...)
├── jcodec/          # JSON codec package
│   ├── docs/        # Documentation
│   └── *.go         # Source code
├── logger/          # Logger package
├── goroutine/       # Goroutine package
├── conflux/         # Config package
└── utils/           # Utilities
```

> Each package folder has its own `docs/README.md` with detailed information.

---

## Need Help?

- **Read the docs**: Each package has detailed guides in `docs/README.md`
- **Check examples**: Look in `docs/example/` folders
- **Found a bug?**: [Open an issue](https://github.com/anthanhphan/gosdk/issues)
- **Have an idea?**: [Start a discussion](https://github.com/anthanhphan/gosdk/discussions)

---

## Contributing

Want to help? We'd love that! 

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/cool-thing`)
3. Make changes and run `make all` to verify
4. Commit (`git commit -m 'Add cool thing'`)
5. Push (`git push origin feature/cool-thing`)
6. Open a Pull Request

See detailed guidelines in each package's documentation.

---

## License

MIT License - see [LICENSE](./LICENSE)

---

<div align="center">

**Made with ❤️ by [anthanhphan](https://github.com/anthanhphan)**

Star this repo if you find it helpful!

</div>

