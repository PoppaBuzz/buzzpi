# Microinteractions

Microinteractions are the small moments that define how BuzzPi feels. They are the difference between a tool that gets the job done and one that is a pleasure to use.

## Principles

### 100ms Rule

Every microinteraction completes in under 100ms or provides immediate feedback that progress has started. If the action takes longer than 100ms, the interface shows a progress indicator within that window.

### No Dead States

Every tap produces a response. If the action is not possible, the button is disabled with an explanation. If the action is possible but takes time, the button transitions to a progress state. There is no "nothing happened" state.

### Predictable

Microinteractions follow consistent patterns. A toggle always responds the same way. A button press always produces the same feedback. The user builds an accurate mental model after two interactions.

---

## Specific Microinteractions

### Device Toggle (Star / Select)

Tapping the star icon on a device card:
1. Haptic tap (if enabled)
2. Icon fills over 100ms (fast-out, linear-in)
3. Card background gains a subtle tint (0 -> 8% accent color overlay)
4. The device moves to the top of the favorites section on next list render

### Pull to Refresh

Dragging down past the threshold:
1. Icon animates from rest to spinning over 300ms
2. Haptic pulse at threshold
3. Animation continues while refreshing
4. Brief checkmark (200ms) on success before returning to rest

### Switch Toggle

Tapping a switch:
1. Thumb slides to target position over 100ms
2. Track color transitions (gray -> accent or accent -> gray) over 150ms
3. Haptic confirmation (if enabled)
4. Label remains visible throughout (no disappearing text)

### Slider

Dragging a slider handle:
1. Handle scales up (1.0 -> 1.2) on touch down over 100ms
2. Value label appears above handle
3. Haptic feedback at each integer boundary (if enabled)
4. Handle scales back on release over 100ms
5. Value label fades out over 150ms

### Search

Typing in the search field:
1. Clear button appears when text is non-empty
2. Results update after 300ms debounce (not on every keystroke)
3. Empty results state fades in over 200ms
4. Search field expands on focus (if compact mode) over 200ms

### Action Execution

Tapping an action button (e.g., "Restart"):
1. Button text fades to spinner over 150ms
2. Spinner animates while action is in progress
3. On completion: spinner fades to checkmark (success) or X (failure) over 150ms
4. Checkmark/X fades back to button label after 1s
5. Duration: entire microinteraction completes within the button bounds

### Long Press

Long pressing a device card:
1. Haptic pulse at 400ms
2. Card lifts (elevation + shadow) over 100ms
3. Action menu appears anchored to the press point over 200ms
4. Background dims (if needed) over 150ms

### Swipe to Dismiss

Swipping a notification or card:
1. Card follows finger position (no lag)
2. Background reveals a subtle action indicator (e.g., dismiss icon)
3. Past threshold: card slides off with acceleration, item removed from list over 300ms
4. Below threshold: card springs back with overshoot over 250ms
5. Undo option appears briefly (3s snackbar)

---

## Haptic Feedback Patterns

| Interaction | Pattern | Duration |
|-------------|---------|----------|
| Tap button | Light click | 20ms |
| Toggle switch | Confirmation | 30ms |
| Long press | Heavy click | 40ms |
| Pull refresh | Threshold reached | 20ms |
| Error | Double pulse | 2 x 20ms |
| Action completed | Soft confirmation | 30ms |
| Slider step | Tick | 10ms |

---

## Empty State Animations

Empty states are not dead states. When a list is empty, the empty state illustration has a subtle breathing animation (scale: 1.0 -> 1.02 -> 1.0, duration: 4s, loop). This conveys that BuzzPi is alive and waiting, not frozen.

When the first device is paired, the empty state illustration fades into the first device card over 300ms — the illustration doesn't disappear, it transitions.
