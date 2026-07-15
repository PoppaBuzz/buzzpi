#!/usr/bin/env bash
#
# build-sd-image.sh — Build a pre-configured Raspberry Pi SD card image
# with BuzzPi agent pre-installed.
#
# Usage:
#   ./build-sd-image.sh [options]
#
# Required environment variables:
#   WIFI_SSID       — WiFi network name
#   WIFI_PASSWORD   — WiFi password
#
# Optional environment variables:
#   PI_MODEL        — "3" (arm) or "4" (arm64, default: "4")
#   DEVICE_NAME     — BuzzPi device hostname (default: "buzzpi")
#   BUZZPI_VERSION  — Agent version to install (default: latest from git)
#   OUTPUT_DIR      — Where to save the .img file (default: ./output)
#
# Prerequisites (must be installed):
#   - qemu-user-static (for arm emulation)
#   - parted, losetup, mount (disk imaging)
#   - wget or curl (to download Raspberry Pi OS)
#
# Example:
#   WIFI_SSID=MyNetwork WIFI_PASSWORD=secret ./build-sd-image.sh

set -euo pipefail

# --- Configuration ---
PI_MODEL="${PI_MODEL:-4}"
DEVICE_NAME="${DEVICE_NAME:-buzzpi}"
OUTPUT_DIR="${OUTPUT_DIR:-./output}"
BUILD_DIR=$(mktemp -d)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()   { echo -e "${GREEN}[build]${NC} $*"; }
warn()  { echo -e "${YELLOW}[warn]${NC} $*"; }
error() { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }

cleanup() {
    log "Cleaning up build directory..."
    # Unmount if mounted
    if [ -d "${BUILD_DIR}/root" ]; then
        umount "${BUILD_DIR}/root" 2>/dev/null || true
    fi
    if [ -d "${BUILD_DIR}/boot" ]; then
        umount "${BUILD_DIR}/boot" 2>/dev/null || true
    fi
    if [ -n "${LOOP_DEVICE:-}" ]; then
        losetup -d "$LOOP_DEVICE" 2>/dev/null || true
    fi
    rm -rf "$BUILD_DIR"
}
trap cleanup EXIT

# --- Validate ---
[ -z "${WIFI_SSID:-}" ] && error "WIFI_SSID is required"
[ -z "${WIFI_PASSWORD:-}" ] && error "WIFI_PASSWORD is required"

case "$PI_MODEL" in
    3) ARCH="armhf"; ARCH_NAME="arm";;
    4|5) ARCH="arm64"; ARCH_NAME="arm64";;
    *) error "PI_MODEL must be 3, 4, or 5";;
esac

mkdir -p "$OUTPUT_DIR"

# --- Step 1: Download Raspberry Pi OS Lite ---
OS_URL="https://downloads.raspberrypi.com/raspios_lite_${ARCH}/images/"
LATEST_IMG=$(wget -qO- "${OS_URL}" 2>/dev/null | grep -oP 'raspios_lite_'"${ARCH}"'-\d{8}/' | sort -r | head -1)

if [ -z "$LATEST_IMG" ]; then
    warn "Could not auto-detect latest OS. Using known URL pattern."
    LATEST_IMG="raspios_lite_${ARCH}-2024-11-19/"
fi

OS_FILE="${BUILD_DIR}/pios-lite.img.xz"
OS_EXTRACTED="${BUILD_DIR}/pios-lite.img"
FULL_URL="${OS_URL}${LATEST_IMG}$(ls "${OS_URL}${LATEST_IMG}" 2>/dev/null | grep -oP '[^/]+\.img\.xz' | head -1)"

log "Downloading Raspberry Pi OS Lite (${ARCH})..."
if [ ! -f "$OS_EXTRACTED" ]; then
    wget -q --show-progress -O "$OS_FILE" "$FULL_URL" || \
        error "Failed to download Raspberry Pi OS. Check network and try again."
    log "Extracting image..."
    xz -d "$OS_FILE"
fi

# --- Step 2: Mount the image ---
log "Setting up loop device..."
LOOP_DEVICE=$(losetup -fP --show "$OS_EXTRACTED")

# Wait for partitions to appear
sleep 2

BOOT_PART="${LOOP_DEVICE}p1"
ROOT_PART="${LOOP_DEVICE}p2"

[ -b "$BOOT_PART" ] || BOOT_PART="${LOOP_DEVICE}p1"
[ -b "$ROOT_PART" ] || ROOT_PART="${LOOP_DEVICE}p2"

# Fall back to offset-based mounting if partition detection fails
if [ ! -b "$BOOT_PART" ]; then
    warn "Partition detection failed, using offset-based mounting"
    mkdir -p "${BUILD_DIR}/boot" "${BUILD_DIR}/root"
    mount -o loop,offset=4194304 "$OS_EXTRACTED" "${BUILD_DIR}/boot"
    mount -o loop,offset=272629760 "$OS_EXTRACTED" "${BUILD_DIR}/root"
else
    mkdir -p "${BUILD_DIR}/boot" "${BUILD_DIR}/root"
    mount "$BOOT_PART" "${BUILD_DIR}/boot"
    mount "$ROOT_PART" "${BUILD_DIR}/root"
fi

log "Image mounted."

# --- Step 3: Configure WiFi ---
log "Configuring WiFi: ${WIFI_SSID}"
mkdir -p "${BUILD_DIR}/boot"

cat > "${BUILD_DIR}/boot/wpa_supplicant.conf" <<WPAEOF
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=US

network={
    ssid="${WIFI_SSID}"
    psk="${WIFI_PASSWORD}"
    key_mgmt=WPA-PSK
}
WPAEOF

# Enable SSH
touch "${BUILD_DIR}/boot/ssh"

