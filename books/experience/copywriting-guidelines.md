# Copywriting Guidelines

BuzzPi's voice is competent, calm, and clear. We are a capable tool that respects the user's intelligence without assuming technical knowledge.

## Voice

| Category | Guideline |
|----------|-----------|
| Tone | Calm, confident, helpful. Never urgent, never playful. |
| Persona | A knowledgeable colleague who explains without condescension. |
| Complexity | Match the user's language. Default to simple terms. Use technical terms only when they are the clearest option — and then explain them. |
| Humor | None. BuzzPi manages hardware. Misunderstandings have real consequences. |

---

## Terminology

### Must Use

| Term | Why |
|------|-----|
| Device | Not "node", "host", "machine", "unit", "endpoint" |
| Pair / Unpair | Not "register", "add", "connect", "link" |
| Workspace | Not "dashboard", "console", "panel", "home" |
| Extension | Not "plugin", "add-on", "module", "pack" |
| Action | Not "command", "task", "job", "operation" |
| Runtime | Not "agent", "daemon", "service" (the on-device software) |
| Workspace | The screen shown when you tap a device |
| Buzz | The brand name for cross-device features ("Buzz from your phone to your Pi") |

### Must Avoid

| Term | Instead Use |
|------|-------------|
| SSH | Remote access (or explain: "a secure way to access your device") |
| IP address | Network address (or omit: just "find your Pi") |
| Port | (omit — the user never needs to know) |
| Certificate | Trust keys, device identity |
| API | (omit — extensions use the protocol) |
| Daemon | Background service, Runtime |
| Kill | Stop, terminate |
| Crash | Stop unexpectedly, fail |
| Error code | (explain what happened) |

---

## Writing Standards

### Button Labels

- Use verbs: "Restart", "Pair Device", "Open Workspace", "Cancel", "Dismiss"
- Never use OK/Cancel alone — the button should communicate the outcome
- Use affirmative labels: "Delete" not "Do Not Delete", "Keep" not "Cancel"
- Maximum 3 words per button

### Error Messages

Follow the Error Philosophy template:
```
[What happened]

[Why it might have happened]

[What the user can do]
```

No exclamation marks in errors. No "Oops!" No "Uh oh."

### Empty States

Empty states are opportunities to guide, not apologies.

| State | Copy |
|-------|------|
| No devices | "No devices yet. Tap the button below to pair your first Raspberry Pi." |
| No extensions | "This device has no extensions installed. Browse available extensions to add functionality." |
| No notifications | "No notifications. You'll see device events here." |
| No search results | "No devices match your search. Try a different name." |

### Progress Messages

| State | Copy |
|-------|------|
| Connecting | "Connecting to Kitchen Pi…" |
| Disconnecting | "Disconnecting…" |
| Updating | "Updating Kitchen Pi to v0.2.5…" |
| Restarting | "Restarting Kitchen Pi…" |
| Pairing | "Waiting for device…" |
| Completed | "Done" (not "Success!" or "Complete!") |
| Failed | "Failed to [action]" (not "Unsuccessful" or "Error") |

---

## Capitalization

- Title case for: screen titles, button labels, dialog titles
- Sentence case for: body text, descriptions, notifications, error messages
- Proper case for: device names (user-defined, preserve their casing), extension names

## Punctuation

- Use sentence-ending periods in body text
- Never use periods in button labels, titles, or list items
- Never use exclamation marks
- Use ellipsis for in-progress actions ("Connecting…")
- Use colons to introduce examples or explanations
- Use dashes for parentheticals — they work well in technical explanations

## Localization

- All strings are externalized to resource files (Android strings.xml, etc.)
- String IDs follow the pattern: `buzzpi_[screen]_[element]_[state]`
- Maximum string length is calculated for German (typically 30-40% longer than English)
- Concatenation is avoided — use placeholders for dynamic content
- Plurals use proper ICU message format
