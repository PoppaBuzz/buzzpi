// Package bpp implements the BuzzPi Protocol (BPP) wire format.
//
// Discovery provides mDNS-based zero-configuration device discovery
// for local network peer finding. It wraps the hashicorp/mdns library
// with BPP-specific TXT record encoding and service integration.
package bpp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

// mDNS service constants.
const (
	// MDNSServiceType is the mDNS service type for BuzzPi devices.
	MDNSServiceType = "_buzzpi._tcp"
	// MDNSDomain is the mDNS domain.
	MDNSDomain = "local"
)

// DiscoveryInfo encodes the metadata a BuzzPi device advertises via mDNS TXT
// records. Each field maps to a key=value entry in the DNS-SD text record.
//
// TXT record format:
//
//	device_id=<id>
//	friendly_name=<name>
//	version=<semver>
//	platform=<platform>
//	caps=<comma-separated capabilities>
type DiscoveryInfo struct {
	DeviceID     string   `txt:"device_id"`
	FriendlyName string   `txt:"friendly_name"`
	Version      string   `txt:"version"`
	Platform     string   `txt:"platform"`
	Capabilities []string `txt:"caps,comma"`
	Port         int      // The BPP WebSocket port (not a TXT record)
}

// ToTXTRecords encodes DiscoveryInfo into a slice of DNS TXT key=value strings.
func (d *DiscoveryInfo) ToTXTRecords() []string {
	records := []string{
		fmt.Sprintf("device_id=%s", d.DeviceID),
		fmt.Sprintf("friendly_name=%s", d.FriendlyName),
		fmt.Sprintf("version=%s", d.Version),
		fmt.Sprintf("platform=%s", d.Platform),
	}
	if len(d.Capabilities) > 0 {
		records = append(records, fmt.Sprintf("caps=%s", strings.Join(d.Capabilities, ",")))
	}
	return records
}

// DiscoveredDevice represents a BuzzPi device discovered on the local network.
type DiscoveredDevice struct {
	DeviceID     string
	FriendlyName string
	Version      string
	Platform     string
	Capabilities []string
	Addr         net.IP
	Port         int
}

// String returns a human-readable representation of the discovered device.
func (d *DiscoveredDevice) String() string {
	return fmt.Sprintf("%s (%s) @ %s:%d [%s]",
		d.FriendlyName, d.DeviceID, d.Addr, d.Port, d.Platform)
}

// Advertiser publishes a BuzzPi device's presence on the LAN via mDNS
// so that clients and other devices can discover it without configuration.
type Advertiser struct {
	server *mdns.Server
	info   *DiscoveryInfo
	log    *slog.Logger
	mu     sync.Mutex
}

// NewAdvertiser creates a new mDNS Advertiser.
func NewAdvertiser(info *DiscoveryInfo) *Advertiser {
	return &Advertiser{
		info: info,
		log:  slog.Default().With("component", "bpp-mdns-adv", "device", info.DeviceID),
	}
}

// Start begins publishing the device's mDNS service. Blocks until the
// advertisement is registered or an error occurs.
func (a *Advertiser) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		return fmt.Errorf("bpp: advertiser already running")
	}

	service, err := mdns.NewMDNSService(
		a.info.FriendlyName,
		MDNSServiceType+".", // service name with trailing dot for FQDN
		"local.",
		"", // auto-detect hostname
		a.info.Port,
		nil, // auto-detect IPs
		a.info.ToTXTRecords(),
	)
	if err != nil {
		return fmt.Errorf("bpp: create mDNS service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("bpp: start mDNS server: %w", err)
	}

	a.server = server
	a.log.Info("mDNS advertiser started",
		"service", MDNSServiceType,
		"name", a.info.FriendlyName,
		"port", a.info.Port,
	)
	return nil
}

// Stop shuts down the mDNS advertiser and sends a goodbye packet.
func (a *Advertiser) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server == nil {
		return nil
	}

	if err := a.server.Shutdown(); err != nil {
		return fmt.Errorf("bpp: stop mDNS: %w", err)
	}
	a.server = nil
	a.log.Info("mDNS advertiser stopped")
	return nil
}

// IsRunning returns true if the advertiser is currently active.
func (a *Advertiser) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.server != nil
}

// Browser discovers BuzzPi devices on the local network using mDNS.
// It sends an mDNS query for the _buzzpi._tcp service type and collects
// responses for the specified duration.
type Browser struct {
	log *slog.Logger
}

// NewBrowser creates a new mDNS Browser.
func NewBrowser() *Browser {
	return &Browser{
		log: slog.Default().With("component", "bpp-mdns-browse"),
	}
}

// Discover scans the local network for BuzzPi devices.
// It blocks for the specified timeout, collecting responses.
// Returns the list of discovered devices (may be empty).
func (b *Browser) Discover(ctx context.Context, timeout time.Duration) ([]DiscoveredDevice, error) {
	resultsCh := make(chan *mdns.ServiceEntry, 64)
	var mu sync.Mutex
	var devices []DiscoveredDevice

	done := make(chan struct{})
	go func() {
		for entry := range resultsCh {
			dev := parseServiceEntry(entry)
			if dev != nil {
				mu.Lock()
				devices = append(devices, *dev)
				mu.Unlock()
			}
		}
		close(done)
	}()

	// Run the mDNS query.
	mdns.Query(&mdns.QueryParam{
		Service: MDNSServiceType,
		Domain:  MDNSDomain,
		Timeout: timeout,
		Entries: resultsCh,
		// Iface: nil  // all interfaces
	})

	close(resultsCh)
	<-done

	// Deduplicate by device ID.
	mu.Lock()
	result := deduplicateDevices(devices)
	mu.Unlock()

	return result, nil
}

