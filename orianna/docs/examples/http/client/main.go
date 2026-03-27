// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"log"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/http/client"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"
)

// Orianna HTTP Client Example
//
// This example demonstrates how to use the Orianna HTTP client with:
// - Built-in tracing (distributed tracing)
// - Metrics (Prometheus-compatible)
// - Retry logic with exponential backoff
// - Circuit breaker pattern
// - Sensitive header redaction

func main() {
	// Initialize logger
	undo := logger.InitLogger(&logger.Config{
		LogLevel:    logger.LevelDebug,
		LogEncoding: logger.EncodingJSON,
	})
	defer undo()

	// Create a metrics client (Prometheus)
	metricsClient := metrics.NewClient("example-client")

	// Create a tracing client (OpenTelemetry)
	tracingClient, err := tracing.NewClient("example-client",
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithInsecure(),
	)
	if err != nil {
		log.Printf("Warning: tracing client not available: %v", err)
		tracingClient = nil
	}
	if tracingClient != nil {
		defer tracingClient.Shutdown(context.Background())
	}

	// Create HTTP client with all options at once
	httpClient, err := client.NewClient(
		client.WithBaseURL("https://jsonplaceholder.typicode.com"),
		client.WithTimeout(30*time.Second),
		client.WithDefaultHeader("X-Client-Name", "orianna-demo"),
		client.WithTracing(tracingClient),
		client.WithMetrics(metricsClient),
		client.WithLogger(logger.NewLoggerWithFields(logger.String("package", "http-client"))),
		client.WithRetry(&resilience.RetryConfig{
			MaxAttempts:          3,
			InitialBackoff:       100 * time.Millisecond,
			MaxBackoff:           5 * time.Second,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
		}),
		client.WithCircuitBreaker(&resilience.CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Example 1: Simple GET request
	logger.Info("Making GET request to /posts/1...")
	resp, err := httpClient.Get(ctx, "/posts/1")
	if err != nil {
		logger.Errorw("GET request failed", "error", err)
	} else {
		logger.Infow("GET request succeeded",
			"status", resp.StatusCode,
			"body_length", len(resp.Body),
		)
	}

	// Example 2: GET with query parameters
	logger.Info("Making GET request with query params...")
	resp, err = httpClient.Get(ctx, "/posts", client.WithQuery("userId", "1"))
	if err != nil {
		logger.Errorw("GET with params failed", "error", err)
	} else {
		logger.Infow("GET with params succeeded",
			"status", resp.StatusCode,
			"body_length", len(resp.Body),
		)
	}

	// Example 3: POST request with body
	logger.Info("Making POST request...")
	type Post struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		UserID int    `json:"userId"`
	}
	post := Post{
		Title:  "Test Post",
		Body:   "This is a test post from Orianna HTTP client",
		UserID: 1,
	}
	resp, err = httpClient.Post(ctx, "/posts", post, client.WithAuth("Bearer test-token"))
	if err != nil {
		logger.Errorw("POST request failed", "error", err)
	} else {
		logger.Infow("POST request succeeded",
			"status", resp.StatusCode,
			"body_length", len(resp.Body),
		)
	}

	// Example 4: PUT request
	logger.Info("Making PUT request...")
	post.Title = "Updated Test Post"
	resp, err = httpClient.Put(ctx, "/posts/1", post)
	if err != nil {
		logger.Errorw("PUT request failed", "error", err)
	} else {
		logger.Infow("PUT request succeeded",
			"status", resp.StatusCode,
		)
	}

	// Example 5: DELETE request
	logger.Info("Making DELETE request...")
	resp, err = httpClient.Delete(ctx, "/posts/1")
	if err != nil {
		logger.Errorw("DELETE request failed", "error", err)
	} else {
		logger.Infow("DELETE request succeeded",
			"status", resp.StatusCode,
		)
	}

	// Example 6: Using standalone functions (for simple use cases)
	logger.Info("Using standalone HTTP functions...")
	standaloneResp, err := client.Get(ctx, "https://jsonplaceholder.typicode.com", "/users/1")
	if err != nil {
		logger.Errorw("Standalone GET failed", "error", err)
	} else {
		logger.Infow("Standalone GET succeeded",
			"status", standaloneResp.StatusCode,
		)
	}

	// Example 7: Using circuit breaker directly
	logger.Info("Testing circuit breaker...")
	cb := resilience.NewCircuitBreaker(resilience.DefaultCircuitBreakerConfig())
	for i := 0; i < 10; i++ {
		allowed := cb.Allow()
		state := cb.State()
		logger.Debugw("Circuit breaker check",
			"attempt", i,
			"allowed", allowed,
			"state", state,
		)
		if !allowed {
			break
		}
		// Record some failures to open the circuit
		cb.RecordResult(i > 2) // First 3 fail, rest succeed
	}

	logger.Info("All examples completed!")
}
