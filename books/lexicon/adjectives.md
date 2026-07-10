# Adjectives

These adjectives describe states, conditions, and relationships in the BuzzPi Platform. Every state machine, status indicator, and notification uses these terms consistently.

---

## Connection States

| Term | Meaning |
|------|---------|
| **Connected** | A session is active. Data can be exchanged. |
| **Discoverable** | Device is visible on the network but not paired. |
| **Trusted** | Device has been paired. Trust relationship exists. |
| **Available** | Device is online and ready to connect. |
| **Offline** | Device is unreachable. Last known state is preserved. |
| **Unreachable** | Device was recently online but cannot be contacted now. |

---

## Device Health States

| Term | Meaning | Status Color |
|------|---------|-------------|
| **Healthy** | All metrics within normal range. | Green |
| **Attention** | A non-critical condition needs review. | Blue / Info |
| **Warning** | A metric is approaching a threshold. | Yellow |
| **Critical** | A threshold has been exceeded. Immediate action needed. | Red |
| **Inactive** | Device has not reported in 7+ days. | Gray |
| **Unknown** | Device state cannot be determined. | Gray / Dashed |

---

## Operational States

| Term | Meaning |
|------|---------|
| **Updating** | Device is installing an update. |
| **Busy** | Device is under load. Operations may be slower. |
| **Degraded** | One or more capabilities are unavailable. |
| **Compatible** | Device supports a given capability or protocol version. |
| **Incompatible** | Device does not support a capability or protocol version. |

---

## Status Color System

The status color system uses semantic names, not color names. Accessibility tools and themes map semantic names to actual colors.

| Semantic Name | Default Color | Use |
|---------------|--------------|-----|
| `healthy` | Green | All normal, no action needed |
| `attention` | Blue | Information available, non-critical |
| `warning` | Yellow/Yellow-orange | Approaching threshold, review suggested |
| `critical` | Red | Threshold exceeded, action required |
| `inactive` | Gray | No recent activity, device may be offline |
| `unknown` | Gray/dashed | State cannot be determined |

These status values are used across all clients, notifications, and APIs. A "Warning" status in the Android app has the same meaning and visual treatment as a "Warning" status in the CLI.

---

## Adjective Usage Map

| UI Context | Canonical Adjective |
|------------|-------------------|
| Device is online and reachable | Available |
| Device has active session | Connected |
| Device has been paired | Trusted |
| Device is on network but not paired | Discoverable |
| Device has no connectivity | Offline |
| Device was connected but lost | Unreachable |
| All metrics normal | Healthy |
| Non-critical issue | Attention |
| Approaching limit | Warning |
| Limit exceeded | Critical |
| No data for 7+ days | Inactive |
| State undetermined | Unknown |
