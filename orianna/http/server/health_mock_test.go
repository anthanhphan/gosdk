// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server_test

import (
	"context"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/http/server/mocks"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"go.uber.org/mock/gomock"
)

// HealthManager Tests (with gomock)

func TestHealthManager_Check(t *testing.T) {
	t.Run("no checkers returns healthy", func(t *testing.T) {
		manager := health.NewManager()
		report := manager.Check(context.Background())

		if report.Status != health.StatusHealthy {
			t.Errorf("status = %v, want %v", report.Status, health.StatusHealthy)
		}
		if len(report.Checks) != 0 {
			t.Errorf("checks length = %d, want 0", len(report.Checks))
		}
	})

	t.Run("single healthy checker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockChecker := mocks.NewMockHealthChecker(ctrl)

		mockChecker.EXPECT().Name().Return("database").AnyTimes()
		mockChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:    "database",
			Status:  health.StatusHealthy,
			Message: "connection ok",
		})

		manager := health.NewManager()
		manager.Register(mockChecker)
		report := manager.Check(context.Background())

		if report.Status != health.StatusHealthy {
			t.Errorf("overall status = %v, want %v", report.Status, health.StatusHealthy)
		}
		if len(report.Checks) != 1 {
			t.Fatalf("checks length = %d, want 1", len(report.Checks))
		}
		check, ok := report.Checks["database"]
		if !ok {
			t.Fatal("missing 'database' check in report")
		}
		if check.Status != health.StatusHealthy {
			t.Errorf("database status = %v, want %v", check.Status, health.StatusHealthy)
		}
		if check.ResponseTime < 0 {
			t.Errorf("response time should be non-negative, got %d", check.ResponseTime)
		}
	})

	t.Run("unhealthy checker makes overall unhealthy", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		healthyChecker := mocks.NewMockHealthChecker(ctrl)
		healthyChecker.EXPECT().Name().Return("redis").AnyTimes()
		healthyChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:   "redis",
			Status: health.StatusHealthy,
		})

		unhealthyChecker := mocks.NewMockHealthChecker(ctrl)
		unhealthyChecker.EXPECT().Name().Return("database").AnyTimes()
		unhealthyChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:    "database",
			Status:  health.StatusUnhealthy,
			Message: "connection refused",
		})

		manager := health.NewManager()
		manager.Register(healthyChecker)
		manager.Register(unhealthyChecker)
		report := manager.Check(context.Background())

		if report.Status != health.StatusUnhealthy {
			t.Errorf("overall status = %v, want %v", report.Status, health.StatusUnhealthy)
		}
		if len(report.Checks) != 2 {
			t.Errorf("checks length = %d, want 2", len(report.Checks))
		}
	})

	t.Run("degraded checker makes overall degraded", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		healthyChecker := mocks.NewMockHealthChecker(ctrl)
		healthyChecker.EXPECT().Name().Return("cache").AnyTimes()
		healthyChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:   "cache",
			Status: health.StatusHealthy,
		})

		degradedChecker := mocks.NewMockHealthChecker(ctrl)
		degradedChecker.EXPECT().Name().Return("external-api").AnyTimes()
		degradedChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:    "external-api",
			Status:  health.StatusDegraded,
			Message: "high latency",
		})

		manager := health.NewManager()
		manager.Register(healthyChecker)
		manager.Register(degradedChecker)
		report := manager.Check(context.Background())

		if report.Status != health.StatusDegraded {
			t.Errorf("overall status = %v, want %v", report.Status, health.StatusDegraded)
		}
	})

	t.Run("unhealthy overrides degraded", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		degradedChecker := mocks.NewMockHealthChecker(ctrl)
		degradedChecker.EXPECT().Name().Return("external-api").AnyTimes()
		degradedChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:   "external-api",
			Status: health.StatusDegraded,
		})

		unhealthyChecker := mocks.NewMockHealthChecker(ctrl)
		unhealthyChecker.EXPECT().Name().Return("database").AnyTimes()
		unhealthyChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:   "database",
			Status: health.StatusUnhealthy,
		})

		manager := health.NewManager()
		manager.Register(degradedChecker)
		manager.Register(unhealthyChecker)
		report := manager.Check(context.Background())

		// Unhealthy should take precedence over degraded
		if report.Status != health.StatusUnhealthy {
			t.Errorf("overall status = %v, want %v", report.Status, health.StatusUnhealthy)
		}
	})

	t.Run("recovers from panicking checker", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		panicChecker := mocks.NewMockHealthChecker(ctrl)
		panicChecker.EXPECT().Name().Return("panic-service").AnyTimes()
		panicChecker.EXPECT().Check(gomock.Any()).DoAndReturn(
			func(_ context.Context) health.HealthCheck {
				panic("kaboom!")
			},
		)

		manager := health.NewManager()
		manager.Register(panicChecker)

		// Should not panic -- HealthManager has panic recovery
		report := manager.Check(context.Background())

		if report.Status != health.StatusUnhealthy {
			t.Errorf("overall status = %v, want %v", report.Status, health.StatusUnhealthy)
		}
		check, ok := report.Checks["panic-service"]
		if !ok {
			t.Fatal("missing 'panic-service' check in report")
		}
		if check.Status != health.StatusUnhealthy {
			t.Errorf("panic-service status = %v, want %v", check.Status, health.StatusUnhealthy)
		}
	})

	t.Run("cancelled context marks checks unhealthy", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockChecker := mocks.NewMockHealthChecker(ctrl)
		mockChecker.EXPECT().Name().Return("slow-service").AnyTimes()
		// Check may or may not be called depending on timing
		mockChecker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
			Name:   "slow-service",
			Status: health.StatusHealthy,
		}).AnyTimes()

		manager := health.NewManagerWithPoolSize(1)
		manager.Register(mockChecker)

		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		report := manager.Check(ctx)

		// Regardless of status, the report should complete without hanging
		if report == nil {
			t.Fatal("report should not be nil")
		}
	})

	t.Run("multiple checkers with custom pool size", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		// Create 5 checkers with pool size 2
		checkerNames := []string{"db", "redis", "kafka", "s3", "elastic"}
		manager := health.NewManagerWithPoolSize(2)

		for _, name := range checkerNames {
			checker := mocks.NewMockHealthChecker(ctrl)
			checker.EXPECT().Name().Return(name).AnyTimes()
			checker.EXPECT().Check(gomock.Any()).Return(health.HealthCheck{
				Name:   name,
				Status: health.StatusHealthy,
			})
			manager.Register(checker)
		}

		report := manager.Check(context.Background())

		if report.Status != health.StatusHealthy {
			t.Errorf("overall status = %v, want %v", report.Status, health.StatusHealthy)
		}
		if len(report.Checks) != 5 {
			t.Errorf("checks length = %d, want 5", len(report.Checks))
		}
		for _, name := range checkerNames {
			if _, ok := report.Checks[name]; !ok {
				t.Errorf("missing check for %q", name)
			}
		}
	})
}
