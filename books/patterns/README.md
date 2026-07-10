# Book 6: BuzzPi Patterns

**How does BuzzPi think?**

This book documents reusable patterns that embody BuzzPi's design philosophy. Every pattern captures a recurring problem, BuzzPi's solution, and the tradeoffs involved.

**Start here:** [North Star](../../NORTH_STAR.md) — [Constitution](../../CONSTITUTION.md) — [Patterns Constraints](constraints.md)

Patterns are not implementation details — they are codified philosophy. New contributors read these to understand *how BuzzPi thinks*, not just how it works.

## Contents

| Pattern | Status | Description |
|---------|--------|-------------|
| [Device Discovery](device-discovery.md) | Complete | Never ask for an IP address. Use mDNS, Bluetooth, QR, or cloud registration automatically. |
| [Progressive Disclosure](progressive-disclosure.md) | Complete | Beginners see what they need. Experts can expand. The interface grows with the user. |
| [Offline First](offline-first.md) | Complete | Work without internet. Cloud is a fallback, not a requirement. |
| [Capability Detection](capability-detection.md) | Complete | Detect what the device supports. Adapt the UI automatically. |
| [Automatic Transport](automatic-transport.md) | Complete | Choose the best connection without user intervention. |
| [Explain, Don't Expose](explain-dont-expose.md) | Complete | Translate technical errors into human language. |
| [Best Effort Reconnect](best-effort-reconnect.md) | Complete | Never lose a session. Reconnect transparently across network changes. |
| [One Action Per Step](one-action-per-step.md) | Complete | Each action does exactly one thing. Compose actions for complex workflows. |
| [Fail Gracefully](fail-gracefully.md) | Complete | Degrade features before crashing. No silent failures. |
| [Configuration by Convention](configuration-by-convention.md) | Complete | Sensible defaults. Configuration files are optional. |
| [Audit by Default](audit-by-default.md) | Complete | Every action is logged. Every state change is recorded. |
| [Privacy by Design](privacy-by-design.md) | Complete | Data stays on-device. Cloud is opt-in, not default. |

*More patterns will be added as the project evolves. Target: 100+ patterns.*

## Pattern Template

Every pattern follows the same structure:

**Problem** — What recurring challenge does this address?

**Solution** — How BuzzPi solves it. What the user experiences.

**Tradeoffs** — What was sacrificed or compromised.

**Examples** — Where this pattern appears in BuzzPi.

**Related Patterns** — How this pattern composes with others.
