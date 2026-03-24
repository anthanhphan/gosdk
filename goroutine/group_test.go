// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelInfo,
		LogEncoding:       logger.EncodingJSON,
		DisableStacktrace: true,
	})
}

// ---------------------------------------------------------------------------
// Group Tests
// ---------------------------------------------------------------------------

func TestGroup_Basic(t *testing.T) {
	var count atomic.Int32

	g := NewGroup()
	for i := 0; i < 10; i++ {
		g.Go(func(_ context.Context) error {
			count.Add(1)
			return nil
		})
	}

	err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, int32(10), count.Load())
}

func TestGroup_FirstError(t *testing.T) {
	expectedErr := errors.New("task failed")

	g := NewGroup()
	g.Go(func(_ context.Context) error {
		return expectedErr
	})
	g.Go(func(_ context.Context) error {
		return nil
	})

	err := g.Wait()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGroup_WithContext_CancelOnError(t *testing.T) {
	ctx := context.Background()
	g := NewGroupWithContext(ctx)

	g.Go(func(_ context.Context) error {
		return errors.New("fail fast")
	})

	// This goroutine should see the context cancelled
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return errors.New("should not reach here")
		}
	})

	err := g.Wait()
	assert.Error(t, err)
}

func TestGroup_WithLimit(t *testing.T) {
	var maxConcurrent atomic.Int32
	var current atomic.Int32

	g := NewGroupWithLimit(context.Background(), 3)

	for i := 0; i < 20; i++ {
		g.Go(func(_ context.Context) error {
			cur := current.Add(1)
			for {
				old := maxConcurrent.Load()
				if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			current.Add(-1)
			return nil
		})
	}

	err := g.Wait()
	assert.NoError(t, err)
	assert.LessOrEqual(t, maxConcurrent.Load(), int32(3))
}

func TestGroup_PanicRecovery(t *testing.T) {
	g := NewGroup()
	g.Go(func(_ context.Context) error {
		panic("group panic test")
	})

	err := g.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "group panic test")
}

func TestGroup_EmptyGroup(t *testing.T) {
	g := NewGroup()
	err := g.Wait()
	assert.NoError(t, err)
}

func TestGroup_ContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	g := NewGroupWithLimit(ctx, 2)

	executed := false
	g.Go(func(_ context.Context) error {
		executed = true // should not reach here -- semaphore select sees ctx.Done
		return nil
	})

	err := g.Wait()
	assert.Error(t, err)
	assert.False(t, executed)
}

// TestGroup_NoLimit_Error tests Group without semaphore collecting errors.
func TestGroup_NoLimit_Error(t *testing.T) {
	g := NewGroupWithContext(context.Background())
	g.Go(func(_ context.Context) error {
		return errors.New("task error")
	})
	g.Go(func(_ context.Context) error {
		return nil
	})
	err := g.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task error")
}

// TestGroup_SetError_AlreadySet tests that setError keeps only the first error.
func TestGroup_SetError_AlreadySet(t *testing.T) {
	g := NewGroup()
	ready := make(chan struct{})
	g.Go(func(_ context.Context) error {
		<-ready
		return errors.New("first")
	})
	close(ready) // let first goroutine run and set error
	time.Sleep(10 * time.Millisecond)
	g.Go(func(_ context.Context) error {
		return errors.New("second")
	})
	err := g.Wait()
	assert.Error(t, err)
	// Either "first" or "second" is acceptable -- we just verify only one error is kept
	assert.True(t, err.Error() == "first" || err.Error() == "second")
}

// TestGroup_WithLimit_NegativeLimit tests that limit <= 0 means no limit.
func TestGroup_WithLimit_NegativeLimit(t *testing.T) {
	g := NewGroupWithLimit(context.Background(), -1)
	var count atomic.Int32
	for i := 0; i < 5; i++ {
		g.Go(func(_ context.Context) error {
			count.Add(1)
			return nil
		})
	}
	err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, int32(5), count.Load())
}

// ---------------------------------------------------------------------------
// WorkerPool additional coverage
// ---------------------------------------------------------------------------

// TestWorkerPool_SubmitAfterContextCancel tests Submit when context is cancelled.
func TestWorkerPool_SubmitAfterContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 1})
	pool.Start(ctx)

	blocker := make(chan struct{})
	pool.Submit(func() { <-blocker }) // block the worker
	pool.Submit(func() {})            // fill the queue (size=1)

	cancel() // cancel context -- next Submit sees ctx.Done
	time.Sleep(10 * time.Millisecond)

	ok := pool.Submit(func() {})
	assert.False(t, ok)

	close(blocker)
	pool.Stop()
}

// TestWorkerPool_TrySubmitQueueFull tests TrySubmit when the queue is full.
func TestWorkerPool_TrySubmitQueueFull(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 1})
	pool.Start(context.Background())

	blocker := make(chan struct{})
	// Block the worker
	pool.Submit(func() { <-blocker })
	// Fill the queue
	pool.Submit(func() {})
	// Queue is now full -- TrySubmit should return false
	ok := pool.TrySubmit(func() {})
	assert.False(t, ok)

	close(blocker)
	pool.Stop()
}

// TestWorkerPool_PendingMetric tests the Pending metric.
func TestWorkerPool_PendingMetric(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 10})
	pool.Start(context.Background())

	blocker := make(chan struct{})
	pool.Submit(func() { <-blocker }) // block worker

	pool.Submit(func() {})
	pool.Submit(func() {})
	pool.Submit(func() {})

	pending := pool.Pending()
	assert.GreaterOrEqual(t, pending, 2)

	close(blocker)
	pool.Stop()
}

// TestWorkerPool_RunningMetric tests the Running metric.
func TestWorkerPool_RunningMetric(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 2, QueueSize: 5})
	pool.Start(context.Background())

	blocker := make(chan struct{})
	pool.Submit(func() { <-blocker })
	pool.Submit(func() { <-blocker })

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, pool.Running())

	close(blocker)
	pool.Stop()
}

// TestWorkerPool_StopCalledTwice tests that Stop is idempotent.
func TestWorkerPool_StopCalledTwice(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 5})
	pool.Start(context.Background())
	pool.Stop()
	pool.Stop() // should not panic
}
