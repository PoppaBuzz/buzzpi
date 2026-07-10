# Network Topologies

**Reference of supported network topologies for BuzzPi device-client communication.** Each topology describes the transports available, how connections are established, and tradeoffs.

---

## Topology 1: LAN Direct

```
┌──────────────────┐                     ┌──────────────────┐
│    Client        │                     │    Runtime       │
│  (Phone/CLI)     │◀───────────────────▶│  (Raspberry Pi)  │
│                  │    WebSocket TLS    │                  │
│                  │    mDNS Discovery   │                  │
│                  │    IP: 192.168.1.42 │                  │
└──────────────────┘                     └──────────────────┘
        │                                      │
   Same subnet (192.168.1.0/24)
```

**Transports:**
- **mDNS discovery:** `_buzzpi._tcp` — automatic, no config
- **WebSocket:** `wss://192.168.1.42:49872/ws` — direct TLS connection
- **Screen streaming:** WebRTC over LAN (sub-10ms latency)

**Latency:** <5ms RTT typical
**Bandwidth:** Full LAN speed (100Mbps+ on Pi Ethernet, 300Mbps+ WiFi 5)
**NAT traversal:** Not required

**Connection flow:**
1. Client joins WiFi network
2. Client browses mDNS for `_buzzpi._tcp`
3. Runtime discovered at `192.168.1.42:49872`
4. Client establishes direct WebSocket connection
5. BPP messages flow directly, no intermediaries

**Best for:** Home use, office LAN, same-network setups

---

## Topology 2: Remote (Cloud Relay)

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│    Client        │     │   Cloud Relay    │     │    Runtime       │
│  (Phone/4G)      │────▶│  relay.buzzpi   │◀────│  (Raspberry Pi)  │
│                  │     │  .cloud:443      │     │                  │
│  IP: 10.0.0.5   │     │  Packet Forward  │     │  IP: 192.168.    │
│  (mobile)        │     │   (no decrypt)   │     │  1.42 (NAT)     │
└──────────────────┘     └──────────────────┘     └──────────────────┘
        │                        │                        │
    Behind CGNAT            Public cloud             Behind NAT
```

**Transports:**
- **Discovery:** Cloud Relay's device registry (requires prior pairing)
- **WebSocket:** `wss://relay.buzzpi.cloud/ws/client` and `/ws/device`
- **Screen streaming:** WebRTC over TURN (if P2P fails)

**Latency:** 20-100ms RTT (depends on relay location and client network)
**Bandwidth:** Limited by relay bandwidth and client connection
**NAT traversal:** Full cone / symmetric NAT via relay

**Connection flow:**
1. Runtime connects to Cloud Relay via persistent WebSocket (`/ws/device`)
2. Client connects to Cloud Relay via WebSocket (`/ws/client`)
3. Relay maps client ↔ device based on pairing registry
4. Messages forwarded bidirectionally; relay does not inspect payloads

**Best for:** Remote access when away from home network

---

## Topology 3: P2P Assisted (ICE/STUN)

```
┌──────────────────┐        STUN Server        ┌──────────────────┐
│    Client        │◀─────────────────────────▶│    Runtime       │
│  (Phone/4G)      │   ICE Candidate Exchange  │  (Raspberry Pi)  │
│                  │◀══════════════════════════▶│                  │
│  IP: 10.0.0.5   │   Direct P2P WebRTC        │  IP: 192.168.   │
│  (mobile)        │   Data Channel             │  1.42 (NAT)     │
└──────────────────┘                            └──────────────────┘
        │                                             │
    Behind CGNAT                                 Behind NAT
         │                                             │
         └──────────── STUN: stun.buzzpi.cloud ────────┘
```

**Transports:**
- **Signaling:** Via Cloud Relay (SDP + ICE candidates)
- **Media:** Direct P2P WebRTC data channel
- **Fallback:** TURN relay if P2P fails

**Latency:** 5-30ms RTT (P2P), 30-100ms (TURN fallback)
**Bandwidth:** P2P: up to client/device link speed. TURN: relay bandwidth
**NAT traversal:** STUN for address discovery, ICE for connectivity checks

**Connection flow:**
1. Client and Runtime already paired
2. Both connect to signaling channel (via relay)
3. STUN server provides public IP:port mapping for each peer
4. ICE connectivity check finds best path
5. Direct P2P channel established
6. Signaling channel closed; data flows P2P

