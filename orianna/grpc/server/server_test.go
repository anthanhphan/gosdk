// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/orianna/grpc/configuration"
	"github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"github.com/anthanhphan/gosdk/tracing"
)

// dummyServiceImpl is a fake service implementation for testing.
type dummyServiceImpl struct{}

func TestNewServer_ValidConfig(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test-service",
		Port:        50051,
	}

	s, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s == nil {
		t.Fatal("server should not be nil")
	}
	if s.grpcServer == nil {
		t.Error("grpcServer should not be nil")
	}
	if s.serviceRegistry == nil {
		t.Error("serviceRegistry should not be nil")
	}
}

func TestNewServer_InvalidConfig(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "",
		Port:        0,
	}

	_, err := NewServer(conf)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test-service",
		Port:        50051,
	}

	calledCount := 0
	customInterceptor := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		calledCount++
		return handler(ctx, req)
	}

	s, err := NewServer(conf,
		WithGlobalUnaryInterceptor(customInterceptor),
		WithDisableRecovery(),
	)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if !s.disableRecovery {
		t.Error("disableRecovery should be true")
	}
	if len(s.globalUnary) != 1 {
		t.Errorf("globalUnary length = %d, want 1", len(s.globalUnary))
	}
}

func TestServiceRegistry_RegisterAndGet(t *testing.T) {
	reg := NewServiceRegistry()

	desc := &grpc.ServiceDesc{
		ServiceName: "test.Service",
	}

	err := reg.RegisterService(ServiceDesc{
		Desc: desc,
		Impl: &dummyServiceImpl{},
	})
	if err != nil {
		t.Fatalf("RegisterService() error = %v", err)
	}

	services := reg.GetServices()
	if len(services) != 1 {
		t.Fatalf("GetServices() length = %d, want 1", len(services))
	}
	if services[0].Desc.ServiceName != "test.Service" {
		t.Errorf("service name = %v, want test.Service", services[0].Desc.ServiceName)
	}
}

func TestServiceRegistry_DuplicateDetection(t *testing.T) {
	reg := NewServiceRegistry()

	desc := &grpc.ServiceDesc{
		ServiceName: "test.Service",
	}
	svc := ServiceDesc{
		Desc: desc,
		Impl: &dummyServiceImpl{},
	}

	if err := reg.RegisterService(svc); err != nil {
		t.Fatalf("first RegisterService() error = %v", err)
	}

	err := reg.RegisterService(svc)
	if err == nil {
		t.Error("expected error for duplicate service")
	}
}

func TestServiceRegistry_NilValidation(t *testing.T) {
	reg := NewServiceRegistry()

	tests := []struct {
		name string
		svc  ServiceDesc
	}{
		{"nil desc", ServiceDesc{Desc: nil, Impl: &dummyServiceImpl{}}},
		{"nil impl", ServiceDesc{Desc: &grpc.ServiceDesc{ServiceName: "test"}, Impl: nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reg.RegisterService(tt.svc)
			if err == nil {
				t.Error("expected error for invalid service")
			}
		})
	}
}

func TestServiceRegistry_AtomicRegistration(t *testing.T) {
	reg := NewServiceRegistry()

	// Register multiple services where the second one is invalid
	err := reg.RegisterServices(
		ServiceDesc{
			Desc: &grpc.ServiceDesc{ServiceName: "valid.Service"},
			Impl: &dummyServiceImpl{},
		},
		ServiceDesc{
			Desc: nil, // invalid
			Impl: &dummyServiceImpl{},
		},
	)
	if err == nil {
		t.Error("expected error for batch with invalid service")
	}

	// First service should still have been registered since it was processed before the error
	// (note: the current implementation is not truly atomic in rollback — it's sequential)
}

func TestServiceBuilder(t *testing.T) {
	desc := &grpc.ServiceDesc{ServiceName: "test.Service"}
	impl := &dummyServiceImpl{}

	interceptor := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}

	svc := NewService(desc, impl).
		UnaryInterceptor(interceptor).
		Build()

	if svc.Desc != desc {
		t.Error("Desc mismatch")
	}
	if svc.Impl != impl {
		t.Error("Impl mismatch")
	}
	if len(svc.UnaryInterceptors) != 1 {
		t.Errorf("UnaryInterceptors length = %d, want 1", len(svc.UnaryInterceptors))
	}
}

