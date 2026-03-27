// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Pre-allocated token auth error sentinels — avoids per-request status.Error allocation.
var (
	errMissingMetadata   = status.Error(codes.Unauthenticated, "missing metadata")
	errMissingAuthHeader = status.Error(codes.Unauthenticated, "missing authorization header")
	errInvalidAuthFormat = status.Error(codes.Unauthenticated, "invalid authorization format, expected 'Bearer <token>'")
	errEmptyBearerToken  = status.Error(codes.Unauthenticated, "empty bearer token")
)

// TokenValidator validates a bearer token and returns the claims.
// Implement this interface to integrate with JWT, OAuth2, or any token-based auth.
type TokenValidator interface {
	// Validate validates the given token string and returns claims on success.
	// Returns an error if the token is invalid, expired, or malformed.
	Validate(ctx context.Context, token string) (claims map[string]any, err error)
}

// Context key for token claims (injected by token auth interceptor for downstream use).
type ctxKeyTokenClaims struct{}

// TokenClaimsFromContext extracts the authenticated token claims from the context.
// Returns nil if no claims were set (e.g., token auth interceptor not in chain).
func TokenClaimsFromContext(ctx context.Context) map[string]any {
	if claims, ok := ctx.Value(ctxKeyTokenClaims{}).(map[string]any); ok {
		return claims
	}
	return nil
}

// TokenAuthInterceptor creates a unary interceptor that:
//  1. Extracts the bearer token from the "authorization" metadata
//  2. Validates the token using the provided TokenValidator
//  3. Injects the token claims into the context for downstream use
//
// The token is expected in the format "Bearer <token>" in the "authorization" metadata.
//
// Example:
//
//	srv := server.NewServer(config,
//	    orianna.WithGrpcGlobalUnaryInterceptor(
//	        interceptor.TokenAuthInterceptor(myJWTValidator),
//	    ),
//	)
func TokenAuthInterceptor(validator TokenValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token, err := extractBearerToken(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := validator.Validate(ctx, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, ctxKeyTokenClaims{}, claims)
		return handler(ctx, req)
	}
}

// StreamTokenAuthInterceptor creates a stream interceptor that:
//  1. Extracts the bearer token from the "authorization" metadata
//  2. Validates the token using the provided TokenValidator
//  3. Injects the token claims into the stream context for downstream use
func StreamTokenAuthInterceptor(validator TokenValidator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := extractBearerToken(ss.Context())
		if err != nil {
			return err
		}

		claims, err := validator.Validate(ss.Context(), token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx := context.WithValue(ss.Context(), ctxKeyTokenClaims{}, claims)
		wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrappedStream)
	}
}

// extractBearerToken extracts the bearer token from gRPC metadata.
// Uses strings.CutPrefix for single-pass prefix stripping.
func extractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}

	authValues := md.Get("authorization")
	if len(authValues) == 0 {
		return "", errMissingAuthHeader
	}

	token, found := strings.CutPrefix(authValues[0], "Bearer ")
	if !found {
		return "", errInvalidAuthFormat
	}

	if token == "" {
		return "", errEmptyBearerToken
	}

	return token, nil
}
