# BPP Chapter 16: System Services

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The System Services service provides management of systemd units (services, timers, sockets) on the device.

## Overview

This service communicates with systemd via D-Bus or the `systemctl` command. It allows clients to view, start, stop, restart, and enable/disable system services.

## Methods

### services.list

List systemd units.

**Request:**
```json
{
  "method": "services.list",
  "params": {
    "type": "service",
    "state": "active",
    "user": false
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `type` | all | `service`, `timer`, `socket`, `target`, `all` |
| `state` | all | `active`, `inactive`, `failed`, `all` |
| `user` | false | List user services (systemd user instance) |

**Response:**
```json
{
  "method": "services.list",
  "result": {
    "units": [
      {
        "name": "buzzpi-runtime.service",
        "description": "BuzzPi Runtime",
        "state": "active",
        "sub_state": "running",
        "enabled": true,
        "load_state": "loaded",
        "pid": 1234,
        "uptime_seconds": 86400,
        "memory_bytes": 31457280,
        "cpu_usage_percent": 0.5,
        "main_pid": 1234
      },
      {
        "name": "ssh.service",
        "description": "OpenSSH server",
        "state": "active",
        "sub_state": "running",
        "enabled": true,
        "load_state": "loaded",
        "pid": 567,
        "uptime_seconds": 86400,
        "memory_bytes": 5242880,
        "cpu_usage_percent": 0.1,
        "main_pid": 567
      }
    ],
    "total": 2
  }
}
```

### services.status

Get detailed status of a specific unit.

**Request:**
```json
{
  "method": "services.status",
  "params": {
    "name": "buzzpi-runtime.service"
  }
}
```

**Response:**
```json
{
  "method": "services.status",
  "result": {
    "name": "buzzpi-runtime.service",
    "description": "BuzzPi Runtime",
    "state": "active",
    "sub_state": "running",
    "enabled": true,
    "load_state": "loaded",
    "pid": 1234,
    "uptime_seconds": 86400,
    "memory_bytes": 31457280,
    "cpu_usage_percent": 0.5,
    "main_pid": 1234,
    "journal_cursor": "s=abc123..."
  }
}
```

### services.start / stop / restart / reload

**Request:**
```json
{
  "method": "services.restart",
  "params": {
    "name": "buzzpi-runtime.service"
  }
}
```

### services.enable / disable

Control whether a service starts at boot.

**Request:**
```json
{
  "method": "services.enable",
  "params": {
    "name": "buzzpi-runtime.service"
  }
}
```

### services.logs

Get journald logs for a unit.

**Request:**
```json
{
  "method": "services.logs",
  "params": {
    "name": "buzzpi-runtime.service",
    "tail": 50,
    "priority": "info",
    "since": "2026-07-07T10:00:00Z",
    "follow": false,
    "cursor": null
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `tail` | all | Number of recent entries |
| `priority` | `info` | Minimum priority (`emerg`, `alert`, `crit`, `err`, `warning`, `notice`, `info`, `debug`) |
| `since` | null | Return entries after this timestamp |
| `follow` | false | Stream new entries |
| `cursor` | null | Return entries after this journal cursor |

**Response:**
```json
{
  "method": "services.logs",
  "result": {
    "name": "buzzpi-runtime.service",
    "entries": [
      {
        "timestamp": "2026-07-07T11:59:00Z",
        "message": "Connection established with relay server",
        "priority": "info",
        "pid": 1234,
        "cursor": "s=abc...i=001"
      }
    ]
  }
}
```

When `follow` is true, log entries are streamed as events.

## Security

| Service | Allowed Operations |
|---------|-------------------|
| `buzzpi-runtime.service` | status, logs, restart |
| `ssh.service` | status, logs, restart, enable, disable |
| `docker.service` | status, logs, restart |
| `bluetooth.service` | status, logs, start, stop, enable, disable |
| `cron.service` | status, logs, start, stop |
| All others | status, logs only (by default) |

The device MAY extend allowed operations via extension configuration.

## Error Codes

| Code | Meaning |
|------|---------|
| SERVICE_NOT_FOUND | Unit does not exist |
| SERVICE_NOT_ACCESSIBLE | User does not have permission |
| SERVICE_OPERATION_FAILED | systemctl command failed |
| SERVICE_NOT_INSTALLED | systemd is not available |
