#!/bin/bash
# Phloem Auto-Discovery and Setup Script
# Zero-configuration setup for Phloem access (local + cloud)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PHLOEM_DIR="$HOME/.phloem"
CONFIG_FILE="$PHLOEM_DIR/config.json"
LOCAL_DB="$PHLOEM_DIR/memories.db"

echo "🧠 Phloem Auto-Setup"
echo "===================="
echo ""

# Create .phloem directory if needed
mkdir -p "$PHLOEM_DIR"

# Step 1: Auto-detect local database
echo "📂 Step 1: Detecting local database..."
if [ -f "$LOCAL_DB" ]; then
    # Count memories in local DB
    MEMORY_COUNT=$(sqlite3 "$LOCAL_DB" "SELECT COUNT(*) FROM memories;" 2>/dev/null || echo "0")
    echo "   ✅ Found local database: $LOCAL_DB"
    echo "   📊 Memories: $MEMORY_COUNT"
    LOCAL_AVAILABLE="true"
else
    echo "   ⚠️  No local database found at $LOCAL_DB"
    echo "   💡 Run 'phloem-cli' to create one"
    LOCAL_AVAILABLE="false"
    MEMORY_COUNT="0"
fi
echo ""

# Step 2: Auto-detect cloud API
echo "🌐 Step 2: Detecting cloud API..."
CLOUD_URLS=(
    "https://canopy-dr-crown-20260125.fly.dev/api/phloem"
    "https://canopyhq.fly.dev/api/phloem"
    "http://localhost:8080/api/phloem"
)

CLOUD_URL=""
CLOUD_AVAILABLE="false"
API_KEY="duncan-canopy-primary"  # Default API key

for url in "${CLOUD_URLS[@]}"; do
    echo "   Trying: $url"
    if curl -s -f -m 2 "$url/stats" -H "X-API-Key: $API_KEY" > /dev/null 2>&1; then
        CLOUD_URL="$url"
        CLOUD_AVAILABLE="true"
        echo "   ✅ Found cloud API: $url"
        
        # Get cloud stats
        CLOUD_STATS=$(curl -s "$url/stats" -H "X-API-Key: $API_KEY")
        CLOUD_MEMORIES=$(echo "$CLOUD_STATS" | grep -o '"total_memories":[0-9]*' | cut -d':' -f2 || echo "0")
        echo "   📊 Cloud memories: $CLOUD_MEMORIES"
        break
    else
        echo "      ❌ Not available"
    fi
done

if [ "$CLOUD_AVAILABLE" = "false" ]; then
    echo "   ⚠️  No cloud API detected"
    echo "   💡 Cloud sync will not be available"
    CLOUD_MEMORIES="0"
fi
echo ""

# Step 3: Generate config
echo "⚙️  Step 3: Generating config..."

cat > "$CONFIG_FILE" << EOF
{
  "version": "1.0",
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "local": {
    "enabled": $LOCAL_AVAILABLE,
    "database_path": "$LOCAL_DB",
    "memory_count": $MEMORY_COUNT
  },
  "cloud": {
    "enabled": $CLOUD_AVAILABLE,
    "api_url": "$CLOUD_URL",
    "api_key": "$API_KEY",
    "memory_count": $CLOUD_MEMORIES
  },
  "access": {
    "mode": "unified",
    "prefer_local": true,
    "auto_sync": true
  }
}
EOF

echo "   ✅ Config generated: $CONFIG_FILE"
echo ""

# Step 4: Summary
echo "===================="
echo "✅ Auto-Setup Complete!"
echo ""
echo "Configuration Summary:"
echo "  Local DB:    $LOCAL_AVAILABLE ($MEMORY_COUNT memories)"
echo "  Cloud API:   $CLOUD_AVAILABLE ($CLOUD_MEMORIES memories)"
echo "  Config:      $CONFIG_FILE"
echo ""

if [ "$LOCAL_AVAILABLE" = "true" ] && [ "$CLOUD_AVAILABLE" = "true" ]; then
    echo "🎉 Full access available (local + cloud)"
    echo ""
    echo "Next steps:"
    echo "  1. Access via HTTP API: curl $CLOUD_URL/stats -H 'X-API-Key: $API_KEY'"
    echo "  2. Access via local DB: sqlite3 $LOCAL_DB"
    echo "  3. Use unified access layer: phloem/internal/access/unified.go ✅"
elif [ "$LOCAL_AVAILABLE" = "true" ]; then
    echo "📂 Local-only access available"
    echo ""
    echo "Next steps:"
    echo "  1. Access via local DB: sqlite3 $LOCAL_DB"
    echo "  2. Set up cloud sync for remote access"
elif [ "$CLOUD_AVAILABLE" = "true" ]; then
    echo "🌐 Cloud-only access available"
    echo ""
    echo "Next steps:"
    echo "  1. Access via HTTP API: curl $CLOUD_URL/stats -H 'X-API-Key: $API_KEY'"
    echo "  2. Create local DB for offline access: phloem-cli"
else
    echo "⚠️  No Phloem access detected"
    echo ""
    echo "Next steps:"
    echo "  1. Install phloem-cli: make install"
    echo "  2. Initialize local DB: phloem-cli"
    echo "  3. Or connect to cloud: export PHLOEM_API_URL=<url>"
fi

echo ""
echo "Config file: $CONFIG_FILE"
echo ""
