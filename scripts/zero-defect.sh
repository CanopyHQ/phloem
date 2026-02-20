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
if go build -o phloem-mcp . 2>&1; then
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
# LAYER 4: INTEGRATION TESTS - NATIVE MESSAGING
# ============================================================================
log_section "LAYER 4: NATIVE MESSAGING INTEGRATION"

echo "Testing native messaging protocol..."

# This is the test that would have caught the bug we shipped
NATIVE_TEST=$(python3 -c '
import struct
import json
import subprocess
import sys

msg = json.dumps({"action": "ping"}).encode("utf-8")
length = struct.pack("<I", len(msg))

proc = subprocess.Popen(
    ["./phloem-mcp", "native-messaging"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
)

proc.stdin.write(length + msg)
proc.stdin.flush()

resp_length_bytes = proc.stdout.read(4)
if len(resp_length_bytes) == 4:
    resp_length = struct.unpack("<I", resp_length_bytes)[0]
    response = proc.stdout.read(resp_length).decode("utf-8")
    data = json.loads(response)
    if data.get("success") and data.get("data") == "pong":
        print("PASS")
        sys.exit(0)
    else:
        print(f"FAIL: unexpected response: {response}")
        sys.exit(1)
else:
    stderr = proc.stderr.read().decode("utf-8")
    print(f"FAIL: no response, stderr: {stderr}")
    sys.exit(1)

proc.terminate()
' 2>&1)

if [ "$NATIVE_TEST" = "PASS" ]; then
    log_pass "Native messaging ping/pong works"
else
    log_fail "Native messaging failed: $NATIVE_TEST"
fi

# Test store_conversation action
echo "Testing store_conversation action..."
STORE_TEST=$(python3 -c '
import struct
import json
import subprocess
import sys

msg = json.dumps({
    "action": "store_conversation",
    "data": {
        "source": "test",
        "title": "Zero Defect Test",
        "url": "http://test.local",
        "messages": [
            {"role": "user", "content": "test message"},
            {"role": "assistant", "content": "test response"}
        ]
    }
}).encode("utf-8")
length = struct.pack("<I", len(msg))

proc = subprocess.Popen(
    ["./phloem-mcp", "native-messaging"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
)

proc.stdin.write(length + msg)
proc.stdin.flush()

resp_length_bytes = proc.stdout.read(4)
if len(resp_length_bytes) == 4:
    resp_length = struct.unpack("<I", resp_length_bytes)[0]
    response = proc.stdout.read(resp_length).decode("utf-8")
    data = json.loads(response)
    if data.get("success"):
        print("PASS")
        sys.exit(0)
    else:
        print(f"FAIL: {data.get(\"error\", \"unknown error\")}")
        sys.exit(1)
else:
    print("FAIL: no response")
    sys.exit(1)

proc.terminate()
' 2>&1)

if [ "$STORE_TEST" = "PASS" ]; then
    log_pass "Native messaging store_conversation works"
else
    log_fail "Native messaging store_conversation failed: $STORE_TEST"
fi

# Test stats action
echo "Testing stats action..."
STATS_TEST=$(python3 -c '
import struct
import json
import subprocess
import sys

msg = json.dumps({"action": "stats"}).encode("utf-8")
length = struct.pack("<I", len(msg))

proc = subprocess.Popen(
    ["./phloem-mcp", "native-messaging"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
)

proc.stdin.write(length + msg)
proc.stdin.flush()

resp_length_bytes = proc.stdout.read(4)
if len(resp_length_bytes) == 4:
    resp_length = struct.unpack("<I", resp_length_bytes)[0]
    response = proc.stdout.read(resp_length).decode("utf-8")
    data = json.loads(response)
    if data.get("success") and "total_memories" in str(data.get("data", {})):
        print("PASS")
        sys.exit(0)
    else:
        print(f"FAIL: {response}")
        sys.exit(1)
else:
    print("FAIL: no response")
    sys.exit(1)

proc.terminate()
' 2>&1)

if [ "$STATS_TEST" = "PASS" ]; then
    log_pass "Native messaging stats works"
else
    log_fail "Native messaging stats failed: $STATS_TEST"
fi

# ============================================================================
# LAYER 5: MCP PROTOCOL COMPLIANCE
# ============================================================================
log_section "LAYER 5: MCP PROTOCOL COMPLIANCE"

echo "Testing MCP initialize..."
MCP_INIT=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | timeout 5 ./phloem-mcp serve 2>/dev/null | head -1)
if echo "$MCP_INIT" | grep -q "protocolVersion"; then
    log_pass "MCP initialize returns protocol version"
else
    log_fail "MCP initialize failed"
fi

echo "Testing MCP tools/list..."
MCP_TOOLS=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | timeout 5 ./phloem-mcp serve 2>/dev/null | head -1)
if echo "$MCP_TOOLS" | grep -q "remember"; then
    log_pass "MCP tools/list returns remember tool"
else
    log_fail "MCP tools/list failed"
fi

# ============================================================================
# LAYER 6: EXTENSION MANIFEST VALIDATION
# ============================================================================
log_section "LAYER 6: EXTENSION VALIDATION"

echo "Checking extension manifest..."
if [ -f "extension/manifest.json" ]; then
    if python3 -c "import json; json.load(open('extension/manifest.json'))" 2>&1; then
        log_pass "Extension manifest is valid JSON"
    else
        log_fail "Extension manifest is invalid JSON"
    fi
    
    # Check required fields
    if grep -q '"manifest_version": 3' extension/manifest.json; then
        log_pass "Extension uses manifest v3"
    else
        log_fail "Extension not using manifest v3"
    fi
    
    if grep -q '"nativeMessaging"' extension/manifest.json; then
        log_pass "Extension requests nativeMessaging permission"
    else
        log_fail "Extension missing nativeMessaging permission"
    fi
else
    log_fail "Extension manifest not found"
fi

echo "Checking content scripts exist..."
for script in content-chatgpt.js content-claude.js content-gemini.js; do
    if [ -f "extension/$script" ]; then
        log_pass "Content script exists: $script"
    else
        log_fail "Content script missing: $script"
    fi
done

# ============================================================================
# LAYER 7: LICENSE SYSTEM
# ============================================================================
log_section "LAYER 7: LICENSE SYSTEM"

echo "Testing license command..."
if ./phloem-mcp license 2>&1 | grep -q "Tier:"; then
    log_pass "License command works"
else
    log_fail "License command failed"
fi

# ============================================================================
# LAYER 8: MEMORY OPERATIONS
# ============================================================================
log_section "LAYER 8: MEMORY OPERATIONS"

echo "Testing memory store and recall..."
# This uses MCP protocol
REMEMBER_TEST=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Zero defect test memory","tags":["test","zero-defect"]}}}' | timeout 10 ./phloem-mcp serve 2>/dev/null | head -1)
if echo "$REMEMBER_TEST" | grep -q "remembered\|stored"; then
    log_pass "Memory remember works"
else
    log_fail "Memory remember failed: $REMEMBER_TEST"
fi

RECALL_TEST=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"recall","arguments":{"query":"zero defect test"}}}' | timeout 10 ./phloem-mcp serve 2>/dev/null | head -1)
if echo "$RECALL_TEST" | grep -q "memories\|Zero defect"; then
    log_pass "Memory recall works"
else
    log_fail "Memory recall failed"
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
    echo "Remember: We shipped a broken extension because we skipped this."
    echo "Never again."
    exit 1
fi
