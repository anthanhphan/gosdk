# Orianna — gRPC

Complete guide to building gRPC services with Orianna. All import paths, types, defaults, and API signatures are sourced directly from the codebase.

## Table of Contents

- [Configuration](#configuration)
  - [Server Config](#server-config)
  - [Config Defaults](#config-defaults)
  - [Server Options](#server-options)
- [Server Lifecycle](#server-lifecycle)
- [TLS / mTLS](#tls--mtls)
  - [Server-side TLS](#server-side-tls)
  - [Mutual TLS (mTLS)](#mutual-tls-mtls)
  - [Certificate-based Authorization](#certificate-based-authorization)
- [Service Registration](#service-registration)
  - [Basic Registration](#basic-registration)
  - [Service Builder (Per-service Interceptors)](#service-builder)
  - [Unary RPC Implementation](#unary-rpc-implementation)
- [Interceptors](#interceptors)
  - [Built-in Interceptors](#built-in-interceptors)
  - [Interceptor Chain Order](#interceptor-chain-order)
  - [Custom Interceptors](#custom-interceptors)
  - [Interceptor Utilities](#interceptor-utilities)
- [Lifecycle Hooks](#lifecycle-hooks)
- [Health Checks](#health-checks)
- [Streaming](#streaming)
  - [Server Streaming](#server-streaming)
  - [Client Streaming](#client-streaming)
  - [Bidirectional Streaming](#bidirectional-streaming)
- [gRPC Client](#grpc-client)
  - [Client Setup](#client-setup)
  - [Client Options](#client-options)
  - [Using the Connection](#using-the-connection)
- [Context Interface](#context-interface)

---

## Configuration

### Server Config

```go
import "github.com/anthanhphan/gosdk/orianna/grpc/configuration"

connectionTimeout := 120 * time.Second
keepaliveTime := 2 * time.Hour
keepaliveTimeout := 20 * time.Second
shutdownTimeout := 30 * time.Second

config := &configuration.Config{
    // ── Required ──
    ServiceName: "user-service",
    Port:        50051,

    // ── Optional ──
    Version: "v1.0.0", // displayed in startup logs

    // ── Timeouts (pointer-based: nil = use default) ──
    ConnectionTimeout:       &connectionTimeout,     // default: 120s
    KeepaliveTime:           &keepaliveTime,          // default: 2h
    KeepaliveTimeout:        &keepaliveTimeout,       // default: 20s
    GracefulShutdownTimeout: &shutdownTimeout,        // default: 30s
    KeepaliveMinTime:        nil,                     // default: 10s (flood protection)

    // ── Limits ──
    MaxRecvMsgSize:      4 * 1024 * 1024,   // default: 4MB
    MaxSendMsgSize:      4 * 1024 * 1024,   // default: 4MB
    MaxConcurrentStreams: 256,                // default: 256

    // ── Buffers (0 = gRPC default: 32KB) ──
    ReadBufferSize:        0,
    WriteBufferSize:       0,
    InitialWindowSize:     0,   // per-stream flow control (gRPC default: 64KB)
    InitialConnWindowSize: 0,   // per-connection flow control (gRPC default: 64KB)

    // ── Features ──
    VerboseLogging:            true,
    VerboseLoggingSkipMethods: []string{"/grpc.health.v1.Health/*"},
    LogPayloads:               false,               // WARNING: may contain PII
    MaxPayloadLogSize:          1024,                // truncate payloads to N bytes
    EnableReflection:           true,                // grpcurl/grpcui support
    EnableCompression:          false,               // gzip payload compression
    SlowRequestThreshold:       2 * time.Second,     // auto-registers SlowRPCDetector

    // ── Keepalive Enforcement ──
    KeepalivePermitWithoutStream: nil,  // default: false (deny pings without streams)

    // ── TLS ──
    TLS:  nil,  // server-side TLS
    MTLS: nil,  // mutual TLS
    Permissions: nil, // certificate-based method authorization
}
```

### Config Defaults

| Field | Default | Notes |
|-------|---------|-------|
| `Port` | `50051` | |
| `ConnectionTimeout` | `120s` | Nil = default |
| `KeepaliveTime` | `2h` | Nil = default |
| `KeepaliveTimeout` | `20s` | Nil = default |
| `GracefulShutdownTimeout` | `30s` | Nil = default |
| `KeepaliveMinTime` | `10s` | Flood protection |
| `KeepalivePermitWithoutStream` | `false` | `*bool`, nil = false |
| `MaxRecvMsgSize` | `4MB` | 0 = default |
| `MaxSendMsgSize` | `4MB` | 0 = default |
| `MaxConcurrentStreams` | `256` | 0 = default |
| `MaxPayloadLogSize` | `1024` bytes | When `LogPayloads` is true |

### Server Options

```go
import "github.com/anthanhphan/gosdk/orianna/grpc/server"

srv, err := server.NewServer(config,
    server.WithMetrics(metricsClient),
    server.WithTracing(tracingClient),
    server.WithHooks(hooks),
    server.WithHealthChecker(dbChecker),
    server.WithHealthChecker(cacheChecker),
    server.WithHealthManager(healthMgr),
    server.WithTokenAuth(jwtValidator),
    server.WithGlobalUnaryInterceptor(customUnary),
    server.WithGlobalStreamInterceptor(customStream),
    server.WithPanicRecover(customUnaryRecover, customStreamRecover),
    server.WithRateLimiter(rateLimiterInterceptor),
    server.WithDisableRecovery(), // WARNING: testing only
)
```

| Option | Description |
|--------|-------------|
| `WithMetrics(client)` | Enable Prometheus metrics (request count, latency, in-flight) |
| `WithTracing(client)` | Enable OpenTelemetry tracing |
| `WithHooks(hooks)` | Set lifecycle hooks |
| `WithHealthChecker(checker)` | Add health checker (auto-creates manager if nil) |
| `WithHealthManager(mgr)` | Set custom health manager |
| `WithTokenAuth(validator)` | Token-based auth (JWT, OAuth2) via `TokenValidator` interface |
| `WithGlobalUnaryInterceptor(i...)` | Add global unary interceptors (appended after built-ins) |
| `WithGlobalStreamInterceptor(i...)` | Add global stream interceptors |
| `WithPanicRecover(unary, stream)` | Custom panic recovery interceptors |
| `WithRateLimiter(interceptor)` | Custom rate limiter interceptor |
| `WithDisableRecovery()` | Disable built-in panic recovery (testing only) |

---

## Server Lifecycle

```go
// Create
srv, err := server.NewServer(config, opts...)

// Register services
pb.RegisterUserServiceServer(srv.GRPCServer(), &userService{})

// Start (blocks until OS signal SIGTERM/SIGINT)
err = srv.Start()

// Programmatic shutdown (applies GracefulShutdownTimeout)
err = srv.Shutdown(ctx)
```

**Startup sequence:**
1. Config validated → `conf.Validate()` (fail-fast)
2. Defaults merged → `MergeConfigDefaults(conf)`
3. Server options applied
4. gRPC server options built (TLS, keepalive, interceptor chains, buffers)
5. Reflection registered (if enabled)
6. `OnServerStart` hooks fire
7. Services registered with their interceptor chains
8. TCP listener starts → blocks

**Shutdown sequence:**
1. OS signal received (SIGTERM/SIGINT)
2. `OnShutdown` hooks fire
3. `GracefulStop()` starts (drains in-flight RPCs)
4. If timeout exceeded → `Stop()` forces shutdown

---

## TLS / mTLS

### Server-side TLS

Encrypts traffic. Clients authenticate via tokens. Supports **hot-reload** — certificates are reloaded from disk on each TLS handshake (no server restart needed for rotation).

```go
config.TLS = &configuration.TLSConfig{
    CertFile: "server.crt",  // required
    KeyFile:  "server.key",  // required
}
```

### Mutual TLS (mTLS)

Both server and client authenticate via certificates. Always enforces `RequireAndVerifyClientCert`. Supports **hot-reload**.

```go
config.MTLS = &configuration.MTLSConfig{
    CertFile: "server.crt",  // required
    KeyFile:  "server.key",  // required
    CAFile:   "ca.crt",      // required — CA that signed client certificates
}
```

> **TLS and mTLS are mutually exclusive.** Config validation fails at startup if both are set.
> **Minimum TLS version:** 1.3 (enforced).

### Certificate-based Authorization

Maps client certificate Common Names (CN) to allowed gRPC methods. Requires mTLS.

```go
config.Permissions = []configuration.ClientPermission{
    {
        ClientIdentity: "service-a",
        AllowedMethods: []string{"/package.Service/*"},     // service wildcard
    },
    {
        ClientIdentity: "service-b",
        AllowedMethods: []string{"/package.Service/GetUser"}, // exact method
    },
    {
        ClientIdentity: "admin-tool",
        AllowedMethods: []string{"/*"},                       // global wildcard
    },
}
```

**Method patterns:**
- Exact: `/package.Service/Method`
- Service wildcard: `/package.Service/*`
- Global wildcard: `/*`

---

## Service Registration

### Basic Registration

```go
pb.RegisterUserServiceServer(srv.GRPCServer(), &userService{})
```

### Service Builder

Register services with **per-service interceptors** (applied only to that service, after global interceptors):

```go
svc := server.NewService(&pb.UserService_ServiceDesc, &userService{}).
    UnaryInterceptor(auditInterceptor, validationInterceptor).
    StreamInterceptor(streamAuditInterceptor).
    Build()

srv.RegisterServices(*svc)
```

> **Atomic registration**: If any service in a batch fails validation (nil desc, nil impl, duplicate name), none are registered.

### Unary RPC Implementation

```go
func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    if req.Id <= 0 {
        return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
    }

    user, err := s.db.FindUser(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "user %d not found", req.Id)
    }

    return &pb.GetUserResponse{User: user}, nil
}
```

---

## Interceptors

### Built-in Interceptors

| Interceptor | Trigger | Description |
|-------------|---------|-------------|
| **Recovery** | Always on (unless `WithDisableRecovery`) | Panic recovery with stack trace logging + request_id/trace_id correlation |
| **Request ID** | Always on | UUIDv7 generation, preserves client `x-request-id` if valid |
| **Cert Auth** | `MTLS` + `Permissions` configured | Certificate CN-based method authorization |
| **Token Auth** | `WithTokenAuth(validator)` | Token-based auth (JWT, OAuth2) via `TokenValidator` interface |
| **Rate Limiter** | `WithRateLimiter(interceptor)` | Custom rate limiter |
| **Metrics** | `WithMetrics(client)` | Counter, histogram, in-flight gauge with per-client identity |
| **Slow RPC** | `SlowRequestThreshold > 0` | Warns when RPCs exceed threshold |
| **Tracing** | `WithTracing(client)` | OpenTelemetry spans with gRPC attributes |
| **Logging** | `VerboseLogging: true` | Full request/response audit trail with PII redaction |

### Interceptor Chain Order

Both unary and stream chains follow the same order (outermost → innermost):

```
1. Recovery (panic safety)
2. Request ID (correlation)
3. Cert Auth (mTLS identity)
4. Token Auth (JWT/OAuth2)
5. Rate Limiter
6. Metrics
7. Slow RPC Detector
8. Tracing
9. Verbose Logging
10. Global custom interceptors
11. Per-service interceptors
```

All interceptors use **zero-allocation dispatch** — indexed recursive dispatch instead of N-closure allocations per request.

### Custom Interceptors

```go
import "github.com/anthanhphan/gosdk/orianna/grpc/interceptor"

// Add globally
server.WithGlobalUnaryInterceptor(myInterceptor)
server.WithGlobalStreamInterceptor(myStreamInterceptor)

// Add per-service
server.NewService(desc, impl).UnaryInterceptor(myInterceptor).Build()
```

### Interceptor Utilities

```go
// Chain multiple interceptors into one
combined := interceptor.Chain(auth, logging, metrics)

// Stream variant
combined := interceptor.StreamChain(streamAuth, streamLogging)

// Timeout wrapper
interceptor.Timeout(5 * time.Second)

// Conditional application
interceptor.Optional(conditionFn, myInterceptor)

// Skip for specific methods
interceptor.SkipForMethods(myInterceptor, "/grpc.health.v1.Health/Check")
```

---

## Lifecycle Hooks

```go
import "github.com/anthanhphan/gosdk/orianna/grpc/core"

hooks := core.NewHooks()

hooks.AddOnRequest(func(ctx core.Context) {
    logger.Debugw("rpc", "method", ctx.FullMethod(), "peer", ctx.Peer())
})

hooks.AddOnResponse(func(ctx core.Context, code string, latency time.Duration) {
    logger.Infow("rpc_response",
        "method", ctx.FullMethod(),
        "code", code,
        "ms", latency.Milliseconds(),
    )
})

hooks.AddOnError(func(ctx core.Context, err error) {
    logger.Errorw("rpc_error", "method", ctx.FullMethod(), "error", err)
})

hooks.AddOnShutdown(func() {
    db.Close()
    logger.Infow("shutting down")
})

hooks.AddOnServerStart(func(server any) error {
    logger.Infow("server started")
    return nil // return error to abort startup
})

srv, _ := server.NewServer(config, server.WithHooks(hooks))
```

---

## Health Checks

Same shared `health` package as HTTP — cross-protocol consistent:

```go
import "github.com/anthanhphan/gosdk/orianna/shared/health"

srv, _ := server.NewServer(config,
    server.WithHealthChecker(
        health.NewCustomChecker("database", func(ctx context.Context) health.HealthCheck {
            if err := db.PingContext(ctx); err != nil {
                return health.HealthCheck{
                    Status:  health.StatusUnhealthy,
                    Message: "PostgreSQL unreachable",
                    Error:   err,
                }
            }
            return health.HealthCheck{
                Status:  health.StatusHealthy,
                Message: "PostgreSQL connected",
            }
        }),
    ),
)

// gRPC-specific health checker (checks target endpoint connectivity)
import grpcServer "github.com/anthanhphan/gosdk/orianna/grpc/server"

server.WithHealthChecker(
    grpcServer.NewGRPCChecker("localhost:50052", "downstream-service", 5*time.Second),
)
```

---

## Streaming

### Server Streaming

Server sends multiple messages to client:

```go
func (s *userService) StreamUsers(req *pb.StreamUsersRequest, stream pb.UserService_StreamUsersServer) error {
    users := s.getAllUsers()
    for _, user := range users {
        if err := stream.Send(user); err != nil {
            return status.Errorf(codes.Internal, "send failed: %v", err)
        }
    }
    return nil
}
```

### Client Streaming

Client sends multiple messages, server responds once:

```go
func (s *userService) BatchCreateUsers(stream pb.UserService_BatchCreateUsersServer) error {
    var created []*pb.User
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&pb.BatchCreateUsersResponse{
                Users: created,
                Count: int32(len(created)),
            })
        }
        if err != nil {
            return err
        }
        user := s.createUser(req)
        created = append(created, user)
    }
}
```

### Bidirectional Streaming

Both sides send and receive concurrently:

```go
func (s *userService) UserChat(stream pb.UserService_UserChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }

        response := &pb.ChatMessage{
            UserId:    msg.UserId,
            Message:   fmt.Sprintf("Echo: %s", msg.Message),
            Timestamp: time.Now().Unix(),
        }
        if err := stream.Send(response); err != nil {
            return err
        }
    }
}
```

---

## gRPC Client

Production-grade gRPC client with connection pooling, retry, circuit breaker, and observability.

### Client Setup

```go
import "github.com/anthanhphan/gosdk/orianna/grpc/client"

grpcClient, err := client.NewClient(
    client.WithAddress("localhost:50051"),
    client.WithServiceName("user-service"),
    client.WithMetrics(metricsClient),
    client.WithTracing(tracingClient),
    client.WithLogger(myLogger),
    client.WithRetry(&resilience.RetryConfig{
        MaxAttempts:          3,
        InitialBackoff:       100 * time.Millisecond,
        MaxBackoff:           5 * time.Second,
        Multiplier:           2.0,
        RetryableStatusCodes: []int{14}, // codes.Unavailable
    }),
    client.WithCircuitBreaker(&resilience.CircuitBreakerConfig{
        FailureThreshold: 5,
        SuccessThreshold: 3,
        Timeout:          30 * time.Second,
    }),
    client.WithTLS(&client.TLSConfig{
        CertFile:           "client.crt",  // for mTLS
        KeyFile:            "client.key",
        CAFile:             "ca.crt",      // server CA verification
        ServerNameOverride: "",
    }),
)
if err != nil {
    log.Fatal(err)
}
defer grpcClient.Close()
```

### Client Options

| Option | Description |
|--------|-------------|
| `WithAddress(addr)` | gRPC server address (required) |
| `WithServiceName(name)` | Service name for metrics/tracing |
| `WithMetrics(client)` | Prometheus metrics (request count, latency, in-flight) |
| `WithTracing(client)` | OpenTelemetry distributed tracing |
| `WithLogger(log)` | Custom structured logger |
| `WithRetry(cfg)` | Retry with exponential backoff + jitter |
| `WithCircuitBreaker(cfg)` | Circuit breaker (fail-fast when downstream is down) |
| `WithTLS(cfg)` | TLS/mTLS configuration |

**Built-in features (always on):**
- Keepalive: `Time=30s`, `Timeout=10s`
- Message size limits: `4MB` recv/send
- Unary + stream client interceptors (observability logging)
- gzip compression support (register-only, enable per-call)

### Using the Connection

```go
conn := grpcClient.Connection()

// Use standard protobuf client
userClient := userpb.NewUserServiceClient(conn)

// Unary call
resp, err := userClient.GetUser(ctx, &userpb.GetUserRequest{Id: 1})

// Server streaming
stream, err := userClient.StreamUsers(ctx, &userpb.StreamUsersRequest{BatchSize: 10})
for {
    user, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(user.Name)
}
```

---

## Context Interface

`core.Context` provides a framework-agnostic interface for gRPC request operations:

| Method | Description |
|--------|-------------|
| `Context()` | Underlying `context.Context` |
| `FullMethod()` | Full gRPC method name (e.g., `/package.Service/Method`) |
| `ServiceName()` | Parsed service name |
| `MethodName()` | Parsed method name |
| `Peer()` | Client address |
| `IsSecure()` | Whether connection uses TLS |
| `GetMetadata(key)` | First value of incoming metadata key |
| `GetMetadataValues(key)` | All values of incoming metadata key |
| `IncomingMetadata()` | Full incoming metadata (cached, parsed once) |
| `SetOutgoingHeader(key, values...)` | Set response header metadata |
| `SetOutgoingTrailer(key, values...)` | Set response trailer metadata |
| `GetLocal(key)` / `SetLocal(key, value)` | Request-scoped local storage (thread-safe) |
| `RequestID()` | Request ID from metadata (or `"unknown"`) |
| `TraceID()` | OpenTelemetry trace ID (or empty) |

**gRPC StatusError** (structured error type):

```go
import "github.com/anthanhphan/gosdk/orianna/grpc/core"

err := core.NewStatusError(codes.NotFound, "user not found").
    WithInternalMsg("db query failed for user %d", userID).
    WithCause(dbErr)

// Convert to gRPC error for returning from handlers
return nil, err.ToGRPCError()

// Check error code
if core.IsCode(err, codes.NotFound) { ... }

// Wrap errors
wrapped := core.WrapError(err, "GetUser failed")
```
