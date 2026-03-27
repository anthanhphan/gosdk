// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package observability

import "testing"

func TestCodeStringCache(t *testing.T) {
	cache := CodeStringCache([]int{200, 404, 500})

	tests := []struct {
		code int
		want string
	}{
		{200, "200"},
		{404, "404"},
		{500, "500"},
		{201, "201"}, // not cached, should fallback
	}

	for _, tt := range tests {
		got := CodeString(cache, tt.code)
		if got != tt.want {
			t.Errorf("CodeString(%d) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestCodeStringCacheEmpty(t *testing.T) {
	cache := CodeStringCache(nil)
	got := CodeString(cache, 42)
	if got != "42" {
		t.Errorf("CodeString(42) with empty cache = %q, want %q", got, "42")
	}
}

func TestAttemptString_Cached(t *testing.T) {
	for i := 0; i <= 10; i++ {
		got := AttemptString(i)
		want := attemptCache[i]
		if got != want {
			t.Errorf("AttemptString(%d) = %q, want %q", i, got, want)
		}
	}
}

func TestAttemptString_Uncached(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{11, "11"},
		{100, "100"},
		{-1, "-1"},
	}
	for _, tt := range tests {
		got := AttemptString(tt.input)
		if got != tt.want {
			t.Errorf("AttemptString(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
