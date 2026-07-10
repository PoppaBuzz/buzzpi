# Device Discovery

**Never ask for an IP address. Use mDNS, Bluetooth, QR, or cloud registration automatically.**

## Problem

The fundamental pain point BuzzPi solves: users should not need to know their device's IP address, remember a hostname, configure port forwarding, or edit network settings. Existing tools (SSH, VNC, web dashboards) require the user to find or remember the device's network address before they can do anything. This is the single largest barrier to managing multiple Raspberry Pis.

## Solution

BuzzPi offers four discovery mechanisms, ordered by preference:

### 1. mDNS (Local Network)

When a device with BuzzPi Runtime boots, it announces itself via mDNS (Multicast DNS, RFC 6762) as `_buzzpi._tcp.local`. The BuzzPi app on the same network discovers the device without any configuration.

```
Device boots → Runtime starts → Publishes mDNS service:
  Instance: buzzpi-5c2a
  Service: _buzzpi._tcp
  Domain: local
  Port: <relay port>
  TXT: device_id, version, pairing_status

App scans → Receives mDNS response → Displays device in list
```

The device appears in the app's device list within 5 seconds of booting, with no user action required.

### 2. Cloud Registration

If the device is paired (previously registered to the user's account), the Relay Server tells the app about the device regardless of network:

```
App connects to Relay → Server sends device list:
  - Kitchen Pi (online via relay)
  - Workshop Pi (offline, last seen 2h ago)
```

The device appears in the app even if it's on a different network, behind CGNAT, or currently offline.

### 3. QR Code (Secondary)

For headless devices without display, the Runtime can print a QR code to serial console or generate one from config:

```
BuzzPi Runtime generates QR:
  buzzpi://pair?device_id=<id>&code=<pairing_code>

App scans QR → Extracts device ID + pairing code → Pairs directly
```

### 4. Bluetooth (Future)

For Zero-config scenarios, the Runtime can broadcast a Bluetooth LE service. The app discovers the device via BLE and initiates pairing without any network dependency.

## User Experience

The user never performs a discovery action. They open the app. If a device is on the same network, it appears. If the device is paired and on a different network, it appears. The only empty state is when the user has no devices and none are currently on the network (see Information Architecture — empty states).

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| mDNS requires device to be on same LAN segment | mDNS does not cross subnets. Devices on different VLANs require cloud registration to appear. Acceptable because: (a) most home users have flat networks, (b) cloud registration covers the remote case. |
| Cloud registration requires BuzzPi Cloud | If the Relay Server is down, remote devices show as offline but are not discoverable. Acceptable because: (a) the Relay Server is designed for high availability, (b) local mDNS continues to work independently. |
| QR code requires a display or serial console | The primary path (mDNS + cloud) does not require a display. QR is a fallback for edge cases. |

## Examples

- Device List screen: shows all devices discovered via mDNS or cloud
- Pairing flow: triggered by mDNS discovery or QR scan
- Offline device list: shows cloud-registered devices even when offline
- First-run experience: empty state with "No devices found. Install BuzzPi Runtime on your Pi."

## Related Patterns

- [Automatic Transport](automatic-transport.md): Once discovered, the best connection path is chosen automatically
- [Explain, Don't Expose](explain-dont-expose.md): Discovery failures are presented in plain language
- [Progressive Disclosure](progressive-disclosure.md): Advanced discovery options (static IP, manual entry) are hidden under "Advanced" for experts
