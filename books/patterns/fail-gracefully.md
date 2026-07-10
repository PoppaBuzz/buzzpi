# Fail Gracefully

**Degrade features before crashing. No silent failures.**

## Problem

In distributed systems, failures are inevitable. A device goes offline. A service crashes. A network partition occurs. The worst possible response is a crash, a frozen UI, or (most insidious) a silent failure where the user thinks something worked but it didn't. BuzzPi must anticipate failure at every level and present a working (if reduced) experience to the user.

## Solution

### Degradation, Not Crash

Every component in BuzzPi has a defined degraded mode:

| Component | Normal | Degraded | Failed |
|-----------|--------|----------|--------|
| Device List | All devices shown with live status | Some devices missing (offline, not yet discovered) | "Unable to load devices. Check connection." |
| Terminal | Full ANSI, 256 color, true color | Reduced color (16 color), no true color | "Terminal unavailable. Device may be offline." |
| Screen | Full resolution, 30fps | Low resolution, 5fps | "Screen streaming not available on this device." |
| File Transfer | Full speed | Throttled (relay bandwidth limit) | "Transfer failed. Retry." |
| Notifications | Push + in-app | In-app only (FCM unavailable) | Notifications disabled |
| Cloud Features | Full | Reduced (rate limited) | "Cloud features unavailable. Local devices still work." |

### Degradation Rules

1. **No silent failures.** If a feature cannot work, the user knows why.
2. **Degrade before failing.** Try a lower quality mode before giving up.
3. **Failure is actionable.** Every failure message includes a recovery suggestion.
4. **Local survives remote.** If the cloud is down, local features continue working.

### Degradation Hierarchy

```
Screen streaming request
├── Try: Full quality (1080p, 30fps)
├── Degrade to: Medium quality (720p, 15fps) [network constrained]
├── Degrade to: Low quality (480p, 10fps) [heavily constrained]
├── Degrade to: Still frames only (1fps) [extreme constraint]
└── Fail: "Screen streaming not available" [codec or hardware issue]
```

Each degradation level shows a brief indicator: "Lowering quality due to network conditions" (auto-dismiss after 3s).

### Circuit Breaker

For cloud services, BuzzPi uses a circuit breaker pattern:

1. **Closed:** Normal operation. Cloud requests proceed.
2. **Open:** After 5 consecutive failures, stop trying for 30 seconds.
3. **Half-open:** After 30 seconds, try a single request.
4. If it succeeds → close the circuit.
5. If it fails → open the circuit again (2x backoff, max 5 minutes).

When the circuit is open, the UI shows "Cloud features temporarily unavailable" and local features work normally.

### Error Recovery

When a failure is detected, the system automatically attempts recovery:

| Failure | Recovery | Timeout |
|---------|----------|---------|
| Device goes offline | Wait in grace period (2 min) | 2 min |
| Connection drops | Reconnect with backoff (1-30s) | 3.5 min (10 attempts) |
| Push notification fails | Retry with backoff | 3 attempts |
| Extension crashes | Auto-restart (up to 3 times) | 5 min cooldown |
| Runtime update fails | Rollback to previous version | Immediate |

## User Experience

A user tries to open a screen stream. Their network is slow. Instead of a spinner that never resolves, they see: "Lowering quality due to network speed" and after 2 seconds, a lower-resolution but working stream appears.

The user tries to open a screen stream on a headless device. Instead of a spinning wheel, they immediately see: "Screen streaming is not available on this device. It may not have a graphical desktop installed."

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Degraded features may confuse users | The degraded indicator is subtle and informative. Users understand "lower quality" more than "failed." |
| Circuit breaker may block legitimate requests | The 2x backoff means longer outages for transient problems. Acceptable because: (a) cloud features are secondary to local features, (b) the backoff prevents cascading failures. |
| Degradation adds code complexity | Every component needs multiple operating modes. This is the cost of reliability in a distributed system. |

## Examples

- Device list: shows cached state when cloud is unreachable
- Terminal: degrades from ANSI 256 to ANSI 16 when bandwidth is low
- Screen: degrades quality before failing
- Notifications: graceful fallback from push to in-app to disabled

## Related Patterns

- [Explain, Don't Expose](explain-dont-expose.md): Degradation messages follow the error template
- [Best Effort Reconnect](best-effort-reconnect.md): Handles reconnection after failures
- [Offline First](offline-first.md): Local features are always available regardless of cloud state
