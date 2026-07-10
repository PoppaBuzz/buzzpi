# RFC-0004: Plugin Architecture

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | RFC-0002 |

## Summary

Define the BuzzPi Plugin Architecture — how plugins are developed, packaged, installed, managed, and communicate with the Runtime. Plugins extend BuzzPi with new capabilities (Docker, GPIO, Pi-hole, Home Assistant, etc.) without modifying the core Runtime.

## Motivation

The Runtime ships with a fixed set of built-in capabilities (screen, terminal, device info). Everything else — Docker management, GPIO control, camera streaming, service monitoring — should be pluggable. A well-designed plugin architecture:

1. **Decouples** capability development from the Runtime release cycle
2. **Lowers the barrier** for community contributions (plugins can be in any language)
3. **Enables a plugin ecosystem** — a marketplace of interoperable capabilities
4. **Preserves reliability** — plugin crashes don't affect core Runtime

Without a plugin architecture, every new capability either bloats the Runtime or requires forking the project.

## Design

### 1. Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                  BuzzPi Runtime                  │
│  ┌───────────────────────────────────────────┐  │
│  │              Plugin Host                   │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐   │  │
│  │  │ Plugin  │  │ Plugin  │  │ Plugin  │   │  │
│  │  │ Manager │  │ Process │  │ Registry│   │  │
│  │  │         │  │ Watcher │  │ Cache   │   │  │
│  │  └────┬────┘  └────┬────┘  └────┬────┘   │  │
│  └───────┼────────────┼────────────┼────────┘  │
│          │            │            │            │
│  ┌───────┼────────────┼────────────┼────────┐  │
│  │       │            │            │         │  │
│  │  ┌────┴────┐  ┌────┴────┐  ┌────┴────┐   │  │
│  │  │ Docker  │  │  GPIO   │  │ Pi-hole │   │  │
│  │  │ Plugin  │  │ Plugin  │  │ Plugin  │   │  │
│  │  │ (Go)    │  │(Python) │  │ (Rust)  │   │  │
│  │  └─────────┘  └─────────┘  └─────────┘   │  │
│  │                                           │  │
│  │         Plugin Processes (sub-processes)    │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### 2. Plugin Manifest

Every plugin bundles a `plugin.yaml` manifest. This is the authoritative definition of the plugin's identity, capabilities, and permissions.

```yaml
# plugin.yaml — required at plugin root
id: docker-manager
name: Docker Manager
version: "1.2.0"
author:
  name: Jane Smith
  email: jane@example.com
  url: https://github.com/jane

description: "Manage Docker containers on your Pi"
homepage: https://github.com/jane/buzzpi-docker
license: MIT
min_runtime_version: "0.5.0"     # Minimum BuzzPi Runtime version

executable:
  command: ./docker-manager        # Relative to plugin directory
  args: ["--config", "config.yaml"]
  env:
    DOCKER_HOST: "unix:///var/run/docker.sock"

capabilities:
  - id: container.list
    name: List Containers
    description: List all Docker containers
    timeout: 10s
  - id: container.logs
    name: Container Logs
    description: View container logs
    timeout: 30s
    streaming: true                # This capability uses server-sent events
  - id: container.exec
    name: Execute in Container
    description: Run a command inside a container
    timeout: 60s
    streaming: true

permissions:
  network:
    level: registry                # none | registry | full
    domains:
      - docker.io
  filesystem:
    paths:
      - /var/run/docker.sock       # Docker socket
      - /var/lib/docker            # Docker data (read-only)
    read_only: true
  system: false

resources:
  max_memory_mb: 128
  max_cpu_percent: 25
  startup_timeout_seconds: 30

lifecycle:
  persistent: true                 # Keep alive after capability call
  restart_on_crash: true
  max_restarts: 3
  restart_window: 5m

tags:
  - docker
  - containers
  - devops

install_hooks:
  post_install: "./scripts/setup.sh"
  pre_remove: "./scripts/cleanup.sh"
```

### 3. IPC Protocol

