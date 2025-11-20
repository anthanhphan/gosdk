// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

// Package jcodec provides an intelligent, architecture-aware JSON marshaling solution
// that automatically selects the optimal JSON library based on the target CPU architecture.
// It seamlessly switches between Sonic (for AMD64/x86_64) and goccy/go-json (for ARM64/others)
// to deliver peak performance on any platform.
package jcodec

import (
	"bytes"
	"encoding/json"
	"io"
	"runtime"
	"sync"
)

const (
	// Architecture constants for engine selection.
	archAMD64 = "amd64"
	arch386   = "386"
)

// Encoder writes JSON values to an output stream.
type Encoder interface {
	Encode(v interface{}) error
	SetIndent(prefix, indent string)
	SetEscapeHTML(on bool)
}

// Decoder reads JSON values from an input stream.
type Decoder interface {
	Decode(v interface{}) error
	Buffered() io.Reader
	DisallowUnknownFields()
	UseNumber()
}

// engine represents a JSON marshaling engine implementation.
type engine interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
	Valid(data []byte) bool
	NewEncoder(w io.Writer) Encoder
	NewDecoder(r io.Reader) Decoder
}

var (
	// defaultEngine is the global default engine instance, lazily initialized.
	defaultEngine engine
	once          sync.Once
)

// getDefaultEngine returns the global default engine instance.
// It is initialized on first use using auto-selection strategy.
func getDefaultEngine() engine {
	once.Do(func() {
		defaultEngine = newEngineForArch(runtime.GOARCH)
	})
	return defaultEngine
}

// newEngineForArch creates an engine optimized for a specific architecture.
func newEngineForArch(arch string) engine {
	switch arch {
	case archAMD64, arch386:
		return newSonicEngine()
	default:
		return newGoccyEngine()
	}
}

// Marshal converts a Go value to JSON bytes using the optimal engine for the current architecture.
// It automatically selects Sonic on AMD64/x86_64 and goccy/go-json on other architectures.
//
// Input:
//   - v: The value to marshal
//
// Output:
//   - []byte: The marshaled JSON bytes
//   - error: Any error that occurred during marshaling
//
// Example:
//
//	user := User{ID: 1, Name: "John"}
//	data, err := jcodec.Marshal(user)
//	if err != nil {
//	    return fmt.Errorf("marshal failed: %w", err)
//	}
func Marshal(v interface{}) ([]byte, error) {
	return getDefaultEngine().Marshal(v)
}

// Unmarshal converts JSON bytes to a Go value using the optimal engine for the current architecture.
// It automatically selects Sonic on AMD64/x86_64 and goccy/go-json on other architectures.
//
// Input:
//   - data: The JSON bytes to unmarshal
//   - v: A pointer to the value to unmarshal into
//
// Output:
//   - error: Any error that occurred during unmarshaling
//
// Example:
//
//	var user User
//	err := jcodec.Unmarshal(data, &user)
//	if err != nil {
//	    return fmt.Errorf("unmarshal failed: %w", err)
//	}
func Unmarshal(data []byte, v interface{}) error {
	return getDefaultEngine().Unmarshal(data, v)
}

// MarshalIndent converts a Go value to pretty-printed JSON bytes using the optimal engine.
// It works like Marshal but applies indentation to format the output for human readability.
// Each JSON element will begin on a new line beginning with prefix followed by one or more
// copies of indent according to the indentation nesting depth.
//
// Input:
//   - v: The value to marshal
//   - prefix: String to prefix each line with
//   - indent: String to use for each indentation level
//
// Output:
//   - []byte: The pretty-printed JSON bytes
//   - error: Any error that occurred during marshaling
//
// Example:
//
//	user := User{ID: 1, Name: "John"}
//	data, err := jcodec.MarshalIndent(user, "", "  ")
//	if err != nil {
//	    return fmt.Errorf("marshal failed: %w", err)
//	}
//	fmt.Println(string(data))
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return getDefaultEngine().MarshalIndent(v, prefix, indent)
}

// CompactString converts a value to a compact JSON string.
// It handles two cases:
// 1. If data is a string, it treats it as a JSON string, validates it, and compacts it.
// 2. For any other type, it marshals the value to JSON.
//
// Input:
//   - data: The value to convert (string or any other type)
//
// Output:
//   - string: The compact JSON string
//   - error: Any error that occurred during validation or marshaling
//
// Example:
//
//	// Case 1: JSON string
//	jsonStr := `{ "id": 1, "name": "John" }`
//	compact, err := jcodec.CompactString(jsonStr)
//	// compact is `{"id":1,"name":"John"}`
//
//	// Case 2: Go struct
//	user := User{ID: 1, Name: "John"}
//	compact, err := jcodec.CompactString(user)
//	// compact is `{"id":1,"name":"John"}`
func CompactString(data interface{}) (string, error) {
	// If data is a string, try to parse it as JSON first to validate and normalize
	if str, ok := data.(string); ok && str != "" {
		var jsonData interface{}
		if err := Unmarshal([]byte(str), &jsonData); err != nil {
			return "", err
		}
		formatted, err := Marshal(jsonData)
		if err != nil {
			return "", err
		}
		return string(formatted), nil
	}

	// For non-string values, marshal directly
	formatted, err := Marshal(data)
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

// Valid reports whether data is a valid JSON encoding.
// This function validates JSON syntax without unmarshaling into a Go value,
// making it efficient for validation-only use cases.
//
// Input:
//   - data: The JSON bytes to validate
//
// Output:
//   - bool: true if data is valid JSON, false otherwise
//
// Example:
//
//	data := []byte(`{"id":1,"name":"John"}`)
//	if !jcodec.Valid(data) {
//	    return errors.New("invalid JSON")
//	}
func Valid(data []byte) bool {
	return getDefaultEngine().Valid(data)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) Encoder {
	return getDefaultEngine().NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) Decoder {
	return getDefaultEngine().NewDecoder(r)
}

// Compact appends to dst the JSON-encoded src with insignificant space characters elided.
func Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

// HTMLEscape appends to dst the JSON-encoded src with <, >, &, U+2028 and U+2029
// characters inside string literals changed to \u003c, \u003e, \u0026, \u2028, \u2029
// so that the JSON can be safely embedded inside HTML <script> tags.
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

// Indent appends to dst an indented form of the JSON-encoded src.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}
