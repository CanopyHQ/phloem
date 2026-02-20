#!/bin/bash
# Phloem MCP Setup Script
# Builds and configures phloem-mcp for Cursor

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="phloem-mcp"

echo "🧠 Phloem MCP Setup"
echo "==================="
echo ""

# Build
echo "📦 Building..."
cd "$PROJECT_DIR"
go build -o "$BINARY_NAME" .
echo "✅ Built: $PROJECT_DIR/$BINARY_NAME"
echo ""

# Create Cursor config
CURSOR_CONFIG="$HOME/.cursor/mcp.json"
CURSOR_DIR="$HOME/.cursor"

echo "📝 Configuring Cursor..."

# Create .cursor directory if needed
mkdir -p "$CURSOR_DIR"

# Check if config exists
if [ -f "$CURSOR_CONFIG" ]; then
    echo "   Found existing $CURSOR_CONFIG"
    
    # Check if phloem is already configured
    if grep -q '"phloem"' "$CURSOR_CONFIG"; then
        echo "   ⚠️  phloem already configured in mcp.json"
        echo "   Please verify the path is correct:"
        grep -A 3 '"phloem"' "$CURSOR_CONFIG"
    else
        echo "   Adding phloem to existing config..."
        # This is simplified - in production use jq
        echo ""
        echo "   ⚠️  Please manually add this to your mcp.json servers:"
        echo ""
        echo '   "phloem": {'
        echo "     \"command\": \"$PROJECT_DIR/$BINARY_NAME\""
        echo '   }'
    fi
else
    echo "   Creating new $CURSOR_CONFIG..."
    cat > "$CURSOR_CONFIG" << EOF
{
  "mcpServers": {
    "phloem": {
      "command": "$PROJECT_DIR/$BINARY_NAME"
    }
  }
}
EOF
    echo "   ✅ Created Cursor MCP config"
fi

echo ""
echo "==================="
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Restart Cursor"
echo "  2. Try asking: 'Remember that this project uses PostgreSQL'"
echo "  3. Later ask: 'What database does this project use?'"
echo ""
echo "Data stored in: ~/.phloem/memories.db"
echo ""
