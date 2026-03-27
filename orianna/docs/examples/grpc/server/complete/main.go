// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	pb "github.com/anthanhphan/gosdk/orianna/docs/examples/grpc/proto"
	"github.com/anthanhphan/gosdk/orianna/grpc/configuration"
	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/orianna/grpc/server"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Orianna gRPC Complete Example - Full-featured gRPC server
//
// This example demonstrates major Orianna gRPC features including:
// - Server configuration (timeouts, limits, TLS)
// - Unary RPC methods
// - Server streaming RPC
// - Client streaming RPC
// - Bidirectional streaming RPC
// - Interceptors (logging, metrics, auth, recovery)
// - Metadata handling
// - Error handling with status codes
// - Health checks
// - Graceful shutdown

func main() {
	// Initialize logger
	undo := logger.InitLogger(&logger.Config{
		LogLevel:          logger.LevelDebug,
		LogEncoding:       logger.EncodingJSON,
		DisableStacktrace: true,
	})
	defer undo()

	srv := createServer()

	if err := srv.Start(); err != nil {
		logger.Fatalw("Server error", "error", err)
	}
}

// Server Configuration

func createServer() *server.Server {
	connectionTimeout := 120 * time.Second
	keepaliveTime := 2 * time.Hour
	keepaliveTimeout := 20 * time.Second
	shutdownTimeout := 30 * time.Second

	config := &configuration.Config{
		ServiceName:             "grpc-complete",
		Version:                 "1.0.0",
		Port:                    50052,
		VerboseLogging:          true,
		SlowRequestThreshold:    2 * time.Second, // Non-zero → auto-registers SlowRPCDetector
		MaxConcurrentStreams:    256,
		MaxRecvMsgSize:          4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:          4 * 1024 * 1024, // 4MB
		ConnectionTimeout:       &connectionTimeout,
		KeepaliveTime:           &keepaliveTime,
		KeepaliveTimeout:        &keepaliveTimeout,
		GracefulShutdownTimeout: &shutdownTimeout,
		EnableReflection:        true, // Enable gRPC reflection for grpcurl
	}

	srv, err := server.NewServer(config,
		// Metrics
		server.WithMetrics(metrics.NewClient("grpc-complete")),

		// Health checks
		server.WithHealthChecker(
			health.NewCustomChecker("database", databaseHealthCheck),
		),
		server.WithHealthChecker(
			health.NewCustomChecker("cache", cacheHealthCheck),
		),

		// Custom hooks
		server.WithHooks(createHooks()),
	)
	if err != nil {
		logger.Fatalw("Failed to create server", "error", err)
	}

	// Register service
	pb.RegisterUserServiceServer(srv.GRPCServer(), &userService{
		users: make(map[int64]*pb.User),
	})

	return srv
}

func createHooks() *core.Hooks {
	hooks := core.NewHooks()

	hooks.AddOnRequest(func(ctx core.Context) {
		logger.Debugw("gRPC REQUEST", "method", ctx.FullMethod())
	})

	hooks.AddOnResponse(func(ctx core.Context, code string, latency time.Duration) {
		logger.Infow("gRPC RESPONSE",
			"method", ctx.FullMethod(),
			"code", code,
			"latency_ms", latency.Milliseconds(),
		)
	})

	hooks.AddOnError(func(ctx core.Context, err error) {
		logger.Errorw("gRPC ERROR", "method", ctx.FullMethod(), "error", err.Error())
	})

	hooks.AddOnShutdown(func() {
		logger.Infow("gRPC server shutting down...")
	})

	return hooks
}

func databaseHealthCheck(_ context.Context) health.HealthCheck {
	return health.HealthCheck{
		Name:    "database",
		Status:  health.StatusHealthy,
		Message: "Connected to PostgreSQL",
	}
}

func cacheHealthCheck(_ context.Context) health.HealthCheck {
	return health.HealthCheck{
		Name:    "cache",
		Status:  health.StatusHealthy,
		Message: "Redis connection OK",
	}
}

// ===========================================================================
// Service Implementation
// ===========================================================================

type userService struct {
	pb.UnimplementedUserServiceServer
	mu     sync.RWMutex
	users  map[int64]*pb.User
	nextID int64
}

