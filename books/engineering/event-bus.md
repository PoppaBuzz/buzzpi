# Event Bus

**The internal pub/sub system that connects all Runtime components.** The Event Bus decouples subsystems within the BuzzPi Runtime, enabling loose coupling, extensibility, and clean separation of concerns. Core services, plugins, and the connection engine all communicate through the bus.

---

## Principles

1. **In-process only** — The Event Bus is internal to a single Runtime process. Cross-process events (to plugins) use the IPC protocol defined in Plugin System.
2. **Fire-and-forget by default** — Publishers do not wait for subscribers. Events are asynchronous and non-blocking.
3. **Ordered delivery per topic** — Events on the same topic are delivered to a subscriber in the order they were published.
4. **No persistence** — Events are not persisted. If no subscriber is listening, the event is lost. For durable event streams, use the session log or database.
5. **Subscriber isolation** — A slow or panicking subscriber never blocks other subscribers or the publisher.
6. **Throttle-safe** — Subscribers receive events via buffered channels. Backpressure is handled by dropping oldest events (configurable per subscriber).

---

## Architecture

```
                    ┌─────────────────────────────────────┐
                    │           Event Bus                  │
                    │                                      │
                    │  ┌──────────────────────────────┐    │
                    │  │        Topic Registry         │    │
                    │  │                              │    │
                    │  │  device.state     → [S1, S2] │    │
                    │  │  connection.*     → [S3]     │    │
                    │  │  service.*        → [S1, S4] │    │
                    │  │  stats.update     → [S2, S4] │    │
                    │  │  extension.*      → [S5]     │    │
                    │  └──────────────────────────────┘    │
                    │                                      │
                    │  ┌────────────┐ ┌────────────────┐   │
                    │  │ Publisher  │ │ Subscriber     │   │
                    │  │ API        │ │ API            │   │
                    │  │ Publish()  │ │ Subscribe()    │   │
                    │  │            │ │ Unsubscribe()  │   │
                    │  └────────────┘ └────────────────┘   │
                    └──────────────────────────────────────┘
                               │           ▲
                               │           │
                    ┌──────────▼───────────┴──────────┐
                    │         Core Runtime             │
                    │  ┌──────┐ ┌──────┐ ┌────────┐   │
                    │  │State │ │Conn  │ │Cap     │   │
                    │  │Mgr   │ │Engine│ │Service │   │
                    │  └──────┘ └──────┘ └────────┘   │
                    │  ┌──────┐ ┌──────┐ ┌────────┐   │
                    │  │Stats │ │Screen│ │Plugin  │   │
                    │  │      │ │Cap   │ │Manager │   │
                    │  └──────┘ └──────┘ └────────┘   │
                    └──────────────────────────────────┘
```

---

## Event Types

Events are organized in a hierarchical namespace using dot notation. Wildcard subscriptions match subtrees.

### Device Lifecycle

```go
// device.state.changed — Device state transitions
Event{Type: "device.state.changed", Data: DeviceStateEvent{
    From: "new",
    To:   "pairing",
}}

// device.registered — Device registered with relay
Event{Type: "device.registered", Data: DeviceRegisteredEvent{
    DeviceID: "dev_abc123",
}}

// device.unpaired — Device pairing removed
Event{Type: "device.unpaired", Data: DeviceUnpairedEvent{
    DeviceID:   "dev_abc123",
    Reason:     "user_initiated",
}}
```

### Connection Events

