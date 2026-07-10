# RFC-0002: Runtime Architecture

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Target** | v0.1.0 (Proof of Concept) |
| **Requires** | RFC-0001 |

## Summary

Define the internal architecture of the BuzzPi Runtime — the Go daemon that runs on every Raspberry Pi. The Runtime is the core of the platform: it owns the device's identity, manages connections, provides capabilities, and runs plugins.

## Motivation

The Runtime is the only software that runs on-device. Its design determines reliability, security, and extensibility of the entire platform. A poorly designed Runtime creates a glass jaw — one crash or compromise takes the device offline.

This RFC defines the Runtime's internal structure, lifecycle, and extension mechanisms so that implementation can proceed with clear boundaries.

## Design

### 1. Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                   BuzzPi Runtime                     │
│  ┌───────────────────────────────────────────────┐  │
│  │              Supervisor (PID 1)                │  │
│  │  • Lifecycle management                        │  │
│  │  • Watchdog / health monitoring                │  │
│  │  • Graceful shutdown (SIGTERM → SIGKILL)       │  │
│  └──────────────┬────────────────────────────────┘  │
│                  │                                    │
│     ┌────────────┼────────────┬─────────────────┐    │
│     ▼            ▼            ▼                  ▼    │
│  ┌──────┐  ┌────────┐  ┌────────┐  ┌──────────────┐ │
│  │ Core │  │Engine  │  │ Plugin │  │ Connection   │ │
│  │ PID  │  │Manager │  │ Host   │  │ Engine       │ │
│  │ File │  │        │  │        │  │ (RFC-0001)   │ │
│  └──────┘  └────────┘  └────────┘  └──────────────┘ │
│  ┌──────┐  ┌────────┐  ┌────────┐  ┌──────────────┐ │
│  │State │  │Device  │  │GPIO    │  │ Camera       │ │
│  │Store │  │Manager │  │Service │  │ Service      │ │
│  └──────┘  └────────┘  └────────┘  └──────────────┘ │
│  ┌──────┐  ┌────────┐  ┌────────┐                    │
│  │mDNS  │  │Screen  │  │Terminal│                    │
│  │Disc. │  │Capture │  │Manager│                    │
│  └──────┘  └────────┘  └────────┘                    │
└─────────────────────────────────────────────────────┘
```

### 2. Component Breakdown

#### 2.1 Supervisor

The Supervisor is the first process. It owns the Runtime's lifecycle.

**Responsibilities:**
- Parse CLI flags and config file on startup
- Initialize logging (structured JSON to stdout + file rotation)
- Start all sub-components in dependency order
- Monitor sub-component health via heartbeat channels
- Handle OS signals (SIGTERM → graceful shutdown with 30s timeout → SIGKILL)
- Crash recovery: restart failed sub-components up to 3 times within 5 minutes

**Shutdown order:** 1. Connection Engine (drain clients) → 2. Plugin Host → 3. Sub-services → 4. State Store (flush) → 5. mDNS (goodbye packet)

```go
type SupervisorConfig struct {
    HeartbeatInterval time.Duration
    ShutdownTimeout   time.Duration
    MaxRestarts       int
    RestartWindow     time.Duration
}

type Component interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() HealthStatus
}
```

#### 2.2 Core P2P Identity

The device identity lives in a **PID file** (`/var/lib/buzzpi/identity/`) with:

```json
{
  "device_id": "dev_a1b2c3d4",
  "pairing_key": "base64-encoded-ed25519-private-key",
  "created_at": "2026-07-07T00:00:00Z",
  "friendly_name": "living-room-pi",
  "platform": "raspberry-pi/5",
  "runtime_version": "0.1.0"
}
```

- `device_id` is derived from the public pairing key (first 12 bytes → base62)
- The private key never leaves the device
- Factory reset = delete the identity directory
- The identity file is created on first boot if it does not exist

#### 2.3 State Store

Embedded key-value store for runtime state. **Does NOT use SQLite** — the state model is simpler than a relational schema.

**Storage engine:** BoltDB (embedded, ACID, single file, pure Go)

**Buckets:**

| Bucket | Content |
|--------|---------|
| `pairs` | Paired account info (account_id, display_name, pairing_timestamp) |
| `sessions` | Active session tokens (session_id → account_id + expiry) |
| `config` | Runtime configuration overrides |
| `plugins` | Plugin manifest cache |
| `telemetry` | Local telemetry buffer (rotated) |
| `state` | Current device state (screen on/off, terminal sessions) |

```go
type StateStore struct {
    db *bolt.DB
}

