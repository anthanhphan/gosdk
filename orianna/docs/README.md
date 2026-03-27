# Orianna

A production-ready Go framework for building high-performance HTTP and gRPC services with hexagonal architecture, zero-allocation hot paths, secure-by-default configuration, and built-in observability.

## Installation

```bash
go get github.com/anthanhphan/gosdk/orianna
```

## Documentation

| Document | Description |
|----------|-------------|
| **[REST.md](REST.md)** | HTTP server — configuration, routing, binding, middleware, auth, health checks, hooks, client |
| **[GRPC.md](GRPC.md)** | gRPC server — configuration, TLS/mTLS, interceptors, streaming, hooks, health checks, client |

## Architecture

```
orianna/
├── http/                        # ── HTTP Module ──
│   ├── configuration/           # Config types, validation, defaults
│   ├── core/                    # Context (11 ISP interfaces), binding, response, errors
│   ├── middleware/              # Composition: Chain, Optional, SkipForPaths, Timeout...
│   ├── routing/                 # Route builder, groups, registry (atomic registration)
│   ├── server/                  # Server lifecycle, options, graceful shutdown
│   ├── client/                  # HTTP client with retry, circuit breaker, observability
│   ├── engine/                  # ServerEngine interface (Strategy pattern)
│   └── platform/fiber/          # Fiber adapter (swappable implementation)
│
├── grpc/                        # ── gRPC Module ──
│   ├── configuration/           # Config types, validation, defaults
│   ├── core/                    # gRPC Context, StatusError, hooks, carrier
│   ├── interceptor/             # Logging, metrics, recovery, auth, tracing, slow RPC
│   ├── server/                  # Server lifecycle, interceptor chain, service registry
│   └── client/                  # gRPC client with retry, circuit breaker, observability
│
└── shared/                      # ── Shared Utilities ──
    ├── errors/                  # Sentinel errors (ErrInvalidConfig, ErrCircuitOpen...)
    ├── health/                  # Worker-pool health check manager
    ├── hooks/                   # Generic lifecycle hooks (Go generics)
    ├── httputil/                # Header/metadata PII redaction
    ├── observability/           # Pre-computed metric names, code string caching
    ├── requestid/               # UUIDv7 generation + validation
    ├── resilience/              # Circuit breaker, retry with exponential backoff
    └── ctxkeys/                 # Context key constants
```

## Quick Start — HTTP

```go
package main

import (
    "log"
    "github.com/anthanhphan/gosdk/orianna/http/configuration"
    "github.com/anthanhphan/gosdk/orianna/http/core"
    "github.com/anthanhphan/gosdk/orianna/http/server"
)

func main() {
    srv, err := server.NewServer(&configuration.Config{
        ServiceName: "my-api",
        Port:        8080,
    })
    if err != nil {
        log.Fatal(err)
    }

    srv.GET("/", func(ctx core.Context) error {
        return ctx.OK(map[string]any{"message": "Hello, World!"})
    })

    log.Fatal(srv.Start())
}
```

→ **[REST.md](REST.md)** for complete HTTP documentation.

## Quick Start — gRPC

```go
package main

import (
    "log"
    pb "your/proto/package"
    "github.com/anthanhphan/gosdk/orianna/grpc/configuration"
    "github.com/anthanhphan/gosdk/orianna/grpc/server"
)

func main() {
    srv, err := server.NewServer(&configuration.Config{
        ServiceName: "user-service",
        Port:        50051,
    })
    if err != nil {
        log.Fatal(err)
    }

    pb.RegisterUserServiceServer(srv.GRPCServer(), &userServiceImpl{})
    log.Fatal(srv.Start())
}
```

→ **[GRPC.md](GRPC.md)** for complete gRPC documentation.

## Design Principles

| Principle | Implementation |
|-----------|---------------|
| **Hexagonal Architecture** | Domain logic is decoupled from transport adapters (Fiber, gRPC). Swap `engine.ServerEngine` without touching business code. |
| **Interface Segregation** | HTTP `Context` is composed of 11 focused interfaces (`RequestInfo`, `BodyReader`, `ResponseWriter`...). Consumers depend only on what they need. |
| **Zero-Allocation Hot Paths** | `sync.Pool` for response objects and log field slices. Indexed recursive dispatch for interceptors (O(1) overhead). Pre-computed metric names at init time. |
| **Secure by Default** | CSRF cookies default to `Secure=true`, `HTTPOnly=true` via `*bool` (nil = true). PII redacted from logs. TLS 1.3 minimum. UUIDv7 for request IDs (not guessable). |
| **Fail-Fast Validation** | `Config.Validate()` runs at startup. Invalid config (missing service name, TLS+mTLS conflict, bad timeout) immediately returns an error — no silent failures. |
| **Config via Pointers** | Timeout fields use `*time.Duration` — `nil` means "use default", zero means "explicitly zero". This prevents the ambiguity of Go zero values. |

## Performance Characteristics

- **Zero-allocation interceptor dispatch** — indexed recursive dispatch instead of N-closure allocations per request
- **sync.Pool for response objects** — `AcquireSuccessResponse` / `AcquireErrorResponse` on framework hot paths
- **sync.Pool for log field slices** — gRPC logging interceptors reuse `[]any` slices
- **Pre-computed metric names** — string concatenation at construction time, zero alloc per request
- **Pre-built retry lookup maps** — O(1) `map[int]struct{}` instead of O(n) linear scan for retryable status codes
- **Status code string cache** — avoids `strconv.Itoa` on every request for common HTTP/gRPC codes
- **Struct tag caching** — validation metadata parsed once per type by the validator package
- **Concurrent health checks** — worker pool with configurable size
- **Race-free** — verified with `-race` detector across all packages
