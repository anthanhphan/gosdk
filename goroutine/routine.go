// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/anthanhphan/gosdk/logger"
	"github.com/anthanhphan/gosdk/utils"
)

const (
	stackTraceBufferSize = 4096
	callerDepth          = 8
)

// frameFilter defines criteria for filtering stack frames.
type frameFilter struct {
	wrapperPatterns []string
	packagePatterns []string
	prefixPatterns  []string
}

// shouldSkip determines if a frame should be skipped based on filter criteria.
func (f *frameFilter) shouldSkip(funcName, filePath string) bool {
	return containsAny(funcName, f.wrapperPatterns) ||
		containsAny(filePath, f.packagePatterns) ||
		hasPrefixAny(funcName, f.prefixPatterns)
}

var (
	defaultFilter = frameFilter{
		wrapperPatterns: []string{
			".invoke.func",
			".Run.func",
			".recoverPanic",
			".capturePanicLocation",
			".callWithPanicRecovery",
		},
		packagePatterns: []string{
			"runtime/",
			"reflect/",
			"/utils/",
		},
		prefixPatterns: []string{
			"runtime.",
			"reflect.",
		},
	}

	// stackBufferPool reuses buffers for stack trace operations to reduce allocations.
	stackBufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, stackTraceBufferSize)
			return &buf
		},
	}
)

// stackTraceParser extracts panic location from stack traces.
type stackTraceParser struct {
	filter *frameFilter
}

// newStackTraceParser creates a new stack trace parser with the default filter.
func newStackTraceParser() *stackTraceParser {
	return &stackTraceParser{filter: &defaultFilter}
}

// extractLocation extracts the panic location from a stack trace string.
// It finds the first user code frame by skipping wrapper functions, system packages, and utility packages.
//
// Input:
//   - stackTrace: The full stack trace string from runtime.Stack()
//
// Output:
//   - string: The panic location in format "file:line" (relative to module root), empty if not found
func (p *stackTraceParser) extractLocation(stackTrace string) string {
	lines := strings.Split(stackTrace, "\n")
	frames := parseStackFrames(lines)

	for _, frame := range frames {
		if p.filter.shouldSkip(frame.funcName, frame.filePath) {
			continue
		}

		if location := frame.location(); location != "" {
			return location
		}
	}

	return ""
}

// stackFrame represents a single frame in the stack trace.
type stackFrame struct {
	funcName string
	filePath string
}

// location extracts and formats the location string for this frame.
// Returns empty string if the frame path is invalid or cannot be parsed.
func (f stackFrame) location() string {
	if !isValidFilePath(f.filePath) {
		return ""
	}

	file, line, ok := parseFilePath(f.filePath)
	if !ok {
		return ""
	}

	return fmt.Sprintf("%s:%s", utils.GetShortPath(file), line)
}

// isValidFilePath checks if a file path string is valid for parsing.
func isValidFilePath(path string) bool {
	return strings.Contains(path, ":") && !strings.HasSuffix(path, ")")
}

// parseFilePath extracts file and line number from a file path string.
// Returns (file, line, true) if successful, otherwise ("", "", false).
func parseFilePath(filePath string) (file, line string, ok bool) {
	filePath = strings.TrimSpace(filePath)
	parts := strings.Split(filePath, ":")
	if len(parts) < 2 {
		return "", "", false
	}

	file = parts[0]
	fields := strings.Fields(parts[1])
	if len(fields) == 0 {
		return "", "", false
	}

	return file, fields[0], true
}

// parseStackFrames parses stack trace lines into frame structures.
// Stack trace format: each frame has 2 lines (function name and file:line)
func parseStackFrames(lines []string) []stackFrame {
	const (
		goroutineHeaderOffset = 1
		framesPerEntry        = 2
	)

	frames := make([]stackFrame, 0, len(lines)/framesPerEntry)

	for i := goroutineHeaderOffset; i < len(lines)-1; i += framesPerEntry {
		if i+1 >= len(lines) {
			break
		}

		frames = append(frames, stackFrame{
			funcName: strings.TrimSpace(lines[i]),
			filePath: strings.TrimSpace(lines[i+1]),
		})
	}

	return frames
}

// containsAny checks if the string contains any of the given patterns.
func containsAny(str string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(str, pattern) {
			return true
		}
	}
	return false
}

// hasPrefixAny checks if the string has any of the given prefixes or contains them.
// Note: strings.Contains already includes strings.HasPrefix, so we only need Contains.
func hasPrefixAny(str string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.Contains(str, prefix) {
			return true
		}
	}
	return false
}

// Run starts a new goroutine and invokes the provided function with the given arguments.
// Any panic that occurs within the goroutine is recovered and logged with panic location.
//
// Input:
//   - fn: The function to be invoked in the new goroutine. It must be a valid function type.
//   - args: A variadic list of arguments to pass to the invoked function.
//
// Output:
//   - None
//
// Example:
//
//	routine.Run(func(msg string) {
//	    logger.Info("Message:", msg)
//	}, "Hello, world!")
//
//	routine.Run(func(a, b int) {
//	    fmt.Printf("Sum: %d\n", a+b)
//	}, 10, 20)
func Run(fn any, args ...any) {
	callerLocation := getCallerLocation()

	go func() {
		defer recoverPanic(callerLocation)
		invoke(fn, args)
	}()
}

