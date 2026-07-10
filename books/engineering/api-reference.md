# API Reference

**Complete catalog of every BPP method, backend endpoint, and IPC message.** This is the definitive reference for developers implementing clients, server components, or plugins against the BuzzPi protocol and platform.

---

## Organization

The API surface is grouped by boundary:

| Layer | Transport | Audience |
|-------|-----------|----------|
| **BPP** (Runtime ↔ Client) | WebSocket / WebRTC | Every client implements these |
| **Relay** (Client ↔ Relay ↔ Runtime) | WebSocket | Signaling only |
| **REST** (Client ↔ Registry) | HTTPS | Account and device management |
| **IPC** (Runtime ↔ Plugin) | stdin/stdout | Plugin developers |

---

# BPP Methods

All communication between clients and the Runtime uses the BPP message envelope over WebSocket (signaling) or WebRTC Data Channel (data). See [API Design](api-design.md) for the envelope format.

---

## Device Methods

### device.info

Get device identity and hardware information.

```json
{
  "v": 1, "id": "msg_001", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "device.info", "rid": "req_001"
}
```

```json
{
  "v": 1, "id": "msg_002", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_001",
  "result": {
    "device_id": "dev_abc123",
    "friendly_name": "Kitchen Pi",
    "model": "Raspberry Pi 5",
    "runtime_version": "0.1.0",
    "uptime_seconds": 86400,
    "capabilities": ["service.terminal", "service.screen", "hardware.gpio"],
    "platform": "linux/arm64"
  }
}
```

### device.stats

Get current system statistics.

```json
{
  "v": 1, "id": "msg_003", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "device.stats", "rid": "req_002"
}
```

```json
{
  "v": 1, "id": "msg_004", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_002",
  "result": {
    "cpu": {
      "usage_percent": 23.5,
      "temperature_celsius": 65.2,
      "frequency_mhz": 1800
    },
    "memory": {
      "total_mb": 8192,
      "used_mb": 3547,
      "available_mb": 4645,
      "percent": 43.3
    },
    "storage": [
      {"mount": "/", "total_mb": 30528, "used_mb": 22112, "available_mb": 8416, "percent": 72.4}
    ],
    "network": {
      "interface": "wlan0",
      "ip": "192.168.1.42",
      "rx_bytes": 1500000000,
      "tx_bytes": 800000000,
      "signal_percent": 85
    },
    "uptime_seconds": 86400
  }
}
```

### device.reboot

Reboot the device (requires confirmation).

```json
{
  "v": 1, "id": "msg_005", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "device.reboot",
  "params": {"confirm": true},
  "rid": "req_003"
}
```

```json
{
  "v": 1, "id": "msg_006", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_003",
  "result": {"status": "rebooting", "estimated_downtime_seconds": 60}
}
```

### device.shutdown

Shut down the device (requires confirmation).

```json
{
  "v": 1, "id": "msg_007", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "device.shutdown",
  "params": {"confirm": true},
  "rid": "req_004"
}
```

```json
{
  "v": 1, "id": "msg_008", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_004",
  "result": {"status": "shutting_down"}
}
```

### device.event

Push event from device to client.

```json
{
  "v": 1, "id": "msg_009", "ts": "2026-07-07T12:00:05Z",
  "type": "event", "method": "device.event",
  "params": {
    "type": "temperature_alert",
    "data": {"temperature": 82.0, "threshold": 80.0}
  }
}
```

---

## Terminal Methods

### terminal.open

Open a new PTY session.

```json
{
  "v": 1, "id": "msg_010", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "terminal.open",
  "params": {
    "rows": 40,
    "cols": 80,
    "shell": "/bin/bash"
  },
  "rid": "req_005"
}
```

```json
{
  "v": 1, "id": "msg_011", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_005",
  "result": {
    "session_id": "term_001",
    "rows": 40,
    "cols": 80
  }
}
```

### terminal.write

Write data to terminal input.

```json
{
  "v": 1, "id": "msg_012", "ts": "2026-07-07T12:00:02Z",
  "type": "request", "method": "terminal.write",
  "params": {
    "session_id": "term_001",
    "data": "bG9nZ2VyIC10YWlsIC01MAo="
  },
  "rid": "req_006"
}
```

### terminal.output

