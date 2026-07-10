# Audit by Default

**Every action is logged. Every state change is recorded.**

## Problem

When something goes wrong with a device — a service crashes, a temperature spikes, an update fails — the user needs to understand what happened and when. Without audit logging, debugging requires SSH access and manual log spelunking. BuzzPi should provide audit trails automatically, without the user having to enable them.

## Solution

Every significant event in the BuzzPi ecosystem is recorded in an audit log.

### What Is Logged

| Category | Events Logged | Retention |
|----------|---------------|-----------|
| Connection | Connected, disconnected (with reason), reconnected | 90 days |
| Pairing | Paired, unpaired, pairing attempted, pairing failed | Lifetime |
| Actions | Action initiated, completed, failed (with action type and user) | 90 days |
| State changes | Online → offline, offline → online, service started/stopped | 30 days |
| Configuration | Config changed, feature enabled/disabled | 90 days |
| Errors | Error code, message, context, recovery action | 30 days |
| Updates | Update checked, downloaded, installed, failed, rolled back | 90 days |
| Extensions | Extension installed, updated, removed, crashed | 90 days |

### What Is NOT Logged

For privacy and security reasons, these events are NOT logged:

- Terminal session content (keystrokes, command output)
- Screen stream content (video frames)
- File content (uploaded/downloaded file data)
- User passwords or authentication tokens
- Personally identifiable information beyond email address

### Log Format

Each audit log entry follows a consistent format:

```json
{
  "timestamp": "2026-07-07T12:00:00Z",
  "event_id": "evt_abc123",
  "category": "action",
  "event": "device.restart",
  "actor": {
    "type": "user",
    "id": "user_abc123",
    "client_id": "client_def456",
    "client_name": "Sarah's Phone"
  },
  "target": {
    "type": "device",
    "id": "dev_789012",
    "name": "Kitchen Pi"
  },
  "result": "completed",
  "details": {
    "duration_ms": 45000,
    "method": "software"
  }
}
```

### Audit Log Access

Users can access the audit log from:
1. **Device info screen:** Recent events for a specific device (last 24 hours)
2. **Device settings → Audit Log:** Full log for a device
3. **User profile → Activity:** All actions across all devices
4. **Notification history:** Important events already shown as notifications

### Audit Log in the App

The audit log is presented as a timeline view:

```
Today, 12:00 PM ──────────────────────
  Kitchen Pi: Restarted by Sarah (completed in 45s)

Today, 11:45 AM ──────────────────────
  Kitchen Pi: Temperature reached 72°C (warning threshold)
  Workshop Pi: Went offline (reason: network lost)

Today, 11:30 AM ──────────────────────
  Docker extension updated to v1.0.2 on Kitchen Pi
  Workshop Pi: Storage at 82% (warning threshold)
```

### Delete on Unpair

When a device is unpaired, its audit log is:
1. Deleted from the user's account (90-day logs)
2. Preserved in anonymized form for aggregate metrics
3. Deleted from the device (local audit log)

## User Experience

A user notices their device restarted overnight. They check the audit log and see: "Kitchen Pi automatically restarted after update installation (v0.1.2 → v0.1.3)." They understand what happened and why. No SSH, no log files, no confusion.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Storage cost for audit logs | Audit logs are small (text, structured). 90 days for a typical user is under 1MB. Cheap storage. |
| Privacy implications of logging everything | No sensitive content is logged. Event metadata (what happened, when, by whom) is not private. |
| Performance impact of logging | Log writes are async and batched. Performance impact is negligible (<1ms per event). |

## Examples

- Device restart: logged with initiator, duration, and result
- Failed action: logged with error code and recovery suggestion
- Offline detection: logged with grace period start and notification sent
- Extension crash: logged with crash count and auto-restart attempt

## Related Patterns

- [Explain, Don't Expose](explain-dont-expose.md): Audit events use human-readable descriptions
- [Privacy by Design](privacy-by-design.md): Audit logs exclude sensitive content
