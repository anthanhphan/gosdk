# GoSDK

<div align="center">

**A collection of high-performance Go packages. Built from scratch, ready for production.**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A590%25-brightgreen)](./code_report/coverage.html)

[**Get Started**](#quick-start) • [**Packages**](#packages) • [**Examples**](#try-it-out)

</div>

---

## What's Inside?

GoSDK provides essential packages that make Go development easier and faster:

| Package | What it does | Why use it |
|---------|-------------|-----------|
| **[orianna](./orianna)** | HTTP + gRPC framework | Hexagonal architecture, mTLS, circuit breaker, interceptors, zero-alloc middleware |
| **[jcodec](./jcodec)** | Fast JSON encoding/decoding | Auto-selects Sonic (AMD64) or goccy (ARM64), 1.5–5× faster than `encoding/json` |
| **[logger](./logger)** | Structured logging | 2.5M logs/sec, 2 allocs/op, async output, zap-level performance |
| **[metrics](./metrics)** | Application metrics | Prometheus counters/gauges/histograms, isolated registries, built-in HTTP handler |
| **[redis](./redis)** | Observable Redis client | Standalone + Sentinel, auto Prometheus metrics + structured logging per command |
| **[tracing](./tracing)** | Distributed tracing | OpenTelemetry with OTLP export to Tempo, Jaeger, or any OTLP collector |
| **[validator](./validator)** | Struct validation | Tag-based rules, zero-alloc after first call, nested struct + slice dive support |
| **[goroutine](./goroutine)** | Safe concurrent code | Run, Group, FanOut, ForEach, WorkerPool — all with auto panic recovery |
| **[conflux](./conflux)** | Config management | Generic `Load[T]` for JSON/YAML with type safety |
| **[utils](./utils)** | Common utilities | Secure file I/O, environment detection, panic location tracking |

> **New to Go?** Each package has its own README with detailed guides and examples. Start with [orianna](./orianna/docs/README.md), [logger](./logger/docs/README.md), or [jcodec](./jcodec/docs/README.md)!

---

## Quick Start

### 1. Install a package

```bash
# Choose what you need — each package is independently importable
go get github.com/anthanhphan/gosdk/orianna
go get github.com/anthanhphan/gosdk/logger
go get github.com/anthanhphan/gosdk/jcodec
go get github.com/anthanhphan/gosdk/tracing
go get github.com/anthanhphan/gosdk/metrics
go get github.com/anthanhphan/gosdk/validator
```

### 2. Use it in your code

```go
package main

import "github.com/anthanhphan/gosdk/logger"

func main() {
    // Initialize logger (2.5M logs/sec, 2 allocs/op)
    undo := logger.InitDefaultLogger()
    defer undo()

    // Three logging styles: simple, formatted, structured
    logger.Info("Hello, GoSDK!")
    logger.Infof("Listening on port %d", 8080)
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

# Orianna — HTTP server examples
go run orianna/docs/examples/http/server/quickstart/main.go     # minimal REST API
go run orianna/docs/examples/http/server/complete/main.go        # full-featured with auth, tracing, metrics

# Orianna — gRPC server examples
go run orianna/docs/examples/grpc/server/quickstart/main.go     # minimal gRPC server
go run orianna/docs/examples/grpc/server/complete/main.go        # full-featured with mTLS, interceptors

# Orianna — client examples
go run orianna/docs/examples/http/client/main.go                 # HTTP client with retry + circuit breaker
go run orianna/docs/examples/grpc/client/main.go                 # gRPC client with observability

# Other packages
go run logger/docs/example/main.go
go run jcodec/docs/example/main.go
go run metrics/docs/example/main.go
go run goroutine/docs/example/main.go
go run redis/docs/example/main.go
```

---

## Packages

### orianna — HTTP + gRPC Framework

Production-ready framework with hexagonal architecture. Supports both REST and gRPC with shared observability, resilience, and security infrastructure.

**HTTP server:**
```go
import (
    "github.com/anthanhphan/gosdk/orianna"
    "github.com/anthanhphan/gosdk/orianna/http/configuration"
    "github.com/anthanhphan/gosdk/orianna/http/core"
)

srv, _ := orianna.NewServer(&configuration.Config{
    ServiceName: "user-api",
    Port:        8080,
})

srv.GET("/users", listUsersHandler)
srv.POST("/users", createUserHandler)
srv.Protected().GET("/profile", profileHandler)
srv.Run()
```

**gRPC server:**
```go
import (
    "github.com/anthanhphan/gosdk/orianna/grpc/server"
    "github.com/anthanhphan/gosdk/orianna/grpc/configuration"
)

srv, _ := server.NewServer(&configuration.Config{
    ServiceName: "user-grpc",
    Port:        50051,
},
    server.WithTracing(tracingClient),
    server.WithMetrics(metricsClient),
    server.WithTokenAuth(jwtValidator),
)

srv.RegisterServices(
    server.NewService(&pb.UserService_ServiceDesc, &userService{}).
        UnaryInterceptor(rateLimiter).
        Build(),
)
srv.Start()
```

**What's included:**

| Feature | HTTP | gRPC |
|---------|------|------|
| Tracing (OpenTelemetry) | ✅ middleware | ✅ interceptor |
| Metrics (Prometheus) | ✅ middleware | ✅ interceptor |
| Structured logging | ✅ middleware | ✅ interceptor |
| Rate limiting | ✅ middleware | ✅ interceptor |
| Panic recovery | ✅ middleware | ✅ interceptor |
| Slow request detection | ✅ middleware | ✅ interceptor |
| Circuit breaker + retry | ✅ HTTP client | ✅ gRPC client |
| mTLS / TLS with cert hot-reload | — | ✅ server |
| Token auth (JWT/OAuth2) | — | ✅ interceptor |
| Certificate-based auth | — | ✅ interceptor |
| Health checks | ✅ HTTP endpoint | ✅ gRPC health protocol |
| PII header redaction | ✅ | ✅ |
| UUIDv7 request IDs | ✅ | ✅ |

Architecture:
```
orianna/
├── http/                    # REST/HTTP framework (Fiber engine)
│   ├── server/              #   Server with lifecycle management
│   ├── client/              #   HTTP client with retry + circuit breaker
│   ├── middleware/           #   Tracing, metrics, logging, request ID, redaction
│   ├── routing/              #   Type-safe route builder with group support
│   ├── core/                 #   ISP-compliant context (11 sub-interfaces), response pooling
│   └── configuration/        #   Typed config with validation + secure defaults
├── grpc/                    # gRPC framework
│   ├── server/              #   Server with per-service interceptor chains
│   ├── client/              #   Client with retry, backoff, circuit breaker
│   ├── interceptor/          #   Auth, token auth, tracing, metrics, logging, recovery
│   └── configuration/        #   Typed config with keepalive, TLS, flow control
└── shared/                  # Shared infrastructure (used by both HTTP and gRPC)
    ├── resilience/           #   Circuit breaker (state machine, injectable clock), retry
    ├── health/               #   Health check manager with checker interface
    ├── observability/        #   Zero-alloc metric helpers (code string cache, attempt cache)
    ├── requestid/            #   UUIDv7 request ID generation
    ├── hooks/                #   Lifecycle hooks (OnStart, OnShutdown)
    ├── errors/               #   Shared sentinel errors
    └── ctxkeys/              #   Typed context keys
```

[Full docs](./orianna/docs/README.md) • [REST guide](./orianna/docs/REST.md) • [gRPC guide](./orianna/docs/GRPC.md)

---

### jcodec — Fast JSON

Architecture-aware JSON codec that auto-selects the fastest engine for your CPU: **Sonic** on AMD64/x86_64, **goccy/go-json** on ARM64/Apple Silicon. Drop-in replacement for `encoding/json`.

```go
import "github.com/anthanhphan/gosdk/jcodec"

// Same API as encoding/json — just swap the import
data, _ := jcodec.Marshal(user)                     // → []byte
jcodec.Unmarshal(data, &user)                       // → struct
pretty, _ := jcodec.MarshalIndent(user, "", "  ")   // → formatted JSON
compact, _ := jcodec.CompactString(user)            // → minified string

// Streaming
encoder := jcodec.NewEncoder(writer)
decoder := jcodec.NewDecoder(reader)

// Validation
ok := jcodec.Valid(data)                            // → bool
```

[Full docs](./jcodec/docs/README.md) • [Examples](./jcodec/docs/example/main.go)

---

### logger — Structured Logging

Built from scratch, zero dependencies, **2.5M logs/sec** at 2 allocs/op. Three logging styles (simple, formatted, structured), async mode, and field-based scoping.

```go
import "github.com/anthanhphan/gosdk/logger"

// Production logger with JSON output
undo := logger.InitProductionLogger()
defer undo()

// Three styles
logger.Info("Server started")                                     // simple
logger.Infof("Listening on port %d", 8080)                       // formatted
logger.Infow("Request processed", "method", "GET", "status", 200) // structured (key-value)

// Scoped logger with pre-set fields
log := logger.NewLoggerWithFields(
    logger.String("service", "user-api"),
    logger.String("version", "1.2.0"),
)
log.Infow("User created", "user_id", 42)
// → {"service":"user-api","version":"1.2.0","msg":"User created","user_id":42}

// Async logger for high-throughput paths
async := logger.NewAsyncLogger(log, 10000)  // 10k buffer
defer async.Close()
async.Infow("Processing event", "event_id", "abc-123")
```

[Full docs](./logger/docs/README.md) • [Examples](./logger/docs/example/main.go)

---

### metrics — Application Metrics

Prometheus-backed metrics with a simplified API. Isolated registries prevent metric collisions. Built-in NoopClient for testing.

```go
import "github.com/anthanhphan/gosdk/metrics"

client := metrics.NewClient("myapp",
    metrics.WithBuckets([]float64{0.01, 0.05, 0.1, 0.5, 1.0}),
    metrics.WithConstLabels(map[string]string{"env": "prod"}),
)

// Counters
client.Inc(ctx, "requests_total", "method", "GET", "status", "200")
client.Add(ctx, "bytes_processed", 1024, "direction", "inbound")

// Gauges
client.SetGauge(ctx, "temperature", 72.5, "sensor", "cpu")
client.GaugeInc(ctx, "in_flight_requests")
client.GaugeDec(ctx, "in_flight_requests")

// Histograms & durations
client.Histogram(ctx, "response_size", 1536.0, "endpoint", "/users")
client.Duration(ctx, "request_duration", startTime, "method", "GET")

// Expose /metrics endpoint for Prometheus scraping
http.Handle("/metrics", client.Handler())
```

[Full docs](./metrics/docs/README.md) • [Examples](./metrics/docs/example/main.go)

---

### tracing — Distributed Tracing

OpenTelemetry interface for distributed tracing. Exports traces via OTLP gRPC to Tempo, Jaeger, Grafana Alloy, or any OTLP-compatible collector. Built-in W3C TraceContext propagation.

```go
import "github.com/anthanhphan/gosdk/tracing"

// Initialize with Tempo exporter
client, _ := tracing.NewClient("order-service",
    tracing.WithEndpoint("localhost:4317"),
    tracing.WithEnvironment("production"),
    tracing.WithServiceVersion("1.2.0"),
    tracing.WithInsecure(),  // for local dev
)
defer client.Shutdown(ctx)

// Create spans
ctx, span := client.StartSpan(ctx, "process-order",
    tracing.WithSpanKind(tracing.SpanKindServer),
    tracing.WithAttributes(attribute.String("order_id", "ORD-123")),
)
defer span.End()

// Record errors
span.RecordError(err)
span.SetStatus(codes.Error, "payment failed")

// Context propagation across services (HTTP/gRPC)
tracing.InjectContext(ctx, carrier)
tracing.ExtractContext(ctx, carrier)

// NoopClient for testing (zero cost)
testClient := tracing.NewNoopClient()
```

[Full docs](./tracing/docs/README.md)

---

### validator — Struct Validation

High-performance struct validation with zero allocation after first call. Rules are defined via struct tags, parsed once, cached forever.

```go
import "github.com/anthanhphan/gosdk/validator"

type CreateUserRequest struct {
    Name     string   `validate:"required,min=2,max=50"`
    Email    string   `validate:"required,email"`
    Age      int      `validate:"required,gte=18,lte=120"`
    Role     string   `validate:"required,oneof=admin user guest"`
    Tags     []string `validate:"dive,min=1,max=20"`  // validate each element
    Address  Address  `validate:"required"`            // nested struct validation
}

v := validator.New(
    validator.WithFieldNameTag("json"),     // use json tag names in errors
    validator.WithStopOnFirstError(false),  // collect all errors
)

err := v.ValidateStruct(&request)
// Returns structured ValidationErrors with field name, rule, and message
```

**Supported rules:** `required`, `min`, `max`, `gte`, `lte`, `gt`, `lt`, `email`, `url`, `oneof`, `len`, `numeric`, `alpha`, `alphanum`, `contains`, `startswith`, `endswith`, `dive`, and more.

[Full docs](./validator/docs/README.md)

---

### goroutine — Safe Concurrency

Production-ready concurrency patterns. Every pattern includes automatic panic recovery — panics never crash your app.

```go
import routine "github.com/anthanhphan/gosdk/goroutine"

// Fire-and-forget (panic-safe)
routine.Run(func() {
    panic("caught and logged, app continues running")
})

// Group — run N tasks, wait for all, first error cancels
g := routine.NewGroupWithContext(ctx)
g.Go(func(ctx context.Context) error { return fetchUser(ctx) })
g.Go(func(ctx context.Context) error { return fetchOrders(ctx) })
err := g.Wait()  // returns first error

// FanOut — parallel map with ordered results (generic)
results, err := routine.FanOut(ctx, userIDs, 10, func(ctx context.Context, id string) (*User, error) {
    return userService.Get(ctx, id)
})

// ForEach — parallel side-effects (generic)
err := routine.ForEach(ctx, emails, 5, func(ctx context.Context, email string) error {
    return sendEmail(ctx, email)
})
```

| Pattern | Use case | Key feature |
|---------|----------|-------------|
| `Run` | Fire-and-forget | Auto panic recovery |
| `RunWithContext` | Context-aware goroutine | Prevents goroutine leaks |
| `RunWithTimeout` | Goroutine with deadline | Auto-cancel + leak detection |
| `Group` | Run N tasks, wait for all | First-error + context cancel |
| `WorkerPool` | Fixed workers + job queue | Graceful shutdown, job timeout |
| `FanOut[T, R]` | Parallel map | Generic, ordered results |
| `ForEach[T]` | Parallel side-effects | Generic |

[Full docs](./goroutine/docs/README.md) • [Examples](./goroutine/docs/example/main.go)

---

### conflux — Easy Config

Type-safe configuration loader with generics. Supports JSON and YAML, detects format from file extension.

```go
import "github.com/anthanhphan/gosdk/conflux"

type Config struct {
    Server struct {
        Port int    `yaml:"port" json:"port"`
        Name string `yaml:"name" json:"name"`
    } `yaml:"server" json:"server"`
    Database struct {
        Host string `yaml:"host" json:"host"`
        Port int    `yaml:"port" json:"port"`
    } `yaml:"database" json:"database"`
}

// Load with error handling
config, err := conflux.Load[Config]("config.yaml")

// Or panic on error (ideal for program init)
config := conflux.MustLoad[Config]("config.yaml")
```

[Full docs](./conflux/docs/README.md) • [Examples](./conflux/docs/example/main.go)

---

### redis — Observable Redis Client

Redis client wrapper built on [go-redis v9](https://github.com/redis/go-redis). Supports **Standalone** and **Sentinel** modes. Every command is automatically instrumented with Prometheus metrics and structured error logging. The `ScopedClient` API removes boilerplate action labeling from service methods.

```go
import "github.com/anthanhphan/gosdk/redis"

// Connect (validates config, applies defaults, pings)
client, err := redis.NewClient(&redis.Config{
    Addr:            "localhost:6379",
    MetricNamespace: "myapp",
}, log)
defer client.Close()

// ScopedClient — define scope once, label every command automatically
scoped := client.Scope("user_svc")

val, err := client.Get(scoped.Ctx(ctx, "get_session"), "session:u-1").Result()
// Prometheus: action="user_svc.get_session", command="get"

client.Set(scoped.Ctx(ctx, "set_session"), "session:u-1", token, time.Hour)
// Prometheus: action="user_svc.set_session", command="set"

// Load from config file via conflux
cfg, _ := conflux.Load[redis.Config]("config/redis.yaml")
client, err = redis.NewClient(cfg, log)
```

**Metrics registered automatically:**

| Metric | Type | Labels |
|---|---|---|
| `<namespace>_command_duration_seconds` | Histogram | `action`, `command` |
| `<namespace>_command_errors_total` | Counter | `action`, `command`, `error_type` |

[Full docs](./redis/docs/README.md) • [Examples](./redis/docs/example/main.go)

---

### utils — Common Utilities

Secure helpers for file I/O, environment detection, and panic location tracking.

```go
import "github.com/anthanhphan/gosdk/utils"

// Environment detection
env := utils.GetEnvironment()              // "development" | "staging" | "production"
err := utils.ValidateEnvironment("staging") // validates against known environments

// Secure file I/O (validates paths, prevents traversal)
data, err := utils.ReadFileSecurely("/etc/app/config.yaml")
file, err := utils.OpenFileSecurely(path, os.O_RDONLY, 0600)

// Panic location (used internally by goroutine package)
location, err := utils.GetPanicLocation()  // → "mypackage/handler.go:42"
shortPath := utils.GetShortPath(fullPath)  // → relative to module root
```

---

## Explore the Code

**For developers:**

```bash
# Run the full 14-step pipeline (recommended before committing)
make all

# Individual targets
make fmt              # go fmt + goimports
make lint             # golangci-lint
make vet              # go vet + staticcheck + ineffassign
make test_race        # tests with race detector
make test_coverage    # coverage report → code_report/coverage.html
make security         # gosec + govulncheck
make help             # show all available targets
```

**Pipeline steps:** `tidy` → `fmt` → `imports` → `lint` → `vet` → `staticcheck` → `ineffassign` → `misspell` → `cyclo` → `race` → `coverage` → `gosec` → `govulncheck`

**Project structure:**
```
gosdk/
├── orianna/                 # HTTP + gRPC framework
│   ├── docs/                #   README.md, REST.md, GRPC.md + examples/
│   ├── http/                #   server, client, middleware, routing, core, configuration
│   ├── grpc/                #   server, client, interceptor, configuration, core
│   └── shared/              #   resilience, health, observability, hooks, errors, ctxkeys
├── jcodec/                  # Architecture-aware JSON codec (Sonic / goccy)
├── logger/                  # Structured logger (sync + async)
├── metrics/                 # Prometheus metrics client
├── tracing/                 # OpenTelemetry tracing client
├── validator/               # Struct validation engine
├── goroutine/               # Concurrency patterns (Run, Group, FanOut, ForEach)
├── conflux/                 # Configuration loader (JSON/YAML)
├── redis/                   # Observable Redis client (Standalone + Sentinel)
└── utils/                   # File I/O, environment, panic tracking
```

> Each package folder has its own `docs/README.md` with detailed information.

---

## Need Help?

- **Read the docs**: Each package has detailed guides in `docs/README.md`
- **Check examples**: Look in `docs/example/` or `docs/examples/` folders
- **Found a bug?**: [Open an issue](https://github.com/anthanhphan/gosdk/issues)
- **Have an idea?**: [Start a discussion](https://github.com/anthanhphan/gosdk/discussions)

---

## License

MIT License — see [LICENSE](./LICENSE)

---

<div align="center">

**Made by [anthanhphan](https://github.com/anthanhphan)**

⭐ Star this repo if you find it helpful!

</div>
