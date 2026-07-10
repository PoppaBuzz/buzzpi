# RFC-0001: Connection Engine

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Core |
| **Created** | 2026-07-07 |
| **Updated** | 2026-07-07 |
| **Requires** | None |
| **Replaces** | None |
| **Superseded by** | None |

## Summary

The Connection Engine is the core networking component that establishes and maintains communication between a BuzzPi client (app/CLI) and a device's Runtime. It handles ICE negotiation, NAT traversal, WebRTC data channels, relay fallback, reconnection, and quality adaptation. This RFC specifies the implementation architecture of the Connection Engine as a Go library and Android module.

## Motivation

The Connection Engine is the highest-risk technical component in the BuzzPi platform. Its correctness determines whether users can reliably connect to devices without networking expertise. Key risks include:

1. **NAT traversal failure** — Devices behind CGNAT, symmetrical NAT, or enterprise firewalls
2. **Connection quality** — High latency or packet loss degrading terminal/screen experience
3. **Reconnection reliability** — Sessions lost during network transitions (WiFi → cellular)
4. **Screen streaming performance** — H.264 encoding on Raspberry Pi hardware

This RFC addresses these risks through a phased implementation with measurable success criteria.

## Design

### Architecture

```
┌───────────────────┐     WebSocket      ┌───────────────────┐
│                   │◀══════════════════▶│                   │
│  ConnectionEngine │    (Signaling)     │  ConnectionEngine │
│  (Client)         │                    │  (Device/Runtime) │
│                   │     WebRTC P2P     │                   │
│                   │◀══════════════════▶│                   │
└───────────────────┘                    └───────────────────┘
         │                                      │
         │ ThroughputMonitor                    │ ThroughputMonitor
         │ ReconnectionManager                  │ ReconnectionManager
         │ QualityAdaptation                    │ QualityAdaptation
```

### Core Interfaces

```go
// ConnectionEngine manages the lifecycle of a connection between client and device.
type ConnectionEngine struct {
    state     ConnectionState
    transport Transport
    monitor   *ThroughputMonitor
    reconn    *ReconnectionManager
    channels  map[string]*DataChannel
}

// Transport abstracts the underlying connection path.
type Transport interface {
    Connect(ctx context.Context) error
    Disconnect() error
    State() ConnectionState
    Type() TransportType // direct | relay
    Stats() TransportStats
}

// DataChannel represents a single logical data channel.
type DataChannel struct {
    Label string          // "terminal", "screen", "files", etc.
    Data  chan []byte     // Raw data (protobuf or JSON framed)
    Done  chan struct{}
}

// TransportStats for quality monitoring.
type TransportStats struct {
    RTTMs            int
    PacketLossPercent float64
    BytesSent         int64
    BytesReceived     int64
    AvailableBandwidth int64
}
```

### Connection Establishment

```
Client                          Relay Server                    Device
  │                                 │                             │
  │  1. relay.connect.request       │                             │
  │ ──────────────────────────────►│                             │
  │                                 │  2. relay.connect.request   │
  │                                 │ ───────────────────────────►│
  │                                 │                             │── Accept/reject
  │                                 │  3. relay.connect.accept    │
  │                                 │ ◀───────────────────────────│
  │  4. relay.connect.accepted      │                             │
  │ ◀──────────────────────────────│                             │
  │                                 │                             │
  │  5. WebRTC SDP Offer            │                             │
  │ ──────────────────────────────►│  6. WebRTC SDP Offer        │
  │                                 │ ───────────────────────────►│
  │                                 │  7. WebRTC SDP Answer       │
  │                                 │ ◀───────────────────────────│
  │  8. WebRTC SDP Answer           │                             │
  │ ◀──────────────────────────────│                             │
  │                                 │                             │
  │  9. ICE Candidates             │                             │
  │ ◀══════════════════════════════╣════════════════════════════►│
  │                                 │                             │
  │ 10. WebRTC P2P Connected       │                             │
  │ ◀══════════════════════════════╣════════════════════════════►│
```

### Transport Selection Algorithm

```go
func (e *ConnectionEngine) selectTransport(ctx context.Context) error {
    // Phase 1: Attempt P2P via STUN
    p2pConfig := webrtc.Configuration{
        ICEServers: []webrtc.ICEServer{
            {URLs: []string{"stun:stun.l.google.com:19302"}},
        },
    }

    pc, err := e.createPeerConnection(p2pConfig)
    if err != nil {
        return err
    }

    // Wait for ICE gathering + connection (timeout: 10s)
    select {
    case <-e.connected():
        e.transport = &DirectTransport{pc: pc}
        return nil
    case <-time.After(10 * time.Second):
        pc.Close()
    }

    // Phase 2: Fallback to TURN relay
    turnConfig := webrtc.Configuration{
        ICEServers: []webrtc.ICEServer{
            {URLs: []string{"stun:stun.l.google.com:19302"}},
            {
                URLs:       []string{"turn:turn.buzzpi.dev:3478"},
                Username:   e.turnCreds.Username,
                Credential: e.turnCreds.Credential,
            },
        },
    }

    pc, err = e.createPeerConnection(turnConfig)
    if err != nil {
        return fmt.Errorf("P2P and relay both failed: %w", err)
    }

    select {
    case <-e.connected():
        e.transport = &RelayTransport{pc: pc}
        return nil
    case <-time.After(30 * time.Second):
        return ErrTransportUnavailable
    }
}
```

