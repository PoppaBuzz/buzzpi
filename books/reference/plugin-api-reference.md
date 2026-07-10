# Plugin API Reference

**Public API surface for developing BuzzPi plugins.** This document defines the interfaces, message protocol, and best practices for writing plugins in any language.

---

## Overview

Plugins are independent processes that communicate with the BuzzPi Runtime over stdin/stdout using a length-prefixed JSON protocol. The Runtime discovers plugins via their manifest file and manages their lifecycle.

---

## IPC Protocol

All messages between the Runtime and a plugin use the same framing:

```
┌──────────────────────────────────┐
│ 4 bytes: payload length          │
│ (big-endian uint32)              │
├──────────────────────────────────┤
│ N bytes: JSON payload            │
└──────────────────────────────────┘
```

### Maximum Message Size

| Direction | Max Size |
|-----------|----------|
| Runtime → Plugin | 1 MB |
| Plugin → Runtime | 1 MB |

---

## Plugin Lifecycle

### Startup Sequence

```
Plugin                            Runtime
  │                                  │
  │  === Phase 1: Handshake ===      │
  │                                  │
  │ ─ ipc.hello ────────────────────►│
  │   {plugin_id: "com.example.x",   │
  │    version: "1.0.0",             │
  │    protocol_version: "1.0"}      │
  │                                  │
  │ ◄── ipc.hello.ack ───────────────│
  │   {runtime_version: "0.1.0",     │
  │    session_id: "ses_abc",        │
  │    config: {...}}                │
  │                                  │
  │  === Phase 2: Capabilities ===   │
  │                                  │
  │ ─ ipc.capabilities ─────────────►│
  │   {capabilities: [...]}          │
  │                                  │
  │  === Phase 3: Operational ===    │
  │                                  │
  │ ◄── ipc.request ─────────────────│  (when client invokes capability)
  │   {method: "ext.x.readings",     │
  │    params: {},                   │
  │    rid: "req_001"}              │
  │                                  │
  │ ─ ipc.response ─────────────────►│
  │   {rid: "req_001",              │
  │    result: {...}}                │
  │                                  │
  │ ─ ipc.event ────────────────────►│  (unsolicited push)
  │   {event_type: "alert",         │
  │    data: {...}}                  │
```

### Shutdown Sequence

```
Plugin                            Runtime
  │                                  │
  │ ◄── ipc.shutdown ────────────────│
  │   {reason: "idle_timeout",       │
  │    grace_period_seconds: 5}      │
  │                                  │
  │ ── [cleanup resources] ─────     │
  │                                  │
  │ ── [exit with code 0] ──────────►│
```

---

## Message Types

### Plugin → Runtime

#### ipc.hello

Sent immediately after plugin process starts.

```json
{
  "type": "ipc.hello",
  "plugin_id": "com.example.weather",
  "version": "1.0.0",
  "protocol_version": "1.0"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.hello"` |
| `plugin_id` | Yes | Must match manifest `plugin.id` |
| `version` | Yes | Plugin version from manifest |
| `protocol_version` | Yes | IPC protocol version (`"1.0"`) |

#### ipc.health

Response to health check ping.

```json
{
  "type": "ipc.health",
  "status": "ok",
  "stats": {
    "uptime_seconds": 3600,
    "memory_mb": 12.5,
    "cpu_percent": 2.3
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.health"` |
| `status` | Yes | `"ok"` or `"degraded"` |
| `stats` | No | Optional resource usage report |

#### ipc.capabilities

Declare plugin capabilities after handshake.

```json
{
  "type": "ipc.capabilities",
  "capabilities": [
    {
      "id": "extension.weather.readings",
      "version": "1.0",
      "available": true,
      "params": {
        "sensors": ["temperature", "humidity"]
      }
    }
  ]
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.capabilities"` |
| `capabilities[].id` | Yes | Fully qualified capability ID |
| `capabilities[].version` | Yes | Capability version |
| `capabilities[].available` | Yes | Currently available |
| `capabilities[].params` | No | Capability-specific parameters |

#### ipc.response

Response to a request from the Runtime.

```json
{
  "type": "ipc.response",
  "rid": "req_001",
  "result": {
    "temperature": 22.5,
    "humidity": 45.0,
    "unit": "celsius"
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.response"` |
| `rid` | Yes | Request ID from the received request |
| `result` | No | Response data (omit on error) |
| `error` | No | Error object (omit on success) |

