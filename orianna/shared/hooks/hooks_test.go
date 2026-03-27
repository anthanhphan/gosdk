package hooks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestHooks_Lifecycle(t *testing.T) {
	h := New[context.Context, int]()

	reqOk := false
	h.AddOnRequest(func(ctx context.Context) {
		reqOk = true
	})

	resOk := false
	h.AddOnResponse(func(ctx context.Context, code int, latency time.Duration) {
		resOk = true
	})

	errOk := false
	h.AddOnError(func(ctx context.Context, err error) {
		errOk = true
	})

	panicOk := false
	h.AddOnPanic(func(ctx context.Context, recovered any, stack []byte) {
		panicOk = true
	})

	shutOk := false
	h.AddOnShutdown(func() {
		shutOk = true
	})

	startErr := errors.New("start error")
	startOk := false
	h.AddOnServerStart(func(server any) error {
		startOk = true
		return startErr
	})

	ctx := context.Background()
	h.ExecuteOnRequest(ctx)
	h.ExecuteOnResponse(ctx, 200, time.Millisecond)
	h.ExecuteOnError(ctx, errors.New("test"))
	h.ExecuteOnPanic(ctx, "panic", []byte{})
	h.ExecuteOnShutdown()
	err := h.ExecuteOnServerStart(nil)

	if !reqOk || !resOk || !errOk || !panicOk || !shutOk || !startOk {
		t.Error("one or more hooks failed to execute properly")
	}
	if err != startErr {
		t.Errorf("expected startErr, got %v", err)
	}
}

func TestHooks_PanicRecovery(t *testing.T) {
	h := New[context.Context, int]()

	h.AddOnRequest(func(ctx context.Context) { panic("boom") })
	h.ExecuteOnRequest(context.Background()) // Should recover

	h.AddOnResponse(func(ctx context.Context, code int, latency time.Duration) { panic("boom") })
	h.ExecuteOnResponse(context.Background(), 200, 0) // Should recover

	h.AddOnError(func(ctx context.Context, err error) { panic("boom") })
	h.ExecuteOnError(context.Background(), nil) // Should recover

	h.AddOnPanic(func(ctx context.Context, r any, s []byte) { panic("boom") })
	h.ExecuteOnPanic(context.Background(), nil, nil) // Should recover

	h.AddOnServerStart(func(server any) error { panic("start payload") })
	err := h.ExecuteOnServerStart(nil)
	if err == nil || err.Error() != "OnServerStart hook panicked: start payload" {
		t.Errorf("expected panic error, got %v", err)
	}
}

func TestHooks_EmptyShutdown(t *testing.T) {
	h := New[context.Context, int]()
	h.ExecuteOnShutdown() // Should return immediately without panic
}
