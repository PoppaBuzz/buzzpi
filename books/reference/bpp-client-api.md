# BPP Client API Reference

**Public API surface for implementing BPP clients.** This document defines the interface any BPP client must implement, regardless of platform (Android, Desktop, CLI). Language-specific implementations should follow these contracts.

---

## Core Interfaces

### BppClient

The primary interface for communicating with a BuzzPi Runtime.

```go
type BppClient interface {
    // Connect establishes a connection to a device.
    // Returns immediately; connection state is tracked via events.
    Connect(ctx context.Context, deviceID string) error

    // Disconnect terminates the connection gracefully.
    Disconnect() error

    // SendRequest sends a BPP request and returns the response.
    // Blocks until response received or timeout.
    SendRequest(ctx context.Context, method string, params interface{}) (*BppResponse, error)

    // SendAndForget sends a request without waiting for a response.
    SendAndForget(method string, params interface{}) error

    // Subscribe registers a handler for events matching the given method prefix.
    // Returns a subscription ID for unsubscription.
    Subscribe(methodPrefix string, handler EventHandler) (string, error)

    // Unsubscribe removes a previously registered event handler.
    Unsubscribe(subscriptionID string) error

    // State returns the current connection state as a readable stream.
    State() <-chan ConnectionState

    // OnStateChange registers a callback for connection state changes.
    OnStateChange(handler func(ConnectionState))

    // Capabilities returns the device's capability list.
    // Updated automatically when capabilities change.
    Capabilities() []Capability

    // PeerConnection returns the underlying WebRTC peer connection
    // for direct access to data channels (advanced use only).
    PeerConnection() *webrtc.PeerConnection

    // Close releases all resources held by the client.
    Close() error
}
```

### BppResponse

```go
type BppResponse struct {
    RequestID string
    Method    string
    Result    json.RawMessage
    Error     *BppError
}

type BppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Data    json.RawMessage `json:"data,omitempty"`
}
```

### EventHandler

```go
type EventHandler func(event BppEvent)

type BppEvent struct {
    ID        string          `json:"id"`
    Timestamp time.Time       `json:"ts"`
    Method    string          `json:"method"`
    Params    json.RawMessage `json:"params"`
}
```

---

## Connection Management

### ConnectionState

```go
type ConnectionState int

const (
    Disconnected  ConnectionState = iota
    Scanning                     // mDNS discovery active
    Discovering                  // Fetching device list from relay
    Connecting                   // WebSocket + WebRTC handshake
    Connected                    // Fully operational
    Degrading                    // Connection quality issues detected
    Reconnecting                 // Attempting to restore lost connection
    Failed                       // Connection unrecoverable
)
```

### Connect Flow

```
1. BppClient.Connect(deviceID)
2.   Resolve device IP (mDNS or relay)
3.   Open WebSocket to device or relay
4.   Authenticate (JWT or pairing code)
5.   Initiate WebRTC handshake
6.   Exchange SDP offers/answers
7.   Exchange ICE candidates
8.   Open data channels
9.   Fetch capabilities
10.  Set state to Connected
```

---

## Device Methods

```go
// DeviceService provides access to device information and control.
type DeviceService interface {
    // Info returns detailed device identity information.
    Info(ctx context.Context) (*DeviceInfo, error)

    // Stats returns current system statistics.
    Stats(ctx context.Context) (*SystemStats, error)

    // Reboot restarts the device.
    Reboot(ctx context.Context) error

    // Shutdown powers off the device.
    Shutdown(ctx context.Context) error

    // OnEvent registers a handler for device-pushed events.
    OnEvent(handler func(DeviceEvent))
}

type DeviceInfo struct {
    DeviceID        string   `json:"device_id"`
    FriendlyName    string   `json:"friendly_name"`
    Model           string   `json:"model"`
    RuntimeVersion  string   `json:"runtime_version"`
    UptimeSeconds   int64    `json:"uptime_seconds"`
    Capabilities    []string `json:"capabilities"`
    Platform        string   `json:"platform"`
}

type SystemStats struct {
    CPU         CPUStats    `json:"cpu"`
    Memory      MemoryStats `json:"memory"`
    Storage     []DiskStats `json:"storage"`
    Network     NetStats    `json:"network"`
    UptimeSeconds int64     `json:"uptime_seconds"`
}

type CPUStats struct {
    UsagePercent    float64 `json:"usage_percent"`
    TemperatureC    float64 `json:"temperature_celsius"`
    FrequencyMHz    int     `json:"frequency_mhz"`
}

type MemoryStats struct {
    TotalMB     int     `json:"total_mb"`
    UsedMB      int     `json:"used_mb"`
    AvailableMB int     `json:"available_mb"`
    Percent     float64 `json:"percent"`
}
```

