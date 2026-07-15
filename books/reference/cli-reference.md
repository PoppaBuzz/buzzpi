# CLI Reference

**Every CLI command, flag, and argument across all BuzzPi platform tools.**

---

## Runtime CLI (`buzzpi-runtime`)

The Runtime binary doubles as a CLI for both daemon mode and administration.

### Main Commands

```
buzzpi-runtime [command] [flags]
```

| Command | Description |
|---------|-------------|
| `start` | Start the Runtime daemon (foreground) |
| `service` | Install/start/stop/restart as systemd service |
| `config` | View and validate configuration |
| `status` | Print current device status |
| `pair` | Initiate or show pairing state |
| `log` | View Runtime logs |
| `update` | Check for and install updates |
| `plugin` | Manage plugins |
| `version` | Print version information |
| `help` | Show help for any command |

---

### buzzpi-runtime start

Start the Runtime daemon in the foreground.

```
buzzpi-runtime start [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `/etc/buzzpi/config.yaml` | Path to configuration file |
| `--no-mdns` | `false` | Disable mDNS advertisement |
| `--no-connect` | `false` | Start without connecting to relay |
| `--log-level` | `"info"` | Log level: debug, info, warn, error |
| `--pprof` | `""` | Enable pprof on address (e.g., `localhost:6060`) |

---

### buzzpi-runtime service

Manage the systemd service.

```
buzzpi-runtime service <command> [flags]
```

| Subcommand | Description |
|------------|-------------|
| `install` | Install systemd unit file and enable on boot |
| `start` | Start the systemd service |
| `stop` | Stop the systemd service |
| `restart` | Restart the systemd service |
| `status` | Show systemd service status |
| `uninstall` | Remove systemd unit file |
| `logs` | Show recent service logs (`journalctl -u buzzpi-runtime`) |

| Flag (install) | Default | Description |
|----------------|---------|-------------|
| `--user` | `buzzpi` | System user to run as |
| `--group` | `buzzpi` | System group to run as |
| `--binary` | `/usr/local/bin/buzzpi-runtime` | Path to Runtime binary |
| `--config` | `/etc/buzzpi/config.yaml` | Path to config file |
| `--env-file` | `/etc/buzzpi/env` | Environment file for secrets |

---

### buzzpi-runtime config

View and validate the Runtime configuration.

```
buzzpi-runtime config <subcommand> [flags]
```

| Subcommand | Description |
|------------|-------------|
| `validate` | Validate configuration file |
| `dump` | Print effective configuration (merged defaults + file + env) |
| `defaults` | Print default configuration |
| `show` | Show a specific config value: `config show network.relay_url` |

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `/etc/buzzpi/config.yaml` | Path to config file |
| `--format` | `"yaml"` | Output format: yaml, json |

---

### buzzpi-runtime status

Display current device status.

```
buzzpi-runtime status [flags]
```

```yaml
# Example output (YAML format)
device_id: "dev_abc12345"
friendly_name: "Kitchen Pi"
model: "Raspberry Pi 5"
runtime_version: "0.1.0"
state: "online"          # online | offline | pairing | unpaired
uptime_seconds: 86400

connection:
  transport: "direct"    # direct | relay | none
  rtt_ms: 12
  relay_connected: true

system:
  cpu_percent: 23
  memory_mb: 512
  memory_total_mb: 8192
  temperature_celsius: 65.2
  disk_percent: 72

services:
  terminal: "running"
  logs: "running"
  screen: "disabled"

