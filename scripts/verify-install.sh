#!/bin/bash
# Phloem Installation Verification
# Builds phloem and verifies setup commands produce correct configs.

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

cleanup() {
    if [ -n "$FAKE_HOME" ] && [ -d "$FAKE_HOME" ]; then
        rm -rf "$FAKE_HOME"
    fi
    if [ -f "$PROJECT_ROOT/phloem-verify" ]; then
        rm -f "$PROJECT_ROOT/phloem-verify"
    fi
}
trap cleanup EXIT

# ============================================================================
# BUILD
# ============================================================================
log_section "BUILD"

echo "Building phloem binary..."
if go build -o phloem-verify . 2>&1; then
    log_pass "Binary builds successfully"
else
    log_fail "Binary failed to build"
    echo "Cannot continue without a binary."
    exit 1
fi

PHLOEM_BIN="$PROJECT_ROOT/phloem-verify"

# Create a fake HOME with the binary in PATH
FAKE_HOME=$(mktemp -d)
mkdir -p "$FAKE_HOME/bin"
cp "$PHLOEM_BIN" "$FAKE_HOME/bin/phloem"
export PATH="$FAKE_HOME/bin:$PATH"

# ============================================================================
# CURSOR SETUP
# ============================================================================
log_section "CURSOR SETUP"

# Create .cursor dir in fake HOME
mkdir -p "$FAKE_HOME/.cursor"

echo "Running phloem setup cursor..."
if HOME="$FAKE_HOME" PHLOEM_DATA_DIR="$FAKE_HOME/.phloem" "$FAKE_HOME/bin/phloem" setup cursor > /dev/null 2>&1; then
    log_pass "Cursor setup exited successfully"
else
    log_fail "Cursor setup exited with error"
fi

CURSOR_CONFIG="$FAKE_HOME/.cursor/mcp.json"

# Check file exists
if [ -f "$CURSOR_CONFIG" ]; then
    log_pass "mcp.json was created"
else
    log_fail "mcp.json was not created"
fi

# Validate JSON
if python3 -c "import json; json.load(open('$CURSOR_CONFIG'))" 2>/dev/null; then
    log_pass "mcp.json is valid JSON"
else
    log_fail "mcp.json is not valid JSON"
fi

# Check for phloem server entry
if python3 -c "
import json, sys
config = json.load(open('$CURSOR_CONFIG'))
servers = config.get('mcpServers', {})
if 'phloem' not in servers:
    sys.exit(1)
p = servers['phloem']
if 'command' not in p or 'args' not in p:
    sys.exit(1)
" 2>/dev/null; then
    log_pass "mcp.json contains phloem server with command and args"
else
    log_fail "mcp.json missing phloem server or required fields"
fi

# ============================================================================
# WINDSURF SETUP
# ============================================================================
log_section "WINDSURF SETUP"

mkdir -p "$FAKE_HOME/.windsurf"

echo "Running phloem setup windsurf..."
if HOME="$FAKE_HOME" PHLOEM_DATA_DIR="$FAKE_HOME/.phloem" "$FAKE_HOME/bin/phloem" setup windsurf > /dev/null 2>&1; then
    log_pass "Windsurf setup exited successfully"
else
    log_fail "Windsurf setup exited with error"
fi

WINDSURF_CONFIG="$FAKE_HOME/.windsurf/mcp_config.json"

if [ -f "$WINDSURF_CONFIG" ]; then
    log_pass "mcp_config.json was created"
else
    log_fail "mcp_config.json was not created"
fi

if python3 -c "import json; json.load(open('$WINDSURF_CONFIG'))" 2>/dev/null; then
    log_pass "mcp_config.json is valid JSON"
else
    log_fail "mcp_config.json is not valid JSON"
fi

if python3 -c "
import json, sys
config = json.load(open('$WINDSURF_CONFIG'))
servers = config.get('mcpServers', {})
if 'phloem' not in servers:
    sys.exit(1)
p = servers['phloem']
if 'command' not in p or 'args' not in p:
    sys.exit(1)
" 2>/dev/null; then
    log_pass "mcp_config.json contains phloem server with command and args"
