# Best Effort Reconnect

**Never lose a session. Reconnect transparently across network changes.**

## Problem

Mobile devices change networks frequently: WiFi to cellular, one access point to another, airplane mode on and off. These transitions should not interrupt an active session. The user should not have to re-tap a device after walking from their desk to the kitchen. Sessions should survive network changes transparently.

## Solution

The connection engine monitors connection health and automatically reconnects when a connection drops, with no user action required.

### Reconnection Strategy

```
┌──────────┐    Connection lost
│          │──────────────────────────────┐
│ ACTIVE   │                              │
│ SESSION  │                              ▼
│          │                       ┌──────────────┐
└──────────┘                       │              │
                                   │ RECONNECTING │
                                   │              │
                                   └──────────────┘
                                    │          │
                           ┌────────┘          └────────┐
                           ▼                            ▼
                    ┌──────────────┐            ┌──────────────┐
                    │              │            │              │
                    │ RECONNECTED  │            │ FAILED       │
                    │ (same        │            │ (exhausted   │
                    │  session)    │            │  retries)    │
                    └──────────────┘            └──────────────┘
                           │                            │
                           ▼                            ▼
                    ┌──────────────┐            ┌──────────────┐
                    │  ACTIVE      │            │  CLOSED      │
                    │  (resumed)   │            │  (new connect│
                    └──────────────┘            │   required)  │
                                                └──────────────┘
```

### Reconnection Parameters

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Initial retry delay | 1 second | Fast reconnection for transient drops |
| Backoff multiplier | 2x | Exponential backoff (1s, 2s, 4s, 8s, 16s) |
| Maximum delay | 30 seconds | Cap to prevent long waits |
| Jitter | ±500ms | Prevents thundering herd |
| Maximum attempts | 10 | ~3.5 minutes of reconnection before giving up |
| Session timeout | 24 hours | Session tokens remain valid across reconnections |

### What Survives a Reconnection

| State | Survives? | Mechanism |
|-------|-----------|-----------|
| Terminal session | Yes (up to 5min gap) | Terminal process persists on device; buffered output replay |
| Screen stream | No | Must be restarted (WebRTC stream is per-connection) |
| File transfer | No | Must be restarted (transfer state lost with data channel) |
| Service watching | Yes | Device continues health checks; events are queued |
| GPIO state | Yes | Pin states persist on device (not session-dependent) |

### Terminal Reconnection

The terminal is the most important state to preserve:

1. When connection drops, the PTY process on the device continues running
2. Terminal output during disconnection is buffered (up to 1MB)
3. On reconnection, the buffered output is replayed to the client
4. The client shows a brief indicator: "Reconnected — 12 lines of output while you were away"

This means a user running a long `apt upgrade` can walk away from their phone, come back, and see the output that happened while they were disconnected.

## User Experience

A user is connected to their device via terminal, running a long compilation. They walk from their desk to the kitchen, and their phone switches from WiFi to cellular. The terminal briefly shows "Reconnecting…" in a subtle banner. Three seconds later, the banner disappears and the terminal continues as if nothing happened. The compilation output from the 3-second gap fills in.

The user does not lose their terminal session, does not re-tap anything, and does not lose their place.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Terminal session held open during disconnection | The PTY process continues running and consuming memory. Mitigated by: (a) 5-minute buffer limit, (b) active session limit per device (10 max), (c) stale session cleanup after 5 minutes. |
| Buffering terminal output may lose data | The buffer is bounded at 1MB. If output exceeds this during disconnection, oldest lines are dropped. The user is informed: "(output truncated — 50 lines omitted)" |
| Reconnection may fail silently | If reconnection fails after all retries, the session is closed and the user sees a final message. The user must reconnect manually. This is rare (network truly dead, not transient). |

## Examples

- Terminal: reconnection preserves the PTY and replays buffered output
- Screen streaming: stops on disconnection, restarts on reconnection (brief "Reconnecting screen…" indicator)
- File transfer: fails on disconnection, user must re-initiate (save option: "Resume transfer" if partial transfer is detectable)
- GPIO monitoring: state is preserved and resumes on reconnection

## Related Patterns

- [Automatic Transport](automatic-transport.md): Reconnection uses the same transport selection logic (may switch from P2P to relay)
- [Offline First](offline-first.md): Reconnection handles the local→offline→local transition
- [Explain, Don't Expose](explain-dont-expose.md): "Reconnecting…" is shown in a subtle, non-alarming way
