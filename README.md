# BuzzPi Platform — Your Raspberry Pi. Anywhere. Instantly.

**v0.1.0** — First end-to-end system. Agent, Android, terminal, file manager, screen viewer.

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Status: Foundation](https://img.shields.io/badge/status-foundation-yellow)](books/product/roadmap.md)

> A Raspberry Pi should feel like a Bluetooth speaker. You plug it in. It appears. You tap it. It works.

Read in this order: [North Star](NORTH_STAR.md) → [Constitution](CONSTITUTION.md) → [Canon](CANON.md). That is the belief system of BuzzPi. Everything else follows from it.

The BuzzPi Platform is an open-source ecosystem for managing Raspberry Pis — and eventually any Linux device — without ever thinking about networking. It is not an app. It is a protocol, an agent, a set of clients, and a community, all designed around one principle: **the user should never have to know an IP address.**

---

## The Platform

```
BuzzPi Platform
│
├── BuzzPi Android    — Native Android companion app
├── BuzzPi Agent      — Go daemon running on each Pi
├── BuzzPi Desktop    — Cross-platform desktop client (future)
├── BuzzPi CLI        — Command-line client (future)
├── BuzzPi Cloud      — Relay and registry services
├── BuzzPi SDK        — Plugin and integration SDK
└── BPP               — BuzzPi Protocol (the contract)
```

Any device that implements BPP can participate in the ecosystem. Any client that implements BPP can manage any device.

---

## Key Features

- **Remote Desktop** — full graphical screen streaming with touch and mouse input. See and control your Pi's desktop from your phone, on LAN or remotely
- **Automatic discovery** — Pis appear on your network without configuration
- **One-tap connection** — tap a device, and you are connected
- **Terminal** — full terminal with ANSI colors, multiple tabs, search, copy/paste
- **File manager** — browse, upload, download, rename, preview files
- **Docker management** — containers, images, compose, logs, resource monitoring
- **GPIO control** — interactive pin controls
- **Camera streaming** — live stream, snapshots, recording
- **System monitoring** — CPU, memory, temperature, storage, network, uptime
- **AI assistant** — diagnose issues, read logs, restart services, answer questions
- **Push notifications** — overheating, low storage, service failures, updates

---

## The Eight Books

BuzzPi is documented as a series of books. Each answers a fundamental question.

| Book | Question | Status | Chapters |
|------|----------|--------|----------|
| [Product](books/product/README.md) | Why does BuzzPi exist? | Complete | 12/12 |
| [Experience](books/experience/README.md) | How should BuzzPi feel? | Complete | 11/11 |
| [Engineering](books/engineering/README.md) | How does BuzzPi work? | Complete | 14/14 |
| [Protocol](books/protocol/README.md) | How do devices communicate? | Complete | 24/24 |
| [Community](books/community/README.md) | How do we build this together? | Complete | 12/12 |
| [Patterns](books/patterns/README.md) | How does BuzzPi think? | Complete | 12/12 |
| [Reference](books/reference/README.md) | Where are the facts? | Complete | 20/20+ |
| [Lexicon](books/lexicon/README.md) | What do we call things? | Complete | 4/4 |

---

## Versioning

BuzzPi uses semantic versioning, but the meaning extends beyond code:

| Version | Meaning |
|---------|---------|
| v0.0.x | Foundation and canon — books, RFCs, governance, architecture |
| v0.1.x | Proofs of concept — discovery, protocol handshake, remote desktop experiments |
| v0.2.x | First end-to-end system — agent to Android, terminal, discovery |
| v0.5.x | Feature-complete MVP |
| v0.9.x | Community preview, stabilization, plugin API freeze |
| v1.0.0 | Stable public release with BPP 1.0 |

v0.0.0 declares that the project began when the architecture began, not when the coding began.

---

## Project Status

v0.1.0 — First end-to-end system. See the [roadmap](books/product/roadmap.md) for details.

---

## Contributing

BuzzPi is open source and welcomes contributors. See the [Community book](books/community/README.md) to get started.

All contributors must follow our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## License

MIT License — see [LICENSE](LICENSE) for details.
