# Configuration by Convention

**Sensible defaults. Configuration files are optional.**

## Problem

Many device management tools require configuration before they do anything useful: "Edit /etc/buzzpi.conf", "Set your relay server URL", "Configure your API key". Every configuration file is a failure point — a typo, a missing value, a deprecated option. The Runtime should work immediately after installation, with zero configuration. Configuration should be an exception, not a requirement.

## Solution

The Runtime has one configuration file (`/var/lib/buzzpi/config.json`) but it is entirely optional. If it does not exist, the Runtime runs with sensible defaults.

### Convention Over Configuration

| Feature | Default Convention | Configurable? |
|---------|-------------------|---------------|
| Relay Server URL | `wss://jphat.net/buzzpi/relay/ws` | Yes |
| Log level | `info` | Yes |
| Log directory | `/var/lib/buzzpi/logs/` | Yes |
| Max log size | 1MB | Yes |
| Heartbeat interval | 30s | Yes |
| Screen capture mode | `auto` | Yes |
| Terminal shell | `/bin/bash` | Yes |
| Extension directory | `/var/lib/buzzpi/extensions/` | Yes |
| Allowed file roots | `/home/*`, `/var/lib/buzzpi/*`, `/tmp/*` | Yes |
| Auto-update | `enabled` | Yes |

If the user installs the Runtime via the standard script (`curl -sSL https://jphat.net/install | bash`), all defaults apply. The Runtime starts, connects to the default relay, and begins accepting pairing requests.

### When Configuration Is Needed

Configuration is required only for:
1. Custom relay server (self-hosted deployment)
2. Restricted file access (enterprise, classroom)
3. Alternative network configuration (proxy, custom DNS)
4. Disabling features (privacy-conscious deployments)

### Configuration Discovery

When a configuration option is missing from the file, the Runtime logs the default value being used at debug level:

```
[DEBUG] Using default relay URL: wss://jphat.net/buzzpi/relay/ws
[DEBUG] Using default log level: info
```

This makes it clear what values are in effect without requiring the user to read a config file.

### First-Run Experience

On first run (no identity key, no config file):
1. Runtime generates identity key
2. Prints pairing code to stdout
3. Connects to default relay server
4. Waits for pairing

The user never sees a configuration editor, never answers setup questions, never reads a configuration guide.

## User Experience

A user installs the Runtime on a fresh Pi. One command, 15 seconds. The Runtime starts and prints a pairing code. The user opens the BuzzPi app, taps "Pair Device", enters the code, and the device is paired. Total time from command to paired: under a minute. No configuration files were created, read, or edited.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Some users need custom configuration | Configuration is supported — it's just optional. Documentation covers common customizations. |
| Defaults may not fit all environments | The defaults are chosen for the 90% case (home users with standard networks). Enterprise/educational deployments can customize. |
| Hidden defaults can confuse troubleshooting | The config file is automatically created with comments when first customization is made. All active settings are visible in the app's "Device Info" screen. |

## Examples

- Runtime installation: one command, no config prompts
- BuzzPi app: no setup wizard, no "Configure your account" screen
- Extension installation: no "Edit this path" — extensions install to the default directory
- Update behavior: auto-update is on by default; users opt out explicitly

## Related Patterns

- [Progressive Disclosure](progressive-disclosure.md): Configuration options are hidden in "Advanced Settings"
- [Explain, Don't Expose](explain-dont-expose.md): When configuration is needed, the UI explains what each option does
