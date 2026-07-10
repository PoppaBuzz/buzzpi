# RFC-0007: Android Client Architecture

| Field | Value |
|-------|-------|
| **Status** | Draft |
| **Author** | BuzzPi Architecture Team |
| **Created** | 2026-07-07 |
| **Last Updated** | 2026-07-07 |
| **Requires** | RFC-0001, RFC-0002, RFC-0003, RFC-0006 |

## Summary

Define the architecture of the BuzzPi Android app — the primary client for the BuzzPi Platform. The Android app is the user's gateway to their devices, providing discovery, pairing, remote desktop, terminal, file management, and system monitoring.

## Motivation

The Android app is the most important piece of user-facing software in BuzzPi. It is where the "never type an IP address" promise is delivered. A well-architected app ensures:

1. **Discoverability** — devices appear without configuration
2. **Reliability** — connections survive network changes
3. **Performance** — screen streaming and terminal are responsive
4. **Maintainability** — the architecture supports years of feature additions

Without a clear architecture, the Android app will become unmaintainable as features accumulate.

## Design

### 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Android Application                      │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Presentation Layer (Jetpack Compose + ViewModels)   │   │
│  │                                                      │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │   │
│  │  │DeviceList│ │ Dashboard│ │Terminal  │ │ Screen │ │   │
│  │  │ Screen   │ │ Screen   │ │ Screen   │ │ Viewer │ │   │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬───┘ │   │
│  │       │            │            │            │      │   │
│  │  ┌────┴────────────┴────────────┴────────────┴───┐ │   │
│  │  │              ViewModels                        │ │   │
│  │  └────────────────────┬──────────────────────────┘ │   │
│  └───────────────────────┼────────────────────────────┘   │
│                          │                                 │
│  ┌───────────────────────┼────────────────────────────┐   │
│  │  Domain Layer         │                            │   │
│  │                       │                            │   │
│  │  ┌────────────────────┴────────────────────────┐  │   │
│  │  │            Use Cases / Repositories          │  │   │
│  │  └────────────────────┬────────────────────────┘  │   │
│  └───────────────────────┼────────────────────────────┘   │
│                          │                                 │
│  ┌───────────────────────┼────────────────────────────┐   │
│  │  Data Layer           │                            │   │
│  │                       │                            │   │
│  │  ┌────────────────────┴────────────────────────┐  │   │
│  │  │           Connection Engine (RFC-0001)       │  │   │
│  │  │  ┌──────────┐ ┌──────────┐ ┌────────────┐  │  │   │
│  │  │  │ Transport│ │ Protocol │ │ Session    │  │  │   │
│  │  │  │ Manager  │ │ Handler  │ │ Manager    │  │  │   │
│  │  │  └──────────┘ └──────────┘ └────────────┘  │  │   │
│  │  └────────────────────────────────────────────┘  │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │   │
│  │  │ Local DB │ │DataStore │ │ WebRTC Peer      │ │   │
│  │  │ (Room)   │ │(Settings)│ │ Connection       │ │   │
│  │  └──────────┘ └──────────┘ └──────────────────┘ │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Platform Layer                                   │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │   │
│  │  │ mDNS     │ │ Notifi-  │ │ OS Keychain      │  │   │
│  │  │ Discover │ │ cations  │ │ (EncryptedShared)│  │   │
│  │  └──────────┘ └──────────┘ └──────────────────┘  │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 2. Layered Architecture

BuzzPi Android follows a strict **3-layer architecture** (Presentation → Domain → Data) plus a Platform layer.

#### Presentation Layer

