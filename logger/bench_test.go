package logger

import (
	"io"
	"testing"
	"time"
)

// Baseline benchmarks to measure current performance before optimization.

var benchConfig = &Config{
	LogLevel:          LevelInfo,
	LogEncoding:       EncodingJSON,
	DisableCaller:     false,
	DisableStacktrace: true,
}

func BenchmarkLogger_Infow_Baseline(b *testing.B) {
	logger := NewLogger(benchConfig, []io.Writer{io.Discard}, String("package", "transport"))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Infow("incoming request",
			"trace_id", "abc123def456",
			"request_id", "req-001",
			"method", "GET",
			"path", "/api/users",
			"ip", "127.0.0.1",
			"http_code", 200,
			"duration_ms", 15,
		)
	}
}

func BenchmarkLogger_Infow_ManyFields(b *testing.B) {
	logger := NewLogger(benchConfig, []io.Writer{io.Discard}, String("package", "transport"))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Infow("request completed",
			"trace_id", "abc123def456",
			"request_id", "req-001",
			"method", "POST",
			"path", "/api/users",
			"ip", "127.0.0.1",
			"user-agent", "curl/8.7.1",
			"http_code", 201,
			"duration_ms", 33,
			"response", `{"http_status":201,"code":"SUCCESS"}`,
			"headers", map[string]string{
				"content-type": "application/json",
				"accept":       "*/*",
				"host":         "localhost:8080",
			},
			"body", `{"name":"Test","email":"test@ex.com"}`,
		)
	}
}

func BenchmarkJSONEncoder_Encode(b *testing.B) {
	encoder := newJSONEncoder(benchConfig)
	entry := &Entry{
		Time:          time.Now(),
		Level:         LevelInfo,
		Message:       "incoming request",
		CallerFile:    "middleware.go",
		CallerLine:    142,
		CallerDefined: true,
		Fields: []Field{
			String("trace_id", "abc123def456"),
			String("request_id", "req-001"),
			String("method", "GET"),
			String("path", "/api/users"),
			String("ip", "127.0.0.1"),
			Int("http_code", 200),
			Int("duration_ms", 15),
			String("package", "transport"),
		},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = encoder.Encode(entry)
	}
}

func BenchmarkLogger_Parallel(b *testing.B) {
	logger := NewLogger(benchConfig, []io.Writer{io.Discard}, String("package", "transport"))
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow("incoming request",
				"trace_id", "abc123def456",
				"request_id", "req-001",
				"method", "GET",
				"path", "/api/users",
				"http_code", 200,
			)
		}
	})
}
