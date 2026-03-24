// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/anthanhphan/gosdk/logger"
)

// ---------------------------------------------------------------------------
// Panic recovery & logger
// ---------------------------------------------------------------------------

// recoverPanic handles panic recovery with lazy location capture.
// Location is only captured when a panic actually occurs (cold path),
// avoiding runtime.Stack/runtime.Callers overhead on every Run() call.
func recoverPanic() {
	r := recover()
	if r == nil {
		return
	}

	location := capturePanicLocation()

	getRecoverLogger().Errorw("panic recovered in goroutine",
		"type", "panic",
		"error", normalizePanicValue(r).Error(),
		"panic_at", location,
	)
}

// normalizePanicValue converts a panic value to an error.
func normalizePanicValue(r any) error {
	switch v := r.(type) {
	case error:
		return v
	case string:
		return fmt.Errorf("%s", v)
	default:
		return fmt.Errorf("%v", v)
	}
}

// capturePanicLocation captures the panic location from the current stack trace.
func capturePanicLocation() string {
	bufPtr := stackBufferPool.Get().(*[]byte)
	defer stackBufferPool.Put(bufPtr)

	buf := *bufPtr
	n := runtime.Stack(buf, false)
	parser := newStackTraceParser()
	stackTrace := string(buf[:n])
	return parser.extractLocation(stackTrace)
}

// ---------------------------------------------------------------------------
// Logger (lazy init)
// ---------------------------------------------------------------------------

var (
	recoverLoggerOnce sync.Once
	recoverLoggerInst *logger.Logger
)

// getRecoverLogger returns the cached recover logger, lazily initialized on first use.
func getRecoverLogger() *logger.Logger {
	recoverLoggerOnce.Do(func() {
		recoverLoggerInst = logger.NewLoggerWithFields(
			logger.String("prefix", "routine::recoverPanic"),
		)
	})
	return recoverLoggerInst
}
