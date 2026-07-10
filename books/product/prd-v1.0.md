# PRD v1.0

**Product Requirements Document for BuzzPi v0.1.0 — the prototype release.** This PRD defines the scope, requirements, and acceptance criteria for the first functional version of BuzzPi. It is the bridge between architecture documentation and implementation.

---

## Release Identity

| Field | Value |
|-------|-------|
| **Release** | BuzzPi v0.1.0 — "Connection" |
| **Theme** | A prototype that proves we can connect a phone to a Pi without any networking knowledge |
| **Tagline** | "Plug it in. Open the app. Your Pi is there." |
| **Audience** | Internal testing, early developer preview |
| **Timeline** | TBD — estimated 8-12 weeks from start |
| **Install Method** | Manual (`curl | bash` for Runtime, APK sideload for Android) |

### Success Criteria

| Criterion | Target |
|-----------|--------|
| Connection success rate (LAN) | >90% on first attempt |
| Connection time (LAN) | <10s from app open to terminal prompt |
| mDNS discovery | Device appears within 5s of app opening |
| Terminal latency | <200ms key-to-display (P95) |
| Reconnection after network drop (<10s) | Session preserved |
| Developer can set up in | <5 minutes (1 terminal command) |
| Code health | >80% test coverage on connection engine |

---

## Functional Requirements

### F1: Device Discovery (P0)

Devices on the local network appear in the app automatically.

```gherkin
Scenario: Device appears on app launch
  Given the Runtime is running on a Raspberry Pi on the LAN
  When the user opens the BuzzPi Android app
  Then the device appears in the discovery list within 5 seconds
  And the device card shows the friendly name and model

Scenario: Device disappears when offline
  Given a device was visible in the discovery list
  When the device is disconnected from the network
  Then the device moves to the "offline" section within 30 seconds
```

**Dependencies:**
- Runtime implements mDNS service advertisement (`_buzzpi._tcp`)
- Android app implements mDNS service browser (JmDNS)
- Both on same subnet / LAN segment

**Acceptance:** 5 Raspberry Pi models (Zero W through Pi 5) all discoverable within 5s on a typical home WiFi network.

### F2: Device Pairing (P0)

Users can pair with a device using a 6-character code displayed on the device.

```gherkin
Scenario: Pair via code
  Given a device is visible in the discovery list
  When the user taps the device in discovery
  Then the app shows a pairing code prompt
  And the device displays a 6-character alphanumeric pairing code on its screen (or via HDMI)
  When the user enters the code displayed on the device
  Then the device is paired to the user's account
  And the device appears in the "My Devices" list

Scenario: Pair via QR code
  Given a device is in pairing mode
  When the user scans the QR code displayed on the device
  Then pairing completes without manual code entry
```

**Dependencies:**
- Backend registry service running
- Runtime generates pairing codes with expiration
- Android app has QR code scanner
- Ed25519 key exchange for cryptographic pairing

**Acceptance:** 100 pairing attempts across 3 network topologies (simple LAN, VLAN, WiFi-isolated) with >95% success rate.

### F3: Terminal Connection (P0)

Paired devices can be accessed via a full terminal session.

```gherkin
Scenario: Open terminal
  Given a device is paired and online
  When the user taps the device from "My Devices"
  Then a terminal session opens within 5 seconds
  And a shell prompt appears (`pi@raspberrypi:~ $`)
  And the user can type commands and see output

Scenario: Terminal keeps working across app background/foreground
  Given a terminal session is active
  When the user backgrounds the app for up to 30 seconds and returns
  Then the terminal session is still active
  And recent output is displayed

Scenario: Multiple terminal tabs
  Given a device connection is active
  When the user opens a second terminal tab
  Then both terminal sessions run independently
```

**Dependencies:**
- Runtime PTY service (Go `os/exec` with PTY)
- WebRTC data channel for terminal I/O (binary, unordered)
- Android terminal renderer (ANSI escape sequence support)
- Terminal input encoding (handling function keys, Ctrl+ combinations, paste)

**Acceptance:** 50 commands per minute typing speed without noticeable lag. ANSI colors render correctly. SSH-like terminal experience.

### F4: Reconnection (P1)

The terminal session survives brief network interruptions.

```gherkin
Scenario: Reconnect after short network drop
  Given a terminal session is active
  When the network is interrupted for 5 seconds
  Then the app shows "Reconnecting..." status
  When the network is restored
  Then the terminal session is restored
  And the user does not need to re-authenticate
```

**Dependencies:**
- Connection Engine reconnection manager
- Runtime session persistence (PTY session not killed on WebRTC disconnect)
- Backend relay for WAN reconnection

**Acceptance:** Session preserved through 3 consecutive 10-second network drops. Terminal state (working directory, environment variables) maintained.

