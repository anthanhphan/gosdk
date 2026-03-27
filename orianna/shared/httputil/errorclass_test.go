// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package httputil

import "testing"

func TestErrorClassFromStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       string
	}{
		{"200 OK", 200, "none"},
		{"201 Created", 201, "none"},
		{"204 No Content", 204, "none"},
		{"301 Redirect", 301, "none"},
		{"400 Bad Request", 400, "client_error"},
		{"401 Unauthorized", 401, "client_error"},
		{"403 Forbidden", 403, "client_error"},
		{"404 Not Found", 404, "client_error"},
		{"429 Too Many Requests", 429, "client_error"},
		{"499 Client Closed", 499, "client_error"},
		{"500 Internal Server Error", 500, "server_error"},
		{"502 Bad Gateway", 502, "server_error"},
		{"503 Service Unavailable", 503, "server_error"},
		{"504 Gateway Timeout", 504, "server_error"},
		{"0 Unknown", 0, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorClassFromStatus(tt.statusCode)
			if got != tt.want {
				t.Errorf("ErrorClassFromStatus(%d) = %q, want %q", tt.statusCode, got, tt.want)
			}
		})
	}
}
