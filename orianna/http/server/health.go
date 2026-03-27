// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

//go:generate mockgen -destination=mocks/mock_health.go -package=mocks -mock_names Checker=MockHealthChecker github.com/anthanhphan/gosdk/orianna/shared/health Checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
)

// HTTPChecker checks HTTP endpoint availability.
type HTTPChecker struct {
	client  *http.Client
	url     string
	name    string
	timeout time.Duration
}

// NewHTTPChecker creates a new HTTP endpoint health checker.
func NewHTTPChecker(url string, name string, timeout time.Duration) *HTTPChecker {
	if name == "" {
		name = configuration.DefaultHealthCheckerName
	}
	if timeout == 0 {
		timeout = configuration.DefaultHealthCheckTimeout
	}
	return &HTTPChecker{
		client:  &http.Client{Timeout: timeout},
		url:     url,
		name:    name,
		timeout: timeout,
	}
}

// Check performs the HTTP health check.
func (h *HTTPChecker) Check(ctx context.Context) health.HealthCheck {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return health.HealthCheck{
			Name:         h.name,
			Status:       health.StatusUnhealthy,
			Message:      fmt.Sprintf(core.HealthMessageCreateRequestFailed, err),
			ResponseTime: time.Since(start).Milliseconds(),
			Error:        err,
		}
	}
	req.Header.Set("User-Agent", "orianna-health-checker/1.0")

	resp, err := h.client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return health.HealthCheck{
			Name:         h.name,
			Status:       health.StatusUnhealthy,
			Message:      err.Error(),
			ResponseTime: responseTime.Milliseconds(),
			Error:        err,
			Details:      map[string]any{"url": h.url, "error": err.Error()},
		}
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	status := health.StatusHealthy
	message := core.HealthMessageHealthy
	if resp.StatusCode >= configuration.HealthStatusThresholdServerError {
		status = health.StatusUnhealthy
		message = fmt.Sprintf(core.HealthMessageUnhealthy, resp.StatusCode)
	} else if resp.StatusCode >= configuration.HealthStatusThresholdClientError {
		status = health.StatusDegraded
		message = fmt.Sprintf(core.HealthMessageDegraded, resp.StatusCode)
	}

	return health.HealthCheck{
		Name:         h.name,
		Status:       status,
		Message:      message,
		ResponseTime: responseTime.Milliseconds(),
		Details:      map[string]any{"url": h.url, "status_code": resp.StatusCode},
	}
}

// Name returns the name of this health checker.
func (h *HTTPChecker) Name() string { return h.name }
