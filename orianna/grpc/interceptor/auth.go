// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/anthanhphan/gosdk/logger"
)

// Context key for client identity (injected by auth interceptor for downstream audit).
type ctxKeyClientIdentity struct{}

// ClientIdentityFromContext extracts the authenticated client identity from the context.
// Returns empty string if no identity was set (e.g., auth interceptor not in chain).
func ClientIdentityFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKeyClientIdentity{}).(string); ok {
		return id
	}
	return ""
}

// CertPermission maps a certificate identity (CN) to allowed gRPC methods.
type CertPermission struct {
	ClientIdentity string
	AllowedMethods []string
}

// Pre-allocated auth error sentinels — avoids per-request status.Error allocation.
var (
	errAuthPermissionDenied = status.Error(codes.PermissionDenied, "permission denied")
	errAuthRequired         = status.Error(codes.Unauthenticated, "authentication required")
)

// certACL maps certificate CN to allowed method patterns.
// Fully stateless: the client certificate IS the identity.
// Pre-builds exact match maps at construction time for O(1) lookups;
// only falls back to pattern matching for wildcards.
type certACL struct {
	exactPerms   map[string]map[string]struct{} // CN -> exact methods (O(1))
	patternPerms map[string][]string            // CN -> wildcard patterns (slow path)
	logger       *logger.Logger
}

// newCertACL builds a certificate-based access control list.
// Splits exact methods and wildcard patterns at construction time.
func newCertACL(perms []CertPermission) *certACL {
	exact := make(map[string]map[string]struct{}, len(perms))
	patterns := make(map[string][]string, len(perms))
	for _, p := range perms {
		for _, m := range p.AllowedMethods {
			if m == "/*" || (len(m) > 1 && m[len(m)-1] == '*') {
				patterns[p.ClientIdentity] = append(patterns[p.ClientIdentity], m)
			} else {
				if exact[p.ClientIdentity] == nil {
					exact[p.ClientIdentity] = make(map[string]struct{}, len(p.AllowedMethods))
				}
				exact[p.ClientIdentity][m] = struct{}{}
			}
		}
	}
	return &certACL{
		exactPerms:   exact,
		patternPerms: patterns,
		logger:       logger.NewLoggerWithFields(logger.String("component", "cert-auth")),
	}
}

// check validates if the given client identity is authorized to call the method.
// Returns nil on success, or a gRPC status error.
// Security: error messages are generic to avoid leaking internal client identities.
// Detailed information is logged server-side for audit.
// Uses O(1) exact match first, falls back to O(n) pattern match only for wildcards.
func (acl *certACL) check(clientIdentity, fullMethod string) error {
	// Fast path: O(1) exact method match
	if exactMethods, ok := acl.exactPerms[clientIdentity]; ok {
		if _, found := exactMethods[fullMethod]; found {
			acl.logger.Debugw("auth granted",
				"client", clientIdentity,
				"method", fullMethod,
				"matched_pattern", fullMethod,
			)
			return nil
		}
	}

	// Slow path: pattern matching for wildcards
	if patterns, ok := acl.patternPerms[clientIdentity]; ok {
		for _, pattern := range patterns {
			if matchMethodPattern(pattern, fullMethod) {
				acl.logger.Debugw("auth granted",
					"client", clientIdentity,
					"method", fullMethod,
					"matched_pattern", pattern,
				)
				return nil
			}
		}
	}

	// If no match at all (unknown identity or method not allowed)
	if _, hasExact := acl.exactPerms[clientIdentity]; !hasExact {
		if _, hasPattern := acl.patternPerms[clientIdentity]; !hasPattern {
			acl.logger.Warnw("auth denied: unknown client identity",
				"client", clientIdentity,
				"method", fullMethod,
			)
			return errAuthPermissionDenied
		}
	}

	acl.logger.Warnw("auth denied: method not allowed",
		"client", clientIdentity,
		"method", fullMethod,
	)
	return errAuthPermissionDenied
}

// extractClientIdentity extracts the Common Name (CN) from the peer's TLS certificate.
// Returns the CN and true on success, or empty string and false if no cert is available.
func extractClientIdentity(ctx context.Context) (string, bool) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", false
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", false
	}

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return "", false
	}

	cn := tlsInfo.State.PeerCertificates[0].Subject.CommonName
	if cn == "" {
		return "", false
	}

	return cn, true
}

// matchMethodPattern matches a gRPC method against a pattern.
// Supported patterns:
//   - "/*" matches everything
//   - "/package.Service/*" matches all methods of a service
//   - "/package.Service/Method" exact match
func matchMethodPattern(pattern, fullMethod string) bool {
	if pattern == "/*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := pattern[:len(pattern)-1] // strip "*", keep "/"
		return strings.HasPrefix(fullMethod, prefix)
	}
	return pattern == fullMethod
}

// CertAuthInterceptor creates a unary interceptor that:
//  1. Extracts client identity (CN) from the mTLS peer certificate
//  2. Checks if the client is allowed to call the requested method
//  3. Injects the client identity into the context for downstream audit
//
// Fully stateless: no tokens, the client certificate IS the identity.
// Error messages are sanitized — no internal identity information is leaked to callers.
func CertAuthInterceptor(permissions []CertPermission) grpc.UnaryServerInterceptor {
	acl := newCertACL(permissions)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		identity, ok := extractClientIdentity(ctx)
		if !ok {
			acl.logger.Warnw("auth failed: no client certificate",
				"method", info.FullMethod,
			)
			return nil, errAuthRequired
		}

		if err := acl.check(identity, info.FullMethod); err != nil {
			return nil, err
		}

		// Inject client identity into context for downstream audit
		ctx = context.WithValue(ctx, ctxKeyClientIdentity{}, identity)
		return handler(ctx, req)
	}
}

// StreamCertAuthInterceptor creates a stream interceptor that:
//  1. Extracts client identity (CN) from the mTLS peer certificate
//  2. Checks if the client is allowed to call the requested method
//  3. Injects the client identity into the stream context for downstream audit
//
// Fully stateless: no tokens, the client certificate IS the identity.
// Error messages are sanitized — no internal identity information is leaked to callers.
func StreamCertAuthInterceptor(permissions []CertPermission) grpc.StreamServerInterceptor {
	acl := newCertACL(permissions)
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		identity, ok := extractClientIdentity(ss.Context())
		if !ok {
			acl.logger.Warnw("auth failed: no client certificate",
				"method", info.FullMethod,
			)
			return errAuthRequired
		}

		if err := acl.check(identity, info.FullMethod); err != nil {
			return err
		}

		// Inject client identity into context for downstream audit
		ctx := context.WithValue(ss.Context(), ctxKeyClientIdentity{}, identity)
		wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrappedStream)
	}
}
