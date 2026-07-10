# One Action Per Step

**Each action does exactly one thing. Compose actions for complex workflows.**

## Problem

Tools that bundle multiple operations into a single button create confusion when something goes wrong. "Update" might mean: check for updates, download, verify, install, reboot. If it fails mid-way, what state is the device in? Was the update installed? Is the device rebooting? Users should never have to wonder what a single action does or what state it leaves their device in.

## Solution

Every action in BuzzPi does exactly one atomic thing. The action name tells the user exactly what will happen.

### Action Atomicity

| Action | What It Does | Composable? |
|--------|-------------|-------------|
| `Restart` | Restarts the device | No (terminal action) |
| `Check for Updates` | Checks if updates are available | Yes → `Download Updates` |
| `Download Updates` | Downloads available updates | Yes → `Install Updates` |
| `Install Updates` | Installs downloaded updates | Yes → `Restart` |
| `Restart` | Restarts the device | No (terminal action) |

A single "Update All" button is NOT provided. Instead, the actions are sequential and the user sees the state at each step:

1. **Check for Updates** → "3 updates available" → **Download Updates** → "Downloaded (45MB)" → **Install Updates** → "Installed. Restart recommended." → **Restart**

Each step is a separate action with its own button, its own progress indicator, and its own result message. The user can stop at any step.

### Composition

For power users, actions can be composed:

- **Scheduled actions:** "Check for updates daily at 3 AM. If updates found, download and install. Restart if required."
- **Batch actions:** "Restart Kitchen Pi AND Workshop Pi" (individual restarts, coordinated from the UI)
- **Action chains:** "Restart Docker, then check container health, then notify me"

Composition is explicit. The user sees the steps and can cancel between them.

### Action State

Every action reports its state clearly:

| State | Meaning | Display |
|-------|---------|---------|
| Available | Action can be triggered | Enabled button |
| Running | Action is in progress | Button → progress indicator |
| Completed | Action finished successfully | Checkmark (auto-dismiss after 2s) |
| Failed | Action encountered an error | Error message with recovery |
| Blocked | Prerequisites not met | Disabled button with explanation |

## User Experience

A user taps "Restart Docker" on their device. The button turns into a spinner. After 3 seconds, a checkmark appears, then the button returns to "Restart Docker." The user knows exactly what happened: Docker restarted successfully.

A different user wants to update their device. They tap "Check for Updates." After 5 seconds, it reports "3 updates available." They tap "Download Updates." Progress appears: "12MB of 45MB." When downloads complete, "Install Updates" becomes available. They tap it. Installation completes in 10 seconds. "Restart recommended" appears. They tap "Restart." Device restarts. Each step was atomic, visible, and cancellable.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| More taps for complex workflows | Each tap confirms intent. An "Update All" button would hide intermediate states and create confusion on failure. |
| State tracking between steps | The client tracks action states and enables/disables dependent actions. This adds UI complexity but prevents errors. |
| Some actions cannot be atomic (Restart is terminal) | Terminal actions return a result immediately ("Restart initiated") because the device cannot report after it shuts down. |

## Examples

- Device actions: Restart, Shutdown, Update are separate buttons with clear labels
- Service actions: Start, Stop, Restart are separate (no combined "Manage" button)
- File actions: Delete shows a confirmation dialog with the file name; no "delete and confirm" ambiguity
- Pairing: each step in the pairing flow is a distinct action (Discover, Select, Enter Code, Confirm)

## Related Patterns

- [Progressive Disclosure](progressive-disclosure.md): Composed actions are an advanced feature hidden from beginners
- [Explain, Don't Expose](explain-dont-expose.md): Each action's failure is explained independently
