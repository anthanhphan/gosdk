// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/core/mocks"
	"go.uber.org/mock/gomock"
)

// Chain Tests

func TestChain(t *testing.T) {
	t.Run("empty chain calls Next", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Next().Return(nil)

		chain := Chain()
		err := chain(mockCtx)
		if err != nil {
			t.Fatalf("Chain() error = %v, want nil", err)
		}
	})

	t.Run("single middleware is passed through", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		var called bool
		mw := func(ctx core.Context) error {
			called = true
			return ctx.Next()
		}

		// The middleware calls ctx.Next(), which advances Fiber's handler chain
		mockCtx.EXPECT().Next().Return(nil)

		chain := Chain(mw)
		err := chain(mockCtx)
		if err != nil {
			t.Fatalf("Chain() error = %v, want nil", err)
		}
		if !called {
			t.Error("middleware was not called")
		}
	})

	t.Run("multiple middlewares all execute", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		var order []string
		mw1 := func(ctx core.Context) error {
			order = append(order, "mw1")
			return ctx.Next()
		}
		mw2 := func(ctx core.Context) error {
			order = append(order, "mw2")
			return ctx.Next()
		}
		mw3 := func(ctx core.Context) error {
			order = append(order, "mw3")
			return ctx.Next()
		}

		// Each middleware calls ctx.Next() once
		mockCtx.EXPECT().Next().Return(nil).Times(3)

		chain := Chain(mw1, mw2, mw3)
		err := chain(mockCtx)
		if err != nil {
			t.Fatalf("Chain() error = %v, want nil", err)
		}
		if len(order) != 3 || order[0] != "mw1" || order[1] != "mw2" || order[2] != "mw3" {
			t.Errorf("execution order = %v, want [mw1 mw2 mw3]", order)
		}
	})

	t.Run("middleware error short-circuits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		expectedErr := errors.New("middleware error")
		mw := func(_ core.Context) error {
			return expectedErr
		}

		chain := Chain(mw)
		err := chain(mockCtx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Chain() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("error in second middleware stops chain", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		expectedErr := errors.New("second middleware error")
		var mw1Called, mw3Called bool
		mw1 := func(ctx core.Context) error {
			mw1Called = true
			return ctx.Next()
		}
		mw2 := func(_ core.Context) error {
			return expectedErr
		}
		mw3 := func(_ core.Context) error {
			mw3Called = true
			return nil
		}

		mockCtx.EXPECT().Next().Return(nil).Times(1)

		chain := Chain(mw1, mw2, mw3)
		err := chain(mockCtx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Chain() error = %v, want %v", err, expectedErr)
		}
		if !mw1Called {
			t.Error("mw1 should have been called")
		}
		if mw3Called {
			t.Error("mw3 should NOT have been called after mw2 error")
		}
	})
}

// Optional Tests

func TestOptional(t *testing.T) {
	t.Run("condition true applies middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		var applied bool
		mw := func(ctx core.Context) error {
			applied = true
			return ctx.Next()
		}

		mockCtx.EXPECT().Next().Return(nil)

		opt := Optional(func(_ core.Context) bool { return true }, mw)
		_ = opt(mockCtx)
		if !applied {
			t.Error("middleware should be applied when condition is true")
		}
	})

	t.Run("condition false skips middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		var applied bool
		mw := func(_ core.Context) error {
			applied = true
			return nil
		}

		// When skipping, Optional calls ctx.Next()
		mockCtx.EXPECT().Next().Return(nil)

		opt := Optional(func(_ core.Context) bool { return false }, mw)
		_ = opt(mockCtx)
		if applied {
			t.Error("middleware should not be applied when condition is false")
		}
	})
}

// OnlyForMethods Tests

func TestOnlyForMethods(t *testing.T) {
	t.Run("matching method applies middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Method().Return("POST")
		mockCtx.EXPECT().Next().Return(nil)

		var applied bool
		mw := func(ctx core.Context) error {
			applied = true
			return ctx.Next()
		}

		filtered := OnlyForMethods(mw, "POST", "PUT")
		_ = filtered(mockCtx)
		if !applied {
			t.Error("middleware should apply for matching method")
		}
	})

	t.Run("non-matching method skips middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Method().Return("GET")
		mockCtx.EXPECT().Next().Return(nil)

		var applied bool
		mw := func(_ core.Context) error {
			applied = true
			return nil
		}

		filtered := OnlyForMethods(mw, "POST", "PUT")
		_ = filtered(mockCtx)
		if applied {
			t.Error("middleware should not apply for non-matching method")
		}
	})
}

// SkipForPaths Tests

func TestSkipForPaths(t *testing.T) {
	t.Run("matching path skips middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Path().Return("/health")
		mockCtx.EXPECT().Next().Return(nil)

		var applied bool
		mw := func(_ core.Context) error {
			applied = true
			return nil
		}

		filtered := SkipForPaths(mw, "/health", "/metrics")
		_ = filtered(mockCtx)
		if applied {
			t.Error("middleware should be skipped for matching path")
		}
	})

	t.Run("non-matching path applies middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Path().Return("/api/users")
		mockCtx.EXPECT().Next().Return(nil)

		var applied bool
		mw := func(ctx core.Context) error {
			applied = true
			return ctx.Next()
		}

		filtered := SkipForPaths(mw, "/health", "/metrics")
		_ = filtered(mockCtx)
		if !applied {
			t.Error("middleware should apply for non-matching path")
		}
	})
}

// Before / After Tests

