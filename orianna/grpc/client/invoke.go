// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"math"
	randv2 "math/rand/v2"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	sharederrors "github.com/anthanhphan/gosdk/orianna/shared/errors"
	"github.com/anthanhphan/gosdk/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Invoke performs a unary gRPC call with observability and optional retry.
func (c *Client) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	// Apply circuit breaker if configured
	var circuitResult bool
	if c.circuit != nil {
		if !c.circuit.Allow() {
			c.logger.Warnw("circuit breaker open, gRPC call rejected",
				"method", method,
			)
			return sharederrors.ErrCircuitOpen
		}
		defer func() {
			c.circuit.RecordResult(circuitResult)
		}()
	}

	// Track in-flight requests
	if c.metrics != nil {
		c.metrics.GaugeInc(ctx, c.metricInFlight)
		defer c.metrics.GaugeDec(ctx, c.metricInFlight)
	}

	// Extract trace context from incoming context
	ctx = c.extractTraceContext(ctx)

	// Start tracing span
	ctx, span := c.startInvokeSpan(ctx, method)

	start := time.Now()

	// Perform the call with retry
	err := c.invokeWithRetry(ctx, method, args, reply, opts)

	duration := time.Since(start)

	// Record observability
	c.recordInvokeMetrics(ctx, method, err, start)
	c.recordInvokeSpan(span, err)

	// Log and record circuit result
	if err != nil {
		circuitResult = false
		c.logger.Warnw("gRPC call failed",
			"method", method,
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
		return err
	}

	circuitResult = true
	c.logger.Debugw("gRPC call succeeded",
		"method", method,
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

// startInvokeSpan starts a tracing span for a unary gRPC call if tracing is configured.
func (c *Client) startInvokeSpan(ctx context.Context, method string) (context.Context, tracing.Span) {
	if c.tracing == nil {
		return ctx, nil
	}
	return c.tracing.StartSpan(ctx, method,
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("rpc.service", c.serviceName),
		),
	)
}

// invokeWithRetry performs the gRPC call with retry logic.
func (c *Client) invokeWithRetry(ctx context.Context, method string, args any, reply any, opts []grpc.CallOption) error {
	maxAttempts := 1
	if c.retry != nil && c.retry.MaxAttempts > 1 {
		maxAttempts = c.retry.MaxAttempts
	}

	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = c.conn.Invoke(ctx, method, args, reply, opts...)
		if err == nil {
			return nil
		}

		// Check if we should retry
		if attempt < maxAttempts-1 && c.retry != nil && c.isRetryable(err) {
			if waitErr := c.waitForRetry(ctx, method, attempt, maxAttempts); waitErr != nil {
				return waitErr
			}
			continue
		}
		break
	}

	return err
}

// waitForRetry waits for the calculated backoff duration or until the context is cancelled.
func (c *Client) waitForRetry(ctx context.Context, method string, attempt, maxAttempts int) error {
	backoff := c.calculateBackoff(attempt)
	c.logger.Warnw("gRPC call failed, retrying",
		"method", method,
		"attempt", attempt+1,
		"max_attempts", maxAttempts,
		"backoff_ms", backoff.Milliseconds(),
	)
	timer := time.NewTimer(backoff)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// recordInvokeMetrics records gRPC call metrics if metrics client is configured.
func (c *Client) recordInvokeMetrics(ctx context.Context, method string, err error, start time.Time) {
	if c.metrics == nil {
		return
	}
	st, _ := status.FromError(err)
	codeStr := st.Code().String()

	c.metrics.Inc(ctx, c.metricRequestsTotal,
		"method", method,
		"code", codeStr,
	)

	c.metrics.Duration(ctx, c.metricRequestDuration, start,
		"method", method,
		"code", codeStr,
	)
}

// recordInvokeSpan records the gRPC call result in the tracing span.
func (c *Client) recordInvokeSpan(span tracing.Span, err error) {
	if span == nil {
		return
	}
	st, _ := status.FromError(err)
	span.SetAttributes(attribute.String("rpc.grpc.status_code", st.Code().String()))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, st.Message())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// =============================================================================
// Retry Logic
// =============================================================================

// isRetryable checks if the error is retryable based on status code.
// Uses pre-built map for O(1) lookup instead of O(n) linear scan.
func (c *Client) isRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	_, retryable := c.retryStatusSet[int(st.Code())]
	return retryable
}

// calculateBackoff computes exponential backoff with jitter.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.retry.InitialBackoff) * math.Pow(c.retry.Multiplier, float64(attempt))
	if backoff > float64(c.retry.MaxBackoff) {
		backoff = float64(c.retry.MaxBackoff)
	}
	// Add 0-25% jitter (not security-sensitive, math/rand is fine for backoff)
	jitter := backoff * 0.25 * randv2.Float64() // #nosec G404
	return time.Duration(backoff + jitter)
}

// =============================================================================
// Streaming
// =============================================================================

// NewStream creates a new streaming gRPC call with observability.
func (c *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// Extract trace context
	ctx = c.extractTraceContext(ctx)

	// Start tracing span
	var span tracing.Span
	if c.tracing != nil {
		var spanCtx context.Context
		spanCtx, span = c.tracing.StartSpan(ctx, method,
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", method),
				attribute.String("rpc.service", c.serviceName),
				attribute.Bool("rpc.grpc.is_client_stream", desc.ClientStreams),
				attribute.Bool("rpc.grpc.is_server_stream", desc.ServerStreams),
			),
		)
		ctx = spanCtx
	}

	stream, err := c.conn.NewStream(ctx, desc, method, opts...)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
		}
		return nil, err
	}

	// Wrap the stream with observability
	wrappedStream := &observedClientStream{
		ClientStream: stream,
		span:         span,
		metrics:      c.metrics,
		serviceName:  c.serviceName,
		method:       method,
	}

	return wrappedStream, nil
}

// extractTraceContext extracts trace context from the given context.
func (c *Client) extractTraceContext(ctx context.Context) context.Context {
	if c.tracing == nil {
		return ctx
	}
	return tracing.ExtractContext(ctx, &core.GRPCMetadataCarrier{})
}

// observedClientStream wraps a gRPC stream for observability.
type observedClientStream struct {
	grpc.ClientStream
	span        tracing.Span
	metrics     metrics.Client
	serviceName string
	method      string
}

func (o *observedClientStream) CloseSend() error {
	err := o.ClientStream.CloseSend()

	if o.span != nil {
		if err != nil {
			o.span.RecordError(err)
			o.span.SetStatus(codes.Error, err.Error())
		}
		o.span.End()
	}

	return err
}
