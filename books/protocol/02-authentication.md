# BPP Chapter 2: Authentication

**Layer:** Identity  
**Status:** Draft  
**Version:** 1.0.0

After pairing, all communication between client and device is authenticated. This chapter defines how the client proves its identity to the device and vice versa.

## Overview

Authentication uses a challenge-response protocol based on the Ed25519 keys exchanged during pairing.

Every authenticated session requires:
1. **Proof of possession** — both sides prove they hold the private key corresponding to the public key exchanged during pairing
2. **Session binding** — the authenticated session is bound to a specific connection (WebRTC data channel or WebSocket)
3. **Freshness** — authentication messages include nonces to prevent replay attacks

## Challenge-Response

### Client Authentication (Client → Device)

When a client connects to a device, the device challenges the client to prove its identity:

```
Device → Client: {
  "type": "auth.challenge",
  "challenge": <random_nonce_32_bytes_hex>,
  "session_id": <uuid>
}

Client → Device: {
  "type": "auth.response",
  "challenge": <same_challenge_from_device>,
  "signature": <Ed25519_Sign(client_private_key, challenge + session_id)>,
  "client_id": <client_public_key_hex>
}

Device: Verifies signature against stored client public key
Device → Client: {
  "type": "auth.success",
  "session_id": <uuid>
}
(or)
Device → Client: {
  "type": "auth.failure",
  "reason": "Invalid signature"
}
```

### Device Authentication (Device → Client)

The same protocol runs in reverse to authenticate the device to the client. This prevents client-side impersonation (e.g., connecting to a fake Runtime):

```
Client → Device: {
  "type": "auth.challenge",
  "challenge": <random_nonce_32_bytes_hex>,
  "session_id": <uuid>
}

Device → Client: {
  "type": "auth.response",
  "challenge": <same_challenge_from_client>,
  "signature": <Ed25519_Sign(device_private_key, challenge + session_id)>,
  "device_id": <device_public_key_hex>
}

Client: Verifies signature against stored device public key
```

## Session Tokens

For WebSocket connections to the Relay Server (not direct device-to-client connections), session tokens are used instead of full challenge-response per message.

### Token Acquisition

After device authentication (during pairing), the user receives a session token from the Relay Server:

```
{
  "type": "auth.token",
  "token": <opaque_session_token>,
  "expires_at": <iso_timestamp>,
  "scope": ["devices:list", "devices:connect", "notifications:receive"]
}
```

### Token Usage

The session token is sent as a header or query parameter on WebSocket connections to the Relay Server:

```
WebSocket URL: wss://relay.buzzpi.dev/ws?token=<session_token>
```

Or as a header:

```
X-Auth-Token: <session_token>
```

### Token Lifetime

| Token Type | Lifetime | Renewal |
|------------|----------|---------|
| Session token | 24 hours | Refresh token (30-day validity) |
| API token (CLI) | 90 days | Manual renewal |
| Pairing token | 5 minutes | Single-use, replaced on expiry |

## Replay Protection

Every challenge-response exchange includes:

1. **Nonce** — 32 random bytes, generated per challenge
2. **Session ID** — UUID v7, generated per connection
3. **Timestamp** — embedded in signed message (optional, for additional protection)

The verifier rejects:
- Duplicate nonces (tracked in a rolling window)
- Nonces older than 30 seconds
- Signatures over session IDs that don't match the current connection

## Mutual Authentication

BPP requires mutual authentication for all direct (WebRTC) connections:

1. Device challenges client → client proves identity
2. Client challenges device → device proves identity
3. Both verifications must pass before any service messages are exchanged

If either verification fails, the connection is terminated immediately.

## Session Binding

The authenticated session is cryptographically bound to the WebRTC connection:

1. The challenge and response messages are sent over the WebRTC data channel (encrypted by DTLS)
2. The session ID is unique per WebRTC connection
3. The signature includes the session ID, binding the authentication to that connection
4. If the WebRTC connection is lost, a new session must be established

This prevents a third party from injecting authentication messages into an established connection.

## Re-authentication

In long-running sessions, re-authentication may be required:

| Trigger | Action |
|---------|--------|
| Every 24 hours of session time | Full challenge-response re-authentication |
| Connection interrupted (WebRTC dropped) | New session, new challenge-response |
| Device identity key rotated | New pairing required |
| Client identity key rotated | Challenge-response with new key; if verification fails, re-pairing required |
