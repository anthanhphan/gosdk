// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"fmt"
	"strconv"

	"github.com/anthanhphan/gosdk/orianna/pkg/validator"
)

// Parameter Parsing Helpers

// GetParamInt gets a route parameter as an integer with error handling
//
// Example:
//
//	userID, err := orianna.GetParamInt(ctx, "id")
//	if err != nil {
//	    return ctx.BadRequest("Invalid user ID")
//	}
func GetParamInt(ctx Context, key string) (int, error) {
	value := ctx.Params(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s not found", key)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, WrapErrorf(err, "parameter %s is not a valid integer", key)
	}

	return intValue, nil
}

// GetParamInt64 gets a route parameter as an int64 with error handling
func GetParamInt64(ctx Context, key string) (int64, error) {
	value := ctx.Params(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s not found", key)
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, WrapErrorf(err, "parameter %s is not a valid integer", key)
	}

	return intValue, nil
}

// GetParamUUID gets a route parameter and validates it as a UUID
// Returns the UUID string if valid, error otherwise
//
// Example:
//
//	uuid, err := orianna.GetParamUUID(ctx, "id")
//	if err != nil {
//	    return ctx.BadRequest("Invalid UUID")
//	}
func GetParamUUID(ctx Context, key string) (string, error) {
	value := ctx.Params(key)
	if value == "" {
		return "", fmt.Errorf("parameter %s not found", key)
	}

	// Validate UUID format
	if !validator.UUIDRegex.MatchString(value) {
		return "", fmt.Errorf("parameter %s is not a valid UUID", key)
	}

	return value, nil
}
