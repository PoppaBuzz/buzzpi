# File Format Reference

**Complete file format specifications for the BuzzPi Platform.**

This document describes every file format used across the BuzzPi ecosystem — naming conventions, storage paths, formats, and schemas.

---

## 1. Configuration Files

### 1.1 `config.yaml` — Main Configuration

- **Location:** `~/.buzzpi/config.yaml` or `$BUZZPI_CONFIG`
- **Format:** YAML
- **Encoding:** UTF-8
- **Schema:** [Configuration Reference](configuration-reference.md)

```yaml
# ~/.buzzpi/config.yaml
default_device: dev_a1b2c3d4
output_format: text
timeout: 30s

connection:
  preferred_transport: auto

logging:
  level: info
```

### 1.2 `agent.yaml` — Agent Configuration

- **Location:** `/etc/buzzpi/agent.yaml` (system) or `./agent.yaml` (cwd)
- **Format:** YAML
- **Encoding:** UTF-8
- **Permissions:** 0644 (system), 0600 (with secrets)

---

## 2. State Files

### 2.1 `state.db` — Persistent State (BoltDB)

- **Location:** `$BUZZPI_DATA_DIR/state/state.db`
- **Format:** BoltDB (Go embedded key-value store)
- **Buckets:**

| Bucket | Key Type | Value Type | Description |
|--------|----------|------------|-------------|
| `devices` | device_id (string) | DeviceEntry | Paired device metadata |
| `sessions` | session_id (string) | SessionEntry | Active session tokens |
| `plugins` | plugin_id (string) | PluginEntry | Installed plugin records |
| `settings` | setting_key (string) | JSON value | Persisted settings |
| `identity` | `keypair` | KeyPair | Agent key material |

**Not intended for direct manipulation.** Use the `buzzpi config` CLI commands or the Agent's state API.

### 2.2 `sessions.json` — Session Cache

- **Location:** `~/.buzzpi/sessions.json`
- **Format:** JSON
- **Permissions:** 0600
- **Schema:**

```json
{
  "version": 1,
  "sessions": [
    {
      "device_id": "dev_a1b2c3d4",
      "token": "sess_eyJhbGciOiJIUzI1NiIs...",
      "paired_at": "2026-07-07T12:00:00Z",
      "expires_at": "2026-07-08T12:00:00Z"
    }
  ]
}
```

---

## 3. Plugin Files

### 3.1 `plugin.yaml` / `plugin.json` — Plugin Manifest

- **Location:** Root of plugin bundle
- **Format:** YAML or JSON
- **Schema:** [Plugin Manifest](plugin-manifest.md)

```yaml
# plugin.yaml
id: com.example.docker-manager
name: Docker Manager
version: 1.2.0
min_agent_version: 0.1.0
entrypoint: docker-manager
```

### 3.2 Plugin Bundle (`*.bpp`)

- **Extension:** `.bpp`
- **Format:** gzipped tar archive
- **Contents:**

```
docker-manager.bpp
├── plugin.yaml              # Manifest (required)
├── docker-manager           # Binary (required)
├── icon.png                 # Plugin icon (optional)
├── README.md                # Documentation (optional)
├── assets/
│   ├── config.schema.json   # Config schema (optional)
│   └── default.conf         # Default config (optional)
└── signatures/
    └── plugin.yaml.sig      # Manifest signature (required if signing enabled)
```

**Layout constraints:**
- All files must be relative to archive root (no `../` paths)
- Maximum archive size: 256MB
- Maximum file count: 500
- Maximum filename length: 255 bytes

### 3.3 Plugin Installation Layout

```
/var/lib/buzzpi/plugins/
├── com.example.docker-manager/
│   ├── plugin.yaml
│   ├── docker-manager        # Binary (executable)
│   ├── data/                 # Plugin writable data
│   │   ├── config.yaml
│   │   └── docker.sock
│   └── logs/                 # Plugin stdout/stderr logs
│       ├── stdout.log
│       └── stderr.log
└── com.example.sensors/
    └── ...
```

---

## 4. Log Files

### 4.1 Agent Log (`buzzpi.log`)

- **Location:** `$BUZZPI_HOME/logs/buzzpi.log`
- **Format:** Structured text (configurable: text or JSON)
- **Rotation:** 10MB per file, 5 rotated files
- **Text format:**

```
2026-07-07T12:00:00.123Z INFO  agent started version=0.1.0 port=10104
2026-07-07T12:00:05.456Z DEBUG mdns registered service=_buzzpi._tcp
2026-07-07T12:00:10.789Z INFO  ws client connected device_id=dev_a1b2c3d4
```

- **JSON format:**

