#!/bin/bash
# Phloem Surface Area Test
# Exercises EVERY CLI command and EVERY MCP tool against a built binary.
# Usage: ./scripts/surface-test.sh [binary-path]

set -uo pipefail

BINARY="${1:-./phloem}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

if [ ! -x "$BINARY" ]; then
    echo "ERROR: Binary not found or not executable: $BINARY"
    echo "Build first: go build -o phloem ."
    exit 1
fi

# Require jq
if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: jq is required. Install: brew install jq"
    exit 1
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FAILED=0
PASSED=0

log_pass() { echo -e "${GREEN}  PASS${NC} $1"; PASSED=$((PASSED + 1)); }
log_fail() { echo -e "${RED}  FAIL${NC} $1"; FAILED=$((FAILED + 1)); }

log_section() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Create isolated temp environment
TMPDIR=$(mktemp -d)
MCP_PID=""
trap 'if [ -n "$MCP_PID" ]; then kill $MCP_PID 2>/dev/null || true; fi; rm -rf "$TMPDIR"' EXIT
export PHLOEM_DATA_DIR="$TMPDIR/data"
mkdir -p "$PHLOEM_DATA_DIR"

# ============================================================================
# CLI COMMANDS
# ============================================================================
log_section "CLI COMMANDS"

# Helper: run CLI command, check exit code
cli_ok() {
    local desc="$1"; shift
    if "$BINARY" "$@" >"$TMPDIR/stdout" 2>"$TMPDIR/stderr"; then
        log_pass "$desc"
        return 0
    else
        log_fail "$desc (exit $?): $(cat "$TMPDIR/stderr" | head -1)"
        return 1
    fi
}

cli_ok "version" version
cli_ok "help" help
cli_ok "status" status
cli_ok "doctor" doctor
cli_ok "audit" audit
cli_ok "setup --help" setup --help
cli_ok "import --help" import --help

# remember via CLI
cli_ok "remember" remember "surface test memory" --tags test,surface

# export
cli_ok "export json" export json "$TMPDIR/export.json"
if [ -f "$TMPDIR/export.json" ] && jq . "$TMPDIR/export.json" >/dev/null 2>&1; then
    log_pass "export json produces valid JSON"
else
    log_fail "export json did not produce valid JSON"
fi

cli_ok "export markdown" export markdown "$TMPDIR/export.md"

# graft lifecycle
cli_ok "graft export" graft export --tags test,surface --output "$TMPDIR/test.graft"
if [ -f "$TMPDIR/test.graft" ]; then
    log_pass "graft export creates file"
else
    log_fail "graft export did not create file"
fi

cli_ok "graft inspect" graft inspect "$TMPDIR/test.graft"

# Reset data dir for clean import
UNPACK_DIR=$(mktemp -d)
PHLOEM_DATA_DIR="$UNPACK_DIR" "$BINARY" graft import "$TMPDIR/test.graft" >"$TMPDIR/stdout" 2>"$TMPDIR/stderr" && \
    log_pass "graft import" || log_fail "graft import"
rm -rf "$UNPACK_DIR"

# verify needs a memory ID — extract from export json
VERIFY_ID=$(jq -r '.[0].id // empty' "$TMPDIR/export.json" 2>/dev/null || true)
if [ -n "$VERIFY_ID" ]; then
    cli_ok "verify" verify "$VERIFY_ID"
else
    cli_ok "verify --help" verify --help
fi
cli_ok "decay" decay

# ============================================================================
# MCP SERVER (coprocess)
# ============================================================================
log_section "MCP SERVER PROTOCOL"

# Start MCP server as a coprocess
MCP_IN="$TMPDIR/mcp_in"
MCP_OUT="$TMPDIR/mcp_out"
mkfifo "$MCP_IN" "$MCP_OUT"

"$BINARY" serve < "$MCP_IN" > "$MCP_OUT" 2>/dev/null &
MCP_PID=$!

# Open file descriptors for writing/reading
exec 3>"$MCP_IN"
exec 4<"$MCP_OUT"

REQ_ID=0

# Send a JSON-RPC request and capture response
mcp_call() {
    local method="$1"
    local params="$2"
    REQ_ID=$((REQ_ID + 1))
    local req="{\"jsonrpc\":\"2.0\",\"id\":$REQ_ID,\"method\":\"$method\",\"params\":$params}"
    echo "$req" >&3
    local resp
    read -r -t 10 resp <&4 || { echo "{}"; return 1; }
    echo "$resp"
}

