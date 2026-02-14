// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

//go:generate mockgen -source=health.go -destination=mocks/mock_health.go -package=mocks

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/orianna/pkg/configuration"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
)

// Health Types

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents the result of a single health check
type HealthCheck struct {
	Name         string         `json:"name"`
	Status       HealthStatus   `json:"status"`
	Message      string         `json:"message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	ResponseTime int64          `json:"response_time_ms,omitempty"`
	Error        error          `json:"-"`
}

// HealthReport represents the overall health status of the system
type HealthReport struct {
	Status      HealthStatus           `json:"status"`
	Timestamp   int64                  `json:"timestamp"`
	Checks      map[string]HealthCheck `json:"checks"`
	Version     string                 `json:"version,omitempty"`
	ServiceName string                 `json:"service_name,omitempty"`
}

// HealthChecker defines the interface for health check implementations
type HealthChecker interface {
	Check(ctx context.Context) HealthCheck
	Name() string
}

// Health Manager

const (
	// DefaultHealthWorkerPoolSize is the default number of concurrent health check workers
	DefaultHealthWorkerPoolSize = 10
)

// HealthManager handles health checks with a worker pool to limit concurrency
type HealthManager struct {
	checkers       []HealthChecker
	mu             sync.RWMutex
	workerPoolSize int
}

// NewHealthManager creates a new health check manager with default worker pool size
func NewHealthManager() *HealthManager {
	return &HealthManager{
		checkers:       make([]HealthChecker, 0),
		workerPoolSize: DefaultHealthWorkerPoolSize,
	}
}

// NewHealthManagerWithPoolSize creates a new health check manager with custom worker pool size
func NewHealthManagerWithPoolSize(poolSize int) *HealthManager {
	if poolSize <= 0 {
		poolSize = DefaultHealthWorkerPoolSize
	}
	return &HealthManager{
		checkers:       make([]HealthChecker, 0),
		workerPoolSize: poolSize,
	}
}

// Register adds a health checker to the manager
func (m *HealthManager) Register(checker HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers = append(m.checkers, checker)
}

// Check performs health checks and returns a report
// Uses a worker pool to limit concurrent goroutines
func (m *HealthManager) Check(ctx context.Context) *HealthReport {
	m.mu.RLock()
	checkers := make([]HealthChecker, len(m.checkers))
	copy(checkers, m.checkers)
	m.mu.RUnlock()

	if len(checkers) == 0 {
		return &HealthReport{
			Status:    HealthStatusHealthy,
			Timestamp: time.Now().UnixMilli(),
			Checks:    make(map[string]HealthCheck),
		}
	}

	type checkResult struct {
		name  string
		check HealthCheck
	}

	// Use buffered channel as a semaphore for worker pool
	poolSize := m.workerPoolSize
	if len(checkers) < poolSize {
		poolSize = len(checkers)
	}
	sem := make(chan struct{}, poolSize)

	results := make(chan checkResult, len(checkers))

	var wg sync.WaitGroup
	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()

			// Recover from panics in custom health checkers
			defer func() {
				if r := recover(); r != nil {
					results <- checkResult{
						name: c.Name(),
						check: HealthCheck{
							Name:    c.Name(),
							Status:  HealthStatusUnhealthy,
							Message: fmt.Sprintf("panic in health check: %v", r),
						},
					}
				}
			}()

			// Acquire semaphore slot (blocks if pool is full)
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release slot when done
			case <-ctx.Done():
				// Context cancelled, skip this check
				results <- checkResult{
					name: c.Name(),
					check: HealthCheck{
						Name:    c.Name(),
						Status:  HealthStatusUnhealthy,
						Message: "health check cancelled",
					},
				}
				return
			}

			start := time.Now()
			check := c.Check(ctx)
			check.ResponseTime = time.Since(start).Milliseconds()
			results <- checkResult{name: c.Name(), check: check}
		}(checker)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	checks := make(map[string]HealthCheck, len(checkers))
	overallStatus := HealthStatusHealthy

	for result := range results {
		checks[result.name] = result.check
		if result.check.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.check.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return &HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now().UnixMilli(),
		Checks:    checks,
	}
}

// Custom Checker

// CustomChecker is a wrapper for custom health check functions
type CustomChecker struct {
	name    string
	checkFn func(ctx context.Context) HealthCheck
}

// NewCustomChecker creates a new custom health checker
func NewCustomChecker(name string, checkFn func(ctx context.Context) HealthCheck) *CustomChecker {
	return &CustomChecker{name: name, checkFn: checkFn}
}

// Check performs the custom health check
func (c *CustomChecker) Check(ctx context.Context) HealthCheck {
	check := c.checkFn(ctx)
	if check.Name == "" {
		check.Name = c.name
	}
	return check
}

// Name returns the name of this health checker
func (c *CustomChecker) Name() string { return c.name }

// HTTP Checker

// HTTPChecker checks HTTP endpoint availability
type HTTPChecker struct {
	client  *http.Client
	url     string
	name    string
	timeout time.Duration
}

// NewHTTPChecker creates a new HTTP endpoint health checker
func NewHTTPChecker(url string, name string, timeout time.Duration) *HTTPChecker {
	if name == "" {
		name = configuration.DefaultHealthCheckerName
	}
	if timeout == 0 {
		timeout = configuration.DefaultHealthCheckTimeout
	}
	return &HTTPChecker{
		client:  &http.Client{Timeout: timeout},
		url:     url,
		name:    name,
		timeout: timeout,
	}
}

// Check performs the HTTP health check
func (h *HTTPChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return HealthCheck{
			Name:         h.name,
			Status:       HealthStatusUnhealthy,
			Message:      fmt.Sprintf(core.HealthMessageCreateRequestFailed, err),
			ResponseTime: time.Since(start).Milliseconds(),
			Error:        err,
		}
	}

	resp, err := h.client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return HealthCheck{
			Name:         h.name,
			Status:       HealthStatusUnhealthy,
			Message:      err.Error(),
			ResponseTime: responseTime.Milliseconds(),
			Error:        err,
			Details:      map[string]any{"url": h.url, "error": err.Error()},
		}
	}
	defer func() {
		_ = resp.Body.Close() // Explicitly ignore error as we're in defer
	}()

	status := HealthStatusHealthy
	message := core.HealthMessageHealthy
	if resp.StatusCode >= configuration.HealthStatusThresholdServerError {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf(core.HealthMessageUnhealthy, resp.StatusCode)
	} else if resp.StatusCode >= configuration.HealthStatusThresholdClientError {
		status = HealthStatusDegraded
		message = fmt.Sprintf(core.HealthMessageDegraded, resp.StatusCode)
	}

	return HealthCheck{
		Name:         h.name,
		Status:       status,
		Message:      message,
		ResponseTime: responseTime.Milliseconds(),
		Details:      map[string]any{"url": h.url, "status_code": resp.StatusCode},
	}
}

// Name returns the name of this health checker
func (h *HTTPChecker) Name() string { return h.name }
