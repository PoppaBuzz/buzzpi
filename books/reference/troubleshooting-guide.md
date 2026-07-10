# Troubleshooting Guide

**Common issues, diagnosis steps, and solutions for the BuzzPi Platform.**

---

## How to Use This Guide

1. **Identify symptoms** — Find the section that matches what you're seeing
2. **Run diagnostics** — Follow the diagnosis steps
3. **Apply the fix** — Try the recommended solution
4. **Verify** — Confirm the issue is resolved

**First step for any issue:** `buzzpi doctor` runs comprehensive diagnostics:

```
$ buzzpi doctor
✓ Config file
✓ mDNS availability
✓ WebSocket connectivity
✓ State store
✓ Paired devices
✗ Agent reachable: living-room-pi — timeout
```

---

## 1. Connection Issues

### 1.1 Cannot Discover Devices

**Symptoms:**
- `buzzpi discover` returns no devices
- `buzzpi doctor` shows `✗ mDNS`

**Diagnosis:**

```bash
# Check if mDNS is available on the system
buzzpi doctor

# On Linux, check avahi/resolved status
systemctl status avahi-daemon
systemctl status systemd-resolved

# Listen for mDNS advertisements
# Linux:
sudo tcpdump -i any port 5353

# macOS:
sudo tcpdump -i en0 port 5353
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| mDNS not installed | `sudo apt install avahi-daemon` (Linux) |
| mDNS service not running | `sudo systemctl start avahi-daemon` |
| Firewall blocking mDNS | `sudo ufw allow 5353/udp` |
| Different subnet | mDNS is link-local — devices must be on same subnet |
| Agent not running | SSH into device and check `systemctl status buzzpi-agent` |

### 1.2 Connection Timeout

**Symptoms:**
- `buzzpi device info <device>` hangs then times out
- `buzzpi doctor` shows `✗ Agent reachable`

**Diagnosis:**

```bash
# Check basic connectivity
ping <device-ip>

# Check if the port is open
nc -zv <device-ip> 10104

# Try connecting directly to the WebSocket
# (requires websocat or similar)
websocat ws://<device-ip>:10104/bpp/v1
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Agent not running | SSH into device and start the agent |
| Firewall blocking 10104 | `sudo ufw allow 10104/tcp` |
| Wrong IP/device | Verify with `buzzpi discover` |
| TLS mismatch | Check if agent expects WSS but CLI sends WS |

### 1.3 Connection Drops Intermittently

**Symptoms:**
- Terminal sessions disconnect after a few minutes
- File transfers fail mid-way

**Diagnosis:**

```bash
# Check agent logs for disconnection reasons
buzzpi doctor --export diagnostics.buzzpi-diag

# Monitor connection stability
buzzpi device monitor <device>
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Network congestion | Switch to wired Ethernet |
| WiFi power saving | Disable WiFi power saving on device |
| Heartbeat timeout | Increase `BUZZPI_AGENT_HEARTBEAT_TIMEOUT` |
| Firewall timeout | Check NAT/firewall TCP timeout settings |

---

## 2. Pairing Issues

### 2.1 PIN Mismatch

**Symptoms:**
- PIN prompt appears but verification fails

**Diagnosis:**

```bash
# Check agent logs
journalctl -u buzzpi-agent --since "5 minutes ago"

# Verify no other pairing in progress
buzzpi pair list
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| PIN entered incorrectly | Try again, ensure numerals match |
| Pairing timed out | PIN is valid for 120s — start over |
| Another client pairing | Wait for other session to complete |
| Token mismatch | Unpair and re-pair from scratch |

### 2.2 Pairing Already Exists

**Symptoms:**
- `Pairing failed: pairing already in progress`
- `Device is already paired`

**Solutions:**

```bash
# List current pairings
buzzpi pair list

# Remove existing pairing
buzzpi unpair <device>

# Reset pairing on the device
# SSH into device:
sudo systemctl stop buzzpi-agent
sudo rm -f /var/lib/buzzpi/state/state.db
sudo systemctl start buzzpi-agent
```

