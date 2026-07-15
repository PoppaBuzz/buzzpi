package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// clientState stores pairing information for CLI-managed devices.
type clientState struct {
	mu      sync.Mutex
	Devices map[string]pairedDevice `json:"devices"`
	path    string
}

// pairedDevice represents a device the CLI has paired with.
type pairedDevice struct {
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	Address      string `json:"address"`
	Port         int    `json:"port"`
	SessionToken string `json:"session_token"`
	PairedAt     string `json:"paired_at"`
}

func loadClientState() (*clientState, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".buzzpi")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, "state.json")

	s := &clientState{
		Devices: make(map[string]pairedDevice),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return s, nil
}

func (s *clientState) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *clientState) addDevice(d pairedDevice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Devices[d.DeviceID] = d
}

func (s *clientState) removeDevice(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Devices, id)
}

func (s *clientState) getDevice(id string) (pairedDevice, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.Devices[id]
	return d, ok
}

func (s *clientState) listDevices() []pairedDevice {
	s.mu.Lock()
	defer s.mu.Unlock()
	var list []pairedDevice
	for _, d := range s.Devices {
		list = append(list, d)
	}
	return list
}
