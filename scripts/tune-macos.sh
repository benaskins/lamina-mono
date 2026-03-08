#!/bin/bash
# tune-macos.sh — raise macOS limits for load testing axon services
#
# Run with: sudo ./scripts/tune-macos.sh
# Revert with: sudo ./scripts/tune-macos.sh --revert
#
# Changes are NOT persistent across reboots.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}Run with sudo${NC}"
    exit 1
fi

show_current() {
    echo "=== Current limits ==="
    echo "File descriptors (soft):  $(ulimit -n)"
    echo "File descriptors (hard):  $(ulimit -Hn)"
    echo "launchctl maxfiles:       $(launchctl limit maxfiles 2>/dev/null | awk '{print $2, $3}')"
    echo "Ephemeral port range:     $(sysctl -n net.inet.ip.portrange.first)-$(sysctl -n net.inet.ip.portrange.last)"
    echo "TCP MSL (TIME_WAIT/2):    $(sysctl -n net.inet.tcp.msl)ms"
    echo "Listen backlog (somaxconn): $(sysctl -n kern.ipc.somaxconn)"
    echo ""
}

tune() {
    echo -e "${YELLOW}Before:${NC}"
    show_current

    # File descriptors: raise soft limit
    launchctl limit maxfiles 65536 200000

    # Ephemeral port range: widen from ~16k to ~32k ports
    sysctl -w net.inet.ip.portrange.first=32768 >/dev/null

    # TCP TIME_WAIT: reduce from 30s to 5s (msl is half the TIME_WAIT duration, in ms)
    sysctl -w net.inet.tcp.msl=2500 >/dev/null

    # Listen backlog: raise from 128 to 2048
    sysctl -w kern.ipc.somaxconn=2048 >/dev/null

    echo -e "${GREEN}After:${NC}"
    show_current
    echo -e "${GREEN}Tuning applied.${NC} Remember to run ${YELLOW}ulimit -n 65536${NC} in your shell before testing."
    echo "Changes revert on reboot, or run: sudo $0 --revert"
}

revert() {
    echo -e "${YELLOW}Before:${NC}"
    show_current

    launchctl limit maxfiles 256 unlimited
    sysctl -w net.inet.ip.portrange.first=49152 >/dev/null
    sysctl -w net.inet.tcp.msl=15000 >/dev/null
    sysctl -w kern.ipc.somaxconn=128 >/dev/null

    echo -e "${GREEN}After:${NC}"
    show_current
    echo -e "${GREEN}Reverted to defaults.${NC}"
}

case "${1:-}" in
    --revert)
        revert
        ;;
    *)
        tune
        ;;
esac
