// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/anthanhphan/gosdk/orianna/grpc/interceptor"
)

// buildGRPCServerOptions builds the gRPC server options from configuration.
func (s *Server) buildGRPCServerOptions() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	// Message size limits
	opts = append(opts, grpc.MaxRecvMsgSize(s.config.MaxRecvMsgSize))
	opts = append(opts, grpc.MaxSendMsgSize(s.config.MaxSendMsgSize))

	// Concurrent streams
	if s.config.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(s.config.MaxConcurrentStreams))
	}

	// Keepalive
	kaParams := keepalive.ServerParameters{}
	if s.config.KeepaliveTime != nil {
		kaParams.Time = *s.config.KeepaliveTime
	}
	if s.config.KeepaliveTimeout != nil {
		kaParams.Timeout = *s.config.KeepaliveTimeout
	}
	opts = append(opts, grpc.KeepaliveParams(kaParams))

	// Keepalive enforcement policy -- protects against client ping floods
	// and prevents idle connections from being held indefinitely.
	kaEnforcement := keepalive.EnforcementPolicy{
		MinTime:             *s.config.KeepaliveMinTime,
		PermitWithoutStream: *s.config.KeepalivePermitWithoutStream,
	}
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(kaEnforcement))

	// Connection timeout
	if s.config.ConnectionTimeout != nil {
		opts = append(opts, grpc.ConnectionTimeout(*s.config.ConnectionTimeout))
	}

	// Buffer sizes
	if s.config.ReadBufferSize > 0 {
		opts = append(opts, grpc.ReadBufferSize(s.config.ReadBufferSize))
	}
	if s.config.WriteBufferSize > 0 {
		opts = append(opts, grpc.WriteBufferSize(s.config.WriteBufferSize))
	}

	// Flow-control window sizes
	if s.config.InitialWindowSize > 0 {
		opts = append(opts, grpc.InitialWindowSize(s.config.InitialWindowSize))
	}
	if s.config.InitialConnWindowSize > 0 {
		opts = append(opts, grpc.InitialConnWindowSize(s.config.InitialConnWindowSize))
	}

	// TLS (server-side only, no client cert required)
	if s.config.TLS != nil {
		creds, err := s.buildTLSCredentials()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// mTLS (mutual TLS, requires client certificate)
	if s.config.MTLS != nil {
		creds, err := s.buildMTLSCredentials()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Pre-build cert permissions once for both chains
	certPermissions := s.buildCertPermissions()

	// Build interceptor chains
	unaryInterceptors := s.buildUnaryInterceptorChain(certPermissions)
	streamInterceptors := s.buildStreamInterceptorChain(certPermissions)

	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	return opts, nil
}

// buildCertReloader returns a GetCertificate callback that reloads TLS certificates
// from disk on each TLS handshake, enabling certificate rotation without server restart.
func (s *Server) buildCertReloader(certFile, keyFile string) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			s.logger.Errorw("failed to reload TLS certificate",
				"cert_file", certFile,
				"error", err,
			)
			return nil, fmt.Errorf("failed to reload TLS certificate: %w", err)
		}
		return &cert, nil
	}
}

// buildMTLSCredentials builds mutual TLS credentials with hot-reload support.
// Always enforces client certificate verification (RequireAndVerifyClientCert).
func (s *Server) buildMTLSCredentials() (credentials.TransportCredentials, error) {
	mtlsConf := s.config.MTLS

	// Validate that initial cert files exist and are loadable
	if _, err := tls.LoadX509KeyPair(mtlsConf.CertFile, mtlsConf.KeyFile); err != nil {
		return nil, fmt.Errorf("failed to load initial TLS key pair: %w", err)
	}

	caCert, err := os.ReadFile(mtlsConf.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file: %w", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	tlsCfg := &tls.Config{
		GetCertificate: s.buildCertReloader(mtlsConf.CertFile, mtlsConf.KeyFile),
		ClientCAs:      certPool,
		ClientAuth:     tls.RequireAndVerifyClientCert,
		MinVersion:     tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsCfg), nil
}

// buildCertPermissions converts config ClientPermission entries to interceptor CertPermissions.
func (s *Server) buildCertPermissions() []interceptor.CertPermission {
	if len(s.config.Permissions) == 0 {
		return nil
	}
	perms := make([]interceptor.CertPermission, len(s.config.Permissions))
	for i, entry := range s.config.Permissions {
		perms[i] = interceptor.CertPermission{
			ClientIdentity: entry.ClientIdentity,
			AllowedMethods: entry.AllowedMethods,
		}
	}
	return perms
}

// buildTLSCredentials builds server-side TLS credentials with hot-reload support.
// Does NOT require client certificates (use buildMTLSCredentials for that).
func (s *Server) buildTLSCredentials() (credentials.TransportCredentials, error) {
	tlsConf := s.config.TLS

	// Validate that initial cert files exist and are loadable
	if _, err := tls.LoadX509KeyPair(tlsConf.CertFile, tlsConf.KeyFile); err != nil {
		return nil, fmt.Errorf("failed to load initial TLS key pair: %w", err)
	}

	tlsCfg := &tls.Config{
		GetCertificate: s.buildCertReloader(tlsConf.CertFile, tlsConf.KeyFile),
		ClientAuth:     tls.NoClientCert,
		MinVersion:     tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsCfg), nil
}
