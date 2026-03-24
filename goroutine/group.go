// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"fmt"
	"sync"
)

// ---------------------------------------------------------------------------
// Group -- fire-and-wait with panic recovery
// ---------------------------------------------------------------------------

// Group runs multiple goroutines concurrently and waits for all to finish.
// It collects the first error and optionally cancels the context on failure.
// Panics inside goroutines are recovered and returned as errors.
//
// Example:
//
//	g := routine.NewGroup()
//	g.Go(func(ctx context.Context) error {
//	    return doWork(ctx)
//	})
//	if err := g.Wait(); err != nil {
//	    log.Fatal(err)
//	}
type Group struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg  sync.WaitGroup
	mu  sync.Mutex
	err error         // first error
	sem chan struct{} // concurrency limiter (nil = unlimited)
}

// NewGroup creates a Group with no concurrency limit and a background context.
func NewGroup() *Group {
	return &Group{ctx: context.Background()}
}

// NewGroupWithContext creates a Group bound to the given context.
// The returned cancel function is called automatically when the first error occurs.
func NewGroupWithContext(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{ctx: ctx, cancel: cancel}
}

// NewGroupWithLimit creates a Group bound to the given context with a maximum
// number of concurrent goroutines. If limit <= 0, there is no limit.
func NewGroupWithLimit(ctx context.Context, limit int) *Group {
	g := NewGroupWithContext(ctx)
	if limit > 0 {
		g.sem = make(chan struct{}, limit)
	}
	return g
}

// Go starts a new goroutine to execute fn. If a concurrency limit is set,
// Go blocks until a slot is available or the context is cancelled.
// Panics inside fn are recovered, logged, and returned as errors.
func (g *Group) Go(fn func(ctx context.Context) error) {
	// Acquire semaphore slot if limit is set
	if g.sem != nil {
		// Fast check -- avoid select non-determinism when already cancelled
		if g.ctx.Err() != nil {
			g.setError(g.ctx.Err())
			return
		}
		select {
		case g.sem <- struct{}{}:
		case <-g.ctx.Done():
			g.setError(g.ctx.Err())
			return
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer func() { <-g.sem }()
		}

		if err := g.safeCall(fn); err != nil {
			g.setError(err)
		}
	}()
}

// Wait blocks until all goroutines added via Go have completed.
// Returns the first error (if any) encountered by any goroutine.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// setError records the first error and cancels the context.
// Uses manual unlock instead of defer for the fast path (error already set).
func (g *Group) setError(err error) {
	g.mu.Lock()
	if g.err != nil {
		g.mu.Unlock()
		return
	}
	g.err = err
	cancel := g.cancel
	g.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// safeCall invokes fn with panic recovery.
// Panics are converted to errors without debug.Stack() overhead.
func (g *Group) safeCall(fn func(ctx context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			getRecoverLogger().Errorw("panic recovered in group goroutine",
				"type", "panic",
				"error", err.Error(),
			)
		}
	}()

	return fn(g.ctx)
}
