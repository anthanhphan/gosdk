// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package routine

import (
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/anthanhphan/gosdk/utils"
)

// ---------------------------------------------------------------------------
// Stack trace parsing & caller location
// ---------------------------------------------------------------------------

const (
	stackTraceBufferSize = 4096
	callerDepth          = 8
)

var (
	// stackBufferPool reuses buffers for stack trace operations to reduce allocations.
	stackBufferPool = sync.Pool{
		New: func() any {
			buf := make([]byte, stackTraceBufferSize)
			return &buf
		},
	}

	// callerPCPool reuses []uintptr slices for runtime.Callers to avoid allocation per Run() call.
	callerPCPool = sync.Pool{
		New: func() any {
			s := make([]uintptr, callerDepth)
			return &s
		},
	}

	// callerPathCache caches file paths -> short paths to avoid repeated GetShortPath calls.
	callerPathCache sync.Map // map[string]string

	// defaultFilter configures which stack frames to skip when locating panic origin.
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

	// defaultParser is a cached singleton stack trace parser (stateless).
	defaultParser = &stackTraceParser{filter: &defaultFilter}
)

// ---------------------------------------------------------------------------
// Caller location (hot path)
// ---------------------------------------------------------------------------

// getCallerLocation captures the caller's location (where Run() was called from).
func getCallerLocation() string {
	pcsPtr := callerPCPool.Get().(*[]uintptr)
	pcs := *pcsPtr
	n := runtime.Callers(2, pcs) // Skip getCallerLocation and caller
	frames := runtime.CallersFrames(pcs[:n])
	callerPCPool.Put(pcsPtr)

	for {
		frame, more := frames.Next()
		if frame.File != "" && !strings.Contains(frame.File, "/goroutine/") {
			short := getCachedShortPath(frame.File)
			return short + ":" + strconv.Itoa(frame.Line)
		}
		if !more {
			break
		}
	}

	return ""
}

// getCachedShortPath returns the cached short path for a file, computing and caching on first call.
func getCachedShortPath(file string) string {
	if cached, ok := callerPathCache.Load(file); ok {
		return cached.(string)
	}
	short := utils.GetShortPath(file)
	callerPathCache.Store(file, short)
	return short
}

// ---------------------------------------------------------------------------
// Stack trace parser (cold path -- only during panic recovery)
// ---------------------------------------------------------------------------

// stackTraceParser extracts panic location from stack traces.
type stackTraceParser struct {
	filter *frameFilter
}

// newStackTraceParser returns the cached singleton stack trace parser.
func newStackTraceParser() *stackTraceParser {
	return defaultParser
}

// extractLocation extracts the panic location from a stack trace string.
// Uses index-based parsing -- no strings.Split, no intermediate slice allocation.
func (p *stackTraceParser) extractLocation(stackTrace string) string {
	// Skip the first line (goroutine header: "goroutine N [status]:")
	idx := strings.IndexByte(stackTrace, '\n')
	if idx < 0 {
		return ""
	}
	rest := stackTrace[idx+1:]

	// Parse frames: each frame has 2 lines (func name, file:line)
	for rest != "" {
		// Line 1: function name
		funcEnd := strings.IndexByte(rest, '\n')
		if funcEnd < 0 {
			break
		}
		funcName := strings.TrimSpace(rest[:funcEnd])
		rest = rest[funcEnd+1:]

		// Line 2: file path
		fileEnd := strings.IndexByte(rest, '\n')
		var filePath string
		if fileEnd < 0 {
			filePath = strings.TrimSpace(rest)
			rest = ""
		} else {
			filePath = strings.TrimSpace(rest[:fileEnd])
			rest = rest[fileEnd+1:]
		}

		if p.filter.shouldSkip(funcName, filePath) {
			continue
		}

		if loc := formatLocation(filePath); loc != "" {
			return loc
		}
	}

	return ""
}

// formatLocation extracts file:line from a stack trace file path string and formats it.
func formatLocation(filePath string) string {
	colonIdx := strings.LastIndexByte(filePath, ':')
	if colonIdx < 0 || strings.HasSuffix(filePath, ")") {
		return ""
	}

	file := filePath[:colonIdx]
	lineStr := filePath[colonIdx+1:]

	// Trim any trailing whitespace or offset info (e.g. " +0x1c")
	if spaceIdx := strings.IndexByte(lineStr, ' '); spaceIdx >= 0 {
		lineStr = lineStr[:spaceIdx]
	}

	if lineStr == "" {
		return ""
	}

	short := getCachedShortPath(file)
	return short + ":" + lineStr
}

// ---------------------------------------------------------------------------
// Frame filtering
// ---------------------------------------------------------------------------

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
		containsAny(funcName, f.prefixPatterns)
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

// ---------------------------------------------------------------------------
// Legacy helpers (used by tests)
// ---------------------------------------------------------------------------

// stackFrame represents a single frame in the stack trace.
type stackFrame struct {
	funcName string
	filePath string
}

// location extracts and formats the location string for this frame.
func (f stackFrame) location() string {
	return formatLocation(f.filePath)
}

// isValidFilePath checks if a file path string is valid for parsing.
func isValidFilePath(path string) bool {
	return strings.Contains(path, ":") && !strings.HasSuffix(path, ")")
}

// parseFilePath extracts file and line number from a file path string.
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
func parseStackFrames(lines []string) []stackFrame {
	const (
		goroutineHeaderOffset = 1
		framesPerEntry        = 2
	)

	frames := make([]stackFrame, 0, len(lines)/framesPerEntry)

	for i := goroutineHeaderOffset; i+1 < len(lines); i += framesPerEntry {
		frames = append(frames, stackFrame{
			funcName: strings.TrimSpace(lines[i]),
			filePath: strings.TrimSpace(lines[i+1]),
		})
	}

	return frames
}
