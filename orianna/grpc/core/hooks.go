// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "github.com/anthanhphan/gosdk/orianna/shared/hooks"

// Type aliases for gRPC-specific hook types.
// gRPC uses string (status code name) as the response code type.
type (
	Hooks             = hooks.Hooks[Context, string]
	OnRequestHook     = hooks.OnRequestHook[Context]
	OnResponseHook    = hooks.OnResponseHook[Context, string]
	OnErrorHook       = hooks.OnErrorHook[Context]
	OnPanicHook       = hooks.OnPanicHook[Context]
	OnShutdownHook    = hooks.OnShutdownHook
	OnServerStartHook = hooks.OnServerStartHook
)

// NewHooks creates a new gRPC Hooks instance.
func NewHooks() *Hooks {
	return hooks.New[Context, string]()
}
