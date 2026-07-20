package device

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/buzzpi/agent/internal/config"
)

const testDeviceID = "test_device_abc123"

func newTestService(t *testing.T) *Service {
	t.Helper()
	cfg := config.DefaultConfig()
	s, err := NewService(cfg, "", testDeviceID, slog.Default())
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
		s, err := NewService(cfg, "", testDeviceID, nil)
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

	if info.DeviceID != testDeviceID {
		t.Errorf("DeviceID = %q, want %q", info.DeviceID, testDeviceID)
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
		s, err := NewService(cfg, "", testDeviceID, slog.Default())
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

	t.Run("falls back to hostname when empty", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Runtime.DeviceName = ""
		s, err := NewService(cfg, "", testDeviceID, slog.Default())
		if err != nil {
			t.Fatalf("NewService() failed: %v", err)
		}
		resp, err := s.HandleInfo(context.Background(), json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("HandleInfo() failed: %v", err)
		}
		info := resp.(*InfoResponse)
		hostname, _ := hostname()
		if info.FriendlyName != hostname {
			t.Errorf("FriendlyName = %q, want hostname %q", info.FriendlyName, hostname)
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

func TestHandleStatsCallsReadCPUStatsTwice(t *testing.T) {
	s := newTestService(t)
	// First call to populate prevCPU baseline
	s.HandleStats(context.Background(), json.RawMessage(`{}`))
	// Second call should calculate delta
	resp, err := s.HandleStats(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("second HandleStats() failed: %v", err)
	}
	stats := resp.(*StatsResponse)
	// On non-Linux this returns 0; on Linux it may return actual usage
	if stats.CPU.UsagePercent < 0 {
		t.Errorf("CPU.UsagePercent = %f, want >= 0", stats.CPU.UsagePercent)
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

// --- Tests for unexported functions in procfs.go and diskstats.go ---

func TestCalcPercent(t *testing.T) {
	tests := []struct {
		used, total int64
		want        float64
	}{
		{0, 0, 0},
		{0, 100, 0},
		{50, 100, 50.0},
		{1, 3, 33.3},
		{100, 100, 100.0},
		{250, 1000, 25.0},
	}
	for _, tt := range tests {
		got := calcPercent(tt.used, tt.total)
		if got != tt.want {
			t.Errorf("calcPercent(%d, %d) = %f, want %f", tt.used, tt.total, got, tt.want)
		}
	}
}

func TestParseMeminfoValue(t *testing.T) {
	tests := []struct {
		line string
		want int64
	}{
		{"MemTotal:       16384000 kB", 16384000},
		{"MemAvailable:   8192000 kB", 8192000},
		{"MemTotal:", 0},           // no value
		{"", 0},                     // empty
		{"just text", 0},            // no numeric
		{"key:  42 kB", 42},        // extra fields
	}
	for _, tt := range tests {
		got := parseMeminfoValue(tt.line)
		if got != tt.want {
			t.Errorf("parseMeminfoValue(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

func TestReadDiskStatsNonLinux(t *testing.T) {
	if isLinux {
		t.Skip("running on Linux, readDiskStats returns real data")
	}
	disks := readDiskStats()
	if len(disks) != 0 {
		t.Errorf("readDiskStats() on non-Linux returned %d entries, want 0", len(disks))
	}
}

func TestReadNetworkStatsNonLinux(t *testing.T) {
	if isLinux {
		t.Skip("running on Linux")
	}
	ifaces := readNetworkStats()
	if len(ifaces) != 0 {
		t.Errorf("readNetworkStats() on non-Linux returned %d entries, want 0", len(ifaces))
	}
}

func TestReadMemoryStatsNonLinux(t *testing.T) {
	if isLinux {
		t.Skip("running on Linux")
	}
	total, used, avail, pct := readMemoryStats()
	if total != 0 || used != 0 || avail != 0 || pct != 0 {
		t.Errorf("readMemoryStats() on non-Linux = (%d, %d, %d, %f), want all zeros", total, used, avail, pct)
	}
}

func TestReadCPUStatsNonLinux(t *testing.T) {
	if isLinux {
		t.Skip("running on Linux")
	}
	s := newTestService(t)
	usage, temp, freq := s.readCPUStats()
	if usage != 0 || temp != 0 || freq != 0 {
		t.Errorf("readCPUStats() on non-Linux = (%f, %f, %d), want all zeros", usage, temp, freq)
	}
}

func TestReadCPUTemperatureNonLinux(t *testing.T) {
	if isLinux {
		t.Skip("running on Linux")
	}
	temp := readCPUTemperature()
	if temp != 0 {
		t.Errorf("readCPUTemperature() on non-Linux = %f, want 0", temp)
	}
}

func TestReadCPUFrequencyNonLinux(t *testing.T) {
	if isLinux {
		t.Skip("running on Linux")
	}
	freq := readCPUFrequency()
	if freq != 0 {
		t.Errorf("readCPUFrequency() on non-Linux = %d, want 0", freq)
	}
}

// hostname helper to match device.go's use of os.Hostname()
func hostname() (string, error) {
	return os_hostname()
}

// Expose os.Hostname for tests
var os_hostname = os.Hostname

func TestServiceFriendlyNameFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Runtime.DeviceName = ""
	s, err := NewService(cfg, "", "dev_test", slog.Default())
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}
	name := s.friendlyName()
	if name == "" {
		t.Error("friendlyName() should return hostname when DeviceName is empty")
	}
}

func TestServiceFriendlyNameFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Runtime.DeviceName = "CustomName"
	s, err := NewService(cfg, "", "dev_test", slog.Default())
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}
	if s.friendlyName() != "CustomName" {
		t.Errorf("friendlyName() = %q, want \"CustomName\"", s.friendlyName())
	}
}
