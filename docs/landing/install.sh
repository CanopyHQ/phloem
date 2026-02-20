#!/bin/bash
# Phloem Installer
# Usage: curl -sSL https://phloem.canopyhq.io/install.sh | bash

set -e

REPO="CanopyHQ/phloem"
INSTALL_DIR="$HOME/.local/bin"
APP_DIR="/Applications"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
echo "  ____  _     _                      "
echo " |  _ \| |__ | | ___   ___ _ __ ___  "
echo " | |_) | '_ \| |/ _ \ / _ \ '_ \` _ \ "
echo " |  __/| | | | | (_) |  __/ | | | | |"
echo " |_|   |_| |_|_|\___/ \___|_| |_| |_|"
echo -e "${NC}"
echo "Your AI Finally Remembers"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
esac

echo -e "${YELLOW}Detected: ${OS}/${ARCH}${NC}"
echo ""

# Check for Homebrew (preferred method on macOS)
if [[ "$OS" == "darwin" ]] && command -v brew &> /dev/null; then
    echo -e "${GREEN}Homebrew detected! Using brew install...${NC}"
    echo ""
    brew tap canopyhq/tap 2>/dev/null || true
    brew install phloem
    echo ""
    echo -e "${GREEN}✅ Phloem installed via Homebrew${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Run: phloem setup"
    echo ""
    exit 0
fi

# Manual installation
echo -e "${YELLOW}Installing manually...${NC}"

# Get latest release
echo "Fetching latest release..."
LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo -e "${RED}Failed to fetch latest release${NC}"
    exit 1
fi

echo -e "Latest version: ${GREEN}$LATEST${NC}"

# Download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/phloem-${OS}-${ARCH}.tar.gz"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract
echo "Downloading..."
TEMP_DIR=$(mktemp -d)
curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/phloem.tar.gz"

echo "Extracting..."
tar -xzf "$TEMP_DIR/phloem.tar.gz" -C "$TEMP_DIR"

# Install binary
echo "Installing to $INSTALL_DIR..."
mv "$TEMP_DIR/phloem" "$INSTALL_DIR/phloem"
chmod +x "$INSTALL_DIR/phloem"

# macOS: Install .app if available
if [[ "$OS" == "darwin" ]] && [ -d "$TEMP_DIR/Phloem.app" ]; then
    echo "Installing Phloem.app to /Applications..."
    cp -R "$TEMP_DIR/Phloem.app" "$APP_DIR/"
fi

# Cleanup
rm -rf "$TEMP_DIR"

# Add to PATH if needed
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo -e "${YELLOW}Add this to your shell profile (~/.zshrc or ~/.bashrc):${NC}"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo ""
echo -e "${GREEN}✅ Phloem installed successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Run: phloem setup"
echo "  2. Start capturing AI conversations!"
echo ""
echo "Documentation: https://phloem.canopyhq.io"
echo "Support: https://github.com/CanopyHQ/phloem/issues"
