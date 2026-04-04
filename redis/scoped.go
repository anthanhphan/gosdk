// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis

import "context"

// ScopedClient wraps *Client with a fixed scope prefix.
// Every call to Ctx prepends the scope to the operation name,
// so callers never manually call WithAction – they only supply
// the short operation name (e.g. "get_session").
//
// ScopedClients can be nested to build hierarchical labels:
//
//	app   := client.Scope("myapp")       // "myapp"
//	cache := app.Scope("user_cache")     // "myapp.user_cache"
//	ctx   = cache.Ctx(ctx, "get")        // action = "myapp.user_cache.get"
type ScopedClient struct {
	*Client
	scope string
}

// Scope creates a ScopedClient that automatically prepends scope to every
// action label. Use Ctx(ctx, "op") to obtain a context ready for a command.
func (c *Client) Scope(scope string) *ScopedClient {
	return &ScopedClient{Client: c, scope: scope}
}

// Scope creates a nested ScopedClient whose action is "<parent>.<child>".
func (s *ScopedClient) Scope(child string) *ScopedClient {
	return &ScopedClient{Client: s.Client, scope: s.scope + "." + child}
}

// Ctx returns a context with action set to "<scope>.<op>".
// Pass the returned context directly to any Redis command.
//
//	val, err := s.redis.Get(s.redis.Ctx(ctx, "get_session"), key).Result()
func (s *ScopedClient) Ctx(ctx context.Context, op string) context.Context {
	return WithAction(ctx, s.scope+"."+op)
}
