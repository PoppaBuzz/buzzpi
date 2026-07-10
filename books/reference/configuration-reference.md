# Configuration Reference

**Every configuration option across all BuzzPi components.** This is the definitive reference for Runtime, client, and backend configuration files, environment variables, and runtime flags.

---

## Runtime Configuration

Location: `/etc/buzzpi/config.yaml`

### Core

```yaml
# Runtime identity and networking
runtime:
  # Friendly name for this device (auto-generated from hostname if empty)
  friendly_name: "Kitchen Pi"

  # Device ID (auto-generated on first start, persisted)
  device_id: "dev_abc123"

  # Data directory for persistent state
  data_dir: "/var/lib/buzzpi"

  # Log level: debug | info | warn | error
  log_level: "info"

  # Log destination: stdout | journald | file
  log_destination: "journald"

  # Log file path (when log_destination = "file")
  log_file: "/var/log/buzzpi/runtime.log"
```

### Networking

```yaml
network:
  # WebSocket signaling server
  relay_url: "wss://relay.buzzpi.dev/v1"

  # mDNS service advertisement
  mdns:
    enabled: true
    service_type: "_buzzpi._tcp"
    port: 9090   # WebSocket signaling port
    ttl_seconds: 120

  # WebRTC ICE servers
  ice_servers:
    - urls: ["stun:stun.l.google.com:19302"]
    - urls: ["stun:stun1.l.google.com:19302"]
    - urls: ["turn:turn.buzzpi.dev:3478"]
      username: "${TURN_USERNAME}"     # From env var
      credential: "${TURN_CREDENTIAL}" # From env var
```

### Services

```yaml
services:
  # Terminal PTY service
  terminal:
    enabled: true
    default_shell: "/bin/bash"
    max_sessions: 10
    idle_timeout_minutes: 30
    max_output_buffer: 65536  # 64KB per session

  # Screen capture service
  screen:
    enabled: false               # Disabled by default in v0.1.x
    capture_method: "auto"       # auto | drm | fbdev | x11
    max_fps: 30
    max_resolution: "1920x1080"
    encoder: "auto"              # auto | mmal | v4l2 | software
    quality: "high"              # high | medium | low | minimum

  # GPIO service
  gpio:
    enabled: false
    driver: "auto"               # auto | sysfs | libgpiod | pigpio

  # Camera service
  camera:
    enabled: false
    default_device: "/dev/video0"

  # Docker service
  docker:
    enabled: false
    socket_path: "/var/run/docker.sock"

  # Systemd service management
  systemd:
    enabled: false

  # Log service
  logs:
    enabled: true
    journald:
      use_system_journal: true
      default_tail: 100
```

### Pairing

```yaml
pairing:
  # Pairing code length (alphanumeric)
  code_length: 6

  # Pairing code expiry
  code_ttl_seconds: 300  # 5 minutes

  # Identity key algorithm
  key_algorithm: "ed25519"

  # Key storage
  key_dir: "/var/lib/buzzpi/keys"
```

### Connection

```yaml
connection:
  # Heartbeat interval for WebRTC data channels
  heartbeat_interval_seconds: 30

  # Reconnection configuration
  reconnection:
    initial_delay_seconds: 1
    max_delay_seconds: 30
    max_attempts: 10
    multiplier: 2.0

  # Transport selection
  transport:
    p2p_timeout_seconds: 10
    relay_timeout_seconds: 30
    prefer_p2p: true
```

### BuzzAI

```yaml
buzzai:
  enabled: false  # Disabled by default

  local:
    enabled: false
    model_path: "/var/lib/buzzpi/ai/models/model.gguf"
    context_length: 4096
    max_tokens: 512
    threads: 4

  cloud:
    enabled: false
    provider: "openai"
    endpoint: "https://api.openai.com/v1"
    model: "gpt-4o-mini"
    api_key: "${OPENAI_API_KEY}"  # From env var
    max_tokens: 1024
    temperature: 0.3

  privacy_mode: false

  # Commands that require user confirmation
  confirm_commands:
    - "write_file"
    - "system_reboot"
    - "system_shutdown"
```

### Plugin System

```yaml
plugins:
  # Plugin storage root
  storage_dir: "/var/lib/buzzpi/plugins"

  # Plugin sources (for install)
  sources:
    - type: "registry"
      url: "https://plugins.buzzpi.dev"

  # Resource limits per plugin
  limits:
    max_cpu_percent: 50
    max_memory_mb: 128
    max_disk_mb: 500
    max_network_kbps: 1000
    max_file_descriptors: 64

  # Plugin lifecycle
  idle_timeout_minutes: 5
  startup_timeout_seconds: 30
  max_restarts: 3
  health_interval_seconds: 15
```

### Updates

```yaml
updates:
  # Update check interval
  check_interval_hours: 24

  # Update channel: stable | beta | nightly
  channel: "stable"

  # Update URL for version manifest
  manifest_url: "https://updates.buzzpi.dev/runtime.json"

  # Automatic update installation (requires restart)
  auto_install: false

  # Download directory for update artifacts
  download_dir: "/var/cache/buzzpi/updates"
```

### Web Server (Embedded)

```yaml
server:
  # Local WebSocket server port (for mDNS + fallback)
  port: 9090

  # Bind address
  bind: "0.0.0.0"

  # TLS (optional, for local connections)
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
```

### Example: Minimal Config

```yaml
# /etc/buzzpi/config.yaml — minimal working config
runtime:
  friendly_name: "Kitchen Pi"
  log_level: "info"
network:
  relay_url: "wss://relay.buzzpi.dev/v1"
services:
  terminal:
    enabled: true
  logs:
    enabled: true
pairing:
  code_length: 6
```

### Example: Full Config

