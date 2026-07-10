// Package bpp implements the BuzzPi Protocol (BPP) wire format.
//
// Handshake defines the capability exchange and session establishment protocol
// that runs when a BPP connection is first established.
//
// Handshake flow:
//
//	Client                       Agent
//	  │                            │
//	  ├──── CapabilityOffer ──────►│
//	  │                            ├── Validate version
//	  │                            ├── Check session token (if provided)
//	  │◄─── CapabilityAccept ─────┤
//	  │       OR                   │
//	  │◄─── AuthChallenge ────────┤  (if no valid session)
//	  │           (PIN prompt)     │
//	  ├──── AuthResponse ─────────►│
//	  │                            ├── Verify PIN
//	  │◄─── SessionEstablished ───┤
//	  │       OR                   │
//	  │◄─── AuthFailed ──────────►│
package bpp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"
)

// Handshake protocol method names.
const (
	MethodHandshake        = "bpp.handshake"
	MethodAuthChallenge    = "bpp.auth.challenge"
	MethodAuthResponse     = "bpp.auth.response"
	MethodCapabilityUpdate = "bpp.capability.update"
)

// Protocol version negotiation constants.
const (
	MinSupportedVersion = 1
	MaxSupportedVersion = 1
)

// Capability IDs — the features a BPP endpoint may support.
const (
	CapTerminal     = "terminal"
	CapScreen       = "screen"
	CapFile         = "file"
	CapSystem       = "system"
	CapPlugin       = "plugin"
	CapGPIO         = "gpio"
	CapCamera       = "camera"
	CapAudio        = "audio"
	CapClipboard    = "clipboard"
	CapNotification = "notification"
	CapEncryption   = "encryption"
	CapCompression  = "compression"
)

// DefaultCapabilities are the capabilities every BuzzPi Agent supports.
var DefaultCapabilities = []string{
	CapTerminal,
	CapFile,
	CapSystem,
	CapPlugin,
}

// HandshakeRole identifies which side of the handshake we are on.
type HandshakeRole int

const (
	RoleClient HandshakeRole = iota
	RoleAgent
)

func (r HandshakeRole) String() string {
	switch r {
	case RoleClient:
		return "client"
	case RoleAgent:
		return "agent"
	default:
		return "unknown"
	}
}

// HandshakeState tracks the progress of the capability exchange and
// session establishment protocol.
type HandshakeState int

const (
	HSWaiting         HandshakeState = iota // Waiting to begin handshake
	HSCapabilityOffer                       // Offered capabilities (sent/received)
	HSAuthenticating                        // Authentication challenge in progress
	HSReady                                 // Handshake complete, ready for BPP messages
	HSFailed                                // Handshake failed
)

