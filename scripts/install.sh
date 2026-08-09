#!/bin/sh
# Agent-Ready POSIX installer (spec §3.1): downloads the release asset,
# verifies its sha256 checksum, and installs the binary into PREFIX/bin.
# Never pipes curl to sh; never requires sudo by default.
set -eu

REPO="JhonMA82/agent-ready"
VERSION="${VERSION:-latest}"
PREFIX="${PREFIX:-$HOME/.local}"
ASSET_URL="${ASSET_URL:-}"
# Private repositories require a token (GITHUB_TOKEN/GH_TOKEN); public
# releases work without one.
AUTH=()
if [ -n "${GITHUB_TOKEN:-}" ]; then
  AUTH=(-H "Authorization: Bearer $GITHUB_TOKEN")
elif [ -n "${GH_TOKEN:-}" ]; then
  AUTH=(-H "Authorization: Bearer $GH_TOKEN")
fi

usage() {
  echo "Usage: VERSION=<semver> PREFIX=<dir> [ASSET_URL=<url>] $0 [--asset-url <url>]" >&2
  echo "  VERSION defaults to 'latest' (resolves the latest release); PREFIX defaults to \$HOME/.local" >&2
  exit 1
}

# Accept --asset-url as a positional argument too (workflow convenience).
while [ "$#" -gt 0 ]; do
  case "$1" in
    --asset-url) ASSET_URL="${2:-}"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; usage ;;
  esac
done

case "$(uname -s)" in
  Linux)  OS=linux ;;
  Darwin) OS=darwin ;;
  MINGW*|MSYS*|CYGWIN*) echo "Windows: download the zip asset from the release page and add it to PATH (no installer in V1)." >&2; exit 1 ;;
  *) echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

# 'latest' resolves the newest release tag via the GitHub API, then downloads
# the versioned asset from that tag (goreleaser emits versioned names).
if [ "$VERSION" = "latest" ]; then
  TAG="$(curl -fsSL "${AUTH[@]}" "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -n1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  if [ -z "$TAG" ]; then
    echo "Could not resolve the latest release tag; set VERSION explicitly." >&2
    exit 1
  fi
  BASE="${ASSET_URL:-https://github.com/$REPO/releases/download/$TAG}"
  ASSET="agent-ready_${TAG#v}_${OS}_${ARCH}.tar.gz"
else
  BASE="${ASSET_URL:-https://github.com/$REPO/releases/download/$VERSION}"
  ASSET="agent-ready_${VERSION}_${OS}_${ARCH}.tar.gz"
fi
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $BASE/$ASSET"
curl -fL "${AUTH[@]}" -o "$TMP/$ASSET" "$BASE/$ASSET"
curl -fL "${AUTH[@]}" -o "$TMP/checksums.txt" "$BASE/checksums.txt"

# Fail closed on checksum mismatch: nothing is installed. The checksums.txt
# line may carry a relative or absolute asset path; compare hashes directly.
EXPECTED="$(grep "${ASSET}$" "$TMP/checksums.txt" | awk '{print $1}' | head -n1)"
ACTUAL="$(sha256sum "$TMP/$ASSET" | awk '{print $1}')"
if [ -z "$EXPECTED" ] || [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Checksum verification failed for $ASSET; nothing installed." >&2
  exit 1
fi

tar -xzf "$TMP/$ASSET" -C "$TMP"
mkdir -p "$PREFIX/bin"
install -m 0755 "$TMP/agent-ready" "$PREFIX/bin/agent-ready"

echo "Installed agent-ready to $PREFIX/bin/agent-ready"
if ! echo "$PATH" | grep -q "$PREFIX/bin"; then
  echo "Add it to your PATH: export PATH=\"$PREFIX/bin:\$PATH\"" >&2
fi
