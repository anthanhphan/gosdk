// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// Duration – JSON unmarshalling
// ============================================================================

type durationWrapper struct {
	D redis.Duration `json:"d" yaml:"d"`
}

func TestDuration_UnmarshalJSON_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  time.Duration
	}{
		{`{"d":"5s"}`, 5 * time.Second},
		{`{"d":"100ms"}`, 100 * time.Millisecond},
		{`{"d":"1m30s"}`, 90 * time.Second},
		{`{"d":"0"}`, 0},
		{`{"d":""}`, 0},
	}
	for _, tc := range cases {
		var w durationWrapper
		require.NoError(t, json.Unmarshal([]byte(tc.input), &w), tc.input)
		assert.Equal(t, tc.want, w.D.Duration, tc.input)
	}
}

func TestDuration_UnmarshalJSON_Invalid(t *testing.T) {
	var w durationWrapper
	err := json.Unmarshal([]byte(`{"d":"notaduration"}`), &w)
	assert.Error(t, err)
}

// ============================================================================
// Duration – YAML unmarshalling
// ============================================================================

func TestDuration_UnmarshalYAML_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  time.Duration
	}{
		{"d: \"5s\"", 5 * time.Second},
		{"d: \"300ms\"", 300 * time.Millisecond},
		{"d: \"\"", 0},
		{"d: \"0\"", 0}, // "0" zero-string branch
	}
	for _, tc := range cases {
		var w durationWrapper
		require.NoError(t, yaml.Unmarshal([]byte(tc.input), &w), tc.input)
		assert.Equal(t, tc.want, w.D.Duration, tc.input)
	}
}

func TestDuration_UnmarshalYAML_Invalid(t *testing.T) {
	var w durationWrapper
	err := yaml.Unmarshal([]byte(`d: "badvalue"`), &w)
	assert.Error(t, err)
}

// UnmarshalYAML – unmarshal() itself errors (wrong YAML type: int → string fails decode)
func TestDuration_UnmarshalYAML_WrongType(t *testing.T) {
	type badWrapper struct {
		D redis.Duration `yaml:"d"`
	}
	// Passing a mapping where a scalar is expected makes gopkg.in/yaml.v3 fail the string decode.
	var w badWrapper
	err := yaml.Unmarshal([]byte("d:\n  nested: true"), &w)
	assert.Error(t, err)
}

// ============================================================================
// Config.Validate
// ============================================================================

func TestConfig_Validate_NilConfig(t *testing.T) {
	var c *redis.Config
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil is not allowed")
}

func TestConfig_Validate_NoAddrNoSentinel(t *testing.T) {
	c := &redis.Config{}
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "addr")
}

func TestConfig_Validate_SentinelMissingMasterName(t *testing.T) {
	c := &redis.Config{SentinelAddrs: []string{"s1:26379"}}
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "master_name")
}

func TestConfig_Validate_SentinelMissingAddrs(t *testing.T) {
	c := &redis.Config{MasterName: "mymaster"}
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sentinel_addrs")
}

func TestConfig_Validate_DBOutOfRange(t *testing.T) {
	c := &redis.Config{Addr: "localhost:6379", DB: 16}
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db")
}

func TestConfig_Validate_NegativePoolSize(t *testing.T) {
	c := &redis.Config{Addr: "localhost:6379", PoolSize: -1}
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pool_size")
}

func TestConfig_Validate_NegativeMinIdleConns(t *testing.T) {
	c := &redis.Config{Addr: "localhost:6379", MinIdleConns: -1}
	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min_idle_conns")
}

func TestConfig_Validate_ValidStandalone(t *testing.T) {
	c := &redis.Config{Addr: "localhost:6379"}
	assert.NoError(t, c.Validate())
}

func TestConfig_Validate_ValidSentinel(t *testing.T) {
	c := &redis.Config{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"s1:26379", "s2:26379"},
	}
	assert.NoError(t, c.Validate())
}

// ============================================================================
// NewFromConfig – nil guards
// ============================================================================

func TestNewClient_NilConfig_FromConfigTest(t *testing.T) {
	_, err := redis.NewClient(nil, nil)
	assert.Error(t, err)
}

func TestNewClient_InvalidConfig_FromConfigTest(t *testing.T) {
	_, err := redis.NewClient(&redis.Config{}, nil)
	assert.Error(t, err)
}

// ============================================================================
// Config parsing from YAML / JSON (simulating conflux.Load)
// ============================================================================

func TestConfig_YAML_Standalone(t *testing.T) {
	raw := `
addr: "localhost:6379"
password: "secret"
db: 1
pool_size: 20
min_idle_conns: 5
dial_timeout: "4s"
read_timeout: "2s"
write_timeout: "2s"
metric_namespace: "myapp"
metric_subsystem: "cache"
`
	var cfg redis.Config
	require.NoError(t, yaml.Unmarshal([]byte(raw), &cfg))
	require.NoError(t, cfg.Validate())

	assert.Equal(t, "localhost:6379", cfg.Addr)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, 1, cfg.DB)
	assert.Equal(t, 20, cfg.PoolSize)
	assert.Equal(t, 5, cfg.MinIdleConns)
	assert.Equal(t, 4*time.Second, cfg.DialTimeout.Duration)
	assert.Equal(t, 2*time.Second, cfg.ReadTimeout.Duration)
	assert.Equal(t, 2*time.Second, cfg.WriteTimeout.Duration)
	assert.Equal(t, "myapp", cfg.MetricNamespace)
	assert.Equal(t, "cache", cfg.MetricSubsystem)
}

func TestConfig_YAML_Sentinel(t *testing.T) {
	raw := `
master_name: "mymaster"
sentinel_addrs:
  - "s1:26379"
  - "s2:26379"
sentinel_password: "sentpass"
password: "secret"
db: 0
pool_size: 10
min_idle_conns: 2
dial_timeout: "5s"
read_timeout: "3s"
write_timeout: "3s"
`
	var cfg redis.Config
	require.NoError(t, yaml.Unmarshal([]byte(raw), &cfg))
	require.NoError(t, cfg.Validate())

	assert.Equal(t, "mymaster", cfg.MasterName)
	assert.Equal(t, []string{"s1:26379", "s2:26379"}, cfg.SentinelAddrs)
	assert.Equal(t, "sentpass", cfg.SentinelPassword)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout.Duration)
}

func TestConfig_JSON_Standalone(t *testing.T) {
	raw := `{
		"addr": "redis:6379",
		"password": "pw",
		"db": 2,
		"pool_size": 15,
		"min_idle_conns": 3,
		"dial_timeout": "3s",
		"read_timeout": "1s",
		"write_timeout": "1s",
		"metric_namespace": "svc"
	}`
	var cfg redis.Config
	require.NoError(t, json.Unmarshal([]byte(raw), &cfg))
	require.NoError(t, cfg.Validate())

	assert.Equal(t, "redis:6379", cfg.Addr)
	assert.Equal(t, 2, cfg.DB)
	assert.Equal(t, 3*time.Second, cfg.DialTimeout.Duration)
	assert.Equal(t, "svc", cfg.MetricNamespace)
}