### Quality Adaptation

```go
// ThroughputMonitor collects WebRTC stats and drives quality adaptation.
type ThroughputMonitor struct {
    stats        []TransportStats
    windowSize   int // 5s window, sampled every 1s
    adaptationFn func(stats TransportStats) AdaptationAction
}

func (m *ThroughputMonitor) tick(stats TransportStats) {
    m.stats = append(m.stats, stats)
    if len(m.stats) > m.windowSize {
        m.stats = m.stats[1:]
    }

    avg := average(m.stats)

    switch {
    case avg.PacketLossPercent > 5:
        m.adaptationFn(AdaptationAction{LowerQuality: true})
    case avg.PacketLossPercent < 1 && avg.RTTMs < 100:
        m.adaptationFn(AdaptationAction{RaiseQuality: true})
    }
}
```

Quality profiles define the adaptation targets:

| Profile | Max Resolution | Max FPS | Max Bitrate | Trigger |
|---------|---------------|---------|-------------|---------|
| High | 1920x1080 | 30 | 5 Mbps | Default on P2P |
| Medium | 1280x720 | 15 | 2 Mbps | Packet loss > 3% |
| Low | 854x480 | 10 | 1 Mbps | Packet loss > 5% |
| Minimum | 640x360 | 5 | 500 Kbps | Packet loss > 10% or relay |

### Reconnection Manager

```go
type ReconnectionManager struct {
    config     ReconnectionConfig
    attempts   int
    backoff    time.Duration
    sessionID  string
}

type ReconnectionConfig struct {
    InitialDelay time.Duration // 1s
    MaxDelay     time.Duration // 30s
    MaxAttempts  int           // 10
    Jitter       time.Duration // 500ms
    Multiplier   float64       // 2.0
}

func (rm *ReconnectionManager) attempt() time.Duration {
    delay := rm.backoff
    jitter := time.Duration(rand.Int63n(int64(rm.config.Jitter)))
    rm.backoff = time.Duration(float64(rm.backoff) * rm.config.Multiplier)
    if rm.backoff > rm.config.MaxDelay {
        rm.backoff = rm.config.MaxDelay
    }
    delay += jitter
    rm.attempts++
    return delay
}
```

### Data Channel Layering

```
┌──────────────────────────────────────────────────────┐
│                 BPP Message Layer                     │
│  (protobuf envelopes: method, params, rid)           │
├──────────────────────────────────────────────────────┤
│              Data Channel Multiplexer                 │
│  (one channel per type: terminal, screen, files...)  │
├──────────────────────────────────────────────────────┤
│              WebRTC Data Channel API                  │
│  (ordered, unreliable SCTP)                          │
├──────────────────────────────────────────────────────┤
│                   ICE / DTLS                          │
├──────────────────────────────────────────────────────┤
│          UDP (P2P)  |  TCP (TURN relay)              │
└──────────────────────────────────────────────────────┘
```

### Screen Capture Subsystem

```go
// Capturer captures device frames for screen streaming.
type Capturer interface {
    Start(resolution Resolution, fps int) (<-chan Frame, error)
    Stop() error
    Methods() []CaptureMethod // ["drm", "fbdev", "x11"]
}

// Encoder encodes frames into H.264 NAL units.
type Encoder interface {
    Encode(frame Frame) ([][]byte, error)
    SetBitrate(bps int) error
    SetFPS(fps int) error
    ForceKeyframe() error
}
```

### Hardware Acceleration Detection

```go
func detectScreenCapture() CaptureMethod {
    // Check DRM/KMS first (best quality)
    if _, err := os.Stat("/dev/dri/card0"); err == nil {
        // Try to open DRM device and check for overlay planes
        if drmAvailable() {
            return CaptureDRM
        }
    }

    // Fall back to fbdev
    if _, err := os.Stat("/dev/fb0"); err == nil {
        return CaptureFBDev
    }

    // Fall back to X11 (if desktop environment)
    if os.Getenv("DISPLAY") != "" {
        return CaptureX11
    }

    return CaptureUnavailable
}

func detectHardwareEncoder() EncoderType {
    // Check VideoCore MMAL on Raspberry Pi
    if hasVideoCore() {
        return EncoderMMAL
    }

    // Check V4L2 hardware encoder
    if hasV4L2Encoder() {
        return EncoderV4L2
    }

    // Software fallback
    return EncoderSoftwareX264
}
```

