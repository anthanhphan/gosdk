// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package server

import (
	"errors"
	"fmt"
	"sync"

	"google.golang.org/grpc"
)

// Service errors
var (
	ErrDuplicateService  = errors.New("duplicate service")
	ErrNilServiceDesc    = errors.New("service descriptor cannot be nil")
	ErrNilImplementation = errors.New("service implementation cannot be nil")
)

// ServiceDesc represents a gRPC service registration with metadata.
type ServiceDesc struct {
	// Desc is the underlying gRPC service descriptor.
	Desc *grpc.ServiceDesc

	// Impl is the service implementation.
	Impl any

	// UnaryInterceptors are interceptors applied to this service's unary RPCs.
	UnaryInterceptors []grpc.UnaryServerInterceptor

	// StreamInterceptors are interceptors applied to this service's stream RPCs.
	StreamInterceptors []grpc.StreamServerInterceptor
}

// ServiceBuilder provides a fluent interface for building service descriptors.
type ServiceBuilder struct {
	service *ServiceDesc
}

// NewService creates a new service builder with the given gRPC service descriptor and implementation.
func NewService(desc *grpc.ServiceDesc, impl any) *ServiceBuilder {
	return &ServiceBuilder{
		service: &ServiceDesc{
			Desc:               desc,
			Impl:               impl,
			UnaryInterceptors:  nil,
			StreamInterceptors: nil,
		},
	}
}

// UnaryInterceptor adds unary interceptors to the service.
func (sb *ServiceBuilder) UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) *ServiceBuilder {
	sb.service.UnaryInterceptors = append(sb.service.UnaryInterceptors, interceptors...)
	return sb
}

// StreamInterceptor adds stream interceptors to the service.
func (sb *ServiceBuilder) StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) *ServiceBuilder {
	sb.service.StreamInterceptors = append(sb.service.StreamInterceptors, interceptors...)
	return sb
}

// Build returns the constructed ServiceDesc.
func (sb *ServiceBuilder) Build() *ServiceDesc {
	return sb.service
}

// ServiceRegistry manages service registration and validation.
// It is safe for concurrent use.
type ServiceRegistry struct {
	mu                 sync.RWMutex
	services           []ServiceDesc
	registeredServices map[string]struct{} // tracks service names to detect duplicates
}

// NewServiceRegistry creates a new service registry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:           nil,
		registeredServices: make(map[string]struct{}),
	}
}

// RegisterService registers a single service.
func (sr *ServiceRegistry) RegisterService(service ServiceDesc) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.registerServiceLocked(&service)
}

// RegisterServices registers multiple services atomically.
func (sr *ServiceRegistry) RegisterServices(services ...ServiceDesc) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for i := range services {
		if err := sr.registerServiceLocked(&services[i]); err != nil {
			return fmt.Errorf("service at index %d: %w", i, err)
		}
	}
	return nil
}

// registerServiceLocked registers a single service. Caller must hold mu.Lock().
func (sr *ServiceRegistry) registerServiceLocked(service *ServiceDesc) error {
	if err := validateService(service); err != nil {
		return fmt.Errorf("invalid service: %w", err)
	}

	// Check for duplicate services
	serviceName := service.Desc.ServiceName
	if _, exists := sr.registeredServices[serviceName]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateService, serviceName)
	}
	sr.registeredServices[serviceName] = struct{}{}

	sr.services = append(sr.services, *service)
	return nil
}

// GetServices returns a copy of all registered services.
func (sr *ServiceRegistry) GetServices() []ServiceDesc {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	out := make([]ServiceDesc, len(sr.services))
	copy(out, sr.services)
	return out
}

// validateService validates a service descriptor.
func validateService(service *ServiceDesc) error {
	if service == nil {
		return errors.New("service cannot be nil")
	}
	if service.Desc == nil {
		return ErrNilServiceDesc
	}
	if service.Impl == nil {
		return ErrNilImplementation
	}
	return nil
}
