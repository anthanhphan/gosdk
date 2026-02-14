// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/anthanhphan/gosdk/metrics/mocks"
	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	coremocks "github.com/anthanhphan/gosdk/orianna/pkg/core/mocks"
	"go.uber.org/mock/gomock"
)

// MetricsMiddleware Tests

func TestMetricsMiddleware(t *testing.T) {
	t.Run("records request count and duration on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockClient := mocks.NewMockClient(ctrl)

		// Setup context behavior
		mockCtx.EXPECT().Next().Return(nil)
		mockCtx.EXPECT().ResponseStatusCode().Return(200)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("GET").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/users").AnyTimes()

		// Expect Inc to be called with correct metric name and tags
		mockClient.EXPECT().Inc(
			gomock.Any(),         // ctx
			"api_requests_total", // name
			"method", "GET",      // tags
			"path", "/api/users",
			"status", "200",
			"error_code", "none",
		)

		// Expect Duration to be called with correct metric name
		mockClient.EXPECT().Duration(
			gomock.Any(),                   // ctx
			"api_request_duration_seconds", // name
			gomock.Any(),                   // start time
			"method", "GET",                // tags
			"path", "/api/users",
			"status", "200",
			"error_code", "none",
		)

		mw := MetricsMiddleware(mockClient, "api")
		err := mw(mockCtx)
		if err != nil {
			t.Errorf("MetricsMiddleware() error = %v, want nil", err)
		}
	})

	t.Run("extracts error code from ErrorResponse", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockClient := mocks.NewMockClient(ctrl)

		errResp := core.NewErrorResponse("VALIDATION_ERROR", 400, "Invalid input")

		mockCtx.EXPECT().Next().Return(errResp)
		mockCtx.EXPECT().ResponseStatusCode().Return(400)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("POST").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/users").AnyTimes()

		// Should use the error code from ErrorResponse
		mockClient.EXPECT().Inc(
			gomock.Any(),
			"svc_requests_total",
			"method", "POST",
			"path", "/api/users",
			"status", "400",
			"error_code", "VALIDATION_ERROR",
		)

		mockClient.EXPECT().Duration(
			gomock.Any(),
			"svc_request_duration_seconds",
			gomock.Any(),
			"method", "POST",
			"path", "/api/users",
			"status", "400",
			"error_code", "VALIDATION_ERROR",
		)

		mw := MetricsMiddleware(mockClient, "svc")
		err := mw(mockCtx)
		// The error should be propagated
		if err == nil {
			t.Error("MetricsMiddleware() should propagate the error")
		}
	})

	t.Run("uses INTERNAL_ERROR for non-ErrorResponse errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockClient := mocks.NewMockClient(ctrl)

		genericErr := errors.New("something went wrong")

		mockCtx.EXPECT().Next().Return(genericErr)
		mockCtx.EXPECT().ResponseStatusCode().Return(500)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("GET").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/items").AnyTimes()

		// Should use INTERNAL_ERROR for generic errors
		mockClient.EXPECT().Inc(
			gomock.Any(),
			"app_requests_total",
			"method", "GET",
			"path", "/api/items",
			"status", "500",
			"error_code", "INTERNAL_ERROR",
		)

		mockClient.EXPECT().Duration(
			gomock.Any(),
			"app_request_duration_seconds",
			gomock.Any(),
			"method", "GET",
			"path", "/api/items",
			"status", "500",
			"error_code", "INTERNAL_ERROR",
		)

		mw := MetricsMiddleware(mockClient, "app")
		err := mw(mockCtx)
		if !errors.Is(err, genericErr) {
			t.Errorf("MetricsMiddleware() error = %v, want %v", err, genericErr)
		}
	})

	t.Run("metric names use subsystem prefix", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockClient := mocks.NewMockClient(ctrl)

		mockCtx.EXPECT().Next().Return(nil)
		mockCtx.EXPECT().ResponseStatusCode().Return(204)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("DELETE").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/resources/:id").AnyTimes()

		// Verify the subsystem prefix is properly used
		mockClient.EXPECT().Inc(
			gomock.Any(),
			"custom_prefix_requests_total", // subsystem + _requests_total
			gomock.Any(), gomock.Any(),     // method tags
			gomock.Any(), gomock.Any(), // path tags
			gomock.Any(), gomock.Any(), // status tags
			gomock.Any(), gomock.Any(), // error_code tags
		)

		mockClient.EXPECT().Duration(
			gomock.Any(),
			"custom_prefix_request_duration_seconds", // subsystem + _request_duration_seconds
			gomock.Any(),                             // start time
			gomock.Any(), gomock.Any(),               // method tags
			gomock.Any(), gomock.Any(), // path tags
			gomock.Any(), gomock.Any(), // status tags
			gomock.Any(), gomock.Any(), // error_code tags
		)

		mw := MetricsMiddleware(mockClient, "custom_prefix")
		err := mw(mockCtx)
		if err != nil {
			t.Errorf("MetricsMiddleware() error = %v, want nil", err)
		}
	})
}
