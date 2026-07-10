# BPP Packet Types

**Complete catalog of all BPP packet types with IDs, descriptions, and usage.** Packets are the atomic units of communication over WebSocket and WebRTC Data Channels. Each packet uses the standard BPP envelope.

---

## Packet ID Registry

Every BPP method has a numeric ID for efficient binary encoding over WebRTC Data Channels. The ID is sent in the envelope's `mid` field.

| ID | Method | Direction | Transport | Layer |
|----|--------|-----------|-----------|-------|
| 0x0001 | `relay.connect` | Both → Relay | WebSocket | Signaling |
| 0x0002 | `relay.disconnect` | Both → Relay | WebSocket | Signaling |
| 0x0003 | `relay.sdp_offer` | Both → Relay | WebSocket | Signaling |
| 0x0004 | `relay.sdp_answer` | Both → Relay | WebSocket | Signaling |
| 0x0005 | `relay.ice_candidate` | Both → Relay | WebSocket | Signaling |
| 0x0006 | `relay.heartbeat` | Both → Relay | WebSocket | Signaling |
| 0x0007 | `relay.device.list` | Client → Relay | WebSocket | Signaling |
| 0x0008 | `relay.error` | Relay → Both | WebSocket | Signaling |
| | | | |
| 0x0101 | `device.info` | Client ↔ Device | WebSocket | Device |
| 0x0102 | `device.stats` | Client → Device | WebSocket | Device |
| 0x0103 | `device.reboot` | Client → Device | WebSocket | Device |
| 0x0104 | `device.shutdown` | Client → Device | WebSocket | Device |
| 0x0105 | `device.event` | Device → Client | WebSocket | Device |
| 0x0106 | `device.log` | Device → Client | WebSocket | Device |
| | | | |
| 0x0201 | `terminal.open` | Client → Device | WebSocket | Terminal |
| 0x0202 | `terminal.close` | Client → Device | WebSocket | Terminal |
| 0x0203 | `terminal.resize` | Client → Device | WebSocket | Terminal |
| 0x0204 | `terminal.write` | Client → Device | Data Channel | Terminal |
| 0x0205 | `terminal.output` | Device → Client | Data Channel | Terminal |
| 0x0206 | `terminal.event` | Device → Client | WebSocket | Terminal |
| | | | |
| 0x0301 | `screen.start` | Client → Device | WebSocket | Screen |
| 0x0302 | `screen.stop` | Client → Device | WebSocket | Screen |
| 0x0303 | `screen.input` | Client → Device | Data Channel | Screen |
| 0x0304 | `screen.quality` | Client → Device | WebSocket | Screen |
| 0x0305 | `screen.event` | Device → Client | WebSocket | Screen |
| | | | |
| 0x0401 | `files.list` | Client → Device | WebSocket | Files |
| 0x0402 | `files.read` | Client → Device | WebSocket | Files |
| 0x0403 | `files.write` | Client → Device | WebSocket | Files |
| 0x0404 | `files.delete` | Client → Device | WebSocket | Files |
| 0x0405 | `files.mkdir` | Client → Device | WebSocket | Files |
| 0x0406 | `files.move` | Client → Device | WebSocket | Files |
| 0x0407 | `files.chunked` | Client ↔ Device | Data Channel | Files |
| 0x0408 | `files.event` | Device → Client | WebSocket | Files |
| | | | |
| 0x0501 | `docker.ps` | Client → Device | WebSocket | Docker |
| 0x0502 | `docker.inspect` | Client → Device | WebSocket | Docker |
| 0x0503 | `docker.logs` | Client → Device | WebSocket | Docker |
| 0x0504 | `docker.start` | Client → Device | WebSocket | Docker |
| 0x0505 | `docker.stop` | Client → Device | WebSocket | Docker |
| 0x0506 | `docker.restart` | Client → Device | WebSocket | Docker |
| 0x0507 | `docker.stats` | Client → Device | WebSocket | Docker |
| 0x0508 | `docker.images` | Client → Device | WebSocket | Docker |
| 0x0509 | `docker.event` | Device → Client | WebSocket | Docker |
| | | | |
| 0x0601 | `gpio.list` | Client → Device | WebSocket | GPIO |
| 0x0602 | `gpio.read` | Client → Device | WebSocket | GPIO |
| 0x0603 | `gpio.write` | Client → Device | WebSocket | GPIO |
| 0x0604 | `gpio.pwm` | Client → Device | WebSocket | GPIO |
| 0x0605 | `gpio.watch` | Client → Device | WebSocket | GPIO |
| 0x0606 | `gpio.unwatch` | Client → Device | WebSocket | GPIO |
| 0x0607 | `gpio.event` | Device → Client | Data Channel | GPIO |
| | | | |
| 0x0701 | `camera.list` | Client → Device | WebSocket | Camera |
| 0x0702 | `camera.start` | Client → Device | WebSocket | Camera |
| 0x0703 | `camera.stop` | Client → Device | WebSocket | Camera |
| 0x0704 | `camera.snapshot` | Client → Device | WebSocket | Camera |
| 0x0705 | `camera.event` | Device → Client | WebSocket | Camera |
| | | | |
| 0x0801 | `logs.list` | Client → Device | WebSocket | Logs |
| 0x0802 | `logs.read` | Client → Device | WebSocket | Logs |
| 0x0803 | `logs.follow` | Client → Device | WebSocket | Logs |
| 0x0804 | `logs.event` | Device → Client | Data Channel | Logs |
| | | | |
| 0x0901 | `systemd.list` | Client → Device | WebSocket | Systemd |
| 0x0902 | `systemd.status` | Client → Device | WebSocket | Systemd |
| 0x0903 | `systemd.start` | Client → Device | WebSocket | Systemd |
| 0x0904 | `systemd.stop` | Client → Device | WebSocket | Systemd |
| 0x0905 | `systemd.restart` | Client → Device | WebSocket | Systemd |
| 0x0906 | `systemd.logs` | Client → Device | WebSocket | Systemd |
| | | | |
| 0x0A01 | `capabilities.list` | Client → Device | WebSocket | Capability |
| 0x0A02 | `capabilities.subscribe` | Client → Device | WebSocket | Capability |
| 0x0A03 | `capabilities.event` | Device → Client | WebSocket | Capability |
| | | | |
| 0x0B01 | `extension.list` | Client → Device | WebSocket | Extension |
| 0x0B02 | `extension.install` | Client → Device | WebSocket | Extension |
| 0x0B03 | `extension.uninstall` | Client → Device | WebSocket | Extension |
| 0x0B04 | `extension.start` | Client → Device | WebSocket | Extension |
| 0x0B05 | `extension.stop` | Client → Device | WebSocket | Extension |
| 0x0B06 | `extension.permissions` | Client → Device | WebSocket | Extension |
| 0x0B07 | `extension.event` | Device → Client | WebSocket | Extension |
| | | | |
| 0x0C01 | `buzzai.ask` | Client → Device | WebSocket | BuzzAI |
| 0x0C02 | `buzzai.confirm` | Client → Device | WebSocket | BuzzAI |
| 0x0C03 | `buzzai.tools.list` | Client → Device | WebSocket | BuzzAI |
| 0x0C04 | `buzzai.event` | Device → Client | WebSocket | BuzzAI |

