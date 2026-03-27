// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	routine "github.com/anthanhphan/gosdk/goroutine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/grpc/configuration"
	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/orianna/grpc/interceptor"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"github.com/anthanhphan/gosdk/tracing"

	_ "google.golang.org/grpc/encoding/gzip" // Register gzip compressor
)

// Server orchestrates gRPC server components.
type Server struct {
	config          *configuration.Config
	grpcServer      *grpc.Server
	serviceRegistry *ServiceRegistry
	hooks           *core.Hooks
	healthManager   *health.Manager
	logger          *logger.Logger
	globalUnary     []grpc.UnaryServerInterceptor
	globalStream    []grpc.StreamServerInterceptor
	panicRecover    grpc.UnaryServerInterceptor
	streamRecover   grpc.StreamServerInterceptor
	rateLimiter     grpc.UnaryServerInterceptor
	disableRecovery bool
	metricsClient   metrics.Client
	tracingClient   tracing.Client
	tokenValidator  interceptor.TokenValidator
}

// NewServer creates a new gRPC server instance with the given configuration and options.
func NewServer(
	conf *configuration.Config,
	options ...ServerOption,
) (*Server, error) {
	// Create logger
	log := logger.NewLoggerWithFields(logger.String("package", "transport"))

	// Validate config
	if conf == nil {
		return nil, fmt.Errorf("invalid config: config cannot be nil")
	}
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Merge with defaults
	conf = configuration.MergeConfigDefaults(conf)

	// Create server
	s := &Server{
		config:          conf,
		serviceRegistry: NewServiceRegistry(),
		hooks:           core.NewHooks(),
		logger:          log,
		globalUnary:     nil,
		globalStream:    nil,
	}

	// Apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Build gRPC server options
	grpcOpts, err := s.buildGRPCServerOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build gRPC options: %w", err)
	}

	// Create the gRPC server
	s.grpcServer = grpc.NewServer(grpcOpts...)

	// Enable reflection if configured
	if conf.EnableReflection {
		reflection.Register(s.grpcServer)
	}

	return s, nil
}

// Start starts the gRPC server and begins listening for incoming requests.
// It blocks until an OS signal (SIGTERM/SIGINT) is received, then performs graceful shutdown.
func (s *Server) Start() error {
	// Invoke OnServerStart hook
	if err := s.hooks.ExecuteOnServerStart(s); err != nil {
		return fmt.Errorf("OnServerStart hook failed: %w", err)
	}

	// Register all services with gRPC server
	for _, svc := range s.serviceRegistry.GetServices() {
		s.registerServiceWithInterceptors(svc)
	}

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
	}

	// Log server start with diagnostic info
	logFields := []any{
		"service", s.config.ServiceName,
		"port", s.config.Port,
		"services", len(s.serviceRegistry.GetServices()),
		"tls", s.config.TLS != nil,
		"mtls", s.config.MTLS != nil,
	}
	if s.config.Version != "" {
		logFields = append(logFields, "version", s.config.Version)
	}
	s.logger.Infow("gRPC server started successfully", logFields...)

	// Start serving in a goroutine with panic recovery
	errChan := make(chan error, 1)
	routine.Run(func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("gRPC server failed: %w", err)
		}
	})

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-errChan:
		s.hooks.ExecuteOnShutdown()
		return err
	case sig := <-quit:
		s.logger.Infow("received shutdown signal", "signal", sig.String())

		timeout := 30 * time.Second
		if s.config.GracefulShutdownTimeout != nil {
			timeout = *s.config.GracefulShutdownTimeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return s.Shutdown(ctx)
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// Invoke OnShutdown hooks
	s.hooks.ExecuteOnShutdown()

	s.logger.Infow("initiating graceful shutdown")

	// Gracefully stop the gRPC server
	stopped := make(chan struct{})
	routine.Run(func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	})

	select {
	case <-stopped:
		s.logger.Infow("graceful shutdown completed")
		return nil
	case <-ctx.Done():
		s.grpcServer.Stop()
		s.logger.Warnw("graceful shutdown timed out, forced stop")
		return fmt.Errorf("graceful shutdown timed out, forced stop")
	}
}

// RegisterServices registers one or more services with the server.
func (s *Server) RegisterServices(services ...ServiceDesc) error {
	return s.serviceRegistry.RegisterServices(services...)
}

// GetHealthManager returns the health check manager.
func (s *Server) GetHealthManager() *health.Manager {
	return s.healthManager
}

// GRPCServer returns the underlying grpc.Server for advanced use cases.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}
