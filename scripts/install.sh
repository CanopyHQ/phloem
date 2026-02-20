#!/bin/bash
# Phloem Complete Installation Script
# Installs everything needed for full Phloem functionality
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
VERSION="1.0.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}🧠 Phloem Complete Installer v$VERSION${NC}"
echo "============================================"
echo ""

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH_NAME="intel"
    GO_ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH_NAME="apple-silicon"
    GO_ARCH="arm64"
else
    echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
    exit 1
fi

echo -e "${BLUE}Detected: macOS $ARCH_NAME ($ARCH)${NC}"
echo ""

# Step 1: Check for existing installation
echo -e "${YELLOW}Step 1: Checking existing installation...${NC}"
if [ -d "/Applications/Phloem.app" ]; then
    echo "  Found existing /Applications/Phloem.app"
    read -p "  Replace existing installation? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "  Removing existing installation..."
        rm -rf "/Applications/Phloem.app"
    else
        echo "  Keeping existing installation"
    fi
fi
echo ""

# Step 2: Build Go binary for correct architecture
echo -e "${YELLOW}Step 2: Building Go binary for $ARCH_NAME...${NC}"
cd "$PROJECT_DIR"
GOOS=darwin GOARCH=$GO_ARCH go build -o phloem-$ARCH_NAME .
echo -e "  ${GREEN}✅ Built: phloem-$ARCH_NAME${NC}"
echo ""

# Step 3: Build Swift app for correct architecture
echo -e "${YELLOW}Step 3: Building Swift app...${NC}"
cd "$PROJECT_DIR/macos"
if [ "$ARCH" = "x86_64" ]; then
    swift build -c release --arch x86_64
else
    swift build -c release
fi
echo -e "  ${GREEN}✅ Built Swift app${NC}"
echo ""

# Step 4: Create app bundle
echo -e "${YELLOW}Step 4: Creating app bundle...${NC}"
BUILD_DIR="$PROJECT_DIR/macos/.build"
APP_BUNDLE="$BUILD_DIR/Phloem.app"
rm -rf "$APP_BUNDLE"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

# Copy Swift binary
if [ "$ARCH" = "x86_64" ]; then
    cp "$BUILD_DIR/x86_64-apple-macosx/release/Phloem" "$APP_BUNDLE/Contents/MacOS/Phloem"
else
    cp "$BUILD_DIR/release/Phloem" "$APP_BUNDLE/Contents/MacOS/Phloem"
fi

# Copy Go binary
cp "$PROJECT_DIR/phloem-$ARCH_NAME" "$APP_BUNDLE/Contents/Resources/phloem"

# Create Info.plist
cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>Phloem</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>io.canopyhq.phloem</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>Phloem</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>12.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2026 CanopyHQ. All rights reserved.</string>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
</dict>
</plist>
EOF

echo -n "APPL????" > "$APP_BUNDLE/Contents/PkgInfo"
echo -e "  ${GREEN}✅ Created app bundle${NC}"
echo ""

# Step 5: Install to /Applications
echo -e "${YELLOW}Step 5: Installing to /Applications...${NC}"
cp -r "$APP_BUNDLE" "/Applications/"
echo -e "  ${GREEN}✅ Installed to /Applications/Phloem.app${NC}"
echo ""

# Step 6: Configure Cursor MCP
echo -e "${YELLOW}Step 6: Configuring Cursor MCP...${NC}"
CURSOR_CONFIG="$HOME/.cursor/mcp.json"
CURSOR_DIR="$HOME/.cursor"
mkdir -p "$CURSOR_DIR"

PHLOEM_BIN="/Applications/Phloem.app/Contents/Resources/phloem"

if [ -f "$CURSOR_CONFIG" ]; then
    # Check if phloem already configured
    if grep -q '"phloem"' "$CURSOR_CONFIG" 2>/dev/null; then
        echo "  Phloem already in mcp.json, updating path..."
        # Use python to update the config
        python3 << PYEOF
import json
with open('$CURSOR_CONFIG', 'r') as f:
    config = json.load(f)
config['mcpServers']['phloem'] = {
    'command': '$PHLOEM_BIN',
    'args': ['serve']
}
with open('$CURSOR_CONFIG', 'w') as f:
    json.dump(config, f, indent=2)