---

## Terminal Service

```go
// TerminalService provides access to remote terminal sessions.
type TerminalService interface {
    // Open creates a new terminal session.
    Open(ctx context.Context, opts TerminalOptions) (*TerminalSession, error)

    // List returns all active terminal sessions.
    List() []*TerminalSession
}

type TerminalOptions struct {
    Rows          int    `json:"rows"`          // 40
    Cols          int    `json:"cols"`          // 80
    Shell         string `json:"shell"`         // "/bin/bash"
    Environment   map[string]string             // [optional]
}

type TerminalSession struct {
    SessionID string

    // Write sends input to the terminal.
    Write(data []byte) error

    // Resize changes the terminal dimensions.
    Resize(rows, cols int) error

    // Close terminates the session.
    Close() error

    // OnOutput registers a handler for terminal output.
    OnOutput(handler func(output []byte))

    // Output returns a channel of terminal output data.
    Output() <-chan []byte
}
```

---

## Screen Streaming Service

```go
// ScreenService provides remote desktop streaming.
type ScreenService interface {
    Start(ctx context.Context, opts ScreenOptions) (*ScreenSession, error)
    Stop(sessionID string) error
    List() []*ScreenSession
}

type ScreenOptions struct {
    Quality      ScreenQuality `json:"quality"`       // "high" | "medium" | "low" | "minimum"
    MaxFPS       int           `json:"max_fps"`        // 30
    MaxResolution string       `json:"max_resolution"` // "1920x1080"
    CursorVisible bool         `json:"cursor_visible"` // true
}

type ScreenQuality string

const (
    ScreenHigh    ScreenQuality = "high"
    ScreenMedium  ScreenQuality = "medium"
    ScreenLow     ScreenQuality = "low"
    ScreenMinimum ScreenQuality = "minimum"
)

type ScreenSession struct {
    SessionID       string
    ActualFPS       int
    ActualResolution string
    Codec           string

    // SendInput sends mouse/touch events.
    SendInput(event InputEvent) error

    // SetQuality changes streaming quality.
    SetQuality(quality ScreenQuality, opts ScreenOptions) error

    // Close stops the session.
    Close() error

    // OnFrame registers a handler for decoded video frames.
    OnFrame(handler func(frame VideoFrame))

    // OnQualityChange registers a handler for quality adaptation events.
    OnQualityChange(handler func(reason string))
}

type InputEvent struct {
    Type     InputType  `json:"type"`     // "mouse" | "touch" | "keyboard"
    Action   string     `json:"action"`   // "mousedown" | "mouseup" | "mousemove" | etc.
    X        int        `json:"x"`
    Y        int        `json:"y"`
    Button   string     `json:"button,omitempty"`   // "left" | "right" | "middle"
    Key      string     `json:"key,omitempty"`      // Keyboard key
    Modifiers []string  `json:"modifiers,omitempty"` // ["ctrl", "alt", "shift"]
}

type VideoFrame struct {
    Timestamp time.Time
    Data      []byte  // Decoded RGBA pixel data
    Width     int
    Height    int
    Keyframe  bool
}
```

---

## File Service

