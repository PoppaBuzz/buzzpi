# BuzzPi Platform Architecture

The BuzzPi Platform is a protocol with multiple clients, not an Android app with a Pi agent. This distinction shapes every architectural decision. The platform functions as the **Raspberry Pi Experience Layer** — the interface between the user and the device that abstracts away everything networking.

---

## System Overview

```
          BuzzPi Protocol
                 |
     +-----------+-----------+
     |           |           |
Android      Desktop        CLI
 Client       Client      Client
     |           |           |
     +-----------+-----------+
                 |
          BuzzPi Agent
                 |
       Raspberry Pi / Linux
```

The BuzzPi Protocol (BPP) is the center of the ecosystem. Any client implementing it can communicate with any agent. Android is not special. The desktop client and CLI client are first-class citizens with identical capabilities.

---

## Core Components

### BuzzPi Protocol (BPP)

- JSON messages over WebSocket, versioned and stable
- Request-response for commands, streaming for events
- Human-readable during development, compact in production
- Designed for third-party clients from day one

### BuzzPi Agent

- Written in Go, distributed as a single binary
- Runs as a systemd service on the Raspberry Pi
- Responsibilities:
  - Device registration and identity management
  - Secure authentication (public/private key pairs)
  - Persistent WebSocket connection for remote access
  - System monitoring (CPU, memory, temperature, storage, network)
  - Remote command execution with sandboxing
  - Plugin hosting and lifecycle management
  - Automatic updates

### Android App

- Written in Kotlin with Jetpack Compose and Material 3
- Connection Engine handles all transport selection automatically
- Background service for notifications and reconnection
- Offline-first with local caching and sync

### Backend Services

- Written in Go for code sharing with the agent
- PostgreSQL for persistent storage
- Redis for caching, rate limiting, and event bus
- Object storage for file transfers and logs
- Cloud registry for device discovery when local methods are unavailable

---

## Connection Engine

The Connection Engine is the most important piece of client-side infrastructure. It automatically selects the best available transport without user input.

### Priority Order

1. **Existing session** — reuse active connections
2. **Local discovery** — mDNS/DNS-SD for same-network devices
3. **Local hostname** — direct hostname resolution
4. **Tailscale/MagicDNS** — zero-config VPN discovery
5. **WebRTC** — peer-to-peer when direct connection is unavailable
6. **Cloud relay** — fallback relay through BuzzPi infrastructure

The engine monitors connectivity continuously. If the network changes (moving from LAN to cellular, for example), it transitions transparently to the best available transport.

---

## Discovery Methods

| Method | Range | Setup Required |
|--------|-------|---------------|
| mDNS/DNS-SD | Local network | None |
| QR Pairing | Physical proximity | Agent generates QR on install |
| Bluetooth LE | Nearby (10m) | BLE on Pi + Android |
| USB Pairing | Direct USB connection | Gadget mode on Pi |
| Cloud Registration | Internet | Outbound connection from Pi |

---

## Agent API

The agent exposes a local REST and WebSocket API on the Pi:

| Endpoint | Purpose |
|----------|---------|
| `/info` | Device identity, model, OS version |
| `/stats` | CPU, memory, temperature, storage, network, uptime |
| `/files` | File browsing, upload, download, rename |
| `/docker` | Container and image management |
| `/services` | Systemd service status and control |
| `/logs` | Journald and application log streaming |
| `/camera` | Video streaming and snapshots |
| `/gpio` | Pin read/write and event monitoring |
| `/terminal` | WebSocket-based PTY for shell access |
| `/screen` | WebRTC-based graphical screen streaming |

---

## Screen Streaming (Remote Desktop)

BuzzPi's remote desktop is not a VNC client wrapper. It is a native screen streaming pipeline integrated into the agent and protocol, designed for low latency and adaptive quality.

### Architecture

```
Android App                     Raspberry Pi
+------------------+            +---------------------------+
|  Screen View     |  WebRTC   |  Screen Capture           |
|  (SurfaceView)   |<--------->|  (DRM / Wayland / X11)   |
|  Touch/Keyboard  |  Data     |  Input Injection           |
|  Input Handler   |  Channel  |  (uinput / libei)         |
+------------------+            +---------------------------+
        |                                |
        |  Signaling / Control           |
        |  (WebSocket via BPP)           |
        +--------------------------------+
```