Stream terminal output to client (pushed).

```json
{
  "v": 1, "id": "msg_013", "ts": "2026-07-07T12:00:03Z",
  "type": "event", "method": "terminal.output",
  "params": {
    "session_id": "term_001",
    "data": "SnVuICBzIDExIDIxOjU3OmJhenogS2l0Y2hlbiBQaVs...
    "encoding": "base64"
  }
}
```

### terminal.resize

Resize terminal dimensions.

```json
{
  "v": 1, "id": "msg_014", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "terminal.resize",
  "params": {"session_id": "term_001", "rows": 50, "cols": 100},
  "rid": "req_007"
}
```

### terminal.close

Close terminal session.

```json
{
  "v": 1, "id": "msg_015", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "terminal.close",
  "params": {"session_id": "term_001"},
  "rid": "req_008"
}
```

---

## Screen / Desktop Methods

### screen.start

Start screen streaming.

```json
{
  "v": 1, "id": "msg_016", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "screen.start",
  "params": {
    "quality": "high",
    "max_fps": 30,
    "max_resolution": "1920x1080",
    "cursor_visible": true
  },
  "rid": "req_009"
}
```

```json
{
  "v": 1, "id": "msg_017", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_009",
  "result": {
    "session_id": "screen_001",
    "actual_fps": 30,
    "actual_resolution": "1920x1080",
    "codec": "h264",
    "capture_method": "drm"
  }
}
```

### screen.stop

Stop screen streaming.

```json
{
  "v": 1, "id": "msg_018", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "screen.stop",
  "params": {"session_id": "screen_001"},
  "rid": "req_010"
}
```

### screen.input

Send mouse/touch input.

```json
{
  "v": 1, "id": "msg_019", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "screen.input",
  "params": {
    "session_id": "screen_001",
    "type": "mouse",
    "action": "mousedown",
    "x": 960,
    "y": 540,
    "button": "left"
  },
  "rid": "req_011"
}
```

### screen.quality

Update streaming quality parameters.

```json
{
  "v": 1, "id": "msg_020", "ts": "2026-07-07T12:01:00Z",
  "type": "request", "method": "screen.quality",
  "params": {
    "session_id": "screen_001",
    "quality": "medium",
    "max_fps": 15,
    "max_resolution": "1280x720"
  },
  "rid": "req_012"
}
```

---

## File Methods

### files.list

List directory contents.

```json
{
  "v": 1, "id": "msg_021", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "files.list",
  "params": {"path": "/home/pi"},
  "rid": "req_013"
}
```

```json
{
  "v": 1, "id": "msg_022", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_013",
  "result": {
    "path": "/home/pi",
    "entries": [
      {"name": "Documents", "type": "directory", "size": 4096, "modified": "2026-07-01T10:00:00Z", "mode": "drwxr-xr-x"},
      {"name": "config.txt", "type": "file", "size": 2048, "modified": "2026-07-05T14:30:00Z", "mode": "-rw-r--r--"},
      {"name": "image.jpg", "type": "file", "size": 5242880, "modified": "2026-06-28T09:15:00Z", "mode": "-rw-r--r--"}
    ]
  }
}
```

### files.read

Read file contents (text or base64).

```json
{
  "v": 1, "id": "msg_023", "ts": "2026-07-07T12:00:02Z",
  "type": "request", "method": "files.read",
  "params": {"path": "/home/pi/config.txt", "encoding": "text"},
  "rid": "req_014"
}
```

```json
{
  "v": 1, "id": "msg_024", "ts": "2026-07-07T12:00:03Z",
  "type": "response", "rid": "req_014",
  "result": {
    "path": "/home/pi/config.txt",
    "content": "# Raspberry Pi configuration\n...",
    "size": 2048,
    "encoding": "text"
  }
}
```

### files.write

Write content to a file.

```json
{
  "v": 1, "id": "msg_025", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "files.write",
  "params": {
    "path": "/home/pi/new_config.txt",
    "content": "new content...",
    "encoding": "text",
    "append": false
  },
  "rid": "req_015"
}
```

### files.delete

Delete a file or empty directory.

```json
{
  "v": 1, "id": "msg_026", "ts": "2026-07-07T12:00:06Z",
  "type": "request", "method": "files.delete",
  "params": {"path": "/home/pi/temp.txt", "recursive": false},
  "rid": "req_016"
}
```

