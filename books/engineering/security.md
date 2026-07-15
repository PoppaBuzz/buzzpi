# Security Policy

BuzzPi takes security seriously. The project handles device access, remote connections, and system management, which means security is a core design requirement, not an afterthought.

---

## Reporting a Vulnerability

If you discover a security vulnerability in BuzzPi, please report it privately.

**Email**: security.buzzpi@jphat.net

**PGP Key**: Available at https://jphat.net/security/pgp-key.asc

Please include the following information in your report:

- Description of the vulnerability
- Steps to reproduce
- Affected versions (agent, Android app, backend)
- Potential impact
- Any suggested fix (optional)

---

## Response Timeline

| Timeframe | Action |
|-----------|--------|
| 48 hours | Acknowledgment of receipt |
| 5 business days | Initial assessment and severity classification |
| Ongoing | Updates every 5 business days until resolution |
| Resolution | Fix released and vulnerability disclosed |

---

## Responsible Disclosure

Please do not disclose the vulnerability publicly until we have had an opportunity to investigate, fix, and release a patch. We will coordinate disclosure timing with you.

---

## Bug Bounty

A formal bug bounty program is not yet available. This will be updated when one is established.

---

## Security Practices

### Communication

- All network communication uses TLS 1.3 minimum
- Certificate pinning prevents MITM attacks on the Android client
- WebSocket connections are always encrypted

### Authentication

- Device identity is established through public/private key pairs
- Keys are generated on-device and never transmitted in plaintext
- Android Keystore provides hardware-backed key storage
- Biometric authentication is supported for sensitive operations

### Session Security

- End-to-end encryption for all session data
- Sessions are individually revocable
- No session data is stored on BuzzPi infrastructure

### Privacy

- No telemetry collected by default
- No analytics or tracking
- No ads
- No data collection without explicit user consent
- Cloud relay forwards encrypted data only, never inspected

### Application Security

- Minimal Android permissions model
- Agent runs as a non-root systemd service where possible
- Command execution is sandboxed per session
- Plugin isolation prevents one plugin from affecting others

---

## Security.txt

See https://jphat.net/.well-known/security.txt for the latest security contact information.
