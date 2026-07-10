package bpp

import (
	"context"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/mdns"
)

func TestDiscoveryInfo_ToTXTRecords(t *testing.T) {
	t.Parallel()

	t.Run("all fields populated", func(t *testing.T) {
		info := &DiscoveryInfo{
			DeviceID:     "dev_abc123",
			FriendlyName: "kitchen-pi",
			Version:      "1.0.0",
			Platform:     "linux/arm64",
			Capabilities: []string{"terminal", "file", "system"},
		}
		records := info.ToTXTRecords()
		expected := []string{
			"device_id=dev_abc123",
			"friendly_name=kitchen-pi",
			"version=1.0.0",
			"platform=linux/arm64",
			"caps=terminal,file,system",
		}
		if len(records) != len(expected) {
			t.Fatalf("got %d records, want %d", len(records), len(expected))
		}
		sort.Strings(records)
		sort.Strings(expected)
		for i := range records {
			if records[i] != expected[i] {
				t.Errorf("record[%d] = %q, want %q", i, records[i], expected[i])
			}
		}
	})

	t.Run("no capabilities", func(t *testing.T) {
		info := &DiscoveryInfo{
			DeviceID:     "dev_abc",
			FriendlyName: "test",
			Version:      "0.1.0",
			Platform:     "linux/amd64",
		}
		records := info.ToTXTRecords()
		for _, r := range records {
			if len(r) > 0 && r[:4] == "caps" {
				t.Errorf("unexpected caps record: %q", r)
			}
		}
	})

	t.Run("empty fields", func(t *testing.T) {
		info := &DiscoveryInfo{}
		records := info.ToTXTRecords()
		if len(records) != 4 {
			t.Fatalf("got %d records, want 4", len(records))
		}
	})
}

func TestDiscoveredDevice_String(t *testing.T) {
	t.Parallel()

	dev := &DiscoveredDevice{
		DeviceID:     "dev_42",
		FriendlyName: "living-room",
		Platform:     "linux/arm64",
		Addr:         net.ParseIP("192.168.1.42"),
		Port:         8080,
	}
	s := dev.String()
	if s == "" {
		t.Fatal("String() returned empty")
	}
	if len(s) < 10 {
		t.Fatalf("String() too short: %q", s)
	}
}

func TestParseServiceEntry(t *testing.T) {
	t.Parallel()

	t.Run("nil entry", func(t *testing.T) {
		if got := parseServiceEntry(nil); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("valid entry with device_id", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name:   "test-pi",
			Addr:   net.ParseIP("10.0.0.5"),
			Port:   8080,
			InfoFields: []string{
				"device_id=dev_555",
				"friendly_name=test-pi",
				"version=2.0.0",
				"platform=linux/arm",
				"caps=terminal,file",
			},
		}
		dev := parseServiceEntry(entry)
		if dev == nil {
			t.Fatal("expected non-nil device")
		}
		if dev.DeviceID != "dev_555" {
			t.Errorf("DeviceID = %q, want %q", dev.DeviceID, "dev_555")
		}
		if dev.FriendlyName != "test-pi" {
			t.Errorf("FriendlyName = %q, want %q", dev.FriendlyName, "test-pi")
		}
		if dev.Version != "2.0.0" {
			t.Errorf("Version = %q, want %q", dev.Version, "2.0.0")
		}
		if dev.Platform != "linux/arm" {
			t.Errorf("Platform = %q, want %q", dev.Platform, "linux/arm")
		}
		if len(dev.Capabilities) != 2 || dev.Capabilities[0] != "terminal" {
			t.Errorf("Capabilities = %v, want [terminal file]", dev.Capabilities)
		}
		if dev.Addr.String() != "10.0.0.5" {
			t.Errorf("Addr = %v, want 10.0.0.5", dev.Addr)
		}
		if dev.Port != 8080 {
			t.Errorf("Port = %d, want 8080", dev.Port)
		}
	})

	t.Run("entry without device_id", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name:       "other-service",
			Addr:       net.ParseIP("10.0.0.99"),
			Port:       1234,
			InfoFields: []string{"version=1.0"},
		}
		if got := parseServiceEntry(entry); got != nil {
			t.Errorf("expected nil for non-BuzzPi entry, got %v", got)
		}
	})

	t.Run("uses AddrV4 when Addr is nil", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name:   "v4-only",
			Addr:   nil,
			AddrV4: net.ParseIP("10.0.0.7"),
			Port:   8080,
			InfoFields: []string{
				"device_id=dev_v4",
			},
		}
		dev := parseServiceEntry(entry)
		if dev == nil {
			t.Fatal("expected non-nil device")
		}
		if dev.Addr.String() != "10.0.0.7" {
			t.Errorf("Addr = %v, want 10.0.0.7", dev.Addr)
		}
	})

	t.Run("skips malformed TXT records", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name: "bad-txt",
			Addr: net.ParseIP("10.0.0.8"),
			Port: 8080,
			InfoFields: []string{
				"no-equals-here",
				"device_id=dev_malformed",
				"=bare_value",
			},
		}
		dev := parseServiceEntry(entry)
		if dev == nil || dev.DeviceID != "dev_malformed" {
			t.Fatalf("expected device dev_malformed, got %v", dev)
		}
	})
}