```go
// connection.established — WebRTC peer connection established
Event{Type: "connection.established", Data: ConnectionEvent{
    DeviceID:    "dev_abc123",
    Transport:   "direct",          // "direct" or "relay"
    RTTMs:       12,
}}

// connection.lost — Connection dropped
Event{Type: "connection.lost", Data: ConnectionLostEvent{
    DeviceID: "dev_abc123",
    Reason:   "timeout",
}}

// connection.quality — Quality metrics update
Event{Type: "connection.quality", Data: QualityEvent{
    DeviceID:          "dev_abc123",
    PacketLossPercent: 2.5,
    RTTMs:             45,
    AvailableBandwidth: 5000000,
}}

// connection.reconnecting — Attempting reconnection
Event{Type: "connection.reconnecting", Data: ReconnectingEvent{
    DeviceID:  "dev_abc123",
    Attempt:   3,
    MaxAttempts: 10,
}}

// connection.transport.switched — Transport type changed
Event{Type: "connection.transport.switched", Data: TransportSwitchEvent{
    DeviceID:   "dev_abc123",
    From:       "direct",
    To:         "relay",
    Reason:     "nat_traversal_failed",
}}
```

### Service Events

```go
// service.started — Runtime service started
Event{Type: "service.started", Data: ServiceEvent{
    Service: "terminal",
}}

// service.stopped — Runtime service stopped
Event{Type: "service.stopped", Data: ServiceEvent{
    Service: "screen",
    Reason:  "encoder_unavailable",
}}

// service.error — Service encountered a recoverable error
Event{Type: "service.error", Data: ServiceErrorEvent{
    Service: "screen",
    Error:   "drm_capture_failed",
    Fatal:   false,
}}
```

### Capability Events

```go
// capability.added — New capability became available
Event{Type: "capability.added", Data: CapabilityEvent{
    Capability: "hardware.camera",
    Params:     map[string]any{"max_resolution": "3280x2464"},
}}

// capability.removed — Capability no longer available
Event{Type: "capability.removed", Data: CapabilityEvent{
    Capability: "hardware.camera",
}}

// capability.updated — Capability parameters changed
Event{Type: "capability.updated", Data: CapabilityEvent{
    Capability: "service.screen",
    Params:     map[string]any{"max_fps": 15},
}}
```

### System Events

```go
// stats.update — Periodic system statistics
Event{Type: "stats.update", Data: StatsEvent{
    CPUPercent:    23.5,
    MemoryMB:      512,
    MemoryTotalMB: 4096,
    TemperatureC:  65.2,
    UptimeSeconds: 86400,
}}

// system.low_disk — Low disk space warning
Event{Type: "system.low_disk", Data: LowDiskEvent{
    AvailableMB: 512,
    ThresholdMB: 1024,
}}

// system.overheating — CPU temperature exceeded threshold
Event{Type: "system.overheating", Data: OverheatingEvent{
    TemperatureC: 85.0,
    ThresholdC:   80.0,
}}

// system.update_available — New Runtime version available
Event{Type: "system.update_available", Data: UpdateEvent{
    CurrentVersion: "0.1.0",
    NewVersion:     "0.1.1",
    Urgency:        "recommended",
}}
```

### Plugin Events

```go
// plugin.installed — Plugin was installed
Event{Type: "plugin.installed", Data: PluginEvent{
    PluginID: "com.example.weather",
    Version:  "1.0.0",
}}

// plugin.failed — Plugin entered failed state
Event{Type: "plugin.failed", Data: PluginFailedEvent{
    PluginID: "com.example.weather",
    Reason:   "out_of_memory",
    RestartCount: 3,
}}

// plugin.event — Plugin-pushed custom event (forwarded to clients)
Event{Type: "plugin.event", Data: PluginCustomEvent{
    PluginID:  "com.example.weather",
    EventType: "temperature_alert",
    Data:      map[string]any{"temp": 35.0, "threshold": 30.0},
}}
```

---

## Implementation

### Core Event Bus

```go
// EventBus is the central pub/sub hub within the Runtime.
type EventBus struct {
    mu          sync.RWMutex
    subscribers map[string][]*Subscriber // topic → subscribers
    wg          sync.WaitGroup
}

// Subscriber wraps a channel with configuration.
type Subscriber struct {
    ID        string
    Topic     string   // Exact or wildcard: "device.*"
    Chan      chan Event
    Buffer    int      // Channel buffer size
    DropPolicy DropPolicy // OldestWhenFull or Block
}

type DropPolicy int
const (
    DropOldest  DropPolicy = iota // Drop oldest event when buffer full
    Block                         // Block publisher until space available
)

// Event is the universal event envelope.
type Event struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"`
    Source    string      `json:"source"`     // Component that published
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

