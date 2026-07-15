# Connection Engine

The Connection Engine is the core networking component that enables BuzzPi's defining promise: **no IP addresses, no port forwarding, no configuration.**

## Architecture

```
┌──────────┐         ┌──────────────┐         ┌──────────┐
│          │  WebRTC │              │  WebRTC │          │
│   App    │◀════════▶│  Relay       │◀════════▶│ Runtime  │
│          │  Data   │  Server      │  Data   │          │
└──────────┘         └──────────────┘         └──────────┘
     │                      │                      │
     │   WebSocket          │   WebSocket          │
     │   (Signaling)        │   (Signaling)        │
     └──────────────────────┴──────────────────────┘
```

## Connection Establishment

### Step 1: Device Connects to Relay

When Runtime starts, it opens a persistent outbound WebSocket connection to the Relay Server:

```yaml
WebSocket handshake:
  url: wss://jphat.net/buzzpi/relay/ws
  headers:
    X-Device-ID: <device_id>
    X-Auth-Token: <session_token>
    
  on_connect:
    - Register device as available
    
  on_disconnect:
    - Mark device as offline (with grace period)
    - Attempt reconnection with exponential backoff
    - Reconnection: 1s, 2s, 4s, 8s, 16s, 30s (cap)
```

This connection is the device's permanent link to BuzzPi. As long as it's alive, the device is reachable. No inbound ports required.

### Step 2: App Connects to Relay

When the user opens the app, it opens its own persistent WebSocket to the Relay Server:

```yaml
WebSocket handshake:
  url: wss://jphat.net/buzzpi/relay/ws
  headers:
    X-User-ID: <user_id>
    X-Auth-Token: <auth_token>
    
  on_connect:
    - Receive device list with states
    - Register for device events
```

### Step 3: Request Connection

When the user taps a device:

1. App sends `request_connection` message to Relay with `device_id`
2. Relay forwards to device's Runtime: `incoming_connection_request`
3. If device accepts (it's online and available), relay notifies both parties
4. Both app and device begin WebRTC ICE negotiation

### Step 4: WebRTC Negotiation

```yaml
WebRTC configuration:
  ICE_servers:
    - stun:stun.l.google.com:19302
    - stun:stun1.l.google.com:19302
    - turn:jphat.net/buzzpi/turn:3478 (relay fallback)
    
  media:
    data_channel: true           # Bidirectional data
    video: true                  # Screen streaming
    audio: false                 # No audio needed
    
  data_channel:
    - terminal: stdin/stdout
    - screen: video frames
    - services: status updates
    - files: file transfer
    - extensions: extension data
```

### Step 5: Connection Type Determination

```
Connection preference:
  1. P2P (STUN only)       ─── Best latency, no relay overhead
  2. P2P (STUN + TURN)     ─── Still direct, TURN as backup
  3. Relay (TURN only)     ─── Through relay server, higher latency
  
Result:
  - App and device negotiate via ICE
  - Best available path is selected automatically
  - Connection type is reported but not user-visible
```

## Data Channels

Each channel is a WebRTC data channel or media track:

| Channel | Type | Direction | Protocol | Use |
|---------|------|-----------|----------|-----|
| Terminal | Data | Bidirectional | Raw (JSON-framed) | stdin/stdout for shell |
| Screen | Video | Device → App | H.264 video | Remote display |
| Services | Data | Bidirectional | BPP messages | Service queries |
| Files | Data | Bidirectional | Chunked binary | File transfer |
| Events | Data | Device → App | BPP messages | Real-time status updates |
| Remote Input | Data | App → Device | Raw | Mouse/keyboard events |

## Relay Server

The Relay Server is a lightweight Go service that:

1. **Maintains WebSocket connections** to all online devices and active app sessions
2. **Routes signaling messages** between app and device during WebRTC negotiation
3. **Acts as TURN server** for connections that cannot establish P2P
4. **Maintains device presence** (heartbeat tracking)
5. **Delivers queued messages** for offline devices (configurable, limited)

### Heartbeat Protocol

```
App ──────────────────────── Relay ──────────────────────── Device
    ◄── heartbeat (30s) ──┄    ◄── heartbeat (30s) ──┄
    ──┄ ack ───────────────►    ──┄ ack ───────────────►
    
    Grace period: 2 minutes
    After 4 missed heartbeats (2 min): mark OFFLINE
    After 1st missed heartbeat: start grace period timer
    Within grace period: any heartbeat restores ONLINE
    After grace period: notification sent, transition complete
```

### Connection Fallback

```
Scenario: App and Device are on different NATs
    
1. STUN discovers public IPs for both
2. ICE attempts P2P connection
3. P2P fails (symmetrical NAT, CGNAT, firewall)
4. ICE falls back to TURN relay
5. Data flows through Relay Server as TURN
6. Latency increases by ~50ms but connection works
```

## CLI Connection

The CLI connects through the same Relay Server but via a different protocol:

```
CLI → Relay (WebSocket) → Runtime (WebSocket)
```

The CLI authenticates with the user's API token (not a session token) and can connect to any of the user's devices. This enables scripts and automation to interact with devices through the same relay infrastructure.

## Security Properties

| Property | Mechanism |
|----------|-----------|
| Authentication | Session tokens + device identity keys |
| Encryption | DTLS-SRTP (WebRTC encryption, mandatory) |
| End-to-end | Data never decrypted by Relay (WebRTC is E2E) |
| Identity binding | Device identity key pinned during pairing |
| Forward secrecy | WebRTC provides per-session key agreement |
| No inbound ports | All connections are outbound (WebSocket from device) |