```
com.buzzpi.android.ui
├── screens/
│   ├── discovery/         # Device discovery screen
│   │   ├── DiscoveryScreen.kt       (Composable)
│   │   ├── DiscoveryViewModel.kt    (ViewModel)
│   │   └── components/
│   │       ├── DeviceCard.kt
│   │       └── PairingDialog.kt
│   ├── dashboard/         # Device dashboard (post-pairing)
│   │   ├── DashboardScreen.kt
│   │   ├── DashboardViewModel.kt
│   │   └── components/
│   │       ├── QuickActions.kt
│   │       └── StatusCards.kt
│   ├── terminal/          # Terminal screen
│   │   ├── TerminalScreen.kt
│   │   ├── TerminalViewModel.kt
│   │   └── components/
│   │       ├── TerminalView.kt      (SurfaceView + custom renderer)
│   │       └── ToolbarOverlay.kt
│   ├── screen/            # Screen streaming viewer
│   │   ├── ScreenViewerScreen.kt
│   │   ├── ScreenViewerViewModel.kt
│   │   └── components/
│   │       ├── RemoteDisplayView.kt (SurfaceView + WebRTC renderer)
│   │       └── TouchOverlay.kt
│   ├── files/             # File browser
│   │   ├── FileBrowserScreen.kt
│   │   ├── FileBrowserViewModel.kt
│   │   └── components/
│   │       ├── FileList.kt
│   │       └── TransferProgress.kt
│   ├── system/            # System monitor
│   │   ├── SystemScreen.kt
│   │   ├── SystemViewModel.kt
│   │   └── components/
│   │       ├── CpuGauge.kt
│   │       └── MemoryChart.kt
│   └── settings/          # Settings
│       ├── SettingsScreen.kt
│       └── SettingsViewModel.kt
├── components/            # Shared UI components
│   ├── BuzzPiTopBar.kt
│   ├── BuzzPiBottomBar.kt
│   ├── ConnectionIndicator.kt
│   └── PermissionDialog.kt
├── theme/                 # Material 3 theming
│   ├── Theme.kt
│   ├── Color.kt
│   ├── Type.kt
│   └── Shape.kt
└── navigation/
    ├── NavGraph.kt
    └── Routes.kt
```

**State management:**
- Each screen has a dedicated ViewModel
- ViewModel state is exposed via `StateFlow<UiState>`
- UiState is a sealed interface with `Loading`, `Success`, `Error`, `Empty` variants
- One-shot events (navigation, snackbar) use `Channel<UiEvent>` + `receiveAsFlow()`

**ViewModel pattern:**

```kotlin
data class DiscoveryUiState(
    val devices: List<DeviceInfo> = emptyList(),
    val isScanning: Boolean = true,
    val error: String? = null
)

sealed interface DiscoveryEvent {
    data class NavigateToPairing(val device: DeviceInfo) : DiscoveryEvent
    data class ShowError(val message: String) : DiscoveryEvent
}

@HiltViewModel
class DiscoveryViewModel @Inject constructor(
    private val discoveryRepository: DiscoveryRepository,
    private val connectionManager: ConnectionManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(DiscoveryUiState())
    val uiState: StateFlow<DiscoveryUiState> = _uiState.asStateFlow()

    private val _events = Channel<DiscoveryEvent>(Channel.BUFFERED)
    val events: Flow<DiscoveryEvent> = _events.receiveAsFlow()

    init {
        viewModelScope.launch {
            discoveryRepository.discoveredDevices.collect { devices ->
                _uiState.update { it.copy(devices = devices, isScanning = false) }
            }
        }
    }

    fun startPairing(device: DeviceInfo) {
        viewModelScope.launch {
            // Navigate to pairing screen
            _events.send(DiscoveryEvent.NavigateToPairing(device))
        }
    }
}
```

#### Domain Layer

```
com.buzzpi.android.domain
├── model/
│   ├── Device.kt
│   ├── Session.kt
│   ├── TerminalSession.kt
│   ├── FileEntry.kt
│   ├── SystemStats.kt
│   └── Capability.kt
├── repository/
│   ├── DeviceRepository.kt          (interface)
│   ├── SessionRepository.kt         (interface)
│   ├── TerminalRepository.kt        (interface)
│   ├── FileRepository.kt            (interface)
│   └── SettingsRepository.kt        (interface)
└── usecase/
    ├── DiscoverDevicesUseCase.kt
    ├── PairDeviceUseCase.kt
    ├── OpenTerminalSessionUseCase.kt
    ├── ObserveConnectionStateUseCase.kt
    └── GetDeviceStatsUseCase.kt
```

**Repository interfaces are in the domain layer; implementations are in the data layer.** Use cases orchestrate repository operations.

