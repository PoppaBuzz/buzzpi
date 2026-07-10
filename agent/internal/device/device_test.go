package device

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/buzzpi/agent/internal/config"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	cfg := config.DefaultConfig()
	s, err := NewService(cfg, slog.Default())
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}
	return s
}

func TestNewService(t *testing.T) {
	t.Run("creates service with config", func(t *testing.T) {
		s := newTestService(t)
		if s == nil {
			t.Fatal("NewService() returned nil")
		}
		if s.hostname == "" {
			t.Error("hostname should not be empty")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		cfg := config.DefaultConfig()
		s, err := NewService(cfg, nil)
		if err != nil {
			t.Fatalf("NewService(nil logger) failed: %v", err)
		}
		if s == nil {
			t.Fatal("NewService(nil logger) returned nil")
		}
	})
}

func TestHandleInfo(t *testing.T) {
	s := newTestService(t)

	resp, err := s.HandleInfo(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("HandleInfo() failed: %v", err)
	}

	info, ok := resp.(*InfoResponse)
	if !ok {
		t.Fatalf("HandleInfo() returned %T, want *InfoResponse", resp)
	}

	if info.DeviceID != "dev_unknown" {
		t.Errorf("DeviceID = %q, want \"dev_unknown\"", info.DeviceID)
	}
	if info.Model != "raspberry-pi/5" {
		t.Errorf("Model = %q, want \"raspberry-pi/5\"", info.Model)
	}
	if info.RuntimeVersion == "" {
		t.Error("RuntimeVersion should not be empty")
	}
	if info.BPPVersion != 1 {
		t.Errorf("BPPVersion = %d, want 1", info.BPPVersion)
	}
	if info.UptimeSeconds < 0 {
		t.Errorf("UptimeSeconds = %d, want >= 0", info.UptimeSeconds)
	}
	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform = %q, should contain '/'", info.Platform)
	}
	if info.FriendlyName == "" {
		t.Error("FriendlyName should not be empty (falls back to hostname)")
	}
}

func TestHandleInfoFriendlyName(t *testing.T) {
	t.Run("uses device_name from config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Runtime.DeviceName = "MyBuzzPi"
		s, err := NewService(cfg, slog.Default())
		if err != nil {
			t.Fatalf("NewService() failed: %v", err)
		}

		resp, err := s.HandleInfo(context.Background(), json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("HandleInfo() failed: %v", err)
		}
		info := resp.(*InfoResponse)
		if info.FriendlyName != "MyBuzzPi" {
			t.Errorf("FriendlyName = %q, want \"MyBuzzPi\"", info.FriendlyName)
		}
	})
}

func TestHandleStats(t *testing.T) {
	s := newTestService(t)

	resp, err := s.HandleStats(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("HandleStats() failed: %v", err)
	}

	stats, ok := resp.(*StatsResponse)
	if !ok {
		t.Fatalf("HandleStats() returned %T, want *StatsResponse", resp)
	}

	if stats.Uptime < 0 {
		t.Errorf("Uptime = %d, want >= 0", stats.Uptime)
	}
	if stats.CPU.UsagePercent < 0 {
		t.Errorf("CPU.UsagePercent = %f, want >= 0", stats.CPU.UsagePercent)
	}
	if stats.CPU.TemperatureC < 0 {
		t.Errorf("CPU.TemperatureC = %f, want >= 0", stats.CPU.TemperatureC)
	}
	if stats.CPU.FrequencyMHz < 0 {
		t.Errorf("CPU.FrequencyMHz = %d, want >= 0", stats.CPU.FrequencyMHz)
	}
	if stats.Memory.TotalMB < 0 {
		t.Errorf("Memory.TotalMB = %d, want >= 0", stats.Memory.TotalMB)
	}
	if stats.Memory.AvailableMB < 0 {
		t.Errorf("Memory.AvailableMB = %d, want >= 0", stats.Memory.AvailableMB)
	}
	if stats.Memory.Percent < 0 || stats.Memory.Percent > 100 {
		t.Errorf("Memory.Percent = %f, want [0, 100]", stats.Memory.Percent)
	}
	if stats.Storage == nil {
		t.Error("Storage should not be nil")
	}
	if stats.Network.Interfaces == nil {
		t.Error("Network.Interfaces should not be nil")
	}
}

func TestHealth(t *testing.T) {
	s := newTestService(t)

	h := s.Health()
	m, ok := h.(map[string]interface{})
	if !ok {
		t.Fatalf("Health() returned %T, want map[string]interface{}", h)
	}
	if m["status"] != "ok" {
		t.Errorf("Health() status = %v, want \"ok\"", m["status"])
	}
	if _, ok := m["uptime_ms"]; !ok {
		t.Error("Health() missing uptime_ms")
	}
	uptime, ok := m["uptime_ms"].(int64)
	if !ok {
		t.Fatalf("Health() uptime_ms type = %T, want int64", m["uptime_ms"])
	}
	if uptime < 0 {
		t.Errorf("Health() uptime_ms = %d, want >= 0", uptime)
	}
}

func TestName(t *testing.T) {
	s := newTestService(t)
	if s.Name() != "device" {
		t.Errorf("Name() = %q, want \"device\"", s.Name())
	}
}

func TestStartStop(t *testing.T) {
	s := newTestService(t)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}
