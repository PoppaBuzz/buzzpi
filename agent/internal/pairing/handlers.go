// Package pairing implements the BPP pairing protocol handlers.
//
// The pairing package manages the lifecycle of client-device trust:
//   - pair.initiate — start pairing, return 6-digit PIN
//   - pair.verify — validate PIN, exchange keys, create session
//   - pair.status — query pairing state for a device
//   - pair.unpair — remove pairing and revoke session
//
// Pairing state is persisted in the state store (BoltDB).
package pairing

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/buzzpi/agent/internal/state"
)

// Default pairing parameters.
const (
	PINDuration     = 2 * time.Minute
	PINLockoutCount = 3
	PINLockoutTime  = 30 * time.Second
	SessionDuration = 24 * time.Hour
	MaxPairings     = 10
)

// Handler implements the BPP pairing methods.
type Handler struct {
	store    *state.Store
	log      *slog.Logger
	identity struct {
		DeviceID string
	}

	mu         sync.Mutex
	pendingPIN map[string]*pinEntry // deviceID/clientID → PIN entry
	lockouts   map[string]time.Time // client key → lockout until
	attempts   map[string]int       // client key → failed attempts
}

type pinEntry struct {
	PIN       string    `json:"pin"`
	ExpiresAt time.Time `json:"expires_at"`
	DeviceID  string    `json:"device_id"`
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}

// PairInitiateParams are the parameters for pair.initiate.
type PairInitiateParams struct {
	DeviceID       string   `json:"device_id"`
	ClientName     string   `json:"client_name,omitempty"`
	SupportedAuths []string `json:"supported_auths,omitempty"`
}

// PairInitiateResult is returned by pair.initiate.
type PairInitiateResult struct {
	Method      string `json:"method"`
	PIN         string `json:"pin"`
	ExpiresAt   int64  `json:"expires_at"`
	SessionID   string `json:"session_id"`
	DeviceID    string `json:"device_id"`
	DeviceName  string `json:"device_name,omitempty"`
}

// PairVerifyParams are the parameters for pair.verify.
type PairVerifyParams struct {
	SessionID       string `json:"session_id"`
	PIN             string `json:"pin"`
	ClientPublicKey string `json:"client_public_key,omitempty"`
	ClientName      string `json:"client_name,omitempty"`
}

// PairVerifyResult is returned by pair.verify.
type PairVerifyResult struct {
	SessionToken string `json:"session_token"`
	ExpiresAt    int64  `json:"expires_at"`
	DeviceID     string `json:"device_id"`
}

// PairStatusResult is returned by pair.status.
type PairStatusResult struct {
	Paired    bool   `json:"paired"`
	ClientID  string `json:"client_id,omitempty"`
	ClientName string `json:"client_name,omitempty"`
	PairedAt  string `json:"paired_at,omitempty"`
}

// NewHandler creates a new pairing handler.
func NewHandler(store *state.Store, deviceID string, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{
		store:      store,
		log:        log.With("component", "pairing"),
		pendingPIN: make(map[string]*pinEntry),
		lockouts:   make(map[string]time.Time),
		attempts:   make(map[string]int),
		identity: struct{ DeviceID string }{
			DeviceID: deviceID,
		},
	}
}

// HandleInitiate processes a pair.initiate request.
// Generates a 6-digit PIN, stores it temporarily, and returns it.
func (h *Handler) HandleInitiate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req PairInitiateParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// Check if device has reached max pairings.
	pairs, err := h.store.List("pairs")
	if err != nil {
		return nil, fmt.Errorf("list pairs: %w", err)
	}
	if len(pairs) >= MaxPairings {
		return nil, fmt.Errorf("device at maximum pairings (%d)", MaxPairings)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check lockout for this client.
	clientKey := req.DeviceID + "/" + req.ClientName
	if until, ok := h.lockouts[clientKey]; ok {
		if time.Now().Before(until) {
			return nil, fmt.Errorf("too many attempts, try again in %v", time.Until(until).Round(time.Second))
		}
		delete(h.lockouts, clientKey)
		delete(h.attempts, clientKey)
	}

	// Generate 6-digit PIN.
	pin := generatePIN()
	sessionID := generateSessionID()

	entry := &pinEntry{
		PIN:       pin,
		ExpiresAt: time.Now().Add(PINDuration),
		DeviceID:  req.DeviceID,
		ClientID:  req.ClientName,
		CreatedAt: time.Now(),
	}

	h.pendingPIN[sessionID] = entry

	h.log.Info("pairing initiated",
		"device", req.DeviceID,
		"client", req.ClientName,
		"session", sessionID,
	)
	fmt.Fprintf(os.Stderr, "\n╔══════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║        BUZZPI PAIRING PIN           ║\n")
	fmt.Fprintf(os.Stderr, "╠══════════════════════════════════════╣\n")
	fmt.Fprintf(os.Stderr, "║                                      ║\n")
	fmt.Fprintf(os.Stderr, "║            %s              ║\n", pin)
	fmt.Fprintf(os.Stderr, "║                                      ║\n")
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════╝\n\n")

	return &PairInitiateResult{
		Method:    "pin",
		PIN:       pin,
		ExpiresAt: entry.ExpiresAt.Unix(),
		SessionID: sessionID,
		DeviceID:  h.identity.DeviceID,
	}, nil
}

