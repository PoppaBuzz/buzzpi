# Typography

## Font Family

**Primary:** Inter

Inter is an open-source, highly legible typeface designed for screens. It has excellent readability at small sizes, clear distinction between similar characters (I/l/1), and extensive language support.

**Monospace:** JetBrains Mono

JetBrains Mono is used for terminal output, code blocks, logs, and any monospaced context. It has clear distinction between 0/O, good readability at terminal font sizes, and programming ligatures (optional).

## Type Scale

| Token | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| `display` | 36sp | Medium 500 | 44sp | Device name on Workspace |
| `headline` | 24sp | Medium 500 | 32sp | Screen titles |
| `title` | 20sp | Medium 500 | 28sp | Card titles, section headers |
| `subtitle` | 16sp | Regular 400 | 24sp | Subtitles, metric labels |
| `body` | 14sp | Regular 400 | 20sp | Primary reading text |
| `body-emphasis` | 14sp | Medium 500 | 20sp | Emphasized body text |
| `caption` | 12sp | Regular 400 | 16sp | Secondary info, timestamps |
| `label` | 11sp | Medium 500 | 16sp | Button text, badges, chips |
| `terminal` | 13sp | Regular 400 via JetBrains Mono | 20sp | Terminal output |
| `mono` | 13sp | Regular 400 via JetBrains Mono | 20sp | Code, logs, IDs |

## Typography Principles

### Hierarchy

Type scale creates clear visual hierarchy without relying on color. The most important information on any screen uses the largest size. Secondary information is smaller. Tertiary details are caption size.

### Density

BuzzPi is information-rich. Terminal output, log streams, Docker process lists, and GPIO states require reading density. The terminal size (13sp) balances readability with information density. Users can adjust terminal font size in settings.

### Consistency

Every client uses the same type scale. The Android app, desktop client, CLI, and website share typographic DNA. Font sizes may adjust for platform conventions (Material 3 on Android, system fonts on CLI), but the hierarchy remains intact.

### Accessibility

- Minimum body text size is 14sp (comfortably readable at arm's length)
- Users can increase font size system-wide — BuzzPi respects system accessibility settings
- Line heights are generous (1.4x-1.5x) for readability
- Font weight is never the only differentiator for interactive elements

## Language Support

Inter supports 437 languages and all major writing systems. JetBrains Mono supports the most common programming languages and terminals. Right-to-left (RTL) layout support is planned for v1.0.
