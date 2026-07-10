# Compatibility Matrix

**Platform, OS, and hardware compatibility for the BuzzPi Platform.**

This document defines which combinations of agent, client, platform, and transport are supported, experimental, or planned.

---

## 1. Agent Platform Support

The **BuzzPi Agent** (Go daemon) runs on the following platforms:

| Platform | Architectures | Min Version | Status | Notes |
|----------|--------------|-------------|--------|-------|
| Raspberry Pi OS (Debian Bookworm) | arm64, armv7l | 12 (Bookworm) | **Supported** | Primary target |
| Raspberry Pi OS (Debian Bullseye) | arm64, armv7l | 11 (Bullseye) | **Supported** | Backward compat |
| Ubuntu Server LTS | amd64, arm64 | 22.04 LTS | **Supported** | Secondary target |
| Ubuntu Server LTS | amd64, arm64 | 24.04 LTS | **Supported** | Ongoing |
| Debian Linux | arm64, amd64 | 12 (Bookworm) | **Supported** | |
| Debian Linux | arm64, amd64 | 11 (Bullseye) | **Supported** | |
| Fedora Linux | amd64, arm64 | 39+ | **Experimental** | Not CI-tested |
| Arch Linux ARM | arm64, armv7l | Rolling | **Experimental** | Community-maintained |
| Alpine Linux | amd64, arm64 | 3.19+ | **Planned** | Container-optimized |
| OpenWrt | mips | 23.05+ | **Planned** | Low-resource target |
| macOS | amd64, arm64 | 14 (Sonoma) | **Development only** | Agent runs but no GPIO/accessory support |
| Windows (WSL2) | amd64 | Windows 10 22H2+ | **Development only** | Agent runs under WSL2 |

---

## 2. Client Platform Support

### 2.1 Android Client

| Android Version | API Level | Architecture | Status | Notes |
|----------------|-----------|--------------|--------|-------|
| 14 | 34 | arm64-v8a | **Supported** | Primary target |
| 13 | 33 | arm64-v8a | **Supported** | |
| 12 | 31-32 | arm64-v8a | **Supported** | |
| 11 | 30 | arm64-v8a | **Supported** | Min recommended |
| 10 | 29 | arm64-v8a | **Experimental** | |
| 9 | 28 | arm64-v8a, armeabi-v7a | **Planned** | |
| 8 | 26-27 | armeabi-v7a | **Not planned** | Insufficient WebRTC support |

### 2.2 CLI Client

The BuzzPi CLI runs anywhere Go compiles, subject to mDNS support:

| OS | Architectures | mDNS | Status | Notes |
|----|--------------|------|--------|-------|
| macOS | amd64, arm64 | Native (resolved) | **Supported** | Primary dev target |
| Linux | amd64, arm64 | avahi-daemon | **Supported** | |
| Linux | armv7l | avahi-daemon | **Supported** | |
| WSL2 (Windows) | amd64 | avahi-daemon via systemd | **Experimental** | Requires WSL2 systemd |
| Windows (native) | amd64 | Third-party | **Not planned** | Use WSL2 instead |

### 2.3 Desktop Client (Planned)

| OS | Architectures | Status | Notes |
|----|--------------|--------|-------|
| macOS | amd64, arm64 | **Planned** | Native app (Swift/SwiftUI) |
| Linux | amd64, arm64 | **Planned** | GTK or Electron |
| Windows | amd64 | **TBD** | After other platforms |

---

## 3. Raspberry Pi Hardware Support

| Model | SoC | RAM | Agent | Screen Stream | Notes |
|-------|-----|-----|-------|------|-------|
| Pi 5 | BCM2712 (Cortex-A76) | 4-8GB | **Full** | **Full** | Recommended |
| Pi 4 B | BCM2711 (Cortex-A72) | 1-8GB | **Full** | **Full** | Supported |
| Pi 400 | BCM2711 (Cortex-A72) | 4GB | **Full** | **Full** | Same as Pi 4 |
| Pi 3 B+ | BCM2837B0 (Cortex-A53) | 1GB | **Full** | **Limited** | Screen at ≤10 FPS |
| Pi 3 B | BCM2837 (Cortex-A53) | 1GB | **Full** | **Limited** | Screen at ≤10 FPS |
| Pi Zero 2 W | RP3A0 (Cortex-A53) | 512MB | **Full** | **Limited** | Headless-optimized |
| Pi Zero W | BCM2835 (ARM11) | 512MB | **Limited** | **No** | Terminal/file only |
| Pi 2 B v1.1 | BCM2836 (Cortex-A7) | 1GB | **Limited** | **No** | Terminal/file only |
| Compute 4 | BCM2711 (Cortex-A72) | 1-8GB | **Full** | **Full** | Same as Pi 4 |
| Compute 3+ | BCM2837B0 (Cortex-A53) | 1GB | **Full** | **Limited** | Same as Pi 3 B+ |

### 3.1 Screen Stream Quality by Hardware

| Model | 720p @ 15fps | 720p @ 30fps | 1080p @ 15fps | 1080p @ 30fps |
|-------|-------------|-------------|--------------|--------------|
| Pi 5 | ✓ | ✓ | ✓ | ✓ |
| Pi 4 | ✓ | ✓ | ✓ | ~ |
| Pi 3 B+ | ✓ | ~ | ~ | ✗ |
| Pi 3 B | ✓ | ~ | ✗ | ✗ |
| Pi Zero 2 W | ✓ | ✗ | ✗ | ✗ |

