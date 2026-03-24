// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// WorkerPool -- fixed workers + job queue
// ---------------------------------------------------------------------------

// PoolConfig configures a WorkerPool.
type PoolConfig struct {
	// Workers is the number of concurrent worker goroutines.
	// Must be >= 1. Defaults to 1 if not set.
	Workers int

	// QueueSize is the capacity of the job queue (buffered channel).
	// If 0, Submit will block until a worker picks up the job.
	// A larger value allows jobs to be enqueued without blocking.
	QueueSize int
}

// WorkerPool manages a fixed number of worker goroutines that process jobs
// from a shared queue. It supports graceful shutdown, panic recovery per job,
// and non-blocking submit.
//
// Example:
//
//	pool := routine.NewWorkerPool(routine.PoolConfig{Workers: 10, QueueSize: 100})
//	pool.Start(ctx)
//
//	pool.Submit(func() { processItem(item) })
//
//	pool.Stop() // graceful: drains queue, waits for workers
type WorkerPool struct {
	config  PoolConfig
	jobs    chan func()
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	started atomic.Bool
	stopped atomic.Bool
	running atomic.Int32
	pending atomic.Int32
}

// NewWorkerPool creates a WorkerPool with the given configuration.
// Call Start() to begin processing jobs.
func NewWorkerPool(config PoolConfig) *WorkerPool {
	if config.Workers < 1 {
		config.Workers = 1
	}
	return &WorkerPool{
		config: config,
		jobs:   make(chan func(), config.QueueSize),
	}
}

// Start launches the worker goroutines. Each worker pulls jobs from the
// shared queue and executes them. Panics are recovered and logged.
// Workers stop when the context is cancelled or Stop() is called.
// Safe to call multiple times — only the first call takes effect.
func (p *WorkerPool) Start(ctx context.Context) {
	if p.started.Swap(true) {
		return // already started
	}

	p.ctx, p.cancel = context.WithCancel(ctx)

	for i := 0; i < p.config.Workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Submit enqueues a job for execution. Blocks if the queue is full.
// Returns false if the pool has been stopped or the context is cancelled.
func (p *WorkerPool) Submit(fn func()) bool {
	p.pending.Add(1)

	// Use a select with a read on ctx.Done to avoid sending on a closed channel.
	// When Stop() closes p.jobs, workers drain it. We never send after close
	// because stopped is set before close.
	if p.stopped.Load() {
		p.pending.Add(-1)
		return false
	}

	select {
	case p.jobs <- fn:
		return true
	case <-p.ctx.Done():
		p.pending.Add(-1)
		return false
	}
}

// TrySubmit enqueues a job without blocking. Returns false if the queue
// is full or the pool has been stopped.
func (p *WorkerPool) TrySubmit(fn func()) bool {
	if p.stopped.Load() {
		return false
	}

	p.pending.Add(1)
	select {
	case p.jobs <- fn:
		return true
	default:
		p.pending.Add(-1)
		return false
	}
}

// SubmitWithTimeout enqueues a context-aware job with a timeout.
// The job receives a context that will be cancelled after the given timeout.
// The job function should respect ctx.Done() to exit promptly on timeout.
// Returns false if the pool has been stopped.
//
// Example:
//
//	pool.SubmitWithTimeout(5*time.Second, func(ctx context.Context) {
//	    resp, err := httpClient.Do(req.WithContext(ctx))
//	})
func (p *WorkerPool) SubmitWithTimeout(timeout time.Duration, fn func(ctx context.Context)) bool {
	return p.Submit(func() {
		ctx, cancel := context.WithTimeout(p.ctx, timeout)
		defer cancel()
		fn(ctx)
	})
}

// Stop gracefully shuts down the pool: closes the job queue, drains
// remaining jobs, and waits for all workers to finish.
// Safe to call multiple times.
func (p *WorkerPool) Stop() {
	if p.stopped.Swap(true) {
		return // already stopped
	}
	if p.cancel != nil {
		p.cancel() // cancel context first — unblocks ctx-aware jobs
	}
	close(p.jobs) // signal workers to drain and exit
	p.wg.Wait()   // wait for all workers to finish
}

// Running returns the number of workers currently executing a job.
func (p *WorkerPool) Running() int {
	return int(p.running.Load())
}

// Pending returns the number of jobs in the queue (approximate).
func (p *WorkerPool) Pending() int {
	return int(p.pending.Load())
}

// worker is the main loop for a single worker goroutine.
// It processes jobs from the queue until the channel is closed.
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for fn := range p.jobs {
		p.pending.Add(-1)
		p.running.Add(1)
		p.safeRun(fn)
		p.running.Add(-1)
	}
}

// safeRun executes a job function with panic recovery.
func (p *WorkerPool) safeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			getRecoverLogger().Errorw("panic recovered in worker pool",
				"type", "panic",
				"error", normalizePanicValue(r).Error(),
			)
		}
	}()

	fn()
}
