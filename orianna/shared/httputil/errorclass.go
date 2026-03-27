// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package httputil

// ErrorClassFromStatus returns a bounded error classification for metric labels.
// This prevents unbounded cardinality from user-defined status codes.
//
// Returns:
//   - "server_error" for 5xx
//   - "client_error" for 4xx
//   - "none" for all others
func ErrorClassFromStatus(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode >= 400:
		return "client_error"
	default:
		return "none"
	}
}
