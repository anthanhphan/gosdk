// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"fmt"
	"sync"
)

// ---------------------------------------------------------------------------
// Pipeline -- Fan-out / Fan-in
// ---------------------------------------------------------------------------

// FanOut processes a slice of items concurrently using the given number of
// workers and returns ordered results. The output[i] corresponds to input[i].
//
// Panics inside the worker function are recovered and returned as errors.
// Processing stops early if the context is cancelled.
//
// Example:
//
//	results, err := routine.FanOut(ctx, userIDs, 5, func(ctx context.Context, id string) (User, error) {
//	    return fetchUser(ctx, id)
//	})
func FanOut[T any, R any](ctx context.Context, items []T, workers int, fn func(ctx context.Context, item T) (R, error)) ([]R, error) {
	n := len(items)
	if n == 0 {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}

	results := make([]R, n)
	errs := make([]error, n)

	// Use small buffer to avoid wasting memory for large n.
	bufSize := workers * 2
	if bufSize > n {
		bufSize = n
	}
	jobs := make(chan int, bufSize)

	var wg sync.WaitGroup
	wg.Add(workers)

	// Start workers
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				select {
				case <-ctx.Done():
					errs[idx] = ctx.Err()
					return
				default:
				}

				val, err := pipelineSafeCall(ctx, items[idx], fn)
				results[idx] = val
				errs[idx] = err
			}
		}()
	}

	// Enqueue indices — respect context cancellation to avoid deadlock.
enqueue:
	for i := 0; i < n; i++ {
		select {
		case jobs <- i:
		case <-ctx.Done():
			break enqueue
		}
	}
	close(jobs)

	wg.Wait()

	// Find first error
	for _, err := range errs {
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

// ForEach processes a slice of items concurrently using the given number of
// workers. Unlike FanOut, it does not return results -- only an error.
//
// Example:
//
//	err := routine.ForEach(ctx, items, 10, func(ctx context.Context, item Item) error {
//	    return processItem(ctx, item)
//	})
func ForEach[T any](ctx context.Context, items []T, workers int, fn func(ctx context.Context, item T) error) error {
	_, err := FanOut(ctx, items, workers, func(ctx context.Context, item T) (struct{}, error) {
		return struct{}{}, fn(ctx, item)
	})
	return err
}

// pipelineSafeCall invokes fn with panic recovery for pipeline workers.
func pipelineSafeCall[T any, R any](ctx context.Context, item T, fn func(ctx context.Context, item T) (R, error)) (result R, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			getRecoverLogger().Errorw("panic recovered in pipeline",
				"type", "panic",
				"error", err.Error(),
			)
		}
	}()

	return fn(ctx, item)
}
