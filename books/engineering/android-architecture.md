# Android Architecture

**The first BPP client.** The BuzzPi Android app is a native Kotlin application that connects to Runtime-equipped devices via the BPP protocol. It is the primary interface for users to discover, pair with, and control their Raspberry Pis.

---

## Design Principles

1. **No networking knowledge required** — Users never enter an IP address, port, or URL. Discovery and connection are fully automatic.
2. **Offline LAN is first-class** — The app works fully on local networks without internet. Cloud features enhance, never gate.
3. **Capability-driven UI** — The interface dynamically adapts to what each device supports. A Pi 5 with GPIO shows a pin control tab. A Pi Zero without Docker hides the containers tab.
4. **Resilient connection** — Network transitions (WiFi → cellular) do not interrupt terminal sessions or file transfers.
5. **UI follows BPP** — Every screen maps to a protocol capability. No client-specific features that aren't exposed through BPP.

---

## Tech Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Kotlin | Modern, coroutine-friendly, first-class for Android |
| UI Framework | Jetpack Compose | Declarative, reactive, state-driven |
| Navigation | Compose Navigation | Type-safe, deep link support |
| DI | Hilt | Industry standard, Google-maintained |
| Networking | OkHttp + Pion/jlibp2p | WebSocket client + WebRTC support |
| Serialization | Kotlinx Serialization | JSON envelopes + protobuf for binary |
| State Management | Kotlin StateFlow | Reactive, lifecycle-aware, testable |
| Local Storage | DataStore + Room | Preferences (DataStore) + device cache (Room) |
| Camera | CameraX | First-party camera API for Pi camera streaming |
| Image Loading | Coil | Compose-native, memory-efficient |
| mDNS Discovery | JmDNS (custom fork) | Zero-conf local device discovery |

---

## Application Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                        UI Layer (Compose)                         │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│ │ Discovery│ │  Device  │ │Terminal  │ │  Files   │ │ Screen │ │
│ │  Screen  │ │  Detail  │ │ Screen   │ │ Manager  │ │Streamer│ │
│ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └───┬────┘ │
│      │            │            │            │            │       │
├──────┴────────────┴────────────┴────────────┴────────────┴───────┤
│                    ViewModel Layer (Hilt)                         │
│ ┌──────────┐ ┌────────────────┐ ┌──────────┐ ┌───────────────┐ │
│ │Discovery │ │DeviceViewModel │ │Terminal  │ │ScreenViewModel│ │
│ │ViewModel │ │                │ │ViewModel │ │               │ │
│ └────┬─────┘ └───────┬────────┘ └────┬─────┘ └──────┬────────┘ │
│      │               │               │               │          │
├──────┴───────────────┴───────────────┴───────────────┴──────────┤
│                    Domain Layer (UseCases)                        │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│ │Discover  │ │ Pair     │ │ConnectTo │ │ Transfer │ │Stream  │ │
│ │Devices   │ │ Device   │ │ Device   │ │  File    │ │Screen  │ │
│ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └───┬────┘ │
│      │            │            │            │            │       │
├──────┴────────────┴────────────┴────────────┴────────────┴───────┤
│                    Data Layer                                     │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ ┌──────┐ │
│ │Connection│ │ Device   │ │ BPP      │ │WebRTC P2P  │ │Local │ │
│ │Manager   │ │Repository│ │Client    │ │Peer        │ │DB    │ │
│ └──────────┘ └──────────┘ └──────────┘ └────────────┘ └──────┘ │
│                    ┌─────────────────────┐                       │
│                    │  mDNS Discovery     │                       │
│                    └─────────────────────┘                       │
└──────────────────────────────────────────────────────────────────┘
```

---

## Module Structure

```
app/
├── app/                          # Application entry, DI, navigation
│   ├── BuzzPiApplication.kt
│   ├── MainActivity.kt
│   ├── navigation/
│   │   └── NavGraph.kt           # Type-safe navigation routes
│   └── di/
│       ├── AppModule.kt          # Hilt: app-wide bindings
│       ├── NetworkModule.kt      # Hilt: OkHttp, WebSocket client
│       └── DatabaseModule.kt     # Hilt: Room database
│
├── discovery/                    # Device discovery (LAN + cloud)
│   ├── DiscoveryScreen.kt        # List of nearby devices
│   ├── DiscoveryViewModel.kt
│   ├── MdnsDiscovery.kt          # mDNS service browser
│   ├── CloudDiscovery.kt         # Relay-registered devices
│   └── DeviceCard.kt             # Card composable for device
│
├── pairing/                      # Device pairing flow
│   ├── PairingScreen.kt          # QR code scan + PIN display
│   ├── PairingViewModel.kt
│   ├── PairingQrCode.kt          # QR code generation/scanning
│   └── PairingService.kt         # BPP pairing protocol client
│
├── device/                       # Connected device interaction
│   ├── DeviceDetailScreen.kt     # Tabbed device interface
│   ├── DeviceViewModel.kt
│   └── tabs/
│       ├── OverviewTab.kt        # System stats, quick actions
│       ├── TerminalTab.kt        # Full terminal emulation
│       ├── FileManagerTab.kt     # File browser + transfer
│       ├── ScreenTab.kt          # Remote desktop stream
│       ├── DockerTab.kt          # Container management
│       ├── GpioTab.kt            # Pin control UI
│       ├── CameraTab.kt          # Camera stream viewer
│       └── LogsTab.kt            # System log viewer
│
├── terminal/                     # Terminal emulator
│   ├── TerminalSession.kt        # PTY session state
│   ├── TerminalRenderer.kt       # ANSI escape sequence renderer
│   ├── TerminalInputHandler.kt   # Keyboard input encoding
│   └── TerminalView.kt           # Compose terminal widget
│
├── screen/                       # Screen streaming
│   ├── ScreenStreamer.kt         # WebRTC media track receiver
│   ├── VideoDecoder.kt           # H.264 hardware decoder
│   ├── TouchInjector.kt          # Touch event → coordinates
│   └── ScreenView.kt             # Compose video surface
│
├── files/                        # File management
│   ├── FileBrowserScreen.kt
│   ├── FileTransferManager.kt    # Upload/download queue
│   └── FilePreviewer.kt          # Preview images, text, etc.
│
├── connection/                   # Connection engine (Android)
│   ├── ConnectionManager.kt      # Connection lifecycle manager
│   ├── WebRtcPeer.kt             # Pion WebRTC wrapper (JNI or pure Kotlin)
│   ├── WebSocketClient.kt        # BPP WebSocket signaling client
│   ├── ReconnectionHandler.kt    # Exponential backoff + reconnect
│   └── ConnectionState.kt        # StateFlow: disconnected → connecting → connected → degrading
│
├── data/                         # Data layer
│   ├── local/
│   │   ├── BuzzPiDatabase.kt     # Room database
│   │   ├── DeviceDao.kt          # Cached devices
│   │   └── SessionDao.kt         # Session persistence
│   └── repository/
│       ├── DeviceRepository.kt   # Device CRUD + caching
│       └── PairingRepository.kt  # Pairing state management
│
└── common/                       # Shared utilities
    ├── ui/
    │   ├── theme/                # Material3 theme (from design tokens)
    │   ├── components/           # Shared composables
    │   └── AnimationSpecs.kt     # Motion language definitions
    ├── extensions/
    │   └── FlowExt.kt            # Kotlin Flow utilities
    └── model/
        └── BppEnvelope.kt        # BPP message envelope model
