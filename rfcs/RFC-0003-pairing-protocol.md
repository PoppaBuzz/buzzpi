# RFC-0003: Pairing Protocol

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | RFC-0001, RFC-0002 |

## Summary

Define the protocol by which a client (Android app, CLI, desktop) discovers, pairs with, and establishes trust with a BuzzPi Runtime. The pairing protocol is the user's first interaction with any device — it must be secure, simple, and work on LAN and remotely.

## Motivation

The entire BuzzPi thesis is that the user should never type an IP address. Pairing is the mechanism that makes this possible. The user taps a device, authenticates, and from that moment forward the device is theirs to manage — no networking knowledge required.

A poorly designed pairing flow destroys the core value proposition. It must be:
1. **Zero-config**: no IPs, no ports, no SSH keys
2. **Secure**: resistant to MITM, replay, and pairing-jacking
3. **Offline-capable**: LAN pairing must work without internet
4. **Portable**: the user's identity follows them across clients

## Design

### 1. Pairing Overview

```
┌─────────┐           ┌─────────┐           ┌──────────┐
│ Client  │           │ Runtime │           │ Cloud    │
│ (Phone) │           │ (Pi)    │           │ Relay    │
└────┬────┘           └────┬────┘           └────┬─────┘
     │                     │                     │
     │  1. Discovery       │                     │
     │  (mDNS or relay)    │                     │
     │◀────────────────────│                     │
     │                     │                     │
     │  2. Pairing Init    │                     │
     │────────────────────▶│                     │
     │                     │                     │
     │  3. Pairing Verify  │                     │
     │◀────────────────────│                     │
     │   (code comparison) │                     │
     │────────────────────▶│                     │
     │                     │                     │
     │  4. Key Exchange    │                     │
     │◀═══════════════════▶│                     │
     │  (SPAKE2+ PAKE)     │                     │
     │                     │                     │
     │  5. Cloud Assn.     │                     │
     │──────────────────────────────────────────▶│
     │                     │                     │
     │  6. Paired ✓        │                     │
     │                     │                     │
```

### 2. Discovery (Pre-Pairing)

Before pairing, the client must discover nearby devices.

#### LAN Discovery (mDNS)

The Runtime advertises via `_buzzpi._tcp` with TXT records containing:
- `device_id` — stable device identifier
- `friendly_name` — human-readable name
- `runtime_version` — semver
- `platform` — hardware identifier

The client listens on the mDNS multicast group and displays discovered devices.

#### Remote Discovery (Cloud Relay)

If a device was previously paired to an account, it registers with the Cloud Relay:
```
Runtime → Cloud: PUT /api/v1/device/{device_id}/presence
                 {"online": true, "relay_session": "sess_..."}
```

The client fetches the user's paired devices:
```
Client → Cloud: GET /api/v1/account/devices
Cloud → Client: [{"device_id":"dev_...","friendly_name":"...","online":true,...}]
```

Remote discovery only works after the initial LAN pairing.

#### Discovery UI

```
┌────────────────────────────┐
│  Available Devices         │
│                            │
│  🔵 Living Room Pi    LAN  │  ← mDNS found
│  🔵 Garage Pi         LAN  │  ← mDNS found
│  ──────────────            │
│  🔵 Office Pi        RMT   │  ← Cloud relay (previously paired)
│  🔵 Lab Pi            RMT  │  ← Cloud relay
│                            │
│  [Searching...]      3s    │
└────────────────────────────┘
```

### 3. Pairing Flow (LAN)

#### Step 1: Initiation

Client connects to Runtime via WebSocket (direct IP:port from mDNS, or via relay).

```
Client → Runtime:
{
  "v": 1,
  "id": "msg_abc123",
  "ts": "2026-07-07T12:00:00Z",
  "type": "request",
  "method": "pair.initiate",
  "remote": {
    "device_id": "dev_a1b2c3d4",
    "client_name": "Pixel 9 Pro",
    "supported_methods": ["spake2p", "pin"]
  }
}
```

The client sends its `supported_methods` — the Runtime selects the most secure mutually-supported method.

#### Step 2: Challenge

```
Runtime → Client:
{
  "v": 1,
  "id": "msg_def456",
  "ts": "2026-07-07T12:00:00Z",
  "type": "response",
  "rid": "req_abc123",
  "result": {
    "method": "pin",
    "challenge": {
      "pin_length": 6,
      "nonce": "a1b2c3d4e5f6..."
    },
    "session_id": "pair_sess_xyz789"
  }
}
```

Two verification methods:

**Method A: PIN (Default)**
- Runtime generates a 6-digit PIN (0-9, unambiguous: no 1/I, 0/O)
- PIN is displayed on the device screen (or spoken via TTS on headless)
- User enters PIN on the client

