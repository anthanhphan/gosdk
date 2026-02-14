// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"io"
	"reflect"
	"strings"
	"testing"
)

type testUser struct {
	Name     string `json:"name"`
	Password string `json:"password" log:"omit"`
	Token    string `json:"token" log:"mask"`
	Age      int    `json:"age"`
}

type testNested struct {
	User    testUser `json:"user"`
	Session string   `json:"session" log:"mask"`
	Debug   string   `json:"debug" log:"omit"`
}

type testNoTags struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type testUnexported struct {
	Name     string `json:"name"`
	password string //nolint:staticcheck
}

// newTestUnexported creates a testUnexported with unexported field set for testing.
func newTestUnexported(name, pass string) testUnexported {
	return testUnexported{Name: name, password: pass}
}

func TestProcessFieldValue_NonStruct(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{"nil value should return nil", nil, nil},
		{"string value should pass through", "hello", "hello"},
		{"int value should pass through", 42, 42},
		{"float value should pass through", 3.14, 3.14},
		{"bool value should pass through", true, true},
		{"slice value should pass through", []string{"a", "b"}, nil}, // just check non-nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processFieldValue(tt.value, "")
			if tt.want != nil {
				if result != tt.want {
					t.Errorf("processFieldValue() = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestProcessFieldValue_OmitTag(t *testing.T) {
	user := testUser{
		Name:     "John",
		Password: "secret123",
		Token:    "abc-token",
		Age:      30,
	}

	result := processFieldValue(user, "")
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map, got %T", result)
	}

	if _, exists := m["password"]; exists {
		t.Error("processFieldValue() should omit field with log:\"omit\" tag")
	}
	if m["name"] != "John" {
		t.Errorf("processFieldValue() name = %v, want John", m["name"])
	}
	if m["age"] != 30 {
		t.Errorf("processFieldValue() age = %v, want 30", m["age"])
	}
}

func TestProcessFieldValue_MaskWithoutKey(t *testing.T) {
	user := testUser{
		Name:     "John",
		Password: "secret123",
		Token:    "abc-token",
		Age:      30,
	}

	result := processFieldValue(user, "")
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map, got %T", result)
	}

	if m["token"] != "***" {
		t.Errorf("processFieldValue() token = %v, want '***' when no mask key", m["token"])
	}
}

func TestProcessFieldValue_MaskWithKey(t *testing.T) {
	maskKey := "0123456789abcdef" // 16 bytes for AES-128

	user := testUser{
		Name:     "John",
		Password: "secret123",
		Token:    "abc-token",
		Age:      30,
	}

	result := processFieldValue(user, maskKey)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map, got %T", result)
	}

	maskedToken, ok := m["token"].(string)
	if !ok {
		t.Fatalf("processFieldValue() token should be string, got %T", m["token"])
	}

	if maskedToken == "abc-token" {
		t.Error("processFieldValue() token should be encrypted, not plaintext")
	}
	if maskedToken == "***" {
		t.Error("processFieldValue() token should be encrypted, not '***' when key is set")
	}

	// Verify we can decrypt it
	decrypted, err := decryptMaskedValue(maskedToken, maskKey)
	if err != nil {
		t.Fatalf("failed to decrypt masked value: %v", err)
	}
	if decrypted != "abc-token" {
		t.Errorf("decrypted value = %v, want 'abc-token'", decrypted)
	}
}

func TestProcessFieldValue_PointerToStruct(t *testing.T) {
	user := &testUser{
		Name:     "Jane",
		Password: "pass",
		Token:    "tok",
		Age:      25,
	}

	result := processFieldValue(user, "")
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map for pointer to struct, got %T", result)
	}

	if _, exists := m["password"]; exists {
		t.Error("processFieldValue() should omit password for pointer to struct")
	}
	if m["name"] != "Jane" {
		t.Errorf("processFieldValue() name = %v, want Jane", m["name"])
	}
}

func TestProcessFieldValue_NilPointer(t *testing.T) {
	var user *testUser
	result := processFieldValue(user, "")
	if result != nil {
		t.Errorf("processFieldValue() = %v, want nil for nil pointer", result)
	}
}