// getCallerLocation captures the caller's location (where Run() was called from).
func getCallerLocation() string {
	pcs := make([]uintptr, callerDepth)
	n := runtime.Callers(2, pcs) // Skip getCallerLocation and Run
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	if frame, _ := frames.Next(); frame.File != "" {
		// Skip if it's still in goroutine package (shouldn't happen)
		if !strings.Contains(frame.File, "/goroutine/") {
			return fmt.Sprintf("%s:%d", utils.GetShortPath(frame.File), frame.Line)
		}
	}

	return ""
}

// functionInvoker handles function invocation with argument validation and type conversion.
type functionInvoker struct {
	logger *logger.Logger
}

// newFunctionInvoker creates a new function invoker.
func newFunctionInvoker() *functionInvoker {
	return &functionInvoker{
		logger: logger.NewLoggerWithFields(
			logger.String("prefix", "routine::invoke"),
		),
	}
}

// invoke validates and converts arguments, then invokes the function.
// It handles type conversion and validates argument counts and types.
//
// Input:
//   - fn: The function to be invoked. It must be a valid function type.
//   - args: A list of arguments to pass to the function.
//
// Output:
//   - None (panics are recovered by the caller)
func invoke(fn any, args []any) {
	invoker := newFunctionInvoker()
	invoker.invoke(fn, args)
}

// invoke validates arguments and invokes the function.
func (f *functionInvoker) invoke(fn any, args []any) {
	funcValue := reflect.ValueOf(fn)
	if funcValue.Kind() != reflect.Func {
		f.logger.Error("provided value is not a function")
		return
	}

	funcType := funcValue.Type()
	numIn := funcType.NumIn()

	if err := f.validateArguments(args, numIn); err != nil {
		f.logger.Errorw("argument validation failed",
			"error", err.Error(),
		)
		return
	}

	funcArgs, err := f.convertArguments(args, funcType, numIn)
	if err != nil {
		f.logger.Errorw("argument conversion failed",
			"error", err.Error(),
		)
		return
	}

	callWithPanicRecovery(funcValue, funcArgs)
}

// validateArguments checks if the argument count is valid.
func (f *functionInvoker) validateArguments(args []any, expected int) error {
	if len(args) < expected {
		return fmt.Errorf("insufficient arguments: expected %d, got %d", expected, len(args))
	}
	if len(args) > expected {
		f.logger.Warnw("excess arguments provided",
			"expected", expected,
			"provided", len(args),
		)
	}
	return nil
}

// convertArguments converts arguments to reflect.Value with type checking and conversion.
func (f *functionInvoker) convertArguments(args []any, funcType reflect.Type, numIn int) ([]reflect.Value, error) {
	funcArgs := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		paramType := funcType.In(i)
		argValue := reflect.ValueOf(args[i])

		// Handle nil values
		if !argValue.IsValid() {
			// Allow nil for pointer and interface types only
			if paramType.Kind() == reflect.Ptr || paramType.Kind() == reflect.Interface {
				funcArgs[i] = reflect.New(paramType).Elem() // Zero value for pointer/interface type
				continue
			}
			return nil, fmt.Errorf("invalid argument at index %d", i)
		}

		var converted reflect.Value
		switch {
		case argValue.Type().AssignableTo(paramType):
			converted = argValue
		case argValue.Type().ConvertibleTo(paramType):
			converted = argValue.Convert(paramType)
		default:
			return nil, fmt.Errorf("type mismatch at index %d: expected %s, got %s",
				i, paramType.String(), argValue.Type().String())
		}

		funcArgs[i] = converted
	}

	return funcArgs, nil
}

// callWithPanicRecovery executes the function and allows panic to propagate.
// Panic will be recovered by the outer recoverPanic defer.
func callWithPanicRecovery(funcValue reflect.Value, funcArgs []reflect.Value) {
	funcValue.Call(funcArgs)
}

// capturePanicLocation captures the panic location from the current stack trace.
func capturePanicLocation() string {
	bufPtr := stackBufferPool.Get().(*[]byte)
	defer stackBufferPool.Put(bufPtr)

	buf := *bufPtr
	n := runtime.Stack(buf, false)
	parser := newStackTraceParser()
	stackTrace := string(buf[:n])
	return parser.extractLocation(stackTrace)
}

// panicRecoverer handles panic recovery and logging.
type panicRecoverer struct {
	callerLocation string
	logger         *logger.Logger
}

// recoverPanic handles panic recovery with optional caller location.
func recoverPanic(callerLocation string) {
	r := recover()
	if r == nil {
		return
	}

	// Capture location from stack trace immediately after recover
	// Stack trace still contains panic information at this point
	location := capturePanicLocation()
	if location == "" {
		location = callerLocation // Fallback to caller location
	}

	recoverer := &panicRecoverer{
		callerLocation: callerLocation,
		logger: logger.NewLoggerWithFields(
			logger.String("prefix", "routine::recoverPanic"),
		),
	}

	err := normalizePanicValue(r)
	recoverer.log(err, location)
}

// log logs the panic with error and location information.
func (p *panicRecoverer) log(err error, location string) {
	fields := []interface{}{
		"error", err.Error(),
	}

	if location != "" {
		fields = append(fields, "panic_at", location)
	}

	p.logger.Errorw("panic recovered in goroutine", fields...)
}

// normalizePanicValue converts a panic value to an error.
// It handles error, string, and other types appropriately.
func normalizePanicValue(r any) error {
	switch v := r.(type) {
	case error:
		return v
	case string:
		return fmt.Errorf("%s", v)
	default:
		return fmt.Errorf("%v", v)
	}
}
