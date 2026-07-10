# BPP Chapter 23: Extensions

**Layer:** Capabilities  
**Status:** Draft  
**Version:** 1.0.0

Extensions are the user-facing packaging of plugins. While plugins are the technical runtime component, extensions are the complete package — plugin, UI components, configuration, and documentation.

## Overview

An extension can add:
- New capabilities to a device (via plugins)
- New UI to the client (via extension views)
- New action types available in the workspace
- New notification categories
- New automation triggers

## Extension Manifest

Extensions are distributed with a manifest that extends the plugin manifest:

```yaml
# docker-manager extension manifest
id: docker-manager
name: Docker Manager
version: 1.0.0
author: BuzzPi Core
description: Manage Docker containers and images on your device
license: MIT

icon: icon.svg
screenshots:
  - screenshot-1.png
  - screenshot-2.png

plugin:
  # Plugin manifest fields (see Chapter 22)

ui:
  workspace_tabs:
    - id: docker
      label: Docker
      icon: docker-icon
      url: "extension://docker-manager/dashboard.html"
      
  settings_screens:
    - id: docker-config
      label: Docker Configuration
      url: "extension://docker-manager/settings.html"
      
  quick_actions:
    - id: docker-prune
      label: Prune Docker
      icon: clean-icon
      confirmation_required: true
      
  notifications:
    - id: docker-container-failed
      label: Container Stopped
      default_priority: warning

automations:
  triggers:
    - id: container-status-change
      label: Container status changes
      params:
        container: string
        status: ["running", "stopped", "failed"]
  actions:
    - id: restart-container
      label: Restart container
      params:
        container: string

permissions:
  - id: docker:read
    label: View Docker containers and images
    description: Allows reading container status, logs, and image list
    danger: false
  - id: docker:write
    label: Manage Docker containers
    description: Allows starting, stopping, and restarting containers
    danger: false
  - id: docker:admin
    label: Administer Docker
    description: Allows image pulls, system prunes, and Compose operations
    danger: true
```

## Extension Types

### Native Extension

A plugin that runs on the device (Chapter 22). Provides device-side capabilities.

### UI Extension

A client-side component that provides user interface without a device plugin. Examples:
- Theme pack (client-only, no device component)
- Notification filter (client-only)
- Keyboard shortcut configuration (client-only)

### Hybrid Extension

Both a device plugin and a client UI component. This is the most common type.

### Bridge Extension

Connects to an external service:
- Home Assistant bridge
- MQTT bridge
- Webhook bridge

## Extension API

Extensions can access BPP APIs on both sides:

### Device-Side API (Plugin)

The Runtime exposes an API to plugins:

```
Runtime API available to plugins:
  - runtime.log(level, message)
  - runtime.get_config(key)
  - runtime.set_config(key, value)
  - runtime.send_event(event_type, payload)
  - runtime.register_method(name, handler)
  - runtime.get_device_info()
  - runtime.get_connected_clients()
  - runtime.call_client(method, params)  // Reverse RPC
```

### Client-Side API (UI Extension)

The client exposes an API to UI extensions loaded in web views:

```javascript
// Available to extension web views
BuzzPiExtension:
  - buzzpi.connect(deviceId)
  - buzzpi.runAction(actionId, params)
  - buzzpi.getDeviceInfo(deviceId)
  - buzzpi.getExtensionConfig()
  - buzzpi.setExtensionConfig(config)
  - buzzpi.showNotification(title, body, priority)
  - buzzpi.on('device:status', callback)
  - buzzpi.on('extension:event', callback)
```

### Extension Context

Extensions receive context about the current state:

```json
{
  "context": {
    "device": {
      "id": "018e0a3f-...",
      "name": "Kitchen Pi",
      "online": true
    },
    "user": {
      "id": "b2c3d4e5-...",
      "preferences": {
        "theme": "dark"
      }
    },
    "extension": {
      "id": "docker-manager",
      "version": "1.0.0",
      "config": {
        "auto_prune": true,
        "prune_interval_hours": 24
      }
    }
  }
}
```

## Extension Distribution

### Registry

Extensions are distributed through the BuzzPi Extension Registry:

| Endpoint | Purpose |
|----------|---------|
| `registry.buzzpi.dev/v1/extensions` | List all extensions |
| `registry.buzzpi.dev/v1/extensions/{id}` | Extension details |
| `registry.buzzpi.dev/v1/extensions/{id}/download/{version}` | Download package |
| `registry.buzzpi.dev/v1/extensions/{id}/versions` | Version history |

### Installation Flow

```
Client                          Relay Server                     Device
  │                                 │                              │
  │  Browse extensions              │                              │
  │ ◀───────────────────────────────│                              │
  │                                 │                              │
  │  Install: docker-manager        │                              │
  │ ───────────────────────────────►│                              │
  │                                 │  relay.extension.install     │
  │                                 │ ────────────────────────────►│
  │                                 │                              │── Download
  │                                 │                              │── Verify
  │                                 │                              │── Extract
  │                                 │                              │── Register
  │                                 │  relay.extension.installed   │
  │                                 │ ◀────────────────────────────│
  │  Extension installed            │                              │
  │ ◀───────────────────────────────│                              │
  │                                 │                              │
  │  Grant permissions?             │                              │
  │ ───── User confirms ───────────►│                              │
  │                                 │  relay.extension.start       │
  │                                 │ ────────────────────────────►│
  │                                 │                              │── Start plugin
  │                                 │  relay.extension.started     │── Register methods
  │                                 │ ◀────────────────────────────│
```

### Update Flow

```
1. Extension registry notifies Relay Server of new version
2. Relay Server notifies clients: extension.update.available
3. Client prompts user to update
4. User confirms
5. Runtime downloads new version
6. Runtime stops old plugin
7. Runtime extracts new plugin
8. Runtime starts new plugin
9. Runtime unregisters old methods, registers new methods
10. Client is notified: extension.updated
```

## Extension Permissions

Extensions declare required permissions in their manifest. When installing, the user must grant these permissions:

| Permission | Risk | Description |
|------------|------|-------------|
| `terminal:read` | Low | Read terminal output |
| `terminal:write` | Medium | Send input to terminal |
| `files:read` | Medium | Read files on the device |
| `files:write` | High | Modify or delete files |
| `docker:read` | Low | List containers and images |
| `docker:write` | Medium | Start, stop, restart containers |
| `docker:admin` | High | Pull images, system prune, manage Compose |
| `gpio:read` | Low | Read GPIO pin states |
| `gpio:write` | High | Write GPIO pin states, configure PWM |
| `camera:read` | Medium | Access camera preview and snapshots |
| `camera:write` | High | Start/stop recording |
| `network:read` | Low | Read network configuration |
| `network:admin` | High | Modify network configuration |
| `system:read` | Low | Read system information and logs |
| `system:admin` | Critical | Restart/shutdown device, manage services |

Permissions are granular and composable. An extension SHOULD request the minimum set of permissions needed.

## Extension Categories

| Category | Description | Example |
|----------|-------------|---------|
| System | Device management | Docker, GPIO, Camera, Services |
| Monitoring | Observability | Temperature graph, network monitor |
| Automation | Scheduled tasks | Backup, cleanup, health check |
| Integration | External services | Home Assistant, MQTT, Slack |
| Development | Coding tools | Code editor, Git client, CI/CD |
| Media | Multimedia | Camera, audio, screensaver |
| Network | Connectivity | WiFi scanner, VPN, DNS |
