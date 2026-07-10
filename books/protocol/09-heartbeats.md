# BPP Chapter 9: Heartbeats

**Layer:** Transport  
**Status:** Draft  
**Version:** 1.0.0

Heartbeats maintain presence — they tell the system that a device or client is still online and reachable.

## WebSocket Heartbeat

WebSocket connections use standard ping/pong frames for connection health:

### Client/Device → Relay Server

| Parameter | Value |
|-----------|-------|
| Interval | 30 seconds |
| Timeout | 10 seconds (no pong received) |
| Max missed pongs before close | 2 consecutive |

After timeout, the client/device:
1. Closes the WebSocket connection
2. Begins reconnection with exponential backoff
3. On reconnection, re-registers presence

### Relay Server → Client/Device

The server also monitors the connection:
- If no ping received for 60 seconds → close connection
- If client reconnects within 30 seconds → treat as brief interruption (no notification)
- If client is gone for > 2 minutes → mark OFFLINE, send notification

## Device Presence Heartbeat

In addition to WebSocket-level pings, the device sends an application-level heartbeat to the Relay Server:

```json
{
  "v": 1,
  "id": "hb_001",
  "ts": "2026-07-07T12:00:00Z",
  "type": "request",
  "method": "relay.heartbeat",
  "params": {
    "uptime_seconds": 86400,
    "temperature_celsius": 45.2,
    "connection_type": "direct"
  }
}
```

| Parameter | Interval | Purpose |
|-----------|----------|---------|
| WebSocket ping | 30s | Connection health |
| Application heartbeat | 60s | Presence + basic device stats |

## Grace Period

When the Relay Server detects a device disconnection:

```
t+0s    WebSocket drops
t+0s    Device marked OFFLINE (internal), grace period starts
t+0s    Start grace period timer (2 minutes)
t+0-2m  If device reconnects → restore ONLINE, cancel timer
t+2m    If no reconnect → send OFFLINE notification to clients
t+2m    Device stays OFFLINE until reconnection
t+24h   If still OFFLINE → send "still offline" summary notification
```

The grace period prevents false alerts during:
- Network interruptions (WiFi flapping, router reboots)
- Device reboots (Runtime restarts within 2 minutes)
- Brief power outages (device restarts)

## Client Presence

Clients (app/CLI) send application-level heartbeats less frequently:

```json
{
  "v": 1,
  "id": "hb_002",
  "ts": "...",
  "type": "request",
  "method": "relay.heartbeat",
  "params": {
    "app_version": "0.1.0",
    "os": "Android 15"
  }
}
```

| Parameter | Interval | Purpose |
|-----------|----------|---------|
| Application heartbeat | 5 minutes | Client presence tracking |

Client disconnection does not trigger notifications (the user does not need to know their phone went offline).

## Dead Peer Detection (WebRTC)

For WebRTC data channels, ICE connectivity checks serve as heartbeats:

| Parameter | Value |
|-----------|-------|
| ICE consent freshness | Every 30 seconds |
| Consent timeout | 30 seconds (no response) |
| Max retries | 3 |

If consent check fails for 3 consecutive attempts:
1. Client considers the WebRTC connection dead
2. Client sends `relay.connect.request` to establish a new WebRTC connection
3. Old connection is torn down

## Summary

| Heartbeat Type | Interval | Timeout | Consequence of Failure |
|---------------|----------|---------|------------------------|
| WebSocket ping | 30s | 10s | WebSocket reconnect |
| Application (device) | 60s | 120s | Device marked OFFLINE |
| Application (client) | 5min | 15min | Client marked inactive |
| ICE consent | 30s | 90s | WebRTC re-negotiation |
