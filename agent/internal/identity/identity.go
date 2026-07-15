// Package identity manages the BuzzPi device identity.
//
// Each BuzzPi Runtime has a persistent identity consisting of an Ed25519
// keypair and a derived device identifier. The identity is:
// - Generated on first boot (or factory reset)
// - Stored in a JSON file in the identity directory
// - Never leaves the device (private key stays local)
// - Used for all pairing and session operations
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// IdentityDir is the default identity directory name under the data root.
	IdentityDir = "identity"
	// IdentityFile is the identity file name.
	IdentityFile = "identity.json"
	// IdentityDirPerm is the permission mode for the identity directory.
	IdentityDirPerm os.FileMode = 0700
	// IdentityFilePerm is the permission mode for the identity file.
	IdentityFilePerm os.FileMode = 0600
)

// DeviceIdentity represents the persistent identity of a BuzzPi Runtime.
type DeviceIdentity struct {
	// DeviceID is the public identifier, derived from the public key.
	DeviceID string `json:"device_id"`
	// PrivateKey is the Ed25519 private key (base64-encoded).
	PrivateKey string `json:"private_key"`
	// PublicKey is the Ed25519 public key (base64-encoded).
	PublicKey string `json:"public_key"`
	// CreatedAt is when this identity was generated.
	CreatedAt time.Time `json:"created_at"`
	// FriendlyName is the human-readable device name.
	FriendlyName string `json:"friendly_name,omitempty"`
	// Platform is the hardware/platform identifier.
	Platform string `json:"platform,omitempty"`
	// RuntimeVersion is the software version that created this identity.
	RuntimeVersion string `json:"runtime_version,omitempty"`
}

// Generate creates a new device identity with a fresh Ed25519 keypair.
func Generate() (*DeviceIdentity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate keypair: %w", err)
	}

	// device_id = "dev_" + first 12 bytes of SHA-256(public key), base62-encoded
	hash := sha256.Sum256(pub)
	id := "dev_" + base62Encode(hash[:12])

	return &DeviceIdentity{
		DeviceID:   id,
		PrivateKey: base64Encode(priv),
		PublicKey:  base64Encode(pub),
		CreatedAt:  time.Now().UTC(),
	}, nil
}

// ParsePrivateKey decodes the base64-encoded Ed25519 private key.
func (d *DeviceIdentity) ParsePrivateKey() (ed25519.PrivateKey, error) {
	return base64DecodeKey(d.PrivateKey, 64) // Ed25519 private key is 64 bytes
}

// ParsePublicKey decodes the base64-encoded Ed25519 public key.
func (d *DeviceIdentity) ParsePublicKey() (ed25519.PublicKey, error) {
	return base64DecodeKey(d.PublicKey, 32) // Ed25519 public key is 32 bytes
}

// IdentityPath returns the full path to the identity file.
func IdentityPath(dataDir string) string {
	return filepath.Join(dataDir, IdentityDir, IdentityFile)
}

// Load reads a device identity from a JSON file.
func Load(path string) (*DeviceIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read identity file: %w", err)
	}

	var id DeviceIdentity
	if err := json.Unmarshal(data, &id); err != nil {
		return nil, fmt.Errorf("parse identity file: %w", err)
	}

	if id.DeviceID == "" {
		return nil, fmt.Errorf("identity file missing device_id")
	}
	if id.PrivateKey == "" || id.PublicKey == "" {
		return nil, fmt.Errorf("identity file missing key material")
	}

	// Verify the key material is valid.
	if _, err := id.ParsePrivateKey(); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return &id, nil
}

// Save writes the device identity to a JSON file.
// Creates the directory if it does not exist.
func (d *DeviceIdentity) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, IdentityDirPerm); err != nil {
		return fmt.Errorf("create identity directory: %w", err)
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal identity: %w", err)
	}

	if err := os.WriteFile(path, data, IdentityFilePerm); err != nil {
		return fmt.Errorf("write identity file: %w", err)
	}

	return nil
}

// Ensure loads an existing identity or generates a new one.
func Ensure(path string) (*DeviceIdentity, error) {
	id, err := Load(path)
	if err == nil {
		return id, nil
	}

	if _, statErr := os.Stat(path); statErr == nil {
		return nil, fmt.Errorf("load identity: %w", err)
	}

	id, err = Generate()
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}

	if err := id.Save(path); err != nil {
		return nil, fmt.Errorf("save identity: %w", err)
	}

	return id, nil
}