```kotlin
// Domain model — no Android dependencies
data class Device(
    val deviceId: String,
    val friendlyName: String,
    val platform: String,
    val runtimeVersion: String,
    val capabilities: List<String>,
    val isOnline: Boolean,
    val transport: Transport
)

enum class Transport { LAN, RELAY, P2P }

// Repository interface
interface DeviceRepository {
    val discoveredDevices: Flow<List<Device>>
    val pairedDevices: Flow<List<Device>>
    suspend fun pair(deviceId: String, pin: String): Session
    suspend fun unpair(deviceId: String)
    suspend fun getDeviceInfo(deviceId: String): Device
    suspend fun observeDevice(deviceId: String): Flow<Device>
}
```

#### Data Layer

```
com.buzzpi.android.data
├── connection/
│   ├── ConnectionEngine.kt          (RFC-0001 implementation)
│   ├── TransportManager.kt          (LAN → Relay → P2P)
│   ├── ProtocolHandler.kt           (BPP message serialization)
│   └── SessionManager.kt            (token storage, refresh)
├── repository/
│   ├── DeviceRepositoryImpl.kt
│   ├── SessionRepositoryImpl.kt
│   ├── TerminalRepositoryImpl.kt
│   └── FileRepositoryImpl.kt
├── local/
│   ├── BuzzPiDatabase.kt            (Room)
│   ├── dao/
│   │   ├── DeviceDao.kt
│   │   └── SessionDao.kt
│   └── entity/
│       ├── DeviceEntity.kt
│       └── SessionEntity.kt
├── discovery/
│   ├── MdnsDiscovery.kt             (JmDNS wrapper)
│   └── CloudDiscovery.kt            (Relay API)
├── webrtc/
│   ├── PeerConnectionFactory.kt
│   ├── ScreenStreamRenderer.kt
│   └── InputForwarder.kt
└── sync/
    └── DeviceSyncManager.kt         (periodic background sync)
```

#### Platform Layer

```
com.buzzpi.android.platform
├── security/
│   ├── KeyStoreManager.kt           (EncryptedSharedPreferences)
│   └── BiometricAuth.kt
├── notifications/
│   ├── NotificationChannels.kt
│   └── PushNotificationService.kt   (FCM)
├── permissions/
│   ├── PermissionManager.kt
│   └── PermissionRationale.kt
└── network/
    ├── NetworkStateObserver.kt      (ConnectivityManager)
    └── WifiLockManager.kt           (screen streaming)
```

### 3. Connection Engine (Client Side)

The client-side Connection Engine mirrors the Runtime's engine (RFC-0001). It manages the transport layer and provides a unified interface to the domain layer.

```kotlin
class ConnectionEngine @Inject constructor(
    private val transportManager: TransportManager,
    private val protocolHandler: ProtocolHandler,
    private val sessionManager: SessionManager
) {
    val connectionState: Flow<ConnectionState>

    // Send a BPP request and await response
    suspend fun request(method: String, params: Any?): BppResponse

    // Send a BPP request and receive streaming events
    suspend fun requestStream(method: String, params: Any?): Flow<BppEvent>

    // Push an event (client→device unsolicited)
    suspend fun pushEvent(method: String, params: Any?)
}
```

**Transport priority:**
1. LAN (direct WebSocket, mDNS discovered)
2. P2P (WebRTC data channel, ICE negotiated)
3. Relay (WebSocket via Cloud Relay)

The Connection Engine continuously evaluates the best available transport and transitions seamlessly.

### 4. Screen Streaming

Screen streaming uses WebRTC with a custom video renderer.

```
┌──────────────┐     ┌──────────────────┐     ┌────────────────┐
│  Runtime     │────▶│   Cloud Relay    │────▶│  Android App   │
│  (Screen     │     │   (or direct)    │     │                │
│   Capture)   │     │                  │     │  SurfaceView   │
│              │     │                  │     │  + WebRTC      │
│  H.264/H.265 │     │                  │     │  Renderer      │
└──────────────┘     └──────────────────┘     └────────────────┘
```

**Client-side pipeline:**
1. WebRTC `PeerConnection` receives encoded video frames
2. Custom `VideoRenderer` decodes and renders to `SurfaceView`
3. Touch events are captured via `TouchOverlay` composable
4. Touch coordinates are scaled and sent back as BPP `screen.input` events
5. Quality feedback (estimated bandwidth, packet loss) is sent via BPP `screen.quality_feedback`

**Touch input protocol:**