### files.mkdir

Create a directory.

```json
{
  "v": 1, "id": "msg_027", "ts": "2026-07-07T12:00:07Z",
  "type": "request", "method": "files.mkdir",
  "params": {"path": "/home/pi/newdir", "mode": "0755"},
  "rid": "req_017"
}
```

### files.move

Move or rename a file.

```json
{
  "v": 1, "id": "msg_028", "ts": "2026-07-07T12:00:08Z",
  "type": "request", "method": "files.move",
  "params": {"from": "/home/pi/temp.txt", "to": "/home/pi/Documents/notes.txt"},
  "rid": "req_018"
}
```

### files.chunked

Chunked file transfer for large files (sent over WebRTC Data Channel).

```json
{
  "v": 1, "id": "msg_029", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "files.chunked",
  "params": {
    "transfer_id": "xfer_001",
    "action": "start",
    "path": "/home/pi/large_file.bin",
    "direction": "upload",
    "total_size": 104857600,
    "checksum_algorithm": "sha256"
  },
  "rid": "req_019"
}
```

---

## Docker Methods

### docker.ps

List Docker containers.

```json
{
  "v": 1, "id": "msg_030", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "docker.ps",
  "params": {"all": false},
  "rid": "req_020"
}
```

```json
{
  "v": 1, "id": "msg_031", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_020",
  "result": {
    "containers": [
      {
        "id": "abc123def456",
        "name": "nginx",
        "image": "nginx:latest",
        "status": "running",
        "created": "2026-07-01T08:00:00Z",
        "ports": ["0.0.0.0:80->80/tcp"],
        "state": "running"
      }
    ]
  }
}
```

### docker.inspect

Inspect a container's details.

```json
{
  "v": 1, "id": "msg_032", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "docker.inspect",
  "params": {"container_id": "abc123def456"},
  "rid": "req_021"
}
```

### docker.logs

Get container logs.

```json
{
  "v": 1, "id": "msg_033", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "docker.logs",
  "params": {"container_id": "abc123def456", "tail": 50, "follow": false},
  "rid": "req_022"
}
```

### docker.start / docker.stop / docker.restart

Container lifecycle management.

```json
{
  "v": 1, "id": "msg_034", "ts": "2026-07-07T12:00:15Z",
  "type": "request", "method": "docker.restart",
  "params": {"container_id": "abc123def456"},
  "rid": "req_023"
}
```

### docker.stats

Container resource usage.

```json
{
  "v": 1, "id": "msg_035", "ts": "2026-07-07T12:00:20Z",
  "type": "request", "method": "docker.stats",
  "params": {"container_id": "abc123def456"},
  "rid": "req_024"
}
```

```json
{
  "v": 1, "id": "msg_036", "ts": "2026-07-07T12:00:21Z",
  "type": "response", "rid": "req_024",
  "result": {
    "cpu_percent": 2.5,
    "memory_mb": 64,
    "memory_percent": 0.8,
    "network_rx_bytes": 5000000,
    "network_tx_bytes": 20000000
  }
}
```

### docker.images

List Docker images.

```json
{
  "v": 1, "id": "msg_037", "ts": "2026-07-07T12:00:25Z",
  "type": "request", "method": "docker.images",
  "params": {},
  "rid": "req_025"
}
```

---

## GPIO Methods

### gpio.list

List available GPIO pins with current states.

```json
{
  "v": 1, "id": "msg_038", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "gpio.list",
  "params": {},
  "rid": "req_026"
}
```

```json
{
  "v": 1, "id": "msg_039", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_026",
  "result": {
    "pins": [
      {"pin": 2, "name": "GPIO2", "mode": "input", "value": 1, "function": "SDA1"},
      {"pin": 3, "name": "GPIO3", "mode": "input", "value": 1, "function": "SCL1"},
      {"pin": 4, "name": "GPIO4", "mode": "output", "value": 0, "function": "GPCLK0"},
      {"pin": 17, "name": "GPIO17", "mode": "output", "value": 1, "function": "GPIO17"},
      {"pin": 27, "name": "GPIO27", "mode": "input", "value": 0, "function": "GPIO27"},
      {"pin": 22, "name": "GPIO22", "mode": "pwm", "value": 128, "function": "GPIO22"}
    ]
  }
}
```

