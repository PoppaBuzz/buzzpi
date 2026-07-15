# Environment Reference

**Complete environment variable registry for the BuzzPi Platform.**

This document catalogs every environment variable recognized by BuzzPi components. All variables follow the `BUZZPI_` prefix convention.

---

## 1. Naming Convention

```
BUZZPI_[COMPONENT]_[KEY]
```

| Component Prefix | Applies To |
|-----------------|------------|
| `BUZZPI_` | Global / cross-component |
| `BUZZPI_CLI_` | CLI client |
| `BUZZPI_AGENT_` | Go Agent runtime |
| `BUZZPI_RELAY_` | Cloud Relay |
| `BUZZPI_PLUGIN_` | Plugin subsystem |
| `BUZZPI_ANDROID_` | Android client |

---

## 2. Global Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_HOME` | path | `~/.buzzpi` | BuzzPi data directory |
| `BUZZPI_CONFIG` | path | `$BUZZPI_HOME/config.yaml` | Config file path |
| `BUZZPI_LOG_LEVEL` | enum | `info` | Log level: `debug`, `info`, `warn`, `error`, `fatal` |
| `BUZZPI_LOG_FILE` | path | `$BUZZPI_HOME/logs/buzzpi.log` | Log file path |
| `BUZZPI_LOG_FORMAT` | enum | `text` | Log format: `text`, `json` |
| `BUZZPI_MODE` | enum | `release` | Runtime mode: `debug`, `release`, `test` |
| `BUZZPI_DATA_DIR` | path | `$BUZZPI_HOME/data` | Persistent data directory |
| `BUZZPI_CACHE_DIR` | path | `$BUZZPI_HOME/cache` | Cache directory |
| `BUZZPI_TEMP_DIR` | path | `$BUZZPI_HOME/tmp` | Temporary files directory |

---

## 3. CLI Client Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_CLI_DEFAULT_DEVICE` | string | — | Default target device ID |
| `BUZZPI_CLI_OUTPUT_FORMAT` | enum | `text` | Output format: `text`, `json`, `yaml` |
| `BUZZPI_CLI_TIMEOUT` | duration | `30s` | Default request timeout |
| `BUZZPI_CLI_NO_COLOR` | bool | `false` | Disable ANSI color output |
| `BUZZPI_CLI_DISABLE_PROGRESS` | bool | `false` | Disable progress bars |
| `BUZZPI_CLI_SESSION_DIR` | path | `$BUZZPI_HOME/sessions` | Session token storage |

---