func (s *StateStore) Write(bucket string, key string, value []byte) error
func (s *StateStore) Read(bucket string, key string) ([]byte, error)
func (s *StateStore) Delete(bucket string, key string) error
func (s *StateStore) List(bucket string) (map[string][]byte, error)
```

#### 2.4 Engine Manager

Manages BPP protocol handlers and routes messages between components.

**Architecture:**

```
┌──────────────┐     ┌──────────────────┐     ┌──────────────┐
│  Connection  │────▶│  Engine Manager  │────▶│  Sub-Services│
│  Engine      │     │  (Message Router)│     │  (Screen,    │
│  (RFC-0001)  │     │                  │     │   Terminal,  │
│              │     │  Method Registry │     │   GPIO, etc) │
│              │     │  capability.map  │     │              │
│              │     │  session.table   │     │              │
└──────────────┘     └──────────────────┘     └──────────────┘
```

- Inbound messages from Connection Engine → Engine Manager → routed to handler
- Outbound messages from components → Engine Manager → Connection Engine → client
- Each BPP method is registered with a handler function at startup

**Method registration pattern:**

```go
type MethodHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

func (em *EngineManager) RegisterMethod(method string, handler MethodHandler) {
    em.methods[method] = handler
}

// Registration at startup:
em.RegisterMethod("device.info", h.DeviceInfo)
em.RegisterMethod("device.stats", h.DeviceStats)
em.RegisterMethod("terminal.open", h.TerminalOpen)
em.RegisterMethod("screen.start", h.ScreenStart)
em.RegisterMethod("file.list", h.FileList)
```

#### 2.5 Plugin Host

Runs plugins as **sub-processes** communicating over stdin/stdout (JSON-RPC over stdio).

**Design rationale:**
- Sub-process isolation: a crash of one plugin does not take down the Runtime
- Language-agnostic: plugins can be written in any language
- Simplified memory management: no GC pressure from arbitrary plugin code
- Standard Unix: stdin/stdout, signals, exit codes — nothing exotic

**Plugin lifecycle:**

```go
type PluginProcess struct {
    Manifest  PluginManifest
    Cmd       *exec.Cmd
    Stdin     io.WriteCloser
    Stdout    io.ReadCloser
    Health    HealthStatus
}

