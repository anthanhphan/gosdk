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
	Encode(v any) error
	SetIndent(prefix, indent string)
	SetEscapeHTML(on bool)
}

// Decoder reads JSON values from an input stream.
type Decoder interface {
	Decode(v any) error
	Buffered() io.Reader
	DisallowUnknownFields()
	UseNumber()
}

// engine represents a JSON marshaling engine implementation.
type engine interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	MarshalIndent(v any, prefix, indent string) ([]byte, error)
	Valid(data []byte) bool
	NewEncoder(w io.Writer) Encoder
	NewDecoder(r io.Reader) Decoder
}

// ============================================================================
// Direct function pointers — eliminates interface dispatch + sync.Once per call
// ============================================================================

var (
	marshalFn       func(any) ([]byte, error)
	unmarshalFn     func([]byte, any) error
	marshalIndentFn func(any, string, string) ([]byte, error)
	validFn         func([]byte) bool
	newEncoderFn    func(io.Writer) Encoder
	newDecoderFn    func(io.Reader) Decoder
)

func init() {
	e := newEngineForArch(runtime.GOARCH)
	marshalFn = e.Marshal
	unmarshalFn = e.Unmarshal
	marshalIndentFn = e.MarshalIndent
	validFn = e.Valid
	newEncoderFn = e.NewEncoder
	newDecoderFn = e.NewDecoder
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

// ============================================================================
// Buffer Pool — reuse buffers for Compact/Indent/HTMLEscape
// ============================================================================

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func getBuf() *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func putBuf(buf *bytes.Buffer) {
	if buf.Cap() > 64*1024 { // don't pool buffers > 64KB
		return
	}
	bufPool.Put(buf)
}

// ============================================================================
// Public API — direct function calls, zero interface dispatch
// ============================================================================

// Marshal converts a Go value to JSON bytes using the optimal engine for the current architecture.
//
// Example:
//
//	data, err := jcodec.Marshal(user)
func Marshal(v any) ([]byte, error) {
	return marshalFn(v)
}

// Unmarshal converts JSON bytes to a Go value using the optimal engine for the current architecture.
//
// Example:
//
//	err := jcodec.Unmarshal(data, &user)
func Unmarshal(data []byte, v any) error {
	return unmarshalFn(data, v)
}

// MarshalIndent converts a Go value to pretty-printed JSON bytes using the optimal engine.
//
// Example:
//
//	data, err := jcodec.MarshalIndent(user, "", "  ")
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return marshalIndentFn(v, prefix, indent)
}

// CompactString converts a value to a compact JSON string.
// For string inputs, it compacts the JSON directly without re-parsing.
// For other types, it marshals to JSON.
//
// Example:
//
//	compact, err := jcodec.CompactString(`{ "id": 1,  "name": "John" }`)
//	// compact is `{"id":1,"name":"John"}`
func CompactString(data any) (string, error) {
	if str, ok := data.(string); ok && str != "" {
		buf := getBuf()
		defer putBuf(buf)
		if err := json.Compact(buf, []byte(str)); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	b, err := marshalFn(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Valid reports whether data is a valid JSON encoding.
//
// Example:
//
//	if !jcodec.Valid(data) {
//	    return errors.New("invalid JSON")
//	}
func Valid(data []byte) bool {
	return validFn(data)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) Encoder {
	return newEncoderFn(w)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) Decoder {
	return newDecoderFn(r)
}

// Compact appends to dst the JSON-encoded src with insignificant space characters elided.
func Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

// HTMLEscape appends to dst the JSON-encoded src with <, >, &, U+2028 and U+2029
// characters inside string literals changed to \u003c, \u003e, \u0026, \u2028, \u2029.
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

// Indent appends to dst an indented form of the JSON-encoded src.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}
