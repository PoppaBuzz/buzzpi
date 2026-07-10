# RFC-0005: Security Model

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | RFC-0001, RFC-0002, RFC-0003 |

## Summary

Define the BuzzPi security model — cryptographic primitives, trust boundaries, threat model, authentication, authorization, and secure defaults. This RFC establishes the security foundations that all other RFCs and implementations must follow.

## Motivation

BuzzPi connects users to their devices over the internet. This creates a large attack surface: network protocols, cloud services, plugins, and physical device access. Security cannot be retrofitted.

A well-defined security model ensures:
1. All implementations follow consistent, audited patterns
2. Reviewers know what to look for
3. Users can trust the platform with device access
4. Vulnerabilities in one domain (e.g., plugin crash) do not compromise other domains (e.g., screen streaming)

## Design

### 1. Threat Model

#### Trust Boundaries

```
┌──────────────────────────┐      ┌──────────────────────┐
│       Client             │      │      Device           │
│  (Phone / Desktop / CLI) │      │  (Raspberry Pi)       │
│                          │      │                       │
│  ┌────────────────────┐  │      │  ┌─────────────────┐  │
│  │ User Context       │  │      │  │ Runtime Process  │  │
│  │ • App sandbox      │  │      │  │ • Supervisor     │  │
│  │ • OS keychain      │─┼──────┼─▶│ • Engine Mgr     │  │
│  │ • User auth        │  │  TLS │  │ • Connection Eng │  │
│  └────────────────────┘  │  WSS │  └────────┬────────┘  │
│                          │      │           │            │
│                          │      │  ┌────────┴────────┐  │
│                          │      │  │   Plugin Proc   │  │
│                          │      │  │   (Docker, GPIO)│  │
│                          │      │  └─────────────────┘  │
│                          │      │                       │
│                          │      │  ┌─────────────────┐  │
│                          │      │  │   Hardware      │  │
│                          │      │  │   Peripherals   │  │
│                          │      │  │   (GPIO, Camera)│  │
│                          │      │  └─────────────────┘  │
└──────────────────────────┘      └──────────────────────┘
         │                                │
         │          ┌──────────┐          │
         │          │  Cloud   │          │
         └──────────▶  Relay   │◀─────────┘
                    │  Service │
                    └──────────┘
```

**Trust Boundary 1: Client ↔ Network**
- Client secrets (session tokens, pairing keys) reside in OS keychain
- Client app runs in sandbox (Android app sandbox, OS process isolation)
- UI auto-lock prevents shoulder-surfing on unattended devices

**Trust Boundary 2: Network ↔ Runtime**
- All communication over TLS 1.3 (WSS)
- Runtime validates session tokens on every request
- Runtime rate-limits authentication attempts

**Trust Boundary 3: Runtime ↔ Plugin**
- Plugin runs as sub-process with restricted permissions
- Plugin cannot access Runtime memory or identity keys
- Plugin cannot interfere with other plugins

**Trust Boundary 4: Runtime ↔ Hardware**
- GPIO access via `/dev/gpiomem` (user-space GPIO)
- Camera access via Video4Linux2
- Screen capture via KMS/DRM or X11

#### Assets to Protect

| Asset | Sensitivity | Where | Compromise Impact |
|-------|-------------|-------|-------------------|
| Device identity key | Critical | Runtime PID file | Permanent device impersonation |
| Session tokens | High | Runtime + Client | Unauthorized device access |
| Pairing keys | Critical | Runtime + Client | Permanent MITM capability |
| Device config | Medium | Runtime state store | Reconnaissance |
| Plugin data | Varies | Plugin process | Depends on plugin |
| Screen stream | High | WebRTC channel | Privacy violation |
| Terminal session | High | WebSocket stream | Full device shell access |

#### Attack Scenarios

| # | Scenario | Severity | Mitigation |
|---|----------|----------|------------|
| 1 | Attacker on same LAN intercepts pairing | Critical | PAKE (SPAKE2+) + PIN verification prevents MITM |
| 2 | Attacker steals client phone | High | OS keychain + app lock + remote session revocation |
| 3 | Attacker gains physical access to Pi | Critical | Identity on encrypted storage; factory reset by SD card reflash |
| 4 | Plugin vulnerability (RCE) | High | Sub-process sandbox + Landlock + seccomp |
| 5 | Cloud relay compromise | Critical | End-to-end encryption — relay sees only encrypted packets |
| 6 | Replay attack on BPP messages | Medium | Nonces, timestamps, message sequence numbers |
| 7 | Brute-force pairing PIN | Medium | Rate limiting + lockout + escalating delay |
| 8 | Man-in-the-middle on relay connection | High | TLS 1.3 + certificate pinning (client) |
| 9 | Malicious plugin exfiltrates data | High | Network permissions declared in manifest, audited |
| 10 | DoS on Runtime via many connections | Medium | Connection limit + rate limiting + resource quotas |

