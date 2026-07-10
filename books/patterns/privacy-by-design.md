# Privacy by Design

**Data stays on-device. Cloud is opt-in, not default.**

## Problem

Many IoT and device management platforms are built around a cloud-first model: all data flows through the cloud, the cloud stores device state, and local operation is an afterthought (or impossible). This creates privacy risks, dependency on a service, and a single point of failure. BuzzPi must be designed from the ground up to respect user privacy, with cloud features as optional enhancements.

## Solution

### Data Minimization

BuzzPi collects the minimum data necessary to function:

| Data | Stored Where | Purpose | Transmitted? |
|------|-------------|---------|--------------|
| Device identity key | Device only | Authentication | Never leaves device |
| Client identity key | Client only | Authentication | Never leaves client |
| Session tokens | Client + Relay | Session management | Over TLS |
| Device telemetry (temp, storage) | Device + Client cache | Display in UI | Optional (anonymized aggregate) |
| Audit events | Device + Cloud (temporary) | User-facing history | 90-day retention |
| Usage analytics | Client only (opt-in) | Product improvement | Only if user opts in |
| Email address | Cloud | Account management | Required for account |

### Data Flow Rules

1. **No data is transmitted to the cloud unless required for the user's current action.**
   - Checking device status: no cloud data needed (local cache)
   - Connecting to a remote device: device ID and session token transmitted
   - Installing an extension: extension name transmitted (no device data)

2. **Data is deleted when no longer needed.**
   - Session tokens: deleted on logout
   - Device audit logs: deleted on unpair
   - Temporary data (connection state): deleted on disconnect

3. **Data is encrypted in transit and at rest.**
   - In transit: TLS (WebSocket), DTLS-SRTP (WebRTC)
   - At rest: platform key storage (Android Keystore), file permissions (0600)

### Privacy Settings

The app provides transparent privacy controls:

| Setting | Default | What It Changes |
|---------|---------|-----------------|
| Usage analytics | Off | Whether app usage patterns are sent anonymously |
| Crash reporting | On (anonymized) | Whether crash reports include device details |
| Error reporting | On (anonymized) | Whether errors are sent to improve the product |
| Telemetry collection | On (device-only) | Whether temperature/storage data is sent to cloud |
| Cloud sync | On | Whether device state is synced across clients |

All settings are accessible from Profile → Settings → Privacy.

### Third-Party Access

BuzzPi does not:
- Sell user data
- Share data with third parties
- Use analytics for advertising
- Require account creation for local operation

BuzzPi Cloud does not:
- Scan device content
- Profile user behavior
- Inject ads or promotions

### Anonymous Mode

BuzzPi can operate entirely without an account:
- Local device discovery (mDNS)
- Local pairing
- Local-only terminal, screen, files
- All data stored on-device

The only features lost are remote access (via relay) and cross-device sync. Anonymous mode has no time limit and no feature degradation beyond cloud-dependent features.

## User Experience

A user installs BuzzPi Runtime and opens the app. They pair a device over local network. Everything works. They have not created an account, not agreed to a privacy policy, and not transmitted any data outside their home network. No cloud dependency.

Months later, they want remote access to their device while traveling. They create an account, enable cloud sync, and the Relay Server routes their connection. They continue using the same app, the same devices — nothing changes except remote connectivity.

At any point, they can delete their account and return to local-only mode. Their devices remain paired and functional.

## Tradeoffs

| Tradeoff | Rationale |
|----------|-----------|
| No cloud means no remote access | Remote access is a premium feature that requires cloud infrastructure. Most users want it. It is opt-in, not default. |
| No analytics makes product improvement harder | BuzzPi relies on opt-in analytics and direct user feedback. This is slower than mandatory telemetry but respects user privacy. |
| Local-first adds implementation complexity | The dual-mode architecture (local + cloud) is more complex than cloud-only. This is a deliberate engineering investment in privacy. |

## Examples

- First-run experience: no account required, no sign-up wall
- Device pairing: works over local network without internet
- Anonymous mode: full functionality (minus relay) without any data leaving home
- Account deletion: deletes cloud data, returns to anonymous mode, devices remain paired
- Privacy settings: visible and configurable from the profile screen, not buried

## Related Patterns

- [Offline First](offline-first.md): Local mode is the default; cloud is optional
- [Configuration by Convention](configuration-by-convention.md): Privacy defaults are chosen for maximum user protection
- [Audit by Default](audit-by-default.md): Audit logs exclude sensitive content by design
