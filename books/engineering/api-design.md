# API Design

BuzzPi defines APIs at three boundaries: **App ↔ Relay**, **Relay ↔ Runtime**, and **Runtime ↔ Extensions**. All follow consistent design principles.

## Principles

### Message-Oriented

All communication uses message passing, not REST or RPC in the traditional sense. Messages are asynchronous by default, with request-response as a pattern built on top.

### Capability-Discoverable

When a connection is established, both sides announce their capabilities. The caller never assumes what the callee supports. This enables forward compatibility: an old app talking to a new Runtime, or vice versa.

### Binary over WebRTC, JSON over WebSocket

- **WebSocket (signaling/control):** JSON messages, human-readable, debuggable.
- **WebRTC Data Channel (data):** Binary protocol (protobuf or flatbuffers), efficient for high-throughput terminal and file data.
- **WebRTC Media Track (screen):** H.264 video, standard codec.

---

## Message Format

### Envelope

Every message (WebSocket or Data Channel) uses a standard envelope:

```json
{
  "v": 1,
  "id": "msg_abc123",
  "ts": "2026-07-07T12:00:00Z",
  "type": "request",
  "method": "device.terminal.write",
  "params": { "data": "ls -la\n" },
  "rid": "req_xyz789"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `v` | Yes | Protocol version (currently 1) |
| `id` | Yes | Unique message identifier |
| `ts` | Yes | ISO 8601 timestamp (UTC) |
| `type` | Yes | `request`, `response`, `event`, `error` |
| `method` | Conditional | The method being called (requests and events) |
| `params` | Conditional | Method parameters (requests and events) |
| `rid` | Conditional | Request ID — sets `type` to `request` when present, `response` when matching |
| `result` | Conditional | Response payload (type: response) |
| `error` | Conditional | Error payload (type: error) |

### Error Response

```json
{
  "v": 1,
  "id": "msg_def456",
  "ts": "2026-07-07T12:00:01Z",
  "type": "error",
  "rid": "req_xyz789",
  "error": {
    "code": "METHOD_NOT_FOUND",
    "message": "device.terminal.clear is not available on this device",
    "data": {}
  }
}
```

---

## Methods by Component

### App ↔ Relay (Signaling)

| Method | Direction | Description |
|--------|-----------|-------------|
| `relay.connect` | App → Relay | Authenticate and open signaling session |
| `relay.disconnect` | App → Relay | Close signaling session |
| `relay.device.list` | App → Relay | Get user's devices with current state |
| `relay.device.connect` | App → Relay | Request connection to a device |
| `relay.device.disconnect` | App → Relay | Disconnect from a device |
| `relay.pair.init` | App → Relay | Initiate pairing |
| `relay.pair.verify` | App → Relay | Submit pairing code |
| `relay.pair.cancel` | App → Relay | Cancel pairing |
| `relay.device.event` | Relay → App | Device state change notification |
| `relay.notification.send` | Relay → App | Send push notification |

### Runtime ↔ Relay (Signaling)

| Method | Direction | Description |
|--------|-----------|-------------|
| `relay.register` | Runtime → Relay | Register device identity and availability |
| `relay.heartbeat` | Runtime → Relay | Keep-alive |
| `relay.unregister` | Runtime → Relay | Device going offline intentionally |
| `relay.connect.request` | Relay → Runtime | Incoming connection request |
| `relay.connect.accept` | Runtime → Relay | Accept incoming connection |
| `relay.connect.reject` | Runtime → Relay | Reject incoming connection |
| `relay.connect.close` | Runtime → Relay | Notify relay of connection close |

### App ↔ Runtime (WebRTC Data Channel)

| Method | Direction | Description |
|--------|-----------|-------------|
| `session.capabilities` | Bidirectional | Announce supported capabilities |
| `session.ping` | Bidirectional | Latency check |
| `terminal.open` | App → Runtime | Open terminal session |
| `terminal.close` | App → Runtime | Close terminal session |
| `terminal.write` | App → Runtime | Write to terminal stdin |
| `terminal.output` | Runtime → App | Terminal stdout |
| `terminal.resize` | App → Runtime | Resize terminal (cols, rows) |
| `screen.start` | App → Runtime | Start screen streaming |
| `screen.stop` | App → Runtime | Stop screen streaming |
| `screen.frame` | Runtime → App | Video frame (H.264, over media track) |
| `screen.input` | App → Runtime | Mouse/keyboard input |
| `services.list` | App → Runtime | List installed extensions |
| `services.status` | App → Runtime | Get service status |
| `services.start` | App → Runtime | Start a service |
| `services.stop` | App → Runtime | Stop a service |
| `services.restart` | App → Runtime | Restart a service |
| `services.logs` | App → Runtime | Get service logs |
| `files.list` | App → Runtime | List directory contents |
| `files.read` | App → Runtime | Read file contents |
| `files.write` | App → Runtime | Write file contents |
| `files.delete` | App → Runtime | Delete file |
| `device.info` | App → Runtime | Get device information |
| `device.update` | App → Runtime | Trigger Runtime update |
| `device.restart` | App → Runtime | Restart device |
| `device.shutdown` | App → Runtime | Shutdown device |
| `action.execute` | App → Runtime | Execute custom action |
| `action.status` | Runtime → App | Action progress/completion |
| `extension.install` | App → Runtime | Install extension |
| `extension.uninstall` | App → Runtime | Remove extension |
| `extension.configure` | App → Runtime | Configure extension |

---

## Versioning

- **Protocol version** is incremented when the message envelope or core methods change
- **Method versioning** is per-method via `v` field inside params (e.g., terminal methods may be at v1 while files may be at v2)
- **Backward compatibility** is required for at least one major version
- Unknown methods are rejected with `METHOD_NOT_FOUND`, not silently ignored

## Rate Limiting

| Component | Limit | Burst |
|-----------|-------|-------|
| App → Relay (signaling) | 100 req/s | 200 |
| Runtime → Relay (signaling) | 10 req/s | 50 |
| App → Runtime (data channel) | 1000 msg/s | 5000 |
| Relay → App (notifications) | 10 msg/s per device | 20 |

## Streaming

Methods that produce streaming data (terminal output, service logs) use a subscription pattern:

1. Caller sends `terminal.open` (request)
2. Runtime sends `terminal.output` (event) messages with the same `rid`
3. Caller sends `terminal.close` to unsubscribe
4. Runtime stops sending events for that `rid`

This avoids separate subscribe/unsubscribe messages — the open/close pattern is self-contained.