**Method B: Code Comparison (Both have screens)**
- Runtime displays a 4-word verification phrase (from BIP-39 wordlist)
- Client displays a different 4-word phrase
- User confirms they match (numeric comparison of the two phrases)

#### Step 3: Verification

```
Client → Runtime:
{
  "v": 1,
  "id": "msg_ghi789",
  "ts": "2026-07-07T12:00:01Z",
  "type": "request",
  "method": "pair.verify",
  "remote": {
    "session_id": "pair_sess_xyz789",
    "pin": "482916",
    "client_public_key": "base64-ed25519-public-key"
  }
}
```

#### Step 4: Key Exchange

The Runtime verifies the PIN. If correct:

```
Runtime → Client:
{
  "v": 1,
  "type": "response",
  "rid": "req_ghi789",
  "result": {
    "server_public_key": "base64-ed25519-public-key",
    "session_token": "sess_...",
    "session_expiry": "2026-07-08T12:00:00Z"
  }
}
```

Both sides now have each other's Ed25519 public keys. Future communication is authenticated by signing BPP envelopes.

```
// Shared secret derivation (both sides independently compute):
shared = X25519(client_sk, server_pk)   // client side
shared = X25519(server_sk, client_pk)   // server side
```

#### Step 5: Cloud Association

The Runtime registers the pairing with the Cloud Relay:

```
Runtime → Cloud: POST /api/v1/pair
                 {"device_id": "dev_...",
                  "account_id": "acct_...",  // from client
                  "client_id": "cli_...",
                  "client_name": "Pixel 9 Pro",
                  "paired_at": "2026-07-07T12:00:00Z"}
```

This enables remote access: the Cloud Relay knows the device is paired to this account.

### 4. Pairing Flow (Remote / Headless)

For headless devices (no display) or remote pairing:

**Method: Claim Code**

1. User generates a claim code on the client:
   ```
   Claim Code = base62(random 8 bytes) → "K7x9mP2q"
   ```

2. Client sends claim to Cloud Relay:
   ```
   Client → Cloud: POST /api/v1/pair/claim
                   {"claim_code": "K7x9mP2q", "device_id": "dev_..."}
   ```

3. Cloud Relay generates a short-lived pairing token and sends it to the Runtime (via existing WebSocket):
   ```
   Cloud → Runtime: {"type": "pair.claim", "token": "...", "account_id": "acct_..."}
   ```

4. Runtime accepts the pairing:
   ```
   Runtime → Cloud: {"type": "pair.confirm", "device_id": "dev_...", "token": "..."}
   ```

5. Cloud notifies the client:
   ```
   Cloud → Client: {"type": "pair.complete", "device_id": "dev_...", "session_token": "sess_..."}
   ```

**Security:** Claim codes expire after 5 minutes. Single-use only. Rate-limited to 3 attempts per code.

### 5. Session Management

After pairing, the client receives a session token.

```
type Session struct {
    Token       string    // 32-byte random, base62
    DeviceID    string    // paired device
    ClientID    string    // client identity
    AccountID   string    // cloud account (optional)
    Role        Role      // admin / member / viewer
    CreatedAt   time.Time
    ExpiresAt   time.Time
    LastUsedAt  time.Time
}

type Role int
const (
    RoleViewer Role = iota
    RoleMember
    RoleAdmin
)
```

**Token lifecycle:**
- Initial expiry: 24 hours
- Refreshed on each authenticated request (sliding window, max 7 days)
- Client can explicitly revoke: `pair.revoke` method
- Runtime can revoke all sessions: `pair.revoke_all` (factory reset)

### 6. Unpairing

```
Client → Runtime:
{
  "type": "request",
  "method": "pair.unpair",
  "remote": {"session_token": "sess_..."}
}
```

Runtime deletes the session, removes the client from its paired table, and sends an acknowledgment. The Cloud Relay is notified and removes the device from the user's account view.

**Factory reset:**
```
Runtime → self: delete /var/lib/buzzpi/identity/*
```
Generates a new identity on next boot. All pairings are lost.

### 7. Multi-Device, Multi-Client

**One device, multiple clients:**
- The Runtime maintains a table of paired clients (up to 10 by default)
- Each client has its own session token and role
- Admin clients can revoke other clients

**One client, multiple devices:**
- The Cloud Relay maintains the device list per account
- The client fetches and caches session tokens per device
- The client display shows all paired devices (LAN and remote)

### 8. BPP Methods Added

| Method | Direction | Description |
|--------|-----------|-------------|
| `pair.initiate` | Client→Runtime | Start pairing, send capabilities |
| `pair.verify` | Client→Runtime | Verify PIN/code, exchange keys |
| `pair.status` | Client→Runtime | Check pairing state |
| `pair.unpair` | Client→Runtime | Remove pairing |
| `pair.revoke` | Client→Runtime | Revoke specific session |
| `device.paired` | Runtime→Client | Event: new client paired |
| `device.unpaired` | Runtime→Client | Event: client removed |