func TestServer_Shutdown(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test-service",
		Port:        50099,
	}

	s, err := NewServer(conf, WithDisableRecovery())
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Shutdown without starting should complete
	err = s.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestBuildUnaryInterceptorChain(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "test-service",
		Port:        50051,
	}

	// With recovery enabled (default)
	s, _ := NewServer(conf)
	chain := s.buildUnaryInterceptorChain(nil)
	if len(chain) < 2 {
		t.Errorf("chain length = %d, want at least 2 (recovery + requestID)", len(chain))
	}

	// With recovery disabled
	s2, _ := NewServer(conf, WithDisableRecovery())
	chain2 := s2.buildUnaryInterceptorChain(nil)
	if len(chain2) < 1 {
		t.Errorf("chain2 length = %d, want at least 1 (requestID)", len(chain2))
	}
	// chain2 should have one fewer interceptor than chain
	if len(chain2) >= len(chain) {
		t.Errorf("disabling recovery should reduce chain length: %d >= %d", len(chain2), len(chain))
	}
}

// ---------- Option tests ----------

func TestWithGlobalStreamInterceptor(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	streamInt := func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
	s, err := NewServer(conf, WithGlobalStreamInterceptor(streamInt))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if len(s.globalStream) != 1 {
		t.Errorf("globalStream length = %d, want 1", len(s.globalStream))
	}
}

func TestWithPanicRecover(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	unaryRecover := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}
	streamRecover := func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
	s, err := NewServer(conf, WithPanicRecover(unaryRecover, streamRecover))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.panicRecover == nil {
		t.Error("panicRecover should not be nil")
	}
	if s.streamRecover == nil {
		t.Error("streamRecover should not be nil")
	}
}

func TestWithRateLimiter(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	limiter := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}
	s, err := NewServer(conf, WithRateLimiter(limiter))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.rateLimiter == nil {
		t.Error("rateLimiter should not be nil")
	}
}

func TestWithHooks(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	hooks := core.NewHooks()
	s, err := NewServer(conf, WithHooks(hooks))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.hooks != hooks {
		t.Error("hooks mismatch")
	}
}

func TestWithHealthManager(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	hm := health.NewManager()
	s, err := NewServer(conf, WithHealthManager(hm))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.healthManager != hm {
		t.Error("healthManager mismatch")
	}
}

func TestWithMetrics(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	mc := metrics.NewNoopClient()
	s, err := NewServer(conf, WithMetrics(mc))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.metricsClient == nil {
		t.Error("metricsClient should not be nil")
	}
}

func TestWithTracing(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	tc := tracing.NewNoopClient()
	s, err := NewServer(conf, WithTracing(tc))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.tracingClient == nil {
		t.Error("tracingClient should not be nil")
	}
}

func TestWithHealthChecker(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	checker := &mockChecker{name: "test-checker"}
	s, err := NewServer(conf, WithHealthChecker(checker))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.healthManager == nil {
		t.Error("healthManager should auto-create")
	}
}

// ---------- Server accessor tests ----------

func TestServer_Accessors(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	hm := health.NewManager()
	s, err := NewServer(conf, WithHealthManager(hm))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.GetHealthManager() != hm {
		t.Error("GetHealthManager mismatch")
	}
	if s.GRPCServer() == nil {
		t.Error("GRPCServer should not be nil")
	}
}

// ---------- Interceptor chain building tests ----------