### 2. Cryptographic Primitives

| Operation | Primitive | Rationale |
|-----------|-----------|-----------|
| Device identity | Ed25519 | Fast keygen, small signatures, Go stdlib (Go 1.20+) |
| Key exchange | X25519 | Widely deployed, well-audited, hardware acceleration on ARM |
| PAKE (pairing) | SPAKE2+ | Best-audited balanced PAKE for PIN-based auth |
| Session tokens | SHA-256 (HMAC) | Fast, no length-extension attack concerns |
| TLS | TLS 1.3 | Mandatory — no fallback to TLS 1.2 |
| Certificate pinning | SPKI hash | Pin relay.buzzpi.dev public key on first connection |
| Random number gen | `crypto/rand` (Go) | OS entropy source |
| Secure compare | Constant-time | All token/code comparisons use constant-time comparison |

#### Key Derivation

```go
// Device identity (generated once)
devicePrivateKey, _ := ed25519.GenerateKey(rand.Reader)
devicePublicKey := devicePrivateKey.Public().(ed25519.PublicKey)

// Device ID from public key
hash := sha256.Sum256(devicePublicKey)
deviceID := "dev_" + base62.Encode(hash[:12])

// Session key derivation (after PAKE)
sessionKey := hkdf.Expand(sha256.New, sharedSecret, []byte("buzzpi-session"), 32)

// Token HMAC
tokenMAC := hmac.New(sha256.New, sessionKey)
tokenMAC.Write([]byte(sessionID))
sessionToken := base62.Encode(tokenMAC.Sum(nil))
```

### 3. Authentication

#### Session Token Format

```
sess_<base62(32-bytes-random)>
```

Tokens are:
- Generated by the Runtime at pairing time
- Validated on every BPP request
- Rotated every 24 hours (sliding window)
- Stored in BoltDB (Runtime) and OS keychain (Client)

#### Token Validation

```go
func (r *Runtime) AuthenticateRequest(ctx context.Context, envelope *Envelope) (*Session, error) {
    token := envelope.Headers.SessionToken
    if token == "" {
        return nil, ErrNotAuthenticated
    }

    session, err := r.stateStore.GetSession(token)
    if err != nil {
        return nil, ErrNotAuthenticated
    }

    if time.Now().After(session.ExpiresAt) {
        r.stateStore.DeleteSession(token)
        return nil, ErrSessionExpired
    }

    // Sliding window refresh
    session.ExpiresAt = time.Now().Add(24 * time.Hour)
    session.LastUsedAt = time.Now()
    r.stateStore.UpdateSession(session)

    return session, nil
}
```

#### Role-Based Access Control

| Role | Can Read Device | Can Execute Actions | Can Configure | Can Manage Pairings |
|------|----------------|---------------------|---------------|---------------------|
| Viewer | ✅ | ❌ | ❌ | ❌ |
| Member | ✅ | ✅ | Limited | ❌ |
| Admin | ✅ | ✅ | ✅ | ✅ |

Roles are assigned during pairing. The first client to pair gets Admin. Subsequent clients get Member by default (Admin can promote).

### 4. Transport Security

#### TLS Configuration

```
Minimum version: TLS 1.3
Cipher suites: TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384
Curves: X25519, P-256
Certificate: ECDSA P-256 (not RSA)
```

#### Certificate Management

- **Runtime self-signed certificate** — generated at first boot, used for LAN connections
- **Cloud Relay certificate** — Let's Encrypt / public CA, used for WSS relay connections
- **Certificate pinning (Android)** — SPKI hash of relay.buzzpi.dev, updated via app update

#### WebSocket Security

```
wss://device.local:49872/    → LAN (self-signed cert, TOFU)
wss://relay.buzzpi.dev/      → Remote (CA-signed cert, pinned)
```

All WebSocket connections use TLS 1.3. The Runtime refuses plain `ws://` connections.

### 5. Plugin Security

#### Process Isolation

```
                ┌──────────────────────┐
                │   Runtime Process    │
                │   PID 1234           │
                │   User: buzzpi       │
                └──────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                 ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Docker Plugin│ │ GPIO Plugin  │ │ Pi-hole Plug │
│ PID 5678     │ │ PID 9012     │ │ PID 3456     │
│ User: buzzpi │ │ User: buzzpi │ │ User: buzzpi │
│ Landlock: ✓  │ │ Landlock: ✓  │ │ Landlock: ✓  │
│ seccomp: ✓   │ │ seccomp: ✓   │ │ seccomp: ✓   │
└──────────────┘ └──────────────┘ └──────────────┘
```