### Capture Pipeline

The agent captures the Pi's graphical output using the best available method for the current display server:

- **Wayland** — wlr-screencopy or PipeWire for direct buffer access
- **X11** — MIT-SHM extension for efficient screen capture
- **Framebuffer** — DRM/KMS dumb buffers for headless or console mode
- **Hardware encoding** — VideoCore IV (Pi 4) or VideoCore VII (Pi 5) H.264 encoder on supported models

Captured frames are hardware-encoded into a video stream and sent over a WebRTC data channel to the client. The client renders frames on a SurfaceView for hardware-accelerated decoding.

### Adaptive Quality

Screen quality adjusts automatically based on network conditions:

| Condition | Resolution | Framerate | Codec |
|-----------|-----------|-----------|-------|
| LAN (wired) | Native | 60 fps | H.264 high |
| Wi-Fi (strong) | 1080p | 30 fps | H.264 high |
| Wi-Fi (weak) | 720p | 15 fps | H.264 main |
| Remote (4G/5G) | 480p | 10 fps | H.264 baseline |
| Remote (slow) | 240p | 5 fps | H.264 baseline |

### Input Handling

Touch and mouse events are captured on the client, serialized over the WebSocket control channel, and injected into the Pi's input subsystem using `uinput` or `libei`. This enables:

- Tap and swipe gestures
- Right-click via long-press
- Keyboard input with modifier keys
- Multi-touch where supported
- Scroll wheel emulation

### Why This Is Built In, Not Bolted On

By integrating screen streaming natively into the BuzzPi Agent and Protocol, we gain capabilities that a standalone VNC solution cannot provide:

- **Automatic transport selection** — screen traffic routes through the same Connection Engine (LAN -> Tailscale -> WebRTC -> relay)
- **Unified security** — screen sessions use the same end-to-end encryption and key pairs
- **Zero configuration** — no VNC server to install, no passwords to set, no ports to open
- **Plugin-aware** — the adaptive dashboard can show a screen preview alongside Docker stats and system health
- **Cross-client** — because screen streaming is part of BPP, desktop and CLI clients get it for free

---

## Plugin Architecture

The agent loads plugins at startup. Each plugin:

- Registers its own API endpoints
- Streams events to connected clients
- Executes commands in a sandboxed environment
- Declares its capabilities for the adaptive dashboard

### Planned Plugins

Docker, Pi-hole, Home Assistant, Frigate, Node-RED, Immich, Jellyfin, OctoPrint, Nextcloud, Mosquitto, WireGuard, Tailscale, Cloudflared.

Plugins are independent processes managed by the agent. A plugin crash does not affect the agent or other plugins.

---

## Security Architecture

| Layer | Mechanism |
|-------|-----------|
| Transport | TLS 1.3 minimum |
| Identity | Public/private key pairs |
| Client auth | Android Keystore + biometrics |
| Session | End-to-end encrypted |
| API | Per-endpoint authorization |
| Code | Signed agent updates |

- All communication is encrypted in transit
- Device identity is rooted in hardware keys
- Sessions are isolated and individually revocable
- No telemetry, no tracking, no analytics

---

## AI Architecture

The built-in assistant is not a chatbot. It is a diagnostics engine that:

1. Detects a problem (high CPU, service failure, disk full)
2. Collects relevant data automatically (logs, stats, dmesg, journalctl)
3. Analyzes the data and produces an explanation in plain language
4. Suggests or performs remediation steps

The AI runs on-device when possible and uses cloud inference only for complex analysis. Users control when and how AI features access their data.

---

## Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Android | Kotlin, Jetpack Compose, Material 3 | Best native Android experience |
| Networking | Ktor | Kotlin-native HTTP/WebSocket client |
| Local storage | Room, DataStore | Type-safe persistence |
| DI | Hilt | Standard Android DI |
| Background | WorkManager | Reliable background processing |
| Agent | Go | Single binary, excellent networking, easy cross-compilation for ARM |
| Backend | Go | Shared libraries with agent |
| Database | PostgreSQL | Mature, reliable, well-understood |
| Cache | Redis | Simple, fast, well-documented |
| Protocol | JSON over WebSocket | Human-readable, easy to debug, extensible |

No Kubernetes. No microservices. The architecture is kept intentionally simple and evolves only when there is a real need.