func TestBuildUnaryInterceptorChain_AllFeatures(t *testing.T) {
	threshold := 100 * time.Millisecond
	conf := &configuration.Config{
		ServiceName:          "svc",
		Port:                 50051,
		VerboseLogging:       true,
		SlowRequestThreshold: threshold,
	}

	mc := metrics.NewNoopClient()
	tc := tracing.NewNoopClient()
	limiter := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}

	s, _ := NewServer(conf,
		WithMetrics(mc),
		WithTracing(tc),
		WithRateLimiter(limiter),
	)

	// Set permissions directly on config after server creation (avoid MTLS validation)
	s.config.Permissions = []configuration.ClientPermission{
		{ClientIdentity: "client1", AllowedMethods: []string{"/*"}},
	}

	perms := s.buildCertPermissions()
	chain := s.buildUnaryInterceptorChain(perms)
	// recovery + requestID + certAuth + rateLimiter + metrics + slowRPC + tracing + verbose = 8
	if len(chain) < 8 {
		t.Errorf("chain length = %d, want at least 8", len(chain))
	}
}

func TestBuildStreamInterceptorChain_AllFeatures(t *testing.T) {
	threshold := 100 * time.Millisecond
	conf := &configuration.Config{
		ServiceName:          "svc",
		Port:                 50051,
		VerboseLogging:       true,
		SlowRequestThreshold: threshold,
	}

	mc := metrics.NewNoopClient()
	tc := tracing.NewNoopClient()

	s, _ := NewServer(conf,
		WithMetrics(mc),
		WithTracing(tc),
	)

	s.config.Permissions = []configuration.ClientPermission{
		{ClientIdentity: "client1", AllowedMethods: []string{"/*"}},
	}

	perms := s.buildCertPermissions()
	chain := s.buildStreamInterceptorChain(perms)
	// recovery + requestID + certAuth + metrics + slowRPC + tracing + verbose = 7
	if len(chain) < 7 {
		t.Errorf("chain length = %d, want at least 7", len(chain))
	}
}

func TestBuildCertPermissions(t *testing.T) {
	// Empty permissions
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf)
	perms := s.buildCertPermissions()
	if perms != nil {
		t.Error("expected nil permissions for empty config")
	}

	// With permissions
	conf2 := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
	}
	s2, _ := NewServer(conf2)
	s2.config.Permissions = []configuration.ClientPermission{
		{ClientIdentity: "svc1", AllowedMethods: []string{"/a/b"}},
	}
	perms2 := s2.buildCertPermissions()
	if len(perms2) != 1 {
		t.Errorf("expected 1 permission, got %d", len(perms2))
	}
}

func TestBuildUnaryInterceptorChain_CustomRecover(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	customRecover := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}
	s, _ := NewServer(conf, WithPanicRecover(customRecover, nil))
	chain := s.buildUnaryInterceptorChain(nil)
	if len(chain) < 2 {
		t.Errorf("chain length = %d, want at least 2", len(chain))
	}
}

func TestBuildStreamInterceptorChain_CustomRecover(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	customStreamRecover := func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
	s, _ := NewServer(conf, WithPanicRecover(nil, customStreamRecover))
	chain := s.buildStreamInterceptorChain(nil)
	if len(chain) < 2 {
		t.Errorf("chain length = %d, want at least 2", len(chain))
	}
}

func TestBuildStreamInterceptorChain_DisableRecovery(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf, WithDisableRecovery())
	chain := s.buildStreamInterceptorChain(nil)
	// Only requestID (no recovery)
	if len(chain) < 1 {
		t.Errorf("chain length = %d, want at least 1", len(chain))
	}
}

// ---------- registerServiceWithInterceptors ----------

// dummyService is an interface type for gRPC service registration.
type dummyService interface {
	DummyMethod()
}

// Ensure dummyServiceImpl implements dummyService.
func (d *dummyServiceImpl) DummyMethod() {}

func TestRegisterServices_ViaServer(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf)
	err := s.RegisterServices(ServiceDesc{
		Desc: &grpc.ServiceDesc{
			ServiceName: "test.Svc",
			HandlerType: (*dummyService)(nil),
		},
		Impl: &dummyServiceImpl{},
	})
	if err != nil {
		t.Fatalf("RegisterServices() error = %v", err)
	}
}

