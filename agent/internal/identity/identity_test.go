package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if id.DeviceID == "" {
		t.Error("DeviceID is empty")
	}
	if len(id.DeviceID) < 10 || id.DeviceID[:4] != "dev_" {
		t.Errorf("DeviceID = %q, want 'dev_...' prefix", id.DeviceID)
	}
	if id.PrivateKey == "" {
		t.Error("PrivateKey is empty")
	}
	if id.PublicKey == "" {
		t.Error("PublicKey is empty")
	}
	if id.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}

	// Verify keys parse correctly.
	priv, err := id.ParsePrivateKey()
	if err != nil {
		t.Fatalf("ParsePrivateKey() error = %v", err)
	}
	if len(priv) != 64 {
		t.Errorf("ParsePrivateKey() len = %d, want 64", len(priv))
	}

	pub, err := id.ParsePublicKey()
	if err != nil {
		t.Fatalf("ParsePublicKey() error = %v", err)
	}
	if len(pub) != 32 {
		t.Errorf("ParsePublicKey() len = %d, want 32", len(pub))
	}

	// Public key should match the private key's public component.
	if len(priv) > 32 {
		expectedPub := priv[32:]
		for i := range pub {
			if pub[i] != expectedPub[i] {
				t.Errorf("Public key mismatch at byte %d", i)
				break
			}
		}
	}
}

func TestGenerateUnique(t *testing.T) {
	id1, _ := Generate()
	id2, _ := Generate()

	if id1.DeviceID == id2.DeviceID {
		t.Error("Two generated identities have the same DeviceID")
	}
	if id1.PrivateKey == id2.PrivateKey {
		t.Error("Two generated identities have the same PrivateKey")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.json")

	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	id.FriendlyName = "test-device"
	id.Platform = "linux/amd64"
	id.RuntimeVersion = "0.2.0"

	if err := id.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	_ = info

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.DeviceID != id.DeviceID {
		t.Errorf("DeviceID = %q, want %q", loaded.DeviceID, id.DeviceID)
	}
	if loaded.PrivateKey != id.PrivateKey {
		t.Error("PrivateKey mismatch")
	}
	if loaded.PublicKey != id.PublicKey {
		t.Error("PublicKey mismatch")
	}
	if loaded.FriendlyName != "test-device" {
		t.Errorf("FriendlyName = %q, want 'test-device'", loaded.FriendlyName)
	}
	if loaded.Platform != "linux/amd64" {
		t.Errorf("Platform = %q, want 'linux/amd64'", loaded.Platform)
	}
	if loaded.RuntimeVersion != "0.2.0" {
		t.Errorf("RuntimeVersion = %q, want '0.2.0'", loaded.RuntimeVersion)
	}
}

func TestLoadNonexistent(t *testing.T) {
	_, err := Load("/nonexistent/path/identity.json")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent path")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for invalid JSON")
	}
}

func TestLoadMissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")

	// Empty object — missing device_id and keys.
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for empty identity")
	}
}

func TestEnsureNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity", "identity.json")

	id, err := Ensure(path)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if id.DeviceID == "" {
		t.Error("DeviceID is empty after Ensure (new)")
	}
}

func TestEnsureExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity", "identity.json")

	// Create identity first.
	id1, err := Ensure(path)
	if err != nil {
		t.Fatalf("Ensure() (first) error = %v", err)
	}

	// Load again — should return the same identity.
	id2, err := Ensure(path)
	if err != nil {
		t.Fatalf("Ensure() (second) error = %v", err)
	}

	if id1.DeviceID != id2.DeviceID {
		t.Error("Ensure() returned different DeviceID on second call")
	}
	if id1.PrivateKey != id2.PrivateKey {
		t.Error("Ensure() returned different PrivateKey on second call")
	}
}

func TestIdentityPath(t *testing.T) {
	path := IdentityPath("/var/lib/buzzpi")
	expected := filepath.Join("/var/lib/buzzpi", "identity", "identity.json")
	if path != expected {
		t.Errorf("IdentityPath = %q, want %q", path, expected)
	}
}

func TestBase62EncodeDeterministic(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	r1 := base62Encode(data)
	r2 := base62Encode(data)
	if r1 != r2 {
		t.Error("base62Encode is not deterministic")
	}
	if r1 == "" {
		t.Error("base62Encode returned empty string")
	}
}

func TestBase62EncodeEmpty(t *testing.T) {
	if r := base62Encode(nil); r != "" {
		t.Errorf("base62Encode(nil) = %q, want empty", r)
	}
	if r := base62Encode([]byte{}); r != "" {
		t.Errorf("base62Encode([]) = %q, want empty", r)
	}
}