```go
// FileService provides remote file system access.
type FileService interface {
    List(ctx context.Context, path string) (*DirectoryListing, error)
    Read(ctx context.Context, path string) (*FileContent, error)
    Write(ctx context.Context, path string, content []byte, opts WriteOptions) error
    Delete(ctx context.Context, path string, recursive bool) error
    Mkdir(ctx context.Context, path string, mode string) error
    Move(ctx context.Context, from, to string) error
    Upload(ctx context.Context, localPath, remotePath string, progress ProgressHandler) (*FileTransfer, error)
    Download(ctx context.Context, remotePath, localPath string, progress ProgressHandler) (*FileTransfer, error)
}

type DirectoryListing struct {
    Path    string        `json:"path"`
    Entries []DirEntry    `json:"entries"`
}

type DirEntry struct {
    Name     string `json:"name"`
    Type     string `json:"type"`     // "file" | "directory"
    Size     int64  `json:"size"`
    Modified string `json:"modified"`
    Mode     string `json:"mode"`
}

type FileContent struct {
    Path     string `json:"path"`
    Content  []byte `json:"content"`
    Size     int64  `json:"size"`
    Encoding string `json:"encoding"` // "text" | "base64"
}

type WriteOptions struct {
    Append bool `json:"append"` // default: false
    Mode   string `json:"mode,omitempty"` // "0644"
}

type FileTransfer struct {
    TransferID   string
    Direction    string  // "upload" | "download"
    RemotePath   string
    LocalPath    string
    TotalBytes   int64
    Transferred  int64
    Progress     float64 // 0.0 - 1.0
    Status       string  // "pending" | "transferring" | "completed" | "failed"
}

type ProgressHandler func(progress *FileTransfer)
```

---

## Docker Service

```go
type DockerService interface {
    PS(ctx context.Context, all bool) ([]Container, error)
    Inspect(ctx context.Context, containerID string) (*ContainerDetail, error)
    Logs(ctx context.Context, containerID string, opts LogOptions) (string, error)
    Start(ctx context.Context, containerID string) error
    Stop(ctx context.Context, containerID string) error
    Restart(ctx context.Context, containerID string) error
    Stats(ctx context.Context, containerID string) (*ContainerStats, error)
    Images(ctx context.Context) ([]Image, error)
}

type Container struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Image   string `json:"image"`
    Status  string `json:"status"`
    Created string `json:"created"`
    Ports   []string `json:"ports"`
}

type ContainerStats struct {
    CPUPercent        float64 `json:"cpu_percent"`
    MemoryMB          int     `json:"memory_mb"`
    MemoryPercent     float64 `json:"memory_percent"`
    NetworkRxBytes    int64   `json:"network_rx_bytes"`
    NetworkTxBytes    int64   `json:"network_tx_bytes"`
}
```

---

## GPIO Service

```go
type GpioService interface {
    List(ctx context.Context) (*GpioList, error)
    Read(ctx context.Context, pin int) (*GpioPin, error)
    Write(ctx context.Context, pin int, value int) error
    PWM(ctx context.Context, pin int, dutyCycle int, frequency int) error
    Watch(ctx context.Context, pin int, edge string) (string, error)
    Unwatch(ctx context.Context, watchID string) error
    OnEvent(handler func(event GpioEvent))
}

type GpioList struct {
    Pins []GpioPin `json:"pins"`
}

type GpioPin struct {
    Pin      int    `json:"pin"`
    Name     string `json:"name"`
    Mode     string `json:"mode"`    // "input" | "output" | "pwm" | "alt"
    Value    int    `json:"value"`   // 0 or 1 (PWM: 0-255)
    Function string `json:"function"`
}

type GpioEvent struct {
    Pin   int    `json:"pin"`
    Value int    `json:"value"`
    Edge  string `json:"edge"`  // "rising" | "falling"
}
```

---

## Capability Service

```go
type CapabilityService interface {
    // List returns all capabilities from the connected device.
    List(ctx context.Context) ([]Capability, error)

    // Subscribe receives capability change events.
    Subscribe(events []string, handler CapabilityHandler) (string, error)

    // Unsubscribe stops receiving capability events.
    Unsubscribe(subscriptionID string) error

    // Available checks if a specific capability is currently available.
    Available(capabilityID string) bool

    // Param gets a specific capability parameter.
    Param(capabilityID, paramName string) interface{}
}

type Capability struct {
    ID        string                 `json:"id"`
    Version   string                 `json:"version"`
    Available bool                   `json:"available"`
    Params    map[string]interface{} `json:"params,omitempty"`
}

type CapabilityHandler func(event CapabilityEvent)

type CapabilityEvent struct {
    Type       string     `json:"type"`       // "added" | "removed" | "updated"
    Capability Capability `json:"capability"`
}
```

---

## Extension Service

