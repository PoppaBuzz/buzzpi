# WebSocket Events

**Complete reference of all WebSocket events flowing between clients, Runtime, and Cloud Relay.** Events are distinct from BPP method calls — they are unsolicited messages pushed from the Runtime to clients (or Relay to clients) rather than request-response pairs.

---

## Event Categories

| Category | Prefix | Direction | Examples |
|----------|--------|-----------|---------|
| Device Lifecycle | `device.*` | Runtime → Client | online, offline, restarting |
| Session | `session.*` | Runtime ↔ Client | created, expired, revoked |
| Screen | `screen.*` | Runtime → Client | frame, quality_changed, stopped |
| Terminal | `terminal.*` | Runtime → Client | output, closed, resize |
| File | `file.*` | Runtime → Client | progress, completed, error |
| GPIO | `gpio.*` | Runtime → Client | change, error |
| Camera | `camera.*` | Runtime → Client | frame, motion_detected, recording_stopped |
| Notification | `notification.*` | Runtime → Client | alert, warning, info |
| Plugin | `plugin.*` | Runtime → Client | loaded, unloaded, crashed |
| Connection | `connection.*` | Relay → Client | disconnected, reconnecting, quality_change |
| System | `system.*` | Runtime → Client | update_available, shutdown, reboot |

---

## Device Lifecycle Events

### device.online

Sent when a device comes online or connects to the relay.

```json
{
  "v": 1,
  "type": "event",
  "method": "device.online",
  "params": {
    "device_id": "dev_a1b2c3d4",
    "friendly_name": "living-room-pi",
    "runtime_version": "0.1.0",
    "relay_session": "sess_abc123"
  }
}
```

### device.offline

Sent when a device disconnects gracefully or heartbeats time out.

```json
{
  "v": 1,
  "type": "event",
  "method": "device.offline",
  "params": {
    "device_id": "dev_a1b2c3d4",
    "reason": "disconnected",
    "last_seen": "2026-07-07T12:00:00Z"
  }
}
```

Reasons: `disconnected` (graceful), `heartbeat_timeout` (180s no heartbeat), `relay_lost` (WebSocket dropped).

### device.restarting

Sent 5 seconds before the Runtime reboots.

```json
{
  "v": 1,
  "type": "event",
  "method": "device.restarting",
  "params": {
    "device_id": "dev_a1b2c3d4",
    "reason": "user_initiated",
    "restart_in_seconds": 5
  }
}
```

### device.disconnecting

Sent before a graceful WebSocket close.

```json
{
  "v": 1,
  "type": "event",
  "method": "device.disconnecting",
  "params": {
    "reason": "shutdown",
    "reconnect_in_seconds": 15
  }
}
```

---

## Session Events

### session.created

Sent when a new session is established.

```json
{
  "v": 1,
  "type": "event",
  "method": "session.created",
  "params": {
    "session_id": "sess_abc123",
    "client_id": "cli_xyz789",
    "client_name": "Pixel 9 Pro",
    "role": "admin",
    "expires_at": "2026-07-08T12:00:00Z"
  }
}
```

### session.expired

Sent when a session expires (sent to all other connected clients).

```json
{
  "v": 1,
  "type": "event",
  "method": "session.expired",
  "params": {
    "session_id": "sess_abc123",
    "client_id": "cli_xyz789"
  }
}
```

### session.revoked

Sent by the Runtime when an admin revokes a session.

```json
{
  "v": 1,
  "type": "event",
  "method": "session.revoked",
  "params": {
    "session_id": "sess_def456",
    "revoked_by": "cli_admin001",
    "reason": "admin_request"
  }
}
```

---

## Screen Events

### screen.frame

The most frequent event during screen streaming. Carries encoded video frame data as a binary WebSocket frame. The event metadata is sent as a JSON header immediately before the binary frame.

```json
// JSON header (text frame)
{
  "v": 1,
  "type": "event",
  "method": "screen.frame",
  "params": {
    "frame_id": 1423,
    "timestamp": 1234567890,
    "width": 1920,
    "height": 1080,
    "codec": "h264",
    "keyframe": false,
    "size_bytes": 24576
  }
}
// Binary frame follows immediately
```

### screen.quality_changed

Sent when the Connection Engine adapts quality.

