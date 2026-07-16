package mdns

import (
	"log/slog"
	"net"
	"testing"

	"github.com/grandcat/zeroconf"
)

func TestServiceType(t *testing.T) {
	if ServiceType != "_buzzpi._tcp" {
		t.Errorf("ServiceType = %q, want \"_buzzpi._tcp\"", ServiceType)
	}
}

func TestNewAdvertiser(t *testing.T) {
	info := &ServiceInfo{
		DeviceID:       "dev_abc",
		FriendlyName:   "test-pi",
		RuntimeVersion: "v0.1.0",
		Platform:       "linux/arm64",
		Port:           8080,
	}
	a := NewAdvertiser(info, slog.Default())
	if a == nil {
		t.Fatal("NewAdvertiser() returned nil")
	}
	if a.Name() != "mdns" {
		t.Errorf("Name() = %q, want \"mdns\"", a.Name())
	}
}

func TestNewAdvertiserNilLogger(t *testing.T) {
	info := &ServiceInfo{DeviceID: "dev_test"}
	a := NewAdvertiser(info, nil)
	if a == nil {
		t.Fatal("NewAdvertiser(nil logger) returned nil")
	}
}

func TestAdvertiserHealthStopped(t *testing.T) {
	info := &ServiceInfo{DeviceID: "dev_test"}
	a := NewAdvertiser(info, nil)
	h := a.Health()
	if h != "stopped" {
		t.Errorf("Health() before Start = %v, want \"stopped\"", h)
	}
}

func TestAdvertiserStopWithoutStart(t *testing.T) {
	info := &ServiceInfo{DeviceID: "dev_test"}
	a := NewAdvertiser(info, nil)
	if err := a.Stop(nil); err != nil {
		t.Fatalf("Stop() on unstarted advertiser failed: %v", err)
	}
}

func TestAdvertiserDoubleStop(t *testing.T) {
	info := &ServiceInfo{
		DeviceID:     "dev_double",
		FriendlyName: "test-double",
		Port:         8080,
	}
	a := NewAdvertiser(info, slog.Default())
	if err := a.Stop(nil); err != nil {
		t.Fatalf("second Stop() failed: %v", err)
	}
}

func TestNewBrowser(t *testing.T) {
	b := NewBrowser(slog.Default())
	if b == nil {
		t.Fatal("NewBrowser() returned nil")
	}
}

func TestNewBrowserNilLogger(t *testing.T) {
	b := NewBrowser(nil)
	if b == nil {
		t.Fatal("NewBrowser(nil logger) returned nil")
	}
}

func TestServiceInfoFields(t *testing.T) {
	info := &ServiceInfo{
		DeviceID:       "dev_fields",
		FriendlyName:   "Fields Test",
		RuntimeVersion: "v0.2.0",
		Platform:       "linux/amd64",
		Port:           9090,
		Capabilities:   []string{"terminal", "files"},
	}
	a := NewAdvertiser(info, slog.Default())
	if a.Name() != "mdns" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestParseEntryNil(t *testing.T) {
	dev := parseEntry(nil)
	if dev != nil {
		t.Error("parseEntry(nil) should return nil")
	}
}

func TestParseEntryValid(t *testing.T) {
	entry := zeroconf.NewServiceEntry("test-pi", ServiceType, "local.")
	entry.Port = 8080
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.100")}
	entry.Text = []string{
		"device_id=dev_abc123",
		"friendly_name=test-pi",
		"runtime_version=v0.1.0",
		"platform=linux/arm64",
	}
	dev := parseEntry(entry)
	if dev == nil {
		t.Fatal("parseEntry(valid) returned nil")
	}
	if dev.DeviceID != "dev_abc123" {
		t.Errorf("DeviceID = %q, want \"dev_abc123\"", dev.DeviceID)
	}
	if dev.FriendlyName != "test-pi" {
		t.Errorf("FriendlyName = %q, want \"test-pi\"", dev.FriendlyName)
	}
	if dev.RuntimeVersion != "v0.1.0" {
		t.Errorf("RuntimeVersion = %q, want \"v0.1.0\"", dev.RuntimeVersion)
	}
	if dev.Platform != "linux/arm64" {
		t.Errorf("Platform = %q, want \"linux/arm64\"", dev.Platform)
	}
	if dev.Port != 8080 {
		t.Errorf("Port = %d, want 8080", dev.Port)
	}
}

func TestParseEntryNoDeviceID(t *testing.T) {
	entry := zeroconf.NewServiceEntry("no-id", ServiceType, "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.101")}
	entry.Text = []string{
		"friendly_name=no-id",
	}
	dev := parseEntry(entry)
	if dev != nil {
		t.Error("parseEntry(no device_id) should return nil")
	}
}