#### ipc.event

Push an unsolicited event to connected clients.

```json
{
  "type": "ipc.event",
  "event_type": "temperature_alert",
  "data": {
    "temperature": 35.0,
    "threshold": 30.0
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.event"` |
| `event_type` | Yes | Event type identifier (plugin-namespaced) |
| `data` | No | Event payload |

#### ipc.log

Send a log message to the Runtime's log system.

```json
{
  "type": "ipc.log",
  "level": "info",
  "message": "Sensor read successful",
  "context": {
    "sensor": "bme280",
    "i2c_addr": "0x76",
    "duration_ms": 12
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.log"` |
| `level` | Yes | `"debug"`, `"info"`, `"warn"`, `"error"` |
| `message` | Yes | Log message |
| `context` | No | Structured context fields |

---

### Runtime → Plugin

#### ipc.hello.ack

Confirms handshake and provides runtime info.

```json
{
  "type": "ipc.hello.ack",
  "runtime_version": "0.1.0",
  "session_id": "ses_abc123",
  "config": {
    "plugin_id": "com.example.weather",
    "storage_dir": "/var/lib/buzzpi/plugins/com.example.weather/data",
    "cache_dir": "/var/lib/buzzpi/plugins/com.example.weather/cache",
    "granted_permissions": ["sensor.read"]
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.hello.ack"` |
| `runtime_version` | Yes | Runtime version string |
| `session_id` | Yes | Unique session identifier |
| `config.storage_dir` | Yes | Plugin's writable data directory |
| `config.cache_dir` | Yes | Plugin's cache directory |
| `config.granted_permissions` | Yes | List of granted permission IDs |

#### ipc.health

Health check ping.

```json
{
  "type": "ipc.health"
}
```

Plugin must respond within **5 seconds** with `ipc.health`.

#### ipc.request

Forward a BPP method invocation to the plugin.

