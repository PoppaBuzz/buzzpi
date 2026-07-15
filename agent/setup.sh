#!/usr/bin/env bash
#
# BuzzPi — One-Line Installer
#
# Install BuzzPi on a Raspberry Pi with a single command:
#
#   curl -sSL https://jphat.net/buzzpi/setup.sh | sudo bash
#
# Or if curl isn't available:
#
#   wget -qO- https://jphat.net/buzzpi/setup.sh | sudo bash
#
# What this does:
#   1. Detects your Pi's architecture (arm64/armv7)
#   2. Downloads the latest BuzzPi binary
#   3. Creates the buzzpi user and directories
#   4. Installs the systemd service
#   5. Starts BuzzPi — it's now discoverable on your network
#
# No configuration needed. No IP addresses. No SSH required (if run locally).
#
set -euo pipefail

# ── Configuration ────────────────────────────────────────────────────
BUZZPI_VERSION="${BUZZPI_VERSION:-latest}"
BUZZPI_REPO="${BUZZPI_REPO:-buzzpi/buzzpi}"
GITHUB_API="https://api.github.com/repos/${BUZZPI_REPO}/releases"

# ── Colors ───────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}!${NC} $*"; }
error() { echo -e "${RED}✗${NC} $*" >&2; }
die()   { error "$*"; exit 1; }
step()  { echo -e "\n${BOLD}${BLUE}▸${NC} ${BOLD}$*${NC}"; }

# ── Banner ───────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}BuzzPi Installer${NC}"
echo -e "Your Raspberry Pi. Anywhere. Instantly."
echo ""

# ── Check root ───────────────────────────────────────────────────────
if [[ $EUID -ne 0 ]]; then
    die "This installer must be run as root.\n\n  sudo bash setup.sh\n  or\n  curl -sSL https://jphat.net/buzzpi/setup.sh | sudo bash"
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

MISSING=()
for cmd in systemctl; do
    if ! command -v "$cmd" &>/dev/null; then
        MISSING+=("$cmd")
    fi
done

if [[ ${#MISSING[@]} -gt 0 ]]; then
    die "Missing required commands: ${MISSING[*]}\n\nThis installer requires a Debian-based system with systemd."
fi
info "All dependencies satisfied"

# ── Determine version ───────────────────────────────────────────────
step "Resolving version"

if [[ "$BUZZPI_VERSION" == "latest" ]]; then
    # Query GitHub API for latest release
    LATEST=$(curl -fsSL "https://api.github.com/repos/${BUZZPI_REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || true)
    if [[ -n "$LATEST" ]]; then
        BUZZPI_VERSION="$LATEST"
        info "Latest version: ${BUZZPI_VERSION}"
    else
        warn "Could not query GitHub API, using nightly build"
        BUZZPI_VERSION="nightly"
    fi
else
    info "Installing version: ${BUZZPI_VERSION}"
fi

# ── Download binary ──────────────────────────────────────────────────
step "Downloading BuzzPi Runtime"

DOWNLOAD_DIR=$(mktemp -d)
trap "rm -rf ${DOWNLOAD_DIR}" EXIT

if [[ "$BUZZPI_VERSION" == "nightly" ]]; then
    # Download from GitHub Actions or nightly builds
    DOWNLOAD_URL="https://github.com/${BUZZPI_REPO}/releases/download/nightly/buzzpi-runtime-linux-${ARCH_NAME}"
else
    DOWNLOAD_URL="https://github.com/${BUZZPI_REPO}/releases/download/${BUZZPI_VERSION}/buzzpi-runtime-linux-${ARCH_NAME}"
fi

BINARY_PATH="${DOWNLOAD_DIR}/buzzpi-runtime"

echo -n "  Downloading from ${DOWNLOAD_URL}... "
if curl -fsSL -o "$BINARY_PATH" "$DOWNLOAD_URL" 2>/dev/null; then
    echo -e "${GREEN}done${NC}"
elif wget -q -O "$BINARY_PATH" "$DOWNLOAD_URL" 2>/dev/null; then
    echo -e "${GREEN}done${NC}"
else
    echo -e "${RED}failed${NC}"
    die "Could not download BuzzPi binary.\n\nURL: ${DOWNLOAD_URL}\n\nIf this is a new release, the binary may not be published yet.\nTry building from source: make cross-runtime-${ARCH_NAME}"
fi

chmod +x "$BINARY_PATH"
info "Binary downloaded ($(du -h "$BINARY_PATH" | cut -f1))"

# ── Create user ──────────────────────────────────────────────────────
step "Setting up system user"

if id buzzpi &>/dev/null; then
    info "User 'buzzpi' already exists"
else
    useradd -r -s /usr/sbin/nologin -d /var/lib/buzzpi buzzpi
    info "Created user 'buzzpi'"
fi

# ── Create directories ───────────────────────────────────────────────
step "Creating directories"

mkdir -p /etc/buzzpi /var/lib/buzzpi /var/log/buzzpi /var/lib/buzzpi/plugins
chown -R buzzpi:buzzpi /var/lib/buzzpi /var/log/buzzpi
chmod 755 /etc/buzzpi
chmod 700 /var/lib/buzzpi
info "Directories ready"

# ── Install binary ──────────────────────────────────────────────────
step "Installing binary"

cp "$BINARY_PATH" /usr/local/bin/buzzpi-runtime
chmod 755 /usr/local/bin/buzzpi-runtime

VERSION_OUTPUT=$(/usr/local/bin/buzzpi-runtime --version 2>&1 || echo "unknown")
info "Installed: ${VERSION_OUTPUT}"

# ── Create configuration ────────────────────────────────────────────
step "Configuring BuzzPi"

if [[ ! -f /etc/buzzpi/runtime.json ]]; then
    tee /etc/buzzpi/runtime.json > /dev/null << 'EOF'
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
    info "Configuration created (edit /etc/buzzpi/runtime.json to customize)"
else
    info "Configuration exists, keeping current settings"
fi

# ── Install systemd service ─────────────────────────────────────────
step "Installing system service"

tee /etc/systemd/system/buzzpi.service > /dev/null << 'EOF'
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
echo -e "  ${BOLD}Configuration:${NC} /etc/buzzpi/runtime.json"
echo -e "  ${BOLD}State database:${NC} /var/lib/buzzpi/state.db"
echo -e "  ${BOLD}Logs:${NC}          /var/log/buzzpi/runtime.log"
echo ""
