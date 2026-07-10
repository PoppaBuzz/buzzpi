# BPP Chapter 13: System Stats Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Stats service provides real-time and historical system metrics from the device.

## Overview

System stats are collected by the Runtime and served to clients on demand or via subscription. Stats cover CPU, memory, storage, temperature, network, and uptime.

## Methods

### stats.get

Get current system metrics.

**Request:**
```json
{
  "method": "stats.get",
  "params": {
    "categories": ["cpu", "memory", "storage", "temperature", "network", "uptime"]
  }
}
```

**Response:**
```json
{
  "method": "stats.get",
  "result": {
    "timestamp": "2026-07-07T12:00:00Z",
    "cpu": {
      "usage_percent": 23.5,
      "temperature_celsius": 45.2,
      "frequency_mhz": 1800,
      "cores": 4,
      "load_average": {
        "1m": 0.85,
        "5m": 0.42,
        "15m": 0.31
      }
    },
    "memory": {
      "total_mb": 8192,
      "used_mb": 3145,
      "available_mb": 5047,
      "swap_total_mb": 1024,
      "swap_used_mb": 128,
      "usage_percent": 38.4
    },
    "storage": {
      "total_mb": 30516,
      "used_mb": 12384,
      "available_mb": 18132,
      "usage_percent": 40.6,
      "mounts": [
        {
          "path": "/",
          "total_mb": 30516,
          "used_mb": 12384,
          "available_mb": 18132,
          "usage_percent": 40.6,
          "filesystem": "ext4"
        },
        {
          "path": "/boot",
          "total_mb": 512,
          "used_mb": 218,
          "available_mb": 294,
          "usage_percent": 42.6,
          "filesystem": "vfat"
        }
      ]
    },
    "temperature": {
      "cpu_celsius": 45.2,
      "gpu_celsius": 42.0,
      "throttled": false,
      "throttle_history": "0x0"
    },
    "network": {
      "interfaces": [
        {
          "name": "eth0",
          "ip_address": "192.168.1.42",
          "mac_address": "dc:a6:32:01:23:45",
          "rx_bytes": 1234567890,
          "tx_bytes": 987654321,
          "rx_packets": 1234567,
          "tx_packets": 987654,
          "link_speed_mbps": 1000,
          "is_up": true,
          "is_wireless": false,
          "signal_percent": null
        },
        {
          "name": "wlan0",
          "ip_address": "192.168.1.43",
          "mac_address": "dc:a6:32:67:89:ab",
          "rx_bytes": 123456,
          "tx_bytes": 98765,
          "rx_packets": 1234,
          "tx_packets": 987,
          "link_speed_mbps": 300,
          "is_up": true,
          "is_wireless": true,
          "signal_percent": 85,
          "ssid": "HomeNetwork"
        }
      ]
    },
    "uptime": {
      "uptime_seconds": 86400,
      "boot_time": "2026-07-06T12:00:00Z",
      "processes": 245,
      "users": ["pi"]
    }
  }
}
```

### stats.subscribe

Subscribe to real-time stat updates.

**Request:**
```json
{
  "method": "stats.subscribe",
  "params": {
    "categories": ["cpu", "temperature", "memory"],
    "interval_seconds": 5
  }
}
```

**Events (streamed every interval):**
```json
{
  "type": "event",
  "method": "stats.update",
  "params": {
    "timestamp": "2026-07-07T12:00:05Z",
    "cpu": { "usage_percent": 24.1, "temperature_celsius": 45.5 },
    "temperature": { "cpu_celsius": 45.5 },
    "memory": { "used_mb": 3150, "available_mb": 5042 }
  }
}
```

Unsubscribe by closing the event stream.

### stats.history

Get historical stat data.

**Request:**
```json
{
  "method": "stats.history",
  "params": {
    "category": "temperature",
    "metric": "cpu_celsius",
    "from": "2026-07-07T10:00:00Z",
    "to": "2026-07-07T12:00:00Z",
    "interval": "5m",
    "aggregation": "avg"
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `interval` | `1m` | `1m`, `5m`, `15m`, `1h`, `1d` |
| `aggregation` | `avg` | `avg`, `min`, `max`, `median`, `p95` |

**Response:**
```json
{
  "method": "stats.history",
  "result": {
    "category": "temperature",
    "metric": "cpu_celsius",
    "from": "2026-07-07T10:00:00Z",
    "to": "2026-07-07T12:00:00Z",
    "interval": "5m",
    "aggregation": "avg",
    "datapoints": [
      { "timestamp": "2026-07-07T10:00:00Z", "value": 44.8 },
      { "timestamp": "2026-07-07T10:05:00Z", "value": 45.1 },
      { "timestamp": "2026-07-07T10:10:00Z", "value": 46.2 }
    ]
  }
}
```

### stats.thresholds

Get or set alert thresholds for metrics.

**Get thresholds:**
```json
{
  "method": "stats.thresholds",
  "params": {}
}
```

**Set thresholds:**
```json
{
  "method": "stats.thresholds",
  "params": {
    "thresholds": [
      {
        "category": "temperature",
        "metric": "cpu_celsius",
        "warning": 70,
        "critical": 80,
        "enabled": true
      },
      {
        "category": "storage",
        "metric": "usage_percent",
        "warning": 80,
        "critical": 90,
        "enabled": true
      }
    ]
  }
}
```

**Response:**
```json
{
  "method": "stats.thresholds",
  "result": {
    "thresholds": [
      {
        "category": "temperature",
        "metric": "cpu_celsius",
        "warning": 70,
        "critical": 80,
        "enabled": true,
        "current_value": 45.2
      },
      {
        "category": "storage",
        "metric": "usage_percent",
        "warning": 80,
        "critical": 90,
        "enabled": true,
        "current_value": 40.6
      }
    ]
  }
}
```

When a threshold is crossed, the device sends an event:

```json
{
  "type": "event",
  "method": "stats.threshold.breached",
  "params": {
    "category": "temperature",
    "metric": "cpu_celsius",
    "value": 71.2,
    "threshold": 70,
    "level": "warning"
  }
}
```

## Collection Interval

| Stat | Collection Interval | Storage Duration |
|------|-------------------|------------------|
| CPU usage | 5 seconds | 24 hours (1m granularity) |
| Memory | 5 seconds | 24 hours (1m granularity) |
| Storage | 60 seconds | 7 days (5m granularity) |
| Temperature | 5 seconds | 24 hours (1m granularity) |
| Network I/O | 60 seconds | 7 days (5m granularity) |
| Uptime | On change | Indefinite |

## Data Storage

System stats are stored on the device:

| Property | Value |
|----------|-------|
| Storage location | `/var/lib/buzzpi/stats/` |
| Storage format | SQLite database per day |
| Max storage | 100MB (oldest files deleted first) |
| Retention | 7 days (configurable) |

Historical data is served from the device's local database. The Relay Server does not store historical stats.