```json
{
  "v": 1,
  "type": "event",
  "method": "screen.quality_changed",
  "params": {
    "quality": "medium",
    "reason": "network_degraded",
    "max_fps": 15,
    "max_resolution": "854x480",
    "estimated_bandwidth_kbps": 1500
  }
}
```

Reasons: `network_degraded`, `network_improved`, `user_request`, `cpu_throttled`.

### screen.stopped

```json
{
  "v": 1,
  "type": "event",
  "method": "screen.stopped",
  "params": {
    "reason": "user_request",
    "frames_sent": 1423,
    "duration_seconds": 47
  }
}
```

Reasons: `user_request`, `connection_lost`, `device_sleep`, `error`.

---

## Terminal Events

### terminal.output

Streaming terminal output (sent for each terminal session).

```json
{
  "v": 1,
  "type": "event",
  "method": "terminal.output",
  "params": {
    "session_id": "term_abc123",
    "data": "bXkgcHJvbXB0ICQg",  // base64-encoded terminal output
    "encoding": "base64"
  }
}
```

Terminal output is base64-encoded to handle binary control sequences safely.

### terminal.closed

```json
{
  "v": 1,
  "type": "event",
  "method": "terminal.closed",
  "params": {
    "session_id": "term_abc123",
    "exit_code": 0,
    "reason": "user_exit"
  }
}
```

Reasons: `user_exit`, `process_exit`, `timeout`, `error`.

### terminal.resize

Sent by the Runtime when the terminal's physical size changes (e.g., window resize on Android). Also sent by the client to request a resize.

```json
{
  "v": 1,
  "type": "event",
  "method": "terminal.resize",
  "params": {
    "session_id": "term_abc123",
    "rows": 40,
    "cols": 80
  }
}
```

---

## GPIO Events

### gpio.change

Sent when a watched GPIO pin changes state.

```json
{
  "v": 1,
  "type": "event",
  "method": "gpio.change",
  "params": {
    "pin": 17,
    "value": 1,
    "timestamp": 1234567890,
    "debounce_us": 200
  }
}
```

---

## Camera Events

### camera.frame

Similar to screen.frame but from the camera. Sent as JSON header + binary frame.

```json
{
  "v": 1,
  "type": "event",
  "method": "camera.frame",
  "params": {
    "frame_id": 89,
    "timestamp": 1234567890,
    "codec": "mjpeg",
    "size_bytes": 65536
  }
}
```

### camera.motion_detected

```json
{
  "v": 1,
  "type": "event",
  "method": "camera.motion_detected",
  "params": {
    "sensitivity": 75,
    "region": {"x": 100, "y": 200, "width": 300, "height": 400},
    "timestamp": 1234567890
  }
}
```

### camera.recording_stopped

```json
{
  "v": 1,
  "type": "event",
  "method": "camera.recording_stopped",
  "params": {
    "duration_seconds": 30,
    "file_size_bytes": 5242880,
    "file_path": "/var/lib/buzzpi/captures/recording_20260707_120000.mp4",
    "reason": "user_request"
  }
}
```

---

## Notification Events

### notification.alert

High-priority notification demanding user attention.

```json
{
  "v": 1,
  "type": "event",
  "method": "notification.alert",
  "params": {
    "id": "notif_abc123",
    "title": "High Temperature",
    "body": "CPU temperature reached 82°C. Check cooling.",
    "severity": "warning",
    "timestamp": 1234567890,
    "action": {
      "type": "deep_link",
      "uri": "buzzpi://device/dev_a1b2c3d4/system"
    }
  }
}
```

Severity values: `info`, `warning`, `critical`.

### notification.info

Informational notification.

```json
{
  "v": 1,
  "type": "event",
  "method": "notification.info",
  "params": {
    "id": "notif_def456",
    "title": "Update Available",
    "body": "BuzzPi Runtime v0.2.0 is available.",
    "timestamp": 1234567890,
    "action": {
      "type": "deep_link",
      "uri": "buzzpi://updates"
    }
  }
}
```

---

## Plugin Events

### plugin.loaded

Sent when the Plugin Host successfully loads a plugin.