// Publish sends an event to all matching subscribers.
func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    var wg sync.WaitGroup

    for topic, subs := range eb.subscribers {
        if !topicMatches(topic, event.Type) {
            continue
        }

        for _, sub := range subs {
            wg.Add(1)
            go func(s *Subscriber) {
                defer wg.Done()
                s.deliver(event)
            }(sub)
        }
    }

    wg.Wait() // Wait for all deliveries (buffered, non-blocking per sub)
}

// Subscribe registers a subscriber for the given topic pattern.
func (eb *EventBus) Subscribe(topic string, buffer int, dropPolicy DropPolicy) *Subscriber {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    sub := &Subscriber{
        ID:         generateID(),
        Topic:      topic,
        Chan:       make(chan Event, buffer),
        Buffer:     buffer,
        DropPolicy: dropPolicy,
    }

    eb.subscribers[topic] = append(eb.subscribers[topic], sub)
    return sub
}

// Unsubscribe removes a subscriber.
func (eb *EventBus) Unsubscribe(sub *Subscriber) {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    for topic, subs := range eb.subscribers {
        for i, s := range subs {
            if s.ID == sub.ID {
                eb.subscribers[topic] = append(subs[:i], subs[i+1:]...)
                close(s.Chan)
                return
            }
        }
    }
}

func (s *Subscriber) deliver(event Event) {
    select {
    case s.Chan <- event:
        // Delivered
    default:
        if s.DropPolicy == DropOldest {
            // Drop oldest event from buffer
            select {
            case <-s.Chan: // Remove oldest
            default:
            }
            select {
            case s.Chan <- event: // Retry
            default:
            }
        }
        // Block policy: the event is dropped silently instead of blocking
        // (Block policy is not recommended for production use)
    }
}
```

### Topic Matching

```go
// topicMatches checks if a topic pattern matches an event type.
// Supports wildcard: "device.*" matches "device.state.changed"
// Supports multi-level: "**" matches everything
func topicMatches(pattern, eventType string) bool {
    if pattern == "**" {
        return true
    }

    patternParts := strings.Split(pattern, ".")
    eventParts := strings.Split(eventType, ".")

    for i, pp := range patternParts {
        if pp == "*" {
            return true // Wildcard matches rest
        }
        if i >= len(eventParts) {
            return false
        }
        if pp != eventParts[i] {
            return false
        }
    }

    return len(patternParts) == len(eventParts)
}
```

---

## Usage Examples

### Subscribing to Connection Events

```go
// NotificationService sends push notifications for relevant events
type NotificationService struct {
    bus     *EventBus
    pusher  *PushNotificationService
}

func (ns *NotificationService) Start(ctx context.Context) {
    // Subscribe to connection events
    sub := ns.bus.Subscribe("connection.*", 32, DropOldest)

    go func() {
        for {
            select {
            case event := <-sub.Chan:
                ns.handleEvent(event)
            case <-ctx.Done():
                ns.bus.Unsubscribe(sub)
                return
            }
        }
    }()
}

func (ns *NotificationService) handleEvent(event Event) {
    switch event.Type {
    case "connection.lost":
        data := event.Data.(ConnectionLostEvent)
        ns.pusher.Send(PushNotification{
            Title: "Device Disconnected",
            Body:  fmt.Sprintf("Your device has disconnected: %s", data.Reason),
            DeviceID: data.DeviceID,
        })

    case "connection.reconnecting":
        data := event.Data.(ReconnectingEvent)
        ns.pusher.Send(PushNotification{
            Title: "Reconnecting...",
            Body:  fmt.Sprintf("Attempt %d of %d", data.Attempt, data.MaxAttempts),
        })

    case "connection.established":
        ns.pusher.ClearNotifications()
    }
}
```

### State Management via Events

```go
// DeviceStateManager maintains device state and publishes transitions
type DeviceStateManager struct {
    bus         *EventBus
    deviceID    string
    currentState DeviceState
}

