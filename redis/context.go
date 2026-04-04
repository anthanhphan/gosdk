// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis

import "context"

// contextKey is an unexported type for context keys in this package
// to avoid key collisions with other packages.
type contextKey uint8

const (
	// actionKey is the context key that stores the Redis action/operation name.
	// The value is used by the metric hook to label latency and error metrics.
	actionKey contextKey = iota
)

// WithAction returns a new context that carries the given Redis action name.
// Callers should wrap their context with this before invoking any Redis command
// so that the metric hook can extract the action label for Prometheus metrics.
//
// Example:
//
//	ctx = redis.WithAction(ctx, "get_user_session")
//	client.Get(ctx, key)
func WithAction(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, actionKey, action)
}

// ActionFromContext extracts the Redis action name from the context.
// Returns an empty string if no action was set.
func ActionFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(actionKey).(string); ok {
		return v
	}
	return ""
}
