// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package ctxkeys provides shared typed context keys used across
// both HTTP and gRPC server implementations.
package ctxkeys

// Key is a typed context key to prevent collisions with other packages.
type Key struct{ name string }

func (k Key) String() string { return "orianna." + k.name }
func (k Key) Key() string    { return k.name }

// Shared context keys used by both HTTP middleware and gRPC interceptors.
var (
	RequestID     = Key{"request_id"}
	TraceID       = Key{"trace_id"}
	UserID        = Key{"user_id"}
	TenantID      = Key{"tenant_id"}
	CorrelationID = Key{"correlation_id"}
)
