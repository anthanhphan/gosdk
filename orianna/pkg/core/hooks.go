// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"sync"
	"time"
)

// Lifecycle Hooks

// Hook types for request/response lifecycle
type (
	// OnRequestHook is called at the start of each request
	OnRequestHook func(ctx Context)

	// OnResponseHook is called after each response is sent
	OnResponseHook func(ctx Context, statusCode int, latency time.Duration)

	// OnErrorHook is called when an error occurs during request handling
	OnErrorHook func(ctx Context, err error)

	// OnPanicHook is called when a panic is recovered during request handling
	OnPanicHook func(ctx Context, recovered any, stack []byte)

	// OnShutdownHook is called when the server is shutting down
	OnShutdownHook func()

	// OnServerStartHook is called when the server starts
	OnServerStartHook func(server any) error
)

// Hooks contains the lifecycle hooks for the server.
// It is safe for concurrent use. Use AddOn* methods to register hooks
// and ExecuteOn* methods to trigger them.
type Hooks struct {
	mu            sync.RWMutex
	onRequest     []OnRequestHook
	onResponse    []OnResponseHook
	onError       []OnErrorHook
	onPanic       []OnPanicHook
	onShutdown    []OnShutdownHook
	onServerStart []OnServerStartHook
}

// NewHooks creates a new Hooks instance
func NewHooks() *Hooks {
	return &Hooks{}
}

// Fluent Builder Methods

// AddOnRequest adds a request hook
func (h *Hooks) AddOnRequest(hook OnRequestHook) *Hooks {
	h.mu.Lock()
	h.onRequest = append(h.onRequest, hook)
	h.mu.Unlock()
	return h
}

// AddOnResponse adds a response hook
func (h *Hooks) AddOnResponse(hook OnResponseHook) *Hooks {
	h.mu.Lock()
	h.onResponse = append(h.onResponse, hook)
	h.mu.Unlock()
	return h
}

// AddOnError adds an error hook
func (h *Hooks) AddOnError(hook OnErrorHook) *Hooks {
	h.mu.Lock()
	h.onError = append(h.onError, hook)
	h.mu.Unlock()
	return h
}

// AddOnPanic adds a panic hook
func (h *Hooks) AddOnPanic(hook OnPanicHook) *Hooks {
	h.mu.Lock()
	h.onPanic = append(h.onPanic, hook)
	h.mu.Unlock()
	return h
}

// AddOnShutdown adds a shutdown hook
func (h *Hooks) AddOnShutdown(hook OnShutdownHook) *Hooks {
	h.mu.Lock()
	h.onShutdown = append(h.onShutdown, hook)
	h.mu.Unlock()
	return h
}

// AddOnServerStart adds a server start hook
func (h *Hooks) AddOnServerStart(hook OnServerStartHook) *Hooks {
	h.mu.Lock()
	h.onServerStart = append(h.onServerStart, hook)
	h.mu.Unlock()
	return h
}

// Execute Methods

// ExecuteOnRequest executes all request hooks
func (h *Hooks) ExecuteOnRequest(ctx Context) {
	h.mu.RLock()
	hooks := h.onRequest
	h.mu.RUnlock()
	for _, hook := range hooks {
		hook(ctx)
	}
}

// ExecuteOnResponse executes all response hooks
func (h *Hooks) ExecuteOnResponse(ctx Context, statusCode int, latency time.Duration) {
	h.mu.RLock()
	hooks := h.onResponse
	h.mu.RUnlock()
	for _, hook := range hooks {
		hook(ctx, statusCode, latency)
	}
}

// ExecuteOnError executes all error hooks
func (h *Hooks) ExecuteOnError(ctx Context, err error) {
	h.mu.RLock()
	hooks := h.onError
	h.mu.RUnlock()
	for _, hook := range hooks {
		hook(ctx, err)
	}
}

// ExecuteOnPanic executes all panic hooks
func (h *Hooks) ExecuteOnPanic(ctx Context, recovered any, stack []byte) {
	h.mu.RLock()
	hooks := h.onPanic
	h.mu.RUnlock()
	for _, hook := range hooks {
		hook(ctx, recovered, stack)
	}
}

// ExecuteOnShutdown executes all shutdown hooks
func (h *Hooks) ExecuteOnShutdown() {
	h.mu.RLock()
	hooks := h.onShutdown
	h.mu.RUnlock()
	for _, hook := range hooks {
		hook()
	}
}

// ExecuteOnServerStart executes all server start hooks
// Returns the first error encountered, stopping execution of remaining hooks
func (h *Hooks) ExecuteOnServerStart(server any) error {
	h.mu.RLock()
	hooks := h.onServerStart
	h.mu.RUnlock()
	for _, hook := range hooks {
		if err := hook(server); err != nil {
			return err
		}
	}
	return nil
}
