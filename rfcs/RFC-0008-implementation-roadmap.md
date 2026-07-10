# RFC-0008: Implementation Roadmap

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | All RFCs 0000-0007 |

## Summary

Define the phased implementation roadmap from v0.0.0 (architecture complete) to v1.0.0 (stable public release). This RFC ties together all prior RFCs and defines the specific deliverables, dependencies, and success criteria for each version milestone.

## Motivation

BuzzPi's architecture is fully defined. The next question is: **what do we build first?** A clear roadmap ensures:

1. **Dependencies are respected** — you cannot implement the Plugin Host before the Engine Manager
2. **Each milestone is shippable** — every version is a working system, even if minimal
3. **Progress is measurable** — clear criteria for what "done" means at each stage
4. **Parallel work is enabled** — independent components can be built simultaneously

## Design

### 1. Development Phases

```
v0.0.0 ──────────────────────────────────────────────────────── v1.0.0
 │                                                                   │
 ├── Foundation (v0.0.x) ────────────── Complete                     │
 │   Books, RFCs, governance, architecture                           │
 │                                                                   │
 ├── Proof of Concept (v0.1.0‐v0.1.x) ───── HERE                    │
 │   Runtime skeleton, mDNS, WebSocket, device.info, CLI             │
 │                                                                   │
 ├── Core (v0.2.0‐v0.2.x) ─────────────────────────────┐            │
 │   Pairing, session management, terminal, Android app │            │
 │                                                        │          │
 ├── Visual (v0.3.0‐v0.3.x) ─────────────────────────┐  │          │
 │   Screen streaming, WebRTC, Remote Desktop          │  │          │
 │                                                      │  │          │
 ├── Rich (v0.4.0‐v0.4.x) ───────────────────────┐   │  │          │
 │   File manager, GPIO, Camera, System monitor   │   │  │          │
 │                                                  │   │  │          │
 ├── Extensible (v0.5.0‐v0.5.x) ─────────────┐   │   │  │          │
 │   Plugin system, SDK, Registry              │   │   │  │          │
 │                                               │   │   │  │          │
 ├── Community (v0.6.0‐v0.9.x) ──────────┐   │   │   │  │          │
 │   Beta testing, hardening, fuzzing     │   │   │   │  │          │
 │                                          │   │   │   │  │          │
 └── Stable (v1.0.0) ───────────────────┴───┴───┴───┴──┴──────────┘
```

### 2. Milestone: v0.1.0 — Proof of Concept

**Tagline:** "The daemon runs, discovers itself, and answers questions."

**Duration:** 4-6 weeks

#### Components

| Component | Language | Owner | Dependencies |
|-----------|----------|-------|-------------|
| Runtime Supervisor | Go | TBD | None |
| mDNS Discovery | Go | TBD | None |
| WebSocket Server | Go | TBD | None |
| Engine Manager (minimal) | Go | TBD | WebSocket server |
| device.info handler | Go | TBD | Engine Manager |
| device.stats handler | Go | TBD | Engine Manager |
| CLI client | Go | TBD | mDNS, WebSocket client |
| Connection Engine (LAN) | Go | TBD | WebSocket, mDNS |
| State Store (BoltDB) | Go | TBD | None |

#### Deliverables

```
buzzpi/                     ─── Go workspace root
├── cmd/
│   ├── runtime/            ─── BuzzPi Runtime binary
│   │   └── main.go
│   └── buzzpi/             ─── BuzzPi CLI client
│       └── main.go
├── internal/
│   ├── supervisor/         ─── Lifecycle management
│   ├── mdns/               ─── mDNS advertiser + browser
│   ├── ws/                 ─── WebSocket server + client
│   ├── engine/             ─── Engine Manager (method routing)
│   ├── device/             ─── Device info + stats handlers
│   ├── state/              ─── BoltDB state store
│   ├── connection/         ─── Connection Engine
│   └── version/            ─── Version string
├── go.mod
└── Makefile
```

#### User-Facing Capabilities

```
✓ Runtime starts with: buzzpi-runtime
✓ Runtime advertises on mDNS as _buzzpi._tcp
✓ CLI discovers devices with: buzzpi discover
✓ CLI queries device info: buzzpi device info <device_id>
✓ CLI queries device stats: buzzpi device stats <device_id>
✓ WebSocket connection survives brief network interruptions
```

