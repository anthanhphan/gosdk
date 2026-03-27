// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package configuration contains configuration types and defaults for the orianna gRPC server.
package configuration

import "time"

// Default server configuration values
const (
	DefaultPort                         = 50051
	DefaultMaxRecvMsgSize               = 4 * 1024 * 1024 // 4MB
	DefaultMaxSendMsgSize               = 4 * 1024 * 1024 // 4MB
	DefaultMaxConcurrentStreams         = 256
	DefaultConnectionTimeout            = 120 * time.Second
	DefaultKeepaliveTime                = 2 * time.Hour
	DefaultKeepaliveTimeout             = 20 * time.Second
	DefaultGracefulShutdownTimeout      = 30 * time.Second
	DefaultHealthCheckTimeout           = 5 * time.Second
	DefaultKeepaliveMinTime             = 10 * time.Second
	DefaultKeepalivePermitWithoutStream = false
	DefaultMaxPayloadLogSize            = 1024 // bytes
)

// Health check defaults
const (
	DefaultHealthCheckerName = "grpc"
)
