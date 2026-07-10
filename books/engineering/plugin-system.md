# Plugin System

**How extensions integrate with BuzzPi Runtime.** The Plugin System is the mechanism by which third-party developers add functionality to a BuzzPi device — new services, hardware integrations, custom UIs, and background automation. Plugins run as isolated child processes and communicate with the Runtime via IPC.

---

## Principles

1. **A plugin crash never takes down the Runtime** — Plugins are isolated processes. They can be killed and restarted without affecting the core Runtime.
2. **No special privileges by default** — Plugins start with minimal permissions. Capabilities must be explicitly requested in the manifest and granted by the user.
3. **IPC over stdio or Unix sockets** — The Runtime communicates with plugins via framed JSON messages over stdin/stdout or Unix domain sockets. No shared memory, no HTTP servers.
4. **Any language, one contract** — Plugins can be written in any language. The only requirement is implementing the BPP plugin IPC protocol.
5. **Versioned capability interface** — Plugins declare what BPP methods they implement. The Runtime routes incoming requests to the correct plugin.
6. **Lazy loading** — Plugins are started on demand when a client requests their capabilities. Idle plugins are terminated after a timeout.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                     BuzzPi Runtime                    │
│                                                       │
│  ┌──────────┐  ┌────────────┐  ┌──────────────────┐ │
│  │ Core     │  │ Capability │  │ Plugin Manager    │ │
│  │ Services │◄─┤ Service    │◄─┤                   │ │
│  │          │  │            │  │ - Start/Stop      │ │
│  │ terminal │  │ - list     │  │ - Health checks   │ │
│  │ screen   │  │ - negotiate│  │ - IPC routing     │ │
│  │ files    │  │ - subscribe│  │ - Restart policy  │ │
│  │ docker   │  └────────────┘  └────────┬──────────┘ │
│  │ stats    │                           │            │
│  └──────────┘                           │            │
└──────────────────────────────────────────┼────────────┘
                                           │
              ┌────────────────────────────┼────────────────────────────┐
              │            stdin/stdout or Unix socket                 │
              │            ┌────────────────────────────┐              │
              │            │     Plugin Process         │              │
              │            │                            │              │
              │            │  ┌──────────────────┐      │              │
              │            │  │ IPC Handler       │      │              │
              │            │  │ - JSON msg loop   │      │              │
              │            │  │ - BPP envelope    │      │              │
              │            │  └────────┬─────────┘      │              │
              │            │           │                │              │
              │            │  ┌────────▼─────────┐      │              │
              │            │  │ Service Logic     │      │              │
              │            │  │ (any language)    │      │              │
              │            │  └──────────────────┘      │              │
              │            └────────────────────────────┘              │
              │                                                       │
              │            ┌────────────────────────────┐              │
              │            │     Plugin Process 2        │              │
              │            │     (or N plugins)          │              │
              │            └────────────────────────────┘              │
              └────────────────────────────────────────────────────────┘
```

---

## Plugin Manifest

Every plugin must include a manifest file that declares its identity, capabilities, permissions, and runtime requirements.

```yaml
# Example: weather sensor plugin
# extensions/weather/manifest.yaml
plugin:
  id: "com.example.weather"
  name: "Weather Station"
  version: "1.0.0"
  author: "Example Developer"
  description: "Read temperature, humidity, and pressure from a connected sensor"

  # Entry point — how the Runtime starts this plugin
  runtime:
    type: process                     # "process" (default), "container" (future)
    command: ["python3", "main.py"]  # Command to start the plugin process
    working_dir: "."                  # Relative to manifest location
    env:
      SENSOR_BUS: "i2c-1"
      SENSOR_ADDR: "0x76"

  # Required capabilities (must be available on the device)
  requires:
    - "hardware.gpio"
    - "transport.websocket"

  # Capabilities this plugin provides
  capabilities:
    - id: "extension.weather.readings"
      version: "1.0"
      description: "Read current sensor values"
      permissions: ["sensor.read"]
      ui:
        component: "WeatherWidget"
        placement: "overview_tab"

    - id: "extension.weather.history"
      version: "1.0"
      description: "Historical sensor data (last 24h)"
      permissions: ["sensor.read", "storage.write"]

  # Permissions this plugin requests
  permissions:
    - id: "sensor.read"
      description: "Read from connected I2C/SPI sensors"
    - id: "storage.write"
      description: "Write data to plugin storage directory"
    - id: "network"
      description: "Make outbound HTTP requests"
