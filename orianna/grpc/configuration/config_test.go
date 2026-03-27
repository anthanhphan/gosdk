// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package configuration

import (
	"testing"
	"time"
)

func TestValidateConfig_EmptyConfig(t *testing.T) {
	err := (&Config{}).Validate()
	if err == nil {
		t.Fatal("expected error for empty config (missing service_name)")
	}
}

func TestValidateConfig_MissingServiceName(t *testing.T) {
	err := (&Config{Port: 50051}).Validate()
	if err == nil {
		t.Fatal("expected error for missing service name")
	}
}

func TestValidateConfig_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"negative port", -1},
		{"port too high", 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&Config{ServiceName: "test", Port: tt.port}).Validate()
			if err == nil {
				t.Fatal("expected error for invalid port")
			}
		})
	}
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	err := (&Config{
		ServiceName: "test-service",
		Port:        50051,
	}).Validate()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateConfig_NegativeTimeouts(t *testing.T) {
	neg := -1 * time.Second

	tests := []struct {
		name   string
		config *Config
	}{
		{"negative connection timeout", &Config{ServiceName: "test", Port: 50051, ConnectionTimeout: &neg}},
		{"negative keepalive time", &Config{ServiceName: "test", Port: 50051, KeepaliveTime: &neg}},
		{"negative keepalive timeout", &Config{ServiceName: "test", Port: 50051, KeepaliveTimeout: &neg}},
		{"negative graceful shutdown timeout", &Config{ServiceName: "test", Port: 50051, GracefulShutdownTimeout: &neg}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Fatal("expected error for negative timeout")
			}
		})
	}
}

func TestValidateConfig_NegativeLimits(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{"negative recv size", &Config{ServiceName: "test", Port: 50051, MaxRecvMsgSize: -1}},
		{"negative send size", &Config{ServiceName: "test", Port: 50051, MaxSendMsgSize: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Fatal("expected error for negative limit")
			}
		})
	}
}

func TestValidateConfig_MTLS(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		hasErr bool
	}{
		{
			"mtls missing cert",
			&Config{ServiceName: "test", Port: 50051, MTLS: &MTLSConfig{KeyFile: "key.pem", CAFile: "ca.pem"}},
			true,
		},
		{
			"mtls missing key",
			&Config{ServiceName: "test", Port: 50051, MTLS: &MTLSConfig{CertFile: "cert.pem", CAFile: "ca.pem"}},
			true,
		},
		{
			"mtls missing ca",
			&Config{ServiceName: "test", Port: 50051, MTLS: &MTLSConfig{CertFile: "cert.pem", KeyFile: "key.pem"}},
			true,
		},
		{
			"valid mtls",
			&Config{ServiceName: "test", Port: 50051, MTLS: &MTLSConfig{CertFile: "cert.pem", KeyFile: "key.pem", CAFile: "ca.pem"}},
			false,
		},
		{
			"permissions without mtls",
			&Config{ServiceName: "test", Port: 50051, Permissions: []ClientPermission{{ClientIdentity: "svc", AllowedMethods: []string{"/*"}}}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.hasErr {
				t.Fatalf("expected hasErr=%v, got error: %v", tt.hasErr, err)
			}
		})
	}
}

func TestMergeConfigDefaults(t *testing.T) {
	conf := &Config{
		ServiceName: "test",
	}

	merged := MergeConfigDefaults(conf)

	// Should not mutate original
	if conf.Port != 0 {
		t.Fatal("original config should not be mutated")
	}

	// Should have defaults
	if merged.Port != DefaultPort {
		t.Fatalf("expected default port %d, got %d", DefaultPort, merged.Port)
	}
	if merged.MaxRecvMsgSize != DefaultMaxRecvMsgSize {
		t.Fatalf("expected default recv size %d, got %d", DefaultMaxRecvMsgSize, merged.MaxRecvMsgSize)
	}
	if merged.MaxSendMsgSize != DefaultMaxSendMsgSize {
		t.Fatalf("expected default send size %d, got %d", DefaultMaxSendMsgSize, merged.MaxSendMsgSize)
	}
	if merged.MaxConcurrentStreams != DefaultMaxConcurrentStreams {
		t.Fatalf("expected default concurrent streams %d, got %d", DefaultMaxConcurrentStreams, merged.MaxConcurrentStreams)
	}
	if merged.ConnectionTimeout == nil || *merged.ConnectionTimeout != DefaultConnectionTimeout {
		t.Fatal("expected default connection timeout")
	}
	if merged.KeepaliveTime == nil || *merged.KeepaliveTime != DefaultKeepaliveTime {
		t.Fatal("expected default keepalive time")
	}
	if merged.KeepaliveTimeout == nil || *merged.KeepaliveTimeout != DefaultKeepaliveTimeout {
		t.Fatal("expected default keepalive timeout")
	}
	if merged.GracefulShutdownTimeout == nil || *merged.GracefulShutdownTimeout != DefaultGracefulShutdownTimeout {
		t.Fatal("expected default graceful shutdown timeout")
	}
}

func TestMergeConfigDefaults_PreservesExistingValues(t *testing.T) {
	customTimeout := 60 * time.Second
	conf := &Config{
		ServiceName:       "test",
		Port:              9090,
		MaxRecvMsgSize:    8 * 1024 * 1024,
		ConnectionTimeout: &customTimeout,
	}

	merged := MergeConfigDefaults(conf)

	if merged.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", merged.Port)
	}
	if merged.MaxRecvMsgSize != 8*1024*1024 {
		t.Fatalf("expected custom recv size, got %d", merged.MaxRecvMsgSize)
	}
	if *merged.ConnectionTimeout != customTimeout {
		t.Fatalf("expected custom timeout, got %v", *merged.ConnectionTimeout)
	}
}

func TestValidateConfig_ZeroPort(t *testing.T) {
	// Port 0 is valid (binds to random port)
	err := (&Config{ServiceName: "test", Port: 0}).Validate()
	if err != nil {
		t.Fatalf("expected no error for port 0, got %v", err)
	}
}