### F5: System Monitoring (P1)

Users can see basic system stats for their devices.

```gherkin
Scenario: View device overview
  Given a device is paired and online
  When the user views the device overview screen
  Then they see:
    - CPU usage %
    - Memory usage (used/total)
    - Storage usage (used/total)
    - Temperature
    - Uptime
    - IP address
    - Runtime version
  And the data refreshes every 30 seconds
```

### F6: Push Notifications (P2)

Users receive notifications for important device events.

```gherkin
Scenario: Temperature alert notification
  Given a device is paired
  When the device temperature exceeds 80°C
  Then the user receives a push notification
  And the notification text includes the current temperature
  And tapping the notification opens the device detail screen
```

### F7: WAN Connection (P2)

Users can connect to devices across the internet (behind NAT).

```gherkin
Scenario: Connect via relay
  Given a device is paired and online
  And the device is on a different network than the app
  When the user opens the terminal
  Then a connection is established through the relay server within 10 seconds
  And the terminal experience is comparable (accepting up to 100ms additional latency)
```

---

## Non-Functional Requirements

### NFR1: Performance

| Metric | Target |
|--------|--------|
| App cold start to discovery list | <3s |
| Pairing flow (scan → connected) | <15s |
| Terminal key press to display | <200ms (P95) |
| Terminal throughput | >1000 lines/second |
| Runtime memory usage (idle) | <30MB |
| Runtime memory usage (active terminal) | <60MB |
| Runtime CPU usage (idle) | <1% |
| App APK size | <25MB |
| Runtime binary size | <15MB (compressed) |

### NFR2: Reliability

| Metric | Target |
|--------|--------|
| App crash-free rate | >99.5% |
| Runtime uptime | >99.9% (only process restarts) |
| Connection success rate | >95% (LAN), >90% (WAN/relay) |
| Pairing success rate | >95% |
| mDNS discovery reliability | >99% devices found |
| Reconnection success (<10s drop) | >95% |

### NFR3: Security

- All communication encrypted in transit (TLS for WebSocket, DTLS for WebRTC)
- Pairing uses Ed25519 key exchange — no password sent over network
- JWT tokens with 15-minute expiry (access) and 7-day expiry (refresh)
- Refresh tokens stored as SHA-256 hashes
- Commands executed through PTY are sandboxed — no direct shell access from relay
- Runtime runs as non-root user
- No telemetry or analytics compiled into release builds

### NFR4: Compatibility

| Component | Minimum Target |
|-----------|---------------|
| Raspberry Pi models | Zero W, 3B+, 4, 5 |
| Raspberry Pi OS | Debian Bookworm (12) arm64 |
| Android API level | 26+ (Android 8.0) |
| Android architecture | arm64-v8a, armeabi-v7a |
| Network | Typical home WiFi, no mDNS VLAN isolation |
| Backend server | 2 vCPU, 4GB RAM, 50GB SSD |

### NFR5: Developer Experience

- Single-command Runtime install: `curl -sS https://get.buzzpi.dev | bash`
- Runtime logs visible via `journalctl -u buzzpi-runtime`
- Runtime configuration via `/etc/buzzpi/config.yaml`
- Android app installable via APK sideload (debug build)
- All Go dependencies vendored, no network required to build

---

## Out of Scope (v0.1.0)

The following features are explicitly excluded from v0.1.0:

- **Screen streaming** — requires DRM/KMS capture, H.264 encoding, hardware decoder on Android. Planned for v0.2.0.
- **File manager** — requires full file system abstraction, chunked transfer protocol. Planned for v0.2.0.
- **Docker management** — requires Docker socket abstraction. Planned for v0.2.0.
- **GPIO control** — requires GPIO sysfs/libgpiod abstraction. Planned for v0.2.0.
- **Camera streaming** — requires V4L2 capture, H.264 encoding. Planned for v0.2.0.
- **Desktop client** — not planned until v0.5.0.
- **CLI client** — not planned until v0.5.0.
- **Plugin system** — requires IPC protocol, permissions model. Planned for v0.3.0.
- **BuzzAI assistant** — requires LLM integration. Planned for v0.5.0+.
- **Push notifications** — requires FCM integration, notification service. Planned for v0.2.0.
- **Multiple user accounts** — v0.1.0 supports one user per device. Multi-user planned for v0.3.0.
- **Windows/macOS Runtime** — Linux-only in v0.1.0.
- **Over-the-air updates** — manual updates in v0.1.0.

---

## User Stories by Persona

### Sarah (DIY Hobbyist)