The complete default configuration is embedded in the Runtime binary and is available by running:

```bash
buzzpi-runtime config --defaults
```

---

## Environment Variables

### Runtime

| Variable | Default | Description |
|----------|---------|-------------|
| `BUZZPI_CONFIG` | `/etc/buzzpi/config.yaml` | Path to config file |
| `BUZZPI_DATA_DIR` | `/var/lib/buzzpi` | Data directory override |
| `BUZZPI_LOG_LEVEL` | — | Override log level (takes precedence over config) |
| `BUZZPI_RELAY_URL` | — | Override relay URL |
| `BUZZPI_FRIENDLY_NAME` | — | Override device name |
| `BUZZPI_DEVICE_ID` | — | Override device ID (testing only) |
| `BUZZPI_NO_MDNS` | — | Set to `1` to disable mDNS |
| `BUZZPI_NO_CONNECT` | — | Set to `1` to start without connecting to relay |
| `TURN_USERNAME` | — | TURN server username |
| `TURN_CREDENTIAL` | — | TURN server credential |
| `OPENAI_API_KEY` | — | OpenAI API key for BuzzAI |

### Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | — | JWT signing secret (required) |
| `DATABASE_URL` | `postgres://localhost/buzzpi` | PostgreSQL connection string |
| `TURN_SECRET` | — | HMAC secret for TURN credentials |
| `TURN_REALM` | `buzzpi.dev` | TURN realm |
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `CORS_ORIGINS` | `*` | Allowed CORS origins |

### Android App

| Property | Method | Default | Description |
|----------|--------|---------|-------------|
| `buzzpi.relay_url` | BuildConfig | `wss://relay.buzzpi.dev/v1` | Signaling server URL |
| `buzzpi.api_url` | BuildConfig | `https://api.buzzpi.dev/v1` | REST API base URL |
| `buzzpi.stun_servers` | BuildConfig | `["stun:stun.l.google.com:19302"]` | STUN servers |
| `buzzpi.turn_servers` | BuildConfig | `[]` | TURN servers |
| `buzzpi.mdns_enabled` | BuildConfig | `true` | Enable mDNS discovery |
| `buzzpi.cloud_discovery` | BuildConfig | `true` | Enable cloud device listing |
| `buzzpi.connection_timeout` | BuildConfig | `15000` | Connection timeout (ms) |
| `buzzpi.reconnect_enabled` | BuildConfig | `true` | Enable auto-reconnection |

---

## Backend Configuration

Location: `backend/config.yaml` (or environment variables)

```yaml
# Backend server configuration
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 10s
  write_timeout: 10s

# PostgreSQL database
database:
  url: "postgres://buzzpi@localhost/buzzpi?sslmode=disable"
  max_conns: 25
  min_conns: 5
  max_conn_lifetime: 30m

# JWT authentication
auth:
  jwt_secret: "${JWT_SECRET}"
  access_token_ttl: 15m
  refresh_token_ttl: 720h  # 30 days

# WebSocket relay
relay:
  max_message_size: 65536  # 64KB
  heartbeat_interval: 30s
  session_timeout: 90s

# TURN server
turn:
  enabled: true
  realm: buzzpi.dev
  secret: "${TURN_SECRET}"
  credential_ttl: 24h
  coturn_binary: /usr/bin/coturn
  coturn_config: /etc/buzzpi/coturn.conf

# Rate limiting
rate_limit:
  signup: "3/hour"
  login: "5/minute"
  device_register: "10/hour"
  device_list: "60/minute"
  relay_connect: "10/minute"
```

---

## docker-compose.yml Reference

```yaml
version: "3.8"

services:
  # Registry + Relay backend service
  backend:
    image: buzzpi/backend:latest
    ports:
      - "8080:8080"    # HTTP + WebSocket
    environment:
      - JWT_SECRET=${JWT_SECRET:?required}
      - DATABASE_URL=postgres://buzzpi@db:5432/buzzpi?sslmode=disable
      - TURN_SECRET=${TURN_SECRET:?required}
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    volumes:
      - turn_data:/var/lib/coturn

  # PostgreSQL
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: buzzpi
      POSTGRES_DB: buzzpi
      POSTGRES_PASSWORD: ${DB_PASSWORD:?required}
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U buzzpi"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # coturn TURN server
  coturn:
    image: coturn/coturn:latest
    network_mode: host  # Needs real UDP ports
    volumes:
      - ./coturn.conf:/etc/coturn/turnserver.conf
    restart: unless-stopped

volumes:
  pg_data:
  turn_data:
```

---

## Configuration Validation

### Runtime

```bash
# Validate config file
buzzpi-runtime config --validate

# Dump effective configuration (merged defaults + file + env)
buzzpi-runtime config --dump

# Show default configuration
buzzpi-runtime config --defaults
```

### Backend

```bash
# Validate config on startup (logged, non-fatal if missing optional fields)
./buzzpi-backend --validate

# Generate default config file
./buzzpi-backend --write-default-config
```

---

## Config File Locations

| Component | Priority Order |
|-----------|---------------|
| **Runtime** | 1. `$BUZZPI_CONFIG` 2. `/etc/buzzpi/config.yaml` 3. `./config.yaml` 4. Embedded defaults |
| **Android** | 1. BuildConfig constants 2. Remote config (Firebase, future) |
| **Backend** | 1. Environment variables 2. `./config.yaml` 3. `$HOME/.buzzpi/config.yaml` 4. `/etc/buzzpi/backend.yaml` |

---

## Schema Versioning

Configuration files include a schema version field for forward compatibility:

```yaml
# v0.1.0 config schema
schema_version: "0.1.0"
```

The Runtime validates the schema version on load. If the config file uses a newer schema version than the Runtime supports, it logs a warning and falls back to defaults for unknown fields.
