# Capability Model

**How BuzzPi discovers, negotiates, and adapts to device capabilities.** The capability model is the core mechanism that enables a single app to work seamlessly across diverse devices — from a Pi Zero with limited hardware to a Pi 5 with GPIO, camera, Docker, and hardware-accelerated video encoding.

---

## Principles

1. **Assume nothing, discover everything** — No client assumes a device supports any capability. Every feature is gated behind capability negotiation.
2. **One protocol, many devices** — The same BPP messages work across all device types. Different capabilities are expressed through negotiation, not protocol forks.
3. **Safe degradation** — If a capability disappears mid-session (e.g., camera disconnected), the client degrades gracefully instead of crashing.
4. **Extensible** — Third-party extensions can introduce new capabilities without modifying core BPP or the client app.
5. **UI follows capability** — The Android/Kotlin Multiplatform client renders UI dynamically based on what the device reports. No hardcoded feature lists per device model.

---

## Capability Definition

```yaml
capability:
  id: "screen.stream"              # Fully qualified BPP method prefix
  version: "1.0"                   # Semver for this capability
  kind: service                    # service | hardware | extension | transport
  description: "Remote desktop streaming"

  # Dependencies on other capabilities
  requires:
    - "transport.webrtc"           # Screen streaming requires WebRTC media tracks
    - "video.encoder"              # H.264 encoder required

  # Negotiation parameters
  params:
    max_fps: 30
    max_resolution: "1920x1080"
    codecs: ["h264", "vp8"]

  # Client rendering hints
  ui:
    component: "ScreenTab"         # Compose component name
    icon: "screen_stream"
    label: "Screen"
    priority: high                 # Tab ordering priority
```

---

## Capability Taxonomy

```
capabilities/
├── transport/
│   ├── websocket                  # Signaling channel available
│   ├── webrtc                     # WebRTC data/media channels
│   └── relay                      # Relay server connectivity
│
├── hardware/
│   ├── gpio                       # Physical GPIO pins
│   ├── camera                     # Camera module attached
│   ├── video.encoder              # Hardware H.264 encoder (MMAL/V4L2)
│   ├── audio                      # Audio output/input
│   ├── display                    # Physical display connected
│   └── hdmi_cec                   # HDMI CEC control
│
├── service/
│   ├── terminal                   # PTY terminal access
│   ├── screen                     # Screen capture (DRM/fbdev/X11)
│   ├── files                      # File system browsing
│   ├── docker                     # Docker management
│   ├── systemd                    # Systemd service management
│   ├── stats                      # System monitoring (CPU/RAM/temp)
│   ├── logs                       # System log access
│   ├── camera.stream              # Camera live streaming
│   ├── bluetooth                  # Bluetooth device management
│   └── wifi                       # WiFi configuration
│
├── extension/
│   ├── plugin.*                   # Third-party extensions (namespace)
│   └── buzzai.*                   # AI assistant tools
│
└── meta/
    ├── pairing                    # Support pairing protocols
    ├── onboarding                 # Initial setup wizard
    └── self_update                # Runtime self-update support
```

---

## Negotiation Flow

```
Client                           Device (Runtime)
  │                                    │
  │  1. capabilities.list              │
  │ ──────────────────────────────────►│
  │                                    │
  │  2. capabilities.list.result       │
  │ ◄──────────────────────────────────│
  │    {                                │
  │      capabilities: [               │
  │        {id: "transport.webrtc",    │
  │         version: "1.0",            │
  │         available: true},          │
  │        {id: "service.terminal",    │
  │         version: "2.1",            │
  │         available: true},          │
  │        {id: "service.screen",      │
  │         version: "1.0",            │
  │         available: true,           │
  │         params: {                  │
  │           capture_method: "drm",   │
  │           max_fps: 30,             │
  │           max_resolution: "1920x1080",
  │           encoder: "mmal"}
  │         },                         │
  │        {id: "hardware.gpio",       │
  │         version: "1.0",            │
  │         available: false},         │
  │        {id: "service.docker",      │
  │         version: "1.0",            │
  │         available: false}          │
  │      ]                             │
  │    }                                │
  │                                    │
  │  3. Build dynamic UI               │
  │     based on available ∩            │
  │     client support                  │
  │                                    │
```