// Unary RPCs

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Check authentication from metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		logger.Debugw("Received metadata", "md", md)
	}

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	s.mu.RLock()
	user, exists := s.users[req.Id]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Errorf(codes.NotFound, "user with ID %d not found", req.Id)
	}

	return &pb.GetUserResponse{User: user}, nil
}

func (s *userService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	users := make([]*pb.User, 0, len(s.users))
	for _, user := range s.users {
		// Filter by active status if specified
		if req.Active && !user.Active {
			continue
		}
		users = append(users, user)
	}
	s.mu.RUnlock()

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

	s.mu.Lock()
	s.nextID++
	user := &pb.User{
		Id:        s.nextID,
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		Role:      req.Role,
		Active:    true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	s.users[user.Id] = user
	s.mu.Unlock()

	logger.Infow("User created", "id", user.Id, "name", user.Name)

	return &pb.CreateUserResponse{User: user}, nil
}

func (s *userService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user with ID %d not found", req.Id)
	}

	// Update fields
	user.Name = req.Name
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	user.UpdatedAt = time.Now().Unix()

	logger.Infow("User updated", "id", user.Id, "name", user.Name)

	return &pb.UpdateUserResponse{User: user}, nil
}

func (s *userService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[req.Id]; !exists {
		return nil, status.Errorf(codes.NotFound, "user with ID %d not found", req.Id)
	}

	delete(s.users, req.Id)
	logger.Infow("User deleted", "id", req.Id)

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// Server Streaming RPC

func (s *userService) StreamUsers(req *pb.StreamUsersRequest, stream pb.UserService_StreamUsersServer) error {
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}

	s.mu.RLock()
	users := make([]*pb.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	s.mu.RUnlock()

	logger.Infow("Streaming users", "total", len(users), "batch_size", batchSize)

	// Stream users in batches with delay to simulate real streaming
	for i, user := range users {
		if err := stream.Send(user); err != nil {
			return status.Errorf(codes.Internal, "failed to send user: %v", err)
		}

		// Small delay to simulate streaming behavior
		if (i+1)%int(batchSize) == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// Client Streaming RPC

func (s *userService) BatchCreateUsers(stream pb.UserService_BatchCreateUsersServer) error {
	createdUsers := make([]*pb.User, 0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			var count int32
			for range createdUsers {
				count++
			}
			logger.Infow("Batch create completed", "count", len(createdUsers))
			return stream.SendAndClose(&pb.BatchCreateUsersResponse{
				Users: createdUsers,
				Count: count,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive user: %v", err)
		}

		// Validate and create user
		if req.Name == "" || req.Email == "" {
			logger.Warnw("Invalid user in batch", "name", req.Name, "email", req.Email)
			continue
		}

		s.mu.Lock()
		s.nextID++
		user := &pb.User{
			Id:        s.nextID,
			Name:      req.Name,
			Email:     req.Email,
			Age:       req.Age,
			Role:      req.Role,
			Active:    true,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		s.users[user.Id] = user
		s.mu.Unlock()

		createdUsers = append(createdUsers, user)
		logger.Debugw("User added to batch", "id", user.Id, "name", user.Name)
	}
}

// Bidirectional Streaming RPC

func (s *userService) UserChat(stream pb.UserService_UserChatServer) error {
	ctx := stream.Context()
	logger.Infow("Chat session started")

	// Echo messages back with timestamp
	for {
		select {
		case <-ctx.Done():
			logger.Infow("Chat session ended")
			return ctx.Err()
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return status.Errorf(codes.Internal, "failed to receive message: %v", err)
			}

			logger.Debugw("Received chat message",
				"user_id", msg.UserId,
				"message", msg.Message,
			)

			// Echo back with server timestamp
			response := &pb.ChatMessage{
				UserId:    msg.UserId,
				Message:   fmt.Sprintf("Echo: %s", msg.Message),
				Timestamp: time.Now().Unix(),
			}

			if err := stream.Send(response); err != nil {
				return status.Errorf(codes.Internal, "failed to send message: %v", err)
			}
		}
	}
}