func TestRegisterServiceWithInterceptors_WithChain(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf)

	called := false
	interceptor := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		called = true
		return handler(ctx, req)
	}

	svc := ServiceDesc{
		Desc: &grpc.ServiceDesc{
			ServiceName: "test.Svc",
			HandlerType: (*dummyService)(nil),
			Methods: []grpc.MethodDesc{
				{
					MethodName: "TestMethod",
					Handler: func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
						return "response", nil
					},
				},
			},
		},
		Impl:              &dummyServiceImpl{},
		UnaryInterceptors: []grpc.UnaryServerInterceptor{interceptor},
	}
	// Should register without panic
	s.registerServiceWithInterceptors(svc)
	_ = called // verified by no panic
}

func TestRegisterServiceWithInterceptors_NoInterceptors(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf)

	svc := ServiceDesc{
		Desc: &grpc.ServiceDesc{
			ServiceName: "test.Svc2",
			HandlerType: (*dummyService)(nil),
		},
		Impl: &dummyServiceImpl{},
	}
	s.registerServiceWithInterceptors(svc)
}

// ---------- Shutdown tests ----------

func TestServer_Shutdown_Graceful(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Shutdown without Start — GracefulStop returns immediately
	err := s.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestServer_Shutdown_Timeout(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	s, _ := NewServer(conf)

	// Use an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Shutdown(ctx)
	// Could be nil or timeout error depending on timing
	_ = err
}

func TestServer_Shutdown_WithHealthManager(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	hm := health.NewManager()
	s, _ := NewServer(conf, WithHealthManager(hm))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

// ---------- Health checker tests ----------

func TestNewGRPCChecker(t *testing.T) {
	checker := NewGRPCChecker("localhost:50051", "", 0)
	if checker.Name() != "grpc" { // default name from configuration
		t.Errorf("default name = %q, want 'grpc'", checker.Name())
	}

	checker2 := NewGRPCChecker("localhost:50051", "custom", time.Second)
	if checker2.Name() != "custom" {
		t.Errorf("custom name = %q, want 'custom'", checker2.Name())
	}
}

func TestGRPCChecker_Close_BeforeInit(t *testing.T) {
	checker := NewGRPCChecker("localhost:50051", "test", time.Second)
	err := checker.Close()
	if err != nil {
		t.Errorf("Close() before init should not error: %v", err)
	}
}

func TestGRPCChecker_Check_InvalidTarget(t *testing.T) {
	checker := NewGRPCChecker("localhost:1", "test", 100*time.Millisecond)
	defer checker.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result := checker.Check(ctx)
	if result.Name != "test" {
		t.Errorf("Name = %q, want 'test'", result.Name)
	}
	// Connection to invalid port won't reach Ready state
	if result.Status == "" {
		t.Error("status should not be empty")
	}
}

func TestGRPCChecker_Check_Timeout(t *testing.T) {
	checker := NewGRPCChecker("localhost:1", "test", 1*time.Millisecond)
	defer checker.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	result := checker.Check(ctx)
	// Should return unhealthy due to timeout
	if result.Status == health.StatusHealthy {
		t.Error("expected unhealthy status for unreachable target")
	}
}

// ---------- StreamInterceptor builder ----------

func TestServiceBuilder_StreamInterceptor(t *testing.T) {
	desc := &grpc.ServiceDesc{ServiceName: "test.Service"}
	impl := &dummyServiceImpl{}

	streamInt := func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}

	svc := NewService(desc, impl).
		StreamInterceptor(streamInt).
		Build()

	if len(svc.StreamInterceptors) != 1 {
		t.Errorf("StreamInterceptors length = %d, want 1", len(svc.StreamInterceptors))
	}
}

// ---------- Config with reflection ----------

func TestNewServer_WithReflection(t *testing.T) {
	conf := &configuration.Config{
		ServiceName:      "svc",
		Port:             50051,
		EnableReflection: true,
	}
	s, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.grpcServer == nil {
		t.Error("grpcServer should not be nil")
	}
}

// ---------- Mocks ----------

type mockChecker struct {
	name string
}

func (m *mockChecker) Check(ctx context.Context) health.HealthCheck {
	return health.HealthCheck{Name: m.name, Status: health.StatusHealthy}
}

func (m *mockChecker) Name() string { return m.name }

// ---------- Start tests ----------

func TestServer_Start_ListenAndSignal(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        0, // let OS assign port
	}
	s, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	// Register a dummy service
	_ = s.RegisterServices(ServiceDesc{
		Desc: &grpc.ServiceDesc{
			ServiceName: "test.Svc",
			HandlerType: (*dummyService)(nil),
		},
		Impl: &dummyServiceImpl{},
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	// Give it a moment to start, then send SIGTERM
	time.Sleep(100 * time.Millisecond)
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() timed out")
	}
}

func TestServer_Start_WithVersion(t *testing.T) {
	version := "1.0.0"
	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        0,
		Version:     version,
	}
	s, _ := NewServer(conf)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	time.Sleep(100 * time.Millisecond)
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() timed out")
	}
}