```

---

## Navigation Architecture

The app uses a single-activity architecture with Compose Navigation and type-safe routes:

```kotlin
sealed class Route(val route: String) {
    data object Splash : Route("splash")
    data object Onboarding : Route("onboarding")
    data object Discovery : Route("discovery")
    data object Pairing : Route("pairing/{deviceId}") {
        fun create(deviceId: String) = "pairing/$deviceId"
    }
    data object DeviceDetail : Route("device/{deviceId}") {
        fun create(deviceId: String) = "device/$deviceId"
    }
    data object FilePreview : Route("device/{deviceId}/file/{path}") {
        fun create(deviceId: String, path: String) = "device/$deviceId/file/${Uri.encode(path)}"
    }
    data object Settings : Route("settings")
}
```

### Navigation Flow

```
                    ┌─────────┐
                    │ Onboarding│  (first launch only)
                    └────┬─────┘
                         │
                    ┌────▼─────┐
                    │ Discovery │◄────── Main entry point
                    │  Screen  │
                    └───┬──┬───┘
                        │  │
              ┌─────────┘  └─────────┐
              │                      │
         ┌────▼────┐          ┌──────▼──────┐
         │ Pairing │          │ DeviceDetail│
         │ QR/PIN  │          │  (Tabbed)   │
         └─────────┘          └──────┬──────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
               ┌────▼───┐    ┌──────▼─────┐   ┌──────▼──────┐
               │Terminal│    │File Manager│   │Screen Stream│
               └────────┘    └──────┬─────┘   └─────────────┘
                                    │
                              ┌─────▼─────┐
                              │FilePreview│
                              └───────────┘