func TestProcessFieldValue_NestedStruct(t *testing.T) {
	nested := testNested{
		User: testUser{
			Name:     "John",
			Password: "secret",
			Token:    "tok",
			Age:      30,
		},
		Session: "sess-123",
		Debug:   "debug-info",
	}

	result := processFieldValue(nested, "")
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map, got %T", result)
	}

	// Debug should be omitted
	if _, exists := m["debug"]; exists {
		t.Error("processFieldValue() should omit debug field")
	}

	// Session should be masked
	if m["session"] != "***" {
		t.Errorf("processFieldValue() session = %v, want '***'", m["session"])
	}

	// User should be a nested map
	userMap, ok := m["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() user should be map, got %T", m["user"])
	}

	if _, exists := userMap["password"]; exists {
		t.Error("processFieldValue() should omit nested password field")
	}
	if userMap["name"] != "John" {
		t.Errorf("processFieldValue() nested name = %v, want John", userMap["name"])
	}
}

func TestProcessFieldValue_NoTags(t *testing.T) {
	data := testNoTags{Name: "test", Age: 10}
	result := processFieldValue(data, "")
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map, got %T", result)
	}

	if m["name"] != "test" {
		t.Errorf("processFieldValue() name = %v, want test", m["name"])
	}
	if m["age"] != 10 {
		t.Errorf("processFieldValue() age = %v, want 10", m["age"])
	}
}

func TestProcessFieldValue_UnexportedFieldsIgnored(t *testing.T) {
	data := newTestUnexported("visible", "hidden")
	result := processFieldValue(data, "")
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("processFieldValue() should return map, got %T", result)
	}

	if m["name"] != "visible" {
		t.Errorf("processFieldValue() name = %v, want visible", m["name"])
	}
	if _, exists := m["password"]; exists {
		t.Error("processFieldValue() should not include unexported fields")
	}
}

func TestResolveFieldName(t *testing.T) {
	type sample struct {
		WithJSON    string `json:"json_name"`
		WithComma   string `json:"comma_name,omitempty"`
		WithDash    string `json:"-"`
		NoTag       string
		EmptyJSON   string `json:""`
		OnlyOptions string `json:",omitempty"`
	}

	st := reflect.TypeOf(sample{})
	tests := []struct {
		fieldName string
		want      string
	}{
		{"WithJSON", "json_name"},
		{"WithComma", "comma_name"},
		{"WithDash", "WithDash"},
		{"NoTag", "NoTag"},
		{"EmptyJSON", "EmptyJSON"},
		{"OnlyOptions", "OnlyOptions"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, _ := st.FieldByName(tt.fieldName)
			got := resolveFieldName(field)
			if got != tt.want {
				t.Errorf("resolveFieldName(%s) = %v, want %v", tt.fieldName, got, tt.want)
			}
		})
	}
}

func TestMaskValue_InvalidKeyLength(t *testing.T) {
	result := maskValue("secret", "short")
	if result != "***" {
		t.Errorf("maskValue() with invalid key = %v, want '***'", result)
	}
}

func TestMaskValue_ValidKey(t *testing.T) {
	key := "0123456789abcdef" // 16 bytes
	result := maskValue("hello", key)

	if result == "hello" {
		t.Error("maskValue() should not return plaintext")
	}
	if result == "***" {
		t.Error("maskValue() should not return '***' with valid key")
	}

	// Verify base64 encoding
	_, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		t.Errorf("maskValue() should return valid base64, got error: %v", err)
	}
}

func TestIntegration_LogStructWithTags(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:      LevelDebug,
		LogEncoding:   EncodingJSON,
		DisableCaller: true,
		MaskKey:       "0123456789abcdef",
	}, []io.Writer{&buf})

	user := testUser{
		Name:     "John",
		Password: "super-secret",
		Token:    "bearer-token-123",
		Age:      30,
	}

	logger.Infow("user login", "user", user)
	output := buf.String()

	if !strings.Contains(output, "John") {
		t.Error("output should contain name")
	}
	if strings.Contains(output, "super-secret") {
		t.Error("output should NOT contain password (omitted)")
	}
	if strings.Contains(output, "bearer-token-123") {
		t.Error("output should NOT contain token in plaintext (masked)")
	}
	if !strings.Contains(output, "user login") {
		t.Error("output should contain the log message")
	}
}

func TestIntegration_LogStructWithTagsNoMaskKey(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:      LevelDebug,
		LogEncoding:   EncodingJSON,
		DisableCaller: true,
	}, []io.Writer{&buf})

	user := testUser{
		Name:     "Jane",
		Password: "pass",
		Token:    "tok-456",
		Age:      25,
	}

	logger.Infow("user action", "user", user)
	output := buf.String()

	if strings.Contains(output, "pass") {
		t.Error("output should NOT contain password (omitted)")
	}
	if !strings.Contains(output, "***") {
		t.Error("output should contain '***' for masked token when no key")
	}
}

// decryptMaskedValue is a test helper to verify AES-GCM encrypted values.
func decryptMaskedValue(encrypted, key string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
