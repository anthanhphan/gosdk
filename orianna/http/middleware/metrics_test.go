// Copyright (c) 2026 anthanhphan <can.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/anthanhphan/gosdk/metrics/mocks"
	coremocks "github.com/anthanhphan/gosdk/orianna/http/core/mocks"
	"github.com/anthanhphan/gosdk/orianna/shared/httputil"
	"go.uber.org/mock/gomock"
)

// MetricsMiddleware Tests

func TestMetricsMiddleware(t *testing.T) {
	t.Run("records request count and duration on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockClient := mocks.NewMockClient(ctrl)

		// Setup context behavior
		mockCtx.EXPECT().Next().Return(nil)
		mockCtx.EXPECT().ResponseStatusCode().Return(200)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("GET").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/users").AnyTimes()

		// Expect in-flight gauge calls
		mockClient.EXPECT().GaugeInc(gomock.Any(), "api_in_flight_requests")
		mockClient.EXPECT().GaugeDec(gomock.Any(), "api_in_flight_requests")

		// Expect Inc to be called with correct metric name and tags
		mockClient.EXPECT().Inc(
			gomock.Any(),         // ctx
			"api_requests_total", // name
			"method", "GET",      // tags
			"path", "/api/users",
			"status", "200",
			"error_class", "none",
		)

		// Expect Duration to be called with correct metric name
		mockClient.EXPECT().Duration(
			gomock.Any(),                   // ctx
			"api_request_duration_seconds", // name
			gomock.Any(),                   // start time
			"method", "GET",                // tags
			"path", "/api/users",
			"status", "200",
			"error_class", "none",
		)

		mw := MetricsMiddleware(mockClient, "api")
		err := mw(mockCtx)
		if err != nil {
			t.Errorf("MetricsMiddleware() error = %v, want nil", err)
		}
	})

	t.Run("uses client_error class for 4xx status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockClient := mocks.NewMockClient(ctrl)

		mockCtx.EXPECT().Next().Return(nil)
		mockCtx.EXPECT().ResponseStatusCode().Return(400)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("POST").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/users").AnyTimes()

		// Expect in-flight gauge calls
		mockClient.EXPECT().GaugeInc(gomock.Any(), "svc_in_flight_requests")
		mockClient.EXPECT().GaugeDec(gomock.Any(), "svc_in_flight_requests")

		// Should use "client_error" for 4xx status codes
		mockClient.EXPECT().Inc(
			gomock.Any(),
			"svc_requests_total",
			"method", "POST",
			"path", "/api/users",
			"status", "400",
			"error_class", "client_error",
		)

		mockClient.EXPECT().Duration(
			gomock.Any(),
			"svc_request_duration_seconds",
			gomock.Any(),
			"method", "POST",
			"path", "/api/users",
			"status", "400",
			"error_class", "client_error",
		)

		mw := MetricsMiddleware(mockClient, "svc")
		_ = mw(mockCtx)
	})

	t.Run("uses server_error class for 5xx status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockCtx := coremocks.NewMockContext(ctrl)
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockClient := mocks.NewMockClient(ctrl)

		genericErr := errors.New("something went wrong")

		mockCtx.EXPECT().Next().Return(genericErr)
		mockCtx.EXPECT().ResponseStatusCode().Return(500)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("GET").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/items").AnyTimes()

		// Expect in-flight gauge calls
		mockClient.EXPECT().GaugeInc(gomock.Any(), "app_in_flight_requests")
		mockClient.EXPECT().GaugeDec(gomock.Any(), "app_in_flight_requests")

		// Should use "server_error" for 5xx status codes
		mockClient.EXPECT().Inc(
			gomock.Any(),
			"app_requests_total",
			"method", "GET",
			"path", "/api/items",
			"status", "500",
			"error_class", "server_error",
		)

		mockClient.EXPECT().Duration(
			gomock.Any(),
			"app_request_duration_seconds",
			gomock.Any(),
			"method", "GET",
			"path", "/api/items",
			"status", "500",
			"error_class", "server_error",
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
		mockCtx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
		mockClient := mocks.NewMockClient(ctrl)

		mockCtx.EXPECT().Next().Return(nil)
		mockCtx.EXPECT().ResponseStatusCode().Return(204)
		mockCtx.EXPECT().Context().Return(context.Background()).AnyTimes()
		mockCtx.EXPECT().Method().Return("DELETE").AnyTimes()
		mockCtx.EXPECT().RoutePath().Return("/api/resources/:id").AnyTimes()

		// Expect in-flight gauge calls
		mockClient.EXPECT().GaugeInc(gomock.Any(), "custom_prefix_in_flight_requests")
		mockClient.EXPECT().GaugeDec(gomock.Any(), "custom_prefix_in_flight_requests")

		// Verify the subsystem prefix is properly used
		mockClient.EXPECT().Inc(
			gomock.Any(),
			"custom_prefix_requests_total", // subsystem + _requests_total
			gomock.Any(), gomock.Any(),     // method tags
			gomock.Any(), gomock.Any(), // path tags
			gomock.Any(), gomock.Any(), // status tags
			gomock.Any(), gomock.Any(), // error_class tags
		)

		mockClient.EXPECT().Duration(
			gomock.Any(),
			"custom_prefix_request_duration_seconds", // subsystem + _request_duration_seconds
			gomock.Any(),                             // start time
			gomock.Any(), gomock.Any(),               // method tags
			gomock.Any(), gomock.Any(), // path tags
			gomock.Any(), gomock.Any(), // status tags
			gomock.Any(), gomock.Any(), // error_class tags
		)

		mw := MetricsMiddleware(mockClient, "custom_prefix")
		err := mw(mockCtx)
		if err != nil {
			t.Errorf("MetricsMiddleware() error = %v, want nil", err)
		}
	})
}

func TestErrorClassFromStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       string
	}{
		{"200 OK", 200, "none"},
		{"201 Created", 201, "none"},
		{"204 No Content", 204, "none"},
		{"301 Redirect", 301, "none"},
		{"400 Bad Request", 400, "client_error"},
		{"401 Unauthorized", 401, "client_error"},
		{"404 Not Found", 404, "client_error"},
		{"429 Too Many Requests", 429, "client_error"},
		{"500 Internal Server Error", 500, "server_error"},
		{"502 Bad Gateway", 502, "server_error"},
		{"503 Service Unavailable", 503, "server_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := httputil.ErrorClassFromStatus(tt.statusCode)
			if got != tt.want {
				t.Errorf("ErrorClassFromStatus(%d) = %q, want %q", tt.statusCode, got, tt.want)
			}
		})
	}
}
