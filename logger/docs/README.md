# Logger Package

A high-performance, zero-dependency structured logging package for Go. Built from scratch to match [zap](https://github.com/uber-go/zap)-level performance — **2.5M logs/sec**, **2 allocs/op**, **248 B/op**.

## Installation

```bash
go get github.com/anthanhphan/gosdk/logger
```

## Quick Start

```go
package main

import "github.com/anthanhphan/gosdk/logger"

func main() {
    undo := logger.InitProductionLogger()
    defer undo()

    logger.Info("Application started")
    logger.Infof("Listening on port %d", 8080)
    logger.Infow("User created", "user_id", 12345, "email", "user@example.com")
}
```

## Logging Styles

### Simple

```go
logger.Info("User logged in")
logger.Error("Connection failed")
```

### Printf-style

```go
logger.Infof("User %s logged in from %s", "john", "192.168.1.1")
logger.Errorf("Failed after %d retries: %v", 3, err)
```

### Structured (Recommended)

```go
logger.Infow("Request completed",
    "method",  "GET",
    "path",    "/api/users",
    "status",  200,
    "latency", 3.14,
)
```

### Typed Fields (Maximum Performance)

```go
log := logger.NewLoggerWithFields(
    logger.String("service", "user-api"),
)
log.Info("request",
    logger.String("method", "GET"),
    logger.Int("status", 200),
    logger.Float64("latency_ms", 3.14),
    logger.Bool("cached", true),
    logger.ErrorField(nil),
)
```

## Configuration

```go
undo := logger.InitLogger(&logger.Config{
    LogLevel:          logger.LevelInfo,       // debug | info | warn | error
    LogEncoding:       logger.EncodingJSON,     // json | console
    OutputPaths:       []string{"log/app.log"}, // stdout, stderr, or file paths
    DisableCaller:     false,
    DisableStacktrace: true,
    IsDevelopment:     false,
    Timezone:          "Asia/Ho_Chi_Minh",
    MaskKey:           "0123456789abcdef",       // AES key for log:"mask" fields
},
    logger.String("app", "my-service"),          // default fields on every log
    logger.String("env", "production"),
)
defer undo()
```

### Preset Configs

```go
undo := logger.InitDevelopmentLogger()  // Debug, Console, color, caller, stacktrace
undo := logger.InitProductionLogger()   // Info, JSON, caller, stacktrace
```

### Async Logger

Non-blocking — log entries are queued and written in a background goroutine:

```go
undo := logger.InitAsyncLogger(&logger.Config{
    LogLevel:    logger.LevelInfo,
    LogEncoding: logger.EncodingJSON,
})
defer undo() // flushes remaining entries on shutdown
```

## Component Loggers

Create loggers with persistent fields for different components:

```go
authLog := logger.NewLoggerWithFields(logger.String("component", "auth"))
authLog.Infow("Login success", "user_id", "u-123")

dbLog := logger.NewLoggerWithFields(logger.String("component", "database"))
dbLog.Infow("Connection established", "pool_size", 10)
```

## Log Levels

| Level | Constant | Exits? |
|---|---|---|
| Debug | `LevelDebug` | No |
| Info | `LevelInfo` | No |
| Warn | `LevelWarn` | No |
| Error | `LevelError` | No |
| Fatal | — | **Yes** (`os.Exit(1)`) |

## Output Formats

### JSON (Ordered Keys)

Keys always ordered: `ts` → `caller` → `level` → `msg` → `trace_id` → `request_id` → rest.

```json
{
  "ts": "2025-11-17T13:57:39.123456+07:00",
  "caller": "handler/user.go:42",
  "level": "info",
  "msg": "User created",
  "trace_id": "abc123",
  "request_id": "req-001",
  "user_id": 12345
}
```

### Console (Colorized in Dev Mode)

```
2025-11-17T13:57:39+07:00  INFO  handler/user.go:42  User created  user_id=12345
```

Colors: Debug=Cyan, Info=Green, Warn=Yellow, Error=Red.

## Sensitive Data Handling

```go
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password" log:"omit"`  // never logged
    Token    string `json:"token"    log:"mask"`  // masked or AES-encrypted
}
```

| Tag | No MaskKey | With MaskKey |
|---|---|---|
| `log:"omit"` | Excluded | Excluded |
| `log:"mask"` | `"***"` | AES-GCM encrypted (base64) |
| _(none)_ | Normal | Normal |

Nested structs processed recursively. MaskKey: 16/24/32 bytes (AES-128/192/256).

## Flushing & Shutdown

```go
logger.Flush()      // flush all buffered output (async + buffer)
myLogger.Sync()     // flush a specific logger instance
```

> Always call `Flush()` or the undo function before program exit.

## Security

- **Directory traversal protection** — file paths validated before creation
- **Secure file permissions** — `0600` (owner read/write only)
- **`log:"omit"` fields** — never reach output at any encoding stage
- **`log:"mask"` fields** — AES-GCM encrypted with configurable key

## Migration from Zap

| Zap | This Logger |
|---|---|
| `zap.String("k", "v")` | `logger.String("k", "v")` |
| `zap.Int("k", 42)` | `logger.Int("k", 42)` |
| `sugar.Infow(...)` | `logger.Infow(...)` |
| `logger.Sync()` | `logger.Flush()` / `myLogger.Sync()` |
| `zap.NewProduction()` | `logger.InitProductionLogger()` |
| `zap.NewDevelopment()` | `logger.InitDevelopmentLogger()` |

---

## Benchmark

Measured on Apple Silicon, Go 1.25:

```
BenchmarkLogger_Infow_Baseline-8     1,550,000    798 ns/op    248 B/op    2 allocs/op
BenchmarkLogger_Infow_ManyFields-8     840,000   1451 ns/op    888 B/op    9 allocs/op
BenchmarkJSONEncoder_Encode-8        4,500,000    262 ns/op    288 B/op    1 allocs/op
BenchmarkLogger_Parallel-8           3,350,000    392 ns/op    248 B/op    2 allocs/op
```

### vs Zap

| Metric | zap.Sugar | This Logger |
|---|---|---|
| Latency | ~800-900 ns | **798 ns** |
| Memory | ~200-300 B | **248 B** |
| Allocs | 1 | 2 |

> The +1 alloc is Go's `...any` variadic — a language-level cost.

## Performance Techniques

### Typed Field Union

Fields use a discriminated union. Common types (string, int, bool, float64) are stored in dedicated slots — **zero `interface{}` boxing**:

```go
type Field struct {
    Key     string
    Type    FieldType  // String | Int64 | Bool | Float64 | Any
    Integer int64      // int64, bool (0/1), float64 (math.Float64bits)
    Str     string
    Iface   any        // fallback for structs, maps, slices
}
```

### Object Pooling (`sync.Pool`)

Three pool layers for near-zero GC pressure:

| Pool | Object | Pre-alloc |
|---|---|---|
| `entryPool` | `*Entry` | 16 fields |
| `fieldSlicePool` | `*[]Field` | 8 fields |
| `bufPool` | `*[]byte` | 1KB |

### Buffered Async I/O

```
Log Call → bufio.Writer (256KB) → Goroutine (100ms flush) → os.Stdout/File
              memcpy only            no syscall on hot path
