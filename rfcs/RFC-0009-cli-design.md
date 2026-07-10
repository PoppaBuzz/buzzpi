# RFC-0009: CLI Design

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | RFC-0002, RFC-0003, RFC-0006 |

## Summary

Define the BuzzPi CLI client — the primary developer tool and secondary user client for the BuzzPi Platform. The CLI enables device discovery, pairing, terminal access, file operations, plugin management, and debugging from the command line.

## Motivation

The CLI serves two critical roles:
1. **Primary development tool** — used throughout the implementation of the Runtime and protocol
2. **Secondary user client** — power users and developers who prefer the terminal over a GUI

The CLI validates the BPP protocol in practice before the Android app is built. Every BPP method the CLI uses is a BPP method the Android app will also use.

## Design

### 1. Command Structure

```
buzzpi [global flags] <command> [subcommand] [arguments]
```

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Config file path | `~/.buzzpi/config.yaml` |
| `--log-level` | Log verbosity (debug, info, warn, error) | `info` |
| `--format` | Output format (text, json, yaml) | `text` |
| `--timeout` | Request timeout | `30s` |
| `--version` | Show version | |

### Commands

```
Device Management:
  discover          Scan for BuzzPi devices on the LAN
  device info       Get device information
  device stats      Get device system statistics
  device monitor    Live device stats (top-like)

Pairing:
  pair              Pair with a device
  unpair            Unpair from a device
  pair list         List paired devices
  pair show         Show pairing details

Terminal:
  term open         Open an interactive terminal session
  term exec         Execute a command and print output
  term list         List active terminal sessions

Screen:
  screen start      Start screen streaming
  screen stop       Stop screen streaming
  screen status     Show screen streaming status

Files:
  file ls           List directory contents
  file cat          Print file contents
  file upload       Upload local file to device
  file download     Download device file to local
  file rm           Delete file on device

System:
  system info       Show system information
  system stats      Show system statistics
  system reboot     Reboot the device
  system shutdown   Shutdown the device

Plugin:
  plugin list       List installed plugins
  plugin install    Install a plugin from registry or file
  plugin uninstall  Remove a plugin
  plugin update     Update a plugin
  plugin show       Show plugin details

Session:
  session list      List active sessions
  session revoke    Revoke a session
  session refresh   Refresh current session

Config:
  config init       Create default configuration
  config show       Show current configuration
  config set        Set a configuration value
  config edit       Open config in $EDITOR

Diagnostics:
  doctor         Run system diagnostics
  ping           Ping a device (connectivity check)
  version        Show version information
  help           Show help for any command
```

### 2. Usage Examples

```
# Discover devices on the LAN
$ buzzpi discover
  DEVICE ID              NAME              PLATFORM        VERSION   TRANSPORT
  dev_a1b2c3d4           living-room-pi    raspberry-pi/5  0.1.0     LAN
  dev_e5f6g7h8           garage-pi         raspberry-pi/4  0.1.0     LAN

# Get device information
$ buzzpi device info dev_a1b2c3d4
  Device ID:       dev_a1b2c3d4
  Name:            living-room-pi
  Platform:        raspberry-pi/5
  Runtime:         0.1.0
  BPP Version:     1
  Uptime:          3d 14h 22m
  Capabilities:    screen, terminal, file, system, gpio

# Interactive terminal session
$ buzzpi term open dev_a1b2c3d4
  Connected to living-room-pi. Type 'exit' to close.
  pi@living-room:~$ neofetch
  pi@living-room:~$ htop

# Execute command non-interactively
$ buzzpi term exec dev_a1b2c3d4 -- "cat /etc/os-release"
  PRETTY_NAME="Raspbian GNU/Linux 12 (bookworm)"

# Pair with a device
$ buzzpi pair dev_a1b2c3d4
  ┌──────────────────────────────┐
  │                              │
  │   Enter PIN shown on device  │
  │                              │
  │   ┌──────────────────────┐   │
  │   │ PIN: [4 8 2 9 1 6]  │   │
  │   └──────────────────────┘   │
  │                              │
  │   [Verify]    [Cancel]       │
  └──────────────────────────────┘
  Device paired successfully!

# Monitor device stats (top-like)
$ buzzpi device monitor dev_a1b2c3d4
  living-room-pi (dev_a1b2c3d4) — LAN — 5ms RTT
  CPU: 12.5% ████░░░░░░░░  Temp: 52°C
  MEM: 45.2% █████████░░░  1.2GB/2.7GB
  DSK: 34.0% ███████░░░░░  12GB/35GB
  NET: ↑ 1.2 Mbps  ↓ 3.4 Mbps

# Install a plugin
$ buzzpi plugin install docker-manager
  Installing docker-manager v1.2.0...
  ✓ Manifest validated
  ✓ Process started
  ✓ Health check passed
  Plugin docker-manager installed successfully
  Capabilities added: container.list, container.logs, container.exec

# File operations
$ buzzpi file ls dev_a1b2c3d4 /home/pi
  -rw-r--r--  pi/pi    config.yaml    1284  2026-07-07
  drwxr-xr-x  pi/pi    photos/        4096  2026-07-06
  -rwxr-xr-x  pi/pi    script.sh      256   2026-07-05

$ buzzpi file upload ./photo.jpg dev_a1b2c3d4 /home/pi/photos/
  Uploading photo.jpg... ████████████████ 100%  2.4MB/s

# Diagnostics
$ buzzpi doctor
  BuzzPi CLI Doctor
  ─────────────────
  ✓ Config file: /home/user/.buzzpi/config.yaml
  ✓ mDNS: available (systemd-resolved)
  ✓ WebSocket: works
  ✓ State store: /home/user/.buzzpi/state.db (12KB)
  ✓ Paired devices: 2

  Recommendations:
  • None — everything looks good!
```

