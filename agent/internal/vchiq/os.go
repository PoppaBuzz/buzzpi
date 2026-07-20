package vchiq

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetLoadAverage returns the 1-minute load average.
func GetLoadAverage() (string, error) {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return "", err
	}
	uptimeResult := string(out)
	loadIndex := strings.Index(uptimeResult, "load average:")
	if loadIndex == -1 {
		return "", errors.New("load average not found")
	}
	loadValue := uptimeResult[loadIndex+len("load average:"):]
	load := strings.TrimSpace(strings.Split(loadValue, ",")[0])
	return load, nil
}

// GetHostname returns the system hostname.
func GetHostname() (string, error) {
	host, err := exec.Command("hostname").Output()
	if err == nil {
		return clean(string(host)), nil
	}
	hostFile, err := exec.Command("cat", "/etc/hostname").Output()
	if err != nil {
		return "", errors.New("couldn't get hostname from both 'hostname' command and /etc/hostname file")
	}
	return clean(string(hostFile)), nil
}

// GetUptime returns system uptime as a duration string.
func GetUptime() (string, error) {
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "", err
	}
	uptimeStr := strings.Split(string(uptimeData), " ")[0]
	uptimeSec, err := strconv.ParseFloat(uptimeStr, 64)
	if err != nil {
		return "", err
	}
	uptimeDur := time.Duration(int64(uptimeSec)) * time.Second
	return uptimeDur.String(), nil
}

// GetKernelVersion returns the kernel version from /proc/version.
func GetKernelVersion() (string, error) {
	version, err := os.ReadFile("/proc/version")
	if err != nil {
		return "", err
	}
	parts := strings.Fields(string(version))
	if len(parts) < 3 {
		return "", errors.New("unexpected /proc/version format")
	}
	return parts[2], nil
}

// GetOSName returns the OS name from /etc/os-release.
func GetOSName() (string, error) {
	osRelease, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	osReleaseStr := string(osRelease)
	lines := strings.Split(osReleaseStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				return strings.Trim(parts[1], "\""), nil
			}
		}
	}
	return "", errors.New("couldn't find OS name in /etc/os-release")
}

// GetFQDN returns the fully qualified domain name.
func GetFQDN() (string, error) {
	fqdn, err := exec.Command("hostname", "-f").Output()
	if err == nil {
		return clean(string(fqdn)), nil
	}
	fqdnFile, err := exec.Command("cat", "/etc/hostname").Output()
	if err != nil {
		return "", errors.New("couldn't get FQDN from both 'hostname -f' command and /etc/hostname file")
	}
	return clean(string(fqdnFile)), nil
}

// GetIPs returns the first non-loopback IPv4 address.
func GetIPs() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		return []net.IP{ipNet.IP}, nil
	}
	return nil, errors.New("couldn't get local IP")
}

// GetNetworkStats returns per-interface network statistics from /sys/class/net.
func GetNetworkStats() ([]NetStatistic, error) {
	files, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return nil, errors.New("permission denied")
	}

	var nets []NetStatistic
	for _, file := range files {
		if file.Type()&os.ModeSymlink == 0 {
			continue
		}
		rxBytes, err := os.ReadFile("/sys/class/net/" + file.Name() + "/statistics/rx_bytes")
		if err != nil {
			return nil, errors.New("permission denied")
		}
		txBytes, err := os.ReadFile("/sys/class/net/" + file.Name() + "/statistics/tx_bytes")
		if err != nil {
			return nil, errors.New("permission denied")
		}
		rx, err := strconv.Atoi(strings.TrimSpace(string(rxBytes)))
		if err != nil {
			return nil, errors.New("error converting rx_bytes to int")
		}
		tx, err := strconv.Atoi(strings.TrimSpace(string(txBytes)))
		if err != nil {
			return nil, errors.New("error converting tx_bytes to int")
		}
		mac, err := os.ReadFile("/sys/class/net/" + file.Name() + "/address")
		if err != nil {
			return nil, errors.New("permission denied")
		}
		nets = append(nets, NetStatistic{
			Interface: file.Name(),
			RxBytes:   rx,
			TxBytes:   tx,
			Mac:       strings.TrimSpace(string(mac)),
		})
	}
	return nets, nil
}

type NetStatistic struct {
	Interface string `json:"interface"`
	Mac       string `json:"mac"`
	RxBytes   int    `json:"rx_bytes"`
	TxBytes   int    `json:"tx_bytes"`
}