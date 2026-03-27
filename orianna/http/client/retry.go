// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"math"
	randv2 "math/rand/v2"
	"time"
)

// isRetryable checks if the request should be retried based on the
// HTTP method, response status code, and error.
// Uses pre-built maps for O(1) lookups instead of linear scans.
func (c *Client) isRetryable(method string, resp *Response, err error) bool {
	if c.retry == nil {
		return false
	}

	// Always retry on network errors for idempotent methods
	if err != nil {
		return c.isMethodRetryable(method)
	}

	if resp == nil {
		return false
	}

	// O(1) status code lookup via pre-built map
	if _, ok := c.retryStatusSet[resp.StatusCode]; !ok {
		return false
	}

	return c.isMethodRetryable(method)
}

// isMethodRetryable checks if the given HTTP method is retryable.
// Uses pre-built map for O(1) lookup. If RetryableMethods was empty,
// retryMethodSet is empty and all methods are retryable.
func (c *Client) isMethodRetryable(method string) bool {
	if len(c.retryMethodSet) == 0 {
		return true
	}
	_, ok := c.retryMethodSet[method]
	return ok
}

// calculateBackoff computes exponential backoff with jitter.
// Adds 0-25% jitter to prevent thundering herd when many clients
// retry simultaneously after a mass failure.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.retry.InitialBackoff) * math.Pow(c.retry.Multiplier, float64(attempt-1))
	if backoff > float64(c.retry.MaxBackoff) {
		backoff = float64(c.retry.MaxBackoff)
	}
	// Add 0-25% jitter (not security-sensitive, math/rand is fine for backoff)
	jitter := backoff * 0.25 * randv2.Float64() // #nosec G404
	return time.Duration(backoff + jitter)
}

// waitForRetry waits for the backoff duration before the next retry attempt.
// Returns ctx.Err() if the context is cancelled, or nil if the wait proceeds or
// the current attempt is already the last one.
func (c *Client) waitForRetry(ctx context.Context, req *Request, statusCode, attempt, maxAttempts int) error {
	if attempt >= maxAttempts {
		return nil
	}
	backoff := c.calculateBackoff(attempt)
	c.logger.Warnw("retrying request",
		"method", req.Method, "path", req.Path,
		"status", statusCode, "attempt", attempt,
		"max_attempts", maxAttempts, "backoff", backoff,
	)
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