### 3. Output Formats

**Text (default):**
```
DEVICE ID            NAME              STATUS    TRANSPORT
dev_a1b2c3d4         living-room-pi    online    LAN
dev_e5f6g7h8         garage-pi         offline   —
```

**JSON (machine-readable):**
```json
{
  "devices": [
    {
      "device_id": "dev_a1b2c3d4",
      "friendly_name": "living-room-pi",
      "online": true,
      "transport": "lan",
      "rtt_ms": 5
    }
  ]
}
```

**YAML (configuration-friendly):**
```yaml
devices:
  - device_id: dev_a1b2c3d4
    friendly_name: living-room-pi
    online: true
    transport: lan
    capabilities:
      - screen
      - terminal
      - file
```

### 4. Interactive Features

**Progress bars for file transfers:**
```
Uploading buzzpi-agent-v0.1.0.tar.gz ... ████████████░░░ 67%  3.2MB/s
```

**Spinner for connection operations:**
```
⟳ Connecting to living-room-pi...
✓ Connected (LAN, 5ms RTT)
```

**Terminal UI for `device monitor`:**
```
living-room-pi (dev_a1b2c3d4) — LAN — 5ms RTT    Last updated: 2s ago
CPU: 12.5% ████░░░░░░░░░░░░░░  Temp: 52°C         Freq: 1.8GHz
MEM: 45.2% █████████░░░░░░░░░░  1.2GB / 2.7GB
DSK: 34.0% ███████░░░░░░░░░░░░  12GB / 35GB
NET: ↑ 1.2 Mbps  ↓ 3.4 Mbps
```

### 5. Configuration File

```yaml
# ~/.buzzpi/config.yaml
default_device: dev_a1b2c3d4
output_format: text
timeout: 30s

connection:
  relay_url: wss://relay.buzzpi.cloud
  preferred_transport: lan       # lan, relay, auto

pairing:
  pin_timeout: 120s

plugins:
  registry_url: https://plugins.buzzpi.dev
  install_dir: /var/lib/buzzpi/plugins

logging:
  level: info
  file: ~/.buzzpi/cli.log
```

### 6. Session Management

The CLI stores session tokens in `~/.buzzpi/sessions.json`:

```json
{
  "dev_a1b2c3d4": {
    "session_token": "sess_abc123...",
    "device_id": "dev_a1b2c3d4",
    "paired_at": "2026-07-07T12:00:00Z",
    "expires_at": "2026-07-08T12:00:00Z"
  }
}
```

Sessions are auto-refreshed on each command. Expired sessions prompt re-pairing.

### 7. Error Handling

```
$ buzzpi device stats unknown_device
  Error: device not found
  Suggestion: Run 'buzzpi discover' to find available devices

$ buzzpi term exec dev_a1b2c3d4 -- "sleep 60"
  Error: request timed out after 30s
  Suggestion: Use --timeout 120s for long-running commands
```

**Exit codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid input |
| 3 | Device not found |
| 4 | Not paired |
| 5 | Session expired |
| 6 | Connection failed |
| 7 | Permission denied |

---

## Drawbacks

1. **CLI is a secondary client** — The CLI will never match the Android app's feature set (no screen streaming, limited touch interaction). This is acceptable — the CLI targets developers.

2. **Session token storage in JSON** — Plaintext tokens on disk are a security concern. Mitigation: file permissions 0600, future support for OS keychain integration.

3. **Terminal UI limits** — The `device monitor` terminal UI is platform-dependent. Windows terminal rendering differs from macOS/Linux. Mitigation: use `bubbletea` for TUI, fall back to simple text on unsupported terminals.

---

## Rationale

1. **Why Cobra CLI?** Most popular Go CLI framework. Auto-generated help, completions, subcommand routing. Well-audited, widely used (Kubernetes, Docker, Hugo).

2. **Why `bubbletea` for TUI?** Pure Go terminal UI framework. Cross-platform, well-maintained, composable. Suitable for `device monitor` and interactive progress.

3. **Why JSON output?** Machine-readable output enables scripting (`jq`, shell pipelines). `--format json` is standard in modern CLIs (Docker, kubectl, gh).

---

## Implementation Plan

| Phase | Milestone | Commands |
|-------|-----------|----------|
| P0 | Skeleton | `discover`, `device info`, `device stats`, `version` |
| P1 | Pairing | `pair`, `unpair`, `pair list`, `session list` |
| P2 | Terminal | `term open`, `term exec`, `term list` |
| P3 | Files | `file ls`, `file cat`, `file upload`, `file download`, `file rm` |
| P4 | Plugins | `plugin list`, `plugin install`, `plugin uninstall`, `plugin update` |
| P5 | System | `system info`, `system stats`, `system reboot`, `system shutdown` |
| P6 | Polish | `doctor`, `config init`, `config set`, `device monitor`, completions |

---

## References

- RFC-0002: Runtime Architecture (BPP methods the CLI calls)
- RFC-0003: Pairing Protocol (pair/unpair flow)
- RFC-0006: Cloud Relay (remote connection)
- Reference: cli-reference.md, bpp-client-api.md
- Go packages: github.com/spf13/cobra, github.com/charmbracelet/bubbletea