## 4. Agent Runtime Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_AGENT_ID` | string | auto | Override agent device ID |
| `BUZZPI_AGENT_NAME` | string | `hostname` | Override agent friendly name |
| `BUZZPI_AGENT_PORT` | port | `10104` | BPP WebSocket listening port |
| `BUZZPI_AGENT_PLATFORM` | string | auto | Platform string, e.g. `raspberry-pi/5` |
| `BUZZPI_AGENT_MAX_SESSIONS` | int | `16` | Maximum concurrent sessions |
| `BUZZPI_AGENT_SESSION_TTL` | duration | `24h` | Session token time-to-live |
| `BUZZPI_AGENT_HEARTBEAT_INTERVAL` | duration | `30s` | WebSocket heartbeat interval |
| `BUZZPI_AGENT_HEARTBEAT_TIMEOUT` | duration | `90s` | Heartbeat timeout |
| `BUZZPI_AGENT_WS_READ_LIMIT` | bytes | `65536` | WebSocket read limit |
| `BUZZPI_AGENT_WS_WRITE_LIMIT` | bytes | `65536` | WebSocket write limit |
| `BUZZPI_AGENT_EXEC_TIMEOUT` | duration | `300s` | Command execution timeout |
| `BUZZPI_AGENT_FILE_CHUNK_SIZE` | bytes | `65536` | File transfer chunk size |
| `BUZZPI_AGENT_FILE_MAX_SIZE` | bytes | `1073741824` | Max file upload size (1GB) |
| `BUZZPI_AGENT_SCREEN_FPS` | int | `15` | Screen capture frame rate |
| `BUZZPI_AGENT_SCREEN_QUALITY` | int `[1-100]` | `60` | Screen capture JPEG quality |
| `BUZZPI_AGENT_SCREEN_MAX_DIM` | int | `1920` | Max screen capture dimension |
| `BUZZPI_AGENT_PLUGIN_DIR` | path | `/var/lib/buzzpi/plugins` | Plugin installation directory |
| `BUZZPI_AGENT_PLUGIN_MAX_MEMORY` | bytes | `268435456` | Per-plugin memory limit (256MB) |
| `BUZZPI_AGENT_PLUGIN_MAX_STARTUP` | duration | `15s` | Plugin max startup time |
| `BUZZPI_AGENT_RECONNECT_BASE` | duration | `1s` | Reconnect exponential backoff base |
| `BUZZPI_AGENT_RECONNECT_MAX` | duration | `60s` | Reconnect max backoff |
| `BUZZPI_AGENT_RECONNECT_JITTER` | float `[0-1]` | `0.1` | Reconnect jitter factor |
| `BUZZPI_AGENT_STATE_DIR` | path | `$BUZZPI_DATA_DIR/state` | State store directory |
| `BUZZPI_AGENT_CERT_DIR` | path | `$BUZZPI_DATA_DIR/certs` | TLS certificate directory |

### 4.1 mDNS Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_AGENT_MDNS_ENABLED` | bool | `true` | Enable mDNS discovery |
| `BUZZPI_AGENT_MDNS_SERVICE` | string | `_buzzpi._tcp` | mDNS service type |
| `BUZZPI_AGENT_MDNS_TTL` | duration | `120s` | mDNS advertisement TTL |
| `BUZZPI_AGENT_MDNS_INTERFACE` | string | `0.0.0.0` | mDNS listening interface |

---

## 5. Cloud Relay Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_RELAY_URL` | url | `wss://relay.buzzpi.cloud` | Relay server WebSocket URL |
| `BUZZPI_RELAY_API_URL` | url | `https://relay.buzzpi.cloud` | Relay REST API base URL |
| `BUZZPI_RELAY_TOKEN` | string | — | Relay authentication token |
| `BUZZPI_RELAY_RECONNECT_BASE` | duration | `1s` | Reconnect backoff base |
| `BUZZPI_RELAY_RECONNECT_MAX` | duration | `120s` | Reconnect max backoff |
| `BUZZPI_RELAY_PING_INTERVAL` | duration | `15s` | WebSocket ping interval |
| `BUZZPI_RELAY_CONNECTION_TIMEOUT` | duration | `10s` | Connection timeout |
| `BUZZPI_RELAY_MAX_RETRANSMIT` | int | `5` | Max message retransmissions |
| `BUZZPI_RELAY_BUFFER_SIZE` | bytes | `262144` | Relay buffer per connection (256KB) |

---

## 6. Plugin Subsystem Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_PLUGIN_REGISTRY_URL` | url | `https://jphat.net/buzzpi/plugins` | Plugin registry URL |
| `BUZZPI_PLUGIN_ALLOW_UNSIGNED` | bool | `false` | Allow unsigned plugins |
| `BUZZPI_PLUGIN_VERIFY_SIGNATURE` | bool | `true` | Verify plugin signatures |
| `BUZZPI_PLUGIN_SANDBOX_ENABLED` | bool | `true` | Enable process sandboxing |
| `BUZZPI_PLUGIN_SANDBOX_TYPE` | enum | `seccomp` | Sandbox type: `seccomp`, `apparmor`, `none` |
| `BUZZPI_PLUGIN_NETWORK_ACCESS` | bool | `false` | Allow plugin network access |
| `BUZZPI_PLUGIN_FS_READ_ONLY` | list | `[]` | Plugin read-only filesystem paths |
| `BUZZPI_PLUGIN_FS_WRITABLE` | list | `[plugin_dir]` | Plugin writable filesystem paths |
| `BUZZPI_PLUGIN_MANIFEST_MAX_SIZE` | bytes | `65536` | Max manifest file size |
| `BUZZPI_PLUGIN_BINARY_MAX_SIZE` | bytes | `268435456` | Max plugin binary size (256MB) |