```kotlin
// Touch event → Runtime
data class TouchEvent(
    val action: Int,       // MotionEvent.ACTION_DOWN/UP/MOVE
    val x: Float,          // 0.0 - 1.0 (normalized)
    val y: Float,          // 0.0 - 1.0 (normalized)
    val pointerId: Int,
    val pressure: Float,
    val timestamp: Long
)

// Keyboard input → Runtime
data class KeyEvent(
    val keyCode: Int,
    val action: Int,       // DOWN / UP
    val modifiers: Int     // bitmask: CTRL, ALT, META
)
```

### 5. Terminal

The terminal is a custom composable that renders ANSI text to a Canvas.

```kotlin
@Composable
fun TerminalView(
    terminalSession: TerminalSessionViewModel,
    modifier: Modifier = Modifier
) {
    val state by terminalSession.displayState.collectAsStateWithLifecycle()

    Canvas(modifier = modifier) {
        // Render character grid with colors
        state.lines.forEachIndexed { row, line ->
            drawText(line, row)
        }
        // Render cursor
        drawCursor(state.cursorPosition, state.cursorVisible)
    }
}
```

**Terminal state:**
```kotlin
data class TerminalDisplayState(
    val lines: List<TerminalLine>,      // Visible lines
    val cursorPosition: Position,
    val cursorVisible: Boolean,
    val scrollbackOffset: Int,
    val dimensions: Dimensions
)

data class TerminalLine(
    val segments: List<StyledSegment>    // Text + foreground + background + bold/italic
)

data class StyledSegment(
    val text: String,
    val fgColor: Color,
    val bgColor: Color,
    val bold: Boolean,
    val italic: Boolean,
    val underline: Boolean
)
```

**Input handling:**
- Tap on the terminal view → show/hide toolbar overlay
- Physical keyboard input → forwarded to Runtime
- On-screen keyboard → custom IME wrapper forwarding keystrokes
- Pinch-to-zoom → font size scaling

### 6. Navigation

```
NavGraph
├── DeviceDiscovery (start destination)
│   ├── DiscoveryScreen
│   └── PairingScreen → PairingDialog
├── DeviceDashboard (deviceId)
│   ├── DashboardScreen
│   ├── TerminalScreen (deviceId, sessionId)
│   ├── ScreenViewerScreen (deviceId)
│   ├── FileBrowserScreen (deviceId, path)
│   ├── SystemScreen (deviceId)
│   └── SettingsScreen
└── Settings
    ├── AppSettingsScreen
    └── PluginSettingsScreen
```

Navigation uses Jetpack Navigation Compose with type-safe arguments:

```kotlin
@Serializable
sealed interface Route {
    @Serializable data object Discovery : Route
    @Serializable data class Pairing(val deviceId: String) : Route
    @Serializable data class Dashboard(val deviceId: String) : Route
    @Serializable data class Terminal(val deviceId: String, val sessionId: String) : Route
    @Serializable data class ScreenViewer(val deviceId: String) : Route
    @Serializable data class FileBrowser(val deviceId: String, val path: String = "/") : Route
    @Serializable data class System(val deviceId: String) : Route
    @Serializable data object AppSettings : Route
}
```

### 7. Background Service

A foreground service maintains the WebSocket connection to the relay when the app is backgrounded.

```kotlin
class BuzzPiConnectionService : Service() {
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        val notification = createOngoingNotification(
            title = "BuzzPi Connected",
            text = "Connected to living-room-pi"
        )
        startForeground(NOTIFICATION_ID, notification)
        return START_STICKY
    }
}
```

The foreground service is active only when there is an active connection (screen streaming or terminal session). When idle, the app uses FCM for push notifications.

### 8. Data Persistence

| Data | Storage | Purpose |
|------|---------|---------|
| Paired devices | Room DB | Device list, display names |
| Session tokens | EncryptedSharedPreferences | Auth for reconnection |
| App settings | DataStore | User preferences |
| Device cache | Room DB | Offline device list |
| Plugin registry | Room DB | Installed plugin cache |

### 9. Permissions

| Permission | When | Why |
|------------|------|-----|
| Internet | Always | Device communication |
| Nearby devices (Wi-Fi) | Discovery | mDNS scanning |
| Foreground service | Active connection | Maintain WebSocket |
| Notifications | Always (opt-out) | Device alerts |
| Camera | When requested | Camera streaming feature |
| Storage | File transfer | Download files from device |
| Biometric | App lock (optional) | Secure app access |

