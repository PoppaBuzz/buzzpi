# Automatic Transport

**Choose the best connection without user intervention.**

## Problem

Users should not care whether their device is connecting over LAN, WAN, direct P2P, or relayed through a cloud server. They should not have to configure port forwarding, select a connection mode, or understand the difference between STUN and TURN. Yet the underlying network conditions vary dramatically — the same device may be reachable via P2P at home and only via relay when the user is traveling.

## Solution

The Connection Engine (see engineering book) automatically selects the best transport path. The user sees a single "Connect" action. Everything else is automatic.

### Transport Selection Logic

```
1. Client initiates connection to device
2. Client and device attempt P2P via ICE (STUN)
   ├── Success → Use P2P (lowest latency)
   └── Failure → Attempt TURN relay
       ├── Success → Use relay (higher latency)
       └── Failure → Report "device unreachable"
3. Connection quality is continuously monitored
4. If quality degrades: attempt transport upgrade (relay → P2P if conditions improve)
```

### Transport Reliability

If the initial connection fails:
1. Retry with the same transport (3 attempts, exponential backoff)
2. Try the next transport preference (P2P → relay)
3. If all transports fail: report a human-readable error with recovery suggestions

### Transport Quality Monitoring

While connected, the system monitors:

| Metric | Good | Degraded | Poor |
|--------|------|----------|------|
| Latency (RTT) | <100ms | 100-500ms | >500ms |
| Packet loss | <1% | 1-5% | >5% |
| Throughput | >5 Mbps | 1-5 Mbps | <1 Mbps |

When quality degrades, the system:
1. Adjusts data stream parameters (lower screen quality, less frequent stats)
2. Attempts to re-establish a P2P connection (if currently on relay)
3. Does NOT switch from P2P to relay (relay is always worse than P2P)

### Transport Indication

The current transport is shown in the Workspace status bar as a small indicator:
- Bolt icon: P2P (fast)
- Cloud icon: Relay (slower)
- No icon: Local connection (fastest, always P2P)

The indicator is informational, not actionable. The user cannot change the transport manually.

## User Experience

A user opens BuzzPi at home, connects to their Kitchen Pi on the same network. Connection is instant, latency is 2ms. They don't see any transport indicator because the local connection is so fast it doesn't warrant attention.

Later, they're at a coffee shop, open BuzzPi, and connect to the same device over cellular. The connection takes 3 seconds (ICE negotiation + relay setup). A small cloud icon appears in the status bar. Terminal feels slightly laggy (100ms latency) but screen streaming adapts to a lower quality automatically.

The user never configures anything. The transport "just works."

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Automatic selection may not choose the optimal path for edge cases | The ICE algorithm is well-tested and optimal for 99%+ of cases. The remaining 1% (e.g., asymmetric routing, network policies) are better handled by automatic fallback than by user configuration. |
| Connection may fail silently (user doesn't know why) | The transport indicator provides feedback. Failure messages are human-readable and suggest recovery actions. |
| Relay adds cost (TURN server bandwidth) | TURN bandwidth is the primary operational cost of BuzzPi Cloud. Managed by rate limiting and quality profiles. |

## Examples

- Device connection: single tap connects, regardless of network topology
- Screen streaming: adaptive quality based on transport latency
- Terminal: automatic latency compensation (keystroke echo)
- File transfer: transfer speed is not artificially limited but shows progress feedback

## Related Patterns

- [Offline First](offline-first.md): Transport selection determines whether remote access is possible
- [Capability Detection](capability-detection.md): Some transports may not support all capabilities (relay may limit screen quality)
- [Best Effort Reconnect](best-effort-reconnect.md): When transport fails, reconnection uses the same automatic selection logic