---

## Packet Class Hierarchy

```
BPP Packet
├── Request      (type: "request", includes rid)
│   ├── Signaling     (0x0001-0x00FF)
│   ├── Device        (0x0101-0x01FF)
│   ├── Terminal      (0x0201-0x02FF)
│   ├── Screen        (0x0301-0x03FF)
│   ├── Files         (0x0401-0x04FF)
│   ├── Docker        (0x0501-0x05FF)
│   ├── GPIO          (0x0601-0x06FF)
│   ├── Camera        (0x0701-0x07FF)
│   ├── Logs          (0x0801-0x08FF)
│   ├── Systemd       (0x0901-0x09FF)
│   ├── Capability    (0x0A01-0x0AFF)
│   ├── Extension     (0x0B01-0x0BFF)
│   └── BuzzAI        (0x0C01-0x0CFF)
│
├── Response     (type: "response", echoes rid)
│   └── (same hierarchy as Request)
│
├── Event        (type: "event", no rid)
│   └── (same hierarchy as Request)
│
└── Error        (type: "error", echoes rid of failed request)
    └── (standard error envelope, independent of hierarchy)
```

---

## Binary Encoding (WebRTC Data Channel)

For high-throughput packets over WebRTC Data Channels, a compact binary encoding is used:

```
┌─────────┬──────────┬──────────┬──────────────────────────────┐
│ MID (2) │ RID (8)  │ FLAGS (1)│ PAYLOAD (variable)           │
├─────────┼──────────┼──────────┼──────────────────────────────┤
│ 0x0001  │ 0x0000.. │ 0b0000000│ CBOR-encoded method params   │
│         │ 0xFFFFFFFF│ 1 = req  │                              │
│         │          │ 0 = resp │                              │
└─────────┴──────────┴──────────┴──────────────────────────────┘
```