// HandleVerify processes a pair.verify request.
// Validates the PIN, creates a session, and stores the pairing.
func (h *Handler) HandleVerify(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req PairVerifyParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.SessionID == "" || req.PIN == "" {
		return nil, fmt.Errorf("session_id and pin are required")
	}

	h.mu.Lock()

	// Look up the pending PIN entry.
	entry, ok := h.pendingPIN[req.SessionID]
	if !ok {
		h.mu.Unlock()
		return nil, fmt.Errorf("invalid session_id")
	}

	// Check expiry.
	if time.Now().After(entry.ExpiresAt) {
		delete(h.pendingPIN, req.SessionID)
		h.mu.Unlock()
		return nil, fmt.Errorf("pin expired")
	}

	// Check lockout for this client.
	clientKey := entry.DeviceID + "/" + entry.ClientID
	if until, ok := h.lockouts[clientKey]; ok {
		if time.Now().Before(until) {
			h.mu.Unlock()
			return nil, fmt.Errorf("too many attempts, try again in %v", time.Until(until).Round(time.Second))
		}
		delete(h.lockouts, clientKey)
		delete(h.attempts, clientKey)
	}

	// Verify PIN.
	if req.PIN != entry.PIN {
		h.attempts[clientKey]++
		if h.attempts[clientKey] >= PINLockoutCount {
			h.lockouts[clientKey] = time.Now().Add(PINLockoutTime)
			delete(h.attempts, clientKey)
		}
		delete(h.pendingPIN, req.SessionID)
		h.mu.Unlock()

		h.log.Warn("invalid pin",
			"device", entry.DeviceID,
			"client", entry.ClientID,
			"attempts", h.attempts[clientKey]+1,
		)
		return nil, fmt.Errorf("invalid pin")
	}

	// PIN correct — clean up and proceed.
	clientName := req.ClientName
	if clientName == "" {
		clientName = entry.ClientID
	}
	delete(h.pendingPIN, req.SessionID)
	delete(h.lockouts, clientKey)
	delete(h.attempts, clientKey)
	h.mu.Unlock()

	// Generate session token.
	sessionToken := generateSessionToken()
	now := time.Now()
	expiresAt := now.Add(SessionDuration)

	// Save pairing in state store.
	pairing := &state.Pairing{
		DeviceID:   h.identity.DeviceID,
		AccountID:  clientName,
		ClientID:   clientName,
		ClientName: clientName,
		Role:       "admin",
		PairedAt:   now,
		PublicKey:  req.ClientPublicKey,
	}
	if err := h.store.SavePairing(pairing); err != nil {
		return nil, fmt.Errorf("save pairing: %w", err)
	}

	// Save session in state store.
	session := &state.Session{
		SessionID:  sessionToken,
		DeviceID:   h.identity.DeviceID,
		ClientID:   clientName,
		ClientName: clientName,
		Role:       "admin",
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastUsed:   now,
	}
	if err := h.store.SaveSession(session); err != nil {
		// Try to clean up the pairing.
		_ = h.store.Delete("pairs", pairing.DeviceID+"/"+pairing.ClientID)
		return nil, fmt.Errorf("save session: %w", err)
	}

	h.log.Info("pairing successful",
		"device", h.identity.DeviceID,
		"client", clientName,
		"session_expires", expiresAt,
	)

	return &PairVerifyResult{
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt.Unix(),
		DeviceID:     h.identity.DeviceID,
	}, nil
}

// HandleStatus processes a pair.status request.
func (h *Handler) HandleStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// Check if there are any pairings for this device.
	pairs, err := h.store.List("pairs")
	if err != nil {
		return nil, fmt.Errorf("list pairs: %w", err)
	}

	if len(pairs) == 0 {
		return &PairStatusResult{Paired: false}, nil
	}

	// Return the first pairing's info.
	var firstPairing state.Pairing
	for _, data := range pairs {
		if err := json.Unmarshal(data, &firstPairing); err == nil {
			break
		}
	}

	return &PairStatusResult{
		Paired:    true,
		ClientID:  firstPairing.ClientID,
		ClientName: firstPairing.ClientName,
		PairedAt:  firstPairing.PairedAt.Format(time.RFC3339),
	}, nil
}

// HandleUnpair processes a pair.unpair request.
func (h *Handler) HandleUnpair(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionToken string `json:"session_token"`
		DeviceID     string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// Find and delete the session.
	if req.SessionToken != "" {
		if err := h.store.DeleteSession(req.SessionToken); err != nil {
			h.log.Warn("session not found for unpair", "session", req.SessionToken)
		}
	}

	// Remove all pairings for this device.
	// This is a simplified unpair — in production we'd remove specific pairings.
	pairs, err := h.store.List("pairs")
	if err == nil {
		for key := range pairs {
			if err := h.store.Delete("pairs", key); err != nil {
				h.log.Warn("failed to delete pairing", "key", key, "error", err)
			}
		}
	}

	h.log.Info("device unpaired",
		"device", req.DeviceID,
	)

	return map[string]interface{}{
		"unpaired": true,
	}, nil
}

// generatePIN creates a random 6-digit PIN.
func generatePIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%900000+100000)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// generateSessionID creates a random session identifier for pairing sessions.
func generateSessionID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return fmt.Sprintf("pair_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("pair_%s", n.Text(36))
}

// generateSessionToken creates a session token for authenticated clients.
func generateSessionToken() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("sess_%x", b)
}
