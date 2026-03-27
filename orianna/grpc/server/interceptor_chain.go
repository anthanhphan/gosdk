// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"context"

	"google.golang.org/grpc"

	"github.com/anthanhphan/gosdk/orianna/grpc/interceptor"
)

// buildUnaryInterceptorChain builds the ordered unary interceptor chain.
func (s *Server) buildUnaryInterceptorChain(certPerms []interceptor.CertPermission) []grpc.UnaryServerInterceptor {
	chain := make([]grpc.UnaryServerInterceptor, 0, 8+len(s.globalUnary))

	// Recovery first (outermost)
	if !s.disableRecovery {
		if s.panicRecover != nil {
			chain = append(chain, s.panicRecover)
		} else {
			chain = append(chain, interceptor.Recover())
		}
	}

	// Request ID generation (after recovery, ensures every request has an ID)
	chain = append(chain, interceptor.RequestIDInterceptor())

	// Certificate-based auth interceptor (after recovery, before token auth)
	if len(certPerms) > 0 {
		chain = append(chain, interceptor.CertAuthInterceptor(certPerms))
	}

	// Token-based auth interceptor (after cert auth, before rate limiter)
	if s.tokenValidator != nil {
		chain = append(chain, interceptor.TokenAuthInterceptor(s.tokenValidator))
	}

	// Rate limiter
	if s.rateLimiter != nil {
		chain = append(chain, s.rateLimiter)
	}

	// Metrics
	if s.metricsClient != nil {
		chain = append(chain, interceptor.MetricsInterceptor(s.metricsClient, s.config.ServiceName))
	}

	// Slow RPC detection (config-driven, non-zero = enabled)
	if s.config.SlowRequestThreshold > 0 {
		chain = append(chain, interceptor.SlowRPCDetector(s.config.SlowRequestThreshold))
	}

	// Tracing
	if s.tracingClient != nil {
		chain = append(chain, interceptor.TracingInterceptor(s.tracingClient))
	}

	// Verbose logging
	if s.config.VerboseLogging {
		chain = append(chain, interceptor.VerboseLoggingInterceptor(s.config.VerboseLoggingSkipMethods))
	}

	// Global interceptors
	chain = append(chain, s.globalUnary...)

	return chain
}

// buildStreamInterceptorChain builds the ordered stream interceptor chain.
func (s *Server) buildStreamInterceptorChain(certPerms []interceptor.CertPermission) []grpc.StreamServerInterceptor {
	chain := make([]grpc.StreamServerInterceptor, 0, 7+len(s.globalStream))

	// Recovery first
	if !s.disableRecovery {
		if s.streamRecover != nil {
			chain = append(chain, s.streamRecover)
		} else {
			chain = append(chain, interceptor.StreamRecover())
		}
	}

	// Request ID generation
	chain = append(chain, interceptor.StreamRequestIDInterceptor())

	// Certificate-based auth interceptor
	if len(certPerms) > 0 {
		chain = append(chain, interceptor.StreamCertAuthInterceptor(certPerms))
	}

	// Token-based auth interceptor (after cert auth, before metrics)
	if s.tokenValidator != nil {
		chain = append(chain, interceptor.StreamTokenAuthInterceptor(s.tokenValidator))
	}

	// Metrics
	if s.metricsClient != nil {
		chain = append(chain, interceptor.StreamMetricsInterceptor(s.metricsClient, s.config.ServiceName))
	}

	// Slow stream detection (config-driven, non-zero = enabled)
	if s.config.SlowRequestThreshold > 0 {
		chain = append(chain, interceptor.StreamSlowRPCDetector(s.config.SlowRequestThreshold))
	}

	// Tracing
	if s.tracingClient != nil {
		chain = append(chain, interceptor.StreamTracingInterceptor(s.tracingClient))
	}

	// Verbose logging
	if s.config.VerboseLogging {
		chain = append(chain, interceptor.StreamVerboseLoggingInterceptor(s.config.VerboseLoggingSkipMethods))
	}

	// Global stream interceptors
	chain = append(chain, s.globalStream...)

	return chain
}

// registerServiceWithInterceptors registers a service and wraps its handlers
// with per-service interceptors.
func (s *Server) registerServiceWithInterceptors(svc ServiceDesc) {
	if len(svc.UnaryInterceptors) == 0 {
		s.grpcServer.RegisterService(svc.Desc, svc.Impl)
		return
	}

	// Build the per-service interceptor chain
	chain := interceptor.Chain(svc.UnaryInterceptors...)

	// Create a new ServiceDesc that wraps each unary handler with the chain.
	// IMPORTANT: We pass nil as the interceptor parameter to originalHandler
	// because global interceptors have already been executed by the gRPC
	// framework before reaching this handler. Passing the interceptor through
	// would cause the global chain to execute twice.
	wrappedDesc := *svc.Desc
	wrappedMethods := make([]grpc.MethodDesc, len(svc.Desc.Methods))
	for i, m := range svc.Desc.Methods {
		originalHandler := m.Handler
		methodName := m.MethodName // capture for closure
		wrappedMethods[i] = grpc.MethodDesc{
			MethodName: methodName,
			Handler: func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
				handler := func(ctx context.Context, req any) (any, error) {
					return originalHandler(srv, ctx, dec, nil)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/" + svc.Desc.ServiceName + "/" + methodName,
				}
				return chain(ctx, nil, info, handler)
			},
		}
	}
	wrappedDesc.Methods = wrappedMethods
	s.grpcServer.RegisterService(&wrappedDesc, svc.Impl)
}
