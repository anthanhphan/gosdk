// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package resilience

import (
	"sync"
	"testing"
	"time"
)

// BenchmarkCircuitBreakerAllow measures the throughput of Allow() under no contention.
// This is the hot-path check on every request.
func BenchmarkCircuitBreakerAllow(b *testing.B) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    3,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Allow()
	}
}

// BenchmarkCircuitBreakerAllowParallel measures concurrent Allow() throughput.
// This simulates production traffic where multiple goroutines check the circuit.
func BenchmarkCircuitBreakerAllowParallel(b *testing.B) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    3,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	})

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Allow()
		}
	})
}

// BenchmarkCircuitBreakerRecordResult measures RecordResult() throughput.
func BenchmarkCircuitBreakerRecordResult(b *testing.B) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    1000000, // high threshold to stay closed
		SuccessThreshold:    3,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Allow()
		cb.RecordResult(true)
	}
}

// BenchmarkCircuitBreakerMetrics measures Metrics() read throughput.
func BenchmarkCircuitBreakerMetrics(b *testing.B) {
	cb := NewCircuitBreaker(nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Metrics()
	}
}

// BenchmarkCircuitBreakerWithCallback measures Allow()+RecordResult() with OnStateChange set.
func BenchmarkCircuitBreakerWithCallback(b *testing.B) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    1000000,
		SuccessThreshold:    3,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
		OnStateChange: func(from, to CircuitBreakerState) {
			// noop callback
		},
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Allow()
		cb.RecordResult(true)
	}
}

// TestCircuitBreaker_ConcurrentRace stress-tests concurrent access for race detection.
// Run with: go test -race -count=1
func TestCircuitBreaker_ConcurrentRace(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    3,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
		OnStateChange: func(from, to CircuitBreakerState) {
			// Verify callback is called safely under contention
		},
	})

	var wg sync.WaitGroup
	goroutines := 100
	iterations := 1000

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				cb.Allow()
				cb.RecordResult(i%3 != 0) // mix of success and failure
				cb.State()
				cb.Metrics()
			}
		}()
	}
	wg.Wait()

	// Verify we can still read state after contention
	_ = cb.State()
	f, s, st := cb.Metrics()
	_ = f
	_ = s
	_ = st
}

// TestCircuitBreaker_OnStateChangeCallback verifies the callback fires on transitions.
func TestCircuitBreaker_OnStateChangeCallback(t *testing.T) {
	var mu sync.Mutex
	transitions := []struct{ from, to CircuitBreakerState }{}

	now := time.Now()
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
		OnStateChange: func(from, to CircuitBreakerState) {
			mu.Lock()
			transitions = append(transitions, struct{ from, to CircuitBreakerState }{from, to})
			mu.Unlock()
		},
	})
	cb.SetNowFn(func() time.Time { return now })

	// Closed -> Open
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	mu.Lock()
	if len(transitions) != 1 || transitions[0].from != StateClosed || transitions[0].to != StateOpen {
		t.Fatalf("expected Closed->Open, got %v", transitions)
	}
	mu.Unlock()

	// Open -> HalfOpen
	cb.SetNowFn(func() time.Time { return now.Add(200 * time.Millisecond) })
	cb.Allow()

	mu.Lock()
	if len(transitions) != 2 || transitions[1].from != StateOpen || transitions[1].to != StateHalfOpen {
		t.Fatalf("expected Open->HalfOpen, got %v", transitions)
	}
	mu.Unlock()

	// HalfOpen -> Closed
	cb.RecordResult(true)
	cb.RecordResult(true)

	mu.Lock()
	if len(transitions) != 3 || transitions[2].from != StateHalfOpen || transitions[2].to != StateClosed {
		t.Fatalf("expected HalfOpen->Closed, got %v", transitions)
	}
	mu.Unlock()
}
