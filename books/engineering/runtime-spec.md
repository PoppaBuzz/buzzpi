# Runtime Specification

The BuzzPi Runtime is the on-device software that runs on every managed Linux device. It is a single binary with minimal dependencies, written in Go.

## Design Goals

| Goal | Requirement |
|------|-------------|
| Minimal footprint | < 10MB binary, < 30MB RAM idle |
| Zero config | No configuration files. Install and run. |
| Self-updating | Runtime can update itself without manual intervention |
| Resilient | Recovers from crashes, network loss, power loss |
| Secure | Minimal attack surface, no open ports, all traffic encrypted |
| Cross-platform | arm64, armv7 (Pi), amd64 (testing) |
| Single binary | No external dependencies (Python, Node, etc.) |

---

## Components

```
buzzpi-runtime
├── main.go                   # Entry point, signal handling
├── relay/
│   ├── client.go             # Relay WebSocket client
│   ├── heartbeat.go          # Heartbeat management
│   └── reconnect.go          # Reconnection with backoff
├── connection/
│   ├── engine.go             # WebRTC connection engine
│   ├── signaling.go          # WebRTC signaling handler
│   ├── datachannel.go        # Data channel management
│   └── screen.go             # Screen capture and encoding
├── terminal/
│   ├── pty.go                # PTY management (go
│   ├── session.go            # Terminal session
│   └── resize.go             # Terminal resize handling
├── services/
│   ├── manager.go            # Service lifecycle manager
│   ├── health.go             # Health check runner
│   └── log.go                # Log streaming
├── extensions/
│   ├── manager.go            # Extension lifecycle
│   ├── sandbox.go            # Extension isolation
│   └── registry.go           # Installed extension registry
├── pairing/
│   ├── handler.go            # Pairing protocol
│   ├── code.go               # Pairing code generation
│   └── identity.go           # Device identity management
├── update/
│   ├── updater.go            # Self-update mechanism
│   └── rollback.go           # Rollback on failed update
├── files/
│   ├── transfer.go           # File transfer handler
│   └── permissions.go        # Path validation
└── device/
    ├── info.go               # Hardware/OS information
    ├── power.go              # Restart/shutdown
    └── monitor.go            # Temperature, storage monitoring
```

---

## Startup Sequence

```
1. Binary starts
2. Read identity key from persistent storage
   │
   ├── First run?
   │   ├── Yes: Generate identity keypair (Ed25519)
   │   │         Save to /var/lib/buzzpi/identity.key
   │   │         Print pairing code to stdout
   │   └── No:  Load existing identity key
   │
   ├── Read device ID from /var/lib/buzzpi/device.id
   │   (Created during first successful pairing)
   │
   ├── Start monitor (temperature, storage)
   │
   ├── Open WebSocket to Relay Server
   │   ├── Success: Register device, start heartbeat
   │   └── Failure: Retry with backoff, log warning
   │
   ├── If registered (paired and known to relay):
   │   ├── Start service manager
   │   └── Start extension manager
   │
   └── If not paired:
       └── Wait for pairing (print code, listen for verification)
```

---

## Screen Capture

Screen streaming is the most technically demanding Runtime component.

### Requirements

| Specification | Target |
|--------------|--------|
| Capture method | DRM/KMS, fbdev, or VNC (fallback) |
| Codec | H.264 (hardware-encoded where available) |
| Resolution | Native device resolution |
| Frame rate | 15 fps (adaptive, 5-30 based on bandwidth) |
| Latency | <500ms p95 from capture to display |
| Bandwidth | 1-5 Mbps adaptive |

### Capture Pipeline

```
 Framebuffer  →  Encoder  →  Packetizer  →  WebRTC Track
      │             │             │              │
      │        Hardware H.264    RTP/RTX     DTLS-SRTP
      │        (VideoCore on Pi)
      │        Software fallback
      │        (x264 baseline)
```

### Hardware Acceleration

| Platform | Encoder | Status |
|----------|---------|--------|
| Raspberry Pi (VideoCore) | MMAL/IL H.264 | Primary target |
| Raspberry Pi (V4L2) | V4L2 H.264 | Fallback |
| Other Linux | Software x264 | Fallback |
| Future: V4L2 | Hardware H.264 | When available |

### Adaptive Bitrate

The Runtime monitors WebRTC statistics and adjusts encoding parameters:
- High bandwidth: 1080p, 30fps, 5 Mbps
- Medium bandwidth: 720p, 15fps, 2 Mbps
- Low bandwidth: 480p, 10fps, 1 Mbps
- Minimum: 240p, 5fps, 500 Kbps

---

## Service Management

The Runtime manages services through the service manager:

| Function | Implementation |
|----------|---------------|
| Start | `systemctl start <service>` or exec process |
| Stop | `systemctl stop <service>` or signal |
| Restart | `systemctl restart <service>` or exec |
| Status | systemctl status or process existence check |
| Logs | journalctl or stdout/stderr capture |
| Auto-start | systemd unit or init script |
| Health check | Configurable command/HTTP endpoint, default 30s interval |

Services are defined by Extensions (see extensions manager).

---

## Self-Update

### Update Mechanism

1. Runtime polls update server daily: `GET https://releases.buzzpi.dev/runtime/latest`
2. If newer version available, downloads binary to `/tmp/buzzpi-runtime-update`
3. Verifies GPG signature
4. Replaces current binary atomically (rename)
5. Sends `SIGTERM` to self
6. Systemd (or equivalent) restarts the new binary

### Rollback

- Previous binary is preserved at `/var/lib/buzzpi/runtime/previous/`
- If new binary crashes within 5 minutes of startup, rollback is automatic
- Manual rollback via `buzzpi-runtime rollback` flag on next restart

---

## Filesystem Layout

```
/var/lib/buzzpi/
├── identity.key          # Ed25519 private key (device identity)
├── device.id             # Device UUID (assigned on pairing)
├── config.json           # Runtime configuration (auto-managed)
├── extensions/           # Installed extensions
│   ├── docker-manager/   # Example extension
│   │   ├── manifest.yaml
│   │   └── extension.so
│   └── ...
├── runtime/              # Update artifacts
│   └── previous/         # Previous binary for rollback
├── logs/                 # Runtime logs (rotated, 1MB)
│   ├── runtime.log
│   └── extensions/
└── tmp/                  # Temporary files (cleared on restart)
```

## Resource Limits

| Resource | Limit | Behavior |
|----------|-------|----------|
| CPU | 10% of one core (idle) | Background priority |
| Memory | 30MB (idle), 100MB (streaming) | Trigger GC if exceeded |
| Storage (logs) | 1MB, rotated | Delete oldest |
| Storage (screenshots) | None explicitly | User-configurable |
| Network (screen) | 5 Mbps max | Adaptive bitrate |
| Open files | 100 max | Runtime-managed |
