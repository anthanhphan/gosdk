// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Helper: create a context with mocked TLS peer certificate.
func contextWithCert(cn string) context.Context {
	cert := &x509.Certificate{
		Subject: pkix.Name{CommonName: cn},
	}
	tlsInfo := credentials.TLSInfo{
		State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert},
		},
	}
	p := &peer.Peer{AuthInfo: tlsInfo}
	return peer.NewContext(context.Background(), p)
}

// Helper: create a context with no TLS info.
func contextWithoutCert() context.Context {
	return context.Background()
}

// certACL tests

func TestCertACL_ValidIdentity(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/*"}},
	}
	acl := newCertACL(perms)

	err := acl.check("user-service", "/pkg.Svc/Method")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCertACL_UnknownIdentity(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/*"}},
	}
	acl := newCertACL(perms)

	err := acl.check("unknown-service", "/pkg.Svc/Method")
	if err == nil {
		t.Fatal("expected error for unknown identity")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
}

func TestCertACL_MethodNotAllowed(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/pkg.Svc/Allowed"}},
	}
	acl := newCertACL(perms)

	err := acl.check("user-service", "/pkg.Svc/NotAllowed")
	if err == nil {
		t.Fatal("expected error for disallowed method")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
}

func TestCertACL_ServiceWildcard(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/pkg.Svc/*"}},
	}
	acl := newCertACL(perms)

	err := acl.check("user-service", "/pkg.Svc/AnyMethod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Different service should fail
	err = acl.check("user-service", "/other.Svc/Method")
	if err == nil {
		t.Fatal("expected error for different service")
	}
}

func TestCertACL_GlobalWildcard(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "admin", AllowedMethods: []string{"/*"}},
	}
	acl := newCertACL(perms)

	err := acl.check("admin", "/any.Service/AnyMethod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// extractClientIdentity tests

func TestExtractClientIdentity_Valid(t *testing.T) {
	ctx := contextWithCert("payment-service")

	identity, ok := extractClientIdentity(ctx)
	if !ok {
		t.Fatal("expected ok for valid cert")
	}
	if identity != "payment-service" {
		t.Fatalf("expected 'payment-service', got %q", identity)
	}
}

func TestExtractClientIdentity_NoPeer(t *testing.T) {
	_, ok := extractClientIdentity(context.Background())
	if ok {
		t.Fatal("expected false for context without peer")
	}
}

func TestExtractClientIdentity_NoTLSInfo(t *testing.T) {
	p := &peer.Peer{} // no AuthInfo
	ctx := peer.NewContext(context.Background(), p)

	_, ok := extractClientIdentity(ctx)
	if ok {
		t.Fatal("expected false for peer without TLS info")
	}
}

func TestExtractClientIdentity_NoCerts(t *testing.T) {
	tlsInfo := credentials.TLSInfo{
		State: tls.ConnectionState{PeerCertificates: nil},
	}
	p := &peer.Peer{AuthInfo: tlsInfo}
	ctx := peer.NewContext(context.Background(), p)

	_, ok := extractClientIdentity(ctx)
	if ok {
		t.Fatal("expected false for TLS info without certs")
	}
}

func TestExtractClientIdentity_EmptyCN(t *testing.T) {
	ctx := contextWithCert("")

	_, ok := extractClientIdentity(ctx)
	if ok {
		t.Fatal("expected false for empty CN")
	}
}

// matchMethodPattern tests

func TestMatchMethodPattern(t *testing.T) {
	tests := []struct {
		pattern    string
		fullMethod string
		want       bool
	}{
		{"/*", "/pkg.Svc/Method", true},
		{"/pkg.Svc/*", "/pkg.Svc/Method", true},
		{"/pkg.Svc/*", "/other.Svc/Method", false},
		{"/pkg.Svc/Method", "/pkg.Svc/Method", true},
		{"/pkg.Svc/Method", "/pkg.Svc/Other", false},
		{"/pkg.Svc/*", "/pkg.Svc/", true},
	}

	for _, tt := range tests {
		got := matchMethodPattern(tt.pattern, tt.fullMethod)
		if got != tt.want {
			t.Errorf("matchMethodPattern(%q, %q) = %v, want %v", tt.pattern, tt.fullMethod, got, tt.want)
		}
	}
}

// CertAuthInterceptor integration tests

func TestCertAuthInterceptor_Success(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/*"}},
	}

	interceptor := CertAuthInterceptor(perms)

	ctx := contextWithCert("user-service")
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	handlerCalled := false
	handler := func(ctx context.Context, req any) (any, error) {
		handlerCalled = true
		// Verify client identity is injected into context
		identity := ClientIdentityFromContext(ctx)
		if identity != "user-service" {
			t.Fatalf("expected 'user-service' in context, got %q", identity)
		}
		return "response", nil
	}

	resp, err := interceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler was not called")
	}
	if resp != "response" {
		t.Fatalf("expected 'response', got %v", resp)
	}
}

func TestCertAuthInterceptor_MissingCert(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/*"}},
	}

	interceptor := CertAuthInterceptor(perms)

	ctx := contextWithoutCert()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	handler := func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	_, err := interceptor(ctx, nil, info, handler)
	if err == nil {
		t.Fatal("expected error for missing cert")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestCertAuthInterceptor_UnauthorizedClient(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "allowed-service", AllowedMethods: []string{"/*"}},
	}

	interceptor := CertAuthInterceptor(perms)

	ctx := contextWithCert("unknown-service")
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	handler := func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	_, err := interceptor(ctx, nil, info, handler)
	if err == nil {
		t.Fatal("expected error for unauthorized client")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
}

// StreamCertAuthInterceptor tests

func TestStreamCertAuthInterceptor_MissingCert(t *testing.T) {
	perms := []CertPermission{
		{ClientIdentity: "user-service", AllowedMethods: []string{"/*"}},
	}

	interceptor := StreamCertAuthInterceptor(perms)

	info := &grpc.StreamServerInfo{FullMethod: "/pkg.Svc/StreamMethod"}

	handler := func(srv any, ss grpc.ServerStream) error {
		t.Fatal("handler should not be called")
		return nil
	}

	err := interceptor(nil, &mockServerStream{ctx: contextWithoutCert()}, info, handler)
	if err == nil {
		t.Fatal("expected error for missing cert")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

// ClientIdentityFromContext tests

func TestClientIdentityFromContext_Set(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKeyClientIdentity{}, "test-svc")
	identity := ClientIdentityFromContext(ctx)
	if identity != "test-svc" {
		t.Fatalf("expected 'test-svc', got %q", identity)
	}
}

func TestClientIdentityFromContext_NotSet(t *testing.T) {
	identity := ClientIdentityFromContext(context.Background())
	if identity != "" {
		t.Fatalf("expected empty string, got %q", identity)
	}
}

// mockServerStream for stream interceptor tests
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}
