# Plugin Manifest Reference

**Complete specification for the plugin manifest file.** Every BuzzPi plugin must include a `manifest.yaml` file. This document defines the schema, all fields, validation rules, and extension points.

---

## Schema Version

Current schema version: `1.0`

Manifests declare their schema version in the root:

```yaml
schema_version: "1.0"
```

---

## Full Schema

```yaml
# Plugin Manifest v1.0
#
# All fields are required unless marked [optional].

plugin:
  # Unique identifier (reverse domain notation)
  # Must match: ^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$
  id: "com.example.weather"

  # Human-readable name [optional, defaults to last segment of ID]
  name: "Weather Station"

  # Semantic version string
  version: "1.0.0"

  # Author information [optional]
  author:
    name: "Jane Doe"           # [optional]
    email: "jane@example.com"   # [optional]
    url: "https://example.com"  # [optional]

  # Short description (max 200 characters)
  description: "Read temperature, humidity, and pressure from a connected I2C sensor"

  # Long description (markdown) [optional]
  # If not provided, README.md in plugin directory is used as fallback
  readme: "README.md"

  # Icon [optional]
  # 48x48 PNG, relative to manifest directory
  icon: "assets/icon.png"

  # Entry point configuration
  runtime:
    # Plugin type: "process" (default) | "container" [future]
    type: "process"

    # Command to start the plugin process
    # First element is the executable, rest are arguments
    command: ["python3", "main.py"]

    # Working directory [optional, defaults to manifest directory]
    working_dir: "."

    # Environment variables [optional]
    env:
      SENSOR_BUS: "i2c-1"
      SENSOR_ADDR: "0x76"
      LOG_LEVEL: "info"

    # Kill signal [optional, default: "SIGTERM"]
    stop_signal: "SIGTERM"

    # Grace period before SIGKILL [optional, default: 10s]
    stop_timeout_seconds: 10

  # Plugin capabilities
  capabilities:
    # Each capability this plugin provides
    - id: "extension.weather.readings"
      version: "1.0"
      description: "Read current temperature, humidity, and pressure"
      permissions: ["sensor.read"]
      ui:                    # [optional] Client rendering hints
        component: "WeatherWidget"
        placement: "overview_tab"   # overview_tab | tab | widget

    - id: "extension.weather.history"
      version: "1.0"
      description: "Historical sensor data (last 24h)"
      permissions: ["sensor.read", "storage.write"]
      ui:
        component: "WeatherHistoryTab"
        placement: "tab"

  # System capabilities this plugin requires
  requires:
    - "hardware.gpio"         # Requires GPIO pins
    - "transport.websocket"   # Requires WebSocket transport

  # Permissions this plugin requests
  permissions:
    - id: "sensor.read"
      description: "Read from connected I2C/SPI sensors"
      risk: "low"

    - id: "storage.write"
      description: "Write data to plugin storage directory"
      risk: "low"

    - id: "network"
      description: "Make outbound HTTP requests"
      risk: "medium"

  # Plugin dependencies [optional]
  dependencies:
    - id: "com.example.sensor-driver"
      version: ">=1.0"
      optional: false

  # Tags for discovery [optional]
  tags:
    - "weather"
    - "sensor"
    - "i2c"

  # License [optional]
  license: "MIT"

  # Homepage URL [optional]
  homepage: "https://github.com/example/weather-plugin"

  # Repository URL [optional]
  repository: "https://github.com/example/weather-plugin"

  # Donation URLs [optional]
  funding:
    - "https://github.com/sponsors/example"
```

---

## Field Validation Rules

### plugin.id

- Reverse domain notation: `com.example.plugin-name`
- Must match: `^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`
- Maximum length: 64 characters
- Reserved prefixes: `com.buzzpi`, `io.buzzpi`, `org.buzzpi`

### plugin.version

- Semantic version: `MAJOR.MINOR.PATCH`
- Must match: `^\d+\.\d+\.\d+$`
- Pre-release tags not supported in current schema

### plugin.runtime.command

- Array of strings, minimum length 1
- First element must be an executable available in the plugin directory or system PATH
- Maximum argument count: 32
- Maximum total command length: 1024 characters

### capabilities[].id

- Plugin-specific ID within the `extension.{plugin_id}.` namespace
- Must match: `^extension\.[a-z0-9.]+\.[a-z][a-zA-Z0-9_]*$`
- The plugin_id part must match `plugin.id` exactly

### permissions[].id

- Built-in permissions are predefined (see [Permissions Registry](#permissions-registry))
- Custom permissions use prefix: `ext.{plugin_id}.{permission_name}`
- Must match: `^[a-z][a-zA-Z0-9_]*(\.[a-z][a-zA-Z0-9_]*)*$`
- Maximum length: 64 characters

---

## Permissions Registry

### Built-in Permissions

| ID | Description | Risk | Default |
|----|-------------|------|---------|
| `sensor.read` | Read from GPIO, I2C, SPI | low | grant |
| `sensor.write` | Write to GPIO, I2C, SPI (can control relays, motors) | medium | prompt |
| `filesystem.read` | Read files outside plugin storage | medium | prompt |
| `filesystem.write` | Write files outside plugin storage | high | deny |
| `network` | Make outbound HTTP/WebSocket connections | medium | prompt |
| `network.listen` | Listen on a TCP/UDP port | high | deny |
| `process.spawn` | Execute arbitrary subprocesses | high | deny |
| `storage.read` | Read plugin's own storage directory | low | grant |
| `storage.write` | Write to plugin's own storage directory | low | grant |

### Risk Levels

| Risk | User Prompt | Auto-grant |
|------|-------------|------------|
| `low` | Never | Always |
| `medium` | On first use | After first grant |
| `high` | Every time | Never |

---

## Capability UI Placement

| Placement | Description | Example |
|-----------|-------------|---------|
| `overview_tab` | Embedded widget in the overview tab | Weather widget showing temperature |
| `tab` | Dedicated tab in device detail screen | Separate "Weather" tab |
| `widget` | Widget on the device list or home screen | Quick-glance temperature on device card |
| `settings` | Settings panel entry | Plugin configuration |
| `context_menu` | Long-press context menu action | "Read sensor now" action |

---

## Minimal Manifest

```yaml
schema_version: "1.0"
plugin:
  id: "com.example.minimal"
  version: "1.0.0"
  description: "A minimal plugin example"
  runtime:
    command: ["./plugin.bin"]
  capabilities:
    - id: "extension.minimal.hello"
      version: "1.0"
      description: "Says hello"
      permissions: []
  permissions: []
```

---

## Validation

The Runtime validates plugin manifests on install:

```bash
# Validate a plugin manifest
buzzpi-runtime plugin validate ./path/to/manifest.yaml

# Expected output on success:
# ✓ manifest.yaml: valid (schema v1.0)

# Expected output on failure:
# ✗ manifest.yaml: validation failed
#   - plugin.id: must match reverse domain notation
#   - plugin.version: must be semver (got "1.0")
```

### Validation Rules Summary

| Rule | Description |
|------|-------------|
| Schema version | Must be a supported version |
| Plugin ID | Must match reverse domain notation |
| Plugin version | Must be semantic version |
| Command | Must be non-empty array of strings |
| Capability ID | Must be in `extension.{plugin_id}.{name}` namespace |
| Permission ID | Must be built-in or custom permission format |
| Dependencies | Referenced plugin IDs must be valid |
| File references | Referenced files (icon, readme) must exist |
| No namespace collision | Capability IDs must not conflict with existing plugins |
| Resource limits | Max declared limits must be within system bounds |
