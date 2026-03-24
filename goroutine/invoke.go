// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"fmt"
	"reflect"
	"sync"
)

// ---------------------------------------------------------------------------
// Reflect-based invocation (generic path for Run)
// ---------------------------------------------------------------------------

// reflectValuePool reuses []reflect.Value slices for invoke to avoid allocation per call.
var reflectValuePool = sync.Pool{
	New: func() any {
		s := make([]reflect.Value, 0, 8)
		return &s
	},
}

// invoke validates and converts arguments, then invokes the function.
func invoke(fn any, args []any) {
	funcValue := reflect.ValueOf(fn)
	if funcValue.Kind() != reflect.Func {
		getRecoverLogger().Error("provided value is not a function")
		return
	}

	funcType := funcValue.Type()
	numIn := funcType.NumIn()

	if err := validateArguments(args, numIn); err != nil {
		getRecoverLogger().Errorw("argument validation failed",
			"error", err.Error(),
		)
		return
	}

	funcArgs, err := convertArguments(args, funcType, numIn)
	if err != nil {
		getRecoverLogger().Errorw("argument conversion failed",
			"error", err.Error(),
		)
		return
	}

	funcValue.Call(funcArgs)
}

// validateArguments checks if the argument count is valid.
func validateArguments(args []any, expected int) error {
	if len(args) < expected {
		return fmt.Errorf("insufficient arguments: expected %d, got %d", expected, len(args))
	}
	if len(args) > expected {
		getRecoverLogger().Warnw("excess arguments provided",
			"expected", expected,
			"provided", len(args),
		)
	}
	return nil
}

// convertArguments converts arguments to reflect.Value with type checking and conversion.
// Uses pooled []reflect.Value slices to reduce allocations.
func convertArguments(args []any, funcType reflect.Type, numIn int) ([]reflect.Value, error) {
	rvPtr := reflectValuePool.Get().(*[]reflect.Value)
	rv := (*rvPtr)[:0]

	for i := 0; i < numIn; i++ {
		paramType := funcType.In(i)
		argValue := reflect.ValueOf(args[i])

		// Handle nil values
		if !argValue.IsValid() {
			if paramType.Kind() == reflect.Ptr || paramType.Kind() == reflect.Interface {
				rv = append(rv, reflect.New(paramType).Elem())
				continue
			}
			*rvPtr = rv
			reflectValuePool.Put(rvPtr)
			return nil, fmt.Errorf("invalid argument at index %d", i)
		}

		switch {
		case argValue.Type().AssignableTo(paramType):
			rv = append(rv, argValue)
		case argValue.Type().ConvertibleTo(paramType):
			rv = append(rv, argValue.Convert(paramType))
		default:
			*rvPtr = rv
			reflectValuePool.Put(rvPtr)
			return nil, fmt.Errorf("type mismatch at index %d: expected %s, got %s",
				i, paramType.String(), argValue.Type().String())
		}
	}

	// Copy out of pool before returning pool -- caller may hold reference
	result := make([]reflect.Value, len(rv))
	copy(result, rv)
	*rvPtr = rv
	reflectValuePool.Put(rvPtr)
	return result, nil
}
