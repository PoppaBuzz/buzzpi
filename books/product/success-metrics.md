# Success Metrics

How we measure whether BuzzPi is achieving its mission.

## North Star Metric

**Active Devices:** Number of Raspberry Pi (or other Linux) devices managed through BuzzPi daily.

This measures the core value proposition: BuzzPi is useful enough to stay connected. Everything else feeds this.

## Key Results (v1.0 Target)

| Metric | Target | Why |
|--------|--------|-----|
| Active Devices (daily) | 10,000 | North star. Meaningful community adoption. |
| Pairing completion rate | >90% | Users who start pairing finish it. |
| 7-day retention | >60% | Users who pair a device are still using it a week later. |
| 30-day retention | >40% | Long-term value. Device management is ongoing. |
| NPS | >40 | Users would recommend BuzzPi. |
| Screen stream latency (p95) | <500ms | Remote desktop must feel responsive. |
| Connection success rate | >99% | Relay must be reliable. |
| Extension install rate | >30% of active users | Ecosystem health. |
| App crash-free rate | >99.5% | Reliability baseline. |
| Time to first pair | <3 minutes (new user) | Setup friction elimination. |
| Support tickets per user/month | <0.1 | Self-serve is working. |

## Leading Indicators (Pre-v1.0)

Before we have 10,000 active devices, these leading indicators tell us if we're on track:

| Indicator | Target | When |
|-----------|--------|------|
| Beta signups | 500 | Pre-alpha |
| First-week pairing rate | >80% | Alpha |
| Users with >1 device | >40% | Alpha |
| Average sessions per user per week | >3 | Beta |
| Community Discord/Forum members | >1,000 | Beta |
| Extension submissions (community) | >10 | Beta |
| GitHub stars | >500 | Beta |
| Documentation completion | 100% (all core chapters) | Alpha |

## Qualitative Signals

Not everything can be measured in numbers. These qualitative signals indicate we're building the right thing:

- Users recommend BuzzPi to other makers without being asked
- Users report discovering new uses for their Pis because BuzzPi made it easier
- Extensions appear that the core team didn't anticipate
- Teachers adopt BuzzPi for classroom management without being targeted
- Users stop using SSH/VPN/Tailscale specifically for Pi access

## Anti-Metrics

Things we optimize against:

| Anti-Metric | Concern |
|-------------|---------|
| Time spent in the app per session | BuzzPi should be efficient, not sticky. Check status, take action, leave. |
| Notification volume per user per day | More notifications = more noise, not more value. |
| Configuration options per screen | Every option is a decision. Clutter indicates poor design. |
| Dependency count (Runtime) | Every dependency is a failure point. Minimal surface area. |
| Support escalations for network issues | Network should "just work." Escalations indicate relay/connection failure. |

## Measurement

| Phase | How We Measure |
|-------|----------------|
| Pre-alpha | Manual tracking, user interviews, session recording |
| Alpha | Built-in anonymous analytics (opt-in), crash reporting |
| Beta | Above + engagement metrics, funnel analysis |
| v1.0 | Above + NPS surveys, support ticket analysis |

All analytics are privacy-respecting, opt-in only, and never sell user data.