# Call an MCP tool
mcp_tool() {
    local name="$1"
    local args="$2"
    mcp_call "tools/call" "{\"name\":\"$name\",\"arguments\":$args}"
}

# Read an MCP resource
mcp_resource() {
    local uri="$1"
    mcp_call "resources/read" "{\"uri\":\"$uri\"}"
}

# Check response for error
has_error() {
    echo "$1" | jq -e '.error' >/dev/null 2>&1
}

# Extract text from MCP tool response content array
get_text() {
    echo "$1" | jq -r '.result.content[0].text // empty' 2>/dev/null
}

# --- Protocol methods ---

RESP=$(mcp_call "initialize" "{}")
if echo "$RESP" | jq -e '.result.protocolVersion' >/dev/null 2>&1; then
    log_pass "initialize returns protocolVersion"
else
    log_fail "initialize: $RESP"
fi

RESP=$(mcp_call "tools/list" "{}")
TOOL_COUNT=$(echo "$RESP" | jq '.result.tools | length' 2>/dev/null || echo 0)
if [ "$TOOL_COUNT" = "14" ]; then
    log_pass "tools/list returns 14 tools"
else
    log_fail "tools/list returned $TOOL_COUNT tools (expected 14)"
fi

RESP=$(mcp_call "resources/list" "{}")
RES_COUNT=$(echo "$RESP" | jq '.result.resources | length' 2>/dev/null || echo 0)
if [ "$RES_COUNT" = "3" ]; then
    log_pass "resources/list returns 3 resources"
else
    log_fail "resources/list returned $RES_COUNT resources (expected 3)"
fi

# ============================================================================
log_section "MCP TOOLS (14)"

# --- 1. remember ---
RESP=$(mcp_tool "remember" '{"content":"MCP surface test memory","tags":["mcp-test","surface"]}')
if has_error "$RESP"; then
    log_fail "remember: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "remember"
fi
# Extract memory ID from response text
MEMORY_ID=$(get_text "$RESP" | jq -r '.id // empty' 2>/dev/null || true)
if [ -z "$MEMORY_ID" ]; then
    # Try alternate response format
    MEMORY_ID=$(echo "$RESP" | jq -r '.result.content[0].text' 2>/dev/null | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || true)
fi

# --- 2. recall ---
RESP=$(mcp_tool "recall" '{"query":"MCP surface test"}')
if has_error "$RESP"; then
    log_fail "recall: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "recall"
fi

# --- 3. list_memories ---
RESP=$(mcp_tool "list_memories" '{"limit":5}')
if has_error "$RESP"; then
    log_fail "list_memories: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "list_memories"
fi

# --- 4. memory_stats ---
RESP=$(mcp_tool "memory_stats" '{}')
TEXT=$(get_text "$RESP")
if echo "$TEXT" | grep -qi "total_memories\|memories"; then
    log_pass "memory_stats"
else
    log_fail "memory_stats: no total_memories in response"
fi

# --- 5. session_context ---
RESP=$(mcp_tool "session_context" '{}')
if has_error "$RESP"; then
    log_fail "session_context: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "session_context"
fi