print("  Updated phloem path in mcp.json")
PYEOF
    else
        # Add phloem to existing config
        python3 << PYEOF
import json
with open('$CURSOR_CONFIG', 'r') as f:
    config = json.load(f)
if 'mcpServers' not in config:
    config['mcpServers'] = {}
config['mcpServers']['phloem'] = {
    'command': '$PHLOEM_BIN',
    'args': ['serve']
}
with open('$CURSOR_CONFIG', 'w') as f:
    json.dump(config, f, indent=2)
print("  Added phloem to mcp.json")
PYEOF
    fi
else
    # Create new config
    cat > "$CURSOR_CONFIG" << EOF
{
  "mcpServers": {
    "phloem": {
      "command": "$PHLOEM_BIN",
      "args": ["serve"]
    }
  }
}
EOF
    echo "  Created new mcp.json"
fi
echo -e "  ${GREEN}✅ Cursor MCP configured${NC}"
echo ""

# Step 7: Install Native Messaging for Chrome extension
echo -e "${YELLOW}Step 7: Installing Native Messaging hosts...${NC}"
"$PHLOEM_BIN" install-native 2>/dev/null || true
echo -e "  ${GREEN}✅ Native Messaging configured${NC}"
echo ""

# Step 8: Initialize data directory
echo -e "${YELLOW}Step 8: Initializing data directory...${NC}"
mkdir -p "$HOME/.phloem"
echo -e "  ${GREEN}✅ Created ~/.phloem${NC}"
echo ""

# Step 9: Verify installation
echo -e "${YELLOW}Step 9: Verifying installation...${NC}"
echo ""

# Check app
if [ -d "/Applications/Phloem.app" ]; then
    echo -e "  ${GREEN}✅ Phloem.app installed${NC}"
else
    echo -e "  ${RED}❌ Phloem.app not found${NC}"
fi

# Check binary architecture
BINARY_ARCH=$(file "/Applications/Phloem.app/Contents/Resources/phloem" | grep -o 'x86_64\|arm64')
if [ "$BINARY_ARCH" = "$ARCH" ] || [ "$BINARY_ARCH" = "x86_64" -a "$ARCH" = "x86_64" ]; then
    echo -e "  ${GREEN}✅ Binary architecture: $BINARY_ARCH (correct)${NC}"
else
    echo -e "  ${YELLOW}⚠️  Binary architecture: $BINARY_ARCH (expected $ARCH)${NC}"
fi

# Check MCP config
if grep -q '"phloem"' "$CURSOR_CONFIG" 2>/dev/null; then
    echo -e "  ${GREEN}✅ Cursor MCP configured${NC}"
else
    echo -e "  ${RED}❌ Cursor MCP not configured${NC}"
fi

# Check native messaging
NM_HOST="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts/com.canopyhq.phloem.json"
if [ -f "$NM_HOST" ]; then
    echo -e "  ${GREEN}✅ Chrome Native Messaging configured${NC}"
else
    echo -e "  ${YELLOW}⚠️  Chrome Native Messaging not found (extension won't work)${NC}"
fi

# Test binary
if "$PHLOEM_BIN" version >/dev/null 2>&1; then
    VERSION_OUT=$("$PHLOEM_BIN" version 2>&1)
    echo -e "  ${GREEN}✅ Binary works: $VERSION_OUT${NC}"
else
    echo -e "  ${RED}❌ Binary test failed${NC}"
fi

# Show status
echo ""
"$PHLOEM_BIN" status 2>&1 | sed 's/^/  /'

echo ""
echo "============================================"
echo -e "${GREEN}🎉 Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. ${YELLOW}Restart Cursor${NC} for MCP to take effect"
echo "  2. Launch Phloem: ${BLUE}open /Applications/Phloem.app${NC}"
echo "  3. Use 'remember' tool in Cursor to store memories"
echo ""
echo "To install Chrome extension:"
echo "  1. Open chrome://extensions"
echo "  2. Enable Developer mode"
echo "  3. Load unpacked: $PROJECT_DIR/extension"
echo ""
echo "To disable Cursor completion sounds:"
echo "  Settings → Features → Chat → 'Play sound on finish' → OFF"
echo ""
