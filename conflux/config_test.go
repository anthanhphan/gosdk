// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package conflux

import (
	"os"
	"strings"
	"testing"
)

// ============================================================================
// Test Config Types
// ============================================================================

type testConfig struct {
	DatabaseURL string `json:"database_url" yaml:"database_url" validate:"required"`
	Port        int    `json:"port" yaml:"port" validate:"required,min=0,max=65535"`
	Debug       bool   `json:"debug" yaml:"debug"`
}

// ============================================================================
// Helpers
// ============================================================================

func setupTempDir(t *testing.T) (cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "conflux_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	return func() {
		_ = os.Chdir(origDir)
		_ = os.RemoveAll(tmpDir)
	}
}

func writeFile(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// ============================================================================
// Load
// ============================================================================

func TestLoad_JSON(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "config.json", `{"database_url":"postgres://localhost","port":8080,"debug":true}`)
	cfg, err := Load[testConfig]("config.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://localhost" {
		t.Errorf("database_url = %q", cfg.DatabaseURL)
	}
	if cfg.Port != 8080 {
		t.Errorf("port = %d", cfg.Port)
	}
	if !cfg.Debug {
		t.Error("debug should be true")
	}
}

func TestLoad_YAML(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "config.yaml", "database_url: postgres://localhost\nport: 8080\ndebug: true")
	cfg, err := Load[testConfig]("config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("port = %d", cfg.Port)
	}
}

func TestLoad_YML(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "config.yml", "database_url: pg://localhost\nport: 3000")
	cfg, err := Load[testConfig]("config.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("port = %d", cfg.Port)
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	_, err := Load[testConfig]("")
	if err == nil || !strings.Contains(err.Error(), "config path is required") {
		t.Errorf("expected path required error, got: %v", err)
	}
}

func TestLoad_UnsupportedExtension(t *testing.T) {
	_, err := Load[testConfig]("config.xml")
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("expected unsupported extension error, got: %v", err)
	}
}

func TestLoad_NoExtension(t *testing.T) {
	_, err := Load[testConfig]("config")
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("expected unsupported extension error, got: %v", err)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	_, err := Load[testConfig]("nonexistent.json")
	if err == nil || !strings.Contains(err.Error(), "failed to read config") {
		t.Errorf("expected file not found error, got: %v", err)
	}
}

func TestLoad_DirectoryTraversal(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	_, err := Load[testConfig]("../config.json")
	if err == nil || !strings.Contains(err.Error(), "invalid path") {
		t.Errorf("expected traversal error, got: %v", err)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "bad.json", `{invalid json}`)
	_, err := Load[testConfig]("bad.json")
	if err == nil || !strings.Contains(err.Error(), "failed to unmarshal json") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "bad.yaml", "invalid: yaml: content")
	_, err := Load[testConfig]("bad.yaml")
	if err == nil || !strings.Contains(err.Error(), "failed to unmarshal yaml") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

func TestLoad_ValidationFailed_MissingRequired(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "missing.yaml", "port: 8080\ndebug: true")
	_, err := Load[testConfig]("missing.yaml")
	if err == nil || !strings.Contains(err.Error(), "config validation failed") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

func TestLoad_ValidationFailed_PortOutOfRange(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "bigport.yaml", "database_url: pg://localhost\nport: 70000")
	_, err := Load[testConfig]("bigport.yaml")
	if err == nil || !strings.Contains(err.Error(), "config validation failed") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

func TestLoad_PortBoundary(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	// Max
	writeFile(t, "max.yaml", "database_url: pg://x\nport: 65535")
	cfg, err := Load[testConfig]("max.yaml")
	if err != nil {
		t.Fatalf("max port should pass: %v", err)
	}
	if cfg.Port != 65535 {
		t.Errorf("port = %d", cfg.Port)
	}

	// Min
	writeFile(t, "min.yaml", "database_url: pg://x\nport: 1")
	cfg, err = Load[testConfig]("min.yaml")
	if err != nil {
		t.Fatalf("min port should pass: %v", err)
	}
	if cfg.Port != 1 {
		t.Errorf("port = %d", cfg.Port)
	}
}

