package pairing

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/buzzpi/agent/internal/state"
)

func newTestHandler(t *testing.T) (*Handler, *state.Store) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := state.Open(dbPath, slog.Default())
	if err != nil {
		t.Fatalf("state.Open() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })

	h := NewHandler(store, "dev_test123", slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn})))
	return h, store
}

func TestNewHandler(t *testing.T) {
	h, store := newTestHandler(t)
	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}
	_ = store
}

func TestHandleInitiate(t *testing.T) {
	h, _ := newTestHandler(t)

	params := PairInitiateParams{
		DeviceID:   "dev_test123",
		ClientName: "test-client",
	}

	data, _ := json.Marshal(params)
	result, err := h.HandleInitiate(context.Background(), data)
	if err != nil {
		t.Fatalf("HandleInitiate() error = %v", err)
	}

	r, ok := result.(*PairInitiateResult)
	if !ok {
		t.Fatalf("HandleInitiate() returned %T, want *PairInitiateResult", result)
	}

	if r.Method != "pin" {
		t.Errorf("Method = %q, want %q", r.Method, "pin")
	}
	if len(r.PIN) != 6 {
		t.Errorf("PIN = %q, want 6 digits", r.PIN)
	}
	if r.DeviceID != "dev_test123" {
		t.Errorf("DeviceID = %q, want %q", r.DeviceID, "dev_test123")
	}
	if r.SessionID == "" {
		t.Error("SessionID is empty")
	}
	if r.ExpiresAt == 0 {
		t.Error("ExpiresAt is zero")
	}
}

func TestHandleInitiateMissingDeviceID(t *testing.T) {
	h, _ := newTestHandler(t)

	params := PairInitiateParams{}
	data, _ := json.Marshal(params)
	_, err := h.HandleInitiate(context.Background(), data)
	if err == nil {
		t.Fatal("HandleInitiate() expected error for missing device_id")
	}
}

func TestHandleVerifySuccess(t *testing.T) {
	h, _ := newTestHandler(t)

	// Initiate pairing.
	initParams := PairInitiateParams{
		DeviceID:   "dev_test123",
		ClientName: "test-client",
	}
	data, _ := json.Marshal(initParams)
	result, err := h.HandleInitiate(context.Background(), data)
	if err != nil {
		t.Fatalf("HandleInitiate() error = %v", err)
	}
	r := result.(*PairInitiateResult)

	// Verify with correct PIN.
	verifyParams := PairVerifyParams{
		SessionID:       r.SessionID,
		PIN:             r.PIN,
		ClientPublicKey: "base64-fake-public-key",
		ClientName:      "test-client",
	}
	data, _ = json.Marshal(verifyParams)
	vresult, err := h.HandleVerify(context.Background(), data)
	if err != nil {
		t.Fatalf("HandleVerify() error = %v", err)
	}

	vr, ok := vresult.(*PairVerifyResult)
	if !ok {
		t.Fatalf("HandleVerify() returned %T, want *PairVerifyResult", vresult)
	}

	if vr.SessionToken == "" {
		t.Error("SessionToken is empty")
	}
	if len(vr.SessionToken) < 10 || vr.SessionToken[:5] != "sess_" {
		t.Errorf("SessionToken = %q, want 'sess_' prefix", vr.SessionToken)
	}
	if vr.DeviceID != "dev_test123" {
		t.Errorf("DeviceID = %q, want %q", vr.DeviceID, "dev_test123")
	}
}

func TestHandleVerifyInvalidPIN(t *testing.T) {
	h, _ := newTestHandler(t)

	// Initiate pairing.
	initParams := PairInitiateParams{
		DeviceID:   "dev_test123",
		ClientName: "test-client",
	}
	data, _ := json.Marshal(initParams)
	result, _ := h.HandleInitiate(context.Background(), data)
	r := result.(*PairInitiateResult)

	// Verify with wrong PIN.
	verifyParams := PairVerifyParams{
		SessionID: r.SessionID,
		PIN:       "000000",
	}
	data, _ = json.Marshal(verifyParams)
	_, err := h.HandleVerify(context.Background(), data)
	if err == nil {
		t.Fatal("HandleVerify() expected error for invalid PIN")
	}
}

