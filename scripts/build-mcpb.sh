#!/bin/bash
# Build MCPB (MCP Bundle) for the official MCP registry
# Usage: ./scripts/build-mcpb.sh <version> <binary-path> <output-path>
# Example: ./scripts/build-mcpb.sh 0.1.1 dist/phloem phloem.mcpb

set -euo pipefail

VERSION="${1:?Usage: build-mcpb.sh <version> <binary-path> <output-path>}"
BINARY="${2:?Usage: build-mcpb.sh <version> <binary-path> <output-path>}"
OUTPUT="${3:?Usage: build-mcpb.sh <version> <binary-path> <output-path>}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# Create manifest.json
cat > "$TMPDIR/manifest.json" <<MANIFEST
{
  "manifest_version": "0.2.0",
  "name": "phloem",
  "display_name": "Phloem",
  "version": "$VERSION",
  "description": "Local-first AI memory with causal graphs",
  "long_description": "Persistent context for AI coding tools via MCP. Fully offline, zero config, your data never leaves your machine. Features causal memory graphs, citation verification with confidence decay, and semantic search.",
  "author": {
    "name": "Canopy HQ LLC",
    "url": "https://github.com/CanopyHQ"
  },
  "license": "Apache-2.0",
  "repository": {
    "type": "git",
    "url": "https://github.com/CanopyHQ/phloem"
  },
  "homepage": "https://github.com/CanopyHQ/phloem",
  "keywords": ["memory", "causal-graphs", "local-first", "ai-memory", "mcp"],
  "server": {
    "type": "binary",
    "entry_point": "server/phloem",
    "mcp_config": {
      "command": "./server/phloem",
      "args": ["serve"]
    }
  },
  "tools": [
    {"name": "remember", "description": "Store a memory for later recall"},
    {"name": "recall", "description": "Search memories by semantic similarity"},
    {"name": "forget", "description": "Delete a specific memory by ID"},
    {"name": "list_memories", "description": "List recent memories"},
    {"name": "memory_stats", "description": "Get memory store statistics"},
    {"name": "session_context", "description": "Load session context"},
    {"name": "add_citation", "description": "Link a memory to a code location"},
    {"name": "verify_citation", "description": "Check if a citation is still valid"},
    {"name": "get_citations", "description": "Get all citations for a memory"},
    {"name": "verify_memory", "description": "Verify memory via citations"},
    {"name": "causal_query", "description": "Query causal graph"},
    {"name": "compose", "description": "Compose two recall queries"},
    {"name": "prefetch", "description": "Preload memory suggestions"},
    {"name": "prefetch_suggest", "description": "Suggest memories for context"}
  ]
}
MANIFEST

# Copy binary
mkdir -p "$TMPDIR/server"
cp "$BINARY" "$TMPDIR/server/phloem"
chmod +x "$TMPDIR/server/phloem"

# Create the .mcpb ZIP bundle
(cd "$TMPDIR" && zip -r - manifest.json server/) > "$OUTPUT"

echo "Built $OUTPUT ($(du -h "$OUTPUT" | cut -f1))"
echo "SHA256: $(openssl dgst -sha256 "$OUTPUT" | awk '{print $NF}')"
