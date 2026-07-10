# Motion Language

Motion in BuzzPi is not decorative. Every animation has a purpose: to communicate state change, maintain context, or guide attention.

## Principles

### Purposeful

Every animation answers a question. "Where did that go?" "What is happening?" "What should I do next?" If an animation does not answer a user question, it does not belong.

### Performant

All animations run at 60 fps. Any animation that drops frames is removed or reworked. On lower-end devices, animations degrade gracefully to opacity-only transitions.

### Subtle

BuzzPi is a utility, not a game. Animations use short durations (150-300ms), standard easings, and minimal motion paths. The interface should feel responsive, not playful.

---

## Durations

| Context | Duration | Easing |
|---------|----------|--------|
| Microinteraction (button press, toggle) | 100ms | Fast-out, linear-in |
| Element appear (card, sheet, dialog) | 200ms | Emphasized decelerate |
| Element disappear (dismiss, close) | 150ms | Emphasized accelerate |
| Screen transition (push, pop) | 300ms | Emphasized decelerate |
| Progress indicator (linear) | Continuous | Linear |
| Device status change (online to offline) | 500ms | Emphasized decelerate |

---

## Easing Curves

| Curve | Cubic Bezier | Usage |
|-------|-------------|-------|
| Emphasized decelerate | (0.05, 0.7, 0.1, 1.0) | Elements entering the screen |
| Emphasized accelerate | (0.3, 0.0, 0.8, 0.15) | Elements leaving the screen |
| Fast-out, linear-in | (0.4, 0.0, 1.0, 1.0) | Microinteractions |
| Linear | (0.0, 0.0, 1.0, 1.0) | Progress bars, continuous motion |

---

## Specific Animations

### Device Discovery

When a device appears on the network, it does not suddenly pop into the list. It fades in with a scale (1.0 -> 1.03 -> 1.0) over 300ms, accompanied by a brief status indicator reveal. This feels like the device is coming online, not being inserted into a list.

### Device Connection

Tapping a device triggers a ripple from the tap point that expands into a full-screen transition. The device card expands into the Workspace. The transition takes 300ms and uses emphasized decelerate.

### Status Change

When a device goes offline, the status indicator transitions over 500ms. The color shifts first (green to gray), then the icon changes. This two-phase transition prevents the UI from feeling jittery during brief connectivity blips. If the device reconnects within 2 seconds, the transition reverses without completing.

### Screen Streaming Start

The screen stream view fades in over 200ms. A brief connecting indicator (circular progress) appears during WebRTC negotiation, then seamlessly transitions to the video stream. The transition between the indicator and the stream is a crossfade, not a cut.

### Notification Arrival

Notifications slide in from the top/bottom over 250ms with emphasized decelerate. They do not interrupt the current action. Critical notifications use a brief haptic pulse in addition to the visual animation.

### Action Execution

When the user executes an action, the button transitions to a progress state over 150ms. The button label fades to a spinner, then fades back to the result (Completed/Failed) over 300ms. The entire microinteraction fits within the button bounds — no dialog, no interruption.

---

## Motion Tokens

| Token | Value |
|-------|-------|
| `duration-micro` | 100ms |
| `duration-short` | 150ms |
| `duration-medium` | 200ms |
| `duration-long` | 300ms |
| `duration-status` | 500ms |
| `easing-emphasized-decelerate` | (0.05, 0.7, 0.1, 1.0) |
| `easing-emphasized-accelerate` | (0.3, 0.0, 0.8, 0.15) |
| `easing-fast-out-linear-in` | (0.4, 0.0, 1.0, 1.0) |
| `easing-linear` | (0.0, 0.0, 1.0, 1.0) |