### Protocol Method

```yaml
capabilities.list:
  description: Returns all capabilities supported by the device
  request: {}
  response:
    capabilities:
      - id: string           # Fully qualified capability ID
        version: string      # Semver
        available: boolean   # Currently available (hardware present, service running)
        params: object       # Capability-specific parameters (optional)

capabilities.subscribe:
  description: Subscribe to capability change events
  request:
    events:
      - "capability.added"
      - "capability.removed"
      - "capability.updated"
  response:
    subscription_id: string

capabilities.event:
  description: Pushed when a capability changes
  payload:
    type: "capability.added" | "capability.removed" | "capability.updated"
    capability:
      id: string
      available: boolean
      params: object (optional)
```

---

## Client-Side Capability Registry

```kotlin
// CapabilityRegistry.kt — Android client
class CapabilityRegistry {
    private val capabilities = MutableStateFlow<Map<String, Capability>>(emptyMap())

    // Subscribe to live capability state
    val availableTabs: StateFlow<List<DeviceTab>> = capabilities.map { caps ->
        TabRegistry.allTabs.filter { tab ->
            caps[tab.requiredCapability]?.available == true
        }
    }

    fun update(deviceCapabilities: List<Capability>) {
        capabilities.value = deviceCapabilities.associateBy { it.id }
    }

    fun refresh(connection: BppClient) {
        scope.launch {
            val result = connection.request("capabilities.list")
            if (result is BppResult.Success) {
                update(result.data.capabilities)
            }
        }
    }
}

// Tab definitions
object TabRegistry {
    val allTabs = listOf(
        DeviceTab("Overview",  "overview",   "stats",          priority = 0),
        DeviceTab("Terminal",  "terminal",   "service.terminal", priority = 1),
        DeviceTab("Screen",    "screen",     "service.screen",   priority = 2),
        DeviceTab("Files",     "files",      "service.files",    priority = 3),
        DeviceTab("Docker",    "docker",     "service.docker",   priority = 4),
        DeviceTab("GPIO",      "gpio",       "hardware.gpio",    priority = 5),
        DeviceTab("Camera",    "camera",     "hardware.camera",  priority = 6),
        DeviceTab("Logs",      "logs",       "service.logs",     priority = 7),
    )
}

data class DeviceTab(
    val label: String,
    val route: String,
    val requiredCapability: String,
    val priority: Int,
    val badge: ((Stats) -> String?)? = null  // Optional badge (e.g., alert count)
)
```

---

## Dynamic UI Rendering

```kotlin
@Composable
fun DeviceDetailScreen(viewModel: DeviceViewModel) {
    val tabs by viewModel.availableTabs.collectAsState()
    val selectedTab by viewModel.selectedTab.collectAsState()

    Scaffold(
        topBar = { DeviceTopBar(viewModel.device) }
    ) {
        Column {
            // Dynamic tab row based on capabilities
            ScrollableTabRow(selectedTabIndex = tabs.indexOf(selectedTab)) {
                tabs.forEach { tab ->
                    Tab(
                        selected = tab == selectedTab,
                        onClick = { viewModel.selectTab(tab) },
                        icon = { Icon(tab.icon()) },
                        text = { Text(tab.label) }
                    )
                }
            }

            // Content area
            when (selectedTab?.route) {
                "overview"  -> OverviewTab(viewModel)
                "terminal"  -> TerminalTab(viewModel.terminalSession)
                "screen"    -> ScreenTab(viewModel.screenStreamer)
                "files"     -> FileManagerTab(viewModel.fileManager)
                "docker"    -> DockerTab(viewModel.dockerService)
                "gpio"      -> GpioTab(viewModel.gpioService)
                "camera"    -> CameraTab(viewModel.cameraStreamer)
                "logs"      -> LogsTab(viewModel.logService)
            }
        }
    }
}
```

---

## Extension Capabilities

Third-party extensions register their capabilities through a plugin manifest:

