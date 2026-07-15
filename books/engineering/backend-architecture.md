# Backend Architecture

**Cloud services for discovery, relay, and device management.** The BuzzPi backend enables users to connect to devices across the internet — not just on their local network. It is designed as a minimal, cost-efficient set of services that augment (never gate) local connectivity.

---

## Principles

1. **Backend is optional** — Devices on the same LAN work fully without any cloud service. The backend only activates for WAN connectivity and cross-network features.
2. **Single-service start** — The entire backend is a single Go binary with PostgreSQL. No microservices, no container orchestration, no Redis in the MVP.
3. **Relay is the expensive operation** — WebRTC P2P is strongly preferred. The TURN relay consumes bandwidth and should be avoided when direct connectivity is possible.
4. **No user data at rest beyond what's required** — The backend stores device metadata, user accounts, and pairing information. It does not store session contents, keystrokes, or screen data.
5. **Stateless where possible** — The relay signaling path is stateless. Device registry is the only stateful component.

---

## Services

```
                ┌─────────────────────────────────────────┐
                │         BuzzPi Cloud (jphat.net/buzzpi)     │
                │                                          │
                │  ┌───────────┐  ┌───────────┐           │
                │  │  Registry │  │  Signaling │           │
                │  │  Service  │  │  Relay     │           │
                │  └─────┬─────┘  └──────┬─────┘           │
                │        │               │                 │
                │  ┌─────▼─────┐  ┌──────▼─────┐           │
                │  │ PostgreSQL │  │  TURN/STUN │           │
                │  └───────────┘  │  (coturn)   │           │
                │                 └────────────┘            │
                └──────────────────────────────────────────┘
                          ▲                ▲
                          │                │
                ┌─────────┴──────┐  ┌──────┴────────┐
                │  Android App   │  │ BuzzPi Runtime │
                └────────────────┘  └───────────────┘
```

### 1. Registry Service

**Purpose:** Device directory, user accounts, pairing management.

```
POST   /v1/auth/signup          # Create account
POST   /v1/auth/login           # Authenticate
POST   /v1/auth/refresh         # Refresh token

POST   /v1/devices              # Register device (claim token)
GET    /v1/devices              # List user's devices
GET    /v1/devices/:id          # Device details
PATCH  /v1/devices/:id          # Update device metadata
DELETE /v1/devices/:id          # Unpair device

POST   /v1/devices/:id/pair     # Initiate pairing
POST   /v1/devices/:id/unpair   # Remove pairing
GET    /v1/devices/:id/status   # Online/offline + last seen

GET    /v1/discover             # User's devices (with online state)
```

### 2. Signaling Relay

**Purpose:** WebSocket-based signaling to establish WebRTC connections between clients and devices.

```
WebSocket /v1/relay              # Upgrade to signaling session
```

The relay does not process or store message contents. It multiplexes signaling messages and ICE candidates between authenticated peers.

### 3. TURN/STUN Server

