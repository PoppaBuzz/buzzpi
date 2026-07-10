# BPP Chapter 20: Capability Discovery

**Layer:** Capabilities  
**Status:** Draft  
**Version:** 1.0.0

Capability discovery allows clients and devices to determine what features, services, and extensions the other party supports.

## Overview

Every BPP connection begins with capability discovery. Both sides announce what they support, and the connection is configured accordingly.

## Capability Exchange

Capability exchange happens immediately after authentication, before any service messages are exchanged:

```
Client                                Device
  │                                     │
  │  session.capabilities               │
  │ ───────────────────────────────────►│
  │                                     │
  │  session.capabilities               │
  │ ◀───────────────────────────────────│
  │                                     │
  │  session.capabilities.ack           │
  │ ────────────────────────────────────│
```

### Client Capabilities

```json
{
  "type": "request",
  "method": "session.capabilities",
  "params": {
    "bpp_version": "1.0.0",
    "client_type": "android",
    "client_version": "0.1.0",
    "features": {
      "terminal": { "version": 1, "supports": ["ansi_256", "true_color"] },
      "screen": { "version": 1, "max_width": 1920, "max_height": 1080 },
      "files": { "version": 1 },
      "docker": { "version": 1 },
      "gpio": null
    },
    "compression": ["none", "gzip"],
    "input_methods": ["touch", "keyboard", "mouse"],
    "display_info": {
      "width_px": 1440,
      "height_px": 3120,
      "density": 2.75,
      "color_depth": 32
    }
  }
}
```

### Device Capabilities

```json
{
  "type": "response",
  "method": "session.capabilities",
  "result": {
    "bpp_version": "1.0.0",
    "device_type": "raspberry_pi_5",
    "runtime_version": "0.1.0",
    "features": {
      "terminal": { "version": 1, "supports": ["ansi_256", "true_color"] },
      "screen": { "version": 1, "max_width": 1920, "max_height": 1080 },
      "files": { "version": 1 },
      "docker": { "version": 1, "available": true },
      "gpio": { "version": 1, "available": true },
      "camera": { "version": 1, "available": false, "reason": "No camera detected" }
    },
    "compression": ["none", "gzip", "zstd"],
    "capture_methods": ["drm", "fbdev"],
    "services": {
      "available": [
        {
          "name": "buzzpi-runtime.service",
          "state": "running",
          "version": "0.1.0"
        }
      ]
    }
  }
}
```

## Feature Flags

Features are represented as a map where:

- `null` or absent: The feature is not supported
- A non-null object: The feature is supported, with version and options

```json
{
  "terminal": { "version": 1, "supports": ["ansi_256"] },
  "gpio": null
}
```

### Versioning

Each feature indicates its maximum supported version:

```json
{
  "terminal": { "version": 2 }
}
```

The negotiated version is `min(client_version, device_version)`. If the client supports v2 but the device supports v1, v1 is used.

## Profiles

Devices may advertise predefined capability profiles:

```json
{
  "profiles": {
    "default": {
      "terminal": { "version": 1 },
      "screen": { "version": 1 },
      "files": { "version": 1 }
    },
    "headless": {
      "terminal": { "version": 1 },
      "files": { "version": 1 },
      "screen": null
    },
    "maker": {
      "terminal": { "version": 1 },
      "screen": { "version": 1 },
      "gpio": { "version": 1 },
      "docker": { "version": 1 }
    }
  }
}
```

The client selects a profile during the capabilities negotiation:

```json
{
  "method": "session.capabilities",
  "params": {
    "profile": "maker",
    "features": { ... }
  }
}
```

## Metadata

Capability discovery also exchanges device metadata:

| Field | Purpose | Example |
|-------|---------|---------|
| `device_type` | Device model | `raspberry_pi_5` |
| `runtime_version` | Runtime version | `0.1.0` |
| `os` | Operating system | `Raspberry Pi OS 12 (bookworm)` |
| `kernel` | Kernel version | `6.6.31` |
| `uptime` | Device uptime | `86400` seconds |
| `display_info` | Client display info | Width, height, density |

## Feature Negotiation

After capability exchange, the connection operates at the intersection of both sides' capabilities:

1. If both sides support a feature → the feature is available
2. If only one side supports a feature → the feature is unavailable
3. If versions differ → the minimum supported version is used
4. If options differ → the intersection is used (e.g., both support gzip, so gzip is used)

Clients SHOULD adapt their UI based on available features:
- If `gpio` is null → hide GPIO controls
- If `docker` is unavailable → prefer service-based controls
- If `screen` is null → hide screen tab, promote terminal

## Dynamic Discovery

Capabilities can change during a session:

- An extension is installed → device sends `session.capabilities.updated`
- A camera is plugged in → device sends `session.capabilities.updated`
- A service fails → device sends `session.capabilities.updated`

```json
{
  "type": "event",
  "method": "session.capabilities.updated",
  "params": {
    "features": {
      "camera": { "version": 1, "available": true }
    }
  }
}
```

The client updates its UI accordingly.
