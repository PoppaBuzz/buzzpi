# Brand Usage

**BuzzPi's brand assets represent the project to the world.** This guide ensures consistent, appropriate use of the BuzzPi name, logo, and visual identity.

---

## The BuzzPi Name

### Correct Usage

- **BuzzPi** — The project name, always written with capital B and P: `BuzzPi`, never `buzzpi`, `Buzz pi`, `Buzz-Pi`, or `buzz_pi`
- **The BuzzPi Platform** — When referring to the full ecosystem
- **BuzzPi Agent** — The Go daemon running on devices
- **BuzzPi Android** — The Android companion app
- **BuzzPi Protocol (BPP)** — The communication protocol

### What BuzzPi Is

BuzzPi is:
- An open-source platform for managing Raspberry Pis
- A protocol (BPP) for device-client communication
- A community of developers and users

BuzzPi is NOT:
- A company or corporation
- A commercial product (though commercial hosting of the relay is permitted)
- An operating system
- A cloud service (the reference relay is open-source, not proprietary)

### Attribution

When using BuzzPi in your own projects:

```
Built with [BuzzPi](https://jphat.net)
Uses the [BuzzPi Protocol](https://jphat.net/buzzpi/bpp)
```

---

## Logo

### Primary Logo

The primary logo is a hexagonal bee head with circuit traces forming the eyes. It is available in:

- **Full color** — for digital use on light backgrounds
- **Monochrome white** — for dark backgrounds
- **Monochrome black** — for print and light backgrounds
- **Icon only** — for favicons, app icons, avatars (square crop)

### Minimum Size

| Format | Minimum Size |
|--------|-------------|
| Digital (screen) | 32px height |
| Print | 0.5 inches (12.7mm) height |
| App icon | 1024×1024px |
| Favicon | 32×32px |

### Clear Space

Maintain clear space equal to the height of the "B" in "BuzzPi" on all sides of the logo. No other elements should intrude into this space.

### Incorrect Usage

- Do not stretch, skew, rotate, or distort the logo
- Do not recolor the logo (use only provided color variants)
- Do not add effects (shadows, gradients, strokes)
- Do not place the logo on busy or low-contrast backgrounds
- Do not combine the logo with other elements to create a new mark
- Do not use the logo as part of your own branding

---

## Colors

### Primary Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Buzz Gold | `#F5A623` | Primary brand color, accents |
| Buzz Dark | `#1A1A2E` | Backgrounds, headers |
| Buzz Light | `#F8F9FA` | Page backgrounds |
| Buzz White | `#FFFFFF` | Card backgrounds, text on dark |

### Secondary Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Circuit Green | `#2ECC71` | Success, online status |
| Signal Blue | `#3498DB` | Links, interactive elements |
| Warning Amber | `#E67E22` | Warnings, degraded status |
| Error Red | `#E74C3C` | Errors, offline status |

---

## Typography

### Digital

| Usage | Font | Weight |
|-------|------|--------|
| Headings | Inter | Bold (700) |
| Body | Inter | Regular (400) |
| Code | JetBrains Mono | Regular (400) |

### Print

| Usage | Font | Weight |
|-------|------|--------|
| Headings | IBM Plex Sans | Bold (700) |
| Body | IBM Plex Sans | Regular (400) |
| Code | IBM Plex Mono | Regular (400) |

---

## Voice and Tone

### Principles

- **Clear** — Prefer simple words over jargon. "Start" not "initiate"
- **Direct** — Say what the user needs to know. "Your device will restart" not "The device may need to be restarted"
- **Human** — Write like a person. "We could not find your device" not "Device discovery returned zero results"
- **Empowering** — Focus on what the user can do. "You can pair up to 5 devices" not "The device limit is 5"

### Vocabulary

| Use | Do Not Use |
|-----|------------|
| Device | Node, host, machine, endpoint |
| Pair | Register, bind, connect (first time) |
| Workspace | Dashboard, panel, console |
| Extension | Plugin (for apps), but Plugin for code |
| Action | Command (depends on context) |
| Runtime | Agent, daemon (code name for the Go process) |
| BuzzPi Protocol | BPP on second reference |
| Account | User (for the person), account (for credentials) |

### Examples

```
✅ "Tap a device to connect."
✅ "Your device is offline. It will reconnect when back online."
✅ "Pairing creates a secure connection between your phone and your Pi."
✅ "This plugin needs access to Docker. Tap Allow to continue."

❌ "Initiate pairing protocol on the target host."
❌ "The node is unreachable due to network connectivity issues."
❌ "Register your endpoint to establish a session."
```

---

## Usage in Derivative Works

### Open Source Projects

You may:
- Use the BuzzPi name to describe compatibility ("compatible with BuzzPi Protocol")
- Use the BuzzPi logo to indicate integration ("works with BuzzPi")
- Fork the project under a different name (change the name and logo)
- Create plugins and distribute them via the BuzzPi plugin registry

You may NOT:
- Use "BuzzPi" as part of your project's name (e.g., "BuzzPi Tools")
- Create the impression that your project is official BuzzPi software
- Use the BuzzPi logo as your project's logo

### Commercial Use

You may:
- Host the BuzzPi Cloud Relay as a commercial service
- Build commercial plugins for BuzzPi
- Offer BuzzPi consulting, support, or managed services
- Use BuzzPi in your commercial product (as a dependency, not as branding)

You may NOT:
- Sell the BuzzPi software itself (it's free and open source)
- Use the BuzzPi brand to imply endorsement unless explicitly authorized

---

## Asset Distribution

Official brand assets are available at:

```
https://jphat.net/brand/
```

Includes:
- Logo (SVG, PNG, ICO)
- Color palette (`.ase`, `.json`)
- Font links
- This brand guide (PDF)

For questions about brand usage not covered here, open a GitHub Discussion.
