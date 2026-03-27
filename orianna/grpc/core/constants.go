// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "github.com/anthanhphan/gosdk/orianna/shared/ctxkeys"

// Context keys — re-exported from shared/ctxkeys for backward compatibility.
var (
	ContextKeyRequestID = ctxkeys.RequestID
	ContextKeyTraceID   = ctxkeys.TraceID
	ContextKeyUserID    = ctxkeys.UserID
)

// Metadata keys for gRPC metadata
const (
	MetadataKeyRequestID = "x-request-id"
	MetadataKeyTraceID   = "x-trace-id"
	MetadataKeyUserAgent = "user-agent"
)

// Default values
const (
	DefaultUnknownRequestID       = "unknown"
	HealthMessageHealthy          = "gRPC endpoint is healthy"
	HealthMessageUnhealthy        = "gRPC endpoint is unhealthy: %v"
	HealthMessageCreateConnFailed = "Failed to create connection: %v"
)
