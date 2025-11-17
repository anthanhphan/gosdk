// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a Field with a string value.
func String(key string, val string) Field {
	return Field{Key: key, Value: val}
}

// Int creates a Field with an int value.
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 creates a Field with an int64 value.
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Float64 creates a Field with a float64 value.
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

// Bool creates a Field with a bool value.
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Any creates a Field with any value.
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}

// ErrorField creates a Field with an error value.
func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}
