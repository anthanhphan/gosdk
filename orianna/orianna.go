// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package orianna is a unified multi-protocol server framework.
//
// For HTTP/REST server:
//   - Server: import "github.com/anthanhphan/gosdk/orianna/http/server"
//   - Client: import "github.com/anthanhphan/gosdk/orianna/http/client"
//   - Routing: import "github.com/anthanhphan/gosdk/orianna/http/routing"
//   - Middleware: import "github.com/anthanhphan/gosdk/orianna/http/middleware"
//   - Configuration: import "github.com/anthanhphan/gosdk/orianna/http/configuration"
//   - Validation: import "github.com/anthanhphan/gosdk/validator"
//
// For gRPC:
//   - Server: import "github.com/anthanhphan/gosdk/orianna/grpc/server"
//   - Client: import "github.com/anthanhphan/gosdk/orianna/grpc/client"
//   - Interceptors: import "github.com/anthanhphan/gosdk/orianna/grpc/interceptor"
//   - Configuration: import "github.com/anthanhphan/gosdk/orianna/grpc/configuration"
//
// For shared utilities:
//   - Health: import "github.com/anthanhphan/gosdk/orianna/shared/health"
package orianna

// Version is the current version of the package.
// Override at build time with:
//
//	-ldflags '-X github.com/anthanhphan/gosdk/orianna.Version=v2.0.0'
var Version = "1.0.0"