func (s HandshakeState) String() string {
	switch s {
	case HSWaiting:
		return "waiting"
	case HSCapabilityOffer:
		return "capability_offer"
	case HSAuthenticating:
		return "authenticating"
	case HSReady:
		return "ready"
	case HSFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// CapabilityOffer is sent by the connecting side to announce its capabilities
// and negotiate the protocol version.
type CapabilityOffer struct {
	Version       int      `json:"v"`                        // Preferred protocol version
	MinVersion    int      `json:"min_v"`                    // Minimum supported version
	Capabilities  []string `json:"caps"`                     // Supported capability IDs
	SessionToken  string   `json:"session,omitempty"`        // Existing session token (if reconnecting)
	DeviceID      string   `json:"device_id,omitempty"`      // Device ID (agent identifies itself)
	ClientName    string   `json:"client_name,omitempty"`    // Client identification
	ClientVersion string   `json:"client_version,omitempty"` // Client software version
}

// CapabilityAccept is sent in response to a CapabilityOffer when the
// handshake can proceed without additional authentication (valid session).
type CapabilityAccept struct {
	Version       int      `json:"v"`                 // Agreed protocol version
	Capabilities  []string `json:"caps"`              // Supported capabilities
	DeviceID      string   `json:"device_id"`         // Agent device identifier
	DeviceName    string   `json:"device_name"`       // Agent friendly name
	Platform      string   `json:"platform"`          // Agent platform string
	SessionToken  string   `json:"session,omitempty"` // New/refreshed session token
	ServerVersion string   `json:"server_version"`    // Agent version
}

// AuthChallenge is sent when the connecting side needs to authenticate
// (no valid session token).
type AuthChallenge struct {
	ChallengeType string `json:"type"`                // "pin" or "signature"
	PIN           string `json:"pin,omitempty"`       // 6-digit numeric PIN (for PIN mode)
	Challenge     string `json:"challenge,omitempty"` // Cryptographic challenge (for signature mode)
	DeviceID      string `json:"device_id"`           // Device identifier
	DeviceName    string `json:"device_name"`         // Device friendly name
	ExpiresAt     int64  `json:"expires_at"`          // Challenge expiry (Unix timestamp)
}

// AuthResponse is sent in reply to an AuthChallenge.
type AuthResponse struct {
	PIN        string `json:"pin,omitempty"`         // PIN entered by user
	Signature  string `json:"signature,omitempty"`   // Cryptographic signature
	DeviceID   string `json:"device_id"`             // Target device ID
	ClientName string `json:"client_name,omitempty"` // Client identification
}

// SessionEstablished is sent when authentication succeeds.
type SessionEstablished struct {
	SessionToken string   `json:"session"`     // New session token
	DeviceID     string   `json:"device_id"`   // Agent device identifier
	DeviceName   string   `json:"device_name"` // Agent friendly name
	Capabilities []string `json:"caps"`        // Full capability list
	Version      int      `json:"v"`           // Agreed protocol version
	ExpiresAt    int64    `json:"expires_at"`  // Token expiry
}

// HandshakeResult captures the outcome of a completed handshake.
type HandshakeResult struct {
	DeviceID     string
	DeviceName   string
	Platform     string
	Version      int
	Capabilities []string
	SessionToken string
	IsNewSession bool // true if a new session was established
}

// HandshakeConfig configures the handshake behaviour for an endpoint.
type HandshakeConfig struct {
	Role          HandshakeRole
	DeviceID      string
	DeviceName    string
	Platform      string
	Version       string // semantic version of the software
	Capabilities  []string
	SessionToken  string // existing token (empty if new)
	TokenSecret   string // HMAC secret for session token generation
	TokenDuration time.Duration
	PINFunc       func() (string, error) // callback to display/get PIN (agent generates, client reads)
}

// Handshake manages the BPP capability exchange and session establishment
// protocol for a single connection.
type Handshake struct {
	config         HandshakeConfig
	state          HandshakeState
	result         *HandshakeResult
	log            *slog.Logger
	mu             sync.Mutex
	pendingPIN     string
	pendingVersion int
}

// NewHandshake creates a new Handshake instance.
func NewHandshake(config HandshakeConfig) *Handshake {
	if config.TokenDuration <= 0 {
		config.TokenDuration = 24 * time.Hour
	}
	if config.Capabilities == nil {
		config.Capabilities = DefaultCapabilities
	}

	return &Handshake{
		config: config,
		state:  HSWaiting,
		log:    slog.Default().With("component", "bpp-handshake", "role", config.Role),
	}
}

// State returns the current handshake state.
func (h *Handshake) State() HandshakeState {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.state
}

// Result returns the handshake result (nil if not yet complete).
func (h *Handshake) Result() *HandshakeResult {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.result == nil {
		return nil
	}
	r := *h.result
	return &r
}

// CreateOffer builds the initial CapabilityOffer message.
func (h *Handshake) CreateOffer() *CapabilityOffer {
	return &CapabilityOffer{
		Version:       MaxSupportedVersion,
		MinVersion:    MinSupportedVersion,
		Capabilities:  h.config.Capabilities,
		SessionToken:  h.config.SessionToken,
		DeviceID:      h.config.DeviceID,
		ClientName:    h.config.DeviceName,
		ClientVersion: h.config.Version,
	}
}

// HandleOffer processes an incoming CapabilityOffer from the peer.
// Returns either a CapabilityAccept (session OK) or an AuthChallenge
// (need authentication), or an error if versions are incompatible.
func (h *Handshake) HandleOffer(offer *CapabilityOffer) (interface{}, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.state = HSCapabilityOffer

	// Version negotiation: find the highest mutually supported version.
	agreed := negotiateVersion(offer.MinVersion, offer.Version, MinSupportedVersion, MaxSupportedVersion)
	if agreed < 0 {
		h.state = HSFailed
		return nil, fmt.Errorf(
			"bpp: incompatible versions: peer [%d-%d], local [%d-%d]",
			offer.MinVersion, offer.Version,
			MinSupportedVersion, MaxSupportedVersion,
		)
	}

	// Check for existing valid session.
	if offer.SessionToken != "" {
		deviceID, ok := h.validateSessionToken(offer.SessionToken)
		if ok {
			// Session is valid — skip authentication.
			newToken := h.generateSessionToken(deviceID)
			accept := &CapabilityAccept{
				Version:       agreed,
				Capabilities:  h.config.Capabilities,
				DeviceID:      h.config.DeviceID,
				DeviceName:    h.config.DeviceName,
				Platform:      h.config.Platform,
				SessionToken:  newToken,
				ServerVersion: h.config.Version,
			}
			h.result = &HandshakeResult{
				DeviceID:     h.config.DeviceID,
				DeviceName:   h.config.DeviceName,
				Platform:     h.config.Platform,
				Version:      agreed,
				Capabilities: intersectCapabilities(h.config.Capabilities, offer.Capabilities),
				SessionToken: newToken,
				IsNewSession: false,
			}
			h.state = HSReady
			return accept, nil
		}
	}

	// Need authentication.
	pin := generatePIN()
	challenge := &AuthChallenge{
		ChallengeType: "pin",
		PIN:           pin,
		DeviceID:      h.config.DeviceID,
		DeviceName:    h.config.DeviceName,
		ExpiresAt:     time.Now().Add(2 * time.Minute).Unix(),
	}

	// Call the PIN callback (agent displays it to the user).
	if h.config.PINFunc != nil {
		if _, err := h.config.PINFunc(); err != nil {
			// The PINFunc callback is expected to handle display internally.
			// This return is only for error handling.
			h.log.Warn("PIN function returned error", "error", err)
		}
	}

	// Store the PIN for later verification.
	h.pendingPIN = pin
	h.pendingVersion = agreed
	h.state = HSAuthenticating

	h.log.Info("auth challenge sent", "type", "pin", "device", h.config.DeviceID)
	return challenge, nil
}

// HandleAuthResponse verifies an authentication response.
// Returns SessionEstablished on success, error on failure.
func (h *Handshake) HandleAuthResponse(resp *AuthResponse) (*SessionEstablished, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.state != HSAuthenticating {
		return nil, fmt.Errorf("bpp: not in authenticating state (state=%s)", h.state)
	}

	// Verify PIN.
	if resp.PIN != h.pendingPIN {
		h.state = HSFailed
		h.pendingPIN = ""
		return nil, fmt.Errorf("bpp: invalid PIN")
	}

	h.pendingPIN = ""

	// Authentication successful. Generate session token.
	sessionToken := h.generateSessionToken(resp.DeviceID)
	expiresAt := time.Now().Add(h.config.TokenDuration)

	est := &SessionEstablished{
		SessionToken: sessionToken,
		DeviceID:     h.config.DeviceID,
		DeviceName:   h.config.DeviceName,
		Capabilities: h.config.Capabilities,
		Version:      h.pendingVersion,
		ExpiresAt:    expiresAt.Unix(),
	}

	h.result = &HandshakeResult{
		DeviceID:     h.config.DeviceID,
		DeviceName:   h.config.DeviceName,
		Platform:     h.config.Platform,
		Version:      h.pendingVersion,
		Capabilities: h.config.Capabilities,
		SessionToken: sessionToken,
		IsNewSession: true,
	}
	h.state = HSReady

	h.log.Info("authentication successful",
		"device", resp.DeviceID,
		"session_expires", expiresAt,
	)
	return est, nil
}

// HandleAccept processes a CapabilityAccept (client-side).
func (h *Handshake) HandleAccept(accept *CapabilityAccept) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.result = &HandshakeResult{
		DeviceID:     accept.DeviceID,
		DeviceName:   accept.DeviceName,
		Platform:     accept.Platform,
		Version:      accept.Version,
		Capabilities: accept.Capabilities,
		SessionToken: accept.SessionToken,
		IsNewSession: false,
	}
	h.state = HSReady

	h.log.Info("handshake accepted",
		"device", accept.DeviceID,
		"version", accept.Version,
	)
}

// HandleSessionEstablished processes a SessionEstablished message (client-side
// after sending auth response).
func (h *Handshake) HandleSessionEstablished(est *SessionEstablished) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.result = &HandshakeResult{
		DeviceID:     est.DeviceID,
		DeviceName:   est.DeviceName,
		Capabilities: est.Capabilities,
		Version:      est.Version,
		SessionToken: est.SessionToken,
		IsNewSession: true,
	}
	h.state = HSReady

	h.log.Info("session established",
		"device", est.DeviceID,
		"version", est.Version,
	)
}

