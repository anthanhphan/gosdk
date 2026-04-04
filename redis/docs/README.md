# Redis Package

An observable Redis client for Go, built on [go-redis v9](https://github.com/redis/go-redis).
Supports **Standalone** and **Sentinel** modes with automatic Prometheus metrics and structured error logging on every command.

## Installation

```bash
go get github.com/anthanhphan/gosdk/redis
```

## Quick Start

```go
log := logger.NewLogger(&logger.ProductionConfig, nil)

client, err := redis.NewClient(&redis.Config{
    Addr:            "localhost:6379",
    MetricNamespace: "myapp",
}, log)
if err != nil {
    panic(err)
}
defer client.Close()

ctx := redis.WithAction(context.Background(), "boot")
client.Set(ctx, "key", "value", time.Minute)
```

## Configuration

Load via [conflux](https://github.com/anthanhphan/gosdk/conflux) from a YAML or JSON file:

```go
cfg, err := conflux.Load[redis.Config]("config/redis.yaml")
if err != nil {
    panic(err)
}
client, err := redis.NewClient(cfg, log)
```

### Standalone

```yaml
addr:             "localhost:6379"
password:         "s3cr3t"
db:               0
pool_size:        10
min_idle_conns:   2
dial_timeout:     "5s"
read_timeout:     "3s"
write_timeout:    "3s"
metric_namespace: "myapp"
metric_subsystem: "cache"
```

### Sentinel

```yaml
master_name:       "mymaster"
sentinel_addrs:
  - "sentinel1:26379"
  - "sentinel2:26379"
sentinel_password: "sentinel-pass"
password:          "s3cr3t"
db:                0
pool_size:         20
min_idle_conns:    5
dial_timeout:      "5s"
read_timeout:      "3s"
write_timeout:     "3s"
metric_namespace:  "myapp"
metric_subsystem:  "cache"
```

### Fields

| Field | Default | Description |
|---|---|---|
| `addr` | — | Standalone server (`host:port`). Required if not using Sentinel. |
| `master_name` | — | Sentinel master name. Required for Sentinel mode. |
| `sentinel_addrs` | — | Sentinel node addresses. Required for Sentinel mode. |
| `sentinel_password` | — | AUTH password for Sentinel nodes. |
| `password` | — | Redis AUTH password. |
| `db` | `0` | Logical database index (0–15). |
| `pool_size` | `10` | Max connections in the pool. |
| `min_idle_conns` | `2` | Minimum idle connections kept open. |
| `dial_timeout` | `5s` | Timeout for establishing a connection. |
| `read_timeout` | `3s` | Timeout for socket reads. |
| `write_timeout` | `3s` | Timeout for socket writes. |
| `metric_namespace` | `redis` | Prometheus namespace prefix. |
| `metric_subsystem` | — | Prometheus subsystem label (optional). |

## Metrics

Every command is automatically instrumented. Two metrics are registered on the default Prometheus registry:

| Metric | Type | Labels |
|---|---|---|
| `<namespace>_command_duration_seconds` | Histogram | `action`, `command` |
| `<namespace>_command_errors_total` | Counter | `action`, `command`, `error_type` |

`error_type` labels: `context` · `connection` · `other`

> `redis.Nil` (cache miss) is never counted as an error.

## Action Label

The `action` label identifies the business operation that triggered a Redis command.
Set it on the context before calling any command:

```go
ctx := redis.WithAction(ctx, "get_user_session")
val, err := client.Get(ctx, "session:u-123").Result()
```

If not set, the label defaults to `"unknown"`.

## ScopedClient

`ScopedClient` removes the need to call `WithAction` in every method.
Define the scope once at the service level; pass only the short operation name per call.

```go
type UserService struct {
    redis  *redis.Client
    scoped *redis.ScopedClient
    log    *logger.Logger
}

func NewUserService(c *redis.Client, log *logger.Logger) *UserService {
    return &UserService{
        redis:  c,
        scoped: c.Scope("user_svc"), // registered once
        log:    log,
    }
}

// action = "user_svc.get_session"
func (s *UserService) GetSession(ctx context.Context, uid string) (string, error) {
    val, err := s.redis.Get(s.scoped.Ctx(ctx, "get_session"), "session:"+uid).Result()
    if errors.Is(err, goredis.Nil) {
        return "", nil // cache miss
    }
    return val, err
}

// action = "user_svc.set_session"
func (s *UserService) SetSession(ctx context.Context, uid, token string, ttl time.Duration) error {
    return s.redis.Set(s.scoped.Ctx(ctx, "set_session"), "session:"+uid, token, ttl).Err()
}
```

Scopes can be nested:

```go
ctx := client.Scope("myapp").Scope("user_cache").Ctx(ctx, "get")
// action = "myapp.user_cache.get"
```

## Error Logging

Command errors are logged automatically by the built-in hook:

```json
{
  "level": "error",
  "msg": "redis: command error",
  "action": "user_svc.get_session",
  "command": "get",
  "error_type": "connection",
  "error": "dial tcp: connection refused"
}
```

## Common Patterns

### Get with cache-miss handling

```go
val, err := client.Get(ctx, "user:1").Result()
switch {
case errors.Is(err, goredis.Nil):
    // cache miss — fetch from DB, populate cache
case err != nil:
    return err
}
```

### Pipeline

```go
pipe := client.Pipeline()
pipe.Set(ctx, "k1", "v1", time.Minute)
pipe.Set(ctx, "k2", "v2", time.Minute)
_, err := pipe.Exec(ctx)
```

### Health check

```go
if err := client.Ping(ctx).Err(); err != nil {
    log.Errorw("redis unhealthy", "error", err)
}
```
