// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package utils

import (
	"encoding/json"
)

// Unmarshal parses a JSON string into the provided struct type.
//
// Input:
//   - jsonStr: The JSON string to parse
//   - target: A struct instance to unmarshal into (used for type inference)
//
// Output:
//   - T: The parsed struct instance, or zero value if parsing failed
//   - error: Any error that occurred during parsing
//
// Example:
//
//	type User struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//
//	user, err := utils.Unmarshal(`{"name":"john","age":30}`, User{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(user.Name) // "john"
func Unmarshal[T any](jsonStr string, target T) (T, error) {
	var zero T
	if jsonStr == "" {
		return zero, nil
	}

	var result T
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return zero, err
	}

	return result, nil
}

// MarshalCompact marshals any Go value into compact single-line JSON.
// It can handle both JSON strings (by parsing first) or Go structs/values (by marshaling directly).
//
// Input:
//   - data: The value to marshal (can be a JSON string or any Go value)
//
// Output:
//   - string: The formatted compact JSON string
//   - error: Any error that occurred during marshaling
//
// Example:
//
//	// Marshal a struct
//	type User struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	user := User{Name: "john", Age: 30}
//	formatted, err := utils.MarshalCompact(user)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(formatted) // `{"age":30,"name":"john"}`
//
//	// Marshal a JSON string (parses and reformats)
//	formatted, err := utils.MarshalCompact(`{"name":"john","age":30}`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(formatted) // `{"age":30,"name":"john"}`
func MarshalCompact(data interface{}) (string, error) {
	// If data is a string, try to parse it as JSON first to validate and normalize
	if str, ok := data.(string); ok && str != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(str), &jsonData); err != nil {
			return "", err
		}
		formatted, err := json.Marshal(jsonData)
		if err != nil {
			return "", err
		}
		return string(formatted), nil
	}

	// For non-string values, marshal directly
	formatted, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}
