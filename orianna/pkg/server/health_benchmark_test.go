// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkHealthCheck benchmarks the health check system with various scenarios
func BenchmarkHealthCheck(b *testing.B) {
	b.Run("SingleChecker", func(b *testing.B) {
		manager := NewHealthManager()
		manager.Register(NewCustomChecker("test", func(ctx context.Context) HealthCheck {
			return HealthCheck{
				Name:    "test",
				Status:  HealthStatusHealthy,
				Message: "OK",
			}
		}))

		ctx := context.Background()
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = manager.Check(ctx)
		}
	})

	b.Run("MultipleCheckers", func(b *testing.B) {
		manager := NewHealthManager()
		// Add 10 health checkers
		for i := 0; i < 10; i++ {
			name := "checker"
			manager.Register(NewCustomChecker(name, func(ctx context.Context) HealthCheck {
				return HealthCheck{
					Name:    name,
					Status:  HealthStatusHealthy,
					Message: "OK",
				}
			}))
		}

		ctx := context.Background()
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = manager.Check(ctx)
		}
	})

	b.Run("SlowCheckers", func(b *testing.B) {
		manager := NewHealthManagerWithPoolSize(5)
		// Add 10 slow health checkers
		for i := 0; i < 10; i++ {
			name := "slow-checker"
			manager.Register(NewCustomChecker(name, func(ctx context.Context) HealthCheck {
				time.Sleep(10 * time.Millisecond)
				return HealthCheck{
					Name:    name,
					Status:  HealthStatusHealthy,
					Message: "OK",
				}
			}))
		}

		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = manager.Check(ctx)
		}
	})
}

// TestHealthCheckMapAllocation tests that the health check results map is properly pre-allocated
func TestHealthCheckMapAllocation(t *testing.T) {
	manager := NewHealthManager()

	// Add multiple checkers with unique names
	numCheckers := 5
	for i := 0; i < numCheckers; i++ {
		// Capture the loop variable for closure
		checkerIndex := i
		checkerName := fmt.Sprintf("checker-%d", checkerIndex)
		manager.Register(NewCustomChecker(checkerName, func(ctx context.Context) HealthCheck {
			return HealthCheck{
				Name:    checkerName,
				Status:  HealthStatusHealthy,
				Message: "OK",
			}
		}))
	}

	ctx := context.Background()
	report := manager.Check(ctx)

	// Verify all checkers reported
	if len(report.Checks) != numCheckers {
		t.Errorf("Expected %d health checks, got %d", numCheckers, len(report.Checks))
	}

	// Verify overall status
	if report.Status != HealthStatusHealthy {
		t.Errorf("Expected overall status %s, got %s", HealthStatusHealthy, report.Status)
	}
}

// TestHealthCheckCancellation tests that health checks respect context cancellation
func TestHealthCheckCancellation(t *testing.T) {
	manager := NewHealthManager()

	// Add a slow checker
	manager.Register(NewCustomChecker("slow", func(ctx context.Context) HealthCheck {
		select {
		case <-time.After(5 * time.Second):
			return HealthCheck{Name: "slow", Status: HealthStatusHealthy}
		case <-ctx.Done():
			return HealthCheck{Name: "slow", Status: HealthStatusUnhealthy, Message: "cancelled"}
		}
	}))

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	report := manager.Check(ctx)

	// Should complete quickly even though checker would take 5 seconds
	if report.Status == HealthStatusHealthy {
		t.Error("Expected unhealthy status due to cancellation")
	}
}