#### Success Criteria

```
[ ] Runtime binary compiles for linux/arm64, linux/amd64
[ ] Runtime starts, announces mDNS, responds to device.info
[ ] CLI discovers the runtime on the same LAN
[ ] CLI queries device info (name, version, platform)
[ ] CLI queries device stats (CPU, memory, uptime)
[ ] WebSocket reconnection handles 5s network drop
[ ] State store persists and retrieves values correctly
[ ] Supervisor handles SIGTERM with graceful shutdown (< 5s)
[ ] Unit test coverage > 70% for all packages
[ ] CI pipeline builds and tests on push
```

### 3. Milestone: v0.2.0 — Core

**Tagline:** "Pair your phone, open a terminal."

**Duration:** 6-8 weeks

#### New Components

| Component | Language | Dependencies |
|-----------|----------|-------------|
| Pairing handlers | Go | Engine Manager, State Store |
| Ed25519 identity | Go | None |
| Session management | Go | State Store |
| Terminal manager | Go | PTY (creack/pty) |
| Android app skeleton | Kotlin | None |
| Android ViewModel layer | Kotlin | App skeleton |
| Cloud Relay (basic) | Go | None |
| Discovery screen | Kotlin | Android app |
| Terminal screen | Kotlin | Android app, WebSocket |

#### Deliverables

```
buzzpi-android/             ─── Android project
├── app/
│   ├── src/main/java/com/buzzpi/android/
│   │   ├── ui/screens/discovery/
│   │   ├── ui/screens/terminal/
│   │   ├── ui/theme/
│   │   └── navigation/
│   ├── build.gradle.kts
│   └── AndroidManifest.xml
├── gradle/
└── settings.gradle.kts
```

#### User-Facing Capabilities

```
✓ Android app discovers devices via mDNS
✓ Android app pairs with device via 6-digit PIN
✓ Paired devices persist across app restarts
✓ Terminal session: open, type, see output, resize, close
✓ Terminal supports ANSI colors, Ctrl+C, tab complete
✓ Cloud Relay: Runtime connects, client connects remotely
✓ Remote (relayed) terminal session
✓ Graceful reconnection when app is backgrounded/foregrounded
```

### 4. Milestone: v0.3.0 — Visual

**Tagline:** "See your Pi's screen on your phone."

**Duration:** 8-10 weeks

#### New Components

| Component | Language | Dependencies |
|-----------|----------|-------------|
| Screen capture service | Go | KMS/DRM or X11 |
| H.264 encoder | Go | Video4Linux, MMAL |
| WebRTC signaling | Go | Connection Engine |
| Screen viewer screen | Kotlin | Android app, WebRTC |
| Touch input handler | Go | Screen capture |
| Quality adaptation | Go | Connection Engine |
| WebRTC Android renderer | Kotlin | Screen viewer |

#### User-Facing Capabilities

```
✓ Remote desktop: see Pi's screen on phone
✓ Touch input: tap, drag, pinch, scroll
✓ Keyboard input for screen viewer
✓ Quality adaptation: auto-switch on network change
✓ Screen streaming at 30fps on LAN (Pi 4+)
✓ Screen streaming at 10fps on 4G
✓ Landscape fullscreen mode
✓ Cursor visibility toggle
```

### 5. Milestone: v0.4.0 — Rich

**Tagline:** "Manage files, GPIO, camera, and system."

**Duration:** 8-10 weeks

#### New Components

| Component | Language | Dependencies |
|-----------|----------|-------------|
| File manager service | Go | Engine Manager |
| GPIO service | Go | /dev/gpiomem |
| Camera service | Go | libcamera |
| System monitor | Go | /proc, /sys |
| File browser screen | Kotlin | Android app |
| System monitor screen | Kotlin | Android app |
| FCM push notifications | Kotlin | Cloud Relay |

#### User-Facing Capabilities

```
✓ File browser: list, navigate, create directories
✓ File upload (phone → device)
✓ File download (device → phone)
✓ GPIO read/write individual pins
✓ GPIO watch (event-driven pin changes)
✓ Camera live preview
✓ Camera snapshot (JPEG)
✓ System stats: CPU, memory, storage, temperature, network
✓ Push notification: overheating, low storage, service failures
```

### 6. Milestone: v0.5.0 — Extensible