**Isolation mechanisms:**

| Mechanism | Protection | Linux Version |
|-----------|------------|---------------|
| Landlock | Filesystem access scoping | 5.13+ |
| seccomp-bpf | System call filtering | 4.8+ |
| Capabilities | Drop `CAP_SYS_ADMIN`, `CAP_NET_RAW` | Always |
| User namespace | Map to unprivileged UID | 3.8+ |
| cgroup v2 | Memory/CPU limits | 4.15+ |

**Minimum Linux requirement for plugin sandboxing:** kernel 5.13 (Raspberry Pi OS ships 6.1+).

#### Plugin Identity

Each plugin is identified by its manifest `id` field. The Runtime verifies:
1. Plugin directory is owned by `root:root` (installed by package manager or `buzzpi plugin install`)
2. Plugin binary is not writable by other users
3. Plugin manifest signature (if signed by author) matches the expected hash

**Plugin signing (future, v1.0):**
```bash
buzzpi plugin sign ./docker-manager.buzzpi-plugin \
    --key ~/.buzzpi/plugin-signing-key.pem

buzzpi plugin verify ./docker-manager.buzzpi-plugin \
    --signer-pubkey ./author.pub
```

### 6. Data Security

#### Data at Rest

| Data | Storage | Encryption |
|------|---------|------------|
| Device identity key | Filesystem (`/var/lib/buzzpi/identity/`) | File permissions 0600 |
| Session tokens | BoltDB (`/var/lib/buzzpi/state.db`) | File permissions 0600 |
| Runtime config | `/etc/buzzpi/runtime.yaml` | File permissions 0644 |
| Plugin data | Plugin directory | Plugin responsibility |
| Logs | `/var/log/buzzpi/` | File permissions 0644 |

**Encryption at rest:** Not required for v0.x. The physical security assumption is that the Pi is in a physically secure location. If the Pi is stolen, the attacker can access the storage directly — the identity key is on an SD card. **Future (v1.0):** LUKS encryption on the data partition with TPM-backed key (Pi 5 with TPM module).

#### Data in Transit

| Channel | Encryption | Authentication |
|---------|-----------|---------------|
| Client ↔ Runtime (LAN) | TLS 1.3 | Self-signed cert + TOFU |
| Client ↔ Relay | TLS 1.3 | CA-signed cert + pinning |
| Relay ↔ Runtime | TLS 1.3 | CA-signed cert + session token |
| Plugin ↔ Runtime | Unix pipes (no network) | Process identity (PID + UID) |
| Screen stream (WebRTC) | DTLS 1.2 + SRTP | Session token via signaling |

#### Data Privacy

| Data Point | Collected? | Purpose | Retention |
|------------|-----------|---------|-----------|
| Device ID | Yes | Device identity, pairing | Until factory reset |
| Device name | Yes | User-friendly identification | Until changed |
| IP address | Transient | Routing packets | Not stored |
| Usage metrics | Opt-in only | Improvement | 90 days |
| Crash reports | Opt-in only | Stability | 90 days |
| Plugin data | No | Plugin responsibility | N/A |

**Opt-in telemetry:**
```
BuzzPi would like to collect anonymous crash reports and usage statistics.
This helps us fix bugs and improve the app.

[Share Analytics] [No Thanks]

You can change this anytime in Settings → Privacy.
```

### 7. Secure Defaults

| Setting | Default | Rationale |
|---------|---------|-----------|
| TLS | Required, no fallback to plaintext | No accidental unencrypted connections |
| Plugin permissions | Minimal (no network, no filesystem) | Least privilege, user must explicitly grant |
| Telemetry | Off | Privacy by default |
| Remote access | Off (LAN only) | User must enable remote access |
| Pairing lockout | 3 attempts → 30s → 5min → 30min | Prevent brute force |
| Session timeout | 24 hours | Limit exposure of stolen tokens |
| Max clients | 5 | Prevent resource exhaustion |
| Log level | info | Diagnostic info without sensitive data |
| Auto-lock (Android) | 30 seconds | Prevent shoulder-surfing |
| Screen capture | Off (user must start) | Prevent accidental streaming |

### 8. Security Audit Checklist

Before each release:

```
[ ] TLS 1.3 verified (no TLS 1.2 fallback)
[ ] No hardcoded credentials in source
[ ] All token comparisons use constant-time
[ ] Plugin manifest permissions validated
[ ] Rate limiting configured and tested
[ ] Session token rotation verified
[ ] Input validation on all BPP methods
[ ] No sensitive data in logs (tokens, keys, passwords)
[ ] File permissions verified (0600 for identity, 0644 for config)
[ ] CGO disabled for Go binaries (pure Go for security-critical paths)
[ ] All dependencies scanned for CVEs
[ ] Fuzz testing on BPP message parser
```

### 9. Vulnerability Reporting

**Reporting channel:** `security@buzzpi.dev` (PGP-encrypted, key published on website)

**Response SLA:**

| Severity | Initial Response | Fix Target |
|----------|-----------------|------------|
| Critical (RCE, key compromise) | 4 hours | 48 hours |
| High (unauthorized access) | 8 hours | 7 days |
| Medium (limited information disclosure) | 24 hours | 30 days |
| Low (minor hardening) | 72 hours | Next release |

**Disclosure policy:** Coordinated disclosure. Fix published before public announcement. CVE assignment for vulnerabilities in the protocol or core Runtime.

---

## Drawbacks

1. **No encryption at rest in v0.x** — Physical device theft = key compromise. Acceptable for v0.x (devices are typically in homes, not data centers). Full disk encryption with TPM is planned for v1.0.

2. **Self-signed certificates for LAN** — Users may see browser warnings if connecting via web. Mitigation: BuzzPi uses native apps, not browsers. The app performs TOFU validation silently.

3. **Seccomp/Landlock require modern kernel** — Pi Zero with kernel 5.10 may not support all sandboxing features. Mitigation: graceful degradation — plugins still run as sub-processes with reduced capabilities even without Landlock.

---

## Rationale

1. **Why Ed25519 over ECDSA P-256?** Ed25519 is faster, has smaller signatures, and avoids P-256's side-channel concerns. Go 1.20+ includes Ed25519 in the standard library.

2. **Why SPAKE2+ over SRP?** SPAKE2+ is simpler, better-audited, and avoids SRP's patent-encumbered history. It is recommended by the CFRG for PAKE applications.

3. **Why no encryption at rest?** The Pi's SD card makes encryption challenging (limited writes, no hardware TPM on Pi 4). Performance impact on flash storage is significant. The threat model assumes physical security for the device.

4. **Why session tokens over JWT?** JWTs are large, require signature verification on every request, and introduce key management complexity. Simple random tokens (32 bytes) in BoltDB are faster and simpler for the Runtime's use case.

---

## Prior Art

- **Tailscale** — Ed25519 node identity, TOFU, coordination server. Inspires our identity model.
- **HomeKit** — SPAKE2+ pairing, Ed25519 identity, MFi coprocessor. Inspires our pairing protocol.
- **WireGuard** — Minimal crypto, ed25519 keys, no CA. Inspires our "no unnecessary complexity" approach.
- **Signal Protocol** — Peer-to-peer key exchange, forward secrecy. Future: we may incorporate Signal's double ratchet for screen streaming.

---

## Unresolved Questions

1. **Forward secrecy for screen streams** — Should WebRTC screen streams use forward-secret keys rotated every frame? This would protect past streams if a session key is compromised. The performance cost is significant. Deferred to v1.0.

2. **Hardware security module** — Should we support TPM or Secure Element on Pi 5 for key storage? The OpenBSD PISCSI work shows this is possible. Leaning toward optional support in v0.5+, default in v1.0.

3. **Audit logging** — Should all sensitive actions (pair, unpair, exec, config change) be logged to an immutable audit trail? Leaning toward yes for Admin actions in v0.5+.

4. **Remote attestation** — Should the Runtime prove its integrity to clients? This requires TPM 2.0 (Pi 5) and is deferred to v1.0.

---

## Implementation Plan

| Phase | Milestone | Details |
|-------|-----------|---------|
| P0 | Identity | Ed25519 key generation, PID file, device ID derivation |
| P1 | Session auth | Token generation, validation, rotation, sliding window |
| P2 | TLS | Self-signed cert generation, WSS enforcement, TOFU |
| P3 | Plugin sandbox | Sub-process isolation, Landlock, seccomp, cgroups |
| P4 | Rate limiting | Connection rate, auth rate, capability rate limits |
| P5 | Audit logging | Immutable audit trail for sensitive operations |
| P6 | Hardening | Fuzz testing, CVE scanning, penetration testing |

---

## References

- RFC-0001: Connection Engine (transport encryption)
- RFC-0002: Runtime Architecture (process model, permissions)
- RFC-0003: Pairing Protocol (key exchange, PIN verification)
- RFC-0004: Plugin Architecture (sandbox, permissions model)
- Engineering Book: capability-model.md, plugin-system.md
- Reference: error-codes.md (auth error codes)
