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

// processFieldValue inspects a value and, if it is a struct (or pointer to struct),
// processes its fields according to `log` struct tags.
// For non-struct values, it returns the value unchanged.
//
// Input:
//   - value: The field value to process
//   - maskKey: The AES encryption key for masking. If empty, masked fields show "***"
//
// Output:
//   - interface{}: The processed value (sanitized map for structs, original value otherwise)
func processFieldValue(value interface{}, maskKey string) interface{} {
	if value == nil {
		return nil
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

// processStruct iterates over struct fields and applies log tag rules.
func processStruct(v reflect.Value, maskKey string) map[string]interface{} {
	t := v.Type()
	result := make(map[string]interface{}, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get(LogTagName)

		switch tag {
		case LogTagOmit:
			continue
		case LogTagMask:
			name := resolveFieldName(field)
			result[name] = maskValue(fieldValue.Interface(), maskKey)
		default:
			name := resolveFieldName(field)
			// Recursively process nested structs
			fv := fieldValue
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					result[name] = nil
					continue
				}
				fv = fv.Elem()
			}
			if fv.Kind() == reflect.Struct {
				result[name] = processStruct(fv, maskKey)
			} else {
				result[name] = fieldValue.Interface()
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
func maskValue(value interface{}, maskKey string) string {
	if maskKey == "" {
		return defaultMaskPlaceholder
	}

	plaintext := fmt.Sprintf("%v", value)

	block, err := aes.NewCipher([]byte(maskKey))
	if err != nil {
		return defaultMaskPlaceholder
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return defaultMaskPlaceholder
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return defaultMaskPlaceholder
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}