## Implementation Plan

### Phase 1: Core Library (Week 1-2)

- Go module: `github.com/buzzpi/buzzpi/connection`
- Transport interface and ICE integration
- WebSocket signaling client (Gorilla WebSocket)
- WebRTC peer connection (Pion)
- Data channel multiplexing
- Tests: unit tests with mock ICE servers

**Deliverable:** Connection engine Go library with P2P connectivity over LAN

### Phase 2: Relay Integration (Week 3-4)

- Relay WebSocket protocol implementation
- TURN credential management
- Signaling message passing
- Connection state machine
- Tests: integration tests against a test relay server

**Deliverable:** End-to-end connectivity via relay when P2P fails

### Phase 3: Quality Monitoring (Week 5)

- WebRTC stats collection (ICE candidate pairs, RTT, packet loss)
- ThroughputMonitor with sliding window
- Quality adaptation triggers
- Transport type detection
- Tests: simulated packet loss and latency

**Deliverable:** Automatic quality adaptation based on network conditions

### Phase 4: Reconnection (Week 6)

- ReconnectionManager with exponential backoff
- Terminal session persistence across reconnects
- Screen stream restart on reconnect
- Buffer and replay terminal output
- Tests: network drop and restore scenarios

**Deliverable:** Seamless reconnection with terminal session preservation

### Phase 5: Screen Capture (Week 7-8)

- DRM/KMS frame capture on Raspberry Pi
- MMAL H.264 hardware encoding
- Software encoding fallback (x264)
- Damage tracking (only encode changed tiles)
- Adaptive bitrate via WebRTC stats
- Tests: frame encoding on Raspberry Pi hardware

**Deliverable:** Screen streaming at 30fps with <500ms latency

## Dependencies

| Dependency | Version | Purpose | License |
|------------|---------|---------|---------|
| [Pion WebRTC](https://github.com/pion/webrtc) | v3 | WebRTC implementation | MIT |
| [Gorilla WebSocket](https://github.com/gorilla/websocket) | v1.5 | WebSocket client | BSD-2 |
| [pion/interceptor](https://github.com/pion/interceptor) | latest | WebRTC stats | MIT |
| [gen2brain/x264-go](https://github.com/gen2brain/x264-go) | latest | Software H.264 encoding | BSD-2 |
| Go standard library | 1.22+ | net, crypto, encoding | BSD |

## Testing Strategy

| Test Type | Scope | Tooling |
|-----------|-------|---------|
| Unit tests | Transport interface, state machine, reconnection logic | Go testing |
| Integration | WebRTC P2P between two processes on same host | Pion + test relay |
| Integration | WebRTC via relay (simulate non-P2P networks) | Test relay server |
| Integration | Screen capture and encoding | Raspberry Pi device |
| Chaos | Packet loss, latency, reconnection scenarios | toxiproxy |
| Chaos | Network transition (WiFi off → cellular on) | Manual on device |

### Success Criteria

| Criterion | Phase | Target |
|-----------|-------|--------|
| P2P connection established | 1 | <5s (LAN) |
| Relay connection established | 2 | <10s (WAN) |
| P2P success rate (lab) | 1 | >99% |
| Relay fallback success | 2 | >99% |
| Screen streaming latency (P95) | 5 | <500ms |
| Screen streaming FPS | 5 | >25fps at 720p |
| Reconnection (network drop <10s) | 4 | 100% session preserved |
| Reconnection (network drop <60s) | 4 | >80% session preserved |

## Open Questions

1. **TURN server scaling:** What is the bandwidth cost per concurrent screen stream? Estimate: 2 Mbps per stream. At 1000 concurrent streams = 2 Gbps egress. Need cost analysis before production deployment.
2. **VideoCodec V4L2 on Pi 5:** Does the Pi 5 VideoCore VII expose an H.264 encoder via V4L2? If not, MMAL or software encoding only.
3. **mDNS library:** Which Go mDNS library for local discovery? Options: `hashicorp/mdns`, `grandcat/zeroconf`. Both need evaluation for reliability on Raspberry Pi.
4. **STUN server dependency:** Using Google's public STUN servers. Acceptable for MVP? Evaluate running our own STUN server for production.

## Alternatives Considered

1. **WebSocket-only tunneling (no WebRTC):** Simpler implementation but higher latency and relay bandwidth costs. Rejected because screen streaming would require proxy server with high egress costs.
2. **libp2p for NAT traversal:** More feature-rich but heavier dependency. Rejected because WebRTC is more widely supported and better documented.
3. **VNC for screen sharing:** No NAT traversal, requires port forwarding or VPN. Rejected per Manifesto principle (no IP addresses).
4. **Software-only H.264 (libx264):** Works everywhere but high CPU usage on Pi Zero. Hardware encoding is preferred but software is acceptable fallback for Pi 4+.