**Tagline:** "Plugins, SDK, and ecosystem."

**Duration:** 8-10 weeks

#### New Components

| Component | Language | Dependencies |
|-----------|----------|-------------|
| Plugin Host | Go | Sub-process manager |
| Plugin IPC (JSON-RPC) | Go | Plugin Host |
| Go SDK | Go | None (external) |
| Python SDK | Python | None (external) |
| Plugin Registry | Go | Cloud Relay |
| Plugin management CLI | Go | CLI |
| Plugin certification | Process | Community |

#### User-Facing Capabilities

```
✓ buzzpi plugin install docker-manager
✓ buzzpi plugin list (installed plugins)
✓ buzzpi plugin uninstall docker-manager
✓ Plugin appears as a capability on the device
✓ Plugin capabilities usable from Android app
✓ Plugin crash does not crash Runtime
✓ Plugin permissions enforced (network, filesystem)
✓ Plugin registry at plugins.buzzpi.dev
```

### 7. Milestones: v0.6.0–v0.9.0 — Community

**Tagline:** "Beta, hardening, polish."

**Duration:** 12-16 weeks total

#### v0.6.0: Community Preview

```
✓ All core features working
✓ Public beta registration
✓ Documentation complete for all features
✓ Plugin SDK documentation
✓ Community contribution guidelines active
✓ Issue tracker organized with labels and milestones
```

#### v0.7.0: Security Audit

```
✓ External security audit completed
✓ All findings addressed or documented
✓ Fuzz testing in CI
✓ CVE scanning in CI
✓ Bug bounty program announced (if funded)
```

#### v0.8.0: Performance

```
✓ Screen streaming: 30fps on Pi 4 LAN, 15fps on 10Mbps remote
✓ Terminal: <50ms latency on LAN
✓ Pairing: <5s on LAN
✓ App startup: <2s cold start
✓ Runtime memory: <40MB idle
✓ Runtime CPU: <5% idle on Pi 4
```

#### v0.9.0: Release Candidate

```
✓ Plugin API frozen
✓ BPP 1.0-rc.1 published
✓ Migration guide for breaking changes
✓ Android app in Play Store closed beta
✓ Docker Compose for self-hosted relay
```

### 8. Milestone: v1.0.0 — Stable

**Tagline:** "The first stable release."

**Duration:** 2-4 weeks (validation only)

#### Requirements

```
[ ] BPP 1.0 protocol specification finalized
[ ] All RFCs for v1.0 features accepted and implemented
[ ] Backward compatibility tested (v0.9.x → v1.0.0)
[ ] Plugin SDK GA (Go + Python)
[ ] Android app in Play Store production
[ ] Self-hosted relay documentation
[ ] Performance targets met
[ ] 30 days since last critical bug
```

### 9. Dependencies Map

```
v0.1.0 (Proof of Concept)
├── mDNS      → v0.2.0 (Android discovery)
├── WebSocket → v0.2.0 (pairing, terminal)
├── Engine    → v0.2.0 (all method handlers)
└── CLI       → v0.2.0 (debugging tool)

v0.2.0 (Core)
├── Pairing   → v0.3.0 (screen auth)
├── Session   → v0.3.0 (all auth)
├── Terminal  → v0.4.0 (file manager, system monitor)
├── Android   → v0.3.0 (screen viewer)
└── Relay     → v0.3.0 (remote screen)

v0.3.0 (Visual)
├── Screen    → v0.4.0 (camera shares capture pipeline)
├── WebRTC    → v1.0.0 (P2P fallback)
└── Touch     → v0.4.0 (GPIO interaction)

v0.4.0 (Rich)
├── File      → v0.5.0 (plugin store files)
├── GPIO      → v0.5.0 (plugin GPIO)
├── Camera    → v0.5.0 (plugin camera)
└── Notifs    → v0.6.0 (community preview)
```

### 10. Parallel Work Streams

| Stream | v0.1.0 | v0.2.0 | v0.3.0 | v0.4.0 | v0.5.0 |
|--------|--------|--------|--------|--------|--------|
| **Runtime** | Skeleton, mDNS, WS | Pairing, Terminal, Sessions | Screen capture, WebRTC | File, GPIO, Camera | Plugin Host, IPC |
| **Android** | — | Discovery, Pairing, Terminal | Screen viewer | File browser, System | Plugin UI |
| **CLI** | Discover, device.info | Pair, terminal (debug) | Screen status | File operations | Plugin management |
| **Relay** | — | Basic WebSocket relay | STUN/TURN | — | Plugin Registry |
| **SDK** | — | — | — | — | Go SDK, Python SDK |

