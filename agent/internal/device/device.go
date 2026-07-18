// Package device implements the device.* BPP method handlers.
// It provides device identity information and real-time system statistics.
package device

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/buzzpi/agent/internal/config"
	"github.com/buzzpi/agent/internal/version"
)

// Service provides device information and system statistics.
type Service struct {
	cfg        *config.Config
	configPath string
	deviceID   string
	log        *slog.Logger
	startTime  time.Time
	hostname   string
	cpuMu      sync.Mutex
	prevCPU    cpuTimes
}

// InfoResponse is the response for device.info.
type InfoResponse struct {
	DeviceID       string   `json:"device_id"`
	FriendlyName   string   `json:"friendly_name"`
	Model          string   `json:"model,omitempty"`
	RuntimeVersion string   `json:"runtime_version"`
	BPPVersion     int      `json:"bpp_version"`
	UptimeSeconds  int64    `json:"uptime_seconds"`
	Capabilities   []string `json:"capabilities"`
	Platform       string   `json:"platform"`
}

// StatsResponse is the response for device.stats.
type StatsResponse struct {
	CPU     CPUStats     `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Storage []DiskStats  `json:"storage"`
	Network NetworkStats `json:"network"`
	Uptime  int64        `json:"uptime_seconds"`
}

// CPUStats contains CPU usage information.
type CPUStats struct {
	UsagePercent float64 `json:"usage_percent"`
	TemperatureC float64 `json:"temperature_celsius"`
	FrequencyMHz int     `json:"frequency_mhz,omitempty"`
}

// MemoryStats contains memory usage information.
type MemoryStats struct {
	TotalMB     int64   `json:"total_mb"`
	UsedMB      int64   `json:"used_mb"`
	AvailableMB int64   `json:"available_mb"`
	Percent     float64 `json:"percent"`
}

// DiskStats contains storage usage for a mount point.
type DiskStats struct {
	Mount       string  `json:"mount"`
	TotalMB     int64   `json:"total_mb"`
	UsedMB      int64   `json:"used_mb"`
	AvailableMB int64   `json:"available_mb"`
	Percent     float64 `json:"percent"`
}

// NetworkStats contains network interface statistics.
type NetworkStats struct {
	Interfaces []InterfaceStats `json:"interfaces"`
}

// InterfaceStats contains stats for a single network interface.
type InterfaceStats struct {
	Name    string `json:"name"`
	RxBytes int64  `json:"rx_bytes"`
	TxBytes int64  `json:"tx_bytes"`
}

// NewService creates a new Device service.
func NewService(cfg *config.Config, configPath string, deviceID string, log *slog.Logger) (*Service, error) {
	if log == nil {
		log = slog.Default()
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	return &Service{
		cfg:        cfg,
		configPath: configPath,
		deviceID:   deviceID,
		log:        log.With("component", "device"),
		startTime:  time.Now(),
		hostname:   hostname,
	}, nil
}

// HandleInfo returns device identity information.
func (s *Service) HandleInfo(ctx context.Context, params json.RawMessage) (interface{}, error) {
	bppVer, _ := strconv.Atoi(version.BPPVersion)
	return &InfoResponse{
		DeviceID:       s.deviceID,
		FriendlyName:   s.friendlyName(),
		Model:          "raspberry-pi/5",
		RuntimeVersion: version.Version,
		BPPVersion:     bppVer,
		UptimeSeconds:  int64(time.Since(s.startTime).Seconds()),
		Capabilities:   []string{},
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}, nil
}

// HandleStats returns real-time device system statistics.
func (s *Service) HandleStats(ctx context.Context, params json.RawMessage) (interface{}, error) {
	cpuUsage, cpuTemp, cpuFreq := s.readCPUStats()
	memTotal, memUsed, memAvail, memPct := readMemoryStats()
	disks := readDiskStats()
	netIfaces := readNetworkStats()

	return &StatsResponse{
		CPU: CPUStats{
			UsagePercent: cpuUsage,
			TemperatureC: cpuTemp,
			FrequencyMHz: cpuFreq,
		},
		Memory: MemoryStats{
			TotalMB:     memTotal,
			UsedMB:      memUsed,
			AvailableMB: memAvail,
			Percent:     memPct,
		},
		Storage: disks,
		Network: NetworkStats{
			Interfaces: netIfaces,
		},
		Uptime: int64(time.Since(s.startTime).Seconds()),
	}, nil
}

func (s *Service) friendlyName() string {
	if s.cfg.Runtime.DeviceName != "" {
		return s.cfg.Runtime.DeviceName
	}
	return s.hostname
}

// Name returns the component name (for the Supervisor).
func (s *Service) Name() string { return "device" }

// HandleRename updates the device's friendly name and persists the change.
func (s *Service) HandleRename(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if len(req.Name) > 64 {
		return nil, fmt.Errorf("name must be 64 characters or fewer")
	}

	s.cfg.Runtime.DeviceName = req.Name

	if s.configPath != "" {
		if err := config.Save(s.cfg, s.configPath); err != nil {
			s.log.Error("failed to persist device name", "error", err)
			return nil, fmt.Errorf("save config: %w", err)
		}
	}

	s.log.Info("device renamed", "name", req.Name)
	return map[string]string{"name": req.Name}, nil
}

// Start initializes the device service.
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("device service started")
	return nil
}

// Stop shuts down the device service.
func (s *Service) Stop(ctx context.Context) error {
	s.log.Info("device service stopped")
	return nil
}

// Health returns the service health.
func (s *Service) Health() interface{} {
	return map[string]interface{}{
		"status":    "ok",
		"uptime_ms": time.Since(s.startTime).Milliseconds(),
	}
}
