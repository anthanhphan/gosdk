// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHooks_NewHooks(t *testing.T) {
	h := NewHooks()
	assert.NotNil(t, h)
}

func TestHooks_OnRequest(t *testing.T) {
	h := NewHooks()
	called := false
	h.AddOnRequest(func(ctx Context) {
		called = true
	})

	sc := NewServerContext(context.Background(), "/svc/m")
	h.ExecuteOnRequest(sc)
	assert.True(t, called)
}

func TestHooks_OnResponse(t *testing.T) {
	h := NewHooks()
	var capturedCode string
	h.AddOnResponse(func(ctx Context, code string, latency time.Duration) {
		capturedCode = code
	})

	sc := NewServerContext(context.Background(), "/svc/m")
	h.ExecuteOnResponse(sc, "OK", time.Millisecond)
	assert.Equal(t, "OK", capturedCode)
}

func TestHooks_OnError(t *testing.T) {
	h := NewHooks()
	var capturedErr error
	h.AddOnError(func(ctx Context, err error) {
		capturedErr = err
	})

	testErr := assert.AnError
	sc := NewServerContext(context.Background(), "/svc/m")
	h.ExecuteOnError(sc, testErr)
	assert.Equal(t, testErr, capturedErr)
}

func TestHooks_OnShutdown(t *testing.T) {
	h := NewHooks()
	called := false
	h.AddOnShutdown(func() {
		called = true
	})

	h.ExecuteOnShutdown()
	assert.True(t, called)
}

func TestHooks_OnServerStart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := NewHooks()
		h.AddOnServerStart(func(server any) error {
			return nil
		})
		assert.NoError(t, h.ExecuteOnServerStart(nil))
	})

	t.Run("error stops chain", func(t *testing.T) {
		h := NewHooks()
		h.AddOnServerStart(func(server any) error {
			return assert.AnError
		})
		secondCalled := false
		h.AddOnServerStart(func(server any) error {
			secondCalled = true
			return nil
		})
		assert.Error(t, h.ExecuteOnServerStart(nil))
		assert.False(t, secondCalled)
	})
}

func TestHooks_OnPanic(t *testing.T) {
	h := NewHooks()
	var capturedVal any
	h.AddOnPanic(func(ctx Context, recovered any, stack []byte) {
		capturedVal = recovered
	})

	sc := NewServerContext(context.Background(), "/svc/m")
	h.ExecuteOnPanic(sc, "panic!", nil)
	assert.Equal(t, "panic!", capturedVal)
}