func TestDeduplicateDevices(t *testing.T) {
	t.Parallel()

	t.Run("removes duplicates", func(t *testing.T) {
		devices := []DiscoveredDevice{
			{DeviceID: "dev_a", FriendlyName: "first"},
			{DeviceID: "dev_b", FriendlyName: "second"},
			{DeviceID: "dev_a", FriendlyName: "duplicate"},
		}
		result := deduplicateDevices(devices)
		if len(result) != 2 {
			t.Fatalf("got %d devices, want 2", len(result))
		}
		// Should keep the first occurrence.
		if result[0].FriendlyName != "first" {
			t.Errorf("kept %q instead of %q", result[0].FriendlyName, "first")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := deduplicateDevices(nil)
		if result == nil || len(result) != 0 {
			t.Errorf("expected empty slice, got %v", result)
		}
	})

	t.Run("all unique", func(t *testing.T) {
		devices := []DiscoveredDevice{
			{DeviceID: "dev_1"},
			{DeviceID: "dev_2"},
			{DeviceID: "dev_3"},
		}
		result := deduplicateDevices(devices)
		if len(result) != 3 {
			t.Errorf("got %d devices, want 3", len(result))
		}
	})
}

func TestNewAdvertiser(t *testing.T) {
	t.Parallel()

	info := &DiscoveryInfo{
		DeviceID: "dev_test", FriendlyName: "test", Version: "1.0", Platform: "linux",
	}
	a := NewAdvertiser(info)
	if a == nil {
		t.Fatal("NewAdvertiser returned nil")
	}
	if a.info.DeviceID != "dev_test" {
		t.Errorf("DeviceID = %q, want %q", a.info.DeviceID, "dev_test")
	}
}

func TestNewBrowser(t *testing.T) {
	t.Parallel()

	b := NewBrowser()
	if b == nil {
		t.Fatal("NewBrowser returned nil")
	}
}

func TestNewDiscoveryService(t *testing.T) {
	t.Parallel()

	info := &DiscoveryInfo{
		DeviceID: "dev_ds", FriendlyName: "ds", Version: "1.0", Platform: "linux",
	}
	ds := NewDiscoveryService(info)
	if ds == nil {
		t.Fatal("NewDiscoveryService returned nil")
	}
	if ds.Advertise() == nil {
		t.Error("Advertise() returned nil")
	}
	if ds.Browse() == nil {
		t.Error("Browse() returned nil")
	}
}

func TestDiscoveryService_CachedDevices(t *testing.T) {
	t.Parallel()

	info := &DiscoveryInfo{
		DeviceID: "dev_ds", FriendlyName: "ds", Version: "1.0", Platform: "linux",
	}
	ds := NewDiscoveryService(info)

	// Initially empty.
	cached := ds.CachedDevices()
	if cached == nil || len(cached) != 0 {
		t.Errorf("expected empty cached devices, got %v", cached)
	}

	// LookupDevice on empty cache.
	_, ok := ds.LookupDevice("nonexistent")
	if ok {
		t.Error("LookupDevice should return false for missing device")
	}
}

func TestDiscoveryService_StartStop(t *testing.T) {
	// Start/Stop require real mDNS — just verify they don't panic on nil server.
	t.Parallel()

	info := &DiscoveryInfo{
		DeviceID: "dev_ds2", FriendlyName: "ds2", Version: "1.0", Platform: "linux",
	}
	ds := NewDiscoveryService(info)

	// Stop should be safe even before Start.
	if err := ds.Stop(); err != nil {
		t.Logf("Stop before Start returned: %v", err)
	}
}

func TestDiscoverOnce_SkipsNetworkCallInShortMode(t *testing.T) {
	// This test would make a real mDNS call — skip in short mode.
	if testing.Short() {
		t.Skip("skipping network-dependent test in short mode")
	}
	t.Parallel()

	b := NewBrowser()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	devices, err := b.DiscoverOnce(ctx, "")
	if err != nil {
		t.Logf("DiscoverOnce returned error (expected on non-mDNS networks): %v", err)
	}
	if devices == nil {
		t.Log("DiscoverOnce returned nil devices (expected without mDNS)")
	}
}
