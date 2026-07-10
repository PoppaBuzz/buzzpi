# UX Principles

Beyond the nine principles in the Manifesto, these UX principles govern every screen, interaction, and microinteraction in BuzzPi.

## Principle 1: The Device Is the Hero

Every screen centers the device, not the app. The app is a window into the Pi. The Workspace should feel like you are looking at the device, not at an application.

**Consequences:** Status indicators are prominent. The device name is always visible. Time since last contact is never hidden. The device's capabilities determine what the UI shows, not the other way around.

---

## Principle 2: Progressive Disclosure

Beginners see what they need. Experts can drill deeper. No one is overwhelmed, and no one is held back.

The primary action on any screen is obvious — usually a single tap. Secondary actions are one tap away. Advanced options are expanded on demand. Terminal commands are hidden until the user asks for them.

**Consequences:** A user who only needs "Restart" sees a restart button. A user who needs "sudo systemctl restart jellyfin" can expand to see it. Both coexist on the same screen without either feeling wrong.

---

## Principle 3: Zero-Error Day

The first error a user sees is the first error we failed to prevent. BuzzPi anticipates failure modes and handles them before the user encounters them.

If a device goes offline, the UI shows it before the user tries to connect. If an action will fail (e.g., Docker not installed), the UI disables it and explains why before the user taps.

**Consequences:** Every error state has a design, not just a dialog box. Every disabled button has a tooltip explaining why. Every timeout has a fallback.

---

## Principle 4: State Is Never Ambiguous

At any moment, the user knows exactly what is happening. Loading states are shown within 200ms. Progress is measurable. Failures are explained.

A progress spinner alone is not enough. The user should know what step is happening (e.g., "Connecting via LAN...", "Trying Tailscale...", "Falling back to relay").

**Consequences:** Every async operation shows progress. Every transition has a label. Every failure has a recovery path.

---

## Principle 5: Connection Is Invisible

The user should never think about how they are connected. The Connection Engine handles it transparently. The UI does not show "Connected via WebRTC" — it shows the device as available.

Network details are available for those who want them (in device settings), but they are never required for normal use.

**Consequences:** Network transport is never shown in the primary UI. Connection errors use human language, not network terminology. Reconnection happens automatically without user intervention.

---

## Principle 6: The Interface Grows With the User

A first-time user sees a welcoming, sparse interface with clear next actions. A power user sees the same interface with expanded capabilities, keyboard shortcuts, and automation options.

The app remembers what the user has done. Common actions rise to the top. Rare actions are accessible but not prominent.

**Consequences:** The home screen adapts to usage patterns. The Workspace rearranges based on capability importance. The terminal remembers command history and frequently accessed paths.

---

## Principle 7: Every Screen Has a Purpose

No screen exists without a clear job. If a screen's purpose can be served by an existing screen, it should be. If a screen's purpose is unclear, it should not exist.

Every screen answers: "What can I do here?" within 2 seconds of looking at it.

**Consequences:** Onboarding is minimal — show the device list and let the user explore. Settings are grouped by purpose, not by technical category. Empty states explain why the screen is empty and what to do next.

---

## Principle 8: Touch First, Precision Second

BuzzPi is primarily used on a phone. Every interaction is designed for thumbs, not mice. Precision interactions (terminal, GPIO pin selection) provide magnification, snapping, or input helpers.

**Consequences:** Interactive elements are at least 48dp. The primary action is within thumb reach. Terminal input provides keyboard shortcuts for common operations. GPIO pin selection uses visual pin diagrams, not pin numbers.
