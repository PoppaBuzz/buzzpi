# State Machines

Correctness in BuzzPi comes from well-defined state machines. Every component with mutable state is modeled as a state machine with explicit states, transitions, and actions.

---

## Device State Machine

The fundamental state machine. Every device tracked by BuzzPi is in one of these states at all times.

```
                    ┌─────────────────────────────────────┐
                    │                                     │
                    v                                     │
 ┌────────┐    ┌──────────┐    ┌─────────┐    ┌──────────┐
 │        │    │          │    │         │    │          │
 │  NEW   │───▶│ PAIRING  │───▶│ ONLINE  │───▶│ OFFLINE  │
 │        │    │          │    │         │    │          │
 └────────┘    └──────────┘    └─────────┘    └──────────┘
    │                                              │
    │                                              │
    │           ┌──────────┐                       │
    │           │          │                       │
    └──────────▶│ UNPAIRED │◀──────────────────────┘
                │          │
                └──────────┘
```

| State | Meaning |
|-------|---------|
| NEW | Device discovered on network but not yet paired |
| PAIRING | User initiated pairing, awaiting code confirmation |
| ONLINE | Device is paired and Runtime is connected |
| OFFLINE | Device is paired but Runtime is disconnected |
| UNPAIRED | Device was removed from user's account |

### Transitions

| From | To | Trigger | Actions |
|------|----|---------|---------|
| NEW | PAIRING | User taps "Pair" | Generate pairing code, show instructions |
| NEW | UNPAIRED | User dismisses | Clean up discovery entry |
| PAIRING | ONLINE | Code verified, device connected | Save device identity, create workspace, send notification |
| PAIRING | NEW | Code expired / rejected | Reset pairing state, notify user |
| PAIRING | UNPAIRED | User cancels pairing | Clean up partial state |
| ONLINE | OFFLINE | Runtime connection lost (no heartbeat for 30s) | Mark offline, trigger notification if grace period exceeded |
| OFFLINE | ONLINE | Runtime reconnects | Mark online, update last seen |
| OFFLINE | UNPAIRED | User unpairs device | Revoke device identity, clear session keys |
| ONLINE | UNPAIRED | User unpairs device | Disconnect Runtime, revoke identity, clear keys |

### Grace Period

When a device transitions from ONLINE to OFFLINE, a 2-minute grace period begins. If the device reconnects within the grace period, the transition is reversed without user notification. This prevents false offline alerts during brief network interruptions.

---

## Pairing State Machine

The pairing flow has its own internal state machine, distinct from the device state machine above.

```
 ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌───────────┐
 │          │    │          │    │          │    │           │
 │ WAITING  │───▶│  CODE    │───▶│ VERIFY  │───▶│ COMPLETED │
 │          │    │ SHOWN    │    │          │    │           │
 └──────────┘    └──────────┘    └──────────┘    └───────────┘
      │               │              │
      │               │              │
      │               v              v
      │          ┌──────────┐   ┌──────────┐
      │          │          │   │          │
      └─────────▶│ EXPIRED  │   │ FAILED   │
                 │          │   │          │
                 └──────────┘   └──────────┘
```

| State | Meaning | Timeout |
|-------|---------|---------|
| WAITING | App scanning for devices, user hasn't selected one | None |
| CODE SHOWN | Device selected, pairing code displayed to user | 5 minutes |
| VERIFY | User entered code, waiting for device confirmation | 30 seconds |
| COMPLETED | Pairing successful | — |
| EXPIRED | User didn't enter code within timeout | — |
| FAILED | Code incorrect or device rejected | — |

### Pairing Protocol

1. Device publishes its identity to the local network via mDNS
2. App discovers device, user selects it
3. Device generates a time-limited pairing code (6 alphanumeric characters)
4. Display shows code; Runtime waits for verification
5. App sends code to Relay Server
6. Relay Server verifies code with device's Runtime
7. On success: device and app exchange public keys
8. On failure: device generates new code, user retries

---

## Connection State Machine

The connection between app and Runtime (direct or relayed) follows this state machine.

```
 ┌──────────┐    ┌──────────┐    ┌───────────┐    ┌──────────┐
 │          │    │          │    │           │    │          │
 │  IDLE    │───▶│NEGOTIATE│───▶│ CONNECTED │───▶│ CLOSED   │
 │          │    │          │    │           │    │          │
 └──────────┘    └──────────┘    └───────────┘    └──────────┘
                      │                │
                      │                │
                      v                v
                 ┌──────────┐    ┌──────────┐
                 │          │    │          │
                 │  FAILED  │    │ RECONNECT│──▶ to NEGOTIATE
                 │          │    │          │
                 └──────────┘    └──────────┘
```

| State | Meaning |
|-------|---------|
| IDLE | No connection active |
| NEGOTIATE | ICE/WebRTC negotiation in progress |
| CONNECTED | Bidirectional communication active |
| CLOSED | Intentional connection teardown |
| FAILED | Connection could not be established |
| RECONNECT | Connection lost, attempting reconnection |

### Connection Preference

1. **Direct (P2P):** App connects directly to Runtime via ICE (STUN only). Preferred.
2. **Relay (TURN):** If direct connection fails (NAT, CGNAT), traffic routes through Relay Server.
3. **Retry:** Exponential backoff with jitter (1s, 2s, 4s, 8s, max 30s).

---

## Service State Machine

Tracks the state of a service running on a device (managed via extensions).

```
 ┌──────────┐    ┌──────────┐    ┌───────────┐    ┌───────────┐
 │          │    │          │    │           │    │           │
 │ STOPPED  │───▶│ STARTING│───▶│ RUNNING   │───▶│ STOPPING  │
 │          │    │          │    │           │    │           │
 └──────────┘    └──────────┘    └───────────┘    └───────────┘
      ▲               │                │               │
      │               │                │               │
      │               │                v               │
      │               │           ┌──────────┐         │
      │               │           │          │         │
      └───────────────┴───────────│ FAILED  │◀────────┘
                                  │          │
                                  └──────────┘
```

| State | Meaning |
|-------|---------|
| STOPPED | Service is installed but not running |
| STARTING | Service launch initiated, awaiting health check |
| RUNNING | Service is running and passing health checks |
| STOPPING | Graceful shutdown initiated |
| FAILED | Service exited unexpectedly or health check failed |

### Health Checks

- Services expose a health endpoint (or process exists check)
- Runtime runs health checks every 30 seconds for RUNNING services
- After 3 consecutive failed health checks → transition to FAILED
- Failed services trigger a notification (configurable)
- Automatic restart is available for non-critical services (configurable)
