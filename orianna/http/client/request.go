// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"time"

	"github.com/anthanhphan/gosdk/jcodec"
	sharederrors "github.com/anthanhphan/gosdk/orianna/shared/errors"
	"github.com/anthanhphan/gosdk/tracing"
)

// Request represents an HTTP request.
type Request struct {
	Method  string
	Path    string
	Query   map[string]string
	Headers map[string]string
	Body    any
}

// RequestOption configures a request.
type RequestOption func(*Request)

// WithQuery adds query parameters.
func WithQuery(key, value string) RequestOption {
	return func(r *Request) {
		if r.Query == nil {
			r.Query = make(map[string]string)
		}
		r.Query[key] = value
	}
}

// WithHeader adds a header to the request.
func WithHeader(key, value string) RequestOption {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers[key] = value
	}
}

// WithAuth adds Authorization header.
func WithAuth(token string) RequestOption {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers["Authorization"] = token
	}
}

// =============================================================================
// Request Execution
// =============================================================================

// Do performs an HTTP request with full observability.
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// Apply circuit breaker if configured
	var circuitResult bool
	if c.circuit != nil {
		if !c.circuit.Allow() {
			c.logger.Warnw("circuit breaker open, request rejected",
				"method", req.Method,
				"path", req.Path,
			)
			return nil, sharederrors.ErrCircuitOpen
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

	// Marshal body once before retry loop to avoid re-serialization per attempt
	bodyBytes, err := c.marshalBody(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.executeWithRetry(ctx, req, bodyBytes)
	if resp != nil {
		circuitResult = resp.StatusCode < 500
	}
	return resp, err
}

// marshalBody serializes the request body if present.
func (c *Client) marshalBody(req *Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	data, err := jcodec.Marshal(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return data, nil
}

// executeWithRetry performs the HTTP request with retry logic.
func (c *Client) executeWithRetry(ctx context.Context, req *Request, bodyBytes []byte) (*Response, error) {
	maxAttempts := 1
	if c.retry != nil {
		maxAttempts = c.retry.MaxAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := c.doRequest(ctx, req, bodyBytes, attempt)

		if err == nil {
			if attempt > 1 {
				c.logger.Infow("request succeeded after retry",
					"method", req.Method, "path", req.Path, "attempts", attempt,
				)
			}
			if c.retry != nil && c.isRetryable(req.Method, resp, nil) {
				lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				if waitErr := c.waitForRetry(ctx, req, resp.StatusCode, attempt, maxAttempts); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return resp, nil
		}

		lastErr = err

		if c.retry == nil || !c.isRetryable(req.Method, resp, err) {
			break
		}

		if waitErr := c.waitForRetry(ctx, req, 0, attempt, maxAttempts); waitErr != nil {
			return nil, waitErr
		}
	}

	return nil, lastErr
}

// doRequest performs a single HTTP request with tracing.
// bodyBytes is the pre-marshaled request body (nil if no body).
func (c *Client) doRequest(ctx context.Context, req *Request, bodyBytes []byte, attempt int) (*Response, error) {
	reqURL := c.buildURL(req)
	httpReq, err := c.buildRequest(ctx, req, reqURL, bodyBytes)
	if err != nil {
		return nil, err
	}

	httpReq, span := c.startTracing(httpReq, req, reqURL, attempt)

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)

	if c.metrics != nil {
		c.recordMetrics(ctx, req, attempt, httpResp, start)
	}

	if err != nil {
		c.handleSpanError(span, err)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	respBody, err := c.readResponseBody(httpResp, span)
	if err != nil {
		return nil, err
	}

	c.endSpan(span, httpResp)

	c.logger.Debugw("HTTP request completed",
		"method", req.Method,
		"path", req.Path,
		"status", httpResp.StatusCode,
		"duration_ms", duration.Milliseconds(),
		"attempt", attempt,
	)

	return &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       respBody,
	}, nil
}

// buildURL builds the full URL with properly encoded query parameters.
// Uses net/url.Values for RFC 3986 compliant encoding.
func (c *Client) buildURL(req *Request) string {
	fullURL := c.baseURL + req.Path
	if len(req.Query) > 0 {
		vals := make(url.Values, len(req.Query))
		for k, v := range req.Query {
			vals.Set(k, v)
		}
		fullURL += "?" + vals.Encode()
	}
	return fullURL
}

// buildRequest builds the HTTP request with pre-marshaled body and headers.
func (c *Client) buildRequest(ctx context.Context, req *Request, reqURL string, bodyBytes []byte) (*http.Request, error) {
	var body io.Reader
	if bodyBytes != nil {
		body = bytes.NewReader(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	if bodyBytes != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	return httpReq, nil
}

// readResponseBody reads and returns the response body.
// It always drains and closes the body to enable HTTP connection reuse.
// Limits read size to MaxResponseBodySize (default 10MB) to prevent OOM
// from malicious or buggy upstream servers.
func (c *Client) readResponseBody(httpResp *http.Response, span tracing.Span) ([]byte, error) {
	defer func() {
		_ = httpResp.Body.Close()
	}()

	maxSize := c.config.MaxResponseBodySize
	if maxSize <= 0 {
		maxSize = DefaultMaxResponseBodySize
	}

	limitedReader := io.LimitReader(httpResp.Body, int64(maxSize))
	result, err := io.ReadAll(limitedReader)
	if err != nil {
		c.handleSpanError(span, err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return result, nil
}

// =============================================================================
// Convenience Methods
// =============================================================================

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method: http.MethodGet,
		Path:   path,
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method: http.MethodPatch,
		Path:   path,
		Body:   body,
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method: http.MethodDelete,
		Path:   path,
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}
