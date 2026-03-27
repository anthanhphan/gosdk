// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package resilience

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker.
type CircuitBreakerState int

const (
	// StateClosed means the circuit is closed, requests pass through normally.
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit is open, requests are rejected.
	StateOpen
	// StateHalfOpen means the circuit is half-open, testing if it should close.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds circuit breaker configuration.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	FailureThreshold int

	// SuccessThreshold is the number of consecutive successes needed to close the circuit.
	SuccessThreshold int

	// Timeout is how long the circuit stays open before attempting to close.
	Timeout time.Duration

	// HalfOpenMaxRequests is the max requests allowed in half-open state.
	HalfOpenMaxRequests int

	// OnStateChange is called when the circuit breaker transitions between states.
	// Useful for alerting (PagerDuty, Grafana), audit logging, and metrics.
	// Called under lock — implementations must not block.
	OnStateChange func(from, to CircuitBreakerState)
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    3,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent cascade failures.
type CircuitBreaker struct {
	config           *CircuitBreakerConfig
	state            CircuitBreakerState
	failures         int
	successes        int
	lastFailure      time.Time
	halfOpenRequests int
	nowFn            func() time.Time // injectable clock for testing
	mu               sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		nowFn:  time.Now,
	}
}

// Allow returns true if a request is allowed to proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has elapsed
		if cb.nowFn().Sub(cb.lastFailure) > cb.config.Timeout {
			cb.transitionTo(StateHalfOpen)
			cb.halfOpenRequests = 0
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenRequests < cb.config.HalfOpenMaxRequests {
			cb.halfOpenRequests++
			return true
		}
		return false
	default:
		return false
	}
}

// RecordResult records the result of a request for circuit breaker logic.
func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		if success {
			cb.successes++
			if cb.successes >= cb.config.SuccessThreshold {
				// Enough successes, close the circuit
				cb.transitionTo(StateClosed)
				cb.failures = 0
				cb.successes = 0
			}
		} else {
			// Failure in half-open, reopen the circuit
			cb.transitionTo(StateOpen)
			cb.lastFailure = cb.nowFn()
			cb.successes = 0
		}
		return
	}

	if cb.state != StateClosed {
		return
	}

	if success {
		// Reset failures on success
		cb.failures = 0
	} else {
		cb.failures++
		if cb.failures >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
			cb.lastFailure = cb.nowFn()
		}
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenRequests = 0
}

// SetNowFn sets the time function used by the circuit breaker.
// This is intended for testing only — inject a mock clock to control time.
func (cb *CircuitBreaker) SetNowFn(fn func() time.Time) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.nowFn = fn
}

// Metrics returns the current circuit breaker counters for external observability.
// Useful for exposing as Prometheus gauges.
func (cb *CircuitBreaker) Metrics() (failures, successes int, state CircuitBreakerState) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures, cb.successes, cb.state
}

// transitionTo changes the circuit breaker state and invokes the OnStateChange callback.
// Must be called while holding cb.mu.
func (cb *CircuitBreaker) transitionTo(newState CircuitBreakerState) {
	oldState := cb.state
	cb.state = newState
	if cb.config.OnStateChange != nil && oldState != newState {
		cb.config.OnStateChange(oldState, newState)
	}
}