func TestHandleVerifyExpiredSession(t *testing.T) {
	h, _ := newTestHandler(t)

	// Use a non-existent session ID.
	verifyParams := PairVerifyParams{
		SessionID: "pair_nonexistent",
		PIN:       "123456",
	}
	data, _ := json.Marshal(verifyParams)
	_, err := h.HandleVerify(context.Background(), data)
	if err == nil {
		t.Fatal("HandleVerify() expected error for invalid session")
	}
}

func TestHandleStatusNotPaired(t *testing.T) {
	h, _ := newTestHandler(t)

	params := map[string]string{"device_id": "dev_test123"}
	data, _ := json.Marshal(params)
	result, err := h.HandleStatus(context.Background(), data)
	if err != nil {
		t.Fatalf("HandleStatus() error = %v", err)
	}

	sr, ok := result.(*PairStatusResult)
	if !ok {
		t.Fatalf("HandleStatus() returned %T, want *PairStatusResult", result)
	}
	if sr.Paired {
		t.Error("Paired = true, want false for new device")
	}
}

func TestHandleStatusAfterPairing(t *testing.T) {
	h, _ := newTestHandler(t)

	// Initiate and verify.
	initParams := PairInitiateParams{DeviceID: "dev_test123", ClientName: "test-client"}
	data, _ := json.Marshal(initParams)
	result, _ := h.HandleInitiate(context.Background(), data)
	r := result.(*PairInitiateResult)

	verifyParams := PairVerifyParams{
		SessionID:       r.SessionID,
		PIN:             r.PIN,
		ClientPublicKey: "key",
		ClientName:      "test-client",
	}
	data, _ = json.Marshal(verifyParams)
	_, err := h.HandleVerify(context.Background(), data)
	if err != nil {
		t.Fatalf("HandleVerify() error = %v", err)
	}

	// Check status.
	statusParams := map[string]string{"device_id": "dev_test123"}
	data, _ = json.Marshal(statusParams)
	sresult, _ := h.HandleStatus(context.Background(), data)
	sr := sresult.(*PairStatusResult)

	if !sr.Paired {
		t.Error("Paired = false, want true after pairing")
	}
}

func TestHandleUnpair(t *testing.T) {
	h, _ := newTestHandler(t)

	// Initiate and verify.
	initParams := PairInitiateParams{DeviceID: "dev_test123", ClientName: "test-client"}
	data, _ := json.Marshal(initParams)
	result, _ := h.HandleInitiate(context.Background(), data)
	r := result.(*PairInitiateResult)

	verifyParams := PairVerifyParams{
		SessionID:       r.SessionID,
		PIN:             r.PIN,
		ClientPublicKey: "key",
		ClientName:      "test-client",
	}
	data, _ = json.Marshal(verifyParams)
	vresult, _ := h.HandleVerify(context.Background(), data)
	vr := vresult.(*PairVerifyResult)

	// Unpair.
	unpairParams := map[string]string{
		"session_token": vr.SessionToken,
		"device_id":     "dev_test123",
	}
	data, _ = json.Marshal(unpairParams)
	uresult, err := h.HandleUnpair(context.Background(), data)
	if err != nil {
		t.Fatalf("HandleUnpair() error = %v", err)
	}

	u, ok := uresult.(map[string]interface{})
	if !ok {
		t.Fatalf("HandleUnpair() returned %T", uresult)
	}
	if u["unpaired"] != true {
		t.Errorf("unpaired = %v, want true", u["unpaired"])
	}
}

func TestGeneratePINLength(t *testing.T) {
	for i := 0; i < 100; i++ {
		pin := generatePIN()
		if len(pin) != 6 {
			t.Fatalf("generatePIN() = %q (len=%d), want 6 digits", pin, len(pin))
		}
	}
}

func TestGenerateSessionTokenFormat(t *testing.T) {
	for i := 0; i < 10; i++ {
		token := generateSessionToken()
		if len(token) < 10 || token[:5] != "sess_" {
			t.Errorf("generateSessionToken() = %q, want 'sess_' prefix", token)
		}
	}
}