// ============================================================================
// Load — nested validation
// ============================================================================

func TestLoad_NestedValidation(t *testing.T) {
	type server struct {
		Port int    `yaml:"port" validate:"required,min=1,max=65535"`
		Host string `yaml:"host" validate:"required"`
	}
	type app struct {
		Server server `yaml:"server"`
	}

	cleanup := setupTempDir(t)
	defer cleanup()

	// Valid
	writeFile(t, "valid.yaml", "server:\n  port: 8080\n  host: localhost")
	cfg, err := Load[app]("valid.yaml")
	if err != nil {
		t.Fatalf("valid nested: %v", err)
	}
	if cfg.Server.Port != 8080 || cfg.Server.Host != "localhost" {
		t.Errorf("got %+v", cfg.Server)
	}

	// Invalid — missing host
	writeFile(t, "bad.yaml", "server:\n  port: 8080")
	_, err = Load[app]("bad.yaml")
	if err == nil || !strings.Contains(err.Error(), "config validation failed") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

// ============================================================================
// MustLoad
// ============================================================================

func TestMustLoad_Success(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "ok.yaml", "database_url: pg://x\nport: 8080")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("should not panic: %v", r)
		}
	}()

	cfg := MustLoad[testConfig]("ok.yaml")
	if cfg.Port != 8080 {
		t.Errorf("port = %d", cfg.Port)
	}
}

func TestMustLoad_Panics_FileNotFound(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	MustLoad[testConfig]("nonexistent.yaml")
}

func TestMustLoad_Panics_ValidationFailed(t *testing.T) {
	cleanup := setupTempDir(t)
	defer cleanup()

	writeFile(t, "bad.yaml", "database_url: pg://x\nport: 99999")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on validation failure")
		}
	}()

	MustLoad[testConfig]("bad.yaml")
}

// ============================================================================
// unmarshal
// ============================================================================

func TestUnmarshal_JSON(t *testing.T) {
	var cfg testConfig
	err := unmarshal([]byte(`{"database_url":"pg","port":8080}`), ExtensionJSON, &cfg)
	if err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("port = %d", cfg.Port)
	}
}

func TestUnmarshal_YAML(t *testing.T) {
	var cfg testConfig
	err := unmarshal([]byte("database_url: pg\nport: 8080"), ExtensionYAML, &cfg)
	if err != nil {
		t.Fatalf("yaml unmarshal: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("port = %d", cfg.Port)
	}
}

func TestUnmarshal_YML(t *testing.T) {
	var cfg testConfig
	err := unmarshal([]byte("database_url: pg\nport: 3000"), ExtensionYML, &cfg)
	if err != nil {
		t.Fatalf("yml unmarshal: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("port = %d", cfg.Port)
	}
}

func TestUnmarshal_Unsupported(t *testing.T) {
	var cfg testConfig
	err := unmarshal([]byte("data"), "xml", &cfg)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected unsupported error, got: %v", err)
	}
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	var cfg testConfig
	err := unmarshal([]byte(`{bad}`), ExtensionJSON, &cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestUnmarshal_InvalidYAML(t *testing.T) {
	var cfg testConfig
	err := unmarshal([]byte("a: b: c"), ExtensionYAML, &cfg)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// ============================================================================
// validExts
// ============================================================================

func TestValidExts(t *testing.T) {
	cases := []struct {
		ext  string
		want bool
	}{
		{ExtensionJSON, true},
		{ExtensionYAML, true},
		{ExtensionYML, true},
		{"xml", false},
		{"txt", false},
		{"", false},
		{"JSON", false},
	}
	for _, c := range cases {
		if got := validExts[c.ext]; got != c.want {
			t.Errorf("validExts[%q] = %v, want %v", c.ext, got, c.want)
		}
	}
}
