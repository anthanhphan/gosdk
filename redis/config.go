// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package redis

import (
	"errors"
	"fmt"
	"time"
)

// ============================================================================
// Config – conflux-compatible struct
// ============================================================================

// Config holds all Redis connection settings.
// It is designed to be loaded via conflux.Load[redis.Config] from JSON or YAML files
// and then passed to NewFromConfig.
//
// Standalone example (YAML):
//
//	redis:
//	  addr: "localhost:6379"
//	  password: "s3cr3t"
//	  db: 0
//	  pool_size: 10
//	  min_idle_conns: 2
//	  dial_timeout: "5s"
//	  read_timeout:  "3s"
//	  write_timeout: "3s"
//	  metric_namespace: "myapp"
//	  metric_subsystem: "cache"
//
// Sentinel example (YAML):
//
//	redis:
//	  master_name: "mymaster"
//	  sentinel_addrs:
//	    - "sentinel1:26379"
//	    - "sentinel2:26379"
//	  sentinel_password: "sentinel-pass"
//	  password: "s3cr3t"
//	  db: 0
//	  pool_size: 20
//	  min_idle_conns: 5
//	  dial_timeout: "5s"
//	  read_timeout:  "3s"
//	  write_timeout: "3s"
//	  metric_namespace: "myapp"
//	  metric_subsystem: "cache"
type Config struct {
	// --- Standalone ---

	// Addr is the address of the standalone Redis server (host:port).
	// Required when not using Sentinel mode.
	Addr string `json:"addr" yaml:"addr"`

	// --- Sentinel ---

	// MasterName is the Sentinel master name (e.g. "mymaster").
	// Required when using Sentinel mode. Must be set together with SentinelAddrs.
	MasterName string `json:"master_name" yaml:"master_name"`

	// SentinelAddrs is the list of Sentinel node addresses (host:port).
	// Required when using Sentinel mode.
	SentinelAddrs []string `json:"sentinel_addrs" yaml:"sentinel_addrs"`

	// SentinelPassword is the AUTH password for Sentinel nodes.
	SentinelPassword string `json:"sentinel_password" yaml:"sentinel_password"`

	// --- Common ---

	// Password is the Redis server AUTH password.
	Password string `json:"password" yaml:"password"`

	// DB is the Redis logical database index (0–15).
	DB int `json:"db" yaml:"db"`

	// --- Pool ---

	// PoolSize is the maximum number of socket connections in the pool.
	// Defaults to 10 when 0.
	PoolSize int `json:"pool_size" yaml:"pool_size"`

	// MinIdleConns is the minimum number of idle connections maintained in the pool.
	// Defaults to 2 when 0.
	MinIdleConns int `json:"min_idle_conns" yaml:"min_idle_conns"`

	// --- Timeouts ---

	// DialTimeout is the timeout for establishing a new connection.
	// Accepts any duration string understood by time.ParseDuration (e.g. "5s", "500ms").
	// Defaults to "5s" when empty.
	DialTimeout Duration `json:"dial_timeout" yaml:"dial_timeout"`

	// ReadTimeout is the timeout for socket reads.
	// Defaults to "3s" when empty.
	ReadTimeout Duration `json:"read_timeout" yaml:"read_timeout"`

	// WriteTimeout is the timeout for socket writes.
	// Defaults to "3s" when empty.
	WriteTimeout Duration `json:"write_timeout" yaml:"write_timeout"`

	// --- Metrics ---

	// MetricNamespace is the Prometheus namespace prefix for Redis metrics.
	// Defaults to "redis" when empty.
	MetricNamespace string `json:"metric_namespace" yaml:"metric_namespace"`

	// MetricSubsystem is the Prometheus subsystem inserted between namespace and metric name.
	MetricSubsystem string `json:"metric_subsystem" yaml:"metric_subsystem"`
}

// Validate checks that the Config is valid and consistent.
// Call this after loading with conflux if you want early feedback,
// or rely on NewFromConfig which calls it automatically.
//
// Example:
//
//	cfg, err := conflux.Load[redis.Config]("config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal("invalid redis config:", err)
//	}
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required, nil is not allowed")
	}

	isSentinel := c.MasterName != "" || len(c.SentinelAddrs) > 0
	isStandalone := c.Addr != ""

	if !isSentinel && !isStandalone {
		return errors.New("either addr (standalone) or master_name + sentinel_addrs (sentinel) must be set")
	}
	if isSentinel {
		if c.MasterName == "" {
			return errors.New("master_name is required in sentinel mode")
		}
		if len(c.SentinelAddrs) == 0 {
			return errors.New("sentinel_addrs must not be empty in sentinel mode")
		}
	}

	if c.DB < 0 || c.DB > 15 {
		return fmt.Errorf("db must be between 0 and 15, got %d", c.DB)
	}
	if c.PoolSize < 0 {
		return fmt.Errorf("pool_size must be >= 0, got %d", c.PoolSize)
	}
	if c.MinIdleConns < 0 {
		return fmt.Errorf("min_idle_conns must be >= 0, got %d", c.MinIdleConns)
	}

	return nil
}

// applyDefaults fills in zero-value fields with sensible production defaults.
func (c *Config) applyDefaults() {
	if c.PoolSize == 0 {
		c.PoolSize = 10
	}
	if c.MinIdleConns == 0 {
		c.MinIdleConns = 2
	}
	if c.DialTimeout.Duration == 0 {
		c.DialTimeout.Duration = 5 * time.Second
	}
	if c.ReadTimeout.Duration == 0 {
		c.ReadTimeout.Duration = 3 * time.Second
	}
	if c.WriteTimeout.Duration == 0 {
		c.WriteTimeout.Duration = 3 * time.Second
	}
	if c.MetricNamespace == "" {
		c.MetricNamespace = "redis"
	}
}

// ============================================================================
// Duration – JSON / YAML friendly time.Duration
// ============================================================================

// Duration is a time.Duration that unmarshals from a human-readable string
// (e.g. "5s", "100ms", "1m30s") in both JSON and YAML.
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	s := string(b)
	// Strip surrounding JSON string quotes.
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	if s == "" || s == "0" {
		d.Duration = 0
		return nil
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" || s == "0" {
		d.Duration = 0
		return nil
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}
