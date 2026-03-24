// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Pipeline Tests
// ---------------------------------------------------------------------------

func TestFanOut_Basic(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results, err := FanOut(ctx(), items, 3, func(_ context.Context, item int) (int, error) {
		return item * 2, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6, 8, 10}, results)
}

func TestFanOut_OrderPreserved(t *testing.T) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	results, err := FanOut(ctx(), items, 10, func(_ context.Context, item int) (string, error) {
		return fmt.Sprintf("item-%d", item), nil
	})

	assert.NoError(t, err)
	for i, r := range results {
		assert.Equal(t, fmt.Sprintf("item-%d", i), r)
	}
}

func TestFanOut_Error(t *testing.T) {
	items := []int{1, 2, 3}

	results, err := FanOut(ctx(), items, 2, func(_ context.Context, item int) (int, error) {
		if item == 2 {
			return 0, errors.New("item 2 failed")
		}
		return item, nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item 2 failed")
	assert.Len(t, results, 3)
}

func TestFanOut_EmptySlice(t *testing.T) {
	results, err := FanOut(ctx(), []int{}, 5, func(_ context.Context, item int) (int, error) {
		return item, nil
	})

	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestFanOut_PanicRecovery(t *testing.T) {
	items := []int{1, 2, 3}

	results, err := FanOut(ctx(), items, 2, func(_ context.Context, item int) (int, error) {
		if item == 2 {
			panic("pipeline panic test")
		}
		return item * 10, nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Len(t, results, 3)
	// Non-panicking items should have correct results
	assert.Equal(t, 10, results[0])
	assert.Equal(t, 30, results[2])
}

func TestFanOut_ContextCancelled(t *testing.T) {
	cctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	items := []int{1, 2, 3}
	_, err := FanOut(cctx, items, 2, func(ctx context.Context, item int) (int, error) {
		return item, nil
	})

	assert.Error(t, err)
}

func TestFanOut_WorkersExceedItems(t *testing.T) {
	items := []int{1, 2}

	results, err := FanOut(ctx(), items, 100, func(_ context.Context, item int) (int, error) {
		return item, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, results)
}

func TestForEach_Basic(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	sum := 0
	ch := make(chan int, len(items))

	err := ForEach(ctx(), items, 3, func(_ context.Context, item int) error {
		ch <- item
		return nil
	})
	close(ch)

	assert.NoError(t, err)
	for v := range ch {
		sum += v
	}
	assert.Equal(t, 15, sum)
}

func TestForEach_Error(t *testing.T) {
	items := []string{"a", "b", "c"}

	err := ForEach(ctx(), items, 2, func(_ context.Context, item string) error {
		if item == "b" {
			return errors.New("b failed")
		}
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "b failed")
}

// ctx returns a background context for tests.
func ctx() context.Context {
	return context.Background()
}

// FanOut with context cancelled during enqueue should not deadlock.
func TestFanOut_ContextCancelledDuringEnqueue(t *testing.T) {
	cctx, cancel := context.WithCancel(context.Background())

	// Large dataset — cancel during processing so enqueue loop hits ctx.Done
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	var processed atomic.Int32
	_, err := FanOut(cctx, items, 2, func(_ context.Context, item int) (int, error) {
		processed.Add(1)
		if processed.Load() == 5 {
			cancel() // cancel after 5 items processed
		}
		time.Sleep(time.Millisecond)
		return item, nil
	})

	assert.Error(t, err, "should return error after context cancelled")
}

// FanOut with zero workers should default to 1.
func TestFanOut_ZeroWorkers(t *testing.T) {
	items := []int{1, 2, 3}
	results, err := FanOut(ctx(), items, 0, func(_ context.Context, item int) (int, error) {
		return item * 2, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6}, results)
}

// FanOut with negative workers should default to 1.
func TestFanOut_NegativeWorkers(t *testing.T) {
	items := []int{1}
	results, err := FanOut(ctx(), items, -5, func(_ context.Context, item int) (int, error) {
		return item, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1}, results)
}

// FanOut with single item.
func TestFanOut_SingleItem(t *testing.T) {
	items := []string{"hello"}
	results, err := FanOut(ctx(), items, 10, func(_ context.Context, s string) (int, error) {
		return len(s), nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{5}, results)
}

// ForEach with context already cancelled.
func TestForEach_ContextCancelled(t *testing.T) {
	cctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ForEach(cctx, []int{1, 2, 3}, 2, func(_ context.Context, _ int) error {
		return nil
	})

	assert.Error(t, err)
}

// ForEach with empty slice.
func TestForEach_EmptySlice(t *testing.T) {
	err := ForEach(ctx(), []int{}, 5, func(_ context.Context, _ int) error {
		t.Error("should not be called")
		return nil
	})

	assert.NoError(t, err)
}

// FanOut with all items returning errors.
func TestFanOut_AllErrors(t *testing.T) {
	items := []int{1, 2, 3}
	results, err := FanOut(ctx(), items, 3, func(_ context.Context, _ int) (int, error) {
		return 0, errors.New("all fail")
	})

	assert.Error(t, err)
	assert.Len(t, results, 3)
}