Legend: ✓ = Smooth | ~ = Usable with stutters | ✗ = Not recommended

---

## 4. Transport Compatibility

| Client | LAN (mDNS) | LAN (Manual IP) | Cloud Relay | P2P (TURN) |
|--------|-----------|----------------|-------------|------------|
| Android | **Supported** | **Supported** | **Supported** | **Planned** |
| CLI | **Supported** | **Supported** | **Supported** | **Planned** |
| Desktop | **Planned** | **Planned** | **Planned** | **Planned** |

---

## 5. Plugin SDK Compatibility

| SDK Language | Status | Notes |
|-------------|--------|-------|
| Go | **Supported** | Primary SDK — best performance and sandboxing |
| Python | **Experimental** | Heavier runtime, subprocess-based |
| Rust | **Planned** | Near Go-level performance, harder sandboxing |
| Node.js | **TBD** | Resource-heavy for constrained devices |

---

## 6. BPP Protocol Versioning

### 6.1 Version Matrix

| BPP Version | Min Agent | Min Android | Min CLI | Status |
|-------------|-----------|-------------|---------|--------|
| 1.0 | 0.1.0 | 0.1.0 | 0.1.0 | **Development** |

All clients and agents negotiate BPP version during the capability exchange handshake. If versions are incompatible, the client receives a clear error:

```
Error: Incompatible protocol version
Client supports BPP 1, Agent requires BPP 2
Update agent: buzzpi agent update
```

### 6.2 Backward Compatibility Policy

- **Patch bumps** (1.0.x): Fully backward-compatible. No client changes needed.
- **Minor bumps** (1.x.0): Backward-compatible with deprecation notices. Old clients continue working but lose access to new features.
- **Major bumps** (2.0.0): Breaking changes. Old clients cannot connect. Migration guide provided.

---

## 7. Dependency Requirements

### 7.1 Agent Dependencies

| Dependency | Min Version | Required | Notes |
|-----------|-------------|----------|-------|
| Linux kernel | 5.10+ | ✓ | For seccomp, pidfd, etc. |
| systemd | 247+ | ✓ | For service management |
| avahi-daemon OR systemd-resolved | — | ✓ for mDNS | For LAN discovery |
| OpenGL ES 2.0 | — | ✗ | For screen capture (optional) |
| libseccomp | 2.5+ | ✗ | For plugin sandboxing |
| libc (glibc) | 2.31+ | ✓ | |

### 7.2 Android Dependencies

| Dependency | Min Version | Required | Notes |
|-----------|-------------|----------|-------|
| Kotlin | 1.9+ | ✓ | For client development |
| WebRTC | M128+ | ✓ | For screen streaming |
| Jetpack Compose | 2024.01+ | ✓ | For UI |
| BouncyCastle | 1.77+ | ✓ | For TLS/crypto |
| Material3 | 1.2+ | ✓ | For UI components |

### 7.3 CLI Dependencies

| Dependency | Min Version | Required | Notes |
|-----------|-------------|----------|-------|
| Go | 1.22+ | ✓ | For compilation |
| Cobra | 1.8+ | ✓ | CLI framework |
| Bubbletea | 0.10+ | ✗ | For terminal UI (optional) |

---

## 8. Endian & Encoding

| Property | Value |
|----------|-------|
| BPP wire format byte order | Big-endian (network byte order) |
| JSON encoding | UTF-8 |
| String encoding in BPP | UTF-8 |
| Numeric encoding | IEEE 754 (float32, float64) |
| Timestamp format | RFC 3339 / ISO 8601 |

---

## 9. Network Compatibility

### 9.1 Ports (Outbound)

| Destination | Port | Protocol | Required For |
|------------|------|----------|-------------|
| Agent (peer) | 10104 | TCP | LAN BPP connections |
| Relay server | 443 / 10106 | TCP (WSS/WS) | Cloud relay connections |
| Plugin registry | 443 | TCP (HTTPS) | Plugin installation |
| mDNS | 5353 | UDP | LAN device discovery |

### 9.2 Network Topologies

See [Network Topologies](network-topologies.md) for details on supported network configurations and requirements.

---

## 10. Version Support Lifecycle

| Phase | Definition | Example |
|-------|-----------|---------|
| **Active** | Full support, CI-tested, bugs fixed | Pi 5, Android 14, Debian 12 |
| **Maintenance** | Critical fixes only, no new features | Pi 4, Android 13 |
| **Legacy** | Best-effort, community support | Pi 3 B+, Android 11 |
| **EOL** | Not supported, may not work | Pi 2, Android 8 |

---

## 11. Future Targets

Platforms under evaluation for future releases:

| Platform | Priority | Rationale |
|----------|----------|-----------|
| Orange Pi 5 (RK3588) | Medium | Popular, powerful SBC alternative |
| RISC-V (VisionFive 2) | Low | Early-stage ecosystem |
| FreeBSD | Low | Different security model |
| Docker (agent container) | Medium | Easy deployment on any Linux |
| Home Assistant Add-on | Medium | Large Raspberry Pi user base |

---

## References

- Reference: Network Topologies (lan/relay/p2p profiles)
- Reference: Environment Reference (version env vars)
- Reference: Platform Reference (build targets)
- RFC-0002: Runtime Architecture
- RFC-0008: Implementation Roadmap