func TestParseEntryEmptyText(t *testing.T) {
	entry := zeroconf.NewServiceEntry("empty-fields", ServiceType, "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.102")}
	entry.Text = []string{}
	dev := parseEntry(entry)
	// Empty text but no device_id means nil
	if dev != nil {
		t.Error("parseEntry(empty Text) should return nil")
	}
}

func TestParseEntryNilText(t *testing.T) {
	entry := zeroconf.NewServiceEntry("nil-fields", ServiceType, "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.103")}
	entry.Text = nil
	dev := parseEntry(entry)
	if dev != nil {
		t.Error("parseEntry(nil Text) should return nil")
	}
}

func TestParseEntryMalformed(t *testing.T) {
	entry := zeroconf.NewServiceEntry("weird", ServiceType, "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.104")}
	entry.Text = []string{
		"device_id=dev_xyz",
		"nope",
		"=badval",
		"",
	}
	dev := parseEntry(entry)
	if dev == nil {
		t.Fatal("parseEntry(malformed) returned nil")
	}
	if dev.DeviceID != "dev_xyz" {
		t.Errorf("DeviceID = %q, want \"dev_xyz\"", dev.DeviceID)
	}
}

func TestParseEntrySingleEquals(t *testing.T) {
	entry := zeroconf.NewServiceEntry("eq-only", ServiceType, "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.105")}
	entry.Text = []string{"="}
	dev := parseEntry(entry)
	if dev != nil {
		t.Error("parseEntry(only equals) should return nil")
	}
}

func TestParseEntrySingleChar(t *testing.T) {
	entry := zeroconf.NewServiceEntry("single-char", ServiceType, "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.106")}
	entry.Text = []string{"a"}
	dev := parseEntry(entry)
	if dev != nil {
		t.Error("parseEntry(single char) should return nil")
	}
}

func TestParseEntryFriendlyNameFromInstance(t *testing.T) {
	entry := zeroconf.NewServiceEntry("my-pi-name", ServiceType, "local.")
	entry.Port = 9090
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.107")}
	entry.Text = []string{
		"device_id=dev_from_name",
	}
	dev := parseEntry(entry)
	if dev == nil {
		t.Fatal("parseEntry returned nil")
	}
	if dev.FriendlyName != "my-pi-name" {
		t.Errorf("FriendlyName = %q, want \"my-pi-name\"", dev.FriendlyName)
	}
}

func TestParseEntryValueWithEquals(t *testing.T) {
	entry := zeroconf.NewServiceEntry("eq-in-val", ServiceType, "local.")
	entry.Port = 8080
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.108")}
	entry.Text = []string{
		"device_id=dev_with=equals",
	}
	dev := parseEntry(entry)
	if dev == nil {
		t.Fatal("parseEntry returned nil")
	}
	if dev.DeviceID != "dev_with=equals" {
		t.Errorf("DeviceID = %q, want \"dev_with=equals\"", dev.DeviceID)
	}
}

func TestParseEntryNoIP(t *testing.T) {
	entry := zeroconf.NewServiceEntry("no-ip", ServiceType, "local.")
	entry.Port = 8080
	entry.Text = []string{
		"device_id=dev_no_ip",
	}
	dev := parseEntry(entry)
	if dev != nil {
		t.Error("parseEntry(no IP) should return nil")
	}
}

func TestParseEntryIPv6Only(t *testing.T) {
	entry := zeroconf.NewServiceEntry("ipv6-only", ServiceType, "local.")
	entry.Port = 8080
	entry.AddrIPv6 = []net.IP{net.ParseIP("fe80::1")}
	entry.Text = []string{
		"device_id=dev_ipv6",
	}
	dev := parseEntry(entry)
	if dev == nil {
		t.Fatal("parseEntry(IPv6) returned nil")
	}
	if !dev.Addr.Equal(net.ParseIP("fe80::1")) {
		t.Errorf("Addr = %v, want fe80::1", dev.Addr)
	}
}

func TestParseEntryCapabilities(t *testing.T) {
	entry := zeroconf.NewServiceEntry("caps", ServiceType, "local.")
	entry.Port = 8080
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.109")}
	entry.Text = []string{
		"device_id=dev_caps",
		"capabilities=terminal,files,docker",
	}
	dev := parseEntry(entry)
	if dev == nil {
		t.Fatal("parseEntry returned nil")
	}
	if len(dev.Capabilities) != 3 {
		t.Fatalf("Capabilities len = %d, want 3", len(dev.Capabilities))
	}
	if dev.Capabilities[0] != "terminal" || dev.Capabilities[1] != "files" || dev.Capabilities[2] != "docker" {
		t.Errorf("Capabilities = %v", dev.Capabilities)
	}
}
