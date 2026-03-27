package ctxkeys

import (
	"context"
	"testing"
)

func TestContextKeys(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		key      Key
		name     string
		expected string
	}{
		{RequestID, "request_id", "orianna.request_id"},
		{TraceID, "trace_id", "orianna.trace_id"},
		{UserID, "user_id", "orianna.user_id"},
		{TenantID, "tenant_id", "orianna.tenant_id"},
		{CorrelationID, "correlation_id", "orianna.correlation_id"},
	}

	for _, tt := range tests {
		val := tt.name + "-val"
		ctxWithVal := context.WithValue(ctx, tt.key, val)
		got := ctxWithVal.Value(tt.key)
		if got != val {
			t.Errorf("expected %s, got %v", val, got)
		}

		if tt.key.String() != tt.expected {
			t.Errorf("String() = %s, want %s", tt.key.String(), tt.expected)
		}
		if tt.key.Key() != tt.name {
			t.Errorf("Key() = %s, want %s", tt.key.Key(), tt.name)
		}
	}
}