```

---

## Plugin Lifecycle

```
           ┌───────────┐
           │ INSTALLED │  (manifest validated, binaries present)
           └─────┬─────┘
                 │ Runtime starts plugin on first capability request
           ┌─────▼─────┐
           │ STARTING  │  (process spawned, health check in progress)
           └─────┬─────┘
        ┌────────┴────────┐
        │                 │
   ┌────▼────┐      ┌────▼────┐
   │ ACTIVE  │      │ FAILED  │  (startup timeout or crash)
   └────┬────┘      └────┬────┘
        │                │
        │ idle timeout   │ retry (max 3)
   ┌────▼────┐           │
   │ IDLE    │───────────┘
   └────┬────┘
        │ capability request
        │
   ┌────▼────┐
   │ ACTIVE  │
   └────┬────┘
        │ user uninstalls
   ┌────▼──────┐
   │ STOPPED   │  (process killed, resources cleaned)
   └───────────┘
```

```go
type PluginState int

const (
    PluginInstalled PluginState = iota
    PluginStarting
    PluginActive
    PluginIdle
    PluginFailed
    PluginStopped
)

type Plugin struct {
    ID          string
    Manifest    PluginManifest
    State       PluginState
    Cmd         *exec.Cmd
    IPC         *IPCConnection
    StartCount  int
    LastActive  time.Time
    HealthFailures int
}
```

### Lifecycle Manager

```go
type PluginManager struct {
    plugins map[string]*Plugin
    config  PluginManagerConfig
}

type PluginManagerConfig struct {
    IdleTimeout     time.Duration // 5 minutes — terminate idle plugins
    StartupTimeout  time.Duration // 30 seconds — process must pass health check
    MaxRestarts     int           // 3 — consecutive failures before disabling
    HealthInterval  time.Duration // 15 seconds — periodic health pings
}

func (pm *PluginManager) StartPlugin(id string) error {
    plugin, exists := pm.plugins[id]
    if !exists {
        return fmt.Errorf("plugin %s not installed", id)
    }

    if plugin.State == PluginActive || plugin.State == PluginStarting {
        return nil // Already running or starting
    }

    plugin.State = PluginStarting
    ctx, cancel := context.WithTimeout(context.Background(), pm.config.StartupTimeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, plugin.Manifest.Runtime.Command[0],
        plugin.Manifest.Runtime.Command[1:]...)
    cmd.Dir = plugin.Manifest.Dir()
    cmd.Env = pm.buildEnv(plugin)

    // Create IPC connection over stdin/stdout
    ipc := NewIPCConnection(cmd)
    plugin.IPC = ipc
    plugin.Cmd = cmd

    if err := cmd.Start(); err != nil {
        plugin.State = PluginFailed
        return fmt.Errorf("failed to start plugin %s: %w", id, err)
    }

    // Wait for health check
    if err := ipc.WaitForHealth(ctx); err != nil {
        cmd.Process.Kill()
        plugin.State = PluginFailed
        plugin.StartCount++
        return fmt.Errorf("plugin %s health check failed: %w", id, err)
    }

    plugin.State = PluginActive
    plugin.LastActive = time.Now()
    plugin.HealthFailures = 0

    // Start health monitoring goroutine
    go pm.monitorHealth(plugin)

    return nil
}

func (pm *PluginManager) StopPlugin(id string) error {
    plugin := pm.plugins[id]
    if plugin.State == PluginStopped || plugin.State == PluginInstalled {
        return nil
    }

    plugin.State = PluginStopped
    if plugin.Cmd != nil && plugin.Cmd.Process != nil {
        plugin.Cmd.Process.Signal(syscall.SIGTERM)
        time.Sleep(5 * time.Second)
        plugin.Cmd.Process.Kill()
    }

    plugin.IPC = nil
    plugin.Cmd = nil
    return nil
}
```

---

## IPC Protocol

Plugins communicate with the Runtime over stdin/stdout using length-prefixed JSON messages. The protocol is a strict subset of BPP.

### Message Frame

```
┌─────────────────────────────────────────────┐
│ 4 bytes: payload length (big-endian uint32) │
├─────────────────────────────────────────────┤
│ N bytes: JSON payload (BPP envelope)        │
└─────────────────────────────────────────────┘
```

### IPC Message Types

| Type | Direction | Purpose |
|------|-----------|---------|
| `ipc.hello` | Plugin → Runtime | Initial handshake, declare identity |
| `ipc.health` | Both | Health check ping/pong |
| `ipc.capabilities` | Plugin → Runtime | Declare provided capabilities |
| `ipc.request` | Runtime → Plugin | Forward BPP method request |
| `ipc.response` | Plugin → Runtime | Return result of request |
| `ipc.event` | Plugin → Runtime | Push unsolicited event |
| `ipc.log` | Plugin → Runtime | Log message at given level |
| `ipc.shutdown` | Runtime → Plugin | Graceful termination signal |

### Handshake Sequence

```
Plugin                            Runtime
  │                                  │
  │  1. ipc.hello                    │
  │     {plugin_id, version}         │
  │ ────────────────────────────────►│
  │                                  │── Validate plugin_id matches manifest
  │  2. ipc.hello.ack                │
  │     {runtime_version,            │
  │      session_id}                 │
  │ ◄────────────────────────────────│
  │                                  │
  │  3. ipc.capabilities             │
  │     {capabilities: [...]}        │
  │ ────────────────────────────────►│
  │                                  │── Register capabilities with CapabilityService
  │                                  │
  │  4. ipc.request                  │── Forward incoming BPP requests for
  │     {method, params, rid}        │   plugin's capabilities
  │ ◄────────────────────────────────│
  │                                  │
  │  5. ipc.response                 │
  │     {result, rid}                │
  │ ────────────────────────────────►│── Route response back to client
