# BuzzPi Platform Roadmap

**Version:** v0.0.0 (architecture exists, code begins)

---

## Phase 0 — Foundation (v0.0.x)

- Repository structure and organization
- Vision, Manifesto, and Architecture documents
- Community guidelines (contributing, code of conduct, governance)
- Security policy
- Development roadmap
- RFC process established
- Branding and design language
- Architecture Decision Records (DECISIONS.md)
- Book structure (Product, Experience, Engineering, Protocol, Community, Patterns, Reference)

Status: **In Progress**

---

## Phase 0.5 — Validation

Before writing production code, validate the ideas:
- Interview Raspberry Pi users
- Watch beginners set up a Pi — identify where they get stuck
- Measure how long common tasks take
- Test wireframes before implementation
- Build proof-of-concepts for risky technical areas (screen streaming, connection engine)
- Competitive analysis against existing tools

Status: **Planned**

---

## Phase 1 — Core (v0.1.x — v0.2.x)

- BuzzPi Agent in Go (systemd service)
- BuzzPi Android app with Jetpack Compose and Material 3
- LAN discovery via mDNS/DNS-SD
- QR code pairing
- **Remote Desktop** — graphical screen streaming via WebRTC with adaptive quality, touch/mouse input, LAN-optimized
- Terminal with full ANSI color support
- File transfer (browse, upload, download, rename)
- Live system stats dashboard (CPU, memory, temperature, storage)
- Connection Engine with local transport selection
- BPP specification draft (connection, authentication, events)

Status: **Planned**

---

## Phase 2 — Remote Access (v0.5.x)

- Backend: single Go service + PostgreSQL (no Redis until needed)
- Cloud registry for device discovery
- Tailscale/MagicDNS integration
- WebRTC peer-to-peer connectivity
- Cloud relay fallback
- Push notifications (overheating, low storage, service failures, updates)
- Auto-reconnect across network changes
- End-to-end encrypted sessions

Status: **Planned**

---

## Phase 3 — Power Features (v0.5.x)

- Docker management (containers, images, compose, logs, shell)
- Camera streaming and snapshots
- GPIO interactive pin controls
- Service management (systemd)
- Log viewer with search and filtering
- Adaptive dashboard that detects installed software

Status: **Planned**

---

## Phase 4 — Intelligence and Ecosystem (v0.9.x)

- BuzzAI assistant — tool-based diagnostics with MCP-like architecture
- Plugin SDK and developer documentation
- Automation recipes
- BuzzPi Desktop (cross-platform)
- BuzzPi CLI
- Plugin certification program
- Developer portal
- Full documentation site
- Community translation platform

Status: **Planned**

---

## v1.0.0 — Stable Release

- BPP 1.0 protocol specification finalized
- Stable plugin API
- Production-ready agent and Android app
- Backend stable and deployable
- Protocol freeze for v1.x cycle

Status: **Planned**

---

## Long-Term Vision

BuzzPi becomes the definitive platform for managing Raspberry Pis — and eventually any Linux device. When someone buys their first Pi and asks "How do I manage it?", the answer is BuzzPi. Not because it is required, but because it is the easiest way.

The BuzzPi Protocol (BPP) enables any device to participate in the ecosystem, expanding from Raspberry Pi to a universal remote device management platform.
