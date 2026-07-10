package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	t.Run("Network defaults", func(t *testing.T) {
		if !cfg.Network.MDNSEnabled {
			t.Error("DefaultConfig().Network.MDNSEnabled should be true")
		}
		if cfg.Network.ListenPort != 0 {
			t.Errorf("DefaultConfig().Network.ListenPort = %d, want 0", cfg.Network.ListenPort)
		}
	})

	t.Run("Screen defaults", func(t *testing.T) {
		if cfg.Screen.MaxFPS != 30 {
			t.Errorf("DefaultConfig().Screen.MaxFPS = %d, want 30", cfg.Screen.MaxFPS)
		}
		if cfg.Screen.DefaultQuality != "high" {
			t.Errorf("DefaultConfig().Screen.DefaultQuality = %q, want \"high\"", cfg.Screen.DefaultQuality)
		}
		if cfg.Screen.CaptureBackend != "auto" {
			t.Errorf("DefaultConfig().Screen.CaptureBackend = %q, want \"auto\"", cfg.Screen.CaptureBackend)
		}
	})

	t.Run("Logging defaults", func(t *testing.T) {
		if cfg.Logging.Level != "info" {
			t.Errorf("DefaultConfig().Logging.Level = %q, want \"info\"", cfg.Logging.Level)
		}
		if cfg.Logging.MaxSizeMB != 100 {
			t.Errorf("DefaultConfig().Logging.MaxSizeMB = %d, want 100", cfg.Logging.MaxSizeMB)
		}
		if cfg.Logging.MaxFiles != 5 {
			t.Errorf("DefaultConfig().Logging.MaxFiles = %d, want 5", cfg.Logging.MaxFiles)
		}
	})

	t.Run("Plugins defaults", func(t *testing.T) {
		if !cfg.Plugins.Enabled {
			t.Error("DefaultConfig().Plugins.Enabled should be true")
		}
	})

	t.Run("Runtime defaults", func(t *testing.T) {
		if cfg.Runtime.DeviceName != "" {
			t.Errorf("DefaultConfig().Runtime.DeviceName = %q, want \"\"", cfg.Runtime.DeviceName)
		}
	})
}

func TestLoadEmpty(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load(\"\") returned nil config")
	}
	// Should return default config
	if cfg.Screen.MaxFPS != 30 {
		t.Errorf("Load(\"\") got MaxFPS = %d, want 30", cfg.Screen.MaxFPS)
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/buzzpi/config.json")
	if err != nil {
		t.Fatalf("Load(nonexistent) returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load(nonexistent) returned nil config")
	}
	// Should return default config
	if !cfg.Network.MDNSEnabled {
		t.Error("Load(nonexistent) should return defaults with MDNSEnabled=true")
	}
}

func TestLoadValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime.json")
	content := `{"screen":{"max_fps":15,"default_quality":"medium"},"logging":{"level":"debug"}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(valid) returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load(valid) returned nil config")
	}

	// Overridden values
	if cfg.Screen.MaxFPS != 15 {
		t.Errorf("Screen.MaxFPS = %d, want 15", cfg.Screen.MaxFPS)
	}
	if cfg.Screen.DefaultQuality != "medium" {
		t.Errorf("Screen.DefaultQuality = %q, want \"medium\"", cfg.Screen.DefaultQuality)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want \"debug\"", cfg.Logging.Level)
	}

	// Default values should remain
	if !cfg.Network.MDNSEnabled {
		t.Error("Network.MDNSEnabled should still be true (default)")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	content := `{invalid json content}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load(invalid JSON) expected an error, got nil")
	}
}

func TestDefaultConfigPaths(t *testing.T) {
	paths := DefaultConfigPaths()
	if len(paths) < 2 {
		t.Errorf("DefaultConfigPaths() returned %d paths, want at least 2", len(paths))
	}
	for _, p := range paths {
		if p == "" {
			t.Error("DefaultConfigPaths() contains empty path")
		}
	}
}
