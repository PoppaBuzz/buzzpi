#!/usr/bin/env bash
#
# BuzzPi Agent — Raspberry Pi Installer
#
# Usage:
#   Interactive:   sudo ./install.sh
#   Uninstall:     sudo ./install.sh --uninstall
#   Upgrade:       sudo ./install.sh --upgrade
#
# This script installs the BuzzPi Runtime daemon on a Debian-based
# system (Raspberry Pi OS). It creates a dedicated user, installs
# the binary, sets up a systemd service, and starts the daemon.
#
set -euo pipefail

# ── Constants ────────────────────────────────────────────────────────
BUZZPI_USER="buzzpi"
BUZZPI_GROUP="buzzpi"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/buzzpi"
DATA_DIR="/var/lib/buzzpi"
LOG_DIR="/var/log/buzzpi"
PLUGIN_DIR="${DATA_DIR}/plugins"
SERVICE_FILE="/etc/systemd/system/buzzpi.service"
CONFIG_FILE="${CONFIG_DIR}/runtime.json"
DB_FILE="${DATA_DIR}/state.db"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Colors ───────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }
die()   { error "$*"; exit 1; }

# ── Pre-flight checks ───────────────────────────────────────────────
check_root() {
    if [[ $EUID -ne 0 ]]; then
        die "This script must be run as root (use sudo)"
    fi
}

check_binary() {
    # Look for the binary in build/ or current directory
    local candidates=(
        "${SCRIPT_DIR}/build/buzzpi-runtime-linux-arm64"
        "${SCRIPT_DIR}/build/buzzpi-runtime-linux-arm"
        "${SCRIPT_DIR}/build/buzzpi-runtime"
        "${SCRIPT_DIR}/buzzpi-runtime"
    )

    BINARY_PATH=""
    for candidate in "${candidates[@]}"; do
        if [[ -f "$candidate" ]]; then
            BINARY_PATH="$candidate"
            break
        fi
    done

    if [[ -z "$BINARY_PATH" ]]; then
        die "BuzzPi binary not found. Build it first:\n  make cross-runtime-arm64\nor copy the binary to: ${SCRIPT_DIR}/"
    fi

    info "Found binary: ${BINARY_PATH}"
}

# ── Detect architecture ──────────────────────────────────────────────
detect_arch() {
    local machine
    machine=$(uname -m)
    case "$machine" in
        aarch64|arm64)  info "Detected architecture: arm64 (Pi 4/5)" ;;
        armv7l|armhf)   info "Detected architecture: armv7 (Pi 3 or older)" ;;
        x86_64)         info "Detected architecture: amd64" ;;
        *)              warn "Unknown architecture: ${machine}" ;;
    esac
}

# ── Create user ──────────────────────────────────────────────────────
create_user() {
    if id "$BUZZPI_USER" &>/dev/null; then
        info "User '${BUZZPI_USER}' already exists"
    else
        info "Creating user '${BUZZPI_USER}'..."
        useradd -r -s /usr/sbin/nologin -d "$DATA_DIR" "$BUZZPI_USER"
        info "User '${BUZZPI_USER}' created"
    fi
}

# ── Create directories ───────────────────────────────────────────────
create_dirs() {
    info "Creating directories..."
    mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR" "$PLUGIN_DIR"
    chown -R "$BUZZPI_USER:$BUZZPI_GROUP" "$DATA_DIR" "$LOG_DIR"
    chmod 755 "$CONFIG_DIR"
    chmod 700 "$DATA_DIR"
    info "Directories created"
}

# ── Install binary ──────────────────────────────────────────────────
install_binary() {
    info "Installing binary..."
    cp "$BINARY_PATH" "${INSTALL_DIR}/buzzpi-runtime"
    chmod 755 "${INSTALL_DIR}/buzzpi-runtime"
    info "Binary installed: ${INSTALL_DIR}/buzzpi-runtime"
}

