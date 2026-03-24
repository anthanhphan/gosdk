// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"
)

// Compile-time interface compliance checks.
var (
	_ Client = (*otelClient)(nil)
	_ Span   = (*otelSpanWrapper)(nil)
)

// ============================================================================
// OTel Client
// ============================================================================

// otelClient is the OpenTelemetry-backed implementation of the Client interface.
type otelClient struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// NewClient creates a new OpenTelemetry tracing client that exports spans
// via OTLP gRPC to a collector (e.g., Tempo, Jaeger, Grafana Alloy).
//
// Input:
//   - serviceName: Name of the service, used as the `service.name` resource attribute
//   - opts: Functional options for configuration
//
// Output:
//   - Client: The tracing client
//   - error: Non-nil if the OTLP exporter fails to initialize
//
// Example:
//
//	client, err := tracing.NewClient("my-service",
//	    tracing.WithEndpoint("otel-collector:4317"),
//	    tracing.WithInsecure(),
//	    tracing.WithEnvironment("production"),
//	    tracing.WithSamplingRate(0.1),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Shutdown(context.Background())
func NewClient(serviceName string, opts ...Option) (Client, error) {
	options := defaultClientOptions()
	options.serviceName = serviceName
	for _, opt := range opts {
		opt(options)
	}

	ctx := context.Background()

	// Build OTLP gRPC exporter options
	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(options.endpoint),
	}
	if options.insecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}
	if len(options.headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithHeaders(options.headers))
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("tracing: failed to create OTLP exporter: %w", err)
	}

	// Build resource attributes
	res, err := buildResource(options)
	if err != nil {
		return nil, fmt.Errorf("tracing: failed to create resource: %w", err)
	}

	// Build TracerProvider options
	providerOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	}
	if options.sampler != nil {
		providerOpts = append(providerOpts, sdktrace.WithSampler(options.sampler))
	}

	// Create TracerProvider
	provider := sdktrace.NewTracerProvider(providerOpts...)

	// Set global TracerProvider and propagators
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := provider.Tracer(serviceName)

	return &otelClient{
		provider: provider,
		tracer:   tracer,
	}, nil
}

// StartSpan starts a new span and returns the updated context and span handle.
func (c *otelClient) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	so := &spanOptions{}
	for _, opt := range opts {
		opt(so)
	}

	// Build OTel start options
	startOpts := []trace.SpanStartOption{
		trace.WithSpanKind(so.kind.toOTelSpanKind()),
	}
	if len(so.attributes) > 0 {
		startOpts = append(startOpts, trace.WithAttributes(so.attributes...))
	}

	ctx, span := c.tracer.Start(ctx, name, startOpts...)
	return ctx, &otelSpanWrapper{span: span}
}

// Shutdown flushes pending spans and releases resources.
func (c *otelClient) Shutdown(ctx context.Context) error {
	return c.provider.Shutdown(ctx)
}

// Tracer returns the underlying OTel tracer.
func (c *otelClient) Tracer() trace.Tracer {
	return c.tracer
}

// ============================================================================
// OTel Span
// ============================================================================

// otelSpanWrapper wraps an OpenTelemetry span.
type otelSpanWrapper struct {
	span trace.Span
}

func (s *otelSpanWrapper) End() {
	s.span.End()
}

func (s *otelSpanWrapper) SetAttributes(attrs ...attribute.KeyValue) {
	s.span.SetAttributes(attrs...)
}

func (s *otelSpanWrapper) SetStatus(code codes.Code, description string) {
	s.span.SetStatus(code, description)
}

func (s *otelSpanWrapper) RecordError(err error) {
	s.span.RecordError(err)
}

func (s *otelSpanWrapper) SetName(name string) {
	s.span.SetName(name)
}

func (s *otelSpanWrapper) AddEvent(name string, attrs ...attribute.KeyValue) {
	s.span.AddEvent(name, trace.WithAttributes(attrs...))
}

func (s *otelSpanWrapper) SpanContext() trace.SpanContext {
	return s.span.SpanContext()
}

// ============================================================================
// Helpers
// ============================================================================

// buildResource builds the OpenTelemetry resource with service metadata.
func buildResource(opts *clientOptions) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(opts.serviceName),
	}

	if opts.serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(opts.serviceVersion))
	}
	if opts.environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment", opts.environment))
	}

	// Use NewSchemaless to avoid schema URL conflicts between semconv versions
	// and the schema URL embedded in resource.Default().
	return resource.Merge(
		resource.Default(),
		resource.NewSchemaless(attrs...),
	)
}
