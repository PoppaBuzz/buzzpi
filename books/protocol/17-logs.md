# BPP Chapter 17: Log Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Log service provides access to system logs (journald), application logs, and the BuzzPi Runtime's own logs.

## Overview

This service unifies log access from multiple sources:
- systemd journal (all system logs)
- Runtime log file
- Extension log files
- Custom log files (specified by path)

## Methods

### logs.query

Query logs from any available source.

**Request:**
```json
{
  "method": "logs.query",
  "params": {
    "source": "journal",
    "filter": {
      "priority": "info",
      "unit": "buzzpi-runtime.service",
      "since": "2026-07-07T10:00:00Z",
      "until": "2026-07-07T12:00:00Z",
      "keyword": "connection",
      "limit": 100,
      "cursor": null
    }
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `source` | `journal` | `journal`, `runtime`, `extension`, `file` |
| `filter.priority` | `info` | Minimum priority level |
| `filter.unit` | null | systemd unit name filter |
| `filter.since` | null | Start timestamp |
| `filter.until` | null | End timestamp |
| `filter.keyword` | null | Text search within logs |
| `filter.limit` | 100 | Maximum entries to return |
| `filter.cursor` | null | Resume from this cursor |

**Response:**
```json
{
  "method": "logs.query",
  "result": {
    "entries": [
      {
        "timestamp": "2026-07-07T11:59:00Z",
        "source": "journal",
        "unit": "buzzpi-runtime.service",
        "priority": "info",
        "message": "Connection established with relay server",
        "cursor": "s=abc...i=001",
        "metadata": {
          "pid": 1234,
          "boot_id": "abc-def-123"
        }
      }
    ],
    "total": 1,
    "truncated": false
  }
}
```

### logs.stream

Stream new log entries in real time.

**Request:**
```json
{
  "method": "logs.stream",
  "params": {
    "source": "journal",
    "filter": {
      "priority": "warning",
      "unit": "buzzpi-runtime.service"
    }
  }
}
```

**Events:**
```json
{
  "type": "event",
  "method": "logs.stream.entry",
  "params": {
    "timestamp": "2026-07-07T12:00:05Z",
    "source": "journal",
    "unit": "buzzpi-runtime.service",
    "priority": "warning",
    "message": "Temperature reading high: 72°C",
    "cursor": "s=abc...i=002"
  }
}
```

Unsubscribe by closing the event stream.

### logs.files

List available log files (for `file` source).

**Request:**
```json
{
  "method": "logs.files",
  "params": {}
}
```

**Response:**
```json
{
  "method": "logs.files",
  "result": {
    "files": [
      {
        "path": "/var/log/syslog",
        "size_bytes": 1048576,
        "modified": "2026-07-07T12:00:00Z",
        "format": "text"
      },
      {
        "path": "/var/log/buzzpi/runtime.log",
        "size_bytes": 524288,
        "modified": "2026-07-07T12:00:00Z",
        "format": "text"
      }
    ]
  }
}
```

### logs.clear

Clear logs (for configured log sources).

**Request:**
```json
{
  "method": "logs.clear",
  "params": {
    "source": "runtime",
    "confirm": true
  }
}
```

## Priority Levels

| Level | Numeric | Description |
|-------|---------|-------------|
| `emerg` | 0 | System is unusable |
| `alert` | 1 | Action must be taken immediately |
| `crit` | 2 | Critical conditions |
| `err` | 3 | Error conditions |
| `warning` | 4 | Warning conditions |
| `notice` | 5 | Normal but significant condition |
| `info` | 6 | Informational |
| `debug` | 7 | Debug-level messages |

## Security

| Concern | Mitigation |
|---------|------------|
| Sensitive data in logs | Logs may contain passwords, tokens, keys; access is gated by authentication |
| Log file traversal | `file` source validates paths against allowed list |
| Log flooding | Rate-limited at 1000 entries per query, 100 entries per second when streaming |
| Journal access | Requires `systemd-journal` group membership or root |

Log access respects the device's configured log retention policies. No logs are transmitted to the Relay Server.