plugins_installed: 0
pairing_code: "A3K9M7"  # Only shown when in pairing mode
```

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `"yaml"` | Output format: yaml, json, text |

---

### buzzpi-runtime pair

Manage device pairing.

```
buzzpi-runtime pair <subcommand> [flags]
```

| Subcommand | Description |
|------------|-------------|
| `show` | Show current pairing information |
| `generate` | Generate a new pairing code (displays on console) |
| `revoke` | Remove all pairings (factory reset pairing state) |
| `display` | Display current pairing code using device output (HDMI blink, LED pattern) |

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `"text"` | Output format: text, json, qr (generate QR code as ASCII) |

```
# Example: generate and display pairing QR code
buzzpi-runtime pair generate --format qr
```

---

### buzzpi-runtime log

View Runtime logs.

```
buzzpi-runtime log [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--tail` | `50` | Number of recent lines |
| `--follow` | `false` | Follow log output |
| `--level` | `""` | Filter by level: debug, info, warn, error |
| `--since` | `""` | Show logs since time: 5m, 2h, 2026-07-07T12:00:00Z |
| `--service` | `true` | Read from systemd journal |

---

### buzzpi-runtime update

Check and install Runtime updates.

```
buzzpi-runtime update <subcommand> [flags]
```

| Subcommand | Description |
|------------|-------------|
| `check` | Check for available updates |
| `apply` | Download and install update (restarts Runtime) |
| `rollback` | Revert to previous version |

| Flag (apply) | Default | Description |
|---------------|---------|-------------|
| `--version` | `"latest"` | Specific version to install |
| `--channel` | `"stable"` | Update channel: stable, beta, nightly |
| `--no-restart` | `false` | Download but don't restart |

```
# Example: update to specific version
buzzpi-runtime update apply --version v0.1.1 --channel stable
```

---

### buzzpi-runtime plugin

Manage plugins/extensions.

```
buzzpi-runtime plugin <subcommand> [flags]
```

| Subcommand | Description |
|------------|-------------|
| `list` | List installed plugins |
| `install` | Install a plugin |
| `uninstall` | Remove a plugin |
| `start` | Start a plugin |
| `stop` | Stop a plugin |
| `restart` | Restart a plugin |
| `permissions` | Manage plugin permissions |

| Flag (install) | Default | Description |
|----------------|---------|-------------|
| `--source` | `"registry"` | Plugin source: registry, path |
| `--version` | `"latest"` | Plugin version |

```
# Example: list plugins
buzzpi-runtime plugin list

# Example: install from filesystem
buzzpi-runtime plugin install --source path ./my-plugin/manifest.yaml
```

---

### buzzpi-runtime version

Print version information.

```
buzzpi-runtime version [flags]
```

```
# Example output
BuzzPi Runtime v0.1.0
Build: go1.22.5 linux/arm64
Commit: a1b2c3d4e5f6
Date: 2026-07-07T12:00:00Z
Channel: stable
BPP: 1.0
```

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `"text"` | Output format: text, json |

---

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error |
| `3` | Connection error |
| `4` | Permission denied |
| `5` | Service error |
| `6` | Plugin error |
| `7` | Update error |
| `8` | Timeout |
| `9` | Signal terminated |

---

### Signal Handling

| Signal | Runtime Behavior |
|--------|------------------|
| `SIGTERM` | Graceful shutdown: stop services → disconnect from relay → save state → exit |
| `SIGINT` | Same as SIGTERM |
| `SIGHUP` | Reload configuration |
| `SIGUSR1` | Rotate log files |
| `SIGUSR2` | Toggle debug logging level |

---

## Install Script (`jphat.net/buzzpi/install`)

The install script supports the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BUZZPI_VERSION` | `latest` | Runtime version to install |
| `BUZZPI_CHANNEL` | `stable` | Release channel |
| `BUZZPI_CONFIG_DIR` | `/etc/buzzpi` | Configuration directory |
| `BUZZPI_DATA_DIR` | `/var/lib/buzzpi` | Data directory |
| `BUZZPI_NO_START` | `""` | Set to `1` to skip starting the service |
| `BUZZPI_DRY_RUN` | `""` | Set to `1` to show what would be installed |

```bash
# Install the latest stable version
curl -sS https://jphat.net/buzzpi/install | bash

# Install a specific version
BUZZPI_VERSION=v0.1.0 curl -sS https://jphat.net/buzzpi/install | bash

# Install without starting the service
BUZZPI_NO_START=1 curl -sS https://jphat.net/buzzpi/install | bash
```

---

## Backend CLI (`buzzpi-backend`)

```
buzzpi-backend [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./config.yaml` | Path to config file |
| `--port` | `8080` | HTTP server port |
| `--validate` | `false` | Validate config and exit |
| `--write-default-config` | `false` | Write default config to stdout |
| `--migrate` | `false` | Run database migrations and exit |
| `--seed` | `false` | Seed database with test data (development) |
| `--log-level` | `"info"` | Log level |

---

## Android App (Build Variants)

Configured via `build.gradle.kts`:

| Build Type | Description | Uses |
|------------|-------------|------|
| `debug` | Development build | Local relay, detailed logging, non-optimized |
| `release` | Production build | Production relay, minified, optimized |

### Build Config Fields

Set in `local.properties` or environment:

| Property | Description |
|----------|-------------|
| `buzzpi.relayUrl` | WebSocket relay URL |
| `buzzpi.apiUrl` | REST API base URL |
| `buzzpi.stunServers` | Comma-separated STUN server URLs |
| `buzzpi.turnServers` | Comma-separated TURN server URLs |

```bash
# Debug build
./gradlew assembleDebug

# Release build
./gradlew assembleRelease

# Install on connected device
./gradlew installDebug
```
