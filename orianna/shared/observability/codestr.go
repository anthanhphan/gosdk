// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

// Package observability provides shared observability utilities used across
// both HTTP middleware and gRPC interceptors.
package observability

import "strconv"

// CodeStringCache builds a map of code → string for the given codes.
// Pre-computing these avoids strconv.Itoa allocation on the hot path.
func CodeStringCache(codes []int) map[int]string {
	m := make(map[int]string, len(codes))
	for _, c := range codes {
		m[c] = strconv.Itoa(c)
	}
	return m
}

// CodeString returns the cached string for the given code, falling back
// to strconv.Itoa for uncached values.
func CodeString(cache map[int]string, code int) string {
	if s, ok := cache[code]; ok {
		return s
	}
	return strconv.Itoa(code)
}

// attemptCache pre-computes string representations of common attempt numbers (0-10).
// Covers 99%+ of real-world retry scenarios without strconv.Itoa allocation.
var attemptCache = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

// AttemptString returns a cached string for attempt numbers 0-10,
// falling back to strconv.Itoa for larger values.
func AttemptString(attempt int) string {
	if attempt >= 0 && attempt < len(attemptCache) {
		return attemptCache[attempt]
	}
	return strconv.Itoa(attempt)
}
