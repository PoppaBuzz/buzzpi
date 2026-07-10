# Event Catalog

**Every event the platform emits, with payload schemas.** Events are the pulse of BuzzPi — they flow through the Event Bus internally and are forwarded to clients via BPP events and push notifications.

---

## Organization

Events are organized by source system. Each event entry includes:

- **Topic**: The Event Bus topic pattern
- **BPP Method**: How the event is forwarded to clients (if applicable)
- **Direction**: `Runtime → Client`, `Internal`, or `Plugin → Runtime`
- **Payload**: The event data schema
- **Frequency**: How often the event fires

---

## Connection Events

### connection.established

Fired when a WebRTC connection is successfully established.

```yaml
topic: connection.established
bpp_method: connection.event
direction: Runtime → Client
frequency: Once per connection

payload:
  type: "connection.established"
  data:
    device_id: string       # Device identifier
    transport: string       # "direct" | "relay"
    rtt_ms: int            # Round-trip time in milliseconds
    ice_role: string       # "controlling" | "controlled"
    local_candidate_type: string  # "host" | "srflx" | "relay"
    remote_candidate_type: string # "host" | "srflx" | "relay"
```

### connection.lost

Fired when the WebRTC connection is lost unexpectedly.

```yaml
topic: connection.lost
bpp_method: connection.event
direction: Runtime → Client
frequency: Once per disconnection event

payload:
  type: "connection.lost"
  data:
    device_id: string
    reason: string         # "timeout" | "ice_disconnected" | "peer_closed" | "transport_error"
    last_rtt_ms: int
    duration_seconds: int  # How long the connection was active
    bytes_sent: int
    bytes_received: int
```

### connection.reconnecting

Fired when reconnection attempt begins.

```yaml
topic: connection.reconnecting
bpp_method: connection.event
direction: Runtime → Client
frequency: Per reconnection attempt

payload:
  type: "connection.reconnecting"
  data:
    device_id: string
    attempt: int
    max_attempts: int
    backoff_seconds: int
    transport_attempt: string  # "direct" | "relay"
```

### connection.transport.switched

Fired when the transport type changes.

```yaml
topic: connection.transport.switched
bpp_method: connection.event
direction: Runtime → Client
frequency: Infrequent

payload:
  type: "connection.transport.switched"
  data:
    device_id: string
    from: string           # "direct" | "relay"
    to: string             # "direct" | "relay"
    reason: string         # "nat_traversal_failed" | "p2p_established"
```

### connection.quality

Fired periodically with transport quality metrics.

```yaml
topic: connection.quality
bpp_method: connection.event
direction: Internal (drives quality adaptation)
frequency: Every 5 seconds during active connection

payload:
  device_id: string
  rtt_ms: int
  packet_loss_percent: float
  jitter_ms: int
  available_send_bandwidth: int     # bps
  available_receive_bandwidth: int  # bps
  bytes_sent: int
  bytes_received: int
  timestamp: string
```

---

## Device State Events

### device.state.changed

Fired when the device transitions between lifecycle states.

```yaml
topic: device.state.changed
bpp_method: device.event
direction: Internal (also forwarded to all connected clients)
frequency: On state transition

payload:
  type: "device.state.changed"
  data:
    from: DeviceState      # "new" | "pairing" | "paired" | "online" | "offline" | "unpaired"
    to: DeviceState
    reason: string         # Optional explanation
```

### device.registered

Fired when the device registers with the relay server.

```yaml
topic: device.registered
bpp_method: device.event
direction: Internal
frequency: Once per device registration

payload:
  device_id: string
  relay_url: string
  registered_at: string    # ISO 8601
```

### device.unpaired

Fired when a pairing is removed.

```yaml
topic: device.unpaired
bpp_method: device.event
direction: Runtime → Client (to the client that initiated unpairing)
frequency: On unpair

payload:
  type: "device.unpaired"
  data:
    device_id: string
    reason: string         # "user_initiated" | "key_compromised" | "admin_revoked"
    initiated_by: string   # "client" | "runtime" | "backend"
```

