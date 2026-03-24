// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Public API -- Run
// ---------------------------------------------------------------------------

// Run starts a new goroutine and invokes the provided function with the given arguments.
// Any panic that occurs within the goroutine is recovered and logged with panic location.
//
// For common function signatures (func(), func(string), func(int), func(error)),
// reflect is bypassed entirely for maximum performance.
//
// WARNING: This function provides no timeout or cancellation. The goroutine
// runs until fn completes. For production use with external calls (API, DB),
// prefer RunWithContext or RunWithTimeout to prevent goroutine leaks.
//
// Example:
//
//	routine.Run(func() {
//	    logger.Info("fire and forget")
//	})
func Run(fn any, args ...any) {
	// Fast paths for common function signatures -- skip reflect entirely
	switch f := fn.(type) {
	case func():
		if len(args) == 0 {
			go func() {
				defer recoverPanic()
				f()
			}()
			return
		}
	case func(error):
		if len(args) == 1 {
			if a, ok := args[0].(error); ok {
				go func() {
					defer recoverPanic()
					f(a)
				}()
				return
			}
		}
	case func(string):
		if len(args) == 1 {
			if a, ok := args[0].(string); ok {
				go func() {
					defer recoverPanic()
					f(a)
				}()
				return
			}
		}
	case func(int):
		if len(args) == 1 {
			if a, ok := args[0].(int); ok {
				go func() {
					defer recoverPanic()
					f(a)
				}()
				return
			}
		}
	}

	// Generic path: use reflect for other function signatures
	go func() {
		defer recoverPanic()
		invoke(fn, args)
	}()
}

// RunWithContext starts a goroutine that executes fn with the given context.
// If the context is cancelled or times out, the function is expected to
// observe ctx.Done() and return. The goroutine logs a warning if fn has not
// returned when the context expires.
//
// This is the recommended way to run goroutines in production — it prevents
// goroutine leaks by making the caller responsible for setting a deadline.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
//	defer cancel()
//
//	routine.RunWithContext(ctx, func(ctx context.Context) {
//	    resp, err := httpClient.Do(req.WithContext(ctx))
//	    // ctx cancellation will abort the HTTP request
//	})
func RunWithContext(ctx context.Context, fn func(ctx context.Context)) {
	go func() {
		defer recoverPanic()
		fn(ctx)
	}()
}

// RunWithTimeout starts a goroutine that executes fn with a timeout.
// A context with the given timeout is created and passed to fn.
// If fn does not complete within the timeout, the context is cancelled,
// and a warning is logged. fn should respect ctx.Done() to exit promptly.
//
// Returns a cancel function that can be called to cancel early.
//
// Example:
//
//	cancel := routine.RunWithTimeout(5*time.Second, func(ctx context.Context) {
//	    // This ctx will be cancelled after 5 seconds
//	    resp, err := http.Get("https://api.example.com/slow")
//	})
//	defer cancel() // optional: cancel early if no longer needed
func RunWithTimeout(timeout time.Duration, fn func(ctx context.Context)) context.CancelFunc {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	go func() {
		defer recoverPanic()
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			defer recoverPanic()
			fn(ctx)
		}()

		select {
		case <-done:
			// fn completed normally
		case <-ctx.Done():
			getRecoverLogger().Warnw("goroutine timed out",
				"timeout", timeout.String(),
			)
			// Wait for fn to actually return (it should observe ctx.Done)
			<-done
		}
	}()

	return cancel
}
