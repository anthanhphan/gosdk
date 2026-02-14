// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/core/mocks"
	"go.uber.org/mock/gomock"
)

func TestBeforeAndAfter_Combined(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCtx := mocks.NewMockContext(ctrl)
	mockCtx.EXPECT().Next().Return(nil)

	var executionOrder []string

	beforeFn := func(_ core.Context) {
		executionOrder = append(executionOrder, "before")
	}

	mw := func(ctx core.Context) error {
		executionOrder = append(executionOrder, "middleware")
		return ctx.Next()
	}

	afterFn := func(_ core.Context, _ error) {
		executionOrder = append(executionOrder, "after")
	}

	wrapped := Before(After(mw, afterFn), beforeFn)
	_ = wrapped(mockCtx)

	expected := []string{"before", "middleware", "after"}
	if len(executionOrder) != len(expected) {
		t.Errorf("execution order length = %d, want %d", len(executionOrder), len(expected))
	}

	for i, step := range expected {
		if i < len(executionOrder) && executionOrder[i] != step {
			t.Errorf("executionOrder[%d] = %s, want %s", i, executionOrder[i], step)
		}
	}
}

func TestSkipForPaths_MultiplePaths(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name        string
		currentPath string
		skipPaths   []string
		shouldApply bool
	}{
		{
			name:        "skip health endpoint",
			currentPath: "/health",
			skipPaths:   []string{"/health", "/metrics", "/ping"},
			shouldApply: false,
		},
		{
			name:        "skip metrics endpoint",
			currentPath: "/metrics",
			skipPaths:   []string{"/health", "/metrics", "/ping"},
			shouldApply: false,
		},
		{
			name:        "apply to API endpoint",
			currentPath: "/api/users",
			skipPaths:   []string{"/health", "/metrics"},
			shouldApply: true,
		},
		{
			name:        "apply to root",
			currentPath: "/",
			skipPaths:   []string{"/health"},
			shouldApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := mocks.NewMockContext(ctrl)
			mockCtx.EXPECT().Path().Return(tt.currentPath)
			mockCtx.EXPECT().Next().Return(nil)

			applied := false
			mw := func(ctx core.Context) error {
				applied = true
				return ctx.Next()
			}

			filtered := SkipForPaths(mw, tt.skipPaths...)
			_ = filtered(mockCtx)

			if applied != tt.shouldApply {
				t.Errorf("middleware applied = %v, want %v", applied, tt.shouldApply)
			}
		})
	}
}

func TestOnlyForMethods_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name           string
		currentMethod  string
		allowedMethods []string
		shouldApply    bool
	}{
		{
			name:           "single allowed method match",
			currentMethod:  "POST",
			allowedMethods: []string{"POST"},
			shouldApply:    true,
		},
		{
			name:           "multiple allowed methods match",
			currentMethod:  "PUT",
			allowedMethods: []string{"POST", "PUT", "PATCH"},
			shouldApply:    true,
		},
		{
			name:           "no match",
			currentMethod:  "GET",
			allowedMethods: []string{"POST", "PUT", "PATCH"},
			shouldApply:    false,
		},
		{
			name:           "case sensitive",
			currentMethod:  "post",
			allowedMethods: []string{"POST"},
			shouldApply:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := mocks.NewMockContext(ctrl)
			mockCtx.EXPECT().Method().Return(tt.currentMethod)
			mockCtx.EXPECT().Next().Return(nil)

			applied := false
			mw := func(ctx core.Context) error {
				applied = true
				return ctx.Next()
			}

			filtered := OnlyForMethods(mw, tt.allowedMethods...)
			_ = filtered(mockCtx)

			if applied != tt.shouldApply {
				t.Errorf("middleware applied = %v, want %v", applied, tt.shouldApply)
			}
		})
	}
}
