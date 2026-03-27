// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"log"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/grpc/client"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"
	"github.com/anthanhphan/gosdk/tracing"

	userpb "github.com/anthanhphan/gosdk/orianna/docs/examples/grpc/proto"
)

// Orianna gRPC Client Example
//
// This example demonstrates how to use the Orianna gRPC client with:
// - Built-in tracing (distributed tracing for unary and stream calls)
// - Metrics (Prometheus-compatible)
// - Retry logic with exponential backoff
// - Circuit breaker pattern
// - TLS/mTLS support

func main() {
	// Initialize logger
	undo := logger.InitLogger(&logger.Config{
		LogLevel:    logger.LevelDebug,
		LogEncoding: logger.EncodingJSON,
	})
	defer undo()

	// Create a metrics client (Prometheus)
	metricsClient := metrics.NewClient("example-grpc-client")

	// Create a tracing client (OpenTelemetry)
	tracingClient, err := tracing.NewClient("example-grpc-client",
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

	// Create gRPC client with all options at once
	grpcCli, err := client.NewClient(
		client.WithAddress("localhost:50051"),
		client.WithServiceName("user-service"),
		client.WithMetrics(metricsClient),
		client.WithTracing(tracingClient),
		// client.WithTLS(&client.TLSConfig{
		// 	CertFile: "cert.pem",
		// 	KeyFile:  "key.pem",
		// 	CAFile:   "ca.pem",
		// }),
	)
	if err != nil {
		logger.Fatalw("Failed to create gRPC client", "error", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the underlying gRPC connection
	conn := grpcCli.Connection()
	defer conn.Close()

	// Create a user service client
	userClient := userpb.NewUserServiceClient(conn)

	// Example 1: Get a single user (unary call)
	logger.Info("Getting user with ID 1...")
	getResp, err := userClient.GetUser(ctx, &userpb.GetUserRequest{Id: 1})
	if err != nil {
		logger.Errorw("GetUser failed", "error", err)
	} else {
		logger.Infow("GetUser succeeded",
			"user_id", getResp.User.Id,
			"name", getResp.User.Name,
			"email", getResp.User.Email,
		)
	}

	// Example 2: List users (unary call with pagination)
	logger.Info("Listing users...")
	listResp, err := userClient.ListUsers(ctx, &userpb.ListUsersRequest{
		Page:   1,
		Limit:  10,
		Active: true,
	})
	if err != nil {
		logger.Errorw("ListUsers failed", "error", err)
	} else {
		logger.Infow("ListUsers succeeded",
			"user_count", len(listResp.Users),
			"total", listResp.Pagination.Total,
		)
	}

	// Example 3: Create a new user (unary call)
	logger.Info("Creating a new user...")
	createResp, err := userClient.CreateUser(ctx, &userpb.CreateUserRequest{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
		Role:  "user",
	})
	if err != nil {
		logger.Errorw("CreateUser failed", "error", err)
	} else {
		logger.Infow("CreateUser succeeded",
			"user_id", createResp.User.Id,
			"name", createResp.User.Name,
		)
	}

	// Example 4: Update a user (unary call)
	logger.Info("Updating user...")
	updateResp, err := userClient.UpdateUser(ctx, &userpb.UpdateUserRequest{
		Id:    1,
		Name:  "Jane Doe",
		Email: "jane.doe@example.com",
		Age:   25,
	})
	if err != nil {
		logger.Errorw("UpdateUser failed", "error", err)
	} else {
		logger.Infow("UpdateUser succeeded",
			"user_id", updateResp.User.Id,
			"name", updateResp.User.Name,
		)
	}

	// Example 5: Delete a user (unary call)
	logger.Info("Deleting user...")
	deleteResp, err := userClient.DeleteUser(ctx, &userpb.DeleteUserRequest{Id: 1})
	if err != nil {
		logger.Errorw("DeleteUser failed", "error", err)
	} else {
		logger.Infow("DeleteUser succeeded",
			"success", deleteResp.Success,
			"message", deleteResp.Message,
		)
	}

	// Example 6: Server streaming - receive multiple users
	logger.Info("Starting server stream...")
	stream, err := userClient.StreamUsers(ctx, &userpb.StreamUsersRequest{BatchSize: 5})
	if err != nil {
		logger.Errorw("StreamUsers failed", "error", err)
	} else {
		for {
			user, err := stream.Recv()
			if err != nil {
				break
			}
			logger.Infow("Received user from stream",
				"user_id", user.Id,
				"name", user.Name,
			)
		}
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
		cb.RecordResult(i > 2)
	}

	logger.Info("All examples completed!")
}
