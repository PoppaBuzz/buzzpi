# Book 4: Protocol

**The BuzzPi Protocol Specification (BPP)**

**Status:** Open Specification
**License:** MIT
**Reference Implementation:** BuzzPi Runtime

BPP is an open protocol for remote device management. It is transport-agnostic, versioned, and extensible. Any device that implements BPP can participate in the ecosystem. Any client that implements BPP can manage any device. The BuzzPi Runtime is one reference implementation — not the only implementation.

Third-party developers should be able to implement a compatible client or agent using only this book, without reading any BuzzPi source code.

**Start here:** [North Star](../../NORTH_STAR.md) — [Constitution](../../CONSTITUTION.md) — [Protocol Constraints](constraints.md)

---

## Protocol Layers

BPP is organized into four independent layers. Each layer can evolve at its own version.

```
BPP
│
├── Identity Layer     — Who are you? Can you prove it?
├── Transport Layer    — How do messages move?
├── Services Layer     — What can you do?
└── Capabilities Layer — What do you support?
```

### Identity Layer

Who the device or user is, and how trust is established.

| Chapter | Status | Description |
|---------|--------|-------------|
| [01-pairing](01-pairing.md) | Complete | Device enrollment, trust-on-first-use, code pairing |
| [02-authentication](02-authentication.md) | Complete | Key exchange, challenge-response, session tokens |
| [03-trust](03-trust.md) | Complete | Key management, rotation, revocation, hardware roots |
| [04-identity](04-identity.md) | Complete | Device IDs, user identities, anonymous modes |

### Transport Layer

How messages move between client and device, regardless of content.

| Chapter | Status | Description |
|---------|--------|-------------|
| [05-websocket](05-websocket.md) | Complete | Primary transport, handshake, TLS, framing |
| [06-webrtc](06-webrtc.md) | Complete | Peer-to-peer, ICE/STUN/TURN, data channels |
| [07-relay](07-relay.md) | Complete | Cloud relay protocol, reconnection, queueing |
| [08-compression](08-compression.md) | Complete | Payload compression, codec negotiation |
| [09-heartbeats](09-heartbeats.md) | Complete | Keep-alive, ping/pong, dead-peer detection |

### Services Layer

The operations a client can perform on a device.

| Chapter | Status | Description |
|---------|--------|-------------|
| [10-terminal](10-terminal.md) | Complete | PTY session, resize, ANSI, copy/paste |
| [11-screen](11-screen.md) | Complete | Frame streaming, adaptive quality, input injection |
| [12-files](12-files.md) | Complete | Browse, upload, download, rename, watch |
| [13-stats](13-stats.md) | Complete | System metrics, history, thresholds |
| [14-gpio](14-gpio.md) | Complete | Pin read, write, PWM, event monitoring |
| [15-docker](15-docker.md) | Complete | Container lifecycle, images, compose, logs |
| [16-services](16-services.md) | Complete | Systemd service status, start, stop, restart |
| [17-logs](17-logs.md) | Complete | Journald streaming, filters, search |
| [18-camera](18-camera.md) | Complete | Video stream, snapshots, recording control |
| [19-rpc](19-rpc.md) | Complete | Request-response for custom commands |

### Capabilities Layer

Discovery, negotiation, and extensibility.

| Chapter | Status | Description |
|---------|--------|-------------|
| [20-discovery](20-discovery.md) | Complete | Capability querying, feature flags, metadata |
| [21-negotiation](21-negotiation.md) | Complete | Version negotiation, profile selection |
| [22-plugins](22-plugins.md) | Complete | Plugin registration, endpoint discovery |
| [23-extensions](23-extensions.md) | Complete | Custom message types, vendor extensions |
| [24-errors](24-errors.md) | Complete | Error codes, error messages, recovery hints |

---

## Protocol Properties

- **Transport-agnostic** — Identity, Services, and Capabilities work over any transport
- **Versioned per layer** — Identity v2 can ship while Transport stays at v1
- **Backward compatible** — Old clients talk to new agents within the supported range
- **Open** — Anyone can implement the spec without license or royalty
- **Extensible** — Custom message types and vendor extensions are first-class
