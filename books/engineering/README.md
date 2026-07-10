# Book 3: Engineering

**How does BuzzPi work?**

This book contains the complete technical architecture of the BuzzPi Platform. Every subsystem, protocol, state machine, and design decision is documented here.

**Start here:** [North Star](../../NORTH_STAR.md) — [Constitution](../../CONSTITUTION.md) — [Engineering Constraints](constraints.md)

## Contents

| Document | Status | Description |
|----------|--------|-------------|
| [Architecture](architecture.md) | Complete | System overview, component diagram, technology stack |
| [Security](security.md) | Complete | Threat model, encryption, authentication, privacy |
| [State Machines](state-machines.md) | Complete | Device lifecycle, pairing, connection, service state machines |
| [Data Model](data-model.md) | Complete | Core entities: Device, User, Session, Extension, Action |
| [Connection Engine](connection-engine.md) | Complete | WebRTC relay, ICE negotiation, NAT traversal |
| [API Design](api-design.md) | Complete | Message-oriented protocol, method catalog, versioning |
| [Runtime Spec](runtime-spec.md) | Complete | On-device daemon architecture, screen capture, self-update |
| [Android Architecture](android-architecture.md) | Complete | Kotlin/Compose app structure, DI, navigation |
| [Backend Architecture](backend-architecture.md) | Complete | Go backend service, relay, TURN, PostgreSQL |
| [Event Bus](event-bus.md) | Complete | Pub/sub system, event types, subscriptions |
| [Plugin System](plugin-system.md) | Complete | Plugin lifecycle, IPC, manifest, sandboxing |
| [Capability Model](capability-model.md) | Complete | Device feature detection, dynamic UI, extension caps |
| [BuzzAI](buzzai.md) | Complete | AI assistant architecture, tool system, MCP-like design |
| [API Reference](api-reference.md) | Complete | Complete BPP, REST, and IPC method catalog |
