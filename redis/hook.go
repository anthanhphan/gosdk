// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/prometheus/client_golang/prometheus"
	goredis "github.com/redis/go-redis/v9"
)

// metricsHook implements goredis.Hook.
// It records per-command latency and error metrics, and logs errors via the gosdk logger.
// The "action" label is extracted from context using ActionFromContext.
type metricsHook struct {
	log        *logger.Logger
	duration   *prometheus.HistogramVec
	errorCount *prometheus.CounterVec
}

var _ goredis.Hook = (*metricsHook)(nil)

// newMetricsHook registers Prometheus metrics on the default registry.
// If the same metrics are already registered (e.g. in tests with -count>1),
// the existing collectors are reused via AlreadyRegisteredError.
func newMetricsHook(log *logger.Logger, namespace, subsystem string) *metricsHook {
	duration := mustOrExisting(prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "command_duration_seconds",
			Help:      "Duration of Redis commands in seconds.",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"action", "command"},
	)).(*prometheus.HistogramVec)

	errorCount := mustOrExisting(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "command_errors_total",
			Help:      "Total Redis command errors.",
		},
		[]string{"action", "command", "error_type"},
	)).(*prometheus.CounterVec)

	return &metricsHook{log: log, duration: duration, errorCount: errorCount}
}

// mustOrExisting registers c on the default Prometheus registry.
// If c is already registered, the existing collector is returned instead of panicking.
func mustOrExisting(c prometheus.Collector) prometheus.Collector {
	if err := prometheus.Register(c); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return are.ExistingCollector
		}
		panic(err)
	}
	return c
}

func (h *metricsHook) DialHook(next goredis.DialHook) goredis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *metricsHook) ProcessHook(next goredis.ProcessHook) goredis.ProcessHook {
	return func(ctx context.Context, cmd goredis.Cmder) error {
		action := actionFromCtx(ctx)
		start := time.Now()

		err := next(ctx, cmd)

		h.duration.WithLabelValues(action, cmd.Name()).Observe(time.Since(start).Seconds())

		if err != nil && err != goredis.Nil {
			errType := classifyError(err)
			h.errorCount.WithLabelValues(action, cmd.Name(), errType).Inc()
			h.log.Errorw("redis: command error",
				"action", action,
				"command", cmd.Name(),
				"error_type", errType,
				"error", err.Error(),
			)
		}

		return err
	}
}

func (h *metricsHook) ProcessPipelineHook(next goredis.ProcessPipelineHook) goredis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []goredis.Cmder) error {
		action := actionFromCtx(ctx)
		start := time.Now()

		err := next(ctx, cmds)

		elapsed := time.Since(start).Seconds()
		perCmd := elapsed / float64(len(cmds))

		for _, cmd := range cmds {
			h.duration.WithLabelValues(action, cmd.Name()).Observe(perCmd)

			if cmdErr := cmd.Err(); cmdErr != nil && cmdErr != goredis.Nil {
				errType := classifyError(cmdErr)
				h.errorCount.WithLabelValues(action, cmd.Name(), errType).Inc()
				h.log.Errorw("redis: pipeline command error",
					"action", action,
					"command", cmd.Name(),
					"error_type", errType,
					"error", cmdErr.Error(),
				)
			}
		}

		if err != nil {
			h.log.Errorw("redis: pipeline error",
				"action", action,
				"pipeline_size", len(cmds),
				"error", err.Error(),
			)
		}

		return err
	}
}

// actionFromCtx returns the action label from context, defaulting to "unknown".
func actionFromCtx(ctx context.Context) string {
	if a := ActionFromContext(ctx); a != "" {
		return a
	}
	return "unknown"
}

// classifyError returns a short error type label for Prometheus.
func classifyError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "context canceled") || strings.Contains(msg, "context deadline exceeded"):
		return "context"
	case strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "dial tcp"):
		return "connection"
	default:
		return "other"
	}
}
