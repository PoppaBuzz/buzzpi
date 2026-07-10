# Port Reference

**Complete port allocation table for the BuzzPi Platform.**

This document catalogs every network port used by BuzzPi components. All ports are registered in the platform's port allocation table to prevent conflicts.

---

## 1. Port Allocation Table

| Port | Protocol | Component | Purpose | Configurable |
|------|----------|-----------|---------|--------------|
| 10104 | TCP | Agent | BPP WebSocket server | Yes |
| 10105 | UDP | Agent | mDNS advertisement | No |
| 10106 | TCP | Relay | Relay WebSocket server | Yes |
| 10107 | TCP | Relay | Relay REST API | Yes |
| 10108 | TCP | Agent | Health check HTTP endpoint | Yes |
| 10109 | TCP | Agent | Metrics endpoint (prometheus) | Yes |
| 10110-10119 | TCP | Agent | Plugin HTTP ports (per-plugin) | Yes |
| 10120-10129 | TCP | Agent | Plugin gRPC ports (per-plugin) | Yes |
| 22 | TCP | Agent | SSH forwarding session | No |
| 5353 | UDP | Agent | mDNS (system-wide) | No |

**Total allocated: 16 ports + dynamic ranges**

---

## 2. Reserved Range

All BuzzPi ports live in the **10104-10129** range (26 ports):

```
10104 ─ Agent BPP WebSocket
10105 ─ Agent mDNS
10106 ─ Relay WebSocket
10107 ─ Relay REST API
10108 ─ Agent Health Check HTTP
10109 ─ Agent Metrics (Prometheus)
10110-10119 ─ Plugin HTTP (10 ports)
10120-10129 ─ Plugin gRPC (10 ports)
```

**Why 10104?** It reads as "10104" = "IO10/4" — pun on "I/O" and "BPP v1.0.4" placeholder. It is also unlikely to conflict with any well-known or ephemeral range.

---

## 3. Port Details

### 3.1 Agent BPP WebSocket (10104/TCP)

- **Component:** `agent/connection`
- **Protocol:** WebSocket over TCP
- **Encryption:** WSS (TLS) by default, WS allowed for local LAN
- **Path:** `/bpp/v1`
- **Config env:** `BUZZPI_AGENT_PORT`

### 3.2 Agent mDNS (10105/UDP)

- **Component:** `agent/mdns`
- **Protocol:** DNS-SD / mDNS (RFC 6762, RFC 6763)
- **Service type:** `_buzzpi._tcp`
- **Service name:** `<device-id>@<friendly-name>`
- **TXT records:** `proto=v1`, `platform=<platform>`, `version=<version>`, `caps=<capabilities>`
- **Config env:** `BUZZPI_AGENT_MDNS_ENABLED`

### 3.3 Relay WebSocket (10106/TCP)

- **Component:** `relay`
- **Protocol:** WebSocket over TCP
- **Path:** `/relay/v1`
- **Config env:** `BUZZPI_RELAY_URL`

### 3.4 Relay REST API (10107/TCP)

- **Component:** `relay`
- **Protocol:** HTTPS
- **Endpoints:** `/api/v1/register`, `/api/v1/devices`, `/api/v1/stats`
- **Config env:** `BUZZPI_RELAY_API_URL`

### 3.5 Agent Health Check (10108/TCP)

- **Component:** `agent`
- **Protocol:** HTTP
- **Endpoints:** `/healthz` (liveness), `/readyz` (readiness), `/live` (simple)
- **Config env:** `BUZZPI_AGENT_HEALTH_PORT`

### 3.6 Agent Metrics (10109/TCP)

- **Component:** `agent`
- **Protocol:** HTTP
- **Format:** Prometheus text format
- **Endpoints:** `/metrics`
- **Config env:** `BUZZPI_AGENT_METRICS_PORT`

### 3.7 Plugin Ports (10110-10129/TCP)

Allocated sequentially per plugin:

| Plugin Index | HTTP Port | gRPC Port |
|-------------|-----------|-----------|
| 1 | 10110 | 10120 |
| 2 | 10111 | 10121 |
| 3 | 10112 | 10122 |
| 4 | 10113 | 10123 |
| 5 | 10114 | 10124 |
| 6 | 10115 | 10125 |
| 7 | 10116 | 10126 |
| 8 | 10117 | 10127 |
| 9 | 10118 | 10128 |
| 10 | 10119 | 10129 |

---

## 4. Ephemeral Considerations

### Operating System Ephemeral Ranges

BuzzPi ports (10104-10129) may conflict with OS ephemeral ranges on some systems:

| OS | Default Ephemeral Range | Conflict? |
|----|------------------------|-----------|
| Linux | 32768-60999 | No |
| macOS | 49152-65535 | No |
| Windows | 49152-65535 | No |
| Windows (old) | 1025-5000 | **Yes** |

**Mitigation for Windows (old):**
```
netsh int ipv4 set dynamicport tcp start=49152 num=16384
```

---

## 5. Docker / Container Networking

When running the Agent in Docker, ports must be explicitly mapped:

```yaml
# docker-compose.yml
services:
  agent:
    ports:
      - "10104:10104"  # BPP WebSocket
      - "10105:10105/udp"  # mDNS
      - "10108:10108"  # Health check
    network_mode: host  # Recommended for mDNS
```

**Note:** mDNS requires `network_mode: host` on Linux for discovery to work.

---

## 6. Firewall Rules

### For the Agent (server):

```bash
# Allow BPP WebSocket
iptables -A INPUT -p tcp --dport 10104 -j ACCEPT

# Allow mDNS advertisements
iptables -A INPUT -p udp --dport 10105 -j ACCEPT

# Allow health checks (internal only)
iptables -A INPUT -p tcp --dport 10108 -s 127.0.0.1 -j ACCEPT
```

### For the CLI (client):

```bash
# Outbound connections only — typically unrestricted
iptables -A OUTPUT -p tcp --dport 10104 -j ACCEPT
iptables -A OUTPUT -p tcp --dport 10106 -j ACCEPT  # Relay
```

---

## 7. Port Conflict Detection

On startup, each component checks its port:

```
INFO[0000] binding BPP WebSocket to :10104
INFO[0000] port 10104 is available
```

On conflict:

```
FATAL[0000] port 10104 already in use (bind: address already in use)
FATAL[0000] suggestion: set BUZZPI_AGENT_PORT to an available port
```

---

## 8. Port Configuration Reference

| Config Key | Env Variable | Default | Component |
|-----------|-------------|---------|-----------|
| `agent.port` | `BUZZPI_AGENT_PORT` | `10104` | Agent |
| `agent.health_port` | `BUZZPI_AGENT_HEALTH_PORT` | `10108` | Agent |
| `agent.metrics_port` | `BUZZPI_AGENT_METRICS_PORT` | `10109` | Agent |
| `relay.ws_port` | `BUZZPI_RELAY_PORT` | `10106` | Relay |
| `relay.api_port` | `BUZZPI_RELAY_API_PORT` | `10107` | Relay |

---

## 9. Port Verification

Run `buzzpi doctor` to verify port availability:

```
$ buzzpi doctor
Port Check:
  ✓ 10104 (BPP WebSocket) — available
  ✓ 10105 (mDNS) — available
  ✓ 10108 (Health) — available
  ✓ 10109 (Metrics) — available
```

---

## References

- Reference: Configuration Reference (configurable ports)
- Reference: Environment Reference (env vars)
- Reference: Platform Reference (filesystem layout)
- RFC-0002: Runtime Architecture
