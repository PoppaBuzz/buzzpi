#!/usr/bin/env bash
#
# BuzzPi Agent — Post-Install Verification
#
# Run this after installation to verify everything is working.
# Usage: sudo ./verify.sh
#
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "  ${GREEN}✓${NC} $*"; }
fail() { echo -e "  ${RED}✗${NC} $*"; FAILURES=$((FAILURES + 1)); }
warn() { echo -e "  ${YELLOW}!${NC} $*"; }
info() { echo -e "  $*"; }

FAILURES=0

echo "BuzzPi Agent — Verification"
echo "==========================="
echo ""

# ── 1. Binary ────────────────────────────────────────────────────────
echo "1. Binary"
if [[ -x /usr/local/bin/buzzpi-runtime ]]; then
    pass "buzzpi-runtime installed at /usr/local/bin/buzzpi-runtime"
    VERSION=$(/usr/local/bin/buzzpi-runtime --version 2>&1 || echo "unknown")
    info "   Version: ${VERSION}"
else
    fail "buzzpi-runtime not found or not executable"
fi
echo ""

# ── 2. User ──────────────────────────────────────────────────────────
echo "2. System user"
if id buzzpi &>/dev/null; then
    pass "User 'buzzpi' exists"
else
    fail "User 'buzzpi' not found"
fi
echo ""

# ── 3. Directories ───────────────────────────────────────────────────
echo "3. Directories"
for dir in /etc/buzzpi /var/lib/buzzpi /var/log/buzzpi /var/lib/buzzpi/plugins; do
    if [[ -d "$dir" ]]; then
        pass "$dir exists"
    else
        fail "$dir missing"
    fi
done
echo ""

# ── 4. Configuration ─────────────────────────────────────────────────
echo "4. Configuration"
if [[ -f /etc/buzzpi/runtime.json ]]; then
    pass "Config file exists"
    if python3 -m json.tool /etc/buzzpi/runtime.json > /dev/null 2>&1; then
        pass "Config is valid JSON"
    else
        fail "Config is not valid JSON"
    fi
else
    fail "Config file missing at /etc/buzzpi/runtime.json"
fi
echo ""

# ── 5. Systemd service ───────────────────────────────────────────────
echo "5. Systemd service"
if [[ -f /etc/systemd/system/buzzpi.service ]]; then
    pass "Service file exists"
else
    fail "Service file missing"
fi

if systemctl is-enabled --quiet buzzpi.service 2>/dev/null; then
    pass "Service is enabled (will start on boot)"
else
    fail "Service is not enabled"
fi

if systemctl is-active --quiet buzzpi.service 2>/dev/null; then
    pass "Service is running"
else
    fail "Service is not running"
    warn "Try: sudo systemctl start buzzpi"
fi
echo ""

# ── 6. State database ────────────────────────────────────────────────
echo "6. State database"
if [[ -f /var/lib/buzzpi/state.db ]]; then
    pass "State DB exists at /var/lib/buzzpi/state.db"
    DBSIZE=$(du -h /var/lib/buzzpi/state.db | cut -f1)
    info "   Size: ${DBSIZE}"
else
    warn "State DB not found (will be created on first start)"
fi
echo ""

# ── 7. Network ───────────────────────────────────────────────────────
echo "7. Network"

# Check if port 8420 (or configured port) is listening
LISTENING=$(ss -tlnp 2>/dev/null | grep -c "buzzpi-runtime" || true)
if [[ "$LISTENING" -gt 0 ]]; then
    pass "BuzzPi is listening on a port"
    ss -tlnp 2>/dev/null | grep "buzzpi-runtime" | awk '{print "   " $4}' || true
else
    warn "No listening port detected (service may not be running yet)"
fi

# Check mDNS
if command -v avahi-browse &>/dev/null; then
    MDNS_RESULT=$(avahi-browse -t -p _buzzpi._tcp 2>/dev/null | head -1 || true)
    if [[ -n "$MDNS_RESULT" ]]; then
        pass "mDNS advertisement visible"
    else
        warn "mDNS advertisement not found (may take a few seconds)"
    fi
else
    warn "avahi-browse not installed, skipping mDNS check"
fi
echo ""

# ── 8. Logs ──────────────────────────────────────────────────────────
echo "8. Logs"
if command -v journalctl &>/dev/null; then
    LOG_ENTRIES=$(journalctl -u buzzpi --no-pager -n 1 2>/dev/null | wc -l || echo "0")
    if [[ "$LOG_ENTRIES" -gt 0 ]]; then
        pass "Logs available via journalctl"
        info "   Latest:"
        journalctl -u buzzpi --no-pager -n 3 2>/dev/null | sed 's/^/   /' || true
    else
        warn "No log entries yet"
    fi
else
    warn "journalctl not available"
fi
echo ""

# ── Summary ──────────────────────────────────────────────────────────
echo "==========================="
if [[ $FAILURES -eq 0 ]]; then
    echo -e "${GREEN}All checks passed!${NC}"
    echo ""
    echo "BuzzPi is installed and running."
    echo "Open the BuzzPi Android app to discover this device."
else
    echo -e "${RED}${FAILURES} check(s) failed${NC}"
    echo ""
    echo "Review the failures above and fix them before using BuzzPi."
    exit 1
fi
