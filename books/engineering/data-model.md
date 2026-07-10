# Data Model

Core data entities in the BuzzPi ecosystem. These are logical models — specific implementations (SQL tables, protobuf messages, JSON schemas) are defined in separate specification documents.

---

## Device

```yaml
Device:
  description: A physical Linux device running BuzzPi Runtime
  identity:
    device_id: UUID v7
    hardware_id: string          # Serial number or MAC-based identifier
    friendly_name: string        # User-assigned name (e.g., "Kitchen Pi")
    model: string                # e.g., "Raspberry Pi 5"
    
  status:
    state: DeviceState           # NEW | PAIRING | ONLINE | OFFLINE | UNPAIRED
    last_seen: timestamp         # Last heartbeat received
    ip_address: string           # Last known IP
    connection_type: enum        # direct | relay | unknown
    
  hardware:
    cpu: string                  # Model + architecture
    memory_mb: integer
    storage_total_mb: integer
    storage_used_mb: integer
    temperature_celsius: float
    
  runtime:
    version: string              # Installed Runtime version
    uptime_seconds: integer
    extensions: Extension[]      # Installed extensions
    
  pairing:
    paired_at: timestamp
    paired_by: user_id
    identity_key: public_key     # Device's identity key
```

---

## User

```yaml
User:
  description: A human using the BuzzPi app or CLI
  identity:
    user_id: UUID v7
    email: string
    display_name: string
    
  account:
    created_at: timestamp
    plan: PlanType               # free | premium | classroom
    device_limit: integer        # Max devices per plan
    
  preferences:
    theme: string                # light | dark | system
    notifications: NotificationSettings
    terminal_font_size: integer
    privacy: PrivacySettings     # Analytics opt-in, etc.
    
  relationships:
    devices: Device[]            # Paired devices
    groups: Group[]              # Created groups
```

---

## Group

```yaml
Group:
  description: A collection of devices managed as a unit
  identity:
    group_id: UUID v7
    name: string                 # e.g., "CS101 — Fall"
    created_by: user_id
    
  membership:
    devices: Device[]
    classroom_mode: boolean      # Restrict student actions
    
  actions:
    - restart_all
    - update_all
    - shutdown_all
    - reset_all (classroom)
    - broadcast_message
    
  permissions:
    # Per-device permissions for classroom mode
    allow_terminal: boolean
    allow_screen: boolean
    allow_settings: boolean
    allow_files: boolean
```

---

## Extension

```yaml
Extension:
  description: A capability add-on installed on a device
  identity:
    extension_id: UUID v7
    name: string                 # e.g., "Docker Manager"
    version: string
    publisher: string            # Publisher's user_id or "official"
    
  metadata:
    description: string
    icon_url: string
    permissions: string[]        # e.g., ["docker:manage", "files:read"]
    homepage_url: string
    source_url: string           # Open source repository
    
  state:
    installed_at: timestamp
    enabled: boolean
    configuration: map           # Extension-specific config
    services: Service[]          # Managed services
```

---

## Session

```yaml
Session:
  description: An active connection between the app (or CLI) and a device's Runtime
  identity:
    session_id: UUID v7
    device_id: UUID v7
    user_id: UUID v7
    
  connection:
    type: enum                   # direct | relay
    started_at: timestamp
    last_activity: timestamp
    bytes_sent: integer
    bytes_received: integer
    
  capabilities:
    terminal: boolean
    screen: boolean
    files: boolean
    extensions: string[]         # Available extensions during this session
```

---

## Notification

```yaml
Notification:
  description: An event notification delivered to the user
  identity:
    notification_id: UUID v7
    user_id: UUID v7
    device_id: UUID v7 (nullable)
    
  content:
    category: NotificationCategory  # device_event | action_result | system
    priority: NotificationPriority   # critical | warning | info | success
    title: string
    body: string
    action_label: string (nullable)  # e.g., "View Device"
    action_deeplink: string (nullable)
    
  state:
    created_at: timestamp
    read_at: timestamp (nullable)
    dismissed_at: timestamp (nullable)
```

---

## Action

```yaml
Action:
  description: A user-initiated operation on a device
  identity:
    action_id: UUID v7
    user_id: UUID v7
    device_id: UUID v7
    type: ActionType              # restart | update | shutdown | custom
    
  execution:
    status: ActionStatus          # pending | in_progress | completed | failed
    started_at: timestamp
    completed_at: timestamp
    duration_ms: integer
    result: string (nullable)     # Success/failure details
    error: string (nullable)      # Error message (if failed)
```

---

## Event Log

```yaml
EventLog:
  description: An immutable record of device events for audit and diagnostics
  identity:
    event_id: UUID v7
    device_id: UUID v7
    timestamp: timestamp
    
  event:
    type: EventType               # state_change | action | system | network
    source: string                # Component that generated the event
    severity: EventSeverity       # info | warning | error
    
  data:
    payload: map                  # Event-specific data
    previous_state: string        # Previous state (for state transitions)
    new_state: string             # New state
```

---

## Storage Strategy

| Data | Location | Retention |
|------|----------|-----------|
| Device metadata | Cloud DB | Lifetime of account |
| Session history | Cloud DB | 90 days |
| Event log (user-facing) | Cloud DB | 30 days |
| Event log (debug) | Device only | 1MB, rotated |
| Notification history | Cloud DB | 30 days |
| Extension registry | Cloud DB | Indefinite |
| User preferences | Cloud DB + local | Indefinite |
| Session tokens | Local (app) | Until logout |
| Device identity keys | Local (device) | Until unpaired |
| Pairing state | Volatile | Until paired/cancelled |