```json
{
  "v": 1,
  "type": "event",
  "method": "plugin.loaded",
  "params": {
    "plugin_id": "docker-manager",
    "version": "1.2.0",
    "capabilities": ["container.list", "container.logs", "container.exec"],
    "pid": 5678
  }
}
```

### plugin.unloaded

```json
{
  "v": 1,
  "type": "event",
  "method": "plugin.unloaded",
  "params": {
    "plugin_id": "docker-manager",
    "reason": "user_disabled"
  }
}
```

### plugin.crashed

```json
{
  "v": 1,
  "type": "event",
  "method": "plugin.crashed",
  "params": {
    "plugin_id": "docker-manager",
    "exit_code": 137,
    "signal": "SIGKILL",
    "restart_count": 3,
    "auto_restart": true,
    "last_logs": "FATAL: out of memory\n"
  }
}
```

---

## Connection Events

### connection.disconnected (Relay → Client)

Sent by the Cloud Relay when the device connection drops.

```json
{
  "v": 1,
  "type": "event",
  "method": "connection.disconnected",
  "params": {
    "device_id": "dev_a1b2c3d4",
    "reason": "relay_lost",
    "auto_reconnect": true
  }
}
```

### connection.reconnecting (Relay → Client)

```json
{
  "v": 1,
  "type": "event",
  "method": "connection.reconnecting",
  "params": {
    "device_id": "dev_a1b2c3d4",
    "attempt": 2,
    "max_attempts": 5,
    "backoff_seconds": 4
  }
}
```

### connection.quality_change

Sent when the Connection Engine switches transport (LAN ↔ Relay) or quality adapts.

```json
{
  "v": 1,
  "type": "event",
  "method": "connection.quality_change",
  "params": {
    "device_id": "dev_a1b2c3d4",
    "transport": "lan",
    "rtt_ms": 5,
    "reason": "device_discovered_on_lan"
  }
}
```

Transport values: `lan`, `relay`, `p2p`.
Reason values: `device_discovered_on_lan`, `relay_fallback`, `p2p_established`, `quality_adapted`.

---

## System Events

### system.update_available

```json
{
  "v": 1,
  "type": "event",
  "method": "system.update_available",
  "params": {
    "current_version": "0.1.0",
    "available_version": "0.2.0",
    "release_date": "2026-08-01",
    "changelog_url": "https://github.com/buzzpi/runtime/releases/v0.2.0",
    "critical": false
  }
}
```

### system.shutdown

```json
{
  "v": 1,
  "type": "event",
  "method": "system.shutdown",
  "params": {
    "reason": "user_initiated",
    "shutdown_in_seconds": 10
  }
}
```

### system.reboot

```json
{
  "v": 1,
  "type": "event",
  "method": "system.reboot",
  "params": {
    "reason": "update_installed",
    "reboot_in_seconds": 5
  }
}
```

---

## Error Events

All error events follow the same schema:

```json
{
  "v": 1,
  "type": "event",
  "method": "error.occurred",
  "params": {
    "code": "device.stats_read_failed",
    "message": "Failed to read CPU temperature",
    "component": "system_monitor",
    "severity": "warning",
    "recoverable": true,
    "timestamp": 1234567890
  }
}
```

---

## Event Subscription

Clients can subscribe to specific event types to reduce bandwidth:

```
Client → Runtime: pair.subscribe
{
  "events": ["device.*", "notification.*", "system.update_available"],
  "exclude": ["screen.frame", "camera.frame"]   // high-frequency events
}
```

Default subscription (when not specified): all events except `screen.*` and `camera.*` (which require explicit opt-in).

---

## Event Delivery Guarantees

| Event Type | Delivery | Persistence |
|------------|----------|-------------|
| `screen.frame` | Best-effort (latest frame only) | Not persisted |
| `camera.frame` | Best-effort (latest frame only) | Not persisted |
| `terminal.output` | Reliable (in-order, buffered) | Not persisted |
| `notification.*` | At-least-once (retry until ACK) | Persisted until delivered |
| `device.*` | At-least-once | Persisted until delivered |
| `session.*` | Exactly-once | Persisted |

For at-least-once delivery, clients must send an ACK:

```
Client → Runtime: {"type":"event","method":"event.ack","params":{"event_id":"evt_abc123"}}
```

Events not ACK'd within 30 seconds are re-delivered (up to 3 times).
