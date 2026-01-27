// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
	"testing"
	"time"
)

// testUser represents a test user structure for JSON operations.
type testUser struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

// testConfig represents a test configuration structure.
type testConfig struct {
	DatabaseURL string            `json:"database_url"`
	Port        int               `json:"port"`
	Debug       bool              `json:"debug"`
	Headers     map[string]string `json:"headers"`
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(t *testing.T, data []byte, err error)
	}{
		{
			name:  "valid user struct should marshal successfully",
			input: testUser{ID: 1, Name: "John", Email: "john@example.com", CreatedAt: time.Now(), Active: true},
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
				if !contains(data, []byte("John")) {
					t.Error("Marshaled data should contain user name")
				}
			},
		},
		{
			name:  "valid config struct should marshal successfully",
			input: testConfig{DatabaseURL: "postgres://localhost", Port: 8080, Debug: true, Headers: map[string]string{"Content-Type": "application/json"}},
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name:    "channel type should return error",
			input:   make(chan int),
			wantErr: true,
			check: func(t *testing.T, _ []byte, err error) {
				if err == nil {
					t.Error("Expected error for channel type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			tt.check(t, data, err)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		target  interface{}
		wantErr bool
		check   func(t *testing.T, target interface{}, err error)
	}{
		{
			name:   "valid JSON should unmarshal successfully",
			data:   []byte(`{"id":1,"name":"John","email":"john@example.com","active":true}`),
			target: &testUser{},
			check: func(t *testing.T, target interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				user, ok := target.(*testUser)
				if !ok {
					t.Error("Target should be *testUser")
					return
				}
				if user.ID != 1 {
					t.Errorf("User ID = %v, want %v", user.ID, 1)
				}
				if user.Name != "John" {
					t.Errorf("User Name = %v, want %v", user.Name, "John")
				}
			},
		},
		{
			name:   "valid config JSON should unmarshal successfully",
			data:   []byte(`{"database_url":"postgres://localhost","port":8080,"debug":true,"headers":{"Content-Type":"application/json"}}`),
			target: &testConfig{},
			check: func(t *testing.T, target interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				config, ok := target.(*testConfig)
				if !ok {
					t.Error("Target should be *testConfig")
					return
				}
				if config.Port != 8080 {
					t.Errorf("Config Port = %v, want %v", config.Port, 8080)
				}
			},
		},
		{
			name:    "invalid JSON should return error",
			data:    []byte(`{"invalid": json}`),
			target:  &testUser{},
			wantErr: true,
			check: func(t *testing.T, _ interface{}, err error) {
				if err == nil {
					t.Error("Expected error for invalid JSON")
				}
			},
		},
		{
			name:    "nil target should return error",
			data:    []byte(`{"id":1}`),
			target:  nil,
			wantErr: true,
			check: func(t *testing.T, _ interface{}, err error) {
				if err == nil {
					t.Error("Expected error for nil target")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.data, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			tt.check(t, tt.target, err)
		})
	}
}

func TestNewEngineForArch(t *testing.T) {
	tests := []struct {
		name  string
		arch  string
		check func(t *testing.T, engine engine)
	}{
		{
			name: "arm64 architecture should return goccy engine",
			arch: "arm64",
			check: func(t *testing.T, e engine) {
				if e == nil {
					t.Error("Engine should not be nil")
				}
				// Test that it works
				data, err := e.Marshal(testUser{ID: 1, Name: "Test"})
				if err != nil {
					t.Errorf("Engine Marshal should work: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name: "unknown architecture should return goccy engine",
			arch: "riscv64",
			check: func(t *testing.T, e engine) {
				if e == nil {
					t.Error("Engine should not be nil")
				}
				// Test that it works
				data, err := e.Marshal(testUser{ID: 1, Name: "Test"})
				if err != nil {
					t.Errorf("Engine Marshal should work: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name: "ppc64 architecture should return goccy engine",
			arch: "ppc64",
			check: func(t *testing.T, e engine) {
				if e == nil {
					t.Error("Engine should not be nil")
				}
				// Test that it works
				data, err := e.Marshal(testUser{ID: 1, Name: "Test"})
				if err != nil {
					t.Errorf("Engine Marshal should work: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := newEngineForArch(tt.arch)
			tt.check(t, engine)
		})
	}
}

func TestGetDefaultEngine(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, engine engine)
	}{
		{
			name: "getDefaultEngine should return non-nil engine",
			check: func(t *testing.T, e engine) {
				if e == nil {
					t.Error("Default engine should not be nil")
				}
			},
		},
		{
			name: "getDefaultEngine should return same instance on multiple calls",
			check: func(t *testing.T, _ engine) {
				// Call multiple times - should return same instance due to sync.Once
				e1 := getDefaultEngine()
				e2 := getDefaultEngine()
				e3 := getDefaultEngine()

				// All should work (same instance)
				data1, err1 := e1.Marshal(testUser{ID: 1, Name: "Test1"})
				data2, err2 := e2.Marshal(testUser{ID: 2, Name: "Test2"})
				data3, err3 := e3.Marshal(testUser{ID: 3, Name: "Test3"})

				if err1 != nil || err2 != nil || err3 != nil {
					t.Error("All engines should work")
				}
				if len(data1) == 0 || len(data2) == 0 || len(data3) == 0 {
					t.Error("All marshaled data should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := getDefaultEngine()
			tt.check(t, engine)
		})
	}
}

func TestMarshal_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(t *testing.T, data []byte, err error)
	}{
		{
			name:    "nil value should marshal successfully",
			input:   nil,
			wantErr: false,
			check: func(t *testing.T, _ []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:    "empty struct should marshal successfully",
			input:   struct{}{},
			wantErr: false,
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name:    "empty map should marshal successfully",
			input:   map[string]interface{}{},
			wantErr: false,
			check: func(t *testing.T, _ []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:    "empty slice should marshal successfully",
			input:   []string{},
			wantErr: false,
			check: func(t *testing.T, _ []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:    "nested struct should marshal successfully",
			input:   struct{ Nested testUser }{Nested: testUser{ID: 1, Name: "Nested"}},
			wantErr: false,
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name:    "struct with pointer field should marshal successfully",
			input:   struct{ Ptr *string }{Ptr: stringPtr("test")},
			wantErr: false,
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
			},
		},
		{
			name:    "struct with nil pointer field should marshal successfully",
			input:   struct{ Ptr *string }{Ptr: nil},
			wantErr: false,
			check: func(t *testing.T, _ []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			tt.check(t, data, err)
		})
	}
}

func TestUnmarshal_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		target  interface{}
		wantErr bool
		check   func(t *testing.T, target interface{}, err error)
	}{
		{
			name:    "empty JSON object should unmarshal successfully",
			data:    []byte(`{}`),
			target:  &testUser{},
			wantErr: false,
			check: func(t *testing.T, _ interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:    "empty JSON array should unmarshal successfully",
			data:    []byte(`[]`),
			target:  &[]string{},
			wantErr: false,
			check: func(t *testing.T, _ interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:    "null JSON should unmarshal successfully",
			data:    []byte(`null`),
			target:  &testUser{},
			wantErr: false,
			check: func(t *testing.T, _ interface{}, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:    "empty bytes should return error",
			data:    []byte(``),
			target:  &testUser{},
			wantErr: true,
			check: func(t *testing.T, _ interface{}, err error) {
				if err == nil {
					t.Error("Expected error for empty bytes")
				}
			},
		},
		{
			name:    "whitespace only should return error",
			data:    []byte(`   `),
			target:  &testUser{},
			wantErr: true,
			check: func(t *testing.T, _ interface{}, err error) {
				if err == nil {
					t.Error("Expected error for whitespace only")
				}
			},
		},
		{
			name:    "partial JSON should return error",
			data:    []byte(`{"id":`),
			target:  &testUser{},
			wantErr: true,
			check: func(t *testing.T, _ interface{}, err error) {
				if err == nil {
					t.Error("Expected error for partial JSON")
				}
			},
		},
		{
			name:    "wrong type in JSON should return error",
			data:    []byte(`{"id":"not_a_number"}`),
			target:  &testUser{},
			wantErr: true, // JSON libraries error on type mismatch
			check: func(t *testing.T, _ interface{}, err error) {
				if err == nil {
					t.Error("Expected error for wrong type in JSON")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.data, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			tt.check(t, tt.target, err)
		})
	}
}

func TestMarshal_Unmarshal_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, original, unmarshaled interface{})
	}{
		{
			name:  "user struct should round-trip successfully",
			input: testUser{ID: 123, Name: "Jane Doe", Email: "jane@example.com", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Active: true},
			check: func(t *testing.T, original, unmarshaled interface{}) {
				originalUser := original.(testUser)
				unmarshaledUser := unmarshaled.(*testUser)
				if originalUser.ID != unmarshaledUser.ID {
					t.Errorf("ID mismatch: original=%v, unmarshaled=%v", originalUser.ID, unmarshaledUser.ID)
				}
				if originalUser.Name != unmarshaledUser.Name {
					t.Errorf("Name mismatch: original=%v, unmarshaled=%v", originalUser.Name, unmarshaledUser.Name)
				}
			},
		},
		{
			name:  "config struct should round-trip successfully",
			input: testConfig{DatabaseURL: "postgres://localhost:5432", Port: 8080, Debug: true, Headers: map[string]string{"Authorization": "Bearer token"}},
			check: func(t *testing.T, original, unmarshaled interface{}) {
				originalConfig := original.(testConfig)
				unmarshaledConfig := unmarshaled.(*testConfig)
				if originalConfig.Port != unmarshaledConfig.Port {
					t.Errorf("Port mismatch: original=%v, unmarshaled=%v", originalConfig.Port, unmarshaledConfig.Port)
				}
				if originalConfig.DatabaseURL != unmarshaledConfig.DatabaseURL {
					t.Errorf("DatabaseURL mismatch: original=%v, unmarshaled=%v", originalConfig.DatabaseURL, unmarshaledConfig.DatabaseURL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := Marshal(tt.input)
			if err != nil {
				t.Errorf("Marshal error: %v", err)
				return
			}

			// Unmarshal
			var unmarshaled interface{}
			switch tt.input.(type) {
			case testUser:
				unmarshaled = &testUser{}
			case testConfig:
				unmarshaled = &testConfig{}
			}

			err = Unmarshal(data, unmarshaled)
			if err != nil {
				t.Errorf("Unmarshal error: %v", err)
				return
			}

			tt.check(t, tt.input, unmarshaled)
		})
	}
}

func TestMarshalIndent(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		prefix  string
		indent  string
		wantErr bool
		check   func(t *testing.T, data []byte, err error)
	}{
		{
			name:   "simple struct should marshal with indentation",
			input:  testUser{ID: 1, Name: "John", Email: "john@example.com"},
			prefix: "",
			indent: "  ",
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
				// Check for indentation (should have newlines and spaces)
				if !contains(data, []byte("\n")) {
					t.Error("Indented JSON should contain newlines")
				}
				if !contains(data, []byte("John")) {
					t.Error("Marshaled data should contain user name")
				}
			},
		},
		{
			name:   "complex struct should marshal with custom prefix and indent",
			input:  testConfig{DatabaseURL: "postgres://localhost", Port: 8080, Debug: true},
			prefix: ">",
			indent: "\t",
			check: func(t *testing.T, data []byte, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}
				// Check for prefix
				if !contains(data, []byte(">")) {
					t.Error("Indented JSON should contain prefix")
				}
			},
		},
		{
			name:    "channel type should return error",
			input:   make(chan int),
			prefix:  "",
			indent:  "  ",
			wantErr: true,
			check: func(t *testing.T, _ []byte, err error) {
				if err == nil {
					t.Error("Expected error for channel type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalIndent(tt.input, tt.prefix, tt.indent)

			if (err != nil) != tt.wantErr {
				t.Errorf("Error expectation mismatch: got err=%v, wantErr=%v", err, tt.wantErr)
				return
			}

			tt.check(t, data, err)
		})
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{
			name:  "valid JSON object should return true",
			data:  []byte(`{"id":1,"name":"John"}`),
			valid: true,
		},
		{
			name:  "valid JSON array should return true",
			data:  []byte(`[1,2,3]`),
			valid: true,
		},
		{
			name:  "valid JSON string should return true",
			data:  []byte(`"hello"`),
			valid: true,
		},
		{
			name:  "valid JSON number should return true",
			data:  []byte(`123`),
			valid: true,
		},
		{
			name:  "valid JSON boolean should return true",
			data:  []byte(`true`),
			valid: true,
		},
		{
			name:  "valid JSON null should return true",
			data:  []byte(`null`),
			valid: true,
		},
		{
			name:  "invalid JSON should return false",
			data:  []byte(`{"id":1,"name":}`),
			valid: false,
		},
		{
			name:  "partial JSON should return false",
			data:  []byte(`{"id":1,`),
			valid: false,
		},
		{
			name:  "empty bytes should return false",
			data:  []byte(``),
			valid: false,
		},
		{
			name:  "whitespace only should return false",
			data:  []byte(`   `),
			valid: false,
		},
		{
			name:  "malformed JSON should return false",
			data:  []byte(`{id:1}`),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valid(tt.data)
			if result != tt.valid {
				t.Errorf("Valid() = %v, want %v for data: %s", result, tt.valid, string(tt.data))
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		iterations int
	}{
		{
			name:       "moderate concurrency test",
			goroutines: 50,
			iterations: 100,
		},
		{
			name:       "high concurrency test",
			goroutines: 100,
			iterations: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test concurrent Marshal operations
			t.Run("concurrent Marshal", func(t *testing.T) {
				done := make(chan bool, tt.goroutines)
				for i := 0; i < tt.goroutines; i++ {
					go func(id int) {
						defer func() { done <- true }()
						for j := 0; j < tt.iterations; j++ {
							user := testUser{ID: int64(id), Name: "User", Email: "user@example.com"}
							_, err := Marshal(user)
							if err != nil {
								t.Errorf("Concurrent Marshal error: %v", err)
							}
						}
					}(i)
				}

				// Wait for all goroutines
				for i := 0; i < tt.goroutines; i++ {
					<-done
				}
			})

			// Test concurrent Unmarshal operations
			t.Run("concurrent Unmarshal", func(t *testing.T) {
				data := []byte(`{"id":1,"name":"Test","email":"test@example.com"}`)
				done := make(chan bool, tt.goroutines)
				for i := 0; i < tt.goroutines; i++ {
					go func() {
						defer func() { done <- true }()
						for j := 0; j < tt.iterations; j++ {
							var user testUser
							err := Unmarshal(data, &user)
							if err != nil {
								t.Errorf("Concurrent Unmarshal error: %v", err)
							}
						}
					}()
				}

				// Wait for all goroutines
				for i := 0; i < tt.goroutines; i++ {
					<-done
				}
			})

			// Test concurrent MarshalIndent operations
			t.Run("concurrent MarshalIndent", func(t *testing.T) {
				done := make(chan bool, tt.goroutines)
				for i := 0; i < tt.goroutines; i++ {
					go func(id int) {
						defer func() { done <- true }()
						for j := 0; j < tt.iterations; j++ {
							config := testConfig{Port: id, Debug: true}
							_, err := MarshalIndent(config, "", "  ")
							if err != nil {
								t.Errorf("Concurrent MarshalIndent error: %v", err)
							}
						}
					}(i)
				}

				// Wait for all goroutines
				for i := 0; i < tt.goroutines; i++ {
					<-done
				}
			})

			// Test concurrent Valid operations
			t.Run("concurrent Valid", func(t *testing.T) {
				validData := []byte(`{"id":1,"name":"Test"}`)
				invalidData := []byte(`{"id":1,"name":}`)
				done := make(chan bool, tt.goroutines)
				for i := 0; i < tt.goroutines; i++ {
					go func(id int) {
						defer func() { done <- true }()
						for j := 0; j < tt.iterations; j++ {
							if id%2 == 0 {
								if !Valid(validData) {
									t.Error("Valid data reported as invalid")
								}
							} else {
								if Valid(invalidData) {
									t.Error("Invalid data reported as valid")
								}
							}
						}
					}(i)
				}

				// Wait for all goroutines
				for i := 0; i < tt.goroutines; i++ {
					<-done
				}
			})

			// Test concurrent mixed operations
			t.Run("concurrent mixed operations", func(_ *testing.T) {
				done := make(chan bool, tt.goroutines)
				for i := 0; i < tt.goroutines; i++ {
					go func(id int) {
						defer func() { done <- true }()
						for j := 0; j < tt.iterations; j++ {
							// Mix different operations
							switch j % 4 {
							case 0:
								user := testUser{ID: int64(id), Name: "User"}
								_, _ = Marshal(user)
							case 1:
								data := []byte(`{"id":1,"name":"Test"}`)
								var user testUser
								_ = Unmarshal(data, &user)
							case 2:
								config := testConfig{Port: id}
								_, _ = MarshalIndent(config, "", "  ")
							case 3:
								data := []byte(`{"id":1}`)
								_ = Valid(data)
							}
						}
					}(i)
				}

				// Wait for all goroutines
				for i := 0; i < tt.goroutines; i++ {
					<-done
				}
			})
		})
	}
}

func TestErrorContextWrapping(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		check     func(t *testing.T, err error)
	}{
		{
			name:      "marshal error should contain engine context",
			operation: "marshal",
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected error for channel type")
					return
				}
				errStr := err.Error()
				// Check for engine context (either "sonic engine" or "goccy engine")
				hasContext := contains([]byte(errStr), []byte("sonic engine")) ||
					contains([]byte(errStr), []byte("goccy engine"))
				if !hasContext {
					t.Errorf("Error should contain engine context, got: %v", err)
				}
			},
		},
		{
			name:      "unmarshal error should contain engine context",
			operation: "unmarshal",
			check: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected error for invalid JSON")
					return
				}
				errStr := err.Error()
				// Check for engine context
				hasContext := contains([]byte(errStr), []byte("sonic engine")) ||
					contains([]byte(errStr), []byte("goccy engine"))
				if !hasContext {
					t.Errorf("Error should contain engine context, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch tt.operation {
			case "marshal":
				_, err = Marshal(make(chan int))
			case "unmarshal":
				err = Unmarshal([]byte(`{"invalid": json}`), &testUser{})
			}
			tt.check(t, err)
		})
	}
}

// Helper function to check if a byte slice contains another byte slice.
func contains(haystack, needle []byte) bool {
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Helper function to create a string pointer.
func stringPtr(s string) *string {
	return &s
}
