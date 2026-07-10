# BPP Chapter 7: Relay Protocol

**Layer:** Transport  
**Status:** Draft  
**Version:** 1.0.0

The Relay Server is the cloud-side component that enables device discovery, signaling, and connectivity when direct P2P is not possible.

## Architecture

```
┌──────────┐    WebSocket     ┌──────────────┐    WebSocket     ┌──────────┐
│          │◀════════════════▶│              │◀════════════════▶│          │
│  Client  │    (Signaling)   │    Relay     │    (Signaling)   │  Device  │
│          │                  │    Server    │                  │ (Runtime)│
└──────────┘                  └──────────────┘                  └──────────┘
     │                             │                                │
     │◀════════════════════════════╣════════════════════════════════▶│
     │         WebRTC (P2P)       ║                                 │
     │                             ║                                 │
     │◀════════════════════════════╣════════════════════════════════▶│
     │      WebRTC via TURN       ║                                 │
```

## Services

The Relay Server provides:

| Service | Description |
|---------|-------------|
| Device registry | Which devices belong to which user, their current online/offline state |
| Presence | Real-time online/offline tracking via WebSocket heartbeats |
| Signaling | WebRTC SDP and ICE candidate relay between client and device |
| TURN | Media/data relay when P2P is not possible |
| Notification | Push notification delivery to clients |
| Pairing | Pairing code verification and key exchange relay |

## Device Registration

When a device's Runtime connects to the Relay Server:

1. WebSocket connection established with device token
2. Server looks up device by Device ID
3. If device is associated with a user account:
   - Device is marked ONLINE
   - Server notifies all connected clients of the user: `relay.device.event` (ONLINE)
   - Server sends current state to device: `relay.device.state` (list of authorized clients)
4. If device is NOT associated with a user account (unpaired):
   - Device is held in an "unpaired" pool
   - Server waits for pairing initiation

## Client Connection

When a client (app/CLI) connects to the Relay Server:

1. WebSocket connection established with session token
2. Server identifies the user
3. Server sends device list: `relay.device.list`
4. Server begins sending real-time device events: ONLINE, OFFLINE, state changes

## Message Routing

The Relay Server routes messages between connected clients and devices:

### Direct Routing (Client ↔ Device)

When the WebRTC connection is active, data flows through the data channels. The Relay Server is not involved in data-plane communication.

### Relayed Routing (Client ↔ Device via Relay)

When WebRTC is being established, or when TURN is used, the Relay Server forwards messages:

```
Client → Relay Server → Device
Device → Relay Server → Client
```

Relayed messages use the same envelope format and are forwarded as-is. The Relay Server does not inspect or modify message payloads (except for required header fields like routing).

### Offline Queuing

If a device is offline when a client sends a message, the Relay Server MAY queue messages:

| Message Type | Queue Behavior |
|-------------|----------------|
| `relay.sdp.*` | Not queued (session-specific, expires immediately) |
| `device.restart` | Not queued (device cannot execute while offline) |
| `action.execute` | Configurable (some extensions MAY accept queued actions) |
| `relay.notification.send` | Queued as push notification (see below) |

Queue size limit: 100 messages per device, max 24 hours.

## Push Notifications

The Relay Server delivers push notifications to clients when:

1. A device goes offline (after grace period)
2. A device comes online
3. A device sends a high-priority event (temperature alert, service failure)
4. An action completes on a device

Push notifications use:
- **Firebase Cloud Messaging (FCM)** for Android
- **Apple Push Notification Service (APNS)** for future iOS support

## TURN Service

The TURN service is a separate component (co-located or separate):

### Allocation

TURN credentials are:
- Generated per-session by the Relay Server
- Passed to both client and device during signaling
- Valid for the session duration (max 24 hours)
- Tied to a specific device+client pair

### Usage

TURN is used when:
1. ICE negotiation determines no P2P path is available
2. One or both sides are behind symmetric NAT
3. One or both sides are behind CGNAT (Carrier-Grade NAT)
4. Firewall rules block UDP (TCP TURN fallback)

### TURN Server Requirements

| Requirement | Specification |
|-------------|---------------|
| Protocol | TURN (RFC 5766) over UDP and TCP |
| Authentication | Time-limited credentials |
| Relay type | UDP (preferred), TCP (fallback) |
| Bandwidth per session | Up to 5 Mbps |
| Concurrent sessions | Proportional to server capacity |

## Connection State Management

The Relay Server maintains a state machine per device:

```
                    ┌────────────┐
                    │            │
         ┌─────────▶│  ONLINE    │◀─────────┐
         │          │            │          │
         │          └────────────┘          │
         │                                    │
    WebSocket                              WebSocket
    connected                              reconnected
         │                                    │
         │          ┌────────────┐            │
         │          │            │            │
         └──────────│  OFFLINE   │────────────┘
                    │            │
                    └────────────┘
                         │
                    Grace period
                    (2 minutes)
                         │
                         v
                    ┌────────────┐
                    │            │
                    │  OFFLINE   │ (notification sent)
                    │ (alerted)  │
                    └────────────┘
```

## Scalability

The Relay Server is designed to be horizontally scalable:

- **Stateless signaling:** All signaling state is ephemeral, held in-memory (no persistence needed)
- **Sticky routing:** A consistent hash of device_id routes a device to a specific relay node
- **TURN farm:** TURN servers are independently scalable, added behind a load balancer
- **No shared state:** Each relay node operates independently; presence is gossiped between nodes (optional, for cross-node signaling)

## Security

The Relay Server:

- Does NOT store identity keys
- Does NOT have access to WebRTC-encrypted data
- Does NOT log message payloads (metadata only: message type, size, timestamp)
- Validates all tokens on every connection
- Rate limits per-connection and per-user
- Supports connection blacklisting