```go
type ExtensionService interface {
    List(ctx context.Context) ([]Plugin, error)
    Install(ctx context.Context, id string, source string, version string) (*Plugin, error)
    Uninstall(ctx context.Context, id string) error
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string) error
    Permissions(ctx context.Context, id string, action string, perms []string) error
    OnEvent(handler func(event ExtensionEvent))
}

type Plugin struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Version      string   `json:"version"`
    State        string   `json:"state"`
    Capabilities []string `json:"capabilities"`
}

type ExtensionEvent struct {
    PluginID  string      `json:"extension_id"`
    EventType string      `json:"event_type"`
    Data      interface{} `json:"data"`
}
```

---

## Connection Events

```go
type ConnectionEvents interface {
    // OnEstablished fires when connection is established.
    OnEstablished(handler func(transport string, rttMs int))

    // OnLost fires when connection is lost.
    OnLost(handler func(reason string))

    // OnReconnecting fires when reconnection begins.
    OnReconnecting(handler func(attempt, maxAttempts int))

    // OnTransportSwitched fires when transport type changes.
    OnTransportSwitched(handler func(from, to, reason string))

    // OnQualityChange fires when connection quality metrics update.
    OnQualityChange(handler func(stats ConnectionQuality))
}

type ConnectionQuality struct {
    RTTMs               int     `json:"rtt_ms"`
    PacketLossPercent   float64 `json:"packet_loss_percent"`
    JitterMs            int     `json:"jitter_ms"`
    AvailableBandwidth  int64   `json:"available_bandwidth"`
}
```

---

## Error Handling

```go
// All API methods can return BppError or one of these typed errors:
var (
    ErrConnectionTimeout    = errors.New("connection timed out")
    ErrDeviceNotFound       = errors.New("device not found on network")
    ErrPairingRejected      = errors.New("device rejected pairing")
    ErrNotConnected         = errors.New("not connected to any device")
    ErrCapabilityUnavailable = errors.New("capability not available on this device")
    ErrPermissionDenied     = errors.New("permission denied")
    ErrInvalidRequest       = errors.New("invalid request parameters")
    ErrTransportUnavailable = errors.New("no transport available (P2P or relay)")
)
```

---

## Usage Examples

### Go Client

```go
client := bpp.NewClient(bpp.ClientConfig{
    RelayURL: "wss://relay.buzzpi.dev/v1",
})

// Connect to device
if err := client.Connect(ctx, "dev_abc12345"); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Open terminal
term, err := client.Terminal().Open(ctx, bpp.TerminalOptions{
    Rows: 40,
    Cols: 80,
})
if err != nil {
    log.Fatal(err)
}

// Handle output
term.OnOutput(func(data []byte) {
    fmt.Print(string(data))
})

// Send command
term.Write([]byte("ls -la\n"))
```

### Kotlin Client (Android)

```kotlin
val client = BppClient(
    relayUrl = "wss://relay.buzzpi.dev/v1",
    context = applicationContext
)

lifecycleScope.launch {
    client.connect("dev_abc12345")

    val stats = client.device().stats()
    Log.d("BuzzPi", "CPU: ${stats.cpu.usagePercent}%")

    client.terminal().open(TerminalOptions(rows = 40, cols = 80)).also { session ->
        session.onOutput { data ->
            Log.d("BuzzPi", "Terminal: ${data.decodeToString()}")
        }
        session.write("htop\n".encodeToByteArray())
    }
}
```

---

## Implementation Requirements

### Thread Safety

All BPP client implementations must be thread-safe. Multiple goroutines/coroutines may call methods concurrently.

### Connection Reuse

The BppClient maintains a single WebSocket + WebRTC connection. All service methods (Terminal, Screen, Files, etc.) share this connection.

### Timeouts

Default timeouts:

| Operation | Default Timeout |
|-----------|----------------|
| Connect | 15 seconds |
| Request/Response | 30 seconds |
| File Transfer | 5 minutes (per file) |
| Screen Start | 10 seconds |

### Cleanup

Calling `Close()` on the BppClient must:
1. Close all active service sessions (terminal, screen, etc.)
2. Close WebRTC peer connection
3. Close WebSocket connection
4. Release all resources
5. Set state to Disconnected