```yaml
# Extension manifest (extensions/my-extension/manifest.yaml)
extension:
  id: "com.example.my-extension"
  name: "My Extension"
  version: "1.0.0"

capabilities:
  - id: "extension.my-extension.weather"
    version: "1.0"
    description: "Weather station data from connected sensor"
    ui:
      component: "WeatherWidget"
      placement: "overview_tab"     # Can embed in existing views
    params:
      sensors: ["temperature", "humidity", "pressure"]

  - id: "extension.my-export.heating"
    version: "1.0"
    description: "Heating controller"
    ui:
      component: "HeatingTab"
      placement: "tab"              # Gets its own tab in device detail
    requires:
      - "hardware.gpio"
```

### API for Extension Capabilities

```go
// Runtime extension capability registration
type ExtensionCapability struct {
    ID          string   `yaml:"id"`
    Version     string   `yaml:"version"`
    Description string   `yaml:"description"`
    Requires    []string `yaml:"requires"`
}

// ExtensionManager aggregates all capabilities from loaded extensions.
func (m *ExtensionManager) AggregateCapabilities() []Capability {
    var caps []Capability
    for _, ext := range m.loaded {
        for _, ec := range ext.Manifest.Capabilities {
            caps = append(caps, Capability{
                ID:          ec.ID,
                Version:     ec.Version,
                Available:   m.dependenciesMet(ec.Requires),
                Description: ec.Description,
            })
        }
    }
    return caps
}

// MergeExtensionCaps merges extension capabilities into the device's
// capability list before sending to the client.
func MergeExtensionCaps(base []Capability, ext []Capability) []Capability {
    // Base capabilities go first, then extension capabilities sorted by ID
    existing := make(map[string]bool)
    for _, c := range base {
        existing[c.ID] = true
    }

    for _, c := range ext {
        if !existing[c.ID] {
            base = append(base, c)
            existing[c.ID] = true
        }
    }

    sort.Slice(base, func(i, j int) bool {
        return base[i].ID < base[j].ID
    })

    return base
}
```

---

## Versioning & Compatibility

```yaml
# Client capability support matrix
client_supports:
  terminal: ">=1.0"
  screen: ">=1.0"
  files: ">=1.0"
  docker: ">=1.0 || 2.0"
  gpio: ">=1.0"
  camera: ">=1.0"

# Negotiation algorithm
negotiate:
  1. Client sends capabilities.supported with its support matrix
  2. Device responds with capabilities.available (intersection of supported ∩ device)
  3. For each capability, select highest mutually supported version
  4. Report capability-specific parameters (codec support, max resolution, etc.)
```

```go
type VersionConstraint struct {
    constraint string // ">=1.0", ">=1.0 || 2.0"
}

func (v VersionConstraint) Matches(version string) bool {
    // Parse semver
    // Evaluate constraint expression
    // Return match result
    return satisfied // boolean
}

func Negotiate(
    clientSupported map[string]string,   // cap_id → constraint
    deviceCapabilities []Capability,
) []Capability {
    var negotiated []Capability

    for _, dc := range deviceCapabilities {
        constraint, exists := clientSupported[dc.ID]
        if !exists {
            continue // Client doesn't support this capability at all
        }

        if VersionConstraint{constraint: constraint}.Matches(dc.Version) {
            negotiated = append(negotiated, dc)
        }
    }

    return negotiated
}
```

---

## Runtime Capability Detection

```go
// DetectCapabilities scans the device and returns all available capabilities.
func DetectCapabilities(ctx context.Context) []Capability {
    var caps []Capability

    // Transport
    caps = append(caps, Capability{ID: "transport.websocket", Available: true})

    // Hardware
    if hasGPIO() {
        caps = append(caps, Capability{ID: "hardware.gpio", Available: true})
    }
    if hasCamera() {
        caps = append(caps, Capability{
            ID: "hardware.camera",
            Available: true,
            Params: map[string]interface{}{
                "max_resolution": "3280x2464",
                "formats":        []string{"mjpeg", "h264"},
            },
        })
    }
    if enc := detectVideoEncoder(); enc != EncoderUnavailable {
        caps = append(caps, Capability{
            ID: "hardware.video.encoder",
            Available: true,
            Params: map[string]interface{}{
                "method": enc.String(),
                "codecs": []string{"h264"},
            },
        })
    }

    // Services
    if hasTerminal() {
        caps = append(caps, Capability{ID: "service.terminal", Available: true})
    }
    if hasDocker() {
        caps = append(caps, Capability{ID: "service.docker", Available: true})
    }
    if hasSystemd() {
        caps = append(caps, Capability{ID: "service.systemd", Available: true})
    }
    // ... detect screen capture, file system, etc.

    return caps
}
```

