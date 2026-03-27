// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewStatusError(t *testing.T) {
	se := NewStatusError(codes.NotFound, "not found")
	assert.Equal(t, codes.NotFound, se.Code)
	assert.Equal(t, "not found", se.Message)
	assert.Contains(t, se.Error(), "NotFound")
}

func TestStatusError_WithInternalMsg(t *testing.T) {
	se := NewStatusError(codes.Internal, "failed").WithInternalMsg("db error: %s", "timeout")
	assert.Contains(t, se.Error(), "internal: db error: timeout")
}

func TestStatusError_WithCause(t *testing.T) {
	cause := errors.New("root cause")
	se := NewStatusError(codes.Internal, "failed").WithCause(cause)
	assert.ErrorIs(t, se, cause)
}

func TestStatusError_GRPCStatus(t *testing.T) {
	se := NewStatusError(codes.NotFound, "not found")
	st := se.GRPCStatus()
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestStatusError_Is(t *testing.T) {
	se1 := NewStatusError(codes.NotFound, "a")
	se2 := NewStatusError(codes.NotFound, "b")
	se3 := NewStatusError(codes.Internal, "c")

	assert.True(t, se1.Is(se2))
	assert.False(t, se1.Is(se3))
}

func TestIsCode(t *testing.T) {
	se := NewStatusError(codes.NotFound, "not found")
	assert.True(t, IsCode(se, codes.NotFound))
	assert.False(t, IsCode(se, codes.Internal))
}

func TestHandleError(t *testing.T) {
	se := NewStatusError(codes.NotFound, "not found")
	grpcErr, ok := HandleError(se)
	assert.True(t, ok)
	assert.NotNil(t, grpcErr)

	_, ok = HandleError(errors.New("plain error"))
	assert.False(t, ok)
}

func TestWrapError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Nil(t, WrapError(nil, "msg"))
	})

	t.Run("plain error", func(t *testing.T) {
		err := WrapError(errors.New("fail"), "context")
		var se *StatusError
		assert.True(t, errors.As(err, &se))
		assert.Equal(t, codes.Internal, se.Code)
	})

	t.Run("status error", func(t *testing.T) {
		original := NewStatusError(codes.NotFound, "not found")
		err := WrapError(original, "extra context")
		var se *StatusError
		assert.True(t, errors.As(err, &se))
		assert.Equal(t, codes.NotFound, se.Code)
	})
}

func TestWrapErrorf(t *testing.T) {
	err := WrapErrorf(errors.New("fail"), "context %d", 42)
	var se *StatusError
	assert.True(t, errors.As(err, &se))
	assert.Equal(t, "context 42", se.InternalMessage)

	// nil returns nil
	assert.Nil(t, WrapErrorf(nil, "msg"))
}

func TestIsCode_GRPCStatusError(t *testing.T) {
	// Test with a standard gRPC status error (not a StatusError)
	grpcErr := status.Error(codes.NotFound, "not found")
	assert.True(t, IsCode(grpcErr, codes.NotFound))
	assert.False(t, IsCode(grpcErr, codes.Internal))
}

func TestIsCode_PlainError(t *testing.T) {
	// Non-StatusError, non-gRPC error
	err := errors.New("plain error")
	assert.False(t, IsCode(err, codes.NotFound))
}

func TestStatusError_Error_AllFields(t *testing.T) {
	se := NewStatusError(codes.Internal, "failed").
		WithInternalMsg("db timeout").
		WithCause(errors.New("connection refused"))
	msg := se.Error()
	assert.Contains(t, msg, "Internal")
	assert.Contains(t, msg, "failed")
	assert.Contains(t, msg, "db timeout")
	assert.Contains(t, msg, "connection refused")
}

func TestIsCode_WrappedStatusError(t *testing.T) {
	se := NewStatusError(codes.PermissionDenied, "denied")
	wrapped := WrapError(se, "extra")
	assert.True(t, IsCode(wrapped, codes.PermissionDenied))
}
