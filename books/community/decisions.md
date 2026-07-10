# Architecture Decision Records

This document records significant architectural decisions made during BuzzPi's development. Each entry explains the context, the decision, alternatives considered, and the rationale.

New decisions should be added as the project evolves. If a decision is later reversed, the reversal is recorded as a new entry with a reference to the original.

---

## 001: Agent Language — Go over Rust

**Date:** 2026-07-07

**Context:** The BuzzPi Agent needs to run on Raspberry Pi (ARM), be a single distributable binary, have excellent networking support, and be maintainable by a broad contributor base.

**Decision:** Go.

**Alternatives considered:**

- **Rust** — Excellent performance and safety, but steeper learning curve for contributors and slower iteration speed. Cross-compilation for ARM is more complex. The agent does not need the level of memory safety Rust provides — it runs in userspace with limited trust boundaries.
- **Python** — Too slow for screen capture encoding, WebSocket handling at scale, and system-level operations. Dependency management on the Pi is fragile.
- **C/C++** — Unnecessarily low-level. Would increase development time without commensurate benefit for this use case.

**Rationale:** Go gives us a single static binary, excellent standard library networking (HTTP/2, WebSocket, TLS), trivial ARM cross-compilation (`GOARCH=arm64`), and a language that most backend and systems developers can read and contribute to. The garbage collector is not a concern for the agent's workload profile.

**Tradeoffs:** Slightly larger binary than Rust. Less fine-grained control over memory layout. FFI with system libraries (e.g., for Wayland screen capture) requires cgo.

---

## 002: Protocol — JSON over WebSocket over gRPC or Protocol Buffers

**Date:** 2026-07-07

**Context:** The BuzzPi Protocol (BPP) needs to carry commands, events, streaming data (terminal output, screen frames, logs), and control messages between clients and agents. It must be debuggable, extensible, and implementable by third-party clients.

**Decision:** JSON messages over WebSocket as the primary protocol, with optional binary frames for high-throughput data (screen video, file transfers). WebSocket handles bidirectional streaming natively.

**Alternatives considered:**

- **gRPC** — Excellent for service-to-service communication, but significantly harder for third-party clients to implement. Requires code generation, protobuf toolchains, and HTTP/2. Mobile clients have weaker gRPC support. Streaming is possible but complex.
- **MQTT** — Good for IoT telemetry, but designed for pub/sub, not request-response or bidirectional streaming. Would require an additional layer for terminal and screen sessions.
- **Raw TCP** — Too low-level. No built-in TLS, framing, or protocol negotiation.
- **SSH** — Excellent for terminal, but does not natively support custom message types, screen streaming, or plugin APIs.

**Rationale:** JSON over WebSocket is the most accessible protocol for third-party clients. A developer can open a WebSocket connection with a few lines of JavaScript, Python, Kotlin, or Swift and start interacting. JSON is human-readable during development, and binary frames can be layered on for performance-critical paths. WebSocket handles reconnection, TLS, and bidirectional streaming out of the box.

**Tradeoffs:** JSON parsing is slower than protobuf. Message sizes are larger. For screen video frames, we use WebRTC data channels (binary), not JSON.

---

## 003: Screen Streaming — WebRTC over VNC

**Date:** 2026-07-07

**Context:** Remote desktop requires streaming screen frames from the Pi to the client with low latency, adaptive quality, and input feedback. The solution must work over LAN and remote connections.

**Decision:** Build native screen capture into the Agent and stream over WebRTC. Do not wrap a VNC server.

**Alternatives considered:**

- **VNC (TigerVNC/RealVNC)** — Mature protocol, but designed for LAN use. No built-in NAT traversal, no adaptive quality based on network conditions, no native hardware encoding support. Requires the user to install and configure a VNC server, set passwords, and open ports. VNC's RFB protocol is chatty and performs poorly over high-latency links.
- **WayVNC** — Wayland-native VNC server. Would simplify capture but inherits all VNC transport limitations.
- **NoVNC (web-based)** — Requires a browser. Not a native Android experience.