// generateSessionToken creates a signed session token for the given device.
func (h *Handshake) generateSessionToken(deviceID string) string {
	// Token format: base64(device_id + ":" + expires + ":" + hmac)
	expires := time.Now().Add(h.config.TokenDuration).Unix()
	payload := fmt.Sprintf("%s:%d", deviceID, expires)
	mac := hmacSHA256(h.config.TokenSecret, payload)
	token := base64.RawURLEncoding.EncodeToString([]byte(payload + ":" + hex.EncodeToString(mac)))
	return "sess_" + token
}

// validateSessionToken checks a session token and returns the device ID
// if valid.
func (h *Handshake) validateSessionToken(token string) (string, bool) {
	if len(token) < 6 || token[:5] != "sess_" {
		return "", false
	}
	raw := token[5:]
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return "", false
	}

	parts := stringsSplitN(string(data), ":", 3)
	if len(parts) != 3 {
		return "", false
	}

	deviceID := parts[0]
	expiresStr := parts[1]
	macHex := parts[2]

	// Check expiration.
	expires, err := parseInt64(expiresStr)
	if err != nil || time.Now().Unix() > expires {
		return "", false
	}

	// Verify HMAC.
	expectedMAC := hmacSHA256(h.config.TokenSecret, parts[0]+":"+parts[1])
	expectedHex := hex.EncodeToString(expectedMAC)
	if !hmac.Equal([]byte(macHex), []byte(expectedHex)) {
		return "", false
	}

	return deviceID, true
}

