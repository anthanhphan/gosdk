// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package health provides shared health check types and a concurrent health manager
// used across all protocol implementations (REST, gRPC, WebSocket, etc.).
package health

import (
	"context"
	"sync"
	"time"

	routine "github.com/anthanhphan/gosdk/goroutine"
)

// HealthStatus represents the health status of a component.
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents the result of a single health check.
type HealthCheck struct {
	Name         string         `json:"name"`
	Status       HealthStatus   `json:"status"`
	Message      string         `json:"message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	ResponseTime int64          `json:"response_time_ms,omitempty"`
	Error        error          `json:"-"`
}

// HealthReport represents the overall health status of the system.
type HealthReport struct {
	Status         HealthStatus           `json:"status"`
	Timestamp      int64                  `json:"timestamp"`
	TotalLatencyMs int64                  `json:"total_latency_ms"`
	Checks         map[string]HealthCheck `json:"checks"`
	Version        string                 `json:"version,omitempty"`
	ServiceName    string                 `json:"service_name,omitempty"`
}

// Checker defines the interface for health check implementations.
type Checker interface {
	Check(ctx context.Context) HealthCheck
	Name() string
}

// Manager

const (
	// DefaultWorkerPoolSize is the default number of concurrent health check workers.
	DefaultWorkerPoolSize = 10
)

// Manager handles health checks with a worker pool to limit concurrency.
type Manager struct {
	mu             sync.RWMutex
	checkers       []Checker
	workerPoolSize int
}

// NewManager creates a new health check manager with default worker pool size.
func NewManager() *Manager {
	return &Manager{
		checkers:       nil,
		workerPoolSize: DefaultWorkerPoolSize,
	}
}

// NewManagerWithPoolSize creates a new health check manager with custom worker pool size.
func NewManagerWithPoolSize(poolSize int) *Manager {
	if poolSize <= 0 {
		poolSize = DefaultWorkerPoolSize
	}
	return &Manager{
		checkers:       nil,
		workerPoolSize: poolSize,
	}
}

// Register adds a health checker to the manager.
func (m *Manager) Register(checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers = append(m.checkers, checker)
}

// WorkerPoolSize returns the current worker pool size.
func (m *Manager) WorkerPoolSize() int {
	return m.workerPoolSize
}

// CheckerCount returns the number of registered health checkers.
func (m *Manager) CheckerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.checkers)
}

// Check performs all health checks concurrently using routine.FanOut
// and returns an aggregated report.
func (m *Manager) Check(ctx context.Context) *HealthReport {
	checkStart := time.Now()

	m.mu.RLock()
	if len(m.checkers) == 0 {
		m.mu.RUnlock()
		return &HealthReport{
			Status:         StatusHealthy,
			Timestamp:      time.Now().UnixMilli(),
			TotalLatencyMs: time.Since(checkStart).Milliseconds(),
			Checks:         make(map[string]HealthCheck),
		}
	}
	checkersCopy := make([]Checker, len(m.checkers))
	copy(checkersCopy, m.checkers)
	m.mu.RUnlock()

	type checkResult struct {
		name  string
		check HealthCheck
	}

	// Use routine.FanOut for concurrent execution with panic recovery.
	// FanOut preserves order: results[i] corresponds to checkers[i].
	// Panicked items produce zero-value results — we detect and mark them unhealthy.
	// Context cancellation may cause FanOut to return fewer results than checkers.
	results, fanoutErr := routine.FanOut(ctx, checkersCopy, m.workerPoolSize,
		func(ctx context.Context, c Checker) (checkResult, error) {
			start := time.Now()
			check := c.Check(ctx)
			check.ResponseTime = time.Since(start).Milliseconds()
			return checkResult{name: c.Name(), check: check}, nil
		},
	)

	checks := make(map[string]HealthCheck, len(checkersCopy))
	overallStatus := StatusHealthy

	// If context was cancelled and no results, mark all checkers as unhealthy
	if fanoutErr != nil && len(results) == 0 {
		for _, c := range checkersCopy {
			checks[c.Name()] = HealthCheck{
				Name:    c.Name(),
				Status:  StatusUnhealthy,
				Message: "health check cancelled",
			}
		}
		return &HealthReport{
			Status:    StatusUnhealthy,
			Timestamp: time.Now().UnixMilli(),
			Checks:    checks,
		}
	}

	for i, result := range results {
		// Zero-value name means the checker panicked (FanOut recovered it)
		if result.name == "" {
			name := checkersCopy[i].Name()
			checks[name] = HealthCheck{
				Name:    name,
				Status:  StatusUnhealthy,
				Message: "health check panicked",
			}
			overallStatus = StatusUnhealthy
			continue
		}
		checks[result.name] = result.check
		if result.check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if result.check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return &HealthReport{
		Status:         overallStatus,
		Timestamp:      time.Now().UnixMilli(),
		TotalLatencyMs: time.Since(checkStart).Milliseconds(),
		Checks:         checks,
	}
}

// CustomChecker wraps a custom health check function as a Checker.
type CustomChecker struct {
	name    string
	checkFn func(ctx context.Context) HealthCheck
}

// NewCustomChecker creates a new custom health checker.
func NewCustomChecker(name string, checkFn func(ctx context.Context) HealthCheck) *CustomChecker {
	return &CustomChecker{name: name, checkFn: checkFn}
}

// Check performs the custom health check.
func (c *CustomChecker) Check(ctx context.Context) HealthCheck {
	check := c.checkFn(ctx)
	if check.Name == "" {
		check.Name = c.name
	}
	return check
}

// Name returns the name of this health checker.
func (c *CustomChecker) Name() string { return c.name }