- **MID (2 bytes)**: Method ID from the registry above
- **RID (8 bytes)**: Request ID (uint64, big-endian)
- **FLAGS (1 byte)**: Bit 0 = request/response, Bits 1-7 = reserved
- **PAYLOAD**: CBOR-encoded parameters (replaces JSON for efficiency)

Binary encoding reduces overhead from ~50 bytes per JSON envelope to ~11 bytes per packet.

---

## Transport Selection

| Packet IDs | Transport | Rationale |
|------------|-----------|-----------|
| 0x0001-0x00FF | WebSocket | Signaling must go through relay |
| 0x0101-0x01FF | WebSocket | Low frequency, needs reliability |
| 0x0204-0x0205 | WebRTC Data Channel | High frequency terminal output |
| 0x0303 | WebRTC Data Channel | Mouse input needs low latency |
| 0x0304-0x0305 | WebSocket | Control messages, reliable delivery |
| 0x0407 | WebRTC Data Channel | Large file chunks |
| 0x0501-0x0509 | WebSocket | Low frequency, CRUD operations |
| 0x0601-0x0606 | WebSocket | Control, low frequency |
| 0x0607 | WebRTC Data Channel | Real-time pin change events |
| 0x0701-0x0705 | WebSocket + Media Track | Camera stream on dedicated media track |
| 0x0801-0x0803 | WebSocket | Control |
| 0x0804 | WebRTC Data Channel | Real-time log streaming |
| 0x0901-0x0906 | WebSocket | Control |
| 0x0A01-0x0A03 | WebSocket | Low frequency, capability negotiation |
| 0x0B01-0x0B07 | WebSocket | Control |
| 0x0C01-0x0C04 | WebSocket | AI queries, low frequency |

---

## Reserved Ranges

| Range | Use | Status |
|-------|-----|--------|
| 0x0001-0x00FF | Signaling / Relay | Allocated |
| 0x0101-0x01FF | Device | Allocated |
| 0x0201-0x02FF | Terminal | Allocated |
| 0x0301-0x03FF | Screen / Desktop | Allocated |
| 0x0401-0x04FF | Files | Allocated |
| 0x0501-0x05FF | Docker | Allocated |
| 0x0601-0x06FF | GPIO | Allocated |
| 0x0701-0x07FF | Camera | Allocated |
| 0x0801-0x08FF | Logs | Allocated |
| 0x0901-0x09FF | Systemd | Allocated |
| 0x0A01-0x0AFF | Capability | Allocated |
| 0x0B01-0x0BFF | Extension / Plugin | Allocated |
| 0x0C01-0x0CFF | BuzzAI | Allocated |
| 0x0D01-0x0DFF | Bluetooth | Reserved for future |
| 0x0E01-0x0EFF | WiFi | Reserved for future |
| 0x0F01-0x0FFF | Audio | Reserved for future |
| 0x1001-0x10FF | Networking | Reserved for future |
| 0x1101-0xEFFF | Extension (dynamic) | Allocated at runtime by plugins |
| 0xF001-0xFFFF | Vendor-specific | Private use, not in registry |
