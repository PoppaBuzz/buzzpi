# Book 7: Reference

**The BuzzPi dictionary.**

Reference is not documentation. Reference is where you look things up. Every fact, schema, token, code, and constant in the BuzzPi Platform is catalogued here.

Think of this as the BuzzPi equivalent of MDN or Android Developers — the definitive source of truth for every concrete detail.

**Start here:** [North Star](../../NORTH_STAR.md) — [Constitution](../../CONSTITUTION.md) — [Reference Constraints](constraints.md)

## Contents

### Protocol

| Document | Status | Description |
|----------|--------|-------------|
| [Protocol Constants](protocol-constants.md) | Complete | Ports, URLs, timing constants, rate limits, capability IDs |
| [Error Codes](error-codes.md) | Complete | Complete error code registry (1xxx-6xxx) with recovery hints |
| [Packet Types](packet-types.md) | Complete | Complete list of BPP packet types with IDs and descriptions |
| [Event Catalog](event-catalog.md) | Complete | Every event the platform emits, with payload schemas |

### Schemas

| Document | Status | Description |
|----------|--------|-------------|
| [JSON Schemas](json-schemas.md) | Complete | Formal JSON Schema definitions for all protocol messages |
| [Plugin Manifest](plugin-manifest.md) | Complete | Plugin manifest schema, fields, validation rules |
| [Config Schema](configuration-reference.md) | Complete | Runtime, backend, and client configuration options |

### Design Tokens

| Document | Status | Description |
|----------|--------|-------------|
| [Design Tokens](design-tokens.md) | Complete | Colors, typography, spacing, animation, elevation tokens |

### Platform

| Document | Status | Description |
|----------|--------|-------------|
| [Platform Reference](platform-reference.md) | Complete | Filesystem layout, env vars, Android permissions, build targets |
| [REST Endpoints](rest-endpoints.md) | Complete | Every backend API endpoint with request/response schemas |
| [CLI Commands](cli-reference.md) | Complete | All CLI commands, flags, and arguments |
| [WebSocket Events](websocket-events.md) | Complete | Every WebSocket event with direction and payload |
| [Network Topologies](network-topologies.md) | Complete | All connection topologies with latency/bandwidth profiles |

### Configuration

| Document | Status | Description |
|----------|--------|-------------|
| [Environment Reference](environment-reference.md) | Complete | Full env variable registry with defaults and validation rules |
| [Port Reference](port-reference.md) | Complete | Complete port allocation table with firewall rules |

### SDK

| Document | Status | Description |
|----------|--------|-------------|
| [BPP Client API](bpp-client-api.md) | Complete | Public API surface for client implementations with Go interfaces |
| [Plugin API](plugin-api-reference.md) | Complete | Public API surface for plugin development with SDK examples |

### Operations

| Document | Status | Description |
|----------|--------|-------------|
| [File Format Reference](file-format-reference.md) | Complete | All file formats, paths, and encoding specifications |
| [Troubleshooting Guide](troubleshooting-guide.md) | Complete | Diagnosis steps and solutions for common issues |
| [Compatibility Matrix](compatibility-matrix.md) | Complete | Platform, OS, and hardware compatibility across all components |
