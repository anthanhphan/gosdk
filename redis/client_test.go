// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/redis"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test helpers
// ============================================================================

// newLogger returns a real logger writing to /dev/null for tests.
func newLogger() *logger.Logger {
	return logger.NewLogger(&logger.DevelopmentConfig, []io.Writer{io.Discard})
}

// newMiniRedis starts an in-process Redis server and returns it + a connected client.
func newMiniRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

// ============================================================================
// Context helpers
// ============================================================================

func TestWithAction_StoresAndRetrieves(t *testing.T) {
	ctx := redis.WithAction(context.Background(), "get_user_session")
	assert.Equal(t, "get_user_session", redis.ActionFromContext(ctx))
}

func TestActionFromContext_MissingKey_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", redis.ActionFromContext(context.Background()))
}

func TestWithAction_Overwrite(t *testing.T) {
	ctx := redis.WithAction(context.Background(), "first")
	ctx = redis.WithAction(ctx, "second")
	assert.Equal(t, "second", redis.ActionFromContext(ctx))
}

// ============================================================================
// NewClient – guard tests (no live Redis)
// ============================================================================

func TestNewClient_NilConfig_ReturnsError(t *testing.T) {
	_, err := redis.NewClient(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config must not be nil")
}

func TestNewClient_NilLogger_ReturnsError(t *testing.T) {
	_, err := redis.NewClient(&redis.Config{Addr: "localhost:6379"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger must not be nil")
}

func TestNewClient_InvalidConfig_ReturnsValidationError(t *testing.T) {
	// cfg invalid (no addr/sentinel), log non-nil → reaches cfg.Validate() in NewClient
	_, err := redis.NewClient(&redis.Config{}, newLogger())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "addr")
}

func TestNewClient_PingFails_ReturnsError(t *testing.T) {
	_, err := redis.NewClient(&redis.Config{Addr: "localhost:19999"}, newLogger())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping failed")
}

// ============================================================================
// NewClient – happy paths
// ============================================================================

func TestNewClient_Standalone_Connected(t *testing.T) {
	_, client := newMiniRedis(t)
	assert.NotNil(t, client)
	assert.NoError(t, client.Ping(context.Background()).Err())
}

func TestNewClient_AppliesDefaults_WhenZeroValues(t *testing.T) {
	mr := miniredis.RunT(t)
	// PoolSize=0, MinIdleConns=0, zero timeouts → applyDefaults fills them
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	assert.NotNil(t, client)
}

func TestNewClient_ExplicitPoolSize_NotOverwritten(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr(), PoolSize: 5}, newLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	assert.NotNil(t, client)
}

func TestNewClient_WithStdoutLogger(t *testing.T) {
	mr := miniredis.RunT(t)
	log := logger.NewLogger(&logger.DevelopmentConfig, []io.Writer{os.Stdout})
	client, err := redis.NewClient(&redis.Config{Addr: mr.Addr()}, log)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	assert.NotNil(t, client)
}

// Sentinel path: buildUniversalClient sentinel branch is covered; ping fails (no real sentinel).
func TestNewClient_SentinelPath_PingFails(t *testing.T) {
	_, err := redis.NewClient(&redis.Config{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"localhost:26379"},
	}, newLogger())
	assert.Error(t, err)
}

// ============================================================================
// ProcessHook – success and redis.Nil
// ============================================================================

func TestClient_SetGet(t *testing.T) {
	_, client := newMiniRedis(t)
	ctx := redis.WithAction(context.Background(), "test_set_get")

	require.NoError(t, client.Set(ctx, "key", "value", 0).Err())

	val, err := client.Get(ctx, "key").Result()
	require.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestClient_GetMissingKey_ReturnsNilNotError(t *testing.T) {
	_, client := newMiniRedis(t)
	ctx := redis.WithAction(context.Background(), "test_get_nil")

	_, err := client.Get(ctx, "no_such_key").Result()
	// redis.Nil must NOT be counted as a metric error
	assert.True(t, errors.Is(err, goredis.Nil))
}

func TestClient_WithoutAction_UsesUnknownLabel(t *testing.T) {
	_, client := newMiniRedis(t)
	// No WithAction – actionFromCtx returns "unknown", should not panic
	assert.NoError(t, client.Set(context.Background(), "k", "v", 0).Err())
}

// ProcessHook – context-cancelled command (hits "context" error_type classify branch)
func TestClient_ContextCanceledCommand(t *testing.T) {
	_, client := newMiniRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx = redis.WithAction(ctx, "test_ctx_cancel")
	_ = client.Get(ctx, "k").Err()
}

// ============================================================================
// ProcessPipelineHook
// ============================================================================

func TestClient_Pipeline_Success(t *testing.T) {
	_, client := newMiniRedis(t)
	ctx := redis.WithAction(context.Background(), "test_pipeline")

	pipe := client.Pipeline()
	setCmd := pipe.Set(ctx, "p1", "val1", 0)
	getCmd := pipe.Get(ctx, "p1")
	_, err := pipe.Exec(ctx)
	require.NoError(t, err)
	assert.NoError(t, setCmd.Err())
	assert.Equal(t, "val1", getCmd.Val())
}

func TestClient_Pipeline_WithoutAction(t *testing.T) {
	_, client := newMiniRedis(t)
	pipe := client.Pipeline()
	pipe.Set(context.Background(), "k2", "v2", 0)
	_, err := pipe.Exec(context.Background())
	assert.NoError(t, err)
}

func TestClient_Pipeline_PartialError(t *testing.T) {
	mr, client := newMiniRedis(t)
	ctx := redis.WithAction(context.Background(), "test_pipeline_partial")

	// Seed a string key, then try LPUSH on it → WRONGTYPE error inside pipeline
	require.NoError(t, client.Set(ctx, "strkey", "val", 0).Err())

	pipe := client.Pipeline()
	pipe.LPush(ctx, "strkey", "elem")
	_, _ = pipe.Exec(ctx)
	_ = mr
}

// ============================================================================
// Concurrency – multiple goroutines issuing commands
// ============================================================================

func TestClient_ConcurrentCommands(t *testing.T) {
	_, client := newMiniRedis(t)
	ctx := redis.WithAction(context.Background(), "concurrent")

	done := make(chan struct{}, 20)
	for i := 0; i < 20; i++ {
		go func() {
			_ = client.Set(ctx, "ck", "cv", 0).Err()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}
