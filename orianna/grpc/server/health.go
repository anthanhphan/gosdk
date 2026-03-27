// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/anthanhphan/gosdk/orianna/grpc/configuration"
	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
)

// GRPCChecker checks gRPC endpoint availability.
// It lazily creates a persistent gRPC connection on the first Check() call
// and reuses it for subsequent checks, avoiding per-check connection overhead.
type GRPCChecker struct {
	target  string
	name    string
	timeout time.Duration

	once    sync.Once
	conn    *grpc.ClientConn
	connErr error
}

// NewGRPCChecker creates a new gRPC endpoint health checker.
func NewGRPCChecker(target string, name string, timeout time.Duration) *GRPCChecker {
	if name == "" {
		name = configuration.DefaultHealthCheckerName
	}
	if timeout == 0 {
		timeout = configuration.DefaultHealthCheckTimeout
	}
	return &GRPCChecker{
		target:  target,
		name:    name,
		timeout: timeout,
	}
}

// Close releases the underlying gRPC connection.
// Safe to call multiple times; no-op if connection was never established.
func (g *GRPCChecker) Close() error {
	// Ensure initConn's sync.Once has settled so g.conn is safe to read.
	g.once.Do(func() {})
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

// initConn lazily creates the gRPC connection once.
func (g *GRPCChecker) initConn() {
	g.once.Do(func() {
		g.conn, g.connErr = grpc.NewClient(g.target,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	})
}

// Check performs the gRPC health check using a persistent connection.
func (g *GRPCChecker) Check(ctx context.Context) health.HealthCheck {
	start := time.Now()

	g.initConn()
	if g.connErr != nil {
		return health.HealthCheck{
			Name:         g.name,
			Status:       health.StatusUnhealthy,
			Message:      fmt.Sprintf(core.HealthMessageCreateConnFailed, g.connErr),
			ResponseTime: time.Since(start).Milliseconds(),
			Error:        g.connErr,
			Details:      map[string]any{"target": g.target, "error": g.connErr.Error()},
		}
	}

	dialCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	beforeState := g.conn.GetState()
	g.conn.Connect()
	g.conn.WaitForStateChange(dialCtx, beforeState)

	if dialCtx.Err() != nil {
		return health.HealthCheck{
			Name:         g.name,
			Status:       health.StatusUnhealthy,
			Message:      fmt.Sprintf(core.HealthMessageUnhealthy, dialCtx.Err()),
			ResponseTime: time.Since(start).Milliseconds(),
			Error:        dialCtx.Err(),
			Details:      map[string]any{"target": g.target},
		}
	}

	currentState := g.conn.GetState()
	if currentState == connectivity.Ready {
		return health.HealthCheck{
			Name:         g.name,
			Status:       health.StatusHealthy,
			Message:      core.HealthMessageHealthy,
			ResponseTime: time.Since(start).Milliseconds(),
			Details:      map[string]any{"target": g.target},
		}
	}

	return health.HealthCheck{
		Name:         g.name,
		Status:       health.StatusUnhealthy,
		Message:      fmt.Sprintf(core.HealthMessageUnhealthy, currentState),
		ResponseTime: time.Since(start).Milliseconds(),
		Details:      map[string]any{"target": g.target, "state": currentState.String()},
	}
}

// Name returns the name of this health checker.
func (g *GRPCChecker) Name() string { return g.name }
