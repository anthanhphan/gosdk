// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// WorkerPool Tests
// ---------------------------------------------------------------------------

func TestWorkerPool_Basic(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 4, QueueSize: 10})
	pool.Start(context.Background())

	var count atomic.Int32
	for i := 0; i < 20; i++ {
		ok := pool.Submit(func() {
			count.Add(1)
		})
		assert.True(t, ok)
	}

	pool.Stop()
	assert.Equal(t, int32(20), count.Load())
}

func TestWorkerPool_TrySubmit(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 1})
	pool.Start(context.Background())

	// Fill queue and worker
	blocker := make(chan struct{})
	pool.Submit(func() { <-blocker }) // blocks the worker
	pool.Submit(func() {})            // fills the queue

	// Queue is full -- TrySubmit should return false
	ok := pool.TrySubmit(func() {})
	assert.False(t, ok)

	close(blocker) // unblock
	pool.Stop()
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 2, QueueSize: 100})
	pool.Start(context.Background())

	var count atomic.Int32
	for i := 0; i < 50; i++ {
		pool.Submit(func() {
			time.Sleep(1 * time.Millisecond)
			count.Add(1)
		})
	}

	pool.Stop() // should drain all 50 jobs
	assert.Equal(t, int32(50), count.Load())
}

func TestWorkerPool_PanicRecovery(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 2, QueueSize: 10})
	pool.Start(context.Background())

	var count atomic.Int32

	// Submit a panicking job followed by normal jobs
	pool.Submit(func() { panic("worker panic test") })
	for i := 0; i < 5; i++ {
		pool.Submit(func() { count.Add(1) })
	}

	pool.Stop()
	// Worker should have recovered from panic and continued processing
	assert.Equal(t, int32(5), count.Load())
}

func TestWorkerPool_SubmitAfterStop(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 1})
	pool.Start(context.Background())
	pool.Stop()

	ok := pool.Submit(func() {})
	assert.False(t, ok)

	ok = pool.TrySubmit(func() {})
	assert.False(t, ok)
}

func TestWorkerPool_DoubleStop(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 1})
	pool.Start(context.Background())

	pool.Stop()
	pool.Stop() // should not panic
}

func TestWorkerPool_Metrics(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 2, QueueSize: 10})
	pool.Start(context.Background())

	blocker := make(chan struct{})
	pool.Submit(func() { <-blocker })
	pool.Submit(func() { <-blocker })

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, pool.Running())

	close(blocker)
	pool.Stop()
}

func TestWorkerPool_DefaultWorkers(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 0}) // should default to 1
	pool.Start(context.Background())

	var count atomic.Int32
	pool.Submit(func() { count.Add(1) })
	pool.Stop()

	assert.Equal(t, int32(1), count.Load())
}

// Start called twice should be idempotent — no duplicate workers.
func TestWorkerPool_DoubleStart(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 2, QueueSize: 10})
	pool.Start(context.Background())
	pool.Start(context.Background()) // second call should be no-op
	pool.Start(context.Background()) // third call should be no-op

	var count atomic.Int32
	for i := 0; i < 10; i++ {
		pool.Submit(func() { count.Add(1) })
	}
	pool.Stop()

	assert.Equal(t, int32(10), count.Load())
}

// Submit should return false when parent context is cancelled.
func TestWorkerPool_SubmitWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPool(PoolConfig{Workers: 1, QueueSize: 0})
	pool.Start(ctx)

	// Block the single worker
	blocker := make(chan struct{})
	pool.Submit(func() { <-blocker })

	// Cancel the context — Submit should return false
	cancel()
	time.Sleep(20 * time.Millisecond)

	ok := pool.Submit(func() {})
	assert.False(t, ok, "Submit should fail when context is cancelled")

	close(blocker)
	pool.Stop()
}

// Concurrent Submit calls should be safe.
func TestWorkerPool_ConcurrentSubmit(t *testing.T) {
	pool := NewWorkerPool(PoolConfig{Workers: 4, QueueSize: 100})
	pool.Start(context.Background())

	var count atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Submit(func() { count.Add(1) })
		}()
	}

	wg.Wait()
	pool.Stop()
	assert.Equal(t, int32(100), count.Load())
}
