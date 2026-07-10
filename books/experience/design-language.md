# Design Language

BuzzPi's design language is guided by the Manifesto: native, beautiful, discoverable, and performant. It draws inspiration from Material 3, Google's ecosystem design, and the industrial design of the Raspberry Pi hardware itself.

## Design Values

**Clear.** Information is organized hierarchically. The most important thing on any screen is visually prominent. Secondary information is accessible but never competes for attention.

**Calm.** BuzzPi does not shout. Status indicators are present but not urgent. Notifications are informative, not alarming. The interface recedes, letting the user focus on their device.

**Capable.** The interface communicates that it can handle whatever the user needs. Empty states are informative. Error states are helpful. Loading states are honest about time.

**Connected.** Visual design reinforces the feeling of being connected to a remote device. Status dots pulse subtly. Transitions between states feel continuous, not abrupt.

---

## Layout Philosophy

Content is organized in a single-column hierarchy on phones, expanding to multi-column on tablets and desktop.

**Information density** decreases as screen size decreases. On a phone, a device shows its name, status, and primary action. On a tablet, it additionally shows live metrics, recent activity, and capability shortcuts.

**Whitespace is intentional.** It separates functional groups. It guides the eye to the next action. It prevents the interface from feeling crowded, even when displaying complex data like Docker container lists or GPIO pin states.

---

## Components

### Cards

Cards group related information. They have:
- Rounded corners (12dp default)
- Subtle elevation (level 1 for default, level 2 for interactive)
- No background color by default (transparent cards on the surface color)
- Optional emphasis through accent color or increased elevation

### Sheets

Bottom sheets are used for:
- Quick actions on a device (Restart, Shutdown, Execute Command)
- Selecting from a list (file picker, Docker image selector)
- Confirmation dialogs with context

### Dialogs

Dialogs are reserved for:
- Destructive actions (factory reset, unpair device)
- Critical confirmations (overwrite file, stop service)
- Error recovery paths

### Snackbars

Snackbars provide brief feedback for completed actions. They disappear automatically after 4 seconds. They never contain critical information or actions.

### Progress Indicators

- **Linear** for determinate progress (file upload, Docker pull)
- **Circular** for indeterminate progress (connecting, discovering)
- **Pulsing dot** for device status (subtle, non-blocking)

---

## Dark Mode

BuzzPi supports light and dark mode. Colors are not inverted — they are thoughtfully mapped to maintain hierarchy and readability in both themes. Status colors remain identifiable in both modes through brightness and saturation differences, not just hue.

Dark mode is the default for the Workspace screen to reduce glare during terminal use and screen streaming.

---

## Responsive Behavior

| Screen Size | Layout |
|-------------|--------|
| Phone (< 600dp) | Single column, bottom navigation |
| Foldable (600-840dp) | Two-column, device list + workspace |
| Tablet (> 840dp) | Multi-column, sidebar + workspace + detail |
| Desktop (> 1200dp) | Full IDE-like layout with panels |

Transitions between layouts preserve context. Opening a device on the phone navigates to a new screen. On a tablet, it opens in the adjacent panel.
