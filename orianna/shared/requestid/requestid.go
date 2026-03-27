// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package requestid provides shared request ID utilities used by both
// HTTP middleware and gRPC interceptors.
package requestid

import "github.com/google/uuid"

// MaxLength is the maximum allowed length for an incoming request ID.
const MaxLength = 128

// IsValid validates that an incoming request ID is safe.
// Accepts alphanumeric characters, dashes, and underscores, up to MaxLength.
// This prevents log injection attacks via crafted X-Request-ID headers.
func IsValid(id string) bool {
	if len(id) == 0 || len(id) > MaxLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// Generate creates a new UUID v7 request ID, falling back to v4 on error.
func Generate() string {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return uuidV7.String()
}
