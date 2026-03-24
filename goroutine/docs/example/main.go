// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	crypto_rand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	routine "github.com/anthanhphan/gosdk/goroutine"
	"github.com/anthanhphan/gosdk/logger"
)

func main() {
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelInfo,
		LogEncoding:       logger.EncodingConsole,
		DisableStacktrace: true,
	})
	defer undo()

	demoRun()
	demoRunWithContext()
	demoRunWithTimeout()
	demoGroup()
	demoGroupWithLimit()
	demoWorkerPool()
	demoSubmitWithTimeout()
	demoFanOut()
	demoForEach()
	demoPanicRecovery()

	logger.Info("All demos completed")
}

// ---------------------------------------------------------------------------
// 1. Run -- fire-and-forget with panic recovery
// ---------------------------------------------------------------------------

func demoRun() {
	section("routine.Run")

	routine.Run(func() { logger.Info("[Run] No-arg fast path") })
	routine.Run(func(s string) { logger.Infof("[Run] String fast path: %s", s) }, "hello")
	routine.Run(func(a, b int) { logger.Infof("[Run] Reflect path: %d+%d=%d", a, b, a+b) }, 10, 20)

	time.Sleep(50 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// 2. RunWithContext -- context-aware goroutine (prevents leak)
// ---------------------------------------------------------------------------

func demoRunWithContext() {
	section("routine.RunWithContext")

	// Scenario 1: Normal completion with context
	done := make(chan struct{})
	routine.RunWithContext(context.Background(), func(_ context.Context) {
		logger.Info("[WithCtx] Task completed normally")
		close(done)
	})
	<-done

	// Scenario 2: Context cancelled mid-work — goroutine exits promptly
	ctx, cancel := context.WithCancel(context.Background())
	exited := make(chan struct{})

	routine.RunWithContext(ctx, func(ctx context.Context) {
		logger.Info("[WithCtx] Long task started, waiting for cancel...")
		select {
		case <-ctx.Done():
			logger.Infof("[WithCtx] Detected cancel: %v — exiting cleanly", ctx.Err())
		case <-time.After(10 * time.Second):
			logger.Info("[WithCtx] Should not reach here")
		}
		close(exited)
	})

	time.Sleep(50 * time.Millisecond)
	cancel() // cancel the goroutine
	<-exited
	logger.Info("[WithCtx] Goroutine exited — no leak")
}

// ---------------------------------------------------------------------------
// 3. RunWithTimeout -- auto-cancel goroutine after deadline
// ---------------------------------------------------------------------------

func demoRunWithTimeout() {
	section("routine.RunWithTimeout")

	// Scenario 1: Task finishes before timeout
	done := make(chan struct{})
	cancelFn := routine.RunWithTimeout(2*time.Second, func(_ context.Context) {
		logger.Info("[Timeout] Fast task done (within 2s deadline)")
		close(done)
	})
	<-done
	cancelFn()

	// Scenario 2: Task exceeds timeout — context cancelled, goroutine exits
	exited := make(chan struct{})
	cancelFn = routine.RunWithTimeout(100*time.Millisecond, func(ctx context.Context) {
		logger.Info("[Timeout] Slow API call started (timeout=100ms)...")
		select {
		case <-ctx.Done():
			logger.Infof("[Timeout] Timed out: %v — releasing resources", ctx.Err())
		case <-time.After(5 * time.Second):
			logger.Info("[Timeout] Should not reach here")
		}
		close(exited)
	})
	<-exited
	cancelFn()
	logger.Info("[Timeout] Goroutine cleaned up — no leak")

	// Scenario 3: Early cancel before timeout
	exited2 := make(chan struct{})
	cancelFn = routine.RunWithTimeout(10*time.Second, func(ctx context.Context) {
		logger.Info("[Timeout] Waiting for early cancel or 10s timeout...")
		<-ctx.Done()
		logger.Infof("[Timeout] Early cancel received: %v", ctx.Err())
		close(exited2)
	})
	time.Sleep(50 * time.Millisecond)
	cancelFn() // cancel early — don't wait 10s
	<-exited2
	logger.Info("[Timeout] Early cancel worked — no 10s wait")
}

// ---------------------------------------------------------------------------

func demoGroup() {
	section("routine.Group")

	g := routine.NewGroupWithContext(context.Background())
	for _, name := range []string{"A", "B", "C"} {
		name := name
		g.Go(func(_ context.Context) error {
			logger.Infof("[Group] Task %s done", name)
			return nil
		})
	}

	logResult("[Group]", g.Wait())
}

// ---------------------------------------------------------------------------
// 3. Group with concurrency limit
// ---------------------------------------------------------------------------

func demoGroupWithLimit() {
	section("routine.GroupWithLimit")

	g := routine.NewGroupWithLimit(context.Background(), 2)
	for i := 1; i <= 6; i++ {
		i := i
		g.Go(func(_ context.Context) error {
			logger.Infof("[Limit] Task %d start", i)
			time.Sleep(20 * time.Millisecond)
			logger.Infof("[Limit] Task %d done", i)
			return nil
		})
	}

	logResult("[Limit]", g.Wait())
}

// ---------------------------------------------------------------------------
// 4. WorkerPool -- fixed workers + job queue
// ---------------------------------------------------------------------------

func demoWorkerPool() {
	section("routine.WorkerPool")

	pool := routine.NewWorkerPool(routine.PoolConfig{Workers: 3, QueueSize: 10})
	pool.Start(context.Background())

	for i := 1; i <= 9; i++ {
		i := i
		pool.Submit(func() {
			logger.Infof("[Pool] Job %d done", i)
			time.Sleep(10 * time.Millisecond)
		})
	}

	pool.Stop()
	logger.Info("[Pool] Stopped gracefully")
	logger.Infof("[Pool] Submit after stop: %v", pool.Submit(func() {}))
}

// ---------------------------------------------------------------------------
// 5. SubmitWithTimeout -- job-level timeout in WorkerPool
// ---------------------------------------------------------------------------

func demoSubmitWithTimeout() {
	section("routine.SubmitWithTimeout")

	pool := routine.NewWorkerPool(routine.PoolConfig{Workers: 2, QueueSize: 10})
	pool.Start(context.Background())

	// Job 1: completes within timeout
	done := make(chan struct{}, 1)
	pool.SubmitWithTimeout(2*time.Second, func(_ context.Context) {
		logger.Info("[PoolTimeout] Fast job completed within 2s deadline")
		done <- struct{}{}
	})
	<-done

	// Job 2: exceeds timeout — ctx cancelled, job exits promptly
	exited := make(chan struct{})
	pool.SubmitWithTimeout(100*time.Millisecond, func(ctx context.Context) {
		logger.Info("[PoolTimeout] Slow job started (timeout=100ms)...")
		<-ctx.Done()
		logger.Infof("[PoolTimeout] Job timed out: %v", ctx.Err())
		close(exited)
	})
	<-exited

	pool.Stop()
	logger.Info("[PoolTimeout] Pool stopped — no leaked workers")
}

// ---------------------------------------------------------------------------
// 6. FanOut -- parallel map with ordered results
// ---------------------------------------------------------------------------

func demoFanOut() {
	section("routine.FanOut")

	ids := []int{101, 102, 103, 104, 105}
	names, err := routine.FanOut(context.Background(), ids, 3,
		func(_ context.Context, id int) (string, error) {
			n, _ := crypto_rand.Int(crypto_rand.Reader, big.NewInt(20))
			time.Sleep(time.Duration(n.Int64()) * time.Millisecond)
			return fmt.Sprintf("User-%d", id), nil
		},
	)
	if err != nil {
		logger.Errorf("[FanOut] %v", err)
		return
	}
	for i, name := range names {
		logger.Infof("[FanOut] %d -> %s", ids[i], name)
	}
}

// ---------------------------------------------------------------------------
// 6. ForEach -- parallel side-effects
// ---------------------------------------------------------------------------

func demoForEach() {
	section("routine.ForEach")

	emails := []string{"email-1", "email-2", "email-3", "email-4", "email-5"}
	err := routine.ForEach(context.Background(), emails, 3,
		func(_ context.Context, e string) error {
			logger.Infof("[ForEach] Sent %s", e)
			return nil
		},
	)

	logResult("[ForEach]", err)
}

// ---------------------------------------------------------------------------
// 7. Panic recovery across all patterns
// ---------------------------------------------------------------------------

func demoPanicRecovery() {
	section("Panic Recovery")

	// Run
	routine.Run(func() { panic("Run panic!") })
	time.Sleep(50 * time.Millisecond)

	// Group -- panic becomes error
	g := routine.NewGroup()
	g.Go(func(_ context.Context) error { panic("Group panic!") })
	logger.Infof("[Panic] Group: %s", firstLine(g.Wait().Error()))

	// WorkerPool -- worker survives
	pool := routine.NewWorkerPool(routine.PoolConfig{Workers: 1, QueueSize: 5})
	pool.Start(context.Background())
	pool.Submit(func() { panic("Pool panic!") })
	pool.Submit(func() { logger.Info("[Panic] Pool worker alive after panic") })
	pool.Stop()

	// FanOut -- other results preserved
	res, err := routine.FanOut(context.Background(), []int{1, 2, 3}, 2,
		func(_ context.Context, n int) (string, error) {
			if n == 2 {
				panic("FanOut panic!")
			}
			return fmt.Sprintf("ok-%d", n), nil
		},
	)
	if err != nil {
		logger.Infof("[Panic] FanOut: %s  results: [%s, _, %s]", firstLine(err.Error()), res[0], res[2])
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func section(name string) { logger.Infof("--- Demo: %s ---", name) }

func logResult(prefix string, err error) {
	if err != nil {
		logger.Errorf("%s Error: %v", prefix, err)
	} else {
		logger.Infof("%s All done", prefix)
	}
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}
