#!/usr/bin/env bash
# One-liner install script for Snippet-Snap
# Usage: curl -sSL https://raw.githubusercontent.com/O-Aditya/snippet-snap/main/scripts/install.sh | bash

set -euo pipefail

REPO="O-Aditya/Snippet-snap"
BINARY="snip"
INSTALL_DIR="${HOME}/.local/bin"

# ── Colors ──
GREEN='\033[0;32m'
CYAN='\033[0;36m'
DIM='\033[0;90m'
BOLD='\033[1m'
RESET='\033[0m'

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

FILENAME="snippet-snap_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

echo ""
echo -e "  ${CYAN}◈  SNIPPET-SNAP${RESET}  ${DIM}installer${RESET}"
echo ""
echo -e "  ${DIM}Version${RESET}   ${LATEST}"
echo -e "  ${DIM}Platform${RESET}  ${OS}/${ARCH}"
echo -e "  ${DIM}Target${RESET}    ${INSTALL_DIR}/${BINARY}"
echo ""

# Download and extract
TMP=$(mktemp -d)
echo -e "  ${DIM}Downloading...${RESET}"
curl -sSL "$URL" -o "${TMP}/${FILENAME}"

mkdir -p "$INSTALL_DIR"
tar -xzf "${TMP}/${FILENAME}" -C "$TMP"
mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"
rm -rf "$TMP"

# ── Auto-add to PATH ──
PATH_ADDED=false
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  # Detect shell and profile file
  SHELL_NAME=$(basename "$SHELL" 2>/dev/null || echo "bash")
  case "$SHELL_NAME" in
    zsh)  PROFILE="$HOME/.zshrc" ;;
    bash)
      if [ -f "$HOME/.bash_profile" ]; then
        PROFILE="$HOME/.bash_profile"
      else
        PROFILE="$HOME/.bashrc"
      fi
      ;;
    fish) PROFILE="$HOME/.config/fish/config.fish" ;;
    *)    PROFILE="$HOME/.profile" ;;
  esac

  EXPORT_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
  if [ "$SHELL_NAME" = "fish" ]; then
    EXPORT_LINE="set -gx PATH ${INSTALL_DIR} \$PATH"
  fi

  # Add to profile if not already there
  if [ -f "$PROFILE" ] && grep -qF "$INSTALL_DIR" "$PROFILE" 2>/dev/null; then
    echo -e "  ${DIM}PATH already configured in ${PROFILE}${RESET}"
  else
    echo "" >> "$PROFILE"
    echo "# Snippet-Snap" >> "$PROFILE"
    echo "$EXPORT_LINE" >> "$PROFILE"
    PATH_ADDED=true
    echo -e "  ${GREEN}✓${RESET} Added to ${BOLD}${PROFILE}${RESET}"
  fi

  # Also add to current session
  export PATH="${INSTALL_DIR}:$PATH"
fi

echo ""
echo -e "  ${GREEN}✓${RESET} ${BOLD}Installed successfully!${RESET}"
echo ""

if [ "$PATH_ADDED" = true ]; then
  echo -e "  ${DIM}PATH was updated. Run this to use immediately:${RESET}"
  echo -e "    ${CYAN}source ${PROFILE}${RESET}"
  echo ""
fi

echo -e "  ${DIM}Get started:${RESET}"
echo -e "    ${CYAN}snip --help${RESET}"
echo -e "    ${CYAN}snip add --name my-snippet --lang bash${RESET}"
echo -e "    ${CYAN}snip find${RESET}"
echo ""