### gpio.read

Read a single pin.

```json
{
  "v": 1, "id": "msg_040", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "gpio.read",
  "params": {"pin": 17},
  "rid": "req_027"
}
```

```json
{
  "v": 1, "id": "msg_041", "ts": "2026-07-07T12:00:06Z",
  "type": "response", "rid": "req_027",
  "result": {"pin": 17, "mode": "input", "value": 1}
}
```

### gpio.write

Write a pin value.

```json
{
  "v": 1, "id": "msg_042", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "gpio.write",
  "params": {"pin": 4, "value": 1},
  "rid": "req_028"
}
```

### gpio.pwm

Set PWM duty cycle.

```json
{
  "v": 1, "id": "msg_043", "ts": "2026-07-07T12:00:15Z",
  "type": "request", "method": "gpio.pwm",
  "params": {"pin": 22, "duty_cycle": 192, "frequency": 1000},
  "rid": "req_029"
}
```

### gpio.watch

Watch a pin for state changes (pushes events).

```json
{
  "v": 1, "id": "msg_044", "ts": "2026-07-07T12:00:20Z",
  "type": "request", "method": "gpio.watch",
  "params": {"pin": 17, "edge": "both"},
  "rid": "req_030"
}
```

### gpio.event

Push event from GPIO watcher.

```json
{
  "v": 1, "id": "msg_045", "ts": "2026-07-07T12:00:25Z",
  "type": "event", "method": "gpio.event",
  "params": {"pin": 17, "value": 0, "edge": "falling"}
}
```

---

## Camera Methods

### camera.list

List available camera devices.

```json
{
  "v": 1, "id": "msg_046", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "camera.list",
  "params": {},
  "rid": "req_031"
}
```

### camera.start

Start camera stream.

```json
{
  "v": 1, "id": "msg_047", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "camera.start",
  "params": {
    "device": "/dev/video0",
    "resolution": "1920x1080",
    "fps": 30,
    "format": "h264"
  },
  "rid": "req_032"
}
```

### camera.stop

Stop camera stream.

```json
{
  "v": 1, "id": "msg_048", "ts": "2026-07-07T12:00:30Z",
  "type": "request", "method": "camera.stop",
  "params": {"session_id": "cam_001"},
  "rid": "req_033"
}
```

---

## Log Methods

### logs.list

List available log sources.

```json
{
  "v": 1, "id": "msg_049", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "logs.list",
  "params": {},
  "rid": "req_034"
}
```

```json
{
  "v": 1, "id": "msg_050", "ts": "2026-07-07T12:00:01Z",
  "type": "response", "rid": "req_034",
  "result": {
    "sources": [
      {"name": "syslog", "description": "System log (syslog)"},
      {"name": "buzzpi", "description": "BuzzPi Runtime log"},
      {"name": "dmesg", "description": "Kernel ring buffer"},
      {"name": "auth", "description": "Authentication log"},
      {"name": "custom:nginx", "description": "Nginx access log"}
    ]
  }
}
```

### logs.read

Read log entries.

```json
{
  "v": 1, "id": "msg_051", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "logs.read",
  "params": {
    "source": "buzzpi",
    "tail": 100,
    "priority": 6,    // 0=emerg .. 7=debug
    "follow": false
  },
  "rid": "req_035"
}
```

```json
{
  "v": 1, "id": "msg_052", "ts": "2026-07-07T12:00:06Z",
  "type": "response", "rid": "req_035",
  "result": {
    "source": "buzzpi",
    "entries": [
      {"timestamp": "2026-07-07T11:59:00Z", "level": "info", "message": "Connection established: client_abc"},
      {"timestamp": "2026-07-07T11:58:30Z", "level": "warn", "message": "Temperature: 75C, approaching threshold"},
      {"timestamp": "2026-07-07T11:58:00Z", "level": "error", "message": "Screen capture failed: DRM buffer unavailable"}
    ]
  }
}
```

### logs.follow

Stream new log entries in real time.

```json
{
  "v": 1, "id": "msg_053", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "logs.follow",
  "params": {"source": "buzzpi", "priority": 6},
  "rid": "req_036"
}
```

