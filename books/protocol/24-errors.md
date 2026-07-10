# BPP Chapter 24: Error Handling

**Layer:** Capabilities  
**Status:** Draft  
**Version:** 1.0.0

Standardized error handling across all BPP services and layers.

## Error Envelope

All errors follow the standard BPP error format:

```json
{
  "v": 1,
  "id": "msg_err_001",
  "ts": "2026-07-07T12:00:00Z",
  "type": "error",
  "rid": "req_xyz789",
  "error": {
    "code": "METHOD_NOT_FOUND",
    "message": "docker.containers.nonexistent is not registered",
    "data": {
      "method": "docker.containers.nonexistent",
      "available_methods": ["docker.containers.list", "docker.containers.logs"]
    }
  }
}
```

## Standard Error Codes

### Transport Errors (Code 1xxx)

| Code | Message | Description |
|------|---------|-------------|
| 1000 | CONNECTION_TIMEOUT | Connection attempt timed out |
| 1001 | CONNECTION_REFUSED | Device rejected the connection |
| 1002 | CONNECTION_LOST | Established connection was lost |
| 1003 | CONNECTION_RATE_LIMITED | Too many connection attempts |
| 1004 | TRANSPORT_UNAVAILABLE | No transport available (P2P and relay both failed) |
| 1005 | RELAY_DISCONNECTED | Relay WebSocket was disconnected |

### Identity Errors (Code 2xxx)

| Code | Message | Description |
|------|---------|-------------|
| 2000 | AUTHENTICATION_FAILED | Challenge-response verification failed |
| 2001 | AUTHENTICATION_EXPIRED | Session has expired, re-authentication needed |
| 2002 | TOKEN_INVALID | Session token is invalid or revoked |
| 2003 | TOKEN_EXPIRED | Session token has expired |
| 2004 | PAIRING_CODE_INVALID | Pairing code is incorrect |
| 2005 | PAIRING_CODE_EXPIRED | Pairing code has expired (5-minute limit) |
| 2006 | PAIRING_ALREADY_PAIRED | Device is already paired with another account |
| 2007 | DEVICE_NOT_FOUND | Device ID is not registered with this account |
| 2008 | KEY_MISMATCH | Identity key does not match stored key |

### Protocol Errors (Code 3xxx)

| Code | Message | Description |
|------|---------|-------------|
| 3000 | INVALID_MESSAGE | Message envelope is malformed |
| 3001 | INVALID_JSON | Message body is not valid JSON (WebSocket) |
| 3002 | VERSION_MISMATCH | BPP versions are incompatible |
| 3003 | METHOD_NOT_FOUND | Requested method is not registered |
| 3004 | METHOD_NOT_AVAILABLE | Method exists but is not available (device offline, feature disabled) |
| 3005 | INVALID_PARAMS | Method parameters failed validation |
| 3006 | MESSAGE_TOO_LARGE | Message exceeds size limit |
| 3007 | UNSUPPORTED_COMPRESSION | Compression method is not supported |

### Service Errors (Code 4xxx)

| Code | Message | Description |
|------|---------|-------------|
| 4000 | SERVICE_ERROR | Generic service error (details in message) |
| 4001 | SERVICE_NOT_AVAILABLE | Service is not available on this device |
| 4002 | SERVICE_NOT_INSTALLED | Extension providing this service is not installed |
| 4003 | SERVICE_BUSY | Service is busy handling another request |
| 4004 | SERVICE_TIMEOUT | Service did not respond within the expected time |
| 4005 | SESSION_LIMIT_EXCEEDED | Too many concurrent sessions (terminal, screen) |
| 4006 | RATE_LIMITED | Too many requests to this method |

### Resource Errors (Code 5xxx)

| Code | Message | Description |
|------|---------|-------------|
| 5000 | NOT_FOUND | Requested resource (file, container, service) was not found |
| 5001 | ALREADY_EXISTS | Resource already exists (file, container name) |
| 5002 | PERMISSION_DENIED | Client does not have the required permission |
| 5003 | ACCESS_DENIED | Device denied access (path restriction, file permissions) |
| 5004 | RESOURCE_EXHAUSTED | Device resource limit reached (disk, memory, CPU) |
| 5005 | STORAGE_FULL | Device has no available storage space |

### Extension Errors (Code 6xxx)

| Code | Message | Description |
|------|---------|-------------|
| 6000 | EXTENSION_ERROR | Generic extension error |
| 6001 | EXTENSION_NOT_INSTALLED | Extension is not installed on the device |
| 6002 | EXTENSION_NOT_RUNNING | Extension is installed but not running |
| 6003 | EXTENSION_CRASHED | Extension process crashed |
| 6004 | EXTENSION_TIMEOUT | Extension did not respond within the expected time |
| 6005 | EXTENSION_PERMISSION_DENIED | Extension does not have the required permission |
| 6006 | PLUGIN_INVALID | Plugin manifest or binary is invalid |
| 6007 | PLUGIN_SIGNATURE_INVALID | Plugin signature verification failed |
| 6008 | EXTENSION_UPDATE_AVAILABLE | Operation failed because an update is available |

## Error Responses by Category

### Timeout

```json
{
  "error": {
    "code": "SERVICE_TIMEOUT",
    "message": "The terminal service did not respond within 5 seconds",
    "data": {
      "timeout_seconds": 5,
      "service": "terminal",
      "method": "terminal.open",
      "retry_allowed": true
    }
  }
}
```

### Permission Denied

```json
{
  "error": {
    "code": "PERMISSION_DENIED",
    "message": "Client does not have docker:admin permission",
    "data": {
      "required_permission": "docker:admin",
      "current_permissions": ["docker:read"],
      "extension": "docker-manager",
      "upgrade_url": "buzzpi://settings/extensions/docker-manager/permissions"
    }
  }
}
```

### Device Offline

```json
{
  "error": {
    "code": "SERVICE_NOT_AVAILABLE",
    "message": "Kitchen Pi is offline. This action cannot be performed until the device reconnects.",
    "data": {
      "device_id": "018e0a3f-...",
      "device_name": "Kitchen Pi",
      "last_seen": "2026-07-07T11:55:00Z",
      "retry_on_reconnect": true
    }
  }
}
```

## Client Error Handling Guidelines

### UI Presentation

Errors presented to the user MUST follow the Error Philosophy (Experience Book):

1. Translate error codes to human-readable messages
2. Provide context (what was happening when the error occurred)
3. Suggest recovery action
4. Never display raw error codes

### Error Presentation Examples

| Error Code | User-Facing Message |
|------------|---------------------|
| CONNECTION_TIMEOUT | "Could not reach Kitchen Pi. It may be too far from the network or powered off. Check that it is plugged in and connected to Wi-Fi." |
| PERMISSION_DENIED | "BuzzPi doesn't have permission to manage Docker on this device. Grant the 'Docker Admin' permission in Settings." |
| STORAGE_FULL | "Kitchen Pi's storage is full. Free up space by deleting old files or logs before trying again." |
| PAIRING_CODE_INVALID | "The pairing code is incorrect. The device should display a new code — try again." |

### Retry Logic

| Error Type | Retry Behavior |
|------------|----------------|
| Connection timeout | Automatic retry (3 attempts, 5s interval) |
| Rate limited | Wait and retry (respect Retry-After header) |
| Permission denied | No retry — prompt user to grant permission |
| Device offline | No retry — wait for device to come online |
| Invalid params | No retry — fix the request |
| Method not found | No retry — update client or device |
| Version mismatch | No retry — update client or device |
