package vchiq

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	ErrExecutingLscpu = errors.New("error executing lscpu")
	ErrParsingLscpu   = errors.New("error parsing lscpu output")
)

type LscpuField struct {
	Field    string       `json:"field"`
	Data     string       `json:"data"`
	Children []LscpuField `json:"children,omitempty"`
}

type LscpuData struct {
	Lscpu []LscpuField `json:"lscpu"`
}

// GetCPUInfo executes `lscpu -J` and returns parsed CPU information.
func GetCPUInfo() ([]LscpuField, error) {
	cmd := exec.Command("lscpu", "-J")
	output, err := cmd.Output()
	if err != nil {
		return nil, ErrExecutingLscpu
	}

	var lscpuData LscpuData
	if err := json.Unmarshal(output, &lscpuData); err != nil {
		return nil, ErrParsingLscpu
	}

	return lscpuData.Lscpu, nil
}

// GetCPUFrequency returns the current CPU frequency in MHz using vcgencmd.
func GetCPUFrequency() (float64, error) {
	out, err := exec.Command("vcgencmd", "measure_clock", "arm").Output()
	if err != nil {
		return 0, err
	}
	freqStr := strings.Split(string(out), "=")[1]
	freq, err := strconv.ParseFloat(strings.TrimSpace(freqStr), 64)
	if err != nil {
		return 0, err
	}
	return freq / 1_000_000, nil
}

// GetCPUTemperature returns the CPU temperature in Celsius using vcgencmd.
func GetCPUTemperature() (string, error) {
	temp, err := exec.Command("vcgencmd", "measure_temp").Output()
	if err != nil {
		return "", errors.New("couldn't run vcgencmd")
	}
	return clean(string(temp), "temp=", "'C"), nil
}

// GetCPUSerial returns the CPU serial from /proc/cpuinfo.
func GetCPUSerial() (string, error) {
	return getCPUInfoValue("Serial")
}

// GetCPURevision returns the CPU revision from /proc/cpuinfo.
func GetCPURevision() (string, error) {
	return getCPUInfoValue("Revision")
}

// GetCPUModel returns the CPU model name from /proc/cpuinfo.
func GetCPUModel() (string, error) {
	return getCPUInfoValue("model name")
}

// GetCPUCores returns the number of CPU cores from /proc/cpuinfo.
func GetCPUCores() (string, error) {
	return getCPUInfoValue("cpu cores")
}

// GetCPUMHz returns the CPU MHz from /proc/cpuinfo.
func GetCPUMHz() (string, error) {
	return getCPUInfoValue("cpu MHz")
}

func getCPUInfoValue(prefix string) (string, error) {
	cpuInfo, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", err
	}

	cpuInfoStr := string(cpuInfo)
	lines := strings.Split(cpuInfoStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", errors.New("couldn't find value for prefix in /proc/cpuinfo")
}