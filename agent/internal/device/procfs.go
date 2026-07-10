// Package device provides device.* BPP method handlers.
package device

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var isLinux = runtime.GOOS == "linux"

// cpuTimes holds a snapshot of /proc/stat CPU line values.
type cpuTimes struct {
	user   uint64
	nice   uint64
	system uint64
	idle   uint64
}

// readCPUStats reads CPU usage percentage, temperature (°C), and frequency (MHz).
// CPU usage is computed as delta between consecutive calls.
// On non-Linux platforms returns zeros.
func (s *Service) readCPUStats() (usagePercent, tempCelsius float64, freqMHz int) {
	s.cpuMu.Lock()
	defer s.cpuMu.Unlock()

	if !isLinux {
		return 0, 0, 0
	}

	curr, err := readCPUTimes()
	if err != nil {
		s.log.Debug("read /proc/stat failed", "error", err)
		return 0, 0, 0
	}

	// Calculate delta if we have a baseline
	if s.prevCPU != (cpuTimes{}) {
		deltaBusy := (curr.user - s.prevCPU.user) +
			(curr.nice - s.prevCPU.nice) +
			(curr.system - s.prevCPU.system)
		deltaIdle := curr.idle - s.prevCPU.idle
		deltaTotal := deltaBusy + deltaIdle
		if deltaTotal > 0 {
			usagePercent = math.Round(float64(deltaBusy)/float64(deltaTotal)*1000) / 10
		}
	}
	s.prevCPU = curr

	tempCelsius = readCPUTemperature()
	freqMHz = readCPUFrequency()
	return
}

// readCPUTimes parses the first "cpu " line from /proc/stat.
func readCPUTimes() (cpuTimes, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return cpuTimes{}, fmt.Errorf("unexpected /proc/stat format: %q", line)
		}
		var ct cpuTimes
		ct.user, _ = strconv.ParseUint(fields[1], 10, 64)
		ct.nice, _ = strconv.ParseUint(fields[2], 10, 64)
		ct.system, _ = strconv.ParseUint(fields[3], 10, 64)
		ct.idle, _ = strconv.ParseUint(fields[4], 10, 64)
		return ct, nil
	}
	return cpuTimes{}, sc.Err()
}

// readCPUTemperature reads CPU temperature from the thermal sysfs interface.
// Returns 0 if unavailable.
func readCPUTemperature() float64 {
	paths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if err != nil {
			continue
		}
		// Values are in millidegrees Celsius (e.g. 55000 = 55°C)
		return float64(val) / 1000.0
	}
	return 0
}

// readCPUFrequency reads the current CPU frequency scaling value.
// Returns 0 if unavailable.
func readCPUFrequency() int {
	paths := []string{
		"/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq",
		"/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_cur_freq",
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		val, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		// Value is in kHz, convert to MHz
		return val / 1000
	}
	return 0
}

// readMemoryStats reads /proc/meminfo and returns memory usage in MB.
// On non-Linux returns zeros.
func readMemoryStats() (totalMB, usedMB, availableMB int64, percent float64) {
	if !isLinux {
		return 0, 0, 0, 0
	}

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, 0
	}

	var memTotal, memAvailable int64
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			memTotal = parseMeminfoValue(line)
		case strings.HasPrefix(line, "MemAvailable:"):
			memAvailable = parseMeminfoValue(line)
		}
	}

	if memTotal == 0 {
		return 0, 0, 0, 0
	}

	totalMB = memTotal / 1024
	availableMB = memAvailable / 1024
	usedMB = totalMB - availableMB
	percent = math.Round(float64(usedMB)/float64(totalMB)*1000) / 10
	return
}

// parseMeminfoValue extracts the numeric value (in kB) from a /proc/meminfo line.
func parseMeminfoValue(line string) int64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	val, _ := strconv.ParseInt(fields[1], 10, 64)
	return val
}

// readNetworkStats reads /proc/net/dev and returns per-interface stats.
// Filters out the loopback interface.
// On non-Linux returns empty slice.
func readNetworkStats() []InterfaceStats {
	if !isLinux {
		return []InterfaceStats{}
	}

	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil
	}

	var ifaces []InterfaceStats
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	lineNum := 0
	for sc.Scan() {
		lineNum++
		// Skip header lines (first two)
		if lineNum <= 2 {
			continue
		}

		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		ifaceName := strings.TrimSpace(parts[0])
		// Skip loopback
		if ifaceName == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		rxBytes, _ := strconv.ParseInt(fields[0], 10, 64)
		txBytes, _ := strconv.ParseInt(fields[8], 10, 64)

		ifaces = append(ifaces, InterfaceStats{
			Name:    ifaceName,
			RxBytes: rxBytes,
			TxBytes: txBytes,
		})
	}

	return ifaces
}