---

## Capability Methods

### capabilities.list

Get all device capabilities.

```json
{
  "v": 1, "id": "msg_054", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "capabilities.list",
  "params": {},
  "rid": "req_037"
}
```

### capabilities.subscribe

Subscribe to capability change events.

```json
{
  "v": 1, "id": "msg_055", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "capabilities.subscribe",
  "params": {"events": ["capability.added", "capability.removed", "capability.updated"]},
  "rid": "req_038"
}
```

### capabilities.event

Push capability change.

```json
{
  "v": 1, "id": "msg_056", "ts": "2026-07-07T12:01:00Z",
  "type": "event", "method": "capabilities.event",
  "params": {
    "type": "capability.updated",
    "capability": {"id": "service.screen", "available": true, "params": {"max_fps": 10}}
  }
}
```

---

## Extension (Plugin) Methods

### extension.list

List installed extensions.

```json
{
  "v": 1, "id": "msg_057", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "extension.list",
  "params": {},
  "rid": "req_039"
}
```

### extension.install

Install an extension.

```json
{
  "v": 1, "id": "msg_058", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "extension.install",
  "params": {"id": "com.example.weather", "source": "registry", "version": "1.0.0"},
  "rid": "req_040"
}
```

### extension.uninstall

Remove an extension.

```json
{
  "v": 1, "id": "msg_059", "ts": "2026-07-07T12:01:00Z",
  "type": "request", "method": "extension.uninstall",
  "params": {"id": "com.example.weather"},
  "rid": "req_041"
}
```

### extension.start / extension.stop

Start or stop a plugin process.

```json
{
  "v": 1, "id": "msg_060", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "extension.start",
  "params": {"id": "com.example.weather"},
  "rid": "req_042"
}
```

### extension.permissions

Manage extension permissions.

```json
{
  "v": 1, "id": "msg_061", "ts": "2026-07-07T12:00:15Z",
  "type": "request", "method": "extension.permissions",
  "params": {"id": "com.example.weather", "action": "grant", "permissions": ["sensor.read"]},
  "rid": "req_043"
}
```

### extension.event

Extension-pushed custom event.

```json
{
  "v": 1, "id": "msg_062", "ts": "2026-07-07T12:05:00Z",
  "type": "event", "method": "extension.event",
  "params": {
    "extension_id": "com.example.weather",
    "event_type": "temperature_alert",
    "data": {"temperature": 35.0, "threshold": 30.0}
  }
}
```

---

## BuzzAI Methods

### buzzai.ask

Send a natural language query.

```json
{
  "v": 1, "id": "msg_063", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "buzzai.ask",
  "params": {
    "message": "Why is my Pi running slow?",
    "conversation_id": null,
    "context": {"device_id": "dev_abc123"}
  },
  "rid": "req_044"
}
```

```json
{
  "v": 1, "id": "msg_064", "ts": "2026-07-07T12:00:03Z",
  "type": "response", "rid": "req_044",
  "result": {
    "response": "Your Pi is thermal throttling at 82°C. The CPU has reduced its frequency from 2.4GHz to 1.2GHz. Check the heatsink and fan, or relocate to a cooler area.\n\nTemperature: 82°C (threshold: 80°C)\nCPU frequency: 1.2GHz (expected: 2.4GHz)\nMemory: 65% used\nDisk: 72% full",
    "conversation_id": "conv_001",
    "actions": [
      {"id": "act_001", "label": "Show temperature graph", "tool": "system_status", "params": {"focus": "temperature"}},
      {"id": "act_002", "label": "Check which processes use CPU", "tool": "execute_command", "params": {"command": "top -bn1 | head -20"}}
    ],
    "requires_confirmation": false
  }
}
```

### buzzai.confirm

Confirm or reject a pending action.

```json
{
  "v": 1, "id": "msg_065", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "buzzai.confirm",
  "params": {"conversation_id": "conv_001", "action_id": "act_001", "confirmed": true},
  "rid": "req_045"
}
```

### buzzai.tools.list

List available AI tools.

```json
{
  "v": 1, "id": "msg_066", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "buzzai.tools.list",
  "params": {},
  "rid": "req_046"
}
```

---

## Systemd Methods

### systemd.list

