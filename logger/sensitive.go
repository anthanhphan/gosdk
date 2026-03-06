// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

const (
	// LogTagName is the struct tag key used for log field processing.
	LogTagName = "log"

	// LogTagOmit indicates the field should be excluded from log output.
	LogTagOmit = "omit"

	// LogTagMask indicates the field value should be masked/encrypted in log output.
	LogTagMask = "mask"

	// defaultMaskPlaceholder is used when no MaskKey is configured.
	defaultMaskPlaceholder = "***"
)

// aeadCache caches cipher.AEAD instances per mask key to avoid
// re-creating AES cipher + GCM on every maskValue call.
var aeadCache sync.Map // map[string]cipher.AEAD

// getOrCreateAEAD returns a cached cipher.AEAD for the given key,
// creating and caching it on first use.
func getOrCreateAEAD(maskKey string) (cipher.AEAD, bool) {
	if v, ok := aeadCache.Load(maskKey); ok {
		return v.(cipher.AEAD), true
	}

	block, err := aes.NewCipher([]byte(maskKey))
	if err != nil {
		return nil, false
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, false
	}

	// Store and return (race-safe: duplicate creation is harmless)
	aeadCache.Store(maskKey, aesGCM)
	return aesGCM, true
}

// processField processes a typed Field for struct tag handling (mask/omit).
// For typed fields (string, int64, bool, float64), no processing is needed
// since they cannot be structs. Only FieldTypeAny is processed through
// processFieldValue for struct tag handling.
func processField(f Field, maskKey string) Field {
	if f.Type != FieldTypeAny {
		// Typed fields cannot be structs, no processing needed
		return f
	}
	// Process the interface value for struct tags
	processed := processFieldValue(f.Iface, maskKey)
	return Field{Key: f.Key, Type: FieldTypeAny, Iface: processed}
}

// processFieldValue inspects a value and, if it is a struct (or pointer to struct),
// processes its fields according to `log` struct tags.
// For non-struct values, it returns the value unchanged.
//
// Input:
//   - value: The field value to process
//   - maskKey: The AES encryption key for masking. If empty, masked fields show "***"
//
// Output:
//   - any: The processed value (sanitized map for structs, original value otherwise)
func processFieldValue(value any, maskKey string) any {
	if value == nil {
		return nil
	}

	// Fast path: skip reflect for common non-struct types
	switch value.(type) {
	case string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool, error:
		return value
	}

	v := reflect.ValueOf(value)

	// Dereference pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return value
	}

	return processStruct(v, maskKey)
}

// structFieldMeta caches per-field metadata to avoid re-parsing struct tags.
type structFieldMeta struct {
	name     string // resolved field name (from json tag or Go name)
	action   byte   // 0=include, 1=omit, 2=mask
	exported bool
}

// structMeta caches per-type metadata.
type structMeta struct {
	fields []structFieldMeta
}

// structMetaCache caches struct reflection metadata per type.
var structMetaCache sync.Map // map[reflect.Type]*structMeta

func getStructMeta(t reflect.Type) *structMeta {
	if cached, ok := structMetaCache.Load(t); ok {
		return cached.(*structMeta)
	}

	numField := t.NumField()
	meta := &structMeta{fields: make([]structFieldMeta, numField)}

	for i := 0; i < numField; i++ {
		field := t.Field(i)
		fm := &meta.fields[i]
		fm.exported = field.IsExported()
		if !fm.exported {
			continue
		}
		fm.name = resolveFieldName(field)
		tag := field.Tag.Get(LogTagName)
		switch tag {
		case LogTagOmit:
			fm.action = 1
		case LogTagMask:
			fm.action = 2
		default:
			fm.action = 0
		}
	}

	structMetaCache.Store(t, meta)
	return meta
}

// processStruct iterates over struct fields and applies log tag rules.
func processStruct(v reflect.Value, maskKey string) map[string]any {
	t := v.Type()
	meta := getStructMeta(t)
	result := make(map[string]any, t.NumField())

	for i, fm := range meta.fields {
		if !fm.exported {
			continue
		}

		switch fm.action {
		case 1: // omit
			continue
		case 2: // mask
			result[fm.name] = maskValue(v.Field(i).Interface(), maskKey)
		default: // include
			fv := v.Field(i)
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					result[fm.name] = nil
					continue
				}
				fv = fv.Elem()
			}
			if fv.Kind() == reflect.Struct {
				result[fm.name] = processStruct(fv, maskKey)
			} else {
				result[fm.name] = v.Field(i).Interface()
			}
		}
	}

	return result
}

// resolveFieldName returns the name to use for a struct field in log output.
// It checks the "json" tag first, then falls back to the field name.
func resolveFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		name, _, _ := strings.Cut(jsonTag, ",")
		if name != "" {
			return name
		}
	}
	return field.Name
}

// maskValue encrypts the string representation of a value using AES-GCM
// and returns a base64-encoded ciphertext. If maskKey is empty, returns "***".
//
// Input:
//   - value: The value to mask
//   - maskKey: The AES key (must be 16, 24, or 32 bytes for AES-128/192/256)
//
// Output:
//   - string: Base64-encoded encrypted value, or "***" if maskKey is empty or encryption fails
func maskValue(value any, maskKey string) string {
	if maskKey == "" {
		return defaultMaskPlaceholder
	}

	aesGCM, ok := getOrCreateAEAD(maskKey)
	if !ok {
		return defaultMaskPlaceholder
	}

	plaintext := fmt.Sprintf("%v", value)

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return defaultMaskPlaceholder
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}