```json
{
  "type": "ipc.request",
  "method": "extension.weather.readings",
  "params": {},
  "rid": "req_001",
  "client_id": "cli_abc123"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `"ipc.request"` |
| `method` | Yes | Method name matching a declared capability |
| `params` | No | Method parameters |
| `rid` | Yes | Request ID (echo in response) |
| `client_id` | No | Client identifier (for context) |

#### ipc.shutdown

Graceful shutdown signal.

```json
{
  "type": "ipc.shutdown",
  "reason": "idle_timeout",
  "grace_period_seconds": 5
}
```

Plugin must exit cleanly within the grace period. After the period, Runtime sends SIGKILL.

---

## Plugin SDK (Python Example)

```python
#!/usr/bin/env python3
"""
BuzzPi Plugin SDK — minimal Python implementation.

Usage:
    class MyPlugin(BuzzPiPlugin):
        def setup(self):
            # Initialize hardware, connections, etc.
            pass

        def handle_request(self, method, params, rid):
            if method == "extension.myplugin.action":
                result = self.do_something()
                self.send_response(rid, result)

    if __name__ == "__main__":
        MyPlugin("com.example.myplugin").run()
```

```python
import json
import struct
import sys
import os
import signal


class BuzzPiPlugin:
    """Base class for BuzzPi plugins."""

    def __init__(self, plugin_id):
        self.plugin_id = plugin_id
        self.stdin = sys.stdin.buffer
        self.stdout = sys.stdout.buffer
        self.storage_dir = None
        self.cache_dir = None
        self.granted_permissions = []
        self.running = True

        signal.signal(signal.SIGTERM, self._handle_signal)

    # ── Public API ──────────────────────────────────────

    def setup(self):
        """Override to initialize plugin resources."""
        pass

    def handle_request(self, method, params, rid):
        """Override to handle requests from the Runtime.
        
        Call self.send_response(rid, result) or self.send_error(rid, code, msg).
        """
        self.send_error(rid, "METHOD_NOT_FOUND", f"Unknown method: {method}")

    def cleanup(self):
        """Override to release resources on shutdown."""
        pass

    def send_response(self, rid, result):
        """Send a successful response."""
        self._send({
            "type": "ipc.response",
            "rid": rid,
            "result": result
        })

    def send_error(self, rid, code, message, data=None):
        """Send an error response."""
        error = {"code": code, "message": message}
        if data:
            error["data"] = data
        self._send({
            "type": "ipc.response",
            "rid": rid,
            "error": error
        })

    def push_event(self, event_type, data=None):
        """Push an unsolicited event to connected clients."""
        msg = {"type": "ipc.event", "event_type": event_type}
        if data is not None:
            msg["data"] = data
        self._send(msg)

    def log(self, level, message, **context):
        """Send a log message to the Runtime."""
        self._send({
            "type": "ipc.log",
            "level": level,
            "message": message,
            "context": context
        })

    def has_permission(self, permission_id):
        """Check if a permission has been granted."""
        return permission_id in self.granted_permissions

    # ── Internal ─────────────────────────────────────────

    def run(self):
        """Main entry point. Call this after construction."""
        # Wait for hello.ack
        hello_ack = self._read_message()
        if not hello_ack or hello_ack.get("type") != "ipc.hello.ack":
            self.log("error", "Handshake failed: expected ipc.hello.ack")
            sys.exit(1)

        self.storage_dir = hello_ack["config"]["storage_dir"]
        self.cache_dir = hello_ack["config"]["cache_dir"]
        self.granted_permissions = hello_ack["config"]["granted_permissions"]

        # Declare capabilities
        self._declare_capabilities()

        # Initialize plugin
        try:
            self.setup()
        except Exception as e:
            self.log("error", f"Setup failed: {e}")
            sys.exit(1)

        self.log("info", f"Plugin {self.plugin_id} started")

        # Message loop
        while self.running:
            try:
                msg = self._read_message()
                if msg is None:
                    break

                msg_type = msg.get("type")

                if msg_type == "ipc.health":
                    self._send({
                        "type": "ipc.health",
                        "status": "ok"
                    })

                elif msg_type == "ipc.request":
                    self.handle_request(
                        msg["method"],
                        msg.get("params", {}),
                        msg["rid"]
                    )

                elif msg_type == "ipc.shutdown":
                    self.log("info", f"Shutting down: {msg.get('reason')}")
                    break

            except Exception as e:
                self.log("error", f"Message loop error: {e}")

        self.cleanup()
        self.log("info", "Plugin stopped")

    def _declare_capabilities(self):
        """Override to send capabilities declaration."""
        self._send({
            "type": "ipc.capabilities",
            "capabilities": self.get_capabilities()
        })

    def get_capabilities(self):
        """Override to provide capability list.
        
        Returns list of dicts with keys: id, version, available, params (optional)
        """
        return []

    def _send(self, msg):
        data = json.dumps(msg).encode("utf-8")
        header = struct.pack(">I", len(data))
        try:
            self.stdout.write(header + data)
            self.stdout.flush()
        except BrokenPipeError:
            self.running = False

    def _read_message(self):
        header = self.stdin.read(4)
        if len(header) < 4:
            return None
        length = struct.unpack(">I", header)[0]
        if length > 1024 * 1024:  # 1MB max
            self.log("error", f"Message too large: {length} bytes")
            return None
        data = self.stdin.read(length)
        return json.loads(data)

    def _handle_signal(self, signum, frame):
        self.running = False
```

---

## Plugin SDK (Go Example)

```go
package plugin

import (
    "encoding/binary"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/signal"
    "syscall"
)

// Plugin is the base interface all plugins implement.
type Plugin interface {
    ID() string
    Setup(config PluginConfig) error
    HandleRequest(method string, params json.RawMessage, rid string) (interface{}, error)
    GetCapabilities() []Capability
    Cleanup()
}

type PluginConfig struct {
    SessionID           string
    StorageDir          string
    CacheDir            string
    GrantedPermissions  []string
}

type Capability struct {
    ID        string      `json:"id"`
    Version   string      `json:"version"`
    Available bool        `json:"available"`
    Params    interface{} `json:"params,omitempty"`
}

// PluginHost manages the IPC connection for a plugin.
type PluginHost struct {
    plugin  Plugin
    stdin   io.Reader
    stdout  io.Writer
    signals chan os.Signal
}

func NewPluginHost(p Plugin) *PluginHost {
    return &PluginHost{
        plugin:  p,
        stdin:   os.Stdin,
        stdout:  os.Stdout,
        signals: make(chan os.Signal, 1),
    }
}

func (h *PluginHost) Run() error {
    signal.Notify(h.signals, syscall.SIGTERM, syscall.SIGINT)

    // Wait for hello
    msg, err := h.readMessage()
    if err != nil {
        return fmt.Errorf("failed to read hello: %w", err)
    }
    if msg.Type != "ipc.hello" {
        return fmt.Errorf("expected ipc.hello, got %s", msg.Type)
    }

    // Send hello.ack
    config := PluginConfig{
        StorageDir: fmt.Sprintf("/var/lib/buzzpi/plugins/%s/data", h.plugin.ID()),
        CacheDir:   fmt.Sprintf("/var/lib/buzzpi/plugins/%s/cache", h.plugin.ID()),
    }
    h.sendMessage(Message{
        Type: "ipc.hello.ack",
        Data: map[string]interface{}{
            "runtime_version": "0.1.0",
            "session_id":      generateID(),
            "config":          config,
        },
    })

    // Receive capabilities
    caps := h.plugin.GetCapabilities()
    h.sendMessage(Message{
        Type: "ipc.capabilities",
        Data: map[string]interface{}{
            "capabilities": caps,
        },
    })

    // Initialize plugin
    if err := h.plugin.Setup(config); err != nil {
        return fmt.Errorf("plugin setup failed: %w", err)
    }

    // Message loop
    for {
        select {
        case <-h.signals:
            h.plugin.Cleanup()
            return nil

        default:
            msg, err := h.readMessage()
            if err != nil {
                return err
            }

            switch msg.Type {
            case "ipc.health":
                h.sendMessage(Message{
                    Type:   "ipc.health",
                    Data:   map[string]string{"status": "ok"},
                })

            case "ipc.request":
                var req Request
                json.Unmarshal(msg.Data, &req)
                result, err := h.plugin.HandleRequest(req.Method, req.Params, req.RID)
                if err != nil {
                    h.sendMessage(Message{
                        Type: "ipc.response",
                        Data: map[string]interface{}{
                            "rid": req.RID,
                            "error": map[string]string{
                                "code":    "INTERNAL_ERROR",
                                "message": err.Error(),
                            },
                        },
                    })
                } else {
                    h.sendMessage(Message{
                        Type: "ipc.response",
                        Data: map[string]interface{}{
                            "rid":    req.RID,
                            "result": result,
                        },
                    })
                }

            case "ipc.shutdown":
                h.plugin.Cleanup()
                return nil
            }
        }
    }
}

// Message is the IPC message envelope.
type Message struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data,omitempty"`
}

type Request struct {
    Method string          `json:"method"`
    Params json.RawMessage `json:"params"`
    RID    string          `json:"rid"`
}
```

---

## Best Practices

### 1. Startup Time

Keep plugin startup under 5 seconds. If initialization is slow (e.g., connecting to remote API), use async initialization and set capability `available: false` until ready, then push a `capability.updated` event.

### 2. Resource Usage

- Target <50MB RSS memory per plugin
- Target <10% CPU per plugin
- Close file descriptors when not in use
- Use the provided `storage_dir` for persistent data, never write outside it

### 3. Error Handling

- Always respond to requests with either `result` or `error`
- Never let exceptions propagate to the message loop
- Log errors with context before returning error responses
- Use standard BuzzPi error codes where applicable

### 4. Event Throttling

- Don't push events more than 10 times per second
- Batch rapid events where possible
- Events are delivered best-effort (may be dropped on slow connections)

### 5. Graceful Degradation

- If a dependency (sensor, network) is unavailable, set capability `available: false`
- Push a `capability.updated` event when dependency recovers
- Never crash — log the error and retry

### 6. Testing

- Test the plugin standalone (without BuzzPi Runtime) by piping IPC messages
- Use the provided test harness to simulate Runtime interactions

```bash
# Test plugin standalone
echo '{"type":"ipc.hello","plugin_id":"com.example.test","version":"1.0","protocol_version":"1.0"}' \
  | python3 my_plugin.py
```

---

## Error Codes for Plugins

| Code | Description | When to Use |
|------|-------------|-------------|
| `INTERNAL_ERROR` | Unexpected plugin error | Catch-all for unhandled errors |
| `METHOD_NOT_FOUND` | Plugin doesn't implement this method | Unknown capability method |
| `INVALID_PARAMS` | Method parameters are invalid | Validation failure |
| `PERMISSION_DENIED` | Missing required permission | Plugin checks via has_permission() |
| `RESOURCE_UNAVAILABLE` | Required hardware/network unavailable | Sensor not connected, API down |
| `TIMEOUT` | Operation took too long | External call timeout |
| `RATE_LIMITED` | Too many requests | Plugin is being overwhelmed |