```

I/O decoupled from request path. Individual `Write()` = in-memory buffer copy.

### Caching

| Cache | Type | Purpose |
|---|---|---|
| `callerCache` | `sync.Map` | `runtime.Caller` file → short path (skip `filepath.Rel`) |
| `structMetaCache` | `sync.Map` | Struct tag metadata per `reflect.Type` (skip re-parsing) |

### Manual JSON Encoding

No `encoding/json`. Direct `append()` on pooled `[]byte`:
- `appendJSONString` — RFC 8259 escaping, zero-alloc
- `appendTypedFieldValue` — type-switch on `FieldType`
- `time.AppendFormat` — avoids `time.Format()` string alloc
- Single-pass priority field ordering (no sort, no nested loops)

### Architecture

```
Application Code
  └─ parseKeysAndValues (pooled []Field)
      └─ createEntry (pooled *Entry, processField, setCallerInfo)
          └─ Encoder (pooled []byte, typed append, single-pass priority)
              └─ BufferedWriteSyncer (256KB bufio, 100ms flush goroutine)
                  └─ os.Stdout / File
```

## API Reference

### Field Constructors

| Function | Type | Alloc |
|---|---|---|
| `String(k, v)` | `string` | **0** |
| `Int(k, v)` | `int` | **0** |
| `Int64(k, v)` | `int64` | **0** |
| `Float64(k, v)` | `float64` | **0** |
| `Bool(k, v)` | `bool` | **0** |
| `ErrorField(err)` | `error` | **0** |
| `Any(k, v)` | `any` | may alloc |