func (dsm *DeviceStateManager) TransitionTo(newState DeviceState) {
    oldState := dsm.currentState
    dsm.currentState = newState

    dsm.bus.Publish(Event{
        ID:        generateID(),
        Type:      "device.state.changed",
        Source:    "device_manager",
        Timestamp: time.Now(),
        Data: DeviceStateEvent{
            From: oldState,
            To:   newState,
        },
    })
}
```

### Stats Aggregation

```go
// StatsCollector gathers system stats and publishes periodically.
type StatsCollector struct {
    bus     *EventBus
    ticker  *time.Ticker
}

func (sc *StatsCollector) Start(ctx context.Context) {
    sc.ticker = time.NewTicker(30 * time.Second)

    go func() {
        for {
            select {
            case <-sc.ticker.C:
                stats := sc.collect()
                sc.bus.Publish(Event{
                    ID:        generateID(),
                    Type:      "stats.update",
                    Source:    "stats_collector",
                    Timestamp: time.Now(),
                    Data:      stats,
                })

                // Check thresholds and publish warnings
                if stats.TemperatureC > 80 {
                    sc.bus.Publish(Event{
                        Type: "system.overheating",
                        Data: OverheatingEvent{
                            TemperatureC: stats.TemperatureC,
                            ThresholdC:   80,
                        },
                    })
                }

            case <-ctx.Done():
                sc.ticker.Stop()
                return
            }
        }
    }()
}
```

### Plugin Event Forwarding

```go
// PluginEventBridge forwards plugin events to the Event Bus
// and publishes them to connected clients via BPP.
type PluginEventBridge struct {
    bus         *EventBus
    pluginMgr   *PluginManager
    bppConn     *BppConnection
}

func (b *PluginEventBridge) HandlePluginEvent(pluginID string, event PluginCustomEvent) {
    busEvent := Event{
        ID:        generateID(),
        Type:      "plugin.event",
        Source:    fmt.Sprintf("plugin:%s", pluginID),
        Timestamp: time.Now(),
        Data: PluginCustomEvent{
            PluginID:  pluginID,
            EventType: event.EventType,
            Data:      event.Data,
        },
    }

    b.bus.Publish(busEvent)

    // Forward to connected client as BPP event
    b.bppConn.Send(BppMessage{
        Type:   "event",
        Method: "plugin.event",
        Params: map[string]interface{}{
            "plugin_id":  pluginID,
            "event_type": event.EventType,
            "data":       event.Data,
        },
    })
}
```

---

## Event Bus Configuration

```yaml
# Runtime event bus configuration
event_bus:
  # Default subscriber buffer sizes
  default_buffer: 16
  stats_buffer: 64        # Stats are high-frequency
  connection_buffer: 32

  # Drop policy: "oldest" or "block"
  default_drop_policy: "oldest"

  # Whether to log all published events (debug only)
  audit_log: false
```

---

## Testing the Event Bus

| Test | Scenario | Expectation |
|------|----------|-------------|
| Publish/subscribe | Publish event on topic | Subscriber receives event |
| Wildcard match | Subscribe to `device.*` | Matches `device.state.changed` |
| Multi-level wildcard | Subscribe to `**` | Matches everything |
| No subscriber | Publish to empty topic | No-op, no error |
| Buffer overflow (oldest) | Publish 20 events to buffer of 10 | Only latest 10 retained |
| Slow subscriber | Subscriber takes 100ms per event | Publisher not blocked |
| Unsubscribe | Remove subscriber | No more events delivered |
| Concurrent publishers | 10 goroutines publishing simultaneously | All events delivered to matching subscribers |
| Topic isolation | Publish to `stats.update` | No delivery to `connection.*` subscribers |
