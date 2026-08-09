#!/bin/sh
# Automated installer harness: proves the clean-install path and the
# tampered-checksum fail-closed path against a locally served fake release.
# Usage: scripts/test-install.sh   (requires go and python3)
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORK="$(mktemp -d)"
REL="$WORK/release"
PREFIX_OK="$WORK/prefix-ok"
PREFIX_BAD="$WORK/prefix-bad"
PORT="${TEST_PORT:-18799}"
trap 'rm -rf "$WORK"' EXIT

mkdir -p "$REL" "$WORK/tarroot"
go build -ldflags \
  "-X github.com/JhonMA82/agent-ready/internal/version.Version=9.9.9 -X github.com/JhonMA82/agent-ready/internal/version.Commit=harness" \
  -o "$WORK/tarroot/agent-ready" "$ROOT/cmd/agent-ready"
tar -czf "$REL/agent-ready_9.9.9_linux_amd64.tar.gz" -C "$WORK/tarroot" agent-ready
sha256sum "$REL/agent-ready_9.9.9_linux_amd64.tar.gz" > "$REL/checksums.txt"

python3 -m http.server "$PORT" --directory "$REL" >/dev/null 2>&1 &
SRV=$!
sleep 1
trap 'kill $SRV 2>/dev/null || true; rm -rf "$WORK"' EXIT

# 1. Clean install via positional --asset-url.
VERSION=9.9.9 PREFIX="$PREFIX_OK" \
  sh "$ROOT/scripts/install.sh" --asset-url "http://127.0.0.1:$PORT" >/dev/null 2>&1
"$PREFIX_OK/bin/agent-ready" --version | grep -q "agent-ready 9.9.9 (harness)" \
  || { echo "FAIL: clean install version mismatch"; exit 1; }
echo "PASS: clean install"

# 2. Tampered checksum must fail closed with nothing installed.
echo "badhash  agent-ready_9.9.9_linux_amd64.tar.gz" > "$REL/checksums.txt"
if VERSION=9.9.9 PREFIX="$PREFIX_BAD" \
   sh "$ROOT/scripts/install.sh" --asset-url "http://127.0.0.1:$PORT" >/dev/null 2>&1; then
  echo "FAIL: tampered checksum must fail"; exit 1
fi
test ! -x "$PREFIX_BAD/bin/agent-ready" || { echo "FAIL: tampered install wrote a binary"; exit 1; }
echo "PASS: tampered checksum fail-closed"

echo "install harness: ALL PASS"