### 11. Resource Estimates

| Phase | Runtime (Go) | Android (Kotlin) | Relay (Go) | CLI (Go) | Total |
|-------|-------------|-------------------|------------|----------|-------|
| v0.1.0 | 4-6 weeks | — | — | 1 week | 5-7 weeks |
| v0.2.0 | 4-6 weeks | 4-6 weeks | 2 weeks | 1 week | 6-8 weeks |
| v0.3.0 | 6-8 weeks | 4-6 weeks | 1 week | — | 7-9 weeks |
| v0.4.0 | 4-6 weeks | 4-6 weeks | — | 1 week | 6-8 weeks |
| v0.5.0 | 4-6 weeks | 2 weeks | 2 weeks | 1 week | 6-8 weeks |
| **Total** | **22-32 wks** | **14-20 wks** | **5 wks** | **4 wks** | **30-40 wks** |

### 12. Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Screen capture on Pi 4 is slow | High | Medium | Fall back to lower resolution; use MMAL encoding |
| WebRTC on Android has compatibility issues | Medium | High | Test on top 10 device models; fall back to MJPEG |
| Plugin ecosystem does not materialize | Medium | Low | Core features are built-in; plugins are additive |
| Developer bandwidth is insufficient | High | High | Prioritize v0.1-v0.2; defer features to later |
| H.264 licensing concerns | Low | Medium | Use openH264 (Cisco patent protection) or switch to VP9 |

---

## Drawbacks

1. **Long time to terminal** — A user must wait for v0.2.0 to get a terminal. v0.1.0 is developer-only. Mitigation: v0.1.0 is a proof of concept for the architecture; visible features start at v0.2.0.

2. **Android app starts late** — No mobile client until v0.2.0. The CLI serves as the primary client during v0.1.0. Mitigation: the CLI is a legitimate BuzzPi client; early adopters are developers comfortable with CLI.

3. **v0.3.0 is the most risky** — Screen streaming via WebRTC is the hardest technical challenge. It depends on both the Runtime (screen capture) and Android (WebRTC rendering). Mitigation: start WebRTC prototyping early (v0.2.0 timeline) to identify issues.

---

## Rationale

1. **Why v0.1.0 is CLI-only?** The CLI is faster to build than a full Android app. It validates the Runtime architecture without the Android development overhead. The CLI also serves as a developer tool throughout the project.

2. **Why terminal before screen streaming?** Terminal is a simpler real-time feature (text vs video). It validates the WebSocket streaming architecture, session management, and connection reliability before adding the complexity of WebRTC.

3. **Why plugins after core features?** Plugins depend on the capability model, Engine Manager, and permission system. These must be stable before external developers build on them.

---

## Prior Art

- **Home Assistant** — Started with core features (automations, sensors), added voice/companion apps later, then add-ons. Similar to our core → visual → plugin progression.
- **ESPHome** — Device firmware first, then dashboard, then cloud. Validates our CLI-first approach.
- **Tailscale** — Minimal viable product (wireguard + coordination server), then added features (ACLs, subnet routing, exit nodes). Inspires our "ship the minimal thing first" approach.

---

## Unresolved Questions

1. **Funding** — Who pays for the Cloud Relay bandwidth during development? Options: (a) free tier with limits, (b) self-host only during dev, (c) sponsorship. Leaning toward (b) for v0.1-v0.4, (a) for v0.5+.

2. **Release management** — How are releases coordinated across Runtime + Android + Relay? The Android app and Runtime have different update cadences. Options: (a) lockstep releases, (b) independent semantic versioning with compatibility contracts. Leaning toward (b).

3. **Platform support** — Should we support macOS/Linux desktop clients before v1.0.0? Leaning toward "no" — Android is the primary client; CLI serves developers. Desktop is post-1.0.

---

## References

- All RFCs 0001-0007
- Product Book: roadmap.md, prd-v1.0.md, release-strategy.md
- Engineering Book: All 14 chapters (architecture details for each component)
- Reference: platform-reference.md (build targets, filesystem layout)
