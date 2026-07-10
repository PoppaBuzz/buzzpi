package mdns

import (
	"log/slog"
	"testing"

	"github.com/hashicorp/mdns"
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

func TestHealth(t *testing.T) {
	info := &ServiceInfo{DeviceID: "dev_test"}
	a := NewAdvertiser(info, nil)

	// Before Start, server is nil so health = "stopped"
	h := a.Health()
	if h != "stopped" {
		t.Errorf("Health() before Start = %v, want \"stopped\"", h)
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

func TestParseEntry(t *testing.T) {
	t.Run("nil entry", func(t *testing.T) {
		dev := parseEntry(nil)
		if dev != nil {
			t.Error("parseEntry(nil) should return nil")
		}
	})

	t.Run("valid entry with all fields", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name: "test-pi",
			Addr: nil,
			Port: 8080,
			InfoFields: []string{
				"device_id=dev_abc123",
				"friendly_name=test-pi",
				"runtime_version=v0.1.0",
				"platform=linux/arm64",
			},
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
	})

	t.Run("entry without device_id returns nil", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name: "no-id",
			InfoFields: []string{
				"friendly_name=no-id",
			},
		}
		dev := parseEntry(entry)
		if dev != nil {
			t.Error("parseEntry(no device_id) should return nil")
		}
	})

	t.Run("malformed TXT records", func(t *testing.T) {
		entry := &mdns.ServiceEntry{
			Name: "weird",
			InfoFields: []string{
				"device_id=dev_xyz",
				"nope",     // no '=' separator
				"=badval",  // empty key
				"",         // empty string
			},
		}
		dev := parseEntry(entry)
		if dev == nil {
			t.Fatal("parseEntry(malformed) returned nil")
		}
		if dev.DeviceID != "dev_xyz" {
			t.Errorf("DeviceID = %q, want \"dev_xyz\"", dev.DeviceID)
		}
	})
}
