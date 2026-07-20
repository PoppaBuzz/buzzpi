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
	"github.com/buzzpi/agent/internal/vchiq"
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

// CPUInfoResponse is the response for device.cpu.
type CPUInfoResponse struct {
	Model      string          `json:"model"`
	Cores      string          `json:"cores"`
	MHz        string          `json:"mhz"`
	Temperature string         `json:"temperature"`
	Frequency  float64         `json:"frequency_mhz"`
	UsagePercent float64       `json:"usage_percent"`
	Serial     string          `json:"serial,omitempty"`
	Revision   string          `json:"revision,omitempty"`
	Info       []vchiq.LscpuField `json:"info,omitempty"`
}

func (s *Service) HandleCPU(ctx context.Context, params json.RawMessage) (interface{}, error) {
	cpuUsage, _, _ := s.readCPUStats()

	model, _ := vchiq.GetCPUModel()
	cores, _ := vchiq.GetCPUCores()
	mhz, _ := vchiq.GetCPUMHz()
	temp, _ := vchiq.GetCPUTemperature()
	freq, _ := vchiq.GetCPUFrequency()
	serial, _ := vchiq.GetCPUSerial()
	revision, _ := vchiq.GetCPURevision()
	info, _ := vchiq.GetCPUInfo()

	return &CPUInfoResponse{
		Model:        model,
		Cores:        cores,
		MHz:          mhz,
		Temperature:  temp,
		Frequency:    freq,
		UsagePercent: cpuUsage,
		Serial:       serial,
		Revision:     revision,
		Info:         info,
	}, nil
}

// ThrottlingResponse is the response for device.throttling.
type ThrottlingResponse struct {
	Throttled    int64  `json:"throttled"`
	ThrottledInfo string `json:"throttled_info"`
}

func (s *Service) HandleThrottling(ctx context.Context, params json.RawMessage) (interface{}, error) {
	throttled, err := vchiq.GetThrottled()
	if err != nil {
		return nil, fmt.Errorf("get throttled: %w", err)
	}
	info, err := vchiq.GetThrottledInfo()
	if err != nil {
		return nil, fmt.Errorf("get throttled info: %w", err)
	}
	return &ThrottlingResponse{
		Throttled:     throttled,
		ThrottledInfo: info,
	}, nil
}

// GPUDetailsResponse is the response for device.gpu.
type GPUDetailsResponse struct {
	GPUTemperature string `json:"gpu_temperature"`
	ARM            string `json:"arm_memory"`
	GPU            string `json:"gpu_memory"`
}

func (s *Service) HandleGPU(ctx context.Context, params json.RawMessage) (interface{}, error) {
	temp, _ := vchiq.GetGPUTemperature()
	arm, gpu, _ := vchiq.GetVCGencmdMemory()
	return &GPUDetailsResponse{
		GPUTemperature: temp,
		ARM:            arm,
		GPU:            gpu,
	}, nil
}

// VoltageResponse is the response for device.voltage.
type VoltageResponse struct {
	CoreVoltage string `json:"core_voltage"`
}

func (s *Service) HandleVoltage(ctx context.Context, params json.RawMessage) (interface{}, error) {
	volt, err := vchiq.GetCoreVoltage()
	if err != nil {
		return nil, fmt.Errorf("get core voltage: %w", err)
	}
	return &VoltageResponse{CoreVoltage: volt}, nil
}

// ProcessResponse is the response for device.processes.
type ProcessResponse struct {
	Processes []vchiq.ProcessInfo `json:"processes"`
}

func (s *Service) HandleProcesses(ctx context.Context, params json.RawMessage) (interface{}, error) {
	procs, err := vchiq.ListProcesses()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}
	return &ProcessResponse{Processes: procs}, nil
}

// USBResponse is the response for device.usb.
type USBResponse struct {
	Devices []vchiq.USBDevice `json:"devices"`
}

func (s *Service) HandleUSB(ctx context.Context, params json.RawMessage) (interface{}, error) {
	devices, err := vchiq.GetUSBList()
	if err != nil {
		return nil, fmt.Errorf("list USB devices: %w", err)
	}
	return &USBResponse{Devices: devices}, nil
}

// OSInfoResponse is the response for device.os.
type OSInfoResponse struct {
	Hostname      string   `json:"hostname"`
	OSName        string   `json:"os_name"`
	KernelVersion string   `json:"kernel_version"`
	Uptime        string   `json:"uptime"`
	LoadAverage   string   `json:"load_average"`
	FQDN          string   `json:"fqdn"`
	IPs           []string `json:"ips"`
}

func (s *Service) HandleOS(ctx context.Context, params json.RawMessage) (interface{}, error) {
	hostname, _ := vchiq.GetHostname()
	osName, _ := vchiq.GetOSName()
	kernel, _ := vchiq.GetKernelVersion()
	uptime, _ := vchiq.GetUptime()
	loadAvg, _ := vchiq.GetLoadAverage()
	fqdn, _ := vchiq.GetFQDN()

	var ips []string
	netIPs, _ := vchiq.GetIPs()
	for _, ip := range netIPs {
		ips = append(ips, ip.String())
	}

	return &OSInfoResponse{
		Hostname:      hostname,
		OSName:        osName,
		KernelVersion: kernel,
		Uptime:        uptime,
		LoadAverage:   loadAvg,
		FQDN:          fqdn,
		IPs:           ips,
	}, nil
}

// ModelResponse is the response for device.model.
type ModelResponse struct {
	DeviceName         string  `json:"device_name"`
	Revision           string  `json:"revision"`
	MinimalPowerSupply float64 `json:"minimal_power_supply_amps"`
	VcgencmdInstalled  bool    `json:"vcgencmd_installed"`
}

func (s *Service) HandleModel(ctx context.Context, params json.RawMessage) (interface{}, error) {
	deviceName, err := vchiq.GetDeviceName()
	if err != nil {
		return nil, fmt.Errorf("get device name: %w", err)
	}
	revision, _ := vchiq.GetCPURevision()
	power, _ := vchiq.GetMinimalPowerSupply(deviceName)
	return &ModelResponse{
		DeviceName:         deviceName,
		Revision:           revision,
		MinimalPowerSupply: power,
		VcgencmdInstalled:  vchiq.IsVcgencmdInstalled(),
	}, nil
}