```

---

## State Management

All state flows through Kotlin `StateFlow` from ViewModels to Compose UI:

```kotlin
// DiscoveryViewModel
@HiltViewModel
class DiscoveryViewModel @Inject constructor(
    private val discoverDevices: DiscoverDevicesUseCase,
    private val getPairedDevices: GetPairedDevicesUseCase
) : ViewModel() {

    data class UiState(
        val localDevices: List<DiscoveredDevice> = emptyList(),
        val cloudDevices: List<PairedDevice> = emptyList(),
        val isScanning: Boolean = true,
        val error: String? = null
    )

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            discoverDevices().collect { device ->
                _uiState.update { state ->
                    state.copy(localDevices = state.localDevices + device)
                }
            }
        }
    }
}
```

### Connection State Machine

```
                    ┌───────────┐
                    │ DISCONNECTED│
                    └─────┬─────┘
                          │ discover / scan
                    ┌─────▼─────┐
                    │ SCANNING  │◄─────────────┐
                    └─────┬─────┘              │
                          │ device found       │
                    ┌─────▼─────┐              │
                    │  FOUND    │              │
                    └─────┬─────┘              │
                          │ tap to connect     │
                    ┌─────▼─────┐              │
                    │CONNECTING │              │
                    └─────┬─────┘              │
                     ┌────┴────┐               │
                     │         │               │
               ┌─────▼──┐ ┌───▼────┐           │
               │CONNECTED│ │ FAILED │───────────┘
               └─────┬──┘ └────────┘ (retry)
                     │ disconnect / lost
               ┌─────▼─────┐
               │RECONNECTING│────► CONNECTED or FAILED
               └───────────┘
```

---

## Connection Management

```kotlin
class ConnectionManager @Inject constructor(
    private val webSocketClient: WebSocketClient,
    private val webRtcPeer: WebRtcPeer,
    private val reconnectionHandler: ReconnectionHandler
) {
    private val _connectionState = MutableStateFlow<ConnectionState>(ConnectionState.Disconnected)
    val connectionState: StateFlow<ConnectionState> = _connectionState

    suspend fun connect(deviceId: String): Result<Unit> {
        _connectionState.value = ConnectionState.Connecting

        // Phase 1: WebSocket signaling to relay
        webSocketClient.connect(deviceId)
            .onFailure { return Result.failure(it) }

        // Phase 2: WebRTC peer connection
        return webRtcPeer.negotiate(webSocketClient.sdpOffer, webSocketClient.sdpAnswer)
            .map {
                _connectionState.value = ConnectionState.Connected
                startMonitoring()
            }
    }

    private fun startMonitoring() {
        viewModelScope.launch {
            webSocketClient.heartbeatFlow.collect { pingMs ->
                if (pingMs > 3000) {
                    _connectionState.value = ConnectionState.Degrading
                }
            }
        }
    }
}
```

### Reconnection Flow

```kotlin
class ReconnectionHandler {
    private val config = ReconnectionConfig(
        initialDelay = 1000L,     // 1 second
        maxDelay = 30_000L,       // 30 seconds
        maxAttempts = 10,
        multiplier = 2.0          // exponential backoff
    )

    suspend fun reconnect(deviceId: String, connectionManager: ConnectionManager): Boolean {
        var delay = config.initialDelay
        for (attempt in 1..config.maxAttempts) {
            delay(delay)
            val result = connectionManager.connect(deviceId)
            if (result.isSuccess) return true
            delay = (delay * config.multiplier).toLong().coerceAtMost(config.maxDelay)
        }
        return false
    }
}
```

---

## Terminal Emulation

The terminal subsystem is the most complex UI component. It must render ANSI escape sequences efficiently on mobile hardware.

```kotlin
class TerminalSession(
    private val connection: ConnectionManager,
    private val bufferSize: Int = 4096
) {
    private val _lines = MutableStateFlow<List<TerminalLine>>(emptyList())
    val lines: StateFlow<List<TerminalLine>> = _lines

    private val parser = AnsiParser()  // Parses ANSI escape sequences
    private val screen = TerminalScreen(rows = 40, cols = 80)

    fun write(data: ByteArray) {
        parser.parse(data) { command ->
            screen.apply(command) // Update internal grid model
            _lines.value = screen.render() // Emit renderable lines
        }
    }

    fun sendInput(text: String) {
        viewModelScope.launch {
            connection.send("terminal.input", mapOf("data" to text))
        }
    }
}
```

### Terminal Screen Model

```kotlin
class TerminalScreen(val rows: Int, val cols: Int) {
    private val buffer: ArrayList<CellState> = ArrayList(rows * cols)

    data class CellState(
        val char: Char = ' ',
        val fg: Color = Color.WHITE,
        val bg: Color = Color.BLACK,
        val bold: Boolean = false,
        val italic: Boolean = false,
        val underline: Boolean = false
    )

