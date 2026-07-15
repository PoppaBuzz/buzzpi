# Platform Reference

Every concrete fact about the BuzzPi Platform — filesystem layout, environment variables, URLs, permissions, and build targets.

## Runtime Filesystem Layout

```
/var/lib/buzzpi/
├── identity.key              # Ed25519 private key (device identity)
│                             # Permissions: 0600, owner: root
├── device.id                 # Device UUID (assigned on first run)
│                             # Format: UUID v7
├── config.json               # Runtime configuration (optional, auto-managed)
│                             # Permissions: 0644, owner: root
├── authorized_clients/       # One file per paired client
│   ├── <client_public_key_hex>  # Client public key, line-delimited
│   └── ...
│                             # Permissions: 0644, owner: root
├── extensions/               # Installed extensions
│   ├── <extension_id>/       # One directory per extension
│   │   ├── manifest.yaml     # Extension manifest
│   │   ├── <binary>          # Plugin binary (if native)
│   │   └── config.json       # Extension-specific config
│   └── ...
│                             # Permissions: 0755 dirs, 0644 files
├── logs/                     # Runtime logs
│   ├── runtime.log           # Main runtime log (rotated, max 1MB)
│   └── extensions/           # Per-extension logs
│       └── <extension_id>.log
│                             # Permissions: 0644
├── stats/                    # System metrics storage
│   └── <yyyy-mm-dd>.db       # SQLite database per day
│                             # Max total: 100MB
├── tmp/                      # Temporary files (cleared on restart)
└── runtime/                  # Update artifacts
    ├── previous/             # Previous binary for rollback
    └── current/              # Current binary (symlink target)
```

## Environment Variables

### Runtime

| Variable | Default | Description |
|----------|---------|-------------|
| `BUZZPI_RELAY_URL` | `wss://jphat.net/buzzpi/relay/ws` | Relay server WebSocket URL |
| `BUZZPI_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warning`, `error` |
| `BUZZPI_LOG_DIR` | `/var/lib/buzzpi/logs` | Log directory |
| `BUZZPI_CONFIG_DIR` | `/var/lib/buzzpi` | Configuration directory |
| `BUZZPI_EXTENSIONS_DIR` | `/var/lib/buzzpi/extensions` | Extensions directory |
| `BUZZPI_DEVICE_ID` | (auto-generated) | Override device ID (testing only) |
| `BUZZPI_UPDATE_INTERVAL` | `24h` | Update check interval |
| `BUZZPI_UPDATE_ENABLED` | `true` | Enable auto-update |
| `BUZZPI_MAX_LOG_SIZE` | `1048576` | Max log size before rotation (bytes) |
| `BUZZPI_HEARTBEAT_INTERVAL` | `30` | WebSocket ping interval (seconds) |

### App

| Variable | Description |
|----------|-------------|
| `BUZZPI_API_BASE` | Relay API base URL (for development) |
| `BUZZPI_REGISTRY_URL` | Extension registry URL |
| `BUZZPI_ENV` | Environment: `production`, `staging`, `development` |

### CLI

| Variable | Default | Description |
|----------|---------|-------------|
| `BUZZPI_TOKEN` | — | API token (or use `--token` flag) |
| `BUZZPI_RELAY_URL` | `wss://jphat.net/buzzpi/relay/ws` | Relay server URL |

## Android Permissions

| Permission | Purpose | Required |
|------------|---------|----------|
| `INTERNET` | WebSocket and WebRTC connectivity | Yes (core) |
| `ACCESS_NETWORK_STATE` | Network status detection | Yes (core) |
| `ACCESS_WIFI_STATE` | Local network mDNS discovery | Yes (discovery) |
| `CHANGE_WIFI_MULTICAST_STATE` | mDNS multicast | Yes (discovery) |
| `FOREGROUND_SERVICE` | Persistent connection for notifications | Yes (notifications) |
| `POST_NOTIFICATIONS` | Push notification display | Yes (notifications) |
| `CAMERA` | QR code scanning (pairing) | No (pairing fallback) |
| `BLUETOOTH` | Future BLE discovery | No (future) |
| `BLUETOOTH_ADMIN` | Future BLE discovery | No (future) |
| `VIBRATE` | Haptic feedback | No (UX preference) |
| `USE_BIOMETRIC` | App unlock | No (security preference) |
| `READ_EXTERNAL_STORAGE` | File uploads | No (file transfer) |
| `WRITE_EXTERNAL_STORAGE` | File downloads | No (file transfer) |

