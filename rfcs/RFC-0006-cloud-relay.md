# RFC-0006: Cloud Relay

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | RFC-0001, RFC-0002, RFC-0003 |

## Summary

Define the Cloud Relay service — the minimal cloud component that enables remote access to BuzzPi devices. The Cloud Relay is a packet-forwarding proxy with identity registration, nothing more. It never decrypts traffic and is designed to be self-hostable.

## Motivation

BuzzPi's core promise is "never type an IP address." On the local network, mDNS makes this work. But remotely — across the internet — devices are behind NATs, firewalls, and CGNAT. The Cloud Relay solves this by providing:

1. **NAT traversal** — a public rendezvous point for devices and clients
2. **Device registry** — maps device IDs to active relay sessions
3. **Account association** — users claim devices via pairing

The relay must be minimal because:
- The fewer features, the fewer vulnerabilities
- Users should be able to self-host without operational burden
- The relay should never be a bottleneck or a single point of failure

## Design

### 1. Architecture

```
┌──────────┐     ┌──────────────────────────────┐     ┌──────────┐
│  Client  │────▶│        Cloud Relay            │◀────│  Runtime │
│  (Phone) │     │                              │     │   (Pi)   │
│          │     │  ┌────────┐  ┌────────────┐  │     │          │
│ WSS──────┼────▶│  │ Packet │  │  Registry  │  │◀────┼──────WSS │
│          │     │  │ Relay  │  │  Service   │  │     │          │
│          │     │  └────────┘  └────────────┘  │     │          │
│          │     │  ┌────────┐  ┌────────────┐  │     │          │
│          │     │  │ Auth   │  │  STUN/TURN │  │     │          │
│          │     │  │ Service│  │  (optional) │  │     │          │
│          │     │  └────────┘  └────────────┘  │     │          │
│          │     └──────────────────────────────┘     │          │
└──────────┘                                          └──────────┘
```

### 2. Core Components

#### Packet Relay

The Packet Relay is a WebSocket proxy. It:
- Accepts WebSocket connections from clients and runtimes
- Authenticates connections via session tokens
- Maintains a mapping of `device_id → WebSocket connection`
- Forwards packets bidirectionally between clients and runtimes

**Connection flow:**

```
Runtime                     Relay                    Client
  │                          │                        │
  │── WSS connect ──────────▶│                        │
  │   /ws/device             │                        │
  │   Header: Bearer <token> │                        │
  │                          │                        │
  │◀─ WSS accepted ────────│                        │
  │                          │                        │
  │                          │◀── WSS connect ────────│
  │                          │   /ws/client            │
  │                          │   Header: Bearer <token>│
  │                          │                        │
  │                          │── WSS accepted ───────▶│
  │                          │                        │
  │── BPP message ──────────▶│── BPP message ────────▶│
  │                          │                        │
  │◀─ BPP message ──────────│◀── BPP message ────────│
```

#### Registry Service

The Registry tracks device presence and account associations.

**Database schema (PostgreSQL):**

```sql
CREATE TABLE devices (
    device_id    VARCHAR(64) PRIMARY KEY,
    friendly_name VARCHAR(128),
    platform     VARCHAR(64),
    runtime_version VARCHAR(16),
    public_key   TEXT,                     -- Ed25519 public key
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ,
    relay_session_id VARCHAR(128)          -- NULL if offline
);

CREATE TABLE accounts (
    account_id   VARCHAR(64) PRIMARY KEY,
    email        VARCHAR(256) UNIQUE,
    display_name VARCHAR(128),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE pairings (
    device_id    VARCHAR(64) REFERENCES devices(device_id),
    account_id   VARCHAR(64) REFERENCES accounts(account_id),
    client_id    VARCHAR(64),
    client_name  VARCHAR(128),
    role         VARCHAR(16) NOT NULL DEFAULT 'member',
    paired_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (device_id, account_id, client_id)
);

CREATE TABLE sessions (
    session_id   VARCHAR(128) PRIMARY KEY,
    device_id    VARCHAR(64) REFERENCES devices(device_id),
    account_id   VARCHAR(64) REFERENCES accounts(account_id),
    client_id    VARCHAR(64),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ
);
```

#### Auth Service

Handles account authentication. Uses OAuth 2.0 with external providers:

| Provider | Purpose |
|----------|---------|
| Google | Primary — broadest reach |
| GitHub | Developer-focused |
| Email + magic link | Privacy-oriented |

**No password-based auth.** BuzzPi uses OAuth or magic links only.