    fun apply(command: AnsiCommand) {
        when (command) {
            is AnsiCommand.PrintChar -> printChar(command.c)
            is AnsiCommand.CursorMove -> cursorMove(command.dx, command.dy)
            is AnsiCommand.ClearScreen -> clear()
            is AnsiCommand.SetStyle -> setStyle(command.style)
            is AnsiCommand.ScrollUp -> scrollUp(command.lines)
            is AnsiCommand.SetCursor -> setCursor(command.row, command.col)
        }
    }
}
```

### Rendering Strategy

- **VirtualScroll**: Only render visible rows + 2 buffer rows above/below
- **Span-based**: Adjacent cells with the same style collapse into text spans
- **Diff-based**: Only emit Compose recomposition when the rendered grid changes
- **Monospace font cache**: Pre-rendered glyph bitmaps for common monospace fonts

---

## Screen Streaming

```kotlin
class ScreenStreamer(
    private val webRtcPeer: WebRtcPeer,
    private val decoder: VideoDecoder
) {
    private val _frameRate = MutableStateFlow<Int>(0)
    val frameRate: StateFlow<Int> = _frameRate

    suspend fun startStreaming() {
        webRtcPeer.onVideoFrame { buffer ->
            decoder.decode(buffer) { bitmap ->
                // Emit decoded frame to UI
                _frameRate.value = decoder.fps
            }
        }
    }

    fun sendTouchEvent(x: Float, y: Float, action: TouchAction) {
        val normalized = TouchEvent(
            x = (x / displayWidth * 1920).toInt(),   // Map to device resolution
            y = (y / displayHeight * 1080).toInt(),
            action = action
        )
        webRtcPeer.sendMediaCommand("screen.input", normalized)
    }
}
```

### Hardware Decoder Selection

```kotlin
fun selectVideoDecoder(): VideoDecoder {
    return when {
        hasMediaCodecH264() -> MediaCodecDecoder()  // Most devices
        else -> SoftwareDecoder()                    // Fallback
    }
}

private fun hasMediaCodecH264(): Boolean {
    return MediaCodecList(MediaCodecList.REGULAR_CODECS)
        .codecInfos
        .any { it.name.contains("h264", ignoreCase = true) && !it.isEncoder }
}
```

---

## Error Handling

Every network operation returns `Result<T>` and exposes failures to the UI layer for user-facing error messages per the Experience book's error philosophy.

```kotlin
sealed class ConnectionError(val message: String, val isRetryable: Boolean) {
    data object DeviceNotFound : ConnectionError("Device not found on network", true)
    data object PairingRejected : ConnectionError("Device rejected pairing", false)
    data object RelayUnreachable : ConnectionError("Relay server unreachable", true)
    data object WebRtcNegotiationFailed : ConnectionError("Could not establish secure connection", true)
    data object Timeout : ConnectionError("Connection timed out", true)
    data class Unknown(val cause: Throwable) : ConnectionError(cause.message ?: "Unknown error", true)
}
```

---

## Testing Strategy

| Test Type | Scope | Tooling |
|-----------|-------|---------|
| Unit tests | ViewModels, UseCases, Repository | JUnit 5 + MockK |
| UI tests | Compose screens (stateless) | Compose UI Test |
| Integration | WebSocket client ↔ test relay server | MockWebServer (OkHttp) |
| Integration | WebRTC P2P ↔ test peer | Custom test harness |
| Screenshot | Visual regression for composables | Roborazzi |
| End-to-end | Full disconnect/scan/pair/connect flow | Real device + emulator |

### Key Test Scenarios

1. **Discovery**: mDNS announcements appear within 3 seconds
2. **Pairing**: QR code scan completes pairing within 10 seconds
3. **Connection**: P2P established within 5 seconds on LAN
4. **Reconnection**: Network drop of 15s restores session
5. **Terminal**: 1000 lines per second rendering without frame drops
6. **Screen stream**: <500ms latency on LAN

---

## Minimum Requirements

| Requirement | Target |
|-------------|--------|
| **Min API Level** | API 26 (Android 8.0) |
| **Target API Level** | API 34+ (Android 14+) |
| **Memory** | <150MB baseline, <300MB with screen streaming |
| **Storage** | <50MB APK |
| **Permissions** | INTERNET, ACCESS_WIFI_STATE, CHANGE_WIFI_MULTICAST_STATE, BLUETOOTH_SCAN, READ_EXTERNAL_STORAGE (file upload/download), CAMERA (QR scan), POST_NOTIFICATIONS |
| **Architecture** | arm64-v8a, armeabi-v7a, x86_64 (emulator) |