## Build Targets

### Runtime

| Target | Architecture | OS | Build Command |
|--------|-------------|----|----------------|
| `runtime-linux-arm64` | ARM64 | Linux | `GOOS=linux GOARCH=arm64 go build` |
| `runtime-linux-armv7` | ARMv7 | Linux | `GOOS=linux GOARCH=arm GOARM=7 go build` |
| `runtime-linux-amd64` | AMD64 | Linux | `GOOS=linux GOARCH=amd64 go build` |

Minimum Go version: 1.22

### App

| Target | Min SDK | Build System |
|--------|---------|-------------|
| Android | Android 13 (API 33) | Gradle + Kotlin 2.0 |
| Target SDK | Android 15 (API 35) | — |

### CLI

| Platform | Package |
|----------|---------|
| Linux (arm64) | `buzzpi-linux-arm64` |
| Linux (amd64) | `buzzpi-linux-amd64` |
| macOS (arm64) | `buzzpi-darwin-arm64` |
| macOS (amd64) | `buzzpi-darwin-amd64` |
| Windows (amd64) | `buzzpi-windows-amd64.exe` |

## Protocol Schema Version Strings

| Schema | Current Version | Location |
|--------|----------------|----------|
| BPP Message Envelope | 1 | All messages |
| BPP Identity Protocol | 1 | Layer 1 specification |
| BPP Transport Protocol | 1 | Layer 2 specification |
| BPP Terminal Service | 1 | Chapter 10 specification |
| BPP Screen Service | 1 | Chapter 11 specification |
| BPP File Service | 1 | Chapter 12 specification |
| BPP Stats Service | 1 | Chapter 13 specification |
| BPP GPIO Service | 1 | Chapter 14 specification |
| BPP Docker Service | 1 | Chapter 15 specification |
| BPP System Services | 1 | Chapter 16 specification |
| BPP Log Service | 1 | Chapter 17 specification |
| BPP Camera Service | 1 | Chapter 18 specification |
| BPP Custom RPC | 1 | Chapter 19 specification |
| BPP Capabilities | 1 | Layer 4 specification |
| Plugin Manifest | 1 | Chapter 22 specification |
| Extension Manifest | 1 | Chapter 23 specification |

## Service Dependencies

| Component | Runtime Dependencies | Build Dependencies |
|-----------|---------------------|-------------------|
| Runtime | None (static binary) | Go 1.22+ |
| App | Android 13+ | JDK 17+, Android SDK 35 |
| Relay Server | None (static binary) | Go 1.22+ |
| CLI | None (static binary) | Go 1.22+ |
| Extensions | Runtime (IPC) | Language-specific |
| SDK | None (library) | Language-specific |

## Supported Devices

| Device | Architecture | Runtime Support | Screen Capture |
|--------|-------------|-----------------|----------------|
| Raspberry Pi 5 | arm64 | Full | DRM (VideoCore VII) |
| Raspberry Pi 4 | arm64 | Full | DRM (VideoCore VI) |
| Raspberry Pi 3 | arm64 / armv7 | Full | DRM (VideoCore IV) |
| Raspberry Pi Zero 2 W | arm64 | Full | DRM (VideoCore IV) |
| Raspberry Pi Zero W | armv6 | Partial (no hw encoding) | fbdev |
| Raspberry Pi 400 | arm64 | Full | DRM (VideoCore VI) |
| Any Linux (x86_64) | amd64 | Full | DRM/x11/Wayland |
| Any Linux (arm64) | arm64 | Full | DRM (if available) |
