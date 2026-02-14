// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"testing"
	"time"
)

// Hooks Constructor Tests

func TestNewHooks(t *testing.T) {
	hooks := NewHooks()
	if hooks == nil {
		t.Fatal("NewHooks() returned nil")
	}
}

// Fluent API Tests

func TestHooks_FluentAPI(t *testing.T) {
	hooks := NewHooks()

	// All Add methods should return the Hooks pointer for chaining
	result := hooks.
		AddOnRequest(func(_ Context) {}).
		AddOnResponse(func(_ Context, _ int, _ time.Duration) {}).
		AddOnError(func(_ Context, _ error) {}).
		AddOnPanic(func(_ Context, _ any, _ []byte) {}).
		AddOnShutdown(func() {}).
		AddOnServerStart(func(_ any) error { return nil })

	if result != hooks {
		t.Error("fluent API should return same *Hooks instance")
	}

	// Verify hooks were registered by checking Execute calls produce side effects
	var requestCalled, responseCalled, errorCalled, panicCalled, shutdownCalled, serverStartCalled bool

	hooks2 := NewHooks()
	hooks2.AddOnRequest(func(_ Context) { requestCalled = true })
	hooks2.AddOnResponse(func(_ Context, _ int, _ time.Duration) { responseCalled = true })
	hooks2.AddOnError(func(_ Context, _ error) { errorCalled = true })
	hooks2.AddOnPanic(func(_ Context, _ any, _ []byte) { panicCalled = true })
	hooks2.AddOnShutdown(func() { shutdownCalled = true })
	hooks2.AddOnServerStart(func(_ any) error { serverStartCalled = true; return nil })

	hooks2.ExecuteOnRequest(nil)
	hooks2.ExecuteOnResponse(nil, 200, 0)
	hooks2.ExecuteOnError(nil, nil)
	hooks2.ExecuteOnPanic(nil, nil, nil)
	hooks2.ExecuteOnShutdown()
	_ = hooks2.ExecuteOnServerStart(nil)

	if !requestCalled || !responseCalled || !errorCalled || !panicCalled || !shutdownCalled || !serverStartCalled {
		t.Error("Not all hooks were called after registration")
	}
}

// Execute Tests

func TestHooks_ExecuteOnRequest(t *testing.T) {
	t.Run("executes all hooks in order", func(t *testing.T) {
		var order []int
		hooks := NewHooks()
		hooks.AddOnRequest(func(_ Context) { order = append(order, 1) })
		hooks.AddOnRequest(func(_ Context) { order = append(order, 2) })
		hooks.AddOnRequest(func(_ Context) { order = append(order, 3) })

		hooks.ExecuteOnRequest(nil)

		if len(order) != 3 {
			t.Fatalf("expected 3 hooks to execute, got %d", len(order))
		}
		for i, v := range order {
			if v != i+1 {
				t.Errorf("order[%d] = %d, want %d", i, v, i+1)
			}
		}
	})

	t.Run("empty hooks no-ops", func(t *testing.T) {
		hooks := NewHooks()
		hooks.ExecuteOnRequest(nil) // should not panic
	})
}

func TestHooks_ExecuteOnResponse(t *testing.T) {
	var calledStatus int
	var calledLatency time.Duration
	hooks := NewHooks()
	hooks.AddOnResponse(func(_ Context, status int, latency time.Duration) {
		calledStatus = status
		calledLatency = latency
	})

	hooks.ExecuteOnResponse(nil, 200, 100*time.Millisecond)

	if calledStatus != 200 {
		t.Errorf("status = %d, want 200", calledStatus)
	}
	if calledLatency != 100*time.Millisecond {
		t.Errorf("latency = %v, want 100ms", calledLatency)
	}
}

func TestHooks_ExecuteOnError(t *testing.T) {
	var calledErr error
	hooks := NewHooks()
	hooks.AddOnError(func(_ Context, err error) {
		calledErr = err
	})

	expectedErr := errors.New("test error")
	hooks.ExecuteOnError(nil, expectedErr)

	if !errors.Is(calledErr, expectedErr) {
		t.Errorf("error = %v, want %v", calledErr, expectedErr)
	}
}

func TestHooks_ExecuteOnPanic(t *testing.T) {
	var calledRecovered any
	var calledStack []byte
	hooks := NewHooks()
	hooks.AddOnPanic(func(_ Context, recovered any, stack []byte) {
		calledRecovered = recovered
		calledStack = stack
	})

	hooks.ExecuteOnPanic(nil, "panic value", []byte("stack trace"))

	if calledRecovered != "panic value" {
		t.Errorf("recovered = %v, want 'panic value'", calledRecovered)
	}
	if string(calledStack) != "stack trace" {
		t.Errorf("stack = %s, want 'stack trace'", calledStack)
	}
}

func TestHooks_ExecuteOnShutdown(t *testing.T) {
	var called bool
	hooks := NewHooks()
	hooks.AddOnShutdown(func() { called = true })

	hooks.ExecuteOnShutdown()

	if !called {
		t.Error("shutdown hook was not called")
	}
}

func TestHooks_ExecuteOnServerStart(t *testing.T) {
	t.Run("all hooks succeed", func(t *testing.T) {
		hooks := NewHooks()
		hooks.AddOnServerStart(func(_ any) error { return nil })
		hooks.AddOnServerStart(func(_ any) error { return nil })

		err := hooks.ExecuteOnServerStart(nil)
		if err != nil {
			t.Errorf("ExecuteOnServerStart() error = %v, want nil", err)
		}
	})

	t.Run("error stops execution", func(t *testing.T) {
		var secondCalled bool
		expectedErr := errors.New("start error")
		hooks := NewHooks()
		hooks.AddOnServerStart(func(_ any) error { return expectedErr })
		hooks.AddOnServerStart(func(_ any) error {
			secondCalled = true
			return nil
		})

		err := hooks.ExecuteOnServerStart(nil)
		if !errors.Is(err, expectedErr) {
			t.Errorf("error = %v, want %v", err, expectedErr)
		}
		if secondCalled {
			t.Error("second hook should not execute after first error")
		}
	})

	t.Run("empty hooks no error", func(t *testing.T) {
		hooks := NewHooks()
		err := hooks.ExecuteOnServerStart(nil)
		if err != nil {
			t.Errorf("ExecuteOnServerStart() error = %v, want nil", err)
		}
	})
}

// Multiple Hooks Tests

func TestHooks_MultipleHooksExecuteAll(t *testing.T) {
	count := 0
	hooks := NewHooks()
	for i := 0; i < 5; i++ {
		hooks.AddOnShutdown(func() { count++ })
	}

	hooks.ExecuteOnShutdown()

	if count != 5 {
		t.Errorf("shutdown hook count = %d, want 5", count)
	}
}
