# Offline First

**Work without internet. Cloud is a fallback, not a requirement.**

## Problem

Many IoT and device management tools require a cloud connection for basic functionality. If the internet is down, the cloud is unreachable, or the service is discontinued, the devices become unmanageable. BuzzPi should never have this failure mode. The local network must work independently of the cloud.

## Solution

BuzzPi operates in two modes, with local-first preference:

### Local Mode (No Internet Required)

When the app and device are on the same local network:

| Feature | Works? | Mechanism |
|---------|--------|-----------|
| Device discovery | Yes | mDNS |
| Pairing | Yes | Local WebSocket |
| Terminal | Yes | WebRTC P2P |
| Screen | Yes | WebRTC P2P |
| File transfer | Yes | WebRTC P2P over data channel |
| All services | Yes | WebRTC P2P (capabilities don't require relay) |

In local mode, the Relay Server is not involved in any data path. It is only used for WebSocket signaling (SDP exchange), and even that can be done locally if needed.

### Cloud Mode (Internet Required)

When the app and device are on different networks:

| Feature | Works? | Mechanism |
|---------|--------|-----------|
| Device discovery | Yes | Cloud registration |
| Pairing | Yes | Relay relayed |
| Terminal | Yes | WebRTC via TURN |
| Screen | Yes | WebRTC via TURN |
| File transfer | Yes | WebRTC via TURN |
| All services | Yes | WebRTC via TURN |

### Graceful Degradation

When internet is lost but local network remains:
1. Cloud-registered devices on other networks show "offline"
2. Local devices continue to work perfectly
3. The app shows "Remote devices are unavailable" — not "connection lost" or an error
4. When internet returns, remote devices reappear without user action

When local network is lost (device disconnects from WiFi):
1. The device shows "offline" in the app
2. The app uses the grace period (2 minutes) before notifying
3. Cloud-registered remote devices remain accessible

### Data Storage

- Device state is cached locally in the app and on the device
- The app works from cache when offline (with appropriate indicators)
- Settings changes made offline sync when connectivity returns
- No critical data depends on cloud persistence

## User Experience

A user opens BuzzPi in their basement where cellular reception is poor but their home WiFi is strong. All their local devices appear, fully functional. A banner at the top says "Remote devices unavailable" in a muted color — informational, not alarming.

When they move upstairs and WiFi reconnects, remote devices silently reappear.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| Local-first means no cloud sync of some data | Device logs, historical stats, and event history are stored on-device. Cloud sync is a premium feature. |
| mDNS requires local network | This is a technical limitation of multicast DNS. Cloud registration fills the gap for remote access. |
| Dual-mode adds implementation complexity | The Runtime and app must handle both paths. The connection engine abstracts this (see Automatic Transport). |

## Examples

- Device list: shows local devices immediately, remote devices load after cloud registration check
- Workspace actions: all work locally, some are disabled with explanation when remote device is offline
- Pairing: works fully offline (local network only) and fully online (relayed)
- Settings: all settings local; cloud sync is explicit ("Sync to Cloud" button)

## Related Patterns

- [Automatic Transport](automatic-transport.md): Decides whether to use P2P or relay
- [Device Discovery](device-discovery.md): mDNS for local, cloud registration for remote
- [Best Effort Reconnect](best-effort-reconnect.md): Handles transitions between online and offline
