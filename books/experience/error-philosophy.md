# Error Philosophy

BuzzPi's error philosophy is defined in the Manifesto: **Explain, Don't Expose.** A raw error code or technical message is a failure of design, not a communication of information.

## Guiding Principles

### Every Error Has a Recovery Path

If BuzzPi can detect an error condition, BuzzPi can suggest what to do about it. Every error message includes:
1. **What happened** — in plain language, without technical jargon
2. **Why it might have happened** — the most likely cause, not a list of possibilities
3. **What the user can do** — concrete next steps, ordered by likelihood of success

**Instead of:** "Connection refused"
**Show:** "BuzzPi couldn't reach your Raspberry Pi. It may be asleep, disconnected from the network, or powered off. Check that it is plugged in and connected to your network, then try again."

---

### Errors Are Human, Not Technical

Error messages are written for a person who does not know what SSH is, what an IP address is, or what a port number is. If the message requires technical knowledge to understand, it needs rewriting.

**Instead of:** "SSH Error Code 255"
**Show:** "BuzzPi couldn't complete the connection. The device may have rejected the request. Try pairing again."

---

### Anticipate, Don't React

The best error handling prevents the error from being visible. Before attempting an action, BuzzPi checks preconditions:

- Is the device online? If not, show it before the user taps.
- Is the capability available? If not, disable the button and explain why.
- Will timeout occur? Set expectations before the operation starts.

**Consequence:** Many "errors" become disabled states with clear explanations.

---

### Error Categories

| Category | UX Treatment | Examples |
|----------|-------------|---------|
| Transient | Automatic retry, brief toast | Network timeout, DNS failure |
| Configuration | Persistent message with fix | Docker not installed, camera disabled |
| Authentication | Clear recovery path | Pairing expired, key mismatch |
| Permission | Explain and redirect | Storage permission denied, camera permission denied |
| Precondition | Disable action, explain why | Device offline, capability unavailable |
| Unexpected | Fallback message with feedback option | Internal error, please report |

---

### What Never to Show

- Raw error codes (except in debug logs, which are off by default)
- Stack traces
- "An error occurred" without elaboration
- Network error details (port numbers, connection refused, ETIMEDOUT)
- JSON parsing errors
- Null pointer or index out of range messages

---

### Error Message Template

```
[What happened, one sentence]

[Why it might have happened, one or two sentences]

[What the user can do, one or two actions]
```

Buttons at the bottom of the error use canonical verbs:
- **Retry** — try the action again
- **Pair Again** — re-establish trust
- **Check Device** — guide the user to inspect their hardware
- **Report** — submit anonymized error details (opt-in only)

---

### Error Logging

Error messages shown to the user are logged locally with sufficient context for debugging. These logs are:
- Stored on-device only
- Never transmitted without explicit opt-in
- Cleared when the user unpairs all devices
- Bounded to 1MB before rotation
