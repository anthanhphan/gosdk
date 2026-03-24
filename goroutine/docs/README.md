# Goroutine Package

Production-ready concurrency patterns for Go with panic recovery, context cancellation, timeout support, and goroutine leak prevention.

## Installation

```bash
go get github.com/anthanhphan/gosdk/goroutine
```

## Patterns

| Pattern | Use case | Key feature |
|---|---|---|
| [`Run`](#run) | Fire-and-forget goroutine | Fast paths for common signatures |
| [`RunWithContext`](#runwithcontext) | Context-aware goroutine | Prevents goroutine leaks |
| [`RunWithTimeout`](#runwithtimeout) | Goroutine with deadline | Auto-cancel + leak detection |
| [`Group`](#group) | Run N tasks, wait for all | First-error + context cancel |
| [`WorkerPool`](#workerpool) | Fixed workers + job queue | Graceful shutdown, job timeout |
| [`FanOut`](#fanout) | Parallel map with ordered results | Generic `[T, R]` |
| [`ForEach`](#foreach) | Parallel side-effects | Generic `[T]` |

All patterns include **panic recovery** — panics never crash your application.

---

## Run

Starts a goroutine with automatic panic recovery and location tracking.

For `func()`, `func(string)`, `func(int)`, `func(error)` — reflect is bypassed entirely.

```go
routine.Run(func(msg string) {
    logger.Info("Message:", msg)
}, "Hello, world!")

routine.Run(func() {
    panic("recovered and logged with panic_at location")
})
```

> [!WARNING]
> `Run` provides **no timeout or cancellation**. The goroutine runs until `fn` completes.
> For production use with external calls (API, DB, gRPC), use `RunWithContext` or `RunWithTimeout`.

---

## RunWithContext

Starts a goroutine that receives a context. When the context is cancelled, the function should observe `ctx.Done()` and return — preventing goroutine leaks.

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

routine.RunWithContext(ctx, func(ctx context.Context) {
    resp, err := httpClient.Do(req.WithContext(ctx))
    // Context cancellation aborts the HTTP request
})
```

**Multiple goroutines sharing one cancel:**

```go
ctx, cancel := context.WithCancel(context.Background())
for _, id := range userIDs {
    routine.RunWithContext(ctx, func(ctx context.Context) {
        fetchUser(ctx, id)
    })
}
// cancel() stops all goroutines
```

---

## RunWithTimeout

Starts a goroutine with an automatic deadline. Returns a `CancelFunc` for early cancellation.

If the function does not complete within the timeout, the context is cancelled and a warning is logged.

```go
cancel := routine.RunWithTimeout(5*time.Second, func(ctx context.Context) {
    // ctx will be cancelled after 5 seconds
    resp, err := httpClient.Do(req.WithContext(ctx))
})
defer cancel() // optional: cancel early if no longer needed
```

**Timeout reached — warning logged:**

```json
{
  "level": "warn",
  "msg": "goroutine timed out",
  "timeout": "5s"
}
```

**Early cancel:**

```go
cancel := routine.RunWithTimeout(30*time.Second, func(ctx context.Context) {
    <-ctx.Done() // exits when cancel() is called
})
cancel() // no need to wait 30s
```

---

## Group

Run multiple goroutines and wait for all to complete. Collects the first error, optionally cancels the context on failure.

```go
g := routine.NewGroupWithContext(ctx)
g.Go(func(ctx context.Context) error { return fetchUsers(ctx) })
g.Go(func(ctx context.Context) error { return fetchOrders(ctx) })
err := g.Wait() // first error cancels the other
```

### With concurrency limit

```go
g := routine.NewGroupWithLimit(ctx, 5) // max 5 concurrent goroutines
for _, id := range userIDs {
    id := id
    g.Go(func(ctx context.Context) error {
        return processUser(ctx, id)
    })
}
err := g.Wait()
```

### API

| Function | Description |
|---|---|
| `NewGroup()` | No limit, background context |
| `NewGroupWithContext(ctx)` | Cancel on first error |
| `NewGroupWithLimit(ctx, n)` | Cancel + concurrency limit |
| `g.Go(fn)` | Add a goroutine (blocks if limit reached) |
| `g.Wait() error` | Wait for all, return first error |

---

## WorkerPool

Fixed pool of worker goroutines processing jobs from a queue.

```go
pool := routine.NewWorkerPool(routine.PoolConfig{
    Workers:   10,
    QueueSize: 100,
})
pool.Start(ctx)

pool.Submit(func() { processItem(item) })

pool.Stop() // graceful: cancels context, drains queue, waits for workers
```

### Job-level timeout

```go
pool.SubmitWithTimeout(5*time.Second, func(ctx context.Context) {
    // ctx is cancelled after 5s OR when pool.Stop() is called
    resp, err := httpClient.Do(req.WithContext(ctx))
})
```

### API

| Function | Description |
|---|---|
| `NewWorkerPool(config)` | Create pool (Workers defaults to 1) |
| `pool.Start(ctx)` | Launch workers (idempotent) |
| `pool.Submit(fn) bool` | Enqueue (blocks if full, false if stopped) |
| `pool.TrySubmit(fn) bool` | Enqueue non-blocking (false if full/stopped) |
| `pool.SubmitWithTimeout(d, fn) bool` | Enqueue with job-level timeout |
| `pool.Stop()` | Graceful shutdown (idempotent) |
| `pool.Running() int` | Workers currently executing |
| `pool.Pending() int` | Jobs waiting in queue |

---

## FanOut

Process a slice in parallel and return **ordered results**. `output[i]` corresponds to `input[i]`.

```go
results, err := routine.FanOut(ctx, userIDs, 5,
    func(ctx context.Context, id string) (User, error) {
        return fetchUser(ctx, id)
    },
)
```

- `workers` is capped at `len(items)`. If `<= 0`, defaults to `1`.
- Returns the first error. Partial results are still available.
- Panics are recovered and returned as errors.
- Context cancellation stops processing.

---

## ForEach

Process a slice in parallel for **side-effects only** (no return values).

```go
err := routine.ForEach(ctx, emails, 10,
    func(ctx context.Context, email string) error {
        return sendEmail(ctx, email)
    },
)
```

---

## Production Safety

| Feature | Details |
|---|---|
| **Panic recovery** | All patterns recover panics — logged with source location |
| **Context propagation** | `RunWithContext`, `Group`, `WorkerPool`, `FanOut` all respect `ctx.Done()` |
| **Timeout support** | `RunWithTimeout`, `SubmitWithTimeout` auto-cancel after deadline |
| **Goroutine leak prevention** | Context cancellation signals goroutines to exit |
| **Graceful shutdown** | `WorkerPool.Stop()` cancels ctx → drains queue → waits for workers |
| **Idempotent operations** | `Start()`, `Stop()`, `cancel()` are safe to call multiple times |
| **Race-safe** | All shared state protected by atomics or mutexes |

## File Structure

```
goroutine/
├── run.go       — Run, RunWithContext, RunWithTimeout
├── recover.go   — Panic recovery + logger
├── invoke.go    — Reflect-based invocation
├── stack.go     — Stack trace parser + caller location
├── group.go     — Group pattern
├── worker.go    — WorkerPool + SubmitWithTimeout
├── pipeline.go  — FanOut / ForEach
└── docs/
    ├── README.md
    └── example/main.go
```

## Examples

See [`docs/example/main.go`](./example/main.go) for complete working examples covering all patterns.

## License

Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>
