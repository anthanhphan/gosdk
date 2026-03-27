// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package observability

// Standard metric name suffixes used across both HTTP and gRPC.
// Using shared constants prevents drift between protocols and ensures
// consistent Prometheus/Grafana dashboard queries.
const (
	// SuffixRequestsTotal is the suffix for the request counter metric.
	SuffixRequestsTotal = "_requests_total"

	// SuffixRequestDurationSeconds is the suffix for the request duration histogram.
	SuffixRequestDurationSeconds = "_request_duration_seconds"

	// SuffixInFlightRequests is the suffix for the in-flight gauge.
	SuffixInFlightRequests = "_in_flight_requests"

	// SuffixStreamsTotal is the suffix for the stream counter (gRPC only).
	SuffixStreamsTotal = "_streams_total"

	// SuffixStreamDurationSeconds is the suffix for the stream duration histogram (gRPC only).
	SuffixStreamDurationSeconds = "_stream_duration_seconds"
)
