// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/anthanhphan/gosdk/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScope_Ctx_SetsAction(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	scoped := client.Scope("user_svc")
	ctx := scoped.Ctx(context.Background(), "get_session")

	assert.Equal(t, "user_svc.get_session", redis.ActionFromContext(ctx))
}

func TestScope_Nested_BuildsLabel(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	ctx := client.Scope("myapp").Scope("cache").Ctx(context.Background(), "get")

	assert.Equal(t, "myapp.cache.get", redis.ActionFromContext(ctx))
}

func TestScope_ExecutesRealCommands(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	scoped := client.Scope("user_svc")

	err = client.Set(scoped.Ctx(context.Background(), "set_session"), "session:u1", "tok", 0).Err()
	require.NoError(t, err)

	val, err := client.Get(scoped.Ctx(context.Background(), "get_session"), "session:u1").Result()
	require.NoError(t, err)
	assert.Equal(t, "tok", val)
}

func TestScope_SharesUnderlyingClient(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	a := client.Scope("svc_a")
	b := client.Scope("svc_b")

	_ = client.Set(a.Ctx(context.Background(), "set"), "shared_key", "v", 0).Err()

	val, err := client.Get(b.Ctx(context.Background(), "get"), "shared_key").Result()
	require.NoError(t, err)
	assert.Equal(t, "v", val)
}