else
    log_fail "mcp_config.json missing phloem server or required fields"
fi

# ============================================================================
# IDEMPOTENCY
# ============================================================================
log_section "IDEMPOTENCY"

echo "Running cursor setup a second time..."
if HOME="$FAKE_HOME" PHLOEM_DATA_DIR="$FAKE_HOME/.phloem" "$FAKE_HOME/bin/phloem" setup cursor > /dev/null 2>&1; then
    log_pass "Cursor setup idempotent (no error on re-run)"
else
    log_fail "Cursor setup failed on second run"
fi

if python3 -c "import json; json.load(open('$CURSOR_CONFIG'))" 2>/dev/null; then
    log_pass "mcp.json still valid JSON after re-run"
else
    log_fail "mcp.json corrupted after re-run"
fi

echo "Running windsurf setup a second time..."
if HOME="$FAKE_HOME" PHLOEM_DATA_DIR="$FAKE_HOME/.phloem" "$FAKE_HOME/bin/phloem" setup windsurf > /dev/null 2>&1; then
    log_pass "Windsurf setup idempotent (no error on re-run)"
else
    log_fail "Windsurf setup failed on second run"
fi

if python3 -c "import json; json.load(open('$WINDSURF_CONFIG'))" 2>/dev/null; then
    log_pass "mcp_config.json still valid JSON after re-run"
else
    log_fail "mcp_config.json corrupted after re-run"
fi

# ============================================================================
# PRESERVATION
# ============================================================================
log_section "PRESERVATION OF EXISTING SERVERS"

echo "Adding a fake other-server to mcp.json..."
python3 -c "
import json
config = json.load(open('$CURSOR_CONFIG'))
config['mcpServers']['other-server'] = {'command': '/usr/bin/other', 'args': ['run']}
with open('$CURSOR_CONFIG', 'w') as f:
    json.dump(config, f, indent=2)
"

echo "Running cursor setup again..."
HOME="$FAKE_HOME" PHLOEM_DATA_DIR="$FAKE_HOME/.phloem" "$FAKE_HOME/bin/phloem" setup cursor > /dev/null 2>&1

if python3 -c "
import json, sys
config = json.load(open('$CURSOR_CONFIG'))
servers = config.get('mcpServers', {})
if 'other-server' not in servers:
    sys.exit(1)
if 'phloem' not in servers:
    sys.exit(1)
" 2>/dev/null; then
    log_pass "Cursor setup preserved existing other-server"
else
    log_fail "Cursor setup removed existing other-server"
fi

echo "Adding a fake other-server to mcp_config.json..."
python3 -c "
import json
config = json.load(open('$WINDSURF_CONFIG'))
config['mcpServers']['other-server'] = {'command': '/usr/bin/other', 'args': ['run']}
with open('$WINDSURF_CONFIG', 'w') as f:
    json.dump(config, f, indent=2)
"

echo "Running windsurf setup again..."
HOME="$FAKE_HOME" PHLOEM_DATA_DIR="$FAKE_HOME/.phloem" "$FAKE_HOME/bin/phloem" setup windsurf > /dev/null 2>&1

if python3 -c "
import json, sys
config = json.load(open('$WINDSURF_CONFIG'))
servers = config.get('mcpServers', {})
if 'other-server' not in servers:
    sys.exit(1)
if 'phloem' not in servers:
    sys.exit(1)
" 2>/dev/null; then
    log_pass "Windsurf setup preserved existing other-server"
else
    log_fail "Windsurf setup removed existing other-server"
fi

# ============================================================================
# SUMMARY
# ============================================================================
log_section "VERIFICATION SUMMARY"

TOTAL=$((PASSED + FAILED))
echo ""
echo "  Passed: $PASSED / $TOTAL"
echo "  Failed: $FAILED / $TOTAL"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 INSTALLATION VERIFICATION PASSED${NC}"
    exit 0
else
    echo -e "${RED}🚫 INSTALLATION VERIFICATION FAILED${NC}"
    echo ""
    echo "Fix the failures above before shipping."
    exit 1
fi
