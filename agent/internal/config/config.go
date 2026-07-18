// Package config provides runtime configuration loading and validation.
// Configuration is loaded from file, environment variables, and CLI flags
// in that order (later overrides earlier).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the full runtime configuration.
type Config struct {
	Runtime RuntimeConfig `json:"runtime"`
	Network NetworkConfig `json:"network"`
	Screen  ScreenConfig  `json:"screen"`
	Plugins PluginsConfig `json:"plugins"`
	Logging LoggingConfig `json:"logging"`
}

// RuntimeConfig contains device identity and naming configuration.
type RuntimeConfig struct {
	DeviceName string `json:"device_name"`
}

// NetworkConfig contains networking and relay configuration.
type NetworkConfig struct {
	RelayServers []string `json:"relay_servers"`
	ListenPort   int      `json:"listen_port"`
	MDNSEnabled  bool     `json:"mdns_enabled"`
}

// ConnectionConfig contains client connection limits.
type ConnectionConfig struct {
	MaxClients        int           `json:"max_clients"`
	SessionTimeout    time.Duration `json:"session_timeout"`
	KeepaliveInterval time.Duration `json:"keepalive_interval"`
}

// ScreenConfig contains screen capture settings.
type ScreenConfig struct {
	CaptureBackend string `json:"capture_backend"`
	MaxFPS         int    `json:"max_fps"`
	DefaultQuality string `json:"default_quality"`
}

// PluginsConfig contains plugin host settings.
type PluginsConfig struct {
	Enabled      bool   `json:"enabled"`
	Directory    string `json:"directory"`
	AllowNetwork bool   `json:"allow_network"`
}

// LoggingConfig contains log output settings.
type LoggingConfig struct {
	Level     string `json:"level"`
	File      string `json:"file"`
	MaxSizeMB int    `json:"max_size_mb"`
	MaxFiles  int    `json:"max_files"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Runtime: RuntimeConfig{
			DeviceName: "",
		},
		Network: NetworkConfig{
			RelayServers: []string{},
			ListenPort:   0, // random port
			MDNSEnabled:  true,
		},
		Screen: ScreenConfig{
			CaptureBackend: "auto",
			MaxFPS:         30,
			DefaultQuality: "high",
		},
		Plugins: PluginsConfig{
			Enabled:      true,
			Directory:    "/var/lib/buzzpi/plugins",
			AllowNetwork: false,
		},
		Logging: LoggingConfig{
			Level:     "info",
			File:      "",
			MaxSizeMB: 100,
			MaxFiles:  5,
		},
	}
}

// Load reads configuration from the specified file path.
// If path is empty, returns default config.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("read config: %w", err)
			}
			// File doesn't exist — proceed with defaults
		} else {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config: %w", err)
			}
		}
	}

	return cfg, nil
}

// DefaultConfigPaths returns the standard config file search paths.
func DefaultConfigPaths() []string {
	return []string{
		"/etc/buzzpi/runtime.yaml",
		"/etc/buzzpi/runtime.json",
		filepath.Join(os.Getenv("HOME"), ".buzzpi", "runtime.yaml"),
		filepath.Join(os.Getenv("HOME"), ".buzzpi", "runtime.json"),
	}
}

// Save writes the configuration to the specified file path as JSON.
func Save(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
