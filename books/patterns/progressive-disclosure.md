# Progressive Disclosure

**Beginners see what they need. Experts can expand. The interface grows with the user.**

## Problem

BuzzPi serves users with widely varying technical skill: a teacher managing 30 classroom Pis, a parent setting up a Pi for their child, and a maker running custom Docker stacks. A single interface cannot serve all of them without either overwhelming beginners or hiding power features from experts. Designing for the lowest common denominator frustrates power users; designing for experts alienates newcomers.

## Solution

Every screen has a default view that shows the essential information and actions for that context. Additional controls are progressively revealed through:

### 1. Default Action, Expanded Details

The device card in the device list shows:
- **Default:** friendly name, status dot (online/offline), temperature indicator
- **Tap to expand:** IP address, uptime, storage, quick actions (terminal, screen, restart)

The user never needs to expand if they only want to check which devices are online. The status is visible at a glance.

### 2. Primary Actions Always Visible, Secondary Actions in Menu

The Workspace action bar shows:
- **Always visible:** Terminal, Screen, Restart
- **In menu:** Shutdown, Update, SSH Key, Diagnostics, Factory Reset

The most common actions (terminal, screen, restart) are one tap away. Destructive or infrequent actions require an extra tap.

### 3. Settings Organization

Settings follow a three-tier structure:

| Tier | Content | Access |
|------|---------|--------|
| Quick settings | Theme, notification toggle | Profile screen, visible immediately |
| Device settings | Friendly name, update channel | Workspace → Device Settings |
| Advanced settings | Network config, debug mode, log level | Device Settings → Advanced (scroll down or tap "Advanced") |

### 4. Contextual Help

When a user encounters a feature for the first time:
- A subtle hint appears below the control ("Restarting will disconnect this device for about 30 seconds")
- The hint disappears after being shown twice
- A "Learn more" link opens the relevant documentation
- Help is always dismissable and never blocks the interface

## User Experience

A first-time user opens BuzzPi and sees "No devices yet. Tap here to pair your first Pi." They pair a device and the Workspace opens with Terminal, Screen, and Status — everything they need. They never see "advanced networking options" or "Docker configuration" until they need it.

Six months later, the same user has 8 devices. They discover they can tap and hold a device card to select multiple devices and run actions in bulk. This was always available — they just didn't need it before.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Power features require more taps | This is intentional. The extra tap confirms deliberate action. Destructive operations (shutdown, factory reset) should never be accidental. |
| Some users never discover advanced features | Discovery is aided by contextual hints and documentation. Users who need advanced features will seek them out. Users who don't need them won't miss them. |
| Progressive disclosure adds UI complexity | The implementation cost is justified by the improved experience across the full skill range. The "advanced" section pattern is well-understood by users. |

## Examples

- Device discovery: "Pair a device" button is prominent; "Enter IP address manually" is under "Need help finding your device?"
- Terminal: basic terminal shows output and input; advanced settings (font, theme, scrollback, confetti mode) are in a settings panel
- Actions: Restart is always visible; Shutdown is in the overflow menu
- GPIO: pin listing is front and center; I2C/SPI protocol settings are in "Advanced"

## Related Patterns

- [Explain, Don't Expose](explain-dont-expose.md): Advanced settings include explanations of what they do
- [Offline First](offline-first.md): Offline indicators are always visible at the top level, even when details are collapsed
