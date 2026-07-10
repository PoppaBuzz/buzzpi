# Engineering Design Constraints

These constraints define the technical boundaries of the BuzzPi Platform. They eliminate architectural options that violate the project's principles or operational requirements.

- The BuzzPi Runtime runs on Raspberry Pi OS Lite. It does not require a desktop environment.
- The Runtime is distributed as a single static Go binary with minimal dependencies.
- The Runtime integrates with systemd. It starts on boot and restarts on failure.
- ARMv6 through ARM64 are supported where practical. The reference architecture is ARM64.
- The Runtime supports offline LAN operation. No internet connection is required for local use.
- The Android client's minimum API level is determined by the oldest version with meaningful market share, updated annually.
- Network transitions (LAN to remote, Wi-Fi to cellular) do not terminate active sessions.
- The backend starts as a single Go service with PostgreSQL. No Redis, no microservices, no container orchestration.
- Every major subsystem has an explicit state machine. Invalid states are unrepresentable.
- All communication is encrypted in transit. End-to-end encryption is the default for all session data.
- The protocol supports version negotiation. Breaking changes require a deprecation window.
- Plugins run as separate processes. A plugin crash never takes down the Runtime.
- No telemetry, analytics, or usage tracking is compiled into release builds. Opt-in crash reporting is the only instrumentation, and it is off by default.
