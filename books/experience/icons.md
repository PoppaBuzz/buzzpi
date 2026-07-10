# Icons

BuzzPi uses a consistent icon language built on Material Design icons (Material 3) with a small set of custom icons for BuzzPi-specific concepts.

## Icon Set

### Primary Set (Material Design Icons)

All interface icons are sourced from Material Design Icons (https://materialdesignicons.com/). This provides a consistent, recognizable, and well-maintained set for all standard UI elements.

### Custom Icons

The following BuzzPi-specific concepts require custom icons:

| Concept | Icon Description | Status |
|---------|-----------------|--------|
| BuzzPi (app) | Stylized bee with Raspberry Pi reference | Needs design |
| BuzzPi (device) | Small bee silhouette | Needs design |
| Runtime | Gear with bee detail | Needs design |
| Pairing | Two nodes with link | Needs design |
| Workspace | Window with bee | Needs design |
| Extension | Puzzle piece with bee | Needs design |
| Screen Stream | Display with signal waves | Needs design |
| Device Group | Stack of device icons | Needs design |
| Buzz Transfer | Arrow through bee | Needs design |

---

## Guidelines

### Style

- Filled (solid) style for navigation bars and primary actions
- Outline style for secondary actions and content areas
- Consistent stroke width (2dp for 24dp icons)
- Consistent corner radius (2dp)
- No gradients, no multi-colored icons (single color currentColor)

### Sizing

| Context | Size |
|---------|------|
| Navigation bar | 24dp |
| Action bar | 24dp |
| List item | 24dp |
| Device card | 40dp (avatar) |
| Notification icon | 24dp (system standard) |
| Empty state illustration | 120dp |
| Status indicator (dot) | 12dp |
| Inline with text | 16dp (aligned to text baseline) |

### Usage

- Icons always convey meaning. If an icon is decorative, it should have no label and be hidden from screen readers.
- Icons with labels always have the label below (button) or to the right (list item).
- Status indicators use the status dot pattern (color + icon), never icon alone.
- Custom BuzzPi icons are reserved for primary BuzzPi concepts. General concepts use standard Material icons.

### Color

- Icons use `currentColor` and inherit from their container
- Interactive icons have a 12dp touch target with 8dp visible icon
- Disabled icons have 38% opacity
- Status indicator icons use the status color (green, yellow, red, gray)

### Custom Icon Requests

When a developer needs a custom icon:
1. Check if a suitable Material Design icon already exists
2. If yes, use it. No custom icon needed.
3. If no, file a design request with the concept description and context
4. Custom icons are designed in SVG, exported as vector drawables (Android)
