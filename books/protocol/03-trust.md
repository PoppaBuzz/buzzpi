# BPP Chapter 3: Trust

**Layer:** Identity  
**Status:** Draft  
**Version:** 1.0.0

Trust management covers key lifecycle, revocation, and recovery — what happens when trust is lost or compromised.

## Key Lifecycle

### Key Generation

Both client and device generate Ed25519 keypairs:

```go
type KeyPair struct {
    PublicKey  [32]byte  // Ed25519 public key
    PrivateKey [64]byte  // Ed25519 private key (seed + public key)
}
```

Generation requirements:
- MUST use a cryptographically secure random number generator (e.g., Go's `crypto/rand`, Android's `SecureRandom`)
- MUST NOT generate keys from weak entropy (timestamps, device IDs, user input)
- SHOULD use hardware-backed key storage where available (Android Keystore, TPM)

### Key Rotation

Keys SHOULD be rotated periodically. Rotation is a pairing-level operation:

1. Client initiates rotation by sending a key update message to the device
2. Device verifies the message is signed by the current (old) client key
3. Both sides exchange new public keys, signed by their old private keys
4. Both sides update their stored public keys
5. Old keys are discarded

```json
{
  "type": "trust.key_rotation",
  "new_public_key": "<hex>",
  "signature": "<signed_with_old_private_key(new_public_key)>"
}
```

Rotation does not require re-pairing. The existing trust relationship is transferred to the new key.

### Key Storage

**Client side:**
- Private key stored in platform secure storage (Android Keystore, iOS Keychain)
- Device public keys stored in app-local encrypted database
- Key export disabled (keys cannot be transferred between devices)

**Device side (Runtime):**
- Private key stored at `/var/lib/buzzpi/identity.key` (file mode 0600)
- Client public keys stored at `/var/lib/buzzpi/authorized_clients/`
- File permissions enforced on every read

## Revocation

### Client-Initiated Unpairing

The user can unpair a device at any time:

1. Client sends `relay.device.unpair` to the Relay Server
2. Relay Server forwards the request to the device
3. Device removes the client's public key from its authorized list
4. Relay Server removes the device from the user's device list
5. Device generates a new pairing code and enters pairing mode
6. Client deletes the device's public key and session state

### Device-Initiated Unpairing

If the Runtime is reset or re-paired with a different account:

1. Device clears its authorized clients list
2. Device generates a new identity key (optional, configurable)
3. Device enters pairing mode
4. Any client attempting to connect receives `auth.failure`
5. Client shows "Device no longer paired. Pair again?"

### Remote Unpairing (via Relay)

A user can unpair a device from a different client:

1. User authenticates to Relay Server (via any client or web portal)
2. User selects device to unpair
3. Relay Server sends `relay.device.force_unpair` to the device
4. Device removes all authorized clients
5. Device enters pairing mode
6. User's other clients receive `device.unpaired` notification

### Account Deletion

When a user deletes their BuzzPi account:

1. Relay Server sends `relay.account.deleted` to all paired devices
2. Each device removes the user's authorized clients
3. Each device enters pairing mode
4. User's session tokens are revoked
5. User's device list is deleted

## Compromise Recovery

### Client Key Compromised

If the user suspects their client's identity key is compromised:

1. User rotates their client key (from a trusted device or web portal)
2. All paired devices receive the new public key
3. Old key is added to a revocation list
4. Sessions authenticated with the old key are terminated
5. User is prompted to re-authenticate their other clients

### Device Key Compromised

If a device's identity key is compromised:

1. User triggers remote unpairing from the web portal or another client
2. Remote command: device generates a new identity key
3. Device enters pairing mode
4. User re-pairs the device from a trusted client
5. Old device key is added to a revocation list

### Relay Server Compromised

In the event of a Relay Server breach:

1. All session tokens are revoked
2. All users are prompted to re-authenticate
3. Device keys are not affected (private keys never leave the device)
4. Client keys are not affected (private keys never leave the client)
5. WebRTC P2P connections are not affected (Relay Server cannot decrypt them)
6. Only relay-routed traffic was exposed (TURN connections, which are encrypted but relay-terminated)

## Trust on First Use (TOFU) Limits

BPP uses TOFU for pairing, which means the first connection to a device establishes trust. To mitigate TOFU risks:

| Risk | Mitigation |
|------|------------|
| User pairs with wrong device | Pairing code is shown on the device's display or stdout; user must enter it on the client |
| Malicious device on network | Code verification proves physical/local access; device must present valid key during key exchange |
| Relay Server serves fake device | Device key is pinned on first exchange; subsequent connections verify against stored key |
| User error (accidental pairing) | Immediate unpairing available; device can be reset |

## Trust Model Summary

```
┌─────────────────────────────────────────────┐
│              Client                          │
│  ┌─────────────────────────────────┐        │
│  │  Identity Key (Ed25519)         │        │
│  │  - Generated on first launch    │        │
│  │  - Stored in platform Keystore  │        │
│  │  - Never transmitted            │        │
│  └─────────────────────────────────┘        │
│  ┌─────────────────────────────────┐        │
│  │  Device Public Keys             │        │
│  │  - One per paired device        │        │
│  │  - Pinned during pairing        │        │
│  │  - Used to verify device        │        │
│  └─────────────────────────────────┘        │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│              Device (Runtime)                │
│  ┌─────────────────────────────────┐        │
│  │  Identity Key (Ed25519)         │        │
│  │  - Generated on first start     │        │
│  │  - Stored at identity.key       │        │
│  │  - Never transmitted            │        │
│  └─────────────────────────────────┘        │
│  ┌─────────────────────────────────┐        │
│  │  Authorized Client Public Keys  │        │
│  │  - One per paired client        │        │
│  │  - Stored in authorized_clients │        │
│  │  - Used to verify client        │        │
│  └─────────────────────────────────┘        │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│              Relay Server                    │
│  ┌─────────────────────────────────┐        │
│  │  Session Tokens                 │        │
│  │  - Opaque, time-limited         │        │
│  │  - Scoped to user's devices     │        │
│  │  - Cannot impersonate identity  │        │
│  │  - Revocable                    │        │
│  └─────────────────────────────────┘        │
│  Does NOT store identity keys.              │
│  Does NOT have access to device keys.       │
│  Does NOT have access to client keys.       │
└─────────────────────────────────────────────┘
```
