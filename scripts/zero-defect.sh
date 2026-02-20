#!/bin/bash
# Phloem Zero-Defect Release Gate
# NOTHING ships without passing ALL of these checks
#
# Based on Lignin testing philosophy:
# - No mocks for critical paths
# - Real integration tests
# - Every failure is a learning opportunity

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAILED=0
PASSED=0

# Portable timeout (macOS doesn't ship GNU timeout)
if command -v timeout >/dev/null 2>&1; then
    TIMEOUT_CMD="timeout"
elif command -v gtimeout >/dev/null 2>&1; then
    TIMEOUT_CMD="gtimeout"
else
    echo "ERROR: 'timeout' not found. Install GNU coreutils: brew install coreutils"
    exit 1
fi

log_pass() {
    echo -e "${GREEN}✅ $1${NC}"
    PASSED=$((PASSED + 1))
}

log_fail() {
    echo -e "${RED}❌ $1${NC}"
    FAILED=$((FAILED + 1))
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_section() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# ============================================================================
# LAYER 1: COMPILATION
# ============================================================================
log_section "LAYER 1: COMPILATION"

echo "Building binary..."
if go build -o phloem . 2>&1; then
    log_pass "Binary builds successfully"
else
    log_fail "Binary failed to build"
fi

echo "Checking all packages compile..."
if go build ./... 2>&1; then
    log_pass "All packages compile"
else
    log_fail "Package compilation failed"
fi

# ============================================================================
# LAYER 2: STATIC ANALYSIS
# ============================================================================
log_section "LAYER 2: STATIC ANALYSIS"

echo "Running go vet..."
if go vet ./... 2>&1; then
    log_pass "go vet passed"
else
    log_fail "go vet found issues"
fi

echo "Running go fmt check..."
UNFMT=$(gofmt -l . 2>&1 | grep -v vendor || true)
if [ -z "$UNFMT" ]; then
    log_pass "Code is formatted"
else
    log_fail "Code needs formatting: $UNFMT"
fi

# ============================================================================
# LAYER 3: UNIT TESTS
# ============================================================================
log_section "LAYER 3: UNIT TESTS"

echo "Running unit tests (this may take a minute)..."
if go test -short ./... 2>&1; then
    log_pass "Unit tests passed"
else
    log_fail "Unit tests failed"
fi

# ============================================================================
# LAYER 4: MCP PROTOCOL COMPLIANCE
# ============================================================================
log_section "LAYER 4: MCP PROTOCOL COMPLIANCE"

echo "Testing MCP initialize..."
MCP_INIT=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | $TIMEOUT_CMD 5 ./phloem serve 2>/dev/null | head -1)
if echo "$MCP_INIT" | grep -q "protocolVersion"; then
    log_pass "MCP initialize returns protocol version"
else
    log_fail "MCP initialize failed"
fi

echo "Testing MCP tools/list..."
MCP_TOOLS=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | $TIMEOUT_CMD 5 ./phloem serve 2>/dev/null | head -1)
if echo "$MCP_TOOLS" | grep -q "remember"; then
    log_pass "MCP tools/list returns remember tool"
else
    log_fail "MCP tools/list failed"
fi

echo "Checking tool count..."
TOOL_COUNT=$(echo "$MCP_TOOLS" | python3 -c "import sys,json; data=json.load(sys.stdin); print(len(data.get('result',{}).get('tools',[])))" 2>/dev/null || echo "0")
if [ "$TOOL_COUNT" = "14" ]; then
    log_pass "MCP tools/list returns 14 tools"
else
    log_fail "MCP tools/list returned $TOOL_COUNT tools, expected 14"
fi

# ============================================================================
# LAYER 5: MEMORY OPERATIONS
# ============================================================================
log_section "LAYER 5: MEMORY OPERATIONS"

echo "Testing memory store and recall..."
# This uses MCP protocol
REMEMBER_TEST=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Zero defect test memory","tags":["test","zero-defect"]}}}' | $TIMEOUT_CMD 10 ./phloem serve 2>/dev/null | head -1)
if echo "$REMEMBER_TEST" | grep -q "remembered\|stored"; then
    log_pass "Memory remember works"
else
    log_fail "Memory remember failed: $REMEMBER_TEST"
fi

RECALL_TEST=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"recall","arguments":{"query":"zero defect test"}}}' | $TIMEOUT_CMD 10 ./phloem serve 2>/dev/null | head -1)
if echo "$RECALL_TEST" | grep -q "memories\|Zero defect"; then
    log_pass "Memory recall works"
else
    log_fail "Memory recall failed"
fi

# ============================================================================
# LAYER 6: PRIVACY VERIFICATION
# ============================================================================
log_section "LAYER 6: PRIVACY VERIFICATION"

echo "Checking phloem makes no network connections during tool call..."
PRIVACY_DIR=$(mktemp -d)
export PHLOEM_DATA_DIR="$PRIVACY_DIR"

# Start phloem serve in background, send a remember, then check for network sockets
REMEMBER_REQ='{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"remember","arguments":{"content":"privacy test memory"}}}'
echo "$REMEMBER_REQ" | $TIMEOUT_CMD 5 ./phloem serve >"$PRIVACY_DIR/out" 2>/dev/null &
PHLOEM_PID=$!
sleep 1

# Check for any network connections from this process
NET_CONNS=$(lsof -i -a -p "$PHLOEM_PID" 2>/dev/null | grep -v "^COMMAND" || true)
kill "$PHLOEM_PID" 2>/dev/null || true
wait "$PHLOEM_PID" 2>/dev/null || true
unset PHLOEM_DATA_DIR
rm -rf "$PRIVACY_DIR"

if [ -z "$NET_CONNS" ]; then
    log_pass "No network connections during tool call"
else
    log_fail "Phloem opened network connections: $NET_CONNS"
fi

# ============================================================================
# LAYER 7: SURFACE AREA
# ============================================================================
log_section "LAYER 7: SURFACE AREA"

echo "Running surface area test (every CLI command + MCP tool)..."
if "$SCRIPT_DIR/surface-test.sh" ./phloem 2>&1; then
    log_pass "Surface area test passed"
else
    log_fail "Surface area test failed"
fi

# ============================================================================
# SUMMARY
# ============================================================================
log_section "ZERO-DEFECT GATE SUMMARY"

TOTAL=$((PASSED + FAILED))
echo ""
echo "  Passed: $PASSED / $TOTAL"
echo "  Failed: $FAILED / $TOTAL"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 ZERO-DEFECT GATE PASSED${NC}"
    echo ""
    echo "This build is ready for release."
    exit 0
else
    echo -e "${RED}🚫 ZERO-DEFECT GATE FAILED${NC}"
    echo ""
    echo "DO NOT RELEASE. Fix the failures above first."
    echo ""
    echo "Zero defects. Every release."
    exit 1
fi