**Best for:** Remote access with lower latency than relay; devices behind moderate NAT

---

## Topology 4: TURN Relay (Symmetric NAT)

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│    Client        │────▶│   TURN Server    │◀────│    Runtime       │
│  (Phone/4G)      │     │  turn.buzzpi.   │     │  (Raspberry Pi)  │
│                  │     │  cloud:3478      │     │                  │
│  Symmetric NAT   │     │  Relays media    │     │  Symmetric NAT   │
└──────────────────┘     └──────────────────┘     └──────────────────┘
```

**Transports:**
- **Signaling:** Via Cloud Relay
- **Media:** TURN relay (all traffic through TURN)
- **No P2P possible** — both peers on symmetric NATs

**Latency:** 30-150ms RTT (depends on TURN server proximity)
**Bandwidth:** Limited by TURN server bandwidth (most expensive topology)
**NAT traversal:** TURN allocates relayed transport addresses

**When P2P fails (ICE concluded):**
1. Both peers behind symmetric NAT (common on 4G/5G)
2. Firewall blocks UDP entirely (corporate networks)
3. STUN-derived candidates unreachable

**Best for:** Fallback when P2P is not possible; enterprise/corporate networks

---

## Topology 5: LAN + Cloud Hybrid

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│    Client        │     │   Cloud Relay    │     │    Runtime       │
│                  │◀───▶│  (Control Only)  │◀───▶│                  │
│  (Phone/CLI)     │     │                  │     │  (Raspberry Pi)  │
│                  │◀════════════════════════════▶│                  │
│  192.168.1.50    │     Direct LAN (Screen/File)  │  192.168.1.42   │
└──────────────────┘                               └──────────────────┘
```

**Transports:**
- **Control plane:** Via Cloud Relay (pairing, session management, presence)
- **Data plane:** Direct LAN (screen streaming, file transfer, terminal)

**Benefits:**
- Low latency for media (LAN direct)
- Presence and notification working even when switching networks
- Seamless transition from remote to LAN and back
- Single connection endpoint for app (always relay, transparently optimized)

**Connection flow:**
1. Client discovers device via relay (previously paired)
2. Client also detects device on LAN via mDNS
3. Control messages route through relay (session heartbeats, presence)
4. Media routes direct via LAN WebSocket
5. When client leaves LAN, media automatically fails over to relay

**Best for:** Devices used both locally and remotely; seamless transition

---

## Topology Comparison

| Topology | Discovery | Latency | Bandwidth | NAT Required | Setup | Best For |
|----------|-----------|---------|-----------|-------------|-------|----------|
| **LAN Direct** | mDNS | <5ms | Full LAN | No | None | Home/office |
| **Cloud Relay** | Registry | 20-100ms | Limited | Yes | Pair once | Remote access |
| **P2P (ICE)** | Signaling | 5-30ms | Link speed | STUN | Pair once | Remote + med NAT |
| **TURN Relay** | Signaling | 30-150ms | TURN BW | Yes | Pair once | Symmetric NAT |
| **LAN+Cloud** | Both | <5ms/20-100ms | Both | Hybrid | Pair once | Best experience |

---

## Connection Priority

The Connection Engine probes transports in this order:

```
1. LAN Direct (mDNS)
   ├── Available → Connect (fastest, lowest latency)
   └── Not found ↓

2. P2P (ICE/STUN)
   ├── Available → Connect (good latency, no relay cost)
   └── NAT too restrictive ↓

3. Cloud Relay
   └── Always available (connected via persistent WebSocket)
```

Transition is seamless: if LAN becomes available during a relay session, the Connection Engine transparently upgrades to LAN without interrupting active streams.

---

## Port Reference

| Service | Transport | Port | Protocol |
|---------|-----------|------|----------|
| WebSocket (LAN) | TCP | 49872 (dynamic) | TLS 1.3 |
| mDNS | UDP | 5353 | Multicast |
| STUN | UDP | 3478 | RFC 8489 |
| TURN | UDP/TCP | 3478-3479 | RFC 8656 |
| Cloud Relay | TCP | 443 | WSS/TLS |
| Health check | TCP | 9100 | HTTP/Prometheus |
| Unix socket | Unix | — | `/var/run/buzzpi.sock` |