### 9. Security Guarantees

| Threat | Mitigation |
|--------|-----------|
| **MITM during pairing** | PIN or code comparison authenticates the exchange |
| **Replay attack** | Nonce in challenge, single-use PIN, timestamps |
| **Brute force PIN** | Rate limit: 3 attempts → 30s lockout → resets after correct PIN or 5 min |
| **Pairing jacking** | Physical access to device needed for PIN display |
| **Offline key compromise** | Ed25519 key derived from per-pairing nonce, not stored after session established |
| **Stolen session token** | Token bound to client identity; Runtime can revoke all tokens |
| **Rogue client on LAN** | PIN displayed on device, attacker cannot guess |

### 10. Error Codes

| Error | HTTP Analog | When |
|-------|-------------|------|
| `pair.pin_expired` | 410 | PIN older than 120 seconds |
| `pair.pin_invalid` | 403 | PIN does not match |
| `pair.pin_locked` | 429 | Too many PIN attempts |
| `pair.session_expired` | 401 | Pairing session timed out |
| `pair.already_paired` | 409 | Device at max pairings |
| `pair.claim_expired` | 410 | Claim code older than 5 minutes |
| `pair.claim_invalid` | 403 | Claim code does not match |
| `pair.method_unsupported` | 501 | Client requests unsupported pairing method |

---

## Drawbacks

1. **PIN-based pairing requires a display.** Headless devices need the claim code flow, which requires cloud connectivity at pairing time. Mitigation: claim code flow is documented and minimal — user generates code on client, enters nothing on device.

2. **Ed25519 key exchange on every pairing is CPU-heavy for a Pi Zero.** Each X25519 operation takes ~5ms on a Pi Zero. Mitigation: this happens once per client, not per request.

3. **Cloud Relay dependency for remote access.** The pairing protocol is intentionally LAN-first, but remote access requires the cloud relay to be available. Mitigation: the relay is minimal (packet forwarding only) and we publish reference implementation so users can self-host.

---

## Rationale

1. **Why PIN over QR code?** QR scanning requires a camera on both sides. A PIN digits-only display works on any hardware, including OLED hats and 7-segment displays. QR is a future addition for phone-to-phone flows.

2. **Why Ed25519 + X25519 over TLS?** No CA dependency, no certificate management, no DNS. The device is its own certificate authority. The pairing handshake is the trust-on-first-use (TOFU) moment.

3. **Why SPAKE2+ for PAKE?** The pairing is inherently asymmetric (device shows PIN, client enters it). SPAKE2+ is the best-audited balanced PAKE, suitable for this low-power use case.

---

## Prior Art

- **HomeKit** — 8-digit PIN displayed on accessory, entered on iPhone. SPAKE2+ for key exchange. Ed25519 identity. This RFC is heavily inspired by HomeKit's pairing protocol.
- **Tailscale** — Pre-shared key + SSO for node authorization. Their `tailscale up` flow informed our claim code design.
- **Bluetooth Pairing** — Numeric comparison and passkey entry modes. Our code comparison (Method B) is directly from BT SSP.
- **Signal Protocol** — X3DH for asynchronous key exchange. We use a simpler synchronous model since both sides are online during pairing.

---

## Unresolved Questions

1. **Screen-less device recovery** — If a headless device needs re-pairing and has no cloud connectivity, how does the user regain access? Options: (a) physical reset button, (b) USB serial claim code, (c) SD card image with pre-loaded claim.

2. **Pairing timeout UX** — What happens if pairing is interrupted (app backgrounded, network drops)? Should we support resumable pairing sessions?

3. **Multi-account device** — Can a device be paired to multiple accounts (e.g., shared family device)? If so, admin roles need per-account assignment.

---

## Implementation Plan

| Phase | Milestone | Details |
|-------|-----------|---------|
| P0 | PIN generation | Runtime generates and displays 6-digit PIN on stdout |
| P1 | WebSocket pair handshake | `pair.initiate` / `pair.verify` / `pair.status` BPP methods |
| P2 | Key storage | Ed25519 generation, PID file, session token management |
| P3 | Cloud Relay integration | Pair registration, remote discovery, claim code flow |
| P4 | Headless support | Claim code generation, no-display fallback |
| P5 | Multi-client | Session table, role management, admin revocation |
| P6 | Polish | Rate limiting, lockout, audit logging |

---

## References

- RFC-0001: Connection Engine
- RFC-0002: Runtime Architecture
- BPP Protocol: `device.discovery`, `device.pair`, `session.*` method specifications
- Engineering Book: architecture.md, capability-model.md
- Reference: packet-types.md (pairing method IDs 0x00A0-0x00AF)