Plugins communicate with the Runtime via **JSON-RPC 2.0 over stdin/stdout**.

#### Message Format

```
Content-Length: 123\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"capability.invoke","params":{...}}
```

#### Runtime → Plugin (Requests)

The Runtime sends capability invocation requests:

```json
{
  "jsonrpc": "2.0",
  "id": "req_a1b2c3",
  "method": "capability.invoke",
  "params": {
    "capability": "container.list",
    "session_id": "sess_abc123",
    "params": {
      "all": true
    }
  }
}
```

#### Plugin → Runtime (Responses)

```json
{
  "jsonrpc": "2.0",
  "id": "req_a1b2c3",
  "result": {
    "containers": [
      {
        "id": "abc123",
        "name": "nginx",
        "status": "running",
        "image": "nginx:latest"
      }
    ]
  }
}
```

#### Plugin → Runtime (Events)

Plugins can push events to the Runtime, which are forwarded to connected clients:

```json
{
  "jsonrpc": "2.0",
  "method": "event.push",
  "params": {
    "type": "container.crash",
    "data": {
      "container_id": "abc123",
      "exit_code": 137
    }
  }
}
```

#### Plugin → Runtime (Logging)

```json
{
  "jsonrpc": "2.0",
  "method": "log",
  "params": {
    "level": "info",
    "message": "Docker plugin initialized",
    "fields": {
      "containers": 5
    }
  }
}
```

### 4. Plugin Lifecycle

```
        ┌────────────┐
        │ Installed   │  — plugin.yaml copied to plugin directory
        └─────┬──────┘
              │ verify
        ┌─────▼──────┐
        │ Verified    │  — manifest validated, permissions checked
        └─────┬──────┘
              │ start (on Runtime boot or demand)
        ┌─────▼──────┐
        │ Starting    │  — process spawned, health check
        └─────┬──────┘
              │ health check OK
        ┌─────▼──────┐
        │ Running     │  — accepting capability requests
        ├─────┬──────┤
        │     │ idle  │  — no active requests, persistent plugin
        │     │ busy  │  — executing capability
        │     │ error │  — recoverable error, retry
        └─────┴──────┘
              │ stop (Runtime shutdown, user disable)
        ┌─────▼──────┐
        │ Stopped     │
        └────────────┘
```

**Startup:**
1. Plugin Manager reads `plugin.yaml`
2. Validates manifest against schema
3. Spawns plugin process with declared permissions
4. Performs health check (sends `health.ping`, expects `health.pong` within configured timeout)
5. Registers capabilities in the Runtime's capability map
6. Notifies connected clients of new capabilities

**Shutdown:**
1. Plugin Manager sends `lifecycle.shutdown` with `graceful_timeout`
2. Plugin cleans up resources (closes connections, saves state)
3. Plugin sends `lifecycle.shutdown_ack`
4. If plugin does not respond within timeout, Plugin Manager sends SIGKILL
5. Capabilities are unregistered from the Runtime

**Crash recovery:**
1. Plugin Manager detects process exit
2. If `restart_on_crash` and under max restart threshold, restart with exponential backoff
3. If over threshold, mark plugin as `crashed` and notify admin client
4. User must manually re-enable the plugin

### 5. Plugin SDK

BuzzPi provides SDKs for plugin development:

#### Go SDK

```go
package main

import (
    "github.com/buzzpi/sdk-go"
)

type DockerPlugin struct{}

func (p *DockerPlugin) Manifest() sdk.PluginManifest {
    return sdk.PluginManifest{
        ID:      "docker-manager",
        Name:    "Docker Manager",
        Version: "1.0.0",
    }
}

func (p *DockerPlugin) Invoke(ctx context.Context, req *sdk.InvokeRequest) (*sdk.InvokeResult, error) {
    switch req.Capability {
    case "container.list":
        return p.listContainers(ctx, req.Params)
    case "container.logs":
        return p.streamLogs(ctx, req.Params)
    }
    return nil, sdk.ErrCapabilityNotFound
}

func main() {
    sdk.Serve(&DockerPlugin{})
}
```

