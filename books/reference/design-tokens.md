# Design Tokens Reference

Source of truth for all visual design tokens. Tokens are organized by category with their implementation values.

## Color Palette

### Brand Colors

| Token | Hex | RGB | Usage |
|-------|-----|-----|-------|
| `color-brand-primary` | `#FFB300` | `rgb(255, 179, 0)` | Primary accent, interactive elements, active states |
| `color-brand-primary-container` | `#FFE082` | `rgb(255, 224, 130)` | Filled buttons, selected states |
| `color-brand-on-primary` | `#1A1A1A` | `rgb(26, 26, 26)` | Text/icon on primary background |
| `color-brand-secondary` | `#FF8F00` | `rgb(255, 143, 0)` | Secondary accent, hover states |
| `color-brand-secondary-container` | `#FFB74D` | `rgb(255, 183, 77)` | Secondary filled states |

### Neutral Colors

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `color-surface` | `#FFFFFF` | `#121212` | Background — main surfaces |
| `color-surface-container` | `#F5F5F5` | `#1E1E1E` | Background — cards, elevated surfaces |
| `color-surface-container-high` | `#EEEEEE` | `#2C2C2C` | Background — dialogs, drawers |
| `color-on-surface` | `#1A1A1A` | `#E0E0E0` | Text/icon on surface |
| `color-on-surface-variant` | `#666666` | `#999999` | Secondary text, captions |
| `color-outline` | `#CCCCCC` | `#444444` | Borders, dividers, disabled state |

### Status Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `color-success` | `#2E7D32` | Online, completed, healthy |
| `color-warning` | `#F57C00` | Warning threshold approached |
| `color-error` | `#C62828` | Offline, failed, error state |
| `color-info` | `#1565C0` | Informational, neutral event |

### Status Indicator Colors

| Status | Hex (Light) | Hex (Dark) | Description |
|--------|-------------|-------------|-------------|
| Online | `#2E7D32` | `#4CAF50` | Device is connected and healthy |
| Offline | `#9E9E9E` | `#757575` | Device is disconnected |
| Warning | `#F57C00` | `#FF9800` | Threshold breached (temp, storage) |
| Error | `#C62828` | `#EF5350` | Device unreachable or failed |
| Pairing | `#1565C0` | `#42A5F5` | Device is in pairing mode |

Status indicators use both color AND icon. Never color alone.

## Typography

### Font Family

| Token | Value | Usage |
|-------|-------|-------|
| `font-family-sans` | `'Inter', 'Roboto', system-ui, sans-serif` | Body text, headings, labels |
| `font-family-mono` | `'JetBrains Mono', 'Cascadia Code', 'Source Code Pro', monospace` | Terminal, code, file paths |

### Type Scale

| Token | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| `text-display` | 32sp | 700 (Bold) | 1.2 | Screen titles, empty state headers |
| `text-headline` | 24sp | 600 (Semibold) | 1.3 | Section headers, dialog titles |
| `text-title` | 18sp | 600 (Semibold) | 1.3 | Card titles, action sheet titles |
| `text-subtitle` | 16sp | 500 (Medium) | 1.4 | Device names, list item titles |
| `text-body` | 14sp | 400 (Regular) | 1.5 | Body text, descriptions, list items |
| `text-body-small` | 12sp | 400 (Regular) | 1.4 | Captions, timestamps, metadata |
| `text-label` | 14sp | 500 (Medium) | 1.2 | Button labels, tab labels, chip text |
| `text-label-small` | 11sp | 500 (Medium) | 1.2 | Small labels, badges, status text |

### Terminal Typography

| Token | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| `text-terminal` | 14sp | 400 (Regular) | 1.4 | Terminal output and input |
| `text-terminal-small` | 12sp | 400 (Regular) | 1.3 | Terminal dense mode |
| `text-terminal-large` | 18sp | 400 (Regular) | 1.4 | Terminal accessibility mode |

Terminal font size is user-adjustable in settings (12sp - 20sp).

## Spacing Scale

Base unit: 4dp

| Token | Value | Usage |
|-------|-------|-------|
| `space-1` | 4dp | Micro spacing, icon padding |
| `space-2` | 8dp | Element spacing, touch target gaps |
| `space-3` | 12dp | Related element grouping |
| `space-4` | 16dp | Standard content padding |
| `space-5` | 20dp | Card inner padding (generous) |
| `space-6` | 24dp | Section spacing |
| `space-8` | 32dp | Screen edge margins, list spacing |
| `space-10` | 40dp | Major section breaks |
| `space-12` | 48dp | Device list item spacing |
| `space-16` | 64dp | Empty state spacing |

### Standard Layout Values

| Context | Value |
|---------|-------|
| Screen edge margin | 16dp |
| Card padding | 16dp |
| List item padding | 12dp vertical, 16dp horizontal |
| Between related elements | 8dp |
| Between unrelated sections | 24dp |
| Button minimum height | 48dp |
| Touch target minimum | 48dp × 48dp |

## Animation Tokens

### Duration

| Token | Value | Usage |
|-------|-------|-------|
| `duration-micro` | 100ms | Button press, toggle, touch feedback |
| `duration-short` | 150ms | Element disappear, dismiss |
| `duration-medium` | 200ms | Element appear, card expansion |
| `duration-long` | 300ms | Screen transitions, status changes |
| `duration-status` | 500ms | Online/offline transition |

### Easing Curves

| Token | Cubic Bezier | Usage |
|-------|-------------|-------|
| `easing-emphasized-decelerate` | `(0.05, 0.7, 0.1, 1.0)` | Elements entering screen |
| `easing-emphasized-accelerate` | `(0.3, 0.0, 0.8, 0.15)` | Elements leaving screen |
| `easing-fast-out-linear-in` | `(0.4, 0.0, 1.0, 1.0)` | Microinteractions |
| `easing-linear` | `(0.0, 0.0, 1.0, 1.0)` | Progress bars, continuous motion |
| `easing-emphasized` | `(0.2, 0.0, 0.0, 1.0)` | Spring-like motion for emphasis |

## Elevation

| Token | Shadow | Usage |
|-------|--------|-------|
| `elevation-0` | None | Flat surfaces |
| `elevation-1` | `0px 1px 2px rgba(0,0,0,0.06)` | Resting cards, list items |
| `elevation-2` | `0px 2px 4px rgba(0,0,0,0.08)` | Hovered cards, raised buttons |
| `elevation-3` | `0px 4px 8px rgba(0,0,0,0.10)` | Dialogs, drawers |
| `elevation-4` | `0px 8px 16px rgba(0,0,0,0.12)` | Modals, snackbars |
| `elevation-5` | `0px 16px 24px rgba(0,0,0,0.14)` | Navigation drawer, FAB |

Dark mode elevation uses lighter shadows (lower opacity) on dark surfaces.

## Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `radius-small` | 4dp | Chips, small elements |
| `radius-medium` | 8dp | Cards, dialogs, buttons |
| `radius-large` | 12dp | Bottom sheets, large containers |
| `radius-full` | 999dp | Pills, circular icons, FAB |

## Icon Sizing

| Context | Size |
|---------|------|
| Navigation bar | 24dp |
| Action bar | 24dp |
| List item | 24dp |
| Device card avatar | 40dp |
| Status indicator | 12dp |
| Inline with text | 16dp |
| Empty state illustration | 120dp |
| Notification icon | 24dp (system) |
