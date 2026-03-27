// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "github.com/anthanhphan/gosdk/orianna/shared/hooks"

// Type aliases for HTTP-specific hook types.
// HTTP uses int (status code) as the response code type.
type (
	Hooks             = hooks.Hooks[Context, int]
	OnRequestHook     = hooks.OnRequestHook[Context]
	OnResponseHook    = hooks.OnResponseHook[Context, int]
	OnErrorHook       = hooks.OnErrorHook[Context]
	OnPanicHook       = hooks.OnPanicHook[Context]
	OnShutdownHook    = hooks.OnShutdownHook
	OnServerStartHook = hooks.OnServerStartHook
)

// NewHooks creates a new HTTP Hooks instance.
func NewHooks() *Hooks {
	return hooks.New[Context, int]()
}