**Rationale:** WebRTC gives us adaptive bitrate, NAT traversal (ICE/STUN/TURN), hardware-accelerated video encoding, and native Android SurfaceView rendering. By building capture into the Agent, we eliminate all user configuration — no VNC server to install, no ports to open, no passwords to manage. Screen traffic routes through the same Connection Engine as everything else.

**Tradeoffs:** Significantly more engineering effort than wrapping VNC. WebRTC signaling adds complexity. Hardware encoding support varies across Pi models (Pi 4's VideoCore IV vs Pi 5's VideoCore VII).

---

## 004: Android UI — Jetpack Compose over XML Views

**Date:** 2026-07-07

**Context:** The Android app needs to be modern, maintainable, and visually consistent with BuzzPi's design language. It must support complex UIs (dashboard, terminal, screen viewer, file manager) and animations.

**Decision:** Jetpack Compose with Material 3.

**Alternatives considered:**

- **XML Views** — Mature and stable, but significantly more verbose. Complex UIs require multiple XML files, adapters, and fragment transactions. Animation and theming are more cumbersome. Harder to maintain consistency across screens.
- **Flutter** — Cross-platform, but not native. Would not feel like a Google-built app. Larger APK, different animation model, separate widget ecosystem.

**Rationale:** Compose is the modern Android standard. It provides declarative UI, built-in Material 3 theming, easy animation APIs, and better development velocity. The BuzzPi brand can be expressed as a Compose theme that applies consistently across all screens.

**Tradeoffs:** Compose is less mature than XML Views. Some advanced platform features (certain accessibility APIs, complex SurfaceView interop for screen streaming) require more work. Minimum API level must be considered.

---

## 005: Plugin Architecture — Separate Processes over In-Process

**Date:** 2026-07-07

**Context:** Plugins extend BuzzPi's capabilities (Docker, Pi-hole, Home Assistant, etc.). They need to be isolated, independently developed, and resilient to crashes.

**Decision:** Plugins run as separate processes managed by the Agent, communicating over a local IPC protocol.

**Alternatives considered:**

- **In-process plugins (shared library/.so)** — Lower overhead and simpler IPC. But a plugin crash takes down the entire Agent. Memory leaks in plugins degrade system stability. Plugin authors must use Go (or cgo), limiting the contributor pool. Restarting a plugin requires restarting the Agent.
- **WebAssembly** — Sandboxed and safe, but limited system access. Would restrict what plugins can do (e.g., Docker plugins need socket access). WASM in Go is still evolving.

**Rationale:** Separate processes provide fault isolation — a crash in the Docker plugin does not affect the terminal session or screen streaming. Plugins can be written in any language. They can be started, stopped, and updated independently. Communication over a well-defined IPC boundary enforces API discipline.

**Tradeoffs:** Higher resource usage per plugin. IPC latency. More complex lifecycle management. Plugin developers must handle serialization.

---

## 006: Connection Engine — Automatic Selection over Manual Mode Switching

**Date:** 2026-07-07

**Context:** Users should never have to think about how they are connected to their Pi. The app must seamlessly handle LAN, VPN, remote, and relay connections without user intervention.

**Decision:** A single Connection Engine that probes transports in priority order and transitions transparently.

**Alternatives considered:**

- **Manual mode selection** — Users pick "Local" or "Remote" or "VPN." Simple to implement, but violates the "no networking knowledge" principle. Users should not need to know whether they are on the same network.
- **Always-cloud** — Route all traffic through the relay. Simple, but introduces latency on LAN and makes offline use impossible.

**Rationale:** Automatic selection is the only approach that satisfies the core principle. The engine tries the fastest available transport first and falls back gracefully. Transition is transparent — a terminal session stays open even if the underlying transport changes from LAN to relay.

**Tradeoffs:** Significantly more complex implementation. The engine must probe transports without introducing noticeable delay. Connection state management requires a robust state machine. Edge cases (split-brain, competing transports) must be handled defensively.