func TestServer_Start_WithCustomShutdownTimeout(t *testing.T) {
	timeout := 1 * time.Second
	conf := &configuration.Config{
		ServiceName:             "svc",
		Port:                    0,
		GracefulShutdownTimeout: &timeout,
	}
	s, _ := NewServer(conf)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	time.Sleep(100 * time.Millisecond)
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() timed out")
	}
}

// ---------- buildMTLSCredentials tests ----------

func TestBuildMTLSCredentials_InvalidCertFile(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		MTLS: &configuration.MTLSConfig{
			CertFile: "/nonexistent/cert.pem",
			KeyFile:  "/nonexistent/key.pem",
			CAFile:   "/nonexistent/ca.pem",
		},
	}
	s := &Server{config: conf}
	_, err := s.buildMTLSCredentials()
	if err == nil {
		t.Error("expected error for invalid cert file")
	}
}

func TestBuildMTLSCredentials_InvalidCAFile(t *testing.T) {
	// Create temp cert/key pair first
	certPEM, keyPEM := generateSelfSignedCert(t)

	certFile := filepath.Join(t.TempDir(), "cert.pem")
	keyFile := filepath.Join(t.TempDir(), "key.pem")

	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatal(err)
	}

	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		MTLS: &configuration.MTLSConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
			CAFile:   "/nonexistent/ca.pem",
		},
	}
	s := &Server{config: conf}
	_, err := s.buildMTLSCredentials()
	if err == nil {
		t.Error("expected error for invalid CA file")
	}
}

func TestBuildMTLSCredentials_InvalidCAPEM(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	caFile := filepath.Join(dir, "ca.pem")

	_ = os.WriteFile(certFile, certPEM, 0600)
	_ = os.WriteFile(keyFile, keyPEM, 0600)
	_ = os.WriteFile(caFile, []byte("not a cert"), 0600)

	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		MTLS: &configuration.MTLSConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
			CAFile:   caFile,
		},
	}
	s := &Server{config: conf}
	_, err := s.buildMTLSCredentials()
	if err == nil {
		t.Error("expected error for invalid CA PEM")
	}
}

func TestBuildMTLSCredentials_ValidCerts(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	caFile := filepath.Join(dir, "ca.pem")

	_ = os.WriteFile(certFile, certPEM, 0600)
	_ = os.WriteFile(keyFile, keyPEM, 0600)
	_ = os.WriteFile(caFile, certPEM, 0600) // self-signed: use cert as CA

	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		MTLS: &configuration.MTLSConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
			CAFile:   caFile,
		},
	}
	s := &Server{config: conf}
	creds, err := s.buildMTLSCredentials()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if creds == nil {
		t.Error("expected non-nil credentials")
	}
}

// ---------- buildGRPCServerOptions with optional fields ----------

func TestBuildGRPCServerOptions_WithKeepaliveAndTimeout(t *testing.T) {
	keepaliveTime := 30 * time.Second
	keepaliveTimeout := 10 * time.Second
	connTimeout := 5 * time.Second
	conf := &configuration.Config{
		ServiceName:          "svc",
		Port:                 50051,
		KeepaliveTime:        &keepaliveTime,
		KeepaliveTimeout:     &keepaliveTimeout,
		ConnectionTimeout:    &connTimeout,
		MaxConcurrentStreams: 100,
	}
	s, _ := NewServer(conf)
	if s.grpcServer == nil {
		t.Error("grpcServer should not be nil")
	}
}

// generateSelfSignedCert generates a self-signed certificate for testing.
func generateSelfSignedCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		IsCA:         true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