**Purpose:** NAT traversal fallback when direct P2P fails. Uses [coturn](https://github.com/coturn/coturn) as a subprocess or sidecar.

- STUN: `stun:jphat.net/buzzpi/stun:3478`
- TURN: `turn:jphat.net/buzzpi/turn:3478` (TCP/UDP)
- TURN credentials: time-limited (REST API authentication)

---

## Registry Service Detail

### Authentication

```go
type AuthHandler struct {
    jwtSecret    []byte
    sessionStore SessionStore
}

func (h *AuthHandler) Signup(ctx context.Context, req SignupRequest) (*AuthResult, error) {
    // Validate email
    // Hash password with bcrypt (cost 12)
    // Create user record in PostgreSQL
    // Generate device claim token
    // Return access + refresh tokens
}

func (h *AuthHandler) Authenticate(ctx context.Context, req LoginRequest) (*AuthResult, error) {
    // Find user by email
    // Verify bcrypt hash
    // Check rate limit (5 attempts per minute)
    // Generate JWT (access: 15min, refresh: 7 days)
    // Return tokens
}
```

### JWT Claims

```go
type AccessClaims struct {
    jwt.RegisteredClaims
    UserID   string   `json:"uid"`
    Devices  []string `json:"dev,omitempty"` // Device IDs this token can access
    Scope    string   `json:"scope"`          // "user" | "device" | "relay"
}
```

### Device Registration Flow

```
Runtime                           Registry
  │                                  │
  │  1. generate keypair              │
  │  2. POST /v1/devices              │
  │     {                              │
  │       claim_token: "xxxx",        │  (from setup/onboarding)
  │       identity_key: "pubkey...",  │
  │       friendly_name: "Kitchen Pi",│
  │       model: "Raspberry Pi 5",    │
  │       runtime_version: "0.1.0"    │
  │     }                             │
  │ ────────────────────────────────► │
  │                                  │
  │                                  │── Validate claim token
  │                                  │── Register device
  │                                  │── Generate device_id
  │                                  │
  │  3. {device_id, pairing_code}    │
  │ ◄──────────────────────────────── │
  │                                  │
  │  4. display QR with pairing_code  │
  │     (on screen or blink LED)     │
```

---

## Signaling Relay Detail

```go
// RelayServer handles WebSocket signaling between clients and devices.
type RelayServer struct {
    sessions sync.Map // map[string]*RelaySession
    upgrader websocket.Upgrader
}

type RelaySession struct {
    ID       string
    Role     string          // "client" or "device"
    DeviceID string
    Conn     *websocket.Conn
    mu       sync.Mutex
}

// HandleConnection upgrades HTTP to WebSocket and manages the session.
func (s *RelayServer) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // Authenticate via JWT in query param or header
    claims, err := s.authenticate(r)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }

    session := &RelaySession{
        ID:       generateID(),
        Role:     claims.Scope,  // "client" or "device"
        DeviceID: claims.DeviceID,
        Conn:     conn,
    }

    s.sessions.Store(session.ID, session)
    defer s.sessions.Delete(session.ID)

    // Read loop: forward signaling messages to the paired peer
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        s.forward(session, msg)
    }
}

func (s *RelayServer) forward(from *RelaySession, msg []byte) {
    var envelope struct {
        TargetID string `json:"target_id"`
        Type     string `json:"type"` // "sdp_offer" | "sdp_answer" | "ice_candidate" | "heartbeat"
        Data     json.RawMessage `json:"data"`
    }
    json.Unmarshal(msg, &envelope)

    // Find target session by device or session ID
    s.sessions.Range(func(key, value interface{}) bool {
        target := value.(*RelaySession)
        if target.DeviceID == envelope.TargetID && target.ID != from.ID {
            target.mu.Lock()
            target.Conn.WriteJSON(envelope)
            target.mu.Unlock()
            return false
        }
        return true
    })
}
```

### Heartbeat & Liveness

```go
const (
    HeartbeatInterval = 30 * time.Second
    SessionTimeout    = 90 * time.Second // 3 missed heartbeats
)

type RelaySession struct {
    // ...
    lastHeartbeat time.Time
}

func (s *RelayServer) heartbeatLoop() {
    ticker := time.NewTicker(HeartbeatInterval)
    for range ticker.C {
        s.sessions.Range(func(key, value interface{}) bool {
            session := value.(*RelaySession)
            if time.Since(session.lastHeartbeat) > SessionTimeout {
                session.Conn.Close()
                s.sessions.Delete(key)
            }
            return true
        })
    }
}
```

---

## TURN Credential Management

```go
// GenerateTURNCredentials creates time-limited TURN credentials.
// coturn supports REST authentication via HMAC-SHA1.
func GenerateTURNCredentials(secret []byte, username string, ttl time.Duration) (string, string) {
    timestamp := time.Now().Add(ttl).Unix()
    user := fmt.Sprintf("%d:%s", timestamp, username)
    mac := hmac.New(sha1.New, secret)
    mac.Write([]byte(user))
    cred := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    return user, cred
}

// VerifyTURNCredential validates that a TURN credential is still valid.
func VerifyTURNCredential(secret []byte, user, cred string) bool {
    parts := strings.SplitN(user, ":", 2)
    if len(parts) != 2 {
        return false
    }
    timestamp, err := strconv.ParseInt(parts[0], 10, 64)
    if err != nil || time.Now().Unix() > timestamp {
        return false
    }
    expected, _ := GenerateTURNCredentials(secret, parts[1], 0)
    return hmac.Equal([]byte(expected), []byte(cred))
}
```

---

## Database Schema

```sql
-- Users
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,           -- bcrypt hash
    display_name TEXT NOT NULL DEFAULT '',
    plan        TEXT NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Devices
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    friendly_name   TEXT NOT NULL,
    model           TEXT NOT NULL DEFAULT '',
    identity_key    TEXT NOT NULL UNIQUE,    -- Ed25519 public key
    runtime_version TEXT NOT NULL DEFAULT '',
    claim_token     TEXT NOT NULL,           -- Token used during registration
    owner_id        UUID REFERENCES users(id),
    state           TEXT NOT NULL DEFAULT 'paired',  -- paired | unpaired | revoked
    last_seen_at    TIMESTAMPTZ,
    last_ip         TEXT,
    paired_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_owner ON devices(owner_id);
CREATE INDEX idx_devices_identity_key ON devices(identity_key);
CREATE UNIQUE INDEX idx_devices_claim_token ON devices(claim_token) WHERE state = 'paired';

-- Pairing codes (ephemeral, for initial device-client pairing)
CREATE TABLE pairing_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID NOT NULL REFERENCES devices(id),
    code        TEXT NOT NULL,              -- 6-character alphanumeric
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pairing_codes_code ON pairing_codes(code);
CREATE INDEX idx_pairing_codes_device ON pairing_codes(device_id);

-- Sessions (active connections via relay)
CREATE TABLE sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID NOT NULL REFERENCES devices(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    transport   TEXT NOT NULL,             -- "direct" | "relay"
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at    TIMESTAMPTZ,
    bytes_sent  BIGINT DEFAULT 0,
    bytes_recv  BIGINT DEFAULT 0
);

CREATE INDEX idx_sessions_active ON sessions(device_id) WHERE ended_at IS NULL;
CREATE INDEX idx_sessions_device ON sessions(device_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    token_hash  TEXT NOT NULL,             -- SHA-256 of refresh token
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
```

---

## Configuration

```yaml
# backend/config.yaml
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 10s
  write_timeout: 10s

database:
  url: "postgres://buzzpi@localhost/buzzpi"
  max_conns: 25
  min_conns: 5

auth:
  jwt_secret: "${JWT_SECRET}"          # Environment variable
  access_token_ttl: 15m
  refresh_token_ttl: 720h              # 30 days

relay:
  max_message_size: 65536              # 64KB max signaling message
  heartbeat_interval: 30s
  session_timeout: 90s

turn:
  enabled: true
  realm: jphat.net
  secret: "${TURN_SECRET}"             # HMAC secret for coturn REST auth
  credential_ttl: 24h
  coturn_binary: /usr/bin/coturn
  coturn_config: /etc/buzzpi/coturn.conf

rate_limit:
  signup: 3/hour                       # Max signups per IP per hour
  login: 5/minute                      # Max login attempts
  device_register: 10/hour            # Max device registrations
```

---

## Deployment

### MVP (Single Machine)

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /buzzpi-backend ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache coturn ca-certificates
COPY --from=build /buzzpi-backend /usr/local/bin/
COPY config.yaml /etc/buzzpi/config.yaml
EXPOSE 8080 3478
ENTRYPOINT ["buzzpi-backend"]
```

### Infrastructure Requirements

| Component | Requirement |
|-----------|-------------|
| Server | 2 vCPU, 4GB RAM |
| Storage | 50GB SSD |
| Bandwidth | 1 Gbps (for relay traffic) |
| PostgreSQL | Managed (RDS, Cloud SQL, or similar) |

The TURN relay is the bandwidth bottleneck. Cost estimate:

| Concurrent Streams | Bandwidth | Monthly Egress (est.) |
|--------------------|-----------|----------------------|
| 10 | 20 Mbps | ~1.5 TB |
| 100 | 200 Mbps | ~15 TB |
| 1000 | 2 Gbps | ~150 TB |

For MVP, we expect <10 concurrent relay streams. P2P should succeed for >80% of connections.

---

## Security

- All API endpoints require TLS (HTTPS/WSS)
- JWT access tokens expire after 15 minutes
- Refresh tokens are stored as SHA-256 hashes
- Passwords hashed with bcrypt (cost 12)
- TURN credentials are time-limited to 24 hours
- Rate limiting on auth endpoints
- Relay signaling messages are not logged or stored
- Session data (keystrokes, screen content) never passes through the backend when P2P is established

---

## Testing Strategy

| Test Type | Scope | Tooling |
|-----------|-------|---------|
| Unit | Auth handler, credential generation, rate limiting | Go testing + testify |
| Unit | Database queries (against test PostgreSQL) | testcontainers-go |
| Integration | Full API round-trip (signup → register device → list) | testcontainers-go |
| Integration | WebSocket relay forward between two peers | Go test with goroutines |
| Load | Concurrent relay connections (1000 virtual peers) | k6 |
| Load | TURN relay bandwidth test | iperf3 |

### Success Criteria

| Criterion | Target |
|-----------|--------|
| API response time (P95) | <200ms |
| Relay message latency (P95) | <50ms |
| Concurrent relay sessions | >1000 |
| TURN bandwidth per session | >5 Mbps |
| Device registration | <2s |
| Database migration | <30s |
| Cold start to serving | <10s |