#### STUN/TURN Server (Optional)

- **STUN** — Built into the Connection Engine (RFC-0001). The relay can optionally host a STUN endpoint to help with NAT reflection.
- **TURN** — Required for symmetric NAT traversal. The relay can integrate coturn as a sub-process. TURN relay is bandwidth-intensive — operators should be aware.

### 3. Packet Forwarding

```
                           ┌──────────────────┐
  Client WebSocket ───────▶│   Packet Relay   │◀─────── Runtime WebSocket
                           │                  │
Id: cli_wss_abc            │  Routing Table   │       Id: dev_wss_xyz
                           │                  │
                           │ dev_a1b2 → wss_1 │
                           │ dev_c3d4 → wss_2 │
                           └──────────────────┘
```

**Forwarding rules:**
1. Runtime sends packet → Relay reads `device_id` from connection metadata → forwards to all connected clients with matching pairing
2. Client sends packet → Relay reads `target_device_id` from the BPP envelope → forwards to the runtime WebSocket
3. If target device is offline, Relay returns error: `device.offline`

**Packet types (transparently forwarded):**
- All BPP requests and responses
- All BPP events (screen frames, terminal output, notifications)
- Connection Engine signaling (ICE candidates, SDP offers)

The relay does NOT inspect packet contents. It reads only the routing fields.

### 4. Registration Flow

**Device registration (on pairing):**

```
POST /api/v1/devices/register
Authorization: Bearer <device-token>

{
  "device_id": "dev_a1b2c3d4",
  "friendly_name": "living-room-pi",
  "platform": "raspberry-pi/5",
  "runtime_version": "0.1.0",
  "public_key": "base64-ed25519-public-key"
}
```

**Device heartbeat (every 60 seconds):**

```
PUT /api/v1/devices/{device_id}/heartbeat
```

If the relay does not receive a heartbeat for 180 seconds, the device is marked offline.

**Device presence check:**

```
GET /api/v1/devices/{device_id}/presence

200 OK
{
  "online": true,
  "relay_session": "sess_abc123",
  "last_seen": "2026-07-07T12:00:00Z"
}
```

### 5. Client → Device Routing

**Client connects and sends message:**

```
1. Client connects to /ws/client, provides session token
2. Client sends BPP envelope with "device_id" in headers:
   {
     "v": 1,
     "type": "request",
     "method": "device.stats",
     "device_id": "dev_a1b2c3d4"
   }
3. Relay:
   a. Validates session token
   b. Looks up device_id in registry
   c. Checks client account is paired to device
   d. Looks up device WebSocket connection
   e. Forwards envelope to device
4. Device responds → relay forwards to client
```

### 6. Graceful Disconnection

```
Runtime                    Relay                     Client
  │── disconnect              │                        │
  │   (close frame)           │                        │
  │                          │── device.offline ──────▶│
  │                          │   {device_id: "dev_..."}│
  │                          │                        │
  │                          │── cleanup ─────────────▶│
  │                          │   registry: offline     │
```

### 7. API Endpoints

#### Health & Management

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (200 OK) |
| GET | `/metrics` | Prometheus metrics |
| GET | `/api/v1/stats` | Server stats (connections, devices) |

#### Device API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/devices/register` | Device token | Register a device |
| PUT | `/api/v1/devices/{id}/heartbeat` | Device token | Update device heartbeat |
| GET | `/api/v1/devices/{id}/presence` | Session | Check device online status |
| GET | `/api/v1/account/devices` | Session | List user's paired devices |

#### Auth API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/oauth/{provider}` | Initiate OAuth login |
| POST | `/api/v1/auth/magic-link` | Request magic link email |
| POST | `/api/v1/auth/verify` | Verify OAuth callback / magic link token |
| POST | `/api/v1/auth/refresh` | Refresh session token |
| POST | `/api/v1/auth/revoke` | Revoke session |

#### WebSocket

| Path | Auth | Description |
|------|------|-------------|
| `/ws/device` | Device token | Device WebSocket connection |
| `/ws/client` | Session token | Client WebSocket connection |

### 8. Self-Hosting

The Cloud Relay is designed to be self-hostable. A single binary with embedded SQLite or optional PostgreSQL.

```bash
# Start the relay
buzzpi-relay --config relay.yaml

# relay.yaml
host: "0.0.0.0"
port: 443
tls_cert: /etc/letsencrypt/live/relay.example.com/fullchain.pem
tls_key: /etc/letsencrypt/live/relay.example.com/privkey.pem
database:
  type: sqlite          # or postgres
  path: /var/lib/buzzpi-relay/data.db
turn:
  enabled: false        # coturn integration, optional
```

