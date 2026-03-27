// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "strconv"

// Query Parameter Helpers

// GetQueryInt gets a query parameter as an integer with default value
//
// Example:
//
//	page := orianna.GetQueryInt(ctx, "page", 1)
//	limit := orianna.GetQueryInt(ctx, "limit", 10)
func GetQueryInt(ctx Context, key string, defaultValue int) int {
	value := ctx.Query(key, "")
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetQueryInt64 gets a query parameter as an int64 with default value
func GetQueryInt64(ctx Context, key string, defaultValue int64) int64 {
	value := ctx.Query(key, "")
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetQueryBool gets a query parameter as a boolean with default value
// Accepts: "true", "1", "yes", "on" for true
// Accepts: "false", "0", "no", "off" for false
//
// Example:
//
//	includeDeleted := orianna.GetQueryBool(ctx, "include_deleted", false)
func GetQueryBool(ctx Context, key string, defaultValue bool) bool {
	value := ctx.Query(key, "")
	if value == "" {
		return defaultValue
	}

	// Check for common truthy values
	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetQueryString gets a query parameter as a string with default value
// This is just an alias for ctx.Query() for consistency
//
// Example:
//
//	sortBy := orianna.GetQueryString(ctx, "sort", "created_at")
func GetQueryString(ctx Context, key string, defaultValue string) string {
	return ctx.Query(key, defaultValue)
}
