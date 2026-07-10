# BPP Chapter 19: Custom RPC

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Custom RPC service allows extensions and clients to define and call arbitrary methods on the device. This is the extensibility mechanism for BPP — any operation not covered by the built-in services can be added via RPC.

## Overview

Custom RPC enables:
- Extension-defined methods (e.g., a Docker extension adds `docker.containers.list`)
- Client-to-device method calls with request/response semantics
- Device-initiated method calls (device sends a request to the client)
- Streaming methods (subscription-based event streams)

## Method Registration

Extensions register custom RPC methods during initialization by sending an extension manifest:

```json
{
  "type": "extension.register",
  "params": {
    "extension_id": "docker-manager",
    "version": "1.0.0",
    "methods": [
      {
        "name": "docker.containers.list",
        "type": "request_response",
        "params_schema": {
          "type": "object",
          "properties": {
            "all": { "type": "boolean" }
          }
        },
        "result_schema": {
          "type": "object",
          "properties": {
            "containers": { "type": "array" }
          }
        },
        "permissions": ["docker:read"],
        "rate_limit": 30
      },
      {
        "name": "docker.containers.logs",
        "type": "stream",
        "params_schema": { ... },
        "permissions": ["docker:read"]
      }
    ]
  }
}
```

| Field | Description |
|-------|-------------|
| `name` | Fully qualified method name (dot-separated, extension-prefixed) |
| `type` | `request_response` or `stream` |
| `params_schema` | JSON Schema for the request parameters |
| `result_schema` | JSON Schema for the response (request_response only) |
| `permissions` | Required permissions to call this method |
| `rate_limit` | Max calls per second (default: 10) |

## Method Invocation

Custom RPC methods are called using the standard BPP message envelope:

```json
{
  "v": 1,
  "id": "msg_abc123",
  "ts": "2026-07-07T12:00:00Z",
  "type": "request",
  "method": "docker.containers.list",
  "params": {
    "all": true
  },
  "rid": "req_xyz789"
}
```

The Runtime routes the method to the registered extension and returns the response:

```json
{
  "v": 1,
  "id": "msg_def456",
  "ts": "2026-07-07T12:00:00.001Z",
  "type": "response",
  "rid": "req_xyz789",
  "result": {
    "containers": [...]
  }
}
```

## Streaming Methods

Streaming methods return results as a series of events:

```json
// Request
{
  "method": "docker.containers.logs",
  "params": { "id": "abc123", "follow": true },
  "rid": "req_stream_001"
}

// First event
{
  "type": "event",
  "method": "docker.containers.logs",
  "params": { "timestamp": "...", "line": "..." },
  "rid": "req_stream_001"
}

// ... more events ...

// Stream complete
{
  "type": "response",
  "method": "docker.containers.logs",
  "rid": "req_stream_001",
  "result": { "status": "completed" }
}
```

The client closes the stream by sending an event stream close message:
```json
{
  "type": "request",
  "method": "__stream.close__",
  "params": { "rid": "req_stream_001" }
}
```

## Device-Initiated RPC

An extension can initiate a method call to the client:

```json
{
  "v": 1,
  "id": "msg_ghi789",
  "ts": "2026-07-07T12:00:00Z",
  "type": "request",
  "method": "extension.alert.show",
  "params": {
    "title": "Motion Detected",
    "message": "Camera detected movement at 12:00"
  },
  "rid": "req_ext_001"
}
```

Client responds:
```json
{
  "v": 1,
  "id": "msg_jkl012",
  "ts": "2026-07-07T12:00:00.1Z",
  "type": "response",
  "rid": "req_ext_001",
  "result": { "dismissed": false }
}
```

## Method Discovery

Clients discover available RPC methods during capability negotiation:

```json
{
  "type": "session.capabilities",
  "params": {
    "rpc_methods": [
      {
        "name": "docker.containers.list",
        "type": "request_response",
        "description": "List Docker containers",
        "extension": "docker-manager",
        "extension_version": "1.0.0"
      }
    ]
  }
}
```

## Error Handling

Custom RPC methods use the standard BPP error format:

```json
{
  "type": "error",
  "rid": "req_xyz789",
  "error": {
    "code": "METHOD_NOT_FOUND",
    "message": "docker.containers.nonexistent is not registered"
  }
}
```

Standard RPC error codes:

| Code | Meaning |
|------|---------|
| `METHOD_NOT_FOUND` | Method is not registered by any extension |
| `METHOD_UNAVAILABLE` | Extension is installed but not running |
| `INVALID_PARAMS` | Parameters failed JSON Schema validation |
| `PERMISSION_DENIED` | Client does not have the required permission |
| `RATE_LIMITED` | Client exceeded the rate limit for this method |
| `EXTENSION_ERROR` | Extension returned an error (details in message) |
| `STREAM_CLOSED` | The stream was closed by the extension |

## Security

| Concern | Mitigation |
|---------|------------|
| Arbitrary method registration | Extensions are sandboxed; method names must use the extension prefix |
| Permission escalation | Each method declares required permissions; client capabilities are verified on every call |
| Extension spoofing | Extensions are verified during install; method-to-extension binding is immutable |
| Resource exhaustion | Rate limiting per method, per extension, per client |