// Protocol: JSON-RPC 2.0 over stdin/stdout
// Runtime → Plugin: {"jsonrpc":"2.0","method":"capability.info","id":1}
// Plugin → Runtime: {"jsonrpc":"2.0","result":{"...":"..."},"id":1}
// Plugin → Runtime: {"jsonrpc":"2.0","method":"event","params":{"type":"gpio.change","pin":12}}
```

#### 2.6 Device Manager

Orchestrates hardware-specific operations through a capability abstraction.

**Built-in capability providers:**

| Provider | Capabilities | Notes |
|----------|-------------|-------|
| Screen Capture | `screen.capture`, `screen.stream` | Uses x11/xvfb or KMS/DRM |
| Terminal | `terminal.open`, `terminal.exec`, `terminal.resize` | PTY multiplexer |
| GPIO | `gpio.read`, `gpio.write`, `gpio.watch` | Uses `/sys/class/gpio` or libgpiod |
| Camera | `camera.stream`, `camera.snapshot`, `camera.record` | Uses libcamera-vid/libcamera-still |
| Docker | `container.list`, `container.logs`, `container.exec` | Docker socket proxy |
| Filesystem | `file.list`, `file.read`, `file.write`, `file.delete` | Sandboxed to allowed paths |
| System | `device.info`, `device.stats`, `device.reboot` | OS-level info |

#### 2.7 mDNS Discovery

Advertises the device on the local network using mDNS/DNS-SD.

**Service type:** `_buzzpi._tcp`

**TXT records:**
```
device_id=dev_a1b2c3d4
friendly_name=living-room-pi
runtime_version=0.1.0
platform=raspberry-pi/5
capabilities=screen,terminal,gpio,camera,docker,filesystem
```

- Implements `github.com/hashicorp/mdns` for mDNS
- Sends goodbye packet on shutdown
- Re-announces on capability change (plugin loaded/unloaded)

#### 2.8 Screen Capture Service

The most performance-critical component. Streams the Raspberry Pi's graphical display to clients.

**Pipeline:**

```
Display (X11/DRM) → Capture → Encode (H.264/H.265) → Packetize → Send
```

**Layers:**
1. **Capture backend:** Platform-specific frame grabber (x11 via XShmGetImage, KMS/DRM via dumb buffers)
2. **Encoder:** Hardware-accelerated video encoding (Raspberry Pi: MMAL/OMX or V4L2 M2M H.264)
3. **Quality adapter:** Dynamically adjusts bitrate, resolution, and FPS based on network conditions
4. **Packetizer:** Splits encoded frames into BPP screen.data messages

**Quality levels:**

| Level | Max FPS | Max Resolution | Bitrate | Use Case |
|-------|---------|----------------|---------|----------|
| High | 30 | 1920x1080 | 5 Mbps | LAN / WiFi |
| Medium | 15 | 1280x720 | 1.5 Mbps | Remote / LTE |
| Low | 8 | 854x480 | 500 Kbps | Slow connection |
| Minimum | 5 | 640x360 | 150 Kbps | Edge / 3G |

#### 2.9 Terminal Manager

Full PTY multiplexer for multiple concurrent terminal sessions.

- Uses `github.com/creack/pty` for PTY creation
- Each terminal session runs as a separate shell process
- Input: stdin forwarding (keystrokes, paste, control sequences)
- Output: ANSI-aware diff encoding (send only changed regions)
- Resize: SIGWINCH forwarding to PTY
- Scrollback buffer: 10,000 lines in-memory per session

### 3. Data Flow: Client Request → Response

```
Client                    Runtime
  │                         │
  │──── BPP Request ───────▶│  Connection Engine receives
  │                         │  ────────────────
  │                         │  Deserialize envelope
  │                         │  Validate signature
  │                         │  Route to Engine Manager
  │                         │
  │                         │  Engine Manager
  │                         │  ────────────────
  │                         │  Lookup method handler
  │                         │  Check capabilities
  │                         │  Check session auth
  │                         │  Call handler
  │                         │
  │                         │  Method Handler
  │                         │  ────────────────
  │                         │  Execute (e.g., read GPIO)
  │                         │  Serialize result
  │                         │
  │◀──── BPP Response ─────│  Engine Manager wraps response
  │                         │  Connection Engine sends
```

### 4. Data Flow: Device Event → Client (Push)

```
Screen Capture              Connection Engine            Client
  │                             │                         │
  │── frame_ready() ──────────▶│                         │
  │                             │── screen.data event ───▶│
  │                             │                         │
  │── frame_ready() ──────────▶│                         │
  │                             │── screen.data event ───▶│
```

Events are pushed over the same WebSocket connection. The Connection Engine multiplexes method responses and events onto a single stream.

### 5. Security Model

**Identity:**
- Ed25519 key pair generated on first boot
- `device_id` = first 12 bytes of SHA-256(public key), base62-encoded
- Private key never leaves the device (stored in PID file with 0600 permissions)

**Pairing:**
- RFC-0003 (Pairing Protocol) — separate RFC for the pairing handshake
- Pairing creates a shared secret via X25519 + PAKE (SPAKE2+)

**Session:**
- Each paired client gets a session token (32-byte random, rotated every 24h)
- Session tokens are stored in BoltDB and sent over encrypted connection
- Token rotation invalidates old tokens

**Capability access:**
- Each capability declares its auth requirements:
  - `Public` — no auth needed (device.info)
  - `Paired` — must be paired (device.stats, terminal, file)
  - `Admin` — must be paired + admin role (reboot, factory reset)

**Plugin sandboxing:**
- Plugins run as sub-processes with reduced privileges
- No direct filesystem access outside plugin directory
- No network access unless explicitly granted in manifest
- Runtime kills plugins that exceed memory/CPU thresholds

### 6. Configuration

**Config file:** `/etc/buzzpi/runtime.yaml`

```yaml
runtime:
  device_name: "living-room-pi"  # Overrides auto-generated name

