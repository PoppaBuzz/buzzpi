#!/usr/bin/env bash
#
# BuzzPi — One-Line Installer
#
# Install BuzzPi on a Raspberry Pi with a single command:
#
#   curl -sSL https://raw.githubusercontent.com/PoppaBuzz/buzzpi/main/install.sh | sudo bash
#
# What this does:
#   1. Detects your Pi's architecture (arm64/armv7/amd64)
#   2. Downloads the latest BuzzPi binary from GitHub releases
#   3. Creates the buzzpi user and directories
#   4. Installs the systemd service
#   5. Starts BuzzPi — it's now discoverable on your network
#
# No configuration needed. No IP addresses. No SSH required (if run locally).
#
set -euo pipefail

# ── Configuration ────────────────────────────────────────────────────
BUZZPI_VERSION="${BUZZPI_VERSION:-v0.1.0}"
BUZZPI_REPO="${BUZZPI_REPO:-PoppaBuzz/buzzpi}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/buzzpi"
DATA_DIR="/var/lib/buzzpi"
LOG_DIR="/var/log/buzzpi"
SERVICE_FILE="/etc/systemd/system/buzzpi.service"

# ── Colors ───────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}!${NC} $*"; }
error() { echo -e "${RED}✗${NC} $*" >&2; }
die()   { error "$*"; exit 1; }
step()  { echo -e "\n${BOLD}▸${NC} ${BOLD}$*${NC}"; }

# ── Banner ───────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}BuzzPi Installer${NC}"
echo -e "Your Raspberry Pi. Anywhere. Instantly."
echo ""

# ── Check root ───────────────────────────────────────────────────────
if [[ $EUID -ne 0 ]]; then
    die "This installer must be run as root.\n\n  sudo bash install.sh"
fi

# ── Detect architecture ──────────────────────────────────────────────
step "Detecting architecture"

ARCH=$(uname -m)
case "$ARCH" in
    aarch64|arm64)
        ARCH_NAME="arm64"
        PI_MODEL="Pi 4/5 (64-bit)"
        ;;
    armv7l|armhf)
        ARCH_NAME="arm"
        PI_MODEL="Pi 3 or older (32-bit)"
        ;;
    x86_64|amd64)
        ARCH_NAME="amd64"
        PI_MODEL="x86_64 (test/dev)"
        ;;
    *)
        die "Unsupported architecture: ${ARCH}\n\nBuzzPi supports: arm64, armv7, amd64"
        ;;
esac
info "Detected: ${PI_MODEL} (${ARCH})"

# ── Check dependencies ──────────────────────────────────────────────
step "Checking dependencies"

for cmd in systemctl curl; do
    if ! command -v "$cmd" &>/dev/null; then
        die "Missing required command: ${cmd}"
    fi
done
info "All dependencies satisfied"

# ── Download binary ──────────────────────────────────────────────────
step "Downloading BuzzPi Runtime"

DOWNLOAD_DIR=$(mktemp -d)
trap "rm -rf ${DOWNLOAD_DIR}" EXIT

ARCHIVE_NAME="buzzpi-runtime-linux-${ARCH_NAME}.tar.gz"
DOWNLOAD_URL="https://github.com/${BUZZPI_REPO}/releases/download/${BUZZPI_VERSION}/${ARCHIVE_NAME}"

echo -n "  Downloading ${ARCHIVE_NAME}... "
if curl -fsSL -o "${DOWNLOAD_DIR}/${ARCHIVE_NAME}" "$DOWNLOAD_URL"; then
    echo -e "${GREEN}done${NC}"
else
    die "Could not download BuzzPi binary.\n\nURL: ${DOWNLOAD_URL}\n\nCheck that release ${BUZZPI_VERSION} exists with asset ${ARCHIVE_NAME}"
fi

echo -n "  Extracting... "
tar xzf "${DOWNLOAD_DIR}/${ARCHIVE_NAME}" -C "${DOWNLOAD_DIR}"
BINARY_PATH="${DOWNLOAD_DIR}/buzzpi-runtime-linux-${ARCH_NAME}"
chmod +x "$BINARY_PATH"
echo -e "${GREEN}done${NC}"
info "Binary downloaded ($(du -h "$BINARY_PATH" | cut -f1))"

# ── Create user ──────────────────────────────────────────────────────
step "Setting up system user"

if id buzzpi &>/dev/null; then
    info "User 'buzzpi' already exists"
else
    useradd -r -s /usr/sbin/nologin -d "$DATA_DIR" buzzpi
    info "Created user 'buzzpi'"
fi

# ── Create directories ───────────────────────────────────────────────
step "Creating directories"

mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR" "${DATA_DIR}/plugins"
chown -R buzzpi:buzzpi "$DATA_DIR" "$LOG_DIR"
chmod 755 "$CONFIG_DIR"
chmod 700 "$DATA_DIR"
info "Directories ready"

# ── Install binary ──────────────────────────────────────────────────
step "Installing binary"

cp "$BINARY_PATH" "${INSTALL_DIR}/buzzpi-runtime"
chmod 755 "${INSTALL_DIR}/buzzpi-runtime"

VERSION_OUTPUT=$("${INSTALL_DIR}/buzzpi-runtime" --version 2>&1 || echo "unknown")
info "Installed: ${VERSION_OUTPUT}"

# ── Create configuration ────────────────────────────────────────────
step "Configuring BuzzPi"

if [[ ! -f "${CONFIG_DIR}/runtime.json" ]]; then
    tee "${CONFIG_DIR}/runtime.json" > /dev/null << 'EOF'
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
    info "Configuration created (edit ${CONFIG_DIR}/runtime.json to customize)"
else
    info "Configuration exists, keeping current settings"
fi

# ── Install systemd service ─────────────────────────────────────────
step "Installing system service"

tee "$SERVICE_FILE" > /dev/null << 'EOF'
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

systemctl daemon-reload
systemctl enable buzzpi.service
info "Service installed and enabled"

# ── Start BuzzPi ─────────────────────────────────────────────────────
step "Starting BuzzPi"

systemctl restart buzzpi.service
sleep 2

if systemctl is-active --quiet buzzpi.service; then
    info "BuzzPi is running!"
else
    warn "BuzzPi may not have started. Checking logs..."
    journalctl -u buzzpi --no-pager -n 10 2>/dev/null || true
fi

# ── Done ─────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}═══════════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  BuzzPi is installed and running!${NC}"
echo -e "${GREEN}${BOLD}═══════════════════════════════════════════${NC}"
echo ""
echo -e "  Your Pi is now discoverable on the network."
echo -e "  Open the BuzzPi app on your phone to connect."
echo ""
echo -e "  ${BOLD}Quick commands:${NC}"
echo -e "    Status:   ${BOLD}sudo systemctl status buzzpi${NC}"
echo -e "    Logs:     ${BOLD}sudo journalctl -u buzzpi -f${NC}"
echo -e "    Restart:  ${BOLD}sudo systemctl restart buzzpi${NC}"
echo -e "    Stop:     ${BOLD}sudo systemctl stop buzzpi${NC}"
echo -e "    Uninstall:${BOLD} sudo /usr/local/bin/buzzpi-runtime --uninstall${NC}"
echo ""
echo -e "  ${BOLD}Configuration:${NC} ${CONFIG_DIR}/runtime.json"
echo -e "  ${BOLD}State database:${NC} ${DATA_DIR}/state.db"
echo -e "  ${BOLD}Logs:${NC}          ${LOG_DIR}/runtime.log"
echo ""
