# BPP Chapter 5: WebSocket Transport

**Layer:** Transport  
**Status:** Draft  
**Version:** 1.0.0

WebSocket is the primary transport for control-plane communication between client/device and the Relay Server. It is used for signaling, presence, and relayed data when a direct WebRTC connection is unavailable.

## Connection

### Endpoint

```
wss://jphat.net/buzzpi/relay/ws
```

All WebSocket connections MUST use WSS (TLS over WebSocket). Plain WS is never used in production.

### Authentication

The WebSocket connection authenticates using one of:

1. **Session token** (client): `wss://jphat.net/buzzpi/relay/ws?token=<session_token>`
2. **Device token** (Runtime): `wss://jphat.net/buzzpi/relay/ws?device_id=<device_id>&token=<device_token>`
3. **API token** (CLI): `wss://jphat.net/buzzpi/relay/ws?token=<api_token>`

Authentication happens immediately after the WebSocket handshake. The server validates the token and, if invalid, closes the connection with code 4001.

### TLS Requirements

| Requirement | Specification |
|-------------|---------------|
| Minimum TLS version | 1.2 |
| Recommended TLS version | 1.3 |
| Certificate validation | Required (both sides) |
| Certificate pinning | Optional (BuzzPi clients MAY pin the relay certificate) |
| Cipher suites | TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 or stronger |

## Message Framing

All WebSocket messages use JSON text frames (opcode 0x1). Binary frames (opcode 0x2) are NOT used over WebSocket — binary data goes through WebRTC data channels.

### Message Envelope

Every message follows the standard BPP envelope (see API Design):

```json
{
  "v": 1,
  "id": "msg_abc123",
  "ts": "2026-07-07T12:00:00Z",
  "type": "request",
  "method": "relay.device.list",
  "params": {}
}
```

### Batch Messages

Multiple messages can be batched in a single WebSocket frame using newline-delimited JSON (NDJSON):

```json
{"v":1,"id":"msg_001","ts":"...","type":"request","method":"device.terminal.write","params":{"data":"ls\n"}}
{"v":1,"id":"msg_002","ts":"...","type":"request","method":"device.terminal.write","params":{"data":"pwd\n"}}
```

The server processes each message independently and sends separate responses. Batching is for throughput optimization only — there is no transactional guarantee across batched messages.

## Connection Lifecycle

### Open

1. Client/device connects to `wss://jphat.net/buzzpi/relay/ws` with authentication token
2. Server validates token
3. Server sends `relay.connected` message:
   ```json
   {
     "v": 1,
     "id": "msg_welcome",
     "ts": "...",
     "type": "event",
     "method": "relay.connected",
     "params": {
       "session_id": "ses_abc123",
       "server_version": "1.0.0",
       "server_time": "2026-07-07T12:00:00Z"
     }
   }
   ```
4. Client/device is now connected and can send/receive messages

### Heartbeat

While connected, both sides send WebSocket ping/pong frames:

- Client sends `ping` every 30 seconds
- Server responds with `pong`
- If no `pong` received within 10 seconds, client considers the connection dead
- Server closes connection if no `ping` received within 60 seconds

Application-level heartbeat is also supported (see Chapter 9).

### Close

The connection can be closed by either side:

| Close Code | Reason | Initiator |
|------------|--------|-----------|
| 1000 | Normal closure | Either |
| 4001 | Authentication failed | Server |
| 4002 | Token expired | Server |
| 4003 | Rate limited | Server |
| 4004 | Server shutting down | Server |
| 4005 | Client going offline | Client/Device |
| 4006 | Protocol version mismatch | Server |

### Reconnection

When the WebSocket connection drops:

1. Client immediately attempts reconnection
2. Uses exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (cap)
3. Backoff resets after 2 minutes of stable connection
4. Jitter: ±500ms on each attempt (prevents thundering herd)
5. On reconnection, client receives fresh `relay.connected` message
6. Any in-flight requests are retried (idempotent methods only)

## Rate Limiting

| Direction | Limit | Burst | Response |
|-----------|-------|-------|----------|
| Client → Relay | 100 msg/s | 200 | 4003 close if exceeded |
| Device → Relay | 10 msg/s | 50 | 4003 close if exceeded |
| Relay → Client | 100 msg/s | — | Backpressure applied |

## Message Size Limits

| Message Type | Max Size |
|-------------|----------|
| Control messages (JSON) | 64 KB |
| Batch payload (aggregate) | 512 KB |
| Terminal output (single frame) | 16 KB |

Messages exceeding these limits are rejected with error code `MESSAGE_TOO_LARGE`.

## Reliability

WebSocket transport provides:
- **In-order delivery** (TCP guarantees ordering within the stream)
- **At-most-once delivery** (no retry for individual messages at the transport layer)
- **Connection-level reliability** (reconnection retries the entire connection)

Application-level retry for important messages is handled by the caller (see Services Layer).