```

```go
// Plugin IPC implementation (Runtime side)
type IPCConnection struct {
    cmd    *exec.Cmd
    stdin  io.Writer
    stdout io.Reader
    mu     sync.Mutex
    handlers map[string]RequestHandler
}

func (ipc *IPCConnection) ReadMessage() (*IPCMessage, error) {
    // Read 4-byte length prefix
    header := make([]byte, 4)
    if _, err := io.ReadFull(ipc.stdout, header); err != nil {
        return nil, err
    }

    length := binary.BigEndian.Uint32(header)
    if length > maxMessageSize {
        return nil, fmt.Errorf("message too large: %d bytes", length)
    }

    // Read payload
    payload := make([]byte, length)
    if _, err := io.ReadFull(ipc.stdout, payload); err != nil {
        return nil, err
    }

    var msg IPCMessage
    if err := json.Unmarshal(payload, &msg); err != nil {
        return nil, err
    }

    return &msg, nil
}

func (ipc *IPCConnection) SendMessage(msg IPCMessage) error {
    payload, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    ipc.mu.Lock()
    defer ipc.mu.Unlock()

    // Write length prefix + payload
    header := make([]byte, 4)
    binary.BigEndian.PutUint32(header, uint32(len(payload)))

    if _, err := ipc.stdin.Write(header); err != nil {
        return err
    }
    if _, err := ipc.stdin.Write(payload); err != nil {
        return err
    }

    return nil
}
```

### Plugin Side (Python Example)

```python
#!/usr/bin/env python3
"""Example weather sensor plugin."""
import json
import struct
import sys


class BuzzPiPlugin:
    def __init__(self, plugin_id):
        self.plugin_id = plugin_id
        self.stdin = sys.stdin.buffer
        self.stdout = sys.stdout.buffer

    def send_message(self, msg_type, payload):
        msg = {"type": msg_type, **payload}
        data = json.dumps(msg).encode("utf-8")
        header = struct.pack(">I", len(data))
        self.stdout.write(header + data)
        self.stdout.flush()

    def read_message(self):
        header = self.stdin.read(4)
        if len(header) < 4:
            return None
        length = struct.unpack(">I", header)[0]
        data = self.stdin.read(length)
        return json.loads(data)

    def health_check(self):
        self.send_message("ipc.health", {"status": "ok"})

    def handle_request(self, method, params, rid):
        if method == "extension.weather.readings":
            # Read sensor via I2C
            temp = self.read_temperature()
            humidity = self.read_humidity()
            self.send_message("ipc.response", {
                "rid": rid,
                "result": {
                    "temperature": temp,
                    "humidity": humidity,
                    "unit": "celsius"
                }
            })

    def run(self):
        # Handshake
        self.send_message("ipc.hello", {"plugin_id": self.plugin_id, "version": "1"})
        hello_ack = self.read_message()

        # Declare capabilities
        self.send_message("ipc.capabilities", {
            "capabilities": [{
                "id": "extension.weather.readings",
                "version": "1.0",
                "available": True
            }]
        })

        # Message loop
        while True:
            msg = self.read_message()
            if msg is None:
                break
            if msg["type"] == "ipc.health":
                self.health_check()
            elif msg["type"] == "ipc.request":
                self.handle_request(msg["method"], msg.get("params"), msg.get("rid"))
            elif msg["type"] == "ipc.shutdown":
                break

    def read_temperature(self):
        # Actual I2C sensor reading
        return 22.5

    def read_humidity(self):
        return 45.0


