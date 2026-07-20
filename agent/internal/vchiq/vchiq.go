package vchiq

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	UnderVoltage          int64 = 1 << 0
	FreqCap                     = 1 << 1
	Throttling                  = 1 << 2
	SoftTempLimitActive         = 1 << 3
	UnderVoltageOccurred        = 1 << 16
	FreqCapOccurred             = 1 << 17
	Throttled                   = 1 << 18
	SoftTempLimitOccurred       = 1 << 19
)

var deviceRevisions = map[string]string{
	"900021": "Raspberry Pi Zero",
	"900032": "Raspberry Pi Zero W",
	"900092": "Raspberry Pi Zero",
	"920092": "Raspberry Pi Zero",
	"900093": "Raspberry Pi Zero",
	"920093": "Raspberry Pi Zero",
	"9000c1": "Raspberry Pi Zero W",
	"9020e0": "Raspberry Pi 3 Model A+",
	"9020e1": "Raspberry Pi 3 Model A+",
	"a01040": "Raspberry Pi 2 Model B",
	"a01041": "Raspberry Pi 2 Model B",
	"a21041": "Raspberry Pi 2 Model B",
	"a02082": "Raspberry Pi 3 Model B",
	"a22082": "Raspberry Pi 3 Model B",
	"a32082": "Raspberry Pi 3 Model B",
	"a52082": "Raspberry Pi 3 Model B",
	"a020a0": "Raspberry Pi Compute Module 3",
	"a220a0": "Raspberry Pi Compute Module 3",
	"a020d3": "Raspberry Pi 3 Model B+",
	"a22083": "Raspberry Pi 3 Model B+",
	"a020d4": "Raspberry Pi 3 Model B+",
	"a02042": "Raspberry Pi 2 Model B (with BCM2837)",
	"a22042": "Raspberry Pi 2 Model B (with BCM2837)",
	"a02100": "Raspberry Pi Compute Module 3+",
	"a03111": "Raspberry Pi 4 Model B",
	"b03111": "Raspberry Pi 4 Model B",
	"c03111": "Raspberry Pi 4 Model B",
	"b03112": "Raspberry Pi 4 Model B",
	"c03112": "Raspberry Pi 4 Model B",
	"b03114": "Raspberry Pi 4 Model B",
	"c03114": "Raspberry Pi 4 Model B",
	"b03115": "Raspberry Pi 4 Model B",
	"c03115": "Raspberry Pi 4 Model B",
	"c03130": "Raspberry Pi 400",
	"a03140": "Raspberry Pi Compute Module 4",
	"b03140": "Raspberry Pi Compute Module 4",
	"c03140": "Raspberry Pi Compute Module 4",
	"d03140": "Raspberry Pi Compute Module 4",
	"902120": "Raspberry Pi Zero 2 W",
	"c04170": "Raspberry Pi 5",
	"d04170": "Raspberry Pi 5",
}

var minimalPowerSupply = map[string]float64{
	"Raspberry Pi Zero":    1.2,
	"Raspberry Pi Zero W":  1.2,
	"Raspberry Pi 3 Model A+": 2.5,
	"Raspberry Pi 2 Model B": 1.8,
	"Raspberry Pi 3 Model B": 2.5,
	"Raspberry Pi 3 Model B+": 2.5,
	"Raspberry Pi Compute Module 3": 2.5,
	"Raspberry Pi Compute Module 3+": 2.5,
	"Raspberry Pi 4 Model B": 3.0,
	"Raspberry Pi 400":     3.0,
	"Raspberry Pi Compute Module 4": 3.0,
	"Raspberry Pi Zero 2 W": 1.2,
	"Raspberry Pi 5":       3.0,
}

func GetThrottled() (int64, error) {
	rawThrottled, err := exec.Command("vcgencmd", "get_throttled").Output()
	if err != nil {
		return 0, fmt.Errorf("couldn't run vcgencmd: %w", err)
	}
	throttled, err := strconv.ParseInt(strings.TrimSpace(string(rawThrottled[12:])), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse throttled output: %w", err)
	}
	return throttled, nil
}

func GetThrottledInfo() (string, error) {
	throttled, err := GetThrottled()
	if err != nil {
		return "", err
	}

	var events []string
	if throttled&UnderVoltage != 0 {
		events = append(events, "Under-voltage detected")
	}
	if throttled&FreqCap != 0 {
		events = append(events, "Frequency capped")
	}
	if throttled&Throttling != 0 {
		events = append(events, "Throttling")
	}
	if throttled&SoftTempLimitActive != 0 {
		events = append(events, "Soft temperature limit active")
	}
	if throttled&UnderVoltageOccurred != 0 {
		events = append(events, "Under-voltage occurred")
	}
	if throttled&FreqCapOccurred != 0 {
		events = append(events, "Frequency cap occurred")
	}
	if throttled&Throttled != 0 {
		events = append(events, "Throttling occurred")
	}
	if throttled&SoftTempLimitOccurred != 0 {
		events = append(events, "Soft temperature limit occurred")
	}

	if len(events) == 0 {
		return "No throttling or voltage issues detected", nil
	}
	return strings.Join(events, "; "), nil
}

// GetGPUTemperature returns the GPU temperature via vcgencmd.
func GetGPUTemperature() (string, error) {
	temp, err := exec.Command("vcgencmd", "measure_temp").Output()
	if err != nil {
		return "", errors.New("couldn't run vcgencmd")
	}
	return clean(string(temp), "temp=", "'C"), nil
}

// GetCoreVoltage returns the core voltage via vcgencmd.
func GetCoreVoltage() (string, error) {
	volt, err := exec.Command("vcgencmd", "measure_volts").Output()
	if err != nil {
		return "", errors.New("couldn't run vcgencmd")
	}
	return clean(string(volt), "volt=", "V"), nil
}

// GetVCGencmdMemory returns ARM and GPU memory allocation via vcgencmd.
func GetVCGencmdMemory() (string, string, error) {
	armMem, err := exec.Command("vcgencmd", "get_mem", "arm").Output()
	if err != nil {
		return "", "", errors.New("couldn't run vcgencmd for arm memory")
	}
	gpuMem, err := exec.Command("vcgencmd", "get_mem", "gpu").Output()
	if err != nil {
		return "", "", errors.New("couldn't run vcgencmd for gpu memory")
	}
	return extractMemorySize(string(armMem)), extractMemorySize(string(gpuMem)), nil
}

// IsVcgencmdInstalled returns true if vcgencmd is available on the system.
func IsVcgencmdInstalled() bool {
	_, err := exec.LookPath("vcgencmd")
	return err == nil
}

// GetDeviceName returns the human-readable Pi model name from the CPU revision.
func GetDeviceName() (string, error) {
	revision, err := GetCPURevision()
	if err != nil {
		return "", err
	}
	if deviceName, exists := deviceRevisions[revision]; exists {
		return deviceName, nil
	}
	return "", fmt.Errorf("unknown revision: %s", revision)
}

// GetMinimalPowerSupply returns the minimum recommended power supply amperage for a Pi model.
func GetMinimalPowerSupply(deviceName string) (float64, bool) {
	power, ok := minimalPowerSupply[deviceName]
	return power, ok
}