// DiscoverOnce is a convenience wrapper that calls Discover with a default
// timeout and returns the first device matching the given deviceID (or all
// if deviceID is empty).
func (b *Browser) DiscoverOnce(ctx context.Context, deviceID string) ([]DiscoveredDevice, error) {
	devices, err := b.Discover(ctx, 3*time.Second)
	if err != nil {
		return nil, err
	}
	if deviceID == "" {
		return devices, nil
	}
	for _, d := range devices {
		if d.DeviceID == deviceID {
			return []DiscoveredDevice{d}, nil
		}
	}
	return nil, nil
}

// DiscoveryService combines an Advertiser and Browser into a single
// coordinated service that both publishes and discovers BuzzPi devices.
type DiscoveryService struct {
	advertiser *Advertiser
	browser    *Browser
	log        *slog.Logger
	devicesMu  sync.RWMutex
	devices    map[string]DiscoveredDevice // device_id -> device
}

// NewDiscoveryService creates a DiscoveryService.
func NewDiscoveryService(info *DiscoveryInfo) *DiscoveryService {
	return &DiscoveryService{
		advertiser: NewAdvertiser(info),
		browser:    NewBrowser(),
		log:        slog.Default().With("component", "bpp-discovery"),
		devices:    make(map[string]DiscoveredDevice),
	}
}

// Start begins advertising this device on the LAN.
func (ds *DiscoveryService) Start() error {
	return ds.advertiser.Start()
}

// Stop stops advertising.
func (ds *DiscoveryService) Stop() error {
	return ds.advertiser.Stop()
}

// Advertise returns the underlying Advertiser (for health checks, etc.).
func (ds *DiscoveryService) Advertise() *Advertiser {
	return ds.advertiser
}

// Browse returns the underlying Browser.
func (ds *DiscoveryService) Browse() *Browser {
	return ds.browser
}

// Discover performs a single scan and caches the results.
func (ds *DiscoveryService) Discover(ctx context.Context, timeout time.Duration) ([]DiscoveredDevice, error) {
	devices, err := ds.browser.Discover(ctx, timeout)
	if err != nil {
		return nil, err
	}

	ds.devicesMu.Lock()
	ds.devices = make(map[string]DiscoveredDevice)
	for _, d := range devices {
		ds.devices[d.DeviceID] = d
	}
	ds.devicesMu.Unlock()

	return devices, nil
}

// CachedDevices returns the last discovered device list without performing
// a new scan.
func (ds *DiscoveryService) CachedDevices() []DiscoveredDevice {
	ds.devicesMu.RLock()
	defer ds.devicesMu.RUnlock()

	result := make([]DiscoveredDevice, 0, len(ds.devices))
	for _, d := range ds.devices {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FriendlyName < result[j].FriendlyName
	})
	return result
}

// LookupDevice returns a cached device by ID.
func (ds *DiscoveryService) LookupDevice(deviceID string) (DiscoveredDevice, bool) {
	ds.devicesMu.RLock()
	defer ds.devicesMu.RUnlock()
	d, ok := ds.devices[deviceID]
	return d, ok
}

// parseServiceEntry converts an mDNS ServiceEntry into a DiscoveredDevice.
// Returns nil if the entry lacks a device_id TXT record (not a BuzzPi device).
func parseServiceEntry(entry *mdns.ServiceEntry) *DiscoveredDevice {
	if entry == nil {
		return nil
	}

	dev := &DiscoveredDevice{
		FriendlyName: entry.Name,
		Addr:         entry.Addr,
		Port:         entry.Port,
	}

	// If Addr is nil, use the entry's IP (sometimes in AddrV4/AddrV6).
	if dev.Addr == nil && entry.AddrV4 != nil {
		dev.Addr = entry.AddrV4
	}
	if dev.Addr == nil && entry.AddrV6 != nil {
		dev.Addr = entry.AddrV6
	}

	// Parse TXT records.
	for _, txt := range entry.InfoFields {
		before, after, found := strings.Cut(txt, "=")
		if !found {
			continue
		}
		switch before {
		case "device_id":
			dev.DeviceID = after
		case "friendly_name":
			dev.FriendlyName = after
		case "version":
			dev.Version = after
		case "platform":
			dev.Platform = after
		case "caps":
			if after != "" {
				dev.Capabilities = strings.Split(after, ",")
			}
		}
	}

	if dev.DeviceID == "" {
		return nil // Not a BuzzPi device.
	}
	return dev
}

// deduplicateDevices removes devices with duplicate IDs, keeping the first
// occurrence (which is typically the most responsive).
func deduplicateDevices(devices []DiscoveredDevice) []DiscoveredDevice {
	seen := make(map[string]bool)
	result := make([]DiscoveredDevice, 0, len(devices))
	for _, d := range devices {
		if !seen[d.DeviceID] {
			seen[d.DeviceID] = true
			result = append(result, d)
		}
	}
	return result
}