---

## 3. Terminal Session Issues

### 3.1 Terminal Shows Garbage Characters

**Symptoms:**
- Terminal output contains `^[[32m` or similar escape sequences

**Solutions:**

| Cause | Solution |
|-------|----------|
| Terminal type mismatch | Set correct TERM: `export TERM=xterm-256color` |
| Color codes not stripped | Use `buzzpi term exec` with `-- TERM=dumb` |
| Encoding issue | Ensure UTF-8 locale on device |

### 3.2 Terminal Disconnects

**Symptoms:**
- Session closes unexpectedly

**Solutions:**

| Cause | Solution |
|-------|----------|
| Session timeout | Increase `BUZZPI_AGENT_SESSION_TTL` |
| Heartbeat timeout | Check network stability |
| Resource limits | Agent may close idle sessions — increase `BUZZPI_AGENT_MAX_SESSIONS` |

---

## 4. File Transfer Issues

### 4.1 Transfer Fails Mid-way

**Symptoms:**
- File upload/download reaches X% then fails

**Diagnosis:**

```bash
# Check available disk space on device
buzzpi term exec <device> -- "df -h"

# Check network reliability
ping <device-ip> -c 100
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Disk full | Free space on device |
| Network timeout | Increase `BUZZPI_CLI_TIMEOUT` |
| File too large | Default max is 1GB — check `BUZZPI_AGENT_FILE_MAX_SIZE` |
| Chunk size mismatch | `BUZZPI_AGENT_FILE_CHUNK_SIZE` must match between client and agent |

### 4.2 Permission Denied

**Symptoms:**
- `buzzpi file upload` fails with permission error

**Solutions:**

```bash
# Check target directory permissions
buzzpi term exec <device> -- "ls -la /target/path"

# Upload to a writable location instead
buzzpi file upload local.txt <device> /tmp/
```

---

## 5. Plugin Issues

### 5.1 Plugin Fails to Install

**Symptoms:**
- `buzzpi plugin install` fails

**Diagnosis:**

```bash
# Check manifest format
buzzpi plugin list

# View agent plugin logs
journalctl -u buzzpi-agent | grep plugin
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Invalid manifest | Validate against Plugin Manifest schema |
| Binary incompatible | Plugin must match agent platform |
| Signature invalid | Check signing key or disable with `BUZZPI_PLUGIN_ALLOW_UNSIGNED=true` |
| Resource limit | Increase `BUZZPI_AGENT_PLUGIN_MAX_MEMORY` |
| Port conflict | Plugin ports 10110-10129 may be in use |

### 5.2 Plugin Runs but Doesn't Respond

**Symptoms:**
- Plugin installed successfully but doesn't handle commands

**Solutions:**

| Cause | Solution |
|-------|----------|
| Health check fails | Check `BUZZPI_AGENT_PLUGIN_MAX_STARTUP` — increase if slow |
| Missing dependency | Plugin logs show dependency errors |
| Sandbox restriction | `BUZZPI_PLUGIN_NETWORK_ACCESS` may block required access |

---

## 6. Relay Issues

### 6.1 Cannot Connect to Relay

**Symptoms:**
- `Relay connection failed`
- Devices not visible when both LAN and relay are up

**Diagnosis:**

```bash
# Check relay connectivity
curl -I https://relay.buzzpi.cloud

# Verify token validity
buzzpi doctor | grep relay
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Relay URL wrong | Check `BUZZPI_RELAY_URL` |
| Token expired | Re-authenticate with relay server |
| Network blocks WebSocket | Corporate firewalls may block port 443 — contact IT |
| Relay server down | Check status page or contact relay operator |

---

## 7. Agent Startup Issues

### 7.1 Agent Fails to Start

**Symptoms:**
- `systemctl start buzzpi-agent` fails
- Agent logs show fatal errors

**Diagnosis:**

```bash
# Check service status
systemctl status buzzpi-agent