---

## System Events

### stats.update

Fired periodically with system statistics.

```yaml
topic: stats.update
bpp_method: none (consumed by internal subscribers)
direction: Internal
frequency: Every 30 seconds

payload:
  cpu_percent: float
  memory_mb: int
  memory_total_mb: int
  temperature_celsius: float
  storage:
    - mount: string
      total_mb: int
      used_mb: int
      available_mb: int
      percent: float
  network:
    - interface: string
      rx_bytes: int
      tx_bytes: int
  uptime_seconds: int
  load_average:
    1m: float
    5m: float
    15m: float
```

### system.low_disk

Fired when disk usage exceeds threshold.

```yaml
topic: system.low_disk
bpp_method: device.event
direction: Runtime → Client (push notification when client not connected)
frequency: Every 30 minutes while condition persists

payload:
  type: "system.low_disk"
  data:
    mount: string
    available_mb: int
    threshold_mb: int
    percent_used: float
```

### system.overheating

Fired when CPU temperature exceeds threshold.

```yaml
topic: system.overheating
bpp_method: device.event
direction: Runtime → Client
frequency: Every 5 minutes while condition persists

payload:
  type: "system.overheating"
  data:
    temperature_celsius: float
    threshold_celsius: float
    throttling_active: boolean
    frequency_reduction_percent: int
```

### system.update_available

Fired when a new Runtime version is available.

```yaml
topic: system.update_available
bpp_method: device.event
direction: Runtime → Client
frequency: Once per update check

payload:
  type: "system.update_available"
  data:
    current_version: string
    new_version: string
    urgency: string        # "optional" | "recommended" | "required"
    changelog_url: string  # Link to release notes
    release_date: string   # ISO 8601
```

### system.update_progress

Fired during Runtime update installation.

```yaml
topic: system.update_progress
bpp_method: device.event
direction: Runtime → Client
frequency: Every 5-10 seconds during update

payload:
  type: "system.update_progress"
  data:
    phase: string          # "downloading" | "verifying" | "installing" | "restarting"
    progress_percent: int  # 0-100
    download_speed_kbps: int (optional)
    estimated_seconds_remaining: int (optional)
```

---

## Service Events

### service.started

Fired when a Runtime service starts.

```yaml
topic: service.started
bpp_method: capabilities.event
direction: Internal + Client (if connected)
frequency: Once per service start

payload:
  service: string          # "terminal" | "screen" | "gpio" | "docker" | etc.
  version: string
  started_at: string
```

### service.stopped

Fired when a Runtime service stops.

```yaml
topic: service.stopped
bpp_method: capabilities.event
direction: Internal + Client

payload:
  service: string
  reason: string           # "disabled" | "error" | "resource_exhausted" | "user_disabled"
  error: string (optional)
  will_restart: boolean
```

### service.error

Fired when a Runtime service encounters a non-fatal error.

```yaml
topic: service.error
bpp_method: capabilities.event
direction: Internal

payload:
  service: string
  error: string
  error_code: string           # From error code registry
  recoverable: boolean
  retry_count: int
```

---

## Capability Events

### capability.added

Fired when a new capability becomes available (e.g., camera plugged in).

```yaml
topic: capability.added
bpp_method: capabilities.event
direction: Runtime → Client
frequency: Infrequent (hardware event)

payload:
  type: "capability.added"
  data:
    capability:
      id: string
      version: string
      params: object (optional)
```

### capability.removed

Fired when a capability becomes unavailable (e.g., camera unplugged).

```yaml
topic: capability.removed
bpp_method: capabilities.event
direction: Runtime → Client

payload:
  type: "capability.removed"
  data:
    capability_id: string
    reason: string (optional)
```

### capability.updated

Fired when capability parameters change (e.g., screen quality degraded).

```yaml
topic: capability.updated
bpp_method: capabilities.event
direction: Runtime → Client

payload:
  type: "capability.updated"
  data:
    capability:
      id: string
      params: object           # New parameters
```

