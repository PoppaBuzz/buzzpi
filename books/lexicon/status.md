# Status

Status values define the canonical state labels used across the entire BuzzPi Platform. Every client, notification, API response, and UI component uses these values.

## Device Status

| Status | Definition | Color | Icon |
|--------|-----------|-------|------|
| Available | Online and ready | Healthy | Circle check |
| Connected | Active session | Healthy | Double circle |
| Discoverable | Visible, not paired | Attention | Eye |
| Offline | No connectivity | Inactive | Circle slash |
| Unreachable | Recently disconnected | Warning | Broken link |
| Updating | Installing update | Attention | Arrow circle |
| Busy | Under load | Warning | Gauge |
| Degraded | Partial capability | Warning | Triangle exclamation |
| Inactive | No data 7+ days | Inactive | Moon |
| Unknown | Cannot determine | Unknown | Question |

## Connection Status

| Status | Definition |
|--------|-----------|
| Direct | Connected over LAN |
| Tailscale | Connected via Tailscale/MagicDNS |
| Relay | Connected via cloud relay |
| WebRTC | Connected via peer-to-peer |

## Capability Status

| Status | Definition |
|--------|-----------|
| Available | Capability is present and working |
| Unavailable | Capability is not present on this device |
| Degraded | Capability is present but partially working |
| Loading | Capability status is being determined |

## Action Status

| Status | Definition |
|--------|-----------|
| Pending | Action is queued |
| Running | Action is executing |
| Completed | Action finished successfully |
| Failed | Action finished with error |
| Cancelled | Action was cancelled by user |
| Timed Out | Action exceeded its time limit |

## Notification Priority

| Priority | Definition | Treatment |
|----------|-----------|-----------|
| Critical | Immediate action required | Full screen + sound + persistent |
| Warning | Review suggested | Notification + sound |
| Info | For awareness only | Silent notification |
| Success | Operation completed | Brief toast |
