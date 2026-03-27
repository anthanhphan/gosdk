// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"log"
	"time"

	pb "github.com/anthanhphan/gosdk/orianna/docs/examples/grpc/proto"
	"github.com/anthanhphan/gosdk/orianna/grpc/configuration"
	"github.com/anthanhphan/gosdk/orianna/grpc/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ===========================================================================
// Orianna gRPC Quickstart - Minimal example to get started quickly
// ===========================================================================
//
// This example demonstrates:
// - Basic gRPC server setup and configuration
// - Simple unary RPC implementation
// - Request validation and error handling
// - Logging interceptor
// - Health checks
// - Graceful shutdown
//
// Prerequisites:
//   - Generate proto files: cd ../proto && make generate
//
// Run:  go run main.go
// Test: grpcurl -plaintext localhost:50051 list
//       grpcurl -plaintext localhost:50051 user.UserService/ListUsers
//
// ===========================================================================

func main() {
	// -------------------------------------------------------------------------
	// Step 1: Create gRPC server with basic configuration
	// -------------------------------------------------------------------------
	srv, err := server.NewServer(&configuration.Config{
		ServiceName:    "grpc-quickstart",
		Port:           50051,
		VerboseLogging: true, // Enable detailed request logging
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// -------------------------------------------------------------------------
	// Step 2: Register service implementation
	// -------------------------------------------------------------------------
	pb.RegisterUserServiceServer(srv.GRPCServer(), &userService{})

	// -------------------------------------------------------------------------
	// Step 3: Start server
	// -------------------------------------------------------------------------
	log.Printf("gRPC server starting on localhost:50051")
	log.Println("Try these commands:")
	log.Println("  grpcurl -plaintext localhost:50051 list")
	log.Println("  grpcurl -plaintext localhost:50051 user.UserService/ListUsers")
	log.Println("  grpcurl -plaintext -d '{\"id\":1}' localhost:50051 user.UserService/GetUser")
	log.Println("  grpcurl -plaintext -d '{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"age\":25,\"role\":\"user\"}' localhost:50051 user.UserService/CreateUser")

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ===========================================================================
// Service Implementation
// ===========================================================================

type userService struct {
	pb.UnimplementedUserServiceServer
}

// GetUser retrieves a single user by ID
func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Validate request
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	// Simulate user lookup
	user := &pb.User{
		Id:        req.Id,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       25,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return &pb.GetUserResponse{User: user}, nil
}

// ListUsers returns a paginated list of users
func (s *userService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Set defaults
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Simulate user data
	users := []*pb.User{
		{
			Id:        1,
			Name:      "Alice",
			Email:     "alice@example.com",
			Age:       25,
			Role:      "admin",
			Active:    true,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
		{
			Id:        2,
			Name:      "Bob",
			Email:     "bob@example.com",
			Age:       30,
			Role:      "user",
			Active:    true,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
		{
			Id:        3,
			Name:      "Charlie",
			Email:     "charlie@example.com",
			Age:       28,
			Role:      "user",
			Active:    false,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
	}

	var total int32
	for range users {
		total++
	}

	return &pb.ListUsersResponse{
		Users: users,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Age <= 0 || req.Age > 150 {
		return nil, status.Error(codes.InvalidArgument, "age must be between 1 and 150")
	}

	// Create user
	user := &pb.User{
		Id:        123, // Simulated ID
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		Role:      req.Role,
		Active:    true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return &pb.CreateUserResponse{User: user}, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Validate request
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Update user
	user := &pb.User{
		Id:        req.Id,
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		Active:    true,
		CreatedAt: time.Now().Add(-24 * time.Hour).Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return &pb.UpdateUserResponse{User: user}, nil
}

// DeleteUser deletes a user by ID
func (s *userService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// Validate request
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	log.Printf("Deleted user %d", req.Id)

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// Streaming methods - Not implemented in quickstart
// See complete example for streaming implementations

func (s *userService) StreamUsers(req *pb.StreamUsersRequest, stream pb.UserService_StreamUsersServer) error {
	return status.Error(codes.Unimplemented, "see complete example for streaming")
}

func (s *userService) BatchCreateUsers(stream pb.UserService_BatchCreateUsersServer) error {
	return status.Error(codes.Unimplemented, "see complete example for streaming")
}

func (s *userService) UserChat(stream pb.UserService_UserChatServer) error {
	return status.Error(codes.Unimplemented, "see complete example for streaming")
}