---

## 007: State Machines over Ad-Hoc State Management

**Date:** 2026-07-07

**Context:** Every major subsystem (connection, pairing, screen streaming, file transfer) has distinct states with valid transitions. Without explicit state machines, invalid states and race conditions proliferate.

**Decision:** All major subsystems use explicit state machines with documented states, transitions, and error handling.

**Rationale:** State machines make impossible states impossible. They are testable, documentable, and auditable. A connection can only transition from "Discovering" to "Connecting" to "Authenticating" to "Connected" — never from "Connected" directly to "Discovering" without going through "Disconnected" first.

**Tradeoffs:** More upfront design work. State machines add boilerplate. Not every minor UI state needs a formal machine — judgment is required to choose the right abstraction level.

---

## 008: Database — PostgreSQL over SQLite for Backend

**Date:** 2026-07-07

**Context:** BuzzPi's backend needs to store device registrations, user accounts, and session metadata. It must be reliable, well-understood, and operate with minimal operational complexity.

**Decision:** PostgreSQL.

**Alternatives considered:**

- **SQLite** — Excellent for embedded use, but does not handle concurrent write workloads well. No built-in networking, replication, or role-based access control.
- **MySQL/MariaDB** — Comparable to PostgreSQL, but historically weaker on advanced features (JSON support, indexing, extensions).
- **MongoDB** — Schema-less design does not fit the relational nature of device registrations and user accounts. Higher operational complexity.

**Rationale:** PostgreSQL is mature, feature-rich, and well-understood by the developer community. JSON columns provide flexibility where needed. Row-level security, extensions, and replication support give room to grow without changing databases.

**Tradeoffs:** More resource-intensive than SQLite. Requires a server process. For the agent's local storage on the Pi, we use SQLite (via Go's standard database/sql).

---

## 009: No Kubernetes, No Microservices

**Date:** 2026-07-07

**Context:** The BuzzPi backend must be deployable by a single developer or small team. It should not require a PhD in cluster management to run in production.

**Decision:** Deploy the backend as a single binary (or small number of services) on a single VM or container. Add horizontal scaling only when there is evidence it is needed.

**Alternatives considered:**

- **Kubernetes** — Adds enormous operational complexity. Requires container images, registrations, ingress controllers, service meshes, and persistent volume management. Appropriate for teams of 10+ infrastructure engineers.
- **Microservices** — Premature decomposition. The backend has a natural bounded context (device registry, auth, relay) that fits a monolith well in the beginning.

**Rationale:** Start with a monolith. Extract services when there is a demonstrated need (performance bottleneck, team scaling, independent deploy requirements). This is the approach used by GitHub, Shopify, and many other successful platforms in their early years.

**Tradeoffs:** Less flexibility for independent scaling. Larger deployment unit. More difficult to experiment with new technologies for specific subsystems.

---

## 010: Privacy by Default over Opt-Out Telemetry

**Date:** 2026-07-07

**Context:** BuzzPi connects users to their devices. It inherently handles sensitive data — device access, credentials, usage patterns. Trust is essential.

**Decision:** Zero telemetry by default. No analytics, no crash reporting, no usage tracking unless the user explicitly opts in. Crash reporting and diagnostics are opt-in at initial setup and can be revoked at any time.

**Alternatives considered:**

- **Opt-out telemetry** (industry default) — Collect data unless the user finds and disables the setting. Better data for the project team, but erodes trust. Violates the privacy-first principle.
- **Anonymous aggregated metrics** — Less invasive, but still collects data without explicit consent. Hard to do meaningfully without unique identifiers.

**Rationale:** Trust is BuzzPi's most valuable asset. Users should never wonder what data is being collected. Opt-in crash reporting is valuable for improving stability, but it must be a conscious choice.

**Tradeoffs:** Less data to guide product decisions. Slower identification of crashes and usability issues. We compensate with a responsive issue tracker, user research, and opt-in feedback channels.