---

## Plugin Events

### plugin.installed

Fired when a plugin is installed.

```yaml
topic: plugin.installed
bpp_method: extension.event
direction: Runtime → Client

payload:
  type: "plugin.installed"
  data:
    plugin_id: string
    name: string
    version: string
    capabilities: string[]
```

### plugin.uninstalled

Fired when a plugin is removed.

```yaml
topic: plugin.uninstalled
bpp_method: extension.event
direction: Runtime → Client

payload:
  type: "plugin.uninstalled"
  data:
    plugin_id: string
```

### plugin.state.changed

Fired when a plugin transitions states.

```yaml
topic: plugin.state.changed
bpp_method: extension.event
direction: Internal + Client

payload:
  plugin_id: string
  from: string             # "installed" | "starting" | "active" | "idle" | "failed"
  to: string
  reason: string (optional)
```

### plugin.failed

Fired when a plugin enters failed state.

```yaml
topic: plugin.failed
bpp_method: extension.event
direction: Runtime → Client (push notification when applicable)

payload:
  type: "plugin.failed"
  data:
    plugin_id: string
    name: string
    reason: string         # "crash" | "oom" | "timeout" | "permission_denied"
    exit_code: int (optional)
    will_restart: boolean
    restart_attempt: int
    max_restarts: int
```

### plugin.event

Custom event pushed by a plugin.

```yaml
topic: plugin.event
bpp_method: extension.event
direction: Plugin → Runtime → Client

payload:
  type: "plugin.event"
  data:
    plugin_id: string
    event_type: string     # Plugin-defined (e.g., "temperature_alert")
    data: object           # Plugin-defined payload
```

---

## Terminal Events

### terminal.output

Fired when the terminal PTY produces output.

```yaml
topic: (per-session, dynamic)
bpp_method: terminal.output
direction: Runtime → Client
frequency: High throughput (every frame/line)

payload:
  session_id: string
  data: string             # Base64-encoded terminal output
  encoding: string         # "base64"
```

### terminal.session.closed

Fired when a terminal session ends.

```yaml
topic: (per-session, dynamic)
bpp_method: terminal.event
direction: Runtime → Client

payload:
  type: "terminal.session.closed"
  data:
    session_id: string
    reason: string         # "user_closed" | "timeout" | "process_exited"
    exit_code: int (optional)
    duration_seconds: int
```

---

## Screen Events

### screen.quality.changed

Fired when screen streaming quality adapts.

```yaml
topic: (per-session, dynamic)
bpp_method: screen.event
direction: Runtime → Client
frequency: On quality adaptation

payload:
  type: "screen.quality.changed"
  data:
    session_id: string
    from_quality: string   # "high" | "medium" | "low" | "minimum"
    to_quality: string
    reason: string         # "bandwidth_limited" | "cpu_throttled" | "user_requested"
    new_fps: int
    new_resolution: string
```

---

## Event Subscription

### Via Event Bus (Internal)

```go
// Subscribe to all connection events
sub := eventBus.Subscribe("connection.*", 32, DropOldest)

// Subscribe to everything
sub := eventBus.Subscribe("**", 64, DropOldest)

// Subscribe to a specific category
sub := eventBus.Subscribe("system.*", 16, DropOldest)
```

### Via BPP (Client)

```json
{
  "v": 1, "id": "msg_001", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "capabilities.subscribe",
  "params": {
    "events": ["capability.added", "capability.removed", "capability.updated"]
  },
  "rid": "req_001"
}
```

### Event Retention

| Category | Retention | Persistence |
|----------|-----------|-------------|
| Connection | None (ephemeral) | In-memory only |
| Device state | Last 100 | Persisted to local DB |
| System stats | Last 1000 | Persisted to local DB (rolling) |
| Service | Last 50 | Persisted to local DB |
| Capability | Last 100 | Persisted to local DB |
| Plugin | Last 100 | Persisted to local DB |
| Terminal | None (ephemeral) | In-memory only |
| Screen | None (ephemeral) | In-memory only |
