# BPP Chapter 22: Plugin System

**Layer:** Capabilities  
**Status:** Draft  
**Version:** 1.0.0

The Plugin System defines how the Runtime can host third-party code that extends device functionality. Plugins run as isolated processes or in a sandboxed environment.

## Overview

Plugins are the mechanism for extending the Runtime with new capabilities. A plugin is an independent binary or script that communicates with the Runtime via a defined IPC protocol.

## Plugin Manifest

Every plugin includes a manifest file that describes its identity, capabilities, and requirements:

```yaml
# /var/lib/buzzpi/extensions/docker-manager/manifest.yaml
id: docker-manager
name: Docker Manager
version: 1.0.0
author: BuzzPi Core
description: Manage Docker containers and images on your device
license: MIT
homepage: https://buzzpi.dev/extensions/docker-manager

runtime:
  type: binary
  entrypoint: ./docker-manager
  args: ["--config", "./config.yaml"]
  
permissions:
  - docker:read
  - docker:write
  
capabilities:
  methods:
    - name: docker.containers.list
      type: request_response
    - name: docker.containers.logs
      type: stream
    - name: docker.images.pull
      type: request_response
  
systemd_services:
  - docker.service
  
resources:
  max_memory_mb: 64
  max_cpu_percent: 25
  allowed_paths:
    - /var/run/docker.sock
    - /usr/bin/docker
  allowed_capabilities: []
  
lifecycle:
  autostart: true
  start_order: 10
  dependencies:
    - docker.service
  health_check:
    type: process
    interval_seconds: 30
    timeout_seconds: 5
```

### Manifest Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique plugin identifier (lowercase, hyphen-separated) |
| `name` | Yes | Human-readable name |
| `version` | Yes | Semantic version |
| `runtime.type` | Yes | `binary`, `script`, `native` (compiled into Runtime) |
| `runtime.entrypoint` | Yes | Path to the plugin executable |
| `permissions` | Yes | List of required permissions |
| `capabilities.methods` | Yes | RPC methods this plugin registers |
| `resources` | Yes | Resource limits |
| `lifecycle` | Yes | Startup/health/restart behavior |

## Plugin Lifecycle

```
                     ┌────────────┐
                     │            │
               ┌────▶│ INSTALLED  │
               │     │            │
               │     └────────────┘
               │           │
               │      Start trigger
               │      (autostart or manual)
               │           │
               │           v
               │     ┌────────────┐
               │     │            │
               │     │ STARTING   │
               │     │            │
               │     └────────────┘
               │           │
               │      Process started
               │      + health check passes
               │           │
               │           v
               │     ┌────────────┐
          ┌────┴────▶│            │
          │          │ RUNNING    │
          │          │            │
          │          └────────────┘
          │                │
     Health check      Stop trigger
     fails (3x)       (manual/update)
          │                │
          v                v
     ┌────────────┐   ┌────────────┐
     │            │   │            │
     │  FAILED    │   │ STOPPING   │
     │            │   │            │
     └────────────┘   └────────────┘
          │                │
     Auto-restart      Process exited
     (configurable)        │
          │                v
          └────────────┐────────────┐
                       │            │
                       │  STOPPED   │
                       │            │
                       └────────────┘
```

### States

| State | Description |
|-------|-------------|
| INSTALLED | Plugin is installed but not running |
| STARTING | Plugin process is being launched |
| RUNNING | Plugin process is running and health checks pass |
| STOPPING | Plugin is being gracefully shut down |
| STOPPED | Plugin process has exited |
| FAILED | Plugin process crashed or health check failed repeatedly |

### Lifecycle Events

| Event | Trigger | Action |
|-------|---------|--------|
| Install | User installs extension | Download plugin, verify signature, extract to extensions directory |
| Start | Autostart or manual | Execute entrypoint with config |
| Stop | Manual or update | Send SIGTERM, wait 5s, SIGKILL |
| Restart | Manual or config change | Stop then start |
| Update | User updates extension | Stop, replace binary, start |
| Uninstall | User removes extension | Stop, remove plugin directory |

## IPC Protocol

Plugins communicate with the Runtime via stdin/stdout using JSON-line protocol:

```
Runtime → Plugin:   {"method":"rpc.call","params":{"method":"docker.containers.list","params":{},"rid":"req_001"}}
Plugin → Runtime:   {"method":"rpc.result","params":{"rid":"req_001","result":{"containers":[]}}}
```

### Message Types

| Direction | Message | Description |
|-----------|---------|-------------|
| Runtime → Plugin | `rpc.call` | Invoke a method on the plugin |
| Plugin → Runtime | `rpc.result` | Return a method result |
| Plugin → Runtime | `rpc.event` | Push an event to the Runtime |
| Runtime → Plugin | `plugin.configure` | Update plugin configuration |
| Runtime → Plugin | `plugin.stop` | Graceful shutdown request |
| Plugin → Runtime | `plugin.ready` | Plugin initialization complete |
| Plugin → Runtime | `plugin.log` | Log message from plugin |

### Plugin Initialization

```
Runtime                          Plugin
  │                                │
  │  Spawn process                 │
  │ ──────────────────────────────►│
  │                                │── Load configuration
  │                                │── Register methods
  │                                │── Open resources
  │                                │
  │  plugin.ready                  │
  │ ◀──────────────────────────────│
  │                                │
  │  (plugin is now ready)         │
```

## Sandboxing

### Process Isolation

| Property | Enforcement |
|----------|-------------|
| Filesystem | Plugin runs in its own directory; access to other paths must be declared in manifest |
| Network | Plugin shares host network (namespace inherited from Runtime, configurable) |
| Processes | Plugin runs as a separate process (not in Runtime's process group) |
| User | Plugin runs as `buzzpi-plugin` user (restricted permissions) |

### Resource Limits

| Resource | Default | Configurable |
|----------|---------|-------------|
| Max memory | 64 MB | In manifest |
| Max CPU | 25% of one core | In manifest |
| Max open files | 32 | In manifest |
| Max child processes | 10 | In manifest |
| Max disk writes | 10 MB/minute | System-wide |
| Max runtime | None | In manifest (for batch plugins) |

## Plugin Store

Plugins are distributed through the BuzzPi Extension Registry (a service of BuzzPi Cloud):

1. Developer publishes plugin to `registry.buzzpi.dev`
2. Plugin is signed with the developer's key
3. User installs from the app's extension browser
4. Runtime downloads, verifies, and extracts the plugin
5. Plugin manifest is validated before first start

### Package Format

```
docker-manager-1.0.0.bpk          # BuzzPi Plugin Package
├── manifest.yaml                  # Plugin manifest
├── docker-manager                 # Binary (ELF, arm64)
├── config.yaml                    # Default configuration
├── assets/                        # Static assets (icons, etc.)
│   └── icon.svg
└── signature.sig                  # Ed25519 signature (over all files)
```

## Security

| Concern | Mitigation |
|---------|------------|
| Malicious plugin | Plugin binary is signed; signature verified before execution |
| Privilege escalation | Plugin runs as restricted user; permissions declared in manifest |
| Resource exhaustion | Resource limits enforced by cgroups (if available) or ulimit |
| Data exfiltration | Network access is host-inherited but can be restricted |
| Crash loop | Auto-restart is limited (max 3 attempts in 5 minutes) |
| Unauthorized install | Plugin installation requires user confirmation |
