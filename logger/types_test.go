// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"testing"
)

func TestLevel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  bool
	}{
		{
			name:  "debug level should be valid",
			level: LevelDebug,
			want:  true,
		},
		{
			name:  "info level should be valid",
			level: LevelInfo,
			want:  true,
		},
		{
			name:  "warn level should be valid",
			level: LevelWarn,
			want:  true,
		},
		{
			name:  "error level should be valid",
			level: LevelError,
			want:  true,
		},
		{
			name:  "invalid level should not be valid",
			level: Level("invalid"),
			want:  false,
		},
		{
			name:  "empty level should not be valid",
			level: Level(""),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.isValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLevelValues(t *testing.T) {
	got := levelValues()
	expected := []string{"debug", "info", "warn", "error"}

	if len(got) != len(expected) {
		t.Errorf("LevelValues() length = %v, want %v", len(got), len(expected))
		return
	}

	for i, v := range got {
		if v != expected[i] {
			t.Errorf("LevelValues()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestEncoding_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		encoding Encoding
		want     bool
	}{
		{
			name:     "json encoding should be valid",
			encoding: EncodingJSON,
			want:     true,
		},
		{
			name:     "console encoding should be valid",
			encoding: EncodingConsole,
			want:     true,
		},
		{
			name:     "invalid encoding should not be valid",
			encoding: Encoding("invalid"),
			want:     false,
		},
		{
			name:     "empty encoding should not be valid",
			encoding: Encoding(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.encoding.isValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodingValues(t *testing.T) {
	got := encodingValues()
	expected := []string{"json", "console"}

	if len(got) != len(expected) {
		t.Errorf("EncodingValues() length = %v, want %v", len(got), len(expected))
		return
	}

	for i, v := range got {
		if v != expected[i] {
			t.Errorf("EncodingValues()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}