| Story | Priority | F# |
|-------|----------|----|
| "I want to see my Pi appear in the app without configuring anything" | P0 | F1 |
| "I want to open a terminal to my Pi from my phone while sitting on the couch" | P0 | F3 |
| "I want to check CPU temperature without SSHing in" | P1 | F5 |
| "When my WiFi drops, I want my terminal to come back automatically" | P1 | F4 |

### Mike (Software Developer)

| Story | Priority | F# |
|-------|----------|----|
| "I want to try BuzzPi on my Pi 5 in under 5 minutes" | P0 | DevEx |
| "I want to tail logs from my Docker containers" | P2 | F5 |
| "I want to connect to my Pi at home from my office" | P2 | F7 |

### Ms. Chen (Educator)

| Story | Priority | F# |
|-------|----------|----|
| "I want 30 students to pair their phones with their Pis quickly" | P0 | F1, F2 |
| "I want to see which students have connected today" | P2 | — |

### Diego (System Administrator)

| Story | Priority | F# |
|-------|----------|----|
| "I need to know this is secure before I deploy it" | P0 | NFR3 |
| "I want to deploy the Runtime via my existing config management" | P1 | DevEx |

---

## Technical Deliverables

### Runtime (Go)

- [ ] Go module: `github.com/buzzpi/buzzpi`
- [ ] mDNS service advertiser (`_buzzpi._tcp`)
- [ ] WebSocket signaling client (relay connection)
- [ ] WebRTC peer connection (Pion)
- [ ] PTY terminal service (per-session, multi-session support)
- [ ] Pairing code generation and verification (Ed25519)
- [ ] System stats collection (CPU, memory, storage, temperature)
- [ ] systemd service unit file
- [ ] Install script (`install.sh`)
- [ ] Config file (`/etc/buzzpi/config.yaml`)
- [ ] Logging to journald

### Android App (Kotlin/Compose)

- [ ] mDNS service browser
- [ ] Pairing flow (QR code + manual code)
- [ ] Terminal view (ANSI escape sequence rendering)
- [ ] Terminal keyboard (special keys, Ctrl combinations, paste)
- [ ] Device list (discovery + paired)
- [ ] Device detail / overview screen
- [ ] WebSocket client (signaling)
- [ ] WebRTC peer connection (Pion jlibp2p or similar)
- [ ] Reconnection handler
- [ ] Navigation graph (discovery → pairing → device detail)
- [ ] DataStore for paired device cache
- [ ] Material3 theme (from design tokens)
- [ ] Single-activity architecture with Compose Navigation

### Backend (Go)

- [ ] Registry service (PostgreSQL)
  - [ ] User signup/login (JWT)
  - [ ] Device registration
  - [ ] Device listing
- [ ] Signaling relay service (WebSocket)
  - [ ] Session management
  - [ ] SDP forwarding
  - [ ] ICE candidate forwarding
  - [ ] Heartbeat/timeout
- [ ] TURN credential generation
- [ ] Schema initialization migration
- [ ] Rate limiting middleware
- [ ] Dockerfile + docker-compose.yml

---

## Release Milestones

```
M1: Runtime MVP ─────────────────── Week 1-3
  - Go module, mDNS, WebSocket client, PTY service
  - Can: curl | bash → device appears on network
  - Playground test: open terminal from CLI
    
M2: Android MVP ────────────────── Week 3-6
  - Discovery, pairing, terminal view
  - Can: open app → see device → pair → type commands
  - Playground test: full round-trip on LAN
  
M3: Backend MVP ────────────────── Week 5-7
  - Registry service, relay signaling, TURN
  - Can: sign up → register device → connect via relay
  - Playground test: WAN connection
  
M4: Polish & Hardening ─────────── Week 7-9
  - Reconnection, error handling, edge cases
  - Performance optimization
  - 80%+ test coverage target
  
M5: Release ────────────────────── Week 9-10
  - Install script, APK build, documentation
  - Internal testing
  - Tag v0.1.0
```

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| WebRTC P2P fails on restrictive networks | Medium | High | TURN relay fallback in M3; STUN from day one |
| mDNS unreliable on enterprise/managed WiFi | Medium | Medium | Backend device list as fallback; document requirement for mDNS-capable WiFi |
| ANSI terminal rendering performance on Android | Low | Medium | Use Span-based rendering; virtual scroll; profile on low-end devices |
| Go WebRTC library (Pion) compatibility issues | Low | High | Prototype Pion P2P in week 1; have alternative (WebSocket tunneling as fallback) |
| Android WebRTC library maturity | Medium | Medium | Evaluate Pion/jlibp2p vs GStreamer vs WebRTC native; prototype both |
| Pi Zero W performance | Medium | Low | Set minimum expectations; optimize critical paths; consider Zero 2 W as minimum |
| JWT secret management | Low | High | Document env-var-based config; use key rotation pattern; audit before release |