if __name__ == "__main__":
    plugin = BuzzPiPlugin("com.example.weather")
    plugin.run()
```

---

## Permissions Model

```go
// Permission represents a single permission a plugin can request.
type Permission struct {
    ID          string `yaml:"id"`
    Description string `yaml:"description"`
}

// PermissionSet represents the permissions granted to a plugin.
type PermissionSet struct {
    Granted     map[string]bool
    Revocable   map[string]bool // Can be revoked at runtime
}

// PermissionEnforcer checks plugin actions against granted permissions.
type PermissionEnforcer struct {
    plugins map[string]PermissionSet
}

func (pe *PermissionEnforcer) Check(pluginID, permission string) error {
    set, exists := pe.plugins[pluginID]
    if !exists {
        return fmt.Errorf("plugin %s not found", pluginID)
    }
    if !set.Granted[permission] {
        return fmt.Errorf("plugin %s does not have permission: %s", pluginID, permission)
    }
    return nil
}
```

### Built-in Permissions

| Permission | Description | Risk |
|------------|-------------|------|
| `sensor.read` | Read from GPIO, I2C, SPI, or similar | Low |
| `sensor.write` | Write to GPIO, I2C, SPI (can control relays, motors) | Medium |
| `filesystem.read` | Read files outside plugin storage | Medium |
| `filesystem.write` | Write files outside plugin storage | High |
| `network` | Make outbound HTTP/WebSocket connections | Medium |
| `network.listen` | Listen on a TCP/UDP port | High |
| `process.spawn` | Execute arbitrary subprocesses | High |
| `storage.read` | Read plugin's own storage directory | Low |
| `storage.write` | Write to plugin's own storage directory | Low |

### Permission Grant Flow

```
1. Plugin manifest declares requested permissions
2. On first start, Runtime checks if permissions were previously granted
3. If not, Runtime sends a permission request to the client
4. Client displays permission dialog to user
5. User grants or denies each permission
6. Runtime stores decision (persisted across restarts)
7. Plugins can check permission grant status at runtime
```

```kotlin
// Permission dialog on Android
@Composable
fun PluginPermissionDialog(
    plugin: PluginInfo,
    permissions: List<Permission>,
    onGrant: (Set<String>) -> Unit,
    onDeny: () -> Unit
) {
    var granted by remember { mutableStateOf(emptySet<String>()) }

    AlertDialog(
        title = { Text("${plugin.name} needs permissions") },
        text = {
            Column {
                permissions.forEach { perm ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Checkbox(
                            checked = perm in granted,
                            onCheckedChange = { checked ->
                                granted = if (checked) granted + perm.id else granted - perm.id
                            }
                        )
                        Column {
                            Text(perm.description, fontWeight = FontWeight.Bold)
                            Text(getRiskLabel(perm), color = getRiskColor(perm))
                        }
                    }
                }
            }
        },
        confirmButton = { TextButton(onClick = { onGrant(granted) }) { Text("Grant") } },
        dismissButton = { TextButton(onClick = onDeny) { Text("Deny") } }
    )
}
```

---

## Resource Controls

Plugins are resource-limited to prevent runaway processes from affecting the system:

```go
type PluginResourceLimits struct {
    MaxCPUPercent   float64 // 50% — single core
    MaxMemoryMB     int     // 128 MB
    MaxDiskMB       int     // 500 MB — plugin storage
    MaxNetworkKbps  int     // 1000 Kbps — outbound bandwidth
    MaxFileDescriptors int  // 64
    MaxChildProcesses   int // 3
}

func applyResourceLimits(cmd *exec.Cmd, limits PluginResourceLimits) {
    // On Linux: use cgroups v2 for CPU and memory limits
    // On other platforms: best-effort ulimit

    cmd.SysProcAttr = &syscall.SysProcAttr{
        // Linux cgroup or rlimit settings
    }
}
```

### Health Monitoring

```go
func (pm *PluginManager) monitorHealth(plugin *Plugin) {
    ticker := time.NewTicker(pm.config.HealthInterval)
    defer ticker.Stop()

    for range ticker.C {
        if plugin.State != PluginActive && plugin.State != PluginIdle {
            return
        }

        // Send health check ping
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        err := plugin.IPC.HealthPing(ctx)
        cancel()

        if err != nil {
            plugin.HealthFailures++
            if plugin.HealthFailures >= 3 {
                pm.handleFailure(plugin, "health check failed 3 times")
                return
            }
        } else {
            plugin.HealthFailures = 0
        }

        // Check resource usage
        if pm.exceedsLimits(plugin) {
            pm.handleFailure(plugin, "resource limit exceeded")
            return
        }
    }
}

