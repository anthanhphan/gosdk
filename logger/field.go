// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"encoding/binary"
	"math"
)

// FieldType indicates how a Field's value is stored, allowing the encoder
// to avoid interface{} type assertions and boxing overhead.
type FieldType uint8

const (
	// FieldTypeAny is the default -- value stored in Iface.
	FieldTypeAny FieldType = iota
	// FieldTypeString -- value stored in Str.
	FieldTypeString
	// FieldTypeInt64 -- value stored in Integer.
	FieldTypeInt64
	// FieldTypeBool -- value stored in Integer (0=false, 1=true).
	FieldTypeBool
	// FieldTypeFloat64 -- value stored in Integer (math.Float64bits).
	FieldTypeFloat64
)

// Field represents a key-value pair for structured logging.
// Uses a typed union (like zap) to avoid interface{} boxing for common types.
type Field struct {
	Key     string
	Type    FieldType
	Integer int64  // int64, bool (as 0/1), float64 (as bits)
	Str     string // string value
	Iface   any    // fallback for complex types
}

// String creates a Field with a string value (zero-alloc).
func String(key string, val string) Field {
	return Field{Key: key, Type: FieldTypeString, Str: val}
}

// Int creates a Field with an int value (zero-alloc).
func Int(key string, val int) Field {
	return Field{Key: key, Type: FieldTypeInt64, Integer: int64(val)}
}

// Int64 creates a Field with an int64 value (zero-alloc).
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: FieldTypeInt64, Integer: val}
}

// float64ToInt64Bits stores a float64's IEEE-754 bits in an int64 field
// without direct uint64↔int64 casts (avoids gosec G115).
// Uses split-uint32 read: uint32→int64 is always safe.
func float64ToInt64Bits(val float64) int64 {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(val))
	lo := int64(binary.LittleEndian.Uint32(buf[:4]))
	hi := int64(binary.LittleEndian.Uint32(buf[4:]))
	return lo | (hi << 32)
}

// int64BitsToFloat64 reverses float64ToInt64Bits.
func int64BitsToFloat64(bits int64) float64 {
	var buf [8]byte
	binary.LittleEndian.PutUint32(buf[:4], uint32(bits&0xFFFFFFFF))
	binary.LittleEndian.PutUint32(buf[4:], uint32((bits>>32)&0xFFFFFFFF))
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:]))
}

// Float64 creates a Field with a float64 value (zero-alloc via math.Float64bits).
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: FieldTypeFloat64, Integer: float64ToInt64Bits(val)}
}

// Bool creates a Field with a bool value (zero-alloc).
func Bool(key string, val bool) Field {
	var i int64
	if val {
		i = 1
	}
	return Field{Key: key, Type: FieldTypeBool, Integer: i}
}

// Any creates a Field with any value type (may allocate for boxing).
func Any(key string, val any) Field {
	// Try to detect common types and use typed fields to avoid boxing
	switch v := val.(type) {
	case string:
		return String(key, v)
	case int:
		return Int(key, v)
	case int32:
		return Int64(key, int64(v))
	case int64:
		return Int64(key, v)
	case uint:
		if v <= math.MaxInt64 {
			return Int64(key, int64(v))
		}
		return Field{Key: key, Type: FieldTypeAny, Iface: val}
	case uint32:
		return Int64(key, int64(v))
	case uint64:
		if v <= math.MaxInt64 {
			return Int64(key, int64(v))
		}
		return Field{Key: key, Type: FieldTypeAny, Iface: val}
	case float32:
		return Float64(key, float64(v))
	case float64:
		return Float64(key, v)
	case bool:
		return Bool(key, v)
	default:
		return Field{Key: key, Type: FieldTypeAny, Iface: val}
	}
}

// ErrorField creates a Field with an error value.
// If the error is nil, the field value will be nil.
func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Type: FieldTypeAny, Iface: nil}
	}
	return String("error", err.Error())
}
