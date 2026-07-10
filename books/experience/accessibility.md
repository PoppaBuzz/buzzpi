# Accessibility

BuzzPi is designed for everyone. Accessibility is not a feature — it is a design requirement that applies to every screen, every interaction, and every piece of content.

## Target Standards

- **WCAG 2.1 AA** minimum for all content
- **WCAG 2.1 AAA** for text contrast and terminal output
- **Material 3 accessibility** guidelines for Android components

---

## Color and Contrast

- All text meets 4.5:1 contrast ratio (normal) or 3:1 (large) against its background
- Status colors are distinguishable by brightness and saturation, not hue alone
- Color is never the only indicator of state — icons, labels, and patterns accompany all status colors
- Terminal output supports configurable color themes, all meeting contrast requirements
- Dark mode is not an alternative theme — it is a first-class design with full accessibility verification

---

## Touch Targets

- All interactive elements are at least 48x48dp (Material 3 minimum)
- Touch targets with visual boundaries smaller than 48dp have invisible extended hit areas
- The minimum spacing between touch targets is 8dp
- Primary actions are positioned within thumb reach (bottom third of screen on phones)

---

## Screen Reader Support

- All meaningful content is exposed to screen readers
- Icons have content descriptions (not "icon" — describe the function: "Pair device", "Open workspace")
- Status indicators announce state changes: "Device offline" not "Gray dot"
- Images have alt text describing their content and purpose
- Decorative elements are hidden from screen readers
- Terminal output is accessible as text, not as a bitmap

---

## Focus and Navigation

- All interactive elements are reachable via keyboard/DPAD navigation
- Focus order follows visual order (left-to-right, top-to-bottom)
- Focus indicators are visible (2dp outline, minimum 2:1 contrast against background)
- Custom focus indicators respect system accessibility settings
- Button labels are descriptive, not generic ("Restart Device" not "Restart")

---

## Motion Sensitivity

- All animations respect the system "Reduce motion" setting
- When reduced motion is enabled:
  - Crossfades replace slide transitions
  - No scaling or parallax effects
  - Status changes appear instantly (no 500ms transition)
  - Notification slide-in becomes instant reveal
- No content or functionality depends on animation

---

## Text and Typography

- BuzzPi respects system font size settings
- Minimum body text size is 14sp (never smaller, even in dense views)
- Terminal font size is adjustable in settings (12sp - 20sp)
- Line spacing is generous (1.4x - 1.5x body text)
- Text can be selected and copied from all information screens

---

## Terminal Accessibility

- Terminal output is available to screen readers as it appears
- Command history is navigable and searchable
- Terminal color themes include high-contrast variants
- Terminal bell/flash events are configurable (visual, haptic, or disabled)
- Copy-paste works with standard platform gestures

---

## Testing Requirements

- All screens are tested with TalkBack (Android)
- All screens are tested with 200% font size
- All screens are tested in landscape orientation
- Color-blind simulation is run on all status indicators
- Focus navigation is verified for every interactive element
- All tests are automated where possible and run in CI