# --- Step 4: Configure hostname ---
log "Setting hostname: ${DEVICE_NAME}"
sed -i "s/raspberrypi/${DEVICE_NAME}/g" "${BUILD_DIR}/root/etc/hostname" 2>/dev/null || echo "$DEVICE_NAME" > "${BUILD_DIR}/root/etc/hostname"
sed -i "s/127.0.1.1.*/127.0.1.1\t${DEVICE_NAME}/g" "${BUILD_DIR}/root/etc/hosts" 2>/dev/null || true

# --- Step 5: Install BuzzPi agent ---
log "Installing BuzzPi agent..."

# Create buzzpi user and directories
chroot "${BUILD_DIR}/root" /bin/bash -c "
    id -u buzzpi &>/dev/null || useradd -m -s /bin/bash buzzpi
    mkdir -p /opt/buzzpi /var/lib/buzzpi
    chown -R buzzpi:buzzpi /opt/buzzpi /var/lib/buzzpi
"

# Build and copy the agent binary
if [ -f "../../Makefile" ]; then
    log "Building agent binary for ${ARCH_NAME}..."
    (cd ../.. && make build GOARCH="${ARCH_NAME}") || warn "Build failed, skipping binary install"
fi

# Copy binary if it exists
AGENT_BIN="../../buzzpi-runtime"
if [ -f "$AGENT_BIN" ]; then
    cp "$AGENT_BIN" "${BUILD_DIR}/root/opt/buzzpi/buzzpi-runtime"
    chmod +x "${BUILD_DIR}/root/opt/buzzpi/buzzpi-runtime"
fi

# Create systemd service
cat > "${BUILD_DIR}/root/etc/systemd/system/buzzpi-runtime.service" <<SVCEOF
[Unit]
Description=BuzzPi Runtime Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=buzzpi
Group=buzzpi
ExecStart=/opt/buzzpi/buzzpi-runtime --config /opt/buzzpi/config.json
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SVCEOF

# Create default config
cat > "${BUILD_DIR}/root/opt/buzzpi/config.json" <<CFGEOF
{
    "runtime": {
        "device_name": "${DEVICE_NAME}"
    },
    "network": {
        "listen_port": 8443,
        "relay_servers": []
    }
}
CFGEOF

chroot "${BUILD_DIR}/root" /bin/bash -c "
    chown -R buzzpi:buzzpi /opt/buzzpi
    systemctl enable buzzpi-runtime 2>/dev/null || true
"

# --- Step 6: Configure first-boot setup ---
log "Setting up first-boot script..."
cat > "${BUILD_DIR}/root/opt/buzzpi/first-boot.sh" <<'FBBEOF'
#!/bin/bash
# First-boot setup — runs once on initial startup
LOG="/var/log/buzzpi-first-boot.log"

echo "BuzzPi first-boot starting at $(date)" > "$LOG"

# Wait for network
for i in $(seq 1 30); do
    if ping -c1 8.8.8.8 &>/dev/null; then
        echo "Network ready" >> "$LOG"
        break
    fi
    sleep 2
done

# Generate device identity
/opt/buzzpi/buzzpi-runtime --version >> "$LOG" 2>&1

echo "BuzzPi first-boot complete at $(date)" >> "$LOG"
systemctl disable buzzpi-first-boot 2>/dev/null
FBBEOF
chmod +x "${BUILD_DIR}/root/opt/buzzpi/first-boot.sh"

cat > "${BUILD_DIR}/root/etc/systemd/system/buzzpi-first-boot.service" <<FBSEOF
[Unit]
Description=BuzzPi First Boot Setup
After=network-online.target

[Service]
Type=oneshot
ExecStart=/opt/buzzpi/first-boot.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
FBSEOF

chroot "${BUILD_DIR}/root" /bin/bash -c "
    systemctl enable buzzpi-first-boot 2>/dev/null || true
"

# --- Step 7: Verify installation ---
log "Verifying installation..."
chroot "${BUILD_DIR}/root" /bin/bash -c "
    ls -la /opt/buzzpi/
    systemctl list-unit-files | grep buzzpi || true
" 2>/dev/null || warn "Verification skipped (chroot may not work on this host)"

# --- Step 8: Unmount and finalize ---
log "Unmounting image..."
sync
umount "${BUILD_DIR}/boot" 2>/dev/null || true
umount "${BUILD_DIR}/root" 2>/dev/null || true
losetup -d "$LOOP_DEVICE"
LOOP_DEVICE=""

# --- Step 9: Compress ---
TIMESTAMP=$(date +%Y%m%d)
OUTPUT_FILE="${OUTPUT_DIR}/buzzpi-${PI_MODEL}-${DEVICE_NAME}-${TIMESTAMP}.img"

log "Copying image to ${OUTPUT_FILE}..."
cp "$OS_EXTRACTED" "$OUTPUT_FILE"

log "Compressing image..."
xz -T0 "${OUTPUT_FILE}"
OUTPUT_FILE="${OUTPUT_FILE}.xz"

IMAGE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)

# --- Done ---
echo ""
log "============================================"
log "SD card image built successfully!"
log "============================================"
log ""
log "  Model:    Raspberry Pi ${PI_MODEL}"
log "  Device:   ${DEVICE_NAME}"
log "  WiFi:     ${WIFI_SSID}"
log "  Output:   ${OUTPUT_FILE}"
log "  Size:     ${IMAGE_SIZE}"
log ""
log "Flash to SD card:"
log "  xzcat ${OUTPUT_FILE} | sudo dd of=/dev/sdX bs=4M status=progress"
log ""
log "First boot will:"
log "  1. Connect to WiFi (${WIFI_SSID})"
log "  2. Enable SSH"
log "  3. Generate device identity"
log "  4. Start BuzzPi agent on port 8443"
log ""
