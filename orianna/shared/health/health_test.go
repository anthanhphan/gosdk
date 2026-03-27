// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	if m.WorkerPoolSize() != DefaultWorkerPoolSize {
		t.Fatalf("expected default pool size %d, got %d", DefaultWorkerPoolSize, m.WorkerPoolSize())
	}
	if m.CheckerCount() != 0 {
		t.Fatalf("expected 0 checkers, got %d", m.CheckerCount())
	}
}

func TestNewManagerWithPoolSize(t *testing.T) {
	m := NewManagerWithPoolSize(5)
	if m.WorkerPoolSize() != 5 {
		t.Fatalf("expected pool size 5, got %d", m.WorkerPoolSize())
	}
}

func TestNewManagerWithPoolSize_Zero(t *testing.T) {
	m := NewManagerWithPoolSize(0)
	if m.WorkerPoolSize() != DefaultWorkerPoolSize {
		t.Fatalf("expected default pool size for zero input, got %d", m.WorkerPoolSize())
	}
}

func TestNewManagerWithPoolSize_Negative(t *testing.T) {
	m := NewManagerWithPoolSize(-1)
	if m.WorkerPoolSize() != DefaultWorkerPoolSize {
		t.Fatalf("expected default pool size for negative input, got %d", m.WorkerPoolSize())
	}
}

func TestManager_Register(t *testing.T) {
	m := NewManager()
	checker := NewCustomChecker("test", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "test", Status: StatusHealthy}
	})

	m.Register(checker)
	if m.CheckerCount() != 1 {
		t.Fatalf("expected 1 checker, got %d", m.CheckerCount())
	}
}

func TestManager_Check_NoCheckers(t *testing.T) {
	m := NewManager()
	report := m.Check(context.Background())

	if report.Status != StatusHealthy {
		t.Fatalf("expected healthy status with no checkers, got %s", report.Status)
	}
	if len(report.Checks) != 0 {
		t.Fatalf("expected 0 checks, got %d", len(report.Checks))
	}
}

func TestManager_Check_AllHealthy(t *testing.T) {
	m := NewManager()
	m.Register(NewCustomChecker("db", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "db", Status: StatusHealthy, Message: "connected"}
	}))
	m.Register(NewCustomChecker("cache", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "cache", Status: StatusHealthy, Message: "online"}
	}))

	report := m.Check(context.Background())

	if report.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", report.Status)
	}
	if len(report.Checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(report.Checks))
	}
}

func TestManager_Check_OneUnhealthy(t *testing.T) {
	m := NewManager()
	m.Register(NewCustomChecker("db", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "db", Status: StatusHealthy}
	}))
	m.Register(NewCustomChecker("cache", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "cache", Status: StatusUnhealthy, Message: "connection refused"}
	}))

	report := m.Check(context.Background())

	if report.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", report.Status)
	}
}

func TestManager_Check_Degraded(t *testing.T) {
	m := NewManager()
	m.Register(NewCustomChecker("db", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "db", Status: StatusHealthy}
	}))
	m.Register(NewCustomChecker("cache", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "cache", Status: StatusDegraded}
	}))

	report := m.Check(context.Background())

	if report.Status != StatusDegraded {
		t.Fatalf("expected degraded, got %s", report.Status)
	}
}

func TestManager_Check_UnhealthyOverridesDegraded(t *testing.T) {
	m := NewManager()
	m.Register(NewCustomChecker("a", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "a", Status: StatusDegraded}
	}))
	m.Register(NewCustomChecker("b", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "b", Status: StatusUnhealthy}
	}))

	report := m.Check(context.Background())

	if report.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy to override degraded, got %s", report.Status)
	}
}

func TestManager_Check_PanicRecovery(t *testing.T) {
	m := NewManager()
	m.Register(NewCustomChecker("panic", func(ctx context.Context) HealthCheck {
		panic("checker panic")
	}))

	report := m.Check(context.Background())

	if report.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy after panic, got %s", report.Status)
	}
	check, ok := report.Checks["panic"]
	if !ok {
		t.Fatal("expected panic checker in results")
	}
	if check.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy status for panicking checker, got %s", check.Status)
	}
}

func TestManager_Check_ContextCancelled(t *testing.T) {
	m := NewManagerWithPoolSize(1)

	// Register a checker that blocks
	m.Register(NewCustomChecker("slow", func(ctx context.Context) HealthCheck {
		<-ctx.Done()
		return HealthCheck{Name: "slow", Status: StatusUnhealthy, Message: "cancelled"}
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	report := m.Check(ctx)
	if report.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy for cancelled context, got %s", report.Status)
	}
}

func TestManager_Check_Concurrent(t *testing.T) {
	m := NewManager()

	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("checker-%d", i)
		m.Register(NewCustomChecker(name, func(ctx context.Context) HealthCheck {
			time.Sleep(10 * time.Millisecond) // Simulate work
			return HealthCheck{Name: name, Status: StatusHealthy}
		}))
	}

	report := m.Check(context.Background())

	if len(report.Checks) != 20 {
		t.Fatalf("expected 20 checks, got %d", len(report.Checks))
	}
}

func TestManager_ConcurrentRegisterAndCheck(t *testing.T) {
	m := NewManager()

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Register(NewCustomChecker(fmt.Sprintf("checker-%d", i), func(ctx context.Context) HealthCheck {
				return HealthCheck{Status: StatusHealthy}
			}))
		}(i)
	}

	// Concurrent checks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			report := m.Check(context.Background())
			if report == nil {
				errs <- errors.New("nil report")
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
}

func TestCustomChecker_Name(t *testing.T) {
	checker := NewCustomChecker("my-checker", func(ctx context.Context) HealthCheck {
		return HealthCheck{Status: StatusHealthy}
	})

	if checker.Name() != "my-checker" {
		t.Fatalf("expected 'my-checker', got %q", checker.Name())
	}
}

func TestCustomChecker_AutoFillsName(t *testing.T) {
	checker := NewCustomChecker("auto-name", func(ctx context.Context) HealthCheck {
		return HealthCheck{Status: StatusHealthy}
	})

	check := checker.Check(context.Background())
	if check.Name != "auto-name" {
		t.Fatalf("expected auto-filled name 'auto-name', got %q", check.Name)
	}
}

func TestCustomChecker_PreservesExistingName(t *testing.T) {
	checker := NewCustomChecker("checker-name", func(ctx context.Context) HealthCheck {
		return HealthCheck{Name: "custom-name", Status: StatusHealthy}
	})

	check := checker.Check(context.Background())
	if check.Name != "custom-name" {
		t.Fatalf("expected preserved name 'custom-name', got %q", check.Name)
	}
}

func TestHealthStatus_Constants(t *testing.T) {
	if StatusHealthy != "healthy" {
		t.Fatalf("expected 'healthy', got %q", StatusHealthy)
	}
	if StatusUnhealthy != "unhealthy" {
		t.Fatalf("expected 'unhealthy', got %q", StatusUnhealthy)
	}
	if StatusDegraded != "degraded" {
		t.Fatalf("expected 'degraded', got %q", StatusDegraded)
	}
}