---

## 7. Android Client Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_ANDROID_RELAY_URL` | url | `wss://relay.buzzpi.cloud` | Default relay server |
| `BUZZPI_ANDROID_DISCOVER_TIMEOUT` | duration | `5s` | mDNS discovery timeout |
| `BUZZPI_ANDROID_TERMINAL_FONT_SIZE` | int | `12` | Terminal emulator font size |
| `BUZZPI_ANDROID_TERMINAL_FONT_FAMILY` | string | `monospace` | Terminal emulator font |
| `BUZZPI_ANDROID_TERMINAL_ROWS` | int | `40` | Terminal default rows |
| `BUZZPI_ANDROID_TERMINAL_COLS` | int | `80` | Terminal default columns |
| `BUZZPI_ANDROID_SCREEN_QUALITY` | int `[1-100]` | `70` | Screen viewer quality |
| `BUZZPI_ANDROID_KEEP_SCREEN_ON` | bool | `true` | Keep screen on during session |
| `BUZZPI_ANDROID_NOTIFICATIONS_ENABLED` | bool | `true` | Enable device notifications |
| `BUZZPI_ANDROID_AUTO_CONNECT` | bool | `false` | Auto-connect to paired devices |
| `BUZZPI_ANDROID_CRASH_REPORTING` | bool | `true` | Enable crash reporting |

---

## 8. Security-Sensitive Variables

These variables require special handling:
- Must never be logged
- Must never appear in stack traces
- Should be sourced from secure storage when possible

| Variable | Sensitivity | Recommended Source |
|----------|-------------|-------------------|
| `BUZZPI_RELAY_TOKEN` | High | Secret manager or keyring |
| `BUZZPI_AGENT_TLS_KEY` | High | File with restricted permissions (0600) |
| `BUZZPI_AGENT_TLS_CERT` | Medium | File with restricted permissions (0644) |
| `BUZZPI_AGENT_ID` | Medium | Machine identity, read-only |

---

## 9. Variable Precedence

Variables are resolved in this order (highest priority first):

1. **CLI flag** — `--log-level debug`
2. **Environment variable** — `BUZZPI_LOG_LEVEL=debug`
3. **Config file** — `~/.buzzpi/config.yaml` → `logging.level: debug`
4. **Default value** — as defined in this document

---

## 10. Validation Rules

- **Duration** values: must be parseable by Go's `time.ParseDuration` (e.g. `30s`, `5m`, `1h`)
- **Port** values: must be in range 1024–65535
- **Path** values: must be absolute paths
- **Enum** values: case-insensitive, validated against allowed values
- **Byte** values: must be valid integers within documented range

---

## 11. Testing Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `BUZZPI_TEST_AGENT_BINARY` | path | `./buzzpi-agent` | Path to agent binary for integration tests |
| `BUZZPI_TEST_RELAY_URL` | url | `ws://localhost:10106` | Test relay server URL |
| `BUZZPI_TEST_SKIP_NETWORK` | bool | `false` | Skip network-dependent tests |
| `BUZZPI_TEST_SKIP_RELAY` | bool | `false` | Skip relay-dependent tests |
| `BUZZPI_TEST_TEMP_DIR` | path | system temp | Temporary test directory |

---

## References

- Reference: Platform Reference (filesystem layout)
- Reference: Configuration Reference (config file format)
- Reference: BPP Client API (programmatic usage)
- RFC-0002: Runtime Architecture
- RFC-0006: Cloud Relay
