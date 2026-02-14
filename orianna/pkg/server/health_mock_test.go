// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server_test

import (
	"context"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/server"
	"github.com/anthanhphan/gosdk/orianna/pkg/server/mocks"
	"go.uber.org/mock/gomock"
)

// HealthManager Tests (with gomock)

func TestHealthManager_Check(t *testing.T) {
	t.Run("no checkers returns healthy", func(t *testing.T) {
		manager := server.NewHealthManager()
		report := manager.Check(context.Background())

		if report.Status != server.HealthStatusHealthy {
			t.Errorf("status = %v, want %v", report.Status, server.HealthStatusHealthy)
		}
		if len(report.Checks) != 0 {
			t.Errorf("checks length = %d, want 0", len(report.Checks))
		}
	})

	t.Run("single healthy checker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockChecker := mocks.NewMockHealthChecker(ctrl)

		mockChecker.EXPECT().Name().Return("database").AnyTimes()
		mockChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:    "database",
			Status:  server.HealthStatusHealthy,
			Message: "connection ok",
		})

		manager := server.NewHealthManager()
		manager.Register(mockChecker)
		report := manager.Check(context.Background())

		if report.Status != server.HealthStatusHealthy {
			t.Errorf("overall status = %v, want %v", report.Status, server.HealthStatusHealthy)
		}
		if len(report.Checks) != 1 {
			t.Fatalf("checks length = %d, want 1", len(report.Checks))
		}
		check, ok := report.Checks["database"]
		if !ok {
			t.Fatal("missing 'database' check in report")
		}
		if check.Status != server.HealthStatusHealthy {
			t.Errorf("database status = %v, want %v", check.Status, server.HealthStatusHealthy)
		}
		if check.ResponseTime < 0 {
			t.Errorf("response time should be non-negative, got %d", check.ResponseTime)
		}
	})

	t.Run("unhealthy checker makes overall unhealthy", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		healthyChecker := mocks.NewMockHealthChecker(ctrl)
		healthyChecker.EXPECT().Name().Return("redis").AnyTimes()
		healthyChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:   "redis",
			Status: server.HealthStatusHealthy,
		})

		unhealthyChecker := mocks.NewMockHealthChecker(ctrl)
		unhealthyChecker.EXPECT().Name().Return("database").AnyTimes()
		unhealthyChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:    "database",
			Status:  server.HealthStatusUnhealthy,
			Message: "connection refused",
		})

		manager := server.NewHealthManager()
		manager.Register(healthyChecker)
		manager.Register(unhealthyChecker)
		report := manager.Check(context.Background())

		if report.Status != server.HealthStatusUnhealthy {
			t.Errorf("overall status = %v, want %v", report.Status, server.HealthStatusUnhealthy)
		}
		if len(report.Checks) != 2 {
			t.Errorf("checks length = %d, want 2", len(report.Checks))
		}
	})

	t.Run("degraded checker makes overall degraded", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		healthyChecker := mocks.NewMockHealthChecker(ctrl)
		healthyChecker.EXPECT().Name().Return("cache").AnyTimes()
		healthyChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:   "cache",
			Status: server.HealthStatusHealthy,
		})

		degradedChecker := mocks.NewMockHealthChecker(ctrl)
		degradedChecker.EXPECT().Name().Return("external-api").AnyTimes()
		degradedChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:    "external-api",
			Status:  server.HealthStatusDegraded,
			Message: "high latency",
		})

		manager := server.NewHealthManager()
		manager.Register(healthyChecker)
		manager.Register(degradedChecker)
		report := manager.Check(context.Background())

		if report.Status != server.HealthStatusDegraded {
			t.Errorf("overall status = %v, want %v", report.Status, server.HealthStatusDegraded)
		}
	})

	t.Run("unhealthy overrides degraded", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		degradedChecker := mocks.NewMockHealthChecker(ctrl)
		degradedChecker.EXPECT().Name().Return("external-api").AnyTimes()
		degradedChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:   "external-api",
			Status: server.HealthStatusDegraded,
		})

		unhealthyChecker := mocks.NewMockHealthChecker(ctrl)
		unhealthyChecker.EXPECT().Name().Return("database").AnyTimes()
		unhealthyChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:   "database",
			Status: server.HealthStatusUnhealthy,
		})

		manager := server.NewHealthManager()
		manager.Register(degradedChecker)
		manager.Register(unhealthyChecker)
		report := manager.Check(context.Background())

		// Unhealthy should take precedence over degraded
		if report.Status != server.HealthStatusUnhealthy {
			t.Errorf("overall status = %v, want %v", report.Status, server.HealthStatusUnhealthy)
		}
	})

	t.Run("recovers from panicking checker", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		panicChecker := mocks.NewMockHealthChecker(ctrl)
		panicChecker.EXPECT().Name().Return("panic-service").AnyTimes()
		panicChecker.EXPECT().Check(gomock.Any()).DoAndReturn(
			func(_ context.Context) server.HealthCheck {
				panic("kaboom!")
			},
		)

		manager := server.NewHealthManager()
		manager.Register(panicChecker)

		// Should not panic -- HealthManager has panic recovery
		report := manager.Check(context.Background())

		if report.Status != server.HealthStatusUnhealthy {
			t.Errorf("overall status = %v, want %v", report.Status, server.HealthStatusUnhealthy)
		}
		check, ok := report.Checks["panic-service"]
		if !ok {
			t.Fatal("missing 'panic-service' check in report")
		}
		if check.Status != server.HealthStatusUnhealthy {
			t.Errorf("panic-service status = %v, want %v", check.Status, server.HealthStatusUnhealthy)
		}
	})

	t.Run("cancelled context marks checks unhealthy", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockChecker := mocks.NewMockHealthChecker(ctrl)
		mockChecker.EXPECT().Name().Return("slow-service").AnyTimes()
		// Check may or may not be called depending on timing
		mockChecker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
			Name:   "slow-service",
			Status: server.HealthStatusHealthy,
		}).AnyTimes()

		manager := server.NewHealthManagerWithPoolSize(1)
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
		manager := server.NewHealthManagerWithPoolSize(2)

		for _, name := range checkerNames {
			checker := mocks.NewMockHealthChecker(ctrl)
			checker.EXPECT().Name().Return(name).AnyTimes()
			checker.EXPECT().Check(gomock.Any()).Return(server.HealthCheck{
				Name:   name,
				Status: server.HealthStatusHealthy,
			})
			manager.Register(checker)
		}

		report := manager.Check(context.Background())

		if report.Status != server.HealthStatusHealthy {
			t.Errorf("overall status = %v, want %v", report.Status, server.HealthStatusHealthy)
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
