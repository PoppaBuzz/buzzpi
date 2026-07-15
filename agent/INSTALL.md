# BuzzPi Agent — Raspberry Pi Installation

This guide covers installing the BuzzPi Runtime daemon on a Raspberry Pi running Raspberry Pi OS (Debian-based).

## Prerequisites

- Raspberry Pi (3B+ or newer recommended) with Raspberry Pi OS (Bookworm or later)
- Network connectivity (Wi-Fi or Ethernet)
- `sudo` access on the Pi
- SSH access (or physical keyboard/monitor)

## Quick Install

From your development machine (Linux/macOS/WSL):

```bash
# 1. Cross-compile for Raspberry Pi
cd agent
make cross-runtime-arm64    # for Pi 4/5 (64-bit OS)
# or
make cross-runtime-arm      # for Pi 3 or older (32-bit OS)

# 2. Copy binary to Pi
scp build/buzzpi-runtime-linux-arm64 pi@<PI_IP>:~/

# 3. SSH into Pi and run installer
ssh pi@<PI_IP>
chmod +x ~/buzzpi-runtime-linux-arm64
sudo ./buzzpi-runtime-linux-arm64 --install
```

## Manual Install

### 1. Create the buzzpi user

```bash
sudo useradd -r -s /usr/sbin/nologin -d /var/lib/buzzpi buzzpi
```

### 2. Create directories

```bash
sudo mkdir -p /etc/buzzpi
sudo mkdir -p /var/lib/buzzpi
sudo mkdir -p /var/log/buzzpi
sudo mkdir -p /var/lib/buzzpi/plugins
sudo chown -R buzzpi:buzzpi /var/lib/buzzpi
sudo chown -R buzzpi:buzzpi /var/log/buzzpi
```

### 3. Install the binary

```bash
sudo cp buzzpi-runtime-linux-arm64 /usr/local/bin/buzzpi-runtime
sudo chmod 755 /usr/local/bin/buzzpi-runtime
```

### 4. Create configuration

```bash
sudo tee /etc/buzzpi/runtime.json > /dev/null << 'EOF'
{
  "runtime": {
    "device_name": ""
  },
  "network": {
    "relay_servers": [],
    "listen_port": 8420,
    "mdns_enabled": true
  },
  "screen": {
    "capture_backend": "auto",
    "max_fps": 30,
    "default_quality": "high"
  },
  "plugins": {
    "enabled": true,
    "directory": "/var/lib/buzzpi/plugins",
    "allow_network": false
  },
  "logging": {
    "level": "info",
    "file": "/var/log/buzzpi/runtime.log",
    "max_size_mb": 100,
    "max_files": 5
  }
}
EOF
```

Set `device_name` to a friendly name (e.g., `"kitchen-pi"`). Leave empty to auto-generate from device ID.

### 5. Create systemd service

```bash
sudo tee /etc/systemd/system/buzzpi.service > /dev/null << 'EOF'
[Unit]
Description=BuzzPi Runtime Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=buzzpi
Group=buzzpi
ExecStart=/usr/local/bin/buzzpi-runtime --config /etc/buzzpi/runtime.json
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/buzzpi /var/log/buzzpi
PrivateTmp=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
WantedBy=multi-user.target
EOF
```

### 6. Enable and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable buzzpi
sudo systemctl start buzzpi
```

## Verify Installation

```bash
# Check service status
sudo systemctl status buzzpi

# View logs
sudo journalctl -u buzzpi -f

# Check mDNS advertisement (from another machine on the network)
avahi-browse -a -t    # Linux
# or
dns-sd -B _tcp .      # macOS
```

The Pi should appear as a BuzzPi device on your local network. Open the BuzzPi Android app and it should discover the device automatically.

## Configuration Reference

| Field | Default | Description |
|-------|---------|-------------|
| `runtime.device_name` | `""` (auto) | Friendly name shown in app |
| `network.listen_port` | `0` (random) | WebSocket port for BPP connections |
| `network.relay_servers` | `[]` | Cloud relay URLs for remote access |
| `network.mdns_enabled` | `true` | Advertise via mDNS on LAN |
| `screen.capture_backend` | `"auto"` | Screen capture method |
| `screen.max_fps` | `30` | Maximum frame rate for screen streaming |
| `plugins.enabled` | `true` | Enable plugin host |
| `logging.level` | `"info"` | Log level: debug, info, warn, error |
| `logging.file` | `""` (stdout only) | Log file path |

## CLI Flags

```
buzzpi-runtime [flags]

  --config string     Path to config file
  --device-name string Override device name
  --db string         Path to state database (default /var/lib/buzzpi/state.db)
  --relay string      Cloud relay server URL
  --version           Show version and exit
```

## Uninstall

```bash
sudo systemctl stop buzzpi
sudo systemctl disable buzzpi
sudo rm /etc/systemd/system/buzzpi.service
sudo rm /usr/local/bin/buzzpi-runtime
sudo rm -rf /etc/buzzpi /var/lib/buzzpi /var/log/buzzpi
sudo userdel buzzpi
sudo systemctl daemon-reload
```

## Troubleshooting

**Service fails to start:**
```bash
sudo journalctl -u buzzpi -n 50 --no-pager
```

**mDNS not visible on network:**
- Ensure port 5353/UDP is open: `sudo ufw allow 5353/udp`
- Check Avahi is running: `sudo systemctl status avahi-daemon`

**Permission denied errors:**
- Verify the `buzzpi` user owns `/var/lib/buzzpi`:
  ```bash
  sudo chown -R buzzpi:buzzpi /var/lib/buzzpi
  ```

**Port already in use:**
- Change `network.listen_port` in config, or set to `0` for random port

## Cross-Compilation Reference

| Target | Command |
|--------|---------|
| Pi 4/5 (64-bit OS) | `make cross-runtime-arm64` |
| Pi 3 or older (32-bit OS) | `make cross-runtime-arm` |
| x86_64 Linux | `make cross-runtime-amd64` |
| All targets | `make cross` |