```json
{"time":"2026-07-07T12:00:00.123Z","level":"info","msg":"agent started","version":"0.1.0","port":10104}
{"time":"2026-07-07T12:00:10.789Z","level":"info","msg":"ws client connected","device_id":"dev_a1b2c3d4"}
```

### 4.2 Plugin Logs

- **Location:** `$BUZZPI_AGENT_PLUGIN_DIR/<plugin-id>/logs/`
- **Format:** Raw stdout/stderr
- **Rotation:** 5MB per file, 3 rotated files

---

## 5. Identity and Certificate Files

### 5.1 Agent Identity (`identity.pem`)

- **Location:** `$BUZZPI_DATA_DIR/certs/identity.pem`
- **Format:** PEM-encoded Ed25519 private key
- **Permissions:** 0600
- **Generated at first startup if missing**

### 5.2 TLS Certificate (`cert.pem`, `key.pem`)

- **Location:** `$BUZZPI_DATA_DIR/certs/{cert,key}.pem`
- **Format:** PEM-encoded X.509
- **Permissions:** cert 0644, key 0600
- **Self-signed for LAN, CA-signed for relay**

### 5.3 Trust Store (`ca-certs.pem`)

- **Location:** `$BUZZPI_DATA_DIR/certs/ca-certs.pem`
- **Format:** PEM-encoded concatenated CA certificates
- **Permissions:** 0644

---

## 6. Data Exchange Formats

### 6.1 BPP Packet

- **Transport:** WebSocket binary frames
- **Format:** BPP binary packet (see [Packet Types](packet-types.md))
- **Encoding:** BPP wire format (fixed header + variable payload)

### 6.2 BPP JSON Messages

- **Transport:** WebSocket text frames (during capability negotiation)
- **Format:** JSON
- **Schema:** [JSON Schemas](json-schemas.md)

### 6.3 File Transfer Chunks

- **Transport:** BPP serial (unreliable data) frames
- **Format:** Binary chunk with metadata header
- **Header (16 bytes):**

```
Offset  Size  Field
0       4     sequence_number (uint32, big-endian)
4       4     offset (uint32, big-endian)
8       4     total_size (uint32, big-endian)
12      1     flags (bit 0: last_chunk)
13      3     reserved (zero)
```

---

## 7. Export and Snapshot Formats

### 7.1 Device Snapshot (`*.buzzpi-snapshot`)

- **Format:** gzipped tar archive
- **Extension:** `.buzzpi-snapshot`
- **Contents:**

```
living-room-pi_2026-07-07.buzzpi-snapshot
├── snapshot.yaml            # Metadata and manifest
├── state.db                 # BoltDB state export
├── config.yaml              # Agent configuration
├── certs/
│   ├── cert.pem
│   └── identity.pem
└── plugins/
    └── com.example.docker-manager/
        └── plugin.yaml
```

- **`snapshot.yaml` schema:**

```yaml
# snapshot.yaml
version: 1
created_at: 2026-07-07T12:00:00Z
device_id: dev_a1b2c3d4
agent_version: 0.1.0
platform: raspberry-pi/5
includes:
  state: true
  config: true
  certs: true
  plugins: false  # Plugin binaries excluded by default
```

---

## 8. Diagnostic Files

### 8.1 Diagnostic Bundle (`*.buzzpi-diag`)

- **Format:** gzipped tar archive
- **Extension:** `.buzzpi-diag`
- **Command:** `buzzpi doctor --export diagnostics.buzzpi-diag`

```
diagnostics_2026-07-07.buzzpi-diag
├── diag.yaml                 # Metadata
├── system.yaml               # System information
├── config.yaml               # Sanitized configuration
├── state.yaml                # State summary (no tokens)
├── logs/
│   ├── buzzpi.log
│   └── agent.log
├── network/
│   ├── interfaces.yaml
│   ├── connections.txt       # ss/netstat output
│   └── mdns.txt              # mDNS browse results
└── plugins/
    ├── installed.yaml
    └── com.example.docker-manager.log
```

---

## 9. Versioning Conventions

| File Type | Version Field | Format | Example |
|-----------|--------------|--------|---------|
| Plugin bundle | `version` in manifest | semver | `1.2.0` |
| Session cache | `version` at root | integer | `1` |
| Snapshot | `version` in snapshot.yaml | integer | `1` |
| Diagnostic bundle | Implicit via metadata | date | `2026-07-07` |

---

## References

- Reference: Plugin Manifest (plugin.yaml schema)
- Reference: Configuration Reference (config.yaml schema)
- Reference: JSON Schemas (BPP message schemas)
- Reference: Platform Reference (filesystem layout)
- RFC-0004: Plugin Architecture
