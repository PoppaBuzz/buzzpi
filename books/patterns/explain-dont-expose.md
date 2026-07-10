# Explain, Don't Expose

**Translate technical errors into human language.**

## Problem

Device management produces deeply technical errors: "Connection refused (errno 111)", "EACCES: permission denied", "SSH Error Code 255", "ETIMEDOUT", "journalctl: Cannot open /var/log/journal". These messages are meaningless or alarming to non-technical users, and even technical users deserve better than a raw error code. Error messages should tell the user what happened and what to do about it — not expose internal system state.

## Solution

Every error condition in BuzzPi passes through an error translation layer that converts technical errors into human-readable messages following a three-part template:

### Error Template

```json
{
  "user_message": {
    "what": "Kitchen Pi stopped responding.",
    "why": "It may have lost power or disconnected from your Wi-Fi network.",
    "action": "Check that it's plugged in and connected to the network. If the issue persists, try restarting your router."
  }
}
```

**What** — One sentence, plain language, no jargon. Describes what the user needs to know.

**Why** — One or two sentences describing the most likely cause. Never a list of possibilities.

**Action** — One or two concrete steps the user can take, ordered by likelihood of success.

### Error Translation Rules

| Raw Error | User Message |
|-----------|--------------|
| ECONNREFUSED | "Kitchen Pi is not accepting connections. It may be starting up or experiencing a temporary issue. Wait a moment and try again." |
| ETIMEDOUT | "Kitchen Pi took too long to respond. It may be busy or having a network issue. Check your connection and try again." |
| EACCES (file) | "BuzzPi doesn't have permission to access this file. Check the file's permissions and try again." |
| ENOSPC | "Kitchen Pi is out of storage space. Free up space by deleting unneeded files before trying again." |
| PAIRING_CODE_INVALID | "The pairing code is incorrect. Your device should show a new code. Try again with the new code." |
| VERSION_MISMATCH | "Your app needs an update to connect to this device's Runtime. Update your app and try again." |

### Error Logging

The original technical error is always logged to the device's debug log (never shown in the UI). The user can access debug logs from Settings → Advanced → View Logs, but they are off by default and hidden from the main interface.

### Technical User Option

Users who enable "Developer Mode" in Settings see:
1. The original error code (in addition to the human message)
2. A copyable error report with full context
3. A "Search for this error" link that searches the BuzzPi forum
4. Additional diagnostic information

Developer mode is a secret setting (tap version number 7 times, Material You style). It does not change core behavior — only adds visibility into the technical layer.

## User Experience

A non-technical user sees: "Your Pi stopped responding. It may have lost power. Check the power cable and try again." They understand what happened and what to do. They don't feel stupid for not knowing what "errno 111" means.

A developer (who has enabled Developer Mode) sees the same message plus: "Debug: Connection refused (0.0.0.0:22, errno 111)". They can copy this into a forum post or bug report.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Error translation is imprecise | The "why" is a best guess, not a certainty. This is acceptable — the user needs a recovery path more than they need certainty. If the first action doesn't work, they can try another (or contact support). |
| Developers lose debug info | Developer Mode restores the technical detail. The default mode optimizes for non-technical users, who are the majority. |
| Translation adds code complexity | Every error path needs a human-readable mapping. This is justified by the improved user experience — it is the difference between a helpful tool and a frustrating one. |

## Examples

- Pairing failure: "BuzzPi couldn't pair with this device. Make sure the device is on and showing a pairing code, then try again."
- Screen stream failure: "The screen stream couldn't start. This device may not have a graphical desktop installed."
- Update failure: "The update couldn't be installed. Kitchen Pi may not have enough storage space or may be low on battery. Free up space and try again."
- Network configuration: "BuzzPi needs to send files over the internet. Your phone may have a weak connection. Move closer to your router or try again later."

## Related Patterns

- [Progressive Disclosure](progressive-disclosure.md): Developer Mode is a progressive disclosure of technical detail
- [Offline First](offline-first.md): Error messages differentiate between local and remote issues without mentioning "offline" or "network"
