# BPP Chapter 21: Version Negotiation

**Layer:** Capabilities  
**Status:** Draft  
**Version:** 1.0.0

Version negotiation ensures that clients and devices with different BPP versions can interoperate.

## Overview

BPP is designed for backward compatibility within the same major version. Version negotiation determines the intersection of supported features and selects the appropriate protocol version for each layer.

## Version Structure

BPP versions follow semantic versioning: `MAJOR.MINOR.PATCH`

| Component | Description |
|-----------|-------------|
| MAJOR | Breaking protocol changes (incompatible) |
| MINOR | Backward-compatible additions |
| PATCH | Bug fixes, no feature changes |

Each service (terminal, screen, files, etc.) also has its own version:

```
bpp: 1.2.3
├── terminal: 1.0.0
├── screen: 2.1.0
├── files: 1.0.0
└── gpio: 1.5.0
```

## Negotiation Flow

### Step 1: Announce Supported Versions

Both sides announce their supported BPP version and per-service versions:

**Client:**
```json
{
  "method": "session.capabilities",
  "params": {
    "bpp_version": "1.0.0",
    "bpp_version_range": ">=1.0.0 <2.0.0",
    "services": {
      "terminal": { "version": 1, "version_range": ">=1.0.0 <3.0.0" },
      "screen": { "version": 1, "version_range": ">=1.0.0 <2.0.0" }
    }
  }
}
```

**Device:**
```json
{
  "method": "session.capabilities",
  "result": {
    "bpp_version": "1.2.0",
    "bpp_version_range": ">=1.0.0 <2.0.0",
    "services": {
      "terminal": { "version": 2, "version_range": ">=1.0.0 <3.0.0" },
      "screen": { "version": 1, "version_range": ">=1.0.0 <2.0.0" }
    }
  }
}
```

### Step 2: Resolve Versions

Both sides independently resolve the negotiated version:

```
BPP version:
  Client: 1.0.0 (supports >=1.0.0 <2.0.0)
  Device: 1.2.0 (supports >=1.0.0 <2.0.0)
  Negotiated: 1.0.0 (minimum of both within intersection)

Terminal service:
  Client: v1 (supports >=1.0.0 <3.0.0)
  Device: v2 (supports >=1.0.0 <3.0.0)
  Negotiated: v1 (minimum, client does not support v2)

Screen service:
  Client: v1 (supports >=1.0.0 <2.0.0)
  Device: v1 (supports >=1.0.0 <2.0.0)
  Negotiated: v1 (both support v1)
```

### Step 3: Confirm

```json
{
  "method": "session.capabilities.ack",
  "params": {
    "bpp_version": "1.0.0",
    "services": {
      "terminal": 1,
      "screen": 1
    }
  }
}
```

## Version Compatibility Rules

| Client Version | Device Version | Result |
|---------------|----------------|--------|
| 1.0.0 | 1.0.0 | Full compatibility |
| 1.0.0 | 1.5.0 | Compatible (device adds features but maintains backward compatibility) |
| 1.0.0 | 2.0.0 | Incompatible (major version mismatch) |
| 1.5.0 | 1.0.0 | Compatible (client must not use features added in 1.1-1.5) |
| 2.0.0 | 1.0.0 | Incompatible |

## Incompatible Versions

If the major versions do not overlap:

```json
{
  "type": "error",
  "method": "session.capabilities",
  "error": {
    "code": "VERSION_MISMATCH",
    "message": "Client BPP v1.0.0 is incompatible with device BPP v2.0.0",
    "data": {
      "client_version": "1.0.0",
      "device_version": "2.0.0",
      "supported_range": ">=2.0.0 <3.0.0",
      "update_available": true,
      "update_url": "https://buzzpi.dev/download/latest"
    }
  }
}
```

The client SHOULD:
1. Prompt the user to update the Runtime
2. Provide the update URL or trigger an update via the relay
3. Block connection until the version is updated

## Layer-Specific Versioning

Each BPP layer has its own version:

| Layer | Version | Notes |
|-------|---------|-------|
| BPP (overall) | 1.0.0 | Changed only when the envelope or core protocol changes |
| Identity | 1.0.0 | Changes when pairing/authentication mechanisms change |
| Transport | 1.0.0 | Changes when WebSocket/WebRTC/Relay protocols change |
| Services | Per-service | Each service versioned independently |
| Capabilities | 1.0.0 | Changes when discovery/negotiation mechanisms change |

Layer versions are nested within the capability exchange:

```json
{
  "bpp_version": "1.0.0",
  "layers": {
    "identity": { "version": 1 },
    "transport": { "version": 1 },
    "capabilities": { "version": 1 }
  },
  "services": {
    "terminal": { "version": 1 }
  }
}
```

## Future-Proofing

When designing extensions to the protocol:

1. **Never remove a field** — mark it as deprecated instead
2. **Never change a field's meaning** — add a new field with a new name
3. **Always send version ranges** — not just the current version
4. **Unknown fields are preserved** — forward them unmodified
5. **Unknown methods are rejected** — with `METHOD_NOT_FOUND`, not silently ignored
