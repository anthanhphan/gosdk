// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package resilience

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	if cb.State() != StateClosed {
		t.Errorf("initial state = %v, want %v", cb.State(), StateClosed)
	}
	if !cb.Allow() {
		t.Error("expected Allow() = true in closed state")
	}
}

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	if cfg.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5", cfg.FailureThreshold)
	}
	if cfg.SuccessThreshold != 3 {
		t.Errorf("SuccessThreshold = %d, want 3", cfg.SuccessThreshold)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.HalfOpenMaxRequests != 3 {
		t.Errorf("HalfOpenMaxRequests = %d, want 3", cfg.HalfOpenMaxRequests)
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	for i := 0; i < 3; i++ {
		cb.Allow()
		cb.RecordResult(false)
	}

	if cb.State() != StateOpen {
		t.Errorf("state = %v, want open", cb.State())
	}
	if cb.Allow() {
		t.Error("expected Allow() = false in open state")
	}
}

func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 2,
	})
	cb.SetNowFn(func() time.Time { return now })

	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	if cb.State() != StateOpen {
		t.Fatalf("expected open, got %v", cb.State())
	}

	cb.SetNowFn(func() time.Time { return now.Add(200 * time.Millisecond) })

	if !cb.Allow() {
		t.Error("expected Allow() = true after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("state = %v, want half-open", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToClose(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	})
	cb.SetNowFn(func() time.Time { return now })

	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	cb.SetNowFn(func() time.Time { return now.Add(200 * time.Millisecond) })
	cb.Allow()

	cb.RecordResult(true)
	cb.RecordResult(true)

	if cb.State() != StateClosed {
		t.Errorf("state = %v, want closed", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          1 * time.Second,
	})

	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("state after reset = %v, want closed", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected Allow() = true after reset")
	}
}

func TestCircuitBreakerState_String(t *testing.T) {
	tests := []struct {
		state CircuitBreakerState
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{CircuitBreakerState(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("CircuitBreakerState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", cfg.MaxAttempts)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("Multiplier = %f, want 2.0", cfg.Multiplier)
	}
	if len(cfg.RetryableStatusCodes) == 0 {
		t.Error("expected non-empty RetryableStatusCodes")
	}
}

func TestRetryOptions(t *testing.T) {
	cfg := &RetryConfig{}

	WithMaxAttempts(5)(cfg)
	if cfg.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", cfg.MaxAttempts)
	}

	WithBackoff(200*time.Millisecond, 10*time.Second)(cfg)
	if cfg.InitialBackoff != 200*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 200ms", cfg.InitialBackoff)
	}
	if cfg.MaxBackoff != 10*time.Second {
		t.Errorf("MaxBackoff = %v, want 10s", cfg.MaxBackoff)
	}
}

func TestSetNowFn(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cb := NewCircuitBreaker(nil)
	cb.SetNowFn(func() time.Time { return fixed })

	// Force a failure to trigger nowFn
	cb.Allow()
	cb.RecordResult(false)

	// If SetNowFn works, the internal lastFailure should use our fixed time
	// We can't access it directly, but we can verify behavior:
	// The circuit should still be closed (only 1 failure, threshold 5)
	if cb.State() != StateClosed {
		t.Errorf("state = %v, want closed (only 1 failure)", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenMaxRequestsExceeded(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    3,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 2, // allow 2 requests in half-open
	})
	cb.SetNowFn(func() time.Time { return now })

	// Trip the circuit
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	// Transition to half-open (this Allow() resets counter to 0 and returns true)
	cb.SetNowFn(func() time.Time { return now.Add(200 * time.Millisecond) })
	if !cb.Allow() {
		t.Fatal("transition to half-open should allow")
	}

	// Second half-open request should be allowed (halfOpenRequests: 0 < 2)
	if !cb.Allow() {
		t.Fatal("second half-open request should be allowed")
	}

	// Third half-open request should be allowed (halfOpenRequests: 1 < 2)
	if !cb.Allow() {
		t.Fatal("third half-open request should be allowed")
	}

	// Fourth request should be denied (halfOpenRequests: 2 >= 2)
	if cb.Allow() {
		t.Error("fourth half-open request should be denied (max 2)")
	}
}

func TestCircuitBreaker_FailureInHalfOpen(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	})
	cb.SetNowFn(func() time.Time { return now })

	// Trip the circuit
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	// Transition to half-open
	cb.SetNowFn(func() time.Time { return now.Add(200 * time.Millisecond) })
	cb.Allow()

	// Record a failure in half-open — should re-open the circuit
	cb.RecordResult(false)

	if cb.State() != StateOpen {
		t.Errorf("state after failure in half-open = %v, want open", cb.State())
	}
}

func TestCircuitBreaker_RecordResultInOpenState(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          1 * time.Second,
	})

	// Trip the circuit
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)
	if cb.State() != StateOpen {
		t.Fatal("expected open state")
	}

	// RecordResult on open state should be no-op
	cb.RecordResult(true)
	cb.RecordResult(false)
	if cb.State() != StateOpen {
		t.Errorf("state = %v, should still be open", cb.State())
	}
}

func TestCircuitBreaker_Metrics(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	failures, successes, state := cb.Metrics()
	if failures != 2 {
		t.Errorf("failures = %d, want 2", failures)
	}
	if successes != 0 {
		t.Errorf("successes = %d, want 0", successes)
	}
	if state != StateClosed {
		t.Errorf("state = %v, want closed", state)
	}
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	var transitions []struct{ from, to CircuitBreakerState }
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    1,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 2,
		OnStateChange: func(from, to CircuitBreakerState) {
			transitions = append(transitions, struct{ from, to CircuitBreakerState }{from, to})
		},
	})
	now := time.Now()
	cb.SetNowFn(func() time.Time { return now })

	// Closed -> Open
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	// Open -> HalfOpen
	cb.SetNowFn(func() time.Time { return now.Add(200 * time.Millisecond) })
	cb.Allow()

	// HalfOpen -> Closed
	cb.RecordResult(true)

	if len(transitions) != 3 {
		t.Fatalf("transitions count = %d, want 3", len(transitions))
	}
	if transitions[0].from != StateClosed || transitions[0].to != StateOpen {
		t.Errorf("transition[0] = %v->%v, want Closed->Open", transitions[0].from, transitions[0].to)
	}
	if transitions[1].from != StateOpen || transitions[1].to != StateHalfOpen {
		t.Errorf("transition[1] = %v->%v, want Open->HalfOpen", transitions[1].from, transitions[1].to)
	}
	if transitions[2].from != StateHalfOpen || transitions[2].to != StateClosed {
		t.Errorf("transition[2] = %v->%v, want HalfOpen->Closed", transitions[2].from, transitions[2].to)
	}
}

func TestCircuitBreaker_SuccessResetFailures(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	// Record 2 failures (not enough to trip)
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)

	// Record a success — should reset failure counter
	cb.Allow()
	cb.RecordResult(true)

	failures, _, _ := cb.Metrics()
	if failures != 0 {
		t.Errorf("failures after success = %d, want 0", failures)
	}

	// Now 2 more failures should not trip (reset worked)
	cb.Allow()
	cb.RecordResult(false)
	cb.Allow()
	cb.RecordResult(false)
	if cb.State() != StateClosed {
		t.Errorf("state = %v, want closed (only 2 failures after reset)", cb.State())
	}
}

func TestCircuitBreaker_DefaultState(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 5,
		Timeout:          1 * time.Second,
	})

	// Verify default state returns via Allow()
	if cb.State() != StateClosed {
		t.Errorf("default state = %v, want closed", cb.State())
	}
}