**Runtime configuration for custom relay:**

```yaml
# /etc/buzzpi/runtime.yaml
network:
  relay_servers:
    - wss://relay.example.com:443
```

### 9. Rate Limiting

| Limit | Per | Value |
|-------|-----|-------|
| Connections per device | Device | 1 |
| Connections per account | Account | 10 |
| Connections per IP | IP | 20 |
| Messages per connection | Second | 1000 |
| Device registrations per hour | IP | 10 |
| Auth attempts per minute | Account | 5 |
| Heartbeat interval | Device | 30-180s |

### 10. Operational Requirements

**buzzpi.cloud (reference hosting):**

| Resource | Target |
|----------|--------|
| Instance type | 2 vCPU, 4 GB RAM |
| Storage | 50 GB SSD |
| Network | 1 Gbps |
| Monthly bandwidth | 10 TB (per 1000 active devices) |
| Database | PostgreSQL 16, 2 vCPU, 4 GB RAM |

**Scaling:** The relay is horizontally scalable behind a load balancer. Device affinity is maintained via consistent hashing on `device_id`. Shared state goes through the database.

---

## Drawbacks

1. **Relay is a bottleneck** — All remote traffic passes through the relay. For screen streaming, this means bandwidth costs are proportional to usage. Mitigation: direct P2P connections are preferred (ICE); the relay is fallback for NAT traversal failures.

2. **Self-hosted relay adds complexity** — Users who self-host must manage TLS certificates, DNS, and uptime. Mitigation: the reference relay is a single binary; deployment docs cover Docker Compose and bare metal.

3. **Relay operator trust** — Users must trust the relay operator not to tamper with packets. Mitigation: end-to-end encryption (Relay only sees encrypted WebSocket payloads, not decrypted BPP messages).

---

## Rationale

1. **Why WebSocket proxy over custom protocol?** WebSocket is universally supported, works through enterprise proxies, and is simple to implement. The relay is a router, not a protocol processor.

2. **Why TURN is optional?** TURN relay is bandwidth-intensive (relays all media traffic). Most NATs can be traversed with STUN + ICE. TURN is needed only for symmetric NATs (rare on mobile networks). The relay operator may choose to deploy coturn separately.

3. **Why SQLite in self-hosted mode?** Self-hosted deployments typically serve a single user or small team. SQLite eliminates the PostgreSQL dependency. PostgreSQL is used only for the reference multi-tenant relay.

---

## Prior Art

- **Tailscale DERP servers** — WebSocket relay with STUN, automatic region selection. Inspires our relay architecture.
- **Cloudflare Argo Tunnel** — Reverse tunnel from device to edge. Inspires our device→relay persistent connection model.
- **Matrix Homeserver** — Federated relay with end-to-end encryption. Inspires our forward-only, no-decrypt relay design.

---

## Unresolved Questions

1. **Relay region selection** — Should clients automatically select the nearest relay (like Tailscale DERP), or use a single configured relay? Leaning toward automatic selection for v0.5+; manual configuration for v0.1.

2. **Relay federation** — Should BuzzPi support federated relays (like Mastodon/Matrix)? This would allow different relay operators to interoperate. Deferred to v1.0 — adds significant complexity.

3. **Bandwidth management** — How should the relay handle bandwidth contention between concurrent screen streams? Options: fair queuing per-device, per-account bandwidth caps, or best-effort. Leaning toward best-effort with optional caps.

---

## Implementation Plan

| Phase | Milestone | Details |
|-------|-----------|---------|
| P0 | Device relay | Basic WebSocket proxy, device↔client forwarding |
| P1 | Registry | Device registration, heartbeat, presence API |
| P2 | Auth | OAuth integration (Google, GitHub), session management |
| P3 | STUN/TURN | coturn integration, ICE helper endpoints |
| P4 | Self-host | Single-binary mode with SQLite, Docker Compose |
| P5 | Scaling | Horizontal scaling, load balancer, consistent hashing |

---

## References

- RFC-0001: Connection Engine (ICE, STUN, TURN, NAT traversal)
- RFC-0002: Runtime Architecture (device identity, registration)
- RFC-0003: Pairing Protocol (account association)
- Engineering Book: backend-architecture.md, buzzai.md
- Reference: rest-endpoints.md, configuration-reference.md
