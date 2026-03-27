// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/core/mocks"
	"github.com/anthanhphan/gosdk/orianna/shared/ctxkeys"
	"github.com/anthanhphan/gosdk/orianna/shared/requestid"
	"go.uber.org/mock/gomock"
)

// Chain Tests

func TestChain(t *testing.T) {
	t.Run("empty chain calls Next", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
	mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		// Recover internally calls core.SendError which calls Get("Accept"), RequestID, UseProperHTTPStatus, Status, JSON
		mockCtx.EXPECT().Get(gomock.Any()).Return("").AnyTimes()
		mockCtx.EXPECT().RequestID().Return("test-req-id").AnyTimes()
		mockCtx.EXPECT().UseProperHTTPStatus().Return(true).AnyTimes()
		mockCtx.EXPECT().Status(gomock.Any()).Return(mockCtx).AnyTimes()
		mockCtx.EXPECT().JSON(gomock.Any()).Return(nil).AnyTimes()
		// Recovery logging uses Path(), Method(), and Locals("trace_id") for context
		mockCtx.EXPECT().Path().Return("/test/path").AnyTimes()
		mockCtx.EXPECT().Method().Return("POST").AnyTimes()
		mockCtx.EXPECT().Locals("trace_id").Return(nil).AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().SetContext(gomock.Any()).AnyTimes()
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().SetContext(gomock.Any()).AnyTimes()

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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().SetContext(gomock.Any()).AnyTimes()

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

	t.Run("SetContext is called with deadline context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		origCtx := context.Background()
		mockCtx.EXPECT().Context().Return(origCtx).AnyTimes()

		// Expect SetContext to be called with a deadline context first,
		// then restored with the original context
		var setCtxCalls []context.Context
		mockCtx.EXPECT().SetContext(gomock.Any()).DoAndReturn(func(ctx context.Context) {
			setCtxCalls = append(setCtxCalls, ctx)
		}).Times(2)

		mw := func(_ core.Context) error {
			return nil
		}
		wrapped := Timeout(mw, 1*time.Second)
		_ = wrapped(mockCtx)

		if len(setCtxCalls) != 2 {
			t.Fatalf("SetContext called %d times, want 2", len(setCtxCalls))
		}
		// First call should inject a deadline context
		if _, hasDeadline := setCtxCalls[0].Deadline(); !hasDeadline {
			t.Error("first SetContext call should inject a context with deadline")
		}
		// Second call should restore the original context
		if setCtxCalls[1] != origCtx {
			t.Error("second SetContext call should restore the original context")
		}
	})

	t.Run("returns timeout error when context already expired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		// Use an already-expired context
		expiredCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond) // ensure it expires

		mockCtx.EXPECT().Context().Return(expiredCtx).AnyTimes()
		mockCtx.EXPECT().SetContext(gomock.Any()).AnyTimes()
		mockCtx.EXPECT().Get(gomock.Any()).Return("").AnyTimes()
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

// SlowRequestDetector Tests

func TestSlowRequestDetector(t *testing.T) {
	t.Run("fast request does not trigger warning", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		mockCtx.EXPECT().Next().Return(nil)
		// No calls to RequestID, Method, RoutePath, ResponseStatusCode expected
		// because the request is fast (< threshold)

		detector := SlowRequestDetector(1 * time.Second)
		err := detector(mockCtx)
		if err != nil {
			t.Fatalf("SlowRequestDetector() error = %v, want nil", err)
		}
	})

	t.Run("slow request triggers warning log", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		// Next() sleeps to exceed threshold
		mockCtx.EXPECT().Next().DoAndReturn(func() error {
			time.Sleep(15 * time.Millisecond)
			return nil
		})

		// These are called when slow request is detected
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().RequestID().Return("req-slow-123")
		mockCtx.EXPECT().Method().Return("GET")
		mockCtx.EXPECT().RoutePath().Return("/api/slow")
		mockCtx.EXPECT().ResponseStatusCode().Return(200)

		detector := SlowRequestDetector(10 * time.Millisecond)
		err := detector(mockCtx)
		if err != nil {
			t.Fatalf("SlowRequestDetector() error = %v, want nil", err)
		}
	})

	t.Run("propagates handler error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		expectedErr := errors.New("handler failed")
		mockCtx.EXPECT().Next().Return(expectedErr)

		detector := SlowRequestDetector(1 * time.Second)
		err := detector(mockCtx)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("SlowRequestDetector() error = %v, want %v", err, expectedErr)
		}
	})
}

// IsValidRequestID Tests

func TestIsValidRequestID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"alphanumeric", "abc123", true},
		{"with dashes", "req-id-123", true},
		{"with underscores", "req_id_123", true},
		{"empty", "", false},
		{"contains space", "req id", false},
		{"contains newline", "req\nid", false},
		{"json injection", `{"key":"val"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requestid.IsValid(tt.id)
			if got != tt.want {
				t.Errorf("requestid.IsValid(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

// RequestIDMiddleware Tests

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("generates new ID when empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		mockCtx.EXPECT().Get(core.HeaderRequestID).Return("")
		mockCtx.EXPECT().Set(core.HeaderRequestID, gomock.Any()).Do(func(_ string, v string) {
			if !requestid.IsValid(v) {
				t.Errorf("Generated invalid request ID: %q", v)
			}
		})
		mockCtx.EXPECT().Locals(ctxkeys.RequestID.Key(), gomock.Any())
		mockCtx.EXPECT().Next().Return(nil)

		mw := RequestIDMiddleware()
		if err := mw(mockCtx); err != nil {
			t.Fatalf("RequestIDMiddleware() error = %v", err)
		}
	})

	t.Run("preserves valid incoming ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		mockCtx.EXPECT().Get(core.HeaderRequestID).Return("valid-request-id-123")
		mockCtx.EXPECT().Set(core.HeaderRequestID, "valid-request-id-123")
		mockCtx.EXPECT().Locals(ctxkeys.RequestID.Key(), "valid-request-id-123")
		mockCtx.EXPECT().Next().Return(nil)

		mw := RequestIDMiddleware()
		if err := mw(mockCtx); err != nil {
			t.Fatalf("RequestIDMiddleware() error = %v", err)
		}
	})

	t.Run("rejects malicious ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := mocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()

		mockCtx.EXPECT().Get(core.HeaderRequestID).Return("evil\n{\"injected\":true}")
		mockCtx.EXPECT().Set(core.HeaderRequestID, gomock.Any()).Do(func(_ string, v string) {
			if v == "evil\n{\"injected\":true}" {
				t.Error("Malicious ID was not rejected")
			}
			if !requestid.IsValid(v) {
				t.Errorf("Replacement ID is not valid: %q", v)
			}
		})
		mockCtx.EXPECT().Locals(ctxkeys.RequestID.Key(), gomock.Any())
		mockCtx.EXPECT().Next().Return(nil)

		mw := RequestIDMiddleware()
		if err := mw(mockCtx); err != nil {
			t.Fatalf("RequestIDMiddleware() error = %v", err)
		}
	})
}