List systemd units.

```json
{
  "v": 1, "id": "msg_067", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "systemd.list",
  "params": {"type": "service", "state": "running"},
  "rid": "req_047"
}
```

### systemd.status

Get unit status.

```json
{
  "v": 1, "id": "msg_068", "ts": "2026-07-07T12:00:05Z",
  "type": "request", "method": "systemd.status",
  "params": {"unit": "buzzpi-runtime.service"},
  "rid": "req_048"
}
```

### systemd.start / systemd.stop / systemd.restart

Manage a systemd unit.

```json
{
  "v": 1, "id": "msg_069", "ts": "2026-07-07T12:00:10Z",
  "type": "request", "method": "systemd.restart",
  "params": {"unit": "nginx.service"},
  "rid": "req_049"
}
```

### systemd.logs

Get journal logs for a unit.

```json
{
  "v": 1, "id": "msg_070", "ts": "2026-07-07T12:00:15Z",
  "type": "request", "method": "systemd.logs",
  "params": {"unit": "buzzpi-runtime.service", "tail": 100},
  "rid": "req_050"
}
```

---

# Relay Signaling Methods

The relay server facilitates WebRTC negotiation between clients and devices over WebSocket.

### relay.connect

Authenticate and open a signaling session.

```json
{
  "v": 1, "id": "msg_071", "ts": "2026-07-07T12:00:00Z",
  "type": "request", "method": "relay.connect",
  "params": {
    "auth_token": "jwt_access_token",
    "role": "client",
    "device_id": "dev_abc123"
  },
  "rid": "req_051"
}
```

### relay.sdp_offer / relay.sdp_answer

Exchange WebRTC SDP.

```json
{
  "v": 1, "id": "msg_072", "ts": "2026-07-07T12:00:01Z",
  "type": "request", "method": "relay.sdp_offer",
  "params": {
    "target_id": "dev_abc123",
    "sdp": "v=0\no=...\n..."
  },
  "rid": "req_052"
}
```

### relay.ice_candidate

Exchange ICE candidates.

```json
{
  "v": 1, "id": "msg_073", "ts": "2026-07-07T12:00:02Z",
  "type": "request", "method": "relay.ice_candidate",
  "params": {
    "target_id": "dev_abc123",
    "candidate": "candidate:1 1 UDP 2122252543 192.168.1.42 54321 typ host"
  },
  "rid": "req_053"
}
```

### relay.heartbeat

Keep-alive ping (every 30 seconds).

```json
{
  "v": 1, "id": "msg_074", "ts": "2026-07-07T12:00:30Z",
  "type": "request", "method": "relay.heartbeat",
  "params": {},
  "rid": "req_054"
}
```

### relay.disconnect

End signaling session.

```json
{
  "v": 1, "id": "msg_075", "ts": "2026-07-07T12:05:00Z",
  "type": "request", "method": "relay.disconnect",
  "params": {},
  "rid": "req_055"
}
```

---

# REST API (Registry Service)

Base URL: `https://api.buzzpi.dev/v1`

### Auth Endpoints

```http
POST /v1/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password",
  "display_name": "User Name"
}
```

```http
201 Created
{
  "user_id": "usr_abc123",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rt_abc123def456",
  "expires_in": 900
}
```

```http
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password"
}
```

```http
POST /v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "rt_abc123def456"
}
```

### Device Endpoints

```http
GET /v1/devices
Authorization: Bearer <access_token>
```

```http
200 OK
{
  "devices": [
    {
      "id": "dev_abc123",
      "friendly_name": "Kitchen Pi",
      "model": "Raspberry Pi 5",
      "state": "online",
      "last_seen_at": "2026-07-07T11:59:00Z",
      "runtime_version": "0.1.0"
    }
  ]
}
```

```http
POST /v1/devices
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "claim_token": "claim_abc123",
  "identity_key": "ssh-ed25519 AAAAC3...",
  "friendly_name": "Kitchen Pi",
  "model": "Raspberry Pi 5",
  "runtime_version": "0.1.0"
}
```

```http
GET /v1/devices/:id
DELETE /v1/devices/:id
PATCH /v1/devices/:id

PATCH /v1/devices/:id
Content-Type: application/json

{
  "friendly_name": "Living Room Pi"
}
```