// NotifyCapabilityUpdate sends a capability update event to the peer.
func NotifyCapabilityUpdate(caps []string) *Envelope {
	evt, _ := NewEvent(MethodCapabilityUpdate, map[string]interface{}{
		"caps": caps,
	})
	return evt
}

// negotiateVersion finds the highest mutually supported version, or -1 if
// the ranges do not overlap.
func negotiateVersion(peerMin, peerMax, localMin, localMax int) int {
	low := peerMin
	if localMin > low {
		low = localMin
	}
	high := peerMax
	if localMax < high {
		high = localMax
	}
	if low > high {
		return -1
	}
	return high
}

// intersectCapabilities returns the intersection of two capability slices.
func intersectCapabilities(a, b []string) []string {
	set := make(map[string]bool)
	for _, c := range b {
		set[c] = true
	}
	var result []string
	for _, c := range a {
		if set[c] {
			result = append(result, c)
		}
	}
	return result
}

// generatePIN creates a random 6-digit numeric PIN.
func generatePIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		// Fallback: use a timestamp-based PIN (not cryptographically secure,
		// but functional).
		return fmt.Sprintf("%06d", time.Now().UnixNano()%900000+100000)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// hmacSHA256 computes an HMAC-SHA256.
func hmacSHA256(secret, data string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// stringsSplitN is a small helper to split strings.
func stringsSplitN(s, sep string, n int) []string {
	// Simple implementation to avoid importing strings in this file.
	// In practice, you'd use the strings package.
	result := make([]string, 0, n)
	start := 0
	for i := 0; i < n-1 && start < len(s); i++ {
		idx := indexOf(s, sep, start)
		if idx < 0 {
			break
		}
		result = append(result, s[start:idx])
		start = idx + len(sep)
	}
	result = append(result, s[start:])
	return result
}

func indexOf(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		n = n*10 + int64(c-'0')
	}
	return n, nil
}