### Dynamic Capability Changes

```go
// CapabilityMonitor watches for hardware changes and pushes capability events.
type CapabilityMonitor struct {
    changes chan CapabilityEvent
}

func (m *CapabilityMonitor) Start(ctx context.Context) {
    // Watch GPIO via udev
    go m.watchUdev("gpio", "gpio*")

    // Watch camera via v4l2
    go m.watchV4L2()

    // Watch display via DRM
    go m.watchDRM()

    // Watch Docker socket
    go m.watchDockerSocket()
}

func (m *CapabilityMonitor) watchV4L2() {
    // Poll /dev/video* every 5 seconds
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        before := m.cameraPresent
        m.cameraPresent = checkV4L2Device()
        if before != m.cameraPresent {
            m.changes <- CapabilityEvent{
                Type: capabilityChanged,
                Capability: Capability{
                    ID:        "hardware.camera",
                    Available: m.cameraPresent,
                },
            }
        }
    }
}
```

---

## Client-Side Adaptation Examples

### Example 1: Pi 5 with everything

| Capability | Available | Tab Shown |
|------------|-----------|-----------|
| service.terminal | ✓ | Terminal |
| service.screen | ✓ | Screen |
| service.files | ✓ | Files |
| service.docker | ✓ | Docker |
| hardware.gpio | ✓ | GPIO |
| hardware.camera | ✓ | Camera |
| service.logs | ✓ | Logs |

→ Full 8-tab interface

### Example 2: Pi Zero W (headless, no camera, no Docker)

| Capability | Available | Tab Shown |
|------------|-----------|-----------|
| service.terminal | ✓ | Terminal |
| service.files | ✓ | Files |
| hardware.gpio | ✓ | GPIO |
| service.stats | ✓ | Overview |
| service.docker | ✗ | (hidden) |
| hardware.camera | ✗ | (hidden) |
| service.screen | ✗ | (hidden) |

→ 4-tab interface (Overview, Terminal, Files, GPIO)

### Example 3: Screen stream degraded mid-session

```
Capability change event:
{
  type: "capability.updated",
  capability: {
    id: "service.screen",
    available: true,
    params: {
      max_fps: 10,         // Dropped from 30 due to thermal throttling
      capture_method: "fbdev"  // DRM failed, fell back to fbdev
    }
  }
}

→ Client shows toast: "Screen streaming quality reduced (device is throttling)"
→ Screen tab switches to lower resolution
→ Frame rate display shows new target
```

---

## Implementation in the Runtime

```go
// Runtime capability service
type CapabilityService struct {
    device    *DeviceInfo
    monitor   *CapabilityMonitor
    extensions *ExtensionManager
    events    chan CapabilityEvent
}

func (s *CapabilityService) HandleList(msg BppMessage) BppMessage {
    base := DetectCapabilities(context.TODO())
    ext := s.extensions.AggregateCapabilities()
    all := MergeExtensionCaps(base, ext)

    return BppMessage{
        Type:   "response",
        RID:    msg.RID,
        Result: map[string]interface{}{
            "capabilities": all,
        },
    }
}

func (s *CapabilityService) HandleSubscribe(msg BppMessage) BppMessage {
    subID := generateID()
    go func() {
        for event := range s.monitor.changes {
            // Push capability events to subscriber
            s.pushEvent(subID, event)
        }
    }()

    return BppMessage{
        Type:   "response",
        RID:    msg.RID,
        Result: map[string]interface{}{
            "subscription_id": subID,
        },
    }
}
```

---

## Testing the Capability Model

| Test | Scenario | Expectation |
|------|----------|-------------|
| Full capability list | Device with all hardware | All 8+ capabilities returned |
| Minimal device | Pi Zero (headless) | Only terminal, files, GPIO |
| Dynamic removal | Camera unplugged mid-session | `capability.removed` event sent |
| Version negotiation | Old client, new device | Minimum compatible version selected |
| Extension registration | Plugin adds capability | Capability appears in list |
| UI adaptation | Capability list changes | Tabs re-render dynamically |
| Degradation | Screen FPS drops | Quality adaptation triggers |
