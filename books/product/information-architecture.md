# Information Architecture

The structure of BuzzPi's app — screens, navigation, and content hierarchy.

## Screen Map

```
BuzzPi App
├── Device List (Home)
│   ├── Device Card (each device)
│   │   ├── Favorites (top section, star toggled)
│   │   └── All Devices (alphabetical, default sort)
│   ├── Search (overlay, full-screen on focus)
│   ├── Add Device (FAB)
│   └── Profile / Settings (top-right menu)
│
├── Workspace (per device, full-screen)
│   ├── Status Bar (collapsed: status, temp, storage)
│   │   └── Expanded: IP, uptime, model, OS version
│   ├── Terminal Tab
│   │   ├── Command input
│   │   ├── Output area (scrollable)
│   │   └── Quick actions bar (predefined commands)
│   ├── Screen Tab
│   │   ├── Video stream (WebRTC)
│   │   ├── Touch/mouse controls overlay
│   │   └── Toolbar (keyboard, home, orientation)
│   ├── Services Tab
│   │   ├── Service list (name, status, uptime)
│   │   └── Service detail (logs, restart)
│   ├── Extensions Tab
│   │   ├── Installed extensions
│   │   ├── Extension detail (configure, remove)
│   │   └── Browse extensions (store)
│   ├── Files Tab
│   │   ├── File browser
│   │   ├── Upload / Download
│   │   └── Recent files
│   └── Actions Bar (bottom)
│       ├── Restart, Shutdown, Update
│       └── More actions (power off, reboot)
│
├── Device Groups (tab bar)
│   ├── Group list
│   ├── Group detail (per group)
│   │   ├── Device list (filtered)
│   │   ├── Group actions (restart all, update all)
│   │   ├── Group settings (classroom mode, permissions)
│   │   └── Group dashboard (aggregate status)
│   └── Create group
│
├── Notifications (tab bar)
│   ├── Notification list (chronological, grouped by device)
│   ├── Notification settings (per category)
│   └── Clear all
│
├── Profile
│   ├── Account settings
│   │   ├── Email, password, delete account
│   │   └── Subscription / plan
│   ├── App settings
│   │   ├── Theme (light, dark, system)
│   │   ├── Notifications
│   │   ├── Privacy (analytics opt-in, crash reporting)
│   │   └── Terminal (font size, theme)
│   ├── About
│   │   ├── Version, licenses, credits
│   │   └── Documentation links
│   └── Help / Support
│       ├── FAQ
│       ├── Contact support
│       └── Report issue
│
└── Pairing Flow (modal)
    ├── Scan / Discover
    ├── Enter code
    ├── Name device
    └── Workspace (deep link after pairing)
```

## Navigation

### Bottom Navigation (Mobile)

| Tab | Icon | Content |
|-----|------|---------|
| Devices | Server icon | Device list, device groups |
| Groups | Layers icon | Device group management (secondary tab) |
| Notifications | Bell icon | Notification list |
| Profile | Person icon | Settings, help, account |

### Device List → Workspace Transition

Tapping a device card pushes the Workspace as a full-screen view. The workspace replaces the entire content area, including bottom navigation, to provide maximum space for terminal/screen content.

Back navigation from Workspace returns to the Device List.

### Workspace Internal Navigation

- **Tabs** (horizontal scrollable, top area): Terminal, Screen, Services, Extensions, Files
- **Actions** (bottom bar, always visible): Restart, Shutdown, Update
- **Status** (collapsible top bar): Online/offline, temperature, storage

## Deep Links

| Intent | URI | Opens |
|--------|-----|-------|
| View device | `buzzpi://device/{id}` | Workspace for device |
| Pair device | `buzzpi://pair` | Pairing flow |
| Open notification settings | `buzzpi://settings/notifications` | Notification settings |
| Device group | `buzzpi://group/{id}` | Group detail |

## Content Hierarchy Rules

1. **The device is the primary object.** Every screen is organized around devices, groups of devices, or device events.
2. **Actions are secondary to information.** Show status first, actions second. The user should see what's happening before being asked to do something.
3. **Settings are tertiary.** Settings are accessed from the profile, not from the workspace. Don't distract from the device.
4. **One level deep for common tasks.** Status check, terminal, screen — all one tap from the device list.
5. **Two levels max for uncommon tasks.** Extension configuration, group settings, file management — never more than two taps from the device list.