func (pm *PluginManager) handleFailure(plugin *Plugin, reason string) {
    plugin.State = PluginFailed
    pm.Logger.Warn("plugin failed", "id", plugin.ID, "reason", reason)

    // Attempt restart
    if plugin.StartCount < pm.config.MaxRestarts {
        go func() {
            time.Sleep(5 * time.Second) // Backoff before restart
            pm.StartPlugin(plugin.ID)
        }()
    } else {
        plugin.State = PluginStopped
        pm.Logger.Error("plugin disabled after max restarts", "id", plugin.ID)
        // Notify user via push notification
    }
}
```

---

## Storage Isolation

Each plugin gets an isolated storage directory:

```
/var/lib/buzzpi/plugins/
└── com.example.weather/
    ├── data/
    │   ├── sensor_history.db
    │   └── config.json
    └── cache/
        └── weather_cache.json
```

```go
func (pm *PluginManager) ensureStorageDir(pluginID string) (string, error) {
    storageDir := filepath.Join(pm.config.StorageRoot, pluginID)

    // Create directories with restrictive permissions
    for _, dir := range []string{"data", "cache"} {
        path := filepath.Join(storageDir, dir)
        if err := os.MkdirAll(path, 0700); err != nil {
            return "", err
        }
    }

    return storageDir, nil
}
```

---

## Plugin Directory Structure

```
extensions/
└── com.example.weather/
    ├── manifest.yaml        # Plugin manifest (required)
    ├── main.py              # Entry point
    ├── requirements.txt     # Language-specific deps
    ├── assets/              # Static assets (icons, etc.)
    │   └── icon.png         # 48x48 plugin icon
    └── README.md            # Human-readable documentation
```

### Installation

```go
// PluginInstaller handles downloading and verifying plugins.
type PluginInstaller struct {
    sources    []PluginSource // Local dirs, remote registries
    verifyKey  crypto.PublicKey
}

type PluginSource struct {
    Type string // "local" | "registry"
    URL  string
}

func (pi *PluginInstaller) Install(id, version string) error {
    // 1. Find plugin in sources
    // 2. Download archive
    // 3. Verify signature
    // 4. Validate manifest schema
    // 5. Extract to /var/lib/buzzpi/plugins/{id}/
    // 6. Verify dependencies
    // 7. Register with PluginManager
    return nil
}

func (pi *PluginInstaller) Uninstall(id string) error {
    // 1. Stop plugin if running
    // 2. Remove plugin directory
    // 3. Remove from PluginManager registry
    // 4. Revoke all permissions
    return nil
}
```

---

## BPP Methods for Plugin Management

```yaml
plugin.list:
  description: List installed plugins
  request: {}
  response:
    plugins:
      - id: string
        name: string
        version: string
        state: "installed" | "starting" | "active" | "failed" | "stopped"
        capabilities: string[]

plugin.install:
  description: Install a plugin from a source
  request:
    id: string
    source: string (optional, default: official registry)
    version: string (optional, default: latest)
  response:
    plugin_id: string
    state: string

plugin.uninstall:
  description: Remove an installed plugin
  request:
    id: string
  response: {}

plugin.start:
  description: Start a stopped/failed plugin
  request:
    id: string
  response:
    state: string

plugin.stop:
  description: Stop a running plugin
  request:
    id: string
  response: {}

plugin.permissions:
  description: List and manage plugin permissions
  request:
    id: string
    action: "list" | "grant" | "revoke"
    permissions: string[] (optional)
  response:
    granted: string[]
    pending: string[] (permissions not yet granted)
```

---

## Testing Strategy

| Test | Scope | Expectation |
|------|-------|-------------|
| Plugin install | Manifest validation, extraction | Plugin registered and ready |
| Plugin start | Process spawn, IPC handshake | Health check passes |
| Plugin IPC request | BPP request forwarded → response returned | Round-trip <100ms |
| Plugin crash | Kill plugin process mid-request | Runtime unaffected, restart triggered |
| Plugin permission deny | Plugin requests without permission | Request rejected with PERMISSION_DENIED error |
| Plugin resource limit | Plugin leaks memory | Plugin killed at 128MB limit |
| Plugin idle timeout | Plugin inactive for 5 minutes | Plugin terminated, restarted on demand |
| Multiple plugins | 10 plugins running concurrently | All functional, no interference |
| Plugin update | Install new version | Old version stopped, new version started |
| Permission persistence | Reboot plugin host | Granted permissions restored |