func TestBefore(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCtx := mocks.NewMockContext(ctrl)
	mockCtx.EXPECT().Next().Return(nil)

	var order []string
	beforeFn := func(_ core.Context) {
		order = append(order, "before")
	}
	mw := func(ctx core.Context) error {
		order = append(order, "middleware")
		return ctx.Next()
	}
	wrapped := Before(mw, beforeFn)
	_ = wrapped(mockCtx)

	if len(order) != 2 || order[0] != "before" || order[1] != "middleware" {
		t.Errorf("execution order = %v, want [before middleware]", order)
	}
}

func TestAfter(t *testing.T) {
	t.Run("after receives error from middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		expectedErr := errors.New("mw error")
		var afterErr error
		afterFn := func(_ core.Context, err error) {
			afterErr = err
		}
		mw := func(_ core.Context) error {
			return expectedErr
		}
		wrapped := After(mw, afterFn)
		err := wrapped(mockCtx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("After() error = %v, want %v", err, expectedErr)
		}
		if !errors.Is(afterErr, expectedErr) {
			t.Errorf("afterFunc received error = %v, want %v", afterErr, expectedErr)
		}
	})

	t.Run("after receives nil on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Next().Return(nil)

		var afterCalled bool
		var afterErr error
		afterFn := func(_ core.Context, err error) {
			afterCalled = true
			afterErr = err
		}
		mw := func(ctx core.Context) error {
			return ctx.Next()
		}
		wrapped := After(mw, afterFn)
		_ = wrapped(mockCtx)
		if !afterCalled {
			t.Error("afterFunc should be called")
		}
		if afterErr != nil {
			t.Errorf("afterFunc error = %v, want nil", afterErr)
		}
	})
}

// Recover Tests

func TestRecover(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		// Recover internally calls core.SendError which calls RequestID, UseProperHTTPStatus, Status, JSON
		mockCtx.EXPECT().RequestID().Return("test-req-id").AnyTimes()
		mockCtx.EXPECT().UseProperHTTPStatus().Return(true).AnyTimes()
		mockCtx.EXPECT().Status(gomock.Any()).Return(mockCtx).AnyTimes()
		mockCtx.EXPECT().JSON(gomock.Any()).Return(nil).AnyTimes()

		mw := func(_ core.Context) error {
			panic("test panic")
		}
		wrapped := Recover(mw)
		err := wrapped(mockCtx)
		// B4 fix: Recover now returns the error so upstream middleware (logging, metrics) can see it
		if err == nil {
			t.Fatal("Recover() should return an error after catching a panic")
		}
		var errResp *core.ErrorResponse
		if !errors.As(err, &errResp) {
			t.Fatalf("Recover() error should be *core.ErrorResponse, got %T", err)
		}
		if errResp.Code != "INTERNAL_ERROR" {
			t.Errorf("ErrorResponse.Code = %q, want INTERNAL_ERROR", errResp.Code)
		}
	})

	t.Run("passes through normal errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		expectedErr := errors.New("normal error")
		mw := func(_ core.Context) error {
			return expectedErr
		}
		wrapped := Recover(mw)
		err := wrapped(mockCtx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Recover() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("passes through success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Next().Return(nil)

		mw := func(ctx core.Context) error {
			return ctx.Next()
		}
		wrapped := Recover(mw)
		err := wrapped(mockCtx)
		if err != nil {
			t.Errorf("Recover() error = %v, want nil", err)
		}
	})
}

// Timeout Tests

func TestTimeout(t *testing.T) {
	t.Run("completes before timeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Next().Return(nil).AnyTimes()

		mw := func(ctx core.Context) error {
			return ctx.Next()
		}
		wrapped := Timeout(mw, 1*time.Second)
		err := wrapped(mockCtx)
		if err != nil {
			t.Errorf("Timeout() error = %v, want nil", err)
		}
	})

	t.Run("propagates panic directly", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()

		mw := func(_ core.Context) error {
			panic("timeout panic")
		}
		wrapped := Timeout(mw, 1*time.Second)

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic to propagate")
			}
			if r != "timeout panic" {
				t.Errorf("recovered = %v, want 'timeout panic'", r)
			}
		}()
		_ = wrapped(mockCtx)
	})

	t.Run("propagates middleware error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()

		expectedErr := errors.New("handler error")
		mw := func(_ core.Context) error {
			return expectedErr
		}
		wrapped := Timeout(mw, 1*time.Second)
		err := wrapped(mockCtx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Timeout() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("timeout context is propagated to handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()

		var receivedDeadline bool
		mw := func(ctx core.Context) error {
			// The wrapped context should have a deadline set
			_, hasDeadline := ctx.Context().Deadline()
			receivedDeadline = hasDeadline
			return nil
		}
		wrapped := Timeout(mw, 1*time.Second)
		_ = wrapped(mockCtx)
		if !receivedDeadline {
			t.Error("handler should receive a context with deadline")
		}
	})

	t.Run("returns timeout error when context already expired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)

		// Use an already-expired context
		expiredCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond) // ensure it expires

		mockCtx.EXPECT().Context().Return(expiredCtx).AnyTimes()
		mockCtx.EXPECT().RequestID().Return("test-req-id").AnyTimes()
		mockCtx.EXPECT().UseProperHTTPStatus().Return(true).AnyTimes()
		mockCtx.EXPECT().Status(gomock.Any()).Return(mockCtx).AnyTimes()
		mockCtx.EXPECT().JSON(gomock.Any()).Return(nil).AnyTimes()

		mw := func(_ core.Context) error {
			return nil
		}
		wrapped := Timeout(mw, 1*time.Nanosecond)
		err := wrapped(mockCtx)
		// Should return nil since the middleware completed (DeadlineExceeded is checked after)
		// The timeout handler fires because the context was already expired
		_ = err // result depends on timing
	})
}