network:
  relay_servers:
    - turn:turn.buzzpi.cloud:3478
  listen_port: 0                  # 0 = random port
  mdns_enabled: true

connection:
  max_clients: 5
  session_timeout: 24h
  keepalive_interval: 30s

screen:
  capture_backend: "auto"        # auto, x11, kms, drm
  max_fps: 30
  default_quality: "high"

plugins:
  enabled: true
  directory: /var/lib/buzzpi/plugins/
  allow_network: false

logging:
  level: "info"                   # debug, info, warn, error
  file: /var/log/buzzpi/runtime.log
  max_size_mb: 100
  max_files: 5
```

**Config override hierarchy:**
1. Defaults (compiled into binary)
2. Config file (`/etc/buzzpi/runtime.yaml`)
3. Environment variables (`BUZZPI_DEVICE_NAME=kitchen-pi`)
4. CLI flags (`--device-name kitchen-pi`)

### 7. Startup Sequence

```
1. Parse config (file → env → flags)
2. Initialize logging
3. Generate/load device identity (PID file)
4. Open state store (BoltDB)
5. Start mDNS advertiser
6. Register all built-in capability providers
7. Start Plugin Host → scan and start plugins
8. Build capability map
9. Start Engine Manager (method handlers)
10. Start Connection Engine (RFC-0001)
11. Signal readiness to OS (systemd notify)
12. Enter main loop (monitor health, handle SIGHUP reload)
```

### 8. Graceful Shutdown Sequence

```
1. Receive SIGTERM (or SIGINT)
2. Stop Connection Engine (send close frames, wait for ACK, max 10s)
3. Notify clients: device.disconnecting event
4. Stop Plugin Host (SIGTERM all plugins, wait 5s, SIGKILL)
5. Stop all sub-services (screen, terminal, etc.)
6. Flush and close state store
7. Send mDNS goodbye packet
8. Exit with code 0
9. If not exited within 30s: SIGKILL self
```

### 9. Error Handling

**Structured error codes (BPP error codes):**

| Code | Description | HTTP Analog |
|------|-------------|-------------|
| `method_not_found` | Unknown method | 404 |
| `invalid_params` | Malformed parameters | 422 |
| `not_authenticated` | Session invalid | 401 |
| `not_authorized` | Insufficient role | 403 |
| `capability_unavailable` | Feature not available | 501 |
| `device_busy` | Resource in use | 409 |
| `internal_error` | Unexpected failure | 500 |
| `rate_limited` | Too many requests | 429 |
| `session_expired` | Token expired | 401 |

**Internal error handling pattern:**

```go
type RuntimeError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e *RuntimeError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
    ErrMethodNotFound     = &RuntimeError{Code: "method_not_found", Message: "method not registered"}
    ErrInvalidParams      = &RuntimeError{Code: "invalid_params", Message: "invalid parameters"}
    ErrNotAuthenticated   = &RuntimeError{Code: "not_authenticated", Message: "session required"}
    ErrNotAuthorized      = &RuntimeError{Code: "not_authorized", Message: "insufficient role"}
    ErrCapUnavailable     = &RuntimeError{Code: "capability_unavailable", Message: "capability not available"}
)
```

### 10. Resource Limits

| Resource | Limit | Behavior When Exceeded |
|----------|-------|------------------------|
| Concurrent clients | 5 (default, configurable) | `device_busy` error |
| Terminal sessions | 10 (default) | New session rejected |
| Plugin CPU | 25% of one core | Plugin killed, restarted once |
| Plugin memory | 128 MB | Plugin killed, restarted once |
| State store | 256 MB | No more writes, log warning |
| Event queue | 10,000 messages | Drop oldest, log warning |

### 11. Monitoring

**Health endpoint** (Unix domain socket: `/var/run/buzzpi/health.sock`):

```
GET /health → {"status":"ok","uptime":3600,"version":"0.1.0","clients":2,"plugins":3}
GET /health/component/{name} → {"name":"screen","status":"ok","fps":28}
```

**Metrics** (exposed via HTTP at localhost:9100/metrics for Prometheus):

```
buzzpi_clients_active{device_id="dev_..."} 2
buzzpi_plugins_running{device_id="dev_..."} 3
buzzpi_screen_fps{device_id="dev_..."} 28
buzzpi_cpu_percent{device_id="dev_..."} 12.5
buzzpi_memory_mb{device_id="dev_..."} 64.2
```

---

## Drawbacks

1. **Sub-process plugin model adds latency** — JSON-RPC over stdio for every plugin call has overhead. For high-frequency operations (GPIO toggling), this may be too slow. Mitigation: hot-path capabilities (screen capture, terminal) are built-in, not plugins.

2. **BoltDB is single-writer** — All writes go through BoltDB's single read-write transaction. This could become a bottleneck under high event throughput. Mitigation: BoltDB handles 10k+ writes/sec on a Pi 4 — sufficient for our volume.

3. **mDNS is LAN-only** — The discovery protocol does not extend to remote networks. Mitigation: pairing creates a persistent association; remote clients connect via relay.

---

## Rationale

1. **Why BoltDB over SQLite?** — Simpler API, pure Go (no CGo), single-file, well-proven on embedded devices. The Runtime does not need relational queries.

2. **Why sub-process plugins over WASM or Lua?** — Language-agnostic (Python, Rust, Node.js all valid plugin languages), stronger isolation (OS-level process boundaries), simpler debugging. WASM has poor I/O support; Lua is niche.

3. **Why Ed25519 over ECDSA?** — Faster key generation, smaller signatures (64 bytes vs 70-72), well-audited implementations in Go standard library (Go 1.20+).

---

## Prior Art

- **Home Assistant** — Python-based, plugin model via custom_components, uses WebSocket for real-time. Inspires our event sub pattern.
- **Pi-hole** — Lightweight Go daemon with embedded HTTP server. Inspires our minimal dependency approach.
- **Tailscale** — WireGuard-based mesh VPN with node identity. Inspires our Ed25519 identity model.
- **Grafana Agent** — Go daemon with component architecture. Inspires our lifecycle management.

---

## Unresolved Questions

1. **Plugin discovery** — Should the Runtime auto-update plugins (over-the-air), or require explicit installation? Resolved: explicit only for v0.x; OTA for v1.0.

2. **Multi-user sessions** — Can two clients simultaneously control the same terminal/screen? Initially: yes, screen streaming is multicast; terminal is single-owner with takeover.

3. **Firmware updates** — Should the Runtime self-update? Initially: no — use system package manager (apt). Future: buzzer daemon for OTA.

---

## Implementation Plan

| Phase | Milestone | Components |
|-------|-----------|------------|
| P0 | Minimal daemon | Supervisor, Core P2P, State Store, mDNS |
| P1 | Interactive | Terminal Manager, Engine Manager, BPP method handlers |
| P2 | Visual | Screen Capture Service, H.264 encoding |
| P3 | Extensible | Plugin Host, plugin API |
| P4 | Rich I/O | GPIO, Camera, Filesystem providers |
| P5 | Manageable | Docker provider, System monitoring, health endpoint |
| P6 | Observable | Metrics, structured logging, telemetry |

---

## References

- RFC-0001: Connection Engine
- RFC-0003: Pairing Protocol (planned)
- Engineering Book: architecture.md, plugin-system.md, event-bus.md, buzzai.md
- Protocol Book: BPP specification (all 24 chapters)
- Reference: configuration-reference.md, packet-types.md, event-catalog.md
