package errors

import (
	"fmt"
	"testing"
)

func TestIsConfigError(t *testing.T) {
	if !IsConfigError(ErrInvalidConfig) {
		t.Error("expected true for ErrInvalidConfig")
	}

	wrapped := fmt.Errorf("wrapper: %w", ErrInvalidConfig)
	if !IsConfigError(wrapped) {
		t.Error("expected true for wrapped ErrInvalidConfig")
	}

	if IsConfigError(ErrHandlerNil) {
		t.Error("expected false for unrelated error")
	}
}

func TestIsServerError(t *testing.T) {
	if !IsServerError(ErrServerNotStarted) {
		t.Error("expected true for ErrServerNotStarted")
	}

	if !IsServerError(ErrServerShutdown) {
		t.Error("expected true for ErrServerShutdown")
	}

	wrappedStarted := fmt.Errorf("wrapper: %w", ErrServerNotStarted)
	if !IsServerError(wrappedStarted) {
		t.Error("expected true for wrapped ErrServerNotStarted")
	}

	wrappedShutdown := fmt.Errorf("wrapper: %w", ErrServerShutdown)
	if !IsServerError(wrappedShutdown) {
		t.Error("expected true for wrapped ErrServerShutdown")
	}

	if IsServerError(ErrInvalidConfig) {
		t.Error("expected false for unrelated error")
	}
}