The Go SDK handles:
- stdin/stdout JSON-RPC framing
- Health check responses
- Graceful shutdown on SIGTERM
- Structured logging to Runtime
- Event push channel

#### Python SDK

```python
from buzzpi_sdk import Plugin, serve

class GpioPlugin(Plugin):
    @property
    def manifest(self):
        return {
            "id": "gpio-control",
            "name": "GPIO Control",
            "version": "1.0.0",
        }

    def invoke(self, capability, params, session_id):
        if capability == "gpio.read":
            return {"value": self.read_pin(params["pin"])}
        elif capability == "gpio.write":
            self.write_pin(params["pin"], params["value"])
            return {"success": True}

serve(GpioPlugin())
```

### 6. Permissions Model

Plugins declare the permissions they need. The Runtime enforces these at the OS level.

| Permission | What It Allows | Enforcement |
|------------|---------------|-------------|
| `network.none` | No network access | Process starts in network namespace with loopback only |
| `network.registry` | Access to declared domains only | iptables/nftables rules per process |
| `network.full` | Full network access | No restriction |
| `filesystem` | Access to declared paths only | Landlock (Linux 5.13+) or seccomp |
| `system` | Process creation, signal sending | seccomp policy |
| `gpio` | `/dev/gpiomem`, `/sys/class/gpio` | Device cgroup |

#### Permission Denial Response

When a plugin attempts an action beyond its declared permissions, the Runtime:

1. Denies the action immediately
2. Logs the denial with context
3. Notifies admin client (optional, configurable)
4. Does not crash the plugin

### 7. Plugin Packaging

Plugins are distributed as **`.buzzpi-plugin`** bundles (tar.gz with a `.buzzpi-plugin` extension):

```
docker-manager-1.2.0.buzzpi-plugin/
├── plugin.yaml           # Manifest (required)
├── docker-manager        # Executable binary (required)
├── config.yaml           # Default config (optional)
├── scripts/
│   ├── setup.sh          # Post-install hook (optional)
│   └── cleanup.sh        # Pre-remove hook (optional)
├── assets/
│   └── icon.png          # Plugin icon (optional)
└── docs/
    └── README.md         # Plugin documentation (recommended)
```

**Installation:**
```bash
buzzpi plugin install ./docker-manager-1.2.0.buzzpi-plugin
# or from registry
buzzpi plugin install docker-manager
```

**Registry URL:** `https://plugins.buzzpi.dev/v1/{plugin-id}/{version}`

### 8. Built-in vs Plugin Capabilities

| Capability | Built-in | Justification |
|------------|----------|---------------|
| Screen capture | ✅ | Performance-critical, hardware-specific encoding |
| Terminal | ✅ | Low-level PTY access, needs tight integration |
| Device info/stats | ✅ | Core Runtime identity |
| File manager | ✅ | Security-sensitive path sandboxing |
| GPIO | ✅ | Hardware access, many plugins depend on it |
| Camera | ✅ | Hardware-specific encoding, real-time |
| Docker | ❌ | Plugin — separate concern, large dependency |
| Pi-hole | ❌ | Plugin — specific to Pi-hole users |
| Home Assistant | ❌ | Plugin — specific to HA users |
| Custom sensors | ❌ | Plugin — use case specific |

### 9. Plugin Discovery

The Runtime discovers plugins by scanning the plugin directory:

```
/var/lib/buzzpi/plugins/
├── docker-manager/
│   ├── plugin.yaml
│   └── docker-manager
├── gpio-custom/
│   ├── plugin.yaml
│   └── gpio-custom.py
└── pihole/
    ├── plugin.yaml
    └── pihole
```

**Discovery flow:**
1. Plugin Manager watches the plugin directory with `fsnotify`
2. On new directory: read `plugin.yaml`, verify, launch
3. On modified `plugin.yaml`: re-read, re-verify, restart plugin
4. On directory removal: stop plugin, unregister capabilities

### 10. Plugin Configuration

Plugins can expose configuration via the Runtime:

```json
// Runtime → Plugin: config.get
{"jsonrpc":"2.0","id":2,"method":"config.get","params":{"key":"docker_host"}}

// Plugin → Runtime: config.get response
{"jsonrpc":"2.0","id":2,"result":{"value":"unix:///var/run/docker.sock"}}
```

Configuration is stored in the Runtime's State Store and exposed to the user through the Android app's plugin settings UI.

---

## Drawbacks

1. **IPC overhead** — Every capability invocation goes through JSON serialization and process boundary. Sub-millisecond operations (simple GPIO toggling) may suffer. Mitigation: hot-path capabilities (GPIO, quick reads) can be built-in; plugins are for higher-latency operations.

2. **Plugin ecosystem fragmentation** — Multiple plugin languages means varying quality. A Python plugin may leak memory in ways a Go plugin would not. Mitigation: the certification process catches common issues; the sandbox limits damage.

3. **Distribution complexity** — Binary plugins must be compiled for each architecture (ARM64, ARM, AMD64). A Python plugin needs the user to have the right Python version. Mitigation: official plugins are pre-compiled; language-runtime plugins declare their dependency in the manifest.

---

## Rationale

1. **Why sub-process over WASM?** Sub-process isolation is stronger (OS-level boundaries), language-agnostic, and debuggable with standard tools. WASM sandboxing is better for memory safety but limits system access (no socket access, no filesystem for Docker plugin). A Docker plugin cannot work in WASM without extensive host function bindings.

2. **Why JSON-RPC over gRPC?** JSON-RPC's simplicity makes plugin development accessible. A plugin can be written in 20 lines of Python without code generation. gRPC is appropriate for internal Runtime services but adds friction for external plugin authors.

3. **Why `event.push` over Pub/Sub?** Simpler mental model. The plugin sends events to the Runtime; the Runtime fans out to connected clients. Plugins don't manage subscriptions. This is sufficient for the initial ecosystem and can evolve to a Pub/Sub model if needed.

---

## Prior Art

- **Home Assistant** — `custom_components` directory, Python-only, YAML config. Inspires our plugin directory scanning and capability registration.
- **VS Code Extensions** — `package.json` manifest, activation events, extensible API. Inspires our capability-based model.
- **Docker Plugins** — Socket activation, plugin discovery. Inspires our sub-process lifecycle.
- **Kong Plugins** — Plugin chains with priority ordering. Inspires our capability namespace approach.

---

## Unresolved Questions

1. **Plugin updates** — Should the Runtime auto-update plugins (like VS Code) or require explicit action (like Docker)? Leaning toward explicit for v0.x with optional auto-update in v1.0.

2. **Plugin dependencies** — Can a plugin depend on another plugin? E.g., a "Smart Home" plugin that depends on the "GPIO" plugin. This adds significant complexity — deferred to v1.0.

3. **Plugin UI** — Can plugins extend the Android app UI? This would require a plugin UI framework (like Home Assistant's Lovelace cards). Deferred to v1.0; initially plugins are capability-only.

---

## Implementation Plan

| Phase | Milestone | Details |
|-------|-----------|---------|
| P0 | Plugin Host | Sub-process spawning, stdin/stdout IPC, health checks |
| P1 | Go SDK | `sdk-go` package with `Serve()`, capability routing, logging |
| P2 | Manifest system | `plugin.yaml` validation, capability registration |
| P3 | Permissions | seccomp/landlock enforcement, filesystem sandboxing |
| P4 | Registry | plugin.buzzpi.dev, `buzzpi plugin install/search/update` commands |
| P5 | Python SDK | `buzzpi-sdk` PyPI package, event push support |
| P6 | Ecosystem | Plugin certification process, CI for plugin authors |

---

## References

- RFC-0002: Runtime Architecture (Plugin Host section)
- Engineering Book: plugin-system.md, capability-model.md
- Community Book: plugin-certification.md
- Reference: plugin-manifest.md, plugin-api-reference.md
- Protocol: BPP event.push specification
