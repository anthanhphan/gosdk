// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

// Package redis provides a Redis client wrapper built on github.com/redis/go-redis/v9.
//
// It supports both standalone and Sentinel setups, and instruments every command with:
//   - Latency histogram: <namespace>_command_duration_seconds{action, command}
//   - Error counter:     <namespace>_command_errors_total{action, command, error_type}
//
// The action label is supplied via context:
//
//	ctx = redis.WithAction(ctx, "get_user_session")
//	val, err := client.Get(ctx, key).Result()
//
// # Usage
//
//	cfg, err := conflux.Load[redis.Config]("config/redis.yaml")
//	client, err := redis.NewClient(cfg, log)
package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/anthanhphan/gosdk/logger"
	goredis "github.com/redis/go-redis/v9"
)

// Client wraps go-redis UniversalClient with metrics and structured logging.
type Client struct {
	goredis.UniversalClient
}

// NewClient creates a Redis client from cfg.
//
// Connection settings (address, pool, timeouts, Sentinel) are read from cfg.
// log is required for structured error logging.
//
// Example:
//
//	cfg, _ := conflux.Load[redis.Config]("config.yaml")
//	client, err := redis.NewClient(cfg, log)
func NewClient(cfg *Config, log *logger.Logger) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("redis: config must not be nil")
	}
	if log == nil {
		return nil, errors.New("redis: logger must not be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	cfg.applyDefaults()

	rdb := buildUniversalClient(cfg)

	rdb.AddHook(newMetricsHook(log, cfg.MetricNamespace, cfg.MetricSubsystem))

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Errorw("redis: ping failed",
			"error", err.Error(),
			"addr", cfg.Addr,
			"master_name", cfg.MasterName,
		)
		return nil, fmt.Errorf("redis: ping failed: %w", err)
	}

	mode := "standalone"
	if cfg.MasterName != "" {
		mode = "sentinel"
	}
	log.Infow("redis: connected",
		"mode", mode,
		"addr", cfg.Addr,
		"master_name", cfg.MasterName,
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)

	return &Client{UniversalClient: rdb}, nil
}

func buildUniversalClient(cfg *Config) goredis.UniversalClient {
	if cfg.MasterName != "" {
		return goredis.NewFailoverClient(&goredis.FailoverOptions{
			MasterName:       cfg.MasterName,
			SentinelAddrs:    cfg.SentinelAddrs,
			SentinelPassword: cfg.SentinelPassword,
			Password:         cfg.Password,
			DB:               cfg.DB,
			PoolSize:         cfg.PoolSize,
			MinIdleConns:     cfg.MinIdleConns,
			DialTimeout:      cfg.DialTimeout.Duration,
			ReadTimeout:      cfg.ReadTimeout.Duration,
			WriteTimeout:     cfg.WriteTimeout.Duration,
		})
	}

	return goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout.Duration,
		ReadTimeout:  cfg.ReadTimeout.Duration,
		WriteTimeout: cfg.WriteTimeout.Duration,
	})
}
