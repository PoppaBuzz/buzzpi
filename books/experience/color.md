# Color

## Brand Palette

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `primary` | #006B5E | #7ED9C8 | Primary actions, active states, links |
| `on-primary` | #FFFFFF | #00382F | Text and icons on primary |
| `primary-container` | #A7F2E0 | #005047 | Containers using primary |
| `secondary` | #4A635C | #ACCCC0 | Secondary actions, badges |
| `tertiary` | #41647B | #AACBE7 | Accent, alternative actions |

BuzzPi's primary color is a deep teal — technical, calm, and distinctive. It evokes the green of circuit boards, the blue of networking, and the calm of a functioning system. It is not blue (too common in tech), not green (too associated with "all good"), and not purple (too abstract).

---

## Neutral Palette

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `background` | #FAFCFA | #121413 | App background |
| `surface` | #F4F7F4 | #1A1C1B | Card and sheet backgrounds |
| `surface-variant` | #DDE4DE | #2E312F | Elevated surfaces |
| `outline` | #717971 | #878D87 | Borders, dividers |
| `on-background` | #191C1A | #E0E4DF | Primary text |
| `on-surface` | #191C1A | #E0E4DF | Text on surfaces |
| `on-surface-variant` | #414941 | #C1C7C0 | Secondary text |

---

## Status Colors

Status colors use semantic names, mapped to actual colors by the theme. This ensures accessibility tools and custom themes can adjust colors without breaking meaning.

| Semantic | Light | Dark | Meaning |
|----------|-------|------|---------|
| `status-healthy` | #2E7D32 | #66BB6A | Device online, all normal |
| `status-attention` | #1565C0 | #42A5F5 | Information available, non-critical |
| `status-warning` | #E65100 | #FFA726 | Approaching threshold |
| `status-critical` | #C62828 | #EF5350 | Threshold exceeded |
| `status-inactive` | #757575 | #9E9E9E | No activity for 7+ days |
| `status-unknown` | #9E9E9E (50%) | #757575 (50%) | State cannot be determined |

Status colors are used consistently across all clients. A warning in the Android app has the same color as a warning in the CLI. See [Status](../lexicon/status.md) for the full status taxonomy.

---

## Accessibility Requirements

- All text meets WCAG AA contrast ratio (4.5:1 for normal text, 3:1 for large text)
- Status colors are distinguishable by brightness and saturation, not hue alone
- Interactive elements have visible focus indicators (2dp outline, 2:1 contrast against background)
- Color is never the only indicator of state — icons and labels accompany all status colors

---

## Usage Guidelines

- Primary color is used sparingly — it signals interactivity
- Status colors are reserved for device state — never used decoratively
- Neutral palette handles 90% of the interface — color draws attention to what matters
- Surface colors differentiate hierarchy, not content — lower surfaces are lighter (light theme) or darker (dark theme)
