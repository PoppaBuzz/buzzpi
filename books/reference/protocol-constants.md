# Protocol Constants Reference

All well-known ports, URLs, identifiers, and protocol constants.

## Network Ports

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Relay WebSocket | 443 | WSS | Primary signaling and control (standard HTTPS) |
| TURN UDP | 3478 | UDP | WebRTC media relay |
| TURN TCP | 3478 | TCP | WebRTC media relay (TCP fallback) |
| TURN TLS | 5349 | TCP | WebRTC media relay (TLS) |
| mDNS Service | 5353 | UDP | Local device discovery |

## Relay URLs

| Environment | URL | Purpose |
|-------------|-----|---------|
| Production | `wss://jphat.net/buzzpi/relay/ws` | Primary relay server |
| Staging | `wss://jphat.net/buzzpi/relay-staging/ws` | Pre-production testing |
| Development | `ws://localhost:8080/ws` | Local development |

## TURN URLs

| Environment | URL |
|-------------|-----|
| Production | `turn:jphat.net/buzzpi/turn:3478` |
| Production (TLS) | `turn:jphat.net/buzzpi/turn:5349?transport=tcp` |
| Staging | `turn:jphat.net/buzzpi/turn-staging:3478` |

## STUN Servers

| URL | Provider |
|-----|----------|
| `stun:stun.l.google.com:19302` | Google |
| `stun:stun1.l.google.com:19302` | Google |

## Registry URLs

| Service | URL |
|---------|-----|
| Extension Registry | `https://jphat.net/buzzpi/registry/v1` |
| Update Server | `https://jphat.net/buzzpi/releases/runtime/latest` |

## Message Envelope Constants

| Field | Values |
|-------|--------|
| BPP Version | `1` |
| Max message size (WebSocket) | 64 KB |
| Max batch payload | 512 KB |
| Max data channel payload | 64 KB |
| Max terminal write payload | 64 KB |

## Timing Constants

| Constant | Value | Context |
|----------|-------|---------|
| WebSocket ping interval | 30s | Connection health |
| WebSocket ping timeout | 10s | Connection dead detection |
| Application heartbeat (device) | 60s | Presence |
| Application heartbeat (client) | 5min | Presence |
| Grace period (offline) | 2min | Before offline notification |
| Pairing code lifetime | 5min | Code validity |
| Pairing code length | 6 chars | Alphanumeric (base30) |
| Session token lifetime | 24h | Client session |
| API token lifetime | 90d | CLI token |
| ICE consent freshness | 30s | WebRTC health |
| Reconnect initial delay | 1s | Exponential backoff start |
| Reconnect max delay | 30s | Exponential backoff cap |
| Reconnect max attempts | 10 | Before session closure |
| Terminal idle timeout | 30min | Auto-close inactive sessions |
| Max terminal sessions (per device) | 10 | Concurrent limit |
| Screen keyframe interval | 2s | H.264 GOP |
| Screen quality adaptation | 5s | Monitoring interval |

## Rate Limits

| Context | Limit | Burst |
|---------|-------|-------|
| Client → Relay signaling | 100 msg/s | 200 |
| Device → Relay signaling | 10 msg/s | 50 |
| App → Runtime (data channel) | 1000 msg/s | 5000 |
| Relay → App notifications | 10 msg/s/device | 20 |
| Terminal writes | 100/s/session | — |
| GPIO reads | 1000/s | — |
| GPIO writes | 1000/s | — |
| Camera snapshots | 10/min | — |
| File transfers (concurrent) | 3/client | — |
| Pairing attempts | 3/code | — |

## Size Limits

| Item | Limit |
|------|-------|
| Device name | 64 chars |
| Device friendly name | 64 chars |
| Pairing code | 6 chars (base30) |
| Terminal dimensions (cols) | 20-500 |
| Terminal dimensions (rows) | 5-200 |
| File listing entries | 1000/directory |
| File chunk | 64 KB |
| File upload max | 100 MB |
| Extension manifest | 64 KB |
| Plugin binary max | 50 MB |
| Config file max | 1 MB |
| Audit log (on-device) | 1 MB (rotated) |

## Well-Known Identifiers

### Device Capability IDs

| ID | Service | Description |
|----|---------|-------------|
| `terminal` | Services | PTY terminal access |
| `screen` | Services | Graphical screen streaming |
| `files` | Services | File system access |
| `stats` | Services | System metrics |
| `gpio` | Services | GPIO pin control |
| `docker` | Services | Docker management |
| `services` | Services | systemd service management |
| `logs` | Services | System and application logs |
| `camera` | Services | Camera access |
| `rpc` | Services | Custom method calls |

### Extension Permission IDs

| ID | Description | Risk |
|----|-------------|------|
| `terminal:read` | Read terminal output | Low |
| `terminal:write` | Send input to terminal | Medium |
| `files:read` | Read files on device | Medium |
| `files:write` | Modify or delete files | High |
| `docker:read` | List containers and images | Low |
| `docker:write` | Manage container lifecycle | Medium |
| `docker:admin` | Pull images, system prune | High |
| `gpio:read` | Read pin states | Low |
| `gpio:write` | Write pin states, PWM | High |
| `camera:read` | Camera preview, snapshots | Medium |
| `camera:write` | Start/stop recording | High |
| `network:read` | Read network configuration | Low |
| `network:admin` | Modify network config | High |
| `system:read` | Read system info and logs | Low |
| `system:admin` | Restart/shutdown, manage services | Critical |

### Notification Category IDs

| ID | Default Priority | Description |
|----|-----------------|-------------|
| `device.online` | Info | Device came online |
| `device.offline` | Info | Device went offline (after grace period) |
| `device.temperature` | Warning | Temperature threshold breached |
| `device.storage` | Warning | Storage threshold breached |
| `device.update_available` | Info | Runtime update available |
| `device.update_completed` | Info | Runtime update installed |
| `action.completed` | Success | Action completed successfully |
| `action.failed` | Warning | Action failed |
| `extension.crashed` | Warning | Extension stopped unexpectedly |
| `pairing.requested` | Action | New device wants to pair |
| `pairing.expired` | Info | Pairing code expired |
| `automation.triggered` | Info | Automation triggered |
| `automation.completed` | Success | Automation completed |
| `automation.failed` | Warning | Automation failed |

### WebSocket Close Codes

| Code | Reason | Sender |
|------|--------|--------|
| 1000 | Normal closure | Either |
| 4001 | Authentication failed | Server |
| 4002 | Token expired | Server |
| 4003 | Rate limited | Server |
| 4004 | Server shutting down | Server |
| 4005 | Client going offline | Client/Device |
| 4006 | Protocol version mismatch | Server |

### Error Code Ranges

| Range | Layer | 
|-------|-------|
| 1xxx | Transport |
| 2xxx | Identity |
| 3xxx | Protocol |
| 4xxx | Service |
| 5xxx | Resource |
| 6xxx | Extension |
