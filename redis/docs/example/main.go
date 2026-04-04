// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

// Package main demonstrates idiomatic usage of the redis package.
//
// Run against a local Redis instance:
//
//	docker run --rm -p 6379:6379 redis:7-alpine
//	go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/redis"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	log := logger.NewLogger(&logger.DevelopmentConfig, []io.Writer{os.Stdout})

	// Create client. Both arguments are required.
	// NewClient validates config, applies defaults, and pings before returning.
	client, err := redis.NewClient(&redis.Config{
		Addr:            "localhost:6379",
		MetricNamespace: "example",
	}, log)
	if err != nil {
		log.Errorw("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Use a real application service to demonstrate the recommended pattern.
	svc := NewSessionService(client, log)

	ctx := context.Background()

	// Store a session.
	if err := svc.Set(ctx, "u-1", "tok-abc", time.Hour); err != nil {
		log.Errorw("set session failed", "error", err)
	}

	// Retrieve it.
	tok, err := svc.Get(ctx, "u-1")
	if err != nil {
		log.Errorw("get session failed", "error", err)
	} else {
		fmt.Println("session:", tok)
	}

	// Handle a cache miss gracefully.
	tok, err = svc.Get(ctx, "u-nonexistent")
	if err != nil {
		log.Errorw("get session failed", "error", err)
	} else if tok == "" {
		fmt.Println("cache miss — no session found")
	}

	// Delete a session.
	if err := svc.Delete(ctx, "u-1"); err != nil {
		log.Errorw("delete session failed", "error", err)
	}

	// Write multiple keys in a single round-trip.
	if err := svc.Warmup(ctx, map[string]string{
		"u-2": "tok-xyz",
		"u-3": "tok-qrs",
	}, time.Hour); err != nil {
		log.Errorw("warmup failed", "error", err)
	}

	// Health check before serving traffic.
	if err := client.Ping(ctx).Err(); err != nil {
		log.Errorw("redis unhealthy", "error", err)
	}
}

// ============================================================================
// SessionService – canonical service pattern
// ============================================================================

// SessionService manages user sessions in Redis.
//
// Design:
//   - Inject *redis.Client + *logger.Logger via constructor.
//   - Create a ScopedClient once with c.Scope("session_svc").
//   - Each method calls scoped.Ctx(ctx, "op") to tag the context.
//   - Prometheus metrics are automatically labeled: action = "session_svc.<op>".
type SessionService struct {
	redis  *redis.Client
	scoped *redis.ScopedClient
	log    *logger.Logger
}

func NewSessionService(c *redis.Client, log *logger.Logger) *SessionService {
	return &SessionService{
		redis:  c,
		scoped: c.Scope("session_svc"),
		log:    log,
	}
}

// Get returns the token for uid, or "" on a cache miss.
func (s *SessionService) Get(ctx context.Context, uid string) (string, error) {
	val, err := s.redis.Get(s.scoped.Ctx(ctx, "get"), "session:"+uid).Result()
	if errors.Is(err, goredis.Nil) {
		return "", nil // cache miss is not an error
	}
	return val, err
}

// Set stores token for uid with the given TTL.
func (s *SessionService) Set(ctx context.Context, uid, token string, ttl time.Duration) error {
	return s.redis.Set(s.scoped.Ctx(ctx, "set"), "session:"+uid, token, ttl).Err()
}

// Delete removes the session for uid.
func (s *SessionService) Delete(ctx context.Context, uid string) error {
	return s.redis.Del(s.scoped.Ctx(ctx, "delete"), "session:"+uid).Err()
}

// Warmup writes multiple sessions in a single pipeline round-trip.
func (s *SessionService) Warmup(ctx context.Context, sessions map[string]string, ttl time.Duration) error {
	pipe := s.redis.Pipeline()
	for uid, token := range sessions {
		pipe.Set(s.scoped.Ctx(ctx, "warmup"), "session:"+uid, token, ttl)
	}
	_, err := pipe.Exec(s.scoped.Ctx(ctx, "warmup"))
	return err
}
