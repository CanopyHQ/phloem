#!/bin/bash
# Quick setup for Phloem MCP in Cursor
# Run: curl -sL https://raw.githubusercontent.com/CanopyHQ/canopy/main/phloem/scripts/setup-cursor.sh | bash

set -e

echo "🧠 Setting up Phloem for Cursor..."

# Build Phloem
cd ~/Documents/GitHub/canopy/phloem
go build -o phloem-cli . 2>/dev/null || { echo "❌ Go build failed"; exit 1; }

# Get Gemini key from 1Password
op signin 2>/dev/null || true
GEMINI_KEY=$(op item get "Gemini API Key" --vault Canopy --fields label=credential --reveal 2>/dev/null)

if [ -z "$GEMINI_KEY" ]; then
  echo "⚠️  No Gemini key found - using local TF-IDF embeddings"
  GEMINI_KEY=""
fi

# Write MCP config
cat > ~/.cursor/mcp.json << EOF
{
  "mcpServers": {
    "phloem": {
      "command": "$HOME/Documents/GitHub/canopy/phloem/phloem-cli",
      "args": ["serve"],
      "env": {
        "GEMINI_API_KEY": "$GEMINI_KEY"
      }
    }
  }
}
EOF

echo "✅ Done! Restart Cursor to activate Phloem."
echo "📊 Database: $(ls -lh ~/.phloem/memories.db 2>/dev/null | awk '{print $5}') via iCloud"
