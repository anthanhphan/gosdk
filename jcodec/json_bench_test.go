// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
	"encoding/json"
	"testing"
	"time"
)

// Benchmark data structures

type benchUser struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

type benchConfig struct {
	DatabaseURL string            `json:"database_url"`
	Port        int               `json:"port"`
	Debug       bool              `json:"debug"`
	Features    []string          `json:"features"`
	Headers     map[string]string `json:"headers"`
	Timeout     int               `json:"timeout"`
}

type benchLargeStruct struct {
	Users    []benchUser            `json:"users"`
	Configs  []benchConfig          `json:"configs"`
	Metadata map[string]interface{} `json:"metadata"`
	Tags     []string               `json:"tags"`
	Settings map[string]benchConfig `json:"settings"`
}

// Test data generators

func getBenchUser() benchUser {
	return benchUser{
		ID:        12345,
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Active:    true,
	}
}

func getBenchConfig() benchConfig {
	return benchConfig{
		DatabaseURL: "postgres://localhost:5432/mydb",
		Port:        8080,
		Debug:       true,
		Features:    []string{"feature1", "feature2", "feature3"},
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
			"X-Request-ID":  "req-123-456",
		},
		Timeout: 30,
	}
}

func getBenchLargeStruct() benchLargeStruct {
	users := make([]benchUser, 10)
	for i := range users {
		users[i] = getBenchUser()
		users[i].ID = int64(i + 1)
	}

	configs := make([]benchConfig, 5)
	for i := range configs {
		configs[i] = getBenchConfig()
		configs[i].Port = 8080 + i
	}

	return benchLargeStruct{
		Users:   users,
		Configs: configs,
		Metadata: map[string]interface{}{
			"version": "1.0.0",
			"env":     "production",
			"region":  "us-east-1",
		},
		Tags: []string{"production", "api", "v2", "stable"},
		Settings: map[string]benchConfig{
			"primary":   getBenchConfig(),
			"secondary": getBenchConfig(),
		},
	}
}

// Marshal Benchmarks

func BenchmarkMarshal_SimpleStruct(b *testing.B) {
	user := getBenchUser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(user)
	}
}

func BenchmarkMarshal_ComplexStruct(b *testing.B) {
	config := getBenchConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(config)
	}
}

func BenchmarkMarshal_LargeStruct(b *testing.B) {
	large := getBenchLargeStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(large)
	}
}

// MarshalIndent Benchmarks

func BenchmarkMarshalIndent_SimpleStruct(b *testing.B) {
	user := getBenchUser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalIndent(user, "", "  ")
	}
}

func BenchmarkMarshalIndent_ComplexStruct(b *testing.B) {
	config := getBenchConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalIndent(config, "", "  ")
	}
}

func BenchmarkMarshalIndent_LargeStruct(b *testing.B) {
	large := getBenchLargeStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalIndent(large, "", "  ")
	}
}

// Unmarshal Benchmarks

func BenchmarkUnmarshal_SimpleStruct(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","created_at":"2024-01-15T10:30:00Z","active":true}`)
	var user benchUser
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Unmarshal(data, &user)
	}
}

func BenchmarkUnmarshal_ComplexStruct(b *testing.B) {
	data := []byte(`{"database_url":"postgres://localhost:5432/mydb","port":8080,"debug":true,"features":["feature1","feature2","feature3"],"headers":{"Authorization":"Bearer token123","Content-Type":"application/json","X-Request-ID":"req-123-456"},"timeout":30}`)
	var config benchConfig
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Unmarshal(data, &config)
	}
}

func BenchmarkUnmarshal_LargeStruct(b *testing.B) {
	large := getBenchLargeStruct()
	data, _ := Marshal(large)
	var result benchLargeStruct
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Unmarshal(data, &result)
	}
}

// Valid Benchmarks

func BenchmarkValid_ValidJSON(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","active":true}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Valid(data)
	}
}

func BenchmarkValid_InvalidJSON(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Valid(data)
	}
}

func BenchmarkValid_LargeJSON(b *testing.B) {
	large := getBenchLargeStruct()
	data, _ := Marshal(large)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Valid(data)
	}
}

// Round-trip Benchmarks

func BenchmarkRoundTrip_SimpleStruct(b *testing.B) {
	user := getBenchUser()
	var result benchUser
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := Marshal(user)
		_ = Unmarshal(data, &result)
	}
}

func BenchmarkRoundTrip_ComplexStruct(b *testing.B) {
	config := getBenchConfig()
	var result benchConfig
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := Marshal(config)
		_ = Unmarshal(data, &result)
	}
}

func BenchmarkRoundTrip_LargeStruct(b *testing.B) {
	large := getBenchLargeStruct()
	var result benchLargeStruct
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := Marshal(large)
		_ = Unmarshal(data, &result)
	}
}

// Comparison Benchmarks with standard library

func BenchmarkCompare_Marshal_Stdlib(b *testing.B) {
	user := getBenchUser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(user)
	}
}

func BenchmarkCompare_Marshal_Jcodec(b *testing.B) {
	user := getBenchUser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(user)
	}
}

func BenchmarkCompare_Unmarshal_Stdlib(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","created_at":"2024-01-15T10:30:00Z","active":true}`)
	var user benchUser
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(data, &user)
	}
}

func BenchmarkCompare_Unmarshal_Jcodec(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","created_at":"2024-01-15T10:30:00Z","active":true}`)
	var user benchUser
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Unmarshal(data, &user)
	}
}

func BenchmarkCompare_MarshalIndent_Stdlib(b *testing.B) {
	config := getBenchConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.MarshalIndent(config, "", "  ")
	}
}

func BenchmarkCompare_MarshalIndent_Jcodec(b *testing.B) {
	config := getBenchConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalIndent(config, "", "  ")
	}
}

// Parallel Benchmarks

func BenchmarkMarshal_Parallel(b *testing.B) {
	user := getBenchUser()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = Marshal(user)
		}
	})
}

func BenchmarkUnmarshal_Parallel(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","created_at":"2024-01-15T10:30:00Z","active":true}`)
	b.RunParallel(func(pb *testing.PB) {
		var user benchUser
		for pb.Next() {
			_ = Unmarshal(data, &user)
		}
	})
}

func BenchmarkValid_Parallel(b *testing.B) {
	data := []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","active":true}`)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Valid(data)
		}
	})
}
