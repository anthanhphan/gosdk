// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a Field with a string value.
//
// Input:
//   - key: Field key name
//   - val: String value for the field
//
// Output:
//   - Field: A Field instance with the string value
//
// Example:
//
//	field := String("service", "user-service")
//	logger.Infow("Request processed", field)
func String(key string, val string) Field {
	return Field{Key: key, Value: val}
}

// Int creates a Field with an int value.
//
// Input:
//   - key: Field key name
//   - val: Integer value for the field
//
// Output:
//   - Field: A Field instance with the int value
//
// Example:
//
//	field := Int("user_id", 12345)
//	logger.Infow("User created", field)
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 creates a Field with an int64 value.
//
// Input:
//   - key: Field key name
//   - val: Int64 value for the field
//
// Output:
//   - Field: A Field instance with the int64 value
//
// Example:
//
//	field := Int64("timestamp", time.Now().Unix())
//	logger.Infow("Event occurred", field)
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Float64 creates a Field with a float64 value.
//
// Input:
//   - key: Field key name
//   - val: Float64 value for the field
//
// Output:
//   - Field: A Field instance with the float64 value
//
// Example:
//
//	field := Float64("duration_ms", 123.45)
//	logger.Infow("Request completed", field)
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

// Bool creates a Field with a bool value.
//
// Input:
//   - key: Field key name
//   - val: Boolean value for the field
//
// Output:
//   - Field: A Field instance with the bool value
//
// Example:
//
//	field := Bool("enabled", true)
//	logger.Infow("Feature status", field)
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Any creates a Field with any value type.
//
// Input:
//   - key: Field key name
//   - val: Any value type for the field
//
// Output:
//   - Field: A Field instance with the value
//
// Example:
//
//	field := Any("metadata", map[string]interface{}{"key": "value"})
//	logger.Infow("Event logged", field)
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}

// ErrorField creates a Field with an error value.
// If the error is nil, the field value will be nil.
//
// Input:
//   - err: Error value to convert to a field
//
// Output:
//   - Field: A Field instance with key "error" and the error message as value
//
// Example:
//
//	if err != nil {
//	    logger.Errorw("Operation failed", ErrorField(err))
//	}
func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}