### Pairing Endpoints

```http
POST /v1/devices/:id/pair
Content-Type: application/json

{
  "pairing_code": "ABC123"
}
```

```http
POST /v1/devices/:id/unpair
```

### Discovery Endpoint

```http
GET /v1/discover
Authorization: Bearer <access_token>
```

```http
200 OK
{
  "devices": [
    {
      "id": "dev_abc123",
      "friendly_name": "Kitchen Pi",
      "state": "online",       // "online" | "offline" | "unknown"
      "connection_type": "direct" | "relay" | null,
      "last_seen_at": "..."
    }
  ]
}
```

### Error Responses

```http
400 Bad Request
{
  "error": "validation_error",
  "message": "email is required",
  "fields": {"email": "required"}
}
```

```http
401 Unauthorized
{
  "error": "unauthorized",
  "message": "Invalid or expired token"
}
```

```http
429 Too Many Requests
{
  "error": "rate_limited",
  "message": "Too many requests. Try again in 60 seconds.",
  "retry_after_seconds": 60
}
```

---

# IPC Messages (Runtime ↔ Plugin)

All messages use 4-byte length prefix + JSON payload. See [Plugin System](plugin-system.md) for the full IPC specification.

### Plugin → Runtime

| Message | Purpose | Payload |
|---------|---------|---------|
| `ipc.hello` | Handshake | `{plugin_id, version}` |
| `ipc.health` | Health pong | `{status: "ok", ...stats}` |
| `ipc.capabilities` | Declare capabilities | `{capabilities: [...]}` |
| `ipc.response` | Return request result | `{rid, result}` |
| `ipc.event` | Push unsolicited event | `{event_type, data}` |
| `ipc.log` | Log message | `{level, message, ...context}` |

### Runtime → Plugin

| Message | Purpose | Payload |
|---------|---------|---------|
| `ipc.hello.ack` | Handshake response | `{runtime_version, session_id}` |
| `ipc.health` | Health ping | `{}` |
| `ipc.request` | Forward BPP request | `{method, params, rid}` |
| `ipc.shutdown` | Graceful termination | `{reason, grace_period_seconds}` |

---

# WebRTC Data Channels

Each service gets its own labeled data channel within the WebRTC peer connection.

| Channel Label | Direction | Reliability | Payload | Purpose |
|---------------|-----------|-------------|---------|---------|
| `control` | Bidirectional | Ordered | JSON | Service control messages |
| `terminal` | Bidirectional | Unordered | Base64 | Terminal I/O (high throughput) |
| `files` | Bidirectional | Ordered | Binary | File transfer (chunked) |
| `screen` | Device → Client | Unordered | H.264 NAL | Video frames |
| `gpio` | Bidirectional | Ordered | JSON | GPIO control events |
| `camera` | Device → Client | Unordered | H.264 NAL | Camera stream |
| `logs` | Device → Client | Unordered | JSON | Log stream |

---

# Error Codes

See [Error Codes](../reference/error-codes.md) in the Reference Book for the complete error code registry.

| Range | Category | Example |
|-------|----------|---------|
| 1xxx | Transport | `1001 CONNECTION_TIMEOUT` |
| 2xxx | Authentication | `2001 PAIRING_REJECTED` |
| 3xxx | Device | `3001 DEVICE_NOT_FOUND` |
| 4xxx | Service | `4001 TERMINAL_UNAVAILABLE` |
| 5xxx | Protocol | `5001 METHOD_NOT_FOUND` |
| 6xxx | Extension | `6001 PLUGIN_CRASHED` |

---

# Rate Limits

| Resource | Limit | Scope |
|----------|-------|-------|
| BPP method calls | 100/second | Per connection |
| Terminal output | 1 MB/s | Per terminal session |
| File transfer | 10 MB/s | Per file transfer |
| REST API (unauthenticated) | 10/minute | Per IP |
| REST API (authenticated) | 100/minute | Per user |
| REST API (device registration) | 10/hour | Per IP |
| Relay connections | 3 concurrent | Per user |
| Relay reconnections | 10/minute | Per session |

---

# Change History

| Date | Change |
|------|--------|
| 2026-07-07 | Initial specification — all BPP methods, REST endpoints, IPC messages defined |
