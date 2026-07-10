# BPP Chapter 4: Identity

**Layer:** Identity  
**Status:** Draft  
**Version:** 1.0.0

Identity in BPP is decentralized. There is no central authority that assigns identities. Instead, identities are self-generated cryptographic keys, bound to human-readable names by the user.

## Device Identity

### Device ID

Every device has a permanent, immutable Device ID assigned on first Runtime startup:

```json
{
  "device_id": "018e0a3f-7b2a-7b00-8000-3a5c3d4e5f6a",
  "format": "UUID v7 (time-ordered)",
  "generated_at": "2026-07-07T12:00:00Z"
}
```

The Device ID is:
- Generated using UUID v7 (time-ordered, random suffix)
- Stored at `/var/lib/buzzpi/device.id`
- Never changes for the life of the device
- Used to identify the device in Relay Server communications
- Independent of the identity key (key can rotate, ID stays the same)

### Device Identity Key

The device's Ed25519 keypair serves as its cryptographic identity:

```json
{
  "public_key": "6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a",
  "algorithm": "Ed25519",
  "stored_at": "/var/lib/buzzpi/identity.key",
  "permissions": "0600 (owner read/write only)"
}
```

Relationship between Device ID and identity key:
- Device ID identifies the device to the Relay Server (for routing, presence)
- Identity key authenticates the device to clients (for trust, encryption)
- They are independent — key rotation does not change the Device ID

### Device Name

The device's friendly name is the only user-visible identifier:

- Assigned by the user during pairing
- Stored on the Relay Server (associated with the user's account)
- Is not exposed to the device (the device does not know its own friendly name)
- Can be changed by the user at any time
- Defaults to the device's hostname if not assigned

Maximum length: 64 characters  
Allowed characters: Unicode letters, digits, spaces, hyphens, apostrophes

### Hardware Identity

The device MAY expose its hardware identity for informational purposes:

```json
{
  "hardware_id": "00000000abcdef01",
  "model": "Raspberry Pi 5",
  "serial": "1234567890abcdef"
}
```

Hardware identity is:
- Read-only, provided by the device
- Not used for authentication
- Displayed in the Workspace for user reference
- Not guaranteed to be unique (some devices don't have serial numbers)

## Client Identity

### Client ID

Every client installation has a permanent Client ID generated on first app launch:

```json
{
  "client_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "format": "UUID v4 (random)",
  "generated_at": "2026-07-07T12:00:00Z"
}
```

### Client Identity Key

The client's Ed25519 keypair serves as its cryptographic identity:

```json
{
  "public_key": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a",
  "algorithm": "Ed25519",
  "storage": "Android Keystore (hardware-backed on supported devices)"
}
```

### Client Name

The client's display name (for multi-client scenarios):

- Set by the user: "Sarah's Phone", "Sarah's Laptop"
- Displayed in the device's authorized clients list
- Used to identify which client performed an action (audit log)
- Defaults to device model name if not set

## User Identity

User identity is managed by the Relay Server (or a BuzzPi Cloud account):

```json
{
  "user_id": "b2c3d4e5-f6a7-8b90-c1d2-e3f4a5b6c7d8",
  "email": "user@example.com",
  "display_name": "Sarah Chen",
  "created_at": "2026-01-15T10:30:00Z"
}
```

User identity is:
- Created when the user first signs up (via app or web)
- Used for Relay Server authentication (session tokens)
- NOT used for device-level authentication (that's the client identity key)
- Recoverable via email (password reset, account recovery)

### Anonymous Mode

BuzzPi supports anonymous/local-only mode:

- No Relay Server account required
- Pairing works over local network (mDNS discovery)
- No remote access (Relay Server features unavailable)
- No push notifications
- All data stored locally
- Can be upgraded to a full account later without re-pairing

## Identity Resolution

When a connection is established, both sides resolve their counterpart's identity:

```
Device → Client: sends device_id, public_key
Client resolves: which device is this? what's its friendly name?

Client → Device: sends client_id, public_key
Device resolves: is this client authorized? what's its client name?
```

Resolution is done locally (both sides have their own lookup tables) — no server query needed during connection.

## Disclosure Rules

| Context | Identity Disclosed |
|---------|-------------------|
| Pairing request (local network) | Device ID, device name (hostname) |
| Pairing request (relay) | Device ID only |
| Authenticated session | Full identity (device ID, public key, friendly name) |
| Unauthenticated connection attempt | Public key only (for verification) |
| Device discovery (mDNS) | Device ID, device name (hostname), protocol version |
| Relay Server presence | Device ID, online/offline status |
