#!/usr/bin/env bash
# One-liner install script for Snippet-Snap
# Usage: curl -sSL https://raw.githubusercontent.com/O-Aditya/snippet-snap/main/scripts/install.sh | bash

set -euo pipefail

REPO="O-Aditya/Snippet-snap"
BINARY="snap"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux)  EXT="tar.gz" ;;
  darwin) EXT="tar.gz" ;;
  *)      echo "Unsupported OS: $OS (use Windows installer instead)"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "Could not determine latest release."
  exit 1
fi
VERSION="${LATEST#v}"

FILENAME="snippet-snap_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

echo "Installing snippet-snap ${LATEST} for ${OS}/${ARCH}..."
echo "  → ${URL}"

# Download and extract
TMP=$(mktemp -d)
curl -sSL "$URL" -o "${TMP}/${FILENAME}"

mkdir -p "$INSTALL_DIR"
tar -xzf "${TMP}/${FILENAME}" -C "$TMP"
mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"
rm -rf "$TMP"

# Check PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "  Add this to your shell profile:"
  echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
  echo ""
fi

echo "✓ Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
echo "  Run 'snap --help' to get started."
