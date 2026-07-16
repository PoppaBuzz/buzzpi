// Package mdns provides mDNS service discovery for BuzzPi devices.
// Devices advertise as _buzzpi._tcp service instances on the local network,
// enabling zero-configuration discovery by clients.
package mdns

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// ServiceType is the mDNS service type for BuzzPi.
const ServiceType = "_buzzpi._tcp"

// ServiceInfo contains the information advertised via mDNS.
type ServiceInfo struct {
	DeviceID       string
	FriendlyName   string
	RuntimeVersion string
	Platform       string
	Capabilities   []string
	Port           int
}

// Advertiser announces a BuzzPi device on the local network via mDNS.
type Advertiser struct {
	server  *zeroconf.Server
	info    *ServiceInfo
	log     *slog.Logger
	closeCh chan struct{}
}

// NewAdvertiser creates a new mDNS advertiser.
func NewAdvertiser(info *ServiceInfo, log *slog.Logger) *Advertiser {
	if log == nil {
		log = slog.Default()
	}
	return &Advertiser{
		info:    info,
		log:     log.With("component", "mdns"),
		closeCh: make(chan struct{}),
	}
}

// Start begins advertising the device on the local network.
func (a *Advertiser) Start(ctx context.Context) error {
	txtRecords := []string{
		fmt.Sprintf("device_id=%s", a.info.DeviceID),
		fmt.Sprintf("friendly_name=%s", a.info.FriendlyName),
		fmt.Sprintf("runtime_version=%s", a.info.RuntimeVersion),
		fmt.Sprintf("platform=%s", a.info.Platform),
	}

	if len(a.info.Capabilities) > 0 {
		caps := strings.Join(a.info.Capabilities, ",")
		txtRecords = append(txtRecords, fmt.Sprintf("capabilities=%s", caps))
	}

	// Get the primary IP address.
	ip, err := getPrimaryIP()
	if err != nil {
		return fmt.Errorf("get primary IP: %w", err)
	}

	// Get the network interface for the primary IP.
	iface, err := getInterfaceForIP(ip)
	if err != nil {
		return fmt.Errorf("get interface for IP: %w", err)
	}
	a.log.Info("mDNS interface selected", "name", iface.Name, "ip", ip)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "buzzpi"
	}

	server, err := zeroconf.RegisterProxy(
		a.info.FriendlyName,
		ServiceType,
		"local.",
		a.info.Port,
		hostname,
		[]string{ip.String()},
		txtRecords,
		[]net.Interface{*iface},
	)
	if err != nil {
		return fmt.Errorf("register mDNS service: %w", err)
	}

	a.server = server
	a.log.Info("mDNS advertiser started",
		"service", ServiceType,
		"name", a.info.FriendlyName,
		"port", a.info.Port,
		"ip", ip,
	)

	return nil
}

// Stop shuts down the mDNS advertiser and sends a goodbye packet.
func (a *Advertiser) Stop(ctx context.Context) error {
	if a.server != nil {
		a.server.Shutdown()
		a.server = nil
	}
	a.log.Info("mDNS advertiser stopped")
	return nil
}

// Name returns the component name for the Supervisor.
func (a *Advertiser) Name() string { return "mdns" }

func getPrimaryIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP, nil
		}
	}
	return nil, fmt.Errorf("no suitable IPv4 address found")
}

func getInterfaceForIP(targetIP net.IP) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.Equal(targetIP) {
				return &iface, nil
			}
		}
	}
	return nil, fmt.Errorf("no interface found for IP %s", targetIP)
}

// Health returns the advertiser's health status.
func (a *Advertiser) Health() interface{} {
	if a.server != nil {
		return "ok"
	}
	return "stopped"
}

// Browser discovers BuzzPi devices on the local network.
type Browser struct {
	log *slog.Logger
}

// DiscoveredDevice represents a device found via mDNS.
type DiscoveredDevice struct {
	DeviceID       string
	FriendlyName   string
	Addr           net.IP
	Port           int
	RuntimeVersion string
	Platform       string
	Capabilities   []string
}

// NewBrowser creates a new mDNS browser.
func NewBrowser(log *slog.Logger) *Browser {
	if log == nil {
		log = slog.Default()
	}
	return &Browser{
		log: log.With("component", "mdns-browser"),
	}
}

// Discover scans the local network for BuzzPi devices.
// It blocks for the specified duration and returns all discovered devices.
func (b *Browser) Discover(ctx context.Context, timeout time.Duration) ([]DiscoveredDevice, error) {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("create mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry, 16)
	var devices []DiscoveredDevice

	done := make(chan struct{})
	go func() {
		defer close(done)
		for entry := range entries {
			dev := parseEntry(entry)
			if dev != nil {
				devices = append(devices, *dev)
			}
		}
	}()

	// Start browsing
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := resolver.Browse(ctx, ServiceType, "local.", entries); err != nil {
		return nil, fmt.Errorf("browse mDNS: %w", err)
	}

	<-done
	return devices, nil
}

func parseEntry(entry *zeroconf.ServiceEntry) *DiscoveredDevice {
	if entry == nil {
		return nil
	}

	dev := &DiscoveredDevice{
		FriendlyName: entry.Instance,
		Port:         entry.Port,
	}

	// Get IP address
	if len(entry.AddrIPv4) > 0 {
		dev.Addr = entry.AddrIPv4[0]
	} else if len(entry.AddrIPv6) > 0 {
		dev.Addr = entry.AddrIPv6[0]
	} else {
		return nil
	}

	// Parse TXT records
	for _, txt := range entry.Text {
		if i := strings.IndexByte(txt, '='); i > 0 {
			key := txt[:i]
			val := txt[i+1:]
			switch key {
			case "device_id":
				dev.DeviceID = val
			case "friendly_name":
				dev.FriendlyName = val
			case "runtime_version":
				dev.RuntimeVersion = val
			case "platform":
				dev.Platform = val
			case "capabilities":
				dev.Capabilities = strings.Split(val, ",")
			}
		}
	}

	if dev.DeviceID == "" {
		return nil // not a valid BuzzPi device
	}
	return dev
}