### 10. Error Handling

```kotlin
sealed interface AppError {
    data class ConnectionLost(val deviceId: String) : AppError
    data class PairingFailed(val reason: String) : AppError
    data class SessionExpired(val deviceId: String) : AppError
    data class DeviceOffline(val deviceId: String) : AppError
    data class PermissionDenied(val permission: String) : AppError
    data class Unknown(val message: String) : AppError
}
```

Error handling strategy:
1. Recoverable errors (connection loss) → automatic retry with backoff
2. Session errors → prompt re-authentication
3. Device errors → show user-friendly message with action
4. Unexpected errors → log, show generic message, option to export logs

---

## Drawbacks

1. **Foreground service requirement** — Android restricts background work. Maintaining a persistent WebSocket requires a foreground notification. Mitigation: notification is minimal ("BuzzPi Connected"), dismissed when no active session.

2. **WebRTC complexity** — The screen streaming implementation requires significant WebRTC expertise. Android's WebRTC stack has platform-specific quirks. Mitigation: wrap WebRTC in a dedicated `ScreenStreamManager` with exhaustive testing.

3. **ANR risk on terminal rendering** — Terminal output at high rates (cat /dev/urandom) could overwhelm the Canvas renderer. Mitigation: frame rate limiting, output buffering, and diff-based rendering.

---

## Rationale

1. **Why Jetpack Compose over XML?** Modern Android standard. Faster development, easier theming, type-safe navigation. BuzzPi's design language is expressed naturally in Compose.

2. **Why Room over raw SQLite?** Type-safe DAOs, Flow integration, migration support. Room is the standard Android persistence library.

3. **Why Hilt over manual DI?** BuzzPi has a deep dependency graph. Hilt provides standard DI patterns, ViewModel injection, and testability.

4. **Why StateFlow over LiveData?** StateFlow is lifecycle-aware via `collectAsStateWithLifecycle()`, supports initial values, and integrates with Kotlin coroutines. LiveData is legacy.

---

## Prior Art

- **Tailscale Android** — WireGuard VPN client with foreground service, mDNS-like discovery. Inspires our connection management and background service patterns.
- **Home Assistant Android** — Compose migration, WebSocket connection, notification-driven. Inspires our WebSocket reconnection and notification architecture.
- **Termius** — Terminal emulator with SSH. Inspires our terminal rendering approach.
- **Microsoft Remote Desktop** — Remote desktop client with touch gestures. Inspires our touch input forwarding.

---

## Unresolved Questions

1. **Widget support** — Should BuzzPi offer home screen widgets (quick-connect to favorite device, system stats)? Leaning toward yes for v0.5+.

2. **Wear OS** — Should we support Wear OS for quick device status glances? Deferred to v1.0.

3. **Companion app** — Should there be a Wear OS companion for device notifications? Deferred to v1.0.

4. **Tablet layout** — Should tablets get a multi-pane layout (device list + detail side-by-side)? Leaning toward yes for v0.5+.

---

## Implementation Plan

| Phase | Milestone | Screens | Features |
|-------|-----------|---------|----------|
| P0 | Skeleton | Discovery | mDNS scan, device list, basic WebSocket connection |
| P1 | Pairing | Discovery + Pairing | PIN entry, session management, simple device info display |
| P2 | Dashboard | Dashboard | Device details, capability display, connection status |
| P3 | Terminal | Terminal | PTY output rendering, keyboard input, ANSI colors |
| P4 | Screen | Screen Viewer | WebRTC peer connection, H.264 rendering, touch input |
| P5 | Files | File Browser | Directory listing, file upload/download, progress |
| P6 | System | System Monitor | CPU/memory/storage gauges, process list |
| P7 | Polish | All | Settings, notifications, tablet layout, widgets |

---

## References

- RFC-0001: Connection Engine (transport layer)
- RFC-0002: Runtime Architecture (BPP methods client calls)
- RFC-0003: Pairing Protocol (pairing flow)
- RFC-0006: Cloud Relay (remote access)
- Experience Book: All 11 chapters (design language, UX patterns)
- Reference: rest-endpoints.md, cli-reference.md, websocket-events.md
