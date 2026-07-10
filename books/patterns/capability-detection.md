# Capability Detection

**Detect what the device supports. Adapt the UI automatically.**

## Problem

A Raspberry Pi 5 with a Camera Module 3 has different capabilities than a Pi Zero 2W with a temperature sensor. A device running Raspberry Pi OS Desktop has different capabilities than one running Raspberry Pi OS Lite. The client cannot assume what any given device supports, and the user should not have to configure which features are available.

## Solution

Every device advertises its capabilities during connection establishment. The client adapts the Workspace UI to match.

### Capability Categories

| Category | Examples | UI Impact |
|----------|----------|-----------|
| Built-in services | terminal, screen, files | Core tabs always visible (terminal always works) |
| Hardware features | gpio, camera, spi, i2c | Tabs/controls shown only if present |
| Software features | docker, services, logs | Tabs shown if compatible software is installed |
| Extensions | docker-manager, gpio-control | Tabs shown after extension is installed and running |

### UI Adaptation Rules

1. **Core services** (terminal) are always available on any device running the Runtime
2. **Hardware-dependent services** (camera, gpio) are shown only when detected
3. **Software-dependent services** (docker) are shown only when installed
4. **Extension services** are shown only when the extension is active

When a capability is missing, the UI does not show a disabled button or a "not available" tab. It simply does not show the tab at all. This prevents the user from wondering why something is grayed out.

### Capability Changes During Session

Capabilities can change during a session (a camera is plugged in, an extension is installed). The device sends a capability update event, and the client updates the UI in real time:

```
Before (camera disconnected):
  [Terminal] [Screen] [Files]

Camera plugged in → capability update received

After (camera connected):
  [Terminal] [Screen] [Camera] [Files]
```

The new tab fades in over 300ms. No reload, no user action required.

## User Experience

A user pairs two devices. One is a Pi 5 with a camera, Docker, and a desktop environment. The other is a Pi Zero 2W running Lite OS. The Workspace for each device shows only what that device supports:

**Pi 5 Workspace:** Terminal, Screen, Files, Docker, Camera, Services
**Pi Zero Workspace:** Terminal, (Screen not available — no desktop), Files, GPIO

The user never sees a disabled "Screen" tab on the Pi Zero. They simply don't see the tab.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| No visibility into missing features | Users may not know their device could support more (e.g., they have a camera but it's disconnected). Tradeoff accepted — clutter of disabled controls is worse. |
| Dynamic UI can be jarring | Tabs appearing/disappearing mid-session could confuse the user. Mitigated by: (a) fade-in/out animation, (b) most capabilities are static (determined at connection time), (c) only hotplug events (camera, USB) cause mid-session changes. |
| Capability detection adds latency | Detection happens during connection establishment, adding ~100ms to initial connection time. Acceptable because it eliminates configuration friction. |

## Examples

- Workspace tabs: dynamically generated based on device capabilities
- Quick actions: "Restart" is always available; "Prune Docker" appears only if Docker extension is installed
- Status indicators: temperature shown only if sensor is available; storage shown only if readable
- GPIO pin list: dynamically populated from device's actual gpiochip

## Related Patterns

- [Progressive Disclosure](progressive-disclosure.md): Once capabilities are detected, the UI shows them at the appropriate detail level
- [Automatic Transport](automatic-transport.md): Transport selection is also capability-driven