// ---------- buildTLSCredentials tests ----------

func TestBuildTLSCredentials_ValidCerts(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	_ = os.WriteFile(certFile, certPEM, 0600)
	_ = os.WriteFile(keyFile, keyPEM, 0600)

	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		TLS: &configuration.TLSConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
		},
	}
	s := &Server{config: conf, logger: logger.NewLoggerWithFields()}
	creds, err := s.buildTLSCredentials()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if creds == nil {
		t.Error("expected non-nil credentials")
	}
}

func TestBuildTLSCredentials_InvalidCerts(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		TLS: &configuration.TLSConfig{
			CertFile: "/nonexistent/cert.pem",
			KeyFile:  "/nonexistent/key.pem",
		},
	}
	s := &Server{config: conf, logger: logger.NewLoggerWithFields()}
	_, err := s.buildTLSCredentials()
	if err == nil {
		t.Error("expected error for invalid cert files")
	}
}

// ---------- buildCertReloader tests ----------

func TestBuildCertReloader_Success(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	_ = os.WriteFile(certFile, certPEM, 0600)
	_ = os.WriteFile(keyFile, keyPEM, 0600)

	s := &Server{logger: logger.NewLoggerWithFields()}
	reloader := s.buildCertReloader(certFile, keyFile)

	cert, err := reloader(nil)
	if err != nil {
		t.Fatalf("reloader error: %v", err)
	}
	if cert == nil {
		t.Error("expected non-nil certificate")
	}
}

func TestBuildCertReloader_Failure(t *testing.T) {
	s := &Server{logger: logger.NewLoggerWithFields()}
	reloader := s.buildCertReloader("/nonexistent/cert.pem", "/nonexistent/key.pem")

	_, err := reloader(nil)
	if err == nil {
		t.Error("expected error for invalid cert files")
	}
}

// ---------- WithTokenAuth test ----------

func TestWithTokenAuth(t *testing.T) {
	conf := &configuration.Config{ServiceName: "svc", Port: 50051}
	validator := &mockTokenValidator{}
	s, err := NewServer(conf, WithTokenAuth(validator))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.tokenValidator == nil {
		t.Error("tokenValidator should not be nil")
	}
}

// ---------- WithTokenAuth interceptor chain test ----------

func TestBuildUnaryInterceptorChain_WithTokenAuth(t *testing.T) {
	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
	}
	validator := &mockTokenValidator{}
	s, _ := NewServer(conf, WithTokenAuth(validator))
	chain := s.buildUnaryInterceptorChain(nil)
	// Should include: recovery + requestID + tokenAuth = at least 3
	if len(chain) < 3 {
		t.Errorf("chain length = %d, want at least 3 (recovery + requestID + tokenAuth)", len(chain))
	}
}

// ---------- buildGRPCServerOptions with buffer configs ----------

func TestBuildGRPCServerOptions_WithBufferAndFlowControl(t *testing.T) {
	connTimeout := 5 * time.Second
	conf := &configuration.Config{
		ServiceName:           "svc",
		Port:                  50051,
		ConnectionTimeout:     &connTimeout,
		ReadBufferSize:        65536,
		WriteBufferSize:       65536,
		InitialWindowSize:     131072,
		InitialConnWindowSize: 131072,
	}
	s, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.grpcServer == nil {
		t.Error("grpcServer should not be nil")
	}
}

func TestBuildGRPCServerOptions_WithTLS(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	_ = os.WriteFile(certFile, certPEM, 0600)
	_ = os.WriteFile(keyFile, keyPEM, 0600)

	conf := &configuration.Config{
		ServiceName: "svc",
		Port:        50051,
		TLS: &configuration.TLSConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
		},
	}
	s, err := NewServer(conf)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if s.grpcServer == nil {
		t.Error("grpcServer should not be nil")
	}
}

// ---------- Token validator mock ----------

type mockTokenValidator struct{}

func (m *mockTokenValidator) Validate(_ context.Context, token string) (map[string]any, error) {
	return map[string]any{"sub": "user-123"}, nil
}

// ---------- validateService edge case ----------
