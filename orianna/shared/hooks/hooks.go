// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package hooks provides generic lifecycle hooks for server implementations.
// Both HTTP and gRPC servers use this package to avoid duplicating hook logic.
package hooks

import (
	"context"
	"fmt"
	"time"

	routine "github.com/anthanhphan/gosdk/goroutine"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/utils"
)

// Hook function types parameterized by context type C and response code type R.
type (
	// OnRequestHook is called at the start of each request.
	OnRequestHook[C any] func(ctx C)

	// OnResponseHook is called after each response is sent.
	// R is int for HTTP (status code) or string for gRPC (status code string).
	OnResponseHook[C any, R any] func(ctx C, code R, latency time.Duration)

	// OnErrorHook is called when an error occurs during request handling.
	OnErrorHook[C any] func(ctx C, err error)

	// OnPanicHook is called when a panic is recovered during request handling.
	OnPanicHook[C any] func(ctx C, recovered any, stack []byte)

	// OnShutdownHook is called when the server is shutting down.
	OnShutdownHook func()

	// OnServerStartHook is called when the server starts.
	OnServerStartHook func(server any) error
)

// Hooks contains the lifecycle hooks for a server.
// C is the context type, R is the response code type.
// Not safe for concurrent use after server start. Register all hooks before calling Start().
type Hooks[C any, R any] struct {
	onRequest     []OnRequestHook[C]
	onResponse    []OnResponseHook[C, R]
	onError       []OnErrorHook[C]
	onPanic       []OnPanicHook[C]
	onShutdown    []OnShutdownHook
	onServerStart []OnServerStartHook
}

// New creates a new Hooks instance.
func New[C any, R any]() *Hooks[C, R] {
	return &Hooks[C, R]{}
}

// --- Fluent Builder Methods ---

// AddOnRequest adds a request hook.
func (h *Hooks[C, R]) AddOnRequest(hook OnRequestHook[C]) *Hooks[C, R] {
	h.onRequest = append(h.onRequest, hook)
	return h
}

// AddOnResponse adds a response hook.
func (h *Hooks[C, R]) AddOnResponse(hook OnResponseHook[C, R]) *Hooks[C, R] {
	h.onResponse = append(h.onResponse, hook)
	return h
}

// AddOnError adds an error hook.
func (h *Hooks[C, R]) AddOnError(hook OnErrorHook[C]) *Hooks[C, R] {
	h.onError = append(h.onError, hook)
	return h
}

// AddOnPanic adds a panic hook.
func (h *Hooks[C, R]) AddOnPanic(hook OnPanicHook[C]) *Hooks[C, R] {
	h.onPanic = append(h.onPanic, hook)
	return h
}

// AddOnShutdown adds a shutdown hook.
func (h *Hooks[C, R]) AddOnShutdown(hook OnShutdownHook) *Hooks[C, R] {
	h.onShutdown = append(h.onShutdown, hook)
	return h
}

// AddOnServerStart adds a server start hook.
func (h *Hooks[C, R]) AddOnServerStart(hook OnServerStartHook) *Hooks[C, R] {
	h.onServerStart = append(h.onServerStart, hook)
	return h
}

// --- Execute Methods ---

// ExecuteOnRequest executes all request hooks.
// Uses a single deferred recover for the entire loop instead of per-hook
// to eliminate N closure allocations + N deferred recovers per request.
func (h *Hooks[C, R]) ExecuteOnRequest(ctx C) {
	defer recoverHookPanic("OnRequest")
	for _, hook := range h.onRequest {
		hook(ctx)
	}
}

// ExecuteOnResponse executes all response hooks.
func (h *Hooks[C, R]) ExecuteOnResponse(ctx C, code R, latency time.Duration) {
	defer recoverHookPanic("OnResponse")
	for _, hook := range h.onResponse {
		hook(ctx, code, latency)
	}
}

// ExecuteOnError executes all error hooks.
func (h *Hooks[C, R]) ExecuteOnError(ctx C, err error) {
	defer recoverHookPanic("OnError")
	for _, hook := range h.onError {
		hook(ctx, err)
	}
}

// ExecuteOnPanic executes all panic hooks.
func (h *Hooks[C, R]) ExecuteOnPanic(ctx C, recovered any, stack []byte) {
	defer recoverHookPanic("OnPanic")
	for _, hook := range h.onPanic {
		hook(ctx, recovered, stack)
	}
}

// ExecuteOnShutdown executes all shutdown hooks in parallel using goroutine.Group.
// Each hook runs in its own goroutine with built-in panic recovery from the routine package.
func (h *Hooks[C, R]) ExecuteOnShutdown() {
	if len(h.onShutdown) == 0 {
		return
	}
	g := routine.NewGroup()
	for _, hook := range h.onShutdown {
		hk := hook
		g.Go(func(_ context.Context) error {
			hk()
			return nil
		})
	}
	_ = g.Wait()
}

// ExecuteOnServerStart executes all server start hooks.
// Returns the first error encountered, stopping execution of remaining hooks.
// Each hook is protected by a recover — a panicking hook will not crash the process.
func (h *Hooks[C, R]) ExecuteOnServerStart(server any) error {
	for _, hook := range h.onServerStart {
		if err := safeExecuteOnServerStartHook(hook, server); err != nil {
			return err
		}
	}
	return nil
}

// recoverHookPanic recovers from a panic in a hook execution loop.
// Replaces per-hook safeExecuteHook to avoid N closure + defer allocations.
func recoverHookPanic(hookName string) {
	if r := recover(); r != nil {
		location, _ := utils.GetPanicLocation()
		logger.Errorw("hook panicked",
			"hook", hookName,
			"error", fmt.Errorf("panic: %v", r),
			"location", location,
		)
	}
}

// safeExecuteOnServerStartHook runs a server start hook inside a deferred recover.
// If the hook panics, the panic is logged and returned as an error.
func safeExecuteOnServerStartHook(hook OnServerStartHook, server any) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			location, _ := utils.GetPanicLocation()
			logger.Errorw("OnServerStart hook panicked",
				"error", fmt.Errorf("panic: %v", r),
				"location", location,
			)
			retErr = fmt.Errorf("OnServerStart hook panicked: %v", r)
		}
	}()
	return hook(server)
}