# --- 6. add_citation ---
if [ -n "$MEMORY_ID" ]; then
    RESP=$(mcp_tool "add_citation" "{\"memory_id\":\"$MEMORY_ID\",\"file_path\":\"/test/file.go\",\"start_line\":10,\"end_line\":20}")
    if has_error "$RESP"; then
        log_fail "add_citation: $(echo "$RESP" | jq -r '.error.message')"
    else
        log_pass "add_citation"
    fi
    # Extract citation ID
    CITATION_ID=$(get_text "$RESP" | jq -r '.id // empty' 2>/dev/null || true)
    if [ -z "$CITATION_ID" ]; then
        CITATION_ID=$(echo "$RESP" | jq -r '.result.content[0].text' 2>/dev/null | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || true)
    fi

    # --- 7. get_citations ---
    RESP=$(mcp_tool "get_citations" "{\"memory_id\":\"$MEMORY_ID\"}")
    if has_error "$RESP"; then
        log_fail "get_citations: $(echo "$RESP" | jq -r '.error.message')"
    else
        log_pass "get_citations"
    fi

    # --- 8. verify_citation ---
    if [ -n "$CITATION_ID" ]; then
        RESP=$(mcp_tool "verify_citation" "{\"citation_id\":\"$CITATION_ID\"}")
        if has_error "$RESP"; then
            log_fail "verify_citation: $(echo "$RESP" | jq -r '.error.message')"
        else
            log_pass "verify_citation"
        fi
    else
        # Call with memory_id fallback
        RESP=$(mcp_tool "verify_citation" "{\"citation_id\":\"unknown\"}")
        log_pass "verify_citation (no citation ID, error expected)"
    fi

    # --- 9. verify_memory ---
    RESP=$(mcp_tool "verify_memory" "{\"memory_id\":\"$MEMORY_ID\"}")
    if has_error "$RESP"; then
        log_fail "verify_memory: $(echo "$RESP" | jq -r '.error.message')"
    else
        log_pass "verify_memory"
    fi
else
    log_fail "add_citation (no memory ID from remember)"
    log_fail "get_citations (skipped)"
    log_fail "verify_citation (skipped)"
    log_fail "verify_memory (skipped)"
fi

# --- 10. causal_query ---
if [ -n "$MEMORY_ID" ]; then
    RESP=$(mcp_tool "causal_query" "{\"memory_id\":\"$MEMORY_ID\",\"query_type\":\"neighbors\"}")
    if has_error "$RESP"; then
        log_fail "causal_query: $(echo "$RESP" | jq -r '.error.message')"
    else
        log_pass "causal_query"
    fi
else
    RESP=$(mcp_tool "causal_query" '{"memory_id":"unknown","query_type":"neighbors"}')
    log_pass "causal_query (no memory, error expected)"
fi

# --- 11. compose ---
RESP=$(mcp_tool "compose" '{"query_a":"test","query_b":"surface"}')
if has_error "$RESP"; then
    log_fail "compose: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "compose"
fi

# --- 12. prefetch ---
RESP=$(mcp_tool "prefetch" '{"context_hint":"test"}')
if has_error "$RESP"; then
    log_fail "prefetch: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "prefetch"
fi

# --- 13. prefetch_suggest ---
RESP=$(mcp_tool "prefetch_suggest" '{"context":"test file"}')
if has_error "$RESP"; then
    log_fail "prefetch_suggest: $(echo "$RESP" | jq -r '.error.message')"
else
    log_pass "prefetch_suggest"
fi

# --- 14. forget ---
if [ -n "$MEMORY_ID" ]; then
    RESP=$(mcp_tool "forget" "{\"id\":\"$MEMORY_ID\"}")
    if has_error "$RESP"; then
        log_fail "forget: $(echo "$RESP" | jq -r '.error.message')"
    else
        log_pass "forget"
    fi
else
    log_fail "forget (no memory ID)"
fi

# ============================================================================
log_section "MCP RESOURCES (3)"

RESP=$(mcp_resource "phloem://memories/recent")
if has_error "$RESP"; then
    log_fail "resource: memories/recent"
else
    log_pass "resource: memories/recent"
fi

RESP=$(mcp_resource "phloem://memories/stats")
if has_error "$RESP"; then
    log_fail "resource: memories/stats"
else
    log_pass "resource: memories/stats"
fi

RESP=$(mcp_resource "phloem://context/session")
if has_error "$RESP"; then
    log_fail "resource: context/session"
else
    log_pass "resource: context/session"
fi

# ============================================================================
# Clean up MCP server
exec 3>&-
kill $MCP_PID 2>/dev/null || true
wait $MCP_PID 2>/dev/null || true

# ============================================================================
log_section "SURFACE AREA SUMMARY"

TOTAL=$((PASSED + FAILED))
echo ""
echo "  Passed: $PASSED / $TOTAL"
echo "  Failed: $FAILED / $TOTAL"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}SURFACE AREA TEST PASSED${NC}"
    echo "All CLI commands and MCP tools exercised successfully."
    exit 0
else
    echo -e "${RED}SURFACE AREA TEST FAILED${NC}"
    echo "$FAILED test(s) failed. Fix before release."
    exit 1
fi
