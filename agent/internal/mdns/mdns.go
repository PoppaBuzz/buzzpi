// Package mdns provides mDNS service discovery for BuzzPi devices.
// Devices advertise as _buzzpi._tcp service instances on the local network,
// enabling zero-configuration discovery by clients.
package mdns

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/hashicorp/mdns"
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
	server  *mdns.Server
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
		caps := ""
		for i, c := range a.info.Capabilities {
			if i > 0 {
				caps += ","
			}
			caps += c
		}
		txtRecords = append(txtRecords, fmt.Sprintf("capabilities=%s", caps))
	}

	host, _ := net.InterfaceAddrs()
	_ = host // TODO: get primary IP

	var ips []net.IP
	if ip, err := getPrimaryIP(); err == nil {
		ips = []net.IP{ip}
	}

	service, err := mdns.NewMDNSService(
		a.info.FriendlyName,
		ServiceType+".",
		"local.",
		"",
		a.info.Port,
		ips,
		txtRecords,
	)
	if err != nil {
		return fmt.Errorf("create mDNS service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{
		Zone: service,
	})
	if err != nil {
		return fmt.Errorf("start mDNS server: %w", err)
	}

	a.server = server
	a.log.Info("mDNS advertiser started",
		"service", ServiceType,
		"name", a.info.FriendlyName,
		"port", a.info.Port,
	)

	return nil
}

// Stop shuts down the mDNS advertiser and sends a goodbye packet.
func (a *Advertiser) Stop(ctx context.Context) error {
	if a.server != nil {
		if err := a.server.Shutdown(); err != nil {
			return fmt.Errorf("stop mDNS: %w", err)
		}
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
	return nil, fmt.Errorf("no suitable IP address found")
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
	resultsCh := make(chan *mdns.ServiceEntry, 16)
	var devices []DiscoveredDevice

	done := make(chan struct{})
	go func() {
		for entry := range resultsCh {
			dev := parseEntry(entry)
			if dev != nil {
				devices = append(devices, *dev)
			}
		}
		close(done)
	}()

	// Start the mDNS query
	mdns.Query(&mdns.QueryParam{
		Service: ServiceType,
		Domain:  "local",
		Timeout: timeout,
		Entries: resultsCh,
	})

	<-done
	return devices, nil
}

func parseEntry(entry *mdns.ServiceEntry) *DiscoveredDevice {
	if entry == nil {
		return nil
	}

	dev := &DiscoveredDevice{
		FriendlyName: entry.Name,
		Addr:         entry.Addr,
		Port:         entry.Port,
	}

	for _, txt := range entry.InfoFields {
		if len(txt) < 2 {
			continue
		}
		// TXT records are key=value
		for i := 0; i < len(txt)-1; i++ {
			if txt[i] == '=' {
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
				}
			}
		}
	}

	if dev.DeviceID == "" {
		return nil // not a valid BuzzPi device
	}
	return dev
}
