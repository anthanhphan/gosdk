package keys

import (
	"testing"
)

func TestHTTPHeaders(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{
			name:     "RequestIDHeader should be X-Request-ID",
			constant: RequestIDHeader,
			want:     "X-Request-ID",
		},
		{
			name:     "TraceIDHeader should be X-Trace-ID",
			constant: TraceIDHeader,
			want:     "X-Trace-ID",
		},
		{
			name:     "TraceParentHeader should be traceparent",
			constant: TraceParentHeader,
			want:     "traceparent",
		},
		{
			name:     "B3TraceIDHeader should be X-B3-TraceId",
			constant: B3TraceIDHeader,
			want:     "X-B3-TraceId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("Constant = %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestContextKeys(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{
			name:     "ContextKeyRequestID should be request_id",
			constant: ContextKeyRequestID,
			want:     "request_id",
		},
		{
			name:     "ContextKeyTraceID should be trace_id",
			constant: ContextKeyTraceID,
			want:     "trace_id",
		},
		{
			name:     "ContextKeyLanguage should be lang",
			constant: ContextKeyLanguage,
			want:     "lang",
		},
		{
			name:     "ContextKeyUserID should be user_id",
			constant: ContextKeyUserID,
			want:     "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("Constant = %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestInternalKeys(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{
			name:     "TrackedLocalsKey should be __aurelion_tracked_locals__",
			constant: TrackedLocalsKey,
			want:     "__aurelion_tracked_locals__",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("Constant = %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestConstantsAreNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		constant string
	}{
		{
			name:     "RequestIDHeader should not be empty",
			constant: RequestIDHeader,
		},
		{
			name:     "TraceIDHeader should not be empty",
			constant: TraceIDHeader,
		},
		{
			name:     "TraceParentHeader should not be empty",
			constant: TraceParentHeader,
		},
		{
			name:     "B3TraceIDHeader should not be empty",
			constant: B3TraceIDHeader,
		},
		{
			name:     "ContextKeyRequestID should not be empty",
			constant: ContextKeyRequestID,
		},
		{
			name:     "ContextKeyTraceID should not be empty",
			constant: ContextKeyTraceID,
		},
		{
			name:     "ContextKeyLanguage should not be empty",
			constant: ContextKeyLanguage,
		},
		{
			name:     "ContextKeyUserID should not be empty",
			constant: ContextKeyUserID,
		},
		{
			name:     "TrackedLocalsKey should not be empty",
			constant: TrackedLocalsKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == "" {
				t.Error("Constant should not be empty")
			}
		})
	}
}
