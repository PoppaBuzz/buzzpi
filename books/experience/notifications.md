# Notifications

Notifications in BuzzPi are designed to be useful without being intrusive. They inform the user of important events without demanding immediate attention unless the situation requires it.

## Notification Categories

### Device Events

| Event | Priority | Content |
|-------|----------|---------|
| Device went offline | Info | "Kitchen Pi went offline" |
| Device came online | Info | "Kitchen Pi is available" |
| High temperature | Warning | "Kitchen Pi temperature is 78°C" |
| Low storage | Warning | "Kitchen Pi has 5% storage remaining" |
| Service failure | Warning | "Docker failed on Kitchen Pi" |
| Update available | Info | "A new Runtime version is available" |
| Update completed | Info | "Kitchen Pi updated to v0.2.5" |
| Pairing requested | Action | "A new device wants to pair: Garage Pi" |
| Pairing expired | Info | "Pairing request for Garage Pi expired" |

### Action Results

| Event | Priority | Content |
|-------|----------|---------|
| Action completed | Success | "Docker restarted on Kitchen Pi" |
| Action failed | Warning | "Failed to restart Docker on Kitchen Pi" |
| Automation triggered | Info | "Automation 'Nightly Backup' started" |
| Automation completed | Success | "Nightly Backup completed" |
| Automation failed | Warning | "Nightly Backup failed: storage full" |

## Notification Behavior

| Priority | Sound | Vibration | Persistent | Dismiss |
|----------|-------|-----------|------------|---------|
| Critical | Yes | Yes | Yes | Only after action taken |
| Warning | Yes | Silent | No | Swipe to dismiss |
| Info | Silent | Silent | No | Swipe to dismiss |
| Success | Silent | Silent | No | Auto-dismiss after 4s |

## Notification Settings

Users control notification behavior per category:

| Setting | Options | Default |
|---------|---------|---------|
| Device offline | On / Off | On |
| High temperature | On / Off | On |
| Low storage | On / Off | On |
| Service failure | On / Off | On |
| Updates | On / Off | On |
| Automation results | On / Off | Off |
| Sound | On / Off | Off |
| Vibration | On / Off | Off |

## Notification Design

- Notifications use the device's friendly name as the title
- The body uses one sentence — no elaboration
- Critical notifications include a primary action button ("View Device", "Restart Service")
- Notifications show the device's current status icon
- Grouped by device — expanding a notification shows all recent events for that device

## Do Not Disturb

BuzzPi respects the system's Do Not Disturb settings. Critical notifications break through only if the user has configured BuzzPi as a priority app in system settings.