# View startup logs
journalctl -u buzzpi-agent -n 50

# Run agent manually to see output
sudo -u buzzpi /usr/bin/buzzpi-agent
```

**Common Causes:**

| Error | Cause | Solution |
|-------|-------|----------|
| `port 10104 already in use` | Port conflict | Kill the process using port 10104, or change `BUZZPI_AGENT_PORT` |
| `config file not found` | Missing config | Copy default config to `/etc/buzzpi/agent.yaml` |
| `state directory not writable` | Permissions | `sudo chown buzzpi:buzzpi /var/lib/buzzpi` |
| `failed to generate identity` | Key failure | Check `/var/lib/buzzpi/certs/` is writable |
| `mDNS registration failed` | mDNS issue | `sudo setcap cap_net_raw+ep /usr/bin/buzzpi-agent` |

---

## 8. Performance Issues

### 8.1 High CPU Usage

**Symptoms:**
- Device fans spinning up while idle
- `top` shows buzzpi-agent using >30% CPU

**Diagnosis:**

```bash
# Check CPU usage
top -p $(pgrep buzzpi-agent)

# Check if screen streaming is active
buzzpi screen status

# Check active sessions
buzzpi session list
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Screen streaming active | Stop unused screen sessions |
| Too many plugins | Disable unnecessary plugins |
| High FPS setting | Reduce `BUZZPI_AGENT_SCREEN_FPS` |
| Large state database | Compact state: `systemctl restart buzzpi-agent` |

### 8.2 High Memory Usage

**Solutions:**

| Cause | Solution |
|-------|----------|
| Plugin memory leak | Restart the plugin: `buzzpi plugin disable && buzzpi plugin enable` |
| Too many sessions | Close idle terminal sessions |
| File buffer sizing | Reduce `BUZZPI_AGENT_FILE_CHUNK_SIZE` |

---

## 9. Platform-Specific Issues

### 9.1 Raspberry Pi

| Issue | Solution |
|-------|----------|
| mDNS not available | `sudo apt install avahi-daemon && sudo systemctl enable avahi-daemon` |
| Permissions for GPIO plugin | Add buzzpi user to `gpio` group: `sudo usermod -aG gpio buzzpi` |
| Swap file needed | `sudo dphys-swapfile setup && sudo dphys-swapfile swapon` |

### 9.2 Android

| Issue | Solution |
|-------|----------|
| Can't discover devices on same WiFi | Check AP isolation (client isolation) in WiFi settings |
| Connection drops on screen off | Disable battery optimization for BuzzPi |
| No notifications for devices | Grant notification permission in Android settings |

### 9.3 Docker

| Issue | Solution |
|-------|----------|
| mDNS not working | Use `network_mode: host` for the agent container |
| Can't access host plugins | Mount `/var/lib/buzzpi/plugins` as a volume |
| Port already in use | Map agent port to different host port |

---

## 10. Diagnostic Commands Reference

```bash
# Run all checks
buzzpi doctor

# Export diagnostics bundle
buzzpi doctor --export buzzpi-diag-$(date +%F).buzzpi-diag

# Check agent logs
journalctl -u buzzpi-agent -f

# Verify port availability
nc -zv localhost 10104

# Capturing WebSocket traffic for debugging
websocat ws://device-ip:10104/bpp/v1 -v

# Reset agent state (preserves pairing)
sudo systemctl stop buzzpi-agent
sudo rm -f /var/lib/buzzpi/state/state.db
sudo systemctl start buzzpi-agent

# Full factory reset
sudo systemctl stop buzzpi-agent
sudo rm -rf /var/lib/buzzpi/*
sudo systemctl start buzzpi-agent
```

---

## References

- Reference: Configuration Reference (tuning options)
- Reference: Environment Reference (env vars for tuning)
- Reference: Port Reference (port conflict resolution)
- Reference: Event Catalog (connection/disconnection events)
- RFC-0008: Implementation Roadmap
