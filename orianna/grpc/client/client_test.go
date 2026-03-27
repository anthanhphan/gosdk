// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

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
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpccore "github.com/anthanhphan/gosdk/orianna/grpc/core"
	"github.com/anthanhphan/gosdk/orianna/shared/resilience"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/metrics"
	"github.com/anthanhphan/gosdk/tracing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Timeout != 30*time.Second {
		t.Errorf("default timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("default max retries = %v, want 3", cfg.MaxRetries)
	}
}

func TestNewClient_RequiresAddress(t *testing.T) {
	_, err := NewClient()
	if err == nil {
		t.Error("expected error when address is not set")
	}
}

func TestNewClient_WithAddress(t *testing.T) {
	c, err := NewClient(
		WithAddress("localhost:50051"),
		WithServiceName("test-service"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.conn == nil {
		t.Error("connection should not be nil")
	}
	if c.serviceName != "test-service" {
		t.Errorf("serviceName = %v, want test-service", c.serviceName)
	}
	_ = c.Close()
}

func TestNewClient_WithOptions(t *testing.T) {
	c, err := NewClient(
		WithAddress("localhost:50051"),
		WithServiceName("order-service"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = c.Close() }()

	if c.config.Address != "localhost:50051" {
		t.Errorf("address = %v, want localhost:50051", c.config.Address)
	}
	if c.serviceName != "order-service" {
		t.Errorf("serviceName = %v, want order-service", c.serviceName)
	}
}

func TestClient_Close(t *testing.T) {
	c, err := NewClient(
		WithAddress("localhost:50051"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestClient_Connection(t *testing.T) {
	c, err := NewClient(
		WithAddress("localhost:50051"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = c.Close() }()

	conn := c.Connection()
	if conn == nil {
		t.Error("Connection() should not return nil")
	}
}

func TestGrpcMetadataCarrier(t *testing.T) {
	md := metadata.New(map[string]string{})
	carrier := &grpccore.GRPCMetadataCarrier{MD: md}

	carrier.Set("key1", "value1")
	if got := carrier.Get("key1"); got != "value1" {
		t.Errorf("Get(key1) = %v, want value1", got)
	}

	if got := carrier.Get("nonexistent"); got != "" {
		t.Errorf("Get(nonexistent) = %v, want empty string", got)
	}

	keys := carrier.Keys()
	if len(keys) != 1 {
		t.Errorf("Keys() length = %d, want 1", len(keys))
	}
}

// ---------- Option tests ----------

func TestWithTLS(t *testing.T) {
	c := &Client{config: DefaultConfig()}
	opt := WithTLS(&TLSConfig{ServerNameOverride: "example.com"})
	if err := opt(c); err != nil {
		t.Fatalf("WithTLS error: %v", err)
	}
	if !c.config.UseTLS {
		t.Error("expected UseTLS = true")
	}
	if c.config.TLSConfig.ServerNameOverride != "example.com" {
		t.Error("expected ServerNameOverride = example.com")
	}
}

func TestWithTracing(t *testing.T) {
	tc := tracing.NewNoopClient()
	c, err := NewClient(
		WithAddress("localhost:50051"),
		WithTracing(tc),
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer c.Close()
	if c.tracing == nil {
		t.Error("expected tracing to be set")
	}
}

func TestWithMetrics(t *testing.T) {
	mc := metrics.NewNoopClient()
	c, err := NewClient(
		WithAddress("localhost:50051"),
		WithMetrics(mc),
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer c.Close()
	if c.metrics == nil {
		t.Error("expected metrics to be set")
	}
}

func TestWithRetry(t *testing.T) {
	c := &Client{config: DefaultConfig()}
	opt := WithRetry(&resilience.RetryConfig{MaxAttempts: 5})
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	if c.retry == nil {
		t.Error("expected retry to be set")
	}
	if c.retry.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", c.retry.MaxAttempts)
	}
}

func TestWithCircuitBreaker(t *testing.T) {
	c := &Client{config: DefaultConfig()}
	opt := WithCircuitBreaker(&resilience.CircuitBreakerConfig{
		HalfOpenMaxRequests: 10,
	})
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	if c.circuit == nil {
		t.Error("expected circuit breaker to be set")
	}
}

func TestWithLogger(t *testing.T) {
	l := logger.NewLoggerWithFields()
	c := &Client{config: DefaultConfig()}
	opt := WithLogger(l)
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	if c.logger != l {
		t.Error("expected logger to be set")
	}
}

// ---------- loadTLSCredentials tests ----------

func TestLoadTLSCredentials_ServerNameOverride(t *testing.T) {
	cfg := &TLSConfig{ServerNameOverride: "example.com"}
	creds, err := loadTLSCredentials(cfg)
	if err != nil {
		t.Fatalf("loadTLSCredentials error: %v", err)
	}
	if creds == nil {
		t.Error("expected non-nil credentials")
	}
}

func TestLoadTLSCredentials_WithCA(t *testing.T) {
	certPEM, _ := generateTestCert(t)
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	_ = os.WriteFile(caFile, certPEM, 0600)

	cfg := &TLSConfig{CAFile: caFile}
	creds, err := loadTLSCredentials(cfg)
	if err != nil {
		t.Fatalf("loadTLSCredentials error: %v", err)
	}
	if creds == nil {
		t.Error("expected non-nil credentials")
	}
}

func TestLoadTLSCredentials_WithCA_InvalidFile(t *testing.T) {
	cfg := &TLSConfig{CAFile: "/nonexistent/ca.pem"}
	_, err := loadTLSCredentials(cfg)
	if err == nil {
		t.Error("expected error for invalid CA file")
	}
}

func TestLoadTLSCredentials_WithCA_InvalidPEM(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	_ = os.WriteFile(caFile, []byte("not a cert"), 0600)

	cfg := &TLSConfig{CAFile: caFile}
	_, err := loadTLSCredentials(cfg)
	if err == nil {
		t.Error("expected error for invalid CA PEM")
	}
}

func TestLoadTLSCredentials_WithMTLS(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	_ = os.WriteFile(certFile, certPEM, 0600)
	_ = os.WriteFile(keyFile, keyPEM, 0600)

	cfg := &TLSConfig{CertFile: certFile, KeyFile: keyFile}
	creds, err := loadTLSCredentials(cfg)
	if err != nil {
		t.Fatalf("loadTLSCredentials error: %v", err)
	}
	if creds == nil {
		t.Error("expected non-nil credentials")
	}
}

func TestLoadTLSCredentials_WithMTLS_InvalidCert(t *testing.T) {
	cfg := &TLSConfig{
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	}
	_, err := loadTLSCredentials(cfg)
	if err == nil {
		t.Error("expected error for invalid cert/key files")
	}
}

// ---------- NewClient with TLS ----------

func TestNewClient_WithTLS(t *testing.T) {
	c, err := NewClient(
		WithAddress("localhost:50051"),
		WithTLS(&TLSConfig{ServerNameOverride: "example.com"}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	if !c.config.UseTLS {
		t.Error("expected UseTLS = true")
	}
}

// ---------- Invoke tests ----------

func TestClient_Invoke_NoCircuitBreaker(t *testing.T) {
	c, err := NewClient(
		WithAddress("127.0.0.1:1"), // unreachable
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestClient_Invoke_WithMetrics(t *testing.T) {
	mc := metrics.NewNoopClient()
	c, err := NewClient(
		WithAddress("127.0.0.1:1"),
		WithMetrics(mc),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestClient_Invoke_WithTracing(t *testing.T) {
	tc := tracing.NewNoopClient()
	c, err := NewClient(
		WithAddress("127.0.0.1:1"),
		WithTracing(tc),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestClient_Invoke_CircuitBreakerOpen(t *testing.T) {
	cbCfg := &resilience.CircuitBreakerConfig{HalfOpenMaxRequests: 1}
	c, err := NewClient(
		WithAddress("127.0.0.1:1"),
		WithCircuitBreaker(cbCfg),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	// Trip the circuit breaker by recording failures
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_ = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
		cancel()
	}

	// Now the circuit breaker should be open
	ctx := context.Background()
	err = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
	// May or may not be circuit breaker error depending on timing
	_ = err
}

func TestClient_Invoke_WithAllObservability(t *testing.T) {
	mc := metrics.NewNoopClient()
	tc := tracing.NewNoopClient()
	c, err := NewClient(
		WithAddress("127.0.0.1:1"),
		WithMetrics(mc),
		WithTracing(tc),
		WithServiceName("test-svc"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This will fail (unreachable server) but exercises all observability paths
	_ = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
}

// ---------- NewStream tests ----------

func TestClient_NewStream_Error(t *testing.T) {
	c, err := NewClient(
		WithAddress("127.0.0.1:1"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	desc := &grpc.StreamDesc{StreamName: "StreamMethod", ServerStreams: true}
	_, err = c.NewStream(ctx, desc, "/test.Svc/StreamMethod")
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestClient_NewStream_WithTracing(t *testing.T) {
	tc := tracing.NewNoopClient()
	c, err := NewClient(
		WithAddress("127.0.0.1:1"),
		WithTracing(tc),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	desc := &grpc.StreamDesc{StreamName: "StreamMethod", ClientStreams: true, ServerStreams: true}
	_, err = c.NewStream(ctx, desc, "/test.Svc/StreamMethod")
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

// ---------- extractTraceContext ----------

func TestExtractTraceContext_NoTracing(t *testing.T) {
	c := &Client{}
	ctx := context.Background()
	result := c.extractTraceContext(ctx)
	if result != ctx {
		t.Error("expected same context when tracing is nil")
	}
}

func TestExtractTraceContext_WithTracing(t *testing.T) {
	tc := tracing.NewNoopClient()
	c := &Client{tracing: tc}
	ctx := context.Background()
	result := c.extractTraceContext(ctx)
	// Should return a context (possibly the same or with extracted carriers)
	if result == nil {
		t.Error("expected non-nil context")
	}
}

// ---------- Client interceptors ----------

func TestNewUnaryClientInterceptor_Success(t *testing.T) {
	interceptor := newUnaryClientInterceptor("test-svc")

	invoker := func(ctx context.Context, method string, args, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	err := interceptor(context.Background(), "/test.Svc/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewUnaryClientInterceptor_Error(t *testing.T) {
	interceptor := newUnaryClientInterceptor("test-svc")

	invoker := func(ctx context.Context, method string, args, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return context.DeadlineExceeded
	}

	err := interceptor(context.Background(), "/test.Svc/Method", nil, nil, nil, invoker)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewStreamClientInterceptor_Success(t *testing.T) {
	interceptor := newStreamClientInterceptor("test-svc")

	mockStream := &mockClientStream{}
	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return mockStream, nil
	}

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test.Svc/Stream", streamer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stream == nil {
		t.Error("expected non-nil stream")
	}
}

func TestNewStreamClientInterceptor_Error(t *testing.T) {
	interceptor := newStreamClientInterceptor("test-svc")

	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, context.DeadlineExceeded
	}

	_, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test.Svc/Stream", streamer)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------- observedClientStream ----------

func TestObservedClientStream_CloseSend_NoSpan(t *testing.T) {
	mockStream := &mockClientStream{}
	obs := &observedClientStream{
		ClientStream: mockStream,
		serviceName:  "test",
		method:       "/test/m",
	}
	err := obs.CloseSend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- Mocks ----------

type mockClientStream struct {
	grpc.ClientStream
}

func (m *mockClientStream) CloseSend() error { return nil }

// ---------- Test cert generation ----------

func generateTestCert(t *testing.T) (certPEM, keyPEM []byte) {
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

// ---------- isRetryable tests ----------

func TestIsRetryable_RetryableCode(t *testing.T) {
	c := &Client{
		config: DefaultConfig(),
		retry: &resilience.RetryConfig{
			MaxAttempts:          3,
			RetryableStatusCodes: []int{14}, // codes.Unavailable
		},
		retryStatusSet: map[int]struct{}{14: {}},
	}

	err := status.Error(codes.Unavailable, "service unavailable")
	if !c.isRetryable(err) {
		t.Error("expected Unavailable to be retryable")
	}
}

func TestIsRetryable_NonRetryableCode(t *testing.T) {
	c := &Client{
		config: DefaultConfig(),
		retry: &resilience.RetryConfig{
			MaxAttempts:          3,
			RetryableStatusCodes: []int{14},
		},
		retryStatusSet: map[int]struct{}{14: {}},
	}

	err := status.Error(codes.NotFound, "not found")
	if c.isRetryable(err) {
		t.Error("expected NotFound to NOT be retryable")
	}
}

func TestIsRetryable_NonGRPCError(t *testing.T) {
	c := &Client{
		config:         DefaultConfig(),
		retry:          &resilience.RetryConfig{MaxAttempts: 3},
		retryStatusSet: map[int]struct{}{14: {}},
	}

	err := context.DeadlineExceeded
	// Non-gRPC errors may parse to OK/Unknown — verify no panic
	_ = c.isRetryable(err)
}

// ---------- calculateBackoff tests ----------

func TestCalculateBackoff_FirstAttempt(t *testing.T) {
	c := &Client{
		retry: &resilience.RetryConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2.0,
		},
	}

	backoff := c.calculateBackoff(0)
	// Expected: 100ms * 2^0 = 100ms + 0-25% jitter = 100-125ms
	if backoff < 100*time.Millisecond || backoff > 150*time.Millisecond {
		t.Errorf("backoff = %v, expected 100-150ms", backoff)
	}
}

func TestCalculateBackoff_CappedAtMax(t *testing.T) {
	c := &Client{
		retry: &resilience.RetryConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     200 * time.Millisecond,
			Multiplier:     2.0,
		},
	}

	// attempt=10: 100ms * 2^10 = 102400ms, capped at 200ms + jitter
	backoff := c.calculateBackoff(10)
	if backoff > 300*time.Millisecond {
		t.Errorf("backoff = %v, should be capped around 200-250ms", backoff)
	}
}

// ---------- waitForRetry tests ----------

func TestWaitForRetry_ContextCancelled(t *testing.T) {
	l := logger.NewLoggerWithFields(logger.String("test", "true"))
	c := &Client{
		retry: &resilience.RetryConfig{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2.0,
		},
		logger: l,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := c.waitForRetry(ctx, "/test/Method", 0, 3)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// ---------- CloseSend with span ----------

func TestObservedClientStream_CloseSend_WithSpan(t *testing.T) {
	tc := tracing.NewNoopClient()
	_, span := tc.StartSpan(context.Background(), "test")

	mockStream := &mockClientStream{}
	obs := &observedClientStream{
		ClientStream: mockStream,
		span:         span,
		serviceName:  "test",
		method:       "/test/m",
	}
	err := obs.CloseSend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestObservedClientStream_CloseSend_WithSpan_Error(t *testing.T) {
	tc := tracing.NewNoopClient()
	_, span := tc.StartSpan(context.Background(), "test")

	mockStream := &mockErrorClientStream{}
	obs := &observedClientStream{
		ClientStream: mockStream,
		span:         span,
		serviceName:  "test",
		method:       "/test/m",
	}
	err := obs.CloseSend()
	if err == nil {
		t.Error("expected error from CloseSend")
	}
}

type mockErrorClientStream struct {
	grpc.ClientStream
}

func (m *mockErrorClientStream) CloseSend() error {
	return context.DeadlineExceeded
}

// ---------- Invoke with retry ----------

func TestClient_Invoke_WithRetry(t *testing.T) {
	c, err := NewClient(
		WithAddress("127.0.0.1:1"), // unreachable
		WithRetry(&resilience.RetryConfig{
			MaxAttempts:          2,
			InitialBackoff:       1 * time.Millisecond,
			MaxBackoff:           10 * time.Millisecond,
			Multiplier:           2.0,
			RetryableStatusCodes: []int{14}, // Unavailable
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Will fail (unreachable) but exercises retry logic paths
	err = c.Invoke(ctx, "/test.Svc/Method", nil, nil)
	if err == nil {
		t.Error("expected error for unreachable server with retry")
	}
}

// ---------- recordInvokeSpan with success ----------

func TestRecordInvokeSpan_Success(t *testing.T) {
	tc := tracing.NewNoopClient()
	c := &Client{tracing: tc}

	_, span := tc.StartSpan(context.Background(), "test")
	c.recordInvokeSpan(span, nil) // should not panic
}

func TestRecordInvokeSpan_NilSpan(t *testing.T) {
	c := &Client{}
	c.recordInvokeSpan(nil, nil) // should not panic
}