# ── Install configuration ───────────────────────────────────────────
install_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        warn "Config file exists at ${CONFIG_FILE}, skipping (use --upgrade to overwrite)"
        return
    fi

    info "Creating default configuration..."
    tee "$CONFIG_FILE" > /dev/null << 'EOF'
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
    chown root:root "$CONFIG_FILE"
    chmod 644 "$CONFIG_FILE"
    info "Configuration created: ${CONFIG_FILE}"
}

# ── Install systemd service ─────────────────────────────────────────
install_service() {
    info "Installing systemd service..."
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
    info "Service installed"
}

# ── Start service ────────────────────────────────────────────────────
start_service() {
    info "Enabling and starting BuzzPi service..."
    systemctl enable buzzpi.service
    systemctl restart buzzpi.service
    sleep 2

    if systemctl is-active --quiet buzzpi.service; then
        info "BuzzPi is running"
    else
        warn "BuzzPi may not have started. Check: sudo journalctl -u buzzpi -n 20"
    fi
}

# ── Print summary ────────────────────────────────────────────────────
print_summary() {
    echo ""
    echo "=========================================="
    info "BuzzPi Agent installed successfully!"
    echo "=========================================="
    echo ""
    echo "  Binary:     ${INSTALL_DIR}/buzzpi-runtime"
    echo "  Config:     ${CONFIG_FILE}"
    echo "  State DB:   ${DB_FILE}"
    echo "  Logs:       journalctl -u buzzpi -f"
    echo "  Service:    sudo systemctl {start|stop|restart} buzzpi"
    echo ""
    echo "  Edit config:  sudo nano ${CONFIG_FILE}"
    echo "  View logs:    sudo journalctl -u buzzpi -f"
    echo "  Check status: sudo systemctl status buzzpi"
    echo ""
}

# ── Uninstall ────────────────────────────────────────────────────────
uninstall() {
    info "Uninstalling BuzzPi Agent..."

    # Stop and disable service
    if systemctl is-active --quiet buzzpi.service 2>/dev/null; then
        systemctl stop buzzpi.service
    fi
    if systemctl is-enabled --quiet buzzpi.service 2>/dev/null; then
        systemctl disable buzzpi.service
    fi

    # Remove files
    rm -f "$SERVICE_FILE"
    rm -f "${INSTALL_DIR}/buzzpi-runtime"
    rm -rf "$CONFIG_DIR"
    rm -rf "$DATA_DIR"
    rm -rf "$LOG_DIR"

    # Remove user
    if id "$BUZZPI_USER" &>/dev/null; then
        userdel "$BUZZPI_USER" 2>/dev/null || true
    fi

    systemctl daemon-reload
    info "BuzzPi Agent uninstalled"
    echo ""
    warn "Note: State database was deleted. Reinstalling will create a new device identity."
}

# ── Upgrade ──────────────────────────────────────────────────────────
upgrade() {
    info "Upgrading BuzzPi Agent..."
    install_binary
    systemctl restart buzzpi.service
    sleep 2

    if systemctl is-active --quiet buzzpi.service; then
        info "Upgrade complete, BuzzPi is running"
    else
        warn "BuzzPi may not have restarted. Check: sudo journalctl -u buzzpi -n 20"
    fi
}

# ── Main ─────────────────────────────────────────────────────────────
main() {
    local action="${1:-install}"

    case "$action" in
        --uninstall|-u)
            check_root
            uninstall
            exit 0
            ;;
        --upgrade|-U)
            check_root
            check_binary
            upgrade
            exit 0
            ;;
        --help|-h)
            echo "BuzzPi Agent Installer"
            echo ""
            echo "Usage:"
            echo "  sudo $0                Install BuzzPi Agent"
            echo "  sudo $0 --upgrade      Upgrade binary and restart"
            echo "  sudo $0 --uninstall    Remove BuzzPi Agent"
            echo "  $0 --help              Show this help"
            exit 0
            ;;
        install)
            check_root
            check_binary
            detect_arch
            create_user
            create_dirs
            install_binary
            install_config
            install_service
            start_service
            print_summary
            ;;
        *)
            die "Unknown option: ${action}. Use --help for usage."
            ;;
    esac
}

main "$@"
